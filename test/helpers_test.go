package test

import (
	"strings"
	"testing"
)

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
