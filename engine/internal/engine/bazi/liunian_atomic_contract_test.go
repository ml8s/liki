package bazi

import (
	"encoding/json"
	"os"
	"strings"
	"testing"
)

// LiuNian atomic facts are canonical RPC facts. Empty groups must remain
// empty JSON arrays/maps: null turns every Python consumer into a runtime
// contract failure instead of a clean zero-hit factor.
func TestLiuNianAtomicFactsCollectionsAreNeverNull(t *testing.T) {
	raw, err := os.ReadFile("../../../../tests/fixtures/domain_oracle/bazi_liunian_atomic.json")
	if err != nil {
		t.Fatalf("read oracle: %v", err)
	}
	var doc struct {
		Cases []struct {
			ID      string   `json:"id"`
			Year    int      `json:"year"`
			Pillars []string `json:"pillars"`
		} `json:"cases"`
	}
	if err := json.Unmarshal(raw, &doc); err != nil {
		t.Fatalf("decode oracle: %v", err)
	}
	if len(doc.Cases) == 0 {
		t.Fatal("oracle has no cases")
	}
	for _, tc := range doc.Cases {
		t.Run(tc.ID, func(t *testing.T) {
			chart := atomicChartFromStrings(t, tc.Pillars)
			got, err := ComputeLiuNian(chart, tc.Year)
			if err != nil {
				t.Fatal(err)
			}
			payload, err := json.Marshal(got.AtomicFacts)
			if err != nil {
				t.Fatalf("marshal atomic facts: %v", err)
			}
			for _, field := range []string{
				"controls_elements",
				"controls_targets",
				"combinations",
				"year_branch_relations",
				"year_branch_controlled_by",
				"day_void_branches",
			} {
				null := `"` + field + `":null`
				if strings.Contains(string(payload), null) {
					t.Fatalf("%s = null, want an empty JSON collection; payload=%s", field, payload)
				}
			}
		})
	}
}
