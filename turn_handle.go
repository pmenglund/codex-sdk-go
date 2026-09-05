package codex

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"log/slog"
	"sync"
	"time"

	"github.com/pmenglund/codex-sdk-go/protocol"
	"github.com/pmenglund/codex-sdk-go/rpc"
)

var turnInterruptCleanupTimeout = 2 * time.Second

// ErrTurnConsumptionMode identifies attempts to mix Run, Next, and Stream on
// the same TurnHandle or to consume it concurrently.
var ErrTurnConsumptionMode = errors.New("turn handle already has a consumer")

type turnConsumptionMode uint8

const (
	turnConsumptionUnset turnConsumptionMode = iota
	turnConsumptionRun
	turnConsumptionNext
	turnConsumptionStream
)

// TurnHandle controls a running turn.
type TurnHandle struct {
	client   *rpc.Client
	threadID string
	logger   *slog.Logger
	stream   *TurnStream

	mu           sync.Mutex
	turnID       string
	closed       bool
	mode         turnConsumptionMode
	runStarted   bool
	streamIssued bool
	consumeMu    sync.Mutex
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
	if response != nil && response.Turn.ID != "" {
		handle.setTurnID(response.Turn.ID)
	}
	return handle, nil
}

// Stream returns the handle's notification stream.
func (h *TurnHandle) Stream() (*TurnStream, error) {
	if err := h.ensureReady(); err != nil {
		return nil, err
	}
	if err := h.beginStream(); err != nil {
		return nil, err
	}
	return &TurnStream{next: func(ctx context.Context) (rpc.Notification, error) {
		return h.next(ctx, turnConsumptionStream)
	}, close: h.Close}, nil
}

func (h *TurnHandle) beginStream() error {
	h.mu.Lock()
	defer h.mu.Unlock()
	if h.mode != turnConsumptionUnset || h.streamIssued {
		return ErrTurnConsumptionMode
	}
	h.mode = turnConsumptionStream
	h.streamIssued = true
	return nil
}

// Next returns the next notification for this turn and updates the handle state.
func (h *TurnHandle) Next(ctx context.Context) (rpc.Notification, error) {
	return h.next(ctx, turnConsumptionNext)
}

func (h *TurnHandle) next(ctx context.Context, mode turnConsumptionMode) (rpc.Notification, error) {
	if err := h.ensureReady(); err != nil {
		return rpc.Notification{}, err
	}
	if err := h.claimConsumption(mode); err != nil {
		return rpc.Notification{}, err
	}
	if !h.consumeMu.TryLock() {
		return rpc.Notification{}, ErrTurnConsumptionMode
	}
	defer h.consumeMu.Unlock()
	note, err := h.stream.Next(ctx)
	if err != nil {
		return note, err
	}
	h.updateFromNotification(note)
	return note, nil
}

// Run waits for this turn to complete and returns its aggregated result. It
// best-effort interrupts remote work after context cancellation, deadline, or
// notification overflow; cleanup is bounded and the original error is returned.
func (h *TurnHandle) Run(ctx context.Context) (*TurnResult, error) {
	if err := h.ensureReady(); err != nil {
		return nil, err
	}
	if err := h.beginRun(); err != nil {
		return nil, err
	}
	defer h.Close()

	result := &TurnResult{}
	for {
		note, err := h.next(ctx, turnConsumptionRun)
		if err != nil {
			if err == context.Canceled || err == context.DeadlineExceeded || errors.Is(err, rpc.ErrNotificationOverflow) {
				h.interruptAfterRunError(ctx, err)
			}
			return nil, err
		}
		result.Notifications = append(result.Notifications, note)
		updateTurnResult(result, note)

		if note.Method == "turn/completed" {
			if turnErr := turnErrorFromNotification(note, result); turnErr != nil {
				resolveLogger(h.logger).Error("codex turn failed", "thread_id", h.threadID, "turn_id", result.TurnID, "error", turnErr)
				return result, turnErr
			}
			resolveLogger(h.logger).Info("codex turn completed", "thread_id", h.threadID, "turn_id", result.TurnID)
			return result, nil
		}
		if note.Method == "turn/failed" {
			turnErr := turnErrorFromNotification(note, result)
			if turnErr == nil {
				turnErr = &TurnError{Result: result, Method: note.Method, Raw: append(json.RawMessage(nil), note.Raw...)}
			}
			resolveLogger(h.logger).Error("codex turn failed", "thread_id", h.threadID, "turn_id", result.TurnID, "error", turnErr)
			return result, turnErr
		}
		if note.Method == "error" {
			if turnErr := turnErrorFromNotification(note, result); turnErr != nil {
				resolveLogger(h.logger).Error("codex turn failed", "thread_id", h.threadID, "turn_id", result.TurnID, "error", turnErr)
				return result, turnErr
			}
		}
	}
}

func (h *TurnHandle) claimConsumption(requested turnConsumptionMode) error {
	h.mu.Lock()
	defer h.mu.Unlock()
	if h.mode == turnConsumptionUnset {
		h.mode = requested
		return nil
	}
	if h.mode != requested {
		return ErrTurnConsumptionMode
	}
	return nil
}

func (h *TurnHandle) beginRun() error {
	h.mu.Lock()
	defer h.mu.Unlock()
	if h.mode != turnConsumptionUnset || h.runStarted {
		return ErrTurnConsumptionMode
	}
	h.mode = turnConsumptionRun
	h.runStarted = true
	return nil
}

func (h *TurnHandle) interruptAfterRunError(ctx context.Context, cause error) {
	turnID := h.currentTurnID()
	logger := resolveLogger(h.logger)
	if turnID == "" {
		logger.Warn("codex turn run failure could not be interrupted", "thread_id", h.threadID, "error", cause)
		return
	}

	cleanupCtx, cancel := context.WithTimeout(context.WithoutCancel(ctx), turnInterruptCleanupTimeout)
	defer cancel()
	interruptResult := make(chan error, 1)
	go func() {
		_, err := h.client.TurnInterrupt(cleanupCtx, protocol.TurnInterruptParams{ThreadID: h.threadID, TurnID: turnID})
		interruptResult <- err
	}()
	select {
	case err := <-interruptResult:
		if err == nil {
			return
		}
		logger.Warn("codex turn interrupt after run failure failed", "thread_id", h.threadID, "turn_id", turnID, "error", err)
	case <-cleanupCtx.Done():
		logger.Warn("codex turn interrupt after run failure timed out", "thread_id", h.threadID, "turn_id", turnID, "error", cleanupCtx.Err())
	}
}

// ID returns the server-assigned turn ID when it is known.
func (h *TurnHandle) ID() string {
	if h == nil {
		return ""
	}
	return h.currentTurnID()
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
		Input:          make([]protocol.UserInput, 0, len(inputs)),
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
		wrapped, err := protocol.NewUserInput(input)
		if err != nil {
			return params, fmt.Errorf("wrap input: %w", err)
		}
		params.Input = append(params.Input, wrapped)
	}
	return params, nil
}
