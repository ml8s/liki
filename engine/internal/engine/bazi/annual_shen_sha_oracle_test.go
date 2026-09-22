package bazi

import (
	"encoding/json"
	"os"
	"testing"

	"liki-engine/internal/engine/ganzhi"
)

func TestDomainOracle_AnnualShenShaPositions(t *testing.T) {
	raw, err := os.ReadFile("../../../../tests/fixtures/domain_oracle/bazi_annual_shen_sha.json")
	if err != nil {
		t.Fatalf("read oracle: %v", err)
	}
	var doc struct {
		Positions map[string]map[string]string `json:"positions"`
	}
	if err := json.Unmarshal(raw, &doc); err != nil {
		t.Fatalf("decode oracle: %v", err)
	}
	if len(doc.Positions) != 4 {
		t.Fatalf("year anchors = %d, want 4", len(doc.Positions))
	}
	for yearText, positions := range doc.Positions {
		year, err := ganzhi.ParseZhi(yearText)
		if err != nil {
			t.Fatalf("parse year %s: %v", yearText, err)
		}
		for name, targetText := range positions {
			t.Run(yearText+"-"+name, func(t *testing.T) {
				target, err := ganzhi.ParseZhi(targetText)
				if err != nil {
					t.Fatalf("parse target %s: %v", targetText, err)
				}
				gan := ganzhi.GanYi
				if int(target)%2 == 1 {
					gan = ganzhi.GanJia
				}
				pillar := ganzhi.Zhu{Gan: gan, Zhi: target}
				bz := ganzhi.Bazi{Nian: pillar, Yue: pillar, Ri: pillar, Shi: pillar}
				got := computeAnnualShenSha(year, bz)
				if len(got) != 1 || got[0].Name != name {
					t.Fatalf("%s年%s位=%s：got %v, want only [%s]", yearText, name, targetText, got, name)
				}
			})
		}
	}
}
