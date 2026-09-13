package protocol

import (
	"encoding/json"
	"reflect"
	"testing"
)

func TestCodex0154PreservesServerRequestAliasIdentity(t *testing.T) {
	for _, tt := range []struct {
		canonical any
		legacy    any
	}{
		{AttestationGenerateParams{}, SanitizedAttestationGenerateParamsJSON{}},
		{ChatgptAuthTokensRefreshParams{}, SanitizedChatgptAuthTokensRefreshParamsJSON{}},
	} {
		if reflect.TypeOf(tt.canonical) != reflect.TypeOf(tt.legacy) {
			t.Errorf("%T must remain an alias for %T so existing callback signatures remain compatible", tt.canonical, tt.legacy)
		}
	}
}

func TestCodex0154ThreadOriginatorRoundTrip(t *testing.T) {
	for _, originator := range []string{`null`, `"codex_cli_rs"`} {
		t.Run(originator, func(t *testing.T) {
			var thread Thread
			if err := json.Unmarshal([]byte(`{"originator":`+originator+`,"status":{"type":"idle"}}`), &thread); err != nil {
				t.Fatal(err)
			}
			data, err := json.Marshal(thread)
			if err != nil {
				t.Fatal(err)
			}
			var fields map[string]json.RawMessage
			if err := json.Unmarshal(data, &fields); err != nil {
				t.Fatal(err)
			}
			if string(fields["originator"]) != originator {
				t.Fatalf("originator = %s, want %s", fields["originator"], originator)
			}
		})
	}
}

func TestCodex0154ConfigurationUpdateResponseItem(t *testing.T) {
	const payload = `{"type":"configuration_update","reasoning":{"effort":"high"}}`
	var item ResponseItem
	if err := json.Unmarshal([]byte(payload), &item); err != nil {
		t.Fatal(err)
	}
	if !item.IsKnown() {
		t.Fatal("configuration_update must decode as a known response item")
	}
	data, err := json.Marshal(item)
	if err != nil {
		t.Fatal(err)
	}
	if string(data) != payload {
		t.Fatalf("round trip = %s", data)
	}
	if err := json.Unmarshal([]byte(`{"type":"configuration_update"}`), &item); err == nil {
		t.Fatal("configuration_update without required reasoning must fail")
	}
}
