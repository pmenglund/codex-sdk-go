---
name: update-codex-protocol
description: Update codex-sdk-go from the latest stable upstream openai/codex Rust release, review compatibility, prove deterministic generation, run the complete local gate, and carry the reviewed change through a protected pull request to GitHub's E2E approval and automatic publication. Use for protocol refreshes or when preparing a new codex-sdk-go release from an upstream rust-vMAJOR.MINOR.PATCH tag.
---

# Update Codex Protocol

## Local update

Use this skill only from the `codex-sdk-go` repository root.

1. Inspect `git status --short --branch`, the active ExecPlan, and remote state. Preserve unrelated changes. For a new update, require clean `main` at `origin/main`, then create `codex/update-protocol-vMAJOR.MINOR.PATCH` before generation. If the invocation continues a reviewed dirty update, keep its files and create the branch before staging.
2. Treat `.envrc` as sensitive. Do not source or print it. The bundled script reads only `CODEX_REPO_ROOT` when the variable is absent.
3. Run `.codex/skills/update-codex-protocol/scripts/update_codex_protocol.sh`. Use `--allow-dirty` only after classifying every existing dirty path as part of this update.
4. Inspect the complete unstaged diff. Generated files under `protocol/` and `rpc/` must name the selected upstream commit. Review changed union and opaque-interface inventories before approving their digests.
5. Update handwritten SDK adapters, tests, documentation, CLI metadata, and the active ExecPlan for intentional compatibility changes. Rerun the updater until its deterministic generation and complete local gate pass.
6. Choose the next SDK version and write it without a `v` prefix to `.github/sdk-version`. Default to the generated Codex version only when that version is greater than the latest immutable SDK tag; otherwise choose a higher semantic version justified by the API change and record the decision. Run `.github/scripts/sdk-release-tag.sh` and `.github/scripts/validate-release-tag.sh "$(.github/scripts/sdk-release-tag.sh)"`.

The script fetches tags in the upstream checkout, chooses the highest tag matching exactly `rust-vMAJOR.MINOR.PATCH`, runs generation twice, and rejects non-deterministic output, including changes to newly created untracked generated files. It then runs formatting, vet, unit tests, race tests, Staticcheck v0.8.1, govulncheck v1.3.0, and `git diff --check`. The analyzers run through pinned `go run ...@version` commands; no separately installed analyzer binaries are required.

If generation reports a changed union inventory, rerun the same generation with
`CODEX_PRINT_UNION_INVENTORY=1`. Review the sorted per-schema hashes and JSON for
every added or changed `oneOf`/`anyOf`. Add typed generation or an explicit opaque
decision as needed, record the decision in the active ExecPlan, and only then
replace `approvedUnionInventorySHA256` with the reported aggregate digest.

Use `--allow-dirty` only after reviewing existing changes. The option does not stage or include files automatically; it only permits the quality run to proceed.

## Protected handoff

An explicit invocation authorizes the skill to prepare and publish the reviewed feature branch and pull request. After the local update passes:

1. Run applicable independent reviews. Address actionable findings and rerun affected gates.
2. Recheck `git status`, `git diff`, and `git diff --check`. Stage only the reviewed paths by name; never use `git add .`. Verify the staged diff contains no unrelated file or secret.
3. Commit in imperative mood, push the feature branch, and open a pull request to `main`. Include the upstream tag, SDK version, compatibility decisions, ExecPlan, and validation evidence.
4. Watch the required pull-request checks. Fix failures on the branch. When checks pass and branch protection permits, merge through the pull request and delete the remote feature branch. If a reviewer is required, leave the pull request pending for that review rather than bypassing it.
5. The merge of `.github/sdk-version` starts `.github/workflows/release.yml`. Find that run and report its URL. Hand control to the user when GitHub requests approval for `e2e`; resume monitoring if the user approves during the task. Publication must run automatically after E2E succeeds. Keep both environments restricted to protected branches, with required reviewers on `e2e` and none on `release`. If GitHub requests a second approval for `release`, report configuration drift rather than approving it on the user's behalf.
6. Confirm the completed run created the expected immutable tag from the merged `main` commit. Never create or move a local tag, push `main` directly, bypass a failed gate, or publish by another path.

The bundled updater remains deliberately non-publishing because generation may require handwritten compatibility work. The skill performs Git and pull-request operations only after reviewing the result. GitHub Actions alone creates and pushes the release tag. Manual Release dispatch requires the exact failed candidate SHA for an unchanged retry, or the exact reviewed corrective merge SHA after fixing a failed release. Retain the unpublished SDK version for a correction. The workflow still derives the SDK tag from the selected protected-main commit rather than accepting a tag input, and reruns every release gate.

If there is no generated or handwritten change, do not create a branch, commit, pull request, or release. If unrelated dirty files, an ambiguous SDK version, missing GitHub permissions, failed checks, or required review block the handoff, stop at that boundary and report the exact recovery action.
