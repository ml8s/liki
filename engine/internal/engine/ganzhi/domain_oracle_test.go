package ganzhi

import (
	"encoding/json"
	"os"
	"testing"
)

type baziOracle struct {
	TenGods []struct {
		Day      string `json:"day"`
		Other    string `json:"other"`
		Expected string `json:"expected"`
	} `json:"ten_gods_all_100"`
	Hidden []struct {
		Branch string  `json:"branch"`
		Main   string  `json:"main"`
		Mid    *string `json:"mid"`
		Minor  *string `json:"minor"`
	} `json:"hidden_stems_all_12"`
	Nayin []struct {
		First    string `json:"first"`
		Second   string `json:"second"`
		Expected string `json:"expected"`
	} `json:"nayin_all_30_pairs"`
	XunKong []struct {
		Leader string   `json:"leader"`
		Void   []string `json:"void"`
	} `json:"xun_kong_all_6"`
	ChangSheng []struct {
		Stem   string `json:"stem"`
		Branch string `json:"branch"`
		Stage  string `json:"stage"`
	} `json:"chang_sheng_all_120"`
	Lu map[string]string `json:"ten_stem_lu"`
}

func loadBaziOracle(t *testing.T) baziOracle {
	t.Helper()
	raw, err := os.ReadFile("../../../../tests/fixtures/domain_oracle/bazi_core.json")
	if err != nil {
		t.Fatalf("read domain oracle: %v", err)
	}
	var doc baziOracle
	if err := json.Unmarshal(raw, &doc); err != nil {
		t.Fatalf("decode domain oracle: %v", err)
	}
	return doc
}

func TestDomainOracle_TenGodsAll100(t *testing.T) {
	doc := loadBaziOracle(t)
	for _, tc := range doc.TenGods {
		dm, err := ParseGan(tc.Day)
		if err != nil {
			t.Fatal(err)
		}
		otherGan, err := ParseGan(tc.Other)
		if err != nil {
			t.Fatal(err)
		}
		if got := ShiShenName(ShiShenFromGan(dm, otherGan)); got != tc.Expected {
			t.Errorf("%s见%s = %s, want %s", tc.Day, tc.Other, got, tc.Expected)
		}
	}
}

func TestDomainOracle_HiddenStemsAll12(t *testing.T) {
	doc := loadBaziOracle(t)
	for _, tc := range doc.Hidden {
		z, err := ParseZhi(tc.Branch)
		if err != nil {
			t.Fatal(err)
		}
		got := CangGanForZhi(z)
		if GanName(*got.Main) != tc.Main ||
			(tc.Mid != nil && GanName(*got.Mid) != *tc.Mid) ||
			(tc.Minor != nil && GanName(*got.Minor) != *tc.Minor) {
			t.Errorf("%s hidden = %+v, want %s/%v/%v", tc.Branch, got, tc.Main, tc.Mid, tc.Minor)
		}
	}
}

func TestDomainOracle_NayinXunKongChangSheng(t *testing.T) {
	doc := loadBaziOracle(t)
	for _, tc := range doc.Nayin {
		first := []rune(tc.First)
		second := []rune(tc.Second)
		a, e1 := ParseGan(string(first[0]))
		if e1 != nil {
			t.Fatal(e1)
		}
		az, e2 := ParseZhi(string(first[1]))
		if e2 != nil {
			t.Fatal(e2)
		}
		b, e3 := ParseGan(string(second[0]))
		if e3 != nil {
			t.Fatal(e3)
		}
		bz, e4 := ParseZhi(string(second[1]))
		if e4 != nil {
			t.Fatal(e4)
		}
		if got := NayinLabel(a, az); got != tc.Expected {
			t.Errorf("%s nayin = %s, want %s", tc.First, got, tc.Expected)
		}
		if got := NayinLabel(b, bz); got != tc.Expected {
			t.Errorf("%s nayin = %s, want %s", tc.Second, got, tc.Expected)
		}
	}

	for _, tc := range doc.XunKong {
		leaderRune := []rune(tc.Leader)
		leader, err := ParseGan(string(leaderRune[0]))
		if err != nil {
			t.Fatal(err)
		}
		z, err2 := ParseZhi(string(leaderRune[1]))
		if err2 != nil {
			t.Fatal(err2)
		}
		got := XunKong(leader, z)
		if GanName(leader) != string(leaderRune[0]) {
			t.Fatal("leader")
		}
		if ZhiName(got[0]) != tc.Void[0] || ZhiName(got[1]) != tc.Void[1] {
			t.Errorf("%s xunkong = %s%s, want %s%s", tc.Leader, ZhiName(got[0]), ZhiName(got[1]), tc.Void[0], tc.Void[1])
		}
	}
	for _, tc := range doc.ChangSheng {
		g, err := ParseGan(tc.Stem)
		if err != nil {
			t.Fatal(err)
		}
		z, err2 := ParseZhi(tc.Branch)
		if err2 != nil {
			t.Fatal(err2)
		}
		want := 0
		for i, stage := range StageNamesZH {
			if stage == tc.Stage {
				want = i
				break
			}
		}
		if gotStage := ChangShengTable[g][want]; gotStage != z {
			t.Errorf("%s stage %s = %s, want %s", tc.Stem, tc.Stage, ZhiName(gotStage), tc.Branch)
		}
	}
	if len(doc.Lu) != 10 {
		t.Fatalf("ten stem lu rows = %d, want 10", len(doc.Lu))
	}
	for ganName, zhiName := range doc.Lu {
		g, err := ParseGan(ganName)
		if err != nil {
			t.Fatal(err)
		}
		z, err := ParseZhi(zhiName)
		if err != nil {
			t.Fatal(err)
		}
		if LuZhi(g) != z {
			t.Errorf("%s lu = %s, want %s", ganName, ZhiName(LuZhi(g)), zhiName)
		}
	}
}
