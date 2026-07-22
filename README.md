# Codex Go SDK

Embed the Codex app-server in Go workflows.

This SDK speaks JSON-RPC to the `codex app-server` process. By default it spawns the CLI and communicates over stdio.

## Requirements

- Go 1.25.12 or newer
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
        break
    }
    fmt.Printf("%s\n", note.Method)
    if note.Method == "turn/completed" {
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

This low-level source change requires SDK version `v0.145.0`
or newer; `protocol.GeneratedCodexVersion` continues to describe the upstream
wire schema, not the SDK release number. The same release adds
`Options.CompatibilityPolicy` and changes `StartLogin` to accept JSON-marshalable
login parameters; code using unkeyed `Options` literals, interface method sets,
or assigned `StartLogin` method values must be updated.

## Approvals

Configure approval handling by supplying a handler when constructing the client.

```go
logger := slog.New(slog.NewTextHandler(os.Stderr, &slog.HandlerOptions{Level: slog.LevelInfo}))
client, err := codex.New(ctx, codex.Options{
    Logger:          logger,
    ApprovalHandler: codex.RejectingApprovalHandler{},
})
```

For custom approval logic, implement `rpc.ServerRequestHandler` (from `rpc`).

`AutoApproveHandler` accepts command, file-change, and permission requests and
must only be used in a trusted environment. Its default logs are redacted. If
sensitive command and path logging is explicitly required, construct
`NewUnsafeLoggingAutoApproveHandler(logger)` and protect those logs accordingly.

## Codex CLI compatibility

When the SDK spawns `codex`, it requires the CLI major/minor version to match
the generated protocol version; patch differences are accepted. A mismatch,
missing binary, or unparseable version returns `*codex.CodexCompatibilityError`
before the process starts. Set `CompatibilityPolicy: codex.Warn` only after
validating compatibility (and provide a logger to see the warning), or
`codex.Ignore` to skip the probe. Custom transports are never probed.

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
