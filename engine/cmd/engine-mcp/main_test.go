package main

import (
	"context"
	"encoding/json"
	"strings"
	"testing"

	"github.com/modelcontextprotocol/go-sdk/mcp"
	"liki-engine/internal/observability"
)

func TestContractToolHandlerRecordsMetrics(t *testing.T) {
	observability.ResetMetrics()
	handler := contractToolHandler(engineContracts, "bazi_chart")
	request := &mcp.CallToolRequest{Params: &mcp.CallToolParamsRaw{}}
	request.Params.Arguments = json.RawMessage(`{
		"standard_time": "1990-05-20T04:00:00Z",
		"longitude": 116.4,
		"gender": "male"
	}`)

	result, err := handler(context.Background(), request)
	if err != nil {
		t.Fatal(err)
	}
	if result.IsError {
		t.Fatalf("tool returned error: %s", result.Content[0].(*mcp.TextContent).Text)
	}

	text := string(observability.RenderMetrics())
	for _, fragment := range []string{
		`domain="bazi",tool="bazi_chart",status="success"`,
		"liki_engine_tool_request_duration_seconds",
	} {
		if !strings.Contains(text, fragment) {
			t.Fatalf("metrics missing %q:\n%s", fragment, text)
		}
	}
}

func TestContractToolHandlerRecordsInvalidArgumentsMetrics(t *testing.T) {
	observability.ResetMetrics()
	handler := contractToolHandler(engineContracts, "bazi_chart")
	request := &mcp.CallToolRequest{Params: &mcp.CallToolParamsRaw{}}
	request.Params.Arguments = json.RawMessage(`{"longitude":500,"gender":"male"}`)

	result, err := handler(context.Background(), request)
	if err != nil {
		t.Fatal(err)
	}
	if !result.IsError {
		t.Fatal("invalid arguments should return tool error")
	}

	text := string(observability.RenderMetrics())
	if !strings.Contains(text, `tool="bazi_chart",status="invalid_arguments"`) {
		t.Fatalf("invalid arguments metric missing:\n%s", text)
	}
}
