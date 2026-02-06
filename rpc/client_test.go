package rpc

import (
	"context"
	"encoding/json"
	"errors"
	"io"
	"strings"
	"sync"
	"testing"
	"time"

	"github.com/pmenglund/codex-sdk-go/protocol"
)

func TestClientCall(t *testing.T) {
	transcript := []TranscriptEntry{
		writeLine(JSONRPCRequest{
			ID:     NewIntRequestID(1),
			Method: "ping",
			Params: mustRaw(map[string]any{"alpha": "a", "beta": 2}),
		}),
		readLine(JSONRPCResponse{
			ID:     NewIntRequestID(1),
			Result: mustRaw(map[string]any{"ok": true}),
		}),
	}

	client := NewClient(NewReplayTransport(transcript), ClientOptions{})
	defer client.Close()

	var result map[string]any
	if err := client.Call(context.Background(), "ping", map[string]any{"alpha": "a", "beta": 2}, &result); err != nil {
		t.Fatalf("call failed: %v", err)
	}
	if result["ok"] != true {
		t.Fatalf("unexpected result: %#v", result)
	}
}

func TestClientCallInvalidParams(t *testing.T) {
	client := NewClient(&stubTransport{}, ClientOptions{})
	defer client.Close()

	var result map[string]any
	if err := client.Call(context.Background(), "ping", map[string]any{"bad": func() {}}, &result); err == nil {
		t.Fatalf("expected marshal error")
	}
}

func TestNotificationDelivery(t *testing.T) {
	transcript := []TranscriptEntry{
		writeLine(JSONRPCRequest{
			ID:     NewIntRequestID(1),
			Method: "ping",
			Params: mustRaw(map[string]any{}),
		}),
		readLine(JSONRPCNotification{
			Method: "turn/started",
			Params: mustRaw(map[string]any{"threadId": "thr_1", "turn": map[string]any{"id": "turn_1"}}),
		}),
		readLine(JSONRPCResponse{
			ID:     NewIntRequestID(1),
			Result: mustRaw(map[string]any{}),
		}),
	}

	client := NewClient(NewReplayTransport(transcript), ClientOptions{})
	defer client.Close()

	iter := client.SubscribeNotifications(1)
	defer iter.Close()

	done := make(chan error, 1)
	go func() {
		var result map[string]any
		done <- client.Call(context.Background(), "ping", map[string]any{}, &result)
	}()

	note, err := iter.Next(context.Background())
	if err != nil {
		t.Fatalf("notification error: %v", err)
	}
	if note.Method != "turn/started" {
		t.Fatalf("unexpected notification: %s", note.Method)
	}

	if err := <-done; err != nil {
		t.Fatalf("call failed: %v", err)
	}
}

func TestNotificationDeliveryDoesNotDropWhenBufferFills(t *testing.T) {
	transport := newChannelTransport()
	client := NewClient(transport, ClientOptions{})
	defer client.Close()

	iter := client.SubscribeNotifications(1)
	defer iter.Close()

	transport.pushReadLine(mustJSON(JSONRPCNotification{
		Method: "turn/started",
		Params: mustRaw(map[string]any{"threadId": "thr_1", "turn": map[string]any{"id": "turn_1"}}),
	}))
	transport.pushReadLine(mustJSON(JSONRPCNotification{
		Method: "turn/completed",
		Params: mustRaw(map[string]any{"threadId": "thr_1", "turn": map[string]any{"id": "turn_1"}}),
	}))

	transport.waitForReads(t, 2)

	ctx, cancel := context.WithTimeout(context.Background(), time.Second)
	defer cancel()
	first, err := iter.Next(ctx)
	if err != nil {
		t.Fatalf("first notification error: %v", err)
	}
	if first.Method != "turn/started" {
		t.Fatalf("unexpected first notification: %s", first.Method)
	}

	ctx2, cancel2 := context.WithTimeout(context.Background(), time.Second)
	defer cancel2()
	second, err := iter.Next(ctx2)
	if err != nil {
		t.Fatalf("second notification error: %v", err)
	}
	if second.Method != "turn/completed" {
		t.Fatalf("unexpected second notification: %s", second.Method)
	}
}

func TestServerRequestDispatch(t *testing.T) {
	resp := protocol.ApplyPatchApprovalResponse(map[string]any{"decision": "approved"})
	handler := &testHandler{
		called: make(chan struct{}, 1),
		applyPatch: func(params protocol.ApplyPatchApprovalParams) (*protocol.ApplyPatchApprovalResponse, error) {
			return &resp, nil
		},
	}

	transcript := []TranscriptEntry{
		readLine(JSONRPCRequest{
			ID:     NewIntRequestID(9),
			Method: "applyPatchApproval",
			Params: mustRaw(map[string]any{"callId": "call", "conversationId": "thr", "fileChanges": map[string]any{}}),
		}),
		writeLine(JSONRPCResponse{
			ID:     NewIntRequestID(9),
			Result: mustRaw(map[string]any{"decision": "approved"}),
		}),
	}

	client := NewClient(NewReplayTransport(transcript), ClientOptions{RequestHandler: handler})
	defer client.Close()

	select {
	case <-handler.called:
	case <-time.After(1 * time.Second):
		t.Fatalf("handler was not called")
	}
}

func TestRecordTransport(t *testing.T) {
	base := &stubTransport{reads: []string{"hello"}}
	recorder := NewRecordTransport(base)

	if err := recorder.WriteLine("ping"); err != nil {
		t.Fatalf("write failed: %v", err)
	}
	if _, err := recorder.ReadLine(); err != nil {
		t.Fatalf("read failed: %v", err)
	}

	transcript := recorder.Transcript()
	if len(transcript) != 2 {
		t.Fatalf("expected 2 transcript entries, got %d", len(transcript))
	}
	if transcript[0].Direction != TranscriptWrite || transcript[0].Line != "ping" {
		t.Fatalf("unexpected write entry: %#v", transcript[0])
	}
	if transcript[1].Direction != TranscriptRead || transcript[1].Line != "hello" {
		t.Fatalf("unexpected read entry: %#v", transcript[1])
	}

	if err := recorder.Close(); err != nil {
		t.Fatalf("close failed: %v", err)
	}
}

func TestReplayTransportMismatch(t *testing.T) {
	replay := NewReplayTransport([]TranscriptEntry{
		{Direction: TranscriptWrite, Line: "expected"},
	})
	if err := replay.WriteLine("different"); err == nil {
		t.Fatalf("expected mismatch error")
	}
}

func TestReplayTransportClosed(t *testing.T) {
	replay := NewReplayTransport([]TranscriptEntry{})
	_ = replay.Close()
	if err := replay.WriteLine("line"); err == nil {
		t.Fatalf("expected error on closed transport")
	}
}

func TestNewRercordTransport(t *testing.T) {
	recorder := NewRercordTransport(&stubTransport{})
	if recorder == nil {
		t.Fatalf("expected recorder")
	}
}

func TestRecordTransportWriteError(t *testing.T) {
	recorder := NewRecordTransport(&errorTransport{})
	if err := recorder.WriteLine("line"); err == nil {
		t.Fatalf("expected write error")
	}
}

func TestNotify(t *testing.T) {
	transcript := []TranscriptEntry{
		writeLine(JSONRPCNotification{
			Method: "notice",
			Params: mustRaw(map[string]any{"ok": true}),
		}),
	}

	client := NewClient(NewReplayTransport(transcript), ClientOptions{})
	defer client.Close()

	if err := client.Notify(context.Background(), "notice", map[string]any{"ok": true}); err != nil {
		t.Fatalf("notify failed: %v", err)
	}
}

func TestCallErrorResponse(t *testing.T) {
	transcript := []TranscriptEntry{
		writeLine(JSONRPCRequest{
			ID:     NewIntRequestID(1),
			Method: "fail",
			Params: mustRaw(map[string]any{}),
		}),
		readLine(JSONRPCError{
			ID: NewIntRequestID(1),
			Error: JSONRPCErrorError{
				Code:    -1,
				Message: "boom",
			},
		}),
	}

	client := NewClient(NewReplayTransport(transcript), ClientOptions{})
	defer client.Close()

	var result map[string]any
	err := client.Call(context.Background(), "fail", map[string]any{}, &result)
	if err == nil {
		t.Fatalf("expected error")
	}
	if !strings.Contains(err.Error(), "boom") {
		t.Fatalf("unexpected error: %v", err)
	}
}

func TestCallContextCancel(t *testing.T) {
	transcript := []TranscriptEntry{
		writeLine(JSONRPCRequest{
			ID:     NewIntRequestID(1),
			Method: "ping",
			Params: mustRaw(map[string]any{}),
		}),
	}
	client := NewClient(NewReplayTransport(transcript), ClientOptions{})
	defer client.Close()

	ctx, cancel := context.WithCancel(context.Background())
	cancel()
	var result map[string]any
	if err := client.Call(ctx, "ping", map[string]any{}, &result); err == nil {
		t.Fatalf("expected context error")
	}
}

func TestCallAfterClose(t *testing.T) {
	client := NewClient(NewReplayTransport(nil), ClientOptions{})
	_ = client.Close()
	var result map[string]any
	if err := client.Call(context.Background(), "ping", map[string]any{}, &result); err == nil {
		t.Fatalf("expected error after close")
	}
}

func TestNotifyContextCancel(t *testing.T) {
	client := NewClient(NewReplayTransport(nil), ClientOptions{})
	defer client.Close()

	ctx, cancel := context.WithCancel(context.Background())
	cancel()
	if err := client.Notify(ctx, "notice", map[string]any{}); err == nil {
		t.Fatalf("expected context error")
	}
}

func TestDispatchServerRequestUnknown(t *testing.T) {
	handler := &recordingHandler{}
	req := JSONRPCRequest{ID: NewIntRequestID(1), Method: "unknown"}
	if _, err := dispatchServerRequest(context.Background(), handler, req); err == nil {
		t.Fatalf("expected error for unknown method")
	}
}

type testHandler struct {
	called     chan struct{}
	applyPatch func(protocol.ApplyPatchApprovalParams) (*protocol.ApplyPatchApprovalResponse, error)
}

func (h *testHandler) ApplyPatchApproval(ctx context.Context, params protocol.ApplyPatchApprovalParams) (*protocol.ApplyPatchApprovalResponse, error) {
	if h.called == nil {
		h.called = make(chan struct{}, 1)
	}
	h.called <- struct{}{}
	if h.applyPatch != nil {
		return h.applyPatch(params)
	}
	resp := protocol.ApplyPatchApprovalResponse(map[string]any{"decision": "approved"})
	return &resp, nil
}

func (h *testHandler) ExecCommandApproval(ctx context.Context, params protocol.ExecCommandApprovalParams) (*protocol.ExecCommandApprovalResponse, error) {
	return nil, errors.New("not implemented")
}

func (h *testHandler) ItemCommandExecutionRequestApproval(ctx context.Context, params protocol.CommandExecutionRequestApprovalParams) (*protocol.CommandExecutionRequestApprovalResponse, error) {
	return nil, errors.New("not implemented")
}

func (h *testHandler) ItemFileChangeRequestApproval(ctx context.Context, params protocol.FileChangeRequestApprovalParams) (*protocol.FileChangeRequestApprovalResponse, error) {
	return nil, errors.New("not implemented")
}

func (h *testHandler) ItemToolRequestUserInput(ctx context.Context, params protocol.ToolRequestUserInputParams) (*protocol.ToolRequestUserInputResponse, error) {
	return nil, errors.New("not implemented")
}

type stubTransport struct {
	reads  []string
	writes []string
}

type channelTransport struct {
	mu       sync.Mutex
	reads    chan string
	observed chan struct{}
	writes   []string
	closed   sync.Once
}

func newChannelTransport() *channelTransport {
	return &channelTransport{
		reads:    make(chan string, 16),
		observed: make(chan struct{}, 16),
	}
}

func (t *channelTransport) pushReadLine(line string) {
	t.reads <- line
}

func (t *channelTransport) waitForReads(testingT *testing.T, count int) {
	testingT.Helper()
	for i := 0; i < count; i++ {
		select {
		case <-t.observed:
		case <-time.After(time.Second):
			testingT.Fatalf("timed out waiting for read %d", i+1)
		}
	}
}

type errorTransport struct{}

func (e *errorTransport) ReadLine() (string, error) {
	return "", io.EOF
}

func (e *errorTransport) WriteLine(line string) error {
	return errors.New("write failed")
}

func (e *errorTransport) Close() error {
	return nil
}

func (s *stubTransport) ReadLine() (string, error) {
	if len(s.reads) == 0 {
		return "", io.EOF
	}
	line := s.reads[0]
	s.reads = s.reads[1:]
	return line, nil
}

func (s *stubTransport) WriteLine(line string) error {
	s.writes = append(s.writes, line)
	return nil
}

func (s *stubTransport) Close() error {
	return nil
}

func (t *channelTransport) ReadLine() (string, error) {
	line, ok := <-t.reads
	if !ok {
		return "", io.EOF
	}
	t.observed <- struct{}{}
	return line, nil
}

func (t *channelTransport) WriteLine(line string) error {
	t.mu.Lock()
	t.writes = append(t.writes, line)
	t.mu.Unlock()
	return nil
}

func (t *channelTransport) Close() error {
	t.closed.Do(func() {
		close(t.reads)
	})
	return nil
}

func writeLine(payload any) TranscriptEntry {
	return TranscriptEntry{Direction: TranscriptWrite, Line: mustJSON(payload)}
}

func readLine(payload any) TranscriptEntry {
	return TranscriptEntry{Direction: TranscriptRead, Line: mustJSON(payload)}
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
