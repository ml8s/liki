package mcpcontract

import (
	"context"
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
