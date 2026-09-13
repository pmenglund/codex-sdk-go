# Update Codex protocol to rust-v0.154.0

This ExecPlan follows `PLANS.md`. Keep Progress, Surprises & Discoveries, Decision Log, and Outcomes & Retrospective current.

## Purpose / Big Picture

Refresh the SDK from Codex 0.153.4 to stable rust-v0.154.0, commit `6b9826e3aa83b1a5947db50f4332cb9c65f1b340`. Users should receive the current protocol types and RPC methods with intentional compatibility in the handwritten SDK. Publish SDK v0.154.0 only through a protected pull request and the complete GitHub Release workflow.

## Tracker Mapping

Workflow: `WORKFLOW.md`. This plan tracks the user's explicit update-codex-protocol invocation on 2026-09-12. Linear search and team lookup both returned UNAUTHORIZED with reauthentication required. The workflow explicitly permits proceeding with available context while documenting the tracker gap. No issue identifier is available; do not invent one or create an item in an unrelated project.

## Progress

- [x] Verified a clean worktree, fetched origin, and fast-forwarded main to `a5ca6c46331f9d5298d0bfc315e3ad41d9cccb18`; the three initially local commits were already merged in PR #8.
- [x] Verified the highest stable upstream tag is rust-v0.154.0 and current SDK version is 0.153.4. Created `codex/update-protocol-v0.154.0` before generation.
- [x] Inspected repository instructions, previous release plan, generator entry point, and updater; documented unavailable Linear access.
- [x] Prepared Rust 1.95.0 and local OpenSSL build files; ran the updater and reproduced both inventory guards.
- [x] Reviewed all changed schemas and printed union/opaque inventories. Implemented originator fields/filtering and nullable rate-limit params, with regression tests for wire behavior and server callback alias compatibility.
- [x] Updated CLI metadata, documentation, and SDK version. The complete updater passed: byte-identical generations, formatting, metadata/installer fixtures, vet, unit/race tests, Staticcheck v0.8.1, govulncheck v1.3.0, and diff hygiene on Go 1.26.8.
- [x] Completed compatibility/security/QA review of the complete diff and independent source-versus-export/AST inventory comparisons; fixed callback alias identity regressions. Verified protected-main checks, immutable tag rules, and automatic protected-branch-only E2E/publication settings.
- [x] Committed the reviewed update as `acf6c73aced1f6487b568e2b4df8fbaca366b950` and pushed `codex/update-protocol-v0.154.0`.
- [ ] Create the PR and pass required hosted checks. Blocked: the connected GitHub integration rejected POST /repos/pmenglund/codex-sdk-go/pulls with HTTP 403, Resource not accessible by integration. Restore pull-request write permission before retrying.
- [ ] Merge through the protected PR, monitor Release, and verify the immutable tag points to the gated merged commit.

## Surprises & Discoveries

The initial origin/main reference was stale. Fetching showed the local commits were already merged remotely, allowing a simple fast-forward without losing work. This Linux environment provides Go 1.26.8 but initially has neither Cargo nor the GitHub CLI. Git transport works and the GitHub connector reports administrator/push access. Upstream requests Rust 1.95.0.

Rust 1.95.0 installed successfully. The first upstream compile stopped because OpenSSL development files were absent. Installed Debian bookworm libssl-dev 3.0.20-1~deb12u2 by downloading and extracting the package under `/tmp/codex-protocol-deps`, without system changes. The retry uses OPENSSL_INCLUDE_DIR, OPENSSL_LIB_DIR, and OPENSSL_STATIC=1. GitHub's connector can read repository rulesets but returned 403 on main's branch-protection endpoint; protection verification remains outstanding before publication.

Committed upstream JSON schemas reproduce the approved 0.153.4 union digest exactly when canonicalized with Go-compatible number/HTML encoding: 407 schemas, aa19cfcd55fcee125e20373138c395e420a673088dc3bb039d372e8d4b726fe2. The 0.154.0 source schemas have 414 entries, 18 added and 11 removed. Actual exported inventory still must be printed and checked before approval. Regression tests reproduced missing originator preservation, unrecognized configuration_update items, and missing originator request filtering before their fixes.

## Decision Log

Use SDK 0.154.0 provisionally because it exceeds the latest recorded immutable SDK tag v0.153.4 and matches the selected upstream release. Recheck remote tags before publication. Install required build tooling outside the repository and use an isolated upstream checkout; never source or print `.envrc`. Run the updater with CODEX_REPO_REF unset so stale configuration cannot override its selected stable tag.

Preserve AccountRateLimitsRead(ctx) and add optional capabilities through the existing generator's AccountRateLimitsReadWithParams(ctx, *protocol.GetAccountRateLimitsParams) mechanism. Preserve null originators explicitly in manual Thread and expose the hosted-backend originator filter in ThreadListOptions and manual ThreadListParams. The local server rejects nonempty originator lists, so document that restriction. Configuration updates use the existing typed ResponseItem wrapper. Guardian command/applyPatch paths change schema aliases from AbsolutePathBuf to LegacyAppPathString but retain Go string wire representations. Detached review delivery remains accepted and gains an upstream deprecation notice; do not remove it.

Approve union digest `214941a1bf78eb429f79fc0599b321d7ab251ab90b3a6b25ab552e095deb2139` after rerunning the exact source generation with CODEX_PRINT_UNION_INVENTORY=1. All 414 printed canonical schemas exactly match the reviewed source inventory. The 18 additions/11 removals cover three nullable account-rate-limit parameter wrapper schemas, two optional account parameter refs, two nullable application network refs, two review-delivery description changes, two Guardian path-alias changes, two ResponseItem unions gaining configuration_update, and five enclosing RPC container schemas whose definitions include these same changes. No RPC method, server request, or notification is added/removed (99/10/81 respectively). Existing typed/discriminator and nullable-reference generation handles all changed unions; no new opaque family is approved.

The printed opaque inventory adds only `type AttestationGenerateParams map[string]interface{}` (101 entries, digest `0c8635d44b86f88d9c69084d54afa00ab258a78e8b81935a84cc87285223f0e9`). An AST inventory of the committed generated types exactly reproduces the old 100-entry digest, proving no other entry changed. The canonical attestation schema is the unchanged unrestricted object `{ "type": "object" }`; the old canonical name aliases the already-opaque SanitizedAttestationGenerateParamsJSON. Approve this equivalent JSON representation for generation, then review the resulting Go alias identity before final compatibility acceptance.

The Go diff revealed that AttestationGenerateParams and ChatgptAuthTokensRefreshParams changed from aliases to named definitions, which would break callback implementations written with the previous target types. Regression tests failed on both changed identities. Preserve both existing aliases in manual_types.go and the reviewed manual-type set instead. Regeneration now reproduces the original 100-entry opaque digest exactly, so retain the previous approval rather than add an opaque entry. Nullable_GetAccountRateLimitsParams likewise uses an explicit pointer alias, matching the existing account-usage pattern. The complete-thread fixture now includes originator:null to reflect intentional nullable serialization. No exported protocol name is intentionally removed.

## Outcomes & Retrospective

Generated protocol/RPC output, compatibility adapters, regression tests, CLI checksums, and SDK version 0.154.0 are prepared, reviewed, committed, and pushed to `codex/update-protocol-v0.154.0`. The complete local updater passed, including deterministic generation and every quality gate; govulncheck reported no vulnerabilities. No exported protocol types were removed, the previous server-request aliases are preserved, and the opaque inventory remains unchanged. GitHub rejected PR creation with HTTP 403 (Resource not accessible by integration). No PR, merge, or release was created. Restore the integration's pull-request write access, retry creating a PR from the existing branch to main, and then follow hosted checks and the automatic Release workflow. Linear separately requires reauthentication.

## Context and Orientation

The root Go package exposes client, thread, turn, and callback APIs. `protocol/` contains generated wire data types and reviewed manual wrappers in `protocol/manual_types.go`; `rpc/` contains transport and generated method/notification dispatch. `internal/codegen/main.go`, invoked by `gen.go`, compiles the selected upstream CLI to export JSON schemas, generates types, and rejects unreviewed unions and opaque fields. A union describes alternative JSON shapes; an opaque field retains JSON for a schema without a safe concrete Go representation. Generated files must carry the exact upstream commit and exported names must have matching GoDoc comments.

`.github/codex/version` and `.github/codex/checksums.txt` pin the CLI and official archive hashes used by E2E. `.github/sdk-version` independently selects the published SDK version. `.github/workflows/release.yml` validates a protected-main candidate, runs Quality and trusted E2E, and alone creates the immutable tag.

## Plan of Work

Milestone 1 establishes a reproducible source/build environment and runs `.codex/skills/update-codex-protocol/scripts/update_codex_protocol.sh`. Its first generation may stop at inventory guards; retain evidence of the exact failure and review every added or changed schema before changing an approved digest. Compare 0.153.4 and 0.154.0 exported schemas where useful. Completion means all schema differences have explicit compatibility decisions.

Milestone 2 updates only affected handwritten adapters, regression tests, generated code, CLI metadata, and docs. Keep manual declarations in the generator's reviewed set and fix generation centrally rather than editing generated output. Fetch the four existing supported CLI archive digests from the official release. Test changed wire behavior, optional/null handling, and unknown-variant preservation. Run generation twice and the complete local gate; completion means deterministic output and green checks.

Milestone 3 reviews the complete diff, resolves independent review findings, stages reviewed paths by name, commits, pushes the feature branch, and opens a PR to main. Required hosted checks must pass before protected merge. Follow the resulting Release run through validation, Quality, E2E, and publication, then verify the remote annotated tag peels to the exact candidate. Stop at a required human review or unavailable permission rather than bypassing protection.

## Concrete Steps

Run from `/workspace/codex-sdk-go`, with Cargo on PATH and an upstream checkout at `/tmp/codex-protocol-upstream`:

    env -u CODEX_REPO_REF PATH=/home/codex/.cargo/bin:$PATH CODEX_REPO_ROOT=/tmp/codex-protocol-upstream CARGO_TARGET_DIR=/tmp/codex-protocol-target OPENSSL_INCLUDE_DIR=/tmp/codex-protocol-deps/root/usr/include OPENSSL_LIB_DIR=/tmp/codex-protocol-deps/root/usr/lib/x86_64-linux-gnu OPENSSL_STATIC=1 .codex/skills/update-codex-protocol/scripts/update_codex_protocol.sh --allow-dirty

The only dirty path initially allowed is this plan. Review and classify every subsequent dirty path before rerunning. On a changed union inventory, rerun the same generation with CODEX_PRINT_UNION_INVENTORY=1, compare sorted schema hashes and JSON with the previous release, document decisions, then update approvedUnionInventorySHA256. Separately review opaque output changes.

    go test ./internal/codegen ./protocol ./rpc .
    .github/scripts/validate-codex-metadata.sh
    .github/scripts/test-validate-codex-metadata.sh
    .github/scripts/test-install-codex-cli.sh
    .github/scripts/sdk-release-tag.sh
    .github/scripts/validate-release-tag.sh "$(.github/scripts/sdk-release-tag.sh)"

The complete updater also runs formatting, vet, all unit tests, race tests, Staticcheck v0.8.1, govulncheck v1.3.0, and git diff --check. Record PR and workflow URLs here as they become available.

## Validation and Acceptance

GeneratedCodexVersion must be 0.154.0 and every generated protocol/RPC header must identify `6b9826e3aa83b1a5947db50f4332cb9c65f1b340`. Two generations must be byte-identical, including new untracked output. New behavior must have meaningful regressions that fail before its implementation and pass afterward. All local checks and required hosted checks must pass, followed by successful trusted E2E and publication. Confirm remote v0.154.0 resolves to the exact protected-main release candidate.

## Idempotence and Recovery

Preserve unrelated work and inspect status before staging. Never stage all paths, create/move local SDK tags, push main, or bypass failed checks. Reuse the unpublished SDK version for a corrective release; retry Release only with the exact failed candidate or reviewed corrective merge SHA. If environment approval unexpectedly blocks execution, restore the skill's configured protected-branches/no-reviewers policy while preserving other protections; do not approve a pending job on the user's behalf. Failures must remain recorded with an exact recovery action.

## Artifacts and Notes

Live upstream tag lookup returned annotated object `36eab01061df3cde5f95ec20a526777b430091ba`, peeled commit `6b9826e3aa83b1a5947db50f4332cb9c65f1b340`. Initial Rust setup and clone logs are outside the repository under `/tmp/codex-protocol-*.log`.

GitHub branch metadata verifies main is protected, with required Coverage report, Go 1.26, and Go 1.27 checks enforced for everyone. Ruleset 20775618 prohibits deletion and non-fast-forward updates to v* tags without bypass actors. Public environment API verification shows e2e and release each have protected_branches=true, custom_branch_policies=false, and only a branch_policy rule; neither has a required reviewer. No GitHub setting was changed.

An export-name comparison found no removed protocol type names and 21 additions; all generated headers identify the selected upstream commit. The complete handwritten/generated diff was reviewed for API compatibility, null/omission behavior, approval handling, and release metadata. SDK tag resolution, version-transition fixtures, and release-tag-state fixtures pass. This review found and corrected the two alias-identity changes described above.

## Interfaces and Dependencies

Retain the existing go-jsonschema dependency and Go 1.26.8 minimum. Keep SDK behavior in the root package, transport in rpc, and data in protocol. Rust 1.95.0 is build tooling for exporting the exact upstream schema, not a new SDK runtime dependency. Determine affected exported APIs from the actual generated diff before specifying adapter changes.

Revision note: Created before generation after live source/remote inspection and tracker access checks.

Revision note: Recorded schema review, compatibility corrections, passing complete local gate, and live GitHub protection evidence before the protected PR handoff. Detailed full-gate output is `/tmp/codex-protocol-full-gate.log`.

Revision note: Recorded the successful branch push and exact GitHub PR permission blocker. The update skill requires stopping at missing GitHub permissions; do not use a direct-main push or an alternate publication path. The prepared comparison is https://github.com/pmenglund/codex-sdk-go/compare/main...codex/update-protocol-v0.154.0?expand=1.
