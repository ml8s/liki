package mcpcontract

import (
	"encoding/json"
	"strings"
	"testing"
)

// TestAllOutputSchemasRootObject locks the MCP outputSchema contract: wire
// models through protocol 2025-11-25 reject any outputSchema whose root is not
// type "object", so a top-level union or array schema would break tools/list
// for standard SDK clients (the experts runtime negotiates those versions).
func TestAllOutputSchemasRootObject(t *testing.T) {
	registry := testEngine()
	for _, name := range registry.Names() {
		definition := mustDefinition(t, registry, name)
		var schema struct {
			Type  json.RawMessage   `json:"type"`
			AnyOf []json.RawMessage `json:"anyOf"`
		}
		if err := json.Unmarshal(definition.OutputSchema, &schema); err != nil {
			t.Fatalf("%s: unmarshal output schema: %v", name, err)
		}
		root := strings.Trim(string(schema.Type), `"`)
		if root != "object" {
			if root == "" && len(schema.AnyOf) > 0 {
				root = "anyOf"
			}
			t.Errorf("%s: output schema root = %q, want object", name, root)
		}
	}
}
