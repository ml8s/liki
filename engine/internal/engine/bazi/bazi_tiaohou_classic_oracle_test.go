package bazi

import (
	"encoding/json"
	"os"
	"testing"

	"liki-engine/internal/engine/ganzhi"
)

// The fixture is deliberately separate from tiaohou.json: it records only
// classical corrections whose source wording has been checked. This test keeps
// those corrected anchors from regressing.
func TestTiaoHou_ClassicCorrectionOracle(t *testing.T) {
	raw, err := os.ReadFile("../../../../tests/fixtures/domain_oracle/bazi_tiaohou_classic_corrections.json")
	if err != nil {
		t.Fatal(err)
	}

	var doc struct {
		Source  string `json:"source"`
		Entries []struct {
			ID                 string `json:"id"`
			RiYuan             string `json:"ri_yuan"`
			MonthBranch        string `json:"month_branch"`
			Primary            string `json:"primary"`
			Secondary          string `json:"secondary"`
			SecondaryCondition string `json:"secondary_condition"`
			Basis              string `json:"basis"`
		} `json:"entries"`
	}
	if err := json.Unmarshal(raw, &doc); err != nil {
		t.Fatal(err)
	}
	if doc.Source == "" || len(doc.Entries) != 4 {
		t.Fatalf("classic oracle entries = %d, want 4", len(doc.Entries))
	}

	for _, tc := range doc.Entries {
		t.Run(tc.ID, func(t *testing.T) {
			dm, err := ganzhi.ParseGan(tc.RiYuan)
			if err != nil {
				t.Fatal(err)
			}
			mb, err := ganzhi.ParseZhi(tc.MonthBranch)
			if err != nil {
				t.Fatal(err)
			}
			got, ok := lookupTiaohou[tiaohouKey{int(dm), int(mb)}]
			if !ok {
				t.Fatalf("missing tiaohou entry for %s%s", tc.RiYuan, tc.MonthBranch)
			}
			if primary := ganzhi.GanName(got.primary); primary != tc.Primary {
				t.Fatalf("primary = %s, want %s; %s", primary, tc.Primary, tc.Basis)
			}
			secondary := ""
			if got.secondary != 0 {
				secondary = ganzhi.GanName(got.secondary)
			}
			if secondary != tc.Secondary {
				t.Fatalf("secondary = %s, want %s; %s", secondary, tc.Secondary, tc.Basis)
			}
			th, ok := queryTiaoHou(dm, mb)
			if !ok {
				t.Fatal("query tiaohou failed")
			}
			if th.SecondaryCondition != tc.SecondaryCondition {
				t.Fatalf("secondary condition = %q, want %q; %s", th.SecondaryCondition, tc.SecondaryCondition, tc.Basis)
			}
		})
	}
}
