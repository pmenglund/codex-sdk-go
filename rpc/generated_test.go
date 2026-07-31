package rpc

import (
	"context"
	"encoding/json"
	"errors"
	"io"
	"os"
	"reflect"
	"regexp"
	"sync"
	"testing"

	"github.com/pmenglund/codex-sdk-go/protocol"
)

func TestGeneratedClientRequests(t *testing.T) {
	transport := newScriptedTransport()
	client := NewClient(transport, ClientOptions{})
	defer client.Close()

	methods := clientRequestMethods()
	for _, name := range methods {
		method := reflect.ValueOf(client).MethodByName(name)
		if !method.IsValid() {
			t.Fatalf("missing method %s", name)
		}

		resultType := method.Type().Out(0)
		var resultValue reflect.Value
		if resultType.Kind() == reflect.Ptr {
			resultValue = reflect.New(resultType.Elem()).Elem()
		} else {
			resultValue = reflect.Zero(resultType)
		}
		transport.enqueueResult(resultValue.Interface())
		transport.enqueueError(-1, "boom")

		args := []reflect.Value{reflect.ValueOf(context.Background())}
		if method.Type().NumIn() == 2 {
			args = append(args, reflect.New(method.Type().In(1)).Elem())
		}
		out := method.Call(args)
		if !out[1].IsNil() {
			t.Fatalf("method %s returned error: %v", name, out[1].Interface())
		}

		out = method.Call(args)
		if out[1].IsNil() {
			t.Fatalf("method %s expected error", name)
		}
	}
}

func TestGeneratedNotifications(t *testing.T) {
	if len(notificationParsers) == 0 {
		t.Fatalf("expected notification methods")
	}
	for method := range notificationParsers {
		note, err := parseServerNotification(method, json.RawMessage("{}"))
		if err != nil {
			t.Fatalf("parseServerNotification %s: %v", method, err)
		}
		if note.Method != method {
			t.Fatalf("unexpected method: %s", note.Method)
		}
	}
}

func TestGeneratedNotificationFallbackAndDecodeError(t *testing.T) {
	note, err := parseServerNotification("custom/unknown", json.RawMessage(`{"ok":true}`))
	if err != nil {
		t.Fatalf("unknown notification returned error: %v", err)
	}
	if note.Method != "custom/unknown" || string(note.Raw) != `{"ok":true}` {
		t.Fatalf("unexpected unknown notification: %#v", note)
	}

	note, err = parseServerNotification("turn/started", json.RawMessage("{bad"))
	if err == nil {
		t.Fatalf("expected decode error")
	}
	if note.Method != "turn/started" || string(note.Raw) != "{bad" {
		t.Fatalf("unexpected failed notification: %#v", note)
	}
}

func TestGeneratedAppRequestWireMethods(t *testing.T) {
	transport := newScriptedTransport()
	client := NewClient(transport, ClientOptions{})
	defer client.Close()

	transport.enqueueResult(protocol.AppsReadResponse{})
	if _, err := client.AppRead(context.Background(), protocol.AppsReadParams{AppIds: []string{"app-one"}}); err != nil {
		t.Fatalf("app/read: %v", err)
	}
	transport.enqueueResult(protocol.AppsInstalledResponse{})
	if _, err := client.AppInstalled(context.Background(), protocol.AppsInstalledParams{}); err != nil {
		t.Fatalf("app/installed: %v", err)
	}

	requests := transport.writtenRequests()
	if len(requests) != 2 {
		t.Fatalf("captured %d requests, want 2", len(requests))
	}
	if requests[0].Method != "app/read" || string(requests[0].Params) != `{"appIds":["app-one"]}` {
		t.Fatalf("unexpected app/read request: %#v", requests[0])
	}
	if requests[1].Method != "app/installed" || string(requests[1].Params) != `{}` {
		t.Fatalf("unexpected app/installed request: %#v", requests[1])
	}
}

func TestGeneratedExternalImportHistoryRequestWireMethod(t *testing.T) {
	transport := newScriptedTransport()
	client := NewClient(transport, ClientOptions{})
	defer client.Close()

	transport.enqueueResult(protocol.ExternalAgentConfigImportHistoryRecordResponse{ImportID: "import-one"})
	result, err := client.ExternalAgentConfigImportRecordHistory(context.Background(), protocol.ExternalAgentConfigImportHistoryRecordParams{
		ItemTypeResults: []protocol.ExternalAgentConfigImportTypeResult{},
		ProviderID:      "provider-one",
	})
	if err != nil {
		t.Fatalf("externalAgentConfig/import/recordHistory: %v", err)
	}
	if result.ImportID != "import-one" {
		t.Fatalf("result = %#v", result)
	}

	requests := transport.writtenRequests()
	if len(requests) != 1 {
		t.Fatalf("captured %d requests, want 1", len(requests))
	}
	if requests[0].Method != "externalAgentConfig/import/recordHistory" || string(requests[0].Params) != `{"itemTypeResults":[],"providerId":"provider-one"}` {
		t.Fatalf("unexpected import history request: %#v", requests[0])
	}
}

func TestGeneratedEnvironmentConnectionNotifications(t *testing.T) {
	for _, method := range []string{"thread/environment/connected", "thread/environment/disconnected"} {
		note, err := parseServerNotification(method, json.RawMessage(`{"environmentId":"environment","threadId":"thread"}`))
		if err != nil {
			t.Fatalf("%s: %v", method, err)
		}
		params, ok := note.Params.(protocol.EnvironmentConnectionNotification)
		if !ok {
			t.Fatalf("%s params type = %T", method, note.Params)
		}
		if params.EnvironmentID != "environment" || params.ThreadID != "thread" {
			t.Fatalf("%s params = %#v", method, params)
		}
	}
}

func TestDispatchServerRequests(t *testing.T) {
	methods := extractCaseMethods(t, "server_requests_gen.go")
	if len(methods) == 0 {
		t.Fatalf("expected server request methods")
	}

	handler := &recordingHandler{}
	for i, method := range methods {
		req := JSONRPCRequest{ID: NewIntRequestID(int64(i + 1)), Method: method, Params: json.RawMessage(`{}`)}
		if _, err := dispatchServerRequest(context.Background(), handler, req); err != nil {
			t.Fatalf("dispatch %s: %v", method, err)
		}
		if handler.lastMethod != method {
			t.Fatalf("handler not invoked for %s", method)
		}
	}
}

func TestUnimplementedServerRequestHandlerSupportsPartialImplementations(t *testing.T) {
	handler := partialServerRequestHandler{}
	var _ ServerRequestHandler = handler
	if _, err := handler.ItemToolCall(context.Background(), protocol.DynamicToolCallParams{}); !errors.Is(err, ErrServerRequestUnsupported) {
		t.Fatalf("expected unsupported default method, got %v", err)
	}
}

func TestCanonicalMCPServerHandlerOverridesEmbeddedLegacyDefault(t *testing.T) {
	handler := &canonicalMCPServerHandler{}
	request := JSONRPCRequest{
		ID:     NewIntRequestID(1),
		Method: "mcpServer/elicitation/request",
		Params: json.RawMessage(`{"message":"choose"}`),
	}
	if _, err := dispatchServerRequest(context.Background(), handler, request); err != nil {
		t.Fatalf("dispatch canonical MCP handler: %v", err)
	}
	if !handler.called {
		t.Fatal("canonical MCP handler was not called")
	}
}

type canonicalMCPServerHandler struct {
	UnimplementedServerRequestHandler
	called bool
}

func (h *canonicalMCPServerHandler) MCPServerElicitationRequest(context.Context, protocol.MCPServerElicitationRequestParams) (*protocol.MCPServerElicitationRequestResponse, error) {
	h.called = true
	return nil, nil
}

type partialServerRequestHandler struct {
	UnimplementedServerRequestHandler
}

func (partialServerRequestHandler) ApplyPatchApproval(context.Context, protocol.ApplyPatchApprovalParams) (*protocol.ApplyPatchApprovalResponse, error) {
	return &protocol.ApplyPatchApprovalResponse{Decision: protocol.MustReviewDecision("approved")}, nil
}

func clientRequestMethods() []string {
	iface := reflect.TypeOf((*ClientRequests)(nil)).Elem()
	methods := make([]string, 0, iface.NumMethod())
	for i := 0; i < iface.NumMethod(); i++ {
		methods = append(methods, iface.Method(i).Name)
	}
	return methods
}

func extractCaseMethods(t *testing.T, filename string) []string {
	t.Helper()
	data, err := os.ReadFile(filename)
	if err != nil {
		t.Fatalf("read %s: %v", filename, err)
	}
	re := regexp.MustCompile(`case "([^"]+)"`)
	matches := re.FindAllStringSubmatch(string(data), -1)
	methods := make([]string, 0, len(matches))
	for _, match := range matches {
		methods = append(methods, match[1])
	}
	return methods
}

type scriptedTransport struct {
	mu        sync.Mutex
	queue     []scriptedResponse
	writes    []JSONRPCRequest
	responses chan string
	closed    chan struct{}
}

func newScriptedTransport() *scriptedTransport {
	return &scriptedTransport{
		responses: make(chan string, 128),
		closed:    make(chan struct{}),
	}
}

func (t *scriptedTransport) enqueueResult(result any) {
	data, err := json.Marshal(result)
	if err != nil {
		panic(err)
	}
	t.mu.Lock()
	t.queue = append(t.queue, scriptedResponse{result: data})
	t.mu.Unlock()
}

func (t *scriptedTransport) enqueueError(code int64, message string) {
	t.mu.Lock()
	t.queue = append(t.queue, scriptedResponse{err: &JSONRPCErrorDetail{Code: code, Message: message}})
	t.mu.Unlock()
}

func (t *scriptedTransport) ReadLine() (string, error) {
	select {
	case line := <-t.responses:
		return line, nil
	case <-t.closed:
		return "", io.EOF
	}
}

func (t *scriptedTransport) WriteLine(line string) error {
	var envelope struct {
		ID json.RawMessage `json:"id"`
	}
	if err := json.Unmarshal([]byte(line), &envelope); err != nil {
		return err
	}
	if len(envelope.ID) == 0 {
		return nil
	}
	id, err := parseRequestID(envelope.ID)
	if err != nil {
		return err
	}
	var request JSONRPCRequest
	if err := json.Unmarshal([]byte(line), &request); err != nil {
		return err
	}

	t.mu.Lock()
	t.writes = append(t.writes, request)
	if len(t.queue) == 0 {
		t.mu.Unlock()
		return errors.New("missing scripted result")
	}
	next := t.queue[0]
	t.queue = t.queue[1:]
	t.mu.Unlock()
	if next.err != nil {
		payload := JSONRPCError{ID: id, Error: *next.err}
		data, err := json.Marshal(payload)
		if err != nil {
			return err
		}
		t.responses <- string(data)
		return nil
	}
	payload := JSONRPCResponse{ID: id, Result: next.result}
	data, err := json.Marshal(payload)
	if err != nil {
		return err
	}
	t.responses <- string(data)
	return nil
}

func (t *scriptedTransport) writtenRequests() []JSONRPCRequest {
	t.mu.Lock()
	defer t.mu.Unlock()
	return append([]JSONRPCRequest(nil), t.writes...)
}

func (t *scriptedTransport) Close() error {
	select {
	case <-t.closed:
	default:
		close(t.closed)
	}
	return nil
}

type recordingHandler struct {
	lastMethod string
}

type scriptedResponse struct {
	result json.RawMessage
	err    *JSONRPCErrorDetail
}

func (h *recordingHandler) ApplyPatchApproval(ctx context.Context, params protocol.ApplyPatchApprovalParams) (*protocol.ApplyPatchApprovalResponse, error) {
	h.lastMethod = "applyPatchApproval"
	resp := protocol.ApplyPatchApprovalResponse{Decision: protocol.MustReviewDecision("approved")}
	return &resp, nil
}

func (h *recordingHandler) AttestationGenerate(ctx context.Context, params protocol.AttestationGenerateParams) (*protocol.AttestationGenerateResponse, error) {
	h.lastMethod = "attestation/generate"
	return nil, nil
}

func (h *recordingHandler) AccountChatgptAuthTokensRefresh(ctx context.Context, params protocol.ChatgptAuthTokensRefreshParams) (*protocol.ChatgptAuthTokensRefreshResponse, error) {
	h.lastMethod = "account/chatgptAuthTokens/refresh"
	return nil, nil
}

func (h *recordingHandler) ExecCommandApproval(ctx context.Context, params protocol.ExecCommandApprovalParams) (*protocol.ExecCommandApprovalResponse, error) {
	h.lastMethod = "execCommandApproval"
	resp := protocol.ExecCommandApprovalResponse{Decision: protocol.MustReviewDecision("approved")}
	return &resp, nil
}

func (h *recordingHandler) ItemCommandExecutionRequestApproval(ctx context.Context, params protocol.CommandExecutionRequestApprovalParams) (*protocol.CommandExecutionRequestApprovalResponse, error) {
	h.lastMethod = "item/commandExecution/requestApproval"
	resp := protocol.CommandExecutionRequestApprovalResponse{Decision: protocol.MustCommandExecutionApprovalDecision("accept")}
	return &resp, nil
}

func (h *recordingHandler) ItemFileChangeRequestApproval(ctx context.Context, params protocol.FileChangeRequestApprovalParams) (*protocol.FileChangeRequestApprovalResponse, error) {
	h.lastMethod = "item/fileChange/requestApproval"
	resp := protocol.FileChangeRequestApprovalResponse{Decision: protocol.FileChangeApprovalDecisionAccept}
	return &resp, nil
}

func (h *recordingHandler) ItemPermissionsRequestApproval(ctx context.Context, params protocol.PermissionsRequestApprovalParams) (*protocol.PermissionsRequestApprovalResponse, error) {
	h.lastMethod = "item/permissions/requestApproval"
	return &protocol.PermissionsRequestApprovalResponse{Permissions: params.Permissions}, nil
}

func (h *recordingHandler) ItemToolCall(ctx context.Context, params protocol.DynamicToolCallParams) (*protocol.DynamicToolCallResponse, error) {
	h.lastMethod = "item/tool/call"
	return nil, nil
}

func (h *recordingHandler) McpServerElicitationRequest(ctx context.Context, params protocol.MCPServerElicitationRequestParams) (*protocol.MCPServerElicitationRequestResponse, error) {
	h.lastMethod = "mcpServer/elicitation/request"
	return nil, nil
}

func (h *recordingHandler) ItemToolRequestUserInput(ctx context.Context, params protocol.ToolRequestUserInputParams) (*protocol.ToolRequestUserInputResponse, error) {
	h.lastMethod = "item/tool/requestUserInput"
	resp := protocol.ToolRequestUserInputResponse{Answers: map[string]protocol.ToolRequestUserInputAnswer{}}
	return &resp, nil
}
