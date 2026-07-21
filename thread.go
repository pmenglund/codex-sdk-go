package codex

import (
	"context"
	"errors"
	"log/slog"

	"github.com/pmenglund/codex-sdk-go/rpc"
)

// Thread represents an active conversation thread.
type Thread struct {
	client *rpc.Client
	id     string
	logger *slog.Logger
}

// ID returns the thread id.
func (t *Thread) ID() string {
	return t.id
}

// Run sends a text prompt and waits for the turn to finish.
func (t *Thread) Run(ctx context.Context, prompt string, opts *TurnOptions) (*TurnResult, error) {
	return t.RunInputs(ctx, []Input{TextInput(prompt)}, opts)
}

// RunInputs sends structured inputs and waits for the turn to finish. If ctx
// is canceled or expires, it best-effort interrupts the remote turn before
// returning the original context error. Use RunStreamed to retain manual
// control over whether remote work is interrupted.
func (t *Thread) RunInputs(ctx context.Context, inputs []Input, opts *TurnOptions) (*TurnResult, error) {
	if err := t.ensureReady(); err != nil {
		return nil, err
	}

	handle, err := t.StartTurn(ctx, inputs, opts)
	if err != nil {
		return nil, err
	}
	return handle.Run(ctx)
}

// RunStreamed sends structured inputs and returns a streaming iterator.
// The iterator includes thread-scoped events and any notifications that omit
// threadId (for example account/session updates).
func (t *Thread) RunStreamed(ctx context.Context, inputs []Input, opts *TurnOptions) (*TurnStream, error) {
	if err := t.ensureReady(); err != nil {
		return nil, err
	}

	handle, err := t.StartTurn(ctx, inputs, opts)
	if err != nil {
		return nil, err
	}
	return handle.Stream()
}

func (t *Thread) ensureReady() error {
	if t == nil {
		return errors.New("thread is nil")
	}
	if t.client == nil {
		return errors.New("thread client is not initialized")
	}
	if t.id == "" {
		return errors.New("thread id is empty")
	}
	return nil
}
