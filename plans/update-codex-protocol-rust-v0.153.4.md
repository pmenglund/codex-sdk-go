# Update Codex protocol to rust-v0.153.4

This living ExecPlan follows `PLANS.md`. Keep Progress, Surprises & Discoveries, Decision Log, and Outcomes & Retrospective current.

## Purpose / Big Picture

Update the SDK from Codex 0.147.0 to stable `rust-v0.153.4`, commit `3d2ee51ca2d5db578f328aa75e20aa22c0197c9a`, and deliver SDK v0.153.4 through a protected pull request, human E2E approval, and automatic publication after E2E succeeds. Users receive current protocol types and RPC methods while existing handwritten APIs retain intentional source and wire compatibility.

## Tracker Mapping

Workflow: `WORKFLOW.md`. This plan records the direct maintenance request of 2026-09-04. The connected Linear server exposes only Colin and Bothnia teams; no matching SDK task was found. Do not create work in an unrelated project. Earlier release records remain historical; this plan owns this update and follows the current repository-local update skill.

## Progress

- [x] (2026-09-05) PR #4 merged as 28b645f with stable Go 1.26/1.27 checks and Codecov project 80.6%/patch 97.0%. Release 33984848890 failed on empty-thread unarchive; v0.153.4 remains unpublished.
- [x] (2026-09-05) User approved the lifecycle correction, local mock provider, required E2E reviewer, and protected release recovery. Focused planner proposal validated against exact upstream 0.153.4 fixtures. Reproduced the old test failure three times against checksum-verified 0.153.4 with credentials removed and provider confined to loopback.
- [x] (2026-09-05) Separated empty-thread rejection from persisted archive/unarchive success. Both API forms verify restored ID, history, and archived/unarchived list membership against the local Responses fixture; ten consecutive race-enabled runs passed on exact CLI 0.153.4.
- [x] (2026-09-05) E2E now requires pmenglund (32618), matching release; before/after verification preserved branch policy, admin bypass, wait behavior, and environment identity. Clarified unchanged-retry versus corrected-candidate dispatch without changing validation logic.
- [x] (2026-09-05) Full Go 1.26.8/1.27.1 vet/unit/race/Staticcheck/govulncheck gates passed; tagged E2E static analysis, actionlint, and release identity/transition fixtures passed. Full tagged E2E passed with race detection on Go 1.25.14 using no credentials and loopback model routing; credentialed cases retained their normal skips.
- [x] (2026-09-05) QA found no issues. Security found mock children inherited hosted login JSON; both credential variables are now cleared and retained only as redaction/leak-check markers. An actual child-process regression failed on both spawn paths before the fix and passed afterward. All three focused tests passed ten race-enabled repetitions; the full uncredentialed Go 1.25.14 tagged suite and tagged static analysis passed again. Security confirmed the finding closed.
- [x] (2026-09-05) Published correction 8ad4181 in PR #5. Hosted Quality and Codecov project passed; patch coverage exposed five shared-helper lines absent from the unit profile. Added a fake-CLI unit regression proving literal config argument forwarding and initialization on both spawn paths; all five changed lines are exercised without changing coverage policy. Full unit/static checks passed; the helper package passed race tests on Go 1.25.14/1.26.8/1.27.1.
- [x] (2026-09-05) PR #5 merged as e0c347a11e86fa1ca24816173bb9f5ae2c337340 after all hosted checks passed; Codecov project reached 81.9%.
- [x] (2026-09-05) Dispatched Release 33989299001 for exact corrective merge e0c347a11e86fa1ca24816173bb9f5ae2c337340. Validation and Quality passed; E2E awaits human approval.
- [x] (2026-09-05) Removed only the release environment required-reviewers rule. Before/after API comparisons prove release identity, branch policy, and admin-bypass setting unchanged and the complete E2E environment unchanged. Unit tests and diff hygiene passed.
- [x] (2026-09-05) Independent security review found no actionable issues with the reviewer removal, successful-job dependencies, candidate/tag validation, or updated guidance.
- [ ] Merge matching automatic-publication guidance through a protected PR after hosted checks pass.

- [x] (2026-09-05) Published reviewed protocol commit c581768 in PR #4; both Go jobs and the coverage-upload job passed. The later Codecov report failed project coverage at 74.7%, despite 95.5% patch coverage.
- [x] (2026-09-05) Inspected the user's Go 1.26/1.27 and Codecov-comment follow-up, obtained the required focused planner proposal, and verified its named functions and fixtures locally.
- [x] (2026-09-05) Added stable Go 1.26/1.27 checks with exact toolchains 1.26.8/1.27.1 and Staticcheck 0.8.1; expanded lifecycle, schema, approval, and overload tests without changing coverage policy. Both full Go gates passed; final added tests passed race/static analysis on both families. Go 1.25.14 unit/race compatibility checks also passed.
- [x] (2026-09-05) Verified both new hosted Go checks and Codecov statuses, updated the two required Go names, merged PR #4, and identified the failed lifecycle release gate.

- [x] (2026-09-05 UTC) Verified clean main and origin/main at `6a14316cd1a6a2cc206b7d91b49b043dc3524d57`; fetched upstream and confirmed latest stable rust-v0.153.4 and latest SDK v0.147.0.
- [x] (2026-09-05 UTC) Completed the required planner pass; verified its generator, metadata, compatibility, and release commands against current files. Created `codex/update-protocol-v0.153.4` before generation.
- [x] (2026-09-05 UTC) Exported exact 0.153.4 schemas; reviewed all 407 printed union hashes/JSON and approved the union digest.
- [x] (2026-09-05 UTC) Reviewed the 100-entry opaque inventory, completed compatibility adapters and wire/RPC regression tests, and documented wrapper migration.
- [x] (2026-09-05 UTC) Updated CLI metadata and SDK version; two generations were byte-identical. Formatting, fixtures, vet, unit/race tests, and Staticcheck passed.
- [x] (2026-09-05 UTC) Complete updater passed with Go 1.26.8: deterministic generation, formatting, installer/metadata fixtures, vet, unit/race tests, Staticcheck, govulncheck, and diff check. Go 1.25.14 independently passed coverage/unit/race tests, vet, Staticcheck, govulncheck, and actionlint. Release metadata fixtures also passed.
- [x] (2026-09-05 UTC) QA, architecture, and security reviews completed; all findings fixed and independently confirmed closed. The reviewed updater passed; all Go gates passed again after the final empty-approval-kind guard, with a focused minimum-Go race check.
- [x] (2026-09-05) Committed and published original protocol update in PR #4.
- [x] (2026-09-05) Published and merged the reviewed Go/coverage follow-up in PR #4 after every required and Codecov check passed.
- [ ] Monitor Release through required human approval and verify the immutable tag and merged commit.

## Surprises & Discoveries

Live main branch protection requires version-specific checks `Go 1.25.12`, `Go 1.26.5`, and `Coverage report`, with strict up-to-date enforcement and administrator enforcement. The patched matrix emits `Go 1.25.14` and `Go 1.26.8`. The later user follow-up authorizes replacing those two required check names with stable Go 1.26 and Go 1.27; preserve all other protection settings and never bypass the gate.

At the lifecycle correction checkpoint, the skill permitted the protected PR handoff and stopped at E2E or release environment approval. The subsequent user request moves the sole human approval to E2E, with publication automatic after it succeeds. Its former dedicated-credential prerequisite is absent from current main. No credential configuration is part of this update.

The local `.envrc` contains an old CODEX_REPO_REF; never source or print it. Explicitly unset that variable for the updater so it chooses the latest stable tag. The verified upstream checkout is `/Users/pme/src/openai/codex`.

Upstream declares Rust 1.95.0, but the installed Homebrew Rust 1.94.0 successfully built the exact CLI and exported schemas in 4m11s. No toolchain change was necessary. The first updater stopped at the union inventory guard before writing generated output.

The new manual wire regression tests first failed on thread model/project/effort, resume/fork excludeTurns, per-turn service tier and trigger, approval kind, and the MCP policy amendment constructor. Generator regression tests also failed because single-variant unions were skipped and shared required fields were not included in variant validation. Both generator fixes now pass the complete generator test package.

## Decision Log

The user's subsequent 2026-09-05 request explicitly authorizes automatic publication when E2E passes. Remove only the required-reviewers rule from GitHub's release environment. Preserve the E2E reviewer, both environments' protected-branch policies, and the existing publish job dependencies on validation, Quality, and E2E. Keep environment: release for deployment restrictions and audit history. No workflow code, SDK version, release candidate, or credential changes are needed. This supersedes the earlier requirement for a second release approval.

The release correction keeps SDK version 0.153.4 and pinned CLI 0.153.4. The previous version is unpublished, and the version-transition validator correctly forbids a further bump. An unchanged retry uses the failed candidate SHA; a code correction uses the exact reviewed corrective merge SHA. Clarify those two paths in the workflow input, WORKFLOW.md, and update skill, without relaxing SHA/main-ancestry/tag or approval gates.

Add test-only ConfigOverrides passthrough and a loopback HTTP Responses fixture matching upstream rust-v0.153.4's response.created/output_item.done/completed sequence. Complete a deterministic turn before archive and require successful archive/unarchive and restored state. Empty-thread failures are asserted separately; do not broaden tolerated archive errors or retry active-writer failures. No production SDK wrapper or generated API change is planned. Detailed baseline logs stay in /private/tmp/codex-sdk-e2e-before.log; final results will be summarized here and in the corrective PR.

User approval includes configuring e2e's required reviewer as pmenglund (32618), matching release's prevent_self_review=false and keeping existing branch policy/admin bypass/wait-timer behavior. Never approve either environment on the user's behalf.

The 2026-09-05 user follow-up explicitly authorizes updating required Go checks to families 1.26 and 1.27 and addressing Codecov comment 5549891090 on PR #4. Use stable check names with exact current toolchain pins 1.26.8 and 1.27.1. Preserve the SDK minimum in go.mod because this request concerns CI. Staticcheck v0.7.0 cannot decode Go 1.27 export data; v0.8.1 passed both CI toolchains and is pinned in CI, the updater, and contributor guidance. Preserve Coverage report, strict checking, app bindings, and all other branch protections when replacing the old Go contexts.

Extend the existing plan with replay-backed direct lifecycle/validation/fork-error tests in account_thread_test.go or a focused thread_lifecycle_test.go, composed-schema/ref/cycle and RPC schema/output failure tests in internal/codegen/schema_coverage_test.go, and malformed approval JSON tests in protocol/protocol_0153_test.go. If needed, add approval/error-classification tests for real uncovered behavior. Codecov counts 2,952 fully covered lines out of 3,948; recover at least 207 lines to pass 80%, aiming for 82% headroom. Go statement percentages are supporting evidence; the hosted Codecov project and patch checks are authoritative. Do not lower targets or expand exclusions. GitHub browser settings verified Codecov installed with repository access on September 5; its old installation warning requires no additional grant.

Independent architecture review identified stale opaque allowances and fail-open parsing of unrecognized RPC params. Remove all five wrapper fallback allowances (including ClientNotification), share a params parser that rejects unsupported present schemas, and use pointer-aware type rendering for nullable inbound params. QA and security review identified omitted approval kind not receiving the upstream command default; apply that default during decoding, retain explicit writeStdin/future string kinds, reject null, non-string, and empty kinds, and verify legacy RPC dispatch. QA also requested successful single-variant constructor round trips and exact pagination request bodies; added both. The new regression tests failed before the fixes and pass afterward. The final updater rerun passed after the generator/default/test corrections. The subsequent explicit-empty-kind guard changed only handwritten decoding; the full Go gate and focused minimum-Go race checks passed afterward, while generated output remained unchanged.

Update the minimum Go patch from 1.25.12 to 1.25.14 and CI newer patch from 1.26.5 to 1.26.8. The complete gate found four reachable standard-library advisories (GO-2026-6218, GO-2026-6090, GO-2026-5972, GO-2026-5026) fixed in 1.25.13/1.26.6. Use the latest patches in the existing two release families, verified against https://go.dev/doc/devel/release, to keep hosted vulnerability checks and runtime builds aligned. Update go.mod, CI, README, APP, and WORKFLOW together; no language minor-version increase or dependency change.

Preserve AccountUsageRead(ctx) and add AccountUsageReadWithParams(ctx, *protocol.GetAccountTokenUsageParams); nil still omits the params member. Reject unsupported multi-reference request unions. Preserve concrete thread/turn/item pagination results with reviewed manual wire structs. Thread.projectId is required nullable; section appearance has outbound preserve/clear/replace semantics. No preexisting exported declaration names were removed.

Approved opaque digest `a6f55f1b748d2bce87b46d67d107373fe4ffa17d13a8e20603a4e58d3d3c9319`, still 100 entries. The exact AST-equivalent baseline reproduces the previous digest. Four additions are reviewed: `MCPServerEventNotification.Params` and `ResponseUsageMetadata.Metadata` have unrestricted JSON schemas; `SanitizedGetAccountRateLimitsResponseJSON.RateLimitUpsell` is a backend-owned banner with no schema constraints; `SanitizedCommandExecutionRequestApprovalParamsJSON.Kind` is a legacy sanitized compatibility surface whose allOf enum is not handled by go-jsonschema. The canonical manual approval params expose typed CommandExecutionApprovalKind. Broad allOf normalization was tested but would introduce unrelated configuration export changes, so it was not retained. Four types in the opaque field inventory (CapabilityRootLocation, DynamicToolNamespaceTool, LocalShellAction, ReasoningItemReasoningSummary) and the previous fallback ClientNotification become raw-preserving typed wrappers through single-variant support. All five constructor migrations are documented.

Reviewed and approved union digest `aa19cfcd55fcee125e20373138c395e420a673088dc3bb039d372e8d4b726fe2` after the source exporter printed all 407 per-schema hashes and JSON, matching the independent committed-schema comparison. The baseline's 334 entries exactly reproduce the prior approved digest; there are 101 added and 28 removed canonical schemas. Sixty-five additions are nullable references (including namespace copies and description/default variants). Remaining changes reduce to 19 structural groups: RPC container changes, login and auth additions, elicitation modes, realtime item/presentation/timeline/transport unions, error and legacy approval enum additions, configuration sources, configured hooks and hook metadata, thread items, guardian stdin approval actions, function-call output optionality, and image-generation failures. New stable discriminators use existing typed union generation; AuthMode/CodexErrorInfo and MCP elicitation retain their existing reviewed opaque representation. No new opaque union family is approved. Section appearance updates require manual three-state serialization; legacy MCP approval needs a recognized constructor value. Opaque generated field inventory was separately reviewed as recorded above.

Use SDK version 0.153.4 because it matches the target protocol and exceeds latest immutable SDK tag v0.147.0. Source compatibility decisions remain subject to the actual generated diff. Preserve old names only when semantics are compatible; removed wire fields may remain deprecated with `json:"-"`.

Use the exact detached worktree `/private/tmp/codex-sdk-01534-upstream` and Cargo target `/private/tmp/codex-sdk-01534-cargo` for compatibility iterations, with CODEX_REPO_REF unset. The first two updater runs discarded their worktrees as designed; a stable diagnostic worktree avoids rebuilding upstream for every Go-only correction. The final authoritative updater still must independently generate twice and run all gates.

Use one flat release plan, matching repository convention. Keep large schema inventories and build logs in temporary storage; retain only decisions and validation evidence in this file. The maintenance invocation authorizes implementation and the skill's protected publication workflow without another plan approval.

## Outcomes & Retrospective

The protocol/SDK 0.153.4 update is merged in PR #4 at 28b645f. Generation was deterministic and all local and hosted Quality checks passed. Codecov project coverage rose to 80.6%, and patch coverage reached 97.0%. Release 33984848890 exposed an invalid test sequence: unarchive followed a tolerated archive failure on a thread without persisted history. No v0.153.4 tag was published.

The corrective branch now uses separate empty-thread rejection tests and strict persisted-thread round trips with a real CLI and loopback mock provider. The original failure reproduced three times; the corrected tests passed ten race-enabled repetitions, and the full uncredentialed tagged suite passed on the release Go version. Both CI Go families passed all local gates. Required human E2E approval is restored. QA and security review are complete; the credential-isolation finding is fixed and verified with a failing-before/passing-after child-process regression. PR #5 passed hosted Quality and Codecov and merged at e0c347a. Corrective Release 33989299001 passed validation and Quality and awaits E2E approval. The user then requested automatic publication after successful E2E; publication remains pending.

## Context and Orientation

`gen.go` invokes `internal/codegen/main.go`, which runs the upstream Codex CLI schema exporter. It checks every JSON `oneOf`/`anyOf` union against a reviewed digest, generates protocol types and RPC dispatch, and rejects unreviewed opaque `interface{}` shapes. Handwritten wire structs live in `protocol/manual_types.go`; adapters live in root Go files including `thread_lifecycle.go`. Generated output belongs in `protocol/*_gen.go` and `rpc/*_gen.go`, with exported GoDoc and the exact source commit.

`.github/codex/version` and `.github/codex/checksums.txt` pin the runtime used for hosted tests. `.github/sdk-version` selects the SDK tag independently. `.github/workflows/release.yml` runs validation, Quality, Trusted E2E, and protected tag publication. Only GitHub Actions may publish a tag.

## Plan of Work

First run the bundled updater. If schema union validation fails, print the complete target and 0.147 inventories outside the repository, compare every added/changed canonical schema, and classify nullable references, discriminated variants, containing-schema changes, and ambiguous shapes. Add typed handling or an explicit opaque rationale before approving any digest. Repeat this review for the generated opaque-interface inventory.

Next inspect the complete generated diff, including RPC additions/removals, exported names, required fields, nullability, manual schema coverage, and generated GoDoc. Reconcile affected thread start/resume/fork/list/metadata/turn adapters and approval types. Add focused generator, JSON wire, routing, and external API tests for actual behavior changes. Never repair generated files manually or add unrelated convenience APIs.

Then update CLI version and four official release archive digests, and set SDK version 0.153.4. Run the updater to prove two identical generations and the full gate. Run independent QA and architecture reviews after implementation and validation; add security review if approval boundaries change. Resolve actionable findings before explicit-path staging and protected PR handoff.

Finally watch required PR checks, fix failures, and merge only through the PR when protection permits. Monitor the automatic Release run and report its URL. Pause at required PR review or E2E approval; after human approval and successful E2E, verify the automatically published immutable tag v0.153.4 resolves to the merged main commit.

## Concrete Steps

Run from `/Users/pme/src/pmenglund/codex-sdk-go`. Initially the only dirty path is this plan, so `--allow-dirty` is authorized for that path.

    env -u CODEX_REPO_REF CODEX_REPO_ROOT=/Users/pme/src/openai/codex GOTOOLCHAIN=go1.26.8 .codex/skills/update-codex-protocol/scripts/update_codex_protocol.sh --allow-dirty

For inventory diagnosis use the exact target ref with `go generate ./...`, adding `CODEX_PRINT_UNION_INVENTORY=1` or `CODEX_PRINT_OPAQUE_INTERFACE_INVENTORY=1`. Capture logs outside the repository. Compare the old inventory using isolated baseline files; never overwrite the update with baseline generation.

    go test ./internal/codegen
    go test . ./protocol ./rpc ./internal/codegen
    .github/scripts/validate-codex-metadata.sh
    .github/scripts/test-validate-codex-metadata.sh
    .github/scripts/test-install-codex-cli.sh
    .github/scripts/sdk-release-tag.sh
    .github/scripts/validate-release-tag.sh v0.153.4

Obtain CLI archive digests from the official `openai/codex` GitHub release API for rust-v0.153.4, using only the four existing supported archive names. Verify returned SHA-256 digests before writing metadata. Re-run all release fixtures and workflow validation where required by CI.

## Validation and Acceptance

All generated headers must name `3d2ee51ca2d5db578f328aa75e20aa22c0197c9a` and GeneratedCodexVersion must be 0.153.4. Every inventory change needs a documented decision. Known variants must decode to concrete types and unknown variants must preserve raw JSON. Tests must cover changed adapter mappings, omission/null behavior, and intentional source compatibility.

The updater must pass two byte-identical generations, formatting, metadata/installer fixtures, vet, unit tests, race tests, Staticcheck v0.8.1, govulncheck v1.3.0, and diff hygiene. Required hosted Quality and E2E checks must pass before GitHub publishes v0.153.4 at the merged candidate. Local gates and hosted acceptance must be reported separately.

## Idempotence and Recovery

Recheck the worktree before every mutation and stage only named reviewed files. Preserve concurrent changes. Do not edit upstream toolchain requirements or downgrade the selected tag to avoid an environment error. If a build fails, record the exact failure and use a compatible toolchain where available. Before publication the feature branch is reversible; after publication repair with a higher SDK version. An unchanged Release retry uses its exact failed candidate SHA; a reviewed code correction uses its exact corrective merge SHA with the unpublished SDK version. Never supply an arbitrary tag. Never push main directly, create or move local SDK tags, approve GitHub environments, or bypass failed checks.

## Artifacts and Notes

The official release API returned all four required archive digests for 0.153.4. Root JSON-RPC schema inspection shows four added client methods and eleven added notifications, with no removed client methods, server requests, or notifications. Actual generated export remains authoritative for compatibility decisions.

## Interfaces and Dependencies

Keep runtime concerns in the root package, transport in rpc, and wire data in protocol. Reuse the existing go-jsonschema dependency and generator mechanisms. No new runtime dependency is planned. Schema compilation requires the upstream Rust toolchain and source checkout; local Go validation uses CI-aligned Go 1.26.8 and Go 1.27.1, with a separate Go 1.25.14 unit/race compatibility check.

Revision note: Created after live baseline inspection and the required planner review, before schema generation or code changes.
