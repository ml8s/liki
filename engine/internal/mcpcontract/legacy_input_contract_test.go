package mcpcontract

import (
	"encoding/json"
	"fmt"
	"slices"
	"testing"
)

// TestPublishedInputSchemasRejectLegacyFields freezes the flattened contract at
// the schema boundary. It catches a reintroduced wrapper or field alias even in
// a branch not covered by a hand-written golden case.
func TestPublishedInputSchemasRejectLegacyFields(t *testing.T) {
	forbidden := map[string]string{
		"birth_time":       "use subject.standard_time",
		"birth_longitude":  "use subject.longitude",
		"qi_gua":           "use flat standard_time/longitude/casting",
		"qi_gua_time":      "use standard_time",
		"qi_gua_longitude": "use longitude",
		"qi_ju":            "use flat standard_time/longitude",
		"qi_ju_time":       "use standard_time",
		"qi_ju_longitude":  "use longitude",
		"location":         "use longitude",
		"place_longitude":  "use longitude",
		"start_time":       "use standard_time",
		"zhai":             "use flat standard_time/longitude/zuo_shan/xiang_shan",
		"yun_start_time":   "use standard_time",
		"civil_time":       "use standard_time",
		"recorded_fold":    "use subject_candidates candidates",
	}

	registry := testEngine()
	names := registry.Names()
	slices.Sort(names)
	if len(names) == 0 {
		t.Fatal("registry has no tools")
	}

	var walk func(path string, value any) []string
	walk = func(path string, value any) []string {
		var violations []string
		object, ok := value.(map[string]any)
		if !ok {
			return violations
		}
		for key, child := range object {
			if reason, blocked := forbidden[key]; blocked {
				violations = append(violations, fmt.Sprintf("%s.%s: %s", path, key, reason))
			}
			violations = append(violations, walk(path+"."+key, child)...)
		}
		return violations
	}

	for _, name := range names {
		t.Run(name, func(t *testing.T) {
			definition, err := registry.Definition(name)
			if err != nil {
				t.Fatalf("definition: %v", err)
			}
			var schema any
			if err := json.Unmarshal(definition.InputSchema, &schema); err != nil {
				t.Fatalf("decode input schema: %v", err)
			}
			if violations := walk("$", schema); len(violations) != 0 {
				t.Fatalf("legacy input fields:\n%s", violations)
			}
		})
	}
}

// TestDomainIrrelevantToolsRejectLongitude prevents event tools from absorbing
// a location fact when their projection never uses local apparent solar time.
func TestDomainIrrelevantToolsRejectLongitude(t *testing.T) {
	for _, name := range []string{"huangli_days", "bazhai_chart", "xuankong_chart", "xuankong_liunian"} {
		t.Run(name, func(t *testing.T) {
			definition, err := testEngine().Definition(name)
			if err != nil {
				t.Fatalf("definition: %v", err)
			}
			var schema any
			if err := json.Unmarshal(definition.InputSchema, &schema); err != nil {
				t.Fatalf("decode input schema: %v", err)
			}
			var walk func(path string, value any)
			walk = func(path string, value any) {
				object, ok := value.(map[string]any)
				if !ok {
					return
				}
				if _, exists := object["longitude"]; exists {
					t.Fatalf("%s.longitude must not exist", path)
				}
				for key, child := range object {
					walk(path+"."+key, child)
				}
			}
			walk("$", schema)
		})
	}
}
