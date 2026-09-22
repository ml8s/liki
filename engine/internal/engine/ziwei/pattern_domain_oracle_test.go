package ziwei

import (
	"encoding/json"
	"os"
	"testing"

	"liki-engine/internal/engine/ganzhi"
)

type patternSemanticPalace struct {
	Name  string                `json:"name"`
	Zhi   string                `json:"zhi"`
	Stars []patternSemanticStar `json:"stars"`
	ZaYao []string              `json:"za_yao"`
}

type patternSemanticStar struct {
	Name       string `json:"name"`
	SiHua      string `json:"si_hua"`
	Brightness string `json:"brightness"`
}

func buildPatternSemanticChart(t *testing.T, input []patternSemanticPalace) Chart {
	t.Helper()
	palaces := blankPalaces()
	for _, item := range input {
		index := -1
		for i, name := range gongLabels {
			if name == item.Name {
				index = i
				break
			}
		}
		if index < 0 {
			t.Fatalf("unknown palace %q", item.Name)
		}
		zhi := Zhi(0)
		if item.Zhi != "" {
			parsed, err := ganzhi.ParseZhi(item.Zhi)
			if err != nil {
				t.Fatalf("parse zhi %q: %v", item.Zhi, err)
			}
			zhi = parsed
		}
		stars := make([]starInfo, 0, len(item.Stars))
		for _, semantic := range item.Stars {
			var star starIndex
			if err := star.fromName(semantic.Name); err != nil {
				t.Fatalf("parse star %q: %v", semantic.Name, err)
			}
			stars = append(stars, starInfo{
				Star: star, Name: semantic.Name, IsMajor: int(star) < 14,
				SiHua: semantic.SiHua, Brightness: semantic.Brightness,
			})
		}
		palaces[index].Zhi = zhi
		palaces[index].Stars = stars
		palaces[index].ZaYao = item.ZaYao
	}
	return Chart{GongWei: palaces}
}

func TestDomainOracle_ZiweiPatternSemantics(t *testing.T) {
	raw, err := os.ReadFile("../../../../tests/fixtures/domain_oracle/ziwei_pattern_semantics.json")
	if err != nil {
		t.Fatalf("read oracle: %v", err)
	}
	var doc struct {
		SanFangCases []struct {
			ID            string                  `json:"id"`
			Palaces       []patternSemanticPalace `json:"palaces"`
			ExpectedSiHua map[string][]string     `json:"expected_si_hua"`
		} `json:"san_fang_cases"`
		PalaceFactCases []struct {
			ID                        string                  `json:"id"`
			Palaces                   []patternSemanticPalace `json:"palaces"`
			ExpectedStarTargets       []string                `json:"expected_star_targets"`
			ExpectedBrightnessTargets []string                `json:"expected_brightness_targets"`
			ExpectedAbsentTargets     []string                `json:"expected_absent_targets"`
		} `json:"palace_fact_cases"`
		PatternCases []struct {
			ID              string                  `json:"id"`
			Palaces         []patternSemanticPalace `json:"palaces"`
			ExpectedPresent []string                `json:"expected_present"`
			ExpectedAbsent  []string                `json:"expected_absent"`
		} `json:"pattern_cases"`
	}
	if err := json.Unmarshal(raw, &doc); err != nil {
		t.Fatalf("decode oracle: %v", err)
	}
	if len(doc.SanFangCases) != 2 || len(doc.PalaceFactCases) != 2 || len(doc.PatternCases) != 17 {
		t.Fatalf("cases = %d/%d/%d, want 2/2/17", len(doc.SanFangCases), len(doc.PalaceFactCases), len(doc.PatternCases))
	}
	for _, tc := range doc.SanFangCases {
		t.Run(tc.ID, func(t *testing.T) {
			chart := buildPatternSemanticChart(t, tc.Palaces)
			got := buildSanFangInfo(chart, sanFang(chart.MingGong))
			for palaceName, want := range tc.ExpectedSiHua {
				var actual []string
				for _, item := range got {
					if item.Name == palaceName {
						actual = item.SiHua
						break
					}
				}
				if actual == nil {
					t.Fatalf("palace %q absent from san fang: %#v", palaceName, got)
				}
				if len(actual) != len(want) {
					t.Fatalf("palace %q si_hua = %#v, want %#v", palaceName, actual, want)
				}
				for i := range want {
					if actual[i] != want[i] {
						t.Fatalf("palace %q si_hua = %#v, want %#v", palaceName, actual, want)
					}
				}
			}
		})
	}
	for _, tc := range doc.PalaceFactCases {
		t.Run(tc.ID, func(t *testing.T) {
			facts := computePalaceFacts(buildPatternSemanticChart(t, tc.Palaces))
			has := func(kind, target string) bool {
				for _, fact := range facts {
					if fact.Kind == kind && fact.Target == target {
						return true
					}
				}
				return false
			}
			for _, target := range tc.ExpectedStarTargets {
				if !has("star", target) {
					t.Fatalf("star target %q absent: %#v", target, facts)
				}
			}
			for _, target := range tc.ExpectedBrightnessTargets {
				if !has("brightness", target) {
					t.Fatalf("brightness target %q absent: %#v", target, facts)
				}
			}
			for _, target := range tc.ExpectedAbsentTargets {
				if has("star", target) || has("brightness", target) {
					t.Fatalf("target %q must be absent: %#v", target, facts)
				}
			}
		})
	}
	for _, tc := range doc.PatternCases {
		t.Run(tc.ID, func(t *testing.T) {
			patterns := findPatterns(buildPatternSemanticChart(t, tc.Palaces).GongWei)
			for _, name := range tc.ExpectedPresent {
				if !hasPattern(patterns, name) {
					t.Fatalf("pattern %q absent: %#v", name, patterns)
				}
			}
			for _, name := range tc.ExpectedAbsent {
				if hasPattern(patterns, name) {
					t.Fatalf("pattern %q must be absent: %#v", name, patterns)
				}
			}
		})
	}
}
