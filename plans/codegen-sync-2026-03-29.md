# Sync SDK Handwritten Code with 2026-03-29 Generated Protocol

This ExecPlan is a living document. The sections `Progress`, `Surprises & Discoveries`, `Decision Log`, and `Outcomes & Retrospective` must be kept up to date as work proceeds.

`PLANS.md` is checked into this repository and this document must be maintained in accordance with `/Users/pme/src/pmenglund/codex-sdk-go/PLANS.md`.

## Purpose / Big Picture

After this change, `go generate ./...` followed by `go test ./...` will succeed against the latest exported Codex app-server protocol. Users of this SDK will be able to consume the regenerated protocol and RPC stubs without local compile failures in approval handling, thread resume options, or RPC server request handler implementations.

## Tracker Mapping

Workflow: `WORKFLOW.md`. Tracker gap: no Linear issue or epic identifier was provided in the user request and none is discoverable from the current branch or repository metadata. This ExecPlan file is the tracked artifact for this maintenance sync. Scope covered by this plan: rerun code generation, update handwritten SDK code to the new generated protocol/RPC surface, and validate with repository tests.

## Progress

- [x] (2026-03-29 20:35Z) Read `AGENTS.md`, `APP.md`, `LANGUAGE.md`, `WORKFLOW.md`, and `PLANS.md`; confirmed the tree was clean before generation.
- [x] (2026-03-29 20:37Z) Ran `go generate ./...` and captured the generated diff in `protocol/` and `rpc/`.
- [x] (2026-03-29 20:39Z) Fixed `internal/codegen/main_test.go` so `TestCodexRepoRootDefault` is hermetic when `CODEX_REPO_ROOT` is set in the shell.
- [x] (2026-03-30 00:40Z) Mapped the handwritten breakages to upstream protocol changes: approval/tool schemas now need manual shims, `thread/resume` dropped `history` and `path`, `thread/start` dropped `experimentalRawEvents`, and `turn/start` dropped `collaborationMode`.
- [x] (2026-03-30 00:52Z) Added manual protocol shims in `protocol/manual_types.go`, updated `internal/codegen/main.go` to reserve those names, and regenerated `protocol/` plus `rpc/`.
- [x] (2026-03-30 00:57Z) Updated the root package and RPC tests to use typed approval responses, added the new handler methods required by `rpc.ServerRequestHandler`, and made removed options fail explicitly instead of silently dropping them.
- [x] (2026-03-30 01:00Z) Ran `gofmt`, `go test ./...`, and `go vet ./...` successfully.
- [x] (2026-03-30 01:00Z) Recorded the final outcome and rationale in this plan.

## Surprises & Discoveries

- Observation: The initial concurrent `go test ./...` run passed the root package because it raced with `go generate ./...` and compiled against the pre-generation tree.
  Evidence: A second `go test ./...` after generation failed with undefined symbols in `approvals.go`, `thread_options.go`, and RPC test handlers.

- Observation: `TestCodexRepoRootDefault` was not hermetic because it inherited the caller's `CODEX_REPO_ROOT`.
  Evidence: The failing test expected a temp sibling checkout but resolved `/Users/pme/src/openai/codex` from the environment.

- Observation: Several approval-related protocol types still exist upstream as concrete Rust structs, but our Go generator now falls back to `interface{}` for them.
  Evidence: `CommandExecutionRequestApprovalParams` and `PermissionsRequestApprovalParams` are present in `codex-rs/app-server-protocol/src/protocol/v2.rs`, while the generated Go output initially emitted them via `protocol/fallback_gen.go`.

- Observation: The latest v2 schema removed `history` and `path` from `ThreadResumeParams`, plus `experimentalRawEvents` from `ThreadStartParams` and `collaborationMode` from `TurnStartParams`.
  Evidence: `protocol/types_gen.go` after regeneration no longer contains those fields, and the source schema at `codex-rs/app-server-protocol/schema/json/v2/ThreadResumeParams.json` only requires `threadId`.

## Decision Log

- Decision: Create an ExecPlan for this task once the generated sync turned into a cross-package handwritten compatibility update.
  Rationale: The work now spans multiple components with uncertain protocol deltas, which meets the repository threshold for non-trivial planned work.
  Date/Author: 2026-03-29 / Codex

- Decision: Fix the test isolation issue immediately before chasing protocol deltas.
  Rationale: The environment-sensitive failure is independent of the generated protocol changes and would otherwise obscure validation.
  Date/Author: 2026-03-29 / Codex

- Decision: Add manual protocol types for the approval/tool request shapes that the generator can no longer emit instead of weakening the SDK call sites to raw `interface{}`.
  Rationale: The upstream protocol remains typed; preserving typed Go structs in `protocol/manual_types.go` keeps the SDK API usable after regeneration and contains the workaround to a small, explicit compatibility layer.
  Date/Author: 2026-03-30 / Codex

- Decision: Keep removed public option fields (`History`, `Path`, `ExperimentalRawEvents`, `CollaborationMode`) source-visible but make them return explicit errors when used.
  Rationale: The current app-server protocol no longer accepts those fields. Failing explicitly is safer than silently ignoring user input, while retaining the fields avoids a harder compile-time break for existing callers.
  Date/Author: 2026-03-30 / Codex

## Outcomes & Retrospective

`go generate ./...` now succeeds, the regenerated `protocol/` and `rpc/` trees are checked in, and the handwritten SDK code compiles and tests cleanly against the updated protocol. The compatibility work fell into three buckets: preserving typed approval/tool request handling via manual protocol shims, adapting the root package to removed protocol fields, and extending handler/test implementations for the new server request method set.

The biggest remaining risk is maintenance burden if the upstream protocol keeps evolving faster than `go-jsonschema` can represent the raw schemas. The manual shim layer keeps the SDK working today, but future protocol syncs should continue to inspect upstream schema diffs first and either replace or expand those manual types intentionally.

## Context and Orientation

This repository has two kinds of code relevant to this task. Generated files live under `protocol/` and `rpc/` and are overwritten by `go generate ./...`, which runs `go run ./internal/codegen`. Handwritten SDK logic lives at the repository root (for example `approvals.go`, `thread_options.go`, and `option_values.go`) plus tests in `rpc/*.go`. The generated protocol types define the wire format exported by the Codex app-server. When those generated names or field layouts change, the handwritten SDK code must adapt.

The failing areas currently fall into four buckets. First, `approvals.go` constructs the high-level `ApprovalRequest` from `protocol.CommandExecutionRequestApprovalParams`, and that generated type no longer exposes the fields expected by the old code. Second, `thread_options.go` references `protocol.ThreadResumeParamsHistoryElem`, which appears to have been renamed or removed in the new schema. Third, `option_values.go` defines approval policy constants and now conflicts with the generated `ApprovalPolicy` representation. Fourth, the generated RPC `ServerRequestHandler` interface gained at least one new request method, which breaks the fake handlers used in `rpc/client_internal_test.go`, `rpc/client_test.go`, and `rpc/generated_test.go`.

## Plan of Work

Start by reading the generated protocol and RPC output around each failing symbol to understand the new source-of-truth type names, enum shapes, and request methods. Update the root package code in the smallest possible way so public SDK behavior remains stable where the protocol still carries equivalent data. For approval requests, map from the new generated fields into the existing `ApprovalRequest` structure, adjusting helper logic and tests only where the protocol genuinely changed. For thread resume history, replace the removed generated type reference with the new generated history item type or adapt the conversion logic if the wire shape now uses a different container. For approval policies, switch from constants to variables or another idiomatic representation if the generated type is not a defined string type anymore. For RPC handlers, extend the test doubles with the new generated method and keep behavior minimal.

After the handwritten fixes are in place, run `gofmt` on each changed handwritten file, then rerun `go test ./...` and `go vet ./...`. If new failures appear in examples or docs, patch only the files required to restore compatibility with the generated protocol. This work ended up requiring changes in `protocol/manual_types.go`, `internal/codegen/main.go`, `approvals.go`, `thread_options.go`, `turn.go`, and the RPC test doubles.

## Concrete Steps

Run these commands from `/Users/pme/src/pmenglund/codex-sdk-go`.

1. Regenerate and capture the baseline failures.

       go generate ./...
       go test ./...

2. Inspect the generated protocol and handwritten call sites.

       rg -n "CommandExecutionRequestApprovalParams|ThreadResumeParams|ApprovalPolicy|AccountChatgptAuthTokensRefresh" protocol rpc *.go
       go test ./... 2>&1 | sed -n '1,220p'

3. Apply targeted handwritten fixes, then format and validate.

       gofmt -w approvals.go option_values.go thread_options.go internal/codegen/main_test.go rpc/*.go examples/**/*.go
       go test ./...
       go vet ./...

Expected final result: `go test ./...` prints only passing packages or `[no test files]`, and `go vet ./...` exits successfully.

Actual final validation:

       go test ./...
       ok  	github.com/pmenglund/codex-sdk-go	1.839s
       ok  	github.com/pmenglund/codex-sdk-go/examples/approvals	1.277s
       ok  	github.com/pmenglund/codex-sdk-go/examples/low_level_rpc	0.569s
       ok  	github.com/pmenglund/codex-sdk-go/examples/quickstart	0.336s
       ok  	github.com/pmenglund/codex-sdk-go/examples/streaming	1.524s
       ok  	github.com/pmenglund/codex-sdk-go/examples/structured_output	0.805s
       ok  	github.com/pmenglund/codex-sdk-go/internal/codegen	(cached)
       ok  	github.com/pmenglund/codex-sdk-go/rpc	(cached)

       go vet ./...
       <no output>

## Validation and Acceptance

Acceptance is behavioral and repository-wide: after running `go generate ./...`, the SDK must build and its full test suite must pass without local environment assumptions. In particular, `internal/codegen` tests must pass even when `CODEX_REPO_ROOT` is set, the root package must compile with the regenerated `protocol` package, and `rpc` test handlers must satisfy the generated `ServerRequestHandler` interface.

## Idempotence and Recovery

`go generate ./...` is safe to rerun; it rewrites checked-in generated files from the current Codex checkout. If a handwritten compatibility fix is wrong, rerun `go test ./...` to get the next failing symbol and adjust incrementally. The only environment-sensitive path in this task is `CODEX_REPO_ROOT`, and the test suite should explicitly set or clear it when behavior depends on it.

## Artifacts and Notes

Initial post-generation failures:

    ./thread_options.go:67:21: undefined: protocol.ThreadResumeParamsHistoryElem
    ./option_values.go:9:26: invalid constant type ApprovalPolicy
    ./approvals.go:22:23: params.ThreadID undefined
    rpc/client_internal_test.go:25:27: ... missing method AccountChatgptAuthTokensRefresh

Final compatibility approach:

    - `protocol/manual_types.go` now provides typed shims for approval/tool request shapes that the generator fell back to `interface{}` for.
    - `internal/codegen/main.go` reserves those manual names so regeneration does not overwrite them with fallback placeholders.
    - `thread_options.go` and `turn.go` now reject removed protocol fields with explicit errors.

## Interfaces and Dependencies

No new dependencies are expected. The relevant interfaces and types are the generated `protocol.CommandExecutionRequestApprovalParams`, the generated thread resume parameter/history types in `protocol/types_gen.go`, the generated approval policy type used by the root `ApprovalPolicy` alias, and the generated `rpc.ServerRequestHandler` interface in `rpc/server_requests_gen.go`. The root package should continue exposing `ApprovalRequest`, thread options helpers, and approval policy values in an idiomatic Go API even if the generated wire types shift.

Change log:

- 2026-03-29: Created the ExecPlan after the task expanded from pure regeneration into a cross-package handwritten compatibility update.
- 2026-03-30: Updated the plan with the completed compatibility strategy, validation results, and the protocol fields that were removed upstream.
