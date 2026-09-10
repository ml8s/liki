package qimen

import (
	"encoding/json"
	"os"
	"testing"
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
	}
	if err := json.Unmarshal(raw, &doc); err != nil {
		t.Fatalf("decode qimen oracle: %v", err)
	}
	if len(doc.Cases) != 72 {
		t.Fatalf("rows = %d, want 72", len(doc.Cases))
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
}
