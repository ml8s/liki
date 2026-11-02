package bazi

import (
	"encoding/json"
	"os"
	"reflect"
	"testing"
)

func TestDomainOracle_GuanXi(t *testing.T) {
	type relationKey struct {
		Kind  string
		Group string
	}
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
	if len(doc.Cases) != 6 {
		t.Fatalf("cases = %d, want 6", len(doc.Cases))
	}
	for _, tc := range doc.Cases {
		t.Run(tc.ID, func(t *testing.T) {
			full := ComputeChartFromCore(pillarsCoreFromStrings(t, tc.Pillars))
			gotSeen := map[relationKey]bool{}
			got := make([]relationKey, 0, len(full.GuanXi))
			for _, group := range full.GuanXi {
				key := relationKey{Kind: group.Kind, Group: group.Group}
				if !gotSeen[key] {
					gotSeen[key] = true
					got = append(got, key)
				}
			}
			want := make([]relationKey, 0, len(tc.Expected))
			for _, group := range tc.Expected {
				want = append(want, relationKey{Kind: group.Field, Group: group.Group})
			}
			if !reflect.DeepEqual(got, want) {
				t.Fatalf("relation groups = %#v, want %#v", got, want)
			}
		})
	}
}
