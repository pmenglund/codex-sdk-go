package protocol

import (
	"bytes"
	"encoding/json"
	"testing"
)

func TestCodex0164ImageGenerationResults(t *testing.T) {
	for _, result := range []string{
		`"result":"cG5n"`,
		`"fileId":"file-1"`,
		`"fileId":"file-1","result":""`,
	} {
		payload := `{"type":"imageGeneration","id":"image-1","status":"completed",` + result + `,"inputImages":[{"path":"/tmp/reference.png","sha256":"abcd"}],"future":9007199254740993}`
		var item ThreadItem
		if err := json.Unmarshal([]byte(payload), &item); err != nil {
			t.Fatal(err)
		}
		encoded, err := json.Marshal(item)
		if err != nil || !bytes.Equal(encoded, []byte(payload)) {
			t.Fatalf("image result changed: %s, %v", encoded, err)
		}
		if item.Kind() != ThreadItemKindImageGeneration || !item.IsKnown() {
			t.Fatalf("unrecognized image item: %s", item.Kind())
		}
	}
	for _, payload := range []string{
		`{"type":"imageGeneration","status":"completed","fileId":"file-1"}`,
		`{"type":"imageGeneration","id":"image-1","status":"completed"}`,
	} {
		var item ThreadItem
		if err := json.Unmarshal([]byte(payload), &item); err == nil {
			t.Fatalf("accepted incomplete image result: %s", payload)
		}
	}
}

func TestCodex0164ApplicationStorage(t *testing.T) {
	for _, storage := range []string{
		`null`,
		`{"defaultProjectParent":"/tmp/projects","projectlessWorkspaceRoot":"/tmp/workspaces"}`,
		`{"defaultProjectParent":"C:\\Users\\test\\projects","projectlessWorkspaceRoot":null}`,
	} {
		payload := `{"storage":` + storage + `}`
		var requirements ApplicationRequirements
		if err := json.Unmarshal([]byte(payload), &requirements); err != nil {
			t.Fatal(err)
		}
		// Optional null paths may be omitted on encoding.
		if storage != `null` {
			encoded, err := json.Marshal(requirements)
			if err != nil {
				t.Fatal(err)
			}
			var object map[string]json.RawMessage
			if err := json.Unmarshal(encoded, &object); err != nil {
				t.Fatal(err)
			}
			var before, after map[string]*string
			if err := json.Unmarshal([]byte(storage), &before); err != nil {
				t.Fatal(err)
			}
			if err := json.Unmarshal(object["storage"], &after); err != nil {
				t.Fatal(err)
			}
			for field, want := range before {
				got := after[field]
				if (got == nil) != (want == nil) || (want != nil && *got != *want) {
					t.Fatalf("storage path %s changed: %s", field, encoded)
				}
			}
		} else {
			check0162Fields(t, payload, requirements, "storage")
		}
	}
}

func TestCodex0164MCPToolListMetadataRemainsLossless(t *testing.T) {
	for _, metadata := range []string{`null`, `{"count":9007199254740993,"future":[true,"value"]}`} {
		payload := `{"name":"server","authStatus":"unsupported","tools":{},"resources":[],"resourceTemplates":[],"toolListMeta":` + metadata + `}`
		var status MCPServerStatus
		if err := json.Unmarshal([]byte(payload), &status); err != nil {
			t.Fatal(err)
		}
		encoded, err := json.Marshal(status)
		if err != nil {
			t.Fatal(err)
		}
		var fields map[string]json.RawMessage
		if err := json.Unmarshal(encoded, &fields); err != nil {
			t.Fatal(err)
		}
		if string(fields["toolListMeta"]) != metadata {
			t.Fatalf("tool-list metadata changed: %s", encoded)
		}
	}
}

func TestCodex0164ManagedModelHooks(t *testing.T) {
	payload := `{"PreModelRequest":[{"hooks":[{"type":"command","async":false,"command":"echo pre"}],"matcher":"pre"}],"PostModelRequest":[{"hooks":[{"type":"command","async":false,"command":"echo post"}],"matcher":"post"}]}`
	var requirements ManagedHooksRequirements
	if err := json.Unmarshal([]byte(payload), &requirements); err != nil {
		t.Fatal(err)
	}
	if len(requirements.PreModelRequest) != 1 || len(requirements.PostModelRequest) != 1 {
		t.Fatalf("managed model hooks were dropped: %#v", requirements)
	}
	check0162Fields(t, payload, requirements, "PreModelRequest", "PostModelRequest")
}

func TestCodex0164SciencePlanTypes(t *testing.T) {
	for plan, want := range map[PlanType]string{PlanTypeSci: "sci", PlanTypeSciPlus: "sci_plus"} {
		encoded, err := json.Marshal(plan)
		if err != nil || string(encoded) != `"`+want+`"` {
			t.Fatalf("plan type %q: %s, %v", plan, encoded, err)
		}
	}
}
