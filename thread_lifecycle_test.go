package codex

import (
	"context"
	"encoding/json"
	"errors"
	"strings"
	"testing"

	"github.com/pmenglund/codex-sdk-go/protocol"
	"github.com/pmenglund/codex-sdk-go/rpc"
)

func lifecycleReplay(t *testing.T, method string, params, result any) (*Codex, *Thread) {
	t.Helper()
	id := rpc.NewIntRequestID(1)
	reply := readLine(rpc.JSONRPCResponse{ID: id, Result: mustRaw(result)})
	if detail, ok := result.(rpc.JSONRPCErrorDetail); ok {
		reply = readLine(rpc.JSONRPCError{ID: id, Error: detail})
	}
	client := rpc.NewClient(rpc.NewReplayTransport([]rpc.TranscriptEntry{
		writeLine(rpc.JSONRPCRequest{ID: id, Method: method, Params: mustRaw(params)}),
		reply,
	}), rpc.ClientOptions{})
	t.Cleanup(func() { _ = client.Close() })
	return &Codex{client: client}, &Thread{client: client, id: "thr_1"}
}

func TestDirectLifecycleRequests(t *testing.T) {
	ctx := context.Background()
	for _, tt := range []struct {
		method string
		params map[string]any
		result map[string]any
		call   func(*testing.T, *Codex) error
	}{
		{"thread/read", map[string]any{"threadId": "thr_1", "includeTurns": true}, map[string]any{"thread": map[string]any{"id": "thr_1"}}, func(t *testing.T, c *Codex) error {
			r, err := c.ReadThread(ctx, "thr_1", ThreadReadOptions{IncludeTurns: true})
			if err == nil && r.Thread.ID != "thr_1" {
				t.Errorf("read returned thread %q", r.Thread.ID)
			}
			return err
		}},
		{"thread/name/set", map[string]any{"threadId": "thr_1", "name": "Work"}, map[string]any{}, func(t *testing.T, c *Codex) error {
			_, err := c.SetThreadName(ctx, "thr_1", "Work")
			return err
		}},
		{"thread/archive", map[string]any{"threadId": "thr_1"}, map[string]any{}, func(t *testing.T, c *Codex) error {
			_, err := c.ArchiveThread(ctx, "thr_1")
			return err
		}},
		{"thread/unarchive", map[string]any{"threadId": "thr_1"}, map[string]any{"thread": map[string]any{"id": "thr_1"}}, func(t *testing.T, c *Codex) error {
			r, err := c.UnarchiveThread(ctx, "thr_1")
			if err == nil && r.Thread.ID != "thr_1" {
				t.Errorf("unarchive returned thread %q", r.Thread.ID)
			}
			return err
		}},
		{"thread/compact/start", map[string]any{"threadId": "thr_1"}, map[string]any{}, func(t *testing.T, c *Codex) error {
			_, err := c.CompactThread(ctx, "thr_1", ThreadCompactOptions{})
			return err
		}},
	} {
		t.Run(tt.method, func(t *testing.T) {
			c, _ := lifecycleReplay(t, tt.method, tt.params, tt.result)
			if err := tt.call(t, c); err != nil {
				t.Fatal(err)
			}
		})
	}
}

func TestLifecycleReadinessAndInputValidation(t *testing.T) {
	ctx := context.Background()
	for _, tt := range []struct {
		name string
		call func(*Codex, *Thread) error
	}{
		{"read", func(c *Codex, th *Thread) error {
			if c != nil {
				_, e := c.ReadThread(ctx, "thr_1", ThreadReadOptions{})
				return e
			}
			_, e := th.Read(ctx, ThreadReadOptions{})
			return e
		}},
		{"name", func(c *Codex, th *Thread) error {
			if c != nil {
				_, e := c.SetThreadName(ctx, "thr_1", "Work")
				return e
			}
			_, e := th.SetName(ctx, "Work")
			return e
		}},
		{"archive", func(c *Codex, th *Thread) error {
			if c != nil {
				_, e := c.ArchiveThread(ctx, "thr_1")
				return e
			}
			_, e := th.Archive(ctx)
			return e
		}},
		{"unarchive", func(c *Codex, th *Thread) error {
			if c != nil {
				_, e := c.UnarchiveThread(ctx, "thr_1")
				return e
			}
			_, e := th.Unarchive(ctx)
			return e
		}},
		{"compact", func(c *Codex, th *Thread) error {
			if c != nil {
				_, e := c.CompactThread(ctx, "thr_1", ThreadCompactOptions{})
				return e
			}
			_, e := th.Compact(ctx, ThreadCompactOptions{})
			return e
		}},
		{"fork", func(c *Codex, th *Thread) error {
			if c != nil {
				_, _, e := c.ForkThread(ctx, "thr_1", ThreadForkOptions{})
				return e
			}
			_, _, e := th.Fork(ctx, ThreadForkOptions{})
			return e
		}},
	} {
		t.Run(tt.name, func(t *testing.T) {
			if err := tt.call(&Codex{}, nil); err == nil || !strings.Contains(err.Error(), "not initialized") {
				t.Errorf("unready client: %v", err)
			}
			if err := tt.call(nil, nil); err == nil || err.Error() != "thread is nil" {
				t.Errorf("nil thread: %v", err)
			}
			if err := tt.call(nil, &Thread{}); err == nil || !strings.Contains(err.Error(), "not initialized") {
				t.Errorf("unready thread: %v", err)
			}
		})
	}
	client := rpc.NewClient(rpc.NewReplayTransport(nil), rpc.ClientOptions{})
	defer client.Close()
	c := &Codex{client: client}
	for _, call := range []func() error{
		func() error { _, e := c.ReadThread(ctx, "", ThreadReadOptions{}); return e },
		func() error { _, e := c.SetThreadName(ctx, "", "Work"); return e },
		func() error { _, e := c.ArchiveThread(ctx, ""); return e },
		func() error { _, e := c.UnarchiveThread(ctx, ""); return e },
		func() error { _, e := c.CompactThread(ctx, "", ThreadCompactOptions{}); return e },
		func() error { _, _, e := c.ForkThread(ctx, "", ThreadForkOptions{}); return e },
	} {
		if err := call(); err == nil || err.Error() != "thread id is required" {
			t.Errorf("missing id: %v", err)
		}
	}
	if _, err := c.SetThreadName(ctx, "thr_1", ""); err == nil || err.Error() != "thread name is required" {
		t.Errorf("missing name: %v", err)
	}
	if _, err := c.ListThreads(ctx, ThreadListOptions{SortKey: "invalid"}); err == nil || !strings.Contains(err.Error(), "invalid thread sort key") {
		t.Errorf("invalid list options: %v", err)
	}
	for _, opts := range []ThreadForkOptions{{ApprovalPolicy: json.RawMessage("{")}, {Sandbox: json.RawMessage("{")}} {
		if _, _, err := c.ForkThread(ctx, "thr_1", opts); err == nil || !strings.Contains(err.Error(), "invalid raw JSON") {
			t.Errorf("invalid fork options: %v", err)
		}
	}
}

func TestForkResponsesAndFailures(t *testing.T) {
	ctx := context.Background()
	for _, viaThread := range []bool{false, true} {
		for _, tt := range []struct {
			name      string
			result    any
			wantID    string
			wantError string
		}{
			{"nested id", map[string]any{"thread": map[string]any{"id": "fork_1"}}, "fork_1", ""},
			{"legacy id", map[string]any{"threadId": "fork_2"}, "fork_2", ""},
			{"missing id", map[string]any{"model": "reported-model"}, "", "thread id not found"},
			{"server error", rpc.JSONRPCErrorDetail{Code: -32000, Message: "fork denied"}, "", "fork denied"},
		} {
			label := "client/"
			if viaThread {
				label = "thread/"
			}
			t.Run(label+tt.name, func(t *testing.T) {
				exclude := true
				c, th := lifecycleReplay(t, "thread/fork", map[string]any{"threadId": "thr_1", "excludeTurns": true}, tt.result)
				var fork *Thread
				var response protocol.ThreadForkResponse
				var err error
				if viaThread {
					fork, response, err = th.Fork(ctx, ThreadForkOptions{ExcludeTurns: &exclude})
				} else {
					fork, response, err = c.ForkThread(ctx, "thr_1", ThreadForkOptions{ExcludeTurns: &exclude})
				}
				if tt.wantError != "" {
					if err == nil || !strings.Contains(err.Error(), tt.wantError) || fork != nil {
						t.Fatalf("fork=%v err=%v", fork, err)
					}
					if tt.name == "server error" {
						var re *rpc.ResponseError
						if !errors.As(err, &re) || re.Detail.Code != -32000 {
							t.Fatalf("lost server error: %v", err)
						}
					}
					if tt.name == "missing id" && response.Model != "reported-model" {
						t.Fatalf("discarded partial response: %#v", response)
					}
				} else if err != nil || fork == nil || fork.ID() != tt.wantID {
					t.Fatalf("fork=%v err=%v", fork, err)
				}
			})
		}
	}
}

func TestStartAndResumeFailures(t *testing.T) {
	ctx := context.Background()
	for _, tt := range []struct {
		method string
		params map[string]any
		call   func(*Codex, bool) (*Thread, error)
	}{
		{"thread/start", map[string]any{}, func(c *Codex, invalid bool) (*Thread, error) {
			opts := ThreadStartOptions{}
			if invalid {
				opts.ApprovalPolicy = json.RawMessage("{")
			}
			return c.StartThread(ctx, opts)
		}},
		{"thread/resume", map[string]any{"threadId": "thr_1"}, func(c *Codex, invalid bool) (*Thread, error) {
			opts := ThreadResumeOptions{ThreadID: "thr_1"}
			if invalid {
				opts.ApprovalPolicy = json.RawMessage("{")
			}
			return c.ResumeThread(ctx, opts)
		}},
	} {
		t.Run(tt.method, func(t *testing.T) {
			for _, unready := range []*Codex{nil, {}} {
				if th, err := tt.call(unready, false); err == nil || th != nil {
					t.Fatalf("unready client returned thread=%v, err=%v", th, err)
				}
			}
			c, _ := lifecycleReplay(t, tt.method, tt.params, map[string]any{})
			if th, err := tt.call(c, true); err == nil || th != nil || !strings.Contains(err.Error(), "invalid raw JSON") {
				t.Fatalf("invalid options returned thread=%v, err=%v", th, err)
			}
			for _, result := range []any{map[string]any{}, rpc.JSONRPCErrorDetail{Code: -32000, Message: "server rejected thread"}} {
				c, _ := lifecycleReplay(t, tt.method, tt.params, result)
				th, err := tt.call(c, false)
				if th != nil || err == nil {
					t.Fatalf("failed creation returned thread=%v, err=%v", th, err)
				}
				if _, serverError := result.(rpc.JSONRPCErrorDetail); serverError {
					var re *rpc.ResponseError
					if !errors.As(err, &re) || re.Detail.Code != -32000 {
						t.Fatalf("lost server failure: %v", err)
					}
				} else if !strings.Contains(err.Error(), "thread id not found") {
					t.Fatalf("missing id: %v", err)
				}
			}
		})
	}
}
