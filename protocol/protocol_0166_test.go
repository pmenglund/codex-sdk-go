package protocol

import (
	"bytes"
	"encoding/json"
	"reflect"
	"testing"
)

func TestCodex0166OptionalFields(t *testing.T) {
	for _, test := range []struct {
		name, payload, field string
		value                any
	}{
		{"process drain", `{"processHandle":"p","exitCode":0,"stdout":"","stderr":"","outputDrainTimedOut":false}`, "outputDrainTimedOut", &ProcessExitedNotification{}},
		{"hook thread", `{"threadId":"thread"}`, "threadId", &HooksListParams{}},
		{"named context", `{"kind":"untrusted","value":"context","name":"source"}`, "name", &AdditionalContextEntry{}},
		{"zero native usage", `{"threadId":"thread","estimatedUsageCreditsMicros":0,"groups":[],"nativeUsageUsdMicros":0}`, "nativeUsageUsdMicros", &ThreadUsage{}},
		{"exact native usage", `{"threadId":"thread","estimatedUsageCreditsMicros":0,"groups":[],"nativeUsageUsdMicros":9007199254740993}`, "nativeUsageUsdMicros", &ThreadUsage{}},
		{"network review", `{"reviewDomains":["*.example.test"]}`, "reviewDomains", &NetworkRequirements{}},
	} {
		t.Run(test.name, func(t *testing.T) {
			if err := json.Unmarshal([]byte(test.payload), test.value); err != nil {
				t.Fatal(err)
			}
			check0162Fields(t, test.payload, test.value, test.field)
		})
	}
	for _, test := range []struct {
		value any
		field string
	}{{&ProcessExitedNotification{}, "outputDrainTimedOut"}, {&ThreadUsage{}, "nativeUsageUsdMicros"}, {&HooksListParams{}, "threadId"}, {&AdditionalContextEntry{}, "name"}} {
		if err := json.Unmarshal([]byte(`{}`), test.value); err != nil {
			t.Fatal(err)
		}
		check0162Fields(t, `{}`, test.value, test.field)
	}
}

func TestCodex0166Accounts(t *testing.T) {
	for _, payload := range []string{
		`{"type":"gcp","future":9007199254740993}`,
		`{"type":"chatgpt","email":null,"planType":"pro","workspaceName":"Editorial"}`,
	} {
		var account Account
		if err := json.Unmarshal([]byte(payload), &account); err != nil {
			t.Fatal(err)
		}
		encoded, err := json.Marshal(account)
		if err != nil || !account.IsKnown() || string(encoded) != payload {
			t.Fatalf("account changed or unrecognized: %s, %v", encoded, err)
		}
	}
}

func TestCodex0166ItemLifecycle(t *testing.T) {
	for _, payload := range []string{
		`{"type":"agentMessage","id":"i","text":"done","startedAtMs":0,"completedAtMs":9007199254740993}`,
		`{"type":"imageGeneration","id":"i","status":"completed","fileId":"file","generationId":"generation","startedAtMs":null,"completedAtMs":12,"future":9007199254740993}`,
	} {
		var item ThreadItem
		if err := json.Unmarshal([]byte(payload), &item); err != nil {
			t.Fatal(err)
		}
		encoded, err := json.Marshal(item)
		if err != nil || !item.IsKnown() || string(encoded) != payload {
			t.Fatalf("item changed: %s, %v", encoded, err)
		}
	}
	var item ThreadItem
	if err := json.Unmarshal([]byte(`{"type":"agentMessage","id":"i","startedAtMs":0}`), &item); err == nil {
		t.Fatal("lifecycle fields allowed an item without its required text")
	}
}

func TestCodex0166OtelRequirements(t *testing.T) {
	for _, exporter := range []string{
		`"none"`, `"statsig"`,
		`{"otlp-http":{"endpoint":"https://example.test","headers":{"authorization":"[REDACTED]"},"protocol":"binary","tls":{"ca-certificate":"C:\\certs\\ca.pem","client-certificate":"/tmp/client.pem","client-private-key":"/tmp/key.pem"}}}`,
		`{"otlp-grpc":{"endpoint":"https://example.test","headers":{},"tls":null}}`,
		`{"future-exporter":{"count":9007199254740993}}`,
	} {
		payload := `{"otel":{"exporter":` + exporter + `,"trace_exporter":` + exporter + `,"metrics_exporter":` + exporter + `,"tool_result":{"max_bytes":0},"log_user_prompt":false,"span_attributes":{"team":"editorial"},"tracestate":{"vendor":{"key":"value"}}}}`
		var requirements ConfigRequirements
		if err := json.Unmarshal([]byte(payload), &requirements); err != nil {
			t.Fatal(err)
		}
		encoded, err := json.Marshal(requirements)
		if err != nil {
			t.Fatal(err)
		}
		var before, after map[string]json.RawMessage
		if err := json.Unmarshal([]byte(payload), &before); err != nil {
			t.Fatal(err)
		}
		if err := json.Unmarshal(encoded, &after); err != nil {
			t.Fatal(err)
		}
		decode := func(data []byte) any {
			t.Helper()
			var result any
			decoder := json.NewDecoder(bytes.NewReader(data))
			decoder.UseNumber()
			if err := decoder.Decode(&result); err != nil {
				t.Fatal(err)
			}
			return result
		}
		if !reflect.DeepEqual(decode(before["otel"]), decode(after["otel"])) {
			t.Fatalf("OTEL requirements changed: %s", encoded)
		}
	}
}

func TestCodex0166HookRunPaths(t *testing.T) {
	for _, path := range []string{"/tmp/hooks.json", `C:\Users\test\hooks.json`, `\\server\share\hooks.json`} {
		encodedPath, err := json.Marshal(path)
		if err != nil {
			t.Fatal(err)
		}
		payload := `{"sourcePath":` + string(encodedPath) + `,"source":"cloudRequirements","eventName":"preModelRequest","startedAt":0}`
		var run HookRunSummary
		if err := json.Unmarshal([]byte(payload), &run); err != nil {
			t.Fatal(err)
		}
		if run.SourcePath != LegacyAppPathString(path) || run.Source != HookSourceCloudRequirements || run.EventName != HookEventNamePreModelRequest {
			t.Fatalf("hook run changed: %#v", run)
		}
		check0162Fields(t, payload, run, "sourcePath", "source", "eventName", "startedAt")
	}
	var metadata HookMetadata
	payload := `{"handlerType":"prompt","currentHash":"hash","displayOrder":0,"enabled":true,"eventName":"stop","isManaged":true,"key":"key","source":"cloudRequirements","sourcePath":"/tmp/hooks.json","timeoutSec":0,"trustStatus":"trusted","environmentId":"environment","future":9007199254740993}`
	if err := json.Unmarshal([]byte(payload), &metadata); err != nil {
		t.Fatal(err)
	}
	encoded, err := json.Marshal(metadata)
	if err != nil || !metadata.IsKnown() || string(encoded) != payload {
		t.Fatalf("hook metadata changed: %s, %v", encoded, err)
	}
}

func TestCodex0166OtelTLSFields(t *testing.T) {
	payload := `{"ca-certificate":"C:\\certs\\ca.pem","client-certificate":"/tmp/client.pem","client-private-key":"/tmp/key.pem"}`
	var config OtelTlsConfig
	if err := json.Unmarshal([]byte(payload), &config); err != nil {
		t.Fatal(err)
	}
	check0162Fields(t, payload, config, "ca-certificate", "client-certificate", "client-private-key")
	for _, protocol := range []OtelHTTPProtocol{OtelHTTPProtocolBinary, OtelHTTPProtocolJSON} {
		encoded, err := json.Marshal(protocol)
		if err != nil || (string(encoded) != `"binary"` && string(encoded) != `"json"`) {
			t.Fatalf("OTEL protocol: %s, %v", encoded, err)
		}
	}
}

func TestCodex0166ModelProviderErrors(t *testing.T) {
	for _, payload := range []string{
		`{"type":"modelProviderMismatch","modelProvider":"openai","currentModelProvider":"gcp","future":9007199254740993}`,
		`{"type":"modelProviderNotConfigured","modelProvider":"gcp"}`,
		`{"type":"modelProviderPolicyConflict","modelProvider":"openai","requiredModelProvider":"gcp"}`,
		`{"type":"futureError","count":9007199254740993}`,
	} {
		var value ModelProviderError
		if err := json.Unmarshal([]byte(payload), &value); err != nil {
			t.Fatal(err)
		}
		encoded, err := json.Marshal(value)
		if err != nil || string(encoded) != payload || value.IsKnown() != (value.Kind() != "futureError") {
			t.Fatalf("provider error changed: %s, %v", encoded, err)
		}
	}
	for _, payload := range []string{
		`{"type":"modelProviderMismatch","modelProvider":"openai"}`,
		`{"type":"modelProviderNotConfigured"}`,
		`{"type":"modelProviderPolicyConflict","modelProvider":"openai"}`,
	} {
		var value ModelProviderError
		if err := json.Unmarshal([]byte(payload), &value); err == nil {
			t.Fatalf("accepted incomplete provider error: %s", payload)
		}
	}
}
