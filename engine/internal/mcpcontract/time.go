package mcpcontract

import (
	"context"
	"fmt"
	"time"
)

// TimeNowInput is deliberately empty: the server owns the clock. The generated
// schema still closes the object so unknown arguments fail validation.
type TimeNowInput struct{}

// TimeNowOutput is the MCP wire shape for the server clock.
type TimeNowOutput struct {
	UTC   string `json:"utc" jsonschema:"Current UTC timestamp in RFC3339 format."`
	Local string `json:"local" jsonschema:"Server-local timestamp in RFC3339 format."`
	CST   string `json:"cst" jsonschema:"Current China Standard Time (UTC+8) timestamp in RFC3339 format."`
}

// TimeNow returns the server clock in UTC, server-local time, and CST.
func TimeNow(ctx context.Context, _ TimeNowInput) (TimeNowOutput, error) {
	if err := ctx.Err(); err != nil {
		return TimeNowOutput{}, err
	}
	now := time.Now()
	return TimeNowOutput{
		UTC:   now.UTC().Format(time.RFC3339),
		Local: now.Format(time.RFC3339),
		CST:   now.In(time.FixedZone("CST", 8*3600)).Format(time.RFC3339),
	}, nil
}

// NewEngine returns the engine's typed MCP contract registry.
func NewEngine() *Registry {
	registry := New()
	err := register(registry, Tool[TimeNowInput, TimeNowOutput]{
		Name:        "time.now",
		Description: "获取服务端当前时间。返回 UTC、本地、北京时间，用于 AI agent 获取准确时间避免幻觉。",
		Handler:     TimeNow,
	})
	if err != nil {
		panic(fmt.Sprintf("register engine MCP contracts: %v", err))
	}
	return registry
}
