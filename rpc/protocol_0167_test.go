package rpc

import (
	"context"
	"encoding/json"
	"testing"

	"github.com/pmenglund/codex-sdk-go/protocol"
)

func TestProtocol0167SectionNotifications(t *testing.T) {
	for _, section := range []string{`"section"`, `null`} {
		payload := `{"threadId":"thread","sectionId":` + section + `}`
		note, err := parseServerNotification("thread/section/changed", json.RawMessage(payload))
		if err != nil {
			t.Fatal(err)
		}
		params, ok := note.Params.(protocol.ThreadSectionChangedNotification)
		if !ok || params.ThreadID != "thread" {
			t.Fatalf("section notification: %#v", note.Params)
		}
		encoded, err := json.Marshal(params)
		if err != nil {
			t.Fatal(err)
		}
		var fields map[string]json.RawMessage
		if err := json.Unmarshal(encoded, &fields); err != nil {
			t.Fatal(err)
		}
		if string(fields["sectionId"]) != section || string(note.Raw) != payload {
			t.Fatalf("section notification changed: %s (raw %s)", encoded, note.Raw)
		}
	}
}

type humanApproval0167Handler struct {
	UnimplementedServerRequestHandler
	kind protocol.CommandExecutionApprovalKind
}

func (h *humanApproval0167Handler) ItemCommandExecutionRequestApproval(_ context.Context, params protocol.CommandExecutionRequestApprovalParams) (*protocol.CommandExecutionRequestApprovalResponse, error) {
	h.kind = params.Kind
	return &protocol.CommandExecutionRequestApprovalResponse{Decision: protocol.MustCommandExecutionApprovalDecision("decline")}, nil
}

func TestProtocol0167HumanApprovalDispatch(t *testing.T) {
	handler := &humanApproval0167Handler{}
	result, err := dispatchServerRequest(context.Background(), handler, JSONRPCRequest{
		ID: NewIntRequestID(1), Method: "item/commandExecution/requestApproval",
		Params: json.RawMessage(`{"threadId":"thread","turnId":"turn","itemId":"item","kind":"humanCommand","startedAtMs":0,"environmentId":null}`),
	})
	if err != nil {
		t.Fatal(err)
	}
	if handler.kind != protocol.CommandExecutionApprovalKindHumanCommand {
		t.Fatalf("approval kind changed: %q", handler.kind)
	}
	encoded, err := json.Marshal(result)
	if err != nil || string(encoded) != `{"decision":"decline"}` {
		t.Fatalf("human approval response: %s, %v", encoded, err)
	}
}
