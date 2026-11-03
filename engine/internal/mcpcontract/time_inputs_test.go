package mcpcontract

import (
	"encoding/json"
	"testing"
)

// TestTimeInputsCanonical freezes the contract: natal methods take a resolved
// subject; event and property tools use domain-owned time/location fields.
func TestTimeInputsCanonical(t *testing.T) {
	registry := testEngine()
	natal := []string{
		"bazi_chart", "ziwei_chart", "bazi_liupan", "ziwei_liupan",
	}
	tools := []string{
		"liuyao_chart", "qimen_chart", "huangli_days", "bazhai_chart",
		"xuankong_chart", "xuankong_liunian",
	}
	for _, name := range natal {
		t.Run(name, func(t *testing.T) {
			def, err := registry.Definition(name)
			if err != nil {
				t.Fatalf("definition: %v", err)
			}
			if len(def.InputSchema) == 0 {
				t.Fatal("missing input schema")
			}
			var subjectSchema struct {
				Required   []string                   `json:"required"`
				Properties map[string]json.RawMessage `json:"properties"`
			}
			var schemaUnion struct {
				OneOf []json.RawMessage `json:"oneOf"`
			}
			if err := json.Unmarshal(def.InputSchema, &subjectSchema); err != nil {
				t.Fatalf("decode schema: %v", err)
			}
			if err := json.Unmarshal(def.InputSchema, &schemaUnion); err != nil {
				t.Fatalf("decode schema union: %v", err)
			}
			subjectSchemas := []struct {
				Required   []string                   `json:"required"`
				Properties map[string]json.RawMessage `json:"properties"`
			}{subjectSchema}
			for _, raw := range schemaUnion.OneOf {
				var branch struct {
					Required   []string                   `json:"required"`
					Properties map[string]json.RawMessage `json:"properties"`
				}
				if err := json.Unmarshal(raw, &branch); err != nil {
					t.Fatalf("decode schema branch: %v", err)
				}
				subjectSchemas = append(subjectSchemas, branch)
			}
			if len(schemaUnion.OneOf) > 0 {
				subjectSchemas = subjectSchemas[1:]
			}
			if len(subjectSchemas) == 0 {
				t.Fatalf("missing subject")
			}
			for schemaIndex, branch := range subjectSchemas {
				if _, ok := branch.Properties["subject"]; !ok {
					t.Fatalf("schema[%d]: missing subject", schemaIndex)
				}
				if !containsString(branch.Required, "subject") {
					t.Fatalf("schema[%d]: subject must be required", schemaIndex)
				}
				var subject struct {
					Required   []string `json:"required"`
					Properties map[string]struct {
						Type string `json:"type"`
					} `json:"properties"`
				}
				if err := json.Unmarshal(branch.Properties["subject"], &subject); err != nil {
					t.Fatal(err)
				}
				for _, field := range []string{"birth_time", "birth_longitude", "gender"} {
					if _, ok := subject.Properties[field]; !ok {
						t.Fatalf("subject missing %s", field)
					}
					if !containsString(subject.Required, field) {
						t.Fatalf("subject.%s must be required", field)
					}
				}
			}
		})
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
				timeObject := map[string]string{
					"liuyao_chart": "qi_gua", "qimen_chart": "qi_ju",
					"xuankong_chart": "zhai",
				}[name]
				if timeObject != "" {
					object, ok := schema.Properties[timeObject]
					if !ok {
						t.Fatalf("schema[%d]: missing %s", schemaIndex, timeObject)
					}
					if !containsString(schema.Required, timeObject) {
						t.Fatalf("schema[%d]: %s must be required", schemaIndex, timeObject)
					}
					var nested struct {
						Required   []string `json:"required"`
						Properties map[string]struct {
							Type json.RawMessage `json:"type"`
						} `json:"properties"`
					}
					if err := json.Unmarshal(object, &nested); err != nil {
						t.Fatalf("decode %s: %v", timeObject, err)
					}
					fields := []string{"longitude"}
					switch timeObject {
					case "qi_gua":
						fields = []string{"qi_gua_time", "longitude"}
					case "qi_ju":
						fields = []string{"qi_ju_time", "longitude"}
					case "zhai":
						fields = []string{"yun_start_time", "zuo_shan", "xiang_shan"}
					}
					for _, field := range fields {
						if _, ok := nested.Properties[field]; !ok {
							t.Fatalf("%s missing %s", timeObject, field)
						}
						if !containsString(nested.Required, field) {
							t.Fatalf("%s.%s must be required", timeObject, field)
						}
					}
				} else {
					timeField := map[string]string{
						"huangli_days":     "start_date",
						"bazhai_chart":     "subject",
						"xuankong_liunian": "start_time",
					}[name]
					if _, ok := schema.Properties[timeField]; !ok {
						t.Fatalf("schema[%d]: missing %s", schemaIndex, timeField)
					}
					if !containsString(schema.Required, timeField) {
						t.Fatalf("schema[%d]: %s must be required", schemaIndex, timeField)
					}
				}
				for _, forbidden := range []string{
					"instant", "civil_time", "fold", "standard_time",
					"period_date", "birth_year", "birth_date",
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
