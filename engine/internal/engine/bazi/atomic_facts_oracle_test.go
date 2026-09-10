package bazi

import (
	"encoding/json"
	"os"
	"reflect"
	"testing"

	"liki-engine/internal/engine/ganzhi"
)

func TestDomainOracle_AtomicFacts(t *testing.T) {
	raw, err := os.ReadFile("../../../../tests/fixtures/domain_oracle/bazi_atomic_facts.json")
	if err != nil {
		t.Fatalf("read oracle: %v", err)
	}
	var doc struct {
		Cases []struct {
			ID       string         `json:"id"`
			Pillars  []string       `json:"pillars"`
			Expected map[string]any `json:"expected"`
		} `json:"cases"`
	}
	if err := json.Unmarshal(raw, &doc); err != nil {
		t.Fatalf("decode oracle: %v", err)
	}
	if len(doc.Cases) != 4 {
		t.Fatalf("cases = %d, want 4", len(doc.Cases))
	}
	for _, tc := range doc.Cases {
		t.Run(tc.ID, func(t *testing.T) {
			var chart Chart
			values := [4]*zhuInfo{&chart.Nian, &chart.Yue, &chart.Ri, &chart.Shi}
			for index, pillar := range tc.Pillars {
				runes := []rune(pillar)
				gan, err := ganzhi.ParseGan(string(runes[0]))
				if err != nil {
					t.Fatal(err)
				}
				zhi, err := ganzhi.ParseZhi(string(runes[1]))
				if err != nil {
					t.Fatal(err)
				}
				values[index].Zhu = ganzhi.Zhu{Gan: gan, Zhi: zhi}
				values[index].NaYin = ganzhi.NayinLabel(gan, zhi)
			}
			chart.Gender = ganzhi.Male
			full := ComputeFullChart(chart)
			got := map[string]any{
				"officer_killing_cleaned": full.AtomicFacts.OfficerKillingCleaned,
				"wealth_tomb_present":     full.AtomicFacts.WealthTombPresent,
				"wealth_star_in_tomb":     full.AtomicFacts.WealthStarInTomb,
				"spouse_palace_state":     full.AtomicFacts.SpousePalaceState,
				"day_branch_type":         full.AtomicFacts.DayBranchType,
				"year_officer_killing":    full.AtomicFacts.YearOfficerKilling,
				"day_master_stem":         full.AtomicFacts.DayMasterStem,
				"day_branch":              full.AtomicFacts.DayBranch,
				"month_longevity":         full.AtomicFacts.MonthLongevity,
				"year_stem_ten_god":       full.AtomicFacts.YearStemTenGod,
				"month_main_ten_god":      full.AtomicFacts.MonthMainTenGod,
				"hour_stem_ten_god":       full.AtomicFacts.HourStemTenGod,
				"pattern_god_transparent": full.AtomicFacts.PatternGodTransparent,
				"pillar_punishments":      full.AtomicFacts.PillarPunishments,
			}
			for field, want := range tc.Expected {
				gotValue, err := json.Marshal(got[field])
				if err != nil {
					t.Fatal(err)
				}
				var normalized any
				if err := json.Unmarshal(gotValue, &normalized); err != nil {
					t.Fatal(err)
				}
				if !reflect.DeepEqual(normalized, want) {
					t.Errorf("%s = %#v, want %#v", field, got[field], want)
				}
			}
		})
	}
}
