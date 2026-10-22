package mcpcontract

import (
	"context"
	"encoding/json"
	"slices"
	"strings"
	"testing"

	"liki-engine/internal/engine/city"
)

func TestCityCoordsInputRequiresCity(t *testing.T) {
	registry := testEngine()
	definition := mustDefinition(t, registry, "city.coords")
	var schema struct {
		Properties map[string]json.RawMessage `json:"properties"`
		Required   []string                   `json:"required"`
	}
	if err := json.Unmarshal(definition.InputSchema, &schema); err != nil {
		t.Fatalf("unmarshal input schema: %v", err)
	}
	if len(schema.Required) != 1 || schema.Required[0] != "city" {
		t.Fatalf("city required = %#v", schema.Required)
	}
	if _, ok := schema.Properties["candidates"]; ok {
		t.Fatalf("candidates flag must be removed: %s", definition.InputSchema)
	}
}

func TestCityCoordsOutputSchema(t *testing.T) {
	registry := testEngine()
	definition := mustDefinition(t, registry, "city.coords")
	var doc map[string]any
	if err := json.Unmarshal(definition.OutputSchema, &doc); err != nil {
		t.Fatalf("unmarshal output schema: %v", err)
	}
	properties, _ := doc["properties"].(map[string]any)
	if len(properties) != 1 {
		t.Fatalf("output must only expose candidates: %s", definition.OutputSchema)
	}
	candidates, ok := properties["candidates"].(map[string]any)
	if !ok || !schemaTypeContains(candidates["type"], "array") {
		t.Fatalf("candidates must be an array: %s", definition.OutputSchema)
	}
	items, _ := candidates["items"].(map[string]any)
	itemProperties, _ := items["properties"].(map[string]any)
	level, ok := itemProperties["level"].(map[string]any)
	if !ok {
		t.Fatalf("level missing from candidate schema: %s", definition.OutputSchema)
	}
	if enum, _ := level["enum"].([]any); len(enum) != 3 {
		t.Errorf("level enum = %v, want city/county/world: %s", level["enum"], definition.OutputSchema)
	}
	for _, name := range []string{"timezone", "longitude_range"} {
		if _, ok := itemProperties[name]; !ok {
			t.Errorf("%s missing from candidate schema: %s", name, definition.OutputSchema)
		}
	}
	required, _ := doc["required"].([]any)
	if !slices.Contains(required, "candidates") {
		t.Errorf("candidates must be required: %s", definition.OutputSchema)
	}
	for _, name := range []string{"ambiguous", "name", "longitude", "latitude", "country"} {
		if _, ok := properties[name]; ok {
			t.Errorf("%s must not be a top-level field: %s", name, definition.OutputSchema)
		}
	}
}

func TestCityCandidateOutputLongitudeRange(t *testing.T) {
	output := cityCandidateOutput(city.Candidate{
		Name: "北京市", Region: "北京市", Longitude: 116.4, Latitude: 39.9,
		Timezone: "Asia/Shanghai",
		BBox:     &city.BBox{West: 115.42, East: 117.51, South: 39.44, North: 41.06},
	})
	if output.Timezone != "Asia/Shanghai" {
		t.Errorf("timezone = %q", output.Timezone)
	}
	if output.LongitudeRange == nil || output.LongitudeRange.West != 115.42 || output.LongitudeRange.East != 117.51 {
		t.Errorf("longitude range = %+v", output.LongitudeRange)
	}
	withoutBox := cityCandidateOutput(city.Candidate{Name: "某地", Longitude: 1, Latitude: 2})
	if withoutBox.LongitudeRange != nil {
		t.Errorf("missing bbox must omit longitude_range: %+v", withoutBox.LongitudeRange)
	}
}

func TestCityCoordsRejectsUnknownPlace(t *testing.T) {
	registry := testEngine()
	_, err := registry.Call(context.Background(), "city.coords",
		mustJSON(t, map[string]any{"city": "不存在的城市xyz"}))
	if err == nil || !strings.Contains(err.Error(), "未找到") {
		t.Fatalf("err=%v, want not-found error", err)
	}
}

func schemaTypeContains(value any, target string) bool {
	switch typed := value.(type) {
	case string:
		return typed == target
	case []any:
		for _, item := range typed {
			if item == target {
				return true
			}
		}
	}
	return false
}
