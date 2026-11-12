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
	serverName  = "hk.liki/engine"
	serverTitle = "Liki Engine"
	serverDesc  = "Liki Engine — deterministic metaphysics computation and auditable divination events: bazi, ziwei, liuyao, qimen, huangli, fengshui."
)

// engineContracts owns every engine tool contract; the legacy raw-JSON
// registry has been fully retired.
var engineContracts = mcpcontract.NewEngine()

// newMCPServer builds a standard MCP server exposing every engine tool.
func newMCPServer(version string, logger *slog.Logger) (*mcp.Server, error) {
	impl := &mcp.Implementation{
		Name:        serverName,
		Title:       serverTitle,
		Description: serverDesc,
		Version:     version,
		WebsiteURL:  "https://liki.hk",
	}
	opts := &mcp.ServerOptions{
		Instructions: "Deterministic Chinese metaphysics computation and auditable divination events. Discover available capabilities with tools/list before use. When required, first resolve time, location, or calendar facts; then use the discovered chart, calendar, divination, or fengshui capability. Auto Liuyao casting creates a new event on each call; coins/yaos receipts reproduce one event. Pass returned artifacts unchanged, interpret only returned facts, and do not invent missing fields. Conclusions are conditioned readings from a traditional culture perspective, not medical, legal or investment advice.",
		Logger:       logger,
		// 2026-07-28 is the preferred stateless surface. Standard clients still
		// discover this endpoint through the initialize handshake, whose newest
		// era is 2025-11-25; advertising only 2026 made the SDK return that
		// handshake fallback and then reject its required header on the next call.
		SupportedProtocolVersions: []string{"2025-11-25", "2026-07-28"},
		// The deterministic tool set is immutable for the life of an image.
		// Advertising listChanged makes modern clients open a long-lived
		// subscriptions/listen stream, which this stateless surface does not
		// need and cannot keep alive behind its request timeout.
		Capabilities: &mcp.ServerCapabilities{
			Tools: &mcp.ToolCapabilities{},
		},
	}
	s := mcp.NewServer(impl, opts)

	for _, name := range engineContracts.Names() {
		definition, err := engineContracts.Definition(name)
		if err != nil {
			return nil, fmt.Errorf("engine MCP tool %s: %w", name, err)
		}
		tool := &mcp.Tool{
			Name:         definition.Name,
			Description:  definition.Description,
			InputSchema:  definition.InputSchema,
			OutputSchema: definition.OutputSchema,
			Annotations:  toolAnnotations(name),
		}
		s.AddTool(tool, contractToolHandler(engineContracts, name))
	}
	return s, nil
}

func toolAnnotations(name string) *mcp.ToolAnnotations {
	// liuyao_qigua's auto mode creates a new random divination event on each
	// call. Method-level hints cannot vary by arguments, so the whole tool is
	// conservatively non-idempotent even though coins/yaos receipts reproduce.
	return &mcp.ToolAnnotations{
		ReadOnlyHint:   true,
		IdempotentHint: name != "liuyao_qigua",
	}
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
