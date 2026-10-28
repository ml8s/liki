package tianwen

import (
	"testing"
	"time"
)

func TestBoundariesInWindowUsesLunarBasis(t *testing.T) {
	ny, err := time.LoadLocation("America/New_York")
	if err != nil {
		t.Fatal(err)
	}
	w := Window{
		Start: time.Date(2024, 2, 8, 0, 0, 0, 0, ny),
		End:   time.Date(2024, 2, 10, 12, 0, 0, 0, ny),
	}

	local := BoundariesInWindow(w, []BoundaryKind{BoundaryDay, BoundaryLunarYear}, BoundaryOptions{
		Timezone: ny, LunarBasis: LunarBasisLocalAstronomical,
	})
	official := BoundariesInWindow(w, []BoundaryKind{BoundaryDay, BoundaryLunarYear}, BoundaryOptions{
		Timezone: ny, LunarBasis: LunarBasisOfficial,
	})

	find := func(items []Boundary, kind BoundaryKind, at time.Time) bool {
		for _, item := range items {
			if item.Kind == kind && item.At.Equal(at) {
				return true
			}
		}
		return false
	}

	if !find(local, BoundaryLunarYear, time.Date(2024, 2, 9, 0, 0, 0, 0, ny)) {
		t.Fatalf("local lunar year boundary missing: %+v", local)
	}
	if !find(local, BoundaryDay, time.Date(2024, 2, 10, 0, 0, 0, 0, ny)) {
		t.Fatalf("local day boundary missing: %+v", local)
	}

	if !find(official, BoundaryLunarYear, time.Date(2024, 2, 9, 16, 0, 0, 0, time.UTC)) {
		t.Fatalf("official lunar year boundary missing: %+v", official)
	}
	if !find(official, BoundaryDay, time.Date(2024, 2, 9, 16, 0, 0, 0, time.UTC)) {
		t.Fatalf("official day boundary missing: %+v", official)
	}
}
