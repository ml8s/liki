package mcpcontract

import (
	"encoding/json"
	"fmt"
	"time"

	"liki-engine/internal/engine/tianwen"
)

// wireFromDomain converts a domain value into its typed wire DTO through JSON:
// the wire DTOs are faithful projections of the domain JSON shapes, and the
// registry validates the result against the DTO schema before it is returned
// as MCP structuredContent.
func wireFromDomain[Out any](value any) (Out, error) {
	raw, err := json.Marshal(value)
	if err != nil {
		var zero Out
		return zero, err
	}
	var out Out
	if err := json.Unmarshal(raw, &out); err != nil {
		var zero Out
		return zero, err
	}
	return out, nil
}

// parseWireSolarTime parses an RFC3339 wire timestamp into engine solar time.
func parseWireSolarTime(value RFC3339Time) (tianwen.SolarTime, error) {
	parsed, err := time.Parse(time.RFC3339, string(value))
	if err != nil {
		return tianwen.SolarTime{}, fmt.Errorf("invalid solar_time %q: %w", value, err)
	}
	return tianwen.SolarTime(parsed), nil
}
