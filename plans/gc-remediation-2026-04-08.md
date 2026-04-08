# Address GC Repository Review Findings

This ExecPlan is a living document. The sections `Progress`, `Surprises & Discoveries`, `Decision Log`, and `Outcomes & Retrospective` must be kept up to date as work proceeds.

`PLANS.md` is checked into this repository and this document must be maintained in accordance with `/Users/pme/src/pmenglund/codex-sdk-go/PLANS.md`.

## Purpose / Big Picture

After this change, the SDK will handle the concrete risks documented in `GC.md`: canceled calls will not be sent, protocol generation will preserve stronger public types where feasible, server request handlers will not block the JSON-RPC reader, high-level builders will reject invalid requests earlier, and stdio shutdown will report cleanup errors and prefer graceful process exit before killing. A user can see the work by reading `GC.md`, inspecting the targeted code paths, and running `go test ./...` plus `go vet ./...`.

## Tracker Mapping

Workflow: `WORKFLOW.md`. Tracker gap: no Linear issue or epic identifier was provided in the user request, and the current branch is `main` with no issue identifier. Tracked Markdown artifacts for this maintenance task: `GC.md` and this ExecPlan file. Scope covered by this plan: address the five findings in `GC.md` one by one and commit after each finding is resolved.

## Progress

- [x] (2026-04-08 02:43Z) Read `AGENTS.md`, `APP.md`, `LANGUAGE.md`, `WORKFLOW.md`, `PLANS.md`, and `GC.md`; confirmed the only existing worktree change was the untracked `GC.md` report from the prior review step.
- [x] (2026-04-08 02:45Z) Resolved GC finding 1 by checking `ctx.Err()` in `rpc.Client.Call` before request registration and before send, and by changing the canceled-call test to assert no write occurs.
- [x] (2026-04-08 02:47Z) Resolved GC finding 2 by teaching codegen to alias fallback titles to generated `Sanitized*JSON` structs when available; `protocol/fallback_gen.go` now has 65 aliases and 55 remaining true `interface{}` fallbacks.
- [x] (2026-04-08 02:48Z) Resolved GC finding 3 by adding a client lifecycle context, canceling it during `finish`, and dispatching server request handling outside the JSON-RPC reader goroutine with a regression test for blocked handlers.
- [x] (2026-04-08 02:49Z) Resolved GC finding 4 by validating input variants in `buildTurnParams` and rejecting empty `ThreadResumeOptions.ThreadID`.
- [ ] Resolve GC finding 5: make stdio shutdown graceful first and return cleanup errors.
- [ ] Run final repository validation and record outcomes.

## Surprises & Discoveries

- Observation: `GC.md` was untracked at the start of remediation because it was created by the immediately preceding review task and had not been committed yet.
  Evidence: `git status --short` printed `?? GC.md` before this plan was created.

- Observation: Strengthening `protocol.ModelListResponse` exposed that the low-level RPC example transcript used an outdated `models` response shape and tested map fallback behavior.
  Evidence: `go test ./internal/codegen ./protocol ./rpc ./examples/low_level_rpc` initially failed with `cannot convert map[string]any{...} ... to type protocol.ModelListResponse`; updating the transcript to `data` with a typed `protocol.Model` made the focused test pass.

## Decision Log

- Decision: Add a focused ExecPlan before code changes.
  Rationale: The requested remediation spans RPC context handling, code generation, server request concurrency, high-level API validation, and process shutdown, which meets the repository threshold for non-trivial planned work.
  Date/Author: 2026-04-08 / Codex

- Decision: Commit the review report and remediation plan as setup before implementation commits.
  Rationale: `GC.md` was an existing untracked artifact from the review step, and keeping setup separate lets each later commit map cleanly to one resolved finding.
  Date/Author: 2026-04-08 / Codex

## Outcomes & Retrospective

Not yet completed. This section will be updated after the five findings have been addressed and final validation has passed.

## Context and Orientation

This repository is a Go SDK for the Codex app-server. The root package, in files such as `codex.go`, `thread.go`, `turn.go`, `thread_options.go`, `input.go`, and `approvals.go`, exposes the high-level SDK API. The `rpc/` package owns JSON-RPC transport and generated RPC stubs. The `protocol/` package contains generated wire-level types plus a small manual compatibility file. The `internal/codegen/` package generates `protocol/` and `rpc/` files from schemas exported by a local Codex checkout.

The review report `GC.md` contains five findings. A JSON-RPC call is a request sent over a line-delimited JSON transport to the app-server. A server request handler is user-supplied Go code called when the app-server asks the SDK to approve a command, answer a tool request, or handle another server-initiated request. A generated fallback type is a public protocol type emitted as `interface{}` when the generator cannot emit a stronger Go representation.

## Plan of Work

First, commit `GC.md` and this plan as setup. Then address each finding in the order listed in `GC.md`, running the smallest useful test set after each change and committing immediately after validation. For finding 1, update `rpc.Client.Call` so cancellation is checked before request registration and again before sending, then replace the current canceled-call test with one that asserts no write occurs. For finding 2, inspect the generated `Sanitized*` types and update codegen/manual protocol shims so public RPC request, response, and notification names do not fall back to unconstrained `interface{}` when a useful sanitized type exists; add or update codegen tests. For finding 3, add a client lifecycle context, cancel it from `finish`, and dispatch server request handling outside `readLoop` while preserving serialized writes. For finding 4, validate `Input` values and required `ThreadID` before sending RPCs. For finding 5, change `StdioTransport.Close` to close stdin, wait briefly, then kill only if needed, returning meaningful cleanup errors.

After all findings are resolved, run `gofmt` on touched Go files, then run `go test ./...` and `go vet ./...`. Update this plan’s progress, discoveries, decision log, and outcomes as each commit lands.

## Concrete Steps

Run these commands from `/Users/pme/src/pmenglund/codex-sdk-go`.

1. Commit setup artifacts.

       git status --short
       git add GC.md plans/gc-remediation-2026-04-08.md
       git commit -m "Document GC remediation plan"

2. For each finding, edit the relevant files, run focused tests, then run at least the affected package tests before committing.

       go test ./rpc
       go test ./...
       go vet ./...
       git add <changed files>
       git commit -m "<short imperative finding-specific message>"

3. After the final finding, run the full validation suite again.

       go test ./...
       go vet ./...

Expected final result: all tests pass, `go vet ./...` exits without output, and `git log --oneline` shows setup plus one commit per resolved finding.

## Validation and Acceptance

Acceptance is repository-wide and behavior-focused. A canceled `Client.Call` must not write to the transport. Generated public protocol names should be stronger than `interface{}` when the repository already has a useful sanitized or manual type. A blocking or nested server request handler must not prevent the reader loop from processing later responses. Invalid high-level inputs and empty resume thread IDs must be rejected before network or process I/O. `StdioTransport.Close` must attempt graceful process exit before forced kill and return cleanup errors when they matter.

Run `go test ./...` and expect all packages to pass. Run `go vet ./...` and expect no output.

## Idempotence and Recovery

All planned code changes are local Go edits and can be retried safely. If a finding-specific change breaks unrelated tests, inspect the failing package before continuing and update this plan with the discovery. If a commit is created for one finding and a later finding needs to touch the same file, make an additive follow-up commit rather than rewriting the earlier history.

## Artifacts and Notes

Initial status before this plan:

    ?? GC.md

Initial validation from the review step:

    go test ./...
    ok  	github.com/pmenglund/codex-sdk-go	0.951s
    ok  	github.com/pmenglund/codex-sdk-go/rpc	(cached)

    go vet ./...
    <no output>

## Interfaces and Dependencies

No new external dependencies are expected. The work should use Go standard library packages such as `context`, `errors`, `os/exec`, and `time`. The key local interfaces and functions are `rpc.Transport`, `rpc.Client.Call`, `rpc.Client.readLoop`, `rpc.ServerRequestHandler`, `codex.Input`, `buildTurnParams`, `ThreadResumeOptions.toParams`, and `rpc.StdioTransport.Close`.

Change log:

- 2026-04-08: Created this ExecPlan to track the five finding remediation task requested from `GC.md`.
