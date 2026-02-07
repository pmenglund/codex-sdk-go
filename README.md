# Codex Go SDK

Embed the Codex app-server in Go workflows.

This SDK speaks JSON-RPC to the `codex app-server` process. By default it spawns the CLI and communicates over stdio.

## Requirements

- Go 1.25+
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

## Approvals

Configure approval handling by supplying a handler when constructing the client.

```go
logger := slog.New(slog.NewTextHandler(os.Stderr, &slog.HandlerOptions{Level: slog.LevelInfo}))
client, err := codex.New(ctx, codex.Options{
    Logger:          logger,
    ApprovalHandler: codex.AutoApproveHandler{Logger: logger},
})
```

For custom approval logic, implement `rpc.ServerRequestHandler` (from `rpc`).

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

## Low-level RPC

Use the RPC client directly for full control.

```go
rpcClient := client.Client()
models, err := rpcClient.ModelList(ctx, protocol.ModelListParams{})
```

## Code generation

Regenerate protocol types and RPC stubs:

```bash
go generate ./...
```

This runs:

- `cargo run -p codex-app-server-protocol --bin export`
- `go-jsonschema` (via `internal/codegen`)

The generator needs a checkout of `openai/codex` to export schemas.
It resolves that checkout in this order:

- `$CODEX_REPO_ROOT` (if set)
- `../codex` (default)

Generated files include a header line with the exact codex commit hash used.

Generated files are checked in under `protocol` and `rpc`.
