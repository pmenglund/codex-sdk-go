package codex

import (
	"errors"
	"fmt"
	"testing"

	"github.com/pmenglund/codex-sdk-go/rpc"
)

func TestRetryableErrorHelpers(t *testing.T) {
	if IsRetryable(nil) || IsOverloaded(nil) {
		t.Fatalf("nil error should not be retryable")
	}
	if !IsOverloaded(ErrOverloaded) {
		t.Fatalf("expected sentinel overload")
	}
	wrapped := fmt.Errorf("wrapped: %w", ErrOverloaded)
	if !IsRetryable(wrapped) {
		t.Fatalf("expected wrapped overload to be retryable")
	}
	responseErr := &rpc.ResponseError{Detail: rpc.JSONRPCErrorError{Code: 429, Message: "too many requests"}}
	if !IsOverloaded(responseErr) {
		t.Fatalf("expected 429 response to be overloaded")
	}
	responseErr = &rpc.ResponseError{Detail: rpc.JSONRPCErrorError{Code: -32000, Message: "server busy"}}
	if !IsRetryable(responseErr) {
		t.Fatalf("expected server busy to be retryable")
	}
	responseErr = &rpc.ResponseError{Detail: rpc.JSONRPCErrorError{Code: -32000, Message: "boom"}}
	if IsRetryable(responseErr) {
		t.Fatalf("generic response should not be retryable")
	}
	if IsRetryable(errors.New("plain boom")) {
		t.Fatalf("plain boom should not be retryable")
	}
	if IsRetryable(errors.New("request was not overloaded")) {
		t.Fatalf("arbitrary outer error text must not imply retry safety")
	}
	if IsOverloaded(fmt.Errorf("operation failed: server busy")) {
		t.Fatalf("only structured response errors should use overload text")
	}
}
