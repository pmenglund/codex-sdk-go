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
	if strings.TrimSpace(output) != "Hello from replay" {
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
}
