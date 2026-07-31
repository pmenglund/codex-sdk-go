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
		"timed_out",
		"abort",
		map[string]any{"denied": map[string]any{"rejection": "approval rejected by policy"}},
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
		"denied",
		map[string]any{"approved": map[string]any{}},
		map[string]any{"approved_for_session": map[string]any{}},
		map[string]any{"denied": map[string]any{}},
		map[string]any{"denied": map[string]any{"rejection": ""}},
		map[string]any{"denied": map[string]any{"rejection": nil}},
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

	const deniedJSON = `{"denied":{"rejection":"approval rejected by policy"}}`
	var denied ReviewDecision
	if err := json.Unmarshal([]byte(deniedJSON), &denied); err != nil {
		t.Fatalf("decode structured denial: %v", err)
	}
	roundTrip, err := json.Marshal(denied)
	if err != nil {
		t.Fatalf("marshal structured denial: %v", err)
	}
	if string(roundTrip) != deniedJSON {
		t.Fatalf("structured denial round trip = %s, want %s", roundTrip, deniedJSON)
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
			name:      "user audio input",
			payload:   `{"type":"audio","url":"https://example.com/audio.mp3"}`,
			newValue:  func(value any) (json.Marshaler, error) { wrapped, err := NewUserInput(value); return wrapped, err },
			kind:      func(value json.Marshaler) string { return string(value.(UserInput).Kind()) },
			knownKind: "audio",
		},
		{
			name:      "user local audio input",
			payload:   `{"type":"localAudio","path":"/tmp/audio.mp3"}`,
			newValue:  func(value any) (json.Marshaler, error) { wrapped, err := NewUserInput(value); return wrapped, err },
			kind:      func(value json.Marshaler) string { return string(value.(UserInput).Kind()) },
			knownKind: "localAudio",
		},
		{
			name:    "amazon bedrock login params",
			payload: `{"type":"amazonBedrock","apiKey":"test-key","region":"eu-north-1"}`,
			newValue: func(value any) (json.Marshaler, error) {
				wrapped, err := NewLoginAccountParams(value)
				return wrapped, err
			},
			kind:      func(value json.Marshaler) string { return string(value.(LoginAccountParams).Kind()) },
			knownKind: "amazonBedrock",
		},
		{
			name:    "amazon bedrock login response",
			payload: `{"type":"amazonBedrock"}`,
			newValue: func(value any) (json.Marshaler, error) {
				wrapped, err := NewLoginAccountResponse(value)
				return wrapped, err
			},
			kind:      func(value json.Marshaler) string { return string(value.(LoginAccountResponse).Kind()) },
			knownKind: "amazonBedrock",
		},
		{
			name:    "daily schedule",
			payload: `{"type":"daily","time":"09:00"}`,
			newValue: func(value any) (json.Marshaler, error) {
				wrapped, err := NewScheduledTaskSchedule(value)
				return wrapped, err
			},
			kind:      func(value json.Marshaler) string { return string(value.(ScheduledTaskSchedule).Kind()) },
			knownKind: "daily",
		},
		{
			name:    "hourly schedule",
			payload: `{"type":"hourly","intervalHours":2}`,
			newValue: func(value any) (json.Marshaler, error) {
				wrapped, err := NewScheduledTaskSchedule(value)
				return wrapped, err
			},
			kind:      func(value json.Marshaler) string { return string(value.(ScheduledTaskSchedule).Kind()) },
			knownKind: "hourly",
		},
		{
			name:    "weekdays schedule",
			payload: `{"type":"weekdays","time":"09:00"}`,
			newValue: func(value any) (json.Marshaler, error) {
				wrapped, err := NewScheduledTaskSchedule(value)
				return wrapped, err
			},
			kind:      func(value json.Marshaler) string { return string(value.(ScheduledTaskSchedule).Kind()) },
			knownKind: "weekdays",
		},
		{
			name:    "weekly schedule",
			payload: `{"type":"weekly","days":["monday"],"time":"09:00"}`,
			newValue: func(value any) (json.Marshaler, error) {
				wrapped, err := NewScheduledTaskSchedule(value)
				return wrapped, err
			},
			kind:      func(value json.Marshaler) string { return string(value.(ScheduledTaskSchedule).Kind()) },
			knownKind: "weekly",
		},
		{
			name:      "content audio",
			payload:   `{"type":"input_audio","audio_url":"https://example.com/audio.mp3"}`,
			newValue:  func(value any) (json.Marshaler, error) { wrapped, err := NewContentItem(value); return wrapped, err },
			kind:      func(value json.Marshaler) string { return string(value.(ContentItem).Kind()) },
			knownKind: "input_audio",
		},
		{
			name:    "dynamic tool audio",
			payload: `{"type":"inputAudio","audioUrl":"https://example.com/audio.mp3"}`,
			newValue: func(value any) (json.Marshaler, error) {
				wrapped, err := NewDynamicToolCallOutputContentItem(value)
				return wrapped, err
			},
			kind:      func(value json.Marshaler) string { return string(value.(DynamicToolCallOutputContentItem).Kind()) },
			knownKind: "inputAudio",
		},
		{
			name:    "function call audio",
			payload: `{"type":"input_audio","audio_url":"https://example.com/audio.mp3"}`,
			newValue: func(value any) (json.Marshaler, error) {
				wrapped, err := NewFunctionCallOutputContentItem(value)
				return wrapped, err
			},
			kind:      func(value json.Marshaler) string { return string(value.(FunctionCallOutputContentItem).Kind()) },
			knownKind: "input_audio",
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
	if _, err := NewLoginAccountParams(map[string]any{"type": "amazonBedrock", "apiKey": "test-key"}); err == nil || !strings.Contains(err.Error(), "region") {
		t.Fatalf("expected missing Amazon Bedrock region validation, got %v", err)
	}
	if _, err := NewScheduledTaskSchedule(map[string]any{"type": "weekly", "time": "09:00"}); err == nil || !strings.Contains(err.Error(), "days") {
		t.Fatalf("expected missing weekly days validation, got %v", err)
	}
	if _, err := NewContentItem(map[string]any{"type": "input_audio"}); err == nil || !strings.Contains(err.Error(), "audio_url") {
		t.Fatalf("expected missing audio_url validation, got %v", err)
	}
}

func TestRawResponseCompletedNotificationUsage(t *testing.T) {
	for name, payload := range map[string]string{
		"absent": `{"responseId":"response","threadId":"thread","turnId":"turn"}`,
		"null":   `{"responseId":"response","threadId":"thread","turnId":"turn","usage":null}`,
	} {
		t.Run(name, func(t *testing.T) {
			var notification RawResponseCompletedNotification
			if err := json.Unmarshal([]byte(payload), &notification); err != nil {
				t.Fatalf("unmarshal notification: %v", err)
			}
			if notification.Usage != nil {
				t.Fatalf("usage = %#v, want nil", notification.Usage)
			}
		})
	}

	var notification RawResponseCompletedNotification
	if err := json.Unmarshal([]byte(`{"responseId":"response","threadId":"thread","turnId":"turn","usage":{"cachedInputTokens":1,"inputTokens":2,"outputTokens":3,"reasoningOutputTokens":4,"totalTokens":5}}`), &notification); err != nil {
		t.Fatalf("unmarshal populated usage: %v", err)
	}
	if notification.Usage == nil || notification.Usage.InputTokens != 2 || notification.Usage.TotalTokens != 5 {
		t.Fatalf("unexpected usage: %#v", notification.Usage)
	}
}

func jsonEqual(left, right any) bool {
	leftJSON, _ := json.Marshal(left)
	rightJSON, _ := json.Marshal(right)
	return string(leftJSON) == string(rightJSON)
}
