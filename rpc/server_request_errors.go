package rpc

import (
	"errors"
	"fmt"
)

// ErrServerRequestUnsupported is returned by default handler methods that an
// application has not overridden.
var ErrServerRequestUnsupported = errors.New("server request is unsupported")

const (
	// ServerRequestMethodNotFoundCode is the JSON-RPC code for an unknown method.
	ServerRequestMethodNotFoundCode int64 = -32601
	// ServerRequestInvalidParamsCode is the JSON-RPC code for malformed parameters.
	ServerRequestInvalidParamsCode int64 = -32602
	// ServerRequestInternalErrorCode is the JSON-RPC code for an internal failure.
	ServerRequestInternalErrorCode int64 = -32603
	// ServerRequestHandlerErrorCode is used when an application handler fails.
	ServerRequestHandlerErrorCode int64 = -32000
	// ServerRequestBusyCode is used when the bounded handler queue is full.
	ServerRequestBusyCode int64 = -32001
)

// ServerRequestMethodError reports an unsupported server-request method.
type ServerRequestMethodError struct{ Method string }

func (e *ServerRequestMethodError) Error() string {
	return fmt.Sprintf("unsupported server request %q", e.Method)
}

// ServerRequestParamsError reports parameters that could not be decoded.
type ServerRequestParamsError struct {
	Method string
	Err    error
}

func (e *ServerRequestParamsError) Error() string {
	return fmt.Sprintf("decode server request %q parameters: %v", e.Method, e.Err)
}
func (e *ServerRequestParamsError) Unwrap() error { return e.Err }

// ServerRequestHandlerError reports an error returned by an application handler.
type ServerRequestHandlerError struct {
	Method string
	Err    error
}

func (e *ServerRequestHandlerError) Error() string {
	return fmt.Sprintf("handle server request %q: %v", e.Method, e.Err)
}
func (e *ServerRequestHandlerError) Unwrap() error { return e.Err }

func classifyServerRequestError(err error) (int64, string) {
	if errors.Is(err, ErrServerRequestUnsupported) {
		return ServerRequestMethodNotFoundCode, "method not found"
	}
	var methodErr *ServerRequestMethodError
	if errors.As(err, &methodErr) {
		return ServerRequestMethodNotFoundCode, "method not found"
	}
	var paramsErr *ServerRequestParamsError
	if errors.As(err, &paramsErr) {
		return ServerRequestInvalidParamsCode, "invalid params"
	}
	return ServerRequestHandlerErrorCode, "server request handler failed"
}
