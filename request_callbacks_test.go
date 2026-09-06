package codex

import (
	"context"
	"errors"
	"testing"
	"testing/synctest"

	"github.com/pmenglund/codex-sdk-go/protocol"
	"github.com/pmenglund/codex-sdk-go/rpc"
)

func TestServerRequestCallbacksAreOptional(t *testing.T) {
	handler := ServerRequestCallbacks{}
	if _, err := handler.ItemToolCall(context.Background(), protocol.DynamicToolCallParams{}); !errors.Is(err, rpc.ErrServerRequestUnsupported) {
		t.Fatalf("expected unsupported callback, got %v", err)
	}

	called := false
	handler.ApproveFileChange = func(context.Context, protocol.FileChangeRequestApprovalParams) (*protocol.FileChangeRequestApprovalResponse, error) {
		called = true
		return &protocol.FileChangeRequestApprovalResponse{Decision: protocol.FileChangeApprovalDecisionAccept}, nil
	}
	if _, err := handler.ItemFileChangeRequestApproval(context.Background(), protocol.FileChangeRequestApprovalParams{}); err != nil || !called {
		t.Fatalf("configured callback was not called: %v", err)
	}
}

func TestNewRejectsConflictingRequestHandlers(t *testing.T) {
	_, err := New(context.Background(), Options{
		ApprovalHandler: RejectingApprovalHandler{},
		RequestHandler:  &ServerRequestCallbacks{},
	})
	if err == nil {
		t.Fatalf("expected request handler conflict")
	}
}

func TestApprovalCallbackAdapters(t *testing.T) {
	rejecting := RejectingApprovalCallbacks()
	response, err := rejecting.ItemFileChangeRequestApproval(context.Background(), protocol.FileChangeRequestApprovalParams{})
	if err != nil || response.Decision != protocol.FileChangeApprovalDecisionDecline {
		t.Fatalf("unexpected rejection: %#v, %v", response, err)
	}

	auto := AutoApproveCallbacks(AutoApproveHandler{})
	response, err = auto.ItemFileChangeRequestApproval(context.Background(), protocol.FileChangeRequestApprovalParams{})
	if err != nil || response.Decision != protocol.FileChangeApprovalDecisionAccept {
		t.Fatalf("unexpected approval: %#v, %v", response, err)
	}
	if _, err := auto.ItemToolCall(context.Background(), protocol.DynamicToolCallParams{}); !errors.Is(err, rpc.ErrServerRequestUnsupported) {
		t.Fatalf("expected non-approval callback to remain unsupported, got %v", err)
	}
}

func TestServerRequestCallbacksRouteEveryMethod(t *testing.T) {
	called := make(map[string]bool)
	mark := func(name string) { called[name] = true }
	handler := ServerRequestCallbacks{
		RefreshAuthTokens: func(context.Context, protocol.ChatgptAuthTokensRefreshParams) (*protocol.ChatgptAuthTokensRefreshResponse, error) {
			mark("auth")
			return nil, nil
		},
		ApprovePatch: func(context.Context, protocol.ApplyPatchApprovalParams) (*protocol.ApplyPatchApprovalResponse, error) {
			mark("patch")
			return nil, nil
		},
		GenerateAttestation: func(context.Context, protocol.AttestationGenerateParams) (*protocol.AttestationGenerateResponse, error) {
			mark("attestation")
			return nil, nil
		},
		ApproveCommand: func(context.Context, protocol.ExecCommandApprovalParams) (*protocol.ExecCommandApprovalResponse, error) {
			mark("command")
			return nil, nil
		},
		ApproveCommandExecution: func(context.Context, protocol.CommandExecutionRequestApprovalParams) (*protocol.CommandExecutionRequestApprovalResponse, error) {
			mark("execution")
			return nil, nil
		},
		ApproveFileChange: func(context.Context, protocol.FileChangeRequestApprovalParams) (*protocol.FileChangeRequestApprovalResponse, error) {
			mark("file")
			return nil, nil
		},
		ApprovePermissions: func(context.Context, protocol.PermissionsRequestApprovalParams) (*protocol.PermissionsRequestApprovalResponse, error) {
			mark("permissions")
			return nil, nil
		},
		CallTool: func(context.Context, protocol.DynamicToolCallParams) (*protocol.DynamicToolCallResponse, error) {
			mark("tool")
			return nil, nil
		},
		RequestUserInput: func(context.Context, protocol.ToolRequestUserInputParams) (*protocol.ToolRequestUserInputResponse, error) {
			mark("input")
			return nil, nil
		},
		RequestMCPServerElicitation: func(context.Context, protocol.MCPServerElicitationRequestParams) (*protocol.MCPServerElicitationRequestResponse, error) {
			mark("mcp")
			return nil, nil
		},
	}
	ctx := context.Background()
	_, _ = handler.AccountChatgptAuthTokensRefresh(ctx, protocol.ChatgptAuthTokensRefreshParams{})
	_, _ = handler.ApplyPatchApproval(ctx, protocol.ApplyPatchApprovalParams{})
	_, _ = handler.AttestationGenerate(ctx, protocol.AttestationGenerateParams{})
	_, _ = handler.ExecCommandApproval(ctx, protocol.ExecCommandApprovalParams{})
	_, _ = handler.ItemCommandExecutionRequestApproval(ctx, protocol.CommandExecutionRequestApprovalParams{})
	_, _ = handler.ItemFileChangeRequestApproval(ctx, protocol.FileChangeRequestApprovalParams{})
	_, _ = handler.ItemPermissionsRequestApproval(ctx, protocol.PermissionsRequestApprovalParams{})
	_, _ = handler.ItemToolCall(ctx, protocol.DynamicToolCallParams{})
	_, _ = handler.ItemToolRequestUserInput(ctx, protocol.ToolRequestUserInputParams{})
	_, _ = handler.McpServerElicitationRequest(ctx, nil)
	for _, name := range []string{"auth", "patch", "attestation", "command", "execution", "file", "permissions", "tool", "input", "mcp"} {
		if !called[name] {
			t.Errorf("callback %q was not routed", name)
		}
	}

	missing := ServerRequestCallbacks{}
	assertUnsupported := func(name string, err error) {
		t.Helper()
		if !errors.Is(err, rpc.ErrServerRequestUnsupported) {
			t.Errorf("%s: expected unsupported, got %v", name, err)
		}
	}
	_, err := missing.AccountChatgptAuthTokensRefresh(ctx, protocol.ChatgptAuthTokensRefreshParams{})
	assertUnsupported("auth", err)
	_, err = missing.ApplyPatchApproval(ctx, protocol.ApplyPatchApprovalParams{})
	assertUnsupported("patch", err)
	_, err = missing.AttestationGenerate(ctx, protocol.AttestationGenerateParams{})
	assertUnsupported("attestation", err)
	_, err = missing.ExecCommandApproval(ctx, protocol.ExecCommandApprovalParams{})
	assertUnsupported("command", err)
	_, err = missing.ItemCommandExecutionRequestApproval(ctx, protocol.CommandExecutionRequestApprovalParams{})
	assertUnsupported("execution", err)
	_, err = missing.ItemFileChangeRequestApproval(ctx, protocol.FileChangeRequestApprovalParams{})
	assertUnsupported("file", err)
	_, err = missing.ItemPermissionsRequestApproval(ctx, protocol.PermissionsRequestApprovalParams{})
	assertUnsupported("permissions", err)
	_, err = missing.ItemToolCall(ctx, protocol.DynamicToolCallParams{})
	assertUnsupported("tool", err)
	_, err = missing.ItemToolRequestUserInput(ctx, protocol.ToolRequestUserInputParams{})
	assertUnsupported("input", err)
	_, err = missing.McpServerElicitationRequest(ctx, nil)
	assertUnsupported("mcp", err)
}

func TestNewInstallsPreferredRequestHandler(t *testing.T) {
	synctest.Test(t, func(t *testing.T) {
		info := protocol.ClientInfo{Name: "callbacks-test", Version: "test"}
		called := make(chan struct{}, 1)
		client, err := New(context.Background(), Options{
			Transport: rpc.NewReplayTransport([]rpc.TranscriptEntry{
				writeLine(rpc.JSONRPCRequest{ID: rpc.NewIntRequestID(1), Method: "initialize", Params: mustRaw(protocol.InitializeParams{ClientInfo: info})}),
				readLine(rpc.JSONRPCResponse{ID: rpc.NewIntRequestID(1), Result: mustRaw(map[string]any{})}),
				writeLine(rpc.JSONRPCNotification{Method: "initialized"}),
				readLine(rpc.JSONRPCRequest{ID: rpc.NewIntRequestID(9), Method: "item/fileChange/requestApproval", Params: mustRaw(map[string]any{"threadId": "thr", "turnId": "turn", "itemId": "item"})}),
				writeLine(rpc.JSONRPCResponse{ID: rpc.NewIntRequestID(9), Result: mustRaw(protocol.FileChangeRequestApprovalResponse{Decision: protocol.FileChangeApprovalDecisionDecline})}),
			}),
			ClientInfo: info,
			RequestHandler: &ServerRequestCallbacks{
				ApproveFileChange: func(context.Context, protocol.FileChangeRequestApprovalParams) (*protocol.FileChangeRequestApprovalResponse, error) {
					called <- struct{}{}
					return &protocol.FileChangeRequestApprovalResponse{Decision: protocol.FileChangeApprovalDecisionDecline}, nil
				},
			},
		})
		if err != nil {
			t.Fatalf("new client: %v", err)
		}
		defer client.Close()
		synctest.Wait()
		select {
		case <-called:
		default:
			t.Fatal("preferred request handler was not called")
		}
	})
}
