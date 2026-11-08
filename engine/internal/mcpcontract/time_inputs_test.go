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
				for _, field := range []string{"standard_time", "longitude", "gender"} {
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
				var fields []string
				switch name {
				case "huangli_days":
					fields = []string{"standard_time"}
				case "xuankong_chart", "xuankong_liunian":
					// 玄空用立春事件 + 坐向，不使用真太阳时；不收无关经度。
					fields = []string{"standard_time"}
				case "bazhai_chart":
					fields = []string{"subject"}
				default:
					fields = []string{"standard_time", "longitude"}
				}
				for _, field := range fields {
					if _, ok := schema.Properties[field]; !ok {
						t.Fatalf("schema[%d]: missing %s", schemaIndex, field)
					}
					if !containsString(schema.Required, field) {
						t.Fatalf("schema[%d]: %s must be required", schemaIndex, field)
					}
				}
				for _, forbidden := range []string{
					"instant", "civil_time", "fold",
					"period_date", "birth_year", "birth_date",
					"lunar_year", "target_lunar", "shi_zhi", "year",
				} {
					// 黄历是民用日期范围查询，前端显式选择 timezone 后投影 standard_time。
					if name == "huangli_days" && forbidden == "standard_time" {
						continue
					}
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
