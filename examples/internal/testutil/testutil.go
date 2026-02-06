package testutil

import (
	"bytes"
	"os"
)

// CaptureOutput runs fn with stdout redirected and returns captured output.
func CaptureOutput(fn func()) string {
	original := os.Stdout
	r, w, err := os.Pipe()
	if err != nil {
		panic(err)
	}
	os.Stdout = w

	fn()

	_ = w.Close()
	os.Stdout = original

	var buf bytes.Buffer
	_, _ = buf.ReadFrom(r)
	_ = r.Close()
	return buf.String()
}
