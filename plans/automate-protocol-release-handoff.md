# Automate Protocol Release Handoff

This ExecPlan follows `PLANS.md` at the repository root. It extends the completed local work in `plans/update-codex-protocol-rust-v0.147.0.md` without weakening the deterministic generator or the protected GitHub release boundary.

## Purpose / Big Picture

A maintainer should be able to invoke the `update-codex-protocol` skill once. Codex will update and review the protocol, run the local gate, create a feature branch, commit only the reviewed files, push it, open a pull request, wait for required checks, and merge when branch protection permits. Merging an SDK version change to `main` will automatically start the protected GitHub Release workflow. The maintainer's remaining actions are GitHub environment approvals for credentialed E2E and tag publication.

The local updater script remains a deterministic, non-publishing primitive. The skill, which can respond to compatibility changes and inspect the complete diff, owns Git and pull-request orchestration. GitHub Actions remains the only component allowed to create or push an SDK release tag.

The user's direct request on 2026-08-12 to make one skill invocation perform as much of the release as possible is the tracked maintenance task. No Linear issue was supplied; this plan is the repository's durable task record.

## Progress

- [x] (2026-08-12) Inspected the current dirty 0.147.0 update, repository policy, skill, updater script, Quality, Trusted E2E, Release workflow, branch protection, environment protection, secret placement, and merge settings.
- [x] (2026-08-12) Chose a dedicated `.github/sdk-version` file so SDK release versions remain distinct from `.github/codex/version`, which pins the upstream Codex CLI.
- [x] (2026-08-12) Added deterministic SDK-tag resolution and valid/missing/malformed/multiline tests.
- [x] (2026-08-12) Made protected Release start when `.github/sdk-version` changes on `main`; removed the duplicate direct-push E2E trigger and retained manual retries.
- [x] (2026-08-12) Revised the skill, UI metadata, workflow documentation, and the 0.147.0 plan handoff.
- [x] (2026-08-12) Passed shell fixtures, official skill validation, Actionlint 1.7.7, focused Go tests, and the final authoritative deterministic updater with the complete repository gate.
- [x] (2026-08-12) Completed security, QA, and SRE reviews. Fixed canonical version parsing, required-check provenance, candidate-SHA races, idempotent annotated-tag verification, approval ordering/identity, mandatory retry identity, and overlapping release prevention. Retained the documented single-maintainer approval model because an independent reviewer does not yet exist.
- [x] (2026-08-12) Created `codex/update-protocol-v0.147.0` and committed the reviewed protocol update as `a731491` and release automation as `c37d76c`.
- [ ] Push the branch, open and merge the protected 0.147.0 pull request, then hand off the GitHub environment approvals.

## Surprises & Discoveries

- Observation: The SDK version cannot always be derived from the pinned Codex CLI version.
  Evidence: SDK tag `v0.145.0` intentionally shipped generated protocol and CLI metadata from Codex 0.144.6 after a source-breaking typed-union migration.

- Observation: Running both the current `Trusted E2E` push trigger and the reusable E2E gate in `Release` would require two credentialed runs and two approvals for one release.
  Evidence: `.github/workflows/e2e.yml` runs on every `main` push, while `.github/workflows/release.yml` calls the same workflow again.

- Observation: GitHub auto-merge is disabled, but branch protection requires no approving review.
  Evidence: Repository settings report `allow_auto_merge: false`; `main` requires a pull request and three Quality checks with zero approving reviews. The skill can wait for checks and issue the protected merge itself, while remaining compatible with future reviewer requirements.

- Observation: Shell command substitution discards trailing newlines, so a regex over its result alone cannot enforce an exact one-line version file.
  Evidence: QA review demonstrated that `0.147.0` followed by a blank line resolved to `v0.147.0`, and that numeric regex components accepted `00.147.0`. The resolver now checks exactly one input line and canonical numeric components; both helper fixture suites cover these cases.

- Observation: The repository is operated by one GitHub maintainer, so an independent human approval requirement would make the current release path unusable.
  Evidence: `pmenglund` is the only configured reviewer for both protected environments. Security review correctly identified the residual risk: the maintainer who merges can also approve the release. The skill never approves an environment deployment; it stops for the user at both GitHub gates. Required Quality checks are now bound to GitHub Actions app ID 15368 rather than accepting unbound status contexts.

- Observation: Requiring the gated commit to remain the head of `main` makes an approval queue fragile, and rejecting every existing tag makes an ambiguous successful push unrecoverable.
  Evidence: SRE review found that an unrelated merge during environment approval would strand the path-filtered run, while a remotely accepted tag push followed by lost runner status would make retries fail. Release now carries an explicit candidate SHA through Quality, E2E, and publication; it accepts that immutable SHA while it remains an ancestor of protected `main`, treats the expected tag at that SHA as success, and fails closed on a conflicting tag.

- Observation: GitHub concurrency retains at most one pending run per group even when `cancel-in-progress` is false; it has no supported unlimited-queue setting.
  Evidence: A second pending SDK release could be replaced by a third while the first awaits approval. The protected PR Quality gate now rejects a new `.github/sdk-version` bump until the previous declared version has an annotated tag in the base branch history, so another release candidate cannot merge into `main` during that window.

- Observation: The first hosted pilot started duplicate Quality runs for the feature branch and pull request.
  Evidence: PR #3 showed two pending copies of each Go matrix job because `tests.yml` still used an unrestricted `push` trigger alongside `pull_request`. Push-triggered Quality is now limited to `main`; pull requests and reusable release candidates each run once.

## Decision Log

- Decision: Add `.github/sdk-version` as the authoritative SDK tag version.
  Rationale: `.github/codex/version` describes the executable used by E2E and can legitimately differ from the SDK's semantic version. A dedicated file makes automatic release intent explicit and provides a narrow `push.paths` trigger.
  Date/Author: 2026-08-12 / Codex

- Decision: Trigger `Release` on protected `main` pushes that change `.github/sdk-version`, with manual dispatch retained as a retry path.
  Rationale: A path-filtered trigger avoids releases for unrelated merges. Deriving the tag from the reviewed file removes a second command and prevents a manual input from disagreeing with the merged release intent.
  Date/Author: 2026-08-12 / Codex

- Decision: Remove the direct `main` push trigger from `Trusted E2E` and call it once from `Release`.
  Rationale: Release remains gated by credentialed E2E, but a single release needs one E2E approval and run rather than two. Manual Trusted E2E dispatch remains available for diagnosis.
  Date/Author: 2026-08-12 / Codex

- Decision: Keep `update_codex_protocol.sh` unable to stage, commit, push, merge, or tag; put reviewed Git and pull-request steps in `SKILL.md`.
  Rationale: The shell script cannot resolve schema compatibility or distinguish handwritten fixes from unrelated changes. The skill can review those boundaries before naming paths for staging. GitHub Actions alone creates the immutable tag.
  Date/Author: 2026-08-12 / Codex

- Decision: Retain a single-maintainer approval model while keeping GitHub environment approval outside the skill.
  Rationale: Requiring a distinct reviewer or preventing self-review would block every release until another trusted maintainer is added. The explicit protected-environment click remains a human checkpoint, but it is not an independent-party review. If a second maintainer is added, enable one required PR approval, stale-review dismissal, last-push approval, code-owner review for release files, environment self-review prevention, and disable administrator bypass.
  Date/Author: 2026-08-12 / Codex

## Outcomes & Retrospective

The implementation and local validation are complete. The official skill validator, all release and transition fixtures, shell syntax checks, Actionlint 1.7.7, focused Go packages, two deterministic exact-0.147.0 generations, formatting, metadata and installer fixtures, vet, all tests, race tests, Staticcheck 0.7.0, govulncheck 1.3.0, and `git diff --check` pass. Independent QA and SRE re-reviews found no unresolved correctness or operability issue after fixes. Security review's unbound-check finding was fixed by binding all three required checks to GitHub Actions app ID 15368. Its independent-human-review recommendation is intentionally deferred until a second trusted maintainer exists; the skill cannot approve GitHub environments and stops for the user's approval.

The reviewed changes are committed on `codex/update-protocol-v0.147.0` as `a731491` and `c37d76c`. No push, merge, or release has yet been made under this plan. Hosted PR checks, the automatic Release trigger, protected E2E, release approval, and final tag identity remain the pilot acceptance steps.

## Context and Orientation

`.codex/skills/update-codex-protocol/SKILL.md` tells Codex how to perform the update. Its bundled `scripts/update_codex_protocol.sh` selects the newest exact stable upstream Rust tag, generates twice, and runs the complete local gate. `.github/workflows/tests.yml` supplies required pull-request checks. `.github/workflows/e2e.yml` owns credentialed tests behind the `e2e` environment. `.github/workflows/release.yml` validates and pushes immutable annotated tags behind the `release` environment.

The current checkout is `main` at the same commit as `origin/main`, with the reviewed 0.147.0 protocol update unstaged. Those files are user-owned task work and must not be reset or discarded. This plan and its workflow changes will join that update in the same release pull request.

## Plan of Work

Add `.github/sdk-version` and a small script that converts its exact semantic version into `vMAJOR.MINOR.PATCH`. Test valid, missing, whitespace, and malformed inputs. Keep the existing release-tag validator responsible for checking the minimum, immutability, and increasing-order rules.

Change `Release` to run on a protected `main` push only when `.github/sdk-version` changes, or on manual retry that requires the exact merged candidate SHA from the failed run. Resolve the tag from that commit, validate it before costly gates, carry the SHA through reusable Quality and Trusted E2E, require it to remain an ancestor of protected `main`, then create and verify only the annotated tag. Verify both the remote tag-object identity and peeled commit for new and idempotent publication. Make E2E wait for successful Quality. Remove `push` from `Trusted E2E`; retain `workflow_dispatch` and `workflow_call`.

Revise `SKILL.md` so explicit invocation authorizes a release branch, reviewed staging, commit, push, pull request, check monitoring, and protected merge. Stop for unrelated dirty files, failed checks, required review, or GitHub environment approval. Never push `main`, create a local tag, bypass protection, or stage with `git add .`. Update `agents/openai.yaml` and `WORKFLOW.md` to describe the same boundary.

Validate shell syntax and fixtures, run Actionlint, run the skill validator, run focused Go tests, then rerun the authoritative updater and complete repository gate. Independent reviewers must inspect credential flow, trigger safety, failure recovery, and user handoff before staging.

## Concrete Steps

Work from `/Users/pme/src/pmenglund/codex-sdk-go`.

    bash -n .github/scripts/sdk-release-tag.sh
    .github/scripts/test-sdk-release-tag.sh
    .github/scripts/test-validate-release-tag.sh
    go run github.com/rhysd/actionlint/cmd/actionlint@v1.7.7
    python3 /Users/pme/.codex/skills/.system/skill-creator/scripts/quick_validate.py .codex/skills/update-codex-protocol
    GOTOOLCHAIN=go1.26.5 go test . ./protocol ./rpc ./internal/codegen
    GOTOOLCHAIN=go1.26.5 .codex/skills/update-codex-protocol/scripts/update_codex_protocol.sh --allow-dirty
    git diff --check

After all review findings are resolved, stage only the paths named by `git status --short`, create `codex/update-protocol-v0.147.0`, commit, push, open the pull request, wait for required checks, and merge through branch protection. Do not create a local tag. Confirm that the automatic Release run reaches the protected `e2e` approval before handing control to the user.

## Validation and Acceptance

The SDK-tag helper must output `v0.147.0` for the committed version file and reject missing, multiline, whitespace-padded, leading-zero, prerelease, or otherwise malformed versions. The tag validator must still reject a decreasing, malformed, or pre-migration tag. Release-state tests must accept an absent valid tag and an existing annotated tag at the expected commit, while rejecting a conflicting or lightweight tag.

Actionlint must accept all workflows. A pull request must continue to receive the three required Quality checks. A `main` merge that does not change `.github/sdk-version` must not start Release. A merge that changes it must run Quality and one protected E2E invocation before the protected tag job becomes reachable. Manual dispatch must retry the same merged SDK version rather than accept an arbitrary tag input.

The skill must leave no command for the maintainer between invocation and GitHub approval unless a compatibility decision, failed check, required PR review, or permission boundary genuinely blocks progress. The local updater remains non-publishing, `main` is never pushed directly, and only the Release workflow creates `v0.147.0`.

## Idempotence and Recovery

The SDK-tag helper and local quality commands are safe to rerun. If publication fails before the remote accepts the tag, retry the exact candidate SHA by manual dispatch. If the remote accepted the tag but the runner lost its result, the retry succeeds only when the annotated tag resolves to that candidate; a conflicting tag fails closed. If the PR checks fail, fix the feature branch and push another commit. An unrelated later `main` merge does not invalidate a candidate that already passed gates, provided the candidate remains in protected `main` history. A subsequent SDK-version PR cannot merge until the prior version is tagged.

If unrelated working-tree files appear, stop before staging. Never use a reset or checkout to remove them. If a GitHub reviewer or protected environment approval is required, report the exact URL and leave the external state pending.

## Interfaces and Dependencies

`.github/sdk-version` contains one unprefixed semantic version. `.github/scripts/sdk-release-tag.sh [REPOSITORY]` prints the corresponding `vMAJOR.MINOR.PATCH` tag. Release continues to depend on the repository `OPENAI_API_KEY`, the `e2e` and `release` environments, the three Quality checks, and GitHub's scoped workflow token. No runtime Go dependency is added.

Revision note: Created on 2026-08-12 after the user requested a one-invocation protocol update and protected release handoff.
