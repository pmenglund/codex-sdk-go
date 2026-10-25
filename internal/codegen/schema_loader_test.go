package main

import (
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/atombender/go-jsonschema/pkg/generator"
	"github.com/atombender/go-jsonschema/pkg/schemas"
)

func TestMCPToolListMetadataUsesRawJSON(t *testing.T) {
	for _, nested := range []bool{false, true} {
		dir := t.TempDir()
		status := `{"title":"McpServerStatus","type":"object","properties":{"toolListMeta":true,"serverCapabilities":true}}`
		document := status
		if nested {
			document = `{"title":"Response","type":"object","properties":{"status":{"$ref":"#/definitions/McpServerStatus"}},"definitions":{"McpServerStatus":` + status + `}}`
		}
		path := filepath.Join(dir, "schema.json")
		if err := os.WriteFile(path, []byte(document), 0o600); err != nil {
			t.Fatal(err)
		}
		gen, err := generator.New(generator.Config{
			DefaultPackageName: "protocol", DefaultOutputName: "types.go",
			Capitalizations: []string{"MCP"}, OnlyModels: true,
			Loader: protocolSchemaLoader{schemas.NewDefaultCacheLoader(nil, nil)},
		})
		if err != nil {
			t.Fatal(err)
		}
		if err := gen.DoFile(path); err != nil {
			t.Fatal(err)
		}
		sources, err := gen.Sources()
		if err != nil {
			t.Fatal(err)
		}
		source := string(sources["types.go"])
		if !strings.Contains(source, "ToolListMeta json.RawMessage") || !strings.Contains(source, "ServerCapabilities interface{}") || !strings.Contains(source, `"encoding/json"`) {
			t.Fatalf("incorrect opaque field representation (nested=%v):\n%s", nested, source)
		}
	}
}

func TestFeedbackSavePathUsesNativePathString(t *testing.T) {
	dir := t.TempDir()
	path := filepath.Join(dir, "FeedbackSaveParams.json")
	if err := os.WriteFile(path, []byte(`{
		"title":"FeedbackSaveParams","type":"object",
		"properties":{"savePath":{"allOf":[{"$ref":"#/definitions/LegacyAppPathString"}]}},
		"required":["savePath"],"definitions":{"LegacyAppPathString":{"type":"string"}}
	}`), 0o600); err != nil {
		t.Fatal(err)
	}
	gen, err := generator.New(generator.Config{
		DefaultPackageName: "protocol", DefaultOutputName: "types.go", OnlyModels: true,
		Loader: protocolSchemaLoader{schemas.NewDefaultCacheLoader(nil, nil)},
	})
	if err != nil {
		t.Fatal(err)
	}
	if err := safeDoFile(gen, path); err != nil {
		t.Fatal(err)
	}
	sources, err := gen.Sources()
	if err != nil {
		t.Fatal(err)
	}
	if source := string(sources["types.go"]); !strings.Contains(source, "SavePath LegacyAppPathString") {
		t.Fatalf("feedback path lost its wire type:\n%s", source)
	}
}
