# Update and publish the Codex 0.167.2 protocol

This living ExecPlan follows `PLANS.md`. Keep Progress, Surprises & Discoveries, Decision Log, and Outcomes & Retrospective current.

## Purpose / Big Picture


Callers will use stable Codex CLI 0.167.2 with matching Go protocol types and RPC bindings. Uploaded audio references, thread section changes, OTEL delta logging, and human approval requests must retain their wire data. Automatic approval handlers must decline requests that upstream requires a person to decide. Publish SDK v0.167.2 only after the protected Quality and trusted E2E checks succeed.

## Tracker Mapping


Workflow: `WORKFLOW.md`. Project: codex-sdk. Issue: COLIN-230, https://linear.app/bothnia/issue/COLIN-230/update-codex-sdk-go-protocol-and-release-to-codex-01672. This single issue covers generation, compatibility, tests, review, protected merge, and publication; there are no child issues.

## Progress


- [x] (2026-11-11 23:35 UTC) Verified clean main and public remote at 21a5e539e703e5ad59b7b61ca033b96880956f1d, latest stable upstream rust-v0.167.2, SDK tags, instructions, and completed prior plans. Created COLIN-230 In Progress.
- [x] (2026-11-11 23:35 UTC) Reviewed the required planner proposal against upstream source, official stable schemas, local generator and approval handlers. Verified live release protections; created codex/update-protocol-v0.167.2 before adding this plan or generating.
- [x] (2026-11-11 23:35 UTC) Installed checksum-verified Rust 1.95.0 in the task temporary directory and independently exported checksum-verified official CLI stable schemas.
- [x] (2026-11-11 23:42 UTC) First authoritative source build completed and stopped at the expected union guard: 471 entries, digest 60f77fa7958aae5e9426404d2d9c0ff93dc66cf3a4b4f54eeee7dbb1c06d8826, matching the independently reviewed official export. Full source inventory print and opaque review remain pending.
- [x] (2026-11-11 23:42 UTC) New protocol tests fail against old generation for uploaded audio and dropped capability/OTEL fields. Sol implemented and tested human-only refusal; generator enum regression exposed scalar-union type loss and now passes after normalization. Metadata and draft compatibility notes updated.
- [x] (2026-11-11 23:53 UTC) Full source union print matches all 471 reviewed official entries; approved union digest. Compared old/new opaque output with a temporary diagnostic that stops before writes. Restricted normalization to the approval enum; all 104 opaque declarations and their approved digest remain identical.
- [x] (2026-11-11 23:54 UTC) Generated and inspected the complete source-export diff. Finished approval refusal, typed enum normalization, versioned protocol/RPC regressions, README and all CLI/SDK metadata. Compatibility suites and five credential-free official-CLI race groups pass. Removed the temporary helper before starting the final updater.
- [x] (2026-11-12 00:00 UTC) Updated Go minimum/matrix/docs; native Go 1.27.4 vet/unit/race/vulnerability checks and Go 1.26.11 production-schema/real-CLI checks pass. No vulnerabilities remain.
- [x] (2026-11-12 00:00 UTC) Final authoritative updater exited 0 on Go 1.26.11: two identical source generations, unchanged 104-entry opaque inventory, formatting, metadata/installer fixtures, vet, all unit/race tests, Staticcheck, govulncheck and diff hygiene pass. Release/version/content-manifest fixtures also pass.
- [x] (2026-11-12 00:05 UTC) Independent security, QA and architecture reviews found no actionable issues. QA reran all four affected packages; security reran focused approval/enum/dispatch tests; those checks and diff hygiene pass. No review-driven code changes were needed.
- [ ] Commit reviewed named paths, publish the feature branch and protected PR, and pass required checks.
- [ ] Merge the exact reviewed head, monitor automatic Release through publication, verify the immutable remote tag, and finish Linear and this plan.

## Surprises & Discoveries


Sandboxed gh cannot access the login keychain and reports an invalid token. With normal keychain access, gh authenticates successfully with repo/workflow permissions. SSH authentication fails; public HTTPS fetching works. Use authenticated HTTPS for branch publication without changing the user's remote configuration. The GitHub connector cannot read branch protection, so gh verified the current settings.

Stable source and the official export add audio alternatives, requiredHumanApproval, humanCommand, log_sse_deltas and thread/section/changed. The source's thread/guardianReview RPC is experimental and is absent from the ordinary stable request schema. High-level initialization currently advertises no capabilities. AutoApproveHandler currently accepts every command approval kind, which conflicts with upstream's requirement for a fresh human decision on humanCommand.

The first source build finished in 4m51s. Both official old/new exports were checksum verified. The old canonical inventory matches the approved c06fc3bf26e02becb5ca915a4363b72ba7957f9bbc1b496470c0f185b355733d exactly. Reviewed all nine removed and twelve added union nodes: audio parent/alternative nodes and their enclosing request/notification schemas, three versions of the new section notification envelope, and the scalar approval kind union. Dense hashes and structural JSON differences are in union-delta.json and union-review.txt under the temporary directory.

The first source-built CLI exported source-stable schemas that are byte-for-byte identical to the official stable directory. Its sorted 471-entry canonical inventory also matches every official hash and JSON entry exactly. Go 1.27.3 is available. All release-tag/version-transition and content-manifest fixture scripts pass.

The first post-approval generation stopped at the opaque guard: generic scalar enum normalization would turn nine existing opaque aliases into typed enums. A temporary inventory-only test proved that the same change occurs with old schemas, so it came from normalization rather than upstream. Restricting normalization to CommandExecutionApprovalKind restores exactly the prior 104-entry inventory and aa9797066f9ef641746535a494de20b16e172ee23352cb1dcd8d1bf398e4a743 digest.

The complete generated diff contains only selected commit/version metadata, typed human approval/capability and OTEL additions, uploaded audio alternatives, and section notification types/parsing. No experimental Guardian RPC appears. Existing protocol/unions_test.go expected the old missing audio_url error; it now asserts the required-reference alternative. Source/manual coverage and all affected suites pass. All five real-CLI groups pass with credentials cleared and a local mock provider under the race detector.

Go 1.27.3 passed vet, unit and race tests but govulncheck reported GO-2026-6859 in net/http. The official Go vulnerability database, published 2026-11-11, identifies patched versions 1.26.11 and 1.27.4. Trusted E2E reads go.mod for its compiler, so changing only the Quality matrix would leave that release check on the old minimum.

## Decision Log


Decision (2026-11-11): Use the established single-file plan convention and keep exported schemas, dense inventories, builds and logs under /private/tmp/codex-01672-review. The planner provides evidence; actual stable generation determines the final change. No user decision is unresolved by this explicitly authorized skill invocation.

Decision (2026-11-11): Target rust-v0.167.2 at ce21072c90d32c40af7a55b60b2fb35c84a43a6a. Select SDK 0.167.2 because it exceeds the latest immutable SDK v0.166.0. Recheck tag availability before publication and record any newer upstream tag without silently changing this update's source.

Decision (2026-11-11): Match upstream's pinned Rust 1.95.0 using the checksum-verified temporary installation. Use Go 1.26.10 for authoritative generation, tests and Staticcheck; use Go 1.27.3 for CI parity. Reuse the existing upstream Cargo target cache and Git CLI dependency fetching. Preserve the upstream checkout branch and files.

Decision (2026-11-11): Preserve CommandExecutionApprovalKind as a typed string and keep the opaque digest unchanged unless the actual reviewed output justifies otherwise. Normalize homogeneous string-enum alternatives in the generator only if the new scalar union requires it. Preserve inline audio and add lossless uploaded audio through existing raw-preserving wrappers. Do not add experimental Guardian endpoints or make the high-level client advertise human approval capability.

Decision (2026-11-11): Escalate the approval-handler change to Sol as required by the repository's model routing for authorization boundaries. Automatic handlers will return a one-time decline for humanCommand and preserve ordinary command/stdin behavior. Clients opting into RequiredHumanApproval need a custom callback that consults a person. Final security review is mandatory for this behavior.

Decision (2026-11-11): Implement only disjoint string-enum oneOf normalization, preserving typed constants without widening overlapping or constrained alternatives. Retain high-level audio constructors unchanged and expose uploaded audio through the existing low-level raw-preserving protocol wrappers; no new convenience API is necessary for this protocol update. All current dirty files are COLIN-230: metadata, README, approvals/tests, generator/tests, versioned protocol tests and this plan.

Decision (2026-11-11): The explicit source generation inventory print matches all 471 canonical hashes and JSON entries from the fully reviewed official export. Approve only 60f77fa7958aae5e9426404d2d9c0ff93dc66cf3a4b4f54eeee7dbb1c06d8826. Keep the existing opaque digest pending confirmation; do not expand it to accommodate loss of the typed approval enum.

Decision (2026-11-11): Invoke the safe scalar-enum normalizer only for the named CommandExecutionApprovalKind schema. Preserve existing opaque aliases rather than introducing unrelated source changes. Add a generator regression for that restriction. Use the temporary helper once to generate from the already verified source-stable export for inspection and compatibility tests while the final updater rebuilds its temporary CLI; remove the helper before that authoritative run. The full updater must independently reproduce those files twice and pass every check before review or handoff.

Decision (2026-11-11): The new standard-library advisory supersedes the original toolchain assumptions. After a read-only Sol assessment and verification at https://vuln.go.dev/ID/GO-2026-6859.json, raise go.mod's minimum to 1.26.11, pin Quality to 1.26.11/1.27.4 and Staticcheck to go1.26.11, and update README, APP.md and WORKFLOW.md. Keep Go series and stable check names unchanged; no toolchain directive or dependency change is needed. This is a required failed-check fix under COLIN-230. The previous updater run cannot qualify the patched baseline; rerun it with go1.26.11 and repeat native 1.27.4 checks.

## Outcomes & Retrospective


Implementation and all local validation are complete. Source and official exports agree exactly; the reviewed enum and automatic approval behavior retain the intended public types and human-decision requirement. The authoritative Go 1.26.11 updater proved deterministic source generation and passed every required check. Native Go 1.27.4 checks, production-schema coverage, release-policy fixtures and credential-free official-CLI race tests pass. Independent security, QA and architecture reviews found no actionable issues. Protected publication remains pending. No release has been published yet.

## Context and Orientation


`gen.go` invokes `internal/codegen/main.go`, which builds upstream's CLI and exports ordinary stable JSON schemas. A schema describes wire messages sent over JSON-RPC. Generation writes checked-in Go types under `protocol/` and request/notification bindings under `rpc/`; generated headers identify the exact source commit. Repair generated behavior in the generator rather than editing output.

A union permits several JSON shapes. `protocol/unions_gen.go` wraps discriminated unions while retaining unknown fields and exact numbers. `protocol/manual_types.go` holds reviewed handwritten declarations checked against exported schemas. Union and opaque-interface inventory digests require explicit review before accepting new schemas or untyped output. `approvals.go` implements automatic and rejecting approval handlers; the unsafe logging handler delegates its command decision to AutoApproveHandler.

`.github/codex/version` and `.github/codex/checksums.txt` select the verified official CLI for E2E. `.github/sdk-version` selects the SDK tag independently. A protected main merge changing that file starts `.github/workflows/release.yml`; GitHub Actions alone publishes an annotated immutable tag after validation, Quality and trusted E2E.

## Plan of Work


Milestone 1 runs the bundled updater on the feature branch with only task-owned files dirty. When inventory guards stop generation, rerun with CODEX_PRINT_UNION_INVENTORY=1 and review every added, removed or changed sorted canonical schema, including nested audio alternatives, the humanCommand enum, notification envelope and required nullable sectionId. Compare the source-built inventory with the independently exported official stable schemas. Review opaque declarations before approving either digest. Acceptance is exact agreement with reviewed ordinary stable schemas.

Milestone 2 fixes any required schema normalization in `internal/codegen/main.go` and adds a failing regression in `internal/codegen/main_test.go`. Preserve the existing typed command approval enum, its old constants and a new typed HumanCommand constant. Regenerate protocol and RPC output. Add `protocol/protocol_0167_test.go` for inline/uploaded audio, missing alternatives, raw data, exact capability/OTEL/approval wire names and section IDs; add `rpc/protocol_0167_test.go` for section notification parsing and humanCommand dispatch. The Sol escalation owns only `approvals.go` and `approvals_test.go`, including failing tests for both automatic handler variants. Update README compatibility notes, all four official CLI archive checksums, CLI version and SDK version. Add high-level audio input fields only if needed for the intended stable compatibility; preserve existing constructors and validation.

Milestone 3 first updates the Go security baseline in go.mod, .github/workflows/tests.yml, README.md, APP.md and WORKFLOW.md, then reruns the authoritative updater until source generation is byte-identical twice and every local check passes. Verify generated commit headers, production-schema/manual coverage, native Go 1.27.4 checks, release fixtures, and applicable credential-free official-CLI tests. Once the work is otherwise ready to commit, run security, QA and architecture reviewers. Address actionable findings and rerun affected checks. Acceptance requires all required checks and no unresolved review finding.

Milestone 4 rechecks concurrent edits, full diff and whitespace. Stage only reviewed paths explicitly, commit with COLIN-230, publish the feature branch and open a PR to main. Watch Go 1.26, Go 1.27 and Coverage report; merge only the exact reviewed head through protection and delete the remote branch. Follow the automatic Release run through validation, Quality, E2E and publication. Independently confirm annotated v0.167.2 peels to the exact gated main merge. Finish Linear and this plan through a protected documentation-only closeout PR if final evidence arrives after merge.

## Concrete Steps


Run from /Users/pme/src/pmenglund/codex-sdk-go. Never source or print .envrc. Before every allow-dirty run, inspect status and classify every dirty path as COLIN-230; preserve unrelated changes and stop at the skill's stated boundary if they appear. The upstream SSH remote needs a command-local HTTPS rewrite; do not persist a remote change.

    env -u CODEX_REPO_REF PATH=/private/tmp/codex-01672-review/rust-toolchain/bin:$PATH CODEX_REPO_ROOT=/Users/pme/src/openai/codex CARGO_TARGET_DIR=/Users/pme/src/openai/codex/codex-rs/target CARGO_NET_GIT_FETCH_WITH_CLI=true GOTOOLCHAIN=go1.26.11 GIT_CONFIG_COUNT=1 GIT_CONFIG_KEY_0=url.https://github.com/.insteadOf GIT_CONFIG_VALUE_0=org-14957082@github.com: .codex/skills/update-codex-protocol/scripts/update_codex_protocol.sh --allow-dirty

Use normal cache/toolchain access outside the sandbox as required. On inventory failure rerun the same command with CODEX_PRINT_UNION_INVENTORY=1 and CODEX_PRINT_OPAQUE_INTERFACE_INVENTORY=1, saving full output under the temporary review directory. Review canonical JSON and per-schema hashes before changing approvedUnionInventorySHA256 or approvedOpaqueInterfaceInventorySHA256.

    GOTOOLCHAIN=go1.26.11 go test ./internal/codegen ./protocol ./rpc .
    CODEX_SCHEMA_TEST_DIR=/private/tmp/codex-01672-review/stable GOTOOLCHAIN=go1.26.11 go test ./internal/codegen
    .github/scripts/sdk-release-tag.sh
    .github/scripts/validate-release-tag.sh "$(.github/scripts/sdk-release-tag.sh)"
    git diff --check

Expect the release scripts to accept v0.167.2, and the updater to print Generation and all quality gates passed. Generated headers must name ce21072c90d32c40af7a55b60b2fb35c84a43a6a. Record exact reviewed head, PR, merge, Release URL, tag object and peeled commit after observation.

## Validation and Acceptance


Inline audio remains compatible; uploaded audio decodes and round-trips with future fields intact. Known audio payloads missing both alternatives fail validation. HumanCommand arrives unchanged at callbacks and automatic handlers decline it without leaking request bodies. High-level initialization does not opt into the human capability. Section notifications preserve string and null IDs, and explicit false OTEL settings survive round-trip. Every exported declaration has exact-name GoDoc. The complete updater checks formatting, metadata/installer fixtures, vet, unit/race tests, Staticcheck 0.8.1, govulncheck 1.3.0 and diff hygiene. Native second-toolchain and relevant release-policy checks pass. Required PR checks and automatic Release validation/Quality/E2E/publication succeed; the immutable annotated SDK tag points at the exact gated protected-main merge.

## Idempotence and Recovery


The updater exports from temporary detached worktrees and never stages or publishes. Repeated generation must match including new untracked files. Preserve concurrent edits and upstream source state. Never hand-edit generated output, push main directly, create or move local tags, bypass failed checks, or approve pending environment jobs. Restore only the previously verified protected-branches-only/no-reviewer environment policy if it drifts, preserving other protections.

For a failed Release, retain unpublished SDK 0.167.2 and retry only the exact failed candidate or exact reviewed corrective merge through the authorized workflow. If no generated or handwritten change is needed, remove only this task's unpublished plan/branch and return to clean main. If permissions, required review, unrelated edits or an ambiguous version block handoff, report the exact recovery action. A later correction after publication requires a higher immutable SDK version.

## Artifacts and Notes


Source release: https://github.com/openai/codex/releases/tag/rust-v0.167.2, published 2026-11-11 22:44:46 UTC. Main base: 21a5e539e703e5ad59b7b61ca033b96880956f1d. Live main protection requires strict Go 1.26, Go 1.27 and Coverage report checks with admin enforcement and zero mandatory reviews; force pushes and deletion are disabled. e2e/release permit protected branches only without reviewers. Active v* tag rules forbid deletion and non-fast-forward changes without bypass actors.

Temporary evidence directory: /private/tmp/codex-01672-review. Official aarch64 Darwin CLI archive SHA256: 7511163b7a0152bcf0e1229092bd2545fdccff64fc8d92ad7c6e01f0ca4681df. Rust 1.95.0 aarch64 Darwin archive SHA256: 0382ac0e7acab0daa24710701442e9fd284148b75265bdb2f59e38b60b03ac3c. Dense command logs and schema comparisons stay outside the final repository diff.

## Interfaces and Dependencies


Keep current Go dependencies, public package boundaries, transport interfaces and lossless union design. Rust is build tooling for schema export only. Determine concrete generated APIs from ordinary stable export. Preserve CommandExecutionApprovalKind's string type and existing constants. RequiredHumanApproval is available to low-level clients; custom callbacks retain responsibility for consulting a person. Do not add the experimental thread/guardianReview endpoint.

Revision note: Created after clean-state inspection and mandatory planner review, before implementation or generation. Records the verified stable target, tracker, toolchains, live protections, compatibility assumptions, approval escalation and release acceptance criteria.

Revision note: Recorded source digest agreement, complete official union delta review, expected pre-fix regression failures, the Sol approval fix and enum normalization. Source inventory printing, digest approval, generated output and final validation remain pending.

Revision note: All 471 printed source inventory entries match the reviewed official inventory exactly. Approved the confirmed union digest; next generation must confirm the existing opaque inventory and complete the local checks.

Revision note: Diagnosed and avoided nine unrelated enum API changes; the final approval-only normalization preserves the exact 104-entry opaque inventory. Temporary source-export generation will permit diff inspection and tests during the final authoritative rebuild; its helper must be removed before handoff.

Revision note: Generated the reviewed source-built export, removed the temporary helper, inspected the complete output, and passed affected compatibility/production-schema and credential-free real-CLI tests. The final updater is independently rebuilding and must still prove deterministic output and the complete quality gate.

Revision note: New GO-2026-6859 blocks the old Go baseline. Recorded the official patched ranges, Sol assessment, required minimum/CI/doc changes and new validation commands. Preserved historical toolchain evidence and corrected one estimated progress timestamp to its observed time.

Revision note: Updated Go minimum/CI/doc pins, passed all native Go 1.27.4 checks without vulnerabilities, and repeated production-schema and credential-free CLI tests on Go 1.26.11. The old-baseline updater proved deterministic source generation but correctly failed vulnerability scanning; final patched-baseline qualification remains in progress.

Revision note: Final patched-baseline updater exited 0 with deterministic source generation and every required check passing. Marked local implementation/validation complete and prepared independent reviews before any staging.

Revision note: All three independent reviews found no actionable findings; QA and security independently passed focused checks. Reconfirmed unchanged protected-main base, available SDK tag and required branch/environment protections before named-path staging.
