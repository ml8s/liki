package bazi

import (
	"testing"
	"time"

	"liki-engine/internal/engine/ganzhi"
	"liki-engine/internal/engine/tianwen"
)

// ── XiaoXian golden test ──

func TestComputeBond_GoldenValues(t *testing.T) {
	// 1984-02-15 08:00 Beijing → 甲子 丙寅 己卯 戊辰
	ca := computeChartForTest(t, 1984, 2, 15, 8, ganzhi.Male)
	// 1990-06-15 12:00 Beijing
	cb := computeChartForTest(t, 1990, 6, 15, 12, ganzhi.Female)

	bond := ComputeBond(ca, cb)

	// ZhuCross: 4x4=16 pairs
	if len(bond.ZhuCross.Pairs) != 16 {
		t.Errorf("ZhuCross.Pairs len=%d, want 16", len(bond.ZhuCross.Pairs))
	}

	// ShiShenCross: A's 日主己土, B's 日主 depends on chart
	if len(bond.ShiShenCross.AToB) != 4 {
		t.Errorf("ShiShenCross.AToB len=%d, want 4", len(bond.ShiShenCross.AToB))
	}
	if len(bond.ShiShenCross.BToA) != 4 {
		t.Errorf("ShiShenCross.BToA len=%d, want 4", len(bond.ShiShenCross.BToA))
	}

	// NayinCross: 4x4=16 pairs
	if len(bond.NayinCross.Pairs) != 16 {
		t.Errorf("NayinCross.Pairs len=%d, want 16", len(bond.NayinCross.Pairs))
	}

	// Each nayin pair should have a valid element relation
	for i, pair := range bond.NayinCross.Pairs {
		if pair.Relation == "" {
			t.Errorf("NayinCross pair %d: empty relation", i)
		}
	}

	// Verify day-gan shiShen A→B and B→A are computed
	aToBDay := bond.ShiShenCross.AToB["ri_gan"]
	bToADay := bond.ShiShenCross.BToA["ri_gan"]
	if aToBDay == "" {
		t.Error("AToB ri_gan is empty")
	}
	if bToADay == "" {
		t.Error("BToA ri_gan is empty")
	}
	t.Logf("A(己) → B day gan: %s | B → A day gan: %s", aToBDay, bToADay)
}

func computeChartForTest(t *testing.T, year, month, day, hour int, g ganzhi.Gender) Chart {
	t.Helper()
	st := tianwen.GregorianToSolar(time.Date(year, time.Month(month), day, hour, 0, 0, 0, time.FixedZone("", int(8*3600))), 116.4, 8)
	return ComputeChart(st, g)
}

// ── bond 实五行 / 用神互见 / 配偶星（数据驱动 / 自洽校验）──
func TestComputeBond_ElementYongShenAndSpouseStar(t *testing.T) {
	ca := computeChartForTest(t, 1984, 2, 15, 8, ganzhi.Male) // 甲子 丙寅 己卯 戊辰
	cb := computeChartForTest(t, 1990, 6, 15, 12, ganzhi.Female)
	bond := ComputeBond(ca, cb)
	aFc, bFc := ComputeFullChart(ca), ComputeFullChart(cb)
	aCount := convertWuxingCount(computeElementCount(ca.ToBazi(), computeCangGan(ca.ToBazi())))
	bCount := convertWuxingCount(computeElementCount(cb.ToBazi(), computeCangGan(cb.ToBazi())))

	if bond.DayMaster.A.Gan != "己" || bond.DayMaster.B.Gan == "" || bond.DayMaster.GanRel.Relation == "" {
		t.Fatalf("day master cross = %+v", bond.DayMaster)
	}
	if bond.SpousePalace.AZhi != "卯" || bond.SpousePalace.BZhi == "" || bond.SpousePalace.Relation.Detail == "" {
		t.Fatalf("spouse palace cross = %+v", bond.SpousePalace)
	}

	combined := map[string]int{}
	for k, v := range aCount {
		combined[k] += v
	}
	for k, v := range bCount {
		combined[k] += v
	}
	for k, v := range combined {
		if bond.ElementCross.Combined[k] != v {
			t.Fatalf("combined wuxing[%s]=%d, want %d", k, bond.ElementCross.Combined[k], v)
		}
	}

	assertFit := func(label string, got yongShenFitEntry, fc FullChart, other map[string]int) {
		want := fc.FuYi
		if got.Yong != want.Yong || got.Xi != want.Xi || got.Ji != want.Ji {
			t.Fatalf("%s fit = %+v, want yong/xi/ji %s/%s/%s", label, got, want.Yong, want.Xi, want.Ji)
		}
		if got.YongInOther != other[got.Yong] || got.XiInOther != other[got.Xi] || got.JiInOther != other[got.Ji] {
			t.Fatalf("%s fit counts = %+v, other=%v", label, got, other)
		}
	}
	assertFit("A", bond.ElementCross.FuYi.A, aFc, bCount)
	assertFit("B", bond.ElementCross.FuYi.B, bFc, aCount)

	if bond.SpouseStar.A.Gender != "male" ||
		bond.SpouseStar.A.Primary.TenGod != "正财" || bond.SpouseStar.A.Secondary.TenGod != "偏财" {
		t.Fatalf("A spouse star = %+v", bond.SpouseStar.A)
	}
	if bond.SpouseStar.B.Gender != "female" ||
		bond.SpouseStar.B.Primary.TenGod != "正官" || bond.SpouseStar.B.Secondary.TenGod != "七杀" {
		t.Fatalf("B spouse star = %+v", bond.SpouseStar.B)
	}
	for _, fact := range []spouseStarFact{bond.SpouseStar.A, bond.SpouseStar.B} {
		for _, group := range []spouseStarGroup{fact.Primary, fact.Secondary} {
			for _, occ := range group.Occurrences {
				if occ.Pillar == "" || occ.Branch == "" || occ.Stem == "" || occ.Source == "" {
					t.Fatalf("%s spouse occurrence incomplete: %+v", group.TenGod, occ)
				}
			}
		}
	}
}
