package protocol

import (
	"encoding/json"
	"reflect"
	"testing"
)

func TestCodex0153ManualFieldsRoundTrip(t *testing.T) {
	for _, tt := range []struct {
		name    string
		payload string
		target  any
	}{
		{"thread", `{"model":"gpt-test","projectId":"project_1","reasoningEffort":"high","status":{"type":"idle"}}`, &Thread{}},
		{"resume", `{"threadId":"thr_1","excludeTurns":true}`, &ThreadResumeParams{}},
		{"fork", `{"threadId":"thr_1","excludeTurns":false}`, &ThreadForkParams{}},
		{"turn", `{"threadId":"thr_1","serviceTierForTurn":"default","turnTrigger":"user"}`, &TurnStartParams{}},
		{"tool output", `{"threadId":"thr_1","toolOutput":{"name":"lookup","namespace":"functions","output":"result"}}`, &TurnStartParams{}},
		{"approval", `{"kind":"command","threadId":"thr_1","turnId":"turn_1","itemId":"item_1"}`, &CommandExecutionRequestApprovalParams{}},
	} {
		t.Run(tt.name, func(t *testing.T) {
			if err := json.Unmarshal([]byte(tt.payload), tt.target); err != nil {
				t.Fatal(err)
			}
			encoded, err := json.Marshal(tt.target)
			if err != nil {
				t.Fatal(err)
			}
			var want, got map[string]any
			if err := json.Unmarshal([]byte(tt.payload), &want); err != nil {
				t.Fatal(err)
			}
			if err := json.Unmarshal(encoded, &got); err != nil {
				t.Fatal(err)
			}
			for field, value := range want {
				if actual, present := got[field]; !present || !reflect.DeepEqual(actual, value) {
					t.Errorf("field %s: got %#v (present %v), want %#v", field, actual, present, value)
				}
			}
		})
	}
}

func TestReviewDecisionMCPPolicyAmendment(t *testing.T) {
	decision, err := NewReviewDecision("approved_mcp_policy_amendment")
	if err != nil {
		t.Fatal(err)
	}
	if !decision.IsKnown() {
		t.Fatal("MCP policy amendment is not recognized")
	}
	encoded, err := json.Marshal(decision)
	if err != nil {
		t.Fatal(err)
	}
	if string(encoded) != `"approved_mcp_policy_amendment"` {
		t.Fatalf("unexpected decision: %s", encoded)
	}
}

func TestThreadSectionAppearanceUpdateStates(t *testing.T) {
	var clear *ThreadSectionAppearance
	color := "blue"
	replacement := &ThreadSectionAppearance{Color: &color}
	for _, tt := range []struct {
		name       string
		appearance **ThreadSectionAppearance
		want       string
	}{
		{"preserve", nil, `{"name":"Work","sectionId":"section_1"}`},
		{"clear", &clear, `{"appearance":null,"name":"Work","sectionId":"section_1"}`},
		{"replace", &replacement, `{"appearance":{"color":"blue"},"name":"Work","sectionId":"section_1"}`},
	} {
		t.Run(tt.name, func(t *testing.T) {
			encoded, err := json.Marshal(ThreadSectionUpdateParams{Appearance: tt.appearance, Name: "Work", SectionID: "section_1"})
			if err != nil {
				t.Fatal(err)
			}
			if string(encoded) != tt.want {
				t.Fatalf("got %s, want %s", encoded, tt.want)
			}
		})
	}
}

func TestCodex0153UnionRequiredFields(t *testing.T) {
	for _, tt := range []struct {
		name    string
		payload string
		target  any
	}{
		{"realtime identity", `{"type":"realtimeSessionStarted"}`, &ThreadRealtimeItem{}},
		{"hook metadata", `{"handlerType":"prompt"}`, &HookMetadata{}},
		{"image failure limit", `{"type":"usageLimitExceeded"}`, &ImageGenerationFailure{}},
	} {
		t.Run(tt.name, func(t *testing.T) {
			if err := json.Unmarshal([]byte(tt.payload), tt.target); err == nil {
				t.Fatal("accepted variant missing required fields")
			}
		})
	}

	payload := `{"type":"realtimeSessionStarted","id":"item_1","realtimeSessionId":"session_1"}`
	var item ThreadRealtimeItem
	if err := json.Unmarshal([]byte(payload), &item); err != nil {
		t.Fatal(err)
	}
	if !item.IsKnown() {
		t.Fatal("known realtime variant not recognized")
	}
	if string(item.RawJSON()) != payload {
		t.Fatal("realtime identity was not retained")
	}

	future := `{"type":"futureRealtimeItem","future":true}`
	if err := json.Unmarshal([]byte(future), &item); err != nil {
		t.Fatal(err)
	}
	if item.IsKnown() || string(item.RawJSON()) != future {
		t.Fatal("unknown realtime variant was not preserved")
	}
}

func TestApprovalKindDefaults(t *testing.T) {
	for _, tt := range []struct {
		payload string
		want    CommandExecutionApprovalKind
	}{
		{`{"threadId":"thr_1"}`, CommandExecutionApprovalKindCommand},
		{`{"kind":"command"}`, CommandExecutionApprovalKindCommand},
		{`{"kind":"writeStdin"}`, CommandExecutionApprovalKindWriteStdin},
		{`{"kind":"futureAction"}`, CommandExecutionApprovalKind("futureAction")},
	} {
		params := CommandExecutionRequestApprovalParams{Kind: CommandExecutionApprovalKindWriteStdin}
		if err := json.Unmarshal([]byte(tt.payload), &params); err != nil {
			t.Fatal(err)
		}
		if params.Kind != tt.want {
			t.Errorf("%s: kind %q, want %q", tt.payload, params.Kind, tt.want)
		}
	}
}

func TestApprovalKindRejectsInvalidValues(t *testing.T) {
	for _, payload := range []string{`{"kind":null}`, `{"kind":false}`, `{"kind":42}`, `{"kind":""}`} {
		var params CommandExecutionRequestApprovalParams
		if err := json.Unmarshal([]byte(payload), &params); err == nil {
			t.Errorf("accepted %s", payload)
		}
	}
}

func TestSingleVariantUnionConstructors(t *testing.T) {
	for _, tt := range []struct {
		name, payload string
		constructor   func(json.RawMessage) (bool, json.RawMessage, error)
	}{
		{"image failure", `{"type":"usageLimitExceeded","limitId":"images"}`, func(raw json.RawMessage) (bool, json.RawMessage, error) {
			v, e := NewImageGenerationFailure(raw)
			return v.IsKnown(), v.RawJSON(), e
		}},
		{"capability", `{"type":"environment","environmentId":"env_1","path":"/tmp"}`, func(raw json.RawMessage) (bool, json.RawMessage, error) {
			v, e := NewCapabilityRootLocation(raw)
			return v.IsKnown(), v.RawJSON(), e
		}},
		{"notification", `{"method":"initialized"}`, func(raw json.RawMessage) (bool, json.RawMessage, error) {
			v, e := NewClientNotification(raw)
			return v.IsKnown(), v.RawJSON(), e
		}},
		{"dynamic tool", `{"type":"function","description":"lookup","inputSchema":{},"name":"lookup"}`, func(raw json.RawMessage) (bool, json.RawMessage, error) {
			v, e := NewDynamicToolNamespaceTool(raw)
			return v.IsKnown(), v.RawJSON(), e
		}},
		{"local shell", `{"type":"exec","command":["pwd"]}`, func(raw json.RawMessage) (bool, json.RawMessage, error) {
			v, e := NewLocalShellAction(raw)
			return v.IsKnown(), v.RawJSON(), e
		}},
		{"summary", `{"type":"summary_text","text":"summary"}`, func(raw json.RawMessage) (bool, json.RawMessage, error) {
			v, e := NewReasoningItemReasoningSummary(raw)
			return v.IsKnown(), v.RawJSON(), e
		}},
	} {
		t.Run(tt.name, func(t *testing.T) {
			known, raw, err := tt.constructor(json.RawMessage(tt.payload))
			if err != nil {
				t.Fatal(err)
			}
			if !known || string(raw) != tt.payload {
				t.Fatalf("known=%v raw=%s", known, raw)
			}
		})
	}
}
