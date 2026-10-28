package tianwen

import (
	"testing"
	"time"
)

func TestEquationOfTimeUsesAbsoluteInstant(t *testing.T) {
	utc := time.Date(2024, 5, 20, 4, 30, 12, 500, time.UTC)
	representations := []time.Time{
		utc,
		utc.In(time.FixedZone("standard", 8*3600)),
		utc.In(time.FixedZone("standard", -5*3600)),
	}

	want := equationOfTime(representations[0])
	for _, at := range representations[1:] {
		if got := equationOfTime(at); got != want {
			t.Fatalf("equationOfTime(%s) = %v, want %v", at, got, want)
		}
	}
}
