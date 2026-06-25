# Create a repo-local Codex protocol update skill

This ExecPlan is a living document. The sections `Progress`, `Surprises & Discoveries`, `Decision Log`, and `Outcomes & Retrospective` must be kept up to date as work proceeds.

This document follows `PLANS.md` in this repository. It is intended to be self-contained for a contributor who only has this working tree.

## Purpose / Big Picture

This repository generates Go protocol and RPC files from the upstream `openai/codex` Rust app-server protocol. Today a contributor has to remember how to fetch tags, choose a stable Rust release tag, set `CODEX_REPO_REF`, run `go generate ./...`, test, and commit the result. After this change, a repo-local Codex skill will encode that workflow and provide a script that performs the fragile pieces consistently.

The behavior is working when a future Codex run can use `.codex/skills/update-codex-protocol/SKILL.md` to update generated files from the latest stable `rust-vX.Y.Z` tag in the checkout named by `CODEX_REPO_ROOT`, run `go test ./...`, and commit only after validation passes.

## Tracker Mapping

Workflow: `WORKFLOW.md`. Linear issue: none. The user explicitly approved doing this maintenance task without a Linear issue on 2026-06-25. This plan file is the tracked record for the work.

## Progress

- [x] (2026-06-25 10:58+02:00) Confirmed the requested skill location is `.codex/skills/update-codex-protocol`, that a script should be included, that stable tags are exactly `rust-vMAJOR.MINOR.PATCH`, that commits should happen after tests pass, and that no Linear issue is needed.
- [x] (2026-06-25 10:58+02:00) Checked `git status --short --branch`; the worktree started clean on `main...origin/main`.
- [x] (2026-06-25 10:58+02:00) Read `AGENTS.md`, `APP.md`, `LANGUAGE.md`, `WORKFLOW.md`, and `PLANS.md`.
- [x] (2026-06-25 11:06+02:00) Created `.codex/skills/update-codex-protocol` with `SKILL.md`, `agents/openai.yaml`, and `scripts/update_codex_protocol.sh`.
- [x] (2026-06-25 11:12+02:00) Validated script syntax with `bash -n`, script help output, and YAML parseability with Ruby. The official `quick_validate.py` could not run because both available Python runtimes lacked the `yaml` module.
- [x] (2026-06-25 11:39+02:00) Ran the skill workflow against the current `CODEX_REPO_ROOT` checkout. It selected `rust-v0.142.2`, generated protocol files from Codex commit `390b0d254d658148751d0cca50ca41832c7894a1`, and exposed required hand-written compatibility updates.
- [x] (2026-06-25 11:39+02:00) Ran `go test ./...`; all packages passed after adding manual params structs and generator overrides.
- [x] (2026-06-25 11:40+02:00) Commit the plan, skill, script, and generated changes after validation passes.

## Surprises & Discoveries

- Observation: The current shell environment did not expose `CODEX_REPO_ROOT`, but repository `.envrc` defines it as `${HOME}/src/openai/codex` and also pins `CODEX_REPO_REF` to `rust-v0.142.1`.
  Evidence: `printenv CODEX_REPO_ROOT` returned no value; reading `.envrc` showed the Codex repository root and release ref. The `.envrc` also contains an API key, so this plan and later logs must not quote the file contents beyond non-secret variable names and non-secret tag values.
- Observation: The repository already supports generating from a specific upstream Codex ref without changing the upstream checkout.
  Evidence: `APP.md` documents `CODEX_REPO_ROOT=../codex CODEX_REPO_REF=<tag> go generate ./...`, and `internal/codegen/main.go` defines `CODEX_REPO_ROOT` and `CODEX_REPO_REF`.
- Observation: The skill-creator validator script is present but cannot run in this environment because PyYAML is unavailable.
  Evidence: `python3 /Users/pme/.codex/skills/.system/skill-creator/scripts/quick_validate.py .codex/skills/update-codex-protocol` and the bundled Python runtime both failed with `ModuleNotFoundError: No module named 'yaml'`. Ruby's standard YAML parser successfully parsed `SKILL.md` frontmatter and `agents/openai.yaml`.
- Observation: Plain `git fetch --tags` failed in the upstream Codex checkout because an unrelated existing local tag would be clobbered by the remote.
  Evidence: The first script run stopped at fetch with `! [rejected] rusty-v8-v147.4.0 -> rusty-v8-v147.4.0 (would clobber existing tag)`.
- Observation: `rust-v0.142.2` added no-params client methods whose response names do not match the generator's existing naming inference.
  Evidence: `go generate ./...` failed until `account/workspaceMessages/read` was mapped to `GetWorkspaceMessagesResponse` and `externalAgentConfig/import/readHistories` was mapped to `ExternalAgentConfigImportHistoriesReadResponse`.
- Observation: `rust-v0.142.2` generated `ThreadGoal` from nested schemas and also left several SDK request params as fallback interfaces.
  Evidence: `go test ./...` initially failed with `ThreadGoal redeclared`, then with invalid composite literals for `ThreadStartParams`, `ThreadResumeParams`, `ThreadForkParams`, and `TurnStartParams`.

## Decision Log

- Decision: Create the repo-local skill at `.codex/skills/update-codex-protocol`.
  Rationale: The user accepted this default, and no repo-local skill convention exists in the checkout yet.
  Date/Author: 2026-06-25 / Codex
- Decision: Include a deterministic shell script in the skill.
  Rationale: Fetching upstream tags, excluding prereleases, selecting the highest semantic version, running generation, running tests, and committing are repeatable steps where a script reduces mistakes.
  Date/Author: 2026-06-25 / Codex
- Decision: Treat a stable upstream release tag as exactly `rust-vMAJOR.MINOR.PATCH`.
  Rationale: The user confirmed this rule. Tags containing prerelease suffixes such as `-alpha`, `-beta`, or `-rc` are excluded because they do not match the exact stable pattern.
  Date/Author: 2026-06-25 / Codex
- Decision: Commit automatically after validation passes.
  Rationale: The user confirmed that the workflow should commit changes after `go test ./...` passes.
  Date/Author: 2026-06-25 / Codex
- Decision: Run the newly created script for this bootstrap pass with `--allow-dirty --no-commit`, then commit explicitly after inspection.
  Rationale: The dirty files are the plan and skill being created in this same task. Running with `--no-commit` proves fetch, generation, and tests before creating a clear final commit that includes both the skill and any generated artifacts.
  Date/Author: 2026-06-25 / Codex
- Decision: Use `git fetch --tags --force` for the upstream Codex tag refresh.
  Rationale: The workflow is explicitly about using the latest upstream tags. A stale or locally divergent tag in the upstream checkout should be refreshed rather than blocking unrelated protocol generation.
  Date/Author: 2026-06-25 / Codex
- Decision: Filter generated declarations for names in `manualProtocolTypes()` after schema generation, rather than skipping entire schema files.
  Rationale: Skipping whole schema files removed sanitized helper aliases needed by existing manual protocol types. Filtering declarations preserves generated helpers while preventing duplicate hand-written public types.
  Date/Author: 2026-06-25 / Codex
- Decision: Maintain concrete request params for thread start/resume/fork and turn start in `protocol/manual_types.go`.
  Rationale: The high-level SDK builds these params as structs. Upstream schemas currently fall back to `interface{}` for these shapes, so manual structs preserve the SDK API and keep tests compiling.
  Date/Author: 2026-06-25 / Codex

## Outcomes & Retrospective

Created a repo-local skill at `.codex/skills/update-codex-protocol` with an executable script that fetches upstream Codex tags, selects the highest stable `rust-vMAJOR.MINOR.PATCH` tag, runs generation with `CODEX_REPO_REF`, runs `go test ./...`, and commits when validation passes.

Regenerated protocol and RPC artifacts from upstream `rust-v0.142.2`, Codex commit `390b0d254d658148751d0cca50ca41832c7894a1`. The generator now handles two additional response-name overrides and filters generated declarations for hand-written manual protocol types. The SDK keeps concrete manual params structs for high-level request builders whose upstream schemas still exceed the generator's capabilities.

Validation completed with `bash -n .codex/skills/update-codex-protocol/scripts/update_codex_protocol.sh`, script help output, Ruby YAML parsing for the skill files, `go test ./internal/codegen`, `go generate ./...` through the script, `go test ./...`, and `git diff --check`. The official `quick_validate.py` could not run because PyYAML is not installed in the available Python runtimes.

## Context and Orientation

`codex-sdk-go` is a Go SDK for the Codex app-server protocol. The generated wire-level protocol files live under `protocol/`, and generated JSON-RPC stubs live under `rpc/`. The generator entrypoint is `gen.go`, which runs `go run ./internal/codegen` when `go generate ./...` is invoked. The generator exports JSON schemas from an upstream `openai/codex` checkout by running Cargo inside that checkout, then rewrites checked-in Go files.

The upstream Codex checkout is selected by the `CODEX_REPO_ROOT` environment variable when set, otherwise by the repository default `../codex`. The upstream ref is selected by `CODEX_REPO_REF`; when set, `internal/codegen` creates a temporary detached worktree at that ref rather than changing the main upstream checkout. That means the update script should set `CODEX_REPO_REF` to the chosen tag and should not check out or reset the upstream repository.

A repo-local skill is a directory under `.codex/skills/<skill-name>/` containing a required `SKILL.md`. Codex discovers skill metadata from `SKILL.md` and loads the body when a user request matches the description. This skill will be named `update-codex-protocol`.

## Plan of Work

First, create the skill directory with the skill-creator initializer so the required files and metadata are well formed. Then replace the generated template content with concise instructions specific to this repository. The `SKILL.md` body will instruct future Codex runs to check `git status`, avoid touching unrelated user changes, treat `.envrc` as sensitive if present, run the bundled script, inspect the generated diff, and only commit after tests pass.

Second, add `scripts/update_codex_protocol.sh`. The script will run from the SDK repository root, locate the upstream Codex repository from `CODEX_REPO_ROOT` or `.envrc`, fetch tags in the upstream repository, list tags matching exactly `rust-v[0-9]+.[0-9]+.[0-9]+`, sort them by semantic version, choose the highest one, and run `CODEX_REPO_REF=<tag> go generate ./...`. It will then run `go test ./...`. If validation succeeds, it will commit the current repository changes with a message naming the selected upstream tag. The script will refuse to run when the SDK worktree has pre-existing changes unless the caller passes an explicit option allowing it, because committing generated changes on top of unrelated user edits would be unsafe.

Third, validate the skill folder using the skill-creator quick validator, and run the script for real from `/Users/pme/src/pmenglund/codex-sdk-go`. Because fetching tags and running Cargo in `/Users/pme/src/openai/codex` need network and access outside this repository, those commands require elevated execution.

Fourth, inspect the generated diff. If generation introduces new protocol shapes that conflict with manual SDK types, update `internal/codegen/main.go` `manualProtocolTypes()` or the hand-written SDK routing code only as needed, following the local generator patterns. Then rerun `go generate ./...` and `go test ./...`.

Finally, after tests pass, commit the plan, skill, script, and generated files. The commit message will be imperative and identify the upstream tag, for example `Update Codex protocol from rust-v0.142.1`.

## Concrete Steps

Run these commands from `/Users/pme/src/pmenglund/codex-sdk-go` unless another working directory is named.

Create and edit the skill:

    /Users/pme/.codex/skills/.system/skill-creator/scripts/init_skill.py update-codex-protocol --path .codex/skills --resources scripts --interface display_name='Update Codex Protocol' --interface short_description='Regenerate codex-sdk-go protocol files from the latest stable upstream Codex Rust release.' --interface default_prompt='Update the codex-sdk-go generated protocol files from the latest stable openai/codex rust-v release and commit after tests pass.'

Validate the skill:

    /Users/pme/.codex/skills/.system/skill-creator/scripts/quick_validate.py .codex/skills/update-codex-protocol

Run the update workflow:

    .codex/skills/update-codex-protocol/scripts/update_codex_protocol.sh

The update script should print the selected tag, run `go generate ./...`, run `go test ./...`, and create one commit if there are changes.

## Validation and Acceptance

The skill is accepted when `.codex/skills/update-codex-protocol/SKILL.md` has valid frontmatter, `.codex/skills/update-codex-protocol/agents/openai.yaml` is present and aligned with the skill, and the bundled script can be run from the SDK repository root.

The update workflow is accepted when the script fetches upstream tags from `CODEX_REPO_ROOT`, selects the highest stable tag matching exactly `rust-vMAJOR.MINOR.PATCH`, runs `go generate ./...`, and runs `go test ./...` successfully. If generation changes checked-in files, those files must be committed only after tests pass. If generation produces no diff, the script should report that there is nothing to commit rather than creating an empty commit.

## Idempotence and Recovery

The script should be safe to rerun. It fetches tags without changing the upstream checkout, uses `CODEX_REPO_REF` so the existing generator creates a temporary worktree, and refuses to start from a dirty SDK worktree by default. If `go generate ./...` fails, fix the generator or manual type handling and rerun the script. If `go test ./...` fails after generation, do not commit; inspect the failing tests, update hand-written code or tests as needed, and rerun the script or the relevant commands.

If a generated diff is wrong, use normal Git inspection to identify the affected files. Do not reset or discard user changes without explicit user approval.

## Artifacts and Notes

The memory system records that previous Codex protocol updates may require adding names to `internal/codegen/main.go` `manualProtocolTypes()` and that generated `thread/goal/*` notifications previously required manual thread-scoped handling in `turn.go`. Treat generation as the first step, then inspect the diff before adding hand-written code.

## Interfaces and Dependencies

The script depends on local `git`, `go`, and `cargo`, plus an upstream `openai/codex` checkout available at `CODEX_REPO_ROOT` or discoverable from `.envrc`. It must not depend on external Go packages or network services beyond `git fetch --tags` in the upstream Codex repository.

The script interface is:

    update_codex_protocol.sh [--allow-dirty] [--no-commit]

`--allow-dirty` permits running when the SDK worktree already has changes, but the skill should avoid using this flag unless the user explicitly asks for it. `--no-commit` runs fetch, generation, and tests but leaves the diff uncommitted for inspection.

## Revision Notes

- 2026-06-25: Initial ExecPlan created after the user confirmed the skill path, script inclusion, stable tag rule, automatic commit policy, and no Linear issue requirement.
- 2026-06-25: Updated after implementation to record the selected upstream tag, generator compatibility fixes, validation results, and the PyYAML validator limitation.
