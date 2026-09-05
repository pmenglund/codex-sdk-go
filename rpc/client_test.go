package rpc

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"io"
	"log/slog"
	"strings"
	"sync"
	"sync/atomic"
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

func TestNotificationOverflowIsBoundedAndIsolated(t *testing.T) {
	transport := newChannelTransport()
	client := NewClient(transport, ClientOptions{})
	defer client.Close()

	slow := client.SubscribeNotifications(1)
	defer slow.Close()
	healthy := client.SubscribeNotifications(4)
	defer healthy.Close()

	callDone := make(chan error, 1)
	go func() {
		var result map[string]any
		callDone <- client.Call(context.Background(), "ping", map[string]any{}, &result)
	}()
	transport.waitForWrites(t, 1)

	transport.pushReadLine(mustJSON(JSONRPCNotification{
		Method: "turn/started",
		Params: mustRaw(map[string]any{"threadId": "thr_1", "turn": map[string]any{"id": "turn_1"}}),
	}))
	transport.pushReadLine(mustJSON(JSONRPCNotification{
		Method: "turn/completed",
		Params: mustRaw(map[string]any{"threadId": "thr_1", "turn": map[string]any{"id": "turn_1"}}),
	}))
	transport.pushReadLine(mustJSON(JSONRPCResponse{
		ID:     NewIntRequestID(1),
		Result: mustRaw(map[string]any{"ok": true}),
	}))

	transport.waitForReads(t, 3)

	ctx, cancel := context.WithTimeout(context.Background(), time.Second)
	defer cancel()
	_, err := slow.Next(ctx)
	var overflow *NotificationOverflowError
	if !errors.As(err, &overflow) {
		t.Fatalf("expected NotificationOverflowError, got %v", err)
	}
	if overflow.Capacity != 1 {
		t.Fatalf("unexpected overflow capacity: %d", overflow.Capacity)
	}
	if !errors.Is(err, ErrNotificationOverflow) {
		t.Fatalf("expected ErrNotificationOverflow compatibility, got %v", err)
	}

	for _, want := range []string{"turn/started", "turn/completed"} {
		note, nextErr := healthy.Next(ctx)
		if nextErr != nil {
			t.Fatalf("healthy notification error: %v", nextErr)
		}
		if note.Method != want {
			t.Fatalf("unexpected healthy notification: got %s want %s", note.Method, want)
		}
	}

	select {
	case callErr := <-callDone:
		if callErr != nil {
			t.Fatalf("call failed after another subscriber overflowed: %v", callErr)
		}
	case <-time.After(time.Second):
		t.Fatalf("reader was blocked by overflowing subscriber")
	}

	client.subsMu.Lock()
	subCount := len(client.subs)
	client.subsMu.Unlock()
	if subCount != 1 {
		t.Fatalf("expected only healthy subscriber to remain, got %d", subCount)
	}
}

func TestNotificationPublishCloseRace(t *testing.T) {
	for iteration := 0; iteration < 100; iteration++ {
		transport := newChannelTransport()
		client := NewClient(transport, ClientOptions{})
		iter := client.SubscribeNotifications(2)
		note := JSONRPCNotification{Method: "turn/started", Params: mustRaw(map[string]any{"threadId": "thr_1"})}

		var wg sync.WaitGroup
		wg.Add(3)
		go func() {
			defer wg.Done()
			for publish := 0; publish < 100; publish++ {
				client.handleNotification(note)
			}
		}()
		go func() {
			defer wg.Done()
			for closeCall := 0; closeCall < 10; closeCall++ {
				iter.Close()
			}
		}()
		go func() {
			defer wg.Done()
			_ = client.Close()
		}()

		done := make(chan struct{})
		go func() {
			wg.Wait()
			close(done)
		}()
		select {
		case <-done:
		case <-time.After(time.Second):
			t.Fatalf("publish/close race did not finish at iteration %d", iteration)
		}
		ctx, cancel := context.WithTimeout(context.Background(), time.Second)
		_, err := iter.Next(ctx)
		cancel()
		if err == nil {
			t.Fatalf("expected terminal iterator error at iteration %d", iteration)
		}
	}
}

func TestSubscribeAfterClientCloseIsNotRetained(t *testing.T) {
	client := NewClient(newChannelTransport(), ClientOptions{})
	if err := client.Close(); err != nil {
		t.Fatalf("close client: %v", err)
	}
	iter := client.SubscribeNotifications(1)
	client.subsMu.Lock()
	subCount := len(client.subs)
	client.subsMu.Unlock()
	if subCount != 0 {
		t.Fatalf("closed client retained %d subscriptions", subCount)
	}
	ctx, cancel := context.WithTimeout(context.Background(), time.Second)
	defer cancel()
	if _, err := iter.Next(ctx); err == nil {
		t.Fatalf("expected closed-client iterator error")
	}
}

func TestServerRequestDispatch(t *testing.T) {
	resp := protocol.ApplyPatchApprovalResponse{Decision: protocol.MustReviewDecision("approved")}
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

func TestServerRequestHandlerDoesNotBlockReaderAndReceivesCloseContext(t *testing.T) {
	transport := newChannelTransport()
	handler := &blockingServerRequestHandler{
		entered: make(chan struct{}),
		done:    make(chan error, 1),
	}
	client := NewClient(transport, ClientOptions{RequestHandler: handler})

	callDone := make(chan error, 1)
	go func() {
		var result map[string]any
		callDone <- client.Call(context.Background(), "ping", map[string]any{}, &result)
	}()
	transport.waitForWrites(t, 1)

	transport.pushReadLine(mustJSON(JSONRPCRequest{
		ID:     NewIntRequestID(9),
		Method: "applyPatchApproval",
		Params: mustRaw(map[string]any{"callId": "call", "conversationId": "thr", "fileChanges": map[string]any{}}),
	}))

	select {
	case <-handler.entered:
	case <-time.After(time.Second):
		t.Fatalf("handler was not called")
	}

	transport.pushReadLine(mustJSON(JSONRPCResponse{
		ID:     NewIntRequestID(1),
		Result: mustRaw(map[string]any{"ok": true}),
	}))

	select {
	case err := <-callDone:
		if err != nil {
			t.Fatalf("call failed while handler was blocked: %v", err)
		}
	case <-time.After(time.Second):
		t.Fatalf("reader was blocked by server request handler")
	}

	if err := client.Close(); err != nil {
		t.Fatalf("close failed: %v", err)
	}
	select {
	case err := <-handler.done:
		if !errors.Is(err, context.Canceled) {
			t.Fatalf("expected canceled handler context, got %v", err)
		}
	case <-time.After(time.Second):
		t.Fatalf("handler did not observe close context")
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
			Error: JSONRPCErrorDetail{
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
	transport := newChannelTransport()
	client := NewClient(transport, ClientOptions{})
	defer client.Close()

	ctx, cancel := context.WithCancel(context.Background())
	cancel()
	var result map[string]any
	if err := client.Call(ctx, "ping", map[string]any{}, &result); err == nil {
		t.Fatalf("expected context error")
	}

	transport.mu.Lock()
	writes := len(transport.writes)
	transport.mu.Unlock()
	if writes != 0 {
		t.Fatalf("expected no writes for canceled call, got %d", writes)
	}
}

func TestCallContextCancelAfterSend(t *testing.T) {
	transport := newChannelTransport()
	client := NewClient(transport, ClientOptions{})
	defer client.Close()

	ctx, cancel := context.WithCancel(context.Background())
	done := make(chan error, 1)
	go func() {
		var result map[string]any
		done <- client.Call(ctx, "ping", map[string]any{}, &result)
	}()
	transport.waitForWrites(t, 1)
	cancel()

	select {
	case err := <-done:
		if !errors.Is(err, context.Canceled) {
			t.Fatalf("expected context canceled, got %v", err)
		}
	case <-time.After(time.Second):
		t.Fatalf("call did not return after context cancellation")
	}
}

func TestBlockedWriteHonorsCallerContext(t *testing.T) {
	for _, test := range []struct {
		name string
		run  func(*Client, context.Context) error
	}{
		{
			name: "call",
			run: func(client *Client, ctx context.Context) error {
				var result map[string]any
				return client.Call(ctx, "ping", map[string]any{}, &result)
			},
		},
		{
			name: "notify",
			run: func(client *Client, ctx context.Context) error {
				return client.Notify(ctx, "notice", map[string]any{})
			},
		},
	} {
		t.Run(test.name, func(t *testing.T) {
			transport := newBlockingWriteTransport()
			client := NewClient(transport, ClientOptions{})

			ctx, cancel := context.WithTimeout(context.Background(), 20*time.Millisecond)
			defer cancel()
			done := make(chan error, 1)
			go func() { done <- test.run(client, ctx) }()

			select {
			case <-transport.started:
			case <-time.After(time.Second):
				t.Fatalf("write did not start")
			}
			select {
			case err := <-done:
				if !errors.Is(err, context.DeadlineExceeded) {
					t.Fatalf("expected deadline exceeded, got %v", err)
				}
			case <-time.After(time.Second):
				t.Fatalf("caller remained blocked on transport write")
			}
			if err := client.Close(); err != nil {
				t.Fatalf("close client: %v", err)
			}
		})
	}
}

func TestOutboundWritesAreSerialized(t *testing.T) {
	transport := newSerialWriteTransport()
	client := NewClient(transport, ClientOptions{})
	defer client.Close()

	const writes = 20
	var wg sync.WaitGroup
	wg.Add(writes)
	for index := range writes {
		go func() {
			defer wg.Done()
			if err := client.Notify(context.Background(), "notice", map[string]any{"index": index}); err != nil {
				t.Errorf("notify: %v", err)
			}
		}()
	}
	wg.Wait()
	if got := transport.maxConcurrent.Load(); got != 1 {
		t.Fatalf("writes were not serialized: max concurrent=%d", got)
	}
}

func TestContextTransportReceivesWriteContext(t *testing.T) {
	transport := newContextWriteTransport()
	client := NewClient(transport, ClientOptions{})
	defer client.Close()

	ctx, cancel := context.WithTimeout(context.Background(), 20*time.Millisecond)
	defer cancel()
	if err := client.Notify(ctx, "notice", nil); !errors.Is(err, context.DeadlineExceeded) {
		t.Fatalf("expected deadline exceeded, got %v", err)
	}
	select {
	case <-transport.contextWrite:
	case <-time.After(time.Second):
		t.Fatalf("context-aware write method was not used")
	}
	if err := client.Notify(context.Background(), "after/canceled/write", nil); !errors.Is(err, context.DeadlineExceeded) {
		t.Fatalf("context-aware write error was not terminal: %v", err)
	}
}

func TestClientRejectsOversizedCustomTransportMessage(t *testing.T) {
	for name, line := range map[string]string{
		"json":       "12345",
		"whitespace": "     ",
	} {
		t.Run(name, func(t *testing.T) {
			client := NewClient(&stubTransport{reads: []string{line}}, ClientOptions{MaxMessageBytes: 4})
			defer client.Close()
			select {
			case <-client.done:
				if !errors.Is(client.errOrClosed(), ErrMessageTooLarge) {
					t.Fatalf("expected ErrMessageTooLarge, got %v", client.errOrClosed())
				}
			case <-time.After(time.Second):
				t.Fatal("client did not reject oversized message")
			}
		})
	}
}

func TestServerRequestQueueIsBounded(t *testing.T) {
	transport := newChannelTransport()
	handler := &queueBlockingHandler{entered: make(chan struct{}, 1), release: make(chan struct{})}
	client := NewClient(transport, ClientOptions{
		RequestHandler:             handler,
		ServerRequestWorkers:       1,
		ServerRequestQueueCapacity: 1,
	})
	defer client.Close()

	request := func(id int64) string {
		return mustJSON(JSONRPCRequest{
			ID:     NewIntRequestID(id),
			Method: "applyPatchApproval",
			Params: mustRaw(map[string]any{"callId": "call", "conversationId": "thr", "fileChanges": map[string]any{}}),
		})
	}
	transport.pushReadLine(request(1))
	select {
	case <-handler.entered:
	case <-time.After(time.Second):
		t.Fatalf("first handler did not start")
	}
	transport.pushReadLine(request(2))
	transport.waitForReads(t, 2)
	transport.pushReadLine(request(3))
	transport.waitForReads(t, 1)

	writes := transport.waitForWrites(t, 1)
	expected := mustJSON(JSONRPCError{
		ID: NewIntRequestID(3),
		Error: JSONRPCErrorDetail{
			Code:    ServerRequestBusyCode,
			Message: "server request queue is full",
		},
	})
	if !equalJSONLine(expected, writes[0]) {
		t.Fatalf("unexpected busy response:\n got: %s\nwant: %s", writes[0], expected)
	}
	close(handler.release)
	transport.waitForWrites(t, 3)
}

func TestServerRequestPanicIsContained(t *testing.T) {
	transport := newChannelTransport()
	client := NewClient(transport, ClientOptions{RequestHandler: &panicServerRequestHandler{}})
	defer client.Close()

	transport.pushReadLine(mustJSON(JSONRPCRequest{
		ID:     NewIntRequestID(9),
		Method: "applyPatchApproval",
		Params: mustRaw(map[string]any{}),
	}))
	writes := transport.waitForWrites(t, 1)
	var response JSONRPCError
	if err := json.Unmarshal([]byte(writes[0]), &response); err != nil {
		t.Fatalf("decode panic response: %v", err)
	}
	if response.Error.Code != ServerRequestInternalErrorCode {
		t.Fatalf("unexpected panic response: %#v", response)
	}
	if err := client.Notify(context.Background(), "still/alive", nil); err != nil {
		t.Fatalf("panic stopped unrelated traffic: %v", err)
	}
}

func TestServerRequestWireErrorContract(t *testing.T) {
	validParams := mustRaw(map[string]any{"callId": "call", "conversationId": "thr", "fileChanges": map[string]any{}})
	tests := []struct {
		name    string
		handler ServerRequestHandler
		method  string
		params  json.RawMessage
		code    int64
		message string
	}{
		{name: "no handler", method: "applyPatchApproval", params: validParams, code: ServerRequestMethodNotFoundCode, message: "no handler configured"},
		{name: "unknown method", handler: &recordingHandler{}, method: "unknown", code: ServerRequestMethodNotFoundCode, message: "method not found"},
		{name: "invalid params", handler: &recordingHandler{}, method: "applyPatchApproval", params: mustRaw([]any{}), code: ServerRequestInvalidParamsCode, message: "invalid params"},
		{name: "handler failure", handler: &testHandler{applyPatch: func(protocol.ApplyPatchApprovalParams) (*protocol.ApplyPatchApprovalResponse, error) {
			return nil, errors.New("secret failure")
		}}, method: "applyPatchApproval", params: validParams, code: ServerRequestHandlerErrorCode, message: "server request handler failed"},
		{name: "unsupported callback", handler: UnimplementedServerRequestHandler{}, method: "applyPatchApproval", params: validParams, code: ServerRequestMethodNotFoundCode, message: "method not found"},
		{name: "panic", handler: &panicServerRequestHandler{}, method: "applyPatchApproval", params: validParams, code: ServerRequestInternalErrorCode, message: "server request handler panicked"},
	}
	for index, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			transport := newChannelTransport()
			client := NewClient(transport, ClientOptions{RequestHandler: test.handler})
			defer client.Close()
			transport.pushReadLine(mustJSON(JSONRPCRequest{ID: NewIntRequestID(int64(index + 1)), Method: test.method, Params: test.params}))
			writes := transport.waitForWrites(t, 1)
			expected := mustJSON(JSONRPCError{
				ID: NewIntRequestID(int64(index + 1)),
				Error: JSONRPCErrorDetail{
					Code:    test.code,
					Message: test.message,
				},
			})
			if !equalJSONLine(expected, writes[0]) {
				t.Fatalf("unexpected response:\n got: %s\nwant: %s", writes[0], expected)
			}
			if err := client.Notify(context.Background(), "still/alive", nil); err != nil {
				t.Fatalf("request error stopped unrelated traffic: %v", err)
			}
		})
	}
}

func TestServerRequestLogsRedactCallbackValues(t *testing.T) {
	const secret = "TOP-SECRET-CALLBACK-VALUE"
	for _, handler := range []ServerRequestHandler{
		&testHandler{applyPatch: func(protocol.ApplyPatchApprovalParams) (*protocol.ApplyPatchApprovalResponse, error) {
			return nil, errors.New(secret)
		}},
		&sensitivePanicHandler{value: secret},
	} {
		var logs bytes.Buffer
		transport := newChannelTransport()
		client := NewClient(transport, ClientOptions{Logger: slog.New(slog.NewTextHandler(&logs, nil)), RequestHandler: handler})
		transport.pushReadLine(mustJSON(JSONRPCRequest{ID: NewIntRequestID(1), Method: "applyPatchApproval", Params: mustRaw(map[string]any{})}))
		transport.waitForWrites(t, 1)
		_ = client.Close()
		if strings.Contains(logs.String(), secret) {
			t.Fatalf("callback value leaked to logs: %s", logs.String())
		}
	}
}

func TestServerRequestReplyFailureIsTerminal(t *testing.T) {
	writeErr := errors.New("reply write failed")
	transport := newFailingWriteTransport(writeErr)
	client := NewClient(transport, ClientOptions{})
	defer client.Close()
	iter := client.SubscribeNotifications(1)
	defer iter.Close()

	transport.reads <- mustJSON(JSONRPCRequest{ID: NewIntRequestID(1), Method: "unknown"})
	ctx, cancel := context.WithTimeout(context.Background(), time.Second)
	defer cancel()
	if _, err := iter.Next(ctx); !errors.Is(err, writeErr) {
		t.Fatalf("expected terminal reply failure, got %v", err)
	}
}

func TestCallInvalidResultJSON(t *testing.T) {
	transcript := []TranscriptEntry{
		writeLine(JSONRPCRequest{
			ID:     NewIntRequestID(1),
			Method: "ping",
			Params: mustRaw(map[string]any{}),
		}),
		readLine(JSONRPCResponse{
			ID:     NewIntRequestID(1),
			Result: mustRaw("not a map"),
		}),
	}

	client := NewClient(NewReplayTransport(transcript), ClientOptions{})
	defer client.Close()

	var result map[string]any
	if err := client.Call(context.Background(), "ping", map[string]any{}, &result); err == nil {
		t.Fatalf("expected invalid result error")
	}
}

func TestCallAfterClose(t *testing.T) {
	client := NewClient(NewReplayTransport(nil), ClientOptions{})
	_ = client.Close()
	var result map[string]any
	if err := client.Call(context.Background(), "ping", map[string]any{}, &result); !errors.Is(err, ErrClientClosed) {
		t.Fatalf("expected ErrClientClosed after close, got %v", err)
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

func TestNotifyInvalidParams(t *testing.T) {
	client := NewClient(newChannelTransport(), ClientOptions{})
	defer client.Close()

	if err := client.Notify(context.Background(), "notice", map[string]any{"bad": func() {}}); err == nil {
		t.Fatalf("expected marshal error")
	}
}

func TestNotifyAfterClose(t *testing.T) {
	client := NewClient(NewReplayTransport(nil), ClientOptions{})
	_ = client.Close()

	if err := client.Notify(context.Background(), "notice", nil); err == nil {
		t.Fatalf("expected error after close")
	}
}

func TestDispatchServerRequestUnknown(t *testing.T) {
	handler := &recordingHandler{}
	req := JSONRPCRequest{ID: NewIntRequestID(1), Method: "unknown"}
	if _, err := dispatchServerRequest(context.Background(), handler, req); err == nil {
		t.Fatalf("expected error for unknown method")
	} else {
		var methodErr *ServerRequestMethodError
		if !errors.As(err, &methodErr) {
			t.Fatalf("expected ServerRequestMethodError, got %T", err)
		}
	}
}

func TestDispatchServerRequestInvalidParams(t *testing.T) {
	handler := &recordingHandler{}
	req := JSONRPCRequest{
		ID:     NewIntRequestID(1),
		Method: "applyPatchApproval",
		Params: json.RawMessage("{bad"),
	}
	if _, err := dispatchServerRequest(context.Background(), handler, req); err == nil {
		t.Fatalf("expected invalid params error")
	} else {
		var paramsErr *ServerRequestParamsError
		if !errors.As(err, &paramsErr) {
			t.Fatalf("expected ServerRequestParamsError, got %T", err)
		}
	}
}

type testHandler struct {
	called     chan struct{}
	applyPatch func(protocol.ApplyPatchApprovalParams) (*protocol.ApplyPatchApprovalResponse, error)
}

type blockingServerRequestHandler struct {
	testHandler
	entered chan struct{}
	done    chan error
	once    sync.Once
}

func (h *blockingServerRequestHandler) ApplyPatchApproval(ctx context.Context, params protocol.ApplyPatchApprovalParams) (*protocol.ApplyPatchApprovalResponse, error) {
	h.once.Do(func() {
		close(h.entered)
	})
	<-ctx.Done()
	err := ctx.Err()
	h.done <- err
	return nil, err
}

func (h *testHandler) ApplyPatchApproval(ctx context.Context, params protocol.ApplyPatchApprovalParams) (*protocol.ApplyPatchApprovalResponse, error) {
	if h.called == nil {
		h.called = make(chan struct{}, 1)
	}
	h.called <- struct{}{}
	if h.applyPatch != nil {
		return h.applyPatch(params)
	}
	resp := protocol.ApplyPatchApprovalResponse{Decision: protocol.MustReviewDecision("approved")}
	return &resp, nil
}

func (h *testHandler) AccountChatgptAuthTokensRefresh(ctx context.Context, params protocol.ChatgptAuthTokensRefreshParams) (*protocol.ChatgptAuthTokensRefreshResponse, error) {
	return nil, errors.New("not implemented")
}

func (h *testHandler) AttestationGenerate(ctx context.Context, params protocol.AttestationGenerateParams) (*protocol.AttestationGenerateResponse, error) {
	return nil, errors.New("not implemented")
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

func (h *testHandler) ItemPermissionsRequestApproval(ctx context.Context, params protocol.PermissionsRequestApprovalParams) (*protocol.PermissionsRequestApprovalResponse, error) {
	return nil, errors.New("not implemented")
}

func (h *testHandler) ItemToolCall(ctx context.Context, params protocol.DynamicToolCallParams) (*protocol.DynamicToolCallResponse, error) {
	return nil, errors.New("not implemented")
}

func (h *testHandler) ItemToolRequestUserInput(ctx context.Context, params protocol.ToolRequestUserInputParams) (*protocol.ToolRequestUserInputResponse, error) {
	return nil, errors.New("not implemented")
}

func (h *testHandler) McpServerElicitationRequest(ctx context.Context, params protocol.MCPServerElicitationRequestParams) (*protocol.MCPServerElicitationRequestResponse, error) {
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

type blockingWriteTransport struct {
	started chan struct{}
	unblock chan struct{}
	once    sync.Once
}

func newBlockingWriteTransport() *blockingWriteTransport {
	return &blockingWriteTransport{started: make(chan struct{}), unblock: make(chan struct{})}
}

func (t *blockingWriteTransport) ReadLine() (string, error) {
	<-t.unblock
	return "", io.EOF
}

func (t *blockingWriteTransport) WriteLine(string) error {
	t.once.Do(func() { close(t.started) })
	<-t.unblock
	return io.ErrClosedPipe
}

func (t *blockingWriteTransport) Close() error {
	select {
	case <-t.unblock:
	default:
		close(t.unblock)
	}
	return nil
}

type serialWriteTransport struct {
	closed        chan struct{}
	closeOnce     sync.Once
	active        atomic.Int32
	maxConcurrent atomic.Int32
}

type contextWriteTransport struct {
	closed       chan struct{}
	closeOnce    sync.Once
	contextWrite chan struct{}
	writeOnce    sync.Once
}

func newContextWriteTransport() *contextWriteTransport {
	return &contextWriteTransport{closed: make(chan struct{}), contextWrite: make(chan struct{})}
}

func (t *contextWriteTransport) ReadLine() (string, error) {
	<-t.closed
	return "", io.EOF
}

func (*contextWriteTransport) WriteLine(string) error {
	panic("legacy WriteLine called on ContextTransport")
}

func (t *contextWriteTransport) WriteLineContext(ctx context.Context, _ string) error {
	t.writeOnce.Do(func() { close(t.contextWrite) })
	<-ctx.Done()
	return ctx.Err()
}

func (t *contextWriteTransport) Close() error {
	t.closeOnce.Do(func() { close(t.closed) })
	return nil
}

func newSerialWriteTransport() *serialWriteTransport {
	return &serialWriteTransport{closed: make(chan struct{})}
}

func (t *serialWriteTransport) ReadLine() (string, error) {
	<-t.closed
	return "", io.EOF
}

func (t *serialWriteTransport) WriteLine(string) error {
	active := t.active.Add(1)
	defer t.active.Add(-1)
	for {
		maximum := t.maxConcurrent.Load()
		if active <= maximum || t.maxConcurrent.CompareAndSwap(maximum, active) {
			break
		}
	}
	time.Sleep(time.Millisecond)
	return nil
}

func (t *serialWriteTransport) Close() error {
	t.closeOnce.Do(func() { close(t.closed) })
	return nil
}

type queueBlockingHandler struct {
	testHandler
	entered chan struct{}
	release chan struct{}
}

func (h *queueBlockingHandler) ApplyPatchApproval(context.Context, protocol.ApplyPatchApprovalParams) (*protocol.ApplyPatchApprovalResponse, error) {
	select {
	case h.entered <- struct{}{}:
	default:
	}
	<-h.release
	return &protocol.ApplyPatchApprovalResponse{Decision: protocol.MustReviewDecision("approved")}, nil
}

type panicServerRequestHandler struct{ testHandler }

func (*panicServerRequestHandler) ApplyPatchApproval(context.Context, protocol.ApplyPatchApprovalParams) (*protocol.ApplyPatchApprovalResponse, error) {
	panic("handler panic")
}

type sensitivePanicHandler struct {
	testHandler
	value string
}

func (h *sensitivePanicHandler) ApplyPatchApproval(context.Context, protocol.ApplyPatchApprovalParams) (*protocol.ApplyPatchApprovalResponse, error) {
	panic(h.value)
}

type failingWriteTransport struct {
	reads     chan string
	closed    chan struct{}
	closeOnce sync.Once
	err       error
}

func newFailingWriteTransport(err error) *failingWriteTransport {
	return &failingWriteTransport{reads: make(chan string, 1), closed: make(chan struct{}), err: err}
}

func (t *failingWriteTransport) ReadLine() (string, error) {
	select {
	case <-t.closed:
		return "", io.EOF
	case line := <-t.reads:
		return line, nil
	}
}
func (t *failingWriteTransport) WriteLine(string) error { return t.err }
func (t *failingWriteTransport) Close() error {
	t.closeOnce.Do(func() { close(t.closed) })
	return nil
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

func (t *channelTransport) waitForWrites(testingT *testing.T, count int) []string {
	testingT.Helper()
	deadline := time.After(time.Second)
	ticker := time.NewTicker(time.Millisecond)
	defer ticker.Stop()
	for {
		t.mu.Lock()
		if len(t.writes) >= count {
			writes := append([]string(nil), t.writes...)
			t.mu.Unlock()
			return writes
		}
		t.mu.Unlock()

		select {
		case <-deadline:
			testingT.Fatalf("timed out waiting for %d writes", count)
		case <-ticker.C:
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
