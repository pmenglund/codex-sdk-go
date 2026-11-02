# Update and publish the Codex 0.166.0 protocol

This living ExecPlan follows `PLANS.md`. Keep Progress, Surprises & Discoveries, Decision Log, and Outcomes & Retrospective current.

## Purpose / Big Picture

Callers will use stable Codex CLI 0.166.x with generated Go protocol types and RPC bindings. Preserve existing source compatibility where practical, expose newly stable fields without losing JSON data, and publish SDK v0.166.0 only after protected Quality and trusted E2E checks succeed.

## Tracker Mapping

Workflow: `WORKFLOW.md`. Project: codex-sdk. Issue: COLIN-229, https://linear.app/bothnia/issue/COLIN-229/update-codex-sdk-go-protocol-and-release-to-codex-01660. This single issue covers generation, compatibility, validation, review, protected merge, and verified publication; there are no child issues.

## Progress

- [x] (2026-11-02 17:23 UTC) Inspected instructions, clean working tree, existing completed plans, current upstream tags and SDK tags. Fast-forwarded clean main over the Dependabot-only merge to 38c43efe8465c8ae0b2cd06bea565ec8cf38a8b0. Created COLIN-229 In Progress.
- [x] (2026-11-02 17:23 UTC) Required planner returned a proposal; checked its paths and assumptions against generator, upstream split v2 sources, release scripts, toolchains, and live protections. Created the update branch before generation.
- [x] (2026-11-02 17:47 UTC) Generated from stable source and reviewed complete union/opaque inventories; source and official exports agree.
- [x] (2026-11-02 17:47 UTC) Completed compatibility, regression tests, documentation, CLI checksums and SDK version.
- [x] (2026-11-02 17:40 UTC) Source-built ordinary stable export prints exactly the same 468 canonical union entries as the fully reviewed official export. Approved c06fc3bf26e02becb5ca915a4363b72ba7957f9bbc1b496470c0f185b355733d and removed the temporary preview helper.
- [x] (2026-11-02 17:38 UTC) Provisional compatibility suites, native Go 1.27.3 vet/unit/race/vulnerability checks, all five credential-free official-CLI race groups, metadata and release-policy fixtures pass. Reviewed the complete provisional generated diff. Source confirmation remains mandatory.
- [x] (2026-11-02 17:47 UTC) Final authoritative updater exited 0 with two identical source generations and every required check passing. Native Go 1.27.3 and all five credential-free official-CLI race groups also pass.
- [x] (2026-11-02 17:49 UTC) QA, architecture and security reviews returned no actionable findings. QA/architecture independently reran affected tests successfully. Confirmed unchanged origin/main and available v0.166.0 before named-path staging.
- [ ] Commit named reviewed paths, push, open protected PR, pass required checks, and merge exact reviewed head.
- [ ] Observe automatic Release success and independently prove annotated v0.166.0 resolves to its exact gated main commit.
- [ ] Finalize this plan and COLIN-229; publish any final evidence through a protected documentation-only PR.

## Surprises & Discoveries

Main had two new commits after the initial local snapshot; the complete delta only updates three GitHub Actions pins in tests.yml. Those changes were fast-forwarded and preserved. The existing upstream checkout is clean on its original main branch; generation must leave that checkout unchanged.

Host Go is 1.27.1 and Cargo/Rust is 1.94.0. Upstream 0.166.0 requests Rust 1.95.0. The previous temporary toolchain directory exists but is empty. Linear advertises get_agent_context but the server returns Tool not found; current issue and repository context are available. GitHub connector cannot read branch protection, so authenticated gh verified the required configuration.

The source updater has started with the host compiler and is fetching fresh rusty_v8 git submodules. While it runs, checksum-verified official 0.164.0 and 0.166.0 CLIs exported ordinary stable schemas. New manual-schema coverage passes unchanged. New protocol regressions fail against old generation for dropped process drain state, hook thread, context name, native usage, review domains, GCP recognition and OTEL requirements. Item raw wrappers already preserve added lifecycle/generation fields and reject incomplete known payloads.

Git CLI fetching completed the large new V8 dependency tree and the pinned Rust exporter is compiling. A sandboxed Go invocation incorrectly reported testing/synctest as missing despite its installed source; the same command succeeds outside the sandbox. Run final Go checks with the needed build-cache/toolchain access instead of modifying tests or toolchain versions.

Reviewed the complete official union delta: 455 to 468 entries, 14 removed and 27 added. Replaced ClientRequest/ServerNotification containers include the nested changes; account adds workspace/GCP, hooks add environment ID, ThreadItem adds outer lifecycle properties and image generationId, Codex error enums add modelNotFound/policyCapacityUnavailable, and review delivery descriptions remove detached. New entries describe ModelProviderError, mixed string/object OTEL exporters and nullable OTEL/account workspace references. Dense hashes and JSON comparisons live in union-delta.json and union-review.txt. Source generation must independently match before final approval.

## Decision Log

Decision (2026-11-02): Use the established single-file plans convention and keep dense logs and exported schemas under /private/tmp/codex-01660-review. The planner is evidence; actual stable exports determine scope. No unresolved user decisions are required by this already authorized skill invocation.

Decision (2026-11-02): Target rust-v0.166.0, peeled commit 6ff1e1ca3b30fc8b914720681fcf5fe62e95f856. Use SDK 0.166.0 because it exceeds the latest immutable SDK tag v0.164.0. Recheck target stability and tag availability before publication.

Decision (2026-11-02): Try the real exporter with the available compiler. If incompatible, install checksum-verified official Rust 1.95.0 in a task-local temporary directory without modifying system tooling. Use Go 1.26.10 for the authoritative updater and Staticcheck, and Go 1.27.3 for matrix parity.

Decision (2026-11-02): The real host-compiler attempt is still fetching upstream git dependencies. Prepared official checksum-verified Rust 1.95.0 under /private/tmp/codex-01660-review/rust-toolchain to match the explicit upstream pin for the next attempt. A temporary internal/codegen/zz_review_test.go generates only a preview outside the repo from official schemas while dependency fetch runs; remove this helper before the final gate and commit. It cannot substitute for required source generation.

Decision (2026-11-02): Review OTEL's externally tagged string/object union explicitly as lossless raw JSON via the documented manual OtelExporter alias. Keep the surrounding requirements, maps, TLS paths and size fields generated and typed. ModelProviderError uses the existing discriminated-union generator. Preview generation exposed HookRunSummary.Source losing an allOf enum reference; add a failing loader regression, restore HookSource through the existing narrow loader, and retain the unchanged 104-entry opaque digest after comparing every entry.

Decision (2026-11-02): The first source attempt stalled in Cargo's default libgit2 submodule fetch and reported a SecureTransport connection error. Stopped that task-owned Cargo process and retried using Git CLI fetching plus pinned Rust 1.95.0. Account-session and GCP setup RPCs are absent from ordinary stable ClientRequest; unused schema definitions do not authorize adding experimental endpoints.

Decision (2026-11-02): Use the reviewed official stable export for provisional generated files so compatibility tests can run while source dependencies download. These artifacts are not accepted for handoff until the source-built inventory agrees and the authoritative updater regenerates twice without changes. Remove the temporary helper first. Preserve ReviewDeliveryDetached as a deprecated handwritten constant. Newly generated hook-run structs intentionally replace the old opaque notification fallbacks in this pre-1.0 minor update; document that source change. HookRunSummary itself did not previously exist as a public struct, so its native path introduces no existing typed-field assignment break.

Decision (2026-11-02): Source-built inventory matches every printed official canonical entry exactly, including hashes and JSON (468/468). Approve its c06fc3bf26e02becb5ca915a4363b72ba7957f9bbc1b496470c0f185b355733d union digest. Retain the original opaque digest; final source generation must confirm it after the hook enum repair. All current dirty paths are COLIN-229: version/checksum metadata, README, generator/loader and regression tests, manual types, generated protocol/RPC files, this plan and versioned protocol/RPC tests.

Decision (2026-11-02): The authoritative source run confirmed both exact inventories and two identical generations, byte-identical to the official provisional output. Formatting, metadata/installer fixtures, vet, unit and race tests pass. Staticcheck correctly flagged the intentional deprecated-constant compatibility fixture (SA1019); annotate that one fixture following existing api_external_test.go precedent. Do not suppress any production finding. Rerun the updater to complete the full gate.

## Outcomes & Retrospective

Implementation, local validation and independent QA/architecture/security review are complete. Stable source and official CLI exports agree on every reviewed union entry, the original opaque inventory remains unchanged, and the authoritative updater passes fully. Native Go 1.27.3, production schema coverage, real-CLI mock-provider tests, and release-policy fixtures pass. No actionable review findings remain. Protected PR merge and automatic publication remain pending; no release has been created.

## Context and Orientation

`gen.go` invokes `internal/codegen/main.go`. The generator builds upstream's CLI and exports ordinary stable JSON schemas describing messages sent over JSON-RPC. It generates checked-in Go wire types under `protocol/` and request/notification bindings under `rpc/`; headers record the exact source commit. Fix generated behavior in the generator or its schema loader, never by editing generated declarations.

`protocol/manual_types.go` contains reviewed handwritten representations checked against exported schemas. A union is a JSON value with several possible shapes. Raw-preserving wrappers retain unknown fields and exact numbers. Inventory digests force review of all changed union schemas and remaining opaque interface declarations before generation can proceed.

`.github/codex/version` and `.github/codex/checksums.txt` pin the official CLI for E2E. `.github/sdk-version` independently selects the SDK release. A protected main merge changing that file automatically starts `.github/workflows/release.yml`; GitHub Actions alone publishes an annotated immutable tag after validation, Quality and trusted E2E.

## Plan of Work

Milestone 1 runs the bundled updater on codex/update-protocol-v0.166.0 with only this task's plan initially dirty. Export checksum-verified official old/new CLI schemas for independent comparison when source inventory guards stop. Print full source inventories and review every added, removed or changed canonical entry. Inspect post-fix opaque output before deciding its digest. Acceptance is exact agreement between reviewed stable schema inventories and source-built generation.

Milestone 2 inspects the complete generated diff and fixes necessary SDK/manual/generator compatibility. Upstream source suggests item lifecycle timestamps, image generation IDs, process output-drain state, hook environment/thread/path changes, native dollar usage, account workspace/GCP/session definitions, model-provider errors, OTEL configuration and network review domains, named additional context, and removal of detached review delivery. Include only what the ordinary stable exporter exposes. Check that the flattened ThreadItem shape retains existing required-field validation and lossless fields. Assess deprecated source compatibility for ReviewDeliveryDetached and the hook path type. Add targeted failing regressions where practical, then fix generation/adapters. Update README, CLI version and all four official archive checksums, and SDK version.

Milestone 3 runs two byte-identical generations and the entire updater gate. Test external schema/manual coverage and exact generated commit headers. Run native Go 1.27.3 vet/unit/race/vulnerability checks, workflow/release fixtures, and applicable credential-free official CLI tests. Run QA and architecture reviewers only when ready to commit; security review covers changed managed network, OTEL/header/TLS, and account representations. Resolve findings and rerun affected checks. Acceptance is all required local checks and independent review findings addressed.

Milestone 4 rechecks concurrent edits and the full final diff, stages reviewed files explicitly, commits with COLIN-229, pushes the feature branch and opens a PR to main. Watch Go 1.26, Go 1.27 and Coverage report; merge only the exact reviewed head through branch protection and delete the remote feature branch. Follow the automatic Release run through validation, Quality, E2E and publication. Prove the remote annotated v0.166.0 peels to the exact gated main merge. Complete Linear and this plan, using a protected documentation-only closeout PR if evidence is available only after merge.

## Concrete Steps

Run from /Users/pme/src/pmenglund/codex-sdk-go. Never source or print .envrc. Before each allow-dirty run, classify every dirty path as COLIN-229 and stop if unrelated edits appear.

    env -u CODEX_REPO_REF CODEX_REPO_ROOT=/Users/pme/src/openai/codex GOTOOLCHAIN=go1.26.10 CARGO_TARGET_DIR=/private/tmp/codex-01640-cargo .codex/skills/update-codex-protocol/scripts/update_codex_protocol.sh --allow-dirty

Reuse the existing Cargo target cache to avoid unnecessary compilation; it is build output, not a source override. If a temporary Rust installation is necessary prepend its bin directory to PATH. On inventory failure rerun with CODEX_PRINT_UNION_INVENTORY=1 and CODEX_PRINT_OPAQUE_INTERFACE_INVENTORY=1. Review full sorted hashes and JSON before updating approvedUnionInventorySHA256 or approvedOpaqueInterfaceInventorySHA256.

    GOTOOLCHAIN=go1.26.10 go test ./internal/codegen ./protocol ./rpc .
    CODEX_SCHEMA_TEST_DIR=/private/tmp/codex-01660-review/stable GOTOOLCHAIN=go1.26.10 go test ./internal/codegen
    .github/scripts/sdk-release-tag.sh
    .github/scripts/validate-release-tag.sh "$(.github/scripts/sdk-release-tag.sh)"
    git diff --check

Expect updater completion text Generation and all quality gates passed. All generated headers must name 6ff1e1ca3b30fc8b914720681fcf5fe62e95f856. Record exact PR, head, merge, Release URL, annotated tag object and peeled commit here once known.

## Validation and Acceptance

New stable fields decode and round-trip through relevant responses and notifications; unknown union data and large integers remain lossless. Wire tests verify changed field names, optional zero/absent behavior and review delivery. Required-field failures continue to fail. Every exported API retains exact-name GoDoc. Metadata/installer fixtures, formatting, vet, all unit and race tests, Staticcheck 0.8.1, govulncheck 1.3.0, release validation and diff hygiene pass. Required PR checks and automatic Release validation/Quality/E2E/publication pass. The expected remote annotated SDK tag resolves to the exact protected main candidate.

## Idempotence and Recovery

The updater exports from a temporary detached worktree and never publishes. Repeated generation must match including untracked generated files. Preserve concurrent edits and do not stage everything. Never push main directly, create or move a local tag, bypass a failed gate, or approve pending environment jobs. Both e2e/release already allow only protected branches with no required reviewers; restore that configured policy if it unexpectedly drifts, preserving other protections.

If no generated or handwritten change is needed, remove only task-owned planning changes and unpublished local branch, returning to clean main without a PR/release. If a new stable tag appears during work, record the change before deciding scope; never mix source versions. For a failed Release, retain the unpublished SDK version; manual retry names the exact failed candidate or exact reviewed corrective main merge and reruns all release gates. If permissions or required review block progress, report the exact recovery action.

## Artifacts and Notes

Source release: https://github.com/openai/codex/releases/tag/rust-v0.166.0, published 2026-10-29. Initial protected main base: 38c43efe8465c8ae0b2cd06bea565ec8cf38a8b0. GitHub main requires strict Go 1.26, Go 1.27 and Coverage report checks with admin enforcement and zero mandatory approving reviewers. Force pushes and deletion are disabled. e2e/release allow protected branches only and have no required reviewers. Active v* tag rules forbid deletion/non-fast-forward updates without bypass actors.

## Interfaces and Dependencies

Keep existing Go dependencies, public package boundaries, transport interfaces and lossless raw union design. Rust is exporter build tooling only. Determine concrete new public types and methods from actual ordinary stable generation; record deliberate raw representations and compatibility exceptions before implementation.

Revision note: Created after current-state inspection and mandatory planner review, before generation. Records verified scope, target, tracker, assumptions, risks, validation and protected publication acceptance.

Revision note: Recorded checksum-verified official exports, full union delta review, pre-fix regression failures, temporary preview generation and task-local pinned Rust. Source generation, approval and compatibility implementation remain pending.

Revision note: Completed provisional compatibility implementation and full diff inspection. Added targeted hook/OTEL/provider-error/RPC tests, preserved detached-review source constant, documented hook fallback narrowing, updated all version/checksum metadata, and recorded successful secondary-toolchain/real-CLI/release-fixture validation. Source exporter confirmation and authoritative deterministic gate remain pending before review.

Revision note: Source-built inventory exactly matches the reviewed official export; approved only that confirmed digest and removed the temporary helper. The authoritative source updater is next, with final opaque proof and complete deterministic gate still required.

Revision note: Authoritative source generation matched twice and exactly matched the provisional files; source opaque inventory is the unchanged reviewed 104 entries. All checks through race tests passed. Added the existing narrow compatibility-fixture lint annotation for intentional deprecated-constant use; full successful updater completion remains pending.

Revision note: Final authoritative updater exited 0 with deterministic generation and all required checks passing. Marked implementation/local validation complete and prepared independent reviews before staging.

Revision note: All three independent reviews returned no actionable findings; recorded their scope and independent affected-test runs. Reconfirmed unchanged remote main, available release tag, clean diff hygiene and the exact task-owned paths. Corrected the final-gate timestamp from an estimate to observed UTC time.
