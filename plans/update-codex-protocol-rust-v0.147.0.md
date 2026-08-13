# Update generated Codex protocol to rust-v0.147.0

This ExecPlan is a living document. The sections `Progress`, `Surprises & Discoveries`, `Decision Log`, and `Outcomes & Retrospective` must remain current while the work proceeds.

This plan follows `PLANS.md` at the repository root. It is a new maintenance record after the completed `plans/update-codex-protocol-rust-v0.146.0.md`; the earlier plan remains unchanged because it accurately describes a different stable release.

## Purpose / Big Picture

After this work, `codex-sdk-go` will generate its checked-in Go protocol and JSON-RPC artifacts from the exact stable upstream tag `rust-v0.147.0` at commit `be6e8eac029b183056b7e4402879f15d2c85f61b`. SDK users will receive the 0.147 thread-section protocol, updated request and notification fields, and source-compatible aliases or deprecated fields where upstream renamed or removed an exported Go surface.

The result is observable by reading `protocol.GeneratedCodexVersion`, compiling the new section APIs, and running `.codex/skills/update-codex-protocol/scripts/update_codex_protocol.sh --allow-dirty`. That updater must generate twice with byte-identical results and pass formatting, metadata fixtures, installer fixtures, vet, unit tests, race tests, Staticcheck, govulncheck, and `git diff --check`. All changes remain unstaged. This task does not commit, tag, push, publish, dispatch a release, or run credentialed hosted end-to-end tests.

## Tracker Mapping

Workflow: `WORKFLOW.md`.

The tracked maintenance request is the user's direct invocation of `.codex/skills/update-codex-protocol/SKILL.md` on 2026-08-12. No Linear identifier was supplied. Following the repository's established protocol-refresh precedent, this file is the durable local task record.

## Progress

- [x] (2026-08-12) Inspected `git status` and confirmed the only pre-existing changes are nine generated protocol/RPC files from an unfinished but related 0.146.1 refresh.
- [x] (2026-08-12) Ran the updater with `--allow-dirty`; it fetched current tags, selected exact stable `rust-v0.147.0`, and stopped before writing 0.147 output because upstream removed the old `codex-app-server-protocol` `export` binary.
- [x] (2026-08-12) Completed the required independent planner pass and validated its exporter recommendation against the exact 0.147 upstream source.
- [x] (2026-08-12) Created this new ExecPlan before changing the generator or approving a changed union inventory.
- [x] (2026-08-12) Replaced the removed schema-export binary command with the upstream Codex CLI schema command, focused tests, documentation, and a single exact-tag worktree reused by both updater generations.
- [x] (2026-08-12) Captured and reviewed the complete 0.147 union inventory before approving digest `8a749e9a763542233e9e554ebe560e1c180338d10a33006f5dfcd637887dbfd4`.
- [x] (2026-08-12) Updated handwritten protocol and high-level thread-list adapters for sections while preserving intentional source compatibility.
- [x] (2026-08-12) Regenerated from exact `rust-v0.147.0`, inspected renamed and removed exports, and approved the reviewed opaque-interface digest `364e37d37ba2acc20715aafc38daa785ba3527eeed29e6aaad056ff5f48e5af7`.
- [x] (2026-08-12) Updated focused protocol, RPC, high-level, external-compilation, and generator tests for the reviewed 0.147 surface.
- [x] (2026-08-12) Aligned `.github/codex/version` and the four official archive checksums with 0.147.0 and passed the secretless metadata and installer fixtures.
- [x] (2026-08-12) Ran the updater's complete deterministic local quality gate and inspected the final unstaged diff and generated provenance.
- [x] (2026-08-12) Ran independent QA and architecture reviews, addressed every actionable finding, and reran the complete quality gate successfully.

## Surprises & Discoveries

- Observation: The checkout began with a related generated-only 0.146.1 diff rather than a clean 0.146.0 baseline.
  Evidence: Nine files under `protocol/` and `rpc/` name commit `79b4f03d35962b005b007a015113b38930711665` and `protocol.GeneratedCodexVersion` 0.146.1. No handwritten or unrelated file was dirty.

- Observation: Upstream removed the package-local `export` binary in 0.147, but the Codex CLI already provides a direct JSON-schema command in both 0.146.1 and 0.147.0.
  Evidence: The failing command reported `no bin target named export in codex-app-server-protocol`. Exact-tag source defines `codex app-server generate-json-schema --out <DIR>` in `codex-rs/cli/src/main.rs`, and `codex-rs/cli/Cargo.toml` names package `codex-cli` and binary `codex`.

- Observation: The 0.147 schema replaces pinned-thread organization with persisted thread sections.
  Evidence: Upstream removes `isPinned` from `Thread`, `ThreadListParams`, and `ThreadMetadataUpdateParams`; it adds `Thread.section`, `Thread.sectionEnteredAt`, a tri-state `ThreadListParams.sectionId`, five section RPC methods, and `section_position` sorting.

- Observation: The protected runtime pin is behind the target protocol.
  Evidence: `.github/codex/version` and `.github/codex/checksums.txt` currently describe 0.146.0, while the local metadata validator requires the generated protocol and pinned CLI major/minor versions to agree.

- Observation: The reviewed 0.147 union inventory is a bounded concrete delta rather than a new ambiguous variant family.
  Evidence: The digest changes from `a3739c2d1017118cbdcac9469cfd770cc2947893fbef4fb12500ee38b6c43aba` across 328 schemas to `8a749e9a763542233e9e554ebe560e1c180338d10a33006f5dfcd637887dbfd4` across 334. The comparison has 17 added and 11 removed canonical schemas. Six net additions are nullable references to `DesktopOnboardingEntrypoint`, `PluginDisabledReason`, and `ThreadSection` in root and `v2` namespaces. The other additions replace prior containing schemas for `ClientRequest`, `ServerRequest`, `ServerNotification`, `ThreadItem`, `CommandAction`, and `ResponseItem`; their changes add section RPCs and fields, a required blocking flag, onboarding and plugin metadata, concrete optional item fields, and existing typed references without introducing a new discriminator strategy.

- Observation: Building the schema command through `codex-cli` compiles substantially more upstream crates than the removed protocol-only binary.
  Evidence: An uncached exact-ref export took about three minutes. Because generation uses a different temporary worktree path on every pass, Cargo's default worktree-local target directory would discard that build before the updater's required second generation.

- Observation: A shared Cargo target directory does not safely solve the changing-worktree-path problem for path dependencies.
  Evidence: Cargo retained duplicate artifacts keyed by temporary upstream source paths; the updater-owned target reached 16 GB, two interrupted worktrees retained another 22 GB, and the linker failed with `No space left on device`. Removing only those exact updater-owned paths recovered about 38 GB.

- Observation: The reviewed opaque-interface inventory shrank from 124 to 100 entries even though three new free-form or empty-response shapes appeared.
  Evidence: Digest `c6fa5c442b4de0dc9cc3000f0adf757a9a891b83b88c69199b4247a7c180be5c` is replaced by `364e37d37ba2acc20715aafc38daa785ba3527eeed29e6aaad056ff5f48e5af7`. The additions are `InitializeCapabilitiesExtensions map[string]interface{}` and empty sanitized responses for section delete and move. Twenty-seven removed opaque aliases are now concrete generated types, including initialization, config, filesystem, hooks, marketplace, plugin, skills, thread metadata, rollback, and notification surfaces.

- Observation: The apparent external-agent import type removals in an isolated schema comparison were not exported Go API removals.
  Evidence: The complete 0.147 bundle still generates `ExternalAgentConfigImportItemTypeSuccess` and `ExternalAgentConfigImportTypeResult` directly for the older import surface, while `ExternalAgentConfigImportHistoryRecordParams.ItemTypeResults` now intentionally uses the distinct concrete `ExternalAgentConfigImportHistoryRecordTypeResultParams`. Proposed compatibility aliases therefore caused duplicate declarations and could not preserve struct-literal assignment compatibility; the focused RPC fixture was updated to the new field type.

- Observation: Preserving removed pin fields with their old JSON tags would preserve compilation while sending an unsupported 0.147 request member.
  Evidence: Architecture review traced `ThreadListOptions.IsPinned` through `ThreadListParams` and the generated metadata request. Both request fields are now manual compatibility-only fields tagged `json:"-"`; external compilation and wire tests prove they remain assignable but never serialize.

- Observation: QA review identified two untested nullability boundaries despite the broad gate already passing.
  Evidence: Focused tests now prove zero-value list params omit `sectionId`, unsectioned listing emits `sectionId:null`, and moving a thread out of a section emits required `sectionId:null` while omitting `beforeThreadId`.

## Decision Log

- Decision: Preserve the related 0.146.1 generated diff as the reviewed input until successful deterministic 0.147 generation supersedes it.
  Rationale: The files belong to the invoked update workflow. Resetting them would discard user-owned work; a successful exact-ref generation can replace them safely and visibly.
  Date/Author: 2026-08-12 / Codex

- Decision: Replace the old exporter with `cargo run -p codex-cli --bin codex -- app-server generate-json-schema --out <DIR>` unconditionally.
  Rationale: This command exists in both the 0.146.1 baseline and 0.147.0 target and calls the same protocol library's JSON generator. A version-specific branch or Python wrapper would add complexity without preserving any needed behavior.
  Date/Author: 2026-08-12 / Codex

- Decision: Keep removed exported pinning fields as deprecated source-compatibility surfaces, but do not describe them as effective 0.147 server filters.
  Rationale: Removing public Go fields immediately would break compilation for existing SDK users. Retaining them with `json:"-"` is narrow and deterministic: old source still compiles, while 0.147 requests cannot contain the removed member. Documentation and tests make clear that sections replace pinning.
  Date/Author: 2026-08-12 / Codex

- Decision: Expose the new section filter without collapsing its three wire states.
  Rationale: `sectionId` distinguishes omitted (all sections), explicit JSON `null` (unsectioned threads), and a string (one section). A two-state pointer would silently make one supported request impossible.
  Date/Author: 2026-08-12 / Codex

- Decision: Keep publication and credentialed acceptance outside this updater task.
  Rationale: The selected skill explicitly leaves files unstaged and delegates release tags to the protected GitHub workflow. Hosted CI, protected trusted E2E, merge, and release approval cannot be proven by the local updater.
  Date/Author: 2026-08-12 / Codex

- Decision: Treat the local update as complete and delegate its subsequent Git/PR/release handoff to `plans/automate-protocol-release-handoff.md`.
  Rationale: The user subsequently authorized the skill to stage reviewed files, commit, push a feature branch, open and merge a protected pull request, and trigger GitHub's release flow. That later plan supersedes this plan's no-stage/no-push boundary only after the local deterministic gate and reviews recorded here have passed; local tagging and direct `main` pushes remain prohibited.
  Date/Author: 2026-08-12 / Codex

- Decision: Approve union inventory digest `8a749e9a763542233e9e554ebe560e1c180338d10a33006f5dfcd637887dbfd4` after reviewing all 17 additions and 11 removals.
  Rationale: The six net-new unions are nullable concrete references. Every remaining changed hash is a replacement for a known containing request, notification, thread-item, command-action, or response-item schema; known discriminators remain typed and unknown union variants remain governed by the existing raw-preserving wrappers. No new opaque exception follows from this approval.
  Date/Author: 2026-08-12 / Codex

- Decision: Make the updater create one exact-tag worktree and reuse it for both deterministic generations, leaving Cargo's target directory local to that worktree.
  Rationale: Both passes now see the same absolute source paths and can reuse one build. The exit trap removes that exact worktree and its target directory. This bounds disk usage and avoids cross-run artifact accumulation while preserving exact-tag generation.
  Date/Author: 2026-08-12 / Codex

- Decision: Approve opaque-interface inventory digest `364e37d37ba2acc20715aafc38daa785ba3527eeed29e6aaad056ff5f48e5af7` after comparing all 100 entries with the 0.146.1 generated surface.
  Rationale: The only new free-form payload is explicitly named extensions metadata, and the two new map responses represent schema-defined empty RPC responses. The other net change removes 27 opaque aliases in favor of concrete types; it does not weaken typing or add a named `interface{}` exception.
  Date/Author: 2026-08-12 / Codex

## Outcomes & Retrospective

The SDK now deterministically targets exact upstream `rust-v0.147.0` at commit `be6e8eac029b183056b7e4402879f15d2c85f61b`. All nine generated protocol/RPC headers record that commit, `protocol.GeneratedCodexVersion` is `0.147.0`, and the protected CLI metadata contains the four official 0.147.0 archive checksums.

The generator follows upstream's current `codex app-server generate-json-schema` entrypoint. The updater creates one exact-tag temporary worktree and reuses it for both generation passes, so Cargo's large CLI build is reused without accumulating cross-worktree artifacts. Its exit trap removed the worktree and target directory after the successful run.

The 0.147 section surface is available through concrete generated CRUD/move RPC methods. Manual `Thread` and thread-list types preserve section metadata and all three `sectionId` states. Deprecated pin fields remain source-compatible but are excluded from 0.147 request JSON. The union and opaque-interface inventory changes were reviewed before their digests were approved.

The final authoritative updater passed deterministic two-pass generation, formatting, metadata validation, installer fixtures, vet, all tests, race tests, Staticcheck 0.7.0, govulncheck 1.3.0, and `git diff --check`. QA review requested explicit omission/null/move-out and external compatibility fixtures; architecture review requested preventing legacy pin serialization and completing this plan. All findings were implemented, and the full updater passed again afterward. No actionable review findings remain.

All changes are unstaged. No commit, tag, push, release dispatch, hosted CI, or credentialed E2E ran.

## Context and Orientation

`codex-sdk-go` starts `codex app-server` and exchanges line-delimited JSON-RPC messages with it. `protocol/` contains checked-in wire types, and `rpc/` contains checked-in clients, notification decoders, and server-request dispatch. `gen.go` invokes `internal/codegen/main.go`, which creates a temporary detached worktree at the requested upstream tag, exports JSON schemas, converts supported shapes to Go, and fails closed on unreviewed unions or `interface{}` output.

`internal/codegen/main.go` currently exports schemas with a Cargo binary upstream deleted in 0.147. Its `approvedUnionInventorySHA256` inventories every canonical `oneOf` and `anyOf` before schema sanitization. `approvedOpaqueInterfaceInventorySHA256` separately inventories weak generated Go types. Neither digest may change until every added and changed entry is reviewed.

`protocol/manual_types.go` owns types whose upstream schemas exceed the generic generator's safe representation. `Thread` and `ThreadListParams` are manual, so the 0.147 section fields must be added there. `thread_lifecycle.go` maps user-facing `ThreadListOptions` to the manual wire params and validates typed sort values.

The pre-existing dirty generated files target 0.146.1. They are related to the update request but are not accepted 0.147 output. Generated files must never be edited by hand; generator or manual-adapter changes must cause deterministic regeneration.

## Plan of Work

First, update `internal/codegen/main.go` so `exportSchemas` invokes the upstream `codex-cli` binary with `app-server generate-json-schema`. Refactor only enough to unit-test the exact command arguments and working directory without launching Cargo. Keep command output visible when real generation runs, and retain contextual execution errors. Update `APP.md` so its code-generation description names the current command.

Run exact-ref generation with `CODEX_PRINT_UNION_INVENTORY=1`. Record the aggregate digest and sorted per-schema entries. Compare the 0.146.1 and 0.147.0 inventories, inspect all added and changed canonical JSON, and update `approvedUnionInventorySHA256` only after proving the new nullable references and containing request/notification bundles retain typed generation. If the opaque-interface gate changes, print that inventory and apply the same per-path review before changing its digest.

Update `protocol/manual_types.go` with `Thread.Section`, `Thread.SectionEnteredAt`, and a tri-state section filter in `ThreadListParams`. Retain removed `IsPinned` fields with deprecation comments for source compilation only. Extend the generator's removed-field sanitizer for generated `ThreadMetadataUpdateParams.IsPinned` if exact generation would otherwise remove that exported field and backing type. Add deterministic compatibility aliases for renamed external-agent import result types when their replacement shapes are compatible.

Update `thread_lifecycle.go` to expose all-sections, unsectioned, and concrete-section filtering without ambiguous input, and accept `protocol.ThreadSortKeySectionPosition`. Keep the new section create/list/update/delete and thread move operations on the generated low-level RPC client in this maintenance task; high-level convenience methods are separate feature work.

Regenerate from exact `rust-v0.147.0`. Inspect every generated header, metadata version, new RPC method, renamed or removed exported declaration, fallback, and opaque field. Add focused tests in `internal/codegen/main_test.go`, `account_thread_test.go`, `protocol/`, `rpc/generated_test.go`, and `api_external_test.go` for the reviewed behavior and compatibility surface.

Update `.github/codex/version` to 0.147.0 and replace the four checksums with SHA-256 digests reported by the official GitHub release assets. Run the repository's secretless metadata and installer fixtures. Finally run the authoritative updater with `--allow-dirty`, inspect the resulting unstaged diff, run independent QA and architecture reviews, address findings, rerun the full gate if any code changes, and complete this plan.

## Concrete Steps

Work from `/Users/pme/src/pmenglund/codex-sdk-go`.

Inspect ownership before each mutation group:

    git status --short --branch
    git diff --check

Run focused generator tests while migrating the exporter:

    GOTOOLCHAIN=go1.26.5 go test ./internal/codegen

Print and review the target inventory before approving it:

    CODEX_REPO_ROOT=/Users/pme/src/openai/codex CODEX_REPO_REF=rust-v0.147.0 CODEX_PRINT_UNION_INVENTORY=1 GOTOOLCHAIN=go1.26.5 go generate ./...

Run focused package tests while adapting the API:

    GOTOOLCHAIN=go1.26.5 go test . ./protocol ./rpc ./internal/codegen

Validate exact metadata and installer inputs:

    .github/scripts/validate-codex-metadata.sh
    .github/scripts/test-validate-codex-metadata.sh
    .github/scripts/test-install-codex-cli.sh

Run the authoritative gate after reviewing every dirty path:

    GOTOOLCHAIN=go1.26.5 .codex/skills/update-codex-protocol/scripts/update_codex_protocol.sh --allow-dirty

Inspect the completed result:

    git status --short --branch
    git diff --stat
    git diff --check
    rg -n "Source codex commit:" protocol/*_gen.go rpc/*_gen.go
    rg -n "GeneratedCodexVersion" protocol/metadata_gen.go

Do not run `git add`, `git commit`, `git tag`, `git push`, a release workflow, or credentialed E2E from this plan.

## Validation and Acceptance

Acceptance requires every generated header to name `be6e8eac029b183056b7e4402879f15d2c85f61b` and `protocol.GeneratedCodexVersion` to equal `0.147.0`.

The schema exporter test must prove the generator runs Cargo from `<codex-root>/codex-rs` with package `codex-cli`, binary `codex`, and arguments `app-server generate-json-schema --out <DIR>`. Real exact-ref generation must succeed through that command twice with identical manifests.

The 0.147 union and opaque-interface digests are accepted only if this plan records their reviewed deltas. Nullable references must remain concrete pointers, known discriminated unions must decode to typed variants, and unknown variants must continue to round-trip their raw JSON.

`Thread` must decode optional `section` and `sectionEnteredAt`. Thread listing must represent omitted `sectionId`, explicit JSON `null`, and a concrete section ID distinctly. `section_position` must pass high-level sort validation. Removed pinning names must still compile with deprecation guidance but must not be presented as effective 0.147 organization behavior.

The generated RPC client must expose concrete methods and params/responses for `threadSection/list`, `threadSection/create`, `threadSection/update`, `threadSection/delete`, and `thread/section/move`. Renamed external-agent import result types must either preserve compatible legacy aliases or have an explicit decision explaining why compatibility is impossible.

`.github/codex/version` and all four checksums must match official 0.147.0 assets, and all three local metadata/installer scripts must pass.

The final updater run must pass formatting, metadata fixtures, installer fixtures, vet, ordinary tests, race tests, Staticcheck v0.7.0, govulncheck v1.3.0, and `git diff --check`. QA and architecture reviews must have no unresolved actionable findings. All files remain unstaged.

Hosted CI, credentialed trusted E2E, pull-request approval, merge, protected release dispatch, and external acceptance remain unverified.

## Idempotence and Recovery

Exact-ref generation uses a temporary detached upstream worktree and is safe to rerun. The updater's second pass verifies byte determinism. `--allow-dirty` only permits reviewed existing changes; it does not stage or silently include them.

If an inventory gate stops generation, preserve its output and inspect every changed entry. Do not update a digest merely to proceed. If generation partially rewrites checked-in files, fix the generator or handwritten adapter and rerun; do not reset, checkout, or manually repair generated files.

If an unrelated dirty path appears, stop before the next generation or review step. If an official release checksum is unavailable, leave metadata alignment incomplete rather than guessing. Never move or recreate an SDK tag.

## Artifacts and Notes

The reviewed input is the generated-only 0.146.1 diff at upstream commit `79b4f03d35962b005b007a015113b38930711665`. The target exact tag is `rust-v0.147.0` at `be6e8eac029b183056b7e4402879f15d2c85f61b`.

The first live updater failure was:

    Selected upstream Codex tag: rust-v0.147.0
    Running first deterministic generation
    error: no bin target named `export` in `codex-app-server-protocol` package

The official 0.147.0 release asset API reported these required SHA-256 values:

    75984b81f92a71b0c0f4b3b5cad80e5c57177e4d8c8b4b1e13db703b20dc4358  codex-aarch64-apple-darwin.tar.gz
    eb677c80f666b1ab8b4b1d083b66e8d614b1281d960bb6f9fd8ca98f58b38b90  codex-aarch64-unknown-linux-musl.tar.gz
    36e782f71d8164cc37c2b89c64948f2180e9a2f8456b27e660da75bc6b5574e2  codex-x86_64-apple-darwin.tar.gz
    0246e2e773834e07f0fb5249ed6ebad12e4591e608f8c7bb97dd6a9690544c36  codex-x86_64-unknown-linux-musl.tar.gz

## Interfaces and Dependencies

`internal/codegen.exportSchemas` continues to accept `(codexRoot, outDir string) error`. Its Cargo command changes to the upstream CLI entrypoint; no runtime Go dependency is added.

The final manual wire representation for `ThreadListParams.sectionId` must distinguish absent, null, and string. The high-level `ThreadListOptions` API may use separate `SectionID string` and `Unsectioned bool` fields, provided it rejects both being set and maps them deterministically to the three wire states.

All section CRUD and move request/response types remain generated from upstream schemas. Validation continues to use the versions pinned in the updater: Staticcheck v0.7.0 and govulncheck v1.3.0.

Revision note: Created on 2026-08-12 after the live updater selected 0.147.0, the removed exporter stopped generation, the required planner pass completed, and the proposed replacement command was verified against exact upstream source. The plan preserves the related 0.146.1 dirty baseline and defines the compatibility, inventory, metadata, review, and local no-publication boundaries for this refresh. The later Git/PR/release handoff is owned by `plans/automate-protocol-release-handoff.md`.
