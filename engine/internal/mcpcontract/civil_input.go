package mcpcontract

import (
	"fmt"
	"time"

	"liki-engine/internal/engine/tianwen"
)

func standardTimeString(at time.Time, loc *time.Location) string {
	return tianwen.StandardLocalTime(at, loc).Format(time.RFC3339Nano)
}

func parseStandardTime(value StandardTime) (time.Time, *time.Location, error) {
	at, err := time.Parse(time.RFC3339Nano, string(value))
	if err != nil {
		return time.Time{}, nil, fmt.Errorf("standard_time must be RFC3339 with offset, got %q: %w", value, err)
	}
	return at, at.Location(), nil
}

func standardTimeValue(at time.Time) StandardTime {
	return StandardTime(at.Format(time.RFC3339Nano))
}
