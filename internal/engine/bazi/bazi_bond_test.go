package bazi

import (
	"time"
	"testing"

	"liki-engine/internal/engine/ganzhi"
	"liki-engine/internal/engine/tianwen"
)

// ── XiaoXian golden test ──

func TestComputeXiaoXian_FemaleFullCycle(t *testing.T) {
	// Female: starts at 申(9), moves backward
	xs := ComputeXiaoXian(ganzhi.Female, 12)
	want := []string{"申", "未", "午", "巳", "辰", "卯", "寅", "丑", "子", "亥", "戌", "酉"}
	for i, x := range xs {
		if ganzhi.ZhiName(x.Zhi) != want[i] {
			t.Errorf("age %d: got %s, want %s", i+1, ganzhi.ZhiName(x.Zhi), want[i])
		}
	}
}

func TestComputeXiaoXian_MaleFullCycle(t *testing.T) {
	// Male: starts at 寅(3), moves forward
	xs := ComputeXiaoXian(ganzhi.Male, 12)
	want := []string{"寅", "卯", "辰", "巳", "午", "未", "申", "酉", "戌", "亥", "子", "丑"}
	for i, x := range xs {
		if ganzhi.ZhiName(x.Zhi) != want[i] {
			t.Errorf("age %d: got %s, want %s", i+1, ganzhi.ZhiName(x.Zhi), want[i])
		}
	}
}

func TestComputeXiaoXian_LargeAge(t *testing.T) {
	// Verify no overflow at large ages
	xs := ComputeXiaoXian(ganzhi.Male, 120)
	if len(xs) != 120 {
		t.Errorf("len=%d, want 120", len(xs))
	}
	// Age 120 should be same as age 12 (modulo 12 cycle)
	if xs[119].Zhi != xs[11].Zhi {
		t.Errorf("age 120: %s != age 12: %s (should be same, 12-year cycle)",
			ganzhi.ZhiName(xs[119].Zhi), ganzhi.ZhiName(xs[11].Zhi))
	}
}

func TestComputeXiaoXian_NegativeMaxAge(t *testing.T) {
	xs := ComputeXiaoXian(ganzhi.Male, -1)
	if len(xs) != 12 {
		t.Errorf("negative maxAge: len=%d, want 12 (default)", len(xs))
	}
}

// ── Bond golden test: known chart cross-verification ──

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
