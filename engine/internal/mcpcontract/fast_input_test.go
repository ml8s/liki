package mcpcontract

import (
	"encoding/json"
	"errors"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/google/jsonschema-go/jsonschema"
)

func TestFastInputValidatorMatchesGenericValidatorOnFixtures(t *testing.T) {
	registry := testEngine()
	entries, err := os.ReadDir(filepath.Join("testdata", "actual"))
	if err != nil {
		t.Fatal(err)
	}
	checked := 0
	for _, entry := range entries {
		if entry.IsDir() || !strings.HasSuffix(entry.Name(), ".json") {
			continue
		}
		raw, err := os.ReadFile(filepath.Join("testdata", "actual", entry.Name()))
		if err != nil {
			t.Fatal(err)
		}
		var fixture struct {
			Input json.RawMessage `json:"input"`
		}
		if err := json.Unmarshal(raw, &fixture); err != nil {
			t.Fatal(err)
		}
		name := fixtureTool(strings.TrimSuffix(entry.Name(), ".json"))
		if name == "" {
			continue
		}
		tool, ok := registry.tools[name]
		if !ok {
			t.Fatalf("%s is not registered", name)
		}
		schemas, err := tool.schemas.get()
		if err != nil {
			t.Fatal(err)
		}
		if !schemas.fastInput {
			continue
		}
		for label, mutation := range inputMutations(t, fixture.Input, schemas.inputTree) {
			t.Run(entry.Name()+"/"+label, func(t *testing.T) {
				compareFastAndGenericInput(t, schemas, mutation)
			})
		}
		compareFastAndGenericInput(t, schemas, fixture.Input)
		checked++
	}
	if checked == 0 {
		t.Fatal("no fast-input fixtures were checked")
	}
}

func TestFastInputScannerJSONSemantics(t *testing.T) {
	stringSchema := &jsonschema.Schema{Type: "string"}
	tests := []struct {
		name   string
		schema *jsonschema.Schema
		raw    string
	}{
		{
			name: "escaped object key resolves to schema property",
			schema: &jsonschema.Schema{
				Type:                 "object",
				Properties:           map[string]*jsonschema.Schema{"chart": stringSchema},
				Required:             []string{"chart"},
				AdditionalProperties: falseSchema(),
			},
			raw: `{"\u0063hart":"complete"}`,
		},
		{
			name:   "schema without type permits scalar",
			schema: &jsonschema.Schema{Properties: map[string]*jsonschema.Schema{"chart": stringSchema}},
			raw:    `"scalar"`,
		},
		{
			name:   "nullable array accepts null",
			schema: &jsonschema.Schema{Types: []string{"null", "array"}, Items: stringSchema},
			raw:    `null`,
		},
		{
			name:   "nullable array accepts items",
			schema: &jsonschema.Schema{Types: []string{"null", "array"}, Items: stringSchema},
			raw:    `["甲","乙"]`,
		},
		{
			name:   "open object accepts additional properties",
			schema: &jsonschema.Schema{Type: "object"},
			raw:    `{"extra":true}`,
		},
		{
			name: "duplicate keys request generic fallback",
			schema: &jsonschema.Schema{
				Type:                 "object",
				Properties:           map[string]*jsonschema.Schema{"chart": stringSchema},
				AdditionalProperties: falseSchema(),
			},
			raw: `{"chart":"first","chart":"second"}`,
		},
		{
			name: "duplicate unknown keys request generic fallback",
			schema: &jsonschema.Schema{
				Type:                 "object",
				Properties:           map[string]*jsonschema.Schema{"chart": stringSchema},
				AdditionalProperties: falseSchema(),
			},
			raw: `{"extra":"first","extra":"second","chart":"complete"}`,
		},
	}
	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			if strings.Contains(tc.name, "duplicate") {
				err := validateFastInput(tc.schema, json.RawMessage(tc.raw), nil)
				if !errors.Is(err, errFastInputFallback) {
					t.Fatalf("duplicate key error = %v, want fallback", err)
				}
				return
			}
			compareFastAndGenericInput(t, derivedSchemas{
				inputTree:     tc.schema,
				inputResolved: resolvedSchemaForTest(t, tc.schema),
			}, json.RawMessage(tc.raw))
		})
	}
}

func TestFastInputFallsBackForUnsupportedNumericBounds(t *testing.T) {
	minimum := 1.0
	exclusiveMinimum := 1.0
	schema := &jsonschema.Schema{
		Type:             "integer",
		Minimum:          &minimum,
		ExclusiveMinimum: &exclusiveMinimum,
	}
	if supportsFastInputSchema(schema) {
		t.Fatal("exclusiveMinimum must use the generic validator")
	}
}

func falseSchema() *jsonschema.Schema {
	return &jsonschema.Schema{Not: &jsonschema.Schema{}}
}

func resolvedSchemaForTest(t *testing.T, schema *jsonschema.Schema) *jsonschema.Resolved {
	t.Helper()
	resolved, err := schema.Resolve(nil)
	if err != nil {
		t.Fatal(err)
	}
	return resolved
}

func compareFastAndGenericInput(t *testing.T, schemas derivedSchemas, raw json.RawMessage) {
	t.Helper()
	raw = normalizeInputRaw(raw)
	fastErr := validateFastInput(schemas.inputTree, raw, nil)
	var document any
	if err := json.Unmarshal(raw, &document); err != nil {
		t.Fatalf("decode mutation: %v", err)
	}
	genericErr := schemas.inputResolved.Validate(document)
	if (fastErr == nil) != (genericErr == nil) {
		t.Fatalf("validator mismatch: fast=%v; generic=%v; input=%s", fastErr, genericErr, raw)
	}
}

func normalizeInputRaw(raw json.RawMessage) json.RawMessage {
	if len(raw) == 0 || string(raw) == "null" {
		return json.RawMessage(`{}`)
	}
	return raw
}

func inputMutations(t *testing.T, raw json.RawMessage, schema *jsonschema.Schema) map[string]json.RawMessage {
	t.Helper()
	var document any
	if err := json.Unmarshal(normalizeInputRaw(raw), &document); err != nil {
		t.Fatal(err)
	}
	out := make(map[string]json.RawMessage)
	collectInputMutations("", document, schema, out)
	return out
}

func collectInputMutations(path string, document any, schema *jsonschema.Schema, out map[string]json.RawMessage) {
	object, ok := document.(map[string]any)
	if !ok || schema == nil {
		return
	}
	for key, value := range object {
		childPath := path + "." + key
		childSchema := schema.Properties[key]

		added := copyJSONObject(object)
		added["__unknown_fast_input__"] = true
		saveInputMutation(out, childPath+"/unknown-parent", added)

		for _, required := range schema.Required {
			if required != key {
				continue
			}
			removed := copyJSONObject(object)
			delete(removed, key)
			saveInputMutation(out, childPath+"/missing-required", removed)
		}
		if childSchema != nil && len(childSchema.Enum) > 0 {
			changed := copyJSONObject(object)
			changed[key] = "__invalid_fast_input_enum__"
			saveInputMutation(out, childPath+"/invalid-enum", changed)
		}
		if childSchema != nil && (childSchema.Type != "" || len(childSchema.Types) > 0) {
			changed := copyJSONObject(object)
			changed[key] = []any{"wrong-type"}
			saveInputMutation(out, childPath+"/wrong-type", changed)
		}
		if childSchema != nil {
			switch childSchema.Format {
			case "date", "date-time":
				changed := copyJSONObject(object)
				changed[key] = "not-a-formatted-time"
				saveInputMutation(out, childPath+"/invalid-format", changed)
			}
			if childSchema.Minimum != nil {
				changed := copyJSONObject(object)
				changed[key] = *childSchema.Minimum - 1
				saveInputMutation(out, childPath+"/below-minimum", changed)
			}
			if childSchema.Maximum != nil {
				changed := copyJSONObject(object)
				changed[key] = *childSchema.Maximum + 1
				saveInputMutation(out, childPath+"/above-maximum", changed)
			}
		}
		collectInputMutations(childPath, value, childSchema, out)
	}
}

func copyJSONObject(value map[string]any) map[string]any {
	out := make(map[string]any, len(value)+1)
	for key, item := range value {
		out[key] = item
	}
	return out
}

func saveInputMutation(out map[string]json.RawMessage, label string, value map[string]any) {
	raw, err := json.Marshal(value)
	if err != nil {
		return
	}
	out[label] = raw
}
