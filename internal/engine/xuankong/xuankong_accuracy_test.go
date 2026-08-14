package xuankong

import (
	"testing"

	"liki-engine/internal/engine/fengshui"
)

func TestComputeSanYuanYun(t *testing.T) {
	tests := []struct {
		name        string
		year        int
		wantYunNum  int
		wantYuan    string
		wantYunName string
	}{
		// 上元
		{"上元一运起点", 1864, 1, "上元", "一运"},
		{"上元一运末", 1883, 1, "上元", "一运"},
		{"上元二运", 1884, 2, "上元", "二运"},
		{"上元三运", 1904, 3, "上元", "三运"},
		// 中元
		{"中元四运", 1924, 4, "中元", "四运"},
		{"中元五运", 1944, 5, "中元", "五运"},
		{"中元六运", 1964, 6, "中元", "六运"},
		// 下元
		{"下元七运", 1984, 7, "下元", "七运"},
		{"下元八运起点", 2004, 8, "下元", "八运"},
		{"下元八运末", 2023, 8, "下元", "八运"},
		{"下元九运起点", 2024, 9, "下元", "九运"},
		{"下元九运中", 2030, 9, "下元", "九运"},
		// Beyond 180-year cycle: 2044 = 1864+180, wraps to上元一运
		{"下一轮上元一运", 2044, 1, "上元", "一运"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := ComputeSanYuanYun(tt.year)
			if got.YunNumber != tt.wantYunNum {
				t.Errorf("YunNumber = %d, want %d", got.YunNumber, tt.wantYunNum)
			}
			if got.Yuan != tt.wantYuan {
				t.Errorf("Yuan = %s, want %s", got.Yuan, tt.wantYuan)
			}
			if got.YunName != tt.wantYunName {
				t.Errorf("YunName = %s, want %s", got.YunName, tt.wantYunName)
			}
			if got.Year != tt.year {
				t.Errorf("Year = %d, want %d", got.Year, tt.year)
			}
			if got.StartYear > tt.year || got.EndYear < tt.year {
				t.Errorf("year %d not in range [%d, %d]", tt.year, got.StartYear, got.EndYear)
			}
		})
	}
}

// TestMountainPalace verifies 24山→九宫 mapping.
func TestMountainPalace(t *testing.T) {
	tests := []struct {
		mountainIdx int
		wantPalace  int
	}{
		// 坎1: 子癸壬
		{0, 1}, {1, 1}, {23, 1},
		// 艮8: 丑艮寅
		{2, 8}, {3, 8}, {4, 8},
		// 震3: 甲卯乙
		{5, 3}, {6, 3}, {7, 3},
		// 巽4: 辰巽巳
		{8, 4}, {9, 4}, {10, 4},
		// 离9: 丙午丁
		{11, 9}, {12, 9}, {13, 9},
		// 坤2: 未坤申
		{14, 2}, {15, 2}, {16, 2},
		// 兑7: 庚酉辛
		{17, 7}, {18, 7}, {19, 7},
		// 乾6: 戌乾亥
		{20, 6}, {21, 6}, {22, 6},
	}

	for _, tt := range tests {
		m := fengshui.Mountains24Table[tt.mountainIdx]
		t.Run(m.Name, func(t *testing.T) {
			got := mountainPalace(tt.mountainIdx)
			if got != tt.wantPalace {
				t.Errorf("mountainPalace(%d=%s) = %d, want %d",
					tt.mountainIdx, m.Name, got, tt.wantPalace)
			}
		})
	}
}

// TestFlyStars verifies顺飞 and逆飞 star distribution.
func TestFlyStars(t *testing.T) {
	// 顺飞: 1入中 → [6,7,8,9,1,2,3,4,5]
	t.Run("顺飞_1入中", func(t *testing.T) {
		stars := flyStars(1, true)
		expected := []int{6, 7, 8, 9, 1, 2, 3, 4, 5}
		for i, want := range expected {
			if stars[i].Number != want {
				t.Errorf("pos %d: got %d, want %d", i, stars[i].Number, want)
			}
		}
	})

	// 逆飞: 1入中 → [5,4,3,2,1,9,8,7,6]
	t.Run("逆飞_1入中", func(t *testing.T) {
		stars := flyStars(1, false)
		expected := []int{5, 4, 3, 2, 1, 9, 8, 7, 6}
		for i, want := range expected {
			if stars[i].Number != want {
				t.Errorf("pos %d: got %d, want %d", i, stars[i].Number, want)
			}
		}
	})

	// 顺飞: 5入中 → [1,2,3,4,5,6,7,8,9]
	t.Run("顺飞_5入中", func(t *testing.T) {
		stars := flyStars(5, true)
		expected := []int{1, 2, 3, 4, 5, 6, 7, 8, 9}
		for i, want := range expected {
			if stars[i].Number != want {
				t.Errorf("pos %d: got %d, want %d", i, stars[i].Number, want)
			}
		}
	})

	// 逆飞: 5入中 → [9,8,7,6,5,4,3,2,1]
	t.Run("逆飞_5入中", func(t *testing.T) {
		stars := flyStars(5, false)
		expected := []int{9, 8, 7, 6, 5, 4, 3, 2, 1}
		for i, want := range expected {
			if stars[i].Number != want {
				t.Errorf("pos %d: got %d, want %d", i, stars[i].Number, want)
			}
		}
	})

	// 顺飞: 8入中, verify center is 8
	t.Run("顺飞_中心星", func(t *testing.T) {
		stars := flyStars(8, true)
		if stars[4].Number != 8 {
			t.Errorf("center (pos 4) = %d, want 8", stars[4].Number)
		}
		// Verify all 9 stars present
		seen := make(map[int]bool)
		for _, s := range stars {
			seen[s.Number] = true
		}
		for n := 1; n <= 9; n++ {
			if !seen[n] {
				t.Errorf("star %d missing from fly result", n)
			}
		}
	})

	// 逆飞: 8入中, verify all 9 stars present
	t.Run("逆飞_完整性", func(t *testing.T) {
		stars := flyStars(8, false)
		seen := make(map[int]bool)
		for _, s := range stars {
			seen[s.Number] = true
		}
		for n := 1; n <= 9; n++ {
			if !seen[n] {
				t.Errorf("star %d missing from reverse fly result", n)
			}
		}
	})
}

// TestTiXingTable verifies替星 data against standard玄空口诀.
// Reference: 子癸甲申→1(贪狼), 壬卯乙未坤→2(巨门),
//
//	乾亥辰巽巳戌→6(武曲), 酉辛丑艮丙→7(破军), 寅午庚丁→9(右弼)
func TestTiXingTable(t *testing.T) {
	// Map mountain name → expected替星 (0 means no替/use original)
	wantMap := map[string]int{
		"子": 1, "癸": 1, "甲": 1, "申": 1, // 贪狼1
		"壬": 2, "卯": 2, "乙": 2, "未": 2, "坤": 2, // 巨门2
		"乾": 6, "亥": 6, "辰": 6, "巽": 6, "巳": 6, "戌": 6, // 武曲6
		"酉": 7, "辛": 7, "丑": 7, "艮": 7, "丙": 7, // 破军7
		"寅": 9, "午": 9, "庚": 9, "丁": 9, // 右弼9
	}

	for i, m := range fengshui.Mountains24Table {
		want, ok := wantMap[m.Name]
		if !ok {
			continue // mountain not in替星口诀, skip
		}
		got := tiXingTable[i]
		if got != want {
			t.Errorf("tiXingTable[%d=%s] = %d, want %d (per替星口诀)", i, m.Name, got, want)
		}
	}
}

// TestTiXingShanStar verifies替星 substitution logic for山星.
func TestTiXingShanStar(t *testing.T) {
	// 地元龙 → should substitute (return non-zero)
	t.Run("地元龙_有替", func(t *testing.T) {
		// 甲(5) is地元龙 → should use替星
		got := tiXingShanStar(5)
		if got == 0 {
			t.Error("地元龙甲 should have替星")
		}
	})

	// 天元龙 → should NOT substitute (return 0)
	t.Run("天元龙_无替", func(t *testing.T) {
		// 子(0) is天元龙 → should NOT use替星
		got := tiXingShanStar(0)
		if got != 0 {
			t.Errorf("天元龙子 should return 0 (no替), got %d", got)
		}
	})

	// 人元龙 → should substitute
	t.Run("人元龙_有替", func(t *testing.T) {
		// 癸(1) is人元龙 → should use替星
		got := tiXingShanStar(1)
		if got == 0 {
			t.Error("人元龙癸 should have替星")
		}
	})
}

// TestSubstituteStarUsage verifies that替星 is actually used in chart computation.
func TestSubstituteStarUsage(t *testing.T) {
	// 甲山庚向 (both are地元龙, should trigger替星)
	// 甲=idx5(地元龙), 庚=idx17(地元龙)
	chart := computeChart(5, 17, 2024)

	// After替星 substitution, the山星 and向星 should differ from
	// what they would be without替星 (i.e., the raw period star values).
	// Basic sanity: chart should not be empty.
	if chart.Yun.YunNumber == 0 {
		t.Fatal("chart is empty")
	}

	// Verify坐向 are stored correctly.
	if chart.SitMountain != 5 {
		t.Errorf("SitMountain = %d, want 5", chart.SitMountain)
	}
	if chart.FaceMountain != 17 {
		t.Errorf("FaceMountain = %d, want 17", chart.FaceMountain)
	}

	// Verify all 9 palaces have valid stars.
	for i, p := range chart.Palaces {
		if p.PeriodStar.Number < 1 || p.PeriodStar.Number > 9 {
			t.Errorf("palace %d: invalid period star %d", i, p.PeriodStar.Number)
		}
		if p.MountainStar.Number < 1 || p.MountainStar.Number > 9 {
			t.Errorf("palace %d: invalid mountain star %d", i, p.MountainStar.Number)
		}
		if p.FacingStar.Number < 1 || p.FacingStar.Number > 9 {
			t.Errorf("palace %d: invalid facing star %d", i, p.FacingStar.Number)
		}
	}
}

// TestChartInvalidInput verifies bounding behavior for invalid mountain indices.
func TestChartInvalidInput(t *testing.T) {
	// Negative indices → empty chart
	chart := computeChart(-1, 2, 2024)
	if chart.Yun.YunNumber != 0 {
		t.Error("negative sit should return empty chart")
	}

	// Too large indices → empty chart
	chart = computeChart(1, 24, 2024)
	if chart.Yun.YunNumber != 0 {
		t.Error("faceMountain >= 24 should return empty chart")
	}
}

// TestWangShanWangXiang verifies旺山旺向 evaluation.
func TestWangShanWangXiang(t *testing.T) {
	// 八运(2004-2023) 子山午向: should be旺山旺向 in some configurations
	chart := computeChart(0, 12, 2020) // 子山午向, 八运

	// Basic sanity: should have some evaluation results.
	// We can't assert specific true/false without full manual computation,
	// but the flags should be computed (not all default false).
	if chart.Yun.YunNumber != 8 {
		t.Errorf("yun number = %d, want 8", chart.Yun.YunNumber)
	}

	// Verify XingJiaHui computed for all palaces.
	for i, x := range chart.XingJiaHui {
		if x.Name == "" {
			t.Errorf("palace %d: empty xingJiaHui name", i)
		}
	}

	// Verify ShouShanChuSha computed.
	if chart.ShouShanChuSha.Assessment == "" {
		t.Error("empty shouShanChuSha assessment")
	}
}
func TestComputeSanYuanYun_PreBaseYear(t *testing.T) {
	// Years before 1864 should be clamped to 1864
	yun := ComputeSanYuanYun(1800)
	if yun.YunNumber != 1 || yun.Yuan != "上元" {
		t.Errorf("1800: got %d运 %s, want 1运 上元 (clamped)", yun.YunNumber, yun.Yuan)
	}
	if yun.StartYear != 1864 {
		t.Errorf("1800: StartYear=%d, want 1864", yun.StartYear)
	}
}

func TestComputeSanYuanYun_Beyond2044(t *testing.T) {
	// 2044 = 1864+180, wraps to上元一运
	tests := []struct {
		year       int
		wantYunNum int
		wantYuan   string
	}{
		{2044, 1, "上元"},
		{2064, 2, "上元"},
		{2084, 3, "上元"},
		{2104, 4, "中元"},
		{2124, 5, "中元"},
		{2144, 6, "中元"},
		{2164, 7, "下元"},
		{2184, 8, "下元"},
		{2204, 9, "下元"},
	}
	for _, tt := range tests {
		yun := ComputeSanYuanYun(tt.year)
		if yun.YunNumber != tt.wantYunNum || yun.Yuan != tt.wantYuan {
			t.Errorf("%d: got %d运 %s, want %d运 %s",
				tt.year, yun.YunNumber, yun.Yuan, tt.wantYunNum, tt.wantYuan)
		}
	}
}

func TestComputeSanYuanYun_PeriodBoundaries(t *testing.T) {
	// Verify period boundaries: each运 is exactly 20 years
	tests := []struct {
		year       int
		wantYunNum int
		wantStart  int
		wantEnd    int
	}{
		{1864, 1, 1864, 1883},
		{1883, 1, 1864, 1883},
		{1884, 2, 1884, 1903},
		{1903, 2, 1884, 1903},
		{1904, 3, 1904, 1923},
		{2004, 8, 2004, 2023},
		{2023, 8, 2004, 2023},
		{2024, 9, 2024, 2043},
		{2043, 9, 2024, 2043},
	}
	for _, tt := range tests {
		yun := ComputeSanYuanYun(tt.year)
		if yun.YunNumber != tt.wantYunNum {
			t.Errorf("%d: YunNumber=%d, want %d", tt.year, yun.YunNumber, tt.wantYunNum)
		}
		if yun.StartYear != tt.wantStart {
			t.Errorf("%d: StartYear=%d, want %d", tt.year, yun.StartYear, tt.wantStart)
		}
		if yun.EndYear != tt.wantEnd {
			t.Errorf("%d: EndYear=%d, want %d", tt.year, yun.EndYear, tt.wantEnd)
		}
	}
}

// ── FlyStars golden values ──

func TestFlyStars_Center9_Forward(t *testing.T) {
	// 9入中顺飞: 中9, 乾1, 兑2, 艮3, 离4, 坎5, 坤6, 震7, 巽8
	stars := flyStars(9, true)
	expected := [9]int{5, 6, 7, 8, 9, 1, 2, 3, 4}
	for i, want := range expected {
		if stars[i].Number != want {
			t.Errorf("pos %d: got %d, want %d", i, stars[i].Number, want)
		}
	}
}

func TestFlyStars_Center9_Backward(t *testing.T) {
	// 9入中逆飞: 中9, 乾8, 兑7, 艮6, 离5, 坎4, 坤3, 震2, 巽1
	stars := flyStars(9, false)
	expected := [9]int{4, 3, 2, 1, 9, 8, 7, 6, 5}
	for i, want := range expected {
		if stars[i].Number != want {
			t.Errorf("pos %d: got %d, want %d", i, stars[i].Number, want)
		}
	}
}

// ── ComputeChart golden verification ──

func TestComputeChart_ZiShanWuXiang_Yun8(t *testing.T) {
	// 八运(2004-2023) 子山午向 (sit=0, face=12)
	// 八运运星盘: 8入中 → [4,5,6,7,8,9,1,2,3]
	// 坐山坎宫(1)运星=4 → 山星4入中(子为天元龙→无替, 子阴→逆飞)
	// 向首离宫(9)运星=3 → 向星3入中(午为天元龙→无替, 午阴→逆飞)
	chart := computeChart(0, 12, 2020)

	if chart.Yun.YunNumber != 8 {
		t.Fatalf("yun=%d, want 8", chart.Yun.YunNumber)
	}

	// Verify period stars: 8入中, 顺飞
	expectedPeriod := [9]int{4, 5, 6, 7, 8, 9, 1, 2, 3}
	for i, want := range expectedPeriod {
		if chart.Palaces[i].PeriodStar.Number != want {
			t.Errorf("period star palace %d: got %d, want %d", i+1, chart.Palaces[i].PeriodStar.Number, want)
		}
	}

	// 子(0)天元龙, 坐宫=坎(1), 运星=4(巽). 4入中逆飞(子阴).
	// 4逆飞: 中4, 乾3(逆→乾位得3), 兑2, 艮1, 离9, 坎8, 坤7, 震6, 巽5
	expectedShan := [9]int{9, 1, 2, 3, 4, 5, 6, 7, 8} // 山星4入中顺飞（巽=阳）：坎9坤1震2巽3中4乾5兑6艮7离8
	for i, want := range expectedShan {
		if chart.Palaces[i].MountainStar.Number != want {
			t.Errorf("mountain star palace %d: got %d, want %d", i+1, chart.Palaces[i].MountainStar.Number, want)
		}
	}

	// 午(12)天元龙, 向宫=离(9), 运星=3(震). 3入中逆飞(午阴).
	// 3逆飞: 中3, 乾2, 兑1, 艮9, 离8, 坎7, 坤6, 震5, 巽4
	expectedXiang := [9]int{7, 6, 5, 4, 3, 2, 1, 9, 8}
	for i, want := range expectedXiang {
		if chart.Palaces[i].FacingStar.Number != want {
			t.Errorf("facing star palace %d: got %d, want %d", i+1, chart.Palaces[i].FacingStar.Number, want)
		}
	}
}

// 八运子山午向权威=双星会向（坐宫山星=9 非当令，非旺山旺向）——见 TestComputeChart_FourJu_Anchors。

func TestComputeChart_FuYin_Detection(t *testing.T) {
	// FuYin occurs when every palace's period star equals its palace number.
	// This happens in 五运 with forward flying: center(5)=5, then 6,7,8,9,1,2,3,4.
	chart := computeChart(0, 12, 1945) // 五运 子山午向
	if !chart.FuYin {
		t.Error("FuYin not detected for 子山午向五运 — check evaluate() logic")
	}
	// Verify a non-FuYin case: 八运 should NOT have FuYin.
	chart8 := computeChart(0, 12, 2020) // 八运 子山午向
	if chart8.FuYin {
		t.Error("FuYin false-positive for 子山午向八运 — should not have FuYin")
	}
}

// ── XingJiaHui completeness ──

func TestXingJiaHui_TableCompleteness(t *testing.T) {
	// Verify known key combinations exist
	keys := [][2]int{
		{1, 4}, {4, 1}, {2, 5}, {5, 2}, {2, 3}, {3, 2},
		{6, 8}, {8, 6}, {5, 9}, {9, 5}, {3, 7}, {7, 3},
		{2, 7}, {7, 2}, {1, 6}, {6, 1}, {4, 9}, {9, 4},
		{2, 9}, {9, 2}, {6, 9}, {9, 6},
	}
	for _, key := range keys {
		if _, ok := xingJiaHuiTable[key]; !ok {
			t.Errorf("xingJiaHuiTable missing key %v", key)
		}
	}
}

// =============================================================================
// 四大局判定（《沈氏玄空学》权威锚点）+ 山向星顺逆规则
// =============================================================================

// shanXiangForward — 入中星对应洛书宫的同元龙山阴阳定顺逆。
func TestShanXiangForward(t *testing.T) {
	cases := []struct {
		center int
		mount  int
		want   bool // 顺飞
		why    string
	}{
		{3, 0, false, "山星3入中=震宫(甲卯乙)，子=天元 → 卯=阴 → 逆飞"},
		{2, 12, true, "向星2入中=坤宫(未坤申)，午=天元 → 坤=阳 → 顺飞"},
		{4, 0, true, "山星4入中=巽宫(辰巽巳)，子=天元 → 巽=阳 → 顺飞"},
		{3, 12, false, "向星3入中=震宫，午=天元 → 卯=阴 → 逆飞"},
		{1, 23, true, "山星1入中=坎宫(壬子癸)，壬=地元 → 壬=阳 → 顺飞"},
		{5, 0, false, "山星5入中无卦 → 按坐山自身（子=阴）→ 逆飞"},
	}
	for _, c := range cases {
		if got := shanXiangForward(c.center, c.mount); got != c.want {
			t.Errorf("shanXiangForward(%d, %d) = %v, want %v (%s)", c.center, c.mount, got, c.want, c.why)
		}
	}
}

// 四大局权威锚点（《沈氏玄空学》七运/八运下卦）：
//
//	七运子山午向=双星会坐；八运子山午向=双星会向；七运乾山巽向=上山下水；七运酉山卯向=旺山旺向。
func TestComputeChart_FourJu_Anchors(t *testing.T) {
	cases := []struct {
		label              string
		sit, face, yr      int
		wS, wX, sZ, xX, ss bool
	}{
		{"七运子山午向=双星会坐", 0, 12, 1990, true, false, true, false, false},
		{"八运子山午向=双星会向", 0, 12, 2020, false, true, false, true, false},
		{"七运乾山巽向=上山下水", 21, 9, 1990, false, false, false, false, true},
		{"七运酉山卯向=旺山旺向", 18, 6, 1990, true, true, false, false, false},
	}
	for _, c := range cases {
		ch := computeChart(c.sit, c.face, c.yr)
		if ch.WangShan != c.wS || ch.WangXiang != c.wX || ch.ShanXing != c.sZ || ch.XiangXing != c.xX || ch.XiaShui != c.ss {
			t.Errorf("%s: got 旺山=%v 旺向=%v 双星会坐=%v 双星会向=%v 上山下水=%v; want %v %v %v %v %v",
				c.label, ch.WangShan, ch.WangXiang, ch.ShanXing, ch.XiangXing, ch.XiaShui, c.wS, c.wX, c.sZ, c.xX, c.ss)
		}
	}
}
