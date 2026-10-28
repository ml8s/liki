package main

import (
	"context"
	"encoding/json"
	"regexp"
	"testing"

	"github.com/modelcontextprotocol/go-sdk/mcp"
)

func TestTypedContract_TimeNowSelfDescribing(t *testing.T) {
	sess, _ := startMCPServer(t)
	ctx := context.Background()

	var found *mcp.Tool
	for tool, err := range sess.Tools(ctx, nil) {
		if err != nil {
			t.Fatalf("tools/list: %v", err)
		}
		if tool.Name == "time_now" {
			found = tool
			break
		}
	}
	if found == nil {
		t.Fatal("time_now is not exposed")
	}

	var input struct {
		Type                 string         `json:"type"`
		AdditionalProperties *bool          `json:"additionalProperties"`
		Properties           map[string]any `json:"properties"`
	}
	if err := decodeJSONSchema(t, found.InputSchema, &input); err != nil {
		t.Fatal(err)
	}
	if input.Type != "object" || input.AdditionalProperties == nil || *input.AdditionalProperties {
		t.Fatalf("time_now input schema is not a closed object: %#v", input)
	}

	var output struct {
		Type       string   `json:"type"`
		Required   []string `json:"required"`
		Properties map[string]struct {
			Description string `json:"description"`
		} `json:"properties"`
	}
	if err := decodeJSONSchema(t, found.OutputSchema, &output); err != nil {
		t.Fatal(err)
	}
	if output.Type != "object" {
		t.Fatalf("time_now output type = %q", output.Type)
	}
	for _, name := range []string{"standard_time"} {
		property, ok := output.Properties[name]
		if !ok {
			t.Fatalf("time_now output lacks %q", name)
		}
		if property.Description == "" {
			t.Errorf("time_now output %q has no description", name)
		}
	}
}

func TestTypedContract_TimeNowStructuredOutput(t *testing.T) {
	sess, _ := startMCPServer(t)
	res, err := sess.CallTool(context.Background(), &mcp.CallToolParams{
		Name:      "time_now",
		Arguments: map[string]any{"timezone": "Asia/Shanghai"},
	})
	if err != nil {
		t.Fatalf("call: %v", err)
	}
	if res.IsError {
		t.Fatalf("call returned tool error: %#v", res.Content)
	}
	raw, err := json.Marshal(res.StructuredContent)
	if err != nil {
		t.Fatalf("marshal structured content: %v", err)
	}
	var output struct {
		StandardTime string `json:"standard_time"`
	}
	if err := json.Unmarshal(raw, &output); err != nil {
		t.Fatalf("decode structured content %s: %v", raw, err)
	}
	rfc3339 := regexp.MustCompile(`^\d{4}-\d{2}-\d{2}T\d{2}:\d{2}:\d{2}(?:\.\d+)?(?:Z|[+-]\d{2}:\d{2})$`)
	if !rfc3339.MatchString(output.StandardTime) {
		t.Errorf("standard_time = %q, want RFC3339", output.StandardTime)
	}
}

func decodeJSONSchema(t *testing.T, value any, target any) error {
	t.Helper()
	raw, err := json.Marshal(value)
	if err != nil {
		return err
	}
	return json.Unmarshal(raw, target)
}
