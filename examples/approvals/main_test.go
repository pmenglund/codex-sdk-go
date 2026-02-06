package main

import (
	"io"
	"log/slog"
	"strings"
	"testing"

	"github.com/pmenglund/codex-sdk-go/examples/internal/testutil"
)

func TestMainReplay(t *testing.T) {
	t.Setenv(exampleReplayEnv, "1")

	output := testutil.CaptureOutput(main)
	if strings.TrimSpace(output) != "Approved summary" {
		t.Fatalf("unexpected output: %q", output)
	}
}

func TestExampleOptionsDefault(t *testing.T) {
	t.Setenv(exampleReplayEnv, "")
	logger := slog.New(slog.NewTextHandler(io.Discard, nil))
	opts := exampleOptions("prompt", logger)
	if opts.Transport != nil {
		t.Fatalf("expected nil transport for default options")
	}
	if opts.ApprovalHandler == nil {
		t.Fatalf("expected approval handler for default options")
	}
}
