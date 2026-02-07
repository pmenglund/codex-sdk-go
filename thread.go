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

// RunInputs sends structured inputs and waits for the turn to finish.
func (t *Thread) RunInputs(ctx context.Context, inputs []Input, opts *TurnOptions) (*TurnResult, error) {
	if err := t.ensureReady(); err != nil {
		return nil, err
	}

	logger := resolveLogger(t.logger)
	stream, err := t.RunStreamed(ctx, inputs, opts)
	if err != nil {
		return nil, err
	}
	defer stream.Close()

	result := &TurnResult{}
	for {
		note, err := stream.Next(ctx)
		if err != nil {
			return nil, err
		}
		result.Notifications = append(result.Notifications, note)
		updateTurnResult(result, note)

		if note.Method == "turn/completed" {
			if turnErr := notificationError(note); turnErr != nil {
				logger.Error("codex turn failed", "thread_id", t.id, "turn_id", result.TurnID, "error", turnErr)
				return nil, turnErr
			}
			logger.Info("codex turn completed", "thread_id", t.id, "turn_id", result.TurnID)
			return result, nil
		}
		if note.Method == "turn/failed" {
			turnErr := notificationError(note)
			if turnErr == nil {
				turnErr = errors.New("turn failed")
			}
			logger.Error("codex turn failed", "thread_id", t.id, "turn_id", result.TurnID, "error", turnErr)
			return nil, turnErr
		}
		if note.Method == "error" {
			if turnErr := notificationError(note); turnErr != nil {
				logger.Error("codex turn failed", "thread_id", t.id, "turn_id", result.TurnID, "error", turnErr)
				return nil, turnErr
			}
		}
	}
}

// RunStreamed sends structured inputs and returns a streaming iterator.
// The iterator includes thread-scoped events and any notifications that omit
// threadId (for example account/session updates).
func (t *Thread) RunStreamed(ctx context.Context, inputs []Input, opts *TurnOptions) (*TurnStream, error) {
	if err := t.ensureReady(); err != nil {
		return nil, err
	}

	logger := resolveLogger(t.logger)
	iter := t.client.SubscribeNotifications(0)

	params, err := buildTurnParams(t.id, inputs, opts)
	if err != nil {
		logger.Error("codex turn start failed", "thread_id", t.id, "error", err)
		iter.Close()
		return nil, err
	}
	logger.Info("codex starting turn", "thread_id", t.id, "input_count", len(inputs))
	if err := t.client.Call(ctx, "turn/start", params, nil); err != nil {
		logger.Error("codex turn start failed", "thread_id", t.id, "error", err)
		iter.Close()
		return nil, err
	}

	return &TurnStream{iter: iter, threadID: t.id}, nil
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
