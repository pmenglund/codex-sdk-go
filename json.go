package codex

import (
	"encoding/json"
	"errors"
	"fmt"
)

// RawJSON represents a pre-serialized JSON value.
type RawJSON = json.RawMessage

// JSON marshals a value into RawJSON.
func JSON(value any) (RawJSON, error) {
	if value == nil {
		return nil, nil
	}
	if raw, ok := value.(json.RawMessage); ok {
		if len(raw) == 0 {
			return nil, nil
		}
		if !json.Valid(raw) {
			return nil, errors.New("invalid raw JSON")
		}
		return raw, nil
	}
	data, err := json.Marshal(value)
	if err != nil {
		return nil, err
	}
	return data, nil
}

// MustJSON marshals a value into RawJSON and panics on error.
func MustJSON(value any) RawJSON {
	raw, err := JSON(value)
	if err != nil {
		panic(err)
	}
	return raw
}

func normalizeJSONValue(label string, value any) (json.RawMessage, error) {
	raw, err := JSON(value)
	if err != nil {
		return nil, fmt.Errorf("%s: %w", label, err)
	}
	return raw, nil
}
