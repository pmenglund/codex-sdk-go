# Repository Review Findings

## Summary
- Total findings by severity: Critical 0, High 3, Medium 2, Low 0
- Highest-value issues first:
- Canceled RPC calls are still written to the transport before `Client.Call` returns the context error.
- Generated protocol typing silently collapses many public request, response, and notification types to `interface{}`.
- Server request handlers execute inline on the JSON-RPC reader goroutine with `context.Background()`.
- High-level turn/resume builders rely on comments and server-side errors for invalid zero values.
- `StdioTransport.Close` kills the app-server immediately and discards shutdown errors.

## Findings

### [High] Canceled RPC calls are still sent
- Category: Correctness risk / context propagation
- File / Symbol: `rpc/client.go:77` `Client.Call`; `rpc/client_requests_gen.go:303` `Client.FsWriteFile`
- Why it matters: Callers reasonably expect a canceled context to prevent a side-effecting RPC from being sent. Today an already-canceled context can still create threads, write files, start turns, or run other generated RPC methods, while the caller only sees `context.Canceled`.
- Evidence: `Client.Call` registers the pending request, builds the payload, and calls `c.send(payload)` at `rpc/client.go:82-98`; it does not check `ctx.Done()` until `rpc/client.go:100-106`. Generated side-effecting methods such as `FsWriteFile` call through `c.Call` at `rpc/client_requests_gen.go:303-306`. `TestCallContextCancel` uses an already-canceled context but provides a replay transcript with an expected outgoing `ping` write at `rpc/client_test.go:267-282`, so the current test allows the write before the context error.
- Recommended fix: Check `ctx.Err()` before allocating/registering the request and again immediately before writing. Add a test transport that asserts no write occurs when the context is canceled before `Call`. If writes can block on stdio, consider extending the transport boundary with a context-aware write or a client-level writer goroutine that can honor cancellation.
- Confidence: High

### [High] Generated protocol typing collapses to interface{}
- Category: Unsafe boundary handling / code generation
- File / Symbol: `protocol/fallback_gen.go`; `internal/codegen/main.go` `writeFallbackTypes`, `stripSubschemas`; `rpc/client_requests_gen.go` `ClientRequests`
- Why it matters: The repository advertises a typed low-level RPC API, but many public protocol types are only `interface{}`. JSON responses unmarshal into maps and slices instead of stable structs, so schema drift is not caught at compile time, callers lose discoverable fields, and invalid wire payloads cross the SDK boundary as untyped data.
- Evidence: `protocol/fallback_gen.go:8-127` defines many public request/response/notification types as `interface{}`, including `InitializeResponse`, `ModelListResponse`, `ThreadListResponse`, and `TurnStartResponse`. The generator emits those fallbacks explicitly in `writeFallbackTypes` at `internal/codegen/main.go:294-304`. The sanitizer removes every `oneOf` and `anyOf` recursively at `internal/codegen/main.go:801-814`, which is a broad source of lost schema detail. Generated RPC methods then present these fallback types as typed APIs, for example `ModelList(ctx, protocol.ModelListParams) (*protocol.ModelListResponse, error)` in `rpc/client_requests_gen.go:44` and `var result protocol.ModelListResponse` in the generated method body pattern at `internal/codegen/main.go:578-586`.
- Recommended fix: Make fallback generation fail for public RPC request/response/notification types unless an explicit manual type or alias exists. Prefer adding union-aware codegen for the schema shapes used by Codex, or alias unsanitized public names to the existing `Sanitized*JSON` structs when they are structurally useful. Keep `interface{}` only for truly open-ended JSON fields and document each exception.
- Confidence: High

### [High] Server request handlers block the JSON-RPC reader and ignore client cancellation
- Category: Concurrency / context propagation
- File / Symbol: `rpc/client.go` `readLoop`, `handleServerRequest`; `rpc/server_requests_gen.go` `ServerRequestHandler`
- Why it matters: Approval, tool-call, and elicitation handlers are user-supplied extension points and may perform I/O or call back into the client. Running them synchronously on the only reader goroutine blocks all response and notification processing; a handler that waits on another RPC response from the same client can deadlock. Because handlers receive `context.Background()`, closing the client or canceling the operation does not cancel the handler work.
- Evidence: `readLoop` dispatches request messages directly with `c.handleServerRequest(msg.request)` at `rpc/client.go:191-199`. `handleServerRequest` calls `dispatchServerRequest(context.Background(), handler, req)` at `rpc/client.go:248-255`. The generated handler interface is context-shaped (`ServerRequestHandler` methods take `ctx context.Context`) at `rpc/server_requests_gen.go:15-25`, but the runtime never supplies a lifecycle or call context.
- Recommended fix: Give `Client` a lifecycle context that is canceled from `finish`/`Close`, pass it into `dispatchServerRequest`, and run server request handling outside the reader loop while keeping response writes serialized through the existing transport lock. Add a regression test with a handler that blocks until client close and another that performs a nested `Call`, so the reader cannot regress back to inline blocking behavior.
- Confidence: High

### [Medium] High-level builders allow invalid zero-value requests through
- Category: Hidden contracts / API validation
- File / Symbol: `input.go` `Input`; `turn.go` `buildTurnParams`; `thread_options.go` `ThreadResumeOptions.toParams`
- Why it matters: The public high-level API exposes invariants as comments and string fields, then sends invalid values to the RPC layer. This makes invalid zero values fail late at the app-server boundary, produces less actionable errors, and makes the SDK easier to misuse as protocol options evolve.
- Evidence: `Input.Type` is a plain exported string with the comment “must be one of the InputType* constants” at `input.go:17-24`. `buildTurnParams` appends each `Input` unchanged at `turn.go:233-240`; there is no validation that `Type` is known or that required fields such as text, URL, path, or skill name are present. `ThreadResumeOptions.toParams` only copies `ThreadID` when it is non-empty at `thread_options.go:94-98`, but the generated protocol marks `ThreadID string` as the required `threadId` field at `protocol/types_gen.go:3243-3244`, and `ResumeThread` proceeds to call `thread/resume` at `codex.go:118-124`.
- Recommended fix: Add SDK-side validation in `buildTurnParams` and `ThreadResumeOptions.toParams`: reject unknown input types, reject missing required fields per input variant, and reject empty `ThreadID` now that history/path resume are explicitly unsupported. Longer term, narrow `Input` behind constructors or variant-specific types with custom marshaling so invalid combinations are harder to construct.
- Confidence: High

### [Medium] Stdio shutdown always kills the app-server and hides errors
- Category: Resource management / shutdown path
- File / Symbol: `rpc/transport.go` `StdioTransport.Close`; `codex.go` `New`, `Close`
- Why it matters: `Codex.New` intentionally detaches the spawned app-server from the constructor context and documents that process lifetime is managed by `Close`. That makes `Close` the important cleanup boundary, but the current implementation immediately kills the process and suppresses close, kill, and wait errors. This can interrupt graceful app-server cleanup and prevents callers from diagnosing shutdown failures.
- Evidence: `codex.New` spawns with `context.WithoutCancel(ctx)` at `codex.go:45-47`, and `Codex.Close` delegates to the RPC client at `codex.go:84-89`. `StdioTransport.Close` closes stdin, calls `Process.Kill`, calls `Wait`, ignores all errors, and always returns nil at `rpc/transport.go:85-93`.
- Recommended fix: Make shutdown graceful first: close stdin or send an explicit shutdown notification if the protocol supports one, wait with a short timeout, and only then kill as a fallback. Return meaningful close/wait errors, using `errors.Join` where multiple cleanup steps fail. Keep forced kill behavior available for stuck processes, but do not make it the unconditional first path.
- Confidence: Medium
