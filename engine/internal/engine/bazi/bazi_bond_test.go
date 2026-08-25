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

	// Verify day-stem shiShen A→B and B→A are computed
	aToBDay := bond.ShiShenCross.AToB["ri_stem"]
	bToADay := bond.ShiShenCross.BToA["ri_stem"]
	if aToBDay == "" {
		t.Error("AToB day_stem is empty")
	}
	if bToADay == "" {
		t.Error("BToA day_stem is empty")
	}
	t.Logf("A(己) → B day stem: %s | B → A day stem: %s", aToBDay, bToADay)
}

func computeChartForTest(t *testing.T, year, month, day, hour int, g ganzhi.Gender) Chart {
	t.Helper()
	st := tianwen.GregorianToSolar(time.Date(year, time.Month(month), day, hour, 0, 0, 0, time.FixedZone("", int(8*3600))), 116.4, 8)
	return ComputeChart(st, g)
}

// ── bond 纳音五行分布 + 用神互见（数据驱动/自洽校验）──
func TestComputeBond_NayinElementsYongShen(t *testing.T) {
	ca := computeChartForTest(t, 1984, 2, 15, 8, ganzhi.Male) // 甲子 丙寅 己卯 戊辰
	cb := computeChartForTest(t, 1990, 6, 15, 12, ganzhi.Female)
	bond := ComputeBond(ca, cb)
	nc := bond.NayinCross

	// 五行分布：四柱纳音五行计数，总和=4，且与各自 NaYinArray 一致
	for _, cc := range []struct {
		label string
		ny    [4]string
		got   map[string]int
	}{
		{"A", ca.NaYinArray(), nc.Elements.A},
		{"B", cb.NaYinArray(), nc.Elements.B},
	} {
		want := map[string]int{}
		for _, s := range cc.ny {
			want[ganzhi.NayinWuxing(s).String()]++
		}
		if len(cc.got) != len(want) {
			t.Errorf("%s 五行分布条目数 = %d, want %d", cc.label, len(cc.got), len(want))
		}
		for k, v := range want {
			if cc.got[k] != v {
				t.Errorf("%s 五行[%s] = %d, want %d", cc.label, k, cc.got[k], v)
			}
		}
	}

	// 用神互见：Yong/Ji 与各自 fullchart 扶抑用神一致；计数 = 对方纳音五行分布中出现次数
	aFc, bFc := ComputeFullChart(ca), ComputeFullChart(cb)
	assertEntry := func(label string, e yongShenEntry, selfYong, selfJi string, other [4]string) {
		if e.Yong != selfYong || e.Ji != selfJi {
			t.Errorf("%s 用神/忌神 = %q/%q, want %q/%q", label, e.Yong, e.Ji, selfYong, selfJi)
		}
		oCount := map[string]int{}
		for _, s := range other {
			oCount[ganzhi.NayinWuxing(s).String()]++
		}
		if e.YongInOther != oCount[e.Yong] || e.JiInOther != oCount[e.Ji] {
			t.Errorf("%s 用神入对方 = %d/%d, want %d/%d", label, e.YongInOther, e.JiInOther, oCount[e.Yong], oCount[e.Ji])
		}
	}
	assertEntry("A", nc.YongShen.A, aFc.YongShen.FuYi.Yong, aFc.YongShen.FuYi.Ji, cb.NaYinArray())
	assertEntry("B", nc.YongShen.B, bFc.YongShen.FuYi.Yong, bFc.YongShen.FuYi.Ji, ca.NaYinArray())
}
