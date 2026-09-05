package codex

import (
	"encoding/json"
	"errors"
	"fmt"
	"testing"

	"github.com/pmenglund/codex-sdk-go/rpc"
)

func TestOverloadedStructuredData(t *testing.T) {
	for _, tt := range []struct {
		name string
		data string
		want bool
	}{
		{"escaped message", `"server\u0020busy"`, true},
		{"nested array", `{"details":[null,42,{"reason":"rate\u0020limit"}]}`, true},
		{"escaped key", `{"too\u0020many\u0020requests":true}`, true},
		{"ordinary structure", `{"details":[null,42,false,"unavailable",{"reason":"permission denied"}]}`, false},
		{"invalid JSON", `{`, false},
	} {
		t.Run(tt.name, func(t *testing.T) {
			err := fmt.Errorf("request failed: %w", &rpc.ResponseError{Detail: rpc.JSONRPCErrorDetail{
				Code: -32000, Message: "request failed", Data: json.RawMessage(tt.data),
			}})
			if got := IsOverloaded(err); got != tt.want {
				t.Fatalf("IsOverloaded(%s) = %v, want %v", tt.data, got, tt.want)
			}
		})
	}
}

func TestCompatibilityErrorPreservesCause(t *testing.T) {
	cause := errors.New("version probe failed")
	err := &CodexCompatibilityError{Path: "codex", GeneratedVersion: "0.153.4", Reason: "cannot probe", Cause: cause, Hint: "install a compatible CLI"}
	if !errors.Is(err, cause) {
		t.Fatal("compatibility error lost the probe failure")
	}
	want := `codex CLI compatibility check failed for "codex": cannot probe (generated 0.153.4): version probe failed; install a compatible CLI`
	if err.Error() != want {
		t.Fatalf("Error() = %q, want %q", err.Error(), want)
	}
}

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
	responseErr := &rpc.ResponseError{Detail: rpc.JSONRPCErrorDetail{Code: 429, Message: "too many requests"}}
	if !IsOverloaded(responseErr) {
		t.Fatalf("expected 429 response to be overloaded")
	}
	responseErr = &rpc.ResponseError{Detail: rpc.JSONRPCErrorDetail{Code: -32000, Message: "server busy"}}
	if !IsRetryable(responseErr) {
		t.Fatalf("expected server busy to be retryable")
	}
	responseErr = &rpc.ResponseError{Detail: rpc.JSONRPCErrorDetail{Code: -32000, Message: "boom"}}
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
