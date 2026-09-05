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

func TestAudioInputsUseProtocolFields(t *testing.T) {
	start, err := buildTurnParams("thr_1", []Input{
		AudioInput("https://example.com/audio.mp3"),
		LocalAudioInput("/tmp/audio.mp3"),
	}, nil)
	if err != nil {
		t.Fatalf("build turn params: %v", err)
	}
	data, err := json.Marshal(start)
	if err != nil {
		t.Fatalf("marshal turn params: %v", err)
	}
	encoded := string(data)
	for _, want := range []string{
		`{"type":"audio","url":"https://example.com/audio.mp3"}`,
		`{"type":"localAudio","path":"/tmp/audio.mp3"}`,
	} {
		if !strings.Contains(encoded, want) {
			t.Fatalf("turn/start payload omitted %s: %s", want, encoded)
		}
	}
}
