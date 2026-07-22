package codex

import (
	"encoding/json"
	"strings"
	"testing"
)

func TestMentionInputUsesProtocolTextElementsKey(t *testing.T) {
	mention := MentionInput("å")
	if got, want := mention.TextElements[0].ByteRange.End, len("@å"); got != want {
		t.Fatalf("mention byte range end = %d, want %d", got, want)
	}

	start, err := buildTurnParams("thr_1", []Input{mention}, nil)
	if err != nil {
		t.Fatalf("build turn params: %v", err)
	}
	steer, err := buildTurnSteerParams("thr_1", "turn_1", []Input{mention})
	if err != nil {
		t.Fatalf("build steer params: %v", err)
	}

	for name, value := range map[string]any{"turn/start": start, "turn/steer": steer} {
		data, err := json.Marshal(value)
		if err != nil {
			t.Fatalf("marshal %s: %v", name, err)
		}
		encoded := string(data)
		if !strings.Contains(encoded, `"text_elements"`) {
			t.Fatalf("%s payload omitted text_elements: %s", name, encoded)
		}
		if strings.Contains(encoded, `"textElements"`) {
			t.Fatalf("%s payload used legacy textElements: %s", name, encoded)
		}
	}
}
