# Go SDK

## Scope
- Applies to `sdk/go`.

## Codegen
- The Go SDK uses `go generate` (see `sdk/go/gen.go`) to:
  - Export schemas via `cargo run -p codex-app-server-protocol --bin export`.
  - Generate protocol types and RPC stubs under `sdk/go/protocol` and `sdk/go/rpc`.
- Generated files must be checked in.
- Before considering a task complete, run in `sdk/go`:
  - `gofmt -w` on any Go files changed.
  - `go mod tidy`
  - `go generate ./...`
  - `go vet ./...`
  - `go test ./...`

## Testing
- For a feature to be considered complete, total test coverage must be > 80%.
  - `go test ./... -coverprofile=coverage.out`.
- The `codex` package should be tested using `_test` extension to enforce the exported functinos.
- Use table driven tests where applicable.

## Go version
- Go 1.25.

## Logging
- Use `log/slog` for logging.

## SDK layout
- `codex` package: user-facing facade and helpers.
- `rpc` package: low-level JSON-RPC client/transport.
- `protocol` package: generated schema types.

## Examples
- The examples in `examples/` must be kept up to date with the ones in `README.md` and `doc.go`.

## Approvals
- Provide an approval handler API.
- Keep the sample auto-approve implementation simple and safe.
