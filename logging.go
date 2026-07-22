package codex

import (
	"io"
	"log/slog"
	"reflect"

	"github.com/pmenglund/codex-sdk-go/rpc"
)

func resolveLogger(logger *slog.Logger) *slog.Logger {
	if logger != nil {
		return logger
	}
	return slog.New(slog.NewTextHandler(io.Discard, nil))
}

func attachApprovalLogger(handler rpc.ServerRequestHandler, logger *slog.Logger) rpc.ServerRequestHandler {
	if isNilServerRequestHandler(handler) {
		return nil
	}
	switch value := handler.(type) {
	case AutoApproveHandler:
		if value.Logger == nil {
			value.Logger = logger
		}
		return value
	case *AutoApproveHandler:
		copy := *value
		if copy.Logger == nil {
			copy.Logger = logger
		}
		return &copy
	case UnsafeLoggingAutoApproveHandler:
		if value.Logger == nil {
			value.Logger = logger
		}
		return value
	case *UnsafeLoggingAutoApproveHandler:
		copy := *value
		if copy.Logger == nil {
			copy.Logger = logger
		}
		return &copy
	default:
		return handler
	}
}

func isNilServerRequestHandler(handler rpc.ServerRequestHandler) bool {
	value := reflect.ValueOf(handler)
	if !value.IsValid() {
		return true
	}
	switch value.Kind() {
	case reflect.Chan, reflect.Func, reflect.Interface, reflect.Map, reflect.Pointer, reflect.Slice:
		return value.IsNil()
	default:
		return false
	}
}
