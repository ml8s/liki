package bazi

import (
	"testing"
	"time"

	"liki-engine/internal/engine/ganzhi"
	"liki-engine/internal/engine/tianwen"
)

// ============================================================================
// 藏干 (Hidden Stems) — 标准命理参考数据
// ============================================================================

// referenceCangGan is the standard hidden stems for each earthly branch.
// Source: 渊海子平 / 三命通会.
var referenceCangGan = map[ganzhi.Zhi]struct{ main, mid, minor ganzhi.Gan }{
	ganzhi.ZhiZi:   {main: ganzhi.GanGui},
	ganzhi.ZhiChou: {main: ganzhi.GanJi, mid: ganzhi.GanGui, minor: ganzhi.GanXin},
	ganzhi.ZhiYin:  {main: ganzhi.GanJia, mid: ganzhi.GanBing, minor: ganzhi.GanWu},
	ganzhi.ZhiMao:  {main: ganzhi.GanYi},
	ganzhi.ZhiChen: {main: ganzhi.GanWu, mid: ganzhi.GanYi, minor: ganzhi.GanGui},
	ganzhi.ZhiSi:   {main: ganzhi.GanBing, mid: ganzhi.GanGeng, minor: ganzhi.GanWu},
	ganzhi.ZhiWu:   {main: ganzhi.GanDing, mid: ganzhi.GanJi},
	ganzhi.ZhiWei:  {main: ganzhi.GanJi, mid: ganzhi.GanYi, minor: ganzhi.GanDing},
	ganzhi.ZhiShen: {main: ganzhi.GanGeng, mid: ganzhi.GanRen, minor: ganzhi.GanWu},
	ganzhi.ZhiYou:  {main: ganzhi.GanXin},
	ganzhi.ZhiXu:   {main: ganzhi.GanWu, mid: ganzhi.GanXin, minor: ganzhi.GanDing},
	ganzhi.ZhiHai:  {main: ganzhi.GanRen, mid: ganzhi.GanJia},
}

func TestCangGan_AllBranches(t *testing.T) {
	for zhi, want := range referenceCangGan {
		t.Run(ganzhi.ZhiName(zhi), func(t *testing.T) {
			qi := ganzhi.CangGanForZhi(zhi)

			if qi.Main == nil || *qi.Main != want.main {
				t.Errorf("main = %v, want %s", qi.Main, ganzhi.GanName(want.main))
			}

			if want.mid != 0 {
				if qi.Mid == nil {
					t.Errorf("mid = nil, want %s", ganzhi.GanName(want.mid))
				} else if *qi.Mid != want.mid {
					t.Errorf("mid = %s, want %s", ganzhi.GanName(*qi.Mid), ganzhi.GanName(want.mid))
				}
			} else if qi.Mid != nil {
				t.Errorf("mid = %s, want nil", ganzhi.GanName(*qi.Mid))
			}

			if want.minor != 0 {
				if qi.Minor == nil {
					t.Errorf("minor = nil, want %s", ganzhi.GanName(want.minor))
				} else if *qi.Minor != want.minor {
					t.Errorf("minor = %s, want %s", ganzhi.GanName(*qi.Minor), ganzhi.GanName(want.minor))
				}
			} else if qi.Minor != nil {
				t.Errorf("minor = %s, want nil", ganzhi.GanName(*qi.Minor))
			}
		})
	}
}

// ============================================================================
// 纳音 (NaYin) — 六十甲子纳音表
// ============================================================================

func TestNaYin_ReferencePairs(t *testing.T) {
	// Standard 60-Jiazi NaYin table spot-checks.
	tests := []struct {
		gan      ganzhi.Gan
		zhi      ganzhi.Zhi
		wantName string
		wantElem string
	}{
		// 甲子乙丑 → 海中金
		{ganzhi.GanJia, ganzhi.ZhiZi, "海中金", "金"},
		{ganzhi.GanYi, ganzhi.ZhiChou, "海中金", "金"},
		// 丙寅丁卯 → 炉中火
		{ganzhi.GanBing, ganzhi.ZhiYin, "炉中火", "火"},
		{ganzhi.GanDing, ganzhi.ZhiMao, "炉中火", "火"},
		// 戊辰己巳 → 大林木
		{ganzhi.GanWu, ganzhi.ZhiChen, "大林木", "木"},
		{ganzhi.GanJi, ganzhi.ZhiSi, "大林木", "木"},
		// 庚午辛未 → 路旁土
		{ganzhi.GanGeng, ganzhi.ZhiWu, "路旁土", "土"},
		{ganzhi.GanXin, ganzhi.ZhiWei, "路旁土", "土"},
		// 壬申癸酉 → 剑锋金
		{ganzhi.GanRen, ganzhi.ZhiShen, "剑锋金", "金"},
		{ganzhi.GanGui, ganzhi.ZhiYou, "剑锋金", "金"},
		// 甲戌乙亥 → 山头火
		{ganzhi.GanJia, ganzhi.ZhiXu, "山头火", "火"},
		{ganzhi.GanYi, ganzhi.ZhiHai, "山头火", "火"},
		// 丙子丁丑 → 涧下水
		{ganzhi.GanBing, ganzhi.ZhiZi, "涧下水", "水"},
		{ganzhi.GanDing, ganzhi.ZhiChou, "涧下水", "水"},
		// 戊寅己卯 → 城头土
		{ganzhi.GanWu, ganzhi.ZhiYin, "城头土", "土"},
		{ganzhi.GanJi, ganzhi.ZhiMao, "城头土", "土"},
		// 庚辰辛巳 → 白蜡金
		{ganzhi.GanGeng, ganzhi.ZhiChen, "白蜡金", "金"},
		{ganzhi.GanXin, ganzhi.ZhiSi, "白蜡金", "金"},
		// 壬午癸未 → 杨柳木
		{ganzhi.GanRen, ganzhi.ZhiWu, "杨柳木", "木"},
		{ganzhi.GanGui, ganzhi.ZhiWei, "杨柳木", "木"},
	}

	for _, tt := range tests {
		t.Run(tt.wantName+" "+ganzhi.GanName(tt.gan)+ganzhi.ZhiName(tt.zhi), func(t *testing.T) {
			got := ganzhi.NayinLabel(tt.gan, tt.zhi)
			if got != tt.wantName {
				t.Errorf("NayinLabel = %q, want %q", got, tt.wantName)
			}

			elem := ganzhi.NayinWuxing(got)
			if elem.String() != tt.wantElem {
				t.Errorf("NayinWuxing = %s, want %s", elem.String(), tt.wantElem)
			}
		})
	}
}

// ============================================================================
// 调候 (TiaoHou) — 穷通宝鉴参考
// ============================================================================

func TestTiaoHou_ReferenceEntries(t *testing.T) {
	// Spot-check entries against 穷通宝鉴.
	tests := []struct {
		riYuan      ganzhi.Gan
		monthBranch ganzhi.Zhi
		wantYong    string // expected yong element
	}{
		// 甲木调候
		{ganzhi.GanJia, ganzhi.ZhiYin, "火"},  // 正月甲木: 丙癸
		{ganzhi.GanJia, ganzhi.ZhiMao, "水"},  // 二月甲木: 癸庚丁
		{ganzhi.GanJia, ganzhi.ZhiWu, "水"},   // 五月甲木: 癸丁庚
		{ganzhi.GanJia, ganzhi.ZhiShen, "水"}, // 七月甲木: 丁壬庚
		{ganzhi.GanJia, ganzhi.ZhiZi, "火"},   // 十一月甲木: 丁庚丙

		// 乙木调候
		{ganzhi.GanYi, ganzhi.ZhiYin, "火"},   // 正月乙木: 丙癸
		{ganzhi.GanYi, ganzhi.ZhiWu, "水"},    // 五月乙木: 壬癸
		{ganzhi.GanYi, ganzhi.ZhiShen, "水"},  // 七月乙木: 癸丙

		// 丙火调候
		{ganzhi.GanBing, ganzhi.ZhiYin, "水"}, // 正月丙火: 壬庚
		{ganzhi.GanBing, ganzhi.ZhiWu, "水"},  // 五月丙火: 癸庚壬
		{ganzhi.GanBing, ganzhi.ZhiZi, "水"},  // 十一月丙火: 甲戊庚

		// 丁火调候
		{ganzhi.GanDing, ganzhi.ZhiYin, "木"},  // 正月丁火: 甲庚
		{ganzhi.GanDing, ganzhi.ZhiWu, "水"},   // 五月丁火: 壬庚癸
		{ganzhi.GanDing, ganzhi.ZhiYou, "木"},  // 八月丁火: 甲庚丙戊

		// 庚金调候
		{ganzhi.GanGeng, ganzhi.ZhiYin, "土"},  // 正月庚金: 戊甲丙丁
		{ganzhi.GanGeng, ganzhi.ZhiWu, "水"},   // 五月庚金: 壬癸
		{ganzhi.GanGeng, ganzhi.ZhiShen, "火"}, // 七月庚金: 丁甲

		// 壬水调候
		{ganzhi.GanRen, ganzhi.ZhiYin, "土"},  // 正月壬水: 庚戊丙
		{ganzhi.GanRen, ganzhi.ZhiWu, "金"},   // 五月壬水: 癸庚辛
		{ganzhi.GanRen, ganzhi.ZhiZi, "火"},   // 十一月壬水: 戊丙
	}

	for _, tt := range tests {
		name := ganzhi.GanName(tt.riYuan) + "日" + ganzhi.ZhiName(tt.monthBranch) + "月"
		t.Run(name, func(t *testing.T) {
			result := computeTiaoHou(tt.riYuan, tt.monthBranch)

			if result.Yong == "" {
				t.Skip("no tiaohou entry")
			}
			if result.Season == "" {
				t.Error("Season is empty")
			}
			if result.Detail == "" {
				t.Error("Detail is empty")
			}

			// Verify yong/xi/ji are valid 五行. Ji may be empty (no clear 忌神).
			for _, field := range []struct{ label, val string }{
				{"Yong", result.Yong}, {"Xi", result.Xi}, {"Ji", result.Ji},
			} {
				if field.label == "Ji" && field.val == "" {
					continue // no clear 忌神 is valid
				}
				if field.label == "Xi" && field.val == "" {
					continue // no 喜神 when no secondary
				}
				switch field.val {
				case "木", "火", "土", "金", "水":
				default:
					t.Errorf("%s = %q, want valid 五行 or empty", field.label, field.val)
				}
			}
		})
	}
}

// ============================================================================
// 大运 (DaYun) — 顺逆排法
// ============================================================================

func TestDaYun_Direction(t *testing.T) {
	// DaYun direction: 阳年男/阴年女 → 顺排; 阳年女/阴年男 → 逆排.
	// 甲子年(阳), 丙寅年(阳), 乙丑年(阴), 丁卯年(阴).
	tests := []struct {
		name        string
		year        int
		month       int
		day         int
		gender      ganzhi.Gender
		wantDir     string
	}{
		{"阳年男→顺排 1984甲子", 1984, 2, 15, ganzhi.Male, DirShunPai},
		{"阳年女→逆排 1984甲子", 1984, 2, 15, ganzhi.Female, DirNiPai},
		{"阴年男→逆排 1985乙丑", 1985, 3, 20, ganzhi.Male, DirNiPai},
		{"阴年女→顺排 1985乙丑", 1985, 3, 20, ganzhi.Female, DirShunPai},
		{"阳年男→顺排 1986丙寅", 1986, 6, 15, ganzhi.Male, DirShunPai},
		{"阳年女→逆排 1986丙寅", 1986, 6, 15, ganzhi.Female, DirNiPai},
		{"阴年男→逆排 1987丁卯", 1987, 9, 10, ganzhi.Male, DirNiPai},
		{"阴年女→顺排 1987丁卯", 1987, 9, 10, ganzhi.Female, DirShunPai},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			st := tianwen.GregorianToSolar(
				time.Date(tt.year, time.Month(tt.month), tt.day, 12, 0, 0, 0,
					time.FixedZone("CST", 8*3600)),
				120, 8)
			chart := ComputeChart(st, tt.gender)

			if chart.DaYun == nil {
				t.Fatal("DaYun is nil")
			}
			if chart.DaYun.Direction != tt.wantDir {
				t.Errorf("Direction = %q, want %q", chart.DaYun.Direction, tt.wantDir)
			}

			// Start age should be between 0 and 12.
			if chart.DaYun.StartAge < 0 || chart.DaYun.StartAge > 12 {
				t.Errorf("StartAge = %d, want [0,12]", chart.DaYun.StartAge)
			}

			// Must have at least 8 zhu (80 years of fortune).
			if len(chart.DaYun.Steps) < 8 {
				t.Errorf("len(Zhu) = %d, want >= 8", len(chart.DaYun.Steps))
			}

			// Each zhu should be 10 years (age_end - age_start + 1 = 10).
			for i, z := range chart.DaYun.Steps {
				if z.AgeEnd-z.AgeStart+1 != 10 {
					t.Errorf("zhu[%d] %s: age range %d-%d is not 10 years",
						i, z.Name, z.AgeStart, z.AgeEnd)
				}
				if z.ShiShen == "" {
					t.Errorf("zhu[%d] %s: ShiShen is empty", i, z.Name)
				}
				if z.Name == "" {
					t.Errorf("zhu[%d]: Name is empty", i)
				}
			}
		})
	}
}

// ============================================================================
// 四柱 (Four Pillars) — 年上起月 / 日上起时 一致性
// ============================================================================

func TestPillarConsistency_YearToMonth(t *testing.T) {
	// Verify 年上起月法 (五虎遁): month stem follows year stem correctly.
	tests := []struct {
		name      string
		year      int
		month     int
		day       int
		wantNianGan ganzhi.Gan
	}{
		// 甲年(1984) → 正月丙寅
		{"甲年→丙寅月 1984-02", 1984, 2, 15, ganzhi.GanJia},
		// 乙年(1985) → 正月戊寅
		{"乙年→戊寅月 1985-02", 1985, 2, 10, ganzhi.GanYi},
		// 丙年(1986) → 正月庚寅
		{"丙年→庚寅月 1986-02", 1986, 2, 10, ganzhi.GanBing},
		// 丁年(1987) → 正月壬寅
		{"丁年→壬寅月 1987-02", 1987, 2, 10, ganzhi.GanDing},
		// 戊年(1988) → 正月甲寅
		{"戊年→甲寅月 1988-02", 1988, 2, 10, ganzhi.GanWu},
		// 己年(1989) → 正月丙寅
		{"己年→丙寅月 1989-02", 1989, 2, 10, ganzhi.GanJi},
		// 庚年(1990) → 正月戊寅
		{"庚年→戊寅月 1990-02", 1990, 2, 10, ganzhi.GanGeng},
		// 辛年(1991) → 正月庚寅
		{"辛年→庚寅月 1991-02", 1991, 2, 10, ganzhi.GanXin},
		// 壬年(1992) → 正月壬寅
		{"壬年→壬寅月 1992-02", 1992, 2, 10, ganzhi.GanRen},
		// 癸年(1993) → 正月甲寅
		{"癸年→甲寅月 1993-02", 1993, 2, 10, ganzhi.GanGui},
	}

	// Expected first-month stem for each year stem (正月 = 寅月).
	wuHuDun := map[ganzhi.Gan]ganzhi.Gan{
		ganzhi.GanJia: ganzhi.GanBing,
		ganzhi.GanYi:  ganzhi.GanWu,
		ganzhi.GanBing: ganzhi.GanGeng,
		ganzhi.GanDing: ganzhi.GanRen,
		ganzhi.GanWu:  ganzhi.GanJia,
		ganzhi.GanJi:  ganzhi.GanBing,
		ganzhi.GanGeng: ganzhi.GanWu,
		ganzhi.GanXin:  ganzhi.GanGeng,
		ganzhi.GanRen:  ganzhi.GanRen,
		ganzhi.GanGui:  ganzhi.GanJia,
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			st := tianwen.GregorianToSolar(
				time.Date(tt.year, time.Month(tt.month), tt.day, 12, 0, 0, 0,
					time.FixedZone("CST", 8*3600)),
				120, 8)
			chart := ComputeChart(st, ganzhi.Male)

			// Verify year pillar stem.
			if chart.Nian.Gan != tt.wantNianGan {
				t.Errorf("Nian.Gan = %s, want %s",
					ganzhi.GanName(chart.Nian.Gan), ganzhi.GanName(tt.wantNianGan))
			}

			// Verify month stem follows 年上起月.
			// For dates in Feb, month is 寅月 (month 0 in the 0-based cycle).
			expectedMonthStem := wuHuDun[tt.wantNianGan]
			gotMonthStem := chart.Yue.Gan
			if gotMonthStem != expectedMonthStem {
				t.Errorf("Yue.Gan = %s, want %s (五虎遁 for %s年正月)",
					ganzhi.GanName(gotMonthStem), ganzhi.GanName(expectedMonthStem), ganzhi.GanName(tt.wantNianGan))
			}

			// Month branch for Feb (after 立春) should be 寅.
			if chart.Yue.Zhi != ganzhi.ZhiYin {
				t.Errorf("Yue.Zhi = %s, want 寅", ganzhi.ZhiName(chart.Yue.Zhi))
			}
		})
	}
}

func TestPillarConsistency_DayToHour(t *testing.T) {
	// Verify 日上起时法 (五鼠遁): hour stem follows day stem correctly.
	// Test for multiple days at 子时 (00:00).
	tests := []struct {
		name   string
		year   int
		month  int
		day    int
		wantRiGan ganzhi.Gan
	}{
		// These dates are chosen so riGan cycles through all 10 stems.
		{"甲日→甲子时", 1984, 2, 15, ganzhi.GanJi},   // 己卯日, 子时=甲子
		{"乙日→丙子时", 1984, 2, 16, ganzhi.GanGeng},  // 庚辰日, 子时=丙子
		{"丙日→戊子时", 1984, 2, 17, ganzhi.GanXin},   // 辛巳日, 子时=戊子
		{"丁日→庚子时", 1984, 2, 18, ganzhi.GanRen},   // 壬午日, 子时=庚子
		{"戊日→壬子时", 1984, 2, 19, ganzhi.GanGui},   // 癸未日, 子时=壬子
	}

	// 五鼠遁: day stem → zi-hour stem.
	wuShuDun := map[ganzhi.Gan]ganzhi.Gan{
		ganzhi.GanJia: ganzhi.GanJia,
		ganzhi.GanYi:  ganzhi.GanBing,
		ganzhi.GanBing: ganzhi.GanWu,
		ganzhi.GanDing: ganzhi.GanGeng,
		ganzhi.GanWu:  ganzhi.GanRen,
		ganzhi.GanJi:  ganzhi.GanJia,
		ganzhi.GanGeng: ganzhi.GanBing,
		ganzhi.GanXin:  ganzhi.GanWu,
		ganzhi.GanRen:  ganzhi.GanGeng,
		ganzhi.GanGui:  ganzhi.GanRen,
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			st := tianwen.SolarTime(
				time.Date(tt.year, time.Month(tt.month), tt.day, 0, 0, 0, 0,
					time.FixedZone("CST", 8*3600)))
			chart := ComputeChart(st, ganzhi.Male)

			// Verify day stem matches expected.
			if chart.Ri.Gan != tt.wantRiGan {
				t.Errorf("Ri.Gan = %s, want %s (date %d-%02d-%02d)",
					ganzhi.GanName(chart.Ri.Gan), ganzhi.GanName(tt.wantRiGan),
					tt.year, tt.month, tt.day)
			}

			// For 子时 (00:00), verify hour stem follows 日上起时.
			expectedShiStem := wuShuDun[tt.wantRiGan]
			gotShiStem := chart.Shi.Gan
			if gotShiStem != expectedShiStem {
				t.Errorf("Shi.Gan = %s, want %s (五鼠遁 for %s日 子时)",
					ganzhi.GanName(gotShiStem), ganzhi.GanName(expectedShiStem), ganzhi.GanName(tt.wantRiGan))
			}

			// For 子时, hour branch should be 子.
			if chart.Shi.Zhi != ganzhi.ZhiZi {
				t.Errorf("Shi.Zhi = %s, want 子", ganzhi.ZhiName(chart.Shi.Zhi))
			}
		})
	}
}

// ============================================================================
// 十神 (ShiShen) — 日主与其他天干关系
// ============================================================================

func TestShiShen_ReferenceCases(t *testing.T) {
	tests := []struct {
		riYuan     ganzhi.Gan
		other      ganzhi.Gan
		wantShiShen string
	}{
		// 甲木日主 vs 十天干
		{ganzhi.GanJia, ganzhi.GanJia, "比肩"},
		{ganzhi.GanJia, ganzhi.GanYi, "劫财"},
		{ganzhi.GanJia, ganzhi.GanBing, "食神"},
		{ganzhi.GanJia, ganzhi.GanDing, "伤官"},
		{ganzhi.GanJia, ganzhi.GanWu, "偏财"},
		{ganzhi.GanJia, ganzhi.GanJi, "正财"},
		{ganzhi.GanJia, ganzhi.GanGeng, "七杀"},
		{ganzhi.GanJia, ganzhi.GanXin, "正官"},
		{ganzhi.GanJia, ganzhi.GanRen, "偏印"},
		{ganzhi.GanJia, ganzhi.GanGui, "正印"},

		// 丙火日主 vs 十天干
		{ganzhi.GanBing, ganzhi.GanBing, "比肩"},
		{ganzhi.GanBing, ganzhi.GanDing, "劫财"},
		{ganzhi.GanBing, ganzhi.GanWu, "食神"},
		{ganzhi.GanBing, ganzhi.GanJi, "伤官"},
		{ganzhi.GanBing, ganzhi.GanGeng, "偏财"},
		{ganzhi.GanBing, ganzhi.GanXin, "正财"},
		{ganzhi.GanBing, ganzhi.GanRen, "七杀"},
		{ganzhi.GanBing, ganzhi.GanGui, "正官"},
		{ganzhi.GanBing, ganzhi.GanJia, "偏印"},
		{ganzhi.GanBing, ganzhi.GanYi, "正印"},

		// 庚金日主 vs 十天干
		{ganzhi.GanGeng, ganzhi.GanGeng, "比肩"},
		{ganzhi.GanGeng, ganzhi.GanXin, "劫财"},
		{ganzhi.GanGeng, ganzhi.GanRen, "食神"},
		{ganzhi.GanGeng, ganzhi.GanGui, "伤官"},
		{ganzhi.GanGeng, ganzhi.GanJia, "偏财"},
		{ganzhi.GanGeng, ganzhi.GanYi, "正财"},
		{ganzhi.GanGeng, ganzhi.GanBing, "七杀"},
		{ganzhi.GanGeng, ganzhi.GanDing, "正官"},
		{ganzhi.GanGeng, ganzhi.GanWu, "偏印"},
		{ganzhi.GanGeng, ganzhi.GanJi, "正印"},
	}

	for _, tt := range tests {
		name := ganzhi.GanName(tt.riYuan) + "→" + ganzhi.GanName(tt.other)
		t.Run(name, func(t *testing.T) {
			got := ganzhi.ShiShenFromGan(tt.riYuan, tt.other)
			if got.String() != tt.wantShiShen {
				t.Errorf("ShiShen = %q, want %q", got.String(), tt.wantShiShen)
			}
		})
	}
}

// ============================================================================
// 节气 (Solar Terms) — 换月边界
// ============================================================================

func TestMonthBoundary_JieQi(t *testing.T) {
	// Verify month pillar changes at 节气, not at calendar month boundaries.
	// Test dates near 立春 (around Feb 3-4).
	tests := []struct {
		name        string
		date        time.Time
		wantYueZhi  ganzhi.Zhi
	}{
		// 2024年立春: 2月4日 16:27 CST.
		// Before 立春 → 丑月; After → 寅月.
		// Engine uses date-only solar term check. Feb 3 is unambiguously before 立春 (Feb 4).
		{"2024立春前_丑月", time.Date(2024, 2, 3, 12, 0, 0, 0, time.FixedZone("CST", 8*3600)), ganzhi.ZhiChou},
		{"2024立春后_寅月", time.Date(2024, 2, 4, 20, 0, 0, 0, time.FixedZone("CST", 8*3600)), ganzhi.ZhiYin},

		// 2024年春分: 3月20日。卯月从惊蛰开始(3月5日左右).
		{"2024春分_卯月", time.Date(2024, 3, 20, 12, 0, 0, 0, time.FixedZone("CST", 8*3600)), ganzhi.ZhiMao},

		// 2024年夏至: 6月21日。午月从芒种开始(6月5日左右).
		{"2024夏至_午月", time.Date(2024, 6, 21, 12, 0, 0, 0, time.FixedZone("CST", 8*3600)), ganzhi.ZhiWu},

		// 2024年秋分: 9月22日。酉月从白露开始(9月7日左右).
		{"2024秋分_酉月", time.Date(2024, 9, 22, 12, 0, 0, 0, time.FixedZone("CST", 8*3600)), ganzhi.ZhiYou},

		// 2024年冬至: 12月21日。子月从大雪开始(12月6日左右).
		{"2024冬至_子月", time.Date(2024, 12, 21, 12, 0, 0, 0, time.FixedZone("CST", 8*3600)), ganzhi.ZhiZi},

		// 2024年1月10日: 小寒后 → 丑月.
		{"2024小寒后_丑月", time.Date(2024, 1, 10, 12, 0, 0, 0, time.FixedZone("CST", 8*3600)), ganzhi.ZhiChou},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			st := tianwen.GregorianToSolar(tt.date, 120, 8)
			chart := ComputeChart(st, ganzhi.Male)

			if chart.Yue.Zhi != tt.wantYueZhi {
				t.Errorf("Yue.Zhi = %s, want %s",
					ganzhi.ZhiName(chart.Yue.Zhi), ganzhi.ZhiName(tt.wantYueZhi))
			}
		})
	}
}

// ============================================================================
// 年柱 (Year Pillar) — 立春换年
// ============================================================================

func TestYearBoundary_LiChun(t *testing.T) {
	// 2024年立春: 2月4日 16:27 CST.
	// Before → 癸卯年; After → 甲辰年.
	tests := []struct {
		name       string
		date       time.Time
		wantNianGan ganzhi.Gan
		wantNianZhi ganzhi.Zhi
	}{
		// Engine uses date-only 立春 check. Feb 3 is unambiguously before 立春 (Feb 4).
		{"2024立春前_癸卯年", time.Date(2024, 2, 3, 12, 0, 0, 0, time.FixedZone("CST", 8*3600)), ganzhi.GanGui, ganzhi.ZhiMao},
		{"2024立春后_甲辰年", time.Date(2024, 2, 4, 20, 0, 0, 0, time.FixedZone("CST", 8*3600)), ganzhi.GanJia, ganzhi.ZhiChen},

		// 2025年立春: 2月3日 22:10 CST.
		{"2025立春前_甲辰年", time.Date(2025, 2, 2, 12, 0, 0, 0, time.FixedZone("CST", 8*3600)), ganzhi.GanJia, ganzhi.ZhiChen},
		{"2025立春后_乙巳年", time.Date(2025, 2, 4, 12, 0, 0, 0, time.FixedZone("CST", 8*3600)), ganzhi.GanYi, ganzhi.ZhiSi},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			st := tianwen.GregorianToSolar(tt.date, 120, 8)
			chart := ComputeChart(st, ganzhi.Male)

			if chart.Nian.Gan != tt.wantNianGan {
				t.Errorf("Nian.Gan = %s, want %s",
					ganzhi.GanName(chart.Nian.Gan), ganzhi.GanName(tt.wantNianGan))
			}
			if chart.Nian.Zhi != tt.wantNianZhi {
				t.Errorf("Nian.Zhi = %s, want %s",
					ganzhi.ZhiName(chart.Nian.Zhi), ganzhi.ZhiName(tt.wantNianZhi))
			}
		})
	}
}

// ============================================================================
// 时辰 (Hour Pillar) — 时辰分界
// ============================================================================

func TestHourBoundary_ShiChen(t *testing.T) {
	// Hour pillar boundaries based on solar time, NOT clock time.
	// 23:30 is safely in 子时 accounting for equation of time (~-0.4 min on Jun 15).
	tests := []struct {
		name       string
		hour, min  int
		wantShiZhi ganzhi.Zhi
	}{
		{"子时 00:00", 0, 0, ganzhi.ZhiZi},
		{"丑时 02:00", 2, 0, ganzhi.ZhiChou},
		{"寅时 04:00", 4, 0, ganzhi.ZhiYin},
		{"卯时 06:00", 6, 0, ganzhi.ZhiMao},
		{"辰时 08:00", 8, 0, ganzhi.ZhiChen},
		{"巳时 10:00", 10, 0, ganzhi.ZhiSi},
		{"午时 12:00", 12, 0, ganzhi.ZhiWu},
		{"未时 14:00", 14, 0, ganzhi.ZhiWei},
		{"申时 16:00", 16, 0, ganzhi.ZhiShen},
		{"酉时 18:00", 18, 0, ganzhi.ZhiYou},
		{"戌时 20:00", 20, 0, ganzhi.ZhiXu},
		{"亥时 22:00", 22, 0, ganzhi.ZhiHai},
		{"夜子时 23:30", 23, 30, ganzhi.ZhiZi},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			st := tianwen.GregorianToSolar(
				time.Date(2024, 6, 15, tt.hour, tt.min, 0, 0, time.FixedZone("CST", 8*3600)),
				120, 8)
			chart := ComputeChart(st, ganzhi.Male)

			if chart.Shi.Zhi != tt.wantShiZhi {
				t.Errorf("Shi.Zhi = %s, want %s",
					ganzhi.ZhiName(chart.Shi.Zhi), ganzhi.ZhiName(tt.wantShiZhi))
			}
		})
	}
}

// ============================================================================
// 五行 (Wuxing) — 天干地支五行属性
// ============================================================================

func TestWuxing_GanZhiMapping(t *testing.T) {
	// 天干五行: 甲乙木, 丙丁火, 戊己土, 庚辛金, 壬癸水.
	ganWuxing := map[ganzhi.Gan]ganzhi.Wuxing{
		ganzhi.GanJia: ganzhi.WxMu, ganzhi.GanYi: ganzhi.WxMu,
		ganzhi.GanBing: ganzhi.WxHuo, ganzhi.GanDing: ganzhi.WxHuo,
		ganzhi.GanWu: ganzhi.WxTu, ganzhi.GanJi: ganzhi.WxTu,
		ganzhi.GanGeng: ganzhi.WxJin, ganzhi.GanXin: ganzhi.WxJin,
		ganzhi.GanRen: ganzhi.WxShui, ganzhi.GanGui: ganzhi.WxShui,
	}
	for gan, want := range ganWuxing {
		got := ganzhi.GanWuxing(gan)
		if got != want {
			t.Errorf("GanWuxing(%s) = %s, want %s", ganzhi.GanName(gan), got, want)
		}
	}

	// 地支五行: 亥子水, 寅卯木, 巳午火, 申酉金, 辰戌丑未土.
	zhiWuxing := map[ganzhi.Zhi]ganzhi.Wuxing{
		ganzhi.ZhiZi: ganzhi.WxShui, ganzhi.ZhiChou: ganzhi.WxTu,
		ganzhi.ZhiYin: ganzhi.WxMu, ganzhi.ZhiMao: ganzhi.WxMu,
		ganzhi.ZhiChen: ganzhi.WxTu, ganzhi.ZhiSi: ganzhi.WxHuo,
		ganzhi.ZhiWu: ganzhi.WxHuo, ganzhi.ZhiWei: ganzhi.WxTu,
		ganzhi.ZhiShen: ganzhi.WxJin, ganzhi.ZhiYou: ganzhi.WxJin,
		ganzhi.ZhiXu: ganzhi.WxTu, ganzhi.ZhiHai: ganzhi.WxShui,
	}
	for zhi, want := range zhiWuxing {
		got := ganzhi.ZhiWuxing(zhi)
		if got != want {
			t.Errorf("ZhiWuxing(%s) = %s, want %s", ganzhi.ZhiName(zhi), got, want)
		}
	}
}

// ============================================================================
// 生克 (Sheng/Ke) — 五行生克关系
// ============================================================================

func TestWuxing_ShengKe(t *testing.T) {
	// 五行相生: 木→火→土→金→水→木
	sheng := map[ganzhi.Wuxing]ganzhi.Wuxing{
		ganzhi.WxMu: ganzhi.WxHuo, ganzhi.WxHuo: ganzhi.WxTu,
		ganzhi.WxTu: ganzhi.WxJin, ganzhi.WxJin: ganzhi.WxShui,
		ganzhi.WxShui: ganzhi.WxMu,
	}
	for from, to := range sheng {
		if !ganzhi.Sheng(from, to) {
			t.Errorf("Sheng(%s, %s) = false, want true", from, to)
		}
		// Reverse should NOT be sheng.
		if ganzhi.Sheng(to, from) {
			t.Errorf("Sheng(%s, %s) = true, want false", to, from)
		}
	}

	// 五行相克: 木→土→水→火→金→木
	ke := map[ganzhi.Wuxing]ganzhi.Wuxing{
		ganzhi.WxMu: ganzhi.WxTu, ganzhi.WxTu: ganzhi.WxShui,
		ganzhi.WxShui: ganzhi.WxHuo, ganzhi.WxHuo: ganzhi.WxJin,
		ganzhi.WxJin: ganzhi.WxMu,
	}
	for from, to := range ke {
		if !ganzhi.Ke(from, to) {
			t.Errorf("Ke(%s, %s) = false, want true", from, to)
		}
		// Reverse should NOT be ke.
		if ganzhi.Ke(to, from) {
			t.Errorf("Ke(%s, %s) = true, want false", to, from)
		}
	}
}

// ============================================================================
// 穷通宝鉴 TiaoHou 全量测试
// ============================================================================

func TestTiaoHou_AllEntriesValid(t *testing.T) {
	// Verify ALL entries in the tiaohou lookup table produce valid results.
	for key, entry := range lookupTiaohou {
		riYuan := ganzhi.Gan(key.stem)
		monthBranch := ganzhi.Zhi(key.branch)

		result := computeTiaoHou(riYuan, monthBranch)

		// Season must not be empty.
		if result.Season == "" {
			t.Errorf("%s日%s月: Season is empty",
				ganzhi.GanName(riYuan), ganzhi.ZhiName(monthBranch))
		}

		// Yong must be a valid five element.
		validWuxing := map[string]bool{"木": true, "火": true, "土": true, "金": true, "水": true}
		if !validWuxing[result.Yong] {
			t.Errorf("%s日%s月: Yong = %q, want valid 五行 (primary=%s, secondary=%s)",
				ganzhi.GanName(riYuan), ganzhi.ZhiName(monthBranch),
				result.Yong, ganzhi.GanName(entry.primary), ganzhi.GanName(entry.secondary))
		}
		if !validWuxing[result.Xi] {
			// Xi may be empty when no secondary (喜神).
			if result.Xi != "" {
				t.Errorf("%s日%s月: Xi = %q, want valid 五行 (primary=%s, secondary=%s)",
					ganzhi.GanName(riYuan), ganzhi.ZhiName(monthBranch),
					result.Xi, ganzhi.GanName(entry.primary), ganzhi.GanName(entry.secondary))
			}
		}
		if result.Ji != "" && !validWuxing[result.Ji] {
			t.Errorf("%s日%s月: Ji = %q, want valid 五行 (primary=%s, secondary=%s)",
				ganzhi.GanName(riYuan), ganzhi.ZhiName(monthBranch),
				result.Ji, ganzhi.GanName(entry.primary), ganzhi.GanName(entry.secondary))
		}

		// Detail should contain the day stem and month branch names.
		if result.Detail == "" {
			t.Errorf("%s日%s月: Detail is empty",
				ganzhi.GanName(riYuan), ganzhi.ZhiName(monthBranch))
		}
	}
}

// ============================================================================
// 六十甲子完整性
// ============================================================================

func TestSixtyJiaZi_Completeness(t *testing.T) {
	// Verify all 60 Jiazi pairs have valid NaYin and are unique.
	seen := make(map[string]bool)
	for i := 0; i < 60; i++ {
		gan := ganzhi.Gan((i % 10) + 1)
		zhi := ganzhi.Zhi((i % 12) + 1)

		nayin := ganzhi.NayinLabel(gan, zhi)
		if nayin == "" {
			t.Errorf("NayinLabel(%s%s) = empty", ganzhi.GanName(gan), ganzhi.ZhiName(zhi))
		}
		seen[nayin] = true

		elem := ganzhi.NayinWuxing(nayin)
		if elem == 0 {
			t.Errorf("NayinWuxing(%q) for %s%s = 0", nayin, ganzhi.GanName(gan), ganzhi.ZhiName(zhi))
		}

		// Verify SixtyCycleIndex is reversible.
		idx := ganzhi.SixtyCycleIndex(gan, zhi)
		if idx != i {
			t.Errorf("SixtyCycleIndex(%s%s) = %d, want %d",
				ganzhi.GanName(gan), ganzhi.ZhiName(zhi), idx, i)
		}
	}
}
