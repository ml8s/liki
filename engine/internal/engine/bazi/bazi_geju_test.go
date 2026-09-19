package bazi

import (
	"liki-engine/internal/engine/ganzhi"
	"liki-engine/internal/engine/tianwen"
	"testing"
	"time"
)

// mkGeJuChart builds a Chart with shi shen entries from both month gan
// and month zhi hidden gan (支藏干). Hidden gan are auto-populated
// from the month zhi using the CangGan lookup.
func mkGeJuChart(riGan, yueGan, nianGan, shiGan ganzhi.Gan, yueZhi ganzhi.Zhi) Chart {
	return Chart{
		Ri: zhuInfo{Zhu: ganzhi.Zhu{Gan: riGan, Zhi: ganzhi.ZhiWu}},
		Yue: zhuInfo{
			Zhu: ganzhi.Zhu{Gan: yueGan, Zhi: yueZhi},
		},
		Nian: zhuInfo{Zhu: ganzhi.Zhu{Gan: nianGan, Zhi: ganzhi.ZhiSi}},
		Shi:  zhuInfo{Zhu: ganzhi.Zhu{Gan: shiGan, Zhi: ganzhi.ZhiShen}},
	}
}
func TestComputeGeJu_AllPatterns(t *testing.T) {
	// Each case: 甲日主, month gan 透月令本气 (month gan == main qi).
	tests := []struct {
		name        string
		riGan       ganzhi.Gan
		yueGan      ganzhi.Gan
		yueZhi      ganzhi.Zhi
		wantPattern string
		wantUsage   string
		note        string
	}{
		// ── 顺用六格 ──
		{
			name:  "正官格 酉月辛透月干",
			riGan: ganzhi.GanJia, yueGan: ganzhi.GanXin, yueZhi: ganzhi.ZhiYou,
			wantPattern: "正官格", wantUsage: "顺用",
		},
		{
			name:  "正财格 未月己透月干",
			riGan: ganzhi.GanJia, yueGan: ganzhi.GanJi, yueZhi: ganzhi.ZhiWei,
			wantPattern: "正财格", wantUsage: "顺用",
		},
		{
			name:  "偏财格 辰月戊透月干",
			riGan: ganzhi.GanJia, yueGan: ganzhi.GanWu, yueZhi: ganzhi.ZhiChen,
			wantPattern: "偏财格", wantUsage: "顺用",
			note: "同正财格公式",
		},
		{
			name:  "正印格 子月癸透月干",
			riGan: ganzhi.GanJia, yueGan: ganzhi.GanGui, yueZhi: ganzhi.ZhiZi,
			wantPattern: "正印格", wantUsage: "顺用",
		},
		{
			name:  "偏印格 亥月壬透月干",
			riGan: ganzhi.GanJia, yueGan: ganzhi.GanRen, yueZhi: ganzhi.ZhiHai,
			wantPattern: "偏印格", wantUsage: "顺用",
			note: "同正印格公式",
		},
		{
			name:  "食神格 巳月丙透月干",
			riGan: ganzhi.GanJia, yueGan: ganzhi.GanBing, yueZhi: ganzhi.ZhiSi,
			wantPattern: "食神格", wantUsage: "顺用",
		},
		// ── 逆用二格 ──
		{
			name:  "七杀格 申月庚透月干",
			riGan: ganzhi.GanJia, yueGan: ganzhi.GanGeng, yueZhi: ganzhi.ZhiShen,
			wantPattern: "七杀格", wantUsage: "逆用",
			note: "逆用不变",
		},
		{
			name:  "伤官格 午月丁透月干",
			riGan: ganzhi.GanJia, yueGan: ganzhi.GanDing, yueZhi: ganzhi.ZhiWu,
			wantPattern: "伤官格", wantUsage: "逆用",
			note: "逆用不变",
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			cb := mkGeJuChart(tt.riGan, tt.yueGan, ganzhi.GanGui, ganzhi.GanRen, tt.yueZhi)
			result := computeGeJu(cb, nil)
			if result.Pattern != tt.wantPattern {
				t.Errorf("Pattern = %q, want %q\n  note: %s", result.Pattern, tt.wantPattern, tt.note)
			}
			if result.Usage != tt.wantUsage {
				t.Errorf("Usage = %q, want %q\n  note: %s", result.Usage, tt.wantUsage, tt.note)
			}
		})
	}
}
func TestComputeGeJu_JianLuYueRen(t *testing.T) {
	// 禄(临官) / 刃(帝旺) per gan — unchanged.
	tests := []struct {
		name        string
		riGan       ganzhi.Gan
		yueZhi      ganzhi.Zhi
		wantPattern string
		wantUsage   string
	}{
		{name: "建禄格 甲日寅月(禄)", riGan: ganzhi.GanJia, yueZhi: ganzhi.ZhiYin,
			wantPattern: "建禄格", wantUsage: "逆用"},
		{name: "月刃格 甲日卯月(刃)", riGan: ganzhi.GanJia, yueZhi: ganzhi.ZhiMao,
			wantPattern: "月刃格", wantUsage: "逆用"},
		{name: "建禄格 乙日卯月(禄)", riGan: ganzhi.GanYi, yueZhi: ganzhi.ZhiMao,
			wantPattern: "建禄格", wantUsage: "逆用"},
		{name: "乙日寅月阴干无刃", riGan: ganzhi.GanYi, yueZhi: ganzhi.ZhiYin,
			wantPattern: "伤官格", wantUsage: "逆用"},
		{name: "建禄格 丙日巳月(禄)", riGan: ganzhi.GanBing, yueZhi: ganzhi.ZhiSi,
			wantPattern: "建禄格", wantUsage: "逆用"},
		{name: "月刃格 丙日午月(刃)", riGan: ganzhi.GanBing, yueZhi: ganzhi.ZhiWu,
			wantPattern: "月刃格", wantUsage: "逆用"},
		{name: "建禄格 丁日午月(禄)", riGan: ganzhi.GanDing, yueZhi: ganzhi.ZhiWu,
			wantPattern: "建禄格", wantUsage: "逆用"},
		{name: "丁日巳月阴干无刃", riGan: ganzhi.GanDing, yueZhi: ganzhi.ZhiSi,
			wantPattern: "杂格", wantUsage: "逆用"},
		{name: "建禄格 戊日巳月(禄)", riGan: ganzhi.GanWu, yueZhi: ganzhi.ZhiSi,
			wantPattern: "建禄格", wantUsage: "逆用"},
		{name: "月刃格 戊日午月(刃)", riGan: ganzhi.GanWu, yueZhi: ganzhi.ZhiWu,
			wantPattern: "月刃格", wantUsage: "逆用"},
		{name: "建禄格 己日午月(禄)", riGan: ganzhi.GanJi, yueZhi: ganzhi.ZhiWu,
			wantPattern: "建禄格", wantUsage: "逆用"},
		{name: "己日巳月阴干无刃", riGan: ganzhi.GanJi, yueZhi: ganzhi.ZhiSi,
			wantPattern: "正印格", wantUsage: "顺用"},
		{name: "建禄格 庚日申月(禄)", riGan: ganzhi.GanGeng, yueZhi: ganzhi.ZhiShen,
			wantPattern: "建禄格", wantUsage: "逆用"},
		{name: "月刃格 庚日酉月(刃)", riGan: ganzhi.GanGeng, yueZhi: ganzhi.ZhiYou,
			wantPattern: "月刃格", wantUsage: "逆用"},
		{name: "建禄格 辛日酉月(禄)", riGan: ganzhi.GanXin, yueZhi: ganzhi.ZhiYou,
			wantPattern: "建禄格", wantUsage: "逆用"},
		{name: "辛日申月阴干无刃", riGan: ganzhi.GanXin, yueZhi: ganzhi.ZhiShen,
			wantPattern: "伤官格", wantUsage: "逆用"},
		{name: "建禄格 壬日亥月(禄)", riGan: ganzhi.GanRen, yueZhi: ganzhi.ZhiHai,
			wantPattern: "建禄格", wantUsage: "逆用"},
		{name: "月刃格 壬日子月(刃)", riGan: ganzhi.GanRen, yueZhi: ganzhi.ZhiZi,
			wantPattern: "月刃格", wantUsage: "逆用"},
		{name: "建禄格 癸日子月(禄)", riGan: ganzhi.GanGui, yueZhi: ganzhi.ZhiZi,
			wantPattern: "建禄格", wantUsage: "逆用"},
		{name: "癸日亥月阴干无刃", riGan: ganzhi.GanGui, yueZhi: ganzhi.ZhiHai,
			wantPattern: "杂格", wantUsage: "逆用"},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			cb := mkGeJuChart(tt.riGan, ganzhi.GanBing, ganzhi.GanGui, ganzhi.GanRen, tt.yueZhi)
			result := computeGeJu(cb, nil)
			if result.Pattern != tt.wantPattern {
				t.Errorf("Pattern = %q, want %q", result.Pattern, tt.wantPattern)
			}
			if result.Usage != tt.wantUsage {
				t.Errorf("Usage = %q, want %q", result.Usage, tt.wantUsage)
			}
		})
	}
}

// TestGeJu_TouTou_Priority verifies that hidden gan 透干 takes priority
// over month gan. 酉月(辛金本气), 月干癸水, 年干辛金透干 → 正官格(不是正印格).
func TestGeJu_TouTou_Priority(t *testing.T) {
	cb := Chart{
		Ri: zhuInfo{Zhu: ganzhi.Zhu{Gan: ganzhi.GanJia, Zhi: ganzhi.ZhiWu}},
		Yue: zhuInfo{
			Zhu: ganzhi.Zhu{Gan: ganzhi.GanGui, Zhi: ganzhi.ZhiYou},
		},
		Nian: zhuInfo{Zhu: ganzhi.Zhu{Gan: ganzhi.GanXin, Zhi: ganzhi.ZhiSi}}, // 年干辛透 → 正官格
		Shi:  zhuInfo{Zhu: ganzhi.Zhu{Gan: ganzhi.GanRen, Zhi: ganzhi.ZhiShen}},
	}
	result := computeGeJu(cb, nil)
	if result.Pattern != "正官格" {
		t.Errorf("Pattern = %q, want 正官格 (酉辛透年干应优先于月干癸水)", result.Pattern)
	}
	if result.Usage != "顺用" {
		t.Errorf("Usage = %q, want 顺用", result.Usage)
	}
}

// TestGeJu_DefaultPattern tests the fallback when the 透干 gan is 劫财/比肩
// (not in the standard 八格).
func TestGeJu_DefaultPattern(t *testing.T) {
	// 甲日主, 辰月(戊乙癸), 月干乙木(劫财), 乙在中气.
	// 辰非甲禄/刃 → 支藏干检查. 戊不透, 乙透(月干=乙) → 劫财 → 杂格.
	cb := Chart{
		Ri: zhuInfo{Zhu: ganzhi.Zhu{Gan: ganzhi.GanJia, Zhi: ganzhi.ZhiWu}},
		Yue: zhuInfo{
			Zhu: ganzhi.Zhu{Gan: ganzhi.GanYi, Zhi: ganzhi.ZhiChen},
		},
		Nian: zhuInfo{Zhu: ganzhi.Zhu{Gan: ganzhi.GanGui, Zhi: ganzhi.ZhiSi}},
		Shi:  zhuInfo{Zhu: ganzhi.Zhu{Gan: ganzhi.GanRen, Zhi: ganzhi.ZhiShen}},
	}
	result := computeGeJu(cb, nil)
	if result.Pattern != "杂格" {
		t.Errorf("Pattern = %q, want 杂格 (劫财不入八格)", result.Pattern)
	}
	if result.Usage != "逆用" {
		t.Errorf("Usage = %q, want 逆用", result.Usage)
	}
}

// TestGeJu_RealCharts tests with real birth charts via ComputeChart.
func TestGeJu_RealCharts(t *testing.T) {
	tests := []struct {
		name        string
		birthTime   time.Time
		gender      ganzhi.Gender
		wantPattern string
		wantUsage   string
		note        string
	}{
		{
			// 甲子 丙寅 己卯 戊辰
			// 己日主, 寅月(甲木正官本气), 年干甲木透 → 正官格
			name:        "己日寅月甲透年干→正官格",
			birthTime:   time.Date(1984, 2, 15, 8, 0, 0, 0, time.FixedZone("CST", 8*3600)),
			gender:      ganzhi.Male,
			wantPattern: "正官格",
			wantUsage:   "顺用",
			note:        "年干甲=寅本气透→正官",
		},
		{
			// 甲子 癸酉 丁未 丙午
			// 丁日主, 酉月(辛金偏财本气), 四干无金 → 虚格偏财
			name:        "丁日酉月辛不透→虚格偏财",
			birthTime:   time.Date(1984, 9, 10, 12, 0, 0, 0, time.FixedZone("CST", 8*3600)),
			gender:      ganzhi.Male,
			wantPattern: "偏财格",
			wantUsage:   "顺用",
			note:        "四干无金→虚格偏财",
		},
		{
			// 壬戌 壬子 壬申 丙午
			// 壬日主, 子月 → 子=壬刃 → 月刃格
			name:        "壬日子月→月刃格",
			birthTime:   time.Date(1982, 12, 15, 12, 0, 0, 0, time.FixedZone("CST", 8*3600)),
			gender:      ganzhi.Male,
			wantPattern: "月刃格",
			wantUsage:   "逆用",
			note:        "月刃不变",
		},
		{
			// 乙丑 己卯 戊午 戊午
			// 戊日主, 卯月(乙木正官本气), 年干乙木透 → 正官格
			name:        "戊日卯月乙透年干→正官格",
			birthTime:   time.Date(1985, 3, 20, 12, 0, 0, 0, time.FixedZone("CST", 8*3600)),
			gender:      ganzhi.Male,
			wantPattern: "正官格",
			wantUsage:   "顺用",
			note:        "年干乙=卯本气透→正官",
		},
		{
			// 戊辰 戊午 辛丑 甲午
			// 辛日主, 午月(丁火七杀本气), 四干无丁 → 虚格七杀
			name:        "辛日午月丁不透→虚格七杀",
			birthTime:   time.Date(1988, 6, 15, 12, 0, 0, 0, time.FixedZone("CST", 8*3600)),
			gender:      ganzhi.Male,
			wantPattern: "七杀格",
			wantUsage:   "逆用",
			note:        "午丁不透→虚格七杀",
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			st := tianwen.GregorianToSolar(tt.birthTime, 116.4, 8)
			chart := ComputeChart(st, tt.gender)
			_, _, geJu := ComputeYongShenSchools(chart)
			if geJu.Pattern != tt.wantPattern {
				t.Errorf("Pattern = %q, want %q\n  note: %s", geJu.Pattern, tt.wantPattern, tt.note)
			}
			if geJu.Usage != tt.wantUsage {
				t.Errorf("Usage = %q, want %q\n  note: %s", geJu.Usage, tt.wantUsage, tt.note)
			}
		})
	}
}

// TestGeJu_StructuralContract verifies that every candidate has an exact
// pattern god and source without emitting school-level yong/xi/ji.
func TestGeJu_StructuralContract(t *testing.T) {
	allTests := []struct {
		riGan  ganzhi.Gan
		yueGan ganzhi.Gan
		yueZhi ganzhi.Zhi
	}{
		{ganzhi.GanJia, ganzhi.GanXin, ganzhi.ZhiYou},
		{ganzhi.GanJia, ganzhi.GanJi, ganzhi.ZhiWei},
		{ganzhi.GanJia, ganzhi.GanWu, ganzhi.ZhiChen},
		{ganzhi.GanJia, ganzhi.GanGui, ganzhi.ZhiZi},
		{ganzhi.GanJia, ganzhi.GanRen, ganzhi.ZhiHai},
		{ganzhi.GanJia, ganzhi.GanBing, ganzhi.ZhiSi},
		{ganzhi.GanJia, ganzhi.GanGeng, ganzhi.ZhiShen},
		{ganzhi.GanJia, ganzhi.GanDing, ganzhi.ZhiWu},
	}
	for _, tt := range allTests {
		cb := mkGeJuChart(tt.riGan, tt.yueGan, ganzhi.GanGui, ganzhi.GanRen, tt.yueZhi)
		result := computeGeJu(cb, nil)
		if result.Pattern == "" || result.PatternGod == "" || result.PatternGodSource == "" {
			t.Fatalf("%s: pattern contract incomplete: %+v", result.Pattern, result)
		}
		if result.Usage != "顺用" && result.Usage != "逆用" {
			t.Fatalf("%s: invalid usage %q", result.Pattern, result.Usage)
		}
		if result.Structure.Pattern.Wuxing == "" || len(result.Structure.Pattern.Roots) == 0 {
			t.Fatalf("%s: structure roots incomplete: %+v", result.Pattern, result.Structure.Pattern)
		}
	}
	jianLuTests := []struct {
		riGan  ganzhi.Gan
		yueZhi ganzhi.Zhi
	}{
		{ganzhi.GanJia, ganzhi.ZhiYin}, {ganzhi.GanJia, ganzhi.ZhiMao},
		{ganzhi.GanYi, ganzhi.ZhiMao}, {ganzhi.GanBing, ganzhi.ZhiSi},
		{ganzhi.GanBing, ganzhi.ZhiWu}, {ganzhi.GanWu, ganzhi.ZhiWu},
		{ganzhi.GanGeng, ganzhi.ZhiYou}, {ganzhi.GanRen, ganzhi.ZhiZi},
	}
	for _, tt := range jianLuTests {
		isLu, isRen := jianLuYueRenZhi(tt.riGan, tt.yueZhi)
		if !isLu && !isRen {
			t.Fatalf("%s%s: expected lu or yang blade", ganzhi.GanName(tt.riGan), ganzhi.ZhiName(tt.yueZhi))
		}
		cb := mkGeJuChart(tt.riGan, ganzhi.GanBing, ganzhi.GanGui, ganzhi.GanRen, tt.yueZhi)
		result := computeGeJu(cb, nil)
		if result.PatternGodSource != "month_lu" && result.PatternGodSource != "month_blade" {
			t.Fatalf("%s%s: source=%q", ganzhi.GanName(tt.riGan), ganzhi.ZhiName(tt.yueZhi), result.PatternGodSource)
		}
	}
}
