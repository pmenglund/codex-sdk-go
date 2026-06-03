package codex

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"log/slog"
	"sync"

	"github.com/pmenglund/codex-sdk-go/protocol"
	"github.com/pmenglund/codex-sdk-go/rpc"
)

// TurnHandle controls a running turn.
type TurnHandle struct {
	client   *rpc.Client
	threadID string
	logger   *slog.Logger
	stream   *TurnStream

	mu       sync.Mutex
	turnID   string
	closed   bool
	closeErr error
}

// StartTurn sends structured inputs and returns a handle for the running turn.
func (t *Thread) StartTurn(ctx context.Context, inputs []Input, opts *TurnOptions) (*TurnHandle, error) {
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
	response, err := t.client.TurnStart(ctx, params)
	if err != nil {
		logger.Error("codex turn start failed", "thread_id", t.id, "error", err)
		iter.Close()
		return nil, err
	}

	handle := &TurnHandle{
		client:   t.client,
		threadID: t.id,
		logger:   logger,
		stream:   &TurnStream{iter: iter, threadID: t.id},
	}
	if id, ok := turnIDFromAny(response); ok {
		handle.setTurnID(id)
	}
	return handle, nil
}

// Stream returns the handle's notification stream.
func (h *TurnHandle) Stream() (*TurnStream, error) {
	if err := h.ensureReady(); err != nil {
		return nil, err
	}
	return h.stream, nil
}

// Next returns the next notification for this turn and updates the handle state.
func (h *TurnHandle) Next(ctx context.Context) (rpc.Notification, error) {
	if err := h.ensureReady(); err != nil {
		return rpc.Notification{}, err
	}
	note, err := h.stream.Next(ctx)
	if err != nil {
		return note, err
	}
	h.updateFromNotification(note)
	return note, nil
}

// Run waits for this turn to complete and returns its aggregated result.
func (h *TurnHandle) Run(ctx context.Context) (*TurnResult, error) {
	if err := h.ensureReady(); err != nil {
		return nil, err
	}
	defer h.Close()

	result := &TurnResult{}
	for {
		note, err := h.Next(ctx)
		if err != nil {
			return nil, err
		}
		result.Notifications = append(result.Notifications, note)
		updateTurnResult(result, note)

		if note.Method == "turn/completed" {
			if turnErr := notificationError(note); turnErr != nil {
				resolveLogger(h.logger).Error("codex turn failed", "thread_id", h.threadID, "turn_id", result.TurnID, "error", turnErr)
				return nil, turnErr
			}
			resolveLogger(h.logger).Info("codex turn completed", "thread_id", h.threadID, "turn_id", result.TurnID)
			return result, nil
		}
		if note.Method == "turn/failed" {
			turnErr := notificationError(note)
			if turnErr == nil {
				turnErr = errors.New("turn failed")
			}
			resolveLogger(h.logger).Error("codex turn failed", "thread_id", h.threadID, "turn_id", result.TurnID, "error", turnErr)
			return nil, turnErr
		}
		if note.Method == "error" {
			if turnErr := notificationError(note); turnErr != nil {
				resolveLogger(h.logger).Error("codex turn failed", "thread_id", h.threadID, "turn_id", result.TurnID, "error", turnErr)
				return nil, turnErr
			}
		}
	}
}

// Steer sends additional input to the active turn.
func (h *TurnHandle) Steer(ctx context.Context, inputs []Input) (*protocol.TurnSteerResponse, error) {
	if err := h.ensureReady(); err != nil {
		return nil, err
	}
	turnID := h.currentTurnID()
	if turnID == "" {
		return nil, errors.New("turn id is not known yet")
	}
	params, err := buildTurnSteerParams(h.threadID, turnID, inputs)
	if err != nil {
		return nil, err
	}
	return h.client.TurnSteer(ctx, params)
}

// Interrupt interrupts the active turn.
func (h *TurnHandle) Interrupt(ctx context.Context) (*protocol.TurnInterruptResponse, error) {
	if err := h.ensureReady(); err != nil {
		return nil, err
	}
	turnID := h.currentTurnID()
	if turnID == "" {
		return nil, errors.New("turn id is not known yet")
	}
	return h.client.TurnInterrupt(ctx, protocol.TurnInterruptParams{ThreadID: h.threadID, TurnID: turnID})
}

// Close releases the handle's notification subscription.
func (h *TurnHandle) Close() {
	if h == nil {
		return
	}
	h.mu.Lock()
	if h.closed {
		h.mu.Unlock()
		return
	}
	h.closed = true
	stream := h.stream
	h.mu.Unlock()

	if stream != nil {
		stream.Close()
	}
}

func (h *TurnHandle) ensureReady() error {
	if h == nil {
		return errors.New("turn handle is nil")
	}
	h.mu.Lock()
	defer h.mu.Unlock()
	if h.closed {
		return errors.New("turn handle is closed")
	}
	if h.client == nil {
		return errors.New("turn handle client is not initialized")
	}
	if h.threadID == "" {
		return errors.New("turn handle thread id is empty")
	}
	if h.stream == nil {
		return errors.New("turn handle stream is not initialized")
	}
	return nil
}

func (h *TurnHandle) currentTurnID() string {
	h.mu.Lock()
	defer h.mu.Unlock()
	return h.turnID
}

func (h *TurnHandle) setTurnID(turnID string) {
	if turnID == "" {
		return
	}
	h.mu.Lock()
	h.turnID = turnID
	h.mu.Unlock()
}

func (h *TurnHandle) updateFromNotification(note rpc.Notification) {
	payload, err := parseTurnNotification(note)
	if err != nil {
		return
	}
	if payload.Turn != nil && payload.Turn.ID != "" {
		h.setTurnID(payload.Turn.ID)
		return
	}
	if payload.TurnID != "" {
		h.setTurnID(payload.TurnID)
	}
}

func buildTurnSteerParams(threadID, turnID string, inputs []Input) (protocol.TurnSteerParams, error) {
	params := protocol.TurnSteerParams{
		ThreadID:       threadID,
		ExpectedTurnID: turnID,
		Input:          make([]protocol.TurnSteerParamsInputElem, 0, len(inputs)),
	}
	if threadID == "" {
		return params, errors.New("thread id is required")
	}
	if turnID == "" {
		return params, errors.New("turn id is required")
	}
	for _, input := range inputs {
		if err := input.validate(); err != nil {
			return params, fmt.Errorf("input: %w", err)
		}
		params.Input = append(params.Input, input)
	}
	return params, nil
}

func turnIDFromAny(value any) (string, bool) {
	if value == nil {
		return "", false
	}
	data, err := JSON(value)
	if err != nil {
		return "", false
	}
	var wrapper struct {
		Turn *protocol.TurnNotificationTurn `json:"turn,omitempty"`
		ID   string                         `json:"id,omitempty"`
	}
	if err := json.Unmarshal(data, &wrapper); err != nil {
		return "", false
	}
	if wrapper.Turn != nil && wrapper.Turn.ID != "" {
		return wrapper.Turn.ID, true
	}
	if wrapper.ID != "" {
		return wrapper.ID, true
	}
	return "", false
}
