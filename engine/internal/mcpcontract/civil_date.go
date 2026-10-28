package mcpcontract

import (
	"time"

	"liki-engine/internal/engine/tianwen"
)

func civilDateString(date tianwen.CivilDate) DateString {
	return DateString(date.String())
}

func instantString(at time.Time) Instant {
	return Instant(at.UTC().Format(time.RFC3339Nano))
}
