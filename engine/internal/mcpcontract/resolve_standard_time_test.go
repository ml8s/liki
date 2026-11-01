package mcpcontract

import (
	"context"
	"testing"
)

func TestTianwenResolveStandardRemovesDSTFold(t *testing.T) {
	fold, err := TianwenResolveStandard(context.Background(), TianwenResolveInput{
		Calendar: "solar", Date: "2024-11-03", Time: "01:30",
		Timezone: "America/New_York", Longitude: -74, Gender: "male",
	})
	if err != nil {
		t.Fatal(err)
	}
	if len(fold.Candidates) != 1 {
		t.Fatalf("standard candidates = %d, want 1", len(fold.Candidates))
	}
	if got := string(fold.Candidates[0].Subject.BirthTime); got != "2024-11-03T01:30:00-05:00" {
		t.Fatalf("standard birth_time = %s, want -05:00 standard time", got)
	}

	gap, err := TianwenResolveStandard(context.Background(), TianwenResolveInput{
		Calendar: "solar", Date: "2024-03-10", Time: "02:30",
		Timezone: "America/New_York", Longitude: -74, Gender: "male",
	})
	if err != nil {
		t.Fatalf("DST gap must normalize to standard time: %v", err)
	}
	if len(gap.Candidates) != 1 {
		t.Fatalf("gap candidates = %d, want 1", len(gap.Candidates))
	}
	if got := string(gap.Candidates[0].Subject.BirthTime); got != "2024-03-10T02:30:00-05:00" {
		t.Fatalf("gap birth_time = %s, want -05:00 standard time", got)
	}
}
