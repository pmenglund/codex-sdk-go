package codex

import (
	"context"
	"errors"
	"log/slog"

	"github.com/pmenglund/codex-sdk-go/protocol"
)

// AutoApproveHandler accepts every approval request it can. Use it only in a
// trusted environment. Logger controls redacted approval logging. Command
// bodies, paths, working directories, and permission payloads are never logged.
// When Logger is nil, logs are discarded.
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
	)
	resp := protocol.CommandExecutionRequestApprovalResponse{Decision: protocol.MustCommandExecutionApprovalDecision("accept")}
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
	)
	resp := protocol.FileChangeRequestApprovalResponse{Decision: protocol.FileChangeApprovalDecisionAccept}
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

// MCPServerElicitationRequest returns an error for MCP elicitation prompts.
func (h AutoApproveHandler) MCPServerElicitationRequest(ctx context.Context, params protocol.MCPServerElicitationRequestParams) (*protocol.MCPServerElicitationRequestResponse, error) {
	logger := resolveLogger(h.Logger)
	logger.Info("codex auto-approve handler cannot answer MCP elicitation prompts")
	return nil, errors.New("mcp elicitation requires a custom handler")
}

// McpServerElicitationRequest preserves the legacy method spelling.
//
// Deprecated: use MCPServerElicitationRequest.
func (h AutoApproveHandler) McpServerElicitationRequest(ctx context.Context, params protocol.MCPServerElicitationRequestParams) (*protocol.MCPServerElicitationRequestResponse, error) {
	return h.MCPServerElicitationRequest(ctx, params)
}

// AccountChatgptAuthTokensRefresh returns an error for auth refresh requests.
func (h AutoApproveHandler) AccountChatgptAuthTokensRefresh(ctx context.Context, params protocol.ChatgptAuthTokensRefreshParams) (*protocol.ChatgptAuthTokensRefreshResponse, error) {
	logger := resolveLogger(h.Logger)
	logger.Info("codex auto-approve handler cannot refresh chatgpt auth tokens")
	return nil, errors.New("chatgpt auth token refresh requires a custom handler")
}

// AttestationGenerate returns an error for attestation generation requests.
func (h AutoApproveHandler) AttestationGenerate(ctx context.Context, params protocol.AttestationGenerateParams) (*protocol.AttestationGenerateResponse, error) {
	logger := resolveLogger(h.Logger)
	logger.Info("codex auto-approve handler cannot generate attestations")
	return nil, errors.New("attestation generation requires a custom handler")
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
	resp := protocol.ApplyPatchApprovalResponse{Decision: protocol.MustReviewDecision("approved")}
	return &resp, nil
}

// ExecCommandApproval approves legacy command requests.
func (h AutoApproveHandler) ExecCommandApproval(ctx context.Context, params protocol.ExecCommandApprovalParams) (*protocol.ExecCommandApprovalResponse, error) {
	logger := resolveLogger(h.Logger)
	logger.Info(
		"codex auto-approving command",
		"conversation_id", params.ConversationID,
		"call_id", params.CallID,
	)
	resp := protocol.ExecCommandApprovalResponse{Decision: protocol.MustReviewDecision("approved")}
	return &resp, nil
}

// RejectingApprovalHandler rejects approval requests and returns errors for
// interactive requests that require an application-specific policy. Its zero
// value is ready for use.
type RejectingApprovalHandler struct{}

const rejectingApprovalReason = "approval rejected by policy"

func rejectingReviewDecision() protocol.ReviewDecision {
	return protocol.MustReviewDecision(map[string]any{
		"denied": map[string]string{"rejection": rejectingApprovalReason},
	})
}

func (RejectingApprovalHandler) ItemCommandExecutionRequestApproval(context.Context, protocol.CommandExecutionRequestApprovalParams) (*protocol.CommandExecutionRequestApprovalResponse, error) {
	return &protocol.CommandExecutionRequestApprovalResponse{Decision: protocol.MustCommandExecutionApprovalDecision("decline")}, nil
}

func (RejectingApprovalHandler) ItemFileChangeRequestApproval(context.Context, protocol.FileChangeRequestApprovalParams) (*protocol.FileChangeRequestApprovalResponse, error) {
	return &protocol.FileChangeRequestApprovalResponse{Decision: protocol.FileChangeApprovalDecisionDecline}, nil
}

func (RejectingApprovalHandler) ItemPermissionsRequestApproval(context.Context, protocol.PermissionsRequestApprovalParams) (*protocol.PermissionsRequestApprovalResponse, error) {
	return &protocol.PermissionsRequestApprovalResponse{Permissions: map[string]any{}}, nil
}

func (RejectingApprovalHandler) ApplyPatchApproval(context.Context, protocol.ApplyPatchApprovalParams) (*protocol.ApplyPatchApprovalResponse, error) {
	return &protocol.ApplyPatchApprovalResponse{Decision: rejectingReviewDecision()}, nil
}

func (RejectingApprovalHandler) ExecCommandApproval(context.Context, protocol.ExecCommandApprovalParams) (*protocol.ExecCommandApprovalResponse, error) {
	return &protocol.ExecCommandApprovalResponse{Decision: rejectingReviewDecision()}, nil
}

// ApprovalCallbackHandler is the stable subset needed to adapt approval
// policies to ServerRequestCallbacks. Both AutoApproveHandler and
// UnsafeLoggingAutoApproveHandler implement it.
type ApprovalCallbackHandler interface {
	ApplyPatchApproval(context.Context, protocol.ApplyPatchApprovalParams) (*protocol.ApplyPatchApprovalResponse, error)
	ExecCommandApproval(context.Context, protocol.ExecCommandApprovalParams) (*protocol.ExecCommandApprovalResponse, error)
	ItemCommandExecutionRequestApproval(context.Context, protocol.CommandExecutionRequestApprovalParams) (*protocol.CommandExecutionRequestApprovalResponse, error)
	ItemFileChangeRequestApproval(context.Context, protocol.FileChangeRequestApprovalParams) (*protocol.FileChangeRequestApprovalResponse, error)
	ItemPermissionsRequestApproval(context.Context, protocol.PermissionsRequestApprovalParams) (*protocol.PermissionsRequestApprovalResponse, error)
}

// AutoApproveCallbacks adapts the approval methods of handler to the preferred
// optional-callback API. Non-approval server requests remain unsupported.
func AutoApproveCallbacks(handler ApprovalCallbackHandler) *ServerRequestCallbacks {
	if isNilValue(handler) {
		return &ServerRequestCallbacks{}
	}
	return &ServerRequestCallbacks{
		ApprovePatch:            handler.ApplyPatchApproval,
		ApproveCommand:          handler.ExecCommandApproval,
		ApproveCommandExecution: handler.ItemCommandExecutionRequestApproval,
		ApproveFileChange:       handler.ItemFileChangeRequestApproval,
		ApprovePermissions:      handler.ItemPermissionsRequestApproval,
	}
}

// RejectingApprovalCallbacks returns preferred optional callbacks that reject
// command, patch, file-change, and permission approval requests.
func RejectingApprovalCallbacks() *ServerRequestCallbacks {
	handler := RejectingApprovalHandler{}
	return &ServerRequestCallbacks{
		ApprovePatch:            handler.ApplyPatchApproval,
		ApproveCommand:          handler.ExecCommandApproval,
		ApproveCommandExecution: handler.ItemCommandExecutionRequestApproval,
		ApproveFileChange:       handler.ItemFileChangeRequestApproval,
		ApprovePermissions:      handler.ItemPermissionsRequestApproval,
	}
}

func (RejectingApprovalHandler) ItemToolCall(context.Context, protocol.DynamicToolCallParams) (*protocol.DynamicToolCallResponse, error) {
	return nil, errors.New("tool calls require a custom handler")
}

func (RejectingApprovalHandler) ItemToolRequestUserInput(context.Context, protocol.ToolRequestUserInputParams) (*protocol.ToolRequestUserInputResponse, error) {
	return nil, errors.New("tool user input requires a custom handler")
}

func (RejectingApprovalHandler) MCPServerElicitationRequest(context.Context, protocol.MCPServerElicitationRequestParams) (*protocol.MCPServerElicitationRequestResponse, error) {
	return nil, errors.New("mcp elicitation requires a custom handler")
}

// McpServerElicitationRequest preserves the legacy method spelling.
//
// Deprecated: use MCPServerElicitationRequest.
func (h RejectingApprovalHandler) McpServerElicitationRequest(ctx context.Context, params protocol.MCPServerElicitationRequestParams) (*protocol.MCPServerElicitationRequestResponse, error) {
	return h.MCPServerElicitationRequest(ctx, params)
}

func (RejectingApprovalHandler) AccountChatgptAuthTokensRefresh(context.Context, protocol.ChatgptAuthTokensRefreshParams) (*protocol.ChatgptAuthTokensRefreshResponse, error) {
	return nil, errors.New("chatgpt auth token refresh requires a custom handler")
}

func (RejectingApprovalHandler) AttestationGenerate(context.Context, protocol.AttestationGenerateParams) (*protocol.AttestationGenerateResponse, error) {
	return nil, errors.New("attestation generation requires a custom handler")
}

// UnsafeLoggingAutoApproveHandler opts into logging sensitive approval payloads.
// Use only when logs have access controls appropriate for command text and paths.
type UnsafeLoggingAutoApproveHandler struct {
	AutoApproveHandler
}

// NewUnsafeLoggingAutoApproveHandler returns an auto-approver that logs
// sensitive command and path details.
func NewUnsafeLoggingAutoApproveHandler(logger *slog.Logger) UnsafeLoggingAutoApproveHandler {
	return UnsafeLoggingAutoApproveHandler{AutoApproveHandler: AutoApproveHandler{Logger: logger}}
}

func (h UnsafeLoggingAutoApproveHandler) ItemCommandExecutionRequestApproval(ctx context.Context, params protocol.CommandExecutionRequestApprovalParams) (*protocol.CommandExecutionRequestApprovalResponse, error) {
	resolveLogger(h.Logger).Info("codex auto-approval sensitive details", "request_kind", "command_execution", "command", optionalString(params.Command), "cwd", optionalString(params.Cwd))
	return h.AutoApproveHandler.ItemCommandExecutionRequestApproval(ctx, params)
}

func (h UnsafeLoggingAutoApproveHandler) ItemFileChangeRequestApproval(ctx context.Context, params protocol.FileChangeRequestApprovalParams) (*protocol.FileChangeRequestApprovalResponse, error) {
	var grantRoot any
	if params.GrantRoot != nil {
		grantRoot = *params.GrantRoot
	}
	resolveLogger(h.Logger).Info("codex auto-approval sensitive details", "request_kind", "file_change", "grant_root", grantRoot)
	return h.AutoApproveHandler.ItemFileChangeRequestApproval(ctx, params)
}

func (h UnsafeLoggingAutoApproveHandler) ItemPermissionsRequestApproval(ctx context.Context, params protocol.PermissionsRequestApprovalParams) (*protocol.PermissionsRequestApprovalResponse, error) {
	resolveLogger(h.Logger).Info("codex auto-approval sensitive details", "request_kind", "permissions", "permissions", params.Permissions)
	return h.AutoApproveHandler.ItemPermissionsRequestApproval(ctx, params)
}

func (h UnsafeLoggingAutoApproveHandler) ApplyPatchApproval(ctx context.Context, params protocol.ApplyPatchApprovalParams) (*protocol.ApplyPatchApprovalResponse, error) {
	resolveLogger(h.Logger).Info("codex auto-approval sensitive details", "request_kind", "legacy_patch", "file_changes", params.FileChanges, "grant_root", params.GrantRoot)
	return h.AutoApproveHandler.ApplyPatchApproval(ctx, params)
}

func (h UnsafeLoggingAutoApproveHandler) ExecCommandApproval(ctx context.Context, params protocol.ExecCommandApprovalParams) (*protocol.ExecCommandApprovalResponse, error) {
	resolveLogger(h.Logger).Info("codex auto-approval sensitive details", "request_kind", "legacy_command", "command", params.Command, "cwd", params.Cwd)
	return h.AutoApproveHandler.ExecCommandApproval(ctx, params)
}

func optionalString(value *string) any {
	if value == nil {
		return nil
	}
	return *value
}
