package bazhai

import (
	"testing"
	"time"

	"liki-engine/internal/engine/fengshui"
	"liki-engine/internal/engine/ganzhi"
	"liki-engine/internal/engine/tianwen"
)

func TestComputeMingGua_WestGroupMembership(t *testing.T) {
	// westGroup: 坤2,乾6,兑7,艮8
	if len(westGroup) != 4 {
		t.Errorf("westGroup size = %d, want 4", len(westGroup))
	}
	for _, n := range []int{2, 6, 7, 8} {
		if !westGroup[n] {
			t.Errorf("gua %d should be in westGroup", n)
		}
	}
	// 东四命不在 westGroup
	for _, n := range []int{1, 3, 4, 9} {
		if westGroup[n] {
			t.Errorf("gua %d should NOT be in westGroup (east group)", n)
		}
	}
}

// =============================================================================
// 命卦 — 寄宫规则
// =============================================================================

func TestComputeMingGua_ZhongGongRule(t *testing.T) {
	// n=5: 男寄坤(2) 女寄艮(8)
	mgMale := ComputeMingGua(ganzhi.Male, 1995) // (100-95)%9=5 → 坤
	if mgMale.Gua.Index != 2 || mgMale.Gua.Name != "坤" {
		t.Errorf("男寄坤: got %d(%s), want 2(坤)", mgMale.Gua.Index, mgMale.Gua.Name)
	}

	mgFemale := ComputeMingGua(ganzhi.Female, 1990) // (90-4)%9=5 → 艮
	if mgFemale.Gua.Index != 8 || mgFemale.Gua.Name != "艮" {
		t.Errorf("女寄艮: got %d(%s), want 8(艮)", mgFemale.Gua.Index, mgFemale.Gua.Name)
	}
}

// =============================================================================
// 命卦 — 公式锚点（《八宅明镜》通行算法示例，2000 年分界）
// =============================================================================

func TestComputeMingGua_FormulaAnchors(t *testing.T) {
	cases := []struct {
		gender ganzhi.Gender
		year   int
		name   string
		group  string
	}{
		{ganzhi.Male, 1966, "兑", "西四命"},   // (100-66)%9=7 兑
		{ganzhi.Male, 1975, "兑", "西四命"},   // (100-75)%9=7 兑
		{ganzhi.Male, 1977, "坤", "西四命"},   // (100-77)%9=5 → 寄坤
		{ganzhi.Male, 1984, "兑", "西四命"},   // (100-84)%9=7 兑
		{ganzhi.Male, 1990, "坎", "东四命"},   // (100-90)%9=1 坎
		{ganzhi.Female, 1962, "巽", "东四命"}, // (62-4)%9=4 巽
		{ganzhi.Female, 1984, "艮", "西四命"}, // (84-4)%9=8 艮
		{ganzhi.Female, 1985, "离", "东四命"}, // (85-4)%9=0→9 离
		{ganzhi.Male, 2000, "离", "东四命"},   // (99-0)%9=0→9 离（2000 后公式）
		{ganzhi.Male, 2009, "离", "东四命"},   // (99-9)%9=0→9 离
		{ganzhi.Female, 2000, "乾", "西四命"}, // (0+6)%9=6 乾
		{ganzhi.Female, 2005, "坤", "西四命"}, // (5+6)%9=2 坤
	}
	for _, c := range cases {
		mg := ComputeMingGua(c.gender, c.year)
		if mg.Gua.Name != c.name || mg.Group != c.group {
			t.Errorf("%d年%v: got %s(%s), want %s(%s)", c.year, c.gender, mg.Gua.Name, mg.Group, c.name, c.group)
		}
	}
}

// =============================================================================
// ComputeChart — 整合测试
// =============================================================================

func TestComputeChart_Integration(t *testing.T) {
	st := tianwen.SolarTime(time.Date(1984, 2, 4, 12, 0, 0, 0, time.UTC))
	chart := ComputeChart(st, ganzhi.Male)

	if chart.MingGua.Gua.Name == "" {
		t.Error("MingGua.Name is empty")
	}
	if chart.YearStars.RuZhong == "" {
		t.Error("YearStars.RuZhong is empty")
	}
	if len(chart.BaZhaiDirs.ShengQi) == 0 {
		t.Error("BaZhaiDirs.ShengQi is empty")
	}
	if len(chart.ZhuBagua) != 4 {
		t.Errorf("ZhuBagua len = %d, want 4", len(chart.ZhuBagua))
	}
	for i, g := range chart.ZhuBagua {
		if g.Name == "" {
			t.Errorf("ZhuBagua[%d] is empty", i)
		}
	}
}

// =============================================================================
// ComputeChart — 年星与年柱一致性
// =============================================================================

func TestComputeChart_YearStarMatches(t *testing.T) {
	st := tianwen.SolarTime(time.Date(2024, 6, 15, 12, 0, 0, 0, time.UTC))
	chart := ComputeChart(st, ganzhi.Female)

	// 2024: 三碧入中
	if chart.YearStars.RuZhong != "三碧禄存" {
		t.Errorf("2024 center star = %s, want 三碧", chart.YearStars.RuZhong)
	}
	// Year stars should have 9 palaces
	if len(chart.YearStars.Palaces) != 9 {
		t.Errorf("YearStars.Palaces len = %d, want 9", len(chart.YearStars.Palaces))
	}
}

// =============================================================================
// 飞星 — Star 对象完整性
// =============================================================================

func TestComputeYearStars_StarIntegrity(t *testing.T) {
	r := computeYearStars(1984)

	for _, ps := range r.Palaces {
		if ps.Xing < 1 || ps.Xing > 9 {
			t.Errorf("palace %d: star number %d out of range", ps.GongNum, ps.Xing)
		}
		if ps.XingName == "" {
			t.Errorf("palace %d: star name empty", ps.GongNum)
		}
		// Verify against authoritative StarByNumber
		ref := fengshui.StarByNumber(ps.Xing)
		if ps.XingName != ref.Name || ps.Wuxing != ref.Element.String() {
			t.Errorf("palace %d star %d: mismatch with StarByNumber", ps.GongNum, ps.Xing)
		}
	}
}
func TestComputeMingGua(t *testing.T) {
	tests := []struct {
		name      string
		gender    ganzhi.Gender
		birthYear int
		wantName  string
		wantNum   int
		wantGroup string
	}{
		// 《八宅明镜》通行公式（2000 前）：男 (100-年后两位)%9；女 (年后两位-4)%9
		{"男1984→兑(7)西四命", ganzhi.Male, 1984, "兑", 7, "西四命"},
		{"男1990→坎(1)东四命", ganzhi.Male, 1990, "坎", 1, "东四命"},
		{"男1986→中宫寄坤(2)西四命", ganzhi.Male, 1986, "坤", 2, "西四命"},
		{"男1993→兑(7)西四命", ganzhi.Male, 1993, "兑", 7, "西四命"},
		{"男1985→乾(6)西四命", ganzhi.Male, 1985, "乾", 6, "西四命"},
		{"男1988→震(3)东四命", ganzhi.Male, 1988, "震", 3, "东四命"},
		{"男1991→离(9)东四命", ganzhi.Male, 1991, "离", 9, "东四命"},
		{"男1997→震(3)东四命", ganzhi.Male, 1997, "震", 3, "东四命"},
		{"女1990→中宫寄艮(8)西四命", ganzhi.Female, 1990, "艮", 8, "西四命"},
		{"女1984→艮(8)西四命", ganzhi.Female, 1984, "艮", 8, "西四命"},
		{"女1982→乾(6)西四命", ganzhi.Female, 1982, "乾", 6, "西四命"},
		{"女1986→坎(1)东四命", ganzhi.Female, 1986, "坎", 1, "东四命"},
		{"女1985→离(9)东四命", ganzhi.Female, 1985, "离", 9, "东四命"},
		{"女1988→震(3)东四命", ganzhi.Female, 1988, "震", 3, "东四命"},
		{"女1991→乾(6)西四命", ganzhi.Female, 1991, "乾", 6, "西四命"},
		{"女1995→坎(1)东四命", ganzhi.Female, 1995, "坎", 1, "东四命"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mg := ComputeMingGua(tt.gender, tt.birthYear)
			if mg.Gua.Name != tt.wantName {
				t.Errorf("Name = %s, want %s", mg.Gua.Name, tt.wantName)
			}
			if mg.Gua.Index != tt.wantNum {
				t.Errorf("Gua.Index = %d, want %d", mg.Gua.Index, tt.wantNum)
			}
			if mg.Group != tt.wantGroup {
				t.Errorf("Group = %s, want %s", mg.Group, tt.wantGroup)
			}
		})
	}
}

// TestGanNaJia verifies 纳甲 mapping from stems to trigrams.
func TestGanNaJia(t *testing.T) {
	tests := []struct {
		stem     ganzhi.Gan
		wantName string
	}{
		{ganzhi.GanJia, "乾"},
		{ganzhi.GanYi, "坤"},
		{ganzhi.GanBing, "艮"},
		{ganzhi.GanDing, "兑"},
		{ganzhi.GanWu, "坎"},
		{ganzhi.GanJi, "离"},
		{ganzhi.GanGeng, "震"},
		{ganzhi.GanXin, "巽"},
		{ganzhi.GanRen, "乾"},
		{ganzhi.GanGui, "坤"},
	}

	for _, tt := range tests {
		t.Run(ganzhi.GanName(tt.stem), func(t *testing.T) {
			got := ganNaJia(tt.stem)
			if got.Name != tt.wantName {
				t.Errorf("ganNaJia(%s) = %s, want %s",
					ganzhi.GanName(tt.stem), got.Name, tt.wantName)
			}
		})
	}
}

// TestBaZhaiDirections verifies eight mansion directions for all gua numbers.
func TestBaZhaiDirections(t *testing.T) {
	for _, num := range []int{1, 2, 3, 4, 6, 7, 8, 9} {
		dirs := baZhaiDirectionsForGua(num)
		if len(dirs.ShengQi) == 0 || len(dirs.TianYi) == 0 ||
			len(dirs.YanNian) == 0 || len(dirs.FuWei) == 0 ||
			len(dirs.HuoHai) == 0 || len(dirs.WuGui) == 0 ||
			len(dirs.LiuSha) == 0 || len(dirs.JueMing) == 0 {
			t.Errorf("gua %d: incomplete directions", num)
		}
	}
}
func TestComputeMingGua_LowYearBoundary(t *testing.T) {
	// 2000 前女公式 (y-4)%9 对 y<4 产生负模，需处理（1900-1903 女）.
	tests := []struct {
		gender    ganzhi.Gender
		birthYear int
		wantNum   int
		wantName  string
	}{
		// 2000 前男: (100-y)%9
		{ganzhi.Male, 1900, 1, "坎"}, // (100-0)%9=1
		{ganzhi.Male, 1901, 9, "离"}, // (100-1)%9=0→9
		{ganzhi.Male, 1902, 8, "艮"}, // (100-2)%9=8
		{ganzhi.Male, 1903, 7, "兑"}, // (100-3)%9=7
		{ganzhi.Male, 1904, 6, "乾"}, // (100-4)%9=6
		// 2000 前女: (y-4)%9（y<4 时负模 → +9）
		{ganzhi.Female, 1900, 8, "艮"}, // (0-4)%9=-4 → 5 → 寄艮(8)
		{ganzhi.Female, 1901, 6, "乾"}, // (1-4)%9=-3 → 6
		{ganzhi.Female, 1904, 9, "离"}, // (4-4)%9=0 → 9
	}
	for _, tt := range tests {
		mg := ComputeMingGua(tt.gender, tt.birthYear)
		if mg.Gua.Index != tt.wantNum {
			t.Errorf("%s %d: num=%d(%s), want %d(%s)",
				tt.gender, tt.birthYear, mg.Gua.Index, mg.Gua.Name, tt.wantNum, tt.wantName)
		}
	}
}

func TestComputeMingGua_FemaleZhongGong(t *testing.T) {
	// 女寄艮(8)：2000 前 (y-4)%9=5 → y%9=0；2000 后 (y+6)%9=5 → y%9=8
	tests := []int{1909, 1918, 1927, 1936, 1945, 1954, 1963, 1972, 1981, 1990, 1999,
		2008, 2017, 2026, 2035, 2044, 2053, 2062, 2071, 2080, 2089, 2098}
	for _, year := range tests {
		mg := ComputeMingGua(ganzhi.Female, year)
		if mg.Gua.Index != 8 || mg.Gua.Name != "艮" {
			t.Errorf("female %d: got %d(%s), want 8(艮) — 中宫寄艮", year, mg.Gua.Index, mg.Gua.Name)
		}
	}
}

func TestComputeMingGua_MaleZhongGong(t *testing.T) {
	// 男寄坤(2)：2000 前 (100-y)%9=5 → y%9=5；2000 后 (99-y)%9=5 → y%9=4
	tests := []int{1905, 1914, 1923, 1932, 1941, 1950, 1959, 1968, 1977, 1986, 1995,
		2004, 2013, 2022, 2031, 2040, 2049, 2058, 2067, 2076, 2085, 2094}
	for _, year := range tests {
		mg := ComputeMingGua(ganzhi.Male, year)
		if mg.Gua.Index != 2 || mg.Gua.Name != "坤" {
			t.Errorf("male %d: got %d(%s), want 2(坤) — 中宫寄坤", year, mg.Gua.Index, mg.Gua.Name)
		}
	}
}

// ── baZhaiDirections golden: all 8 gua ──

func TestBaZhaiDirectionsForGua_AllEight(t *testing.T) {
	// Golden values for all 8 gua numbers against standard 大游年歌.
	// Order: 生气,天医,延年,伏位,祸害,五鬼,六煞,绝命
	tests := []struct {
		num  int
		want [8]string // direction names
	}{
		{1, [8]string{"东南", "东", "南", "北", "西", "东北", "西北", "西南"}}, // 坎
		{2, [8]string{"东北", "西", "西北", "西南", "东", "东南", "南", "北"}}, // 坤
		{3, [8]string{"南", "北", "东南", "东", "西南", "西北", "东北", "西"}}, // 震
		{4, [8]string{"北", "南", "东", "东南", "西北", "西南", "西", "东北"}}, // 巽（祸害乾、五鬼坤）
		{6, [8]string{"西", "东北", "西南", "西北", "东南", "东", "北", "南"}}, // 乾（祸害巽、五鬼震、六煞坎）
		{7, [8]string{"西北", "西南", "东北", "西", "北", "南", "东南", "东"}}, // 兑
		{8, [8]string{"西南", "西北", "西", "东北", "南", "北", "东", "东南"}}, // 艮
		{9, [8]string{"东", "东南", "北", "南", "东北", "西", "西南", "西北"}}, // 离
	}
	for _, tt := range tests {
		dirs := baZhaiDirectionsForGua(tt.num)
		got := [8]string{
			dirs.ShengQi[0], dirs.TianYi[0], dirs.YanNian[0], dirs.FuWei[0],
			dirs.HuoHai[0], dirs.WuGui[0], dirs.LiuSha[0], dirs.JueMing[0],
		}
		if got != tt.want {
			t.Errorf("gua %d:\n  got  %v\n  want %v", tt.num, got, tt.want)
		}
	}
}

// ── guaTable completeness ──

func TestGuaTable_WuxingCorrectness(t *testing.T) {
	// Verify five-element attributes against standard bagua 五行
	want := map[int]string{
		1: "水", 2: "土", 3: "木", 4: "木",
		6: "金", 7: "金", 8: "土", 9: "火",
	}
	for num, wx := range want {
		if guaTable[num].Wuxing != wx {
			t.Errorf("guaTable[%d].Wuxing=%s, want %s", num, guaTable[num].Wuxing, wx)
		}
	}
}

func TestGuaTable_YinYangCorrectness(t *testing.T) {
	want := map[int]string{
		1: "阳", 2: "阴", 3: "阳", 4: "阴",
		6: "阳", 7: "阴", 8: "阳", 9: "阴",
	}
	for num, yy := range want {
		if guaTable[num].YinYang != yy {
			t.Errorf("guaTable[%d].YinYang=%s, want %s", num, guaTable[num].YinYang, yy)
		}
	}
}
