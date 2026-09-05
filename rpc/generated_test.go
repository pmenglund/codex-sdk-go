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

func TestAccountUsageReadOptionalFilter(t *testing.T) {
	transport := newScriptedTransport()
	client := NewClient(transport, ClientOptions{})
	defer client.Close()
	for range 3 {
		transport.enqueueResult(map[string]any{})
	}
	ctx := context.Background()
	if _, err := client.AccountUsageRead(ctx); err != nil {
		t.Fatal(err)
	}
	if _, err := client.AccountUsageReadWithParams(ctx, nil); err != nil {
		t.Fatal(err)
	}
	threadID := "thr_1"
	if _, err := client.AccountUsageReadWithParams(ctx, &protocol.GetAccountTokenUsageParams{ThreadID: &threadID}); err != nil {
		t.Fatal(err)
	}
	requests := transport.writtenRequests()
	if len(requests) != 3 {
		t.Fatalf("requests = %d", len(requests))
	}
	for i, want := range []string{"", "", `{"threadId":"thr_1"}`} {
		if requests[i].Method != "account/usage/read" || string(requests[i].Params) != want {
			t.Errorf("request %d = %#v", i, requests[i])
		}
	}
}

func TestPaginatedThreadHistoryRPC(t *testing.T) {
	transport := newScriptedTransport()
	client := NewClient(transport, ClientOptions{})
	defer client.Close()
	transport.enqueueResult(json.RawMessage(`{"data":[{"id":"turn_1","items":[],"status":"completed"}],"nextCursor":"turn-next","backwardsCursor":"turn-back"}`))
	transport.enqueueResult(json.RawMessage(`{"data":[{"turnId":"turn_1","item":{"type":"agentMessage","id":"item_1","text":"done"}}],"nextCursor":"item-next","backwardsCursor":"item-back"}`))
	transport.enqueueResult(json.RawMessage(`{"thread":{"id":"thr_1","projectId":null,"turns":[]},"itemsBackwardsCursor":"item-back","turnsBackwardsCursor":"turn-back"}`))
	ctx := context.Background()
	cursor, turnID := "cursor_1", "turn_1"
	limit := 7
	direction := protocol.SortDirectionAsc
	turns, err := client.ThreadTurnsList(ctx, protocol.ThreadTurnsListParams{ThreadID: "thr_1", Cursor: &cursor, Limit: &limit, SortDirection: &direction, ItemsView: protocol.TurnItemsViewSummary})
	if err != nil {
		t.Fatal(err)
	}
	if len(turns.Data) != 1 || turns.Data[0].ID != "turn_1" || turns.NextCursor == nil || *turns.NextCursor != "turn-next" || turns.BackwardsCursor == nil || *turns.BackwardsCursor != "turn-back" {
		t.Fatalf("turn page = %#v", turns)
	}
	items, err := client.ThreadItemsList(ctx, protocol.ThreadItemsListParams{ThreadID: "thr_1", Cursor: &cursor, Limit: &limit, SortDirection: &direction, TurnID: &turnID})
	if err != nil {
		t.Fatal(err)
	}
	if len(items.Data) != 1 || items.Data[0].TurnID != "turn_1" || !items.Data[0].Item.IsKnown() || items.NextCursor == nil || *items.NextCursor != "item-next" || items.BackwardsCursor == nil || *items.BackwardsCursor != "item-back" {
		t.Fatalf("item page = %#v", items)
	}
	reverted, err := client.ThreadRevert(ctx, protocol.ThreadRevertParams{ThreadID: "thr_1", BeforeTurnID: "turn_2"})
	if err != nil {
		t.Fatal(err)
	}
	if reverted.Thread.ID != "thr_1" || reverted.ItemsBackwardsCursor == nil || *reverted.ItemsBackwardsCursor != "item-back" || reverted.TurnsBackwardsCursor == nil || *reverted.TurnsBackwardsCursor != "turn-back" {
		t.Fatalf("revert = %#v", reverted)
	}
	requests := transport.writtenRequests()
	wantParams := []string{`{"cursor":"cursor_1","itemsView":"summary","limit":7,"sortDirection":"asc","threadId":"thr_1"}`, `{"cursor":"cursor_1","limit":7,"sortDirection":"asc","threadId":"thr_1","turnId":"turn_1"}`, `{"beforeTurnId":"turn_2","threadId":"thr_1"}`}
	for i, method := range []string{"thread/turns/list", "thread/items/list", "thread/revert"} {
		if string(requests[i].Params) != wantParams[i] {
			t.Errorf("request %s: params %s, want %s", method, requests[i].Params, wantParams[i])
		}
		if requests[i].Method != method {
			t.Fatalf("request %d = %s", i, requests[i].Method)
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
		ItemTypeResults: []protocol.ExternalAgentConfigImportHistoryRecordTypeResultParams{},
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

func TestGeneratedThreadSectionRequestWireMethods(t *testing.T) {
	transport := newScriptedTransport()
	client := NewClient(transport, ClientOptions{})
	defer client.Close()

	transport.enqueueResult(protocol.ThreadSectionCreateResponse{Section: protocol.ThreadSection{ID: "section-one", Name: "Work"}})
	if _, err := client.ThreadSectionCreate(context.Background(), protocol.ThreadSectionCreateParams{Name: "Work"}); err != nil {
		t.Fatalf("threadSection/create: %v", err)
	}
	transport.enqueueResult(protocol.ThreadSectionDeleteResponse{})
	if _, err := client.ThreadSectionDelete(context.Background(), protocol.ThreadSectionDeleteParams{SectionID: "section-one"}); err != nil {
		t.Fatalf("threadSection/delete: %v", err)
	}
	cursor := "cursor-one"
	limit := 2
	transport.enqueueResult(protocol.ThreadSectionListResponse{Data: []protocol.ThreadSection{{ID: "section-one", Name: "Work"}}})
	if _, err := client.ThreadSectionList(context.Background(), protocol.ThreadSectionListParams{
		Cursor: protocol.ThreadSectionListParamsCursor(&cursor),
		Limit:  protocol.ThreadSectionListParamsLimit(&limit),
	}); err != nil {
		t.Fatalf("threadSection/list: %v", err)
	}
	beforeThreadID := "thread-two"
	sectionID := "section-one"
	transport.enqueueResult(protocol.ThreadSectionMoveResponse{})
	if _, err := client.ThreadSectionMove(context.Background(), protocol.ThreadSectionMoveParams{
		BeforeThreadID: protocol.ThreadSectionMoveParamsBeforeThreadID(&beforeThreadID),
		SectionID:      protocol.ThreadSectionMoveParamsSectionID(&sectionID),
		ThreadID:       "thread-one",
	}); err != nil {
		t.Fatalf("thread/section/move: %v", err)
	}
	transport.enqueueResult(protocol.ThreadSectionMoveResponse{})
	if _, err := client.ThreadSectionMove(context.Background(), protocol.ThreadSectionMoveParams{ThreadID: "thread-one"}); err != nil {
		t.Fatalf("thread/section/move out: %v", err)
	}
	transport.enqueueResult(protocol.ThreadSectionUpdateResponse{Section: protocol.ThreadSection{ID: "section-one", Name: "Projects"}})
	if _, err := client.ThreadSectionUpdate(context.Background(), protocol.ThreadSectionUpdateParams{Name: "Projects", SectionID: "section-one"}); err != nil {
		t.Fatalf("threadSection/update: %v", err)
	}

	want := []struct {
		method string
		params string
	}{
		{method: "threadSection/create", params: `{"name":"Work"}`},
		{method: "threadSection/delete", params: `{"sectionId":"section-one"}`},
		{method: "threadSection/list", params: `{"cursor":"cursor-one","limit":2}`},
		{method: "thread/section/move", params: `{"beforeThreadId":"thread-two","sectionId":"section-one","threadId":"thread-one"}`},
		{method: "thread/section/move", params: `{"sectionId":null,"threadId":"thread-one"}`},
		{method: "threadSection/update", params: `{"name":"Projects","sectionId":"section-one"}`},
	}
	requests := transport.writtenRequests()
	if len(requests) != len(want) {
		t.Fatalf("captured %d requests, want %d", len(requests), len(want))
	}
	for i := range want {
		if requests[i].Method != want[i].method || string(requests[i].Params) != want[i].params {
			t.Fatalf("request %d = %#v, want method %q params %s", i, requests[i], want[i].method, want[i].params)
		}
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

func TestDispatchOlderCommandApprovalDefaultsKind(t *testing.T) {
	handler := &recordingHandler{}
	req := JSONRPCRequest{ID: NewIntRequestID(1), Method: "item/commandExecution/requestApproval", Params: json.RawMessage(`{"threadId":"thr_1","turnId":"turn_1","itemId":"item_1"}`)}
	if _, err := dispatchServerRequest(context.Background(), handler, req); err != nil {
		t.Fatal(err)
	}
	if handler.approvalKind != protocol.CommandExecutionApprovalKindCommand {
		t.Fatalf("approval kind = %q", handler.approvalKind)
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
	approvalKind protocol.CommandExecutionApprovalKind
	lastMethod   string
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
	h.approvalKind = params.Kind
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
