package fengshui

import (
	"encoding/json"
	"os"
	"strconv"
	"testing"
)

func loadFengshuiOracle(t *testing.T) struct {
	Mountains []struct {
		Index    int    `json:"index"`
		Name     string `json:"name"`
		Angle    int    `json:"angle"`
		Trigram  string `json:"trigram"`
		YuanLong string `json:"yuan_long"`
		YinYang  string `json:"yin_yang"`
		Wuxing   string `json:"wuxing"`
	} `json:"mountains_24"`
	Annual []struct {
		Year    int    `json:"year"`
		RuZhong string `json:"ru_zhong"`
	} `json:"annual_flying_star_anchors"`
	Distribution map[string]int `json:"annual_distribution_1984"`
} {
	t.Helper()
	raw, err := os.ReadFile("../../../../tests/fixtures/domain_oracle/fengshui_core.json")
	if err != nil {
		t.Fatalf("read oracle: %v", err)
	}
	var doc struct {
		Mountains []struct {
			Index    int    `json:"index"`
			Name     string `json:"name"`
			Angle    int    `json:"angle"`
			Trigram  string `json:"trigram"`
			YuanLong string `json:"yuan_long"`
			YinYang  string `json:"yin_yang"`
			Wuxing   string `json:"wuxing"`
		} `json:"mountains_24"`
		Annual []struct {
			Year    int    `json:"year"`
			RuZhong string `json:"ru_zhong"`
		} `json:"annual_flying_star_anchors"`
		Distribution map[string]int `json:"annual_distribution_1984"`
	}
	if err := json.Unmarshal(raw, &doc); err != nil {
		t.Fatalf("decode oracle: %v", err)
	}
	return doc
}

func TestDomainOracle_Mountains24(t *testing.T) {
	doc := loadFengshuiOracle(t)
	if len(doc.Mountains) != 24 {
		t.Fatalf("mountains = %d, want 24", len(doc.Mountains))
	}
	for _, want := range doc.Mountains {
		got := Mountains24Table[want.Index]
		if got.Name != want.Name || got.Angle != want.Angle || got.Trigram != want.Trigram ||
			got.YuanLong != want.YuanLong || got.YinYang != want.YinYang || got.Element.String() != want.Wuxing {
			t.Fatalf("mount %d = %+v, want %+v", want.Index, got, want)
		}
	}
}

func TestDomainOracle_AnnualFlyingStars(t *testing.T) {
	doc := loadFengshuiOracle(t)
	for _, want := range doc.Annual {
		got := ComputeAnnualFlyingStars(want.Year)
		if got.RuZhong != want.RuZhong {
			t.Errorf("year %d ru_zhong = %s, want %s", want.Year, got.RuZhong, want.RuZhong)
		}
	}
	board := ComputeAnnualFlyingStars(1984)
	if board.YearBoundary != "gregorian_calendar_year" {
		t.Fatalf("year_boundary = %s, want explicit gregorian_calendar_year", board.YearBoundary)
	}
	if board.RuZhong != "七赤破军" {
		t.Fatalf("1984 ru_zhong = %s", board.RuZhong)
	}
	for _, got := range board.GongWei {
		key := strconv.Itoa(got.GongNum)
		want, ok := doc.Distribution[key]
		if !ok {
			t.Fatalf("unexpected palace %d", got.GongNum)
		}
		if got.Xing != want {
			t.Errorf("palace %d star = %d, want %d", got.GongNum, got.Xing, want)
		}
	}
}
