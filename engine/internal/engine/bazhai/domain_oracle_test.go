package bazhai

import (
	"encoding/json"
	"os"
	"testing"

	"liki-engine/internal/engine/ganzhi"
)

func TestDomainOracle_BazhaiCore(t *testing.T) {
	raw, err := os.ReadFile("../../../../tests/fixtures/domain_oracle/fengshui_core.json")
	if err != nil {
		t.Fatalf("read oracle: %v", err)
	}
	var doc struct {
		Patterns map[string]struct {
			ShengQi int `json:"sheng_qi"`
			TianYi  int `json:"tian_yi"`
			YanNian int `json:"yan_nian"`
			FuWei   int `json:"fu_wei"`
			HuoHai  int `json:"huo_hai"`
			WuGui   int `json:"wu_gui"`
			LiuSha  int `json:"liu_sha"`
			JueMing int `json:"jue_ming"`
		} `json:"eight_mansion_patterns"`
		MingGua []struct {
			Gender string `json:"gender"`
			Year   int    `json:"year"`
			Gua    string `json:"gua"`
			Group  string `json:"group"`
		} `json:"ming_gua_anchors"`
	}
	if err := json.Unmarshal(raw, &doc); err != nil {
		t.Fatalf("decode oracle: %v", err)
	}
	nameToNum := map[string]int{"坎": 1, "坤": 2, "震": 3, "巽": 4, "乾": 6, "兑": 7, "艮": 8, "离": 9}
	for name, want := range doc.Patterns {
		num := nameToNum[name]
		got := eightMansionPatterns[num]
		if got.shengQi != want.ShengQi || got.tianYi != want.TianYi || got.yanNian != want.YanNian || got.fuWei != want.FuWei ||
			got.huoHai != want.HuoHai || got.wuGui != want.WuGui || got.liuSha != want.LiuSha || got.jueMing != want.JueMing {
			t.Fatalf("%s pattern = %+v, want %+v", name, got, want)
		}
	}
	for _, want := range doc.MingGua {
		gender := ganzhi.Male
		if want.Gender == "female" {
			gender = ganzhi.Female
		}
		got := ComputeMingGua(gender, want.Year)
		if got.YearBoundary != "gregorian_calendar_year" {
			t.Fatalf("%s %d year_boundary = %s", want.Gender, want.Year, got.YearBoundary)
		}
		if got.Gua.Name != want.Gua || got.Group != want.Group {
			t.Errorf("%s %d = %s/%s, want %s/%s", want.Gender, want.Year, got.Gua.Name, got.Group, want.Gua, want.Group)
		}
	}
}
