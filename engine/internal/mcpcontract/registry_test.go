package mcpcontract

import (
	"context"

	"encoding/json"
	"fmt"
	"github.com/google/jsonschema-go/jsonschema"
	"strings"
	"testing"
)

func TestEngineRegistry_TimeNowContract(t *testing.T) {
	registry := testEngine()
	definition, err := registry.Definition("time_now")
	if err != nil {
		t.Fatalf("time.now definition: %v", err)
	}
	if definition.Description == "" {
		t.Error("time.now has no description")
	}

	var input struct {
		Type                 string         `json:"type"`
		AdditionalProperties *bool          `json:"additionalProperties"`
		Properties           map[string]any `json:"properties"`
	}
	decode(t, definition.InputSchema, &input)
	if input.Type != "object" || input.AdditionalProperties == nil || *input.AdditionalProperties {
		t.Fatalf("input schema is not closed: %s", definition.InputSchema)
	}

	var output struct {
		Type       string   `json:"type"`
		Required   []string `json:"required"`
		Properties map[string]struct {
			Description string `json:"description"`
		} `json:"properties"`
	}
	decode(t, definition.OutputSchema, &output)
	if output.Type != "object" {
		t.Fatalf("output schema type = %q", output.Type)
	}
	for _, name := range []string{"local_standard_time"} {
		property, ok := output.Properties[name]
		if !ok {
			t.Fatalf("output schema lacks %q: %s", name, definition.OutputSchema)
		}
		if property.Description == "" {
			t.Errorf("output field %q has no description", name)
		}
	}
	for _, name := range output.Required {
		if _, ok := output.Properties[name]; !ok {
			t.Errorf("required output field %q is undeclared", name)
		}
	}
}

func TestEngineRegistry_TimeNowCall(t *testing.T) {
	registry := testEngine()
	_, err := registry.Call(context.Background(), "time_now", json.RawMessage(`{"ignored":true}`))
	if err == nil {
		t.Fatal("unknown input field was accepted")
	}
	if !strings.Contains(err.Error(), "invalid arguments") {
		t.Fatalf("unexpected input error: %v", err)
	}

	raw, err := registry.Call(context.Background(), "time_now", json.RawMessage(`{"timezone":"Asia/Shanghai"}`))
	if err != nil {
		t.Fatalf("call: %v", err)
	}
	var output TimeNowOutput
	if err := json.Unmarshal(raw, &output); err != nil {
		t.Fatalf("decode output: %v", err)
	}
	if output.LocalStandardTime == "" {
		t.Fatalf("incomplete output: %#v", output)
	}
}

func decode(t *testing.T, raw json.RawMessage, target any) {
	t.Helper()
	if err := json.Unmarshal(raw, target); err != nil {
		t.Fatalf("decode JSON %s: %v", raw, err)
	}
}

// ValidateOutput checks an actual JSON result against a tool's published
// output schema. It is intentionally separate from Call: output is produced by
// trusted typed code and is covered by contract tests, so production requests
// do not reconstruct the document for a reflective validation pass.
func (r *Registry) ValidateOutput(name string, output json.RawMessage) error {
	tool, ok := r.tools[name]
	if !ok {
		return fmt.Errorf("unknown tool: %s", name)
	}
	schemas, err := tool.schemas.get()
	if err != nil {
		return fmt.Errorf("mcp tool %s: derive schemas: %w", tool.name, err)
	}
	return schemas.outputValidator.validate(output)
}

func (v *lazyOutputValidator) validate(raw json.RawMessage) error {
	if err := v.resolve(); err != nil {
		return err
	}
	return validateJSONDocument(v.resolved, raw)
}

func validateJSONDocument(resolved *jsonschema.Resolved, raw json.RawMessage) error {
	var document any
	if err := json.Unmarshal(raw, &document); err != nil {
		return err
	}
	if err := resolved.Validate(document); err != nil {
		return fmt.Errorf("output contract violation: %w", err)
	}
	return nil
}
