package rpc

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"log/slog"
	"strings"
	"sync"
	"sync/atomic"
)

// ClientOptions configures a low-level JSON-RPC client.
type ClientOptions struct {
	Logger *slog.Logger
	// RequestHandler handles app-server initiated requests. A typed-nil handler
	// is treated as nil.
	RequestHandler ServerRequestHandler
	// ServerRequestWorkers bounds concurrent app-server request handlers. Zero
	// uses the default of four.
	ServerRequestWorkers int
	// ServerRequestQueueCapacity bounds queued app-server requests. Zero uses
	// the default of 64. A full queue receives JSON-RPC code -32001.
	ServerRequestQueueCapacity int
}

const (
	defaultWriteQueueCapacity         = 64
	defaultServerRequestWorkers       = 4
	defaultServerRequestQueueCapacity = 64
)

// ErrClientClosed identifies a client closed by its caller.
var ErrClientClosed = errors.New("rpc client closed")

// ErrConnectionClosed identifies a client whose connection closed without a
// more specific transport error.
var ErrConnectionClosed = errors.New("rpc connection closed")

// ErrOutboundQueueFull identifies a client whose bounded outbound write queue
// could not accept an asynchronous protocol reply.
var ErrOutboundQueueFull = errors.New("rpc outbound write queue is full")

// Client manages JSON-RPC requests over a Transport.
type Client struct {
	transport Transport
	logger    *slog.Logger

	nextID int64

	pendingMu sync.Mutex
	pending   map[string]chan response

	subsMu  sync.Mutex
	subs    map[int]*notificationSubscription
	nextSub int

	handlerMu sync.RWMutex
	handler   ServerRequestHandler
	requests  chan JSONRPCRequest

	writes chan writeRequest

	lifecycle context.Context
	cancel    context.CancelFunc
	done      chan struct{}
	doneOnce  sync.Once
	errMu     sync.RWMutex
	err       error
}

type writeRequest struct {
	ctx  context.Context
	line string
	done chan error
}

// NewClient creates a JSON-RPC client over a Transport. It panics if transport
// is nil. Use NewClientChecked when the dependency is not statically known.
func NewClient(transport Transport, options ClientOptions) *Client {
	client, err := NewClientChecked(transport, options)
	if err != nil {
		panic(err)
	}
	return client
}

// NewClientChecked creates a JSON-RPC client over a validated Transport.
func NewClientChecked(transport Transport, options ClientOptions) (*Client, error) {
	if isNilInterface(transport) {
		return nil, errors.New("rpc transport is nil")
	}
	if options.ServerRequestWorkers < 0 {
		return nil, errors.New("server request workers cannot be negative")
	}
	if options.ServerRequestQueueCapacity < 0 {
		return nil, errors.New("server request queue capacity cannot be negative")
	}
	workers := options.ServerRequestWorkers
	if workers == 0 {
		workers = defaultServerRequestWorkers
	}
	queueCapacity := options.ServerRequestQueueCapacity
	if queueCapacity == 0 {
		queueCapacity = defaultServerRequestQueueCapacity
	}
	if isNilInterface(options.RequestHandler) {
		options.RequestHandler = nil
	}
	logger := options.Logger
	if logger == nil {
		logger = slog.New(slog.NewTextHandler(io.Discard, nil))
	}

	lifecycle, cancel := context.WithCancel(context.Background())

	client := &Client{
		transport: transport,
		logger:    logger,
		pending:   make(map[string]chan response),
		subs:      make(map[int]*notificationSubscription),
		handler:   options.RequestHandler,
		requests:  make(chan JSONRPCRequest, queueCapacity),
		writes:    make(chan writeRequest, defaultWriteQueueCapacity),
		lifecycle: lifecycle,
		cancel:    cancel,
		done:      make(chan struct{}),
	}

	go client.writeLoop()
	for range workers {
		go client.serverRequestWorker()
	}
	go client.readLoop()

	return client, nil
}

// Close shuts down the client and transport.
func (c *Client) Close() error {
	c.finish(ErrClientClosed)
	return c.transport.Close()
}

// SetRequestHandler replaces the server request handler.
func (c *Client) SetRequestHandler(handler ServerRequestHandler) {
	if isNilInterface(handler) {
		handler = nil
	}
	c.handlerMu.Lock()
	defer c.handlerMu.Unlock()
	c.handler = handler
}

// Call sends a JSON-RPC request and decodes the response into result.
func (c *Client) Call(ctx context.Context, method string, params any, result any) error {
	if err := ctx.Err(); err != nil {
		return err
	}
	if err := c.ensureOpen(); err != nil {
		return err
	}

	id := c.nextRequestID()
	respCh := make(chan response, 1)

	c.pendingMu.Lock()
	c.pending[id.Key()] = respCh
	c.pendingMu.Unlock()

	payload, err := BuildClientRequest(method, params, id)
	if err != nil {
		c.deletePending(id)
		return err
	}

	if err := ctx.Err(); err != nil {
		c.deletePending(id)
		return err
	}
	if err := c.send(ctx, payload); err != nil {
		c.deletePending(id)
		return err
	}

	select {
	case <-c.done:
		c.deletePending(id)
		return c.errOrClosed()
	case <-ctx.Done():
		c.deletePending(id)
		return ctx.Err()
	case resp := <-respCh:
		if resp.err != nil {
			return resp.err
		}
		if result == nil {
			return nil
		}
		return json.Unmarshal(resp.result, result)
	}
}

// Notify sends a JSON-RPC notification.
func (c *Client) Notify(ctx context.Context, method string, params any) error {
	if err := c.ensureOpen(); err != nil {
		return err
	}

	payload := JSONRPCNotification{Method: method}
	if params != nil {
		data, err := json.Marshal(params)
		if err != nil {
			return err
		}
		payload.Params = data
	}

	data, err := json.Marshal(payload)
	if err != nil {
		return err
	}

	return c.write(ctx, string(data))
}

// SubscribeNotifications creates an iterator over server notifications.
// Buffer is a hard pending-notification capacity. A non-positive value uses
// the default capacity of 64. If a consumer falls behind, only that iterator
// is closed and Next returns a NotificationOverflowError.
func (c *Client) SubscribeNotifications(buffer int) *NotificationIterator {
	sub := newNotificationSubscription(buffer)

	c.subsMu.Lock()
	id := -1
	select {
	case <-c.done:
		sub.close(c.errOrClosed())
	default:
		id = c.nextSub
		c.nextSub++
		c.subs[id] = sub
	}
	c.subsMu.Unlock()

	return &NotificationIterator{
		ch:         sub.out,
		subDone:    sub.done,
		clientDone: c.done,
		subErr:     sub.terminalError,
		clientErr:  c.errOrClosed,
		cancel: func() {
			if id < 0 {
				return
			}
			c.subsMu.Lock()
			sub := c.subs[id]
			delete(c.subs, id)
			c.subsMu.Unlock()
			if sub != nil {
				sub.close(ErrNotificationSubscriptionClosed)
			}
		},
	}
}

func (c *Client) readLoop() {
	for {
		line, err := c.transport.ReadLine()
		if err != nil {
			c.finish(err)
			return
		}
		if strings.TrimSpace(line) == "" {
			continue
		}

		msg, err := parseMessage([]byte(line))
		if err != nil {
			c.logger.Warn("failed to parse json-rpc message", slog.Any("error", err))
			continue
		}

		switch msg.kind {
		case messageResponse:
			c.handleResponse(msg.response)
		case messageError:
			c.handleError(msg.error)
		case messageRequest:
			c.enqueueServerRequest(msg.request)
		case messageNotification:
			c.handleNotification(msg.notification)
		}
	}
}

func (c *Client) handleResponse(resp JSONRPCResponse) {
	c.pendingMu.Lock()
	ch := c.pending[resp.ID.Key()]
	delete(c.pending, resp.ID.Key())
	c.pendingMu.Unlock()

	if ch == nil {
		return
	}

	ch <- response{result: resp.Result}
}

func (c *Client) handleError(resp JSONRPCError) {
	c.pendingMu.Lock()
	ch := c.pending[resp.ID.Key()]
	delete(c.pending, resp.ID.Key())
	c.pendingMu.Unlock()

	if ch == nil {
		return
	}

	ch <- response{err: &ResponseError{ID: resp.ID, Detail: resp.Error}}
}

func (c *Client) handleNotification(note JSONRPCNotification) {
	notification, err := parseServerNotification(note.Method, note.Params)
	if err != nil {
		c.logger.Warn("failed to decode notification", slog.String("method", note.Method), slog.Any("error", err))
	}

	type subscriptionEntry struct {
		id  int
		sub *notificationSubscription
	}
	c.subsMu.Lock()
	subs := make([]subscriptionEntry, 0, len(c.subs))
	for id, sub := range c.subs {
		subs = append(subs, subscriptionEntry{id: id, sub: sub})
	}
	c.subsMu.Unlock()

	for _, entry := range subs {
		if entry.sub.publish(notification) {
			continue
		}
		c.subsMu.Lock()
		if c.subs[entry.id] == entry.sub {
			delete(c.subs, entry.id)
		}
		c.subsMu.Unlock()
		var overflow *NotificationOverflowError
		if errors.As(entry.sub.terminalError(), &overflow) {
			c.logger.Warn("notification subscription overflow",
				slog.Int("subscription_id", entry.id),
				slog.Int("capacity", overflow.Capacity),
				slog.String("method", note.Method),
			)
		}
	}
}

func (c *Client) serverRequestWorker() {
	for {
		select {
		case <-c.done:
			return
		case req := <-c.requests:
			c.handleServerRequest(req)
		}
	}
}

func (c *Client) enqueueServerRequest(req JSONRPCRequest) {
	select {
	case <-c.done:
		return
	case c.requests <- req:
	default:
		response := JSONRPCError{
			ID: req.ID,
			Error: JSONRPCErrorDetail{
				Code:    ServerRequestBusyCode,
				Message: "server request queue is full",
			},
		}
		if err := c.sendAsync(c.requestContext(), response); err != nil {
			c.finish(err)
		}
	}
}

func (c *Client) handleServerRequest(req JSONRPCRequest) {
	defer func() {
		if recovered := recover(); recovered != nil {
			c.logger.Error("server request handler panicked", slog.String("method", req.Method), slog.Any("panic", recovered))
			if err := c.replyError(c.requestContext(), req.ID, ServerRequestInternalErrorCode, "server request handler panicked", nil); err != nil {
				c.finish(err)
			}
		}
	}()

	handler := c.currentHandler()
	if handler == nil {
		if err := c.replyError(c.requestContext(), req.ID, ServerRequestMethodNotFoundCode, "no handler configured", nil); err != nil {
			c.finish(err)
		}
		return
	}

	result, err := dispatchServerRequest(c.requestContext(), handler, req)
	if err != nil {
		code, message := classifyServerRequestError(err)
		c.logger.Warn("server request failed", slog.String("method", req.Method), slog.Any("error", err))
		if replyErr := c.replyError(c.requestContext(), req.ID, code, message, nil); replyErr != nil {
			c.finish(replyErr)
		}
		return
	}

	if err := c.replyResult(c.requestContext(), req.ID, result); err != nil {
		c.finish(err)
	}
}

func (c *Client) replyResult(ctx context.Context, id RequestID, result any) error {
	data, err := json.Marshal(result)
	if err != nil {
		return err
	}
	resp := JSONRPCResponse{ID: id, Result: data}
	return c.send(ctx, resp)
}

func (c *Client) replyError(ctx context.Context, id RequestID, code int64, message string, data json.RawMessage) error {
	resp := JSONRPCError{
		ID: id,
		Error: JSONRPCErrorError{
			Code:    code,
			Message: message,
			Data:    data,
		},
	}
	return c.send(ctx, resp)
}

func (c *Client) send(ctx context.Context, payload any) error {
	data, err := json.Marshal(payload)
	if err != nil {
		return err
	}
	return c.write(ctx, string(data))
}

func (c *Client) write(ctx context.Context, line string) error {
	request := writeRequest{ctx: ctx, line: line, done: make(chan error, 1)}
	select {
	case <-ctx.Done():
		return ctx.Err()
	case <-c.done:
		return c.errOrClosed()
	case c.writes <- request:
	}
	select {
	case <-ctx.Done():
		return ctx.Err()
	case <-c.done:
		return c.errOrClosed()
	case err := <-request.done:
		return err
	}

}

func (c *Client) sendAsync(ctx context.Context, payload any) error {
	data, err := json.Marshal(payload)
	if err != nil {
		return err
	}
	request := writeRequest{ctx: ctx, line: string(data), done: make(chan error, 1)}
	select {
	case <-ctx.Done():
		return ctx.Err()
	case <-c.done:
		return c.errOrClosed()
	case c.writes <- request:
		return nil
	default:
		return ErrOutboundQueueFull
	}
}

func (c *Client) writeLoop() {
	for {
		select {
		case <-c.done:
			return
		case request := <-c.writes:
			if err := request.ctx.Err(); err != nil {
				request.done <- err
				continue
			}
			var err error
			if transport, ok := c.transport.(ContextTransport); ok {
				err = transport.WriteLineContext(request.ctx, request.line)
			} else {
				err = c.transport.WriteLine(request.line)
			}
			request.done <- err
			if err != nil && request.ctx.Err() == nil {
				c.finish(err)
				return
			}
		}
	}
}

func (c *Client) nextRequestID() RequestID {
	next := atomic.AddInt64(&c.nextID, 1)
	return NewIntRequestID(next)
}

func (c *Client) deletePending(id RequestID) {
	c.pendingMu.Lock()
	delete(c.pending, id.Key())
	c.pendingMu.Unlock()
}

func (c *Client) currentHandler() ServerRequestHandler {
	c.handlerMu.RLock()
	defer c.handlerMu.RUnlock()
	return c.handler
}

func (c *Client) requestContext() context.Context {
	if c.lifecycle != nil {
		return c.lifecycle
	}
	return context.Background()
}

func (c *Client) ensureOpen() error {
	select {
	case <-c.done:
		return c.errOrClosed()
	default:
		return nil
	}
}

func (c *Client) errOrClosed() error {
	c.errMu.RLock()
	defer c.errMu.RUnlock()
	if c.err != nil {
		return c.err
	}
	return ErrConnectionClosed
}

func (c *Client) finish(err error) {
	c.doneOnce.Do(func() {
		c.errMu.Lock()
		c.err = err
		c.errMu.Unlock()
		if c.cancel != nil {
			c.cancel()
		}
		close(c.done)
		c.pendingMu.Lock()
		for _, ch := range c.pending {
			ch <- response{err: err}
		}
		c.pending = map[string]chan response{}
		c.pendingMu.Unlock()

		c.subsMu.Lock()
		subs := make([]*notificationSubscription, 0, len(c.subs))
		for _, sub := range c.subs {
			subs = append(subs, sub)
		}
		c.subs = map[int]*notificationSubscription{}
		c.subsMu.Unlock()

		for _, sub := range subs {
			sub.close(err)
		}
	})
}

type response struct {
	result json.RawMessage
	err    error
}

// ErrNotificationOverflow identifies a notification subscription that was
// closed because its configured capacity was exhausted.
var ErrNotificationOverflow = errors.New("notification subscription overflow")

// ErrNotificationSubscriptionClosed identifies an iterator closed by its caller.
var ErrNotificationSubscriptionClosed = errors.New("notification subscription closed")

// NotificationOverflowError reports the hard capacity of an overflowing
// notification subscription.
type NotificationOverflowError struct {
	Capacity int
}

func (e *NotificationOverflowError) Error() string {
	return fmt.Sprintf("%s (capacity %d)", ErrNotificationOverflow, e.Capacity)
}

// Is allows errors.Is(err, ErrNotificationOverflow).
func (e *NotificationOverflowError) Is(target error) bool {
	return target == ErrNotificationOverflow
}

type notificationSubscription struct {
	mu       sync.Mutex
	out      chan Notification
	done     chan struct{}
	closed   bool
	err      error
	capacity int
}

func newNotificationSubscription(buffer int) *notificationSubscription {
	if buffer <= 0 {
		buffer = 64
	}
	return &notificationSubscription{
		out:      make(chan Notification, buffer),
		done:     make(chan struct{}),
		capacity: buffer,
	}
}

func (s *notificationSubscription) publish(note Notification) bool {
	s.mu.Lock()
	defer s.mu.Unlock()
	if s.closed {
		return false
	}
	select {
	case s.out <- note:
		return true
	default:
		s.closeLocked(&NotificationOverflowError{Capacity: s.capacity})
		return false
	}
}

func (s *notificationSubscription) close(err error) {
	s.mu.Lock()
	defer s.mu.Unlock()
	s.closeLocked(err)
}

func (s *notificationSubscription) closeLocked(err error) {
	if s.closed {
		return
	}
	s.closed = true
	s.err = err
	close(s.done)
}

func (s *notificationSubscription) terminalError() error {
	s.mu.Lock()
	defer s.mu.Unlock()
	if s.err != nil {
		return s.err
	}
	return ErrNotificationSubscriptionClosed
}

// NotificationIterator iterates notifications from the server.
type NotificationIterator struct {
	ch         <-chan Notification
	subDone    <-chan struct{}
	clientDone <-chan struct{}
	subErr     func() error
	clientErr  func() error
	cancel     func()
}

// Next returns the next notification or an error.
func (it *NotificationIterator) Next(ctx context.Context) (Notification, error) {
	select {
	case <-it.subDone:
		return Notification{}, it.subErr()
	default:
	}
	select {
	case <-ctx.Done():
		return Notification{}, ctx.Err()
	case <-it.subDone:
		return Notification{}, it.subErr()
	case <-it.clientDone:
		return Notification{}, it.clientErr()
	case note := <-it.ch:
		select {
		case <-it.subDone:
			return Notification{}, it.subErr()
		default:
		}
		return note, nil
	}
}

// Close unsubscribes the iterator.
func (it *NotificationIterator) Close() {
	if it.cancel != nil {
		it.cancel()
	}
}

// parseMessage decodes a JSON-RPC line into a typed message.
func parseMessage(data []byte) (message, error) {
	var envelope struct {
		ID     json.RawMessage    `json:"id"`
		Method string             `json:"method"`
		Params json.RawMessage    `json:"params"`
		Result json.RawMessage    `json:"result"`
		Error  *JSONRPCErrorError `json:"error"`
	}

	if err := json.Unmarshal(data, &envelope); err != nil {
		return message{}, err
	}

	if envelope.Method != "" {
		if len(envelope.ID) > 0 {
			id, err := parseRequestID(envelope.ID)
			if err != nil {
				return message{}, err
			}
			return message{kind: messageRequest, request: JSONRPCRequest{ID: id, Method: envelope.Method, Params: envelope.Params}}, nil
		}
		return message{kind: messageNotification, notification: JSONRPCNotification{Method: envelope.Method, Params: envelope.Params}}, nil
	}

	if len(envelope.Result) > 0 {
		id, err := parseRequestID(envelope.ID)
		if err != nil {
			return message{}, err
		}
		return message{kind: messageResponse, response: JSONRPCResponse{ID: id, Result: envelope.Result}}, nil
	}

	if envelope.Error != nil {
		id, err := parseRequestID(envelope.ID)
		if err != nil {
			return message{}, err
		}
		return message{kind: messageError, error: JSONRPCError{ID: id, Error: *envelope.Error}}, nil
	}

	return message{}, fmt.Errorf("unrecognized json-rpc message")
}

func parseRequestID(raw json.RawMessage) (RequestID, error) {
	var id RequestID
	if err := id.UnmarshalJSON(raw); err != nil {
		return RequestID{}, err
	}
	return id, nil
}

type messageKind int

const (
	messageResponse messageKind = iota
	messageError
	messageRequest
	messageNotification
)

type message struct {
	kind         messageKind
	response     JSONRPCResponse
	error        JSONRPCError
	request      JSONRPCRequest
	notification JSONRPCNotification
}
