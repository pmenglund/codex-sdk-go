package protocol

import (
	"bytes"
	"encoding/json"
	"testing"
)

func TestCodex0163ThreadExclusions(t *testing.T) {
	for _, excluded := range []string{"", `null`, `[]`, `["thread-a","thread-b"]`} {
		payload := `{}`
		if excluded != "" {
			payload = `{"excludedThreadIds":` + excluded + `}`
		}
		var params ThreadListParams
		if err := json.Unmarshal([]byte(payload), &params); err != nil {
			t.Fatal(err)
		}
		check0162Fields(t, payload, params, "excludedThreadIds")
	}
}

func TestCodex0163RetainedToolOutput(t *testing.T) {
	for _, retain := range []string{"", `false`, `true`} {
		payload := `{"name":"tool","output":"result"`
		if retain != "" {
			payload += `,"retain":` + retain
		}
		payload += `}`
		var output TurnToolOutput
		if err := json.Unmarshal([]byte(payload), &output); err != nil {
			t.Fatal(err)
		}
		check0162Fields(t, payload, output, "retain")
	}
}

func TestCodex0163ResponseMetadataRemainsLossless(t *testing.T) {
	for _, payload := range []string{
		`{"type":"message","role":"assistant","content":[{"type":"output_text","text":"answer","annotations":[{"count":9007199254740993}],"logprobs":[9007199254740993]}],"status":"completed","encrypted_content":"encrypted"}`,
		`{"type":"function_call","call_id":"call","name":"tool","arguments":"{}","status":"completed","encrypted_content":"encrypted"}`,
		`{"type":"context_compaction","encrypted_content":"encrypted","future":9007199254740993}`,
	} {
		var item ResponseItem
		if err := json.Unmarshal([]byte(payload), &item); err != nil {
			t.Fatal(err)
		}
		encoded, err := json.Marshal(item)
		if err != nil || !bytes.Equal(encoded, []byte(payload)) {
			t.Fatalf("response metadata changed: %s, %v", encoded, err)
		}
		if !item.IsKnown() {
			t.Fatalf("unrecognized response item %s", item.Kind())
		}
	}
}

func TestCodex0163ReadStateUnions(t *testing.T) {
	for _, payload := range []string{
		`{"type":"threadStart"}`,
		`{"type":"turn","turnId":"turn"}`,
		// Raw wrappers validate required-key presence; values remain server-validated.
		`{"type":"turn","turnId":null}`,
		`{"type":"future","count":9007199254740993}`,
	} {
		var position ThreadUnreadPosition
		if err := json.Unmarshal([]byte(payload), &position); err != nil {
			t.Fatal(err)
		}
		encoded, err := json.Marshal(position)
		if err != nil || !bytes.Equal(encoded, []byte(payload)) {
			t.Fatalf("unread position changed: %s, %v", encoded, err)
		}
		if position.IsKnown() != (string(position.Kind()) != "future") {
			t.Fatalf("incorrect known position: %s", position.Kind())
		}
	}
	for _, payload := range []string{`{}`, `{"type":"turn"}`} {
		var position ThreadUnreadPosition
		if err := json.Unmarshal([]byte(payload), &position); err == nil {
			t.Fatalf("accepted incomplete unread position: %s", payload)
		}
	}
	for _, payload := range []string{`{"type":"read"}`, `{"type":"unread"}`, `{"type":"future","count":9007199254740993}`} {
		var operation ThreadReadStateOperation
		if err := json.Unmarshal([]byte(payload), &operation); err != nil {
			t.Fatal(err)
		}
		encoded, err := json.Marshal(operation)
		if err != nil || !bytes.Equal(encoded, []byte(payload)) {
			t.Fatalf("read-state operation changed: %s, %v", encoded, err)
		}
		if operation.IsKnown() != (string(operation.Kind()) != "future") {
			t.Fatalf("incorrect known operation: %s", operation.Kind())
		}
	}
}
