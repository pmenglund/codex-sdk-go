package codex

import (
	"context"
	"encoding/json"
	"errors"
	"testing"
	"testing/synctest"

	"github.com/pmenglund/codex-sdk-go/protocol"
	"github.com/pmenglund/codex-sdk-go/rpc"
)

type lifecycleOperation struct {
	name, method   string
	params, result map[string]any
	call           func(*Codex, string) (any, error)
}

func lifecycleOperations() []lifecycleOperation {
	return []lifecycleOperation{
		{"read", "thread/read", map[string]any{"threadId": "source"}, map[string]any{"thread": map[string]any{"id": "source"}}, func(c *Codex, id string) (any, error) {
			return c.ReadThread(context.Background(), id, ThreadReadOptions{})
		}},
		{"read_with_turns", "thread/read", map[string]any{"threadId": "source", "includeTurns": true}, map[string]any{"thread": map[string]any{"id": "source"}}, func(c *Codex, id string) (any, error) {
			return c.ReadThread(context.Background(), id, ThreadReadOptions{IncludeTurns: true})
		}},
		{"name", "thread/name/set", map[string]any{"threadId": "source", "name": "New name"}, map[string]any{}, func(c *Codex, id string) (any, error) { return c.SetThreadName(context.Background(), id, "New name") }},
		{"archive", "thread/archive", map[string]any{"threadId": "source"}, map[string]any{}, func(c *Codex, id string) (any, error) { return c.ArchiveThread(context.Background(), id) }},
		{"unarchive", "thread/unarchive", map[string]any{"threadId": "source"}, map[string]any{"thread": map[string]any{"id": "source"}}, func(c *Codex, id string) (any, error) { return c.UnarchiveThread(context.Background(), id) }},
		{"compact", "thread/compact/start", map[string]any{"threadId": "source"}, map[string]any{}, func(c *Codex, id string) (any, error) {
			return c.CompactThread(context.Background(), id, ThreadCompactOptions{})
		}},
		{"fork", "thread/fork", map[string]any{"threadId": "source", "model": "test-model"}, map[string]any{"thread": map[string]any{"id": "forked"}}, func(c *Codex, id string) (any, error) {
			thread, _, err := c.ForkThread(context.Background(), id, ThreadForkOptions{Model: "test-model"})
			return thread, err
		}},
	}
}

func TestPublicThreadLifecycleRequestsAndServerErrors(t *testing.T) {
	for _, operation := range lifecycleOperations() {
		for _, serverError := range []bool{false, true} {
			suffix := "success"
			if serverError {
				suffix = "server_error"
			}
			t.Run(operation.name+"/"+suffix, func(t *testing.T) {
				synctest.Test(t, func(t *testing.T) {
					id := rpc.NewIntRequestID(1)
					reply := readLine(rpc.JSONRPCResponse{ID: id, Result: mustRaw(operation.result)})
					if serverError {
						reply = readLine(rpc.JSONRPCError{ID: id, Error: rpc.JSONRPCErrorDetail{Code: -32001, Message: "lifecycle rejected", Data: mustRaw(map[string]any{"reason": "fixture"})}})
					}
					c := &Codex{client: rpc.NewClient(rpc.NewReplayTransport([]rpc.TranscriptEntry{
						writeLine(rpc.JSONRPCRequest{ID: id, Method: operation.method, Params: mustRaw(operation.params)}), reply,
					}), rpc.ClientOptions{})}
					defer c.Close()
					response, err := operation.call(c, "source")
					if serverError {
						var responseErr *rpc.ResponseError
						if !errors.As(err, &responseErr) || responseErr.Detail.Code != -32001 || responseErr.Detail.Message != "lifecycle rejected" || string(responseErr.Detail.Data) != `{"reason":"fixture"}` {
							t.Fatalf("server error not preserved: %v", err)
						}
						return
					}
					if err != nil {
						t.Fatal(err)
					}
					switch response := response.(type) {
					case *protocol.ThreadReadResponse:
						if response == nil || response.Thread.ID != "source" {
							t.Fatalf("read response = %#v", response)
						}
					case *protocol.ThreadUnarchiveResponse:
						if response == nil || response.Thread.ID != "source" {
							t.Fatalf("unarchive response = %#v", response)
						}
					case *protocol.ThreadSetNameResponse:
						if response == nil {
							t.Fatal("nil name response")
						}
					case *protocol.ThreadArchiveResponse:
						if response == nil {
							t.Fatal("nil archive response")
						}
					case *protocol.ThreadCompactStartResponse:
						if response == nil {
							t.Fatal("nil compact response")
						}
					case *Thread:
						if response == nil || response.ID() != "forked" || response.client != c.client {
							t.Fatalf("forked thread = %#v", response)
						}
					default:
						t.Fatalf("unexpected response type %T", response)
					}
				})
			})
		}
	}
}

func TestPublicThreadLifecycleValidation(t *testing.T) {
	for _, operation := range lifecycleOperations() {
		t.Run(operation.name, func(t *testing.T) {
			synctest.Test(t, func(t *testing.T) {
				for _, c := range []*Codex{nil, {}} {
					if _, err := operation.call(c, "source"); err == nil {
						t.Fatal("uninitialized client accepted")
					}
				}
				recorder := rpc.NewRecordTransport(rpc.NewReplayTransport(nil))
				c := &Codex{client: rpc.NewClient(recorder, rpc.ClientOptions{})}
				defer c.Close()
				if _, err := operation.call(c, ""); err == nil || err.Error() != "thread id is required" {
					t.Fatalf("empty ID error = %v", err)
				}
				if len(recorder.Transcript()) != 0 {
					t.Fatal("invalid input produced RPC traffic")
				}
			})
		})
	}
	synctest.Test(t, func(t *testing.T) {
		recorder := rpc.NewRecordTransport(rpc.NewReplayTransport(nil))
		c := &Codex{client: rpc.NewClient(recorder, rpc.ClientOptions{})}
		defer c.Close()
		if _, err := c.SetThreadName(context.Background(), "source", ""); err == nil || err.Error() != "thread name is required" {
			t.Fatalf("empty name error = %v", err)
		}
		if _, _, err := c.ForkThread(context.Background(), "source", ThreadForkOptions{ApprovalPolicy: json.RawMessage("{bad")}); err == nil {
			t.Fatal("invalid fork policy accepted")
		}
		if len(recorder.Transcript()) != 0 {
			t.Fatal("invalid options produced RPC traffic")
		}
	})
}

func TestPublicForkThreadResponseIDs(t *testing.T) {
	for _, test := range []struct {
		name   string
		result map[string]any
		want   string
	}{
		{"top_level", map[string]any{"threadId": "top"}, "top"},
		{"nested", map[string]any{"thread": map[string]any{"id": "nested"}}, "nested"},
		{"precedence", map[string]any{"threadId": "top", "thread": map[string]any{"id": "nested"}}, "top"},
		{"missing", map[string]any{}, ""},
		{"empty_nested", map[string]any{"thread": map[string]any{"id": ""}}, ""},
	} {
		t.Run(test.name, func(t *testing.T) {
			synctest.Test(t, func(t *testing.T) {
				c := &Codex{client: rpc.NewClient(rpc.NewReplayTransport([]rpc.TranscriptEntry{
					writeLine(rpc.JSONRPCRequest{ID: rpc.NewIntRequestID(1), Method: "thread/fork", Params: mustRaw(map[string]any{"threadId": "source"})}),
					readLine(rpc.JSONRPCResponse{ID: rpc.NewIntRequestID(1), Result: mustRaw(test.result)}),
				}), rpc.ClientOptions{})}
				defer c.Close()
				thread, response, err := c.ForkThread(context.Background(), "source", ThreadForkOptions{})
				if test.want == "" {
					if err == nil || err.Error() != "thread id not found in response" || thread != nil {
						t.Fatalf("missing ID result: thread=%v err=%v", thread, err)
					}
				} else if err != nil || thread == nil || thread.ID() != test.want || thread.client != c.client || thread.logger != c.logger {
					t.Fatalf("fork result: thread=%v err=%v", thread, err)
				}
				// Even an invalid fork result must preserve the server's typed response.
				if _, present := test.result["thread"]; present && response.Thread == nil {
					t.Fatal("nested server response was discarded")
				}
				if top, ok := test.result["threadId"].(string); ok && response.ThreadID != top {
					t.Fatal("top-level server ID was discarded")
				}
			})
		})
	}
}
