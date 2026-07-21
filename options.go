package codex

import (
	"io"
	"log/slog"

	"github.com/pmenglund/codex-sdk-go/protocol"
	"github.com/pmenglund/codex-sdk-go/rpc"
)

// Options configures the Codex client.
type Options struct {
	// Transport overrides the default stdio spawn.
	Transport rpc.Transport

	// Spawn controls how the default stdio process is launched.
	Spawn SpawnOptions

	// Logger receives SDK logs. If nil, logging is disabled.
	Logger *slog.Logger

	// ClientInfo identifies this SDK to the app-server.
	ClientInfo protocol.ClientInfo

	// ApprovalHandler handles server approval requests.
	ApprovalHandler rpc.ServerRequestHandler

	// CompatibilityPolicy controls validation of a spawned Codex CLI. The zero
	// value, RequireMajorMinor, rejects binaries whose major/minor version cannot
	// be verified against the generated protocol. Custom transports are not probed.
	CompatibilityPolicy CompatibilityPolicy
}

// CompatibilityPolicy controls spawned Codex CLI version validation.
type CompatibilityPolicy uint8

const (
	// RequireMajorMinor requires the runtime and generated protocol to have the
	// same major and minor version. Patch differences are allowed.
	RequireMajorMinor CompatibilityPolicy = iota
	// Warn logs compatibility failures and continues. Supply Options.Logger to
	// observe the warning.
	Warn
	// Ignore disables the Codex CLI version probe.
	Ignore
)

// SpawnOptions configures the spawned codex app-server process.
type SpawnOptions struct {
	// CodexPath is the path to the codex binary (defaults to "codex").
	CodexPath string
	// ConfigOverrides are passed as --config key=value flags.
	ConfigOverrides []string
	// ExtraArgs are appended to the command line.
	ExtraArgs []string
	// Stderr captures stderr from the codex process (defaults to os.Stderr).
	Stderr io.Writer
	// Env holds extra environment entries for the spawned process, in
	// "KEY=value" form. They are appended to the parent process environment,
	// so an entry overrides an inherited variable of the same name (useful
	// for CODEX_HOME or CODEX_API_KEY). Nil inherits the parent environment
	// unchanged.
	Env []string
}
