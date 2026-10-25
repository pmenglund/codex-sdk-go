package main

import "github.com/atombender/go-jsonschema/pkg/schemas"

// protocolSchemaLoader supplies reviewed Go representations without changing
// the upstream schemas used by the inventory and manual-type checks.
type protocolSchemaLoader struct {
	schemas.Loader
}

// Load supplies lossless MCP metadata and the native feedback-save path type.
func (loader protocolSchemaLoader) Load(uri, parentURI string) (*schemas.Schema, error) {
	schema, err := loader.Loader.Load(uri, parentURI)
	if err != nil {
		return nil, err
	}
	status := protocolSchemaObject(schema, "McpServerStatus")
	if status != nil {
		if metadata := status.Properties["toolListMeta"]; metadata != nil {
			rawType := "json.RawMessage"
			metadata.GoJSONSchemaExtension = &schemas.GoJSONSchemaExtension{
				Type: &rawType, Nillable: true, Imports: []string{"encoding/json"},
			}
		}
	}
	if feedback := protocolSchemaObject(schema, "FeedbackSaveParams"); feedback != nil {
		if path := feedback.Properties["savePath"]; path != nil {
			// Schema sanitization strips this property's allOf reference. Restore
			// its wire string type instead of exposing an opaque interface{}.
			pathType := "LegacyAppPathString"
			path.GoJSONSchemaExtension = &schemas.GoJSONSchemaExtension{Type: &pathType}
		}
	}
	return schema, nil
}

func protocolSchemaObject(schema *schemas.Schema, name string) *schemas.Type {
	if schema.Title == name {
		return (*schemas.Type)(schema.ObjectAsType)
	}
	return schema.Definitions[name]
}
