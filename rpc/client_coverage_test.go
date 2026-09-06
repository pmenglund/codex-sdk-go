package rpc

import (
	"context"
	"errors"
	"strings"
	"sync"
	"testing"
	"testing/synctest"
)

// gatedRecordingTransport holds writes until released, then records successful writes.
type gatedRecordingTransport struct {
	*channelTransport
	release     chan struct{}
	releaseOnce sync.Once
}

func newGatedRecordingTransport() *gatedRecordingTransport {
	return &gatedRecordingTransport{channelTransport: newChannelTransport(), release: make(chan struct{})}
}
func (tr *gatedRecordingTransport) unblock() { tr.releaseOnce.Do(func() { close(tr.release) }) }
func (tr *gatedRecordingTransport) WriteLine(line string) error {
	<-tr.release
	return tr.channelTransport.WriteLine(line)
}
func (tr *gatedRecordingTransport) Close() error {
	tr.unblock()
	return tr.channelTransport.Close()
}

func requireCallResult(t *testing.T, done <-chan error, want error) {
	t.Helper()
	synctest.Wait()
	select {
	case err := <-done:
		if !errors.Is(err, want) {
			t.Fatalf("call error = %v, want %v", err, want)
		}
	default:
		t.Fatal("call is still blocked")
	}
}

func requireNoPendingCalls(t *testing.T, client *Client) {
	t.Helper()
	client.pendingMu.Lock()
	defer client.pendingMu.Unlock()
	if len(client.pending) != 0 {
		t.Fatalf("retained %d pending calls", len(client.pending))
	}
}

func TestQueuedCallCancellationDoesNotWriteOrPoisonClient(t *testing.T) {
	for _, full := range []bool{false, true} {
		name := "already_enqueued"
		if full {
			name = "waiting_for_queue_space"
		}
		t.Run(name, func(t *testing.T) {
			synctest.Test(t, func(t *testing.T) {
				tr := newGatedRecordingTransport()
				client := NewClient(tr, ClientOptions{})
				defer client.Close()
				first := make(chan error, 1)
				go func() { first <- client.Notify(context.Background(), "first", nil) }()
				synctest.Wait()
				if len(client.writes) != 0 {
					t.Fatal("first write was not dequeued")
				}
				fillers := 0
				if full {
					fillers = cap(client.writes)
					for i := 0; i < fillers; i++ {
						if err := client.sendAsync(context.Background(), JSONRPCNotification{Method: "filler"}); err != nil {
							t.Fatal(err)
						}
					}
				}
				ctx, cancel := context.WithCancel(context.Background())
				defer cancel()
				done := make(chan error, 1)
				go func() { done <- client.Call(ctx, "must/not/write", nil, nil) }()
				synctest.Wait()
				wantQueued := fillers
				if !full {
					wantQueued = 1
				}
				if len(client.writes) != wantQueued {
					t.Fatalf("queued writes = %d, want %d", len(client.writes), wantQueued)
				}
				select {
				case err := <-done:
					t.Fatalf("call returned before cancellation: %v", err)
				default:
				}
				cancel()
				requireCallResult(t, done, context.Canceled)
				requireNoPendingCalls(t, client)
				tr.unblock()
				requireCallResult(t, first, nil)
				if err := client.Notify(context.Background(), "after/cancel", nil); err != nil {
					t.Fatal(err)
				}
				writes := tr.waitForWrites(t, fillers+2)
				if len(writes) != fillers+2 {
					t.Fatalf("unexpected writes: %v", writes)
				}
				for _, line := range writes {
					if strings.Contains(line, "must/not/write") {
						t.Fatalf("canceled call was transmitted: %s", line)
					}
				}
			})
		})
	}
}

func TestOutboundQueueFull(t *testing.T) {
	for _, busyReply := range []bool{false, true} {
		name := "async_rejection_can_recover"
		if busyReply {
			name = "busy_reply_failure_is_terminal"
		}
		t.Run(name, func(t *testing.T) {
			synctest.Test(t, func(t *testing.T) {
				tr := newGatedRecordingTransport()
				handler := &queueBlockingHandler{entered: make(chan struct{}, 1), release: make(chan struct{})}
				client := NewClient(tr, ClientOptions{RequestHandler: handler, ServerRequestWorkers: 1, ServerRequestQueueCapacity: 1})
				defer client.Close()
				defer close(handler.release)
				done := make(chan error, 1)
				go func() { done <- client.Notify(context.Background(), "blocked", nil) }()
				synctest.Wait()
				for i := 0; i < cap(client.writes); i++ {
					if err := client.sendAsync(context.Background(), JSONRPCNotification{Method: "filler"}); err != nil {
						t.Fatal(err)
					}
				}
				if !busyReply {
					if err := client.sendAsync(context.Background(), JSONRPCNotification{Method: "overflow"}); !errors.Is(err, ErrOutboundQueueFull) {
						t.Fatalf("queue error = %v", err)
					}
					tr.unblock()
					requireCallResult(t, done, nil)
					if err := client.Notify(context.Background(), "after/full", nil); err != nil {
						t.Fatal(err)
					}
					writes := tr.waitForWrites(t, cap(client.writes)+2)
					if len(writes) != cap(client.writes)+2 {
						t.Fatalf("unexpected write count: %d", len(writes))
					}
					for _, line := range writes {
						if strings.Contains(line, "overflow") {
							t.Fatalf("rejected reply was transmitted: %s", line)
						}
					}
					return
				}
				request := func(id int64) string {
					return mustJSON(JSONRPCRequest{ID: NewIntRequestID(id), Method: "applyPatchApproval", Params: mustRaw(map[string]any{"callId": "call", "conversationId": "thread", "fileChanges": map[string]any{}})})
				}
				tr.pushReadLine(request(1))
				synctest.Wait()
				select {
				case <-handler.entered:
				default:
					t.Fatal("worker did not start")
				}
				tr.pushReadLine(request(2))
				synctest.Wait()
				if len(client.requests) != 1 {
					t.Fatal("request queue was not filled")
				}
				tr.pushReadLine(request(3))
				requireCallResult(t, done, ErrOutboundQueueFull)
				if err := client.Notify(context.Background(), "after/failure", nil); !errors.Is(err, ErrOutboundQueueFull) {
					t.Fatalf("terminal error = %v", err)
				}
			})
		})
	}
}

func TestReaderIgnoresMalformedAndLateResponses(t *testing.T) {
	synctest.Test(t, func(t *testing.T) {
		tr := newChannelTransport()
		client := NewClient(tr, ClientOptions{})
		defer client.Close()
		ctx, cancel := context.WithCancel(context.Background())
		defer cancel()
		canceled := make(chan error, 1)
		go func() { canceled <- client.Call(ctx, "canceled", nil, nil) }()
		tr.waitForWrites(t, 1)
		cancel()
		requireCallResult(t, canceled, context.Canceled)
		requireNoPendingCalls(t, client)
		var result struct {
			OK bool `json:"ok"`
		}
		done := make(chan error, 1)
		go func() { done <- client.Call(context.Background(), "healthy", nil, &result) }()
		tr.waitForWrites(t, 2)
		for _, line := range []string{
			"  ", "{not json",
			mustJSON(JSONRPCResponse{ID: NewIntRequestID(1), Result: mustRaw(map[string]any{"ok": false})}),
			mustJSON(JSONRPCError{ID: NewIntRequestID(1), Error: JSONRPCErrorDetail{Code: -32000, Message: "late"}}),
			mustJSON(JSONRPCResponse{ID: NewIntRequestID(999), Result: mustRaw(map[string]any{})}),
			mustJSON(JSONRPCError{ID: NewIntRequestID(999), Error: JSONRPCErrorDetail{Code: -32000, Message: "unknown"}}),
		} {
			tr.pushReadLine(line)
		}
		tr.waitForReads(t, 6)
		select {
		case err := <-done:
			t.Fatalf("unrelated message completed the healthy call: %v", err)
		default:
		}
		tr.pushReadLine(mustJSON(JSONRPCResponse{ID: NewIntRequestID(2), Result: mustRaw(map[string]any{"ok": true})}))
		requireCallResult(t, done, nil)
		if !result.OK {
			t.Fatal("valid response was not delivered")
		}
		requireNoPendingCalls(t, client)
	})
}
