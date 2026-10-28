package mcpcontract

import (
	"encoding/json"
	"testing"
)

// TestTimeInputsCanonical freezes the contract: time-bearing domain tools take
// standard_time (plus longitude where LAST is required). Calendar dates
// are projections, not inputs.
func TestTimeInputsCanonical(t *testing.T) {
	registry := testEngine()
	tools := []string{
		"bazi_chart", "ziwei_chart", "liuyao_chart", "qimen_chart",
		"bazi_liunian", "bazi_liuyue", "bazi_liuri", "bazi_liushi",
		"ziwei_liunian", "ziwei_liuyue", "ziwei_liuri", "ziwei_liushi",
		"huangli_days", "bazhai_chart", "xuankong_chart", "xuankong_liunian",
	}
	for _, name := range tools {
		t.Run(name, func(t *testing.T) {
			def, err := registry.Definition(name)
			if err != nil {
				t.Fatalf("definition: %v", err)
			}
			if len(def.InputSchema) == 0 {
				t.Fatal("missing input schema")
			}
			type inputSchema struct {
				Required   []string                   `json:"required"`
				Properties map[string]json.RawMessage `json:"properties"`
			}
			var root inputSchema
			var rootUnion struct {
				OneOf []json.RawMessage `json:"oneOf"`
			}
			if err := json.Unmarshal(def.InputSchema, &root); err != nil {
				t.Fatalf("decode schema: %v", err)
			}
			if err := json.Unmarshal(def.InputSchema, &rootUnion); err != nil {
				t.Fatalf("decode schema union: %v", err)
			}
			schemas := []inputSchema{root}
			for _, raw := range rootUnion.OneOf {
				var branch inputSchema
				if err := json.Unmarshal(raw, &branch); err != nil {
					t.Fatalf("decode oneOf branch: %v", err)
				}
				schemas = append(schemas, branch)
			}
			if len(rootUnion.OneOf) > 0 {
				schemas = schemas[1:]
			}

			for schemaIndex, schema := range schemas {
				if _, ok := schema.Properties["standard_time"]; !ok {
					t.Fatalf("schema[%d]: missing standard_time", schemaIndex)
				}
				if !containsString(schema.Required, "standard_time") {
					t.Fatalf("schema[%d]: standard_time must be required", schemaIndex)
				}
				for _, forbidden := range []string{
					"instant", "civil_time", "timezone", "fold",
					"date", "start_date", "period_date", "birth_year", "birth_date",
					"lunar_year", "target_lunar", "shi_zhi", "year",
				} {
					if _, ok := schema.Properties[forbidden]; ok {
						t.Fatalf("schema[%d]: obsolete time input %q must not be exposed", schemaIndex, forbidden)
					}
				}
			}
		})
	}
}

func containsString(values []string, target string) bool {
	for _, value := range values {
		if value == target {
			return true
		}
	}
	return false
}
