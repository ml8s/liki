package mcpcontract

import (
	"context"
	"testing"
)

func TestTianwenResolveDSTFoldAndGap(t *testing.T) {
	fold, err := TianwenResolve(context.Background(), TianwenResolveInput{
		Calendar: "solar", Date: "2024-11-03", Time: "01:30",
		Timezone: "America/New_York", Longitude: -74,
	})
	if err != nil {
		t.Fatal(err)
	}
	if len(fold.Candidates) != 2 {
		t.Fatalf("fold candidates = %d, want 2", len(fold.Candidates))
	}
	seen := map[string]bool{}
	for _, candidate := range fold.Candidates {
		seen[candidate.StandardTime] = true
	}
	for _, want := range []string{"2024-11-03T00:30:00-05:00", "2024-11-03T01:30:00-05:00"} {
		if !seen[want] {
			t.Fatalf("missing fold candidate %s: %+v", want, fold.Candidates)
		}
	}

	if _, err := TianwenResolve(context.Background(), TianwenResolveInput{
		Calendar: "solar", Date: "2024-03-10", Time: "02:30",
		Timezone: "America/New_York", Longitude: -74,
	}); err == nil {
		t.Fatal("DST gap should fail")
	}
}
