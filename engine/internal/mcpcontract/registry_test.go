package mcpcontract

import (
	"context"
	"encoding/json"
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
	for _, name := range []string{"standard_time"} {
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
	if output.StandardTime == "" {
		t.Fatalf("incomplete output: %#v", output)
	}
}

func decode(t *testing.T, raw json.RawMessage, target any) {
	t.Helper()
	if err := json.Unmarshal(raw, target); err != nil {
		t.Fatalf("decode JSON %s: %v", raw, err)
	}
}
