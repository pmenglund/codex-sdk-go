# Update Codex protocol to rust-v0.160.0

This living ExecPlan follows `PLANS.md`. Maintain Progress, Surprises & Discoveries, Decision Log, and Outcomes & Retrospective throughout execution.

## Purpose / Big Picture

Refresh the SDK from Codex 0.154.0 to the latest stable rust-v0.160.0. Users receive current wire types and RPC methods while existing handwritten APIs retain intentional compatibility. Publish SDK v0.160.0 through a protected pull request and the automated Release gates.

## Tracker Mapping

Workflow: `WORKFLOW.md`. This plan records the explicit maintenance invocation on 2026-10-01. Linear authentication was restored during reconnaissance; broad Codex and exact codex-sdk-go searches returned unrelated Colin/Bothnia issues, without a matching SDK maintenance issue. The user explicitly authorized skipping a Linear issue for this update on 2026-10-01. This exception resolves the tracker boundary and permits implementation. Do not invent identifiers or attach work to an unrelated project.

## Progress

- [x] (2026-10-01) Inspected instructions and clean checkout; fetched HTTPS origin and fast-forwarded main to afbdfdacdc1894f66736a254fd35c6ff2fb94ca5.
- [x] (2026-10-01) Verified upstream highest exact stable tag rust-v0.159.3 and current immutable SDK v0.154.0; completed and validated the required planner proposal.
- [x] (2026-10-01) Created codex/update-protocol-v0.159.3 before generation.
- [x] (2026-10-01) Compiled exact upstream exporter with installed Rust 1.94.0; reproduced union guard at 440 schemas. Source comparison reproduces the old approved 414-schema digest and target export digest.
- [x] (2026-10-01) Completed printed-inventory rerun; all 440 printed canonical schemas exactly match the independently extracted target source inventory. No digest has been approved.
- [x] (2026-10-01) User explicitly waived a Linear issue for this update.
- [x] (2026-10-01) Revised the plan with planner advice and independently verified 0.160.0 source schemas exactly match the reviewed intermediate target.
- [x] (2026-10-01) Generated exact 0.160.0 output; reviewed all final printed union and opaque deltas before approvals.
- [x] (2026-10-01) Added manual compatibility and wire/RPC regressions, reproduced failures before fixes, and updated official CLI digests, SDK 0.160.0, and migration documentation.
- [x] (2026-10-01) Authoritative updater proved two identical generations and passed the complete Go 1.26.8 local gate, including race tests, Staticcheck, govulncheck, metadata fixtures, and diff hygiene.
- [x] (2026-10-01) Completed QA and architecture reviews; addressed guard coverage and attachment precision findings, with reviewer confirmation.
- [ ] Complete protected PR handoff.
- [ ] Monitor Release through publication and verify immutable tag identity.

## Surprises & Discoveries

Configured SDK SSH authentication failed, but GitHub CLI keyring authentication and HTTPS Git transport work. GitHub reports administrator and push access. Remote main included the completed 0.154.0 update, seven commits beyond the initial checkout. Installed Go is 1.27.1 and Cargo is 1.94.0; the target declares Rust 1.95.0, but installed Rust built the exact exporter successfully in 3m54s. No tooling installation was needed. Never source or print sensitive `.envrc`.

The first exact export stopped before writing generated output: union digest 0b3fbe996fafc3b6a356b643f52a9777cb49b5e89b8b08089e213c55281577ad across 440 schemas. The committed 0.154.0 schemas reproduce approved digest 214941a1bf78eb429f79fc0599b321d7ab251ab90b3a6b25ab552e095deb2139 across 414 entries. The target committed schemas independently reproduce the exporter digest, with 48 added and 22 removed canonical entries.

## Decision Log

Initially approved opaque digest `d0e53d1e866ff1d8aa7813ee13b81fad87527497b510b029c502e4d98fb608de` after the printed 108-entry inventory comparison; architecture review subsequently superseded this approval as described below. Baseline AST extraction reproduces the prior 100-entry digest exactly; no entry was removed. Eight additions are intentional: MCPServerStatus.ServerCapabilities has no constrained schema; three attachment payload fields are unrestricted JSON; ConfigRequirementsModelProviders explicitly permits arbitrary provider definitions; and three gateway/attachment empty response objects use the existing map representation for open response objects. No cursor or tool-surface opaque fallback is approved: those have manual typed representations.

Architecture review identified precision loss in unrestricted attachment payloads decoded into interface{}: integers above 2^53 become rounded float64 values. Use manual ThreadAttachment and ThreadAttachmentAddParams with json.RawMessage Payload, retaining the sanitized params spelling as an alias and enforcing source schema field coverage. The focused regression preserves 9007199254740993, both as a scalar and nested object, plus null through response-to-request round trips. Printed opaque inventory removes exactly the three attachment payload fields and adds nothing; approve the resulting 105-entry digest `3dd6c1d855fb38cc550c690e767621de8d9a8c8bc7befcfa9d6cdf38eb152539`. Five newly opaque entries remain justified by unconstrained schemas. No separate security issue was found in the generated gateway bindings.

The full 0.160.0 exporter printed 440 canonical union entries exactly equal to the independently extracted 0.160.0 source inventory and the reviewed intermediate inventory. The union approval therefore applies directly to the final target. Preserve ThreadItemsListParams.Cursor and its canonical/sanitized string helper types; add CursorAnchor with nonempty turnId and conflict validation. Raw cursor wrappers exclude struct-field coverage because UnmarshalJSON validates the union, with focused round-trip tests. Nested image alternatives validate at least one required source-field group while retaining raw future members.

Remove upstream-deleted thread/rollback generated RPC/types rather than issuing an unsupported call or guessing a numTurns-to-turn-ID translation. Document ThreadRevert migration. This is an intentional low-level API change in the 0.x minor update; no high-level rollback method exists. Managed Windows sandbox requirements now describe implementation choices including mxc rather than the removed private-desktop flag; document the corresponding low-level requirement field/type migration.

Retarget to rust-v0.160.0, source commit a956835d020762cb2b570053af06f643a11c0ecc, because the updater refreshed tags and found a newer exact stable release before any generated output was accepted. Rename the existing branch and plan. Independently extracted 0.160.0 JSON schemas exactly equal every 0.159.3 schema; their 440-entry canonical inventory is identical. The direct 0.154.0-to-0.160.0 comparison therefore has the same 48 additions/22 removals already reviewed. Confirm printed exporter inventory and opaque/API output separately; do not rely on version equality alone.

Approve union digest `0b3fbe996fafc3b6a356b643f52a9777cb49b5e89b8b08089e213c55281577ad` after comparing all 440 printed exporter entries to target source and all 48 added/22 removed canonical schemas to the exactly reproduced 414-entry baseline. Changes comprise nullable refs and description changes, nested image URL/file-ID alternatives, item cursor/anchor, tool surface string alternatives, error enum additions, MCP UI fields, and enclosing containers listed in Artifacts. No new opaque union family is approved: the heterogeneous cursor needs an explicit manual wire wrapper, while images retain existing raw discriminator wrappers with alternative required-field validation. Existing reviewed opaque CodexErrorInfo remains raw-preserving.

Use one flat plan matching existing repository convention; keep large logs and inventories in temporary storage. The planner's proposed paths and commands match the updater and generator. Use SDK 0.160.0 provisionally because it exceeds v0.154.0; recheck tags before publication. Unset CODEX_REPO_REF so stale local configuration cannot select a different release. If the updater selects a newer stable tag, revise the target, branch, and plan before proceeding. The explicit skill invocation authorizes implementation and protected publication without another plan approval.

Preserve aliases where Go type identity matters, and retain obsolete source fields only when their wire behavior remains truthful. Never approve a digest until every changed canonical schema has been reviewed. Known discriminator variants remain typed; unknown variants retain raw JSON. No unrelated runtime dependency or architecture changes are planned.

## Outcomes & Retrospective

Compatibility implementation and complete local verification are complete. The authoritative updater passed deterministic generation, formatting, metadata/installer fixtures, vet, unit/race tests, Staticcheck 0.8.1, govulncheck 1.3.0, and diff hygiene using Go 1.26.8. Four release/tag fixture suites also pass. Final protocol race tests pass after QA requested explicit empty-turn-ID and zero-anchor marshal rejection cases.

QA independently verified source-schema coverage and focused generator/protocol/RPC tests. Its two missing guard regressions were added. Raw union constructors intentionally validate discriminators and required-member presence, following the existing raw-preserving convention; they do not perform complete schema type validation. Architecture review identified and resolved attachment number precision loss. The complete authoritative gate passed again after this fix, including two identical generations and no reported vulnerabilities. Protected handoff and publication remain pending.

## Context and Orientation

`gen.go` invokes `internal/codegen/main.go`, which compiles the selected upstream exporter and checks reviewed union and opaque-interface inventories. Generated wire types are in `protocol/*_gen.go`; RPC methods and dispatch are in `rpc/*_gen.go`. Handwritten protocol wrappers live in `protocol/manual_types.go`, with SDK adapters in the root package. Generated declarations must be fixed in the generator, never by hand, and every exported declaration needs exact-name GoDoc.

`.github/codex/version` and `.github/codex/checksums.txt` pin the checksum-verified CLI used by E2E. `.github/sdk-version` selects the SDK tag independently. `.github/workflows/release.yml` validates the protected main candidate, runs Quality and trusted E2E, and alone publishes the immutable tag.

## Plan of Work

Milestone 1 runs the bundled updater from the required feature branch and records the exact selected commit. Inspect the target toolchain declaration if compilation fails. On inventory guards, export printed hashes/JSON and compare the 0.154.0 baseline against the target. Classify every added, changed, and removed union, then record typed handling or explicit opaque decisions before approving digests in `internal/codegen/main.go`.

Milestone 2 inspects the complete generated diff and changes only affected handwritten wrappers, SDK adapters, generator policy, and tests. Use regression tests for changed omission/null semantics, alias identity, required members, routing, and unknown-variant preservation. Obtain the four supported CLI archive digests from the official GitHub release and update metadata, SDK version, and affected documentation. Run generation twice and the complete gate.

Milestone 3 runs QA and architecture review once implementation and local verification are complete; include security review for approval/auth/permission/command/network changes. Resolve actionable findings before named-path staging. Push the reviewed feature branch, open and attach its PR, watch required checks, and merge through protection. Monitor automatic Release and verify the tag peels to the exact gated main commit. Record the final publication evidence through a small protected documentation closeout PR if necessary.

## Concrete Steps

Run from `/Users/pme/src/pmenglund/codex-sdk-go`. Initially this plan is the sole permitted dirty path. Classify every additional dirty path before a rerun.

    env -u CODEX_REPO_REF CODEX_REPO_ROOT=/Users/pme/src/openai/codex GOTOOLCHAIN=go1.26.8 .codex/skills/update-codex-protocol/scripts/update_codex_protocol.sh --allow-dirty

Use CODEX_PRINT_UNION_INVENTORY=1 and CODEX_PRINT_OPAQUE_INTERFACE_INVENTORY=1 for diagnosis. Keep raw inventories outside the repository. HTTPS Git URL rewrites may be supplied for this invocation without changing repository remotes.

    go test ./internal/codegen ./protocol ./rpc .
    .github/scripts/validate-codex-metadata.sh
    .github/scripts/test-validate-codex-metadata.sh
    .github/scripts/test-install-codex-cli.sh
    .github/scripts/sdk-release-tag.sh
    .github/scripts/validate-release-tag.sh v0.160.0
    git diff --check

The final authoritative updater runs two generations, formatting, metadata/installer fixtures, vet, all unit/race tests, Staticcheck v0.8.1, govulncheck v1.3.0, and diff hygiene.

## Validation and Acceptance

All generated protocol/RPC headers must name the peeled target commit and GeneratedCodexVersion must equal 0.160.0. Generation must be byte-identical including new files. Meaningful regressions must demonstrate handwritten fixes. All local gates and required hosted checks must pass before successful trusted E2E and publication. The remote annotated v0.160.0 tag must peel to the protected-main candidate used by Release.

## Idempotence and Recovery

Recheck status before mutation and stage only reviewed named paths. Preserve concurrent changes. Never push main, create/move local SDK tags, bypass failed gates, or manually approve pending deployment jobs. If environment approval appears, restore only the configured protected-branch/no-reviewer policy while preserving other protections. An unpublished version may be retained for an exact-candidate retry or reviewed corrective merge. After publication, repair through a higher immutable version and a retract directive where appropriate.

## Artifacts and Notes

Upstream tag rust-v0.159.3 object 8e46774a94a745ffdf676bd7a8aa36466bbd4f99 peels to 01fc69f4026735edfdf6789820549727a4867b11, confirmed by the updater and git independently. First-run log is `/private/tmp/codex-01593-update.log`; printed-inventory rerun uses `/private/tmp/codex-01593-inventory.log` and persistent temporary Cargo cache `/private/tmp/codex-01593-cargo`.

Source research identifies six new client methods (three account/gatewayOAuth methods and three thread/attachment methods), removed thread/rollback, two new notifications, and unchanged ten server requests. Changed union groups cover nullable collaboration mode, gateway status, MCP UI/resource target, model access programs, plugin onboarding skill; personality/error descriptions; nested image URL/file-ID alternatives in UserInput, ContentItem, and FunctionCallOutputContentItem; heterogeneous item cursor and single-variant item anchor; tool exposure string alternatives; CodexErrorInfo additions flexUnavailable and tooManyDenials; MCP tool-call UI fields; and enclosing RPC containers. Full printed-export review must precede digest approval. Source-compatible item cursor handling, removed rollback API policy, nested image required alternatives, manual history-entry timestamps, and new nullable target linkId need explicit compatibility decisions and regression tests.

The official GitHub release API supplied the four supported CLI archive digests now pinned for 0.160.0.

## Interfaces and Dependencies

Keep behavior in the root SDK, transport in rpc, and wire data in protocol. Reuse existing go-jsonschema and CI-aligned Go toolchains. Rust is build tooling for schema export. Determine required adapters from actual generated differences rather than adding speculative APIs.

Revision note: Created after repository inspection and validated planner advice, before generation.

Revision note: Recorded completed exact schema research and the tracker boundary. Continue after the user supplies the Linear mapping or authorizes creation of an appropriately placed issue; do not treat the incomplete plan as release acceptance.

Revision note: User explicitly waived Linear tracking for this maintenance update; resumed implementation using the existing reviewed plan.

Revision note: Retargeted the existing plan and branch to the newer stable 0.160.0 release before accepted generated output; preserved the intermediate 0.159.3 research and log paths.

Revision note: Recorded exact final-target inventory approval, manual wire regressions, intended removed-method migration, official CLI digest pins, and live protected-main/environment settings.

Revision note: Focused internal/codegen, protocol, rpc, and root tests pass, including final external source-schema inventory/manual coverage checks. All generated headers name a956835d020762cb2b570053af06f643a11c0ecc. Live main protection requires strict Coverage report, Go 1.26, and Go 1.27 checks with zero approving reviews; both release environments permit protected branches and have no reviewer rules. Immutable v* tag deletion/non-fast-forward rules are active without bypass actors. No GitHub settings changed.

Revision note: Recorded the successful authoritative local gate and addressed QA guard coverage findings before staging.

Revision note: Addressed architecture review attachment payload precision loss with raw JSON and reviewed the reduced opaque inventory; source coverage and focused regression tests pass.

Revision note: Final authoritative updater passed after the attachment fix; architecture reviewer confirmed resolution and no new findings. Final gate log: /private/tmp/codex-01600-final-gate.log.
