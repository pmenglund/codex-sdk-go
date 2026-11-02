package rpc

import (
	"context"
	"encoding/json"
	"testing"

	"github.com/pmenglund/codex-sdk-go/protocol"
)

func TestProtocol0166ThreadHooks(t *testing.T) {
	transport := newScriptedTransport()
	client := NewClient(transport, ClientOptions{})
	t.Cleanup(func() { _ = client.Close() })
	transport.enqueueResult(json.RawMessage(`{"data":[]}`))
	var params protocol.HooksListParams
	if err := json.Unmarshal([]byte(`{"threadId":"thread"}`), &params); err != nil {
		t.Fatal(err)
	}
	if _, err := client.HooksList(context.Background(), params); err != nil {
		t.Fatal(err)
	}
	requests := transport.writtenRequests()
	if len(requests) != 1 || requests[0].Method != "hooks/list" || string(requests[0].Params) != `{"threadId":"thread"}` {
		t.Fatalf("hook requests: %#v", requests)
	}
}

func TestProtocol0166LifecycleNotifications(t *testing.T) {
	item := `{"type":"imageGeneration","id":"image","status":"completed","fileId":"file","generationId":"generation","startedAtMs":0,"completedAtMs":9007199254740993}`
	for _, method := range []string{"item/started", "item/completed"} {
		payload := `{"threadId":"thread","turnId":"turn","item":` + item + `}`
		note, err := parseServerNotification(method, json.RawMessage(payload))
		if err != nil {
			t.Fatal(err)
		}
		if string(note.Raw) != payload {
			t.Fatalf("lifecycle notification raw data changed: %s", note.Raw)
		}
		if completed, ok := note.Params.(protocol.ItemCompletedNotification); ok {
			encoded, err := json.Marshal(completed.Item)
			if err != nil || string(encoded) != item || !completed.Item.IsKnown() {
				t.Fatalf("completed item changed: %s, %v", encoded, err)
			}
		} else if method == "item/completed" {
			t.Fatalf("unexpected completed notification type: %T", note.Params)
		}
	}
}

func TestProtocol0166InlineReview(t *testing.T) {
	transport := newScriptedTransport()
	client := NewClient(transport, ClientOptions{})
	t.Cleanup(func() { _ = client.Close() })
	transport.enqueueResult(json.RawMessage(`{"turn":{"id":"turn","items":[],"status":"inProgress"},"reviewThreadId":"thread"}`))
	target, err := protocol.NewReviewTarget(json.RawMessage(`{"type":"uncommittedChanges"}`))
	if err != nil {
		t.Fatal(err)
	}
	delivery := protocol.ReviewDeliveryInline
	if _, err := client.ReviewStart(context.Background(), protocol.ReviewStartParams{ThreadID: "thread", Target: target, Delivery: &delivery}); err != nil {
		t.Fatal(err)
	}
	requests := transport.writtenRequests()
	if len(requests) != 1 || requests[0].Method != "review/start" {
		t.Fatalf("review requests: %#v", requests)
	}
	var fields map[string]json.RawMessage
	if err := json.Unmarshal(requests[0].Params, &fields); err != nil {
		t.Fatal(err)
	}
	if string(fields["delivery"]) != `"inline"` || string(fields["threadId"]) != `"thread"` {
		t.Fatalf("review params: %s", requests[0].Params)
	}
	//lint:ignore SA1019 compatibility fixture proves the retired review constant remains available
	if protocol.ReviewDeliveryDetached != "detached" {
		t.Fatal("deprecated review constant changed")
	}
}

func TestProtocol0166HookNotifications(t *testing.T) {
	for _, method := range []string{"hook/started", "hook/completed"} {
		note, err := parseServerNotification(method, json.RawMessage(`{"threadId":"thread","run":{"id":"hook","source":"cloudRequirements","sourcePath":"C:\\hooks.json","eventName":"preModelRequest"}}`))
		if err != nil {
			t.Fatal(err)
		}
		var run protocol.HookRunSummary
		switch value := note.Params.(type) {
		case protocol.HookStartedNotification:
			run = value.Run
		case protocol.HookCompletedNotification:
			run = value.Run
		default:
			t.Fatalf("unexpected hook notification: %T", note.Params)
		}
		if run.ID != "hook" || run.SourcePath != `C:\hooks.json` || run.Source != protocol.HookSourceCloudRequirements {
			t.Fatalf("hook notification changed: %#v", run)
		}
	}
}

func TestProtocol0166ProviderErrorData(t *testing.T) {
	for _, data := range []string{
		`{"type":"modelProviderNotConfigured","modelProvider":"gcp","future":9007199254740993}`,
		`{"unknown":9007199254740993}`,
	} {
		var response JSONRPCError
		if err := json.Unmarshal([]byte(`{"id":1,"error":{"code":-32000,"message":"provider error","data":`+data+`}}`), &response); err != nil {
			t.Fatal(err)
		}
		if string(response.Error.Data) != data {
			t.Fatalf("JSON-RPC error data changed: %s", response.Error.Data)
		}
	}
}
