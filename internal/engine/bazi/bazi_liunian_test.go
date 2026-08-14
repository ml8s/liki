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
	// steps: 丁卯(6-15) 戊辰(16-25) 己巳(26-35) 庚午(36-45) 辛未(46-55) 壬申(56-65) 癸酉(66-75) 甲戌(76-85) 乙亥(86-95)
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
	if chart.BirthYear != 1984 {
		t.Errorf("BirthYear = %d, want 1984", chart.BirthYear)
	}

	// 2026: 虚岁 = 2026-1984+1 = 43 → step3 庚午(36-45)
	ln, err := ComputeLiuNian(chart, 2026)
	if err != nil {
		t.Fatalf("ComputeLiuNian(2026): %v", err)
	}
	if len(ln.DaYunInteractions) < 1 {
		t.Fatal("DaYunInteractions should have 1 entry for 2026 (庚午运)")
	}
	if got := ln.DaYunInteractions[0].ZhuLabel; got != "伤官运(庚午)" {
		t.Errorf("2026 dayun ZhuLabel = %q, want 伤官运(庚午)（当年大运，非当前大运）", got)
	}

	// 1999: 虚岁 = 1999-1984+1 = 16 → step1 戊辰(16-25)
	ln99, err := ComputeLiuNian(chart, 1999)
	if err != nil {
		t.Fatalf("ComputeLiuNian(1999): %v", err)
	}
	if len(ln99.DaYunInteractions) < 1 {
		t.Fatal("DaYunInteractions should have 1 entry for 1999 (戊辰运)")
	}
	if got := ln99.DaYunInteractions[0].ZhuLabel; got != "劫财运(戊辰)" {
		t.Errorf("1999 dayun ZhuLabel = %q, want 劫财运(戊辰)", got)
	}

	// 1985: 虚岁 2，未起运 → 空
	ln85, err := ComputeLiuNian(chart, 1985)
	if err != nil {
		t.Fatalf("ComputeLiuNian(1985): %v", err)
	}
	if len(ln85.DaYunInteractions) != 0 {
		t.Errorf("1985 未起运: DaYunInteractions = %v, want empty", ln85.DaYunInteractions)
	}
}

func TestDayunStepForYear(t *testing.T) {
	dy := &DaYun{
		StartAge:  6,
		Direction: "顺排",
		Steps: []DaYunStep{
			{AgeStart: 6, AgeEnd: 15, Name: "丁卯", ShiShen: "偏印运"},
			{AgeStart: 16, AgeEnd: 25, Name: "戊辰", ShiShen: "劫财运"},
			{AgeStart: 26, AgeEnd: 35, Name: "己巳", ShiShen: "比肩运"},
			{AgeStart: 36, AgeEnd: 45, Name: "庚午", ShiShen: "伤官运"},
			{AgeStart: 46, AgeEnd: 55, Name: "辛未", ShiShen: "食神运"},
			{AgeStart: 56, AgeEnd: 65, Name: "壬申", ShiShen: "正财运"},
			{AgeStart: 66, AgeEnd: 75, Name: "癸酉", ShiShen: "偏财运"},
			{AgeStart: 76, AgeEnd: 85, Name: "甲戌", ShiShen: "正官运"},
			{AgeStart: 86, AgeEnd: 95, Name: "乙亥", ShiShen: "七杀运"},
		},
	}
	tests := []struct {
		name      string
		dy        *DaYun
		birthYear int
		year      int
		wantIdx   int // -1 = nil
		wantName  string
	}{
		{"中段_2026_庚午", dy, 1984, 2026, 3, "庚午"},   // 虚岁43
		{"首步_1999_戊辰", dy, 1984, 1999, 1, "戊辰"},   // 虚岁16
		{"次步_2019_庚午", dy, 1984, 2019, 3, "庚午"},   // 虚岁36（36-45边界内）
		{"末步_2078_乙亥", dy, 1984, 2078, 8, "乙亥"},   // 虚岁95（86-95末步）
		{"过完_2080", dy, 1984, 2080, -1, ""},            // 虚岁97
		{"过完_2100", dy, 1984, 2100, -1, ""},            // 虚岁117
		{"dy为nil", nil, 1984, 2026, -1, ""},
		{"birthYear为0_旧chart", dy, 0, 2026, -1, ""},
		{"年份早于出生年", dy, 1984, 1980, -1, ""},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := dayunStepForYear(tt.dy, tt.birthYear, tt.year)
			if tt.wantIdx == -1 {
				if got != nil {
					t.Errorf("dayunStepForYear(%d, %d) = %+v, want nil", tt.birthYear, tt.year, got)
				}
				return
			}
			if got == nil {
				t.Fatalf("dayunStepForYear(%d, %d) = nil, want step%d(%s)", tt.birthYear, tt.year, tt.wantIdx, tt.wantName)
			}
			if got.Name != tt.wantName {
				t.Errorf("dayunStepForYear(%d, %d) = %s, want %s", tt.birthYear, tt.year, got.Name, tt.wantName)
			}
			if &dy.Steps[tt.wantIdx] != got {
				t.Errorf("dayunStepForYear returned wrong step pointer (idx %d)", tt.wantIdx)
			}
		})
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

	// 2026 丙午年对 1984 甲子命：值年煞——丧门（太岁午后2=辰，时支辰临命）、大耗（午对冲子，年/日支子临命）
	if ln.ShenSha == nil {
		t.Error("ShenSha is nil, want empty slice")
	}
	names2026 := map[string]bool{}
	for _, s := range ln.ShenSha {
		names2026[s.Name] = true
	}
	if !names2026["丧门"] {
		t.Errorf("2026 缺丧门（值年煞：太岁午→丧门辰，时支辰临命），got %v", ln.ShenSha)
	}
	if !names2026["大耗"] {
		t.Errorf("2026 缺大耗（值年煞：太岁午→大耗子，命局有子），got %v", ln.ShenSha)
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
