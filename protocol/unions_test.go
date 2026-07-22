package protocol

import (
	"encoding/json"
	"strings"
	"testing"
)

func TestCommandExecutionApprovalDecision(t *testing.T) {
	for _, value := range []any{
		"accept",
		"acceptForSession",
		"decline",
		"cancel",
		map[string]any{"acceptWithExecpolicyAmendment": map[string]any{"execpolicyAmendment": []string{"git status"}}},
		map[string]any{"applyNetworkPolicyAmendment": map[string]any{"networkPolicyAmendment": map[string]any{}}},
	} {
		decision, err := NewCommandExecutionApprovalDecision(value)
		if err != nil {
			t.Fatalf("new decision %#v: %v", value, err)
		}
		if len(decision.RawJSON()) == 0 {
			t.Fatalf("decision %#v lost raw JSON", value)
		}
	}
	for _, value := range []any{
		"approve",
		"acceptWithExecpolicyAmendment",
		"applyNetworkPolicyAmendment",
		map[string]any{"accept": map[string]any{}},
		map[string]any{"acceptForSession": map[string]any{}},
		map[string]any{"decline": map[string]any{}},
		map[string]any{"cancel": map[string]any{}},
		map[string]any{"unknown": map[string]any{}},
		1,
	} {
		if _, err := NewCommandExecutionApprovalDecision(value); err == nil {
			t.Fatalf("expected invalid decision %#v to fail", value)
		}
		data, err := json.Marshal(value)
		if err != nil {
			t.Fatal(err)
		}
		var decoded CommandExecutionApprovalDecision
		if err := json.Unmarshal(data, &decoded); err == nil && isKnownCommandExecutionApprovalDecisionKind(decoded.Kind()) {
			t.Fatalf("wire decoder accepted known decision with invalid representation %#v", value)
		}
	}
}

func TestReviewDecision(t *testing.T) {
	for _, value := range []any{
		"approved",
		"approved_for_session",
		"denied",
		"timed_out",
		"abort",
		map[string]any{"approved_execpolicy_amendment": map[string]any{"proposed_execpolicy_amendment": []string{"git status"}}},
		map[string]any{"network_policy_amendment": map[string]any{"network_policy_amendment": map[string]any{}}},
	} {
		decision, err := NewReviewDecision(value)
		if err != nil {
			t.Fatalf("NewReviewDecision(%#v): %v", value, err)
		}
		if len(decision.RawJSON()) == 0 {
			t.Fatalf("NewReviewDecision(%#v) returned empty JSON", value)
		}
	}

	for _, value := range []any{
		"approve",
		"approved_execpolicy_amendment",
		"network_policy_amendment",
		map[string]any{"approved": map[string]any{}},
		map[string]any{"approved_for_session": map[string]any{}},
		map[string]any{"denied": map[string]any{}},
		map[string]any{"timed_out": map[string]any{}},
		map[string]any{"abort": map[string]any{}},
		map[string]any{"unknown": map[string]any{}},
		1,
	} {
		if _, err := NewReviewDecision(value); err == nil {
			t.Fatalf("NewReviewDecision(%#v) unexpectedly succeeded", value)
		}
		data, err := json.Marshal(value)
		if err != nil {
			t.Fatal(err)
		}
		var decoded ReviewDecision
		if err := json.Unmarshal(data, &decoded); err == nil && isKnownReviewDecisionKind(decoded.Kind()) {
			t.Fatalf("wire decoder accepted known decision with invalid representation %#v", value)
		}
	}
}

func TestGeneratedDiscriminatedUnionsRoundTrip(t *testing.T) {
	tests := []struct {
		name      string
		payload   string
		newValue  func(any) (json.Marshaler, error)
		kind      func(json.Marshaler) string
		knownKind string
	}{
		{
			name:      "sandbox policy",
			payload:   `{"type":"workspaceWrite","writableRoots":["/tmp/work"]}`,
			newValue:  func(value any) (json.Marshaler, error) { wrapped, err := NewSandboxPolicy(value); return wrapped, err },
			kind:      func(value json.Marshaler) string { return string(value.(SandboxPolicy).Kind()) },
			knownKind: "workspaceWrite",
		},
		{
			name:      "user input",
			payload:   `{"type":"text","text":"hello","text_elements":[]}`,
			newValue:  func(value any) (json.Marshaler, error) { wrapped, err := NewUserInput(value); return wrapped, err },
			kind:      func(value json.Marshaler) string { return string(value.(UserInput).Kind()) },
			knownKind: "text",
		},
		{
			name:      "thread status",
			payload:   `{"type":"active","activeFlags":["waitingOnApproval"]}`,
			newValue:  func(value any) (json.Marshaler, error) { wrapped, err := NewThreadStatus(value); return wrapped, err },
			kind:      func(value json.Marshaler) string { return string(value.(ThreadStatus).Kind()) },
			knownKind: "active",
		},
		{
			name:      "response item",
			payload:   `{"type":"message","role":"assistant","content":[]}`,
			newValue:  func(value any) (json.Marshaler, error) { wrapped, err := NewResponseItem(value); return wrapped, err },
			kind:      func(value json.Marshaler) string { return string(value.(ResponseItem).Kind()) },
			knownKind: "message",
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			var decoded any
			if err := json.Unmarshal([]byte(tt.payload), &decoded); err != nil {
				t.Fatalf("decode fixture: %v", err)
			}
			value, err := tt.newValue(decoded)
			if err != nil {
				t.Fatalf("wrap fixture: %v", err)
			}
			if got := tt.kind(value); got != tt.knownKind {
				t.Fatalf("unexpected kind: got %q want %q", got, tt.knownKind)
			}
			data, err := json.Marshal(value)
			if err != nil {
				t.Fatalf("marshal wrapped value: %v", err)
			}
			var roundTrip any
			if err := json.Unmarshal(data, &roundTrip); err != nil {
				t.Fatalf("decode round trip: %v", err)
			}
			if !jsonEqual(decoded, roundTrip) {
				t.Fatalf("round trip changed JSON: got %s want %s", data, tt.payload)
			}
		})
	}
}

func TestGeneratedUnionPreservesUnknownFutureVariant(t *testing.T) {
	value, err := NewUserInput(map[string]any{"type": "futureInput", "payload": map[string]any{"x": 1}})
	if err != nil {
		t.Fatalf("wrap future variant: %v", err)
	}
	if value.Kind() != "futureInput" || value.IsKnown() {
		t.Fatalf("unexpected future variant state: kind=%q known=%v", value.Kind(), value.IsKnown())
	}
	data, err := json.Marshal(value)
	if err != nil || !strings.Contains(string(data), `"payload"`) {
		t.Fatalf("future payload was not preserved: %s err=%v", data, err)
	}
}

func TestGeneratedUnionRejectsInvalidDiscriminator(t *testing.T) {
	for _, payload := range []string{`{}`, `{"type":null}`, `{"type":""}`, `[]`} {
		var value UserInput
		if err := json.Unmarshal([]byte(payload), &value); err == nil {
			t.Fatalf("expected %s to fail", payload)
		}
	}
	data, err := json.Marshal(UserInput{})
	if err != nil || string(data) != "null" {
		t.Fatalf("expected zero union value to marshal as null, got %s err=%v", data, err)
	}
}

func TestGeneratedUnionRejectsKnownVariantMissingRequiredField(t *testing.T) {
	if _, err := NewLoginAccountParams(map[string]any{"type": "apiKey"}); err == nil || !strings.Contains(err.Error(), "apiKey") {
		t.Fatalf("expected missing apiKey validation, got %v", err)
	}
	if _, err := NewUserInput(map[string]any{"type": "text"}); err == nil || !strings.Contains(err.Error(), "text") {
		t.Fatalf("expected missing text validation, got %v", err)
	}
}

func jsonEqual(left, right any) bool {
	leftJSON, _ := json.Marshal(left)
	rightJSON, _ := json.Marshal(right)
	return string(leftJSON) == string(rightJSON)
}
