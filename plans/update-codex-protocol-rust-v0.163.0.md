# Update Codex protocol to rust-v0.163.0

This living ExecPlan follows `PLANS.md`. Keep Progress, discoveries, decisions, and outcomes current.

## Purpose / Big Picture


Update the Go SDK from Codex 0.162.1 to stable 0.163.0. Callers should be able to exclude threads from listings, retain supplied tool output, and receive the current response metadata and notifications exposed by the stable exporter. Publish SDK v0.163.0 through protected GitHub checks and trusted end-to-end tests. Experimental source APIs are included only when the repository's normal exporter actually exposes them.

## Tracker Mapping


Workflow: `WORKFLOW.md`. Linear project: P-COLIN-6 (codex-sdk). Issue: COLIN-226 owns all milestones below. The completed 0.162.1 plan remains historical.

## Progress


- [x] (2026-10-19 UTC) Verified clean main at origin/main, ebc984e331c824e2e8fcc6c68e37c01282bc63ea, instructions, latest SDK v0.162.1, and highest exact stable upstream rust-v0.163.0.
- [x] (2026-10-19 UTC) Created COLIN-226, validated the required planner proposal, checked live release protections, and created codex/update-protocol-v0.163.0 before generation.
- [x] (2026-10-19 UTC) Source-built exporter printed all 451 reviewed union schemas; 104-entry opaque inventory is unchanged. Two identical generations passed; full gate reached unit tests.
- [x] (2026-10-19 UTC) Repaired exclusion pass-through, added protocol/RPC/replay and real-CLI regressions, updated verified CLI metadata and SDK 0.163.0, and passed focused tests/schema coverage.
- [x] (2026-10-19 UTC) Final authoritative updater passed two identical generations, formatting, metadata/installer fixtures, vet, all unit/race tests, Staticcheck 0.8.1, govulncheck 1.3.0, and diff hygiene. Native Go 1.27.2 vet/unit/race/vulnerability checks and release-policy fixtures also pass.
- [x] (2026-10-19 UTC) Completed independent QA and architecture reviews. QA requested the documented present-null raw-union regression; added it and passed the full unit suite plus protocol race checks on Go 1.26.9/1.27.2. Architecture found no actionable issue and reviewed the QA addition.
- [x] (2026-10-19 UTC) Staged 21 reviewed named paths, committed 8e114032766ebd63be6965b71cd64697217c291e, pushed the feature branch, and opened PR #18. Version-transition validation passed. Codecov project/patch checks pass.
- [x] (2026-10-19 UTC) Required Go 1.26, Go 1.27, Coverage report, and Codecov statuses passed. Merged exact reviewed PR #18 head through protection at 649395559fa6e36219c70e219865d4b4ce4efd66 and deleted the remote feature branch. Preserved the local plan edit and fast-forwarded main to the merge.
- [x] (2026-10-19 UTC) Automatic Release 39315374167 passed validation, both Quality toolchains, coverage, trusted E2E, and publication. Verified annotated v0.163.0 peels to exact gated merge 649395559fa6e36219c70e219865d4b4ce4efd66 using GitHub API and independent git ls-remote.
- [x] (2026-10-19 UTC) Recorded final release evidence and completed outcomes for the protected documentation closeout; completed COLIN-226 after publication verification.

## Surprises & Discoveries


Installed Go is 1.27.1 and Cargo is 1.94.0. The repository's supported patched toolchains are Go 1.26.9 and 1.27.2; Staticcheck 0.8.1 must run with Go 1.26.9. Upstream requests Rust 1.95.0; test the actual exporter before changing tooling. Git metadata writes and network access require sandbox escalation. Never source or print `.envrc`.

Initial protocol regressions fail because ThreadListParams drops empty/populated excludedThreadIds and TurnToolOutput drops retain=true. Upstream defaults retain to false; generated Retain is an optional bool pointer and preserves explicit false as well as true. Existing raw ResponseItem wrappers already preserve the new encrypted/status/annotations/logprobs metadata, including 9007199254740993. The sandbox's first Go test attempt could not load testmain; the escalated rerun produced the actual expected regression failures.

The first full gate failed one new test that incorrectly expected the raw union wrapper to reject a present turnId:null. Existing wrappers validate the discriminator and required-key presence, rather than performing full schema-value validation. Corrected that new test to exercise the established contract; missing turnId still fails. Nullable firstUnread decodes correctly through the existing null-aware union wrapper. No production change or broader validation policy change is needed.

The official 0.163.0 CLI has a reproduced listing limitation: two freshly archived same-second threads appear in an unbounded listing, but excluding its first row with limit=1 on the default scan-and-repair path returns no data/cursor. A direct credential-free JSON-RPC reproducer, independent of this SDK, confirms default limits 2/10 and indexed limits 1/2/10 return the remaining thread. Upstream thread_processor.rs list_threads_common filters exclusions while paging the store; the exact deeper cursor cause is not proven. Do not claim the SDK fixes upstream pagination. README documents the observed case and that indexed listing skips rollout scanning/repair.

GitHub's Actions/Pages/Webhooks incident began 2026-10-19 15:49 UTC: https://www.githubstatus.com/incidents/8tzbpcnl75jp. During the first handoff, GitHub CLI and the independent connector reported required jobs queued, so protection blocked merge. On resumption, all required Quality 39285854839 and Codecov statuses were successful and the reviewed PR was mergeable. No gate or protection was bypassed. Repository auto-merge remains disabled.

The planner's proposed experimental read-state response fields, item-read/update RPCs, prediction fork mode, and environment websocket metadata must be checked against the actual stable exporter. Checked source JSON omits read-state fields from ThreadListResponse and ThreadReadResponse, so Rust source alone does not justify adding those fields. Also reject the proposed standalone `go test ./thread_lifecycle_test.go`; run package tests instead. Validate script argument contracts before using proposed commands.

## Decision Log


Use upstream rust-v0.163.0 at 83086188ae0614bd882818ff1aa23f3194962012. Choose SDK 0.163.0 because it exceeds latest immutable v0.162.1; recheck before staging. The explicit skill invocation authorizes implementation and protected publication without another approval. Use one file in the existing flat plans/ convention and temporary research/build logs under /private/tmp.

The required planner supplied workflow evidence; root independently validated paths, scripts, source schemas, branch protection, environments, and tag rules. Implement only actual stable exported compatibility needs. Preserve raw JSON precision and the existing high-level/transport boundaries. Extend existing high-level listing options for stable exclusions; do not invent experimental high-level wrappers.

Independent Go canonical extraction reproduces the approved baseline of 447 union schemas and digest 0a713f8e5cdc656a268e0a295111528b7cf7787cf8bf9e61c286f396c0bb310b. Target checked source contains 451, digest d1e367d3013d9aaa81fc25baff6c81ce52140d96abfe0155fefa4eb9947f373c, with 12 added and 8 removed entries. Both the official CLI and first source-built generation report this target digest. The source build succeeded with installed Cargo 1.94.0; no toolchain change was needed.

Approve union digest d1e367d3013d9aaa81fc25baff6c81ce52140d96abfe0155fefa4eb9947f373c after the checksum-verified official CLI exporter printed all 451 entries and every canonical entry exactly matched target source. Reviewed all 12 additions and 8 removals: ClientRequest and three ServerNotification containers, root/v2 ResponseItem and ContentItem unions, nullable root/v2 ThreadUnreadPosition references, ThreadUnreadPosition (threadStart/turn with required turnId), and ThreadReadStateOperation (read/unread). ResponseItem gains optional encrypted_content/status; ContentItem gains arbitrary annotations/logprobs, preserved by existing raw wrappers. New read-state unions use existing discriminated raw-preserving generation and required-member validation. The request container also includes excludedThreadIds, retain, and expanded realtime voice values. No generic interface fallback is approved. The bundled source-built exporter must confirm the same inventory on its next run.

Live main protection requires strict Coverage report, Go 1.26, and Go 1.27 checks, zero approving reviews, and resolved conversations. Both e2e and release allow protected branches without required reviewers. The active v* tag rules prohibit deletion and non-fast-forward updates and have no bypass actors. GitHub Actions alone creates the release tag.

The complete printed opaque inventory remains 104 entries with approved digest aa9797066f9ef641746535a494de20b16e172ee23352cb1dcd8d1bf398e4a743. No new opaque declaration or fallback was introduced. Stable generation adds read-state notifications/types, but not read-state update/item-read RPCs, read-state list/read response fields, or experimental fork prediction options.

Preserve upstream listing behavior and defaults; add no SDK-side pagination workaround. The real-CLI exclusion test covers default listing at limit=2 and indexed listing at limit=1, while replay independently verifies exact limit=1 plus exclusion forwarding. This replaces the initial overly broad acceptance claim that default upstream small-page behavior is always correct. Both real paths pass under the race detector without credentials or model calls. The existing credential-isolation and empty/full archive tests also pass. The existing Sol planner independently endorsed this amendment, and root validated its advice against the direct reproduction and tests.

## Outcomes & Retrospective


Generated protocol/RPC output now targets exact upstream 83086188ae0614bd882818ff1aa23f3194962012. Stable exclusions, retained tool-output requests, read-state notification/unions, and voice constants are covered. All local gates, both supported Go toolchains, schema coverage, release fixtures, and credential-free real-CLI tests pass. The upstream pagination limitation is reproduced and documented. QA's sole coverage request is fixed and verified; architecture found no actionable issue.

SDK v0.163.0 is published from protected-main commit 649395559fa6e36219c70e219865d4b4ce4efd66. [PR #18](https://github.com/pmenglund/codex-sdk-go/pull/18), reviewed commit 8e114032766ebd63be6965b71cd64697217c291e, passed required Go 1.26, Go 1.27, and Coverage report checks plus Codecov project/patch checks, merged through protection, and had its remote feature branch deleted. [Automatic Release 39315374167](https://github.com/pmenglund/codex-sdk-go/actions/runs/39315374167) passed candidate validation, both Quality toolchains, coverage, trusted E2E, and publication.

GitHub API and independent git ls-remote confirm annotated tag object 9ea9443fe83a99e45b46ee771166ce69124291f1 peels to exact gated merge 649395559fa6e36219c70e219865d4b4ce4efd66. The annotation identifies github-actions[bot] and publication at 2026-10-19 20:01:17 UTC. The tag's SDK version is 0.163.0 and generated upstream source is 83086188ae0614bd882818ff1aa23f3194962012. No local tag was created or moved, no direct main push or alternative publication path was used, no gate was bypassed, and no environment approval or policy change was needed. COLIN-226 is complete. This documentation closeout changes no release input and triggers no additional release.

## Context and Orientation


`gen.go` invokes `internal/codegen/main.go`, which compiles the selected upstream Codex CLI and exports its JSON schemas. A schema describes the JSON fields exchanged over the local app-server's request/response transport. Generated wire types live in `protocol/*_gen.go`; generated request calls and notification parsing live in `rpc/*_gen.go`. Fix generated code through the generator. Every exported generated declaration needs exact-name GoDoc.

`protocol/manual_types.go` contains reviewed types that the generator cannot fully express. Its schema-coverage guard rejects missing fields. Existing high-level helpers in `thread_lifecycle.go` translate user options into those protocol requests. Raw-preserving union wrappers represent JSON values with several possible shapes, retaining unknown fields and exact numbers.

`.github/codex/version` and `.github/codex/checksums.txt` pin the verified CLI. `.github/sdk-version` independently selects the SDK release. A protected main merge changing that file starts `.github/workflows/release.yml`; validation, Quality, and trusted E2E precede annotated immutable tag publication.

## Plan of Work


Milestone 1 runs the bundled updater on the feature branch with only this plan initially dirty. If the union guard stops it, print the canonical inventory and compare all added/removed entries against baseline and target source JSON under /private/tmp/codex-01630-review. Review nested definitions and request/notification containers as well as individual unions. Record typed handling or deliberately lossless raw representation before changing either approval digest. Verify the upstream checkout stays unchanged.

Milestone 2 reviews the entire generated diff and fixes handwritten types, existing option adapters, tests, and documentation. Check stable thread exclusions, read-state notifications and unread-position variants, tool-output retain, output-text annotations/logprobs, response status/encrypted content, and realtime voice changes. Confirm which experimental source fields are absent from stable output. Add meaningful wire, routing, and omission regressions, failing before fixes where feasible. Update four official platform checksums and both version files to 0.163.0. Do not add runtime dependencies.

Milestone 3 runs focused tests and the authoritative updater until deterministic generation and every local quality gate pass. Validate release/version scripts, source-schema coverage, and exact generated commit headers. Use Go 1.27.2 native runtime checks where needed for matrix parity and the supported patched Go 1.26.9 for Staticcheck. Exercise the official checksum-verified CLI with existing credential-free lifecycle tests before hosted E2E.

Milestone 4 runs QA and architecture reviews after implementation and local verification; run security review if changed code affects permission or command boundaries. Resolve actionable findings and rerun affected gates. Inspect and stage only named reviewed files. Commit with COLIN-226, push the feature branch, open a PR with compatibility and validation evidence, watch required checks, then merge through protection and delete the remote branch. Monitor automatic Release and independently compare its candidate, annotated tag object, and peeled commit. Record URLs/SHAs and close Linear after publication. Use a protected documentation closeout for final evidence if needed.

## Concrete Steps


Run from `/Users/pme/src/pmenglund/codex-sdk-go`. Reclassify every dirty path before each allow-dirty rerun. Use task-owned logs and Cargo output outside the repository.

    env -u CODEX_REPO_REF CODEX_REPO_ROOT=/Users/pme/src/openai/codex GOTOOLCHAIN=go1.26.9 CARGO_TARGET_DIR=/private/tmp/codex-01630-cargo .codex/skills/update-codex-protocol/scripts/update_codex_protocol.sh --allow-dirty

Add CODEX_PRINT_UNION_INVENTORY=1 or CODEX_PRINT_OPAQUE_INTERFACE_INVENTORY=1 when the corresponding guard fails. Compare every printed canonical entry before approving its digest.

    GOTOOLCHAIN=go1.26.9 go test ./internal/codegen ./protocol ./rpc .
    CODEX_SCHEMA_TEST_DIR=/private/tmp/codex-01630-review/target/codex-rs/app-server-protocol/schema/json GOTOOLCHAIN=go1.26.9 go test ./internal/codegen
    .github/scripts/validate-codex-metadata.sh
    .github/scripts/sdk-release-tag.sh
    .github/scripts/validate-release-tag.sh "$(.github/scripts/sdk-release-tag.sh)"
    git diff --check

The final updater must report two identical generations and success for formatting, metadata and installer fixtures, vet, all unit/race tests, Staticcheck 0.8.1, govulncheck 1.3.0, and diff hygiene. All generated headers must name 83086188ae0614bd882818ff1aa23f3194962012.

## Validation and Acceptance


Requests preserve omission/empty/populated exclusion lists, and existing high-level listing helpers send exactly the requested exclusions and page limit. Real default listing excludes the requested ID, and indexed listing also fills the requested single-row page; the reproduced upstream scan-and-repair small-page limitation remains documented. Tool output can request retention. New stable notifications parse their fields, known discriminated variants require their member keys, and unknown fields/variants retain raw JSON including 9007199254740993. Existing user flows and source-schema coverage pass. All local and required PR gates pass, followed by trusted E2E and publication. Remote v0.163.0 must be annotated and peel to the exact Release candidate on protected main.

## Idempotence and Recovery


Recheck status before mutations and preserve concurrent edits. Never stage everything, push main, create or move a local tag, bypass a failed gate, or approve environment jobs. If a newer stable upstream appears before accepted generation, revise the target and plan before continuing. If unrelated edits appear, stop at that boundary. For an unpublished failed Release, retry only the exact failed candidate or a reviewed corrective main merge, retaining the unpublished version. After publication use a higher corrective immutable version.

PR #18 is merged, its remote branch is deleted, and Release 39315374167 successfully published the verified immutable tag. Preserve v0.163.0 permanently. Any future correction must use a higher SDK version and the protected release workflow. A repeat check of the completed run, tag object, and peeled commit is read-only. Publish this plan-only documentation closeout through a protected PR without changing SDK or CLI metadata.

## Artifacts and Notes


Evidence directory: /private/tmp/codex-01630-review. Official release https://github.com/openai/codex/releases/tag/rust-v0.163.0 was published 2026-10-15. API asset SHA-256 digests: 61c4d9a92521bdb42b0c1d9f87c673f6da628a06de5945cbdd9a0010632f6d10 (Darwin arm64), 562a1aea0e5686283a467687d268ae33705f02fd4d40f1bc4110a415c5c1c7b1 (Darwin x86_64), a7261e65f4a96d8f4ca65a069c6cf2dc2c33defc341e48078acefda9461fffe4 (Linux arm64 musl), beb339e8ecd9ba057a6bb2dc81ad8ca474812edfbf0fc5056ca818797648dde4 (Linux x86_64 musl).

Authoritative success log: final-updater.log. Every generated header names the target commit. Both printed source-exported inventories exactly equal all 451 reviewed source entries; the opaque digest stays unchanged. Direct upstream-only pagination reproducer: reproduce-exclusions.py. Temporary logs/reproducers stay outside source control.

Publication evidence: PR #18 merge 649395559fa6e36219c70e219865d4b4ce4efd66; automatic push-triggered Release 39315374167; annotated tag v0.163.0 object 9ea9443fe83a99e45b46ee771166ce69124291f1; peeled commit 649395559fa6e36219c70e219865d4b4ce4efd66. The live v* ruleset remains active, rejects deletion/non-fast-forward updates, and has no bypass actors. Both e2e/release environments remain protected-branches-only without required reviewers.

## Interfaces and Dependencies


Retain the current Go packages, generator dependencies, transport interfaces, and raw-preserving union design. Rust is exporter build tooling. Determine exact exported APIs from actual stable output, not unfiltered experimental Rust declarations.

Revision note: Created after required planner review and independent validation, before generation. Rejected unsupported experimental assumptions and a non-compiling proposed test command.

Revision note: Added failing protocol regressions, implemented stable exclusion pass-through, and recorded official CLI checksums and SDK 0.163.0. Initial updater is compiling the exact upstream exporter; inventory approval remains pending.

Revision note: Reviewed and approved all union deltas after exact official exporter/source comparison. Manual struct coverage against the official exported schemas passes; the source-built updater will independently confirm the approved inventory.

Revision note: First source-built generation confirmed the exact reviewed inventory and stopped at the expected old compiled guard. All current dirty paths belong to COLIN-226: this plan, protocol regression/manual field, lifecycle adapter/tests and credential-free E2E, README, generator digest, CLI metadata, and SDK version. Rerun with both printed inventory flags; no unrelated changes are present.

Revision note: Source output and all printed inventories agree with review; two generations were identical. Corrected one new test's full-validation assumption to the established raw-wrapper contract. Regeneration is unchanged; final complete gate remains required.

Revision note: Final full gate and Go 1.27.2 checks passed with no vulnerabilities. Documented independently reproduced upstream pagination behavior, retained exact SDK wire verification, and passed both supported real listing paths under the race detector. Ready for independent QA/architecture review before staging.

Revision note: QA requested a positive regression for the documented required-key boundary. Added turnId:null to known unread-position round trips, asserting exact raw preservation and IsKnown while retaining the missing-key rejection. No production or generated code changed; affected unit/race checks will be rerun before staging.

Revision note: Review is complete. Full unit tests and affected protocol race tests on both supported toolchains pass after the QA test-only addition. Architecture reviewed the final diff and found no actionable issues. Rechecked clean scope, unchanged origin/main, available v0.163.0, release tag validation, and diff hygiene before named-path staging.

Revision note: Recorded committed/pushed PR #18, passing Codecov statuses, and the current GitHub Actions incident blocking required checks. Left this plan-only status update local to preserve the queued candidate. Protected merge, automatic trusted E2E/publication, immutable-tag verification, and final documentation closeout remain required; this update is not marked complete.

Revision note: Resumed after successful required check reporting, preserved the plan-only edit, merged the exact reviewed PR head through protection, deleted its remote branch, and located automatic Release 39315374167 for the exact merged commit. Candidate validation passed; final release gates and immutable publication remain under observation.

Revision note: Automatic Release passed every gate and published v0.163.0. Independently verified its annotated object and exact peeled commit, recorded final outcomes and protections, and completed Linear. This plan-only closeout records the finished release without changing release inputs.
