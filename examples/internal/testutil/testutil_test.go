package testutil

import (
	"fmt"
	"os"
	"testing"
)

func TestCaptureOutput(t *testing.T) {
	original := os.Stdout
	output := CaptureOutput(func() {
		fmt.Print("hello")
	})
	if output != "hello" {
		t.Fatalf("unexpected output: %q", output)
	}
	if os.Stdout != original {
		t.Fatalf("expected stdout restored")
	}
}
