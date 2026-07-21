package codex

import (
	"context"
	"encoding/json"
	"errors"
	"io"
	"sync"
	"testing"
	"time"

	"github.com/pmenglund/codex-sdk-go/protocol"
	"github.com/pmenglund/codex-sdk-go/rpc"
)

func TestTurnHandleRunSteerInterruptWithReplay(t *testing.T) {
	ctx := context.Background()
	info := protocol.ClientInfo{Name: "codex-go-test", Version: "test"}
	client, err := New(ctx, Options{
		Transport: rpc.NewReplayTransport([]rpc.TranscriptEntry{
			writeLine(rpc.JSONRPCRequest{ID: rpc.NewIntRequestID(1), Method: "initialize", Params: mustRaw(protocol.InitializeParams{ClientInfo: info})}),
			readLine(rpc.JSONRPCResponse{ID: rpc.NewIntRequestID(1), Result: mustRaw(map[string]any{})}),
			writeLine(rpc.JSONRPCNotification{Method: "initialized"}),
			writeLine(rpc.JSONRPCRequest{ID: rpc.NewIntRequestID(2), Method: "thread/start", Params: mustRaw(map[string]any{})}),
			readLine(rpc.JSONRPCResponse{ID: rpc.NewIntRequestID(2), Result: mustRaw(map[string]any{"thread": map[string]any{"id": "thr_123"}})}),
			writeLine(rpc.JSONRPCRequest{ID: rpc.NewIntRequestID(3), Method: "turn/start", Params: mustRaw(turnStartParams("hello"))}),
			readLine(rpc.JSONRPCResponse{ID: rpc.NewIntRequestID(3), Result: mustRaw(map[string]any{"turn": turnPayload("turn_1", "inProgress")})}),
			writeLine(rpc.JSONRPCRequest{ID: rpc.NewIntRequestID(4), Method: "turn/steer", Params: mustRaw(map[string]any{"threadId": "thr_123", "expectedTurnId": "turn_1", "input": []Input{TextInput("more")}})}),
			readLine(rpc.JSONRPCResponse{ID: rpc.NewIntRequestID(4), Result: mustRaw(map[string]any{})}),
			writeLine(rpc.JSONRPCRequest{ID: rpc.NewIntRequestID(5), Method: "turn/interrupt", Params: mustRaw(map[string]any{"threadId": "thr_123", "turnId": "turn_1"})}),
			readLine(rpc.JSONRPCResponse{ID: rpc.NewIntRequestID(5), Result: mustRaw(map[string]any{})}),
			readLine(rpc.JSONRPCNotification{Method: "turn/started", Params: mustRaw(map[string]any{"threadId": "thr_123", "turn": turnPayload("turn_1", "inProgress")})}),
			readLine(rpc.JSONRPCNotification{Method: "thread/tokenUsage/updated", Params: mustRaw(map[string]any{
				"threadId": "thr_123",
				"turnId":   "turn_1",
				"tokenUsage": map[string]any{
					"last":  tokenUsageBreakdown(1, 2, 3),
					"total": tokenUsageBreakdown(4, 5, 9),
				},
			})}),
			readLine(rpc.JSONRPCNotification{Method: "item/completed", Params: mustRaw(map[string]any{"threadId": "thr_123", "item": map[string]any{"text": "final"}})}),
			readLine(rpc.JSONRPCNotification{Method: "turn/completed", Params: mustRaw(map[string]any{"threadId": "thr_123", "turn": turnPayload("turn_1", "completed")})}),
		}),
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
	handle, err := thread.StartTurn(ctx, []Input{TextInput("hello")}, nil)
	if err != nil {
		t.Fatalf("start turn error: %v", err)
	}
	if _, err := handle.Steer(ctx, []Input{TextInput("more")}); err != nil {
		t.Fatalf("steer error: %v", err)
	}
	if _, err := handle.Interrupt(ctx); err != nil {
		t.Fatalf("interrupt error: %v", err)
	}
	result, err := handle.Run(ctx)
	if err != nil {
		t.Fatalf("run error: %v", err)
	}
	if result.TurnID != "turn_1" || result.Status != "completed" || result.FinalResponse != "final" {
		t.Fatalf("unexpected result: %#v", result)
	}
	if result.TokenUsage == nil || result.TokenUsage.Total.TotalTokens != 9 {
		t.Fatalf("expected token usage, got %#v", result.TokenUsage)
	}
}

func TestTurnSteerParamsInputElemCompatibility(t *testing.T) {
	input, err := protocol.NewUserInput(TextInput("compatible"))
	if err != nil {
		t.Fatalf("new user input: %v", err)
	}
	var elem protocol.TurnSteerParamsInputElem = input
	params := protocol.TurnSteerParams{
		ThreadID:       "thr_123",
		ExpectedTurnID: "turn_123",
		Input:          []protocol.UserInput{elem},
	}
	if len(params.Input) != 1 {
		t.Fatalf("expected compatibility input")
	}
}

func TestTurnHandleRejectsUnknownTurnIDAndInvalidInputs(t *testing.T) {
	handle := &TurnHandle{
		client:   rpc.NewClient(rpc.NewReplayTransport(nil), rpc.ClientOptions{}),
		threadID: "thr_123",
		stream:   &TurnStream{},
	}
	defer handle.client.Close()

	if _, err := handle.Steer(context.Background(), []Input{TextInput("hi")}); err == nil {
		t.Fatalf("expected missing turn id error")
	}
	handle.setTurnID("turn_1")
	if _, err := handle.Steer(context.Background(), []Input{ImageInput("")}); err == nil {
		t.Fatalf("expected invalid input error")
	}
	handle.Close()
	handle.Close()
	if _, err := handle.Next(context.Background()); err == nil {
		t.Fatalf("expected closed handle error")
	}
}

func TestTurnHandleContextCancellation(t *testing.T) {
	client := rpc.NewClient(rpc.NewReplayTransport(nil), rpc.ClientOptions{})
	defer client.Close()
	handle := &TurnHandle{
		client:   client,
		threadID: "thr_123",
		stream:   &TurnStream{iter: client.SubscribeNotifications(0), threadID: "thr_123"},
	}
	ctx, cancel := context.WithTimeout(context.Background(), time.Nanosecond)
	defer cancel()
	_, err := handle.Next(ctx)
	if !errors.Is(err, context.DeadlineExceeded) && !errors.Is(err, context.Canceled) {
		t.Fatalf("expected context error, got %v", err)
	}
}

func TestTurnHandleRunCancellationInterruptsAndPreservesContextError(t *testing.T) {
	info := protocol.ClientInfo{Name: "codex-go-test", Version: "test"}
	client, err := New(context.Background(), Options{
		Transport: rpc.NewReplayTransport([]rpc.TranscriptEntry{
			writeLine(rpc.JSONRPCRequest{ID: rpc.NewIntRequestID(1), Method: "initialize", Params: mustRaw(protocol.InitializeParams{ClientInfo: info})}),
			readLine(rpc.JSONRPCResponse{ID: rpc.NewIntRequestID(1), Result: mustRaw(map[string]any{})}),
			writeLine(rpc.JSONRPCNotification{Method: "initialized"}),
			writeLine(rpc.JSONRPCRequest{ID: rpc.NewIntRequestID(2), Method: "thread/start", Params: mustRaw(map[string]any{})}),
			readLine(rpc.JSONRPCResponse{ID: rpc.NewIntRequestID(2), Result: mustRaw(map[string]any{"thread": map[string]any{"id": "thr_123"}})}),
			writeLine(rpc.JSONRPCRequest{ID: rpc.NewIntRequestID(3), Method: "turn/start", Params: mustRaw(turnStartParams("hello"))}),
			readLine(rpc.JSONRPCResponse{ID: rpc.NewIntRequestID(3), Result: mustRaw(map[string]any{"turn": turnPayload("turn_1", "inProgress")})}),
			writeLine(rpc.JSONRPCRequest{ID: rpc.NewIntRequestID(4), Method: "turn/interrupt", Params: mustRaw(map[string]any{"threadId": "thr_123", "turnId": "turn_1"})}),
			readLine(rpc.JSONRPCError{ID: rpc.NewIntRequestID(4), Error: rpc.JSONRPCErrorError{Code: -32000, Message: "interrupt failed"}}),
		}),
		ClientInfo: info,
	})
	if err != nil {
		t.Fatalf("new client error: %v", err)
	}
	defer client.Close()
	thread, err := client.StartThread(context.Background(), ThreadStartOptions{})
	if err != nil {
		t.Fatalf("start thread error: %v", err)
	}

	handle, err := thread.StartTurn(context.Background(), []Input{TextInput("hello")}, nil)
	if err != nil {
		t.Fatalf("start turn error: %v", err)
	}
	runCtx, cancel := context.WithCancel(context.Background())
	cancel()
	_, err = handle.Run(runCtx)
	if err != context.Canceled {
		t.Fatalf("expected exact context.Canceled, got %v", err)
	}
}

func TestTurnHandleRunCancellationWithoutTurnIDPreservesContextError(t *testing.T) {
	client := rpc.NewClient(rpc.NewReplayTransport(nil), rpc.ClientOptions{})
	defer client.Close()
	handle := &TurnHandle{
		client:   client,
		threadID: "thr_123",
		stream:   &TurnStream{iter: client.SubscribeNotifications(0), threadID: "thr_123"},
	}
	ctx, cancel := context.WithCancel(context.Background())
	cancel()
	_, err := handle.Run(ctx)
	if err != context.Canceled {
		t.Fatalf("expected exact context.Canceled, got %v", err)
	}
}

func TestTurnHandleCancellationCleanupIsBoundedWhenTransportWriteBlocks(t *testing.T) {
	originalTimeout := turnInterruptCleanupTimeout
	turnInterruptCleanupTimeout = 25 * time.Millisecond
	t.Cleanup(func() { turnInterruptCleanupTimeout = originalTimeout })

	transport := newBlockingWriteTransport()
	client := rpc.NewClient(transport, rpc.ClientOptions{})
	handle := &TurnHandle{
		client:   client,
		threadID: "thr_123",
		turnID:   "turn_1",
		stream:   &TurnStream{iter: client.SubscribeNotifications(1), threadID: "thr_123"},
	}
	ctx, cancel := context.WithCancel(context.Background())
	cancel()
	started := time.Now()
	_, err := handle.Run(ctx)
	if err != context.Canceled {
		t.Fatalf("expected exact context.Canceled, got %v", err)
	}
	if elapsed := time.Since(started); elapsed > 250*time.Millisecond {
		t.Fatalf("cancellation cleanup was not bounded: %v", elapsed)
	}
	select {
	case <-transport.writeStarted:
	case <-time.After(time.Second):
		t.Fatalf("interrupt write was not attempted")
	}
	_ = client.Close()
}

func TestTurnHandleNotificationOverflowInterruptsRemoteTurn(t *testing.T) {
	transport := newOverflowTurnTransport()
	client := rpc.NewClient(transport, rpc.ClientOptions{})
	defer client.Close()
	handle := &TurnHandle{
		client:   client,
		threadID: "thr_123",
		turnID:   "turn_1",
		stream:   &TurnStream{iter: client.SubscribeNotifications(1), threadID: "thr_123"},
	}
	note := mustJSON(rpc.JSONRPCNotification{Method: "turn/started", Params: mustRaw(map[string]any{
		"threadId": "thr_123",
		"turn":     turnPayload("turn_1", "inProgress"),
	})})
	for index := 0; index < 3; index++ {
		transport.reads <- note
	}

	ctx, cancel := context.WithTimeout(context.Background(), time.Second)
	defer cancel()
	_, err := handle.Run(ctx)
	if !errors.Is(err, rpc.ErrNotificationOverflow) {
		t.Fatalf("expected notification overflow, got %v", err)
	}
	select {
	case line := <-transport.writes:
		var request rpc.JSONRPCRequest
		if err := json.Unmarshal([]byte(line), &request); err != nil {
			t.Fatalf("decode interrupt request: %v", err)
		}
		if request.Method != "turn/interrupt" {
			t.Fatalf("expected turn/interrupt, got %s", request.Method)
		}
	case <-time.After(time.Second):
		t.Fatalf("overflow did not trigger turn interruption")
	}
}

type blockingWriteTransport struct {
	writeStarted chan struct{}
	release      chan struct{}
	startOnce    sync.Once
	closeOnce    sync.Once
}

type overflowTurnTransport struct {
	reads     chan string
	writes    chan string
	closed    chan struct{}
	closeOnce sync.Once
}

func newOverflowTurnTransport() *overflowTurnTransport {
	return &overflowTurnTransport{reads: make(chan string), writes: make(chan string, 4), closed: make(chan struct{})}
}

func (t *overflowTurnTransport) ReadLine() (string, error) {
	select {
	case line := <-t.reads:
		return line, nil
	case <-t.closed:
		return "", io.EOF
	}
}

func (t *overflowTurnTransport) WriteLine(line string) error {
	t.writes <- line
	var request rpc.JSONRPCRequest
	if json.Unmarshal([]byte(line), &request) == nil && request.Method == "turn/interrupt" {
		go func() {
			t.reads <- mustJSON(rpc.JSONRPCResponse{ID: request.ID, Result: mustRaw(map[string]any{})})
		}()
	}
	return nil
}

func (t *overflowTurnTransport) Close() error {
	t.closeOnce.Do(func() { close(t.closed) })
	return nil
}

func newBlockingWriteTransport() *blockingWriteTransport {
	return &blockingWriteTransport{writeStarted: make(chan struct{}), release: make(chan struct{})}
}

func (t *blockingWriteTransport) ReadLine() (string, error) {
	<-t.release
	return "", io.EOF
}

func (t *blockingWriteTransport) WriteLine(string) error {
	t.startOnce.Do(func() { close(t.writeStarted) })
	<-t.release
	return errors.New("transport closed")
}

func (t *blockingWriteTransport) Close() error {
	t.closeOnce.Do(func() { close(t.release) })
	return nil
}

func TestRunStreamedStillFiltersThreadAndGlobalNotifications(t *testing.T) {
	ctx := context.Background()
	info := protocol.ClientInfo{Name: "codex-go-test", Version: "test"}
	client, err := New(ctx, Options{
		Transport: rpc.NewReplayTransport([]rpc.TranscriptEntry{
			writeLine(rpc.JSONRPCRequest{ID: rpc.NewIntRequestID(1), Method: "initialize", Params: mustRaw(protocol.InitializeParams{ClientInfo: info})}),
			readLine(rpc.JSONRPCResponse{ID: rpc.NewIntRequestID(1), Result: mustRaw(map[string]any{})}),
			writeLine(rpc.JSONRPCNotification{Method: "initialized"}),
			writeLine(rpc.JSONRPCRequest{ID: rpc.NewIntRequestID(2), Method: "thread/start", Params: mustRaw(map[string]any{})}),
			readLine(rpc.JSONRPCResponse{ID: rpc.NewIntRequestID(2), Result: mustRaw(map[string]any{"thread": map[string]any{"id": "thr_123"}})}),
			writeLine(rpc.JSONRPCRequest{ID: rpc.NewIntRequestID(3), Method: "turn/start", Params: mustRaw(turnStartParams("hello"))}),
			readLine(rpc.JSONRPCResponse{ID: rpc.NewIntRequestID(3), Result: mustRaw(map[string]any{"turn": turnPayload("turn_1", "inProgress")})}),
			readLine(rpc.JSONRPCNotification{Method: "turn/started", Params: mustRaw(map[string]any{"threadId": "other", "turn": turnPayload("turn_other", "inProgress")})}),
			readLine(rpc.JSONRPCNotification{Method: "account/updated", Params: mustRaw(map[string]any{})}),
			readLine(rpc.JSONRPCNotification{Method: "turn/started", Params: mustRaw(map[string]any{"threadId": "thr_123", "turn": turnPayload("turn_1", "inProgress")})}),
		}),
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
	stream, err := thread.RunStreamed(ctx, []Input{TextInput("hello")}, nil)
	if err != nil {
		t.Fatalf("run streamed error: %v", err)
	}
	defer stream.Close()
	note, err := stream.Next(ctx)
	if err != nil {
		t.Fatalf("next global error: %v", err)
	}
	if note.Method != "account/updated" {
		t.Fatalf("expected global notification, got %s", note.Method)
	}
	note, err = stream.Next(ctx)
	if err != nil {
		t.Fatalf("next thread error: %v", err)
	}
	if note.Method != "turn/started" {
		t.Fatalf("expected turn started, got %s", note.Method)
	}
}

func tokenUsageBreakdown(input, output, total int) map[string]any {
	return map[string]any{
		"inputTokens":           input,
		"cachedInputTokens":     0,
		"outputTokens":          output,
		"reasoningOutputTokens": 0,
		"totalTokens":           total,
	}
}
