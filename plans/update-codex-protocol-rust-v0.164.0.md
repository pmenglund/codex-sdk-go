# Update Codex protocol and release SDK 0.164.0

This living ExecPlan follows `PLANS.md`. Maintain Progress, discoveries, decisions, and outcomes throughout the update.

## Purpose / Big Picture


Update the SDK from Codex 0.163.0 to stable 0.164.0 so callers can use the current stable app-server requests and decode its responses without losing fields or numeric precision. Publish SDK v0.164.0 only after deterministic generation, local checks, independent review, protected pull-request checks, and trusted end-to-end tests succeed.

## Tracker Mapping


Workflow: `WORKFLOW.md`. Linear project: P-COLIN-6 (codex-sdk). COLIN-228 owns every milestone in this plan. The completed 0.163.0 and API documentation plans remain historical.

## Progress


- [x] (2026-10-25 21:15 UTC) Inspected clean checkout, instructions, completed plans, upstream tags, immutable SDK versions, release protections, and required planner recommendation. Verified stable rust-v0.164.0 at 6aaac694803004a63c6f393d153c96d22521a6b9 and latest SDK v0.163.0.
- [x] (2026-10-25 21:15 UTC) Created COLIN-228. Removed only the verified stale worktree registration, preserved its directory, switched to clean main equal to origin/main at d380a4706990d793349c9f62621ac857907c2a05, and created codex/update-protocol-v0.164.0 before generation.
- [x] (2026-10-25 21:25 UTC) Source-built exporter under verified Rust 1.99.1 exactly matched all 455 reviewed official canonical union entries; approved e648ec95284956a8bff9f105b8f9beed1d701f3e98766b1ebee682f03aacd4d9. Opaque inventory review remains part of regeneration below.
- [x] (2026-10-25 21:36 UTC) Reviewed all generated changes and exact commit headers, retained all 104 baseline opaque declarations, passed new protocol/RPC regressions, and updated official CLI metadata, SDK 0.164.0, and README.
- [x] (2026-10-25 21:42 UTC) Final authoritative patched-updater.log passes two identical generations, matching reviewed 455-union and 104-opaque inventories, formatting, metadata/installer fixtures, vet, all unit/race tests, Staticcheck 0.8.1, govulncheck 1.3.0 (no vulnerabilities), and diff hygiene. Native Go 1.27.3 and official CLI race checks, actionlint, and release-policy fixtures also pass.
- [x] (2026-10-25 21:47 UTC) Completed independent QA, architecture, and security review. Architecture/security found no actionable issues. Added QA-requested nonempty managed-hook key round trips and sci/sci_plus constant regressions; QA confirmed resolution. Full unit suites and protocol race checks pass on both supported Go patches. Named-path staging remains the next handoff action.
- [ ] Commit, push, open protected PR, pass checks, merge, and delete the remote feature branch.
- [ ] Verify automatic Release validation, Quality, E2E, publication, and immutable tag at the exact merged commit; record evidence and complete COLIN-228.

## Surprises & Discoveries


The original checkout was clean but on the completed release-record branch. Local main already equaled origin/main and was registered in an expired temporary worktree without a .git file. Pruning its registration left its remaining directories untouched. Current main includes public API documentation and generator comment tests that this update must preserve.

The host has Go 1.27.1 and Homebrew Rust/Cargo 1.94.0 without rustup. Upstream requests Rust 1.99.1. Try the real exporter before deciding whether temporary tooling is required. Use patched Go 1.26.10 for the authoritative updater and Staticcheck, and Go 1.27.3 for native matrix parity. Linear's advertised get_agent_context tool is unavailable; issue and repository context are available.

The actual Rust 1.94 exporter build failed with E0658 for atomic_try_update in codex-diagnostics. Downloaded upstream Rust 1.99.1 from static.rust-lang.org, verified SHA-256 30976378b6aaefd1861ba475905c6ba22a44c981418ce52285006af26337f98c, and installed only rustc/cargo/standard library under /private/tmp/codex-01640-review/rust-toolchain. The system toolchain is unchanged. The source exporter is rebuilding with the pinned compiler.

Checksum-verified official 0.163.0/0.164.0 CLI exports give 451/455 union schemas. Reviewed all six replaced and ten added canonical entries in union-review.txt: three ClientRequest containers add feedback/save; ServerNotification changes nested hook/plan enums and image definitions; two ThreadItem representations accept fileId/result and inputImages; two added image unions and two nullable storage references complete the delta. No thread sampling, realtime voice-session ID, or metadata-patch field appears in the stable output. Manual schema coverage passes unchanged.

New protocol regressions fail against the old generation for file-only image results, omitted storage paths, and dropped MCP tool-list metadata. The generator now uses its dependency's supported schema-loader extension to represent only McpServerStatus.toolListMeta as json.RawMessage; a focused generator regression passes for both direct and nested schemas and retains other existing opaque representations.

The first complete generated run passed two identical generations, formatting, metadata and installer fixtures, vet, unit tests, race tests, and Staticcheck, then govulncheck correctly stopped on GO-2026-6809 in crypto/x509. The Go vulnerability database (https://pkg.go.dev/vuln/GO-2026-6809) confirms fixes in 1.26.10 and 1.27.3. Updated go.mod, CI pins, README, APP.md, and WORKFLOW.md to those patches under COLIN-228. The hosted release must use the fixed patches too. Original failed-gate evidence remains in validated-updater.log. The patched 1.26.10 vulnerability check and native 1.27.3 vet/unit/race/vulnerability checks pass, as do actionlint and release-policy fixtures. All five credential-free real-CLI groups pass under the race detector, including feedback ZIP creation/collision handling, credential isolation, empty-thread archive, exclusions, and mock-provider archive round trips.

## Decision Log


Decision (2026-10-25): Use the existing single-file plans convention and COLIN-228, after validating the required planner proposal against the actual checkout. Keep command transcripts, source comparisons, and build output under /private/tmp/codex-01640-review, outside the final repository diff.

Decision (2026-10-25): Target the highest exact stable rust-vMAJOR.MINOR.PATCH tag, rust-v0.164.0. SDK 0.164.0 is the default candidate because it exceeds the highest existing immutable SDK tag, v0.163.0. Confirm again before handoff.

Decision (2026-10-25): Assess the ordinary stable exporter, rather than assuming every Rust declaration is exported. Expected areas include feedback/save, application storage settings, MCP tool-list metadata, new hook events, and file-backed image results. Thread sampling and metadata patches remain out of scope unless the normal exporter exposes them. Preserve arbitrary JSON through raw wrappers, including integers above 2^53.

Decision (2026-10-25): Use the existing raw ThreadItem wrapper and nested required-alternative handling for image results. Keep storage paths typed and toolListMeta deliberately raw to preserve arbitrary JSON without float64 conversion. A narrow schema loader avoids duplicating the whole MCPServerStatus declaration or editing generated fields. Approve the union digest only after the source-built inventory independently matches the reviewed official export; inspect the remaining opaque inventory before deciding its digest.

Decision (2026-10-25): Approved the 455-entry union digest after comparing every printed canonical entry from source-inventory-updater.log with the reviewed official CLI export; both sets are identical. Added protocol/RPC regressions for image alternatives, native feedback path spelling, explicit includeLogs=false, notification routing, and lossless metadata. Added a credential-free real-CLI feedback test that verifies a readable ZIP, existing-path rejection, and zero model calls.

Decision (2026-10-25): Rejected the printed opaque FeedbackSaveParams.SavePath field. Sanitization strips its allOf reference; restore LegacyAppPathString through the narrow loader and test that actual sanitization/generation path. The subsequent source generation prints exactly the original 104 opaque declarations, digest aa9797066f9ef641746535a494de20b16e172ee23352cb1dcd8d1bf398e4a743. Compared every line with the checked-in baseline and retained the original approval. An intermediate predicted 103-entry digest was rejected by the guard: fixing the path also restores existing fallback declaration choices, so inventory changes cannot be predicted by simply subtracting one line. No generated files from that rejected attempt were accepted.

Decision (2026-10-25): Raise the minimum supported Go patch to 1.26.10 and CI's second toolchain to 1.27.3 to fix the newly reported standard-library vulnerability. Preserve required check names Go 1.26/Go 1.27, the pinned analyzers, and all protections. Rerun the entire updater under 1.26.10 and native 1.27.3 checks. Add security review for this necessary toolchain change. Do not suppress the finding or publish with vulnerable tooling.

## Outcomes & Retrospective


Local implementation, validation, and independent review are complete: stable 0.164.0 APIs are generated from the exact source commit, raw metadata and native path types are preserved, and the new Go standard-library advisory is fixed in supported CI toolchains. QA's test-coverage finding is resolved; architecture and security found no actionable issues. Protected merge and publication remain unfinished. No release success is claimed until the remote immutable tag is independently verified.

## Context and Orientation


`gen.go` invokes `internal/codegen/main.go`, which builds the selected upstream CLI and exports JSON schemas describing the app-server wire format. It writes Go wire types under `protocol/*_gen.go` and request/notification stubs under `rpc/*_gen.go`. Generated headers identify the exact upstream commit. Fix generated behavior in the generator, never in generated declarations.

`protocol/manual_types.go` holds reviewed handwritten representations. The generator checks their fields against exported schemas. A union is a JSON value with several possible shapes; reviewed raw-preserving wrappers retain unknown fields and exact numbers. Inventory digests force deliberate review of changed unions and remaining opaque interface declarations.

`.github/codex/version` and `.github/codex/checksums.txt` pin the verified official CLI. `.github/sdk-version` independently sets the SDK release. A protected main merge changing that file starts `.github/workflows/release.yml`. GitHub Actions alone creates the annotated immutable tag after validation, Quality, and trusted E2E.

## Plan of Work


Milestone 1 runs the bundled updater with only this plan initially dirty. Use the existing task-safe Cargo cache to avoid unnecessary rebuilds. If the exporter fails, diagnose the actual error without weakening upstream or bypassing source generation. If a union guard fails, rerun with its print flag and compare every added, removed, or changed canonical JSON entry with the previous stable export. Review opaque inventory changes similarly. Record typed or deliberately lossless raw handling before updating either digest.

Milestone 2 inspects the complete generated diff and updates only required handwritten adapters, protocol/RPC tests, README, CLI metadata, and SDK version. Add regressions for newly stable requests and response fields, raw union round trips, and routing where relevant. Verify failing behavior before production fixes when feasible. Preserve exact path spellings and opaque JSON, including 9007199254740993. Do not add runtime dependencies or unrelated features.

Milestone 3 reruns the authoritative updater until two generations match and all checks pass. Validate source-schema coverage, exact generated commit headers, official CLI checksums, release scripts, and the Go 1.27.3 matrix checks. Then run QA and architecture reviews; add security review if compatibility code changes permissions, execution, or file-write boundaries. Address actionable findings and rerun affected checks.

Milestone 4 rechecks concurrent changes and the full diff, stages reviewed paths by name, commits with COLIN-228, pushes, and opens a pull request to main. Watch required checks and merge only the exact reviewed head through protection. Delete the remote branch. Monitor the automatic Release and independently verify the expected annotated tag peels to its gated main commit. Complete Linear and the plan after publication. If final repository evidence needs a documentation closeout, use a protected documentation-only PR.

## Concrete Steps


Run from `/Users/pme/src/pmenglund/codex-sdk-go`. Before every allow-dirty run, classify every dirty path as belonging to COLIN-228. Never source or print .envrc.

    env -u CODEX_REPO_REF PATH=/private/tmp/codex-01640-review/rust-toolchain/bin:$PATH CODEX_REPO_ROOT=/Users/pme/src/openai/codex GOTOOLCHAIN=go1.26.10 CARGO_TARGET_DIR=/private/tmp/codex-01640-cargo .codex/skills/update-codex-protocol/scripts/update_codex_protocol.sh --allow-dirty

If an inventory guard reports a change, rerun the same generation with CODEX_PRINT_UNION_INVENTORY=1 or CODEX_PRINT_OPAQUE_INTERFACE_INVENTORY=1, review its complete canonical entries, and only then update the corresponding approved digest in internal/codegen/main.go.

    GOTOOLCHAIN=go1.26.10 go test ./internal/codegen ./protocol ./rpc .
    CODEX_SCHEMA_TEST_DIR=/private/tmp/codex-01640-review/stable GOTOOLCHAIN=go1.26.10 go test ./internal/codegen
    .github/scripts/sdk-release-tag.sh
    .github/scripts/validate-release-tag.sh "$(.github/scripts/sdk-release-tag.sh)"
    git diff --check

The updater must end with “Generation and all quality gates passed.” Run native Go 1.27.3 vet, unit/race tests, and govulncheck where needed for matrix parity. Both release version files should report 0.164.0; every generated header must identify 6aaac694803004a63c6f393d153c96d22521a6b9.

## Validation and Acceptance


New stable requests send the exact method and JSON fields, responses and notifications decode their added fields, and unknown union data remains lossless. All exported APIs retain exact-name GoDoc. Two generations match. Formatting, metadata/installer fixtures, vet, all unit and race tests, Staticcheck 0.8.1, govulncheck 1.3.0, release validation, and diff hygiene pass. Required protected PR checks pass, followed by automatic Release validation, Quality, trusted E2E, and publication. Remote v0.164.0 must be annotated and peel to the exact Release candidate on protected main.

## Idempotence and Recovery


Recheck status before mutations and preserve concurrent edits. Never stage everything, push main directly, create or move a local tag, bypass a failed gate, or approve environment jobs. Live e2e/release environments allow only protected branches and have no required reviewers; restore that configured policy if approval unexpectedly blocks the authorized workflow while preserving other protections. An unchanged failed Release retry must name the exact failed candidate; a corrective retry must name the reviewed corrective main merge and retain the unpublished SDK version.

If generation and required handwritten work are unchanged, remove only this task-owned plan, return to clean main, and delete only this unpublished local branch, leaving no PR or release. If unrelated dirty files appear, stop at the skill's boundary and report their exact paths. The shared upstream checkout remains on its existing branch; the updater uses a temporary detached worktree.

## Artifacts and Notes


Source: https://github.com/openai/codex/releases/tag/rust-v0.164.0, published 2026-10-23. Exact peeled source commit: 6aaac694803004a63c6f393d153c96d22521a6b9. Evidence and scratch output live under /private/tmp/codex-01640-review. Live main requires Go 1.26, Go 1.27, and Coverage report with strict/admin enforcement, zero mandatory reviewers, and no force pushes or deletion. Active v* tag rules prohibit deletion/non-fast-forward updates without bypass actors.

## Interfaces and Dependencies


Keep the existing Go module dependencies, public package boundaries, transport interfaces, and raw union design. Rust is exporter build tooling. Determine concrete new types and methods from actual stable generation, and record any manual representation before implementation.

Revision note: Created after repository inspection and required planner review, before generation. Recorded the exact target, branch preparation, toolchain uncertainty, stable-export boundary, tracker, and protected publication acceptance.

Revision note: Recorded the actual compiler failure, verified temporary toolchain, complete official union review, stable-export exclusions, failing compatibility regressions, narrow raw-metadata generation decision, and official CLI/SDK metadata updates. Source inventory approval and complete validation remain pending.

Revision note: Source-built inventory matches the reviewed official export exactly; approved the digest and restarted the authoritative updater with both inventory flags. Every current dirty path belongs to COLIN-228: this plan, version/checksum metadata, README, generator/loader and tests, protocol/RPC regressions, and the real-CLI feedback test. Opaque review, generated diff inspection, and full local gates remain pending.

Revision note: The opaque guard exposed a newly erased native path reference. Added typed generation and a regression, reviewed every remaining opaque change, and set the fail-closed expected post-fix digest. Rerun the updater to confirm that exact inventory and inspect all emitted declarations.

Revision note: The actual post-fix source inventory exactly matches the original 104-entry baseline, rather than the predicted 103 entries. Restored its original digest after full line-by-line comparison. The guard behaved correctly; generation and full checks still require a successful rerun.

Revision note: Reviewed the complete emitted diff and confirmed stable additive APIs, typed save paths, exact raw metadata, image required-alternative handling, and unchanged compatibility aliases. Determinism and all local checks except the vulnerability scan passed. Added the required Go security patch update to COLIN-228 and this plan before rerunning validation.

Revision note: Patched vulnerability checks, native Go 1.27.3 verification, workflow/release fixtures, and the official CLI credential-free race tests pass. The authoritative updater is rerunning with Go 1.26.10; full completion and independent reviews remain pending.

Revision note: The authoritative updater passed every check under Go 1.26.10. Both generations print the exact reviewed inventories and remain byte-identical. Implementation is ready for QA, architecture, and security review before named-path staging.

Revision note: Completed all three reviews. Added QA's missing managed-model-hook and plan-type regressions, reran full unit and affected protocol race checks on Go 1.26.10/1.27.3, and received QA confirmation. No production or generated code changed after the authoritative gate. Rechecked unchanged origin/main, available SDK v0.164.0, exact generated headers, valid release tag, and protected automatic-release environments.
