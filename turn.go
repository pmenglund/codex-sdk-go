package codex

import (
	"context"
	"encoding/json"
	"errors"

	"github.com/pmenglund/codex-sdk-go/protocol"
	"github.com/pmenglund/codex-sdk-go/rpc"
)

// TurnOptions configures a turn/start request.
type TurnOptions struct {
	Cwd string
	// ApprovalPolicy is marshaled as JSON and sent as "approvalPolicy".
	// Prefer ApprovalPolicy* constants for simple policies.
	ApprovalPolicy any
	// SandboxPolicy is marshaled as JSON and sent as "sandboxPolicy".
	// Prefer SandboxMode* constants for simple policies.
	SandboxPolicy any
	Model         string
	// Effort is marshaled as JSON and sent as "effort".
	// Prefer ReasoningEffort* constants for standard values.
	Effort any
	// Summary is marshaled as JSON and sent as "summary".
	Summary any
	// OutputSchema is marshaled as JSON and sent as "outputSchema".
	OutputSchema any
	// CollaborationMode is marshaled as JSON and sent as "collaborationMode".
	CollaborationMode any
}

// TurnResult aggregates notifications for a completed turn.
type TurnResult struct {
	TurnID        string
	Notifications []rpc.Notification
	// Items holds the raw JSON payloads for completed items.
	Items         []json.RawMessage
	FinalResponse string
}

// TurnStream iterates notifications for a running turn.
// Notifications that omit threadId are still emitted to avoid dropping
// global events sent during the turn.
type TurnStream struct {
	iter     *rpc.NotificationIterator
	threadID string
}

// Next returns the next notification for this turn.
// Notifications without threadId are treated as belonging to the active stream.
func (s *TurnStream) Next(ctx context.Context) (rpc.Notification, error) {
	if s == nil || s.iter == nil {
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
	if s == nil || s.iter == nil {
		return
	}
	s.iter.Close()
}

func updateTurnResult(result *TurnResult, note rpc.Notification) {
	if note.Method != "item/completed" && note.Method != "turn/started" && note.Method != "turn/completed" && note.Method != "turn/failed" {
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
	}
}

func notificationError(note rpc.Notification) error {
	if note.Method == "error" {
		payload, err := parseTurnNotification(note)
		if err != nil {
			return errors.New("turn error")
		}
		if payload.WillRetry != nil && *payload.WillRetry {
			return nil
		}
		if payload.Error != nil && payload.Error.Message != "" {
			return errors.New(payload.Error.Message)
		}
		return errors.New("turn error")
	}
	if note.Method == "turn/completed" {
		payload, err := parseTurnNotification(note)
		if err != nil {
			return nil
		}
		if payload.Turn != nil && payload.Turn.Status == "failed" {
			if message := payloadErrorMessage(payload); message != "" {
				return errors.New(message)
			}
			return errors.New("turn failed")
		}
	}
	if note.Method == "turn/failed" {
		payload, err := parseTurnNotification(note)
		if err != nil {
			return errors.New("turn failed")
		}
		if message := payloadErrorMessage(payload); message != "" {
			return errors.New(message)
		}
		return errors.New("turn failed")
	}
	return nil
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
		Text string `json:"text"`
	}
	if err := json.Unmarshal(raw, &direct); err == nil && direct.Text != "" {
		return direct.Text, true
	}

	var wrapper map[string]json.RawMessage
	if err := json.Unmarshal(raw, &wrapper); err != nil || len(wrapper) != 1 {
		return "", false
	}
	for _, inner := range wrapper {
		var nested struct {
			Text string `json:"text"`
		}
		if err := json.Unmarshal(inner, &nested); err == nil && nested.Text != "" {
			return nested.Text, true
		}
	}
	return "", false
}

type turnNotificationPayload struct {
	ThreadID  string                          `json:"threadId,omitempty"`
	Turn      *protocol.TurnNotificationTurn  `json:"turn,omitempty"`
	Item      json.RawMessage                 `json:"item,omitempty"`
	WillRetry *bool                           `json:"willRetry,omitempty"`
	Error     *protocol.TurnNotificationError `json:"error,omitempty"`
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
			return turnNotificationPayload{ThreadID: value.ThreadID, Item: value.Item}, nil
		case *protocol.ItemCompletedNotification:
			if value != nil {
				return turnNotificationPayload{ThreadID: value.ThreadID, Item: value.Item}, nil
			}
		case protocol.ErrorNotification:
			return turnNotificationPayload{ThreadID: value.ThreadID, WillRetry: value.WillRetry, Error: value.Error}, nil
		case *protocol.ErrorNotification:
			if value != nil {
				return turnNotificationPayload{ThreadID: value.ThreadID, WillRetry: value.WillRetry, Error: value.Error}, nil
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
		params.Input = append(params.Input, input)
	}

	if opts == nil {
		return params, nil
	}

	if opts.Cwd != "" {
		params.Cwd = stringPtr(opts.Cwd)
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
	if raw, err := normalizeJSONValue("collaborationMode", opts.CollaborationMode); err != nil {
		return params, err
	} else if raw != nil {
		params.CollaborationMode = raw
	}

	return params, nil
}
