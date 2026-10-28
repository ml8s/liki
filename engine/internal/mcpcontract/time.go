package mcpcontract

import (
	"context"
	"fmt"
	"time"

	"liki-engine/internal/engine/tianwen"
)

// TimeNowInput requires an IANA timezone; the tool layer never guesses.
type TimeNowInput struct {
	Timezone string `json:"timezone" jsonschema:"IANA timezone, e.g. Asia/Shanghai. Required."`
}

// TimeNowOutput is the canonical clock fact in standard and civil time.
type TimeNowOutput struct {
	StandardTime string `json:"standard_time" jsonschema:"当前时区标准时间（无 DST），RFC3339。"`
}

// TimeNow returns the current standard time in the required IANA timezone. The
// process wall clock is never used as an implicit source.
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
		StandardTime: tianwen.StandardLocalTime(now, zone).Format(time.RFC3339Nano),
	}, nil
}
