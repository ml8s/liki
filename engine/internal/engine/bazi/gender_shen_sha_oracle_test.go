package bazi

import (
	"encoding/json"
	"os"
	"reflect"
	"testing"

	"liki-engine/internal/engine/ganzhi"
)

func TestDomainOracle_GenderShenSha(t *testing.T) {
	raw, err := os.ReadFile("../../../../tests/fixtures/domain_oracle/bazi_gender_shen_sha.json")
	if err != nil {
		t.Fatalf("read oracle: %v", err)
	}
	var doc struct {
		Cases []struct {
			ID       string              `json:"id"`
			Pillars  []string            `json:"pillars"`
			Gender   string              `json:"gender"`
			Expected map[string][]string `json:"expected"`
			Why      string              `json:"why"`
		} `json:"cases"`
	}
	if err := json.Unmarshal(raw, &doc); err != nil {
		t.Fatalf("decode oracle: %v", err)
	}
	if len(doc.Cases) != 4 {
		t.Fatalf("cases = %d, want 4", len(doc.Cases))
	}

	pillarKey := []string{"nian", "yue", "ri", "shi"}
	for _, tc := range doc.Cases {
		t.Run(tc.ID, func(t *testing.T) {
			chart := atomicChartFromStrings(t, tc.Pillars)
			switch tc.Gender {
			case string(ganzhi.Male):
				chart.Gender = ganzhi.Male
			case string(ganzhi.Female):
				chart.Gender = ganzhi.Female
			default:
				t.Fatalf("gender = %q", tc.Gender)
			}
			full := ComputeFullChart(chart)
			got := map[string][]string{}
			pillars := [4][]shenShaEntry{full.Nian.ShenSha, full.Yue.ShenSha, full.Ri.ShenSha, full.Shi.ShenSha}
			for name := range tc.Expected {
				got[name] = []string{}
				for pi, entries := range pillars {
					for _, item := range entries {
						if item.Name == name {
							got[name] = append(got[name], pillarKey[pi])
						}
					}
				}
			}
			if !reflect.DeepEqual(got, tc.Expected) {
				t.Fatalf("gender shensha = %v, want %v; why: %s", got, tc.Expected, tc.Why)
			}
		})
	}
}
