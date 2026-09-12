package bazi

import (
	"encoding/json"
	"os"
	"reflect"
	"testing"

	"liki-engine/internal/engine/ganzhi"
)

func TestDomainOracle_NayinShenSha(t *testing.T) {
	raw, err := os.ReadFile("../../../../tests/fixtures/domain_oracle/bazi_nayin_shen_sha.json")
	if err != nil {
		t.Fatalf("read oracle: %v", err)
	}
	var doc struct {
		Cases []struct {
			ID       string              `json:"id"`
			Pillars  []string            `json:"pillars"`
			Expected map[string][]string `json:"expected"`
		} `json:"cases"`
	}
	if err := json.Unmarshal(raw, &doc); err != nil {
		t.Fatalf("decode oracle: %v", err)
	}
	if len(doc.Cases) != 6 {
		t.Fatalf("cases = %d, want 6", len(doc.Cases))
	}
	keys := []string{"nian", "yue", "ri", "shi"}
	for _, tc := range doc.Cases {
		t.Run(tc.ID, func(t *testing.T) {
			chart := atomicChartFromStrings(t, tc.Pillars)
			chart.Gender = ganzhi.Male
			full := ComputeFullChart(chart)
			pillars := [4][]shenShaEntry{full.Nian.ShenSha, full.Yue.ShenSha, full.Ri.ShenSha, full.Shi.ShenSha}
			got := map[string][]string{"学堂": {}, "词馆": {}}
			for name := range got {
				for pi, entries := range pillars {
					for _, item := range entries {
						if item.Name == name {
							got[name] = append(got[name], keys[pi])
						}
					}
				}
			}
			if !reflect.DeepEqual(got, tc.Expected) {
				t.Fatalf("nayin shensha = %v, want %v", got, tc.Expected)
			}
		})
	}
}
