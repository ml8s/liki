package main

import (
	"context"
	"encoding/json"
	"io"
	"log/slog"
	"net/http"
	"net/http/httptest"
	"regexp"
	"strings"
	"testing"

	"github.com/modelcontextprotocol/go-sdk/mcp"

	"liki-engine/internal/agent"
)

const testVersion = "test-version"

var toolNameRe = regexp.MustCompile(`^[a-z0-9_]+$`)

func newTestLogger() *slog.Logger {
	return slog.New(slog.NewTextHandler(io.Discard, nil))
}

// startMCPServer starts an httptest server exposing the standard MCP
// Streamable HTTP endpoint and returns a connected client session.
func startMCPServer(t *testing.T) (*mcp.ClientSession, *mcp.Server) {
	t.Helper()
	reg := agent.NewRPCRegistry()
	reg.SetVersion(testVersion)
	srv := newMCPServer(reg, testVersion, newTestLogger())

	ts := httptest.NewServer(mcp.NewStreamableHTTPHandler(func(*http.Request) *mcp.Server {
		return srv
	}, &mcp.StreamableHTTPOptions{}))
	t.Cleanup(ts.Close)

	client := mcp.NewClient(&mcp.Implementation{Name: "liki-mcp-test", Version: "0.0.1"}, &mcp.ClientOptions{Logger: newTestLogger()})
	ctx := context.Background()
	sess, err := client.Connect(ctx, &mcp.StreamableClientTransport{Endpoint: ts.URL + "/"}, nil)
	if err != nil {
		t.Fatalf("connect: %v", err)
	}
	t.Cleanup(func() { _ = sess.Close() })
	return sess, srv
}

func TestMCPServer_Initialize(t *testing.T) {
	sess, _ := startMCPServer(t)
	info := sess.InitializeResult().ServerInfo
	if info == nil {
		t.Fatalf("missing serverInfo")
	}
	if got, want := info.Name, serverName; got != want {
		t.Errorf("server name = %q, want %q", got, want)
	}
	if got, want := info.Version, testVersion; got != want {
		t.Errorf("server version = %q, want %q", got, want)
	}
}

func TestTools_ListAllRegisteredMethods(t *testing.T) {
	reg := agent.NewRPCRegistry()
	want := len(reg.Names())

	sess, _ := startMCPServer(t)
	tools := make(map[string]*mcp.Tool)
	ctx := context.Background()
	for tool, err := range sess.Tools(ctx, nil) {
		if err != nil {
			t.Fatalf("tools/list: %v", err)
		}
		tools[tool.Name] = tool
	}

	if got := len(tools); got != want {
		t.Fatalf("tool count = %d, want %d", got, want)
	}
	for _, name := range reg.Names() {
		if _, ok := tools[strings.ReplaceAll(name, ".", "_")]; !ok {
			t.Errorf("missing tool for RPC method %q", name)
		}
	}
}

func TestTools_StandardNameAndSchema(t *testing.T) {
	sess, _ := startMCPServer(t)
	ctx := context.Background()
	for tool, err := range sess.Tools(ctx, nil) {
		if err != nil {
			t.Fatalf("tools/list: %v", err)
		}
		if !toolNameRe.MatchString(tool.Name) {
			t.Errorf("tool %q: name must match %v", tool.Name, toolNameRe)
		}
		if strings.Contains(tool.Name, ".") {
			t.Errorf("tool %q: MCP tool names must not contain dots", tool.Name)
		}
		if tool.Description == "" {
			t.Errorf("tool %q: missing description", tool.Name)
		}
		if tool.InputSchema == nil {
			t.Errorf("tool %q: missing inputSchema", tool.Name)
		}
	}
}

// callToolText invokes a tool and returns the text content plus the IsError flag.
func callToolText(t *testing.T, sess *mcp.ClientSession, name string, args any) (string, bool) {
	t.Helper()
	res, err := sess.CallTool(context.Background(), &mcp.CallToolParams{Name: name, Arguments: args})
	if err != nil {
		t.Fatalf("call %s: %v", name, err)
	}
	if len(res.Content) == 0 {
		t.Fatalf("call %s: empty content", name)
	}
	tc, ok := res.Content[0].(*mcp.TextContent)
	if !ok {
		t.Fatalf("call %s: content is %T, want *TextContent", name, res.Content[0])
	}
	return tc.Text, res.IsError
}

func TestCallTool_BaziChart(t *testing.T) {
	sess, _ := startMCPServer(t)
	text, isErr := callToolText(t, sess, "bazi_chart", map[string]any{
		"solar_time": "1984-02-04T06:00:00+08:00",
		"gender":     "male",
	})
	if isErr {
		t.Fatalf("bazi_chart returned IsError: %s", text)
	}
	if !json.Valid([]byte(text)) {
		t.Fatalf("bazi_chart result is not valid JSON: %q", text)
	}
	var doc map[string]any
	if err := json.Unmarshal([]byte(text), &doc); err != nil {
		t.Fatalf("unmarshal result: %v", err)
	}
	if _, hasProduct := doc["_product"]; hasProduct {
		t.Errorf("result leaked RPC _product envelope")
	}
	if _, ok := doc["ri"]; !ok {
		t.Errorf("result missing 日柱 (ri): %v", doc)
	}
}

func TestCallTool_InvalidParamsIsError(t *testing.T) {
	sess, _ := startMCPServer(t)
	// bazi_chart requires solar_time + gender; missing args must surface as
	// an IsError result (visible to the LLM), not a protocol-level error.
	res, err := sess.CallTool(context.Background(), &mcp.CallToolParams{
		Name:      "bazi_chart",
		Arguments: map[string]any{},
	})
	if err != nil {
		t.Fatalf("call: %v", err)
	}
	if !res.IsError {
		t.Errorf("expected IsError for invalid params")
	}
	if len(res.Content) == 0 {
		t.Fatalf("expected error message content")
	}
}

func TestCallTool_NullArguments(t *testing.T) {
	sess, _ := startMCPServer(t)
	// MCP spec allows omitting arguments; a client may also send explicit null.
	// Both must be treated as an empty object.
	for name, args := range map[string]any{
		"omitted":       nil,
		"explicit null": json.RawMessage(`null`),
	} {
		t.Run(name, func(t *testing.T) {
			res, err := sess.CallTool(context.Background(), &mcp.CallToolParams{
				Name:      "time_now",
				Arguments: args,
			})
			if err != nil {
				t.Fatalf("call: %v", err)
			}
			if res.IsError {
				t.Errorf("%s: null/omitted arguments should not be an error: %v", name, res.Content)
			}
		})
	}
}

func TestCallTool_UnknownTool(t *testing.T) {
	sess, _ := startMCPServer(t)
	_, err := sess.CallTool(context.Background(), &mcp.CallToolParams{Name: "no_such_tool", Arguments: map[string]any{}})
	if err == nil {
		t.Fatalf("expected protocol error for unknown tool")
	}
}

// TestCallTool_ConsistencyWithRPC verifies that MCP tool output equals the
// data payload of the same RPC call (envelope stripped, nothing else changed).
func TestCallTool_ConsistencyWithRPC(t *testing.T) {
	reg := agent.NewRPCRegistry()
	reg.SetVersion(testVersion)
	srv := newMCPServer(reg, testVersion, newTestLogger())
	ts := httptest.NewServer(mcp.NewStreamableHTTPHandler(func(*http.Request) *mcp.Server {
		return srv
	}, &mcp.StreamableHTTPOptions{}))
	t.Cleanup(ts.Close)

	client := mcp.NewClient(&mcp.Implementation{Name: "liki-mcp-test", Version: "0.0.1"}, &mcp.ClientOptions{Logger: newTestLogger()})
	ctx := context.Background()
	sess, err := client.Connect(ctx, &mcp.StreamableClientTransport{Endpoint: ts.URL + "/"}, nil)
	if err != nil {
		t.Fatalf("connect: %v", err)
	}
	t.Cleanup(func() { _ = sess.Close() })

	cases := []struct {
		method string
		args   any
	}{
		{"bazi.chart", map[string]any{"solar_time": "1984-02-04T06:00:00+08:00", "gender": "male"}},
		{"time.now", map[string]any{}},
		{"bazhai.chart", map[string]any{"birth_year": 1990, "gender": "male"}},
		{"tianwen.time", map[string]any{"time": "1984-02-04T06:00:00+08:00", "longitude": 116.4}},
	}
	for _, tc := range cases {
		t.Run(tc.method, func(t *testing.T) {
			rpcRaw, err := reg.Execute(ctx, tc.method, mustMarshal(tc.args))
			if err != nil {
				t.Fatalf("rpc %s: %v", tc.method, err)
			}
			wantData, err := unwrapResult(rpcRaw)
			if err != nil {
				t.Fatalf("unwrap rpc %s: %v", tc.method, err)
			}
			var want any
			if err := json.Unmarshal(wantData, &want); err != nil {
				t.Fatalf("rpc %s data is not JSON: %v", tc.method, err)
			}

			text, isErr := callToolText(t, sess, toolName(tc.method), tc.args)
			if isErr {
				t.Fatalf("mcp %s returned IsError: %s", tc.method, text)
			}
			var got any
			if err := json.Unmarshal([]byte(text), &got); err != nil {
				t.Fatalf("mcp %s result not JSON: %v", tc.method, err)
			}
			if !jsonEqual(got, want) {
				t.Errorf("mcp %s result != rpc %s data", tc.method, tc.method)
			}
		})
	}
}

func TestUnwrapResult(t *testing.T) {
	raw := json.RawMessage(`{"_product":"chart","data":{"ri":{"gan":"戊"}}}`)
	got, err := unwrapResult(raw)
	if err != nil {
		t.Fatalf("unwrap: %v", err)
	}
	if string(got) != `{"ri":{"gan":"戊"}}` {
		t.Errorf("unwrap = %s", got)
	}

	if _, err := unwrapResult(json.RawMessage(`{"data":null}`)); err == nil {
		t.Errorf("expected error for null data")
	}
	if _, err := unwrapResult(json.RawMessage(`{}`)); err == nil {
		t.Errorf("expected error for missing data")
	}
}

func mustMarshal(v any) json.RawMessage {
	b, err := json.Marshal(v)
	if err != nil {
		panic(err)
	}
	return b
}

func jsonEqual(a, b any) bool {
	ab, _ := json.Marshal(a)
	bb, _ := json.Marshal(b)
	return string(ab) == string(bb)
}

func TestTools_OutputSchemaComplete(t *testing.T) {
	// object 结构的工具必须有 outputSchema（从 RPC Result envelope 提取）；
	// array/primitive 返回的工具按 MCP 规范省略（description 承担返回说明）。
	sess, _ := startMCPServer(t)
	ctx := context.Background()
	objectCount := 0
	withOutput := 0
	for tool, err := range sess.Tools(ctx, nil) {
		if err != nil {
			t.Fatalf("tools/list: %v", err)
		}
		if tool.OutputSchema == nil {
			continue
		}
		withOutput++
		if os, ok := tool.OutputSchema.(map[string]any); ok {
			if os["type"] != "object" {
				t.Errorf("tool %q: outputSchema.type = %v, want object", tool.Name, os["type"])
			}
			objectCount++
		}
	}
	if withOutput == 0 {
		t.Fatal("no tools expose outputSchema")
	}
	if objectCount != withOutput {
		t.Errorf("non-object outputSchema found: %d/%d", withOutput-objectCount, withOutput)
	}
}
