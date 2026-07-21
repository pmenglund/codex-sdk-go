---
name: update-codex-protocol
description: Update codex-sdk-go generated protocol and RPC artifacts from the latest stable upstream openai/codex Rust release, prove deterministic generation, and run the complete local quality gate. This workflow never publishes, tags, commits, stages, or pushes.
---

# Update Codex Protocol

## Workflow

Use this skill only from the `codex-sdk-go` repository root.

1. Check `git status --short --branch`. Preserve unrelated user changes. If unrelated files are dirty, stop before using `--allow-dirty`.
2. Treat `.envrc` as sensitive. Do not source or print it. The bundled script reads only `CODEX_REPO_ROOT` when the variable is absent.
3. Run `.codex/skills/update-codex-protocol/scripts/update_codex_protocol.sh`.
4. Inspect the unstaged generated diff. Generated files under `protocol/` and `rpc/` must name the upstream Codex commit in their headers.
5. If generation exposes an intentional compatibility change, update the handwritten SDK adapters and the active ExecPlan, then rerun the script.
6. Leave all changes unstaged for review. Release tags are created only by the protected GitHub release workflow.

The script fetches tags in the upstream checkout, chooses the highest tag matching exactly `rust-vMAJOR.MINOR.PATCH`, runs generation twice, and rejects non-deterministic output, including changes to newly created untracked generated files. It then runs formatting, vet, unit tests, race tests, Staticcheck v0.7.0, govulncheck v1.3.0, and `git diff --check`. The analyzers run through pinned `go run ...@version` commands; no separately installed analyzer binaries are required.

If generation reports a changed union inventory, rerun the same generation with
`CODEX_PRINT_UNION_INVENTORY=1`. Review the sorted per-schema hashes and JSON for
every added or changed `oneOf`/`anyOf`. Add typed generation or an explicit opaque
decision as needed, record the decision in the active ExecPlan, and only then
replace `approvedUnionInventorySHA256` with the reported aggregate digest.

Use `--allow-dirty` only after reviewing existing changes. The option does not stage or include files automatically; it only permits the quality run to proceed.

The removed `--tag-release`, `--push-release`, and implicit commit behavior are intentionally unsupported. Do not recreate tags or push from a local updater as a fallback.
