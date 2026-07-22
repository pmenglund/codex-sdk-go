package codex

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"io"
	"log/slog"
	"os"
	"path/filepath"
	"reflect"
	"runtime"
	"strings"
	"testing"
	"time"

	"github.com/pmenglund/codex-sdk-go/protocol"
	"github.com/pmenglund/codex-sdk-go/rpc"
)

func TestThreadStartOptionsToParams(t *testing.T) {
	ephemeral := true
	opts := ThreadStartOptions{
		Model:                 "gpt-test",
		ModelProvider:         "openai",
		ServiceTier:           "priority",
		Cwd:                   "/tmp/project",
		ApprovalPolicy:        "never",
		SandboxPolicy:         map[string]any{"type": "readOnly"},
		Config:                map[string]any{"foo": "bar"},
		ServiceName:           "codex-sdk-go",
		BaseInstructions:      "base",
		DeveloperInstructions: "dev",
		Ephemeral:             &ephemeral,
	}

	params, err := opts.toParams()
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	assertEqual(t, "model", params.Model, stringPtr("gpt-test"))
	assertEqual(t, "modelProvider", params.ModelProvider, stringPtr("openai"))
	assertEqual(t, "serviceTier", params.ServiceTier, stringPtr("priority"))
	assertEqual(t, "cwd", params.Cwd, stringPtr("/tmp/project"))
	assertRawEqual(t, "approvalPolicy", params.ApprovalPolicy, MustJSON("never"))
	assertRawEqual(t, "sandbox", params.Sandbox, MustJSON(map[string]any{"type": "readOnly"}))
	if params.Config == nil {
		t.Fatalf("expected config")
	}
	assertEqual(t, "config", *params.Config, map[string]any{"foo": "bar"})
	assertEqual(t, "serviceName", params.ServiceName, stringPtr("codex-sdk-go"))
	assertEqual(t, "baseInstructions", params.BaseInstructions, stringPtr("base"))
	assertEqual(t, "developerInstructions", params.DeveloperInstructions, stringPtr("dev"))
	assertEqual(t, "ephemeral", params.Ephemeral, &ephemeral)
}

func TestNewRejectsTypedNilTransport(t *testing.T) {
	var transport *rpc.ReplayTransport
	if _, err := New(context.Background(), Options{Transport: transport}); err == nil {
		t.Fatal("expected typed-nil transport error")
	}
}

func TestThreadStartOptionsRejectExperimentalRawEvents(t *testing.T) {
	_, err := (ThreadStartOptions{ExperimentalRawEvents: true}).toParams()
	if err == nil {
		t.Fatalf("expected experimental raw events error")
	}
}

func TestThreadResumeOptionsToParams(t *testing.T) {
	opts := ThreadResumeOptions{
		ThreadID:              "thr_123",
		Model:                 "gpt-test",
		ModelProvider:         "openai",
		ServiceTier:           "priority",
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
	assertEqual(t, "model", params.Model, stringPtr("gpt-test"))
	assertEqual(t, "modelProvider", params.ModelProvider, stringPtr("openai"))
	assertEqual(t, "serviceTier", params.ServiceTier, stringPtr("priority"))
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

func TestThreadResumeOptionsRejectEmptyThreadID(t *testing.T) {
	_, err := (ThreadResumeOptions{}).toParams()
	if err == nil {
		t.Fatalf("expected empty thread id error")
	}
}

func TestThreadResumeOptionsRejectHistory(t *testing.T) {
	_, err := (ThreadResumeOptions{
		ThreadID: "thr_123",
		History:  []ThreadResumeHistoryElem{MustJSON("h1")},
	}).toParams()
	if err == nil {
		t.Fatalf("expected history resume error")
	}
}

func TestThreadResumeOptionsRejectPath(t *testing.T) {
	_, err := (ThreadResumeOptions{
		ThreadID: "thr_123",
		Path:     "/tmp/rollout",
	}).toParams()
	if err == nil {
		t.Fatalf("expected path resume error")
	}
}

func TestBuildTurnParams(t *testing.T) {
	opts := &TurnOptions{
		ClientUserMessageID: "msg_123",
		Cwd:                 "/tmp",
		ApprovalPolicy:      "never",
		SandboxPolicy:       map[string]any{"type": "readOnly"},
		Model:               "gpt-test",
		ServiceTier:         "priority",
		Effort:              "medium",
		Summary:             "short",
		OutputSchema:        map[string]any{"type": "object"},
	}

	params, err := buildTurnParams("thr_123", []Input{TextInput("hello")}, opts)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	assertEqual(t, "threadId", params.ThreadID, "thr_123")
	assertEqual(t, "input", params.Input, []protocol.TurnStartParamsInputElem{TextInput("hello")})
	assertEqual(t, "clientUserMessageId", params.ClientUserMessageID, stringPtr("msg_123"))
	assertEqual(t, "cwd", params.Cwd, stringPtr("/tmp"))
	assertRawEqual(t, "approvalPolicy", params.ApprovalPolicy, MustJSON("never"))
	assertRawEqual(t, "sandboxPolicy", params.SandboxPolicy, MustJSON(map[string]any{"type": "readOnly"}))
	assertEqual(t, "model", params.Model, stringPtr("gpt-test"))
	assertEqual(t, "serviceTier", params.ServiceTier, stringPtr("priority"))
	assertRawEqual(t, "effort", params.Effort, MustJSON("medium"))
	assertRawEqual(t, "summary", params.Summary, MustJSON("short"))
	assertRawEqual(t, "outputSchema", params.OutputSchema, MustJSON(map[string]any{"type": "object"}))
}

func TestBuildTurnParamsRejectCollaborationMode(t *testing.T) {
	_, err := buildTurnParams("thr_123", []Input{TextInput("hello")}, &TurnOptions{CollaborationMode: "default"})
	if err == nil {
		t.Fatalf("expected collaboration mode error")
	}
}

func TestBuildTurnParamsRejectInvalidInputs(t *testing.T) {
	tests := []struct {
		name  string
		input Input
	}{
		{name: "unknown type", input: Input{Type: "bogus"}},
		{name: "empty text", input: Input{Type: InputTypeText}},
		{name: "empty image url", input: ImageInput("")},
		{name: "empty local image path", input: LocalImageInput("")},
		{name: "empty skill name", input: SkillInput("", "/tmp/skill")},
		{name: "empty skill path", input: SkillInput("skill", "")},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if _, err := buildTurnParams("thr_123", []Input{tt.input}, nil); err == nil {
				t.Fatalf("expected invalid input error")
			}
		})
	}
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
	if _, err := c.ResumeThread(ctx, ThreadResumeOptions{ThreadID: "thr_123", ApprovalPolicy: json.RawMessage("{bad")}); err == nil {
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
	raw := MustJSON(map[string]any{"type": "agentMessage", "text": "hello"})
	if text, ok := extractTextFromItemRaw(raw); !ok || text != "hello" {
		t.Fatalf("expected text from raw")
	}

	raw = MustJSON(map[string]any{"wrapped": map[string]any{"type": "agentMessage", "text": "inner"}})
	if text, ok := extractTextFromItemRaw(raw); !ok || text != "inner" {
		t.Fatalf("expected text from nested raw")
	}

	if _, ok := extractTextFromItemRaw(MustJSON(map[string]any{"type": "plan", "text": "not final"})); ok {
		t.Fatalf("plan text must not be treated as a final response")
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

	note = rpc.Notification{Method: "turn/failed", Raw: MustJSON(map[string]any{"threadId": "thr_123", "turn": map[string]any{"status": "failed", "error": map[string]any{"message": "failed hard"}}})}
	if err := notificationError(note); err == nil || err.Error() != "failed hard" {
		t.Fatalf("expected error failed hard, got %v", err)
	}

	note = rpc.Notification{Method: "error", Raw: json.RawMessage("{bad")}
	if err := notificationError(note); err == nil || err.Error() != "turn error" {
		t.Fatalf("expected generic turn error for malformed payload, got %v", err)
	}

	note = rpc.Notification{Method: "turn/completed", Raw: MustJSON(map[string]any{"turn": map[string]any{"status": "failed"}})}
	if err := notificationError(note); err == nil || err.Error() != "turn failed" {
		t.Fatalf("expected generic completed failure, got %v", err)
	}

	note = rpc.Notification{Method: "turn/failed", Raw: MustJSON(map[string]any{"error": map[string]any{"message": "payload boom"}})}
	if err := notificationError(note); err == nil || err.Error() != "payload boom" {
		t.Fatalf("expected payload error message, got %v", err)
	}

	note = rpc.Notification{Method: "turn/failed", Raw: json.RawMessage("{bad")}
	if err := notificationError(note); err == nil || err.Error() != "turn failed" {
		t.Fatalf("expected generic turn failed for malformed payload, got %v", err)
	}
}

func TestParseTurnNotificationTypedParams(t *testing.T) {
	willRetry := true
	itemRaw := MustJSON(map[string]any{"type": "agentMessage", "id": "item_1", "text": "done"})
	var item protocol.ThreadItem
	if err := json.Unmarshal(itemRaw, &item); err != nil {
		t.Fatalf("decode thread item: %v", err)
	}
	tests := []struct {
		name string
		note rpc.Notification
		want turnNotificationPayload
	}{
		{
			name: "turn value",
			note: rpc.Notification{Params: protocol.TurnNotification{
				ThreadID: "thr_1",
				Turn:     &protocol.Turn{ID: "turn_1"},
			}},
			want: turnNotificationPayload{ThreadID: "thr_1", Turn: &protocol.Turn{ID: "turn_1"}},
		},
		{
			name: "turn pointer",
			note: rpc.Notification{Params: &protocol.TurnNotification{
				ThreadID: "thr_2",
				Turn:     &protocol.Turn{ID: "turn_2"},
			}},
			want: turnNotificationPayload{ThreadID: "thr_2", Turn: &protocol.Turn{ID: "turn_2"}},
		},
		{
			name: "item value",
			note: rpc.Notification{Params: protocol.ItemCompletedNotification{
				ThreadID: "thr_3",
				Item:     item,
			}},
			want: turnNotificationPayload{ThreadID: "thr_3", Item: itemRaw},
		},
		{
			name: "item pointer",
			note: rpc.Notification{Params: &protocol.ItemCompletedNotification{
				ThreadID: "thr_4",
				Item:     item,
			}},
			want: turnNotificationPayload{ThreadID: "thr_4", Item: itemRaw},
		},
		{
			name: "error value",
			note: rpc.Notification{Params: protocol.ErrorNotification{
				ThreadID:  "thr_5",
				WillRetry: willRetry,
				Error:     protocol.TurnError{Message: "retrying"},
			}},
			want: turnNotificationPayload{
				ThreadID:  "thr_5",
				WillRetry: &willRetry,
				Error:     &protocol.TurnError{Message: "retrying"},
			},
		},
		{
			name: "error pointer",
			note: rpc.Notification{Params: &protocol.ErrorNotification{
				ThreadID:  "thr_6",
				WillRetry: willRetry,
				Error:     protocol.TurnError{Message: "retrying"},
			}},
			want: turnNotificationPayload{
				ThreadID:  "thr_6",
				WillRetry: &willRetry,
				Error:     &protocol.TurnError{Message: "retrying"},
			},
		},
		{
			name: "thread goal cleared value",
			note: rpc.Notification{Params: protocol.ThreadGoalClearedNotification{
				ThreadID: "thr_7",
			}},
			want: turnNotificationPayload{ThreadID: "thr_7"},
		},
		{
			name: "thread goal cleared pointer",
			note: rpc.Notification{Params: &protocol.ThreadGoalClearedNotification{
				ThreadID: "thr_8",
			}},
			want: turnNotificationPayload{ThreadID: "thr_8"},
		},
		{
			name: "thread goal updated value",
			note: rpc.Notification{Params: protocol.ThreadGoalUpdatedNotification{
				ThreadID: "thr_9",
				Goal: protocol.ThreadGoal{
					ThreadID:  "thr_9",
					Objective: "ship",
					Status:    protocol.ThreadGoalStatusActive,
				},
			}},
			want: turnNotificationPayload{ThreadID: "thr_9"},
		},
		{
			name: "thread goal updated pointer",
			note: rpc.Notification{Params: &protocol.ThreadGoalUpdatedNotification{
				ThreadID: "thr_10",
				Goal: protocol.ThreadGoal{
					ThreadID:  "thr_10",
					Objective: "ship",
					Status:    protocol.ThreadGoalStatusActive,
				},
			}},
			want: turnNotificationPayload{ThreadID: "thr_10"},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := parseTurnNotification(tt.note)
			if err != nil {
				t.Fatalf("parseTurnNotification error: %v", err)
			}
			assertEqual(t, "payload", got, tt.want)
		})
	}
}

func TestTurnStreamNilAndClose(t *testing.T) {
	var stream *TurnStream
	if _, err := stream.Next(context.Background()); err == nil {
		t.Fatalf("expected nil stream error")
	}
	stream.Close()

	stream = &TurnStream{}
	if _, err := stream.Next(context.Background()); err == nil {
		t.Fatalf("expected uninitialized stream error")
	}
	stream.Close()
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

	ptr := &AutoApproveHandler{}
	attached = attachApprovalLogger(ptr, logger)
	if attached == ptr {
		t.Fatalf("expected pointer handler to be copied")
	}
	if ptr.Logger != nil {
		t.Fatalf("caller-owned pointer was mutated")
	}
	if copied, ok := attached.(*AutoApproveHandler); !ok || copied.Logger == nil {
		t.Fatalf("expected copied pointer logger to be attached: %#v", attached)
	}

	customLogger := slog.New(slog.NewTextHandler(io.Discard, nil))
	ptr = &AutoApproveHandler{Logger: customLogger}
	attached = attachApprovalLogger(ptr, logger)
	if attached == ptr || ptr.Logger != customLogger {
		t.Fatalf("expected existing pointer logger to be copied without mutation")
	}

	gotNil := attachApprovalLogger((*AutoApproveHandler)(nil), logger)
	if gotNil != nil {
		t.Fatalf("expected typed nil handler to normalize to nil, got %#v", gotNil)
	}

	var typedNilCustom *testServerRequestHandler
	if got := attachApprovalLogger(typedNilCustom, logger); got != nil {
		t.Fatalf("expected custom typed nil handler to normalize to nil, got %#v", got)
	}

	other := &testServerRequestHandler{}
	if got := attachApprovalLogger(other, logger); got != other {
		t.Fatalf("expected non-auto-approve handler to pass through")
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

	permissions := map[string]any{"sandbox": "workspace-write"}
	permResp, err := handler.ItemPermissionsRequestApproval(context.Background(), protocol.PermissionsRequestApprovalParams{
		ItemID:      "item",
		ThreadID:    "thr",
		TurnID:      "turn",
		Permissions: permissions,
	})
	if err != nil {
		t.Fatalf("unexpected permissions error: %v", err)
	}
	assertEqual(t, "permissions", permResp.Permissions, permissions)
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
	if _, err := handler.ItemToolCall(context.Background(), protocol.DynamicToolCallParams{}); err == nil {
		t.Fatalf("expected error for dynamic tool call")
	}
	if _, err := handler.McpServerElicitationRequest(context.Background(), protocol.MCPServerElicitationRequestParams(nil)); err == nil {
		t.Fatalf("expected error for mcp elicitation")
	}
	if _, err := handler.AccountChatgptAuthTokensRefresh(context.Background(), protocol.ChatgptAuthTokensRefreshParams{}); err == nil {
		t.Fatalf("expected error for auth token refresh")
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
	// Compatibility probing is covered separately; this test isolates whether
	// canceling the initialization context terminates the spawned app-server.
	client, err := New(ctx, Options{
		Spawn:               SpawnOptions{CodexPath: writeFakeCodexBinary(t)},
		Logger:              slog.New(slog.NewTextHandler(io.Discard, nil)),
		CompatibilityPolicy: Ignore,
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

func TestStartThreadOnUninitializedClient(t *testing.T) {
	_, err := (&Codex{}).StartThread(context.Background(), ThreadStartOptions{})
	if err == nil || err.Error() != "codex client is not initialized" {
		t.Fatalf("unexpected error: %v", err)
	}
}

func TestResumeThreadOnUninitializedClient(t *testing.T) {
	_, err := (&Codex{}).ResumeThread(context.Background(), ThreadResumeOptions{ThreadID: "thr_123"})
	if err == nil || err.Error() != "codex client is not initialized" {
		t.Fatalf("unexpected error: %v", err)
	}
}

func TestThreadRunOnUninitializedThread(t *testing.T) {
	_, err := (&Thread{}).Run(context.Background(), "hi", nil)
	if err == nil || err.Error() != "thread client is not initialized" {
		t.Fatalf("unexpected error: %v", err)
	}
}

func TestNilThreadRun(t *testing.T) {
	var thread *Thread
	_, err := thread.Run(context.Background(), "hi", nil)
	if err == nil || err.Error() != "thread is nil" {
		t.Fatalf("unexpected error: %v", err)
	}
}

func TestRunStreamedStartCallError(t *testing.T) {
	client := rpc.NewClient(rpc.NewReplayTransport([]rpc.TranscriptEntry{
		writeLine(rpc.JSONRPCRequest{
			ID:     rpc.NewIntRequestID(1),
			Method: "turn/start",
			Params: mustRaw(turnStartParams("hello")),
		}),
		readLine(rpc.JSONRPCError{
			ID: rpc.NewIntRequestID(1),
			Error: rpc.JSONRPCErrorDetail{
				Code:    -1,
				Message: "start failed",
			},
		}),
	}), rpc.ClientOptions{})
	defer client.Close()

	thread := &Thread{client: client, id: "thr_123", logger: slog.New(slog.NewTextHandler(io.Discard, nil))}
	if _, err := thread.RunStreamed(context.Background(), []Input{TextInput("hello")}, nil); err == nil || !strings.Contains(err.Error(), "start failed") {
		t.Fatalf("expected start failed error, got %v", err)
	}
}

func TestNewLogsCodexVersionMismatch(t *testing.T) {
	if runtime.GOOS == "windows" {
		t.Skip("spawn script test is unix-only")
	}

	var logs bytes.Buffer
	logger := slog.New(slog.NewTextHandler(&logs, nil))
	client, err := New(context.Background(), Options{
		Spawn:               SpawnOptions{CodexPath: writeFakeCodexBinaryWithVersion(t, "999.999.999")},
		Logger:              logger,
		CompatibilityPolicy: Warn,
	})
	if err != nil {
		t.Fatalf("new client error: %v", err)
	}
	defer client.Close()

	logOutput := logs.String()
	if !strings.Contains(logOutput, "codex binary compatibility could not be guaranteed") {
		t.Fatalf("expected version mismatch warning, got %s", logOutput)
	}
	if !strings.Contains(logOutput, "generated_version="+protocol.GeneratedCodexVersion) {
		t.Fatalf("expected generated version in warning, got %s", logOutput)
	}
}

func TestWarnIfCodexVersionCannotBeVerified(t *testing.T) {
	var logs bytes.Buffer
	logger := slog.New(slog.NewTextHandler(&logs, nil))
	if err := checkCodexCompatibility(context.Background(), logger, filepath.Join(t.TempDir(), "missing-codex"), Warn); err != nil {
		t.Fatalf("warn policy returned error: %v", err)
	}

	logOutput := logs.String()
	if !strings.Contains(logOutput, "level=WARN") {
		t.Fatalf("expected warning level, got %s", logOutput)
	}
	if !strings.Contains(logOutput, "codex binary compatibility could not be guaranteed") {
		t.Fatalf("expected verification warning, got %s", logOutput)
	}
	if !strings.Contains(logOutput, "generated_version="+protocol.GeneratedCodexVersion) {
		t.Fatalf("expected generated version in warning, got %s", logOutput)
	}
}

func TestSpawnLogsDoNotExposeArgumentValues(t *testing.T) {
	const secret = "spawn-secret-marker"
	var logs bytes.Buffer
	client, err := New(context.Background(), Options{
		Logger: slog.New(slog.NewTextHandler(&logs, nil)),
		Spawn: SpawnOptions{
			CodexPath:       writeFakeCodexBinary(t),
			ConfigOverrides: []string{"provider_token=" + secret},
			ExtraArgs:       []string{"--credential", secret},
		},
	})
	if err != nil {
		t.Fatalf("new client: %v", err)
	}
	defer client.Close()
	if strings.Contains(logs.String(), secret) {
		t.Fatalf("spawn logs exposed argument values: %s", logs.String())
	}
}

func TestRequireMajorMinorRejectsMismatch(t *testing.T) {
	_, err := New(context.Background(), Options{
		Spawn: SpawnOptions{CodexPath: writeFakeCodexBinaryWithVersion(t, "999.999.999")},
	})
	var compatibilityErr *CodexCompatibilityError
	if !errors.As(err, &compatibilityErr) {
		t.Fatalf("expected CodexCompatibilityError, got %v", err)
	}
	if compatibilityErr.RuntimeVersion != "999.999.999" || compatibilityErr.GeneratedVersion != protocol.GeneratedCodexVersion {
		t.Fatalf("unexpected compatibility error: %#v", compatibilityErr)
	}
}

func TestCompatibilityPoliciesAndPatchTolerance(t *testing.T) {
	parts := strings.Split(protocol.GeneratedCodexVersion, ".")
	if len(parts) < 3 {
		t.Fatalf("generated version lacks patch: %q", protocol.GeneratedCodexVersion)
	}
	patchVersion := strings.Join([]string{parts[0], parts[1], "999"}, ".")

	for _, tt := range []struct {
		name    string
		version string
		policy  CompatibilityPolicy
	}{
		{name: "same major minor patch", version: patchVersion, policy: RequireMajorMinor},
		{name: "warn unparseable", version: "dev", policy: Warn},
		{name: "ignore unparseable", version: "dev", policy: Ignore},
	} {
		t.Run(tt.name, func(t *testing.T) {
			client, err := New(context.Background(), Options{
				Spawn:               SpawnOptions{CodexPath: writeFakeCodexBinaryWithVersion(t, tt.version)},
				CompatibilityPolicy: tt.policy,
			})
			if err != nil {
				t.Fatalf("new client error: %v", err)
			}
			defer client.Close()
		})
	}
}

func TestRequireMajorMinorRejectsUnprobeableAndUnparseable(t *testing.T) {
	for _, path := range []string{
		filepath.Join(t.TempDir(), "missing-codex"),
		writeFakeCodexBinaryWithVersion(t, "dev"),
	} {
		err := checkCodexCompatibility(context.Background(), nil, path, RequireMajorMinor)
		var compatibilityErr *CodexCompatibilityError
		if !errors.As(err, &compatibilityErr) {
			t.Fatalf("expected typed compatibility error for %q, got %v", path, err)
		}
		if compatibilityErr.Cause != nil && !strings.Contains(err.Error(), compatibilityErr.Cause.Error()) {
			t.Fatalf("compatibility error omitted its underlying cause: %v", err)
		}
	}
}

func TestCustomTransportSkipsCompatibilityProbe(t *testing.T) {
	client, err := New(context.Background(), Options{
		Transport: rpc.NewReplayTransport(initializeTranscript()),
		Spawn:     SpawnOptions{CodexPath: filepath.Join(t.TempDir(), "missing-codex")},
	})
	if err != nil {
		t.Fatalf("custom transport should skip probe: %v", err)
	}
	defer client.Close()
}

func TestParseCodexVersionOutput(t *testing.T) {
	if got := parseCodexVersionOutput("codex-cli 0.137.0\n"); got != "0.137.0" {
		t.Fatalf("unexpected version: %q", got)
	}
	if got := parseCodexVersionOutput("warning\ncodex-cli v1.2.3\n"); got != "1.2.3" {
		t.Fatalf("unexpected version with prefix: %q", got)
	}
	if got := parseCodexVersionOutput("codex-cli dev\n"); got != "" {
		t.Fatalf("expected empty version, got %q", got)
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
	return writeFakeCodexBinaryWithVersion(t, protocol.GeneratedCodexVersion)
}

func writeFakeCodexBinaryWithVersion(t *testing.T, version string) string {
	t.Helper()

	path := filepath.Join(t.TempDir(), "fake-codex")
	script := `#!/bin/sh
if [ "$1" = "--version" ]; then
	printf 'codex-cli ` + version + `\n'
	exit 0
fi

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
	mention := MentionInput("thread")
	if mention.Text != "@thread" || len(mention.TextElements) != 1 {
		t.Fatalf("unexpected mention input: %#v", mention)
	}
	if err := mention.validate(); err != nil {
		t.Fatalf("mention should validate: %v", err)
	}
	invalid := Input{Type: InputTypeText, Text: "x", TextElements: []protocol.TextElement{{
		ByteRange: protocol.TextElementByteRange{Start: 2, End: 3},
	}}}
	if err := invalid.validate(); err == nil {
		t.Fatalf("expected invalid text element range")
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

	typed := rpc.Notification{Params: protocol.ThreadGoalUpdatedNotification{ThreadID: "thr_3"}}
	if !matchesThreadID(typed, "thr_3") {
		t.Fatalf("expected typed goal notification to match thread id")
	}
	if matchesThreadID(typed, "thr_4") {
		t.Fatalf("expected typed goal notification not to match another thread id")
	}

	empty := rpc.Notification{Raw: MustJSON(map[string]any{})}
	if !matchesThreadID(empty, "thr_1") {
		t.Fatalf("expected match when thread id missing")
	}
}

type testServerRequestHandler struct{}

func (h *testServerRequestHandler) AccountChatgptAuthTokensRefresh(ctx context.Context, params protocol.ChatgptAuthTokensRefreshParams) (*protocol.ChatgptAuthTokensRefreshResponse, error) {
	return nil, nil
}

func (h *testServerRequestHandler) ApplyPatchApproval(ctx context.Context, params protocol.ApplyPatchApprovalParams) (*protocol.ApplyPatchApprovalResponse, error) {
	return nil, nil
}

func (h *testServerRequestHandler) AttestationGenerate(ctx context.Context, params protocol.AttestationGenerateParams) (*protocol.AttestationGenerateResponse, error) {
	return nil, nil
}

func (h *testServerRequestHandler) ExecCommandApproval(ctx context.Context, params protocol.ExecCommandApprovalParams) (*protocol.ExecCommandApprovalResponse, error) {
	return nil, nil
}

func (h *testServerRequestHandler) ItemCommandExecutionRequestApproval(ctx context.Context, params protocol.CommandExecutionRequestApprovalParams) (*protocol.CommandExecutionRequestApprovalResponse, error) {
	return nil, nil
}

func (h *testServerRequestHandler) ItemFileChangeRequestApproval(ctx context.Context, params protocol.FileChangeRequestApprovalParams) (*protocol.FileChangeRequestApprovalResponse, error) {
	return nil, nil
}

func (h *testServerRequestHandler) ItemPermissionsRequestApproval(ctx context.Context, params protocol.PermissionsRequestApprovalParams) (*protocol.PermissionsRequestApprovalResponse, error) {
	return nil, nil
}

func (h *testServerRequestHandler) ItemToolCall(ctx context.Context, params protocol.DynamicToolCallParams) (*protocol.DynamicToolCallResponse, error) {
	return nil, nil
}

func (h *testServerRequestHandler) ItemToolRequestUserInput(ctx context.Context, params protocol.ToolRequestUserInputParams) (*protocol.ToolRequestUserInputResponse, error) {
	return nil, nil
}

func (h *testServerRequestHandler) McpServerElicitationRequest(ctx context.Context, params protocol.MCPServerElicitationRequestParams) (*protocol.MCPServerElicitationRequestResponse, error) {
	return nil, nil
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
