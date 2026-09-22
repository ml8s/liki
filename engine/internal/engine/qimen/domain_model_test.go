package qimen

import (
	"testing"
	"time"

	"liki-engine/internal/engine/ganzhi"
	"liki-engine/internal/engine/tianwen"
)

func TestPalaceIdentityAndBranchFacts(t *testing.T) {
	chart := chartForTest()
	for i, palace := range chart.Pan.GongWei {
		expected := GongIndex(i + 1)
		if palace.Gong.Name != expected.String() || palace.Gong.Luoshu != int(expected) {
			t.Fatalf("palace[%d] identity = %+v", i, palace.Gong)
		}
		if expected == GongZhong && palace.Gong.RingIndex != nil {
			t.Fatal("central palace must not claim an outer-ring index")
		}
		if expected != GongZhong && palace.Gong.RingIndex == nil {
			t.Fatalf("%s lacks ring index", expected)
		}
	}
	for _, void := range chart.Pan.KongWang {
		if void.Branch == 0 || void.Gong == 0 || void.Gong == GongZhong {
			t.Fatalf("invalid void branch fact %+v", void)
		}
	}
	if chart.Pan.MaXing.Branch == 0 || chart.Pan.MaXing.Gong == 0 {
		t.Fatalf("invalid horse fact %+v", chart.Pan.MaXing)
	}
}

func TestYingQiUsesYongShenFocus(t *testing.T) {
	chart := chartForTest()
	focuses := map[GongIndex][]RelatedSymbol{
		chart.Pan.KongWang[0].Gong: {{Symbol: "生门", Role: "yong_shen"}},
	}
	result := computeYingQi(chart, time.Date(
		1984, 2, 15, 8, 0, 0, 0, time.FixedZone("CST", 8*3600),
	), focuses)
	found := false
	for _, candidate := range result.Candidates {
		if candidate.Gong != chart.Pan.KongWang[0].Gong {
			continue
		}
		for _, related := range candidate.RelatedTo {
			if related.Symbol == "生门" && related.Role == "yong_shen" {
				found = true
			}
		}
	}
	if !found {
		t.Fatal("void candidate did not retain its selected yong-shen focus")
	}
}

func TestComputeChartWithYongShenProjectsTiming(t *testing.T) {
	st := tianwen.GregorianToSolar(
		time.Date(1984, 2, 15, 8, 0, 0, 0, time.FixedZone("CST", 8*3600)),
		116.4, 8,
	)
	symbol, err := ParseYongShen("生门")
	if err != nil {
		t.Fatal(err)
	}
	projected, err := ComputeChartWithYongShenAndMethod(
		st, []YongShenSymbol{symbol}, BirthDate{},
		Method{Scope: ScopeHour, School: SchoolZhuanPan, Dingju: DingjuChaiBu},
	)
	if err != nil {
		t.Fatal(err)
	}
	if projected.YongShen == nil || len(projected.YongShen.Symbols) != 1 {
		t.Fatal("selected yong shen missing")
	}
	palace := projected.YongShen.Symbols[0].Palace
	for _, candidate := range projected.YingQi.Candidates {
		if candidate.Gong != palace {
			continue
		}
		found := false
		for _, related := range candidate.RelatedTo {
			if related.Symbol == "生门" && related.Role == "yong_shen" {
				found = true
			}
		}
		if !found {
			t.Fatalf("candidate at %s lacks 生门 focus: %+v", palace, candidate)
		}
	}
}

func TestParseBirthDatePrecision(t *testing.T) {
	if _, err := ParseBirthDate("2024-02-04"); err == nil {
		t.Fatal("date-only Lichun boundary should be rejected")
	}
	dateOnly, err := ParseBirthDate("1984-02-15")
	if err != nil {
		t.Fatal(err)
	}
	if !dateOnly.Has || dateOnly.Precision != BirthDateOnly {
		t.Fatalf("date precision = %+v", dateOnly)
	}
	moment, err := ParseBirthDate("2024-02-04T02:00:00Z")
	if err != nil {
		t.Fatal(err)
	}
	if moment.Precision != BirthDateMoment {
		t.Fatalf("moment precision = %+v", moment)
	}
}

func TestTableDrivenSpiritNamesAndRelations(t *testing.T) {
	if SpiritGouChen.YangName() != "勾陈" || SpiritGouChen.YinName() != "白虎" {
		t.Fatal("slot-5 spirit school names are not table-driven")
	}
	if got := xingGongRelation(StarTianRui, GongGen); got != "same" {
		t.Fatalf("TianRui/Gen relation = %s, want same", got)
	}
	if wuxingRelations["same"].TraditionalLabel != "旺" {
		t.Fatal("same-element traditional label is missing")
	}
	if riShiRelations["ri_controls_shi"] == "" {
		t.Fatal("ri-shi relation label is missing")
	}
}

func TestFuTouYuanTableCoversAllBranches(t *testing.T) {
	for branch := ganzhi.ZhiZi; branch <= ganzhi.ZhiHai; branch++ {
		if _, ok := fuTouYuanTable[branch]; !ok {
			t.Errorf("branch %s has no fu-tou yuan", branch)
		}
	}
}
