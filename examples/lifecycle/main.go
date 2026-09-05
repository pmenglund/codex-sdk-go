package main

import (
	"context"
	"encoding/json"
	"fmt"
	"log/slog"
	"os"

	"github.com/pmenglund/codex-sdk-go"
	"github.com/pmenglund/codex-sdk-go/protocol"
	"github.com/pmenglund/codex-sdk-go/rpc"
)

func main() {
	ctx := context.Background()
	logger := slog.New(slog.NewTextHandler(os.Stderr, &slog.HandlerOptions{Level: slog.LevelInfo}))

	client, err := codex.New(ctx, exampleOptions(logger))
	if err != nil {
		panic(err)
	}
	defer client.Close()

	account, err := client.Account(ctx, codex.AccountOptions{})
	if err != nil {
		panic(err)
	}

	_, err = client.ListThreads(ctx, codex.ThreadListOptions{})
	if err != nil {
		panic(err)
	}

	thread, err := client.StartThread(ctx, codex.ThreadStartOptions{})
	if err != nil {
		panic(err)
	}
	if _, err := thread.SetName(ctx, "SDK lifecycle example"); err != nil {
		panic(err)
	}

	forked, _, err := thread.Fork(ctx, codex.ThreadForkOptions{})
	if err != nil {
		panic(err)
	}

	handle, err := forked.StartTurn(ctx, []codex.Input{codex.TextInput("Inspect the repo")}, nil)
	if err != nil {
		panic(err)
	}
	if _, err := handle.Steer(ctx, []codex.Input{codex.TextInput("Focus on public API ergonomics")}); err != nil {
		panic(err)
	}

	result, err := handle.Run(ctx)
	if err != nil {
		panic(err)
	}

	fmt.Printf("requires_auth=%t fork=%s final=%s\n", account.RequiresOpenaiAuth, forked.ID(), result.FinalResponse)
}

const exampleReplayEnv = "CODEX_EXAMPLE_REPLAY"

func exampleOptions(logger *slog.Logger) codex.Options {
	if os.Getenv(exampleReplayEnv) == "" {
		return codex.Options{Logger: logger}
	}

	info := exampleClientInfo()
	return codex.Options{
		Transport:  rpc.NewReplayTransport(exampleTranscript(info)),
		ClientInfo: info,
	}
}

func exampleClientInfo() protocol.ClientInfo {
	return protocol.ClientInfo{
		Name:    "codex-go-example",
		Title:   stringPtr("Codex Go SDK Example"),
		Version: "test",
	}
}

func exampleTranscript(info protocol.ClientInfo) []rpc.TranscriptEntry {
	return []rpc.TranscriptEntry{
		writeLine(rpc.JSONRPCRequest{ID: rpc.NewIntRequestID(1), Method: "initialize", Params: mustRaw(protocol.InitializeParams{ClientInfo: info})}),
		readLine(rpc.JSONRPCResponse{ID: rpc.NewIntRequestID(1), Result: mustRaw(map[string]any{})}),
		writeLine(rpc.JSONRPCNotification{Method: "initialized"}),
		writeLine(rpc.JSONRPCRequest{ID: rpc.NewIntRequestID(2), Method: "account/read", Params: mustRaw(map[string]any{})}),
		readLine(rpc.JSONRPCResponse{ID: rpc.NewIntRequestID(2), Result: mustRaw(map[string]any{"requiresOpenaiAuth": false})}),
		writeLine(rpc.JSONRPCRequest{ID: rpc.NewIntRequestID(3), Method: "thread/list", Params: mustRaw(map[string]any{})}),
		readLine(rpc.JSONRPCResponse{ID: rpc.NewIntRequestID(3), Result: mustRaw(map[string]any{"data": []any{}})}),
		writeLine(rpc.JSONRPCRequest{ID: rpc.NewIntRequestID(4), Method: "thread/start", Params: mustRaw(map[string]any{})}),
		readLine(rpc.JSONRPCResponse{ID: rpc.NewIntRequestID(4), Result: mustRaw(map[string]any{"thread": map[string]any{"id": "thr_123"}})}),
		writeLine(rpc.JSONRPCRequest{ID: rpc.NewIntRequestID(5), Method: "thread/name/set", Params: mustRaw(map[string]any{"threadId": "thr_123", "name": "SDK lifecycle example"})}),
		readLine(rpc.JSONRPCResponse{ID: rpc.NewIntRequestID(5), Result: mustRaw(map[string]any{})}),
		writeLine(rpc.JSONRPCRequest{ID: rpc.NewIntRequestID(6), Method: "thread/fork", Params: mustRaw(map[string]any{"threadId": "thr_123"})}),
		readLine(rpc.JSONRPCResponse{ID: rpc.NewIntRequestID(6), Result: mustRaw(map[string]any{"thread": map[string]any{"id": "thr_fork"}})}),
		writeLine(rpc.JSONRPCRequest{ID: rpc.NewIntRequestID(7), Method: "turn/start", Params: mustRaw(map[string]any{"threadId": "thr_fork", "input": []codex.Input{codex.TextInput("Inspect the repo")}})}),
		readLine(rpc.JSONRPCResponse{ID: rpc.NewIntRequestID(7), Result: mustRaw(map[string]any{"turn": turnPayload("turn_1", "inProgress")})}),
		writeLine(rpc.JSONRPCRequest{ID: rpc.NewIntRequestID(8), Method: "turn/steer", Params: mustRaw(map[string]any{"threadId": "thr_fork", "expectedTurnId": "turn_1", "input": []codex.Input{codex.TextInput("Focus on public API ergonomics")}})}),
		readLine(rpc.JSONRPCResponse{ID: rpc.NewIntRequestID(8), Result: mustRaw(map[string]any{})}),
		readLine(rpc.JSONRPCNotification{Method: "turn/started", Params: mustRaw(map[string]any{"threadId": "thr_fork", "turn": turnPayload("turn_1", "inProgress")})}),
		readLine(rpc.JSONRPCNotification{Method: "item/completed", Params: mustRaw(map[string]any{"threadId": "thr_fork", "turnId": "turn_fork", "completedAtMs": 1, "item": map[string]any{"type": "agentMessage", "id": "item_1", "text": "Lifecycle replay complete"}})}),
		readLine(rpc.JSONRPCNotification{Method: "turn/completed", Params: mustRaw(map[string]any{"threadId": "thr_fork", "turn": turnPayload("turn_1", "completed")})}),
	}
}

func turnPayload(turnID, status string) map[string]any {
	return map[string]any{
		"id":     turnID,
		"status": status,
		"items":  []any{},
		"error":  nil,
	}
}

func writeLine(payload any) rpc.TranscriptEntry {
	return rpc.TranscriptEntry{Direction: rpc.TranscriptWrite, Line: mustJSON(payload)}
}

func readLine(payload any) rpc.TranscriptEntry {
	return rpc.TranscriptEntry{Direction: rpc.TranscriptRead, Line: mustJSON(payload)}
}

func mustJSON(payload any) string {
	data, err := json.Marshal(payload)
	if err != nil {
		panic(err)
	}
	return string(data)
}

func mustRaw(payload any) json.RawMessage {
	if payload == nil {
		return nil
	}
	data, err := json.Marshal(payload)
	if err != nil {
		panic(err)
	}
	return data
}

func stringPtr(value string) *string {
	return &value
}
