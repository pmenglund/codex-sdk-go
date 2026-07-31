package codex

import (
	"bytes"
	"context"
	"log/slog"
	"strings"
	"testing"

	"github.com/pmenglund/codex-sdk-go/protocol"
)

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
