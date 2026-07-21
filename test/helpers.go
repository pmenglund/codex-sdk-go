package test

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"os"
	"os/exec"
	"strings"
	"sync"
	"testing"
	"time"

	codex "github.com/pmenglund/codex-sdk-go"
	"github.com/pmenglund/codex-sdk-go/protocol"
	"github.com/pmenglund/codex-sdk-go/rpc"
)

// LockedBuffer is a concurrency-safe buffer for app-server stderr capture.
type LockedBuffer struct {
	mu      sync.Mutex
	buf     bytes.Buffer
	secrets []string
}

func (b *LockedBuffer) Write(p []byte) (int, error) {
	b.mu.Lock()
	defer b.mu.Unlock()
	return b.buf.Write(p)
}

func (b *LockedBuffer) String() string {
	b.mu.Lock()
	defer b.mu.Unlock()
	output := b.buf.String()
	for _, secret := range b.secrets {
		if secret != "" {
			output = strings.ReplaceAll(output, secret, "[REDACTED]")
		}
	}
	return output
}

func (b *LockedBuffer) rawString() string {
	b.mu.Lock()
	defer b.mu.Unlock()
	return b.buf.String()
}

// RealClientOptions configures a real codex app-server test client.
type RealClientOptions struct {
	Timeout          time.Duration
	DisableAutoClose bool
	Secrets          []string
	RequestRecorder  *RequestRecorder
}

// RequestRecorder captures outbound JSON-RPC requests without changing them.
type RequestRecorder struct {
	mu    sync.Mutex
	lines []string
}

// Request returns the first captured request with method.
func (r *RequestRecorder) Request(method string) (rpc.JSONRPCRequest, bool) {
	if r == nil {
		return rpc.JSONRPCRequest{}, false
	}
	r.mu.Lock()
	defer r.mu.Unlock()
	for _, line := range r.lines {
		var request rpc.JSONRPCRequest
		if json.Unmarshal([]byte(line), &request) == nil && request.Method == method {
			return request, true
		}
	}
	return rpc.JSONRPCRequest{}, false
}

type recordingTransport struct {
	transport rpc.Transport
	recorder  *RequestRecorder
}

func (t *recordingTransport) ReadLine() (string, error) { return t.transport.ReadLine() }

func (t *recordingTransport) WriteLine(line string) error {
	t.recorder.mu.Lock()
	t.recorder.lines = append(t.recorder.lines, line)
	t.recorder.mu.Unlock()
	return t.transport.WriteLine(line)
}

func (t *recordingTransport) Close() error { return t.transport.Close() }

// NewRealClient starts a real codex app-server with an isolated CODEX_HOME.
func NewRealClient(t testing.TB, opts RealClientOptions) (*codex.Codex, context.Context, *LockedBuffer) {
	t.Helper()

	codexPath, err := exec.LookPath("codex")
	if err != nil {
		t.Fatalf("codex must be available on PATH for e2e tests: %v", err)
	}

	codexHome, err := os.MkdirTemp("", "codex-sdk-go-e2e-")
	if err != nil {
		t.Fatalf("create temporary CODEX_HOME: %v", err)
	}
	t.Cleanup(func() { removeTemporaryCodexHome(t, codexHome) })
	t.Setenv("CODEX_HOME", codexHome)

	timeout := opts.Timeout
	if timeout == 0 {
		timeout = 20 * time.Second
	}
	ctx, cancel := context.WithTimeout(context.Background(), timeout)
	t.Cleanup(cancel)

	stderr := LockedBuffer{secrets: secretMarkers(opts.Secrets...)}
	clientOptions := codex.Options{Spawn: codex.SpawnOptions{CodexPath: codexPath, Stderr: &stderr}}
	if opts.RequestRecorder != nil {
		transport, spawnErr := rpc.SpawnStdio(context.WithoutCancel(ctx), codexPath, []string{"app-server"}, &stderr)
		if spawnErr != nil {
			t.Fatalf("spawn real codex app-server: %v\nstderr:\n%s", spawnErr, stderr.String())
		}
		clientOptions.Transport = &recordingTransport{transport: transport, recorder: opts.RequestRecorder}
	}
	client, err := codex.New(ctx, clientOptions)
	if err != nil {
		t.Fatalf("initialize real codex app-server: %v\nstderr:\n%s", err, stderr.String())
	}
	t.Cleanup(func() { AssertNoSecretLeak(t, &stderr, opts.Secrets...) })
	if !opts.DisableAutoClose {
		t.Cleanup(func() {
			if err := client.Close(); err != nil {
				t.Errorf("close real codex app-server: %v\nstderr:\n%s", err, stderr.String())
			}
		})
	}
	return client, ctx, &stderr
}

func removeTemporaryCodexHome(t testing.TB, path string) {
	t.Helper()
	deadline := time.Now().Add(5 * time.Second)
	for {
		err := os.RemoveAll(path)
		if err == nil {
			return
		}
		if time.Now().After(deadline) {
			t.Errorf("remove temporary CODEX_HOME %s: %v", path, err)
			return
		}
		time.Sleep(25 * time.Millisecond)
	}
}

func secretMarkers(values ...string) []string {
	seen := make(map[string]struct{})
	var markers []string
	add := func(value string) {
		if value == "" {
			return
		}
		if _, ok := seen[value]; ok {
			return
		}
		seen[value] = struct{}{}
		markers = append(markers, value)
	}
	var visit func(string, any)
	visit = func(key string, value any) {
		switch value := value.(type) {
		case string:
			lowerKey := strings.ToLower(key)
			if strings.Contains(lowerKey, "key") || strings.Contains(lowerKey, "token") || strings.Contains(lowerKey, "secret") || strings.Contains(lowerKey, "password") {
				add(value)
			}
		case []any:
			for _, item := range value {
				visit(key, item)
			}
		case map[string]any:
			for itemKey, item := range value {
				visit(itemKey, item)
			}
		}
	}
	for _, value := range values {
		add(value)
		var decoded any
		if json.Unmarshal([]byte(value), &decoded) == nil {
			visit("", decoded)
		}
	}
	return markers
}

// StartThread starts a real Codex thread and fails the test if no id is returned.
func StartThread(t testing.TB, client *codex.Codex, ctx context.Context, stderr *LockedBuffer, cwd string) *codex.Thread {
	t.Helper()

	thread, err := client.StartThread(ctx, codex.ThreadStartOptions{Cwd: cwd})
	if err != nil {
		t.Fatalf("start real codex thread: %v\nstderr:\n%s", err, stderr.String())
	}
	if thread.ID() == "" {
		t.Fatalf("expected real codex thread id\nstderr:\n%s", stderr.String())
	}
	return thread
}

// RequireLoginParams reads CODEX_E2E_LOGIN_PARAMS_JSON or skips the test.
func RequireLoginParams(t testing.TB) (any, string) {
	t.Helper()

	raw := os.Getenv("CODEX_E2E_LOGIN_PARAMS_JSON")
	if raw == "" {
		t.Skip("set CODEX_E2E_LOGIN_PARAMS_JSON to run credential-dependent live e2e tests")
	}
	decoder := json.NewDecoder(strings.NewReader(raw))
	decoder.UseNumber()
	var params any
	if err := decoder.Decode(&params); err != nil {
		t.Fatalf("decode CODEX_E2E_LOGIN_PARAMS_JSON: %v", err)
	}
	if params == nil {
		t.Fatalf("CODEX_E2E_LOGIN_PARAMS_JSON must decode to a non-null JSON value")
	}
	return params, raw
}

// LiveTurnOptions returns conservative defaults for credential-backed live turns.
func LiveTurnOptions(t testing.TB) *codex.TurnOptions {
	t.Helper()

	opts := &codex.TurnOptions{
		Cwd:            t.TempDir(),
		ApprovalPolicy: codex.ApprovalPolicyNever,
		SandboxPolicy:  map[string]any{"type": "readOnly"},
	}
	if model := os.Getenv("CODEX_E2E_MODEL"); model != "" {
		opts.Model = model
	}
	return opts
}

// CollectTurnStream consumes a turn stream and returns a public TurnResult.
func CollectTurnStream(ctx context.Context, stream *codex.TurnStream) (*codex.TurnResult, error) {
	defer stream.Close()

	result := &codex.TurnResult{}
	for {
		note, err := stream.Next(ctx)
		if err != nil {
			return nil, err
		}
		result.Notifications = append(result.Notifications, note)
		updateTurnResult(result, note)
		if note.Method == "turn/completed" {
			if turnErr := notificationError(note); turnErr != nil {
				return nil, turnErr
			}
			return result, nil
		}
		if note.Method == "turn/failed" {
			if turnErr := notificationError(note); turnErr != nil {
				return nil, turnErr
			}
			return nil, errors.New("turn failed")
		}
		if note.Method == "error" {
			if turnErr := notificationError(note); turnErr != nil {
				return nil, turnErr
			}
		}
	}
}

// AssertCompletedTurnResult checks the stable public effects of a completed turn.
func AssertCompletedTurnResult(t testing.TB, label string, result *codex.TurnResult) {
	t.Helper()

	if result == nil {
		t.Fatalf("%s returned nil result", label)
		return
	}
	if len(result.Notifications) == 0 {
		t.Fatalf("%s returned no notifications", label)
	}
	if result.TurnID == "" {
		t.Fatalf("%s returned empty turn id", label)
	}
	if result.Status != "" && result.Status != "completed" {
		t.Fatalf("%s returned unexpected status %q", label, result.Status)
	}
	if result.FinalResponse == "" && len(result.Items) == 0 {
		t.Fatalf("%s returned no final response or completed items", label)
	}
}

// WaitForKnownTurnID advances a handle until the SDK has observed a turn id.
func WaitForKnownTurnID(ctx context.Context, handle *codex.TurnHandle) error {
	for {
		note, err := handle.Next(ctx)
		if err != nil {
			return err
		}
		if turnIDFromNotification(note) != "" {
			return nil
		}
		if note.Method == "turn/completed" || note.Method == "turn/failed" {
			return fmt.Errorf("turn id was not observed")
		}
	}
}

// IsExpectedTurnControlStateError reports whether a control call raced turn completion.
func IsExpectedTurnControlStateError(err error) bool {
	if err == nil {
		return false
	}
	message := strings.ToLower(err.Error())
	expected := []string{
		"already completed",
		"completed",
		"expected turn",
		"not active",
		"not found",
		"not running",
		"turn state",
	}
	for _, text := range expected {
		if strings.Contains(message, text) {
			return true
		}
	}
	return false
}

// IsExpectedUnmaterializedThreadError identifies expected empty-thread state errors.
func IsExpectedUnmaterializedThreadError(err error) bool {
	if err == nil {
		return false
	}
	message := strings.ToLower(err.Error())
	return strings.Contains(message, "not materialized") || strings.Contains(message, "no rollout found")
}

// IsExpectedUnmaterializedArchiveError identifies expected empty archive state errors.
func IsExpectedUnmaterializedArchiveError(err error) bool {
	if err == nil {
		return false
	}
	message := strings.ToLower(err.Error())
	return strings.Contains(message, "failed to read unarchived thread") ||
		strings.Contains(message, "no archived rollout found") ||
		strings.Contains(message, "thread-store internal error")
}

// AssertJSONContains marshals value and checks that it contains want.
func AssertJSONContains(t testing.TB, label string, value any, want string) {
	t.Helper()

	data, err := json.Marshal(value)
	if err != nil {
		t.Fatalf("marshal %s: %v", label, err)
	}
	if !strings.Contains(string(data), want) {
		t.Fatalf("expected %s to contain %q, got %s", label, want, data)
	}
}

// AssertNoSecretLeak fails if captured output contains any provided secret.
func AssertNoSecretLeak(t testing.TB, stderr *LockedBuffer, secrets ...string) {
	t.Helper()

	output := stderr.rawString()
	for _, secret := range secretMarkers(secrets...) {
		if secret == "" {
			continue
		}
		if strings.Contains(output, secret) {
			t.Fatalf("app-server stderr leaked CODEX_E2E_LOGIN_PARAMS_JSON")
		}
	}
}

type turnPayload struct {
	ThreadID   string                          `json:"threadId,omitempty"`
	TurnID     string                          `json:"turnId,omitempty"`
	Turn       *protocol.TurnNotificationTurn  `json:"turn,omitempty"`
	Item       json.RawMessage                 `json:"item,omitempty"`
	WillRetry  *bool                           `json:"willRetry,omitempty"`
	Error      *protocol.TurnNotificationError `json:"error,omitempty"`
	TokenUsage *protocol.ThreadTokenUsage      `json:"tokenUsage,omitempty"`
}

func updateTurnResult(result *codex.TurnResult, note rpc.Notification) {
	if note.Method != "item/completed" && note.Method != "turn/started" &&
		note.Method != "turn/completed" && note.Method != "turn/failed" &&
		note.Method != "thread/tokenUsage/updated" {
		return
	}

	payload, err := parseTurnNotification(note)
	if err != nil {
		return
	}
	if note.Method == "item/completed" && len(payload.Item) > 0 {
		result.Items = append(result.Items, payload.Item)
		if text, ok := extractTextFromItemRaw(payload.Item); ok {
			result.FinalResponse = text
		}
	}
	if note.Method == "turn/started" || note.Method == "turn/completed" || note.Method == "turn/failed" {
		if payload.Turn != nil && payload.Turn.ID != "" {
			result.TurnID = payload.Turn.ID
		}
		if payload.Turn != nil && payload.Turn.Status != "" {
			result.Status = payload.Turn.Status
		}
		if payload.Turn != nil && payload.Turn.Error != nil && payload.Turn.Error.Message != "" {
			result.ErrorMessage = payload.Turn.Error.Message
		}
	}
	if note.Method == "thread/tokenUsage/updated" && payload.TokenUsage != nil {
		result.TokenUsage = payload.TokenUsage
	}
}

func notificationError(note rpc.Notification) error {
	payload, err := parseTurnNotification(note)
	if err != nil {
		if note.Method == "error" || note.Method == "turn/failed" {
			return errors.New("turn error")
		}
		return nil
	}
	if note.Method == "error" {
		if payload.WillRetry != nil && *payload.WillRetry {
			return nil
		}
		if payload.Error != nil && payload.Error.Message != "" {
			return errors.New(payload.Error.Message)
		}
		return errors.New("turn error")
	}
	if note.Method == "turn/completed" && payload.Turn != nil && payload.Turn.Status == "failed" {
		if payload.Turn.Error != nil && payload.Turn.Error.Message != "" {
			return errors.New(payload.Turn.Error.Message)
		}
		return errors.New("turn failed")
	}
	if note.Method == "turn/failed" {
		if payload.Turn != nil && payload.Turn.Error != nil && payload.Turn.Error.Message != "" {
			return errors.New(payload.Turn.Error.Message)
		}
		return errors.New("turn failed")
	}
	return nil
}

func turnIDFromNotification(note rpc.Notification) string {
	payload, err := parseTurnNotification(note)
	if err != nil {
		return ""
	}
	if payload.Turn != nil && payload.Turn.ID != "" {
		return payload.Turn.ID
	}
	return payload.TurnID
}

func parseTurnNotification(note rpc.Notification) (turnPayload, error) {
	if note.Params != nil {
		switch value := note.Params.(type) {
		case protocol.TurnNotification:
			return turnPayload{ThreadID: value.ThreadID, Turn: value.Turn}, nil
		case protocol.ItemCompletedNotification:
			return turnPayload{ThreadID: value.ThreadID, Item: value.Item}, nil
		case protocol.ErrorNotification:
			return turnPayload{ThreadID: value.ThreadID, WillRetry: value.WillRetry, Error: value.Error}, nil
		case protocol.ThreadTokenUsageUpdatedNotification:
			return turnPayload{ThreadID: value.ThreadID, TurnID: value.TurnID, TokenUsage: &value.TokenUsage}, nil
		}
	}

	var payload turnPayload
	if len(note.Raw) == 0 {
		return payload, nil
	}
	if err := json.Unmarshal(note.Raw, &payload); err != nil {
		return payload, err
	}
	return payload, nil
}

func extractTextFromItemRaw(raw json.RawMessage) (string, bool) {
	if len(raw) == 0 {
		return "", false
	}
	var direct struct {
		Text string `json:"text"`
	}
	if err := json.Unmarshal(raw, &direct); err == nil && direct.Text != "" {
		return direct.Text, true
	}

	var wrapper map[string]json.RawMessage
	if err := json.Unmarshal(raw, &wrapper); err != nil || len(wrapper) != 1 {
		return "", false
	}
	for _, inner := range wrapper {
		var nested struct {
			Text string `json:"text"`
		}
		if err := json.Unmarshal(inner, &nested); err == nil && nested.Text != "" {
			return nested.Text, true
		}
	}
	return "", false
}
