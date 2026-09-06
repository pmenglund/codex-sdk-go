# Make concurrency tests deterministic with testing/synctest

This living ExecPlan follows PLANS.md. Maintain Progress, Surprises & Discoveries, Decision Log, and Outcomes & Retrospective throughout implementation.

## Purpose / Big Picture

Replace wall-clock timing in in-memory concurrency tests with virtual time and explicit synchronization. Tests should verify deadline boundaries and concurrency behavior without scheduler-dependent sleeps or polling. Production behavior and public APIs remain unchanged.

## Tracker Mapping

Workflow: WORKFLOW.md. This user-approved task is recorded in plans/testing-synctest.md. No parent epic or child issue identifiers are available: Linear lookup on 2026-09-06 returned reauthentication required. Proceed under WORKFLOW.md's unavailable-tracker fallback and record the gap in delivery notes.

## Progress

- [x] (2026-09-06) Inspected tests and transports; root and RPC baseline tests pass.
- [x] (2026-09-06) Migrated RPC timing tests and all shared waiting-helper callers.
- [x] (2026-09-06) Migrated turn deadlines, cleanup, callbacks, and replay coordination.
- [x] (2026-09-06) Passed repeated race tests and complete local validation; reviewed changes for the delivery commit.

## Surprises & Discoveries

ReplayTransport blocks on sync.Cond and fake transports block on channels, so both can participate in synctest. Client.Close does not join goroutines; bubble exit additionally checks their termination.

## Decision Log

Decision (2026-09-06): Implement the user's broader migration, including timer-only hang guards in in-memory tests. Keep subprocess and credentialed E2E tests unchanged because external I/O cannot be synchronized through a bubble.

Decision (2026-09-06): Preserve production timeouts and test exact virtual boundaries. Do not mutate the global cleanup timeout or introduce production clock injection.

## Outcomes & Retrospective

Migration is complete across five test files. The final 50-repeat race matrix passed at CPU counts 1, 2, and 4 (root package 23.298s, RPC 9.381s). Full ordinary and race suites, vet, Staticcheck v0.7.0, govulncheck v1.3.0, installer fixtures, formatting, and diff checks passed. Production and generated files, dependencies, subprocess tests, and E2E tests are unchanged. Linear remains unavailable due to reauthentication; no tracker status was changed.

## Context and Orientation

The root package controls turns and callbacks; rpc implements concurrent JSON-RPC reads, writes, notifications, and request workers. rpc/client_test.go contains channel-based fake transports and polling helpers. turn_handle_test.go and turn_test.go exercise cancellation and deadlines. request_callbacks_test.go and rpc/message_test.go contain callback and replay hang guards.

A synctest bubble is the isolated group of goroutines created by synctest.Test. Time advances virtually when all goroutines block. synctest.Wait waits for other goroutines to finish or block, but assertions must still verify the expected result. Construct all mutable fixtures inside the bubble and register cleanup there; release fake handlers and close clients even on failure.

## Plan of Work

Milestone 1 migrates RPC tests with time-based guards and all callers of waitForReads/waitForWrites. Wrap individual tests or table subtests in synctest.Test, replace timers with Wait and nonblocking assertions, and replace polling helpers with Wait plus count checks. Preserve concurrency in publish/close race tests and virtual sleep in the serialization fake. Assert blocked writes before and at context deadline expiration. Validate with go test ./rpc.

Milestone 2 migrates root deadline tests, bounded cancellation cleanup, notification-overflow interruption, preferred-handler callback, and replay coordination. Use the unchanged two-second production cleanup timeout, assert interrupt initiation and pending state just before expiration, and preserve the original cancellation error at expiration. Deadline tests must return DeadlineExceeded. Validate with go test . ./rpc.

Milestone 3 runs repeated race tests and the full gate, reviews diffs for retained assertions and cleanup, and commits the test-only migration with this updated plan.

## Concrete Steps

Run from /workspace/codex-sdk-go using Go 1.26.8:

    go test . ./rpc
    go test -race -count=50 -cpu=1,2,4 . ./rpc
    go test ./...
    go test -race ./...
    go vet ./...
    go run honnef.co/go/tools/cmd/staticcheck@v0.7.0 ./...
    go run golang.org/x/vuln/cmd/govulncheck@v1.3.0 ./...
    bash .github/scripts/test-install-codex-cli.sh
    git diff --check

Format changed Go tests with gofmt and verify all tracked Go files have no formatting differences. Tests should report ok, and govulncheck should report no vulnerabilities.

## Validation and Acceptance

Preserve callback, transcript, overflow, queue, cancellation, serialization, and race assertions. Migrated tests must have no wall-clock polling or tolerance windows, must check deadline boundaries, and must release all goroutines at bubble exit. Run the repeat matrix at CPU counts 1, 2, and 4. Existing subprocess timing remains outside scope. No production, generated, or dependency changes are permitted.

## Idempotence and Recovery

Checks can be rerun safely. If a bubble deadlocks, inspect fake transport/handler cleanup and ensure all resources were created inside the bubble. Fix test coordination without weakening assertions or changing production behavior.

## Artifacts and Notes

Baseline: go test . ./rpc passed on 2026-09-06. Linear access returned oauth_token_invalid_grant.

Final evidence:

    go test -race -count=50 -cpu=1,2,4 . ./rpc
    ok github.com/pmenglund/codex-sdk-go 23.298s
    ok github.com/pmenglund/codex-sdk-go/rpc 9.381s
    govulncheck: No vulnerabilities found.

A final timing search found only virtual deadline/serialization sleeps in migrated tests, plus the intentionally unchanged subprocess sleep and E2E logout deadlines. The turn-handle deadline test is named TestTurnHandleContextDeadline to distinguish it from explicit cancellation coverage.

## Interfaces and Dependencies

Use only standard-library testing/synctest.Test(t, func(t *testing.T)) and synctest.Wait. Shared RPC waiting helpers remain test-only and require callers to run inside a bubble. No public interface or dependency changes.

Revision note (2026-09-06): Created from the approved plan before implementation.

Revision note (2026-09-06): Implemented both migration milestones. Added failure-path cleanup for blocked fake writers/handlers and shutdown-aware fake interrupt responses; these keep bubble cleanup safe even when assertions fail.

Revision note (2026-09-06): Recorded final validation and outcomes after cleanup review; all implementation and validation milestones are complete.
