package codex

import (
	"context"
	"errors"

	"github.com/pmenglund/codex-sdk-go/protocol"
)

// AccountOptions configures an account/read request.
type AccountOptions struct {
	// RefreshToken requests a proactive token refresh before account data is returned.
	RefreshToken bool
}

// Account returns account state from the app-server.
func (c *Codex) Account(ctx context.Context, opts AccountOptions) (*protocol.GetAccountResponse, error) {
	if err := c.ensureReady(); err != nil {
		return nil, err
	}
	return c.client.AccountRead(ctx, protocol.GetAccountParams{RefreshToken: opts.RefreshToken})
}

// StartLogin starts an app-server account login flow using protocol login params.
func (c *Codex) StartLogin(ctx context.Context, params protocol.LoginAccountParams) (*protocol.LoginAccountResponse, error) {
	if err := c.ensureReady(); err != nil {
		return nil, err
	}
	if params == nil {
		return nil, errors.New("login params are required")
	}
	return c.client.AccountLoginStart(ctx, params)
}

// CancelLogin cancels an in-progress account login flow.
func (c *Codex) CancelLogin(ctx context.Context, loginID string) (*protocol.CancelLoginAccountResponse, error) {
	if err := c.ensureReady(); err != nil {
		return nil, err
	}
	if loginID == "" {
		return nil, errors.New("login id is required")
	}
	return c.client.AccountLoginCancel(ctx, protocol.CancelLoginAccountParams{LoginID: loginID})
}

// Logout logs out the current account.
func (c *Codex) Logout(ctx context.Context) (*protocol.LogoutAccountResponse, error) {
	if err := c.ensureReady(); err != nil {
		return nil, err
	}
	return c.client.AccountLogout(ctx)
}

// ListModelsOptions configures a model/list request.
type ListModelsOptions struct {
	// Cursor continues listing after a previous response cursor.
	Cursor string
	// IncludeHidden includes models hidden from the default picker list.
	IncludeHidden *bool
	// Limit caps the number of models returned by the app-server.
	Limit *int
}

// ListModels returns available models from the app-server.
func (c *Codex) ListModels(ctx context.Context, opts ListModelsOptions) (*protocol.ModelListResponse, error) {
	if err := c.ensureReady(); err != nil {
		return nil, err
	}
	params := protocol.ModelListParams{
		IncludeHidden: opts.IncludeHidden,
		Limit:         opts.Limit,
	}
	if opts.Cursor != "" {
		params.Cursor = stringPtr(opts.Cursor)
	}
	return c.client.ModelList(ctx, params)
}
