//go:build e2e

package e2e

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"net/http"
	"net/http/httptest"
	"os"
	"os/exec"
	"path/filepath"
	"runtime"
	"sync/atomic"
	"testing"

	codex "github.com/pmenglund/codex-sdk-go"
	"github.com/pmenglund/codex-sdk-go/protocol"
	"github.com/pmenglund/codex-sdk-go/rpc"
	"github.com/pmenglund/codex-sdk-go/test"
)

func TestRealCodexMockProviderDoesNotInheritCredentials(t *testing.T) {
	if runtime.GOOS == "windows" {
		t.Skip("credential-checking CLI wrapper requires a POSIX shell")
	}
	realCLI, err := exec.LookPath("codex")
	if err != nil {
		t.Fatal(err)
	}
	bin := t.TempDir()
	// This wrapper rejects credentials before either version probing or app-server startup.
	script := "#!/bin/sh\n[ -z \"$OPENAI_API_KEY\" ] && [ -z \"$CODEX_E2E_LOGIN_PARAMS_JSON\" ] || exit 42\nexec \"$CODEX_TEST_REAL_CLI\" \"$@\"\n"
	if err := os.WriteFile(filepath.Join(bin, "codex"), []byte(script), 0o700); err != nil {
		t.Fatal(err)
	}
	t.Setenv("CODEX_TEST_REAL_CLI", realCLI)
	t.Setenv("PATH", bin+string(os.PathListSeparator)+os.Getenv("PATH"))
	for _, recorded := range []bool{false, true} {
		name := "spawn"
		if recorded {
			name = "recorder"
		}
		t.Run(name, func(t *testing.T) {
			t.Setenv("OPENAI_API_KEY", "synthetic-api-secret")
			t.Setenv("CODEX_E2E_LOGIN_PARAMS_JSON", `{"type":"apiKey","apiKey":"synthetic-login-secret"}`)
			var recorder *test.RequestRecorder
			if recorded {
				recorder = &test.RequestRecorder{}
			}
			client, ctx, stderr, _ := newArchiveTestClient(t, recorder)
			test.StartThread(t, client, ctx, stderr, t.TempDir())
		})
	}
}

func newArchiveTestClient(t *testing.T, recorder *test.RequestRecorder) (*codex.Codex, context.Context, *test.LockedBuffer, *atomic.Int64) {
	t.Helper()
	secrets := []string{os.Getenv("OPENAI_API_KEY"), os.Getenv("CODEX_E2E_LOGIN_PARAMS_JSON")}
	t.Setenv("OPENAI_API_KEY", "")
	t.Setenv("CODEX_E2E_LOGIN_PARAMS_JSON", "")
	var requests atomic.Int64
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodPost || r.URL.Path != "/v1/responses" {
			t.Errorf("unexpected model request: %s %s", r.Method, r.URL.Path)
			http.NotFound(w, r)
			return
		}
		if r.Header.Get("Authorization") != "" {
			t.Error("local provider received an authorization header")
			http.Error(w, "unexpected authorization", http.StatusBadRequest)
			return
		}
		var request struct {
			Model string `json:"model"`
		}
		if err := json.NewDecoder(r.Body).Decode(&request); err != nil || request.Model != "mock-model" {
			t.Errorf("invalid local model request: model=%q, err=%v", request.Model, err)
			http.Error(w, "invalid request", http.StatusBadRequest)
			return
		}
		requests.Add(1)
		w.Header().Set("Content-Type", "text/event-stream")
		// Match the deterministic Responses fixture used by upstream rust-v0.153.4.
		_, err := fmt.Fprint(w, "data: "+`{"type":"response.created","response":{"id":"resp-sdk-e2e"}}`+"\n\n"+
			"data: "+`{"type":"response.output_item.done","item":{"type":"message","role":"assistant","id":"msg-sdk-e2e","content":[{"type":"output_text","text":"Done"}]}}`+"\n\n"+
			"data: "+`{"type":"response.completed","response":{"id":"resp-sdk-e2e","usage":{"input_tokens":0,"input_tokens_details":null,"output_tokens":0,"output_tokens_details":null,"total_tokens":0}}}`+"\n\n")
		if err != nil {
			t.Errorf("write local model response: %v", err)
		}
	}))
	t.Cleanup(server.Close)
	client, ctx, stderr := test.NewRealClient(t, test.RealClientOptions{
		Secrets:         secrets,
		RequestRecorder: recorder,
		ConfigOverrides: []string{
			`model="mock-model"`, `model_provider="mock_provider"`,
			`approval_policy="never"`, `sandbox_mode="read-only"`,
			fmt.Sprintf(`model_providers.mock_provider={name="Local lifecycle test",base_url=%q,wire_api="responses",requires_openai_auth=false,supports_websockets=false,request_max_retries=0,stream_max_retries=0}`, server.URL+"/v1"),
		},
	})
	return client, ctx, stderr, &requests
}

func TestRealCodexUnmaterializedArchiveRejected(t *testing.T) {
	client, ctx, stderr, requests := newArchiveTestClient(t, nil)
	for _, viaThread := range []bool{false, true} {
		thread := test.StartThread(t, client, ctx, stderr, t.TempDir())
		var err error
		if viaThread {
			_, err = thread.Archive(ctx)
		} else {
			_, err = client.ArchiveThread(ctx, thread.ID())
		}
		var responseErr *rpc.ResponseError
		if !errors.As(err, &responseErr) || responseErr.Detail.Code != -32600 || !test.IsExpectedUnmaterializedThreadError(err) {
			t.Fatalf("empty thread archive (thread helper=%v): %v\nstderr:\n%s", viaThread, err, stderr.String())
		}
		// A rejected archive is not followed by unarchive: the thread still owns its writer.
	}
	if requests.Load() != 0 {
		t.Fatalf("empty-thread test unexpectedly sent %d model requests", requests.Load())
	}
}

func TestRealCodexThreadArchiveRoundTripWithMockProvider(t *testing.T) {
	for _, viaThread := range []bool{false, true} {
		name := "client"
		if viaThread {
			name = "thread"
		}
		t.Run(name, func(t *testing.T) {
			// Exercise provider overrides with both the default spawn and recording transport.
			var recorder *test.RequestRecorder
			if viaThread {
				recorder = &test.RequestRecorder{}
			}
			client, ctx, stderr, requests := newArchiveTestClient(t, recorder)
			cwd := t.TempDir()
			thread := test.StartThread(t, client, ctx, stderr, cwd)
			result, err := thread.Run(ctx, "Materialize this test thread.", nil)
			if err != nil {
				t.Fatalf("materialize thread: %v\nstderr:\n%s", err, stderr.String())
			}
			test.AssertCompletedTurnResult(t, "local turn", result)
			if result.FinalResponse != "Done" || requests.Load() != 1 {
				t.Fatalf("local response=%q, requests=%d", result.FinalResponse, requests.Load())
			}
			if viaThread {
				_, err = thread.Archive(ctx)
			} else {
				_, err = client.ArchiveThread(ctx, thread.ID())
			}
			if err != nil {
				t.Fatalf("archive persisted thread: %v\nstderr:\n%s", err, stderr.String())
			}
			assertArchiveListing(t, ctx, client, thread.ID(), cwd, true)
			var restored *protocol.ThreadUnarchiveResponse
			if viaThread {
				restored, err = thread.Unarchive(ctx)
			} else {
				restored, err = client.UnarchiveThread(ctx, thread.ID())
			}
			if err != nil {
				t.Fatalf("unarchive persisted thread: %v\nstderr:\n%s", err, stderr.String())
			}
			if restored.Thread.ID != thread.ID() {
				t.Fatalf("restored thread %q, want %q", restored.Thread.ID, thread.ID())
			}
			assertArchiveListing(t, ctx, client, thread.ID(), cwd, false)
			read, err := client.ReadThread(ctx, thread.ID(), codex.ThreadReadOptions{IncludeTurns: true})
			if err != nil || len(read.Thread.Turns) != 1 || read.Thread.Turns[0].ID != result.TurnID {
				t.Fatalf("restored history=%#v, err=%v", read, err)
			}
		})
	}
}

func assertArchiveListing(t *testing.T, ctx context.Context, client *codex.Codex, id, cwd string, archived bool) {
	t.Helper()
	for _, filter := range []bool{archived, !archived} {
		response, err := client.ListThreads(ctx, codex.ThreadListOptions{Archived: &filter, Cwd: cwd})
		if err != nil {
			t.Fatalf("list archived=%v: %v", filter, err)
		}
		found := false
		for _, thread := range response.Data {
			if thread.ID == id {
				found = true
			}
		}
		if found != (filter == archived) {
			t.Fatalf("thread %s found=%v in archived=%v, want archived=%v", id, found, filter, archived)
		}
	}
}
