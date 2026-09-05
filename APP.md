# codex-sdk-go Architecture Notes

This file captures application-specific context that should stay stable across tasks.

## Purpose

`codex-sdk-go` is a Go SDK for embedding the Codex app-server into Go workflows.
It provides a high-level `codex` client API, streaming turn execution, approval handling, and optional low-level JSON-RPC access.

## System Boundaries

- Primary runtime(s): Go 1.25.14+ and the `codex` CLI process (spawned and managed by the SDK).
- External services: Local `codex app-server` over JSON-RPC (stdio transport); local `openai/codex` checkout for schema export during codegen.
- Data stores: No persistent datastore in this repo; generated artifacts are checked into source control.

## Repository Layout

- Repository root (`*.go`) - main `codex` package and user-facing API.
- `rpc/` - low-level JSON-RPC transport/client, bounded writer and server-request queues, and generated stubs.
- `protocol/` - generated protocol schema types plus reviewed manual wrappers for schemas the generator cannot yet express safely.
- `internal/codegen/` - code generation implementation invoked by `go generate`.
- `examples/` - runnable usage examples.
- `*_test.go` files at root plus `examples_test.go`/`turn_test.go` - automated tests.
- `README.md` and `doc.go` - user/developer documentation.

## Core Components

- `codex` package: High-level SDK facade (client/thread/turn APIs, options, approvals integration).
- `rpc` package: JSON-RPC client/server plumbing with serialized context-bounded writes, bounded server-request workers, and request/notification handling.
- `protocol` package: Generated wire-level types shared by SDK and app-server protocol. Canonical Go initialisms are generated centrally; raw-preserving wrappers retain unresolved union data.

## Architecture Rules

- Keep user-facing behavior in the `codex` package and transport concerns in `rpc`; avoid leaking RPC details into high-level APIs.
- Prefer extension through existing abstractions before introducing new top-level modules.
- Record significant architecture tradeoffs in the active ExecPlan decision log.
- Generated files in `protocol/` and `rpc/` must be checked in.
- Keep manual protocol declarations in `protocol/manual_types.go` listed in the generator's reviewed manual-type set. Do not hand-edit generated declarations.
- Prefer `ServerRequestCallbacks` in the root package or embedding `rpc.UnimplementedServerRequestHandler`; broad generated handler interfaces are compatibility surfaces only.
- A `TurnHandle` has exactly one notification consumer (`Run`, `Next`, or `Stream`).
- Keep examples in `examples/` aligned with `README.md` and `doc.go`.

## Local Development

- Install dependencies: `go mod tidy`
- Run example locally: `go run ./examples/quickstart` (requires `codex` on `PATH`).
- Run the local quality gate: formatting, installer fixtures, `go vet ./...`,
  `go test ./...`, `go test -race ./...`, `staticcheck ./...`,
  `govulncheck ./...`, and `git diff --check`.
- Use Staticcheck v0.7.0 and govulncheck v1.3.0, matching CI.

### Code Generation

Regenerate protocol types and RPC stubs:

```bash
go generate ./...
```

Generate from a specific Codex tag, branch, or commit without changing the
checkout at `$CODEX_REPO_ROOT`:

```bash
CODEX_REPO_ROOT=../codex CODEX_REPO_REF=<tag> go generate ./...
```

If `$CODEX_REPO_REF` is set, generation runs from a temporary detached git
worktree at that ref. Fetch the desired tag or ref in the Codex checkout before
running generation.

This runs:

- `cargo run -p codex-cli --bin codex -- app-server generate-json-schema`
- `go-jsonschema` (via `internal/codegen`)

The generator needs a checkout of `openai/codex` to export schemas.
It resolves that checkout in this order:

- `$CODEX_REPO_ROOT` (if set)
- `../codex` (default)

Generated files include a header line with the exact codex commit hash used.
Generated files are checked in under `protocol` and `rpc`.
Every generated Go file is formatted before it is written. The repository-owned
update skill runs generation twice and rejects a second diff. Discriminated
unions are rendered as raw-preserving wrapper types; any remaining
`interface{}` output must be present in a reviewed generator allowlist.

## Operational Constraints

- Security and privacy requirements: Approval handling must remain explicit and safe; sample auto-approve behavior should stay minimal and conservative.
- Performance expectations: Streaming APIs should remain responsive and avoid unnecessary buffering/copying for turn notifications.
- Compatibility constraints: Support Go 1.25.14 or newer and maintain protocol compatibility with generated schema versions.

## Change Checklist for Contributors

- Update this file when architecture, paths, or commands change.
- Keep examples and commands copy/paste ready.
- Ensure this file stays consistent with `README.md`, `WORKFLOW.md`, and `LANGUAGE.md`.
