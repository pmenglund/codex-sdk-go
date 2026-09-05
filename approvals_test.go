package codex

import (
	"bytes"
	"context"
	"encoding/json"
	"log/slog"
	"strings"
	"testing"

	"github.com/pmenglund/codex-sdk-go/protocol"
)

func TestUnsafeApprovalLoggingAndDecisions(t *testing.T) {
	const marker = "explicit-sensitive-payload"
	ctx := context.Background()
	for _, tt := range []struct {
		name string
		call func(UnsafeLoggingAutoApproveHandler) (any, error)
		want string
	}{
		{"file change", func(h UnsafeLoggingAutoApproveHandler) (any, error) {
			path := marker
			return h.ItemFileChangeRequestApproval(ctx, protocol.FileChangeRequestApprovalParams{GrantRoot: &path})
		}, `{"decision":"accept"}`},
		{"permissions", func(h UnsafeLoggingAutoApproveHandler) (any, error) {
			return h.ItemPermissionsRequestApproval(ctx, protocol.PermissionsRequestApprovalParams{Permissions: map[string]any{"path": marker}})
		}, `{"permissions":{"path":"explicit-sensitive-payload"}}`},
		{"patch", func(h UnsafeLoggingAutoApproveHandler) (any, error) {
			return h.ApplyPatchApproval(ctx, protocol.ApplyPatchApprovalParams{FileChanges: map[string]any{marker: "add"}})
		}, `{"decision":"approved"}`},
		{"command", func(h UnsafeLoggingAutoApproveHandler) (any, error) {
			return h.ExecCommandApproval(ctx, protocol.ExecCommandApprovalParams{Command: []string{"echo", marker}})
		}, `{"decision":"approved"}`},
	} {
		t.Run(tt.name, func(t *testing.T) {
			var logs bytes.Buffer
			result, err := tt.call(NewUnsafeLoggingAutoApproveHandler(slog.New(slog.NewTextHandler(&logs, nil))))
			if err != nil {
				t.Fatal(err)
			}
			data, err := json.Marshal(result)
			if err != nil || string(data) != tt.want {
				t.Fatalf("approval=%s err=%v, want %s", data, err, tt.want)
			}
			if !strings.Contains(logs.String(), marker) {
				t.Fatalf("explicit unsafe logger omitted payload: %s", logs.String())
			}
		})
	}
	result, err := NewUnsafeLoggingAutoApproveHandler(nil).ItemFileChangeRequestApproval(ctx, protocol.FileChangeRequestApprovalParams{})
	if err != nil || result.Decision != protocol.FileChangeApprovalDecisionAccept {
		t.Fatalf("nil logger and grant root: %#v, %v", result, err)
	}
}

func TestRejectingHandlerRequiresCustomInteractivePolicy(t *testing.T) {
	h := RejectingApprovalHandler{}
	ctx := context.Background()
	for _, tt := range []struct {
		name string
		call func() (bool, error)
		want string
	}{
		{"tool", func() (bool, error) {
			r, e := h.ItemToolCall(ctx, protocol.DynamicToolCallParams{})
			return r == nil, e
		}, "tool calls require a custom handler"},
		{"input", func() (bool, error) {
			r, e := h.ItemToolRequestUserInput(ctx, protocol.ToolRequestUserInputParams{})
			return r == nil, e
		}, "tool user input requires a custom handler"},
		{"elicitation", func() (bool, error) {
			r, e := h.MCPServerElicitationRequest(ctx, protocol.MCPServerElicitationRequestParams{})
			return r == nil, e
		}, "mcp elicitation requires a custom handler"},
		{"auth", func() (bool, error) {
			r, e := h.AccountChatgptAuthTokensRefresh(ctx, protocol.ChatgptAuthTokensRefreshParams{})
			return r == nil, e
		}, "chatgpt auth token refresh requires a custom handler"},
		{"attestation", func() (bool, error) {
			r, e := h.AttestationGenerate(ctx, protocol.AttestationGenerateParams{})
			return r == nil, e
		}, "attestation generation requires a custom handler"},
	} {
		t.Run(tt.name, func(t *testing.T) {
			empty, err := tt.call()
			if !empty || err == nil || err.Error() != tt.want {
				t.Fatalf("empty response=%v, err=%v, want %s", empty, err, tt.want)
			}
		})
	}
}

func TestAutoApproveHandlerRedactsSensitivePayloads(t *testing.T) {
	const marker = "TOP-SECRET-MARKER"
	var logs bytes.Buffer
	handler := AutoApproveHandler{Logger: slog.New(slog.NewTextHandler(&logs, nil))}
	ctx := context.Background()
	path := "/tmp/" + marker
	command := "echo " + marker

	_, _ = handler.ItemCommandExecutionRequestApproval(ctx, protocol.CommandExecutionRequestApprovalParams{ThreadID: "thr", TurnID: "turn", ItemID: "item", Command: &command, Cwd: &path})
	_, _ = handler.ItemFileChangeRequestApproval(ctx, protocol.FileChangeRequestApprovalParams{ThreadID: "thr", TurnID: "turn", ItemID: "item", GrantRoot: &path})
	_, _ = handler.ItemPermissionsRequestApproval(ctx, protocol.PermissionsRequestApprovalParams{ThreadID: "thr", TurnID: "turn", ItemID: "item", Permissions: map[string]any{"path": path}})
	_, _ = handler.ApplyPatchApproval(ctx, protocol.ApplyPatchApprovalParams{ConversationID: "thr", CallID: "call", FileChanges: map[string]any{path: marker}, GrantRoot: &path})
	_, _ = handler.ExecCommandApproval(ctx, protocol.ExecCommandApprovalParams{ConversationID: "thr", CallID: "call", Command: []string{"echo", marker}, Cwd: path})

	if strings.Contains(logs.String(), marker) {
		t.Fatalf("default approval logs leaked sensitive marker: %s", logs.String())
	}
}

func TestUnsafeLoggingAutoApproveHandlerEmitsSensitivePayloads(t *testing.T) {
	const marker = "TOP-SECRET-MARKER"
	var logs bytes.Buffer
	handler := NewUnsafeLoggingAutoApproveHandler(slog.New(slog.NewTextHandler(&logs, nil)))
	command := "echo " + marker
	_, _ = handler.ItemCommandExecutionRequestApproval(context.Background(), protocol.CommandExecutionRequestApprovalParams{Command: &command})
	if !strings.Contains(logs.String(), marker) {
		t.Fatalf("expected explicit unsafe handler to log marker, got %s", logs.String())
	}
	callbacks := AutoApproveCallbacks(handler)
	if callbacks.ApproveCommandExecution == nil || callbacks.ApproveFileChange == nil || callbacks.ApprovePermissions == nil {
		t.Fatalf("unsafe handler did not adapt to preferred callbacks: %#v", callbacks)
	}
}

func TestAutoApproveCallbacksAcceptsTypedNilHandler(t *testing.T) {
	var handler *AutoApproveHandler
	callbacks := AutoApproveCallbacks(handler)
	if callbacks == nil || callbacks.ApprovePatch != nil || callbacks.ApproveCommandExecution != nil {
		t.Fatalf("typed nil handler produced active callbacks: %#v", callbacks)
	}
}

func TestRejectingApprovalHandler(t *testing.T) {
	handler := RejectingApprovalHandler{}
	ctx := context.Background()
	command, err := handler.ItemCommandExecutionRequestApproval(ctx, protocol.CommandExecutionRequestApprovalParams{})
	if err != nil || string(command.Decision.RawJSON()) != `"decline"` {
		t.Fatalf("unexpected command rejection: %#v err=%v", command, err)
	}
	file, err := handler.ItemFileChangeRequestApproval(ctx, protocol.FileChangeRequestApprovalParams{})
	if err != nil || file.Decision != protocol.FileChangeApprovalDecisionDecline {
		t.Fatalf("unexpected file rejection: %#v err=%v", file, err)
	}
	patch, err := handler.ApplyPatchApproval(ctx, protocol.ApplyPatchApprovalParams{})
	if err != nil || string(patch.Decision.RawJSON()) != `{"denied":{"rejection":"approval rejected by policy"}}` {
		t.Fatalf("unexpected patch rejection: %#v err=%v", patch, err)
	}
	legacy, err := handler.ExecCommandApproval(ctx, protocol.ExecCommandApprovalParams{})
	if err != nil || string(legacy.Decision.RawJSON()) != `{"denied":{"rejection":"approval rejected by policy"}}` {
		t.Fatalf("unexpected legacy rejection: %#v err=%v", legacy, err)
	}
	permissions, err := handler.ItemPermissionsRequestApproval(ctx, protocol.PermissionsRequestApprovalParams{})
	if err != nil || permissions.Permissions == nil {
		t.Fatalf("unexpected permissions rejection: %#v err=%v", permissions, err)
	}
}
