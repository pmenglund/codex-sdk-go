package codex

import (
	"context"
	"errors"
	"fmt"
	"log/slog"
	"os/exec"
	"runtime/debug"
	"strings"
	"time"

	"github.com/pmenglund/codex-sdk-go/protocol"
	"github.com/pmenglund/codex-sdk-go/rpc"
)

const codexVersionProbeTimeout = 2 * time.Second

// Codex is the main entrypoint for the Go SDK.
type Codex struct {
	client *rpc.Client
	logger *slog.Logger
}

// New creates a new Codex client and performs the initialize handshake.
func New(ctx context.Context, opts Options) (*Codex, error) {
	logger := resolveLogger(opts.Logger)
	if opts.RequestHandler != nil && !isNilServerRequestHandler(opts.ApprovalHandler) {
		return nil, errors.New("request handler conflicts with deprecated approval handler")
	}
	requestHandler := opts.ApprovalHandler
	if opts.RequestHandler != nil {
		requestHandler = *opts.RequestHandler
	}

	transport := opts.Transport
	if transport == nil {
		spawn := opts.Spawn
		if spawn.CodexPath == "" {
			spawn.CodexPath = "codex"
		}
		args := []string{"app-server"}
		for _, override := range spawn.ConfigOverrides {
			args = append(args, "--config", override)
		}
		args = append(args, spawn.ExtraArgs...)

		logger.Info("codex checking CLI compatibility", "path", spawn.CodexPath)
		if err := checkCodexCompatibility(ctx, logger, spawn.CodexPath, opts.CompatibilityPolicy); err != nil {
			return nil, err
		}

		var err error
		if spawn.Stderr == nil {
			spawn.Stderr = rpc.DefaultStderr()
		}
		if err := ctx.Err(); err != nil {
			return nil, err
		}
		logger.Info("codex starting app-server", "path", spawn.CodexPath, "argument_count", len(args))
		// The constructor context is only for initialization; process lifetime is managed by Close.
		transport, err = rpc.SpawnStdio(context.WithoutCancel(ctx), spawn.CodexPath, args, spawn.Stderr)
		if err != nil {
			return nil, err
		}
	} else {
		logger.Info("codex using custom transport")
	}

	client, err := rpc.NewClientChecked(transport, rpc.ClientOptions{
		Logger:         logger,
		RequestHandler: attachApprovalLogger(requestHandler, logger),
	})
	if err != nil {
		return nil, err
	}

	info := opts.ClientInfo
	if info.Name == "" {
		info = defaultClientInfo()
	}

	if _, err := client.Initialize(ctx, protocol.InitializeParams{ClientInfo: info}); err != nil {
		_ = client.Close()
		return nil, err
	}

	if err := client.Notify(ctx, "initialized", nil); err != nil {
		_ = client.Close()
		return nil, err
	}

	logger.Info("codex initialized")

	return &Codex{client: client, logger: logger}, nil
}

// Client exposes the underlying RPC client for low-level access.
func (c *Codex) Client() *rpc.Client {
	return c.client
}

// Close closes the underlying transport.
func (c *Codex) Close() error {
	if err := c.ensureReady(); err != nil {
		return err
	}
	return c.client.Close()
}

// StartThread starts a new thread using the app-server.
func (c *Codex) StartThread(ctx context.Context, options ThreadStartOptions) (*Thread, error) {
	if err := c.ensureReady(); err != nil {
		return nil, err
	}
	params, err := options.toParams()
	if err != nil {
		return nil, err
	}
	var response protocol.ThreadStartResponse
	if err := c.client.Call(ctx, "thread/start", params, &response); err != nil {
		return nil, err
	}
	threadID, err := threadIDFromResponse(response.ThreadID, response.Thread)
	if err != nil {
		return nil, err
	}
	c.logger.Info("codex thread started", "thread_id", threadID)
	return &Thread{client: c.client, id: threadID, logger: c.logger}, nil
}

// ResumeThread resumes an existing thread.
func (c *Codex) ResumeThread(ctx context.Context, options ThreadResumeOptions) (*Thread, error) {
	if err := c.ensureReady(); err != nil {
		return nil, err
	}
	params, err := options.toParams()
	if err != nil {
		return nil, err
	}
	var response protocol.ThreadResumeResponse
	if err := c.client.Call(ctx, "thread/resume", params, &response); err != nil {
		return nil, err
	}
	threadID, err := threadIDFromResponse(response.ThreadID, response.Thread)
	if err != nil {
		return nil, err
	}
	c.logger.Info("codex thread resumed", "thread_id", threadID)
	return &Thread{client: c.client, id: threadID, logger: c.logger}, nil
}

func defaultClientInfo() protocol.ClientInfo {
	version := "dev"
	if info, ok := debug.ReadBuildInfo(); ok && info.Main.Version != "" {
		version = info.Main.Version
	}
	return protocol.ClientInfo{
		Name:    "codex-go-sdk",
		Title:   stringPtr("Codex Go SDK"),
		Version: version,
	}
}

func stringPtr(value string) *string {
	return &value
}

func boolPtr(value bool) *bool {
	return &value
}

func checkCodexCompatibility(ctx context.Context, logger *slog.Logger, codexPath string, policy CompatibilityPolicy) error {
	if policy == Ignore {
		return nil
	}
	if policy != RequireMajorMinor && policy != Warn {
		return fmt.Errorf("invalid compatibility policy %d", policy)
	}

	generatedVersion := protocol.GeneratedCodexVersion
	if generatedVersion == "" {
		return nil
	}
	runtimeVersion, err := probeCodexVersion(ctx, codexPath)
	if err != nil {
		return handleCompatibilityFailure(logger, policy, newCompatibilityError(codexPath, "", "version probe failed", err))
	}
	if runtimeVersion == "" {
		return handleCompatibilityFailure(logger, policy, newCompatibilityError(codexPath, "", "version output was not parseable", nil))
	}
	if sameMajorMinor(runtimeVersion, generatedVersion) {
		return nil
	}
	return handleCompatibilityFailure(logger, policy, newCompatibilityError(codexPath, runtimeVersion, "major/minor version mismatch", nil))
}

func newCompatibilityError(path, runtimeVersion, reason string, cause error) *CodexCompatibilityError {
	return &CodexCompatibilityError{
		Path:             path,
		RuntimeVersion:   runtimeVersion,
		GeneratedVersion: protocol.GeneratedCodexVersion,
		GeneratedCommit:  protocol.GeneratedCodexCommit,
		Reason:           reason,
		Hint:             "install a matching Codex CLI or explicitly set CompatibilityPolicy to Warn or Ignore after validating protocol compatibility",
		Cause:            cause,
	}
}

func handleCompatibilityFailure(logger *slog.Logger, policy CompatibilityPolicy, compatibilityErr *CodexCompatibilityError) error {
	if policy == RequireMajorMinor {
		return compatibilityErr
	}
	resolveLogger(logger).Warn(
		"codex binary compatibility could not be guaranteed",
		"path", compatibilityErr.Path,
		"runtime_version", compatibilityErr.RuntimeVersion,
		"generated_version", compatibilityErr.GeneratedVersion,
		"generated_commit", compatibilityErr.GeneratedCommit,
		"reason", compatibilityErr.Reason,
		"error", compatibilityErr.Cause,
	)
	return nil
}

func sameMajorMinor(runtimeVersion, generatedVersion string) bool {
	runtimeMajor, runtimeMinor, runtimeOK := majorMinor(runtimeVersion)
	generatedMajor, generatedMinor, generatedOK := majorMinor(generatedVersion)
	return runtimeOK && generatedOK && runtimeMajor == generatedMajor && runtimeMinor == generatedMinor
}

func majorMinor(version string) (string, string, bool) {
	parts := strings.Split(version, ".")
	if len(parts) < 2 || !isDottedVersion(version) {
		return "", "", false
	}
	return strings.TrimLeft(parts[0], "0"), strings.TrimLeft(parts[1], "0"), true
}

func probeCodexVersion(parent context.Context, codexPath string) (string, error) {
	ctx, cancel := context.WithTimeout(parent, codexVersionProbeTimeout)
	defer cancel()

	out, err := exec.CommandContext(ctx, codexPath, "--version").Output()
	if err != nil {
		return "", err
	}
	return parseCodexVersionOutput(string(out)), nil
}

func parseCodexVersionOutput(output string) string {
	for _, field := range strings.Fields(output) {
		field = strings.TrimPrefix(field, "v")
		if isDottedVersion(field) {
			return field
		}
	}
	return ""
}

func isDottedVersion(value string) bool {
	parts := strings.Split(value, ".")
	if len(parts) < 2 {
		return false
	}
	for _, part := range parts {
		if part == "" {
			return false
		}
		for _, r := range part {
			if r < '0' || r > '9' {
				return false
			}
		}
	}
	return true
}

func threadIDFromResponse(threadID string, thread *protocol.Thread) (string, error) {
	if threadID != "" {
		return threadID, nil
	}
	if thread != nil && thread.ID != "" {
		return thread.ID, nil
	}
	return "", errors.New("thread id not found in response")
}

func (c *Codex) ensureReady() error {
	if c == nil {
		return errors.New("codex client is nil")
	}
	if c.client == nil {
		return errors.New("codex client is not initialized")
	}
	return nil
}
