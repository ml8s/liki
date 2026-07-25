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
	mgMale := ComputeMingGua(ganzhi.Male, 1990) // n=5
	if mgMale.GuaNumber != 2 || mgMale.Gua.Name != "坤" {
		t.Errorf("男寄坤: got %d(%s), want 2(坤)", mgMale.GuaNumber, mgMale.Gua.Name)
	}

	mgFemale := ComputeMingGua(ganzhi.Female, 1982) // n=5
	if mgFemale.GuaNumber != 8 || mgFemale.Gua.Name != "艮" {
		t.Errorf("女寄艮: got %d(%s), want 8(艮)", mgFemale.GuaNumber, mgFemale.Gua.Name)
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
	if chart.YearStars.CenterStar.Number == 0 {
		t.Error("YearStars.CenterStar.Number is zero")
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
	if chart.YearStars.CenterStar.Number != 3 {
		t.Errorf("2024 center star = %d, want 3", chart.YearStars.CenterStar.Number)
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
		s := ps.Star
		if s.Number < 1 || s.Number > 9 {
			t.Errorf("palace %d: star number %d out of range", ps.PalaceNum, s.Number)
		}
		if s.Name == "" || s.Color == "" {
			t.Errorf("palace %d: star name or color empty", ps.PalaceNum)
		}
		// Verify against authoritative StarByNumber
		ref := fengshui.StarByNumber(s.Number)
		if s.Name != ref.Name || s.Color != ref.Color || s.Element != ref.Element {
			t.Errorf("palace %d star %d: mismatch with StarByNumber", ps.PalaceNum, s.Number)
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
		// Male cases: n = (Y-4)%9, if n<=0 add 9, no further adjustment
		{"男1984→艮(8)西四命", ganzhi.Male, 1984, "艮", 8, "西四命"},
		{"男1990→中宫寄坤(2)西四命", ganzhi.Male, 1990, "坤", 2, "西四命"},
		{"男1986→坎(1)东四命", ganzhi.Male, 1986, "坎", 1, "东四命"},
		{"男1993→艮(8)西四命", ganzhi.Male, 1993, "艮", 8, "西四命"},
		{"男1985→离(9)东四命", ganzhi.Male, 1985, "离", 9, "东四命"},
		{"男1988→震(3)东四命", ganzhi.Male, 1988, "震", 3, "东四命"},
		{"男1991→乾(6)西四命", ganzhi.Male, 1991, "乾", 6, "西四命"},
		{"男1997→震(3)东四命", ganzhi.Male, 1997, "震", 3, "东四命"},
		// Female cases: n = 11 - ((Y-4)%9), adjust >9
		{"女1990→乾(6)西四命", ganzhi.Female, 1990, "乾", 6, "西四命"},
		{"女1984→震(3)东四命", ganzhi.Female, 1984, "震", 3, "东四命"},
		{"女1982→中宫寄艮(8)西四命", ganzhi.Female, 1982, "艮", 8, "西四命"},
		{"女1986→坎(1)东四命", ganzhi.Female, 1986, "坎", 1, "东四命"},
		{"女1985→坤(2)西四命", ganzhi.Female, 1985, "坤", 2, "西四命"},
		{"女1988→艮(8)西四命", ganzhi.Female, 1988, "艮", 8, "西四命"},
		{"女1991→中宫寄艮(8)西四命", ganzhi.Female, 1991, "艮", 8, "西四命"},
		{"女1995→坎(1)东四命", ganzhi.Female, 1995, "坎", 1, "东四命"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mg := ComputeMingGua(tt.gender, tt.birthYear)
			if mg.Gua.Name != tt.wantName {
				t.Errorf("Name = %s, want %s", mg.Gua.Name, tt.wantName)
			}
			if mg.GuaNumber != tt.wantNum {
				t.Errorf("GuaNumber = %d, want %d", mg.GuaNumber, tt.wantNum)
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
	// birthYear%100 < 4 → Go's negative modulo kicks in
	tests := []struct {
		gender    ganzhi.Gender
		birthYear int
		wantNum   int
		wantName  string
	}{
		// 1900: (0-4)%9 = -4 → n=5 → 男寄坤(2)
		{ganzhi.Male, 1900, 2, "坤"},
		// 1901: (1-4)%9 = -3 → n=6 → 乾(6)
		{ganzhi.Male, 1901, 6, "乾"},
		// 1902: (2-4)%9 = -2 → n=7 → 兑(7)
		{ganzhi.Male, 1902, 7, "兑"},
		// 1903: (3-4)%9 = -1 → n=8 → 艮(8)
		{ganzhi.Male, 1903, 8, "艮"},
		// 1904: (4-4)%9 = 0 → n=9 → 离(9)
		{ganzhi.Male, 1904, 9, "离"},
		// 1904 female: male n=9, female 11-9=2 → 坤(2)
		{ganzhi.Female, 1904, 2, "坤"},
		// 1900 female: male n=5→2(寄坤), female 11-5=6 → 乾(6)
		{ganzhi.Female, 1900, 6, "乾"},
	}
	for _, tt := range tests {
		mg := ComputeMingGua(tt.gender, tt.birthYear)
		if mg.GuaNumber != tt.wantNum {
			t.Errorf("%s %d: num=%d(%s), want %d(%s)",
				tt.gender, tt.birthYear, mg.GuaNumber, mg.Gua.Name, tt.wantNum, tt.wantName)
		}
	}
}

func TestComputeMingGua_FemaleHighYear(t *testing.T) {
	// Female n>9 adjustment: when male_n=1, female 11-1=10 → 1
	// male_n=1 happens when (YY-4)%9=1 → YY%9=5
	tests := []struct {
		birthYear int
		wantNum   int
		wantName  string
	}{
		{1905, 1, "坎"}, // male: (5-4)%9=1, female: 11-1=10→1
		{1914, 1, "坎"}, // (14-4)%9=1
		{1923, 1, "坎"}, // (23-4)%9=10%9=1
		{2005, 1, "坎"}, // 2005
	}
	for _, tt := range tests {
		mg := ComputeMingGua(ganzhi.Female, tt.birthYear)
		if mg.GuaNumber != tt.wantNum {
			t.Errorf("female %d: num=%d(%s), want %d(%s)",
				tt.birthYear, mg.GuaNumber, mg.Gua.Name, tt.wantNum, tt.wantName)
		}
	}
}

func TestComputeMingGua_FemaleZhongGong(t *testing.T) {
	// Female 中宫寄艮(8): male_n=6 → female 11-6=5 → 寄艮(8)
	// male_n=6 when (YY-4)%9=6 → YY%9=1
	tests := []int{1901, 1910, 1919, 1928, 1937, 1946, 1955, 1964, 1973, 1982, 1991}
	for _, year := range tests {
		mg := ComputeMingGua(ganzhi.Female, year)
		if mg.GuaNumber != 8 || mg.Gua.Name != "艮" {
			t.Errorf("female %d: got %d(%s), want 8(艮) — 中宫寄艮", year, mg.GuaNumber, mg.Gua.Name)
		}
	}
}

func TestComputeMingGua_MaleZhongGong(t *testing.T) {
	// Male 中宫寄坤(2): (YY-4)%9=5 → YY%9=0 or 9
	tests := []int{1900, 1909, 1918, 1927, 1936, 1945, 1954, 1963, 1972, 1981, 1990, 1999, 2000, 2009}
	for _, year := range tests {
		mg := ComputeMingGua(ganzhi.Male, year)
		if mg.GuaNumber != 2 || mg.Gua.Name != "坤" {
			t.Errorf("male %d: got %d(%s), want 2(坤) — 中宫寄坤", year, mg.GuaNumber, mg.Gua.Name)
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
		{1, [8]string{"东南", "东", "南", "北", "西", "东北", "西北", "西南"}},  // 坎
		{2, [8]string{"东北", "西", "西北", "西南", "东", "东南", "南", "北"}},   // 坤
		{3, [8]string{"南", "北", "东南", "东", "西南", "西北", "东北", "西"}},   // 震
		{4, [8]string{"北", "南", "东", "东南", "西南", "西北", "西", "东北"}},   // 巽
		{6, [8]string{"西", "东北", "西南", "西北", "东", "北", "东南", "南"}},   // 乾
		{7, [8]string{"西北", "西南", "东北", "西", "北", "南", "东南", "东"}},   // 兑
		{8, [8]string{"西南", "西北", "西", "东北", "南", "北", "东", "东南"}},   // 艮
		{9, [8]string{"东", "东南", "北", "南", "东北", "西", "西南", "西北"}},   // 离
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
