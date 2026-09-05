# Update generated Codex protocol to rust-v0.146.0

This ExecPlan is a living document. The sections `Progress`, `Surprises & Discoveries`, `Decision Log`, and `Outcomes & Retrospective` must remain current while the work proceeds.

This plan follows `PLANS.md` at the repository root. It completes a new maintenance refresh after `plans/update-codex-protocol-rust-v0.145.0.md`; that earlier plan remains a completed record and must not be rewritten to claim that it targeted a tag that did not yet exist.

## Purpose / Big Picture

After this work, the Go SDK will target the latest exact stable Codex release, `rust-v0.146.0`, at upstream commit `e363b08c9175ac1cbe5893615dd2cb9ddf95043b`. Callers will receive the new external-agent import-history method, pinned-thread fields and filters, configuration requirement types, plugin command provenance, and the other compatible 0.146 schema additions through generated or high-level Go APIs.

The result is observable by inspecting `protocol.GeneratedCodexVersion`, compiling and testing the new API, running the repository updater through two identical generations and the complete local quality gate, and pushing a reviewed feature-branch commit that can enter `main` through a pull request. The protected release workflow, not a local updater or local tag, remains the only publication path.

## Tracker Mapping

Workflow: `WORKFLOW.md`.

The tracked task is the user's direct request on 2026-07-31 to finish the current branch so it can be merged, following the earlier direct invocation of `.codex/skills/update-codex-protocol/SKILL.md`. Following the repository's existing local-plan precedent, this file is the durable task record; no Linear issue was supplied or created.

## Progress

- [x] (2026-07-31) Inspected the dirty worktree and confirmed every existing edit belongs to the completed 0.145 refresh; no unrelated file was found.
- [x] (2026-07-31) Ran focused root, protocol, RPC, and generator tests against the existing 0.145 changes; all passed.
- [x] (2026-07-31) Fetched upstream tags through the updater and verified that new exact stable tag `rust-v0.146.0` supersedes the four-day-old 0.145 target.
- [x] (2026-07-31) Let the updater stop at the changed-union safety gate before it wrote 0.146 generated artifacts.
- [x] (2026-07-31) Captured complete 0.145 and 0.146 union inventories and compared all hashes and canonical JSON.
- [x] (2026-07-31) Completed the required independent planner audit and validated its compatibility findings against the repository and upstream source.
- [x] (2026-07-31) Created this separate ExecPlan before approving the new inventory digest or editing compatibility code.
- [x] (2026-07-31) Added focused coverage for omitted, false, and true pinned-thread filters; the new import-history RPC method; plugin command provenance; nullable requirements; the pinned response field; and the removed app-metadata export.
- [x] (2026-07-31) Updated handwritten high-level and protocol adapters for the 0.146 pinned-thread fields.
- [x] (2026-07-31) Recorded the typed union decision, approved digest `a3739c2d1017118cbdcac9469cfd770cc2947893fbef4fb12500ee38b6c43aba`, and verified that the opaque-interface inventory stayed unchanged.
- [x] (2026-07-31) Regenerated from `rust-v0.146.0`, inspected the generated API, and preserved the removed app-metadata field deterministically with a deprecated compatibility declaration.
- [x] (2026-07-31) Aligned the protected trusted-E2E CLI pin and four official release-archive checksums with 0.146.0; the secretless installer and release-tag fixtures passed.
- [x] (2026-07-31) Ran focused tests and the updater's complete deterministic quality gate with Go 1.26.5; the final run passed every gate.
- [x] (2026-07-31) Ran independent architecture, QA, and security reviews after local verification; resolved every QA finding and completed clean follow-up reviews.
- [x] (2026-07-31) Updated this plan's outcomes and evidence, committed the reviewed branch as `705c240`, and pushed `codex/awesome-go-quality` to `origin` without creating a release tag.

## Surprises & Discoveries

- Observation: The live stable target changed after the 0.145 work was completed but before it was committed.
  Evidence: On 2026-07-31 the updater fetched `rust-v0.146.0`, selected it as the highest exact `rust-vMAJOR.MINOR.PATCH` tag, and resolved it to `e363b08c9175ac1cbe5893615dd2cb9ddf95043b`.

- Observation: The union inventory change is small but not a header-only change.
  Evidence: The aggregate digest changed from `80fb7367165ee22e9e50553f7d6463ce0667e029c91b58950b6fa63d6e8ffcab` across 324 schemas to `a3739c2d1017118cbdcac9469cfd770cc2947893fbef4fb12500ee38b6c43aba` across 328 schemas. Ten hashes were added and six removed.

- Observation: Four genuinely new union schemas are nullable references; the remaining changed hashes are containing schemas.
  Evidence: The added standalone schemas are `BrowserUseRequirements | null` and `FeedbackRequirements | null` in both unprefixed and `v2` definition namespaces. Changed `ClientRequest` bundles contain `externalAgentConfig/import/recordHistory`; changed `ThreadItem` bundles add optional `pluginId` and `scriptPath` fields to the known command-execution variant. `ServerNotification` changes because its embedded definitions changed, not because it gained an unreviewed notification variant.

- Observation: The protected release path is not currently aligned with a 0.146 generated protocol.
  Evidence: `.github/codex/version` records `0.144.4`, while `.github/workflows/release.yml` rejects a generated/CLI major-minor mismatch. The pin and its four archive checksums must be updated before a 0.146 release can pass the protected gate.

- Observation: `protocol.Thread` is handwritten, so exact-ref generation failed closed until its new `isPinned` response field was added manually.
  Evidence: The first 0.146 generation reported `Thread missing isPinned`; adding `IsPinned bool` to the existing manual type satisfied the coverage gate without weakening it.

- Observation: Upstream removed `AppMetadata.firstPartyType`, which would otherwise remove two exported Go declarations during this maintenance refresh.
  Evidence: A schema-sanitization compatibility hook now restores the optional nullable field with a deprecation description. Generated `AppMetadata.FirstPartyType` and `AppMetadataFirstPartyType` remain source-compatible and are covered by generator, protocol, and external-package compile tests.

- Observation: The 0.146 refresh does not expand the generator's opaque-interface exception set.
  Evidence: The post-generation inventory remained 97 entries with digest `c6fa5c442b4de0dc9cc3000f0adf757a9a891b83b88c69199b4247a7c180be5c`.

## Decision Log

- Decision: Keep the completed 0.145 plan and create this new 0.146 plan.
  Rationale: The earlier plan accurately records a completed, reviewed maintenance event. Rewriting it would erase when the new stable tag became available and make its validation transcript misleading.
  Date/Author: 2026-07-31 / Codex

- Decision: Treat `rust-v0.146.0` at commit `e363b08c9175ac1cbe5893615dd2cb9ddf95043b` as the only generation source.
  Rationale: The updater fetched tags and selected the highest tag matching the exact stable-release pattern while ignoring alpha tags.
  Date/Author: 2026-07-31 / Codex

- Decision: Approve only the four new nullable-reference union schemas and the reviewed containing-schema changes; do not add an opaque exception preemptively.
  Rationale: Nullable concrete references must remain typed pointers through the generator's existing nullable-reference collapse. The new request and command fields are ordinary concrete schema additions covered by existing generation. Any new opaque output must stop generation and receive its own focused review and test.
  Date/Author: 2026-07-31 / Codex

- Decision: Preserve source compatibility for exported declarations removed upstream when a narrow generated tombstone is possible.
  Rationale: A protocol refresh should not silently break callers of an existing exported SDK symbol. Any tombstone must retain the old name and values, carry a deprecation comment, be generated deterministically, and have a regression test.
  Date/Author: 2026-07-31 / Codex

- Decision: Align the trusted-E2E CLI pin in this branch but do not tag or publish locally.
  Rationale: A merged 0.146 protocol with a 0.144 CLI pin cannot pass the protected release preflight and would test the wrong runtime. Updating immutable input metadata is merge preparation; release publication still requires `main`, hosted quality and trusted E2E, and protected-environment approval.
  Date/Author: 2026-07-31 / Codex

- Decision: Validate live pinned metadata separately from the installer's synthetic fixtures.
  Rationale: The installer fixture proves platform mapping, archive extraction, checksum enforcement, and failure behavior using temporary archives; it cannot prove that the committed version and four official digests are complete and aligned. The new validator checks major/minor protocol alignment, the exact supported archive set, digest syntax, and uniqueness, with focused positive and negative fixtures in the updater and Quality workflow.
  Date/Author: 2026-07-31 / Codex, after QA review

## Outcomes & Retrospective

The 0.146 implementation and local verification are complete. Generated metadata targets exact stable `rust-v0.146.0` at commit `e363b08c9175ac1cbe5893615dd2cb9ddf95043b`. The generated surface includes the external-agent import-history request, typed browser and feedback requirements, plugin command provenance, plan type `ent26`, and other upstream additions. Handwritten adapters expose pinned-thread listing and response state, while generator sanitization retains the removed app-metadata field and type as deprecated source-compatibility declarations.

The CLI pin is `0.146.0`, with all four archive digests taken from the official release assets. The live metadata validator, its negative fixture suite, the installer fixture, and `validate-release-tag.sh v0.146.0` passed. The authoritative updater completed two byte-identical generations and passed formatting, metadata and installer fixtures, vet, all unit tests, race tests, Staticcheck v0.7.0, govulncheck v1.3.0, and `git diff --check` with Go 1.26.5; govulncheck reported no vulnerabilities. The modified workflows also passed actionlint v1.7.7.

Architecture review found no actionable issues in the manual/generated boundary or compatibility sanitizer. QA initially found missing live-metadata validation, incomplete legacy command coverage, and a missing high-level pinned-thread replay assertion; all were implemented, and follow-up QA found no remaining actionable issue. Security independently reconciled the tag, commit, and four archive hashes with GitHub's official APIs, found no actionable security issue, and confirmed the new validator fails closed.

The reviewed implementation was committed as `705c240` (`Update Codex protocol from rust-v0.146.0`) and pushed to `origin/codex/awesome-go-quality`. No local or remote release tag was created. Hosted CI, credentialed trusted E2E, pull-request approval, merge, and protected release dispatch remain external acceptance steps.

## Context and Orientation

`codex-sdk-go` exchanges line-delimited JSON-RPC messages with `codex app-server`. Checked-in generated wire types live in `protocol/*_gen.go`; generated RPC client methods, notification parsers, and server-request dispatch live in `rpc/*_gen.go`. `gen.go` runs `internal/codegen`, which exports schemas from an exact local OpenAI Codex ref and writes deterministic Go artifacts.

`internal/codegen/main.go` contains `approvedUnionInventorySHA256`. Before compatibility sanitization, generation inventories every `oneOf` and `anyOf`. A changed digest stops generation so ambiguous wire shapes cannot silently degrade into `interface{}`. `approvedOpaqueInterfaceInventorySHA256` separately inventories every generated empty-interface use, including nested fields and collection elements.

`protocol/manual_types.go` holds compatibility declarations that cannot be generated safely. It includes both `ThreadListParams` and `Thread`, so the upstream `isPinned` filter and response field must be added there manually. `thread_lifecycle.go` exposes the corresponding high-level `ThreadListOptions` adapter.

`internal/codegen/main.go` already emits deprecated compatibility declarations for removed upstream exports such as `AmazonBedrockCredentialSource`. The 0.146 diff must be checked for any newly removed public symbol, particularly the app-metadata first-party type, before generation is accepted.

`.github/codex/version` and `.github/codex/checksums.txt` pin the exact Codex CLI archives used by secretless installer tests and protected trusted E2E. The four checksums must come from authoritative 0.146 release assets and must be validated by `.github/scripts/test-install-codex-cli.sh`; no credential or `.envrc` content is needed.

## Plan of Work

Begin with focused tests. Extend `account_thread_test.go` so an explicit pinned filter survives `ThreadListOptions.toParams` and appears as `isPinned` in the wire request. Extend `rpc/generated_test.go` for `externalAgentConfig/import/recordHistory` and for decoding the new optional command-execution plugin provenance. Extend generator tests if an upstream-removed exported app-metadata declaration requires a compatibility tombstone.

Update `protocol.ThreadListParams` in `protocol/manual_types.go` and `ThreadListOptions` in `thread_lifecycle.go` with `IsPinned *bool`, preserving omitted, explicit false, and explicit true values. Keep the public high-level API independent of raw RPC details.

Review the complete union-inventory delta recorded in `/tmp/codex-union-inventory-0.145.0-current.log` and `/tmp/codex-union-inventory-0.146.0.log`. After confirming that the four new nullable references remain typed and that every containing-schema change uses existing concrete generation, replace `approvedUnionInventorySHA256` with `a3739c2d1017118cbdcac9469cfd770cc2947893fbef4fb12500ee38b6c43aba`.

Run exact-ref generation from `rust-v0.146.0`. If generation reports a changed opaque-interface digest, inspect every added and removed declaration path before approving it. If a removed upstream schema would remove an existing exported SDK symbol, add the smallest deterministic compatibility declaration and generator test rather than editing generated output by hand. Resolve any generated/manual name conflict in the generator's reviewed manual-type set.

Inspect the generated diff. Every generated header must name `e363b08c9175ac1cbe5893615dd2cb9ddf95043b`, and `protocol.GeneratedCodexVersion` must be `0.146.0`. Verify the new import-history method and concrete params/response types, typed browser and feedback requirements, pinned thread fields, optional command plugin provenance, and the absence of new unreviewed empty interfaces.

Update `.github/codex/version` to `0.146.0` and `.github/codex/checksums.txt` to the four authoritative release-archive checksums. Run the installer fixture so archive names, digests, extraction, target selection, and failure cases remain verified without credentials.

Run focused tests, then the updater with `--allow-dirty` and `GOTOOLCHAIN=go1.26.5`. The successful updater run is the authoritative two-pass deterministic generation and complete local gate. After it passes, run architecture and QA review agents, fix findings, and rerun the updater if any code or generated artifact changes.

Finish by updating this plan's progress and outcomes. Stage and commit the reviewed files with an imperative message, then push `codex/awesome-go-quality` to `origin` for a pull request. Do not create, move, or push an SDK release tag; the protected release workflow owns that action after merge.

## Concrete Steps

Work from `/Users/pme/src/pmenglund/codex-sdk-go`.

Before each mutation group, inspect:

    git status --short --branch
    git diff --check

Run focused tests while implementing:

    GOTOOLCHAIN=go1.26.5 go test . ./protocol ./rpc ./internal/codegen

Generate from the reviewed exact ref while diagnosing compatibility:

    CODEX_REPO_ROOT=/Users/pme/src/openai/codex CODEX_REPO_REF=rust-v0.146.0 GOTOOLCHAIN=go1.26.5 go generate ./...

Validate the pinned CLI installer metadata:

    .github/scripts/test-install-codex-cli.sh

After reviewing the intentional dirty diff, run the authoritative gate:

    GOTOOLCHAIN=go1.26.5 .codex/skills/update-codex-protocol/scripts/update_codex_protocol.sh --allow-dirty

Inspect the result with:

    git status --short --branch
    git diff --stat
    git diff --check
    rg -n "Source codex commit:" protocol/*_gen.go rpc/*_gen.go
    git diff -- protocol rpc internal/codegen thread_lifecycle.go account_thread_test.go .github/codex

Only after local verification and independent review:

    git add <reviewed paths>
    git commit -m "Update Codex protocol from rust-v0.146.0"
    git push -u origin codex/awesome-go-quality

Do not run `git tag`, push a tag, dispatch the Release workflow, or merge into `main` from this plan.

## Validation and Acceptance

The branch is accepted only when all of the following are true:

- `protocol.GeneratedCodexVersion` is `0.146.0`, and every generated header names upstream commit `e363b08c9175ac1cbe5893615dd2cb9ddf95043b`.
- All four new nullable browser/feedback requirement schemas generate concrete pointer fields rather than `interface{}`.
- `ThreadListOptions.IsPinned` preserves omitted, explicit false, and explicit true values through the handwritten protocol params and JSON-RPC request.
- The generated import-history record method uses concrete params and response types and the wire name `externalAgentConfig/import/recordHistory`.
- Command-execution thread items decode optional `pluginId` and `scriptPath` without changing existing required-field validation or unknown-variant round trips.
- Any exported declaration removed upstream is either proven unused/private or retained through a tested, deprecated deterministic compatibility declaration.
- No new opaque union or empty-interface exception exists unless this plan records the exact declaration paths, rationale, and regression test.
- `.github/codex/version` and all four checksums match authoritative 0.146.0 release assets, and `.github/scripts/test-install-codex-cli.sh` passes.
- The updater produces identical first and second generations and passes formatting, installer fixtures, vet, all unit tests, race tests, Staticcheck v0.7.0, govulncheck v1.3.0, and `git diff --check` with Go 1.26.5.
- Architecture and QA reviews have no unresolved actionable findings.
- The ExecPlan outcomes match the final diff, the feature branch is committed and pushed, and no release tag has been created or moved.

Credentialed trusted E2E, hosted CI, pull-request approval, merge, protected release dispatch, and external acceptance remain explicitly unverified until they occur in their protected environments.

## Idempotence and Recovery

The exact-ref generation and updater are safe to rerun after the dirty diff has been reviewed and `--allow-dirty` is supplied. The updater writes generated artifacts deterministically and compares a second generation with the first.

If generation stops at a digest gate, preserve the exact error and inspect every changed schema or declaration path before updating the digest. Do not recover by editing `protocol/*_gen.go` or `rpc/*_gen.go` manually. If an unrelated change appears, stop before the next generation or staging operation and separate ownership with the user.

CLI version and checksum metadata are plain text and can be restored from the pre-change diff if an authoritative asset is unavailable. A missing asset blocks release-pin alignment but must not be replaced with a guessed digest.

No SDK tag may be moved. If a release fails after merge, fix forward on `main` and publish a higher immutable version through the protected workflow.

## Artifacts and Notes

The reviewed baseline inventory is `80fb7367165ee22e9e50553f7d6463ce0667e029c91b58950b6fa63d6e8ffcab` across 324 schemas.

The target inventory is `a3739c2d1017118cbdcac9469cfd770cc2947893fbef4fb12500ee38b6c43aba` across 328 schemas. The raw comparison contains 10 added hashes and 6 removed hashes; four net-new schemas are the two nullable requirement references in two namespaces.

Temporary per-schema evidence:

    /tmp/codex-union-inventory-0.145.0-current.log
    /tmp/codex-union-inventory-0.146.0.log

The initial live updater stopped before writing 0.146 generated artifacts:

    schema union inventory changed: got sha256 a3739c2d1017118cbdcac9469cfd770cc2947893fbef4fb12500ee38b6c43aba across 328 unique union schemas

## Interfaces and Dependencies

The completed high-level addition in `thread_lifecycle.go` must be:

    type ThreadListOptions struct {
        IsPinned *bool
        // existing fields remain
    }

The matching manual wire field in `protocol/manual_types.go` must be:

    IsPinned *bool `json:"isPinned,omitempty"`

All other 0.146 protocol types and RPC methods should remain generated unless a fail-closed generator error proves a narrow manual compatibility declaration is required.

No runtime dependency is introduced. Generation continues to use the repository's existing Go generator and exact upstream Rust schema exporter. Validation continues to use pinned Staticcheck v0.7.0 and govulncheck v1.3.0.

Revision note: Created on 2026-07-31 after the live updater discovered exact stable `rust-v0.146.0`, captured the changed union inventory, and completed the required planner audit. Updated after implementation, review remediation, follow-up review, and the final complete local gate to record the manual `Thread` coverage discovery, source-compatible app-metadata tombstone, unchanged opaque inventory, CLI pin validation, and passing evidence. This plan preserves the finished 0.145 record while making the feature branch and protected release inputs consistent with the new stable protocol.
