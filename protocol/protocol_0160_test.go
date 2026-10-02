package protocol

import (
	"encoding/json"
	"reflect"
	"testing"
)

func TestProtocol0160ManualFieldsRoundTrip(t *testing.T) {
	for _, tt := range []struct {
		name   string
		value  any
		input  string
		fields []string
	}{
		{"start", &ThreadStartResponse{}, `{"disabledPluginIds":["plugin_1"]}`, []string{"disabledPluginIds"}},
		{"fork", &ThreadForkResponse{}, `{"disabledPluginIds":["plugin_1"]}`, []string{"disabledPluginIds"}},
		{"resume", &ThreadResumeResponse{}, `{"disabledPluginIds":[],"collaborationMode":{"mode":"default","settings":{"model":"model_1","reasoning_effort":"high","developer_instructions":"continue"}}}`, []string{"disabledPluginIds", "collaborationMode"}},
		{"item", &ThreadItemEntry{}, `{"turnId":"t","item":{"type":"agentMessage","id":"i","text":"done"},"startedAtMs":0,"completedAtMs":12}`, []string{"startedAtMs", "completedAtMs"}},
		{"turn", &TurnStartParams{}, `{"threadId":"t","input":[],"disabledPluginIds":[]}`, []string{"disabledPluginIds"}},
	} {
		t.Run(tt.name, func(t *testing.T) {
			if err := json.Unmarshal([]byte(tt.input), tt.value); err != nil {
				t.Fatal(err)
			}
			data, err := json.Marshal(tt.value)
			if err != nil {
				t.Fatal(err)
			}
			var got, want map[string]json.RawMessage
			if err := json.Unmarshal(data, &got); err != nil {
				t.Fatal(err)
			}
			if err := json.Unmarshal([]byte(tt.input), &want); err != nil {
				t.Fatal(err)
			}
			for _, field := range tt.fields {
				var g, w any
				if err := json.Unmarshal(got[field], &g); err != nil {
					t.Errorf("decode %s: %v", field, err)
					continue
				}
				if err := json.Unmarshal(want[field], &w); err != nil {
					t.Fatal(err)
				}
				if !reflect.DeepEqual(g, w) {
					t.Errorf("%s = %s, want %s", field, got[field], want[field])
				}
			}
		})
	}
}

func TestProtocol0160ImageAlternatives(t *testing.T) {
	for _, tt := range []struct {
		name     string
		newValue func() any
		kind     string
		url      string
		fileID   string
	}{
		{"input", func() any { return &UserInput{} }, "image", "url", "fileId"},
		{"content", func() any { return &ContentItem{} }, "input_image", "image_url", "file_id"},
		{"output", func() any { return &FunctionCallOutputContentItem{} }, "input_image", "image_url", "file_id"},
	} {
		t.Run(tt.name, func(t *testing.T) {
			for _, field := range []string{tt.url, tt.fileID} {
				payload := `{"type":"` + tt.kind + `","` + field + `":"image_1","future":true}`
				value := tt.newValue()
				if err := json.Unmarshal([]byte(payload), value); err != nil {
					t.Fatal(err)
				}
				data, err := json.Marshal(value)
				if err != nil || string(data) != payload {
					t.Fatalf("round trip = %s, %v", data, err)
				}
			}
			if err := json.Unmarshal([]byte(`{"type":"`+tt.kind+`"}`), tt.newValue()); err == nil {
				t.Fatal("accepted image without URL or file ID")
			}
		})
	}
}

func TestProtocol0160HistoryCursor(t *testing.T) {
	text, limit := "next", 7
	legacy := SanitizedThreadItemsListParamsJSON{
		ThreadID: "thread",
		Cursor:   SanitizedThreadItemsListParamsJSONCursor(&text),
		Limit:    SanitizedThreadItemsListParamsJSONLimit(&limit),
		TurnID:   SanitizedThreadItemsListParamsJSONTurnID(&text),
	}
	if _, err := json.Marshal(legacy); err != nil {
		t.Fatal(err)
	}
	for _, payload := range []string{
		`{"cursor":"next","threadId":"thread"}`,
		`{"cursor":{"type":"item","itemId":"item"},"threadId":"thread","turnId":"turn"}`,
		`{"cursor":{"type":"future","position":12},"threadId":"thread","turnId":"turn"}`,
	} {
		var params ThreadItemsListParams
		if err := json.Unmarshal([]byte(payload), &params); err != nil {
			t.Fatal(err)
		}
		data, err := json.Marshal(params)
		if err != nil || string(data) != payload {
			t.Fatalf("round trip = %s, %v", data, err)
		}
	}
	for _, payload := range []string{
		`{"cursor":{"type":"item","itemId":"item"},"threadId":"thread"}`,
		`{"cursor":{"type":"item"},"threadId":"thread","turnId":"turn"}`,
		`{"cursor":12,"threadId":"thread"}`,
	} {
		var params ThreadItemsListParams
		if err := json.Unmarshal([]byte(payload), &params); err == nil {
			t.Fatalf("accepted %s", payload)
		}
	}
	anchor, err := NewThreadItemsListAnchor(map[string]any{"type": "item", "itemId": "item"})
	if err != nil {
		t.Fatal(err)
	}
	cursor, turnID := "next", "turn"
	emptyTurnID := ""
	for _, tt := range []struct {
		name   string
		params ThreadItemsListParams
	}{
		{"conflicting cursors", ThreadItemsListParams{ThreadID: "thread", TurnID: &turnID, Cursor: &cursor, CursorAnchor: &anchor}},
		{"empty turn ID", ThreadItemsListParams{ThreadID: "thread", TurnID: &emptyTurnID, CursorAnchor: &anchor}},
		{"zero anchor", ThreadItemsListParams{ThreadID: "thread", TurnID: &turnID, CursorAnchor: &ThreadItemsListAnchor{}}},
	} {
		t.Run(tt.name, func(t *testing.T) {
			if _, err := json.Marshal(tt.params); err == nil {
				t.Fatal("accepted invalid anchor request")
			}
		})
	}
	if _, err := NewThreadItemsListCursor(nil); err == nil {
		t.Fatal("accepted null cursor")
	}
	var value ThreadItemsListCursor
	if err := json.Unmarshal([]byte("  null  "), &value); err == nil {
		t.Fatal("accepted whitespace-padded null cursor")
	}
	for _, payload := range []string{`"next"`, `{"type":"item","itemId":"item","future":true}`} {
		if err := json.Unmarshal([]byte(payload), &value); err != nil {
			t.Fatal(err)
		}
		data, err := json.Marshal(value)
		if err != nil || string(data) != payload {
			t.Fatalf("cursor round trip = %s, %v", data, err)
		}
	}
	data, err := json.Marshal(ThreadItemsListCursor{})
	if err != nil || string(data) != "null" {
		t.Fatalf("zero cursor = %s, %v", data, err)
	}
}

func TestProtocol0160AttachmentPayloadPreservesNumbers(t *testing.T) {
	for _, payload := range []string{`{"id":9007199254740993}`, `9007199254740993`, `null`} {
		var attachment ThreadAttachment
		if err := json.Unmarshal([]byte(`{"payload":`+payload+`}`), &attachment); err != nil {
			t.Fatal(err)
		}
		if string(attachment.Payload) != payload {
			t.Fatalf("attachment payload = %s", attachment.Payload)
		}
		params := ThreadAttachmentAddParams{ThreadID: "thread", Payload: attachment.Payload}
		data, err := json.Marshal(params)
		if err != nil {
			t.Fatal(err)
		}
		var decoded SanitizedThreadAttachmentAddParamsJSON
		if err := json.Unmarshal(data, &decoded); err != nil {
			t.Fatal(err)
		}
		if string(decoded.Payload) != payload {
			t.Fatalf("attachment request round trip = %s", decoded.Payload)
		}
	}
}

func TestProtocol0160ResourceTargetNullableLink(t *testing.T) {
	params := MCPResourceReadParams{Server: "apps", Uri: "ui://view"}
	for _, tt := range []struct {
		target *MCPResourceReadTarget
		want   string
	}{
		{nil, `{"server":"apps","uri":"ui://view"}`},
		{&MCPResourceReadTarget{ConnectorID: "app"}, `{"server":"apps","target":{"connectorId":"app","linkId":null},"uri":"ui://view"}`},
	} {
		params.Target = tt.target
		data, err := json.Marshal(params)
		if err != nil || string(data) != tt.want {
			t.Fatalf("resource request = %s, %v", data, err)
		}
	}
}

func TestProtocol0160PluginListOmissionAndClear(t *testing.T) {
	params := TurnStartParams{ThreadID: "thread"}
	data, err := json.Marshal(params)
	if err != nil {
		t.Fatal(err)
	}
	var object map[string]json.RawMessage
	if err := json.Unmarshal(data, &object); err != nil {
		t.Fatal(err)
	}
	if _, exists := object["disabledPluginIds"]; exists {
		t.Fatal("nil list must preserve saved settings")
	}
	empty := []string{}
	params.DisabledPluginIDs = &empty
	data, err = json.Marshal(params)
	if err != nil {
		t.Fatal(err)
	}
	if err := json.Unmarshal(data, &object); err != nil {
		t.Fatal(err)
	}
	if string(object["disabledPluginIds"]) != "[]" {
		t.Fatalf("clear list = %s", object["disabledPluginIds"])
	}
}
