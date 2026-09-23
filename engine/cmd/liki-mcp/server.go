package main

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"log/slog"
	"os"
	"strings"

	"github.com/modelcontextprotocol/go-sdk/mcp"

	"liki-engine/internal/agent"
)

const (
	serverName = "liki-mcp"
	serverDesc = "liki.hk Metaphysics Engine — deterministic Chinese metaphysics computation: bazi, ziwei, liuyao, qimen, huangli, fengshui and naming."
)

// toolName converts an RPC method name to a valid MCP tool name.
// RPC method names use dot-separated domains (bazi.chart) while MCP tool names
// only allow [a-zA-Z0-9_-], so each dot is replaced with an underscore.
func toolName(method string) string {
	return strings.ReplaceAll(method, ".", "_")
}

// mcpDomains maps an MCP endpoint suffix to the RPC method prefixes it exposes.
// Each domain is a self-contained tool set (one 术数 or shared aux) so the LLM
// only sees relevant tools per expert.
var mcpDomains = []struct {
	Suffix   string
	Prefixes []string
}{
	{"bazi", []string{"bazi."}},
	{"ziwei", []string{"ziwei."}},
	{"qimen", []string{"qimen."}},
	{"liuyao", []string{"liuyao."}},
	{"aux", []string{"time.", "tianwen.", "city."}},
}

// newMCPServer builds a standard MCP server exposing every registered RPC
// method as an MCP tool. RPC-specific envelopes ({"_product","data"}) are
// unwrapped: tools return the raw engine data, not the private RPC envelope.
func newMCPServer(reg *agent.RPCRegistry, version string, logger *slog.Logger) *mcp.Server {
	return newMCPServerFor(reg, version, logger, func(string) bool { return true })
}

// newDomainServer builds an MCP server exposing only the methods matching the
// given RPC method prefixes (one 术数 domain or the shared aux tools).
func newDomainServer(reg *agent.RPCRegistry, prefixes []string, version string, logger *slog.Logger) *mcp.Server {
	return newMCPServerFor(reg, version, logger, func(method string) bool {
		for _, p := range prefixes {
			if strings.HasPrefix(method, p) {
				return true
			}
		}
		return false
	})
}

func newMCPServerFor(reg *agent.RPCRegistry, version string, logger *slog.Logger, match func(string) bool) *mcp.Server {
	impl := &mcp.Implementation{
		Name:        serverName,
		Title:       "Liki Metaphysics Engine",
		Description: serverDesc,
		Version:     version,
		WebsiteURL:  "https://liki.hk",
	}
	opts := &mcp.ServerOptions{
		Instructions: "Deterministic Chinese metaphysics computation. Gather birth info (solar time, gender, birth place) via tianwen_time/city_coords first, then compute charts (bazi_chart, ziwei_chart, liuyao_qigua, qimen_chart, bazhai_chart, xuankong_chart, huangli_days) and interpret their results. Conclusions are conditioned readings from a traditional culture perspective, not medical, legal or investment advice.",
		Logger:       logger,
		// 干净设计：只支持新协议（2026-07-28，server/discover 协商），不做旧协议向后兼容。
		SupportedProtocolVersions: []string{"2026-07-28"},
	}
	s := mcp.NewServer(impl, opts)

	for _, name := range reg.Names() {
		if !match(name) {
			continue
		}
		m, ok := reg.Method(name)
		if !ok {
			continue
		}
		tool := &mcp.Tool{
			Name:        toolName(name),
			Description: m.Description,
			InputSchema: m.Params,
		}
		if os := outputSchemaFromResult(m.Result); len(os) > 0 {
			tool.OutputSchema = os
		}
		s.AddTool(tool, toolHandler(reg, name))
	}
	return s
}

// outputSchemaFromResult extracts the engine data schema from the RPC result
// envelope ({"_product":...,"data":<schema>}). MCP tools return the unwrapped
// data, so the data sub-schema becomes the tool's outputSchema, giving the
// LLM the same structural knowledge the OpenRPC document provides. Only
// object-typed schemas are exposed: MCP conformance requires outputSchema to
// be an object schema, so array/primitive-typed returns omit it (the tool
// description carries their shape). Returns nil when absent or unsuitable.
func outputSchemaFromResult(result json.RawMessage) json.RawMessage {
	if len(result) == 0 {
		return nil
	}
	var env struct {
		Properties struct {
			Data json.RawMessage `json:"data"`
		} `json:"properties"`
	}
	if err := json.Unmarshal(result, &env); err != nil || len(env.Properties.Data) == 0 {
		return nil
	}
	var doc struct {
		Type string `json:"type"`
	}
	if err := json.Unmarshal(env.Properties.Data, &doc); err != nil || doc.Type != "object" {
		return nil
	}
	return env.Properties.Data
}

// toolHandler maps a registered RPC method to an MCP tool handler. Business
// errors surface as IsError results (visible to the LLM); only internal
// failures become protocol-level errors.
func toolHandler(reg *agent.RPCRegistry, method string) mcp.ToolHandler {
	return func(ctx context.Context, req *mcp.CallToolRequest) (*mcp.CallToolResult, error) {
		args := req.Params.Arguments
		if len(args) == 0 || string(args) == "null" {
			args = json.RawMessage(`{}`)
		}
		result, err := reg.Execute(ctx, method, args)
		if err != nil {
			msg := err.Error()
			var rpcErr *agent.RPCError
			if errors.As(err, &rpcErr) {
				msg = rpcErr.Message
			}
			return &mcp.CallToolResult{
				IsError: true,
				Content: []mcp.Content{&mcp.TextContent{Text: msg}},
			}, nil
		}
		data, err := unwrapResult(result)
		if err != nil {
			return nil, err
		}
		return &mcp.CallToolResult{
			Content: []mcp.Content{&mcp.TextContent{Text: string(data)}},
		}, nil
	}
}

// unwrapResult strips the RPC {"_product":...,"data":...} envelope and returns
// the engine data alone. The envelope is a private JSON-RPC design; standard
// MCP tools expose the raw result.
func unwrapResult(raw json.RawMessage) (json.RawMessage, error) {
	var env struct {
		Data json.RawMessage `json:"data"`
	}
	if err := json.Unmarshal(raw, &env); err != nil {
		return nil, fmt.Errorf("decode engine result: %w", err)
	}
	if len(env.Data) == 0 || string(env.Data) == "null" {
		return nil, fmt.Errorf("engine result missing data")
	}
	return env.Data, nil
}

func runStdio(reg *agent.RPCRegistry, version string, logger *slog.Logger) error {
	s := newMCPServer(reg, version, logger)
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()
	return s.Run(ctx, &mcp.StdioTransport{})
}

func envOr(key, def string) string {
	if v := os.Getenv(key); v != "" {
		return v
	}
	return def
}
