# codex-sdk-go Architecture Notes

This file captures application-specific context that should stay stable across tasks.

## Purpose

`codex-sdk-go` is a Go SDK for embedding the Codex app-server into Go workflows.
It provides a high-level `codex` client API, streaming turn execution, approval handling, and optional low-level JSON-RPC access.

## System Boundaries

- Primary runtime(s): Go 1.25+ and the `codex` CLI process (spawned and managed by the SDK).
- External services: Local `codex app-server` over JSON-RPC (stdio transport); local `openai/codex` checkout for schema export during codegen.
- Data stores: No persistent datastore in this repo; generated artifacts are checked into source control.

## Repository Layout

- Repository root (`*.go`) - main `codex` package and user-facing API.
- `rpc/` - low-level JSON-RPC transport/client and generated stubs.
- `protocol/` - generated protocol schema types.
- `internal/codegen/` - code generation implementation invoked by `go generate`.
- `examples/` - runnable usage examples.
- `*_test.go` files at root plus `examples_test.go`/`turn_test.go` - automated tests.
- `README.md` and `doc.go` - user/developer documentation.

## Core Components

- `codex` package: High-level SDK facade (client/thread/turn APIs, options, approvals integration).
- `rpc` package: JSON-RPC client/server plumbing and request/notification handling.
- `protocol` package: Generated wire-level types shared by SDK and app-server protocol.

## Architecture Rules

- Keep user-facing behavior in the `codex` package and transport concerns in `rpc`; avoid leaking RPC details into high-level APIs.
- Prefer extension through existing abstractions before introducing new top-level modules.
- Record significant architecture tradeoffs in the active ExecPlan decision log.
- Generated files in `protocol/` and `rpc/` must be checked in.
- Keep examples in `examples/` aligned with `README.md` and `doc.go`.

## Local Development

- Install dependencies: `go mod tidy`
- Run example locally: `go run ./examples/quickstart` (requires `codex` on `PATH`).
- Run tests locally: `go test ./...`
- Lint/format checks: `gofmt -w ./...` (changed files) and `go vet ./...`
- Regenerate protocol and RPC code: `go generate ./...`

## Operational Constraints

- Security and privacy requirements: Approval handling must remain explicit and safe; sample auto-approve behavior should stay minimal and conservative.
- Performance expectations: Streaming APIs should remain responsive and avoid unnecessary buffering/copying for turn notifications.
- Compatibility constraints: Support Go 1.25+ and maintain protocol compatibility with generated schema versions.

## Change Checklist for Contributors

- Update this file when architecture, paths, or commands change.
- Keep examples and commands copy/paste ready.
- Ensure this file stays consistent with `README.md`, `WORKFLOW.md`, and `LANGUAGE.md`.
