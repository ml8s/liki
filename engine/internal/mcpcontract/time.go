package mcpcontract

import (
	"context"
	"fmt"
	"time"
)

// TimeNowInput requires an IANA timezone; the tool layer never guesses.
type TimeNowInput struct {
	Timezone string `json:"timezone" jsonschema:"IANA timezone, e.g. Asia/Shanghai. Required."`
}

// TimeNowOutput is the canonical clock fact: UTC instant + IANA zone.
// Display local time/offset is the frontend's concern (Intl.DateTimeFormat).
type TimeNowOutput struct {
	Instant  string `json:"instant" jsonschema:"Current instant in UTC (RFC3339, Z)."`
	Timezone string `json:"timezone" jsonschema:"Effective IANA timezone."`
}

// TimeNow returns the current instant plus its representation in the required
// IANA timezone. The wall clock is never used as an implicit source.
func TimeNow(ctx context.Context, input TimeNowInput) (TimeNowOutput, error) {
	if err := ctx.Err(); err != nil {
		return TimeNowOutput{}, err
	}
	if input.Timezone == "" {
		return TimeNowOutput{}, fmt.Errorf("time.now: timezone is required")
	}
	zone, err := loadTimezone(input.Timezone)
	if err != nil {
		return TimeNowOutput{}, fmt.Errorf("time.now: %w", err)
	}
	now := time.Now().Round(0).UTC()
	return TimeNowOutput{
		Instant:  now.Format(time.RFC3339Nano),
		Timezone: zone.String(),
	}, nil
}
