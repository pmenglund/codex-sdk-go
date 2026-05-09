//go:build e2e

package codex

import (
	"bytes"
	"context"
	"os/exec"
	"sync"
	"testing"
	"time"
)

func TestRealCodexStartThread(t *testing.T) {
	client, ctx, stderr := newRealCodexTestClient(t)

	thread, err := client.StartThread(ctx, ThreadStartOptions{Cwd: t.TempDir()})
	if err != nil {
		t.Fatalf("start real codex thread: %v\nstderr:\n%s", err, stderr.String())
	}
	if thread.ID() == "" {
		t.Fatalf("expected real codex thread id\nstderr:\n%s", stderr.String())
	}
}

func TestRealCodexGeneratedConfigRequirementsRead(t *testing.T) {
	client, ctx, stderr := newRealCodexTestClient(t)

	if _, err := client.Client().ConfigRequirementsRead(ctx); err != nil {
		t.Fatalf("read config requirements through generated RPC client: %v\nstderr:\n%s", err, stderr.String())
	}
}

func newRealCodexTestClient(t *testing.T) (*Codex, context.Context, *lockedBuffer) {
	t.Helper()

	codexPath, err := exec.LookPath("codex")
	if err != nil {
		t.Fatalf("codex must be available on PATH for e2e tests: %v", err)
	}

	t.Setenv("CODEX_HOME", t.TempDir())

	ctx, cancel := context.WithTimeout(context.Background(), 20*time.Second)
	t.Cleanup(cancel)

	var stderr lockedBuffer
	client, err := New(ctx, Options{
		Spawn: SpawnOptions{
			CodexPath: codexPath,
			Stderr:    &stderr,
		},
	})
	if err != nil {
		t.Fatalf("initialize real codex app-server: %v\nstderr:\n%s", err, stderr.String())
	}
	t.Cleanup(func() {
		if err := client.Close(); err != nil {
			t.Errorf("close real codex app-server: %v\nstderr:\n%s", err, stderr.String())
		}
	})
	return client, ctx, &stderr
}

type lockedBuffer struct {
	mu  sync.Mutex
	buf bytes.Buffer
}

func (b *lockedBuffer) Write(p []byte) (int, error) {
	b.mu.Lock()
	defer b.mu.Unlock()
	return b.buf.Write(p)
}

func (b *lockedBuffer) String() string {
	b.mu.Lock()
	defer b.mu.Unlock()
	return b.buf.String()
}
