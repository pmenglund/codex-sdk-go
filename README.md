# Codex Go SDK

[![Go Reference](https://pkg.go.dev/badge/github.com/pmenglund/codex-sdk-go.svg)](https://pkg.go.dev/github.com/pmenglund/codex-sdk-go)
[![Codecov](https://codecov.io/gh/pmenglund/codex-sdk-go/graph/badge.svg)](https://codecov.io/gh/pmenglund/codex-sdk-go)

Embed the Codex app-server in Go workflows.

This SDK speaks JSON-RPC to the `codex app-server` process. By default it spawns the CLI and communicates over stdio.

## Requirements

- Go 1.25.14 or newer
- `codex` available on your `PATH`

## Install

```bash
go get github.com/pmenglund/codex-sdk-go
```

## Quickstart

```go
package main

import (
    "context"
    "fmt"
    "log/slog"
    "os"

    "github.com/pmenglund/codex-sdk-go"
)

func main() {
    ctx := context.Background()
    logger := slog.New(slog.NewTextHandler(os.Stderr, &slog.HandlerOptions{Level: slog.LevelInfo}))
    prompt := "Diagnose the test failure and propose a fix"

    client, err := codex.New(ctx, codex.Options{Logger: logger})
    if err != nil {
        panic(err)
    }
    defer client.Close()

    thread, err := client.StartThread(ctx, codex.ThreadStartOptions{})
    if err != nil {
        panic(err)
    }

    result, err := thread.Run(ctx, prompt, nil)
    if err != nil {
        panic(err)
    }

    fmt.Println(result.FinalResponse)
}
```

`New` uses its `context.Context` for initialization requests (`initialize`/`initialized`).
After `New` returns successfully, the spawned app-server lifetime is managed by `Close`, so canceling the constructor context later does not terminate the process.

## Streaming

Use `RunStreamed` to receive notifications as the turn progresses.

```go
prompt := "Inspect the repo"
stream, err := thread.RunStreamed(ctx, []codex.Input{codex.TextInput(prompt)}, nil)
if err != nil {
    panic(err)
}

defer stream.Close()

for {
    note, err := stream.Next(ctx)
    if err != nil {
		panic(err)
    }
    fmt.Printf("%s\n", note.Method)
	if note.Method == "turn/failed" {
		panic("turn failed")
	}
	if note.Method == "turn/completed" {
		if completed, ok := note.Params.(protocol.TurnCompletedNotification); ok &&
			completed.Turn != nil && completed.Turn.Status == protocol.TurnStatusFailed {
			panic(fmt.Errorf("turn failed: %v", completed.Turn.Error))
		}
		break
    }
}
```

`RunStreamed` returns thread-scoped events plus notifications that omit `threadId` (for example account/session updates) so global events are not silently dropped.

## Turn handles

Use `StartTurn` when you need to steer or interrupt a running turn.

```go
handle, err := thread.StartTurn(ctx, []codex.Input{codex.TextInput("Inspect the repo")}, nil)
if err != nil {
    panic(err)
}

if _, err := handle.Steer(ctx, []codex.Input{codex.TextInput("Focus on tests")}); err != nil {
    panic(err)
}

result, err := handle.Run(ctx)
if err != nil {
    panic(err)
}

fmt.Println(result.FinalResponse)
```

`TurnHandle` owns its notification subscription. Call `Close` if you stop before `Run` returns.

Choose exactly one consumption style per handle: `Run`, repeated `Next` calls,
or the stream returned by `Stream`. Mixing styles or consuming concurrently
returns `codex.ErrTurnConsumptionMode` immediately.

If a turn fails, `Run` returns both the partial result and a `*codex.TurnError`.
Use `errors.Is(err, codex.ErrTurnFailed)` to classify the failure and
`errors.As` to inspect its structured details, retry metadata, and raw payload:

```go
result, err := handle.Run(ctx)
if errors.Is(err, codex.ErrTurnFailed) {
    var turnErr *codex.TurnError
    if errors.As(err, &turnErr) {
        fmt.Printf("received %d items before failure: %v\n", len(result.Items), turnErr.Detail)
    }
}
```

`FinalResponse` is populated only by completed `agentMessage` items. Plan,
reasoning, and commentary items remain available in `Items` but are never
reported as the final assistant answer.

Canceling or expiring the context passed to `Run`, `RunInputs`, or
`TurnHandle.Run` best-effort interrupts the remote turn before returning the
original context error. Cleanup is bounded to two seconds. To detach without
interrupting remote work, use `RunStreamed` (or `StartTurn` plus manual `Next`)
and close only the local stream or handle.

Low-level `rpc.Client.SubscribeNotifications` buffers at most the requested number
of pending notifications (64 by default). If a consumer falls behind, only that
iterator closes and `Next` returns an `rpc.NotificationOverflowError`; JSON-RPC
responses and other subscribers continue normally.

```go
note, err := iterator.Next(ctx)
var overflow *rpc.NotificationOverflowError
if errors.As(err, &overflow) {
    // Events were lost. Increase the capacity, drain faster, and subscribe
    // again; overflow.Capacity reports the exhausted hard limit.
}
```

## Account, models, and threads

High-level helpers wrap common app-server operations without requiring direct JSON-RPC calls.

```go
account, err := client.Account(ctx, codex.AccountOptions{})
models, err := client.ListModels(ctx, codex.ListModelsOptions{})
threads, err := client.ListThreads(ctx, codex.ThreadListOptions{})
```

Thread list, read, fork, and unarchive helpers return concrete protocol values.
List sorting uses `protocol.SortDirection` and `protocol.ThreadSortKey` rather
than arbitrary strings.

Thread values also expose lifecycle helpers:

```go
if _, err := thread.SetName(ctx, "Investigation"); err != nil {
    panic(err)
}

forked, _, err := thread.Fork(ctx, codex.ThreadForkOptions{})
if err != nil {
    panic(err)
}

_ = forked
```

For lower-level or less stable protocol features, use `client.Client()` and the generated `rpc` package.

### Low-level union migration

Discriminated protocol unions are concrete generated wrapper types rather than
`interface{}`. Construct values with functions such as
`protocol.NewUserInput`, inspect the typed `Kind`, and use `RawJSON` when a
variant-specific payload is needed:

```go
input, err := protocol.NewUserInput(map[string]any{
    "type": "text",
    "text": "Inspect the repository",
})
if err != nil {
    return err
}
if input.Kind() == protocol.UserInputKindText {
    var payload struct {
        Text string `json:"text"`
    }
    if err := json.Unmarshal(input.RawJSON(), &payload); err != nil {
        return err
    }
}
```

Constructors validate JSON encoding, the required non-empty discriminator, and
required-field presence for known variants. Field value types and other schema
constraints remain server-validated. Known variants have generated kind
constants; a well-formed future discriminator remains round-trippable and
reports `IsKnown() == false`. Existing opaque schemas remain `interface{}` only
through reviewed generator allowlists, and generation fails when a new weak
fallback appears. The 31 affected wrappers—including `UserInput`,
`ResponseItem`, `SandboxPolicy`, `ThreadStatus`, and `LoginAccountParams`—are
listed in `protocol/unions_gen.go`.

Generated protocol declarations use Go initialisms consistently, including
`MCP`, `OAuth`, `ID`, `URL`, and `JSON`. Older `Mcp...`/`Oauth...` spellings and
`Sanitized...JSON` implementation names remain deprecated aliases for a
migration window. Prefer canonical declarations such as
`protocol.MCPServerOAuthLoginParams` and canonical methods such as
`rpc.Client.MCPServerOAuthLogin`.

These API changes are currently unreleased and require the first SDK release
after `v0.145.0`; `protocol.GeneratedCodexVersion` continues to describe the
upstream wire schema, not the SDK release number. This API-hardening work also
includes `Options.CompatibilityPolicy` and changes `StartLogin` to accept
JSON-marshalable login parameters; code using unkeyed `Options` literals,
interface method sets, or assigned `StartLogin` method values must be updated.

## Approvals

Configure app-server requests with optional callbacks. Unset callbacks return
`rpc.ErrServerRequestUnsupported`, so adding a future server request does not
break existing applications at compile time.

```go
logger := slog.New(slog.NewTextHandler(os.Stderr, &slog.HandlerOptions{Level: slog.LevelInfo}))
client, err := codex.New(ctx, codex.Options{
    Logger:         logger,
    RequestHandler: codex.RejectingApprovalCallbacks(),
})
```

For custom logic, set only the callbacks your application supports:

```go
callbacks := &codex.ServerRequestCallbacks{
    ApproveFileChange: func(ctx context.Context, request protocol.FileChangeRequestApprovalParams) (*protocol.FileChangeRequestApprovalResponse, error) {
        return &protocol.FileChangeRequestApprovalResponse{
            Decision: protocol.FileChangeApprovalDecisionDecline,
        }, nil
    },
}
client, err := codex.New(ctx, codex.Options{RequestHandler: callbacks})
```

`Options.ApprovalHandler` and the broad generated `rpc.ServerRequestHandler`
remain as deprecated compatibility surfaces. Low-level handlers that must
implement only selected methods can embed `rpc.UnimplementedServerRequestHandler`.
For methods whose historical Go spelling used non-canonical initialisms, such
as `McpServerElicitationRequest`, implement the exported canonical capability
interface (for example, `rpc.MCPServerElicitationRequestHandler`); dispatch
prefers that method over the legacy spelling. The client may invoke different
server-request callbacks concurrently, so callbacks that share state must be
concurrency-safe.

`AutoApproveHandler` accepts command, file-change, and permission requests and
must only be used in a trusted environment. Its default logs are redacted. If
sensitive command and path logging is explicitly required, construct
`NewUnsafeLoggingAutoApproveHandler(logger)` and protect those logs accordingly.
Use `codex.AutoApproveCallbacks(codex.AutoApproveHandler{Logger: logger})` to
attach a safe auto-approver through the preferred callback API. The unsafe
handler can be passed to the same adapter when explicitly required.

## Codex CLI compatibility

When the SDK spawns `codex`, it requires the CLI major/minor version to match
the generated protocol version; patch differences are accepted. A mismatch,
missing binary, or unparseable version returns `*codex.CodexCompatibilityError`
before the process starts. Set `CompatibilityPolicy: codex.Warn` only after
validating compatibility (and provide a logger to see the warning), or
`codex.Ignore` to skip the probe. Custom transports are never probed.

### Updating to protocol 0.153.4

This SDK targets Codex CLI 0.153.x. Low-level RPC now includes paginated turn
and item history, thread revert, plugin reconciliation, and new realtime,
project, queue, and authentication-recovery notifications.

`AccountUsageRead(ctx)` remains available. Use
`AccountUsageReadWithParams(ctx, &protocol.GetAccountTokenUsageParams{ThreadID: &id})`
for thread-specific usage. `ThreadForkOptions.ExcludeTurns` is supported again.
Low-level resume and turn-start parameters expose the new history and per-turn
options. `Thread.ProjectID` is nullable and always serializes as `projectId`, as
required by the new protocol. Approval requests that omit `kind` decode as
`command`, matching the upstream default for older servers.

Section appearance updates distinguish omission, clearing, and replacement:
leave `ThreadSectionUpdateParams.Appearance` nil to preserve it, assign a pointer
to a nil `*protocol.ThreadSectionAppearance` to clear it, or point to an appearance
value to replace it. Existing section response names remain available.

`CapabilityRootLocation`, `ClientNotification`, `DynamicToolNamespaceTool`,
`LocalShellAction`, and `ReasoningItemReasoningSummary` now use the same raw-preserving union wrappers as
the other discriminated types. Replace direct map assignments with the
corresponding `protocol.New<Type>(value)` constructor, and use `RawJSON()` to
inspect the payload. Known variants validate shared required fields; unknown
future variants continue to round-trip unchanged.

## Structured Output

Provide a JSON Schema to constrain the final assistant message.

```go
prompt := "Summarize repo status"
schema := codex.MustJSON(map[string]any{
    "type": "object",
    "properties": map[string]any{
        "summary": map[string]any{"type": "string"},
        "status": map[string]any{"type": "string", "enum": []string{"ok", "action_required"}},
    },
    "required": []string{"summary", "status"},
    "additionalProperties": false,
})

_, err := thread.RunInputs(ctx, []codex.Input{codex.TextInput(prompt)}, &codex.TurnOptions{
    OutputSchema: schema,
})
```

## JSON-typed options

Fields like `ApprovalPolicy`, `SandboxPolicy`, `Effort`, `Summary`, and `OutputSchema` accept any JSON-marshalable value. If you already have raw JSON, pass a `json.RawMessage` (or `codex.MustJSON(...)`) to avoid double encoding.

For common values, prefer typed constants from this package:

- `codex.ApprovalPolicyNever`, `codex.ApprovalPolicyOnFailure`, `codex.ApprovalPolicyOnRequest`, `codex.ApprovalPolicyUntrusted`
- `codex.SandboxModeReadOnly`, `codex.SandboxModeWorkspaceWrite`, `codex.SandboxModeDangerFullAccess`
- `codex.ReasoningEffortNone`, `codex.ReasoningEffortMinimal`, `codex.ReasoningEffortLow`, `codex.ReasoningEffortMedium`, `codex.ReasoningEffortHigh`, `codex.ReasoningEffortXHigh`

## Inputs and overload errors

Use helpers to build structured inputs:

```go
inputs := []codex.Input{
    codex.TextInput("Inspect this file"),
    codex.MentionInput("AGENTS.md"),
}
```

Overload classification uses ordinary Go errors:

```go
if codex.IsOverloaded(err) && operationIsIdempotent {
    // Retry according to your caller policy.
}
```

`IsOverloaded` classifies structured overload failures; it does not prove that
repeating an operation is safe. Base retries on idempotency or an
application-level idempotency key. `IsRetryable` remains as a deprecated alias
for source compatibility.

## Low-level RPC

Use the RPC client directly for full control.

```go
rpcClient := client.Client()
models, err := rpcClient.ModelList(ctx, protocol.ModelListParams{})
```

Use checked constructors such as `rpc.NewClientChecked`,
`rpc.NewConnTransportChecked`, and `rpc.NewReplayTransportChecked` when a
dependency is dynamic. Their legacy counterparts panic immediately on invalid
programmer input instead of failing later in a background goroutine.

All outbound writes are serialized through a bounded queue. `Call` and
`Notify` return when their contexts end; with a legacy `rpc.Transport`, the
underlying write may remain blocked until `Close`. Implement
`rpc.ContextTransport` when writes themselves must observe cancellation.
App-server request handling defaults to four workers and a queue of 64; tune
`rpc.ClientOptions.ServerRequestWorkers` and
`ServerRequestQueueCapacity` when using the low-level client.

Inbound JSON-RPC messages are limited to 8 MiB by default. Set
`rpc.ClientOptions.MaxMessageBytes` to apply a smaller client-side acceptance
limit. Built-in line transports retain the 8 MiB allocation ceiling even if
the client option is larger. Custom transports return an already allocated
string, so they must enforce their own read/allocation limit; the client option
then rejects oversized returned values. Oversized messages close the client
with an error matching `rpc.ErrMessageTooLarge`.
