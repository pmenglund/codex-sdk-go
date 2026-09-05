package test

import (
	"os"
	"path/filepath"
	"runtime"
	"strings"
	"testing"

	"github.com/pmenglund/codex-sdk-go/protocol"
)

func TestNewRealClientForwardsConfigOverrides(t *testing.T) {
	if runtime.GOOS == "windows" {
		t.Skip("fake CLI fixture requires a POSIX shell")
	}
	bin := t.TempDir()
	// Reject missing, split, reordered, or altered overrides before initialization.
	// No model provider or real Codex installation is needed for this transport check.
	script := `#!/bin/sh
if [ "$1" = "--version" ]; then
  printf 'codex-cli %s\n' '` + protocol.GeneratedCodexVersion + `'
  exit 0
fi
[ "$#" = 5 ] && [ "$1" = app-server ] &&
  [ "$2" = --config ] && [ "$3" = 'model="test model"' ] &&
  [ "$4" = --config ] && [ "$5" = 'model_provider="local"' ] || exit 43
IFS= read -r request || exit 44
case "$request" in
  *'"method":"initialize"'*) printf '%s\n' '{"id":1,"result":{}}' ;;
  *) exit 45 ;;
esac
while IFS= read -r request; do :; done
`
	if err := os.WriteFile(filepath.Join(bin, "codex"), []byte(script), 0o700); err != nil {
		t.Fatal(err)
	}
	t.Setenv("PATH", bin+string(os.PathListSeparator)+os.Getenv("PATH"))
	for _, recorded := range []bool{false, true} {
		name := "spawn"
		if recorded {
			name = "recorder"
		}
		t.Run(name, func(t *testing.T) {
			var recorder *RequestRecorder
			if recorded {
				recorder = &RequestRecorder{}
			}
			client, _, _ := NewRealClient(t, RealClientOptions{
				RequestRecorder: recorder,
				ConfigOverrides: []string{`model="test model"`, `model_provider="local"`},
			})
			if client.Client() == nil {
				t.Fatal("client did not initialize")
			}
			if recorded {
				if _, ok := recorder.Request("initialize"); !ok {
					t.Fatal("recorder missed initialization")
				}
			}
		})
	}
}

func TestLockedBufferRedactsStructuredAndLeafSecrets(t *testing.T) {
	raw := `{"type":"apiKey","apiKey":"secret-leaf"}`
	buffer := LockedBuffer{secrets: secretMarkers(raw)}
	if _, err := buffer.Write([]byte("payload=" + raw + " key=secret-leaf")); err != nil {
		t.Fatalf("write buffer: %v", err)
	}
	output := buffer.String()
	if strings.Contains(output, raw) || strings.Contains(output, "secret-leaf") {
		t.Fatalf("redacted output contains a credential: %q", output)
	}
	if !strings.Contains(output, "[REDACTED]") {
		t.Fatalf("expected redaction marker, got %q", output)
	}
	if !strings.Contains(buffer.rawString(), "secret-leaf") {
		t.Fatalf("raw capture should remain available for leak assertions")
	}
}
