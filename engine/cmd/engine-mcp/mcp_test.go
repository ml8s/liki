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

	apphttp "liki-engine/internal/http"
)

const testBuildID = "test-build-id"

var toolNameRe = regexp.MustCompile(`^[a-z0-9_]+$`)

func newTestLogger() *slog.Logger {
	return slog.New(slog.NewTextHandler(io.Discard, nil))
}

// startMCPServer starts an httptest server exposing the standard MCP
// Streamable HTTP endpoint and returns a connected client session.
func startMCPServer(t *testing.T) (*mcp.ClientSession, *mcp.Server) {
	t.Helper()
	srv, err := newMCPServer(testBuildID, newTestLogger())
	if err != nil {
		t.Fatalf("build MCP server: %v", err)
	}

	ts := httptest.NewServer(apphttp.UnsupportedMethods(mcp.NewStreamableHTTPHandler(func(*http.Request) *mcp.Server {
		return srv
	}, &mcp.StreamableHTTPOptions{Stateless: true, JSONResponse: true}), "subscriptions/listen"))
	t.Cleanup(ts.Close)

	client := mcp.NewClient(&mcp.Implementation{Name: "engine-mcp-test", Version: "0.0.1"}, &mcp.ClientOptions{Logger: newTestLogger()})
	ctx := context.Background()
	sess, err := client.Connect(ctx, &mcp.StreamableClientTransport{Endpoint: ts.URL + "/"}, nil)
	if err != nil {
		t.Fatalf("connect: %v", err)
	}
	t.Cleanup(func() { _ = sess.Close() })
	return sess, srv
}

func TestCallTool_StructuredOnlyOmitsTextContent(t *testing.T) {
	sess, _ := startMCPServer(t)
	res, err := sess.CallTool(context.Background(), &mcp.CallToolParams{
		Name: "time_now",
		Arguments: map[string]any{
			"timezone": "Asia/Shanghai",
		},
	})
	if err != nil {
		t.Fatalf("call time_now: %v", err)
	}
	if res.IsError {
		t.Fatalf("time_now returned IsError")
	}
	if len(res.Content) != 0 {
		t.Fatalf("structured-only result has %d content blocks, want 0", len(res.Content))
	}
	if res.StructuredContent == nil {
		t.Fatal("structured-only result has no StructuredContent")
	}
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
	if got, want := info.Title, serverTitle; got != want {
		t.Errorf("server title = %q, want %q", got, want)
	}
	if got, want := info.Version, testBuildID; got != want {
		t.Errorf("server build ID = %q, want %q", got, want)
	}
}

func TestTools_ListAllRegisteredMethods(t *testing.T) {
	want := len(engineContracts.Names())

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
	for _, name := range engineContracts.Names() {
		if _, ok := tools[strings.ReplaceAll(name, ".", "_")]; !ok {
			t.Errorf("missing typed contract tool %q", name)
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

// callToolStructured invokes a tool and returns the structured content plus
// the IsError flag; successful tool calls are structured-only on the wire.
func callToolStructured(t *testing.T, sess *mcp.ClientSession, name string, args any) (json.RawMessage, bool) {
	t.Helper()
	res, err := sess.CallTool(context.Background(), &mcp.CallToolParams{Name: name, Arguments: args})
	if err != nil {
		t.Fatalf("call %s: %v", name, err)
	}
	if res.IsError {
		if len(res.Content) == 0 {
			t.Fatalf("call %s: error result has no text detail", name)
		}
		return json.RawMessage(res.Content[0].(*mcp.TextContent).Text), true
	}
	if res.StructuredContent == nil {
		t.Fatalf("call %s: success result has no structuredContent", name)
	}
	return json.RawMessage(mustJSONMarshal(t, res.StructuredContent)), false
}

func mustJSONMarshal(t *testing.T, value any) []byte {
	t.Helper()
	raw, err := json.Marshal(value)
	if err != nil {
		t.Fatalf("marshal structured content: %v", err)
	}
	return raw
}

func TestCallTool_BaziChart(t *testing.T) {
	sess, _ := startMCPServer(t)
	text, isErr := callToolStructured(t, sess, "bazi_chart", map[string]any{
		"subject": map[string]any{
			"standard_time": "1984-02-04T06:00:00+08:00", "longitude": 116.4,
			"gender": "male",
		},
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
	// bazi_chart requires subject.standard_time, longitude, and gender; missing args must surface as
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
			if !res.IsError {
				t.Errorf("%s: strict time_now must reject missing timezone", name)
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

func TestTools_OutputSchemaComplete(t *testing.T) {
	sess, _ := startMCPServer(t)
	ctx := context.Background()
	count := 0
	for tool, err := range sess.Tools(ctx, nil) {
		if err != nil {
			t.Fatalf("tools/list: %v", err)
		}
		if tool.OutputSchema == nil {
			t.Errorf("tool %q: outputSchema is missing", tool.Name)
			continue
		}
		if _, ok := tool.OutputSchema.(map[string]any); !ok {
			t.Errorf("tool %q: outputSchema is not an object-like schema: %T", tool.Name, tool.OutputSchema)
			continue
		}
		count++
	}
	if count == 0 {
		t.Fatal("no tools expose outputSchema")
	}
}

// Tools that declare an outputSchema must also return structuredContent.
// Returning JSON only in text blocks violates MCP structured-output promises.
func TestTools_DeclaredOutputSchemaReturnsStructuredContent(t *testing.T) {
	sess, _ := startMCPServer(t)
	ctx := context.Background()

	tools := make(map[string]*mcp.Tool)
	for tool, err := range sess.Tools(ctx, nil) {
		if err != nil {
			t.Fatalf("tools/list: %v", err)
		}
		if tool.OutputSchema != nil {
			tools[tool.Name] = tool
		}
	}
	if len(tools) == 0 {
		t.Fatal("no tools declare outputSchema")
	}

	// time_now is deterministic enough for transport-shape validation and
	// does not require engine-specific fixture data.
	res, err := sess.CallTool(ctx, &mcp.CallToolParams{
		Name:      "time_now",
		Arguments: map[string]any{"timezone": "Asia/Shanghai"},
	})
	if err != nil {
		t.Fatalf("tools/call time_now: %v", err)
	}
	if res.IsError {
		t.Fatalf("tools/call time_now returned IsError")
	}
	if res.StructuredContent == nil {
		t.Fatalf("tool with outputSchema returned no structuredContent")
	}
}

// Standard clients negotiate through initialize. Even though this server is
// configured around the stateless 2026 protocol, it must not advertise a
// version it will subsequently reject in the required protocol header.
func TestStreamableHTTP_AcceptsNegotiatedProtocolHeader(t *testing.T) {
	srv, err := newMCPServer(testBuildID, newTestLogger())
	if err != nil {
		t.Fatalf("build MCP server: %v", err)
	}
	ts := httptest.NewServer(apphttp.UnsupportedMethods(mcp.NewStreamableHTTPHandler(func(*http.Request) *mcp.Server {
		return srv
	}, &mcp.StreamableHTTPOptions{Stateless: true, JSONResponse: true}), "subscriptions/listen"))
	defer ts.Close()

	initialize := `{"jsonrpc":"2.0","id":1,"method":"initialize","params":{"protocolVersion":"2025-11-25","capabilities":{},"clientInfo":{"name":"test","version":"1.0"}}}`
	req, err := http.NewRequest(http.MethodPost, ts.URL+"/mcp", strings.NewReader(initialize))
	if err != nil {
		t.Fatal(err)
	}
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Accept", "application/json, text/event-stream")
	res, err := ts.Client().Do(req)
	if err != nil {
		t.Fatal(err)
	}
	defer func() { _ = res.Body.Close() }()
	body, err := io.ReadAll(res.Body)
	if err != nil {
		t.Fatal(err)
	}
	if res.StatusCode != http.StatusOK {
		t.Fatalf("initialize status=%d body=%s", res.StatusCode, body)
	}
	var document struct {
		Result struct {
			ProtocolVersion string `json:"protocolVersion"`
		} `json:"result"`
	}
	if err := json.Unmarshal(body, &document); err != nil {
		t.Fatalf("decode initialize response %q: %v", body, err)
	}
	negotiated := document.Result.ProtocolVersion
	if negotiated == "" {
		t.Fatalf("initialize returned no protocol version: %s", body)
	}

	list := `{"jsonrpc":"2.0","id":2,"method":"tools/list","params":{}}`
	req, err = http.NewRequest(http.MethodPost, ts.URL+"/mcp", strings.NewReader(list))
	if err != nil {
		t.Fatal(err)
	}
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Accept", "application/json, text/event-stream")
	req.Header.Set("MCP-Protocol-Version", negotiated)
	res, err = ts.Client().Do(req)
	if err != nil {
		t.Fatal(err)
	}
	defer func() { _ = res.Body.Close() }()
	body, err = io.ReadAll(res.Body)
	if err != nil {
		t.Fatal(err)
	}
	if res.StatusCode != http.StatusOK {
		t.Fatalf("tools/list status=%d negotiated=%q body=%s", res.StatusCode, negotiated, body)
	}
	var listDocument struct {
		Result struct {
			Tools []struct {
				Name string `json:"name"`
			} `json:"tools"`
		} `json:"result"`
	}
	if err := json.Unmarshal(body, &listDocument); err != nil {
		t.Fatalf("decode tools/list response %q: %v", body, err)
	}
	if len(listDocument.Result.Tools) == 0 {
		t.Fatalf("tools/list returned no tools: %s", body)
	}
}

func TestAggregateServerExposesAllToolsAndEventHints(t *testing.T) {
	sess, _ := startMCPServer(t)
	tools := map[string]*mcp.Tool{}
	for tool, err := range sess.Tools(context.Background(), nil) {
		if err != nil {
			t.Fatalf("tools/list: %v", err)
		}
		tools[tool.Name] = tool
	}
	for _, name := range []string{
		"bazi_chart", "ziwei_chart", "liuyao_qigua", "liuyao_chart",
		"qimen_chart", "huangli_days", "bazhai_chart", "xuankong_chart",
		"time_now", "subject_candidates", "city_info",
	} {
		if tools[name] == nil {
			t.Fatalf("aggregate surface missing tool %q", name)
		}
	}
	if got := tools["liuyao_qigua"].Annotations; got == nil || got.IdempotentHint {
		t.Fatal("liuyao_qigua must be non-idempotent: auto mode creates a new event")
	}
	if got := tools["liuyao_chart"].Annotations; got == nil || !got.IdempotentHint {
		t.Fatal("liuyao_chart must be idempotent")
	}
	if got := tools["time_now"].Annotations; got == nil || !got.IdempotentHint {
		t.Fatal("time_now must be idempotent")
	}
}

// Modern clients open subscriptions/listen only when the server advertises a
// change stream. The engine tool set is immutable per image, so it must not
// advertise one; if a client asks anyway, the transport rejects it with the
// protocol-mandated 404 JSON-RPC method error.
func TestStreamableHTTP_DoesNotAdvertiseChangeStreams(t *testing.T) {
	srv, err := newMCPServer(testBuildID, newTestLogger())
	if err != nil {
		t.Fatalf("build MCP server: %v", err)
	}
	ts := httptest.NewServer(apphttp.UnsupportedMethods(mcp.NewStreamableHTTPHandler(func(*http.Request) *mcp.Server {
		return srv
	}, &mcp.StreamableHTTPOptions{Stateless: true, JSONResponse: true}), "subscriptions/listen"))
	defer ts.Close()

	meta := `"_meta":{"io.modelcontextprotocol/protocolVersion":"2026-07-28",` +
		`"io.modelcontextprotocol/clientInfo":{"name":"test","version":"1.0"},` +
		`"io.modelcontextprotocol/clientCapabilities":{}}`
	discover := `{"jsonrpc":"2.0","id":1,"method":"server/discover","params":{` + meta + `}}`
	body, status := postMCPJSON(t, ts, discover, "server/discover", "")
	if status != http.StatusOK {
		t.Fatalf("discover status=%d body=%s", status, body)
	}
	var document struct {
		Result struct {
			Capabilities struct {
				Tools struct {
					ListChanged bool `json:"listChanged"`
				} `json:"tools"`
			} `json:"capabilities"`
		} `json:"result"`
	}
	if err := json.Unmarshal([]byte(body), &document); err != nil {
		t.Fatalf("decode discover response %q: %v", body, err)
	}
	if document.Result.Capabilities.Tools.ListChanged {
		t.Fatalf("discover advertised tool list changes: %s", body)
	}

	listen := `{"jsonrpc":"2.0","id":2,"method":"subscriptions/listen","params":{"notifications":{},` + meta + `}}`
	body, status = postMCPJSON(t, ts, listen, "subscriptions/listen", "")
	if status != http.StatusNotFound {
		t.Fatalf("subscriptions/listen status=%d body=%s", status, body)
	}
	var errorDocument struct {
		Error struct {
			Code    int    `json:"code"`
			Message string `json:"message"`
		} `json:"error"`
	}
	if err := json.Unmarshal([]byte(body), &errorDocument); err != nil {
		t.Fatalf("decode listen response %q: %v", body, err)
	}
	if errorDocument.Error.Code != -32601 {
		t.Fatalf("subscriptions/listen error code=%d body=%s", errorDocument.Error.Code, body)
	}
}

func postMCPJSON(t *testing.T, ts *httptest.Server, payload string, method string, name string) (string, int) {
	t.Helper()
	req, err := http.NewRequest(http.MethodPost, ts.URL+"/mcp", strings.NewReader(payload))
	if err != nil {
		t.Fatal(err)
	}
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Accept", "application/json, text/event-stream")
	req.Header.Set("MCP-Protocol-Version", "2026-07-28")
	req.Header.Set("Mcp-Method", method)
	if name != "" {
		req.Header.Set("Mcp-Name", name)
	}
	res, err := ts.Client().Do(req)
	if err != nil {
		t.Fatal(err)
	}
	defer func() { _ = res.Body.Close() }()
	body, err := io.ReadAll(res.Body)
	if err != nil {
		t.Fatal(err)
	}
	return string(body), res.StatusCode
}
