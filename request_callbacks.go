package codex

import (
	"context"

	"github.com/pmenglund/codex-sdk-go/protocol"
	"github.com/pmenglund/codex-sdk-go/rpc"
)

// ServerRequestCallbacks is a forward-compatible set of optional callbacks for
// requests initiated by the app-server. Unset callbacks report
// rpc.ErrServerRequestUnsupported. The client may invoke different callbacks
// concurrently; callbacks that share state must synchronize access to it.
type ServerRequestCallbacks struct {
	// RefreshAuthTokens handles account/chatgptAuthTokens/refresh.
	RefreshAuthTokens func(context.Context, protocol.ChatgptAuthTokensRefreshParams) (*protocol.ChatgptAuthTokensRefreshResponse, error)
	// ApprovePatch handles the legacy applyPatchApproval request.
	ApprovePatch func(context.Context, protocol.ApplyPatchApprovalParams) (*protocol.ApplyPatchApprovalResponse, error)
	// GenerateAttestation handles attestation/generate.
	GenerateAttestation func(context.Context, protocol.AttestationGenerateParams) (*protocol.AttestationGenerateResponse, error)
	// ApproveCommand handles the legacy execCommandApproval request.
	ApproveCommand func(context.Context, protocol.ExecCommandApprovalParams) (*protocol.ExecCommandApprovalResponse, error)
	// ApproveCommandExecution handles item/commandExecution/requestApproval.
	ApproveCommandExecution func(context.Context, protocol.CommandExecutionRequestApprovalParams) (*protocol.CommandExecutionRequestApprovalResponse, error)
	// ApproveFileChange handles item/fileChange/requestApproval.
	ApproveFileChange func(context.Context, protocol.FileChangeRequestApprovalParams) (*protocol.FileChangeRequestApprovalResponse, error)
	// ApprovePermissions handles item/permissions/requestApproval.
	ApprovePermissions func(context.Context, protocol.PermissionsRequestApprovalParams) (*protocol.PermissionsRequestApprovalResponse, error)
	// CallTool handles item/tool/call.
	CallTool func(context.Context, protocol.DynamicToolCallParams) (*protocol.DynamicToolCallResponse, error)
	// RequestUserInput handles item/tool/requestUserInput.
	RequestUserInput func(context.Context, protocol.ToolRequestUserInputParams) (*protocol.ToolRequestUserInputResponse, error)
	// RequestMCPServerElicitation handles mcpServer/elicitation/request.
	RequestMCPServerElicitation func(context.Context, protocol.MCPServerElicitationRequestParams) (*protocol.MCPServerElicitationRequestResponse, error)
}

func (h ServerRequestCallbacks) AccountChatgptAuthTokensRefresh(ctx context.Context, params protocol.ChatgptAuthTokensRefreshParams) (*protocol.ChatgptAuthTokensRefreshResponse, error) {
	if h.RefreshAuthTokens == nil {
		return nil, rpc.ErrServerRequestUnsupported
	}
	return h.RefreshAuthTokens(ctx, params)
}
func (h ServerRequestCallbacks) ApplyPatchApproval(ctx context.Context, params protocol.ApplyPatchApprovalParams) (*protocol.ApplyPatchApprovalResponse, error) {
	if h.ApprovePatch == nil {
		return nil, rpc.ErrServerRequestUnsupported
	}
	return h.ApprovePatch(ctx, params)
}
func (h ServerRequestCallbacks) AttestationGenerate(ctx context.Context, params protocol.AttestationGenerateParams) (*protocol.AttestationGenerateResponse, error) {
	if h.GenerateAttestation == nil {
		return nil, rpc.ErrServerRequestUnsupported
	}
	return h.GenerateAttestation(ctx, params)
}
func (h ServerRequestCallbacks) ExecCommandApproval(ctx context.Context, params protocol.ExecCommandApprovalParams) (*protocol.ExecCommandApprovalResponse, error) {
	if h.ApproveCommand == nil {
		return nil, rpc.ErrServerRequestUnsupported
	}
	return h.ApproveCommand(ctx, params)
}
func (h ServerRequestCallbacks) ItemCommandExecutionRequestApproval(ctx context.Context, params protocol.CommandExecutionRequestApprovalParams) (*protocol.CommandExecutionRequestApprovalResponse, error) {
	if h.ApproveCommandExecution == nil {
		return nil, rpc.ErrServerRequestUnsupported
	}
	return h.ApproveCommandExecution(ctx, params)
}
func (h ServerRequestCallbacks) ItemFileChangeRequestApproval(ctx context.Context, params protocol.FileChangeRequestApprovalParams) (*protocol.FileChangeRequestApprovalResponse, error) {
	if h.ApproveFileChange == nil {
		return nil, rpc.ErrServerRequestUnsupported
	}
	return h.ApproveFileChange(ctx, params)
}
func (h ServerRequestCallbacks) ItemPermissionsRequestApproval(ctx context.Context, params protocol.PermissionsRequestApprovalParams) (*protocol.PermissionsRequestApprovalResponse, error) {
	if h.ApprovePermissions == nil {
		return nil, rpc.ErrServerRequestUnsupported
	}
	return h.ApprovePermissions(ctx, params)
}
func (h ServerRequestCallbacks) ItemToolCall(ctx context.Context, params protocol.DynamicToolCallParams) (*protocol.DynamicToolCallResponse, error) {
	if h.CallTool == nil {
		return nil, rpc.ErrServerRequestUnsupported
	}
	return h.CallTool(ctx, params)
}
func (h ServerRequestCallbacks) ItemToolRequestUserInput(ctx context.Context, params protocol.ToolRequestUserInputParams) (*protocol.ToolRequestUserInputResponse, error) {
	if h.RequestUserInput == nil {
		return nil, rpc.ErrServerRequestUnsupported
	}
	return h.RequestUserInput(ctx, params)
}
func (h ServerRequestCallbacks) MCPServerElicitationRequest(ctx context.Context, params protocol.MCPServerElicitationRequestParams) (*protocol.MCPServerElicitationRequestResponse, error) {
	if h.RequestMCPServerElicitation == nil {
		return nil, rpc.ErrServerRequestUnsupported
	}
	return h.RequestMCPServerElicitation(ctx, params)
}

// McpServerElicitationRequest preserves the legacy method spelling.
//
// Deprecated: use MCPServerElicitationRequest.
func (h ServerRequestCallbacks) McpServerElicitationRequest(ctx context.Context, params protocol.MCPServerElicitationRequestParams) (*protocol.MCPServerElicitationRequestResponse, error) {
	return h.MCPServerElicitationRequest(ctx, params)
}
