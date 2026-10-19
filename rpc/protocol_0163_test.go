package rpc

import (
	"bytes"
	"encoding/json"
	"testing"

	"github.com/pmenglund/codex-sdk-go/protocol"
)

func TestProtocol0163ReadStateNotification(t *testing.T) {
	for _, position := range []string{`null`, `{"type":"threadStart"}`, `{"type":"turn","turnId":"turn"}`, `{"type":"future","count":9007199254740993}`} {
		payload := `{"threadId":"thread","readState":{"revision":"opaque-revision","firstUnread":` + position + `}}`
		note, err := parseServerNotification("thread/readState/changed", json.RawMessage(payload))
		if err != nil {
			t.Fatal(err)
		}
		state, ok := note.Params.(protocol.ThreadReadStateChangedNotification)
		if !ok || state.ThreadID != "thread" || state.ReadState.Revision != "opaque-revision" {
			t.Fatalf("read-state notification: %#v", note)
		}
		encoded, err := json.Marshal(state.ReadState.FirstUnread)
		if err != nil || !bytes.Equal(encoded, []byte(position)) {
			t.Fatalf("unread position changed: %s, %v", encoded, err)
		}
	}
	if _, err := parseServerNotification("thread/readState/changed", json.RawMessage(`{"threadId":"thread","readState":{"revision":"revision","firstUnread":{"type":"turn"}}}`)); err == nil {
		t.Fatal("accepted unread turn without turnId")
	}
}
