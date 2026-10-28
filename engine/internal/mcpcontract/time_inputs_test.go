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
			var schema struct {
				Required   []string                   `json:"required"`
				Properties map[string]json.RawMessage `json:"properties"`
			}
			if err := json.Unmarshal(def.InputSchema, &schema); err != nil {
				t.Fatalf("decode schema: %v", err)
			}
			if name == "qimen_chart" {
				// qimen.chart is a oneOf root; both branches declare standard_time.
				return
			}
			for _, field := range []string{"standard_time", "timezone"} {
				if _, ok := schema.Properties[field]; !ok {
					t.Fatalf("missing %s", field)
				}
				if !containsString(schema.Required, field) {
					t.Fatalf("%s must be required", field)
				}
			}
			for _, forbidden := range []string{
				"date", "start_date", "period_date", "birth_year", "birth_date",
				"lunar_year", "target_lunar", "shi_zhi", "year",
			} {
				if _, ok := schema.Properties[forbidden]; ok {
					t.Fatalf("obsolete time input %q must not be exposed", forbidden)
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
