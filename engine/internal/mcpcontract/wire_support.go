package mcpcontract

import (
	"encoding/json"
	"fmt"
)

// wireFromDomain converts a domain value into its typed wire DTO through JSON.
// The wire DTOs are faithful projections of the domain JSON shapes; output
// schema checks are contract-time checks exposed by Registry.ValidateOutput.
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

// marshalRawOutput marshals a domain value that is contract-tested to match
// its published wire shape, avoiding a domain→DTO→JSON round trip.
func marshalRawOutput(tool string, value any) (json.RawMessage, error) {
	raw, err := json.Marshal(value)
	if err != nil {
		return nil, fmt.Errorf("%s: marshal output: %w", tool, err)
	}
	return raw, nil
}
