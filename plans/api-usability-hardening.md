# Make the exported Go API safe, typed, and predictable

This ExecPlan is a living document. The sections `Progress`, `Surprises & Discoveries`, `Decision Log`, and `Outcomes & Retrospective` must be kept up to date as work proceeds.

This plan follows `PLANS.md` at the repository root. It is self-contained and describes the API-usability remediation requested after the review in `GC.md`.

## Purpose / Big Picture

After this work, Go callers can use mention inputs, thread lifecycle operations, turn results, request handlers, and low-level RPC transports without silent wire-data loss, opaque `interface{}` responses, background panics, unbounded request-handler concurrency, or misleading retry guidance. The protocol package will expose canonical typed declarations while retaining deprecated aliases long enough for callers to migrate. The behavior is observable through exact JSON transcript tests, external-package compile tests, concurrency tests under the race detector, deterministic generation from Codex `rust-v0.144.6`, and the repository's complete local quality gate.

The work intentionally does not refresh to a newer upstream Codex release. All generated changes must use Codex version `0.144.6`, commit `5d1fbf26c43abc65a203928b2e31561cb039e06d`, so API cleanup is not mixed with protocol drift.

## Tracker Mapping

Workflow: `WORKFLOW.md`, with the explicit local-tracking exception recorded below. By the user's standing decision, this repository is the only tracker: no Linear project or issues will be created. This ExecPlan is the parent tracker, and the following exact `GC.md` headings are its child task identifiers:

- `[High] MentionInput emits the wrong JSON field name`
- `[High] Core thread lifecycle APIs return opaque interfaces`
- `[High] Hand-written protocol types silently truncate current payloads`
- `[High] Request contexts do not bound outbound RPC writes`
- `[High] Typed-nil approval handlers can panic the process`
- `[High] IsRetryable claims safety that error classification cannot prove`
- `[Medium] Generated broad interfaces make protocol updates source-breaking`
- `[Medium] TurnHandle exposes competing consumers for one iterator`
- `[Medium] Failed turns discard their accumulated result and structured details`
- `[Medium] FinalResponse can report plan text as the final answer`
- `[Medium] MCP acronym normalization creates duplicate opaque APIs`
- `[Medium] Fixed-choice protocol values are exposed as any`
- `[Medium] Server-request dispatch lacks a reliable error and concurrency contract`
- `[Medium] Nil transport constructors fail asynchronously`
- `[Medium] ReplayTransport can hang on invalid input and lose JSON number precision`
- `[Low] Generated implementation names pollute the public protocol namespace`
- `[Low] Compatibility shims do not use Go deprecation and error conventions`

`GC.md` remains an untracked review artifact and must never be staged. This plan restates the full implementation scope so future work does not depend on the untracked file.

## Progress

- [x] (2026-07-22 09:04Z) Rechecked the repository instructions, current `main` at `31fb6e9`, existing plans, generated version metadata, untracked `GC.md`, and the 17 finding locations.
- [x] (2026-07-22 09:04Z) Completed the required independent planner pass and validated its proposed milestones against the current repository.
- [ ] Create and switch to local branch `codex/api-usability-hardening`; keep `GC.md` untracked.
- [ ] Complete Milestone 1: patch-safe wire, panic, retry, constructor, replay, deprecation, and error-sentinel fixes with tests.
- [ ] Complete Milestone 2: context-bounded outbound writes and bounded, correctly classified server-request dispatch with concurrency tests.
- [ ] Complete Milestone 3: canonical complete protocol generation and upgrade-safe handler/client interfaces from pinned `rust-v0.144.6`.
- [ ] Complete Milestone 4: dependable high-level lifecycle and turn behavior with typed results, errors, and single-consumer enforcement.
- [ ] Complete Milestone 5: migration documentation, examples, external-package compile fixtures, full local quality gate, and final independent reviews.
- [ ] Update this plan's outcomes, ensure all 17 tracker headings are addressed, and leave only intended committed changes plus untracked `GC.md`.

## Surprises & Discoveries

- Observation: The previous `plans/repository-hardening.md` covers a separate 10-finding runtime, security, CI, and release program and should not be reopened for this API program.
  Evidence: `main` already contains hardening commit `baa1d0a` and release tag `v0.145.0`, while the current `GC.md` contains a new 17-finding exported-API review.
- Observation: Several apparently independent findings share one generator naming and schema-sanitization path.
  Evidence: `internal/codegen/main.go` normalizes initialisms for generated declarations but fallback resolution uses raw schema titles; it also removes nested `oneOf`/`anyOf`, producing both the duplicate `Mcp...` names and constrained `interface{}` fields.
- Observation: Some public signatures can become typed without changing the method declaration text.
  Evidence: `ListThreads`, `ReadThread`, `ForkThread`, and `UnarchiveThread` refer to named `protocol.Thread*Response` declarations. Replacing those declarations from empty interfaces with structs preserves the method names and source shape for callers that did not depend on the old dynamic values.

## Decision Log

- Decision: Track this program only in `plans/api-usability-hardening.md`, using the 17 exact `GC.md` headings as child identifiers.
  Rationale: The user explicitly requested local-repository tracking only. A self-contained ExecPlan satisfies the repository workflow without creating external state.
  Date/Author: 2026-07-22 / Codex
- Decision: Keep `GC.md` untracked and commit the ExecPlan plus implementation files.
  Rationale: `GC.md` is the inspection artifact; the plan is the durable implementation record and restates every finding.
  Date/Author: 2026-07-22 / Codex
- Decision: Generate only from `rust-v0.144.6` during this work.
  Rationale: A protocol refresh would mix upstream schema changes with local API corrections and make compatibility review unreliable.
  Date/Author: 2026-07-22 / Codex
- Decision: Deliver compatible fixes in place and use additive checked APIs plus deprecated aliases where possible; allow intentional source breaks only for the typed protocol migration.
  Rationale: Wrong wire keys, typed-nil panics, misleading retry classification, and deadlocking constructors can be fixed without forcing callers to rewrite correct code. Replacing weak public fields and unstable broad interfaces requires a documented migration and at least a new minor version in this pre-v1 module.
  Date/Author: 2026-07-22 / Codex
- Decision: Preserve existing lifecycle method names and signatures while making their named protocol response types concrete.
  Rationale: Callers gain typed fields without an unnecessary duplicate high-level method family.
  Date/Author: 2026-07-22 / Codex
- Decision: Add preferred narrow/capability request-handler APIs, retain broad legacy interfaces as deprecated compatibility surfaces, and provide an embeddable default handler.
  Rationale: Producer-owned generated interfaces otherwise break every implementer when upstream adds a method.
  Date/Author: 2026-07-22 / Codex
- Decision: Make `TurnHandle` single-consumer by mode and return partial `TurnResult` values with typed terminal errors.
  Rationale: Competing drains are inherently ambiguous, while partial results contain useful IDs, items, usage, timestamps, and error details that should not be discarded.
  Date/Author: 2026-07-22 / Codex

## Outcomes & Retrospective

Implementation has not started. At completion, this section will record the user-visible API changes, compatibility shims, generator changes, test evidence, review findings, commits, and any deliberately deferred cleanup.

## Context and Orientation

The repository root is the high-level `codex` package. `codex.New` constructs a client, `Codex` exposes account and thread lifecycle helpers, `Thread` starts turns, and `TurnHandle` owns a live notification stream. The `rpc` directory contains the low-level line-oriented JSON-RPC client and transports. The `protocol` directory contains checked-in generated wire types. `internal/codegen/main.go` exports schemas from a local `openai/codex` checkout and writes the generated protocol and RPC files.

The current protocol metadata in `protocol/metadata_gen.go` is version `0.144.6` at commit `5d1fbf26c43abc65a203928b2e31561cb039e06d`. Regeneration must use `/Users/pme/src/openai/codex` with `CODEX_REPO_REF=rust-v0.144.6`. Generated files are not edited directly unless the generator explicitly treats the file as a compatibility shim; durable generator changes belong in `internal/codegen/main.go` and its tests.

An empty-interface response such as `type ThreadListResponse interface{}` accepts any value and makes a generated RPC method return `*interface{}`, which is not a useful Go result. A context-bounded write means the caller of `Call` or `Notify` returns when its context ends even if a legacy transport remains blocked; it does not mean the remote app-server canceled a request already received. A capability interface contains only one coherent group of server callbacks, unlike the current generated `ServerRequestHandler` that requires every callback.

Compatibility work follows a two-level rule. Existing correct patch-level behavior should continue to compile. Where the current API is weak by design—empty interfaces, fixed-choice `any`, duplicate acronym names, and producer-owned broad interfaces—the plan may introduce source changes in a documented minor release while preserving deprecated aliases and forwarding methods when Go permits it.

## Plan of Work

### Milestone 1: Correct patch-safe API defects

Begin with failing unit tests. In `input.go`, correct `Input.TextElements` to `json:"text_elements,omitempty"` and assert the exact `turn/start` and `turn/steer` JSON for a non-ASCII mention. In `logging.go` and `codex.go`, normalize typed-nil handlers before the RPC client sees them and copy known pointer handler values before attaching inherited loggers.

In `errors.go`, deprecate `IsRetryable`, remove arbitrary matching against an outer error string, and keep overload detection limited to `ErrOverloaded` plus structured `rpc.ResponseError` evidence. Update `README.md` and `doc.go` so classification is never described as proof that an operation is idempotent.

In `rpc`, add stable closure sentinels and canonical error-detail naming. Add checked, error-returning constructors for client, connection, record, and replay dependencies. Existing constructors remain available but reject invalid programmer input synchronously. Validate replay directions and use `json.Decoder.UseNumber` when normalizing JSON. Add canonical `Deprecated:` comments to unsupported option fields, misspelled transport aliases, and superseded names.

This milestone addresses the mention, typed-nil, retry, nil-constructor, replay, and compatibility-convention findings. Run focused root and RPC tests plus the race detector before committing it.

### Milestone 2: Bound RPC writes and server-request work

Add an optional context-aware transport interface in `rpc/transport.go`, while retaining the existing `Transport`. Refactor `rpc.Client` so all outbound requests, notifications, and replies pass through one serialized writer. Each caller waits on writer completion, client shutdown, or its own context; a canceled caller returns promptly even when a legacy transport write remains blocked until `Close`.

Replace one-goroutine-per-server-request dispatch with a fixed worker pool and bounded queue configured by documented finite defaults in `ClientOptions`. The read loop must remain responsive when the queue is full. Change generated dispatch support so unknown methods, malformed parameters, and application-handler failures are distinct typed conditions mapped to JSON-RPC `-32601`, `-32602`, and a documented application code. Recover handler panics at the worker boundary. A reply write failure is terminal and must wake pending calls and iterators with its cause.

Write deterministic blocked-write, serialized-write, queue-saturation, panic, error-code, and reply-failure tests. Repeat the concurrency-focused cases under `-race` and `-count=100` before committing.

### Milestone 3: Generate canonical complete protocol APIs

This is the semver-sensitive milestone. Add one schema-title-to-Go-name function in `internal/codegen/main.go` and use it for declarations, fallback lookup, aliases, RPC parameter/response names, notification names, and method names. Cover `MCP`, `OAuth`, `ID`, `URL`, `JSON`, `RPC`, `HTTP`, `SSE`, and `UUID` with generator fixtures. Generate canonical MCP declarations and methods, retaining the old `Mcp...` spellings as deprecated aliases or forwarding methods.

Teach the generator to preserve ordinary object responses even when nested members include unions. Generate complete canonical thread lifecycle responses and current thread, turn, item, and error notification payloads. Remove the corresponding partial declarations from `protocol/manual_types.go`; keep manual request shapes only where the high-level SDK still needs a concrete builder and the source schema remains unsupported. Preserve genuinely unresolved nested values as `json.RawMessage` behind a reviewed allowlist rather than degrading entire responses to `interface{}`.

Convert constrained weak fields such as thread-list sort direction/key and approval decisions to named enums, nullable references, or raw-preserving mixed-union wrappers. Shrink the allowlists and fail generation when a new weak type is not explicitly reviewed. Generate canonical names directly in public fields. Retain existing `Sanitized...`, `...JSON`, and compatibility spellings as deprecated aliases for a migration window. Add `protocol/doc.go` explaining package stability and raw-union escape hatches.

Stop growing the legacy broad `ClientRequests` and `ServerRequestHandler` interfaces. Keep them as deprecated hand-written compatibility surfaces, generate capability-specific interfaces, and provide an embeddable `UnimplementedServerRequestHandler` whose default methods report unsupported operations. Concrete `*rpc.Client` methods remain generated.

Run generator unit tests, generate twice from exactly `rust-v0.144.6`, and require no second diff. Then run protocol/RPC tests and external-package compile fixtures before committing.

### Milestone 4: Make lifecycle and turn behavior dependable

Update `thread_lifecycle.go` to use the new concrete protocol responses directly and remove JSON re-marshaling helpers such as `threadIDFromAny`. Change thread-list sort options to the generated enum types. Add a preferred root request-handler option backed by stable capability interfaces while retaining the old approval-handler path as deprecated; reject conflicting configuration.

Add a consumption-mode state machine to `TurnHandle`. `Run`, manual `Next`, and `Stream` claim exactly one mode, and mixed or concurrent drains return a stable typed error immediately. A stream view obtained from a handle must consume through the handle so turn IDs and other state remain current for `Steer` and `Interrupt`.

Introduce `ErrTurnFailed` and an exported `TurnError` that carries the accumulated `*TurnResult`, the terminal notification method, complete protocol error details, retry metadata, and raw payload. Terminal failures return both the non-nil partial result and the typed error. Decode completed items by discriminator and update `FinalResponse` only from the final assistant-message variant; retain every raw item in `Items`.

Add tests for all mode-mixing orders, concurrent consumers, stream-observed turn IDs, partial failed results, `errors.Is`/`errors.As`, plan/commentary/final-answer item sequences, and concrete list/read/fork/unarchive responses. Replace E2E JSON substring checks with typed assertions where practical.

### Milestone 5: Document migration and prove the complete result

Update `README.md`, `doc.go`, `APP.md`, examples, and exported GoDoc. Document the mention wire correction, retry/idempotency distinction, checked constructors, partial turn failures, single-consumer rule, request-handler migration, canonical MCP names, fixed-choice enums, concrete lifecycle responses, and deprecated alias window.

Add an external-package compile fixture that demonstrates keyed options, a narrow consumer-defined client interface, an embedded default server handler, canonical MCP names, enum options, and `errors.Is`/`errors.As`. Run formatting, installer fixtures if touched, vet, unit tests, race tests, Staticcheck v0.7.0, govulncheck v1.3.0, deterministic generation, and `git diff --check`. Run credentialed E2E only through the repository's trusted environment and pinned verified CLI.

After implementation and local validation, run the applicable independent review agents required by `AGENTS.md`: `review-security` for handler panics/concurrency and transport cancellation, `review-qa` for regression coverage, `review-architecture` for generated/hand-written boundaries, and `review-ux-specialist` for the exported API and migration docs. Address actionable findings before final staging or commits that are intended for handoff.

## Concrete Steps

Run commands from `/Users/pme/src/pmenglund/codex-sdk-go` unless another directory is stated.

Create the local implementation branch after preserving the untracked review artifact:

    git switch -c codex/api-usability-hardening
    git status --short --branch

Use fast focused loops during the milestones:

    go test . ./rpc
    go test -race . ./rpc
    go test ./internal/codegen ./protocol ./rpc

For pinned regeneration, run twice and inspect the second result:

    CODEX_REPO_ROOT=/Users/pme/src/openai/codex CODEX_REPO_REF=rust-v0.144.6 go generate ./...
    git status --short
    CODEX_REPO_ROOT=/Users/pme/src/openai/codex CODEX_REPO_REF=rust-v0.144.6 go generate ./...
    git diff --exit-code -- protocol rpc

Run concurrency tests repeatedly using the exact test names introduced by Milestones 2 and 4:

    go test -race -count=100 ./rpc -run 'BlockedWrite|ServerRequest|ReplyFailure'
    go test -race -count=100 . -run 'TurnHandleConsumption|TurnFailure'

Run the full local quality gate before completion:

    test -z "$(gofmt -l $(git ls-files '*.go'))"
    go vet ./...
    go test ./...
    go test -race ./...
    staticcheck ./...
    govulncheck ./...
    git diff --check
    git status --short --branch

`GC.md` must remain listed as `?? GC.md` and must not appear in `git diff --cached --name-only`.

## Validation and Acceptance

Every finding is accepted only when its pre-fix regression test fails for the expected reason and passes after the change, unless the finding is documentation-only. The following behavior must be demonstrated:

- Mention inputs encode `text_elements`, never `textElements`, for start and steer requests.
- Typed-nil handlers and nil transports fail synchronously without background panics; shared handler pointers are not mutated.
- `IsRetryable` is deprecated and cannot classify arbitrary outer error text as proof of safe retry.
- Replay rejects invalid directions and distinguishes adjacent integers above `2^53`.
- `Call` and `Notify` return on context deadline during a blocked write; all outbound writes remain serialized.
- Server request concurrency and queue memory are bounded; unknown method, invalid params, handler error, and handler panic produce the documented responses without stopping unrelated traffic.
- Thread list/read/fork/unarchive responses are concrete typed values, and current thread/turn/error fields survive decode and re-encode.
- Canonical initialisms and constrained fields appear in generated signatures; compatibility spellings remain deprecated aliases or forwarders.
- Adding a simulated new server method does not break a handler that embeds the generated default implementation.
- Mixed `TurnHandle` consumption modes fail immediately; stream consumption still updates the turn ID.
- Failed turns return a non-nil partial result and a typed error; plan or commentary text never becomes `FinalResponse`.
- All package examples and external-package compile fixtures use the preferred API.
- The pinned generator is deterministic and the complete local quality gate passes.

Credential-backed `go test -tags=e2e ./test/e2e` is accepted only when run by the trusted workflow with the pinned Codex CLI and login contract. Its absence locally does not invalidate unit acceptance, but it must be reported explicitly at handoff.

## Idempotence and Recovery

Unit tests, formatting, vet, Staticcheck, govulncheck, and pinned generation are safe to rerun. The generator uses a temporary detached worktree when `CODEX_REPO_REF` is set and must not change the user's upstream Codex checkout. If generation creates a second-pass diff, stop and fix nondeterminism before proceeding.

Each milestone should produce a small logical local commit only after its focused tests pass. Do not stage `GC.md`. If a milestone fails, keep its changes uncommitted, update `Progress` and `Surprises & Discoveries`, and continue from the smallest failing package. Do not reset, checkout, or overwrite unrelated user changes.

The principal rollback boundary is Milestone 3. Before publishing its source-breaking typed migration, retain the prior API through deprecated aliases and compile fixtures. If the object renderer cannot represent the verified `rust-v0.144.6` schemas without broad new weak types, pause that milestone and record the exact unsupported schemas rather than adding permissive fallback entries.

Published tags are immutable. A bad release is corrected by pinning consumers to the prior version, using a `retract` directive where appropriate, fixing forward, and publishing a higher version; never move an existing tag.

## Artifacts and Notes

Starting commit: `31fb6e9` on `main`. Starting worktree: only `?? GC.md`. Generated protocol: Codex `0.144.6`, commit `5d1fbf26c43abc65a203928b2e31561cb039e06d`.

Expected commit grouping is one local commit per completed milestone. Commit messages use imperative mood and reference this local plan, for example `Harden patch-safe API behavior`, `Bound RPC request handling`, `Generate canonical protocol APIs`, `Make turn results dependable`, and `Document API migration`.

No separate `record.md` is needed initially. If generated API inventories or before/after compatibility evidence become too large for this plan, create `plans/api-usability-hardening-record.md`, define its entry format here, and keep only concise conclusions in this ExecPlan.

## Interfaces and Dependencies

Milestone 2 adds an optional low-level interface:

    type ContextTransport interface {
        Transport
        WriteLineContext(context.Context, string) error
    }

Checked constructors return errors while legacy constructors remain compatibility shims. Exact names may be adjusted once call-site ergonomics are tested, but the end state must include checked construction for `Client`, `ConnTransport`, `RecordTransport`, and `ReplayTransport`.

Milestone 3 produces concrete named protocol response and notification structs, canonical initialism names, capability-specific server handler interfaces, and an embeddable `UnimplementedServerRequestHandler`. The broad legacy interfaces remain deprecated compatibility surfaces and are no longer the generator's preferred extension point.

Milestone 4 exports stable turn failure inspection:

    var ErrTurnFailed error

    type TurnError struct {
        Result       *TurnResult
        Method       string
        Detail       *protocol.TurnError
        WillRetry    bool
        Raw          json.RawMessage
    }

`TurnError` implements `error`, supports `errors.Is(err, ErrTurnFailed)`, and is discoverable through `errors.As`. The exact representation may use private fields plus accessors if that better preserves invariants, but callers must be able to obtain the partial result and complete wire error details.

No new third-party runtime dependency is expected. Use the standard library and existing generator dependency. Staticcheck and govulncheck versions remain those pinned by `WORKFLOW.md`.

## Revision Notes

- 2026-07-22: Created this ExecPlan from the 17-finding exported-API review, the required independent planner pass, and the user's local-only tracking decision. The plan separates patch-safe fixes from pinned semver-sensitive generation work so implementation can remain testable and recoverable.
