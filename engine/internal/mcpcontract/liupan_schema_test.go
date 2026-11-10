package mcpcontract

import "testing"

func TestLiupanOutputSchemasFlattenNamedRoots(t *testing.T) {
	registry := testEngine()
	wantRoots := map[string][]string{
		"bazi_liupan":  {"liu_nian", "liu_yue", "liu_ri", "liu_shi", "xiao_yun"},
		"ziwei_liupan": {"liu_nian", "liu_yue", "liu_ri", "liu_shi"},
	}
	for name, roots := range wantRoots {
		t.Run(name, func(t *testing.T) {
			definition := mustDefinition(t, registry, name)
			var schema map[string]any
			decode(t, definition.OutputSchema, &schema)
			if schema["type"] != "object" {
				t.Fatalf("%s output root type = %#v, want object", name, schema["type"])
			}
			if _, exists := schema["oneOf"]; exists {
				t.Fatalf("%s output schema must not use root oneOf", name)
			}
			if schema["additionalProperties"] != false {
				t.Fatalf("%s output properties are not closed", name)
			}
			if minProperties, ok := schema["minProperties"].(float64); !ok || minProperties != 1 {
				t.Fatalf("%s output minProperties = %#v, want 1", name, schema["minProperties"])
			}
			if maxProperties, ok := schema["maxProperties"].(float64); !ok || maxProperties != 1 {
				t.Fatalf("%s output maxProperties = %#v, want 1", name, schema["maxProperties"])
			}
			properties, _ := schema["properties"].(map[string]any)
			if len(properties) != len(roots) {
				t.Fatalf("%s output root count = %d, want %d", name, len(properties), len(roots))
			}
			for _, root := range roots {
				raw, ok := properties[root]
				if !ok {
					t.Fatalf("%s output lacks %q", name, root)
				}
				var property map[string]any
				decode(t, mustJSON(t, raw), &property)
				if property["type"] != "array" {
					t.Fatalf("%s.%s type = %#v, want array", name, root, property["type"])
				}
			}
		})
	}
}
