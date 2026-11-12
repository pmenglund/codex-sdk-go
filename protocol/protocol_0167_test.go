package protocol

import (
	"encoding/json"
	"testing"
)

func TestCodex0167AudioReferences(t *testing.T) {
	for _, test := range []struct {
		name, inline, uploaded, missing string
		newValue                        func() any
	}{
		{"user input", `{"type":"audio","url":"data:audio/wav;base64,YQ=="}`, `{"type":"audio","fileId":"file-audio","future":9007199254740993}`, `{"type":"audio"}`, func() any { return &UserInput{} }},
		{"content item", `{"type":"input_audio","audio_url":"data:audio/wav;base64,YQ=="}`, `{"type":"input_audio","file_id":"file-audio","future":9007199254740993}`, `{"type":"input_audio"}`, func() any { return &ContentItem{} }},
	} {
		t.Run(test.name, func(t *testing.T) {
			for _, payload := range []string{test.inline, test.uploaded} {
				value := test.newValue()
				if err := json.Unmarshal([]byte(payload), value); err != nil {
					t.Fatalf("decode %s: %v", payload, err)
				}
				encoded, err := json.Marshal(value)
				if err != nil || string(encoded) != payload {
					t.Fatalf("audio reference changed: %s, %v", encoded, err)
				}
			}
			if err := json.Unmarshal([]byte(test.missing), test.newValue()); err == nil {
				t.Fatal("accepted audio without a URL or file ID")
			}
		})
	}
}

func TestCodex0167OptionalFields(t *testing.T) {
	for _, test := range []struct {
		name, payload, field string
		newValue             func() any
	}{
		{"human approval capability", `{"requiredHumanApproval":true}`, "requiredHumanApproval", func() any { return &InitializeCapabilities{} }},
		{"false SSE deltas", `{"log_sse_deltas":false}`, "log_sse_deltas", func() any { return &OtelRequirements{} }},
	} {
		t.Run(test.name, func(t *testing.T) {
			value := test.newValue()
			if err := json.Unmarshal([]byte(test.payload), value); err != nil {
				t.Fatal(err)
			}
			check0162Fields(t, test.payload, value, test.field)
			value = test.newValue()
			if err := json.Unmarshal([]byte(`{}`), value); err != nil {
				t.Fatal(err)
			}
			check0162Fields(t, `{}`, value, test.field)
		})
	}
}
