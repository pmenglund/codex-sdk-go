---
name: update-codex-protocol
description: Update codex-sdk-go generated protocol and RPC artifacts from the latest stable upstream openai/codex Rust release. Use when asked to pull upstream Codex tags, choose the latest non-alpha rust-vMAJOR.MINOR.PATCH tag, run go generate ./..., run Go tests, and commit the regenerated SDK files.
---

# Update Codex Protocol

## Workflow

Use this skill only from the `codex-sdk-go` repository root.

1. Check `git status --short --branch` before doing anything. Preserve unrelated user changes. If the worktree is dirty from unrelated files, stop and ask before running the script with `--allow-dirty`.
2. Treat `.envrc` as sensitive. Do not source it and do not print its full contents. The bundled script reads only `CODEX_REPO_ROOT` from `.envrc` when the variable is not already present in the environment.
3. Run the update script:

       .codex/skills/update-codex-protocol/scripts/update_codex_protocol.sh

   The script fetches tags in the upstream Codex checkout, selects the highest tag matching exactly `rust-vMAJOR.MINOR.PATCH`, runs `CODEX_REPO_REF=<tag> go generate ./...`, runs `go test ./...`, and commits the resulting repository changes if validation passes.

4. Inspect the generated diff or commit. Generated files under `protocol/` and `rpc/` should have headers naming the upstream Codex commit used by generation.
5. If generation exposes a new protocol shape that conflicts with existing manual SDK types, update `internal/codegen/main.go` `manualProtocolTypes()` or the relevant hand-written SDK code, then rerun the script or rerun `go generate ./...` and `go test ./...`.

## Script Options

Use `--no-commit` to run fetch, generation, and tests but leave the diff uncommitted for inspection.

Use `--allow-dirty` only when the existing dirty files are intentional and should be included in the final commit. The default refusal on dirty worktrees protects user edits from accidental commits.

## Expected Result

A successful run prints the selected stable upstream tag, completes `go generate ./...`, completes `go test ./...`, and either creates a commit named `Update Codex protocol from <tag>` or reports that there were no changes to commit.
