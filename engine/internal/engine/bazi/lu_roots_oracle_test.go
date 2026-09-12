package bazi

import (
	"encoding/json"
	"os"
	"reflect"
	"testing"
)

func TestDomainOracle_LuRoots(t *testing.T) {
	raw, err := os.ReadFile("../../../../tests/fixtures/domain_oracle/bazi_lu_roots.json")
	if err != nil {
		t.Fatalf("read oracle: %v", err)
	}
	var doc struct {
		Cases []struct {
			ID       string   `json:"id"`
			Pillars  []string `json:"pillars"`
			Expected []LuRoot `json:"expected"`
			Why      string   `json:"why"`
		} `json:"cases"`
	}
	if err := json.Unmarshal(raw, &doc); err != nil {
		t.Fatalf("decode oracle: %v", err)
	}
	if len(doc.Cases) != 2 {
		t.Fatalf("cases = %d, want 2", len(doc.Cases))
	}
	for _, tc := range doc.Cases {
		t.Run(tc.ID, func(t *testing.T) {
			chart := atomicChartFromStrings(t, tc.Pillars)
			full := ComputeFullChart(chart)
			got := make([]LuRoot, 0, len(full.LuRoots))
			for _, root := range full.LuRoots {
				got = append(got, LuRoot{ShiShen: root.ShiShen, Gan: root.Gan, Zhi: root.Zhi})
			}
			want := make([]LuRoot, 0, len(tc.Expected))
			for _, root := range tc.Expected {
				want = append(want, LuRoot{ShiShen: root.ShiShen, Gan: root.Gan, Zhi: root.Zhi})
			}
			if !reflect.DeepEqual(got, want) {
				t.Fatalf("lu_roots = %#v, want %#v; why: %s", got, want, tc.Why)
			}
		})
	}
}
