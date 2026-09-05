package main

import (
	"encoding/json"
	"os"
	"path/filepath"
	"reflect"
	"runtime"
	"strings"
	"testing"
)

func TestManualSchemaCompositionAndCycles(t *testing.T) {
	root := t.TempDir()
	if err := os.MkdirAll(filepath.Join(root, "protocol"), 0o755); err != nil {
		t.Fatal(err)
	}
	schemas := t.TempDir()
	schema := `{
		"title":"Example",
		"definitions":{
			"Part/~":{"properties":{"id":{"type":"string"}},"required":["id"],"allOf":[{"$ref":"#/definitions/Cycle"}]},
			"Cycle":{"properties":{"shared":{"type":"string"}},"required":["shared"],"allOf":[{"$ref":"#/definitions/Part~1~0"}]}
		},
		"allOf":[
			{"$ref":"#/definitions/Part~1~0"},
			{"properties":{"extra":{"type":"string"}},"required":["extra"]},
			{"anyOf":[{"properties":{"optionalA":{"type":"string"}}},{"oneOf":[{"properties":{"optionalB":{"type":"string"}}}]}]}
		]
	}`
	if err := os.WriteFile(filepath.Join(schemas, "Example.json"), []byte(schema), 0o600); err != nil {
		t.Fatal(err)
	}
	old := requiredManualSchemaCoverage
	requiredManualSchemaCoverage = map[string]struct{}{"Example": {}}
	t.Cleanup(func() { requiredManualSchemaCoverage = old })
	for _, tt := range []struct{ name, extra, shared, want string }{
		{"complete", "extra", "shared", ""},
		{"missing inherited field", "-", "shared", "Example missing: extra"},
		{"required omitted", "extra", "shared,omitempty", "Example required with omitempty: shared"},
	} {
		t.Run(tt.name, func(t *testing.T) {
			source := "package protocol\ntype Example struct {\n" +
				"ID string " + "`json:\"id\"`" + "\n" +
				"Shared string " + "`json:\"" + tt.shared + "\"`" + "\n" +
				"Extra string " + "`json:\"" + tt.extra + "\"`" + "\n" +
				"OptionalA string " + "`json:\"optionalA,omitempty\"`" + "\n" +
				"OptionalB string " + "`json:\"optionalB,omitempty\"`" + "\n}\n"
			if err := os.WriteFile(filepath.Join(root, "protocol", "manual_types.go"), []byte(source), 0o600); err != nil {
				t.Fatal(err)
			}
			err := validateManualStructSchemaCoverage(schemas, root)
			if tt.want == "" {
				if err != nil {
					t.Fatal(err)
				}
			} else if err == nil || !strings.Contains(err.Error(), tt.want) {
				t.Fatalf("got %v, want %s", err, tt.want)
			}
		})
	}
	var document map[string]any
	if err := json.Unmarshal([]byte(schema), &document); err != nil {
		t.Fatal(err)
	}
	for _, ref := range []string{"#/definitions/missing", "#/title/child", "#/title"} {
		if _, ok := resolveSchemaReference(document, ref); ok {
			t.Errorf("resolved non-object reference %s", ref)
		}
	}
}

func TestCollectUnionDefinitionsAndDiscriminators(t *testing.T) {
	dir := t.TempDir()
	single := map[string]any{"required": []string{"id"}, "oneOf": []any{map[string]any{
		"properties": map[string]any{"handlerType": map[string]any{"enum": []string{"command"}}},
		"required":   []string{"handlerType"},
	}}}
	writeJSON(t, filepath.Join(dir, "Root.json"), map[string]any{
		"title": "Root", "oneOf": []any{
			map[string]any{"properties": map[string]any{"type": map[string]any{"enum": []string{"first"}}}},
			map[string]any{"properties": map[string]any{"type": map[string]any{"enum": []string{"second"}}}},
		},
		"definitions": map[string]any{"Handler": single, "McpServerElicitationRequestParams": single},
	})
	unions, err := collectDiscriminatedUnions(dir, false)
	if err != nil {
		t.Fatal(err)
	}
	if len(unions) != 2 || !reflect.DeepEqual(unions["Root"].Kinds, []string{"first", "second"}) {
		t.Fatalf("unions=%#v", unions)
	}
	if unions["Handler"].Discriminator != "handlerType" || !reflect.DeepEqual(unions["Handler"].Required["command"], []string{"handlerType", "id"}) {
		t.Fatalf("handler=%#v", unions["Handler"])
	}
	if _, err := collectDiscriminatedUnions(dir, true); err == nil || !strings.Contains(err.Error(), "required discriminated union") {
		t.Fatalf("missing core unions: %v", err)
	}
	duplicate := map[string]any{"oneOf": []any{
		map[string]any{"properties": map[string]any{"type": map[string]any{"enum": []string{"same"}}}},
		map[string]any{"properties": map[string]any{"type": map[string]any{"enum": []string{"same"}}}},
	}}
	for _, definitions := range []bool{false, true} {
		doc := map[string]any{"title": "Broken", "oneOf": duplicate["oneOf"]}
		if definitions {
			doc = map[string]any{"definitions": map[string]any{"Broken": duplicate}}
		}
		writeJSON(t, filepath.Join(dir, "Broken.json"), doc)
		if _, err := collectDiscriminatedUnions(dir, false); err == nil || !strings.Contains(err.Error(), "repeated type discriminator") {
			t.Fatalf("duplicate discriminator: %v", err)
		}
	}
}

func TestCodexPackageVersionFromManifest(t *testing.T) {
	root := t.TempDir()
	if _, err := codexPackageVersion(root); err == nil || !strings.Contains(err.Error(), "read codex Cargo.toml") {
		t.Fatalf("missing manifest: %v", err)
	}
	if err := os.Mkdir(filepath.Join(root, "codex-rs"), 0o755); err != nil {
		t.Fatal(err)
	}
	for _, tt := range []struct{ manifest, want string }{
		{"[package]\nversion = \"9.9.9\"\n[workspace.package]\nversion = \"0.153.4\"\n", "0.153.4"},
		{"[package]\nversion = \"9.9.9\"\n", ""},
	} {
		if err := os.WriteFile(filepath.Join(root, "codex-rs", "Cargo.toml"), []byte(tt.manifest), 0o600); err != nil {
			t.Fatal(err)
		}
		got, err := codexPackageVersion(root)
		if tt.want == "" {
			if err == nil || !strings.Contains(err.Error(), "workspace.package") {
				t.Fatalf("missing workspace version: %v", err)
			}
		} else if err != nil || got != tt.want {
			t.Fatalf("version=%q err=%v", got, err)
		}
	}
}

func TestExportSchemasRunsCLIAndPropagatesFailure(t *testing.T) {
	if runtime.GOOS == "windows" {
		t.Skip("fake cargo fixture requires a POSIX shell")
	}
	root := t.TempDir()
	cli := filepath.Join(root, "codex-rs", "cli")
	if err := os.MkdirAll(cli, 0o755); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(filepath.Join(cli, "Cargo.toml"), []byte("[package]\nname = \"codex-cli\"\n"), 0o600); err != nil {
		t.Fatal(err)
	}
	bin := t.TempDir()
	t.Setenv("PATH", bin+string(os.PathListSeparator)+os.Getenv("PATH"))
	output := t.TempDir()
	script := "#!/bin/sh\n[ \"$1 $2 $3 $4 $5 $6 $7 $8 $9\" = \"run -p codex-cli --bin codex -- app-server generate-json-schema --out\" ] || exit 13\nprintf '%s' \"$PWD\" > \"${10}/working-dir\"\n"
	if err := os.WriteFile(filepath.Join(bin, "cargo"), []byte(script), 0o700); err != nil {
		t.Fatal(err)
	}
	if err := exportSchemas(root, output); err != nil {
		t.Fatal(err)
	}
	cwd, err := os.ReadFile(filepath.Join(output, "working-dir"))
	if err != nil || string(cwd) != filepath.Join(root, "codex-rs") {
		t.Fatalf("export working directory=%q err=%v", cwd, err)
	}
	if err := os.WriteFile(filepath.Join(bin, "cargo"), []byte("#!/bin/sh\nexit 23\n"), 0o700); err != nil {
		t.Fatal(err)
	}
	if err := exportSchemas(root, output); err == nil || !strings.Contains(err.Error(), "exit status 23") {
		t.Fatalf("CLI failure: %v", err)
	}
}

func TestRPCGenerationFailsOnIncompleteSchemasAndBlockedOutputs(t *testing.T) {
	for _, tt := range []struct{ name, missing, blocked string }{
		{name: "missing client", missing: "ClientRequest.json"},
		{name: "missing server", missing: "ServerRequest.json"},
		{name: "missing notifications", missing: "ServerNotification.json"},
		{name: "output directory is a file", blocked: "rpc"},
		{name: "client output is a directory", blocked: "rpc/client_requests_gen.go"},
		{name: "server output is a directory", blocked: "rpc/server_requests_gen.go"},
		{name: "notification output is a directory", blocked: "rpc/notifications_gen.go"},
	} {
		t.Run(tt.name, func(t *testing.T) {
			root, schemas := t.TempDir(), t.TempDir()
			for _, file := range []string{"ClientRequest.json", "ServerRequest.json", "ServerNotification.json"} {
				if file != tt.missing {
					writeJSON(t, filepath.Join(schemas, file), map[string]any{"oneOf": []any{}})
				}
			}
			if tt.blocked == "rpc" {
				if err := os.WriteFile(filepath.Join(root, "rpc"), []byte("preserve"), 0o600); err != nil {
					t.Fatal(err)
				}
			} else if tt.blocked != "" {
				if err := os.MkdirAll(filepath.Join(root, tt.blocked), 0o755); err != nil {
					t.Fatal(err)
				}
			}
			err := generateRPCStubs(schemas, root, testCodexCommit)
			if err == nil {
				t.Fatal("generation succeeded despite missing input or blocked output")
			}
			wantPath := filepath.Join(schemas, tt.missing)
			if tt.blocked != "" {
				wantPath = filepath.Join(root, tt.blocked)
			}
			if !strings.Contains(err.Error(), wantPath) {
				t.Fatalf("error lost failing path %s: %v", wantPath, err)
			}
			if tt.blocked == "rpc" {
				data, err := os.ReadFile(filepath.Join(root, "rpc"))
				if err != nil || string(data) != "preserve" {
					t.Fatalf("overwrote blocking file: %q, %v", data, err)
				}
			}
		})
	}
}
