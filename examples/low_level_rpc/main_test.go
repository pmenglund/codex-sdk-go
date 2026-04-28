package main

import (
	"io"
	"log/slog"
	"strings"
	"testing"

	"github.com/pmenglund/codex-sdk-go/examples/internal/testutil"
	"github.com/pmenglund/codex-sdk-go/protocol"
)

func TestMainReplay(t *testing.T) {
	t.Setenv(exampleReplayEnv, "1")

	output := testutil.CaptureOutput(main)
	expected := `models: {
  "data": [
    {
      "defaultReasoningEffort": "medium",
      "description": "Test Model",
      "displayName": "Test Model",
      "hidden": false,
      "id": "model-1",
      "isDefault": true,
      "model": "model-1",
      "supportedReasoningEfforts": [
        {
          "description": "Medium",
          "reasoningEffort": "medium"
        }
      ]
    }
  ]
}`
	if strings.TrimSpace(output) != expected {
		t.Fatalf("unexpected output: %q", output)
	}
}

func TestExampleOptionsDefault(t *testing.T) {
	t.Setenv(exampleReplayEnv, "")
	logger := slog.New(slog.NewTextHandler(io.Discard, nil))
	opts := exampleOptions(logger)
	if opts.Transport != nil {
		t.Fatalf("expected nil transport for default options")
	}

	info := exampleClientInfo()
	if info.Name == "" || info.Version == "" {
		t.Fatalf("unexpected client info: %#v", info)
	}
	if len(exampleTranscript(info)) == 0 {
		t.Fatalf("expected transcript entries")
	}
	if stringPtr("x") == nil {
		t.Fatalf("expected stringPtr value")
	}

	if formatModels(nil) != "models: <nil>" {
		t.Fatalf("unexpected nil format")
	}
	models := protocol.ModelListResponse{}
	if !strings.HasPrefix(formatModels(&models), "models: {") {
		t.Fatalf("expected structured formatting")
	}
}

func TestMustRawNil(t *testing.T) {
	if raw := mustRaw(nil); raw != nil {
		t.Fatalf("expected nil raw message, got %s", raw)
	}
}
