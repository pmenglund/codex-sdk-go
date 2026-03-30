package codex

import (
	"context"
	"errors"
	"log/slog"

	"github.com/pmenglund/codex-sdk-go/protocol"
)

// AutoApproveHandler accepts every approval request it can.
// Logger controls approval logging. When nil, logs are discarded.
type AutoApproveHandler struct {
	Logger *slog.Logger
}

// ItemCommandExecutionRequestApproval approves command execution requests.
func (h AutoApproveHandler) ItemCommandExecutionRequestApproval(ctx context.Context, params protocol.CommandExecutionRequestApprovalParams) (*protocol.CommandExecutionRequestApprovalResponse, error) {
	logger := resolveLogger(h.Logger)
	logger.Info(
		"codex auto-approving command execution",
		"thread_id", params.ThreadID,
		"turn_id", params.TurnID,
		"item_id", params.ItemID,
		"command", params.Command,
		"cwd", params.Cwd,
	)
	resp := protocol.CommandExecutionRequestApprovalResponse{Decision: "accept"}
	return &resp, nil
}

// ItemFileChangeRequestApproval approves file change requests.
func (h AutoApproveHandler) ItemFileChangeRequestApproval(ctx context.Context, params protocol.FileChangeRequestApprovalParams) (*protocol.FileChangeRequestApprovalResponse, error) {
	logger := resolveLogger(h.Logger)
	logger.Info(
		"codex auto-approving file change",
		"thread_id", params.ThreadID,
		"turn_id", params.TurnID,
		"item_id", params.ItemID,
		"grant_root", params.GrantRoot,
	)
	resp := protocol.FileChangeRequestApprovalResponse{Decision: "accept"}
	return &resp, nil
}

// ItemPermissionsRequestApproval approves permission escalation requests.
func (h AutoApproveHandler) ItemPermissionsRequestApproval(ctx context.Context, params protocol.PermissionsRequestApprovalParams) (*protocol.PermissionsRequestApprovalResponse, error) {
	logger := resolveLogger(h.Logger)
	logger.Info(
		"codex auto-approving permission request",
		"thread_id", params.ThreadID,
		"turn_id", params.TurnID,
		"item_id", params.ItemID,
	)
	resp := protocol.PermissionsRequestApprovalResponse{Permissions: params.Permissions}
	return &resp, nil
}

// ItemToolCall returns an error for dynamic tool calls.
func (h AutoApproveHandler) ItemToolCall(ctx context.Context, params protocol.DynamicToolCallParams) (*protocol.DynamicToolCallResponse, error) {
	logger := resolveLogger(h.Logger)
	logger.Info("codex auto-approve handler cannot execute tool calls")
	return nil, errors.New("tool calls require a custom handler")
}

// ItemToolRequestUserInput returns an error for tool user input prompts.
func (h AutoApproveHandler) ItemToolRequestUserInput(ctx context.Context, params protocol.ToolRequestUserInputParams) (*protocol.ToolRequestUserInputResponse, error) {
	logger := resolveLogger(h.Logger)
	logger.Info(
		"codex auto-approve handler cannot answer tool user input",
		"thread_id", params.ThreadID,
		"turn_id", params.TurnID,
		"item_id", params.ItemID,
		"questions", len(params.Questions),
	)
	return nil, errors.New("tool user input requires a custom handler")
}

// McpServerElicitationRequest returns an error for MCP elicitation prompts.
func (h AutoApproveHandler) McpServerElicitationRequest(ctx context.Context, params protocol.McpServerElicitationRequestParams) (*protocol.McpServerElicitationRequestResponse, error) {
	logger := resolveLogger(h.Logger)
	logger.Info("codex auto-approve handler cannot answer MCP elicitation prompts")
	return nil, errors.New("mcp elicitation requires a custom handler")
}

// AccountChatgptAuthTokensRefresh returns an error for auth refresh requests.
func (h AutoApproveHandler) AccountChatgptAuthTokensRefresh(ctx context.Context, params protocol.ChatgptAuthTokensRefreshParams) (*protocol.ChatgptAuthTokensRefreshResponse, error) {
	logger := resolveLogger(h.Logger)
	logger.Info("codex auto-approve handler cannot refresh chatgpt auth tokens")
	return nil, errors.New("chatgpt auth token refresh requires a custom handler")
}

// ApplyPatchApproval approves legacy patch requests.
func (h AutoApproveHandler) ApplyPatchApproval(ctx context.Context, params protocol.ApplyPatchApprovalParams) (*protocol.ApplyPatchApprovalResponse, error) {
	logger := resolveLogger(h.Logger)
	logger.Info(
		"codex auto-approving patch",
		"conversation_id", params.ConversationID,
		"call_id", params.CallID,
		"file_changes", len(params.FileChanges),
	)
	resp := protocol.ApplyPatchApprovalResponse{Decision: "approved"}
	return &resp, nil
}

// ExecCommandApproval approves legacy command requests.
func (h AutoApproveHandler) ExecCommandApproval(ctx context.Context, params protocol.ExecCommandApprovalParams) (*protocol.ExecCommandApprovalResponse, error) {
	logger := resolveLogger(h.Logger)
	logger.Info(
		"codex auto-approving command",
		"conversation_id", params.ConversationID,
		"call_id", params.CallID,
		"command", params.Command,
		"cwd", params.Cwd,
	)
	resp := protocol.ExecCommandApprovalResponse{Decision: "approved"}
	return &resp, nil
}
