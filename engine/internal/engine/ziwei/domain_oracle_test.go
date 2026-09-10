package ziwei

import (
	"encoding/json"
	"os"
	"testing"

	"liki-engine/internal/engine/ganzhi"
)

func loadZiweiOracle(t *testing.T) struct {
	SchemaVersion string `json:"schema_version"`
	SiHua         []struct {
		Stem            string            `json:"stem"`
		Transformations map[string]string `json:"transformations"`
	} `json:"si_hua_all_10"`
	Soul []struct {
		Branch   string `json:"palace_branch"`
		Expected string `json:"expected"`
	} `json:"soul_star_all_12"`
	Body []struct {
		Branch   string `json:"year_branch"`
		Expected string `json:"expected"`
	} `json:"body_star_all_12"`
} {
	t.Helper()
	raw, err := os.ReadFile("../../../../tests/fixtures/domain_oracle/ziwei_core.json")
	if err != nil {
		t.Fatalf("read ziwei oracle: %v", err)
	}
	var doc struct {
		SchemaVersion string `json:"schema_version"`
		SiHua         []struct {
			Stem            string            `json:"stem"`
			Transformations map[string]string `json:"transformations"`
		} `json:"si_hua_all_10"`
		Soul []struct {
			Branch   string `json:"palace_branch"`
			Expected string `json:"expected"`
		} `json:"soul_star_all_12"`
		Body []struct {
			Branch   string `json:"year_branch"`
			Expected string `json:"expected"`
		} `json:"body_star_all_12"`
	}
	if err := json.Unmarshal(raw, &doc); err != nil {
		t.Fatalf("decode ziwei oracle: %v", err)
	}
	return doc
}

func TestDomainOracle_ZiweiCoreTables(t *testing.T) {
	doc := loadZiweiOracle(t)
	for _, tc := range doc.SiHua {
		g, err := ganzhi.ParseGan(tc.Stem)
		if err != nil {
			t.Fatal(err)
		}
		want := [4]string{tc.Transformations["禄"], tc.Transformations["权"], tc.Transformations["科"], tc.Transformations["忌"]}
		got := siHuaTable[g]
		for i, name := range want {
			if starName(got[i]) != name {
				t.Errorf("%s transformation %d = %s, want %s", tc.Stem, i, starName(got[i]), name)
			}
		}
	}
	for _, tc := range doc.Soul {
		z, err := ganzhi.ParseZhi(tc.Branch)
		if err != nil {
			t.Fatal(err)
		}
		if got := soulStar(z); got != tc.Expected {
			t.Errorf("soul %s = %s, want %s", tc.Branch, got, tc.Expected)
		}
	}
	for _, tc := range doc.Body {
		z, err := ganzhi.ParseZhi(tc.Branch)
		if err != nil {
			t.Fatal(err)
		}
		if got := bodyStar(z); got != tc.Expected {
			t.Errorf("body %s = %s, want %s", tc.Branch, got, tc.Expected)
		}
	}
}
