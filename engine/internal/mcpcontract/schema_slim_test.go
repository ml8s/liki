package mcpcontract

import (
	"encoding/json"
	"testing"
)

func TestSlimSchemaRawCompactsNullableEnums(t *testing.T) {
	raw := json.RawMessage(`{
		"type":"object",
		"additionalProperties":false,
		"properties":{
			"a":{"description":"甲。","anyOf":[{"type":"null"},{"type":"string","enum":["甲","乙"]}]},
			"b":{"description":"乙。","anyOf":[{"type":"null"},{"type":"string","enum":["甲","乙"]}]}
		}
	}`)
	slimmed, err := slimSchemaRaw(raw)
	if err != nil {
		t.Fatal(err)
	}
	var schema struct {
		Properties map[string]map[string]any `json:"properties"`
		Defs       map[string]map[string]any `json:"$defs"`
	}
	if err := json.Unmarshal(slimmed, &schema); err != nil {
		t.Fatal(err)
	}
	for name, property := range schema.Properties {
		ref, ok := property["$ref"].(string)
		if !ok || ref == "" {
			t.Fatalf("%s has no enum $ref: %#v", name, property)
		}
		if property["description"] == "" {
			t.Fatalf("%s lost its field description", name)
		}
	}
	if len(schema.Defs) != 1 {
		t.Fatalf("shared enum definitions = %d, want 1", len(schema.Defs))
	}
	for _, definition := range schema.Defs {
		types, ok := definition["type"].([]any)
		if !ok || len(types) != 2 || types[0] != "null" || types[1] != "string" {
			t.Fatalf("shared schema type = %#v, want [null string]", definition["type"])
		}
		values, ok := definition["enum"].([]any)
		if !ok || len(values) != 2 || values[0] != "甲" || values[1] != "乙" {
			t.Fatalf("shared enum = %#v", definition["enum"])
		}
	}
}

func TestSlimSchemaRawKeepsUniqueEnumsInline(t *testing.T) {
	raw := json.RawMessage(`{
		"type":"object",
		"properties":{
			"a":{"description":"甲。","type":"string","enum":["甲"]}
		}
	}`)
	slimmed, err := slimSchemaRaw(raw)
	if err != nil {
		t.Fatal(err)
	}
	var schema map[string]any
	if err := json.Unmarshal(slimmed, &schema); err != nil {
		t.Fatal(err)
	}
	if _, exists := schema["$defs"]; exists {
		t.Fatalf("unique enum was hoisted: %s", slimmed)
	}
}
