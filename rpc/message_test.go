package rpc

import (
	"context"
	"encoding/json"
	"errors"
	"testing"
	"time"
)

func TestRequestIDJSON(t *testing.T) {
	empty := RequestID{}
	if !empty.IsZero() {
		t.Fatalf("expected zero request id")
	}
	if empty.Key() != "" || empty.String() != "" {
		t.Fatalf("expected empty key/string")
	}

	stringID := NewStringRequestID("abc")
	if stringID.IsZero() {
		t.Fatalf("expected non-zero string id")
	}
	if stringID.Key() != "s:abc" {
		t.Fatalf("unexpected key: %s", stringID.Key())
	}
	if stringID.String() != "abc" {
		t.Fatalf("unexpected string: %s", stringID.String())
	}
	if data, err := json.Marshal(stringID); err != nil || string(data) != `"abc"` {
		t.Fatalf("unexpected string marshal: %s err=%v", data, err)
	}

	id := NewIntRequestID(42)
	data, err := json.Marshal(id)
	if err != nil {
		t.Fatalf("marshal error: %v", err)
	}
	if string(data) != "42" {
		t.Fatalf("unexpected int json: %s", data)
	}

	var parsed RequestID
	if err := json.Unmarshal([]byte(`"abc"`), &parsed); err != nil {
		t.Fatalf("unmarshal error: %v", err)
	}
	if parsed.String() != "abc" {
		t.Fatalf("unexpected string id: %s", parsed.String())
	}

	if err := json.Unmarshal([]byte(`{}`), &parsed); err == nil {
		t.Fatalf("expected invalid request id error")
	}
}

func TestParseMessageVariants(t *testing.T) {
	msg, err := parseMessage([]byte(`{"id":1,"method":"ping","params":{"ok":true}}`))
	if err != nil || msg.kind != messageRequest {
		t.Fatalf("expected request message, got %#v err=%v", msg, err)
	}

	msg, err = parseMessage([]byte(`{"method":"notify","params":{"ok":true}}`))
	if err != nil || msg.kind != messageNotification {
		t.Fatalf("expected notification message, got %#v err=%v", msg, err)
	}

	msg, err = parseMessage([]byte(`{"id":2,"result":{"ok":true}}`))
	if err != nil || msg.kind != messageResponse {
		t.Fatalf("expected response message, got %#v err=%v", msg, err)
	}

	msg, err = parseMessage([]byte(`{"id":3,"error":{"code":-1,"message":"bad"}}`))
	if err != nil || msg.kind != messageError {
		t.Fatalf("expected error message, got %#v err=%v", msg, err)
	}

	if _, err := parseMessage([]byte(`{"jsonrpc":"2.0"}`)); err == nil {
		t.Fatalf("expected unrecognized message error")
	}
	if _, err := parseMessage([]byte(`{"id":{},"method":"ping"}`)); err == nil {
		t.Fatalf("expected invalid request id error")
	}
	if _, err := parseMessage([]byte(`{"id":{},"result":{}}`)); err == nil {
		t.Fatalf("expected invalid response id error")
	}
	if _, err := parseMessage([]byte(`{"id":{},"error":{"code":-1,"message":"bad"}}`)); err == nil {
		t.Fatalf("expected invalid error id error")
	}
}

func TestNotificationUnmarshalParams(t *testing.T) {
	var payload map[string]bool
	note := Notification{Raw: json.RawMessage(`{"ok":true}`)}
	if err := note.UnmarshalParams(&payload); err != nil {
		t.Fatalf("unmarshal params: %v", err)
	}
	if !payload["ok"] {
		t.Fatalf("unexpected payload: %#v", payload)
	}

	payload = map[string]bool{"kept": true}
	note = Notification{}
	if err := note.UnmarshalParams(&payload); err != nil {
		t.Fatalf("empty raw params: %v", err)
	}
	if !payload["kept"] {
		t.Fatalf("expected empty raw params to leave target alone")
	}

	note = Notification{Raw: json.RawMessage("{bad")}
	if err := note.UnmarshalParams(&payload); err == nil {
		t.Fatalf("expected invalid raw params error")
	}
}

func TestReplayJSONLineHelpers(t *testing.T) {
	if !equalJSONLine(`{"a":1,"b":2}`, `{"b":2,"a":1}`) {
		t.Fatalf("expected equal json lines")
	}
	if equalJSONLine(``, `{}`) {
		t.Fatalf("expected empty expected line to fail normalization")
	}
	if equalJSONLine(`{bad}`, `{}`) {
		t.Fatalf("expected invalid expected json to fail normalization")
	}
	if equalJSONLine(`{}`, `{bad}`) {
		t.Fatalf("expected invalid actual json to fail normalization")
	}

	if got, ok := normalizeJSONLine(" \n "); ok || got != "" {
		t.Fatalf("expected blank line normalization to fail")
	}
}

func TestNotificationIteratorNext(t *testing.T) {
	done := make(chan struct{})
	errFn := func() error { return errors.New("closed") }
	ch := make(chan Notification, 1)
	iter := NotificationIterator{ch: ch, done: done, err: errFn}

	ch <- Notification{Method: "note"}
	note, err := iter.Next(context.Background())
	if err != nil || note.Method != "note" {
		t.Fatalf("unexpected note: %#v err=%v", note, err)
	}

	ctx, cancel := context.WithCancel(context.Background())
	cancel()
	if _, err := iter.Next(ctx); err == nil {
		t.Fatalf("expected context error")
	}

	close(done)
	if _, err := iter.Next(context.Background()); err == nil {
		t.Fatalf("expected done error")
	}
}

func TestReplayTransportWaitsForRead(t *testing.T) {
	replay := NewReplayTransport([]TranscriptEntry{
		{Direction: TranscriptWrite, Line: "write"},
		{Direction: TranscriptRead, Line: "read"},
	})

	done := make(chan struct{})
	go func() {
		defer close(done)
		if err := replay.WriteLine("write"); err != nil {
			t.Errorf("write error: %v", err)
		}
	}()

	select {
	case <-done:
	case <-time.After(1 * time.Second):
		t.Fatalf("write did not complete")
	}

	line, err := replay.ReadLine()
	if err != nil {
		t.Fatalf("read error: %v", err)
	}
	if line != "read" {
		t.Fatalf("unexpected line: %s", line)
	}
}
