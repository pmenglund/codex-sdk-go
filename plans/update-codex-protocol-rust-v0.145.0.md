# Update generated Codex protocol to rust-v0.145.0

This ExecPlan is a living document. The sections `Progress`, `Surprises & Discoveries`, `Decision Log`, and `Outcomes & Retrospective` must remain current while the work proceeds.

This plan follows `PLANS.md` at the repository root. It records one maintenance refresh and does not reopen the completed plans that created the updater, hardened protocol unions, improved API usability, or enforced generated Go documentation.

## Purpose / Big Picture

After this work, the Go SDK will use the latest stable upstream Codex app-server protocol, `rust-v0.145.0`, instead of `rust-v0.144.6`. Callers will be able to use the newly published app requests, environment notifications, audio inputs, Amazon Bedrock login variants, scheduled-task schedule values, and other protocol additions without sending wire shapes that Codex 0.145 rejects.

The result is observable by inspecting `protocol/metadata_gen.go`, compiling the new generated RPC methods and notification parsers, exercising the changed handwritten approval and input adapters, and running the repository updater successfully. The updater must generate twice with identical output and pass formatting, vet, unit tests, race tests, Staticcheck v0.7.0, govulncheck v1.3.0, and `git diff --check`.

This workflow leaves every change unstaged. It does not commit, tag, push, publish a release, or run credentialed end-to-end tests.

## Tracker Mapping

Workflow: `WORKFLOW.md`.

The tracked task is the user's direct invocation of `.codex/skills/update-codex-protocol/SKILL.md` on 2026-07-27. Following the repository's existing local-plan precedent, this file is the durable task record; no Linear issue was supplied or created. The completed `plans/update-codex-protocol-skill.md`, `plans/repository-hardening.md`, `plans/api-usability-hardening.md`, and `plans/coverage-and-generated-godoc.md` remain historical records and are not extended by this refresh.

## Progress

- [x] (2026-07-27) Confirmed the checkout was clean on `codex/awesome-go-quality` before running the updater.
- [x] (2026-07-27) Ran the updater, selected stable `rust-v0.145.0`, and stopped at the changed-union safety gate without approving the new digest.
- [x] (2026-07-27) Captured complete union inventories for `rust-v0.144.6` and `rust-v0.145.0`; compared 317 old schemas with 324 new schemas.
- [x] (2026-07-27) Completed the required independent planner pass and validated its recommendations against the live tag and repository. Corrected its stale proposed upstream commit using the local tagged checkout.
- [x] (2026-07-27) Created this separate ExecPlan before compatibility edits.
- [x] (2026-07-27) Added tests that failed against the old simple denial and missing high-level audio constructors, then passed after the compatibility edits.
- [x] (2026-07-27) Updated the handwritten review-decision, rejecting-approval, and high-level input adapters.
- [x] (2026-07-27) Recorded the final union decision and updated the approved aggregate inventory digest.
- [x] (2026-07-27) Regenerated from `rust-v0.145.0`, inspected the generated API, and completed focused protocol, RPC, and generator tests.
- [x] (2026-07-27) Ran architecture and QA reviews after the initial successful full gate; fixed their nullable-reference, opaque-inventory, removed-enum, and focused-test findings.
- [x] (2026-07-27) Reran the updater's complete deterministic quality gate after the review fixes; both generations matched and every gate passed with Go 1.26.5.
- [x] (2026-07-27) Finished the compatibility summary and confirmed all changes remain unstaged.

## Surprises & Discoveries

- Observation: The protocol refresh is not a generated-header-only patch.
  Evidence: The updater stopped with `schema union inventory changed` and reported digest `80fb7367165ee22e9e50553f7d6463ce0667e029c91b58950b6fa63d6e8ffcab` across 324 schemas.

- Observation: Comparing full inventory hashes overstates the number of independent wire changes because large schema bundles embed the same nested unions.
  Evidence: The old and new inventories differ by 27 removed and 34 added hashes, but the recurring changes reduce to app and environment methods, audio inputs/content, Amazon Bedrock login, structured denial, scheduled-task schedules, nullable path/token-usage references, and changed containing schemas.

- Observation: The independent planner returned an upstream commit that did not match the live stable tag.
  Evidence: `git -C /Users/pme/src/openai/codex rev-list -n 1 rust-v0.145.0` returned `25af12f7e61572b0bc18ddb1008be543b91519b0`; this verified value governs generation and header review.

- Observation: The existing high-level account login adapter is already forward-compatible with a new typed login variant.
  Evidence: `account.go` accepts any JSON-marshalable value through `protocol.NewLoginAccountParams`, so no Amazon Bedrock-specific root adapter is needed unless generation disproves that assumption.

- Observation: The compatibility tests failed for the intended pre-change reasons.
  Evidence: `go test ./protocol` rejected the new structured denial as requiring a string, and `go test .` failed to compile because `AudioInput`, `LocalAudioInput`, `InputTypeAudio`, and `InputTypeLocalAudio` did not exist. Both focused packages passed after the handwritten edits.

- Observation: Schema sanitization discarded nullable `$ref` type information before Go generation.
  Evidence: The first generated `RawResponseCompletedNotification.Usage` field was `interface{}` even though the 0.145 schema is `TokenUsageBreakdown | null`. Collapsing that nullable reference before stripping unsupported unions now generates `*TokenUsageBreakdown`; absent, null, and populated payload tests pass.

- Observation: The pre-existing opaque-output guard covered only top-level `type X interface{}` declarations.
  Evidence: Architecture and QA review identified the untyped `Usage` struct field after the initial full gate. The generator now inventories empty interfaces in struct fields and collection elements as well as top-level declarations; the reviewed 0.145 inventory is 97 entries at digest `c6fa5c442b4de0dc9cc3000f0adf757a9a891b83b88c69199b4247a7c180be5c`.

- Observation: The first complete gate used Go 1.26.4 and stopped only at govulncheck.
  Evidence: GO-2026-5856 is fixed in Go 1.26.5. Re-running with `GOTOOLCHAIN=go1.26.5` passed generation, formatting, vet, unit tests, race tests, Staticcheck, govulncheck, and diff checks before the independent reviews.

- Observation: The first post-review full gate exposed one stale replay fixture after deterministic generation succeeded.
  Evidence: `TestAccountAndModelHelpersWithReplay` omitted the new `Account.type` discriminator and `planType` required by the `chatgpt` account variant. Updating the fixture to the reviewed 0.145 wire shape made the focused test and the subsequent complete gate pass.

## Decision Log

- Decision: Track this refresh in a new plan instead of rewriting a completed plan.
  Rationale: The earlier plans describe completed work with deliberate `rust-v0.144.6` pins. This invocation is a new maintenance event with its own protocol compatibility decisions and validation evidence.
  Date/Author: 2026-07-27 / Codex

- Decision: Treat `rust-v0.145.0` at commit `25af12f7e61572b0bc18ddb1008be543b91519b0` as the only generation source.
  Rationale: The updater selected the highest exact stable `rust-vMAJOR.MINOR.PATCH` tag and ignored `rust-v0.146.0-alpha.*`; the local tag resolution is direct, current evidence.
  Date/Author: 2026-07-27 / Codex

- Decision: Review the changed union inventory by comparing sorted per-schema hashes and JSON from the old and new stable tags.
  Rationale: Approving only the aggregate digest would hide which wire shapes changed. The per-schema comparison separates actual additions from repeated changes in aggregate bundles.
  Date/Author: 2026-07-27 / Codex

- Decision: Do not add an opaque-union or `interface{}` exception preemptively.
  Rationale: The generator already emits raw-preserving typed wrappers for object-discriminated unions and fails closed for unreviewed weak output. New exceptions require a concrete generator error, a narrow rationale, and a focused regression test.
  Date/Author: 2026-07-27 / Codex

- Decision: Approve union inventory digest `80fb7367165ee22e9e50553f7d6463ce0667e029c91b58950b6fa63d6e8ffcab` after the per-schema review, with no new opaque exception.
  Rationale: New object-discriminated variants for audio, Amazon Bedrock login, scheduled-task schedules, client requests, notifications, account, thread items, response items, and related content are covered by the existing raw-preserving typed generator. New nullable reference schemas remain concrete. The denial representation is intentionally handwritten because `ReviewDecision` is an existing compatibility surface, and its tests now enforce the 0.145 object payload with a non-empty `rejection` string. The remaining changed hashes are containing schemas that repeat these reviewed additions.
  Date/Author: 2026-07-27 / Codex

- Decision: Preserve nullable `$ref | null` schemas as references during sanitization and approve an AST-derived inventory of every remaining generated empty-interface use.
  Rationale: A nullable concrete reference should generate a pointer, not erase the referenced type. The remaining opaque fields and collection elements are existing dynamic protocol surfaces; hashing their sorted declaration paths makes additions, removals, and representation changes fail closed without maintaining a long, error-prone field allowlist.
  Date/Author: 2026-07-27 / Codex

- Decision: Retain `AmazonBedrockCredentialSource` and its two constants as deprecated generated compatibility declarations when upstream no longer emits the enum.
  Rationale: Removing an exported SDK type is a source-breaking change. The tombstone preserves compilation for existing callers while directing new code to the direct Amazon Bedrock login payload introduced by protocol 0.145.
  Date/Author: 2026-07-27 / Codex

## Outcomes & Retrospective

The SDK now targets Codex `0.145.0` at upstream commit `25af12f7e61572b0bc18ddb1008be543b91519b0`. Generated artifacts add the new app requests, environment notifications, audio and local-audio variants, Amazon Bedrock login variants, scheduled-task schedules, structured account variants, and the related containing-schema changes. The handwritten compatibility layer now emits structured approval denials and exposes high-level remote and local audio inputs.

The generator preserves nullable references as pointers, so `RawResponseCompletedNotification.Usage` is `*TokenUsageBreakdown` rather than `interface{}`. It also inventories all remaining generated empty-interface uses, including struct fields and collection elements, at reviewed digest `c6fa5c442b4de0dc9cc3000f0adf757a9a891b83b88c69199b4247a7c180be5c`. No new opaque union exception or RPC response override was needed. Deprecated `AmazonBedrockCredentialSource` declarations preserve source compatibility after upstream removed the enum.

Focused tests passed for the root package, `internal/codegen`, `protocol`, and `rpc`. The final updater run with `GOTOOLCHAIN=go1.26.5` selected `rust-v0.145.0`, produced identical first and second generations, and passed formatting, vet, all unit tests, race tests, Staticcheck v0.7.0, govulncheck v1.3.0 with “No vulnerabilities found,” and `git diff --check`.

All generated and handwritten changes, this plan, and the updated replay fixture remain unstaged. No commit, tag, push, release, credentialed end-to-end run, or hosted CI run was performed.

## Context and Orientation

`codex-sdk-go` starts `codex app-server` and exchanges line-delimited JSON-RPC messages. Checked-in generated wire types live in `protocol/*_gen.go`; checked-in RPC client methods, server-request dispatch, and notification parsers live in `rpc/*_gen.go`. `gen.go` runs `internal/codegen`, which exports JSON schemas from a local upstream OpenAI Codex checkout and writes the generated Go artifacts.

`internal/codegen/main.go` contains `approvedUnionInventorySHA256`. Before compatibility sanitization, generation recursively inventories every `oneOf` and `anyOf` schema. A changed aggregate digest stops generation so new weak or ambiguous unions cannot silently become `interface{}`. `CODEX_PRINT_UNION_INVENTORY=1` prints each canonical schema with its individual hash for review.

The generator discovers object-discriminated unions and emits raw-preserving wrappers with typed kind constants in `protocol/unions_gen.go`. Known variants validate required fields; unknown future discriminators preserve their complete JSON. The existing explicit opaque discriminated-union exception is `McpServerElicitationRequestParams`. Other weak generated or fallback types are separately allowlisted and fail closed when a new unreviewed type appears.

`protocol/manual_types.go` contains compatibility types whose wire semantics need handwritten validation. In particular, `ReviewDecision` currently treats `"denied"` as a simple string, while the 0.145 schema requires a single-key object whose `denied` body contains a non-empty `rejection` string. `approvals.go` uses this type for the safe zero-value `RejectingApprovalHandler`.

`input.go` is the high-level input adapter used by turn start and steer operations. It currently supports text, remote image, local image, skill, and mention inputs. Upstream 0.145 adds remote audio and local audio variants to `UserInput`, so the root adapter must expose and validate those variants instead of forcing callers down to low-level protocol values.

The union inventory comparison identified these reviewed categories:

- `ClientRequest` adds `app/read` and `app/installed`.
- `ServerNotification` adds thread environment connected and disconnected notifications.
- `LoginAccountParams` and `LoginAccountResponse` add Amazon Bedrock variants.
- `UserInput`, dynamic tool output content, function-call output content, and general content items add audio shapes; `UserInput` also adds local audio.
- `ReviewDecision` replaces the simple `"denied"` form with a structured denial containing `rejection`.
- `ScheduledTaskSchedule` is a new four-variant object-discriminated union.
- `LegacyAppPathString` and `TokenUsageBreakdown` gain nullable reference positions.
- `FileSystemSpecialPath`, configured hooks, `ThreadItem`, `ResponseItem`, `Account`, and aggregate bundles change because they contain the additions above or add fields to existing variants.

## Plan of Work

Begin with tests. In `protocol/unions_test.go`, replace the accepted simple denied decision with a structured object and add cases proving that the old string, an empty denial object, a missing rejection, and an empty rejection are rejected. Prove that the valid structured object preserves exact JSON through marshal and unmarshal.

In `approvals_test.go`, change the expected legacy patch and command rejection payloads to the structured denial shape. Use a stable non-sensitive rejection explanation and assert it exactly so the handler cannot regress to schema-invalid output.

In `codex_test.go` and the smallest relevant input-focused test, add constructor, validation, and marshaling cases for remote audio and local audio. Empty URL and path values must fail before an RPC request is written. Existing image, text, skill, and mention behavior must remain unchanged.

Update `protocol/manual_types.go` so `ReviewDecisionKindDenied` is a structured kind. Require a non-empty string field named `rejection` in its body. Continue accepting unknown future string or single-key object decisions during wire decoding, while checked construction accepts only known valid variants.

Update `approvals.go` so `RejectingApprovalHandler` returns a structured denial with a fixed explanation that contains no command, path, grant, or user-supplied content.

Update `input.go` with `InputTypeAudio`, `InputTypeLocalAudio`, `AudioInput`, and `LocalAudioInput`. Reuse the existing `URL` and `Path` JSON fields and validate them with audio-specific error messages.

After the changed-union categories and typed-versus-opaque decision are complete in this plan, update `approvedUnionInventorySHA256` in `internal/codegen/main.go` to `80fb7367165ee22e9e50553f7d6463ce0667e029c91b58950b6fa63d6e8ffcab`. Do not hand-edit generated files.

Run focused tests, review the intentional handwritten diff, then run the updater with `--allow-dirty`. This flag is justified only after the existing changes have been inspected; it does not stage or include files. If generation fails because a response type, manual type, or weak output is unresolved, stop at that exact error. Add the smallest generator change and regression test, record the decision here, and rerun the updater.

Inspect every generated file. Headers must name commit `25af12f7e61572b0bc18ddb1008be543b91519b0`, and `protocol/metadata_gen.go` must report version `0.145.0`. Verify concrete `AppRead` and `AppInstalled` methods, environment notification parsers, the new typed union kinds, changed payload fields, generated name-leading GoDoc, and the absence of new unreviewed `interface{}` declarations. No unrelated file may change.

Finish by running the updater again if any handwritten compatibility edit follows generation. A successful updater run is the authoritative complete quality gate and deterministic-generation proof.

## Concrete Steps

Work from `/Users/pme/src/pmenglund/codex-sdk-go`.

Before each mutation group, inspect:

    git status --short --branch
    git diff --check

Run focused tests while implementing:

    go test ./protocol
    go test .
    go test ./rpc
    go test ./internal/codegen

After reviewing the intentional dirty diff, run:

    CODEX_PRINT_UNION_INVENTORY=1 .codex/skills/update-codex-protocol/scripts/update_codex_protocol.sh --allow-dirty

The script must select `rust-v0.145.0`, generate twice, and run its full local gate.

Inspect the result with:

    git status --short
    git diff --stat
    git diff -- protocol rpc internal/codegen/main.go input.go approvals.go
    git diff --check

Do not run `git add`, `git commit`, `git tag`, or `git push`.

## Validation and Acceptance

The refresh is accepted only when all of the following are true:

- `protocol.GeneratedCodexVersion` is `0.145.0`, and every generated header names upstream commit `25af12f7e61572b0bc18ddb1008be543b91519b0`.
- Known new discriminated variants decode as known, validate their required fields, and preserve complete JSON. Unknown future discriminator values still round-trip losslessly.
- `ReviewDecision` accepts `{"denied":{"rejection":"..."}}` and rejects the old simple `"denied"` form plus missing or empty rejection values.
- `RejectingApprovalHandler` sends protocol-valid, non-sensitive structured denials for legacy patch and command approvals.
- High-level remote and local audio inputs marshal to the upstream `UserInput` shapes and reject empty URL or path values before transport.
- Generated `AppRead` and `AppInstalled` methods and environment notifications use concrete protocol payloads without speculative response overrides.
- No new opaque union or unreviewed `interface{}` exception exists unless this plan records the exact generator failure, rationale, and regression test.
- The updater's deterministic second generation and complete quality gate pass.
- `git diff --check` passes and `git status --short` shows all refresh files unstaged.

Credentialed e2e, hosted CI, release publication, and external acceptance are explicitly unverified by this workflow.

## Idempotence and Recovery

The updater is safe to rerun after reviewing the dirty diff and passing `--allow-dirty`. It writes generated artifacts deterministically and compares a second generation pass with the first. If it stops at a compatibility gate, preserve the exact error, make only the reviewed generator or adapter change, and rerun.

The inventory logs are temporary evidence under `/tmp` and are not repository artifacts. They can be regenerated from exact tags with `CODEX_PRINT_UNION_INVENTORY=1`.

Do not recover by editing `protocol/*_gen.go` or `rpc/*_gen.go`, moving an existing SDK tag, or using removed local publish flags. If an unrelated user change appears, stop and separate it before the next `--allow-dirty` run.

## Artifacts and Notes

The reviewed baseline inventory is `3e9c78fe0a86958dcbd03657f4d426277dc7427d1f0ca1d390b732d704359f23` across 317 schemas.

The reviewed target inventory is `80fb7367165ee22e9e50553f7d6463ce0667e029c91b58950b6fa63d6e8ffcab` across 324 schemas.

Temporary per-schema evidence:

    /tmp/codex-union-inventory-0.144.6.log
    /tmp/codex-union-inventory-0.145.0.log

The initial target updater stopped before writing generated artifacts:

    schema union inventory changed: got sha256 80fb7367165ee22e9e50553f7d6463ce0667e029c91b58950b6fa63d6e8ffcab across 324 unique union schemas

## Interfaces and Dependencies

In `input.go`, the completed public additions must be:

    const InputTypeAudio = "audio"
    const InputTypeLocalAudio = "localAudio"
    func AudioInput(url string) Input
    func LocalAudioInput(path string) Input

`AudioInput` uses the existing `Input.URL` JSON field; `LocalAudioInput` uses `Input.Path`.

`protocol.ReviewDecision` retains its public constructor, JSON methods, kind access, and raw-preserving behavior. The `ReviewDecisionKindDenied` constant remains named and valued the same, but its required wire representation changes from a simple string to a single-key object with a non-empty `rejection` string.

No new runtime dependency is introduced. Generation continues to use the repository's existing Go generator and the upstream Rust schema exporter. Validation continues to use the pinned analyzer versions in the updater script.

Revision note: Created on 2026-07-27 after the first updater run exposed a changed union inventory. This plan isolates the 0.145.0 maintenance refresh from completed earlier plans and records the evidence required before approving the new digest.
