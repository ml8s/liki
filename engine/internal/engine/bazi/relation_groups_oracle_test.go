package bazi

import (
	"encoding/json"
	"os"
	"reflect"
	"testing"
)

func TestDomainOracle_RelationGroups(t *testing.T) {
	raw, err := os.ReadFile("../../../../tests/fixtures/domain_oracle/bazi_relation_groups.json")
	if err != nil {
		t.Fatalf("read oracle: %v", err)
	}
	var doc struct {
		Cases []struct {
			ID       string   `json:"id"`
			Pillars  []string `json:"pillars"`
			Expected []struct {
				Field string `json:"field"`
				Group string `json:"group"`
			} `json:"expected"`
		} `json:"cases"`
	}
	if err := json.Unmarshal(raw, &doc); err != nil {
		t.Fatalf("decode oracle: %v", err)
	}
	if len(doc.Cases) != 5 {
		t.Fatalf("cases = %d, want 5", len(doc.Cases))
	}
	for _, tc := range doc.Cases {
		t.Run(tc.ID, func(t *testing.T) {
			full := ComputeFullChart(atomicChartFromStrings(t, tc.Pillars))
			got := make([]RelationGroup, 0, len(full.RelationGroups))
			for _, group := range full.RelationGroups {
				got = append(got, RelationGroup{Field: group.Field, Group: group.Group})
			}
			want := make([]RelationGroup, 0, len(tc.Expected))
			for _, group := range tc.Expected {
				want = append(want, RelationGroup{Field: group.Field, Group: group.Group})
			}
			if !reflect.DeepEqual(got, want) {
				t.Fatalf("relation groups = %#v, want %#v", got, want)
			}
		})
	}
}
