package mcpcontract

import (
	"fmt"
	"strings"
	"time"

	"liki-engine/internal/engine/tianwen"
)

func standardTimeString(at time.Time, loc *time.Location) string {
	return tianwen.StandardLocalTime(at, loc).Format(time.RFC3339Nano)
}

func parseStandardTime(value StandardTime) (time.Time, *time.Location, error) {
	if strings.HasSuffix(string(value), "-00:00") {
		return time.Time{}, nil, fmt.Errorf(
			"standard_time must use a known RFC3339 offset; -00:00 means unknown offset",
		)
	}
	at, err := time.Parse(time.RFC3339Nano, string(value))
	if err != nil {
		return time.Time{}, nil, fmt.Errorf(
			"standard_time must be RFC3339 with a known offset, got %q: %w", value, err)
	}
	return at, at.Location(), nil
}

func standardTimeValue(at time.Time) StandardTime {
	return StandardTime(at.Format(time.RFC3339Nano))
}
