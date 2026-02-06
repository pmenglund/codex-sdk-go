package codex

import (
	"context"
	"encoding/json"
	"io"
	"log/slog"
	"os"
	"path/filepath"
	"reflect"
	"runtime"
	"testing"
	"time"

	"github.com/pmenglund/codex-sdk-go/protocol"
	"github.com/pmenglund/codex-sdk-go/rpc"
)

func TestThreadStartOptionsToParams(t *testing.T) {
	opts := ThreadStartOptions{
		Model:                 "gpt-test",
		Cwd:                   "/tmp/project",
		ApprovalPolicy:        "never",
		SandboxPolicy:         map[string]any{"type": "readOnly"},
		Config:                map[string]any{"foo": "bar"},
		BaseInstructions:      "base",
		DeveloperInstructions: "dev",
		ExperimentalRawEvents: true,
	}

	params, err := opts.toParams()
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	assertEqual(t, "model", params.Model, stringPtr("gpt-test"))
	assertEqual(t, "cwd", params.Cwd, stringPtr("/tmp/project"))
	assertRawEqual(t, "approvalPolicy", params.ApprovalPolicy, MustJSON("never"))
	assertRawEqual(t, "sandbox", params.Sandbox, MustJSON(map[string]any{"type": "readOnly"}))
	if params.Config == nil {
		t.Fatalf("expected config")
	}
	assertEqual(t, "config", *params.Config, map[string]any{"foo": "bar"})
	assertEqual(t, "baseInstructions", params.BaseInstructions, stringPtr("base"))
	assertEqual(t, "developerInstructions", params.DeveloperInstructions, stringPtr("dev"))
	assertEqual(t, "experimentalRawEvents", params.ExperimentalRawEvents, true)
}

func TestThreadResumeOptionsToParams(t *testing.T) {
	opts := ThreadResumeOptions{
		ThreadID:              "thr_123",
		History:               []protocol.ThreadResumeParamsHistoryElem{"h1"},
		Path:                  "/tmp/rollout",
		Model:                 "gpt-test",
		ModelProvider:         "openai",
		Cwd:                   "/tmp/project",
		ApprovalPolicy:        "never",
		Sandbox:               map[string]any{"type": "readOnly"},
		Config:                map[string]any{"foo": "bar"},
		BaseInstructions:      "base",
		DeveloperInstructions: "dev",
	}

	params, err := opts.toParams()
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	assertEqual(t, "threadId", params.ThreadID, "thr_123")
	assertEqual(t, "history", params.History, []protocol.ThreadResumeParamsHistoryElem{MustJSON("h1")})
	assertEqual(t, "path", params.Path, stringPtr("/tmp/rollout"))
	assertEqual(t, "model", params.Model, stringPtr("gpt-test"))
	assertEqual(t, "modelProvider", params.ModelProvider, stringPtr("openai"))
	assertEqual(t, "cwd", params.Cwd, stringPtr("/tmp/project"))
	assertRawEqual(t, "approvalPolicy", params.ApprovalPolicy, MustJSON("never"))
	assertRawEqual(t, "sandbox", params.Sandbox, MustJSON(map[string]any{"type": "readOnly"}))
	if params.Config == nil {
		t.Fatalf("expected config")
	}
	assertEqual(t, "config", *params.Config, map[string]any{"foo": "bar"})
	assertEqual(t, "baseInstructions", params.BaseInstructions, stringPtr("base"))
	assertEqual(t, "developerInstructions", params.DeveloperInstructions, stringPtr("dev"))
}

func TestBuildTurnParams(t *testing.T) {
	opts := &TurnOptions{
		Cwd:               "/tmp",
		ApprovalPolicy:    "never",
		SandboxPolicy:     map[string]any{"type": "readOnly"},
		Model:             "gpt-test",
		Effort:            "medium",
		Summary:           "short",
		OutputSchema:      map[string]any{"type": "object"},
		CollaborationMode: "default",
	}

	params, err := buildTurnParams("thr_123", []Input{TextInput("hello")}, opts)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	assertEqual(t, "threadId", params.ThreadID, "thr_123")
	assertEqual(t, "input", params.Input, []protocol.TurnStartParamsInputElem{TextInput("hello")})
	assertEqual(t, "cwd", params.Cwd, stringPtr("/tmp"))
	assertRawEqual(t, "approvalPolicy", params.ApprovalPolicy, MustJSON("never"))
	assertRawEqual(t, "sandboxPolicy", params.SandboxPolicy, MustJSON(map[string]any{"type": "readOnly"}))
	assertEqual(t, "model", params.Model, stringPtr("gpt-test"))
	assertRawEqual(t, "effort", params.Effort, MustJSON("medium"))
	assertRawEqual(t, "summary", params.Summary, MustJSON("short"))
	assertRawEqual(t, "outputSchema", params.OutputSchema, MustJSON(map[string]any{"type": "object"}))
	assertRawEqual(t, "collaborationMode", params.CollaborationMode, MustJSON("default"))
}

func TestThreadResponseID(t *testing.T) {
	response := protocol.ThreadStartResponse{Thread: &protocol.Thread{ID: "thr_1"}}
	id, err := threadIDFromResponse(response.ThreadID, response.Thread)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if id != "thr_1" {
		t.Fatalf("expected thread id thr_1, got %q", id)
	}

	if _, err := threadIDFromResponse("", nil); err == nil {
		t.Fatalf("expected error for missing thread id")
	}
}

func TestJSONHelpers(t *testing.T) {
	if raw, err := JSON(nil); err != nil || raw != nil {
		t.Fatalf("expected nil JSON, got %v err=%v", raw, err)
	}

	raw, err := JSON(json.RawMessage(`{"ok":true}`))
	if err != nil || string(raw) != `{"ok":true}` {
		t.Fatalf("unexpected raw JSON: %s err=%v", string(raw), err)
	}

	if _, err := JSON(json.RawMessage("{bad")); err == nil {
		t.Fatalf("expected error for invalid raw JSON")
	}

	if _, err := normalizeJSONValue("value", json.RawMessage("{bad")); err == nil {
		t.Fatalf("expected normalize error for invalid raw JSON")
	}

	if raw := MustJSON(map[string]any{"ok": true}); !json.Valid(raw) {
		t.Fatalf("expected valid JSON")
	}
}

func TestStartThreadInvalidJSONOptions(t *testing.T) {
	ctx := context.Background()
	c := &Codex{logger: slog.New(slog.NewTextHandler(io.Discard, nil))}
	if _, err := c.StartThread(ctx, ThreadStartOptions{ApprovalPolicy: json.RawMessage("{bad")}); err == nil {
		t.Fatalf("expected error for invalid approval policy")
	}
}

func TestResumeThreadInvalidJSONOptions(t *testing.T) {
	ctx := context.Background()
	c := &Codex{logger: slog.New(slog.NewTextHandler(io.Discard, nil))}
	if _, err := c.ResumeThread(ctx, ThreadResumeOptions{ApprovalPolicy: json.RawMessage("{bad")}); err == nil {
		t.Fatalf("expected error for invalid approval policy")
	}
}

func TestRunStreamedInvalidJSONOptions(t *testing.T) {
	ctx := context.Background()
	client, err := New(ctx, Options{Transport: rpc.NewReplayTransport(initializeTranscript())})
	if err != nil {
		t.Fatalf("new client error: %v", err)
	}
	defer client.Close()

	thread := &Thread{client: client.Client(), id: "thr_123", logger: slog.New(slog.NewTextHandler(io.Discard, nil))}
	if _, err := thread.RunStreamed(ctx, []Input{TextInput("hi")}, &TurnOptions{ApprovalPolicy: json.RawMessage("{bad")}); err == nil {
		t.Fatalf("expected error for invalid approval policy")
	}
}

func TestExtractTextFromItemRaw(t *testing.T) {
	raw := MustJSON(map[string]any{"text": "hello"})
	if text, ok := extractTextFromItemRaw(raw); !ok || text != "hello" {
		t.Fatalf("expected text from raw")
	}

	raw = MustJSON(map[string]any{"wrapped": map[string]any{"text": "inner"}})
	if text, ok := extractTextFromItemRaw(raw); !ok || text != "inner" {
		t.Fatalf("expected text from nested raw")
	}
}

func TestNotificationError(t *testing.T) {
	note := rpc.Notification{Method: "error", Raw: MustJSON(map[string]any{"willRetry": true})}
	if err := notificationError(note); err != nil {
		t.Fatalf("expected nil error for willRetry")
	}

	note = rpc.Notification{Method: "error", Raw: MustJSON(map[string]any{"error": map[string]any{"message": "boom"}})}
	if err := notificationError(note); err == nil || err.Error() != "boom" {
		t.Fatalf("expected error boom, got %v", err)
	}

	note = rpc.Notification{Method: "turn/completed", Raw: MustJSON(map[string]any{"turn": map[string]any{"status": "failed", "error": map[string]any{"message": "fail"}}})}
	if err := notificationError(note); err == nil || err.Error() != "fail" {
		t.Fatalf("expected error fail, got %v", err)
	}
}

func TestResolveLogger(t *testing.T) {
	logger := resolveLogger(nil)
	if logger == nil {
		t.Fatalf("expected non-nil logger")
	}
	logger.Info("silenced")
}

func TestAttachApprovalLogger(t *testing.T) {
	logger := slog.New(slog.NewTextHandler(io.Discard, nil))
	handler := AutoApproveHandler{}
	attached := attachApprovalLogger(handler, logger)
	typed, ok := attached.(AutoApproveHandler)
	if !ok {
		t.Fatalf("expected AutoApproveHandler")
	}
	if typed.Logger == nil {
		t.Fatalf("expected logger to be attached")
	}
}

func TestAutoApproveResponses(t *testing.T) {
	handler := AutoApproveHandler{}
	resp, err := handler.ItemCommandExecutionRequestApproval(context.Background(), protocol.CommandExecutionRequestApprovalParams{ItemID: "item", ThreadID: "thr", TurnID: "turn"})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if resp == nil {
		t.Fatalf("expected response")
	}
}

func TestAutoApproveLegacyResponses(t *testing.T) {
	handler := AutoApproveHandler{}
	if _, err := handler.ItemFileChangeRequestApproval(context.Background(), protocol.FileChangeRequestApprovalParams{ItemID: "item", ThreadID: "thr", TurnID: "turn"}); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if _, err := handler.ApplyPatchApproval(context.Background(), protocol.ApplyPatchApprovalParams{CallID: "call", ConversationID: "thr", FileChanges: map[string]any{}}); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if _, err := handler.ExecCommandApproval(context.Background(), protocol.ExecCommandApprovalParams{CallID: "call", ConversationID: "thr", Command: []string{"echo"}}); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if _, err := handler.ItemToolRequestUserInput(context.Background(), protocol.ToolRequestUserInputParams{ItemID: "item", ThreadID: "thr", TurnID: "turn"}); err == nil {
		t.Fatalf("expected error for tool user input")
	}
}

func TestNewUsesDefaultClientInfo(t *testing.T) {
	ctx := context.Background()
	client, err := New(ctx, Options{
		Transport: rpc.NewReplayTransport(initializeTranscript()),
	})
	if err != nil {
		t.Fatalf("new client error: %v", err)
	}
	if client.Client() == nil {
		t.Fatalf("expected rpc client")
	}
	_ = client.Close()
}

func TestNewSpawnError(t *testing.T) {
	ctx := context.Background()
	_, err := New(ctx, Options{
		Spawn: SpawnOptions{CodexPath: "codex-missing-binary"},
	})
	if err == nil {
		t.Fatalf("expected spawn error")
	}
}

func TestNewSpawnSurvivesInitContextCancellation(t *testing.T) {
	if runtime.GOOS == "windows" {
		t.Skip("spawn script test is unix-only")
	}

	ctx, cancel := context.WithCancel(context.Background())
	client, err := New(ctx, Options{
		Spawn:  SpawnOptions{CodexPath: writeFakeCodexBinary(t)},
		Logger: slog.New(slog.NewTextHandler(io.Discard, nil)),
	})
	if err != nil {
		t.Fatalf("new client error: %v", err)
	}
	defer client.Close()

	cancel()
	time.Sleep(100 * time.Millisecond)

	thread, err := client.StartThread(context.Background(), ThreadStartOptions{})
	if err != nil {
		t.Fatalf("start thread after init context cancel failed: %v", err)
	}
	if thread.ID() != "thr_test" {
		t.Fatalf("unexpected thread id: %s", thread.ID())
	}
}

func initializeTranscript() []rpc.TranscriptEntry {
	info := defaultClientInfo()
	return []rpc.TranscriptEntry{
		writeLine(rpc.JSONRPCRequest{
			ID:     rpc.NewIntRequestID(1),
			Method: "initialize",
			Params: mustRaw(protocol.InitializeParams{ClientInfo: info}),
		}),
		readLine(rpc.JSONRPCResponse{
			ID:     rpc.NewIntRequestID(1),
			Result: mustRaw(map[string]any{}),
		}),
		writeLine(rpc.JSONRPCNotification{Method: "initialized"}),
	}
}

func writeFakeCodexBinary(t *testing.T) string {
	t.Helper()

	path := filepath.Join(t.TempDir(), "fake-codex")
	script := `#!/bin/sh
extract_id() {
	printf '%s\n' "$1" | sed -n 's/.*"id":\([0-9][0-9]*\).*/\1/p'
}

while IFS= read -r line; do
	case "$line" in
		*'"method":"initialize"'*)
			id=$(extract_id "$line")
			if [ -z "$id" ]; then id=1; fi
			printf '{"jsonrpc":"2.0","id":%s,"result":{}}\n' "$id"
			;;
		*'"method":"thread/start"'*)
			id=$(extract_id "$line")
			if [ -z "$id" ]; then id=2; fi
			printf '{"jsonrpc":"2.0","id":%s,"result":{"threadId":"thr_test"}}\n' "$id"
			;;
	esac
done
`
	if err := os.WriteFile(path, []byte(script), 0o755); err != nil {
		t.Fatalf("write fake codex: %v", err)
	}

	return path
}

func TestInputHelpers(t *testing.T) {
	if input := TextInput("hi"); input.Type != InputTypeText || input.Text != "hi" {
		t.Fatalf("unexpected text input: %#v", input)
	}
	if input := ImageInput("https://example.com"); input.Type != InputTypeImage || input.URL != "https://example.com" {
		t.Fatalf("unexpected image input: %#v", input)
	}
	if input := LocalImageInput("/tmp/img.png"); input.Type != InputTypeLocalImage || input.Path != "/tmp/img.png" {
		t.Fatalf("unexpected local image input: %#v", input)
	}
	if input := SkillInput("skill", "/tmp/skill"); input.Type != InputTypeSkill || input.Name != "skill" || input.Path != "/tmp/skill" {
		t.Fatalf("unexpected skill input: %#v", input)
	}
}

func TestMatchThreadID(t *testing.T) {
	note := rpc.Notification{Raw: MustJSON(map[string]any{"threadId": "thr_1"})}
	if !matchesThreadID(note, "thr_1") {
		t.Fatalf("expected matching thread id")
	}
	if matchesThreadID(note, "thr_2") {
		t.Fatalf("expected non-matching thread id")
	}

	empty := rpc.Notification{Raw: MustJSON(map[string]any{})}
	if !matchesThreadID(empty, "thr_1") {
		t.Fatalf("expected match when thread id missing")
	}
}

func assertEqual(t *testing.T, name string, got, want any) {
	t.Helper()
	if !reflect.DeepEqual(got, want) {
		t.Fatalf("unexpected %s: %#v (want %#v)", name, got, want)
	}
}

func assertRawEqual(t *testing.T, name string, got any, want json.RawMessage) {
	t.Helper()
	if got == nil {
		t.Fatalf("expected %s to be set", name)
	}
	raw, ok := got.(json.RawMessage)
	if !ok {
		t.Fatalf("expected %s to be json.RawMessage, got %T", name, got)
	}
	if string(raw) != string(want) {
		t.Fatalf("unexpected %s: %s (want %s)", name, string(raw), string(want))
	}
}
