package main

import (
	"context"
	"encoding/json"
	"fmt"
	"log/slog"
	"os"
	"strings"
	"time"

	"github.com/modelcontextprotocol/go-sdk/mcp"

	"liki-engine/internal/mcpcontract"
	"liki-engine/internal/observability"
)

const (
	serverName = "engine"
	serverDesc = "Liki Chart Engine — deterministic Chinese metaphysics computation: bazi, ziwei, liuyao, qimen, huangli, fengshui."
)

// toolName converts an RPC method name to a valid MCP tool name.
// RPC method names use dot-separated domains (bazi.chart) while MCP tool names
// only allow [a-zA-Z0-9_-], so each dot is replaced with an underscore.
// mcpDomains maps an MCP endpoint suffix to the RPC method prefixes it exposes.
// Each domain is a self-contained tool set (one 术数 or shared aux) so the LLM
// only sees relevant tools per expert.
var mcpDomains = []struct {
	Suffix   string
	Prefixes []string
}{
	{"bazi", []string{"bazi_"}},
	{"ziwei", []string{"ziwei_"}},
	{"qimen", []string{"qimen_"}},
	{"liuyao", []string{"liuyao_"}},
	{"huangli", []string{"huangli_"}},
	{"fengshui", []string{"bazhai_", "xuankong_", "time_"}},
	{"aux", []string{"time_", "tianwen_", "city_"}},
}

// engineContracts owns every engine tool contract; the legacy raw-JSON
// registry has been fully retired.
var engineContracts = mcpcontract.NewEngine()

// newMCPServer builds a standard MCP server exposing every engine tool.
func newMCPServer(version string, logger *slog.Logger) (*mcp.Server, error) {
	return newMCPServerFor(version, logger, func(string) bool { return true })
}

// newDomainServer builds an MCP server exposing only the methods matching the
// given RPC method prefixes (one 术数 domain or the shared aux tools).
func newDomainServer(prefixes []string, version string, logger *slog.Logger) (*mcp.Server, error) {
	return newMCPServerFor(version, logger, func(method string) bool {
		for _, p := range prefixes {
			if strings.HasPrefix(method, p) {
				return true
			}
		}
		return false
	})
}

func newMCPServerFor(
	version string, logger *slog.Logger, match func(string) bool,
) (*mcp.Server, error) {
	impl := &mcp.Implementation{
		Name:        serverName,
		Title:       "Liki Chart Engine",
		Description: serverDesc,
		Version:     version,
		WebsiteURL:  "https://liki.hk",
	}
	opts := &mcp.ServerOptions{
		Instructions: "Deterministic Chinese metaphysics computation. Discover available capabilities with tools/list before use. When required, first resolve time, location, or calendar facts; then use the discovered chart, calendar, divination, or fengshui capability. Pass returned artifacts unchanged, interpret only returned facts, and do not invent missing fields. Conclusions are conditioned readings from a traditional culture perspective, not medical, legal or investment advice.",
		Logger:       logger,
		// 2026-07-28 is the preferred stateless surface. Standard clients still
		// discover this endpoint through the initialize handshake, whose newest
		// era is 2025-11-25; advertising only 2026 made the SDK return that
		// handshake fallback and then reject its required header on the next call.
		SupportedProtocolVersions: []string{"2025-11-25", "2026-07-28"},
	}
	s := mcp.NewServer(impl, opts)

	for _, name := range engineContracts.Names() {
		if !match(name) {
			continue
		}
		definition, err := engineContracts.Definition(name)
		if err != nil {
			return nil, fmt.Errorf("engine MCP tool %s: %w", name, err)
		}
		tool := &mcp.Tool{
			Name:         definition.Name,
			Description:  definition.Description,
			InputSchema:  definition.InputSchema,
			OutputSchema: definition.OutputSchema,
			// 全部 engine 工具都是确定性计算：无副作用、幂等，如实标注。
			Annotations: &mcp.ToolAnnotations{
				ReadOnlyHint:   true,
				IdempotentHint: true,
			},
		}
		s.AddTool(tool, contractToolHandler(engineContracts, name))
	}
	return s, nil
}

// contractToolHandler invokes a typed MCP contract and answers structured-only:
// every tool publishes a typed outputSchema, the text-content duplication was
// removed as a payload optimization, and every consumer (browser, experts SDK,
// external agents) reads structuredContent. Validation and business failures
// stay visible to the model as text tool errors.
func contractToolHandler(contracts *mcpcontract.Registry, name string) mcp.ToolHandler {
	return func(ctx context.Context, req *mcp.CallToolRequest) (*mcp.CallToolResult, error) {
		started := time.Now()
		status := "success"
		args := req.Params.Arguments
		if len(args) == 0 || string(args) == "null" {
			args = json.RawMessage(`{}`)
		}
		result, err := contracts.Call(ctx, name, args)
		if err != nil {
			status = "invalid_arguments"
			if !strings.Contains(err.Error(), "invalid arguments") {
				status = "handler_error"
			}
			duration := time.Since(started).Seconds()
			defer observability.ObserveTool(toolDomain(name), name, status, duration)
			return &mcp.CallToolResult{
				IsError: true,
				Content: []mcp.Content{&mcp.TextContent{Text: err.Error()}},
			}, nil
		}
		duration := time.Since(started).Seconds()
		defer observability.ObserveTool(toolDomain(name), name, status, duration)
		return &mcp.CallToolResult{StructuredContent: result}, nil
	}
}

func toolDomain(name string) string {
	if domain, _, found := strings.Cut(name, "_"); found {
		return domain
	}
	return "other"
}

func runStdio(version string, logger *slog.Logger) error {
	s, err := newMCPServer(version, logger)
	if err != nil {
		return err
	}
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
