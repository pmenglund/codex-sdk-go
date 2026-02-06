package rpc

import "encoding/json"

// UnmarshalParams decodes the raw notification params into v.
func (n Notification) UnmarshalParams(v any) error {
	if len(n.Raw) == 0 {
		return nil
	}
	return json.Unmarshal(n.Raw, v)
}
