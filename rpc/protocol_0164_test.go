package rpc

import (
	"bytes"
	"context"
	"encoding/json"
	"testing"

	"github.com/pmenglund/codex-sdk-go/protocol"
)

func TestProtocol0164FeedbackSave(t *testing.T) {
	for _, path := range []string{"/tmp/feedback.zip", `C:\Users\test\feedback.zip`, `\\server\share\feedback.zip`} {
		transport := newScriptedTransport()
		client := NewClient(transport, ClientOptions{})
		t.Cleanup(func() { _ = client.Close() })
		encodedPath, err := json.Marshal(path)
		if err != nil {
			t.Fatal(err)
		}
		transport.enqueueResult(json.RawMessage(`{"threadId":"thread","savedPath":` + string(encodedPath) + `,"promptHash":null}`))
		includeLogs := false
		response, err := client.FeedbackSave(context.Background(), protocol.FeedbackSaveParams{
			Classification: "bug", SavePath: protocol.LegacyAppPathString(path), IncludeLogs: &includeLogs,
		})
		if err != nil {
			t.Fatal(err)
		}
		if response == nil || response.ThreadID != "thread" || response.SavedPath != protocol.LegacyAppPathString(path) || response.PromptHash != nil {
			t.Fatalf("feedback response: %#v", response)
		}
		requests := transport.writtenRequests()
		if len(requests) != 1 || requests[0].Method != "feedback/save" {
			t.Fatalf("feedback requests: %#v", requests)
		}
		var fields map[string]json.RawMessage
		if err := json.Unmarshal(requests[0].Params, &fields); err != nil {
			t.Fatal(err)
		}
		if len(fields) != 3 || string(fields["classification"]) != `"bug"` || string(fields["includeLogs"]) != "false" || !bytes.Equal(fields["savePath"], encodedPath) {
			t.Fatalf("feedback params: %s", requests[0].Params)
		}
	}
}

func TestProtocol0164FileBackedImageNotification(t *testing.T) {
	item := `{"type":"imageGeneration","id":"image-1","status":"completed","fileId":"file-1","inputImages":[{"path":"/tmp/reference.png","sha256":"abcd"}],"future":9007199254740993}`
	note, err := parseServerNotification("item/completed", json.RawMessage(`{"threadId":"thread","turnId":"turn","item":`+item+`}`))
	if err != nil {
		t.Fatal(err)
	}
	completed, ok := note.Params.(protocol.ItemCompletedNotification)
	if !ok || completed.ThreadID != "thread" || completed.TurnID != "turn" {
		t.Fatalf("image notification: %#v", note)
	}
	encoded, err := json.Marshal(completed.Item)
	if err != nil || string(encoded) != item || !completed.Item.IsKnown() {
		t.Fatalf("image payload changed: %s, %v", encoded, err)
	}
}
