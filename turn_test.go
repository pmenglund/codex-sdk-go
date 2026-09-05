package codex

import (
	"context"
	"encoding/json"
	"errors"
	"testing"
	"time"

	"github.com/pmenglund/codex-sdk-go/protocol"
	"github.com/pmenglund/codex-sdk-go/rpc"
)

func TestThreadRunWithReplay(t *testing.T) {
	ctx := context.Background()
	info := protocol.ClientInfo{
		Name:    "codex-go-test",
		Title:   stringPtr("Codex Go SDK Test"),
		Version: "test",
	}

	client, err := New(ctx, Options{
		Transport:  rpc.NewReplayTransport(runTranscript(info, "hello", "final")),
		ClientInfo: info,
	})
	if err != nil {
		t.Fatalf("new client error: %v", err)
	}
	defer client.Close()

	thread, err := client.StartThread(ctx, ThreadStartOptions{})
	if err != nil {
		t.Fatalf("start thread error: %v", err)
	}

	result, err := thread.Run(ctx, "hello", nil)
	if err != nil {
		t.Fatalf("run error: %v", err)
	}
	if result.FinalResponse != "final" {
		t.Fatalf("unexpected final response: %s", result.FinalResponse)
	}
}

func TestThreadRunFailsOnTurnFailedNotification(t *testing.T) {
	ctx := context.Background()
	info := protocol.ClientInfo{
		Name:    "codex-go-test",
		Title:   stringPtr("Codex Go SDK Test"),
		Version: "test",
	}

	client, err := New(ctx, Options{
		Transport:  rpc.NewReplayTransport(runFailedTranscript(info, "hello", "boom")),
		ClientInfo: info,
	})
	if err != nil {
		t.Fatalf("new client error: %v", err)
	}
	defer client.Close()

	thread, err := client.StartThread(ctx, ThreadStartOptions{})
	if err != nil {
		t.Fatalf("start thread error: %v", err)
	}

	result, err := thread.Run(ctx, "hello", nil)
	if err == nil || err.Error() != "boom" {
		t.Fatalf("expected boom error, got %v", err)
	}
	if result == nil || result.TurnID != "turn_1" || result.Status != "failed" {
		t.Fatalf("expected partial failed result, got %#v", result)
	}
	if !errors.Is(err, ErrTurnFailed) {
		t.Fatalf("expected ErrTurnFailed, got %v", err)
	}
	var turnErr *TurnError
	if !errors.As(err, &turnErr) || turnErr.Result != result || turnErr.Detail == nil || turnErr.Detail.Message != "boom" {
		t.Fatalf("expected structured TurnError, got %#v", err)
	}
}

func TestThreadRunFailsOnCompletedFailedStatus(t *testing.T) {
	ctx := context.Background()
	info := protocol.ClientInfo{
		Name:    "codex-go-test",
		Title:   stringPtr("Codex Go SDK Test"),
		Version: "test",
	}

	client, err := New(ctx, Options{
		Transport:  rpc.NewReplayTransport(runCompletedFailedTranscript(info, "hello", "completed boom")),
		ClientInfo: info,
	})
	if err != nil {
		t.Fatalf("new client error: %v", err)
	}
	defer client.Close()

	thread, err := client.StartThread(ctx, ThreadStartOptions{})
	if err != nil {
		t.Fatalf("start thread error: %v", err)
	}

	result, err := thread.Run(ctx, "hello", nil)
	if err == nil || err.Error() != "completed boom" {
		t.Fatalf("expected completed boom error, got %v", err)
	}
	if result == nil || result.ErrorMessage != "completed boom" {
		t.Fatalf("expected partial completed-failure result, got %#v", result)
	}
}

func TestTurnHandleFailedResultRetainsAccumulatedDetails(t *testing.T) {
	client := rpc.NewClient(rpc.NewReplayTransport(nil), rpc.ClientOptions{})
	defer client.Close()
	notes := []rpc.Notification{
		{Method: "turn/started", Raw: mustRaw(map[string]any{"threadId": "thr_1", "turn": map[string]any{"id": "turn_1", "items": []any{}, "status": "inProgress", "startedAt": 10}})},
		{Method: "item/completed", Raw: mustRaw(map[string]any{"threadId": "thr_1", "turnId": "turn_1", "completedAtMs": 11, "item": map[string]any{"type": "agentMessage", "id": "item_1", "text": "partial"}})},
		{Method: "thread/tokenUsage/updated", Raw: mustRaw(map[string]any{"threadId": "thr_1", "turnId": "turn_1", "tokenUsage": map[string]any{"last": tokenUsageBreakdown(1, 2, 3), "total": tokenUsageBreakdown(4, 5, 9)}})},
		{Method: "error", Raw: mustRaw(map[string]any{"threadId": "thr_1", "turnId": "turn_1", "willRetry": true, "error": map[string]any{"message": "retrying"}})},
		{Method: "turn/failed", Raw: mustRaw(map[string]any{"threadId": "thr_1", "turn": map[string]any{"id": "turn_1", "items": []any{}, "status": "failed", "completedAt": 20, "error": map[string]any{"message": "terminal", "additionalDetails": "details"}}})},
	}
	index := 0
	handle := &TurnHandle{
		client:   client,
		threadID: "thr_1",
		stream: &TurnStream{next: func(context.Context) (rpc.Notification, error) {
			note := notes[index]
			index++
			return note, nil
		}},
	}
	result, err := handle.Run(context.Background())
	var turnErr *TurnError
	if !errors.As(err, &turnErr) || !errors.Is(err, ErrTurnFailed) {
		t.Fatalf("expected typed terminal failure, got %v", err)
	}
	if result == nil || turnErr.Result != result || result.FinalResponse != "partial" || len(result.Items) != 1 || len(result.Notifications) != len(notes) {
		t.Fatalf("partial result was not retained: %#v", result)
	}
	if result.TokenUsage == nil || result.TokenUsage.Total.TotalTokens != 9 || result.CreatedAt == nil || result.CreatedAt.Unix() != 10 || result.CompletedAt == nil || result.CompletedAt.Unix() != 20 {
		t.Fatalf("usage or timestamps were not retained: %#v", result)
	}
	if turnErr.Detail == nil || turnErr.Detail.Message != "terminal" || turnErr.WillRetry || len(turnErr.Raw) == 0 {
		t.Fatalf("terminal details were not retained: %#v", turnErr)
	}
}

func TestTurnHandleTerminalErrorRetainsAccumulatedResult(t *testing.T) {
	client := rpc.NewClient(rpc.NewReplayTransport(nil), rpc.ClientOptions{})
	defer client.Close()
	notes := []rpc.Notification{
		{Method: "item/completed", Raw: mustRaw(map[string]any{"threadId": "thr_1", "turnId": "turn_1", "completedAtMs": 11, "item": map[string]any{"type": "agentMessage", "id": "item_1", "text": "partial"}})},
		{Method: "error", Raw: mustRaw(map[string]any{"threadId": "thr_1", "turnId": "turn_1", "willRetry": false, "error": map[string]any{"message": "terminal error", "additionalDetails": "details"}})},
	}
	index := 0
	handle := &TurnHandle{
		client:   client,
		threadID: "thr_1",
		stream: &TurnStream{next: func(context.Context) (rpc.Notification, error) {
			note := notes[index]
			index++
			return note, nil
		}},
	}

	result, err := handle.Run(context.Background())
	var turnErr *TurnError
	if !errors.As(err, &turnErr) || !errors.Is(err, ErrTurnFailed) {
		t.Fatalf("expected typed terminal error notification, got %v", err)
	}
	if result == nil || turnErr.Result != result || result.FinalResponse != "partial" || len(result.Items) != 1 || len(result.Notifications) != len(notes) {
		t.Fatalf("terminal error discarded accumulated result: %#v", result)
	}
	if turnErr.Method != "error" || turnErr.WillRetry || turnErr.Detail == nil || turnErr.Detail.Message != "terminal error" {
		t.Fatalf("terminal error details were not retained: %#v", turnErr)
	}
}

func TestTurnResultUsesOnlyAgentMessageAsFinalResponse(t *testing.T) {
	result := &TurnResult{}
	for _, item := range []map[string]any{
		{"type": "plan", "id": "plan_1", "text": "intermediate plan"},
		{"type": "agentMessage", "id": "message_1", "text": "final answer"},
		{"type": "reasoning", "id": "reasoning_1", "text": "internal text"},
	} {
		raw := mustRaw(map[string]any{"threadId": "thr_1", "turnId": "turn_1", "completedAtMs": 1, "item": item})
		updateTurnResult(result, rpc.Notification{Method: "item/completed", Raw: raw})
	}
	if result.FinalResponse != "final answer" {
		t.Fatalf("unexpected final response: %q", result.FinalResponse)
	}
	if len(result.Items) != 3 {
		t.Fatalf("expected every raw item, got %d", len(result.Items))
	}
}

func TestThreadRunReturnsStreamNextError(t *testing.T) {
	ctx := context.Background()
	info := protocol.ClientInfo{
		Name:    "codex-go-test",
		Title:   stringPtr("Codex Go SDK Test"),
		Version: "test",
	}

	client, err := New(ctx, Options{
		Transport:  rpc.NewReplayTransport(runWithoutCompletionTranscript(info, "hello")),
		ClientInfo: info,
	})
	if err != nil {
		t.Fatalf("new client error: %v", err)
	}
	defer client.Close()

	thread, err := client.StartThread(ctx, ThreadStartOptions{})
	if err != nil {
		t.Fatalf("start thread error: %v", err)
	}

	runCtx, cancel := context.WithTimeout(ctx, 10*time.Millisecond)
	defer cancel()
	_, err = thread.Run(runCtx, "hello", nil)
	if err == nil || err != context.DeadlineExceeded {
		t.Fatalf("expected context deadline exceeded error, got %v", err)
	}
}

func TestResumeThreadWithReplay(t *testing.T) {
	ctx := context.Background()
	info := protocol.ClientInfo{
		Name:    "codex-go-test",
		Title:   stringPtr("Codex Go SDK Test"),
		Version: "test",
	}

	client, err := New(ctx, Options{
		Transport:  rpc.NewReplayTransport(resumeTranscript(info)),
		ClientInfo: info,
	})
	if err != nil {
		t.Fatalf("new client error: %v", err)
	}
	defer client.Close()

	thread, err := client.ResumeThread(ctx, ThreadResumeOptions{ThreadID: "thr_123"})
	if err != nil {
		t.Fatalf("resume thread error: %v", err)
	}
	if thread.ID() != "thr_123" {
		t.Fatalf("unexpected thread id: %s", thread.ID())
	}
}

func TestCloseNilClient(t *testing.T) {
	c := &Codex{}
	if err := c.Close(); err == nil {
		t.Fatalf("expected error for nil client")
	}
}

func runTranscript(info protocol.ClientInfo, prompt, finalResponse string) []rpc.TranscriptEntry {
	return []rpc.TranscriptEntry{
		writeLine(rpc.JSONRPCRequest{
			ID:     rpc.NewIntRequestID(1),
			Method: "initialize",
			Params: mustRaw(protocol.InitializeParams{ClientInfo: info}),
		}),
		readLine(rpc.JSONRPCResponse{
			ID:     rpc.NewIntRequestID(1),
			Result: mustRaw(map[string]any{}),
		}),
		writeLine(rpc.JSONRPCNotification{Method: "initialized"}),
		writeLine(rpc.JSONRPCRequest{
			ID:     rpc.NewIntRequestID(2),
			Method: "thread/start",
			Params: mustRaw(map[string]any{}),
		}),
		readLine(rpc.JSONRPCResponse{
			ID:     rpc.NewIntRequestID(2),
			Result: mustRaw(map[string]any{"thread": map[string]any{"id": "thr_123"}}),
		}),
		writeLine(rpc.JSONRPCRequest{
			ID:     rpc.NewIntRequestID(3),
			Method: "turn/start",
			Params: mustRaw(turnStartParams(prompt)),
		}),
		readLine(rpc.JSONRPCResponse{
			ID:     rpc.NewIntRequestID(3),
			Result: mustRaw(map[string]any{"turn": turnPayload("turn_1", "inProgress")}),
		}),
		readLine(rpc.JSONRPCNotification{
			Method: "turn/started",
			Params: mustRaw(map[string]any{"threadId": "thr_123", "turn": turnPayload("turn_1", "inProgress")}),
		}),
		readLine(rpc.JSONRPCNotification{
			Method: "item/completed",
			Params: mustRaw(map[string]any{"threadId": "thr_123", "turnId": "turn_1", "completedAtMs": 1, "item": map[string]any{"type": "agentMessage", "id": "item_1", "text": finalResponse}}),
		}),
		readLine(rpc.JSONRPCNotification{
			Method: "turn/completed",
			Params: mustRaw(map[string]any{"threadId": "thr_123", "turn": turnPayload("turn_1", "completed")}),
		}),
	}
}

func runFailedTranscript(info protocol.ClientInfo, prompt, failureMessage string) []rpc.TranscriptEntry {
	return []rpc.TranscriptEntry{
		writeLine(rpc.JSONRPCRequest{
			ID:     rpc.NewIntRequestID(1),
			Method: "initialize",
			Params: mustRaw(protocol.InitializeParams{ClientInfo: info}),
		}),
		readLine(rpc.JSONRPCResponse{
			ID:     rpc.NewIntRequestID(1),
			Result: mustRaw(map[string]any{}),
		}),
		writeLine(rpc.JSONRPCNotification{Method: "initialized"}),
		writeLine(rpc.JSONRPCRequest{
			ID:     rpc.NewIntRequestID(2),
			Method: "thread/start",
			Params: mustRaw(map[string]any{}),
		}),
		readLine(rpc.JSONRPCResponse{
			ID:     rpc.NewIntRequestID(2),
			Result: mustRaw(map[string]any{"thread": map[string]any{"id": "thr_123"}}),
		}),
		writeLine(rpc.JSONRPCRequest{
			ID:     rpc.NewIntRequestID(3),
			Method: "turn/start",
			Params: mustRaw(turnStartParams(prompt)),
		}),
		readLine(rpc.JSONRPCResponse{
			ID:     rpc.NewIntRequestID(3),
			Result: mustRaw(map[string]any{"turn": turnPayload("turn_1", "inProgress")}),
		}),
		readLine(rpc.JSONRPCNotification{
			Method: "turn/started",
			Params: mustRaw(map[string]any{"threadId": "thr_123", "turn": turnPayload("turn_1", "inProgress")}),
		}),
		readLine(rpc.JSONRPCNotification{
			Method: "turn/failed",
			Params: mustRaw(map[string]any{
				"threadId": "thr_123",
				"turn": map[string]any{
					"id":     "turn_1",
					"status": "failed",
					"error":  map[string]any{"message": failureMessage},
				},
			}),
		}),
	}
}

func runCompletedFailedTranscript(info protocol.ClientInfo, prompt, failureMessage string) []rpc.TranscriptEntry {
	entries := runTranscript(info, prompt, "partial")
	entries[len(entries)-1] = readLine(rpc.JSONRPCNotification{
		Method: "turn/completed",
		Params: mustRaw(map[string]any{
			"threadId": "thr_123",
			"turn": map[string]any{
				"id":     "turn_1",
				"status": "failed",
				"error":  map[string]any{"message": failureMessage},
			},
		}),
	})
	return entries
}

func runWithoutCompletionTranscript(info protocol.ClientInfo, prompt string) []rpc.TranscriptEntry {
	return []rpc.TranscriptEntry{
		writeLine(rpc.JSONRPCRequest{
			ID:     rpc.NewIntRequestID(1),
			Method: "initialize",
			Params: mustRaw(protocol.InitializeParams{ClientInfo: info}),
		}),
		readLine(rpc.JSONRPCResponse{
			ID:     rpc.NewIntRequestID(1),
			Result: mustRaw(map[string]any{}),
		}),
		writeLine(rpc.JSONRPCNotification{Method: "initialized"}),
		writeLine(rpc.JSONRPCRequest{
			ID:     rpc.NewIntRequestID(2),
			Method: "thread/start",
			Params: mustRaw(map[string]any{}),
		}),
		readLine(rpc.JSONRPCResponse{
			ID:     rpc.NewIntRequestID(2),
			Result: mustRaw(map[string]any{"thread": map[string]any{"id": "thr_123"}}),
		}),
		writeLine(rpc.JSONRPCRequest{
			ID:     rpc.NewIntRequestID(3),
			Method: "turn/start",
			Params: mustRaw(turnStartParams(prompt)),
		}),
		readLine(rpc.JSONRPCResponse{
			ID:     rpc.NewIntRequestID(3),
			Result: mustRaw(map[string]any{"turn": turnPayload("turn_1", "inProgress")}),
		}),
		writeLine(rpc.JSONRPCRequest{
			ID:     rpc.NewIntRequestID(4),
			Method: "turn/interrupt",
			Params: mustRaw(map[string]any{"threadId": "thr_123", "turnId": "turn_1"}),
		}),
		readLine(rpc.JSONRPCResponse{ID: rpc.NewIntRequestID(4), Result: mustRaw(map[string]any{})}),
	}
}

func resumeTranscript(info protocol.ClientInfo) []rpc.TranscriptEntry {
	return []rpc.TranscriptEntry{
		writeLine(rpc.JSONRPCRequest{
			ID:     rpc.NewIntRequestID(1),
			Method: "initialize",
			Params: mustRaw(protocol.InitializeParams{ClientInfo: info}),
		}),
		readLine(rpc.JSONRPCResponse{
			ID:     rpc.NewIntRequestID(1),
			Result: mustRaw(map[string]any{}),
		}),
		writeLine(rpc.JSONRPCNotification{Method: "initialized"}),
		writeLine(rpc.JSONRPCRequest{
			ID:     rpc.NewIntRequestID(2),
			Method: "thread/resume",
			Params: mustRaw(map[string]any{"threadId": "thr_123"}),
		}),
		readLine(rpc.JSONRPCResponse{
			ID:     rpc.NewIntRequestID(2),
			Result: mustRaw(map[string]any{"thread": map[string]any{"id": "thr_123"}}),
		}),
	}
}

func turnStartParams(prompt string) map[string]any {
	return map[string]any{
		"threadId": "thr_123",
		"input":    []Input{TextInput(prompt)},
	}
}

func turnPayload(turnID, status string) map[string]any {
	return map[string]any{
		"id":     turnID,
		"status": status,
		"items":  []any{},
		"error":  nil,
	}
}

func writeLine(payload any) rpc.TranscriptEntry {
	return rpc.TranscriptEntry{Direction: rpc.TranscriptWrite, Line: mustJSON(payload)}
}

func readLine(payload any) rpc.TranscriptEntry {
	return rpc.TranscriptEntry{Direction: rpc.TranscriptRead, Line: mustJSON(payload)}
}

func mustJSON(payload any) string {
	data, err := json.Marshal(payload)
	if err != nil {
		panic(err)
	}
	return string(data)
}

func mustRaw(payload any) json.RawMessage {
	if payload == nil {
		return nil
	}
	data, err := json.Marshal(payload)
	if err != nil {
		panic(err)
	}
	return data
}
