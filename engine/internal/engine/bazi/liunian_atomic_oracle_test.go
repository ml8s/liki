package bazi

import (
	"encoding/json"
	"os"
	"testing"

	"liki-engine/internal/engine/ganzhi"
)

func TestDomainOracle_LiuNianAtomicFacts(t *testing.T) {
	raw, err := os.ReadFile("../../../../tests/fixtures/domain_oracle/bazi_liunian_atomic.json")
	if err != nil {
		t.Fatalf("read oracle: %v", err)
	}
	var doc struct {
		Cases []struct {
			ID       string         `json:"id"`
			Year     int            `json:"year"`
			Pillars  []string       `json:"pillars"`
			Expected map[string]any `json:"expected"`
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
			chart := atomicChartFromStrings(t, tc.Pillars)
			got, err := ComputeLiuNian(chart, tc.Year)
			if err != nil {
				t.Fatal(err)
			}
			if values, ok := tc.Expected["controls_elements"].([]any); ok {
				expected := make([]string, 0, len(values))
				for _, value := range values {
					expected = append(expected, value.(string))
				}
				if got.AtomicFacts.ControlsElements == nil {
					got.AtomicFacts.ControlsElements = []string{}
				}
				assertStrings(t, got.AtomicFacts.ControlsElements, expected)
			}
			if groups, ok := tc.Expected["combination_kinds"].(map[string]any); ok {
				for kind, rawNames := range groups {
					names := make([]string, 0)
					for _, value := range rawNames.([]any) {
						names = append(names, value.(string))
					}
					actual := make([]string, 0)
					for _, combination := range got.AtomicFacts.Combinations {
						if combination.Kind == kind && combination.IncludesYear {
							actual = append(actual, combination.Group)
						}
					}
					assertStrings(t, actual, names)
				}
			}
			if targets, ok := tc.Expected["controls_targets"].(map[string]any); ok {
				for target, want := range targets {
					if got.AtomicFacts.ControlsTargets[target] != want.(bool) {
						t.Errorf("controls_targets[%s] = %v, want %v", target, got.AtomicFacts.ControlsTargets[target], want)
					}
				}
			}
			if relations, ok := tc.Expected["year_branch_relations"].(map[string]any); ok {
				for branch, relation := range relations {
					if got.AtomicFacts.YearBranchRelations[branch] != relation.(string) {
						t.Errorf("relation %s = %q, want %q", branch, got.AtomicFacts.YearBranchRelations[branch], relation)
					}
				}
			}
			if values, ok := tc.Expected["year_branch_controlled_by"].([]any); ok {
				expected := make([]string, 0, len(values))
				for _, value := range values {
					expected = append(expected, value.(string))
				}
				assertStrings(t, got.AtomicFacts.YearBranchControlledBy, expected)
			}
			boolFields := map[string]func(LiuNianAtomicFacts) bool{
				"year_gan_controls_day_gan":      func(f LiuNianAtomicFacts) bool { return f.YearGanControlsDayGan },
				"day_gan_controls_year_gan":      func(f LiuNianAtomicFacts) bool { return f.DayGanControlsYearGan },
				"gan_combines":                   func(f LiuNianAtomicFacts) bool { return f.GanCombines },
				"dayun_gan_combines":             func(f LiuNianAtomicFacts) bool { return f.DayunGanCombines },
				"dayun_zhi_clashes_year":         func(f LiuNianAtomicFacts) bool { return f.DayunZhiClashesYear },
				"year_equals_day_pillar":         func(f LiuNianAtomicFacts) bool { return f.YearEqualsDayPillar },
				"dayun_equals_year_pillar":       func(f LiuNianAtomicFacts) bool { return f.DayunEqualsYearPillar },
				"year_gan_equals_natal_year_gan": func(f LiuNianAtomicFacts) bool { return f.YearGanEqualsNatalYearGan },
			}
			for field, getter := range boolFields {
				if want, ok := tc.Expected[field].(bool); ok && getter(got.AtomicFacts) != want {
					t.Errorf("%s = %v, want %v", field, getter(got.AtomicFacts), want)
				}
			}
			if values, ok := tc.Expected["day_void_branches"].([]any); ok {
				expected := make([]string, 0, len(values))
				for _, value := range values {
					expected = append(expected, value.(string))
				}
				assertStrings(t, got.AtomicFacts.DayVoidBranches, expected)
			}
			if want, ok := tc.Expected["wealth_breaks_seal"].(bool); ok {
				if got.AtomicFacts.WealthBreaksSeal != want {
					t.Errorf("wealth_breaks_seal = %v, want %v", got.AtomicFacts.WealthBreaksSeal, want)
				}
			}
		})
	}
}

func atomicChartFromStrings(t *testing.T, pillars []string) Chart {
	t.Helper()
	if len(pillars) != 4 {
		t.Fatalf("pillars = %d, want 4", len(pillars))
	}
	var chart Chart
	values := [4]*zhuInfo{&chart.Nian, &chart.Yue, &chart.Ri, &chart.Shi}
	for index, pillar := range pillars {
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
	return chart
}

func assertStrings(t *testing.T, got, want []string) {
	t.Helper()
	if len(got) != len(want) {
		t.Fatalf("strings = %v, want %v", got, want)
	}
	for index := range want {
		if got[index] != want[index] {
			t.Fatalf("strings = %v, want %v", got, want)
		}
	}
}
