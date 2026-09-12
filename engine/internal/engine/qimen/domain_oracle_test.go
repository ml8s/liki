package qimen

import (
	"encoding/json"
	"os"
	"testing"
	"time"
)

func TestDomainOracle_QimenAll72Ju(t *testing.T) {
	raw, err := os.ReadFile("../../../../tests/fixtures/domain_oracle/qimen_core.json")
	if err != nil {
		t.Fatalf("read qimen oracle: %v", err)
	}
	var doc struct {
		Cases []struct {
			Term    string `json:"term"`
			Yuan    string `json:"yuan"`
			Ju      int    `json:"ju"`
			YangDun bool   `json:"yang_dun"`
		} `json:"solar_term_ju_all_72"`
		YingQi []struct {
			ID             string `json:"id"`
			DutyStarPalace string `json:"duty_star_palace"`
			DutyDoorPalace string `json:"duty_door_palace"`
			Expected       []struct {
				Type   string `json:"type"`
				Branch string `json:"branch"`
				Gong   string `json:"gong"`
			} `json:"expected"`
		} `json:"yingqi_cases"`
	}
	if err := json.Unmarshal(raw, &doc); err != nil {
		t.Fatalf("decode qimen oracle: %v", err)
	}
	if len(doc.Cases) != 72 {
		t.Fatalf("rows = %d, want 72", len(doc.Cases))
	}
	if len(doc.YingQi) != 2 {
		t.Fatalf("yingqi rows = %d, want 2", len(doc.YingQi))
	}
	for _, tc := range doc.Cases {
		entry, ok := solarTermJuTable[tc.Term]
		if !ok {
			t.Fatalf("missing term %s", tc.Term)
		}
		yuanIdx := map[string]int{"上元": 0, "中元": 1, "下元": 2}[tc.Yuan]
		if entry[yuanIdx] != tc.Ju {
			t.Errorf("%s %s ju = %d, want %d", tc.Term, tc.Yuan, entry[yuanIdx], tc.Ju)
		}
		if (entry[3] != 0) != tc.YangDun {
			t.Errorf("%s yang dun = %v, want %v", tc.Term, entry[3] != 0, tc.YangDun)
		}
	}
	for _, tc := range doc.YingQi {
		t.Run(tc.ID, func(t *testing.T) {
			chart := Chart{
				DutyStarPalace: oracleGong(t, tc.DutyStarPalace),
				DutyDoorPalace: oracleGong(t, tc.DutyDoorPalace),
			}
			got := computeYingQi(chart, time.Date(2026, 9, 9, 12, 0, 0, 0, time.FixedZone("CST", 8)), baseYingQiFocuses(chart))
			if len(got.Candidates) != len(tc.Expected) {
				t.Fatalf("candidates = %#v, want %d rows", got.Candidates, len(tc.Expected))
			}
			for i, want := range tc.Expected {
				item := got.Candidates[i]
				if item.Type != want.Type || item.Branch != want.Branch || item.Gong.String() != want.Gong {
					t.Fatalf("candidate %d = %+v, want %+v", i, item, want)
				}
			}
		})
	}
}
