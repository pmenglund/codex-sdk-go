# Harden runtime safety, protocol types, CI, and releases

This ExecPlan is a living document. The sections `Progress`, `Surprises & Discoveries`, `Decision Log`, and `Outcomes & Retrospective` must be kept up to date as work proceeds.

This document follows `PLANS.md` in this repository. It is self-contained for a contributor who has only this working tree. The local `GC.md` review artifact supplied the initial findings, but implementation must not depend on that file being committed or staged.

## Purpose / Big Picture

After this work, canceling a synchronous SDK turn will stop the corresponding Codex work, a slow notification consumer will not exhaust process memory or stall unrelated RPC traffic, and callers will receive safer approval and version-compatibility behavior. Generated discriminated unions use raw-preserving wrappers with typed kind constants, known-variant required-key checks, and fail-closed schema inventory review instead of silently degrading new unions into `interface{}`. Pull requests run useful secretless checks, trusted end-to-end tests use a pinned verified Codex CLI, and an immutable release tag is publishable only after protected quality and end-to-end gates succeed.

The program is complete when all ten findings summarized below have tests and observable acceptance evidence, the final implementation is merged into `main`, and the next release is published through the protected workflow rather than a local direct push. Per the recorded tracking exception, this plan replaces Linear for this program.

## Tracker Mapping

Workflow: `WORKFLOW.md`.

Tracking exception: By explicit user decision on 2026-07-21, this ExecPlan is the sole tracker for this hardening program. No Linear epic or child issues will be created, and milestone updates that would normally be posted to Linear will instead be recorded in `Progress`, `Decision Log`, and `Outcomes & Retrospective` below.

Local work items:

- interrupt active turns when synchronous runs are canceled.
- pin and verify the Codex CLI and remove broad CI secret exposure.
- preserve `oneOf` and `anyOf` as typed generated unions.
- split secretless quality checks from trusted end-to-end tests.
- publish releases only after protected checks.
- bound notification queues without blocking the RPC reader.
- raise the Go security baseline and add vulnerability scanning.
- redact approval logs and lead documentation with a safe handler.
- make runtime protocol compatibility explicit and actionable.
- enforce formatting, vet, race, and static-analysis gates.

The user approved plan-file-only tracking on 2026-07-21. Each implementation commit and pull request must reference this ExecPlan.

## Progress

- [x] (2026-07-21) Re-read `AGENTS.md`, `APP.md`, `LANGUAGE.md`, `WORKFLOW.md`, `PLANS.md`, the current code, CI, updater, tests, and all ten `GC.md` findings.
- [x] (2026-07-21) Confirmed `main` is at `8725d1a` and the only pre-existing worktree change is untracked `GC.md`.
- [x] (2026-07-21) Obtained an independent planning-only review from the required planner agent and validated its proposed files and assumptions against the current tree.
- [x] (2026-07-21) Created this proposed ExecPlan without changing implementation files.
- [x] (2026-07-21) Recorded the user's explicit approval to use this ExecPlan as the sole tracker instead of Linear.
- [x] (2026-07-21) Accepted the recommended policy assumptions in `Open Decisions Before Implementation`; the user asked to implement the plan without revisions.
- [x] (2026-07-21) Completed Milestone 1: pinned Go 1.25.12 and 1.26.5, split secretless quality CI, added a checksum-verified Codex installer with fixture tests, formatted generated Go, and proved deterministic generation.
- [x] (2026-07-21) Completed Milestone 2 implementation: synchronous runs delegate to `TurnHandle.Run`, cancellation best-effort interrupts with a bounded cleanup context, replay tests preserve exact context errors, and a trusted live test is present. Execution of the credentialed live test remains an external validation step.
- [x] (2026-07-21) Completed Milestone 3: notification buffers are hard capacities, overflow is typed and isolated, RPC progress continues, and race tests pass.
- [x] (2026-07-21) Completed Milestone 4: approval logs are redacted, safe rejection leads the example, sensitive logging is explicit, and spawned-client compatibility is enforced with typed errors and tested policies.
- [x] (2026-07-21) Completed Milestone 5: 31 discriminated unions are concrete generated raw-preserving wrappers, future variants round-trip, and every remaining weak type is covered by an explicit reviewed allowlist that rejects new entries by default.
- [x] (2026-07-21) Completed Milestone 6 repository changes: trusted e2e and release workflows reuse pinned CLI metadata and protected environments, and the local updater can no longer stage, commit, tag, or push.
- [x] (2026-07-21) Ran security, QA, architecture, UX, and SRE reviews; addressed all release-blocking findings. Follow-up security and QA reviews found no remaining blockers, and architecture accepted the documented raw-wrapper boundary.
- [x] (2026-07-21) Completed the post-review matrix on Go 1.25.12 and 1.26.5: vet, unit, race, Staticcheck v0.7.0, and govulncheck v1.3.0 pass; tagged e2e, lifecycle repetition, actionlint, installer/release/manifest fixtures, formatting, regeneration, and diff checks pass. Credentialed live e2e remains protected-environment evidence.
- [x] (2026-07-21) Staged the reviewed repository changes while leaving `GC.md` unstaged.
- [ ] Merge the reviewed implementation, publish the next release through the protected workflow, and record the final external evidence in `Outcomes & Retrospective`.

## Open Decisions Before Implementation

The following are recommended assumptions. The user may revise them before implementation; otherwise the implementer must record explicit acceptance in the Decision Log before starting the affected milestone.

1. `Thread.Run` and `Thread.RunInputs` interrupt the active turn when their context is canceled or expires. `RunStreamed` and manual `TurnHandle.Next` plus `Close` remain the explicit detach-and-continue APIs.
2. `SubscribeNotifications(buffer)` treats `buffer` as a hard pending-notification limit, with the existing default of 64 when the caller supplies a non-positive value. Overflow closes only that subscription with a typed error and never blocks the JSON-RPC reader.
3. Default spawned-client compatibility requires the runtime Codex CLI to match the generated protocol's major and minor version. Same-major/minor patch differences are allowed. Callers may explicitly choose `Warn` or `Ignore`; custom transports are not probed.
4. Generated discriminated unions use wrapper types with custom JSON methods, typed kind constants, known-variant required-key validation, `RawJSON` access, and raw preservation for a well-formed future discriminator. Full field type and constraint validation remains a server boundary. Genuinely opaque, non-discriminated schemas require an explicit reviewed allowlist, and the full `oneOf`/`anyOf` inventory is fail-closed behind a reviewed digest.
5. `AutoApproveHandler` retains its current exported struct shape and decisions for source compatibility, but default logs omit command bodies, paths, and working directories. A new explicitly named opt-in type or constructor enables sensitive logging. Documentation presents a rejecting or callback-policy handler first.
6. The trusted end-to-end job runs on protected `main` pushes and manual dispatch through a protected `e2e` GitHub Environment. It is required by the release workflow; whether it is also required for internal pull requests must be chosen when branch protection is configured.
7. A protected `release` GitHub Environment requires human approval. The release workflow may write tags with minimal `contents: write` permission but never pushes `main`.
8. The exact versioned Codex archive digest is committed in a repository-owned checksum file and reviewed independently of the download step. Implementation must confirm which upstream checksum or GitHub asset digest covers the exact archive; publication remains blocked if no trustworthy digest can be established.
9. The generated `protocol` package's union migration is a source-breaking low-level API correction. The next SDK tag must be a new minor release, at least `v0.145.0`, and SDK release versioning must be explicitly separated from `protocol.GeneratedCodexVersion` if upstream Codex has not supplied the same minor version.
10. `GC.md` remains an unstaged review artifact unless the user explicitly asks to commit it. The ExecPlan embeds the relevant findings so future work does not depend on `GC.md` being present.

## Surprises & Discoveries

- Observation: The CI workflow installs and executes a real Codex CLI and requires `OPENAI_API_KEY`, but the only command is `go test ./...`, which excludes `test/e2e` because that package is guarded by the `e2e` build tag.
  Evidence: `.github/workflows/tests.yml` runs no `-tags=e2e` command, while `test/e2e/codex_test.go` starts with `//go:build e2e`.
- Observation: The CI secret name and the e2e helper contract do not match.
  Evidence: CI exports `OPENAI_API_KEY`; `test.RequireLoginParams` reads `CODEX_E2E_LOGIN_PARAMS_JSON` and skips credentialed tests when it is absent.
- Observation: The buffer accepted by `SubscribeNotifications` is not a memory bound.
  Evidence: `notificationSubscription.run` appends every event to an unlimited private slice after the public `out` channel fills.
- Observation: The generator does not merely fall back on a few difficult schemas; it recursively deletes every `oneOf` and `anyOf` before generation.
  Evidence: `stripSubschemas` deletes both keys from every map. Current generated files contain 99 named `interface{}` fallback types and 233 total `interface{}` occurrences.
- Observation: Fixing `AutoApproveHandler` by adding a configuration field could break callers that use unkeyed composite literals.
  Evidence: The exported struct currently has only `Logger`; preserve the struct shape and put new configuration on a separate type or constructor.
- Observation: The updater's atomic push prevents a partial branch/tag push but is not a release gate.
  Evidence: `--push-release` runs only generation and `go test ./...`, stages with `git add .`, and pushes `main` plus the tag before hosted checks or review.
- Observation: The existing SDK tag is derived from the upstream Codex version, but typed union corrections can change the SDK API without changing the upstream protocol version.
  Evidence: `protocol.GeneratedCodexVersion` is currently `0.144.6`, and the updater derives `v0.144.6` directly from `rust-v0.144.6`. The versioning rules must be decoupled before a source-breaking generator correction is published.
- Observation: The locally installed Go 1.26.4 toolchain still contains GO-2026-5856, while the declared minimum Go 1.25.12 and CI's Go 1.26.5 are patched.
  Evidence: local govulncheck under Go 1.26.4 reported the called `crypto/tls` issue; identical scans under Go 1.25.12 and Go 1.26.5 reported no vulnerabilities.
- Observation: Generalized discriminated-union generation covered 31 current schemas without hand-writing payload structs. One server-request union, `McpServerElicitationRequestParams`, required an explicit opaque exception because existing callers select its variant through request context and rely on nil/zero behavior.
  Evidence: exact-tag generation at `rust-v0.144.6` and the full test suite passed with that single discriminated-union exception plus explicit allowlists for other non-discriminated weak schemas.
- Observation: Staticcheck exposed a nil-dereference path after `testing.TB.Fatalf` because the interface contract does not tell the analyzer that `Fatalf` terminates execution.
  Evidence: adding an explicit return after the nil result failure made Staticcheck v0.7.0 clean on both supported toolchains.
- Observation: The exact tagged e2e command initially failed because Codex plugin-clone activity raced `testing.TempDir` cleanup.
  Evidence: retrying cleanup of an explicit temporary `CODEX_HOME` made the lifecycle test pass three consecutive runs and the full tagged package pass.
- Observation: A cleanup context alone did not bound cancellation because `Transport.WriteLine` has no context.
  Evidence: interruption now runs behind a bounded wait with a blocking-transport regression test; custom transports are documented to unblock writes from `Close`.
- Observation: Independent review found that checking out moving `main` after gates could tag untested code and that credentialed manual e2e could run branch code.
  Evidence: release publication now pins `github.sha`, verifies `origin/main`, serializes dispatches, and trusted e2e has a secretless main-ref prerequisite.

## Decision Log

- Decision: Use one new ExecPlan, `plans/repository-hardening.md`, rather than extending the completed protocol-update plan.
  Rationale: The completed plan documents a narrow updater workflow. This work crosses runtime behavior, concurrency, API generation, security, CI, documentation, and releases.
  Date/Author: 2026-07-21 / Codex
- Decision: Deliver the work as incremental milestones and pull requests instead of one large change.
  Rationale: Immediate security and runtime fixes should not wait for the higher-risk union generator, and every merged state must remain releasable.
  Date/Author: 2026-07-21 / Codex
- Decision: Keep `GC.md` outside the plan's required artifact set.
  Rationale: It is an untracked review artifact. This plan restates the evidence needed for implementation and must remain usable if `GC.md` is removed.
  Date/Author: 2026-07-21 / Codex
- Decision: Require a prototype promotion gate for typed unions.
  Rationale: `go-jsonschema` v0.23.1 does not model the current union schemas adequately, and a broad rewrite without representative fixtures would create a large public API risk.
  Date/Author: 2026-07-21 / Codex
- Decision: Preserve exact context-error identity during best-effort interrupt cleanup.
  Rationale: Existing tests and callers may compare directly with `context.Canceled` or `context.DeadlineExceeded`. Interrupt failure is secondary cleanup evidence and belongs in structured logs rather than the returned error.
  Date/Author: 2026-07-21 / Codex
- Decision: Track this hardening program only in the local ExecPlan and do not create Linear issues.
  Rationale: The user explicitly selected local-repository tracking. Milestone status and evidence will be recorded in this file in place of Linear comments.
  Date/Author: 2026-07-21 / User and Codex
- Decision: Accept all ten recommended policy assumptions as the implementation baseline.
  Rationale: After reviewing the implementation-ready plan, the user instructed Codex to implement it and did not request changes to the listed policies.
  Date/Author: 2026-07-21 / User and Codex
- Decision: Pin the trusted Codex CLI to published stable version 0.144.4 while the generated protocol remains 0.144.6.
  Rationale: 0.144.4 is the latest published stable GitHub release with authoritative asset digests available during implementation, and the compatibility policy explicitly accepts patch differences within major/minor 0.144.
  Date/Author: 2026-07-21 / Codex
- Decision: Generate raw-preserving discriminated-union wrappers with typed kind constants instead of hand-maintaining every variant payload struct.
  Rationale: This representation covers 31 current unions deterministically, keeps unknown future variants lossless, validates required keys for known variants, avoids coupling the SDK generator to hundreds of nested Rust schema types, and makes any remaining `interface{}` an explicit review decision. A reviewed inventory digest makes every upstream `oneOf`/`anyOf` change fail closed; `CODEX_PRINT_UNION_INVENTORY=1` prints per-schema review evidence.
  Date/Author: 2026-07-21 / Codex
- Decision: Separate SDK release numbering from `protocol.GeneratedCodexVersion` in the protected release workflow and documentation.
  Rationale: The union migration changes the Go API without changing the upstream wire version. The release workflow refuses tags older than v0.145.0 while keeping generated wire metadata at 0.144.6.
  Date/Author: 2026-07-21 / Codex

## Outcomes & Retrospective

All six repository milestones are implemented on `codex/repository-hardening`. Local validation is green under Go 1.25.12 and Go 1.26.5 for unit tests, race tests, vet, Staticcheck v0.7.0, and govulncheck v1.3.0. Installer, release-tag, and complete-content manifest fixtures, actionlint v1.7.7, the exact tagged e2e package, repeated lifecycle e2e, formatting, `git diff --check`, exact-tag regeneration, and deterministic generated-content checks also pass.

The remaining work is external: run the credentialed trusted e2e workflow, configure and verify GitHub branch/tag/environment protections, merge to `main`, and publish at least v0.145.0 through the protected release workflow. No local release or direct push is an accepted substitute.

## Context and Orientation

The repository is a Go SDK that starts `codex app-server` and communicates over line-delimited JSON-RPC. The root `codex` package exposes clients, threads, turn handles, streaming, lifecycle methods, and approval helpers. `rpc/client.go` owns request correlation, server-request dispatch, and notification subscriptions. `rpc/transport.go` owns spawned-process and connection transports. `protocol/` contains checked-in generated wire types. `internal/codegen/main.go` exports upstream Codex schemas and writes the generated `protocol` and `rpc` files. `test/e2e/` contains the real CLI/auth/turn tests behind the `e2e` build tag.

Before implementation, `Thread.RunInputs` obtained only a `TurnStream` from `RunStreamed`; when its context ended, it closed the local iterator but could not call `TurnHandle.Interrupt`.

Before implementation, each `rpc.Client` notification subscription had a buffered public channel, an unbuffered inbox, and an unbounded goroutine-owned slice.

Code generation still sanitizes union keywords for the underlying `go-jsonschema` library, but it first validates the exact full union inventory, requires the four core wrappers, generates 31 discriminated wrappers, and rejects new weak output outside reviewed allowlists. The raw-wrapper boundary and incomplete field-type validation are documented as low-level API behavior rather than described as full schema validation.

The current GitHub Actions workflow downloads the mutable `releases/latest` Codex asset, makes `OPENAI_API_KEY` available to the entire job, then runs only the untagged unit suite. The local updater can generate, commit, tag, and push `main` in one command. These paths must be redesigned together so the same pinned CLI metadata and validation commands are used by trusted e2e and release publication.

## Plan of Work

### Milestone 1: Establish secure, reproducible quality foundations

This milestone fixes the toolchain vulnerability and makes ordinary pull-request validation meaningful without secrets. It does not yet change runtime behavior.

Update `go.mod` from Go 1.25.5 to at least Go 1.25.12, choosing the latest patched 1.25 release available when implementation begins. Keep the supported language baseline in Go 1.25; test both the exact minimum and the then-current stable Go release in CI when practical.

Split `.github/workflows/tests.yml` into a required secretless quality job and a trusted e2e job or reusable workflow. The quality job must not install Codex and must not receive any API credential. Pin `actions/checkout`, `actions/setup-go`, and any new action to reviewed commit SHAs. Run a no-output formatting check, `go vet ./...`, `go test ./...`, `go test -race ./...`, a Staticcheck version known to support the selected Go toolchain, and `govulncheck ./...`. Pin analyzer versions rather than downloading mutable latest versions.

Add a reusable installer such as `.github/scripts/install-codex-cli.sh` and repository-owned version/checksum metadata under `.github/codex/`. The installer must accept an exact version, choose the platform asset, verify its pinned digest before extraction, and fail closed if metadata is absent. It must never receive API credentials. Add shell tests or fixture-driven checks for supported architectures, absent metadata, and checksum mismatch.

Change `internal/codegen/main.go` so every generated `.go` file passes through `go/format.Source` before `os.WriteFile`. Keep non-Go writes separate. Tests in `internal/codegen/main_test.go` must prove invalid Go fails generation and all rendered outputs are formatted. Add a deterministic generation check using the exact `protocol.GeneratedCodexVersion` upstream tag: two consecutive runs must leave no second diff.

Update `APP.md` and `WORKFLOW.md` with the exact local and hosted quality gates. The milestone is accepted when a fork-style secretless CI run can pass without Codex, formatting produces no filenames, the analyzers pass on supported toolchains, the known called standard-library vulnerability is gone, corrupt CLI archives fail before execution, and generation is formatted and idempotent.

### Milestone 2: Interrupt synchronous turns on cancellation

Refactor `Thread.RunInputs` in `thread.go` to retain the `TurnHandle` returned by `StartTurn` and delegate aggregation to `TurnHandle.Run`. Remove the duplicated completion loop only after replay tests prove results, failures, notifications, timestamps, token usage, and final response remain unchanged.

In `turn_handle.go`, when `TurnHandle.Run` receives exactly `context.Canceled` or `context.DeadlineExceeded` from `Next`, best-effort call `turn/interrupt` with a short cleanup context derived from `context.WithoutCancel(ctx)`. Use a named timeout constant. Return the original context error unchanged even when interrupt fails; log the cleanup failure with thread and turn IDs. If no turn ID is known, log that interruption was impossible and return the original error. Keep `RunStreamed` and manual handle iteration as the documented way to detach without implicit interruption. Remove the unused `closeErr` field unless it receives a tested responsibility.

Extend `turn_test.go` and `turn_handle_test.go` with replay transcripts for successful interrupt, failed interrupt, missing turn ID, repeated close, and normal completion. Preserve direct equality assertions for context errors. Add a trusted e2e case that records transport output, waits until a turn ID is known, cancels the synchronous run, and observes `turn/interrupt`.

The milestone is accepted when cancellation always attempts interruption when an ID exists, the original context error is returned unchanged, no interrupt is sent after a normal completion, and all previous aggregation behavior remains green.

### Milestone 3: Bound notification delivery without blocking RPC

Replace the unbounded queue in `rpc/client.go` with a fixed-capacity subscription. Export a sentinel-compatible typed error, for example `NotificationOverflowError`, containing the configured capacity. `errors.As` must identify the detailed type and `errors.Is` must work with an exported overflow sentinel if both are provided.

Make `notificationSubscription.publish` non-blocking. When a subscription is full, atomically store its terminal overflow error, close only that subscription, and remove it from `Client.subs`. Extend `NotificationIterator` to distinguish a subscription error from global client shutdown. Preserve idempotent concurrent `Close` and prevent send-on-closed-channel races. Do not block `readLoop`: a response, server request, and healthy subscriber must continue after another subscriber overflows.

Replace `TestNotificationDeliveryDoesNotDropWhenBufferFills` with tests for the new bounded contract. Add cases proving retained messages never exceed capacity, `Next` returns the typed overflow error, another subscriber continues, a pending `Client.Call` completes, and concurrent publish/close is race-free. Document the hard-cap behavior in `rpc/doc.go` and the README low-level section.

The milestone is accepted when an abandoned iterator has bounded memory, overflow is observable and isolated, and `go test -race ./...` passes stress cases without blocking the reader.

### Milestone 4: Make approval and compatibility UX safe by default

Preserve the exported shape and approval decisions of `AutoApproveHandler`. Change its default structured logs in `approvals.go` so they include request kind and stable IDs but omit command bodies, cwd, grant roots, file paths, and permission payloads. Do not add a field to `AutoApproveHandler`, because an extra field breaks unkeyed composite literals. Add a separate explicitly named sensitive-logging wrapper or constructor and test that it is the only path that emits those details.

Add a new rejecting or callback-policy handler whose zero value rejects supported command, file-change, permission, and legacy approvals with protocol-valid decisions. Do not present shell-string prefix matching as an authorization boundary. Update `README.md` and `examples/approvals/` so the safe handler appears first and `AutoApproveHandler` carries a prominent trusted-environment warning.

Add `CompatibilityPolicy` to `Options` in `options.go`, with `RequireMajorMinor`, `Warn`, and `Ignore`. Use a zero value that implements the approved default. Add typed errors in `errors.go` containing the binary path, runtime version, generated version and commit, and an installation hint. In `codex.go`, apply the policy before spawning the default process; do not probe a caller-supplied custom transport. Under `RequireMajorMinor`, accept patch differences within the same major/minor, reject known incompatible versions and unparseable/unprobeable binaries, and make `errors.As` useful. Under `Warn`, retain structured logging and document that a logger is required to see the warning. `Ignore` performs no compatibility probe.

Add unit tests for exact versions, same-major/minor patch differences, incompatible versions, missing binaries, unparseable output, each policy, nil logging, and custom transports. Add log-capture tests with unique secret markers in modern and legacy command/cwd/path fields. The milestone is accepted when default client construction cannot silently continue with a known incompatible binary, safe approval logs contain no markers, and existing auto-approval source and decision behavior remains compatible.

### Milestone 5: Prototype and promote typed protocol unions

Begin with fixtures under `internal/codegen/testdata/` copied from schemas exported from the exact current generated Codex tag. The prototype must cover `SandboxPolicy`, `UserInput`, `ThreadStatus`, and `ResponseItem`, including unit variants, object variants, arrays, nested references, and the discriminator naming styles present upstream.

Prototype a focused renderer that produces deterministic formatted wrapper types with custom `MarshalJSON` and `UnmarshalJSON`, typed kind constants, raw access, known-variant required-key checks, and raw preservation for well-formed future discriminator values. Compare this with keeping focused manual types. Record the chosen representation, validation boundary, unknown-variant behavior, and source-compatibility impact in the Decision Log before changing broad generation.

Promote the prototype only if representative known variants round-trip with equivalent JSON, invalid or ambiguous discriminators and missing required keys fail with schema/type context, unknown future discriminators are preserved where forward compatibility requires it, generated code is deterministic and formatted, and the high-level `Input` validation remains at least as strict as today. Field value types and other complete schema constraints remain server-validated for this low-level raw-wrapper release.

After promotion, validate the complete upstream `oneOf`/`anyOf` inventory before the compatibility sanitization pass, generate raw-preserving typed-kind wrappers for supported discriminated alternatives, and validate required-key presence for known variants. Maintain reviewed allowlists for genuinely opaque or non-discriminated schemas and fail when the inventory or weak generated output changes. Print per-schema inventory evidence for review before accepting a new aggregate digest. Update hand-written compatibility types and high-level adapters deliberately.

Add a generated API comparison against the latest published SDK tag. Document every intended low-level source break and the migration path. Do not publish the union change under an existing or patch release tag; resolve the SDK/upstream version decoupling decision and release it as a new SDK minor version. The milestone is accepted when no union is silently erased, no named weak fallback exists outside the allowlist, fixtures and current upstream schemas pass, `go generate ./...` is deterministic, and high-level examples still compile and run.

### Milestone 6: Run trusted e2e and publish only after protected gates

Configure the trusted e2e job to use the same exact version and checksum-verified Codex installer as release validation. Scope `OPENAI_API_KEY` to the single shell step that constructs `CODEX_E2E_LOGIN_PARAMS_JSON`; do not expose the key to checkout, setup, download, or verification steps. Construct the JSON without echoing it, unset `OPENAI_API_KEY`, then execute `go test -tags=e2e ./test/e2e` with only the helper's expected variable. Retain and extend stderr secret-leak assertions. Ensure credentialed tests fail rather than silently skip in trusted CI when the expected secret is absent.

Change `.codex/skills/update-codex-protocol/scripts/update_codex_protocol.sh` and its `SKILL.md` so local updating prepares reviewed changes but cannot publish a branch or tag. Remove direct `--push-release`, stop using `git add .`, and either leave changes for explicit review or stage only named generated artifacts after verifying that no unrelated files, including `GC.md`, are included. Run the complete local quality gate before reporting readiness.

Add `.github/workflows/release.yml` with manual dispatch through a protected `release` environment. It must run from `main`, validate that the requested new SDK tag is greater than existing tags, validate the recorded upstream protocol version and CLI checksum, reuse all quality checks, require the trusted e2e result, and only then create and push the annotated tag. The publish job receives minimal `contents: write`; it pushes only the tag and refuses to move an existing tag.

Configure branch protection plus `e2e` and `release` environment approvers in GitHub settings. Record the settings and required checks in `WORKFLOW.md`. If this external configuration cannot be completed, release publication remains blocked; local direct push is not a fallback.

The milestone is accepted when unit CI has no secrets, the trusted job proves credentialed tests actually ran, failed quality or e2e makes publication unreachable, updater flags cannot push, `main` is never pushed by the release job, and existing tags cannot be changed.

## Concrete Steps

Work from `/Users/pme/src/pmenglund/codex-sdk-go` on the feature branch recorded in this plan. Before each milestone, run:

    git status --short --branch
    git diff --check

Confirm that `GC.md` is not staged. Implement one milestone per pull request unless a tracker issue explicitly records a smaller split. Use tests first for cancellation, overflow, redaction, compatibility, and union behavior.

The local quality loop, once Milestone 1 establishes its pinned tool versions, is:

    test -z "$(gofmt -l $(git ls-files '*.go'))"
    go vet ./...
    go test ./...
    go test -race ./...
    staticcheck ./...
    govulncheck ./...
    git diff --check

For generator changes, resolve `CODEX_REPO_ROOT` without printing `.envrc`, use the exact upstream tag recorded by `protocol.GeneratedCodexVersion`, and run:

    CODEX_REPO_REF=rust-v<generated-version> go generate ./...
    git diff --check
    CODEX_REPO_REF=rust-v<generated-version> go generate ./...
    git status --short

The second generation must produce no additional diff. The implementation should provide a repository command or script that performs this check without requiring contributors to interpolate the version manually.

The trusted environment, not an ordinary local shell, runs:

    go test -tags=e2e ./test/e2e

Do not claim e2e success from `go test ./...`; the build tag is mandatory.

At the end of every milestone, update `Progress`, `Surprises & Discoveries`, `Decision Log`, and `Outcomes & Retrospective`, then attach the exact command results to the pull request.

## Validation and Acceptance

The final validation matrix is cumulative:

- Cancellation: replay and real transport evidence show `turn/interrupt` after a synchronous context cancellation, while the returned error remains exactly the original context error.
- Notifications: an intentionally abandoned subscription terminates at its hard capacity with a typed overflow error; another subscriber and a concurrent RPC call continue.
- Security baseline: the chosen patched Go minimum passes `govulncheck`, and no called known vulnerability remains. Any exception is a documented release blocker.
- Approval logs: secret-marker tests cover current and legacy command, cwd, path, grant, and permission payloads; default logs contain none.
- Compatibility: default construction rejects incompatible major/minor versions with a typed actionable error; same-major/minor patches, Warn, Ignore, and custom transports behave as documented.
- Generation: representative and current upstream unions round-trip, unsupported unions fail generation, remaining opaque types are allowlisted, generated files are formatted, and a second generation is clean.
- CI: fork-style pull requests complete required checks without secrets or Codex installation. Trusted e2e uses the pinned verified CLI and does not skip credentialed tests.
- Release: a deliberately failed quality or e2e job makes the tag step unreachable. A successful protected run creates only the requested new immutable tag from `main`.
- Regression: all examples compile; `go test ./...`, race, vet, Staticcheck, vulnerability scanning, formatting, deterministic generation, and trusted tagged e2e are green.

After implementation and local verification are otherwise complete, run `review-security`, `review-qa`, `review-architecture`, `review-ux-specialist`, and `review-sre` agents in batches allowed by the concurrency limit. Address actionable findings before staging. Record each review's findings and resulting changes in this plan.

## Idempotence and Recovery

All local validation commands are safe to repeat. Generation must be deterministic, and the pinned installer must overwrite only a temporary destination or an explicitly named runner path after checksum verification.

Deliver milestones as separate pull requests so cancellation, notification delivery, CI, UX, unions, and release automation can be reverted independently before publication. The union milestone is the primary compatibility risk and must not be combined with an unrelated upstream protocol refresh.

If cancellation cleanup fails, preserve the caller's original context error, log the interrupt failure without sensitive payloads, and keep the client usable. If notification overflow logic destabilizes the reader, revert that milestone rather than restoring an unbounded hidden queue without an explicit decision.

If a CI or e2e log contains a credential, stop the workflow, rotate the credential immediately, preserve only redacted evidence, and treat the release as blocked. Never print `.envrc`, the login JSON, or archive digests alongside secret values.

Never move or delete a published release tag. If publication is wrong, fix forward and publish the next valid SDK version. If branch or GitHub Environment protection is unavailable, keep release publication blocked rather than restoring local direct push.

## Artifacts and Notes

Durable artifacts expected by completion are this updated ExecPlan, pinned Codex version/checksum metadata, the reusable installer, union generator fixtures, migration notes for low-level protocol callers, updated workflow documentation, protected quality/e2e/release workflows, and concise review evidence.

`GC.md` is not a required completion artifact and must remain outside commits unless the user explicitly changes that instruction. Do not create a separate `record.md` unless implementation evidence becomes too dense for this plan; if one is added, define its format here first and keep only concise durable evidence before merging.

## Interfaces and Dependencies

The final implementation should expose these concepts, with exact names confirmed during the relevant milestone:

- In `turn_handle.go`, a private best-effort cancellation helper with a bounded cleanup timeout. Public `Run` signatures and exact context errors remain unchanged.
- In `rpc`, an exported notification-overflow sentinel and/or `NotificationOverflowError` usable with `errors.Is` and `errors.As`. `SubscribeNotifications(buffer int)` documents the hard capacity.
- In the root package, `CompatibilityPolicy` values for RequireMajorMinor, Warn, and Ignore, plus a typed compatibility error containing runtime/generated versions, commit, path, and remediation.
- In approvals, a safe rejecting or callback-policy handler, unchanged `AutoApproveHandler` struct shape and decision behavior, redacted default logging, and an explicitly unsafe sensitive-logging opt-in.
- In generated `protocol`, typed union wrappers with deterministic JSON methods and an explicit opaque-schema allowlist. No upstream union may silently become `interface{}`.
- In CI and releases, one repository-pinned source of truth for Codex versioned asset digests shared by trusted e2e and protected publication.

Prefer the standard library for cancellation, bounded queues, hashing, JSON, and formatting. New build-time tools must be pinned, compatible with the selected Go versions, and must not become runtime dependencies of SDK consumers. Any source-breaking low-level API change requires migration notes and a new SDK minor version.

Plan revision note (2026-07-21): Initial plan created from the ten-item repository review, current code inspection, repository workflow rules, and an independent planner proposal. The user subsequently approved local-plan-only tracking and implementation under the recommended policy assumptions.
