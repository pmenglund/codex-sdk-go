package codex_test

import (
	"context"
	"errors"
	"testing"

	"github.com/pmenglund/codex-sdk-go"
	"github.com/pmenglund/codex-sdk-go/protocol"
	"github.com/pmenglund/codex-sdk-go/rpc"
)

type threadReader interface {
	ThreadRead(context.Context, protocol.ThreadReadParams) (*protocol.ThreadReadResponse, error)
}

type partialServerRequestHandler struct {
	rpc.UnimplementedServerRequestHandler
}

var (
	_ threadReader = (*rpc.Client)(nil)
	_              = partialServerRequestHandler{}
	//lint:ignore SA1019 compatibility fixture proves the former OAuth spelling remains available
	_ protocol.MCPServerOauthLoginParams
	//lint:ignore SA1019 compatibility fixture proves former generated backing names remain available
	_ protocol.SanitizedMCPServerOauthLoginResponseJSON
	//lint:ignore SA1019 compatibility fixture proves former unsuffixed generated aliases remain available
	_ protocol.SanitizedMCPServerOauthLoginCompletedNotification
	//lint:ignore SA1019 compatibility fixture proves former unsuffixed generated aliases remain available
	_ protocol.SanitizedMCPServerOauthLoginParams
	//lint:ignore SA1019 compatibility fixture proves former unsuffixed generated aliases remain available
	_ protocol.SanitizedMCPServerOauthLoginResponse
	//lint:ignore SA1019 compatibility fixture proves former union constants remain available
	_ = protocol.ThreadItemKindMcpToolCall
)

func TestPreferredAPISurfaceCompilesExternally(t *testing.T) {
	options := codex.Options{
		RequestHandler: &codex.ServerRequestCallbacks{
			ApproveFileChange: func(context.Context, protocol.FileChangeRequestApprovalParams) (*protocol.FileChangeRequestApprovalResponse, error) {
				return &protocol.FileChangeRequestApprovalResponse{Decision: protocol.FileChangeApprovalDecisionDecline}, nil
			},
		},
	}
	listOptions := codex.ThreadListOptions{
		SortDirection: protocol.SortDirectionDesc,
		SortKey:       protocol.ThreadSortKeyUpdatedAt,
	}
	oauth := protocol.MCPServerOAuthLoginParams{}

	var turnErr *codex.TurnError
	err := error(&codex.TurnError{})
	if !errors.Is(err, codex.ErrTurnFailed) || !errors.As(err, &turnErr) {
		t.Fatal("TurnError must support errors.Is and errors.As")
	}

	_ = options
	_ = listOptions
	_ = oauth
}
