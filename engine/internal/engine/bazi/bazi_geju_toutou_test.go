package bazi

import (
	"testing"
	"time"

	"liki-engine/internal/engine/ganzhi"
	"liki-engine/internal/engine/tianwen"
)

// chartFromGanZhi builds a full Chart from explicit 4 pillars + dummy time.
// DaYun data is meaningless but geju calculation doesn't use it.
func chartFromGanZhi(nianGan, nianZhi, yueGan, yueZhi, riGan, riZhi, shiGan, shiZhi int, g ganzhi.Gender) Chart {
	bz := ganzhi.Bazi{
		Nian: ganzhi.Zhu{Gan: ganzhi.Gan(nianGan), Zhi: ganzhi.Zhi(nianZhi)},
		Yue:  ganzhi.Zhu{Gan: ganzhi.Gan(yueGan), Zhi: ganzhi.Zhi(yueZhi)},
		Ri:   ganzhi.Zhu{Gan: ganzhi.Gan(riGan), Zhi: ganzhi.Zhi(riZhi)},
		Shi:  ganzhi.Zhu{Gan: ganzhi.Gan(shiGan), Zhi: ganzhi.Zhi(shiZhi)},
	}
	st := tianwen.SolarTime(time.Date(2000, 1, 1, 12, 0, 0, 0, time.UTC))
	return computeChartCore(bz, st, g)
}

// ── 透干法格局测试 ──────────────────────────────────────────────

// 子平正法：以月令支藏干透干定格局。
// 顺序：本气→中气→余气，第一个透干者定格；都不透则月令本气虚格。
// 建禄格/月刃格优先级高于透干检查。

type toutouTestCase struct {
	name                                                           string
	nianGan, nianZhi, yueGan, yueZhi, riGan, riZhi, shiGan, shiZhi int
	gender                                                         ganzhi.Gender
	wantPattern                                                    string
	wantUsage                                                      string
	wantYong                                                       string
	wantXi                                                         string
	wantJi                                                         string
	note                                                           string
}

func TestGeJu_TouTou_AllPatterns(t *testing.T) {
	// All tests use 甲木日主, months that are NOT 甲禄(寅)刃(卯).
	tests := []toutouTestCase{
		// ── 正官格 — 酉月辛金本气透月干 ──
		// 四干: 癸/辛/甲/壬 → 辛透(月干), 辛克甲=正官
		{
			name:    "正官格_酉月辛透月干",
			nianGan: 10, nianZhi: 6, // 癸巳
			yueGan: 8, yueZhi: 10, // 辛酉
			riGan: 1, riZhi: 1, // 甲子
			shiGan: 9, shiZhi: 9, // 壬申
			gender:      ganzhi.Male,
			wantPattern: "正官格",
			wantUsage:   "顺用",
			// pattern=金, yong=生金=土, ji=克金=火, xi=制火=水
			wantYong: "土",
			wantXi:   "水",
			wantJi:   "火",
			note:     "酉本气辛透月干→正官。xi=水(印制伤护官), 不等于ji=火",
		},
		// ── 七杀格 — 申月庚金本气透月干 ──
		// 四干: 癸/庚/甲/壬 → 庚透(月干), 庚克甲=七杀
		{
			name:    "七杀格_申月庚透月干",
			nianGan: 10, nianZhi: 6, // 癸巳
			yueGan: 7, yueZhi: 9, // 庚申
			riGan: 1, riZhi: 1, // 甲子
			shiGan: 9, shiZhi: 9, // 壬申
			gender:      ganzhi.Male,
			wantPattern: "七杀格",
			wantUsage:   "逆用",
			// pattern=金, yong=克金=火, xi=生火=木, ji=生金=土
			wantYong: "火",
			wantXi:   "木",
			wantJi:   "土",
			note:     "申本气庚透月干→七杀。逆用逻辑不变",
		},
		// ── 正印格 — 子月癸水本气透月干 ──
		// 四干: 癸/癸/甲/壬 → 癸透(月干+年干), 癸生甲=正印
		{
			name:    "正印格_子月癸透月干",
			nianGan: 10, nianZhi: 6, // 癸巳
			yueGan: 10, yueZhi: 1, // 癸子
			riGan: 1, riZhi: 1, // 甲子
			shiGan: 9, shiZhi: 9, // 壬申
			gender:      ganzhi.Male,
			wantPattern: "正印格",
			wantUsage:   "顺用",
			// pattern=水, yong=生水=金, ji=克水=土, xi=制土=木
			wantYong: "金",
			wantXi:   "木",
			wantJi:   "土",
			note:     "子本气癸透月干→正印。xi=木(比劫抗财护印), 不等于ji=土",
		},
		// ── 偏印格 — 亥月壬水本气透年干 ──
		// 四干: 壬/丙/甲/壬 → 壬透(年干+时干), 壬生甲=偏印
		{
			name:    "偏印格_亥月壬透年干",
			nianGan: 9, nianZhi: 6, // 壬巳
			yueGan: 3, yueZhi: 12, // 丙亥
			riGan: 1, riZhi: 1, // 甲子
			shiGan: 9, shiZhi: 9, // 壬申
			gender:      ganzhi.Male,
			wantPattern: "偏印格",
			wantUsage:   "顺用",
			// pattern=水, yong=生水=金, ji=克水=土, xi=制土=木
			wantYong: "金",
			wantXi:   "木",
			wantJi:   "土",
			note:     "亥本气壬透年干(壬生甲=偏印)→偏印格。月干丙不入格",
		},
		// ── 食神格 — 巳月丙火本气透月干 ──
		// 四干: 癸/丙/甲/壬 → 丙透(月干), 丙生甲=食神
		{
			name:    "食神格_巳月丙透月干",
			nianGan: 10, nianZhi: 6, // 癸巳
			yueGan: 3, yueZhi: 6, // 丙巳
			riGan: 1, riZhi: 1, // 甲子
			shiGan: 9, shiZhi: 9, // 壬申
			gender:      ganzhi.Male,
			wantPattern: "食神格",
			wantUsage:   "顺用",
			// pattern=火, yong=生火=木, ji=克火=水, xi=制水=土
			wantYong: "木",
			wantXi:   "土",
			wantJi:   "水",
			note:     "巳本气丙透月干→食神。xi=土(财泄食制印), 不等于ji=水",
		},
		// ── 伤官格 — 午月丁火本气透月干 ──
		// 四干: 癸/丁/甲/壬 → 丁透(月干), 丁生甲=伤官
		{
			name:    "伤官格_午月丁透月干",
			nianGan: 10, nianZhi: 6, // 癸巳
			yueGan: 4, yueZhi: 7, // 丁午
			riGan: 1, riZhi: 1, // 甲子
			shiGan: 9, shiZhi: 9, // 壬申
			gender:      ganzhi.Male,
			wantPattern: "伤官格",
			wantUsage:   "逆用",
			// pattern=火, yong=克火=水, xi=生水=金, ji=生火=木
			wantYong: "水",
			wantXi:   "金",
			wantJi:   "木",
			note:     "午本气丁透月干→伤官。逆用逻辑不变",
		},
		// ── 正财格 — 未月己土本气透月干 ──
		// 四干: 癸/己/甲/壬 → 己透(月干), 甲克己=正财
		{
			name:    "正财格_未月己透月干",
			nianGan: 10, nianZhi: 6, // 癸巳
			yueGan: 6, yueZhi: 8, // 己未
			riGan: 1, riZhi: 1, // 甲子
			shiGan: 9, shiZhi: 9, // 壬申
			gender:      ganzhi.Male,
			wantPattern: "正财格",
			wantUsage:   "顺用",
			// pattern=土, yong=生土=火, ji=克土=木, xi=制木=金
			wantYong: "火",
			wantXi:   "金",
			wantJi:   "木",
			note:     "未本气己透月干→正财。xi=金(官杀制比劫护财), 不等于ji=木",
		},
		// ── 偏财格 — 辰月戊土本气透月干 ──
		// 四干: 癸/戊/甲/壬 → 戊透(月干), 甲克戊=偏财
		{
			name:    "偏财格_辰月戊透月干",
			nianGan: 10, nianZhi: 6, // 癸巳
			yueGan: 5, yueZhi: 5, // 戊辰
			riGan: 1, riZhi: 1, // 甲子
			shiGan: 9, shiZhi: 9, // 壬申
			gender:      ganzhi.Male,
			wantPattern: "偏财格",
			wantUsage:   "顺用",
			// pattern=土, yong=生土=火, ji=克土=木, xi=制木=金
			wantYong: "火",
			wantXi:   "金",
			wantJi:   "木",
			note:     "辰本气戊透月干→偏财。公式同正财",
		},
		// ── 中气透干 — 辰月乙木中气透时干 ──
		// 辰: 戊(本气) 乙(中气) 癸(余气)
		// 四干: 癸/甲/甲/乙 → 戊不透, 乙透(时干), 乙=劫财
		{
			name:    "杂格_辰月中气乙透时干",
			nianGan: 10, nianZhi: 6, // 癸巳
			yueGan: 1, yueZhi: 5, // 甲辰
			riGan: 1, riZhi: 1, // 甲子
			shiGan: 2, shiZhi: 9, // 乙申
			gender:      ganzhi.Male,
			wantPattern: "杂格",
			wantUsage:   "逆用",
			// 劫财不在顺用/逆用中→default 杂格
			// yong=克日主=金, xi=生金=土, ji=生日主=水
			wantYong: "金",
			wantXi:   "土",
			wantJi:   "水",
			note:     "辰中气乙透时干(劫财)→杂格(比劫不入八格)",
		},
		// ── 余气透干 — 辰月癸水余气透时干 ──
		// 辰: 戊(本气) 乙(中气) 癸(余气)
		// 四干: 甲/甲/甲/癸 → 戊不透, 乙不透, 癸透(时干), 癸生甲=正印
		{
			name:    "正印格_辰月余气癸透时干",
			nianGan: 1, nianZhi: 6, // 甲巳
			yueGan: 1, yueZhi: 5, // 甲辰
			riGan: 1, riZhi: 1, // 甲子
			shiGan: 10, shiZhi: 9, // 癸申
			gender:      ganzhi.Male,
			wantPattern: "正印格",
			wantUsage:   "顺用",
			// pattern=水, yong=金, ji=土, xi=木
			wantYong: "金",
			wantXi:   "木",
			wantJi:   "土",
			note:     "辰戊(本气)乙(中气)均不透, 癸(余气)透时干→正印格",
		},
		// ── 虚格 — 酉月四干无金 ──
		// 酉月辛金本气, 四干: 丙/丙/甲/丙 → 辛不透 → 虚格正官
		{
			name:    "虚格正官_酉月四干无金",
			nianGan: 3, nianZhi: 6, // 丙巳
			yueGan: 3, yueZhi: 10, // 丙酉
			riGan: 1, riZhi: 1, // 甲子
			shiGan: 3, shiZhi: 9, // 丙申
			gender:      ganzhi.Male,
			wantPattern: "正官格",
			wantUsage:   "顺用",
			// 虚格: 月令本气辛金=正官, 规用
			wantYong: "土",
			wantXi:   "水",
			wantJi:   "火",
			note:     "酉辛不透干→虚格正官。同正官格公式",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			c := chartFromGanZhi(
				tt.nianGan, tt.nianZhi,
				tt.yueGan, tt.yueZhi,
				tt.riGan, tt.riZhi,
				tt.shiGan, tt.shiZhi,
				tt.gender,
			)
			result := ComputeYongShen(c)

			got := result.GeJu
			if got.Pattern != tt.wantPattern {
				t.Errorf("Pattern = %q, want %q\n  note: %s", got.Pattern, tt.wantPattern, tt.note)
			}
			if got.Usage != tt.wantUsage {
				t.Errorf("Usage = %q, want %q\n  note: %s", got.Usage, tt.wantUsage, tt.note)
			}
			if got.Yong != tt.wantYong {
				t.Errorf("Yong = %q, want %q\n  note: %s", got.Yong, tt.wantYong, tt.note)
			}
			if got.Xi != tt.wantXi {
				t.Errorf("Xi = %q, want %q\n  note: %s", got.Xi, tt.wantXi, tt.note)
			}
			if got.Ji != tt.wantJi {
				t.Errorf("Ji = %q, want %q\n  note: %s", got.Ji, tt.wantJi, tt.note)
			}
		})
	}
}
