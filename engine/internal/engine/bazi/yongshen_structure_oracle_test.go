package bazi

import (
	"encoding/json"
	"os"
	"reflect"
	"testing"
)

func TestDomainOracle_BaziYongShenStructure(t *testing.T) {
	raw, err := os.ReadFile("../../../../tests/fixtures/domain_oracle/bazi_yongshen_structure.json")
	if err != nil {
		t.Fatalf("read oracle: %v", err)
	}
	var doc struct {
		GeJuCases []struct {
			ID                            string             `json:"id"`
			Pillars                       []string           `json:"pillars"`
			ExpectedPattern               string             `json:"expected_pattern"`
			ExpectedPatternGod            string             `json:"expected_pattern_god"`
			ExpectedTransparentPillars    []string           `json:"expected_transparent_pillars"`
			ExpectedControllerTransparent []string           `json:"expected_controller_transparent_pillars"`
			ExpectedPatternRoots          []StemOccurrence   `json:"expected_pattern_roots"`
			ExpectedPatternTimely         bool               `json:"expected_pattern_timely"`
			ExpectedRelationFacts         []expectedRelation `json:"expected_relation_facts"`
		} `json:"ge_ju_cases"`
		FuYiCases []struct {
			ID                    string             `json:"id"`
			Pillars               []string           `json:"pillars"`
			ExpectedRootType      string             `json:"expected_root_type"`
			ExpectedYinBiCount    int                `json:"expected_yin_bi_count"`
			ExpectedDayRoots      []StemOccurrence   `json:"expected_day_master_roots"`
			ExpectedRelationFacts []expectedRelation `json:"expected_relation_facts"`
		} `json:"fu_yi_cases"`
		TiaoHouCases []struct {
			ID                         string              `json:"id"`
			Pillars                    []string            `json:"pillars"`
			ExpectedPrimary            expectedTiaoHouGod  `json:"expected_primary"`
			ExpectedSecondary          *expectedTiaoHouGod `json:"expected_secondary"`
			ExpectedPrimaryRelationIDs []string            `json:"expected_primary_relation_groups"`
		} `json:"tiao_hou_cases"`
	}
	if err := json.Unmarshal(raw, &doc); err != nil {
		t.Fatalf("decode oracle: %v", err)
	}
	if len(doc.GeJuCases) != 4 || len(doc.FuYiCases) != 2 || len(doc.TiaoHouCases) != 1 {
		t.Fatalf("cases = %d/%d/%d, want 4/2/1", len(doc.GeJuCases), len(doc.FuYiCases), len(doc.TiaoHouCases))
	}
	for _, tc := range doc.GeJuCases {
		t.Run(tc.ID, func(t *testing.T) {
			got := ComputeYongShen(atomicChartFromStrings(t, tc.Pillars)).GeJu
			if got.Pattern != tc.ExpectedPattern || got.PatternGod != tc.ExpectedPatternGod {
				t.Fatalf("pattern = %s/%s, want %s/%s", got.Pattern, got.PatternGod, tc.ExpectedPattern, tc.ExpectedPatternGod)
			}
			assertStrings(t, got.Structure.Pattern.TransparentPillars, tc.ExpectedTransparentPillars)
			if got.Structure.Pattern.Timely != tc.ExpectedPatternTimely {
				t.Fatalf("pattern timely = %v, want %v", got.Structure.Pattern.Timely, tc.ExpectedPatternTimely)
			}
			if !reflect.DeepEqual(got.Structure.Pattern.Roots, tc.ExpectedPatternRoots) {
				t.Fatalf("pattern roots = %#v, want %#v", got.Structure.Pattern.Roots, tc.ExpectedPatternRoots)
			}
			if tc.ExpectedControllerTransparent != nil {
				assertStrings(t, got.Structure.Controller.TransparentPillars, tc.ExpectedControllerTransparent)
			}
			assertRelationFacts(t, got.Structure.RelationFacts, tc.ExpectedRelationFacts)
		})
	}
	for _, tc := range doc.FuYiCases {
		t.Run(tc.ID, func(t *testing.T) {
			got := ComputeYongShen(atomicChartFromStrings(t, tc.Pillars)).FuYi.Basis
			if got.RootType != tc.ExpectedRootType || got.YinBiCount != tc.ExpectedYinBiCount {
				t.Fatalf("basis = %s/%d, want %s/%d", got.RootType, got.YinBiCount, tc.ExpectedRootType, tc.ExpectedYinBiCount)
			}
			if !reflect.DeepEqual(got.DayMasterRoots, tc.ExpectedDayRoots) {
				t.Fatalf("day roots = %#v, want %#v", got.DayMasterRoots, tc.ExpectedDayRoots)
			}
			assertRelationFacts(t, got.RelationFacts, tc.ExpectedRelationFacts)
		})
	}
	for _, tc := range doc.TiaoHouCases {
		t.Run(tc.ID, func(t *testing.T) {
			got := ComputeYongShen(atomicChartFromStrings(t, tc.Pillars)).TiaoHou
			assertTiaoHouGod(t, got.Primary, tc.ExpectedPrimary)
			if tc.ExpectedSecondary == nil {
				if got.Secondary != nil {
					t.Fatalf("secondary = %#v, want nil", got.Secondary)
				}
				return
			}
			if got.Secondary == nil {
				t.Fatalf("secondary = nil, want %#v", tc.ExpectedSecondary)
			}
			assertTiaoHouGod(t, *got.Secondary, *tc.ExpectedSecondary)
			for _, group := range tc.ExpectedPrimaryRelationIDs {
				found := false
				for _, fact := range got.Primary.RelationFacts {
					if fact.Group == group {
						found = true
						break
					}
				}
				if !found {
					t.Fatalf("primary relation %q absent: %#v", group, got.Primary.RelationFacts)
				}
			}
		})
	}
}

type expectedRelation struct {
	Field   string   `json:"field"`
	Group   string   `json:"group"`
	Targets []string `json:"targets"`
}

type expectedTiaoHouGod struct {
	Stem        string           `json:"stem"`
	Transparent bool             `json:"transparent"`
	Hidden      bool             `json:"hidden"`
	Occurrences []StemOccurrence `json:"occurrences"`
}

func assertRelationFacts(t *testing.T, facts []RelationFact, expected []expectedRelation) {
	t.Helper()
	for _, want := range expected {
		found := false
		for _, fact := range facts {
			if fact.Field != want.Field || fact.Group != want.Group {
				continue
			}
			found = true
			if len(fact.Targets) != len(want.Targets) {
				t.Fatalf("relation %s targets = %v, want %v", want.Group, fact.Targets, want.Targets)
			}
			for index := range want.Targets {
				if fact.Targets[index] != want.Targets[index] {
					t.Fatalf("relation %s targets = %v, want %v", want.Group, fact.Targets, want.Targets)
				}
			}
			break
		}
		if !found {
			t.Fatalf("relation %s/%s absent: %#v", want.Field, want.Group, facts)
		}
	}
}

func assertTiaoHouGod(t *testing.T, got TiaoHouAvailability, want expectedTiaoHouGod) {
	t.Helper()
	if got.Stem != want.Stem || got.Transparent != want.Transparent || got.Hidden != want.Hidden {
		t.Fatalf("availability = %+v, want %+v", got, want)
	}
	if !reflect.DeepEqual(got.Occurrences, want.Occurrences) {
		t.Fatalf("occurrences = %#v, want %#v", got.Occurrences, want.Occurrences)
	}
}
