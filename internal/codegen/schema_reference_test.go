package main

import (
	"encoding/json"
	"reflect"
	"testing"
)

func TestResolveSchemaReference(t *testing.T) {
	target := map[string]any{"type": "string"}
	root := map[string]any{"defs": map[string]any{"a/b": map[string]any{"~key": target}, "~1": target, "scalar": 42}}
	for _, test := range []struct {
		name, ref string
		want      bool
	}{
		{"nested_escaped", "#/defs/a~1b/~0key", true},
		{"escape_order", "#/defs/~01", true},
		{"missing", "#/defs/missing", false},
		{"scalar_target", "#/defs/scalar", false},
		{"traverse_scalar", "#/defs/scalar/child", false},
	} {
		t.Run(test.name, func(t *testing.T) {
			got, ok := resolveSchemaReference(root, test.ref)
			if ok != test.want {
				t.Fatalf("resolved = %v, want %v", ok, test.want)
			}
			if test.want && !reflect.DeepEqual(got, target) {
				t.Fatalf("target = %#v", got)
			}
		})
	}
}

func TestSchemaObjectFieldsThroughReferencesAndComposition(t *testing.T) {
	for _, test := range []struct {
		name, schema         string
		properties, required []string
	}{
		{"nested_and_shared_references", `{"$ref":"#/defs/child","defs":{"base":{"properties":{"base":{}},"required":["base"]},"child":{"allOf":[{"$ref":"#/defs/base"},{"$ref":"#/defs/base"}],"properties":{"child":{}},"required":["child"]}}}`, []string{"base", "child"}, []string{"base", "child"}},
		{"reference_cycle", `{"$ref":"#/defs/a","defs":{"a":{"$ref":"#/defs/b","properties":{"a":{}},"required":["a"]},"b":{"$ref":"#/defs/a","properties":{"b":{}},"required":["b"]}}}`, []string{"a", "b"}, []string{"a", "b"}},
		{"composite_required_semantics", `{"properties":{"own":{}},"required":["own",123],"allOf":[{"properties":{"always":{}},"required":["always"]},null],"anyOf":[{"properties":{"choice":{}},"required":["choice"]}],"oneOf":[{"properties":{"alternative":{}},"required":["alternative"]}]}`, []string{"own", "always", "choice", "alternative"}, []string{"own", "always"}},
		{"missing_reference_keeps_local_fields", `{"$ref":"#/defs/missing","properties":{"local":{}},"required":["local"]}`, []string{"local"}, []string{"local"}},
		{"external_reference_is_not_resolved", `{"$ref":"other.json#/defs/base","properties":{"local":{}}}`, []string{"local"}, nil},
		{"escaped_reference", `{"$ref":"#/defs/a~1b~0c","defs":{"a/b~c":{"properties":{"inherited":{}},"required":["inherited"]}}}`, []string{"inherited"}, []string{"inherited"}},
	} {
		t.Run(test.name, func(t *testing.T) {
			var root map[string]any
			if err := json.Unmarshal([]byte(test.schema), &root); err != nil {
				t.Fatal(err)
			}
			for _, check := range []struct {
				name string
				got  map[string]struct{}
				want []string
			}{
				{"properties", schemaObjectProperties(root, root, make(map[string]bool)), test.properties},
				{"required", schemaObjectRequired(root, root, make(map[string]bool)), test.required},
			} {
				want := make(map[string]struct{})
				for _, name := range check.want {
					want[name] = struct{}{}
				}
				if !reflect.DeepEqual(check.got, want) {
					t.Fatalf("%s = %v, want %v", check.name, check.got, want)
				}
			}
		})
	}
}
