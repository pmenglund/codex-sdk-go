---
name: update-codex-protocol
description: Update codex-sdk-go generated protocol and RPC artifacts from the latest stable upstream openai/codex Rust release. Use when asked to pull upstream Codex tags, choose the latest non-alpha rust-vMAJOR.MINOR.PATCH tag, run go generate ./..., run Go tests, commit the regenerated SDK files, or create and push an SDK release tag for the generated protocol version.
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
6. For a release, rerun or run the script with `--tag-release` to create a local annotated SDK tag after validation. Use `--push-release` only when the user explicitly wants to publish; it implies `--tag-release` and pushes the current branch plus tag to `origin` atomically.

## Script Options

Use `--no-commit` to run fetch, generation, and tests but leave the diff uncommitted for inspection.

Use `--allow-dirty` only when the existing dirty files are intentional and should be included in the final commit. The default refusal on dirty worktrees protects user edits from accidental commits.

Use `--tag-release` to create or verify an annotated SDK release tag derived from the selected upstream Codex tag: `rust-vMAJOR.MINOR.PATCH` becomes `vMAJOR.MINOR.PATCH`. The script refuses to move an existing tag; if the tag already exists on `HEAD`, it reports success. If generation produces no repository diff, this option tags the current `HEAD` after validation. Do not combine this option with `--no-commit`.

Use `--push-release` to create or verify the release tag and push both the current branch and `refs/tags/<tag>` to `origin` with `git push --atomic`. Use it only when the user wants to publish the release. Do not combine this option with `--no-commit`.

## Expected Result

A successful run prints the selected stable upstream tag, completes `go generate ./...`, completes `go test ./...`, and either creates a commit named `Update Codex protocol from <tag>` or reports that there were no changes to commit. With `--tag-release`, it also leaves an annotated SDK tag such as `v0.142.5` on the release commit. With `--push-release`, the branch and SDK tag are pushed to `origin` atomically.
