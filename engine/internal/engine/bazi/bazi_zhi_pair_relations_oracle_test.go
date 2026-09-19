package bazi

import (
	"encoding/json"
	"os"
	"strings"
	"testing"

	"liki-engine/internal/engine/ganzhi"
)

func TestDomainOracle_ZhiPairRelations(t *testing.T) {
	raw, err := os.ReadFile("../../../../tests/fixtures/domain_oracle/bazi_zhi_pair_relations.json")
	if err != nil {
		t.Fatal(err)
	}
	var doc struct {
		Source string `json:"source"`
		Basis  string `json:"basis"`
		Cases  []struct {
			ID                  string   `json:"id"`
			A                   string   `json:"a"`
			B                   string   `json:"b"`
			ExpectedType        string   `json:"expected_type"`
			Imperial            string   `json:"imperial"`
			ExpectedDetailParts []string `json:"expected_detail_contains"`
		} `json:"cases"`
		ForbidCompleteTypesForPairOnly []string `json:"forbid_complete_types_for_pair_only"`
	}
	if err := json.Unmarshal(raw, &doc); err != nil {
		t.Fatal(err)
	}
	if doc.Basis == "" || len(doc.Cases) != 8 || len(doc.ForbidCompleteTypesForPairOnly) != 2 {
		t.Fatalf("oracle coverage/anchor incomplete: cases=%d forbids=%d", len(doc.Cases), len(doc.ForbidCompleteTypesForPairOnly))
	}
	for _, tc := range doc.Cases {
		t.Run(tc.ID, func(t *testing.T) {
			a, err := ganzhi.ParseZhi(tc.A)
			if err != nil {
				t.Fatal(err)
			}
			b, err := ganzhi.ParseZhi(tc.B)
			if err != nil {
				t.Fatal(err)
			}
			got := analyzeZhiRelation(a, b)
			if got.Type != tc.ExpectedType {
				t.Fatalf("type = %q, want %q; detail=%q", got.Type, tc.ExpectedType, got.Detail)
			}
			for _, part := range tc.ExpectedDetailParts {
				if !strings.Contains(got.Detail, part) {
					t.Fatalf("detail = %q, want contains %q", got.Detail, part)
				}
			}
			if tc.Imperial != "" && !strings.Contains(got.Detail, tc.Imperial) {
				t.Fatalf("detail = %q, want imperial %q", got.Detail, tc.Imperial)
			}
		})
	}

	// A two-branch analyzer is not a whole-chart group detector. Full 三合 / 三会
	// are owned by ComputeHeHui and must never be emitted by pair-only logic.
	for a := 1; a <= 12; a++ {
		for b := 1; b <= 12; b++ {
			got := analyzeZhiRelation(ganzhi.Zhi(a), ganzhi.Zhi(b))
			for _, forbidden := range doc.ForbidCompleteTypesForPairOnly {
				if got.Type == forbidden {
					t.Fatalf("pair-only analyzer emitted %s for %s%s: %+v", forbidden, ganzhi.ZhiName(ganzhi.Zhi(a)), ganzhi.ZhiName(ganzhi.Zhi(b)), got)
				}
			}
		}
	}
}
