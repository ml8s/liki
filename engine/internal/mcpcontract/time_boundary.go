package mcpcontract

import (
	"errors"
	"fmt"
	"strings"
	"time"
)

// parseInstant accepts an RFC3339 absolute instant. Canonical wire form is
// UTC and ends with Z, but numeric offsets are accepted as equivalent
// representations of the same point and normalized here. The returned
// time.Time is always UTC.
func parseInstant(value string) (time.Time, error) {
	if strings.HasSuffix(value, "-00:00") {
		return time.Time{}, fmt.Errorf("instant %q uses unknown offset -00:00", value)
	}

	at, err := time.Parse(time.RFC3339Nano, value)
	if err != nil {
		return time.Time{}, fmt.Errorf("invalid instant %q: %w", value, err)
	}

	return at.UTC(), nil
}

// loadTimezone loads an explicit IANA timezone. Fixed offsets, Local, and bare
// abbreviations are not accepted; timezone identity and offset are separate.
func loadTimezone(value string) (*time.Location, error) {
	if value == "" {
		return nil, errors.New("timezone is required")
	}

	if strings.EqualFold(value, "Local") {
		return nil, fmt.Errorf("timezone %q is not allowed", value)
	}

	loc, err := time.LoadLocation(value)
	if err != nil {
		return nil, fmt.Errorf("invalid timezone %q: %w", value, err)
	}

	return loc, nil
}
