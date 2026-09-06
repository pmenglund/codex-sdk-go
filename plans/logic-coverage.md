# Cover untested RPC, schema, lifecycle, and process logic

This living ExecPlan follows PLANS.md. Maintain its progress, discoveries, decisions, and outcomes as work proceeds.

## Purpose / Big Picture

Protect behavior found untested during the coverage review: bounded outbound queues, canceled and late RPC requests, malformed input recovery, schema reference traversal, public thread lifecycle helpers, and forced subprocess shutdown. Tests must assert observable behavior, not merely execute lines. No production behavior changes are planned.

## Tracker Mapping

WORKFLOW.md governs this user-authorized task, recorded in plans/logic-coverage.md. Linear lookup on 2026-09-06 again required reauthentication; no parent or child issue identifiers are available. Proceed under the workflow's unavailable-tracker fallback.

## Progress

- [x] (2026-09-06) Reviewed implementation and existing tests; baseline coverage is 84.1% after repository exclusions.
- [x] (2026-09-06) Add RPC queue and reader recovery tests.
- [x] (2026-09-06) Add schema reference and inheritance fixtures.
- [x] (2026-09-06) Cover public lifecycle requests, validation, errors, and fork response IDs.
- [x] (2026-09-06) Cover forced process shutdown with a real child process.
- [x] (2026-09-06) Run validation, review the diff, record coverage changes, and commit.

## Surprises & Discoveries

The existing empty thread-ID test uses an uninitialized client and returns before reaching ID validation. Outbound capacity differs from server-request capacity and requires separate fixtures. Process shutdown uses a real two-second timeout.

## Decision Log

Decision (2026-09-06): Use testing/synctest for in-memory RPC coordination and replay tests; use a real self-hosted test subprocess for shutdown. Preserve the production timeout. Keep every new resource's cleanup local to its test.

Decision (2026-09-06): Assert semantic behavior and failure recovery, including subsequent traffic, exact errors, transmitted IDs/parameters, inherited schema fields, and child reaping. Do not add superficial coverage-only tests.

## Outcomes & Retrospective

All five coverage gaps are covered without changing production code. Filtered statement coverage increased from 84.1% to 87.2%. RPC client coverage increased from 89.4% to 92.9%, transport from 87.7% to 92.6%, lifecycle from 61.0% to 91.9%, and codegen from 82.6% to 85.0%. The targeted public by-ID lifecycle methods, schemaObjectRequired, schemaObjectProperties, and resolveSchemaReference now have 100% statement coverage. Linear remains unavailable pending reauthentication.

## Context and Orientation

rpc/client.go owns bounded write/request queues and a reader that ignores malformed or unmatched responses. rpc/transport.go closes child stdin then waits and kills after a timeout. internal/codegen/main.go collects schema fields through JSON references and composite schemas. thread_lifecycle.go exposes methods on Codex as well as Thread; prior tests primarily exercise Thread.

## Plan of Work

Milestone 1 adds rpc/client_coverage_test.go with a releasable, recording fake writer. Fill the real outbound queue while one write blocks. Test a caller canceled before enqueue, a canceled queued write skipped after release, async queue-full errors and terminal shutdown when busy replies cannot enqueue. Verify pending-call cleanup, late/unknown success and error responses, and malformed/blank messages followed by valid traffic. Use synctest.Wait plus state assertions and release all resources on failure.

Milestone 2 adds internal/codegen/schema_reference_test.go fixtures for nested and escaped references, missing/non-object targets, cycles, and inherited properties/required fields across composite schemas. Assert exact field sets and bounded traversal.

Milestone 3 adds thread_lifecycle_coverage_test.go replay tables for public by-ID methods, both IncludeTurns modes, exact request parameters, preserved server errors, fork response ID precedence/fallback/missing IDs, readiness and input validation. Repair the existing misleading ID test with an initialized RPC client and exact error assertion.

Milestone 4 adds rpc/transport_shutdown_test.go. Launch the current test binary as a helper child which signals readiness, consumes stdin to EOF, signals EOF, then remains alive. Close the transport and assert forced-kill error and reaped process state. Use a generous context watchdog and cleanup that kills/reaps if assertions fail, without altering production timeouts.

Milestone 5 runs the complete ordinary and race suites, repeated focused concurrency tests, vet, pinned static analyzers, installer fixtures, formatting and diff checks. Recompute coverage with the same exclusions and record results.

## Concrete Steps

From /workspace/codex-sdk-go with Go 1.26.8:

    go test ./...
    go test -race ./...
    go test -race -count=50 -cpu=1,2,4 ./rpc -run 'TestOutbound|TestQueued|TestReader'
    go vet ./...
    go run honnef.co/go/tools/cmd/staticcheck@v0.7.0 ./...
    go run golang.org/x/vuln/cmd/govulncheck@v1.3.0 ./...
    bash .github/scripts/test-install-codex-cli.sh
    go test -coverprofile=/tmp/codex-logic-coverage.out ./...
    git diff --check

## Validation and Acceptance

All tests must pass. Queued cancellation must not transmit canceled calls or poison subsequent traffic. Malformed/late messages must not strand unrelated calls. Queue saturation must produce the documented error. Schema field unions and reference decoding must match explicit expected sets. Lifecycle tests must check actual wire methods and values, not use the implementation's parameter builder for expected values. Shutdown must use its force-kill path and reap the child. Only tests and this plan should change unless a concrete bug requires an explicitly recorded adjustment.

## Idempotence and Recovery

Tests use local fixtures and temporary child processes and can be rerun safely. Register cleanup before assertions or blocking operations. Failure-path cleanup must release writers and handlers and kill/reap children. Do not alter real services or generated files.

## Artifacts and Notes

Baseline filtered coverage: 84.1%. Baseline thread_lifecycle.go: 61.0%; RPC client: 89.4%; schema reference resolver: 0%; forced-kill branch: uncovered.

## Interfaces and Dependencies

Use Go's standard testing, testing/synctest, os/exec, context, and existing replay transports. No dependencies or public interfaces change.

Revision note (2026-09-06): Created before implementation from the user-approved review findings.

Final validation (2026-09-06): Full ordinary and race suites passed, as did 50 focused concurrency repetitions at CPU counts 1, 2, and 4, vet, Staticcheck v0.7.0, govulncheck v1.3.0 (no vulnerabilities), installer fixtures, formatting, and diff checks. The real forced-shutdown test passed in ordinary and race runs with the unchanged production timeout.

Temporary compiler overlays independently disabled skipping canceled queued requests and reversed JSON-pointer escape decoding. The new tests rejected both regressions; the original production files were never modified. The first overlay check's output matcher expected a later assertion, but the earlier write-count assertion correctly rejected the regression; the diagnostic matcher was corrected and both checks passed.

Revision note (2026-09-06): Recorded completion, coverage measurements, and regression-sensitivity evidence. No implementation deviation or production fix was necessary.
