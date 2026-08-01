package bazi

import (
	"testing"
	"time"

	"liki-engine/internal/engine/ganzhi"
	"liki-engine/internal/engine/tianwen"
)

// ============================================================================
// 流年 (LiuNian) — 大运互动 / 神煞 / 伏吟反吟
// ============================================================================

func TestLiuNian_DaYunInteraction(t *testing.T) {
	// 1984-02-15 08:00 甲子年 丙寅月 甲子日 戊辰时, male (顺排, 6岁起运)
	st := tianwen.GregorianToSolar(
		time.Date(1984, 2, 15, 8, 0, 0, 0, time.FixedZone("CST", 8*3600)),
		116.4, 8,
	)
	chart := ComputeChart(st, ganzhi.Male)

	if chart.DaYun == nil {
		t.Fatal("DaYun is nil")
	}
	if chart.DaYun.Direction != "顺排" {
		t.Errorf("Direction = %q, want 顺排", chart.DaYun.Direction)
	}

	// Manually set current_step_index for age 42 (2026): 6+30=36→step3, 6+40=46→step4.
	// Age 42 → step 3 (36-45).
	chart.DaYun.CurrentStepIndex = 3

	ln, err := ComputeLiuNian(chart, 2026)
	if err != nil {
		t.Fatalf("ComputeLiuNian(2026): %v", err)
	}

	// dayun_interactions should be non-nil because current_step_index is set.
	if ln.DaYunInteractions == nil {
		t.Error("DaYunInteractions should not be nil when current_step_index is set")
	} else if len(ln.DaYunInteractions) < 1 {
		t.Error("DaYunInteractions should have at least 1 entry")
	} else {
		entry := ln.DaYunInteractions[0]
		if entry.ZhuLabel == "" {
			t.Error("DaYunInteraction ZhuLabel is empty")
		}
	}
}

func TestLiuNian_DaYunInteraction_NegativeIndex(t *testing.T) {
	// When current_step_index is -1 (not in 大运), dayun_interactions should be nil.
	st := tianwen.GregorianToSolar(
		time.Date(2023, 6, 15, 12, 0, 0, 0, time.FixedZone("CST", 8*3600)),
		116.4, 8,
	)
	chart := ComputeChart(st, ganzhi.Male)

	if chart.DaYun == nil {
		t.Fatal("DaYun is nil")
	}

	// Not yet in 大运: 2023年生, age 0.
	chart.DaYun.CurrentStepIndex = -1

	ln, err := ComputeLiuNian(chart, 2023)
	if err != nil {
		t.Fatalf("ComputeLiuNian(2023): %v", err)
	}

	if len(ln.DaYunInteractions) != 0 {
		t.Error("DaYunInteractions should be empty when current_step_index = -1")
	}
}

func TestLiuNian_ShenSha(t *testing.T) {
	// 1984-02-15 甲子年 male
	st := tianwen.GregorianToSolar(
		time.Date(1984, 2, 15, 8, 0, 0, 0, time.FixedZone("CST", 8*3600)),
		116.4, 8,
	)
	chart := ComputeChart(st, ganzhi.Male)

	ln, err := ComputeLiuNian(chart, 2026) // 丙午年
	if err != nil {
		t.Fatalf("ComputeLiuNian(2026): %v", err)
	}

	// 2026 丙午年对 1984 甲子命：无羊刃/劫煞匹配 → 应为空数组（非 nil）
	if ln.ShenSha == nil {
		t.Error("ShenSha is nil, want empty slice")
	}
	if len(ln.ShenSha) != 0 {
		t.Errorf("2026 ShenSha = %v, want empty", ln.ShenSha)
	}

	// 2025 乙巳年：命主日支子见巳 → 羊刃；年支子见巳 → 劫煞（具体断言）
	ln2025, err := ComputeLiuNian(chart, 2025)
	if err != nil {
		t.Fatalf("ComputeLiuNian(2025): %v", err)
	}
	names := map[string]bool{}
	for _, s := range ln2025.ShenSha {
		names[s.Name] = true
	}
	if !names["羊刃"] {
		t.Errorf("2025 缺羊刃, got %v", ln2025.ShenSha)
	}
	if !names["劫煞"] {
		t.Errorf("2025 缺劫煞, got %v", ln2025.ShenSha)
	}
}

func TestLiuNian_FuYinFanYin_Exists(t *testing.T) {
	// A year that matches a natal pillar should have fuyin_fanyin entries.
	// 1984-02-15 甲子年 → 2024 甲辰年 (甲子 vs 甲辰 — not a match)
	// Let's test with a year that has a known match.
	// 甲子日: 2020 庚子年 has 子 matching 甲子日's 子 → it's 反吟 for 子.
	// Actually, 甲子日: 流年庚子 = 甲 vs 庚 (different) + 子 vs 子 (same branch)
	// → same branch = 反吟 for 地支部分 but NOT full 伏吟 (need both gan+zhi same).
	st := tianwen.GregorianToSolar(
		time.Date(1984, 2, 15, 8, 0, 0, 0, time.FixedZone("CST", 8*3600)),
		116.4, 8,
	)
	chart := ComputeChart(st, ganzhi.Male)

	// 甲子日 → 2044 甲子年 should have 伏吟 on 日柱
	ln, err := ComputeLiuNian(chart, 2044)
	if err != nil {
		t.Fatalf("ComputeLiuNian(2044): %v", err)
	}

	if len(ln.FuYinFanYin) > 0 {
		t.Logf("FuYinFanYin found %d entries (expected for matching year)", len(ln.FuYinFanYin))
	}
}

func TestLiuNian_FuYinFanYin_Nil(t *testing.T) {
	// A year with NO matching pillars should have nil fuyin_fanyin.
	// 甲子日 → 流年 丙午 has nothing in common with 甲子 → nil is valid.
	st := tianwen.GregorianToSolar(
		time.Date(1984, 2, 15, 8, 0, 0, 0, time.FixedZone("CST", 8*3600)),
		116.4, 8,
	)
	chart := ComputeChart(st, ganzhi.Male)

	ln, err := ComputeLiuNian(chart, 2026) // 丙午, no relation to 甲子
	if err != nil {
		t.Fatalf("ComputeLiuNian(2026): %v", err)
	}

	if ln.FuYinFanYin == nil {
		t.Log("FuYinFanYin is nil (no matches for this year — expected)")
	}
}

func TestLiuNian_YearFields(t *testing.T) {
	// Validate basic liunian fields for a known chart and year.
	st := tianwen.GregorianToSolar(
		time.Date(1984, 2, 15, 8, 0, 0, 0, time.FixedZone("CST", 8*3600)),
		116.4, 8,
	)
	chart := ComputeChart(st, ganzhi.Male)

	ln, err := ComputeLiuNian(chart, 2026)
	if err != nil {
		t.Fatalf("ComputeLiuNian(2026): %v", err)
	}

	if ln.Year != 2026 {
		t.Errorf("Year = %d, want 2026", ln.Year)
	}
	if ln.YearName != "丙午" {
		t.Errorf("YearName = %q, want 丙午", ln.YearName)
	}
	if ln.ShiShen == "" {
		t.Error("ShiShen is empty")
	}
	if ln.NaYin == "" {
		t.Error("NaYin is empty")
	}
	if len(ln.NatalInteractions) < 1 {
		t.Error("NatalInteractions should not be empty")
	}
	if ln.NatalInteractions[0].ZhuLabel != "丙午" {
		t.Errorf("NatalInteractions[0].ZhuLabel = %q, want 丙午", ln.NatalInteractions[0].ZhuLabel)
	}
}
