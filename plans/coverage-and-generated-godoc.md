# Publish meaningful coverage and document generated protocol APIs

This ExecPlan is a living document. The sections `Progress`, `Surprises & Discoveries`, `Decision Log`, and `Outcomes & Retrospective` must be kept up to date as work proceeds.

This plan follows `PLANS.md` at the repository root. It is self-contained and covers the quality work requested after comparing this repository with the Awesome Go quality standard.

## Purpose / Big Picture

After this work, contributors and prospective users can open the README and reach both the package documentation and a hosted Codecov report. Continuous integration will publish a reproducible coverage profile and reject meaningful handwritten coverage below 80 percent. Protocol regeneration will also guarantee that every exported declaration in `protocol/*_gen.go` has a GoDoc-style comment beginning with the exact exported name, so a future Codex schema refresh cannot silently degrade pkg.go.dev documentation.

The work does not refresh the upstream Codex schema, change runtime wire behavior, add coverage-only product branches, or hand-edit generated declarations. Regeneration remains pinned to Codex `rust-v0.144.6`, version `0.144.6`, commit `5d1fbf26c43abc65a203928b2e31561cb039e06d`.

## Tracker Mapping

Workflow: `WORKFLOW.md`. By the repository-local tracking decision already recorded in the completed `plans/api-usability-hardening.md`, this ExecPlan is the parent tracker and no Linear issue is created. Its child tasks are the user's three numbered requests: publish meaningful coverage, generate Go-style comments and preserve the rule in `AGENTS.md`, and add pkg.go.dev plus coverage links to `README.md`.

## Progress

- [x] (2026-07-22 12:20Z) Inspected the clean current repository, completed plans, generator output paths, CI/release permissions, public Codecov state, and current coverage.
- [x] (2026-07-22 12:20Z) Completed the required independent planner pass and validated its recommendations against current action releases and workflow files.
- [x] (2026-07-22 12:20Z) Created branch `codex/awesome-go-quality` from local `main`, which was four commits ahead of `origin/main`.
- [x] (2026-07-22 12:17Z) Added failing generator tests for name-leading GoDoc on every exported generated protocol declaration, including preservation, invalid-source, idempotence, and committed-artifact cases.
- [x] (2026-07-22 12:17Z) Implemented centralized protocol-output documentation, regenerated from pinned `rust-v0.144.6`, and proved a second generation pass produces the same protocol diff hash.
- [x] (2026-07-22 12:17Z) Added Codecov configuration and SHA-pinned, OIDC-authenticated coverage publishing to CI and its reusable release caller.
- [x] (2026-07-22 12:17Z) Added the durable generator rule to `AGENTS.md`, documented the CI coverage policy in `WORKFLOW.md`, and added pkg.go.dev/Codecov badges to `README.md`.
- [x] (2026-07-22 12:22Z) Ran the full local quality gate and independent security, QA, and architecture reviews; pinned the Codecov CLI and added validator failure-path tests in response to review findings.

## Surprises & Discoveries

- Observation: The completed API-usability work already raised the root package above the Awesome Go threshold.
  Evidence: `go test -coverprofile=/tmp/codex-sdk-go-quality.cover ./...` reported 82.3 percent for the root package, 88.9 percent for `rpc`, and 81.9 percent for `internal/codegen`.
- Observation: The raw 58.2 percent aggregate is not a useful project gate because it counts generated protocol/RPC source and executable examples/test helpers as handwritten shipping logic.
  Evidence: Excluding only `**/*_gen.go`, `examples/**`, and `test/**` yields approximately 83.9 percent while retaining handwritten root, `rpc`, `protocol/manual_types.go`, and `internal/codegen` code.
- Observation: Codecov knows the public repository but has no active report yet.
  Evidence: Codecov's public repository listing returned `active:false`, `activated:false`, and the badge currently renders `unknown`; the first accepted CI upload must establish the report.
- Observation: The independent planner's download-artifact version was already one patch behind.
  Evidence: The current official release is `actions/download-artifact` v8.0.1 at commit `3e5f45b2cfb9172054b4087a40e8e0b5a5461e7c`, so this plan uses that verified SHA rather than v8.0.0.
- Observation: The completed implementation and review fixes improve the scoped coverage baseline.
  Evidence: The final atomic profile covers 2,296 of 2,727 in-scope statements, or 84.2 percent, under the exact validated Codecov exclusions.
- Observation: The local `govulncheck` invocation initially could not read the sandboxed shared Go build cache.
  Evidence: Repeating the identical pinned Go 1.25.12 scan with approved cache access completed successfully and reported no vulnerabilities.

## Decision Log

- Decision: Measure and gate handwritten implementation code, excluding generated `*_gen.go`, executable examples, and the credentialed/test-support tree.
  Rationale: Generated declarations are verified through deterministic generation and generator tests; examples and test helpers are not shipped library logic. The exclusions are explicit in `codecov.yml`, while handwritten protocol union logic remains in scope through `protocol/manual_types.go`.
  Date/Author: 2026-07-22 / Codex
- Decision: Set Codecov project and patch targets to 80 percent with zero tolerance below the target.
  Rationale: This directly enforces the Awesome Go threshold without turning the current 83.9 percent snapshot into an unnecessarily brittle hidden requirement.
  Date/Author: 2026-07-22 / Codex
- Decision: Generate the coverage profile in the secretless quality matrix, store it for one day, and upload it from a separate OIDC-only job.
  Rationale: The tests do not need `id-token: write`. Isolating Codecov authentication from test execution preserves the repository's least-privilege CI design and avoids a repository secret.
  Date/Author: 2026-07-22 / Codex
- Decision: Pin both the Codecov action commit and the downloaded Codecov CLI version, currently v11.3.1.
  Rationale: The action's default `latest` CLI would leave mutable runtime code executing in an OIDC-enabled job even though the action itself is immutable.
  Date/Author: 2026-07-22 / Codex
- Decision: Centralize protocol GoDoc repair and validation in the generator rather than editing each output template or generated file by hand.
  Rationale: `types_gen.go` comes from a third-party schema generator while aliases, fallbacks, metadata, compatibility declarations, and unions use repository templates. One final protocol-output path can preserve useful schema prose, prepend only missing name-leading sentences, and enforce the invariant uniformly.
  Date/Author: 2026-07-22 / Codex
- Decision: Apply the invariant only to exported top-level types, aliases, constants, variables, functions, and methods in `protocol/*_gen.go`.
  Rationale: These declarations form the package-level public API named by the user. Generated exported struct fields already receive field-specific comments from the schema generator; RPC cleanup is outside this request.
  Date/Author: 2026-07-22 / Codex

## Outcomes & Retrospective

All requested repository changes are implemented and locally verified. The final coverage profile reports 84.2 percent for handwritten implementation code under the validated `codecov.yml` exclusions. Both project and patch reports will be gated at 80 percent. The README links to pkg.go.dev and Codecov, and `WORKFLOW.md` explains what the published result measures.

Protocol generation now routes every generated protocol artifact through one AST-backed documentation and validation boundary. Tests cover preservation of existing schema and deprecation prose, aliases, constants, variables, functions, methods, idempotence, invalid source, multi-name failure, direct validator rejection, and every committed `protocol/*_gen.go` file. Two pinned `rust-v0.144.6` generation passes produced the same protocol diff hash, and generated metadata remained at version `0.144.6` and commit `5d1fbf26c43abc65a203928b2e31561cb039e06d`.

The local gate passed formatting, installer fixtures, release-tag and content-manifest tests, Actionlint, Codecov's public validator, vet, ordinary and race tests, Staticcheck v0.7.0, govulncheck v1.3.0 with patched Go 1.25.12, and `git diff --check`. Security review found that the SHA-pinned Codecov action would otherwise download a mutable latest CLI; pinning CLI v11.3.1 resolved it. QA review found an untested validator failure path; direct negative tests and the multi-export fail-closed test resolved it and raised validator coverage to 92.9 percent. Architecture review found no actionable issues and confirmed generated diffs are non-semantic.

One external acceptance step remains by design: after this branch is pushed, the first GitHub Actions run must complete the OIDC upload so Codecov activates the public report and replaces the current unknown badge. No token, push, tag, or release was performed locally.

## Context and Orientation

The repository's `internal/codegen/main.go` exports schemas from a local OpenAI Codex checkout, delegates ordinary model generation to `go-jsonschema`, and writes checked-in files under `protocol/` and `rpc/`. Protocol output is split across `types_gen.go`, `aliases_gen.go`, `fallback_gen.go`, `compatibility_gen.go`, `metadata_gen.go`, and `unions_gen.go`. The current writers format files but do not guarantee that exported declarations have name-leading GoDoc comments.

The secretless `.github/workflows/tests.yml` workflow runs formatting, installer fixtures, vet, tests, race tests, Staticcheck, and govulncheck on Go 1.25.12 and 1.26.5. `.github/workflows/release.yml` calls it as a reusable workflow. A reusable caller caps permissions granted to the called workflow, so the release caller must explicitly allow `id-token: write` for the quality call once Codecov uses GitHub OpenID Connect, abbreviated OIDC, to authenticate without a stored token.

Codecov consumes the standard Go coverage profile written by `go test -covermode=atomic -coverprofile=coverage.out ./...`. Root `codecov.yml` determines which paths are counted and defines the project and changed-line status thresholds. The README badge links to `https://codecov.io/gh/pmenglund/codex-sdk-go`; it reports `unknown` until a pushed CI run successfully uploads the first profile.

## Plan of Work

Begin in `internal/codegen/main_test.go` with fixtures covering an undocumented exported type, alias, grouped constant, top-level function, and method. The test must also prove that existing schema prose and `Deprecated:` paragraphs survive, unexported declarations remain untouched, the transform is idempotent, invalid Go returns a useful error, and every committed `protocol/*_gen.go` artifact satisfies the invariant.

In `internal/codegen/main.go`, parse generated Go with comments, inspect exported top-level declarations and exported methods, and prepend a concise name-leading sentence only where the current documentation does not begin with the exact name. Preserve the original comment group after the inserted sentence. Add a protocol-specific writer that applies the transformation, formats the result, validates the completed source, and is used for all six protocol generated files. Leave the RPC writer unchanged. Regenerate twice from `rust-v0.144.6`; the first diff must be comment-only and the second pass must be empty.

Add `codecov.yml` at the repository root. Ignore `**/*_gen.go`, `examples/**/*`, and `test/**/*`; set both project and patch coverage to 80 percent with zero threshold. In `.github/workflows/tests.yml`, make the Go 1.25.12 test step write `coverage.out`, upload it as a one-day artifact with `actions/upload-artifact` v7.0.1 commit `043fb46d1a93c77aae656e7c1c64a875d1fc6a0a`, and add a job that depends on the complete matrix. That job has only `contents: read` and `id-token: write`, downloads the artifact with `actions/download-artifact` v8.0.1 commit `3e5f45b2cfb9172054b4087a40e8e0b5a5461e7c`, and uploads through `codecov/codecov-action` v7.0.0 commit `fb8b3582c8e4def4969c97caa2f19720cb33a72f` with explicit file search disabled and CI failure enabled. Grant the reusable `quality` call in `.github/workflows/release.yml` the same OIDC permission cap.

Add the exact durable rule to `AGENTS.md`: every exported type, alias, constant, variable, function, and method emitted into `protocol/*_gen.go` must have a GoDoc comment beginning with its exact exported name; enforce it in `internal/codegen` and tests and never repair generated comments manually. Add pkg.go.dev and Codecov badges directly below the README title without hard-coding a percentage.

## Concrete Steps

Run all commands from `/Users/pme/src/pmenglund/codex-sdk-go`.

First run the focused generator test before and after implementation:

    go test ./internal/codegen

Regenerate twice from the pinned upstream ref:

    CODEX_REPO_ROOT=/Users/pme/src/openai/codex CODEX_REPO_REF=rust-v0.144.6 go generate ./...
    git diff -- protocol
    CODEX_REPO_ROOT=/Users/pme/src/openai/codex CODEX_REPO_REF=rust-v0.144.6 go generate ./...
    git diff --exit-code -- protocol

Produce the same profile CI will upload and inspect the raw package results:

    go test -covermode=atomic -coverprofile=/tmp/codex-sdk-go-quality.cover ./...
    go tool cover -func=/tmp/codex-sdk-go-quality.cover

Validate Codecov configuration with its public validator and validate workflows with the pinned project tool:

    curl --data-binary @codecov.yml https://codecov.io/validate
    go run github.com/rhysd/actionlint/cmd/actionlint@v1.7.7

Run the complete local gate:

    test -z "$(gofmt -l $(git ls-files '*.go'))"
    .github/scripts/test-install-codex-cli.sh
    .github/scripts/test-validate-release-tag.sh
    .github/scripts/test-content-manifest.sh
    go vet ./...
    go test ./...
    go test -race ./...
    staticcheck ./...
    GOTOOLCHAIN=go1.25.12 govulncheck ./...
    git diff --check

## Validation and Acceptance

The generated-artifact invariant must report zero undocumented exported declarations across every committed `protocol/*_gen.go`. A representative type, alias, constant, function, and method must have a comment whose first word is its exact identifier. Existing richer schema descriptions and deprecation paragraphs must remain visible. Running the pinned generator twice must leave no second-pass diff and must keep `protocol/metadata_gen.go` at Codex version `0.144.6` and source commit `5d1fbf26c43abc65a203928b2e31561cb039e06d`.

The local handwritten coverage calculation must remain at or above 80 percent under the exact three `codecov.yml` exclusions. The Codecov validator and Actionlint must accept their files. The README badges must resolve to pkg.go.dev and the public Codecov project. The first pushed CI run must upload `coverage.out`; the badge must stop displaying `unknown` before Codecov status contexts are considered for branch protection.

All repository tests, race tests, vet, Staticcheck, the patched Go 1.25.12 vulnerability scan, installer fixtures, and `git diff --check` must pass. Because the CI change introduces OIDC and downloaded third-party actions, run `review-security`; because coverage policy and generated-comment enforcement are regression-sensitive, run `review-qa`; because the generator gains a centralized output boundary, run `review-architecture`. Address actionable findings before the final commit.

## Idempotence and Recovery

Coverage generation, validation, and pinned code generation are safe to repeat. The generated-comment transform must be idempotent by test. If a generation pass changes declarations, JSON tags, method bodies, metadata, or the upstream commit, stop because the pinned schema boundary has been violated. If Codecov rejects OIDC before merge, keep the local coverage profile and configuration but do not weaken CI with a secret or mutable action tag; report the repository activation requirement. Published tags remain immutable and are outside this plan.

## Artifacts and Notes

Starting branch: `codex/awesome-go-quality`. Starting commit: `5165507`. Starting worktree: clean. Starting raw coverage: 58.2 percent aggregate. Final explicitly scoped handwritten coverage: 84.2 percent. Codecov is publicly discoverable but inactive, and its badge remains `unknown` until the first pushed upload.

No separate `record.md` is needed. Keep concise validation evidence, discoveries, and reviewer outcomes in this file.

## Interfaces and Dependencies

No runtime dependency is added. `internal/codegen` will gain a protocol-specific source transform and validator using only `go/ast`, `go/parser`, `go/token`, and existing `go/format` support. CI adds only immutable GitHub Action dependencies: upload-artifact v7.0.1, download-artifact v8.0.1, and codecov-action v7.0.0 at the exact SHAs recorded above.

## Revision Notes

- 2026-07-22: Created this ExecPlan after live repository inspection, an independent planner pass, current action-release verification, and a fresh coverage baseline. The plan chooses a direct 80 percent gate and isolates OIDC from test execution.
- 2026-07-22: Updated the plan after implementation, deterministic regeneration, Codecov validation, and the complete local quality gate. The measured in-scope result is 84.2 percent.
- 2026-07-22: Recorded final security, QA, and architecture outcomes. Review led to pinning Codecov CLI v11.3.1 and adding negative validator and multi-export regression tests; no findings remain.
