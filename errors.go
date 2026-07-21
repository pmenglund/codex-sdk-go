package codex

import (
	"bytes"
	"encoding/json"
	"errors"
	"fmt"
	"strings"

	"github.com/pmenglund/codex-sdk-go/rpc"
)

// ErrOverloaded identifies retryable overload or server-busy failures.
var ErrOverloaded = errors.New("codex overloaded")

// CodexCompatibilityError reports why a spawned Codex CLI could not be proven
// compatible with the generated protocol package.
type CodexCompatibilityError struct {
	Path             string
	RuntimeVersion   string
	GeneratedVersion string
	GeneratedCommit  string
	Reason           string
	Hint             string
	Cause            error
}

func (e *CodexCompatibilityError) Error() string {
	message := fmt.Sprintf("codex CLI compatibility check failed for %q: %s", e.Path, e.Reason)
	if e.RuntimeVersion != "" {
		message += fmt.Sprintf(" (runtime %s, generated %s)", e.RuntimeVersion, e.GeneratedVersion)
	} else if e.GeneratedVersion != "" {
		message += fmt.Sprintf(" (generated %s)", e.GeneratedVersion)
	}
	if e.Cause != nil {
		message += ": " + e.Cause.Error()
	}
	if e.Hint != "" {
		message += "; " + e.Hint
	}
	return message
}

// Unwrap exposes a version-probe failure, when present.
func (e *CodexCompatibilityError) Unwrap() error {
	return e.Cause
}

// IsOverloaded reports whether err indicates an overload or server-busy failure.
func IsOverloaded(err error) bool {
	if err == nil {
		return false
	}
	if errors.Is(err, ErrOverloaded) {
		return true
	}
	var responseErr *rpc.ResponseError
	if errors.As(err, &responseErr) {
		if responseErr.Detail.Code == 429 {
			return true
		}
		if containsOverloadText(responseErr.Detail.Message) || containsOverloadText(string(responseErr.Detail.Data)) {
			return true
		}
		if responseErr.Detail.Data != nil && json.Valid(responseErr.Detail.Data) {
			var payload any
			if json.Unmarshal(responseErr.Detail.Data, &payload) == nil && containsOverloadJSON(payload) {
				return true
			}
		}
	}
	return containsOverloadText(err.Error())
}

// IsRetryable reports whether err is safe for SDK callers to retry.
func IsRetryable(err error) bool {
	return IsOverloaded(err)
}

func containsOverloadJSON(value any) bool {
	switch value := value.(type) {
	case string:
		return containsOverloadText(value)
	case []any:
		for _, item := range value {
			if containsOverloadJSON(item) {
				return true
			}
		}
	case map[string]any:
		for key, item := range value {
			if containsOverloadText(key) || containsOverloadJSON(item) {
				return true
			}
		}
	}
	return false
}

func containsOverloadText(value string) bool {
	value = strings.ToLower(value)
	value = string(bytes.Map(func(r rune) rune {
		if r == '-' || r == '_' {
			return ' '
		}
		return r
	}, []byte(value)))
	return strings.Contains(value, "overload") ||
		strings.Contains(value, "server busy") ||
		strings.Contains(value, "too many requests") ||
		strings.Contains(value, "rate limit")
}
