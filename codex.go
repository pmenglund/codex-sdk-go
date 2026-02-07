package codex

import (
	"context"
	"errors"
	"log/slog"
	"runtime/debug"
	"strings"

	"github.com/pmenglund/codex-sdk-go/protocol"
	"github.com/pmenglund/codex-sdk-go/rpc"
)

// Codex is the main entrypoint for the Go SDK.
type Codex struct {
	client *rpc.Client
	logger *slog.Logger
}

// New creates a new Codex client and performs the initialize handshake.
func New(ctx context.Context, opts Options) (*Codex, error) {
	logger := resolveLogger(opts.Logger)

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

		logger.Info("codex starting app-server", "path", spawn.CodexPath, "args", strings.Join(args, " "))

		var err error
		if spawn.Stderr == nil {
			spawn.Stderr = rpc.DefaultStderr()
		}
		if err := ctx.Err(); err != nil {
			return nil, err
		}
		// The constructor context is only for initialization; process lifetime is managed by Close.
		transport, err = rpc.SpawnStdio(context.WithoutCancel(ctx), spawn.CodexPath, args, spawn.Stderr)
		if err != nil {
			return nil, err
		}
	} else {
		logger.Info("codex using custom transport")
	}

	client := rpc.NewClient(transport, rpc.ClientOptions{
		Logger:         logger,
		RequestHandler: attachApprovalLogger(opts.ApprovalHandler, logger),
	})

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
