package codex

import (
	"context"
	"encoding/json"
	"errors"
	"log/slog"
	"strings"
	"testing"

	"github.com/pmenglund/codex-sdk-go/protocol"
	"github.com/pmenglund/codex-sdk-go/rpc"
)

func TestAccountAndModelHelpersWithReplay(t *testing.T) {
	ctx := context.Background()
	info := protocol.ClientInfo{Name: "codex-go-test", Version: "test"}
	includeHidden := true
	limit := 2

	client, err := New(ctx, Options{
		Transport: rpc.NewReplayTransport([]rpc.TranscriptEntry{
			writeLine(rpc.JSONRPCRequest{ID: rpc.NewIntRequestID(1), Method: "initialize", Params: mustRaw(protocol.InitializeParams{ClientInfo: info})}),
			readLine(rpc.JSONRPCResponse{ID: rpc.NewIntRequestID(1), Result: mustRaw(map[string]any{})}),
			writeLine(rpc.JSONRPCNotification{Method: "initialized"}),
			writeLine(rpc.JSONRPCRequest{ID: rpc.NewIntRequestID(2), Method: "account/read", Params: mustRaw(map[string]any{"refreshToken": true})}),
			readLine(rpc.JSONRPCResponse{ID: rpc.NewIntRequestID(2), Result: mustRaw(map[string]any{"requiresOpenaiAuth": false, "account": map[string]any{"type": "chatgpt", "email": "dev@example.com", "planType": "plus"}})}),
			writeLine(rpc.JSONRPCRequest{ID: rpc.NewIntRequestID(3), Method: "model/list", Params: mustRaw(map[string]any{"cursor": "cur_1", "includeHidden": true, "limit": 2})}),
			readLine(rpc.JSONRPCResponse{ID: rpc.NewIntRequestID(3), Result: mustRaw(map[string]any{"data": []any{}, "nextCursor": nil})}),
		}),
		ClientInfo: info,
	})
	if err != nil {
		t.Fatalf("new client error: %v", err)
	}
	defer client.Close()

	account, err := client.Account(ctx, AccountOptions{RefreshToken: true})
	if err != nil {
		t.Fatalf("account error: %v", err)
	}
	if account.RequiresOpenaiAuth {
		t.Fatalf("expected no auth requirement")
	}

	models, err := client.ListModels(ctx, ListModelsOptions{Cursor: "cur_1", IncludeHidden: &includeHidden, Limit: &limit})
	if err != nil {
		t.Fatalf("list models error: %v", err)
	}
	if len(models.Data) != 0 {
		t.Fatalf("expected no models in replay")
	}
}

func TestLoginHelpersWithReplay(t *testing.T) {
	ctx := context.Background()
	info := protocol.ClientInfo{Name: "codex-go-test", Version: "test"}
	loginParams := map[string]any{"type": "chatgpt"}

	client, err := New(ctx, Options{
		Transport: rpc.NewReplayTransport([]rpc.TranscriptEntry{
			writeLine(rpc.JSONRPCRequest{ID: rpc.NewIntRequestID(1), Method: "initialize", Params: mustRaw(protocol.InitializeParams{ClientInfo: info})}),
			readLine(rpc.JSONRPCResponse{ID: rpc.NewIntRequestID(1), Result: mustRaw(map[string]any{})}),
			writeLine(rpc.JSONRPCNotification{Method: "initialized"}),
			writeLine(rpc.JSONRPCRequest{ID: rpc.NewIntRequestID(2), Method: "account/login/start", Params: mustRaw(loginParams)}),
			readLine(rpc.JSONRPCResponse{ID: rpc.NewIntRequestID(2), Result: mustRaw(map[string]any{"type": "chatgpt", "loginId": "login_1", "authUrl": "https://example.test/login"})}),
			writeLine(rpc.JSONRPCRequest{ID: rpc.NewIntRequestID(3), Method: "account/login/cancel", Params: mustRaw(map[string]any{"loginId": "login_1"})}),
			readLine(rpc.JSONRPCResponse{ID: rpc.NewIntRequestID(3), Result: mustRaw(map[string]any{"status": "canceled"})}),
			writeLine(rpc.JSONRPCRequest{ID: rpc.NewIntRequestID(4), Method: "account/logout"}),
			readLine(rpc.JSONRPCResponse{ID: rpc.NewIntRequestID(4), Result: mustRaw(map[string]any{"ok": true})}),
		}),
		ClientInfo: info,
	})
	if err != nil {
		t.Fatalf("new client error: %v", err)
	}
	defer client.Close()

	if _, err := client.StartLogin(ctx, loginParams); err != nil {
		t.Fatalf("start login error: %v", err)
	}
	if _, err := client.CancelLogin(ctx, "login_1"); err != nil {
		t.Fatalf("cancel login error: %v", err)
	}
	if _, err := client.Logout(ctx); err != nil {
		t.Fatalf("logout error: %v", err)
	}
}

func TestLoginHelpersRejectInvalidInputsAndDoNotLeakSecrets(t *testing.T) {
	var logOutput strings.Builder
	client := &Codex{
		logger: slog.New(slog.NewTextHandler(&logOutput, nil)),
		client: rpc.NewClient(rpc.NewReplayTransport(nil), rpc.ClientOptions{}),
	}
	defer client.Client().Close()

	if _, err := client.StartLogin(context.Background(), nil); err == nil {
		t.Fatalf("expected nil login params error")
	}
	if _, err := client.CancelLogin(context.Background(), ""); err == nil {
		t.Fatalf("expected empty login id error")
	}
	if strings.Contains(logOutput.String(), "secret-api-key") {
		t.Fatalf("unexpected secret in logs: %s", logOutput.String())
	}
}

func TestThreadLifecycleHelpersWithReplay(t *testing.T) {
	ctx := context.Background()
	info := protocol.ClientInfo{Name: "codex-go-test", Version: "test"}
	archived := false
	pinned := false
	limit := 10

	client, err := New(ctx, Options{
		Transport: rpc.NewReplayTransport([]rpc.TranscriptEntry{
			writeLine(rpc.JSONRPCRequest{ID: rpc.NewIntRequestID(1), Method: "initialize", Params: mustRaw(protocol.InitializeParams{ClientInfo: info})}),
			readLine(rpc.JSONRPCResponse{ID: rpc.NewIntRequestID(1), Result: mustRaw(map[string]any{})}),
			writeLine(rpc.JSONRPCNotification{Method: "initialized"}),
			writeLine(rpc.JSONRPCRequest{ID: rpc.NewIntRequestID(2), Method: "thread/start", Params: mustRaw(map[string]any{})}),
			readLine(rpc.JSONRPCResponse{ID: rpc.NewIntRequestID(2), Result: mustRaw(map[string]any{"thread": map[string]any{"id": "thr_123"}})}),
			writeLine(rpc.JSONRPCRequest{ID: rpc.NewIntRequestID(3), Method: "thread/list", Params: mustRaw(map[string]any{"archived": false, "isPinned": false, "limit": 10, "searchTerm": "sdk"})}),
			readLine(rpc.JSONRPCResponse{ID: rpc.NewIntRequestID(3), Result: mustRaw(map[string]any{"data": []any{map[string]any{"id": "thr_123", "sessionId": "session_1", "preview": "SDK preview", "isPinned": true, "modelProvider": "openai", "createdAt": 1, "updatedAt": 2, "cwd": "/tmp/project", "cliVersion": "0.144.6", "turns": []any{}}}, "nextCursor": "next_1"})}),
			writeLine(rpc.JSONRPCRequest{ID: rpc.NewIntRequestID(4), Method: "thread/read", Params: mustRaw(map[string]any{"threadId": "thr_123", "includeTurns": true})}),
			readLine(rpc.JSONRPCResponse{ID: rpc.NewIntRequestID(4), Result: mustRaw(map[string]any{"thread": map[string]any{"id": "thr_123", "name": "SDK work", "turns": []any{map[string]any{"id": "turn_1", "items": []any{}, "itemsView": "full", "status": "completed"}}}})}),
			writeLine(rpc.JSONRPCRequest{ID: rpc.NewIntRequestID(5), Method: "thread/name/set", Params: mustRaw(map[string]any{"threadId": "thr_123", "name": "SDK work"})}),
			readLine(rpc.JSONRPCResponse{ID: rpc.NewIntRequestID(5), Result: mustRaw(map[string]any{})}),
			writeLine(rpc.JSONRPCRequest{ID: rpc.NewIntRequestID(6), Method: "thread/archive", Params: mustRaw(map[string]any{"threadId": "thr_123"})}),
			readLine(rpc.JSONRPCResponse{ID: rpc.NewIntRequestID(6), Result: mustRaw(map[string]any{})}),
			writeLine(rpc.JSONRPCRequest{ID: rpc.NewIntRequestID(7), Method: "thread/unarchive", Params: mustRaw(map[string]any{"threadId": "thr_123"})}),
			readLine(rpc.JSONRPCResponse{ID: rpc.NewIntRequestID(7), Result: mustRaw(map[string]any{"thread": map[string]any{"id": "thr_123"}})}),
			writeLine(rpc.JSONRPCRequest{ID: rpc.NewIntRequestID(8), Method: "thread/compact/start", Params: mustRaw(map[string]any{"threadId": "thr_123"})}),
			readLine(rpc.JSONRPCResponse{ID: rpc.NewIntRequestID(8), Result: mustRaw(map[string]any{})}),
			writeLine(rpc.JSONRPCRequest{ID: rpc.NewIntRequestID(9), Method: "thread/fork", Params: mustRaw(map[string]any{"threadId": "thr_123", "model": "gpt-test"})}),
			readLine(rpc.JSONRPCResponse{ID: rpc.NewIntRequestID(9), Result: mustRaw(map[string]any{"thread": map[string]any{"id": "thr_fork"}})}),
		}),
		ClientInfo: info,
	})
	if err != nil {
		t.Fatalf("new client error: %v", err)
	}
	defer client.Close()

	thread, err := client.StartThread(ctx, ThreadStartOptions{})
	if err != nil {
		t.Fatalf("start thread error: %v", err)
	}
	page, err := client.ListThreads(ctx, ThreadListOptions{Archived: &archived, IsPinned: &pinned, Limit: &limit, SearchTerm: "sdk"})
	if err != nil {
		t.Fatalf("list threads error: %v", err)
	}
	if len(page.Data) != 1 || page.Data[0].Preview != "SDK preview" || !page.Data[0].IsPinned || page.NextCursor == nil || *page.NextCursor != "next_1" {
		t.Fatalf("unexpected typed thread page: %#v", page)
	}
	read, err := thread.Read(ctx, ThreadReadOptions{IncludeTurns: true})
	if err != nil {
		t.Fatalf("read thread error: %v", err)
	}
	if len(read.Thread.Turns) != 1 || read.Thread.Turns[0].Status != protocol.TurnStatusCompleted {
		t.Fatalf("unexpected typed thread snapshot: %#v", read)
	}
	if _, err := thread.SetName(ctx, "SDK work"); err != nil {
		t.Fatalf("set name error: %v", err)
	}
	if _, err := thread.Archive(ctx); err != nil {
		t.Fatalf("archive error: %v", err)
	}
	unarchived, err := thread.Unarchive(ctx)
	if err != nil {
		t.Fatalf("unarchive error: %v", err)
	}
	if unarchived.Thread.ID != "thr_123" {
		t.Fatalf("unexpected typed unarchive response: %#v", unarchived)
	}
	if _, err := thread.Compact(ctx, ThreadCompactOptions{}); err != nil {
		t.Fatalf("compact error: %v", err)
	}
	forked, forkResponse, err := thread.Fork(ctx, ThreadForkOptions{Model: "gpt-test"})
	if err != nil {
		t.Fatalf("fork error: %v", err)
	}
	if forked.ID() != "thr_fork" {
		t.Fatalf("expected forked thread id, got %s", forked.ID())
	}
	if forkResponse.Thread == nil || forkResponse.Thread.ID != "thr_fork" {
		t.Fatalf("unexpected typed fork response: %#v", forkResponse)
	}
	if thread.ID() != "thr_123" {
		t.Fatalf("original thread mutated: %s", thread.ID())
	}
}

func TestThreadListOptionsToParamsPreservesExplicitEmptyFilters(t *testing.T) {
	params, err := (ThreadListOptions{}).toParams()
	if err != nil {
		t.Fatalf("to params: %v", err)
	}
	if params.ModelProviders != nil {
		t.Fatalf("expected omitted model providers for zero-value options")
	}
	if params.SourceKinds != nil {
		t.Fatalf("expected omitted source kinds for zero-value options")
	}
	if params.IsPinned != nil {
		t.Fatalf("expected omitted pinned filter for zero-value options")
	}

	pinned := false
	params, err = (ThreadListOptions{
		ModelProviders: []string{},
		SourceKinds:    []protocol.ThreadSourceKind{},
		IsPinned:       &pinned,
		SortDirection:  protocol.SortDirectionAsc,
		SortKey:        protocol.ThreadSortKeyUpdatedAt,
	}).toParams()
	if err != nil {
		t.Fatalf("to params with explicit filters: %v", err)
	}
	if params.ModelProviders == nil {
		t.Fatalf("expected explicit model providers filter")
	}
	if len(*params.ModelProviders) != 0 {
		t.Fatalf("expected empty model providers filter, got %#v", *params.ModelProviders)
	}
	if params.SourceKinds == nil {
		t.Fatalf("expected explicit source kinds filter")
	}
	if len(*params.SourceKinds) != 0 {
		t.Fatalf("expected empty source kinds filter, got %#v", *params.SourceKinds)
	}
	if params.IsPinned == nil || *params.IsPinned {
		t.Fatalf("expected explicit false pinned filter, got %#v", params.IsPinned)
	}
	wire, err := json.Marshal(params)
	if err != nil {
		t.Fatalf("marshal params: %v", err)
	}
	if !strings.Contains(string(wire), `"isPinned":false`) {
		t.Fatalf("pinned filter missing from wire params: %s", wire)
	}
	if params.SortDirection != protocol.SortDirectionAsc || params.SortKey != protocol.ThreadSortKeyUpdatedAt {
		t.Fatalf("typed sort options were not preserved: %#v", params)
	}
	pinned = true
	params, err = (ThreadListOptions{IsPinned: &pinned}).toParams()
	if err != nil {
		t.Fatalf("to params with true pinned filter: %v", err)
	}
	wire, err = json.Marshal(params)
	if err != nil {
		t.Fatalf("marshal true pinned filter: %v", err)
	}
	if !strings.Contains(string(wire), `"isPinned":true`) {
		t.Fatalf("true pinned filter missing from wire params: %s", wire)
	}
	if _, err := (ThreadListOptions{SortDirection: "sideways"}).toParams(); err == nil {
		t.Fatal("invalid sort direction was accepted")
	}
	if _, err := (ThreadListOptions{SortKey: "unknown"}).toParams(); err == nil {
		t.Fatal("invalid sort key was accepted")
	}
}

func TestThreadForkOptionsToParams(t *testing.T) {
	ephemeral := true
	opts := ThreadForkOptions{
		Model:                 "gpt-test",
		ModelProvider:         "openai",
		ServiceTier:           "priority",
		LastTurnID:            "turn_123",
		Cwd:                   "/tmp/project",
		ApprovalPolicy:        "never",
		Sandbox:               map[string]any{"type": "readOnly"},
		Config:                map[string]any{"foo": "bar"},
		BaseInstructions:      "base",
		DeveloperInstructions: "dev",
		Ephemeral:             &ephemeral,
	}

	params, err := opts.toParams("thr_123")
	if err != nil {
		t.Fatalf("to params: %v", err)
	}
	assertEqual(t, "threadId", params.ThreadID, "thr_123")
	assertEqual(t, "model", params.Model, stringPtr("gpt-test"))
	assertEqual(t, "modelProvider", params.ModelProvider, stringPtr("openai"))
	assertEqual(t, "serviceTier", params.ServiceTier, stringPtr("priority"))
	assertEqual(t, "lastTurnId", params.LastTurnID, stringPtr("turn_123"))
	assertEqual(t, "cwd", params.Cwd, stringPtr("/tmp/project"))
	assertRawEqual(t, "approvalPolicy", params.ApprovalPolicy, MustJSON("never"))
	assertRawEqual(t, "sandbox", params.Sandbox, MustJSON(map[string]any{"type": "readOnly"}))
	if params.Config == nil {
		t.Fatalf("expected config")
	}
	assertEqual(t, "config", *params.Config, map[string]any{"foo": "bar"})
	assertEqual(t, "baseInstructions", params.BaseInstructions, stringPtr("base"))
	assertEqual(t, "developerInstructions", params.DeveloperInstructions, stringPtr("dev"))
	assertEqual(t, "ephemeral", params.Ephemeral, &ephemeral)
}

func TestThreadGoalResponsesWithReplay(t *testing.T) {
	ctx := context.Background()
	tokenBudget := int64(100)
	goal := protocol.ThreadGoal{
		ThreadID:        "thr_123",
		Objective:       "ship",
		Status:          protocol.ThreadGoalStatusActive,
		TokenBudget:     &tokenBudget,
		TokensUsed:      10,
		TimeUsedSeconds: 20,
		CreatedAt:       30,
		UpdatedAt:       40,
	}
	client := rpc.NewClient(rpc.NewReplayTransport([]rpc.TranscriptEntry{
		writeLine(rpc.JSONRPCRequest{ID: rpc.NewIntRequestID(1), Method: "thread/goal/set", Params: mustRaw(protocol.ThreadGoalSetParams{ThreadID: "thr_123"})}),
		readLine(rpc.JSONRPCResponse{ID: rpc.NewIntRequestID(1), Result: mustRaw(map[string]any{"goal": goal})}),
		writeLine(rpc.JSONRPCRequest{ID: rpc.NewIntRequestID(2), Method: "thread/goal/get", Params: mustRaw(protocol.ThreadGoalGetParams{ThreadID: "thr_123"})}),
		readLine(rpc.JSONRPCResponse{ID: rpc.NewIntRequestID(2), Result: mustRaw(map[string]any{"goal": goal})}),
	}), rpc.ClientOptions{})
	defer client.Close()

	setResp, err := client.ThreadGoalSet(ctx, protocol.ThreadGoalSetParams{ThreadID: "thr_123"})
	if err != nil {
		t.Fatalf("set goal: %v", err)
	}
	assertEqual(t, "set goal", setResp.Goal, goal)

	getResp, err := client.ThreadGoalGet(ctx, protocol.ThreadGoalGetParams{ThreadID: "thr_123"})
	if err != nil {
		t.Fatalf("get goal: %v", err)
	}
	if getResp.Goal == nil {
		t.Fatalf("expected goal")
	}
	assertEqual(t, "get goal", *getResp.Goal, goal)
}

func TestThreadLifecycleRejectsInvalidInputs(t *testing.T) {
	ctx := context.Background()
	c := &Codex{logger: slog.New(slog.NewTextHandler(&strings.Builder{}, nil))}
	if _, err := c.ListThreads(ctx, ThreadListOptions{}); err == nil {
		t.Fatalf("expected uninitialized client error")
	}
	if _, err := c.ReadThread(ctx, "", ThreadReadOptions{}); err == nil {
		t.Fatalf("expected read thread id error")
	}

	thread := &Thread{client: rpc.NewClient(rpc.NewReplayTransport(nil), rpc.ClientOptions{}), id: "thr_123"}
	defer thread.client.Close()
	if _, err := thread.SetName(ctx, ""); err == nil {
		t.Fatalf("expected empty name error")
	}
	if _, _, err := thread.Fork(ctx, ThreadForkOptions{ApprovalPolicy: json.RawMessage("{bad")}); err == nil {
		t.Fatalf("expected invalid approval policy error")
	}

	var nilThread *Thread
	if _, err := nilThread.Archive(ctx); err == nil {
		t.Fatalf("expected nil thread error")
	}
}

func TestThreadLifecyclePropagatesServerErrors(t *testing.T) {
	ctx := context.Background()
	info := protocol.ClientInfo{Name: "codex-go-test", Version: "test"}
	client, err := New(ctx, Options{
		Transport: rpc.NewReplayTransport([]rpc.TranscriptEntry{
			writeLine(rpc.JSONRPCRequest{ID: rpc.NewIntRequestID(1), Method: "initialize", Params: mustRaw(protocol.InitializeParams{ClientInfo: info})}),
			readLine(rpc.JSONRPCResponse{ID: rpc.NewIntRequestID(1), Result: mustRaw(map[string]any{})}),
			writeLine(rpc.JSONRPCNotification{Method: "initialized"}),
			writeLine(rpc.JSONRPCRequest{ID: rpc.NewIntRequestID(2), Method: "thread/list", Params: mustRaw(map[string]any{})}),
			readLine(rpc.JSONRPCError{ID: rpc.NewIntRequestID(2), Error: rpc.JSONRPCErrorDetail{Code: -32000, Message: "boom"}}),
		}),
		ClientInfo: info,
	})
	if err != nil {
		t.Fatalf("new client error: %v", err)
	}
	defer client.Close()

	_, err = client.ListThreads(ctx, ThreadListOptions{})
	var responseErr *rpc.ResponseError
	if !errors.As(err, &responseErr) {
		t.Fatalf("expected response error, got %v", err)
	}
}
