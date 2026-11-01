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
			item := newStarInfo(star, int(star) < 14)
			item.SiHua = semantic.SiHua
			item.Brightness = semantic.Brightness
			stars = append(stars, item)
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
	if len(doc.PatternCases) != 17 {
		t.Fatalf("pattern cases = %d, want 17", len(doc.PatternCases))
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
