package codex

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"time"

	"github.com/pmenglund/codex-sdk-go/protocol"
	"github.com/pmenglund/codex-sdk-go/rpc"
)

// TurnOptions configures a turn/start request.
type TurnOptions struct {
	ClientUserMessageID string
	Cwd                 string
	// ApprovalPolicy is marshaled as JSON and sent as "approvalPolicy".
	// Prefer ApprovalPolicy* constants for simple policies.
	ApprovalPolicy any
	// SandboxPolicy is marshaled as JSON and sent as "sandboxPolicy".
	// Prefer SandboxMode* constants for simple policies.
	SandboxPolicy any
	Model         string
	ServiceTier   string
	// Effort is marshaled as JSON and sent as "effort".
	// Prefer ReasoningEffort* constants for standard values.
	Effort any
	// Summary is marshaled as JSON and sent as "summary".
	Summary any
	// OutputSchema is marshaled as JSON and sent as "outputSchema".
	OutputSchema any
	// CollaborationMode is retained for source compatibility, but the current
	// app-server protocol no longer supports this option. Setting it returns an
	// error from buildTurnParams.
	//
	// Deprecated: collaboration mode is no longer supported by the app-server protocol.
	CollaborationMode any
}

// TurnResult aggregates notifications for a completed turn.
type TurnResult struct {
	TurnID        string
	Status        string
	ErrorMessage  string
	Notifications []rpc.Notification
	// Items holds the raw JSON payloads for completed items.
	Items         []json.RawMessage
	FinalResponse string
	TokenUsage    *protocol.ThreadTokenUsage
	CreatedAt     *time.Time
	CompletedAt   *time.Time
}

// ErrTurnFailed identifies a terminal turn failure.
var ErrTurnFailed = errors.New("turn failed")

// TurnError describes a terminal turn failure and retains the partial result
// and complete structured wire details available at failure time.
type TurnError struct {
	Result    *TurnResult
	Method    string
	Detail    *protocol.TurnError
	WillRetry bool
	Raw       json.RawMessage
}

func (e *TurnError) Error() string {
	if e != nil && e.Detail != nil && e.Detail.Message != "" {
		return e.Detail.Message
	}
	if e != nil && e.Result != nil && e.Result.ErrorMessage != "" {
		return e.Result.ErrorMessage
	}
	if e != nil && e.Method == "error" {
		return "turn error"
	}
	return ErrTurnFailed.Error()
}

// Is allows errors.Is(err, ErrTurnFailed).
func (e *TurnError) Is(target error) bool { return target == ErrTurnFailed }

// TurnStream iterates notifications for a running turn.
// Notifications that omit threadId are still emitted to avoid dropping
// global events sent during the turn.
type TurnStream struct {
	iter     *rpc.NotificationIterator
	threadID string
	next     func(context.Context) (rpc.Notification, error)
	close    func()
}

// Next returns the next notification for this turn.
// Notifications without threadId are treated as belonging to the active stream.
func (s *TurnStream) Next(ctx context.Context) (rpc.Notification, error) {
	if s == nil {
		return rpc.Notification{}, errors.New("turn stream is not initialized")
	}
	if s.next != nil {
		return s.next(ctx)
	}
	if s.iter == nil {
		return rpc.Notification{}, errors.New("turn stream is not initialized")
	}

	for {
		note, err := s.iter.Next(ctx)
		if err != nil {
			return note, err
		}
		if s.threadID == "" {
			return note, nil
		}
		if matchesThreadID(note, s.threadID) {
			return note, nil
		}
	}
}

// Close stops the iterator.
func (s *TurnStream) Close() {
	if s == nil {
		return
	}
	if s.close != nil {
		s.close()
		return
	}
	if s.iter == nil {
		return
	}
	s.iter.Close()
}

func updateTurnResult(result *TurnResult, note rpc.Notification) {
	if note.Method != "item/completed" && note.Method != "turn/started" && note.Method != "turn/completed" && note.Method != "turn/failed" && note.Method != "thread/tokenUsage/updated" {
		return
	}

	payload, err := parseTurnNotification(note)
	if err != nil {
		return
	}

	if note.Method == "item/completed" {
		if len(payload.Item) > 0 {
			result.Items = append(result.Items, payload.Item)
			if text, ok := extractTextFromItemRaw(payload.Item); ok {
				result.FinalResponse = text
			}
		}
	}

	if note.Method == "turn/started" || note.Method == "turn/completed" || note.Method == "turn/failed" {
		if payload.Turn != nil && payload.Turn.ID != "" {
			result.TurnID = payload.Turn.ID
		}
		if payload.Turn != nil && payload.Turn.Status != "" {
			result.Status = string(payload.Turn.Status)
		}
		if payload.Turn != nil && payload.Turn.StartedAt != nil {
			startedAt := time.Unix(*payload.Turn.StartedAt, 0)
			result.CreatedAt = &startedAt
		}
		if payload.Turn != nil && payload.Turn.CompletedAt != nil {
			completedAt := time.Unix(*payload.Turn.CompletedAt, 0)
			result.CompletedAt = &completedAt
		}
		if message := payloadErrorMessage(payload); message != "" {
			result.ErrorMessage = message
		}
		if payload.CreatedAt != nil {
			result.CreatedAt = payload.CreatedAt
		}
		if payload.CompletedAt != nil {
			result.CompletedAt = payload.CompletedAt
		}
	}

	if note.Method == "thread/tokenUsage/updated" && payload.TokenUsage != nil {
		result.TokenUsage = payload.TokenUsage
	}
}

func notificationError(note rpc.Notification) error {
	return turnErrorFromNotification(note, nil)
}

func turnErrorFromNotification(note rpc.Notification, result *TurnResult) error {
	payload, err := parseTurnNotification(note)
	if err != nil {
		if note.Method == "error" || note.Method == "turn/failed" {
			return &TurnError{Result: result, Method: note.Method, Raw: append(json.RawMessage(nil), note.Raw...)}
		}
		return nil
	}
	willRetry := payload.WillRetry != nil && *payload.WillRetry
	if note.Method == "error" && willRetry {
		return nil
	}
	failed := note.Method == "error" || note.Method == "turn/failed" ||
		(note.Method == "turn/completed" && payload.Turn != nil && payload.Turn.Status == protocol.TurnStatusFailed)
	if !failed {
		return nil
	}
	detail := payload.Error
	if detail == nil && payload.Turn != nil {
		detail = payload.Turn.Error
	}
	return &TurnError{
		Result:    result,
		Method:    note.Method,
		Detail:    detail,
		WillRetry: willRetry,
		Raw:       append(json.RawMessage(nil), note.Raw...),
	}
}

func matchesThreadID(note rpc.Notification, threadID string) bool {
	// Some notifications omit threadId; treat those as matching to avoid dropping global events.
	payload, err := parseTurnNotification(note)
	if err != nil || payload.ThreadID == "" {
		return true
	}
	return payload.ThreadID == threadID
}

func extractTextFromItemRaw(raw json.RawMessage) (string, bool) {
	if len(raw) == 0 {
		return "", false
	}
	var direct struct {
		Type string `json:"type"`
		Text string `json:"text"`
	}
	if err := json.Unmarshal(raw, &direct); err == nil && direct.Type == "agentMessage" && direct.Text != "" {
		return direct.Text, true
	}

	var wrapper map[string]json.RawMessage
	if err := json.Unmarshal(raw, &wrapper); err != nil || len(wrapper) != 1 {
		return "", false
	}
	for _, inner := range wrapper {
		var nested struct {
			Type string `json:"type"`
			Text string `json:"text"`
		}
		if err := json.Unmarshal(inner, &nested); err == nil && nested.Type == "agentMessage" && nested.Text != "" {
			return nested.Text, true
		}
	}
	return "", false
}

type turnNotificationPayload struct {
	ThreadID    string                     `json:"threadId,omitempty"`
	TurnID      string                     `json:"turnId,omitempty"`
	Turn        *protocol.Turn             `json:"turn,omitempty"`
	Item        json.RawMessage            `json:"item,omitempty"`
	WillRetry   *bool                      `json:"willRetry,omitempty"`
	Error       *protocol.TurnError        `json:"error,omitempty"`
	TokenUsage  *protocol.ThreadTokenUsage `json:"tokenUsage,omitempty"`
	CreatedAt   *time.Time                 `json:"createdAt,omitempty"`
	CompletedAt *time.Time                 `json:"completedAt,omitempty"`
}

func parseTurnNotification(note rpc.Notification) (turnNotificationPayload, error) {
	if note.Params != nil {
		switch value := note.Params.(type) {
		case protocol.TurnNotification:
			return turnNotificationPayload{ThreadID: value.ThreadID, Turn: value.Turn}, nil
		case *protocol.TurnNotification:
			if value != nil {
				return turnNotificationPayload{ThreadID: value.ThreadID, Turn: value.Turn}, nil
			}
		case protocol.ItemCompletedNotification:
			return turnNotificationPayload{ThreadID: value.ThreadID, TurnID: value.TurnID, Item: value.Item.RawJSON()}, nil
		case *protocol.ItemCompletedNotification:
			if value != nil {
				return turnNotificationPayload{ThreadID: value.ThreadID, TurnID: value.TurnID, Item: value.Item.RawJSON()}, nil
			}
		case protocol.ErrorNotification:
			willRetry := value.WillRetry
			detail := value.Error
			return turnNotificationPayload{ThreadID: value.ThreadID, TurnID: value.TurnID, WillRetry: &willRetry, Error: &detail}, nil
		case *protocol.ErrorNotification:
			if value != nil {
				willRetry := value.WillRetry
				detail := value.Error
				return turnNotificationPayload{ThreadID: value.ThreadID, TurnID: value.TurnID, WillRetry: &willRetry, Error: &detail}, nil
			}
		case protocol.ThreadGoalClearedNotification:
			return turnNotificationPayload{ThreadID: value.ThreadID}, nil
		case *protocol.ThreadGoalClearedNotification:
			if value != nil {
				return turnNotificationPayload{ThreadID: value.ThreadID}, nil
			}
		case protocol.ThreadGoalUpdatedNotification:
			return turnNotificationPayload{ThreadID: value.ThreadID}, nil
		case *protocol.ThreadGoalUpdatedNotification:
			if value != nil {
				return turnNotificationPayload{ThreadID: value.ThreadID}, nil
			}
		case protocol.ThreadTokenUsageUpdatedNotification:
			return turnNotificationPayload{ThreadID: value.ThreadID, TurnID: value.TurnID, TokenUsage: &value.TokenUsage}, nil
		case *protocol.ThreadTokenUsageUpdatedNotification:
			if value != nil {
				return turnNotificationPayload{ThreadID: value.ThreadID, TurnID: value.TurnID, TokenUsage: &value.TokenUsage}, nil
			}
		}
	}

	var payload turnNotificationPayload
	if len(note.Raw) == 0 {
		return payload, nil
	}
	if err := note.UnmarshalParams(&payload); err != nil {
		return payload, err
	}
	return payload, nil
}

func payloadErrorMessage(payload turnNotificationPayload) string {
	if payload.Turn != nil && payload.Turn.Error != nil && payload.Turn.Error.Message != "" {
		return payload.Turn.Error.Message
	}
	if payload.Error != nil && payload.Error.Message != "" {
		return payload.Error.Message
	}
	return ""
}

func buildTurnParams(threadID string, inputs []Input, opts *TurnOptions) (protocol.TurnStartParams, error) {
	params := protocol.TurnStartParams{
		ThreadID: threadID,
		Input:    make([]protocol.TurnStartParamsInputElem, 0, len(inputs)),
	}
	for _, input := range inputs {
		if err := input.validate(); err != nil {
			return params, fmt.Errorf("input: %w", err)
		}
		params.Input = append(params.Input, input)
	}

	if opts == nil {
		return params, nil
	}

	if opts.Cwd != "" {
		params.Cwd = stringPtr(opts.Cwd)
	}
	if opts.ClientUserMessageID != "" {
		params.ClientUserMessageID = stringPtr(opts.ClientUserMessageID)
	}
	if raw, err := normalizeJSONValue("approvalPolicy", opts.ApprovalPolicy); err != nil {
		return params, err
	} else if raw != nil {
		params.ApprovalPolicy = raw
	}
	if raw, err := normalizeJSONValue("sandboxPolicy", opts.SandboxPolicy); err != nil {
		return params, err
	} else if raw != nil {
		params.SandboxPolicy = raw
	}
	if opts.Model != "" {
		params.Model = stringPtr(opts.Model)
	}
	if opts.ServiceTier != "" {
		params.ServiceTier = stringPtr(opts.ServiceTier)
	}
	if raw, err := normalizeJSONValue("effort", opts.Effort); err != nil {
		return params, err
	} else if raw != nil {
		params.Effort = raw
	}
	if raw, err := normalizeJSONValue("summary", opts.Summary); err != nil {
		return params, err
	} else if raw != nil {
		params.Summary = raw
	}
	if raw, err := normalizeJSONValue("outputSchema", opts.OutputSchema); err != nil {
		return params, err
	} else if raw != nil {
		params.OutputSchema = raw
	}
	if opts.CollaborationMode != nil {
		if _, err := normalizeJSONValue("collaborationMode", opts.CollaborationMode); err != nil {
			return params, err
		}
		return params, errors.New("collaboration mode is no longer supported by the current app-server protocol")
	}

	return params, nil
}
