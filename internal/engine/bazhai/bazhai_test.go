package bazhai

import (
	"testing"

	"liki-engine/internal/engine/fengshui"
	"liki-engine/internal/engine/ganzhi"
)

// =============================================================================
// 紫白年飞星 — 入中星验证
// =============================================================================

func TestComputeYearStars_CenterStarKnownYears(t *testing.T) {
	// 下元甲子(1984)七赤入中, 每年逆减1
	// 三元九运: 上元1864一白, 中元1924四绿, 下元1984七赤
	tests := []struct {
		year          int
		wantCenterNum int
		wantColor     string
	}{
		{1984, 7, "赤"}, // 下元甲子 七赤入中
		{1985, 6, "白"}, // 六白入中
		{1986, 5, "黄"}, // 五黄入中
		{1987, 4, "绿"}, // 四绿入中
		{1988, 3, "碧"}, // 三碧入中
		{1989, 2, "黑"}, // 二黑入中
		{1990, 1, "白"}, // 一白入中
		{1991, 9, "紫"}, // 九紫入中
		{1992, 8, "白"}, // 八白入中
		{1993, 7, "赤"}, // 回到七赤 (9年周期)
		{2000, 9, "紫"}, // 九紫入中
		{2010, 8, "白"}, // 八白入中
		{2024, 3, "碧"}, // 2024年三碧入中
		{1924, 4, "绿"}, // 中元甲子 四绿入中
		{1864, 1, "白"}, // 上元甲子 一白入中
	}

	for _, tt := range tests {
		r := computeYearStars(tt.year)
		wantName := fengshui.StarByNumber(tt.wantCenterNum).Name
		if r.RuZhong != wantName {
			t.Errorf("year %d: center star = %s, want %d(%s)",
				tt.year, r.RuZhong, tt.wantCenterNum, tt.wantColor)
		}
		if r.Year != tt.year {
			t.Errorf("year %d: result.Year = %d", tt.year, r.Year)
		}
	}
}

// =============================================================================
// 紫白年飞星 — 全盘飞布验证
// =============================================================================

func TestComputeYearStars_FullDistribution_1984(t *testing.T) {
	// 1984年七赤入中, 按洛书飞行: 6→7→8→9→1→2→3→4
	// 坎1=三碧, 坤2=四绿, 震3=五黄, 巽4=六白, 中5=七赤
	// 乾6=八白, 兑7=九紫, 艮8=一白, 离9=二黑
	want := map[int]int{
		1: 3, // 坎 三碧
		2: 4, // 坤 四绿
		3: 5, // 震 五黄
		4: 6, // 巽 六白
		5: 7, // 中 七赤
		6: 8, // 乾 八白
		7: 9, // 兑 九紫
		8: 1, // 艮 一白
		9: 2, // 离 二黑
	}

	r := computeYearStars(1984)
	for _, ps := range r.Palaces {
		expectedNum, ok := want[ps.GongNum]
		if !ok {
			t.Errorf("unexpected palace num %d", ps.GongNum)
			continue
		}
		if ps.Xing != expectedNum {
			t.Errorf("palace %d: star=%d(%s), want %d",
				ps.GongNum, ps.Xing, ps.XingName, expectedNum)
		}
	}
}

func TestComputeYearStars_FullDistribution_2024(t *testing.T) {
	// 2024年三碧入中, 飞星依次: 4→5→6→7→8→9→1→2
	// 乾6=四绿, 兑7=五黄, 艮8=六白, 离9=七赤, 坎1=八白, 坤2=九紫, 震3=一白, 巽4=二黑, 中5=三碧
	want := map[int]int{
		1: 8, 2: 9, 3: 1, 4: 2, 5: 3, 6: 4, 7: 5, 8: 6, 9: 7,
	}

	r := computeYearStars(2024)
	for _, ps := range r.Palaces {
		if expectedNum := want[ps.GongNum]; ps.Xing != expectedNum {
			t.Errorf("palace %d: star=%d(%s), want %d",
				ps.GongNum, ps.Xing, ps.XingName, expectedNum)
		}
	}
}

// =============================================================================
// 紫白年飞星 — 九宫全
// =============================================================================

func TestComputeYearStars_AllNinePalaces(t *testing.T) {
	r := computeYearStars(1984)
	seen := make(map[int]bool)
	for _, ps := range r.Palaces {
		if ps.GongNum < 1 || ps.GongNum > 9 {
			t.Errorf("invalid palace num %d", ps.GongNum)
		}
		if seen[ps.GongNum] {
			t.Errorf("duplicate palace %d", ps.GongNum)
		}
		seen[ps.GongNum] = true
	}
	for i := 1; i <= 9; i++ {
		if !seen[i] {
			t.Errorf("missing palace %d", i)
		}
	}
}

// =============================================================================
// 紫白年飞星 — 1864年前 (上元前)
// =============================================================================

func TestComputeYearStars_Pre1864(t *testing.T) {
	// 1804年 = 1864 - 60 (一次上推), 逆行3→2→1→9→8→7→... = 7
	// 1804 = 上元甲子前推60年 = 七赤入中
	r := computeYearStars(1804)
	if r.RuZhong != "七赤破军" {
		t.Errorf("year 1804: center star=%s, want 七赤", r.RuZhong)
	}

	// 1844年 (1864前20年): 逆推得三碧入中
	r = computeYearStars(1844)
	if r.RuZhong != "三碧禄存" {
		t.Errorf("year 1844: center star=%s, want 三碧", r.RuZhong)
	}
}

// =============================================================================
// 八宅大游年 — 特定命卦方位验证
// =============================================================================

func TestEightMansionPatterns_KanGua(t *testing.T) {
	// 坎1: 生气=巽4, 天医=震3, 延年=离9, 伏位=坎1
	//      祸害=兑7, 五鬼=艮8, 六煞=乾6, 绝命=坤2
	aus, inaus := eightMansionDirs(1)
	wantAus := [4]int{4, 3, 9, 1}
	wantInaus := [4]int{7, 8, 6, 2}
	for i, v := range wantAus {
		if aus[i] != v {
			t.Errorf("坎 auspicious[%d] = %d, want %d", i, aus[i], v)
		}
	}
	for i, v := range wantInaus {
		if inaus[i] != v {
			t.Errorf("坎 inauspicious[%d] = %d, want %d", i, inaus[i], v)
		}
	}
}

func TestEightMansionPatterns_LiGua(t *testing.T) {
	// 离9: 生气=震3, 天医=巽4, 延年=坎1, 伏位=离9
	//      祸害=艮8, 五鬼=兑7, 六煞=坤2, 绝命=乾6
	aus, inaus := eightMansionDirs(9)
	wantAus := [4]int{3, 4, 1, 9}
	wantInaus := [4]int{8, 7, 2, 6}
	for i, v := range wantAus {
		if aus[i] != v {
			t.Errorf("离 auspicious[%d] = %d, want %d", i, aus[i], v)
		}
	}
	for i, v := range wantInaus {
		if inaus[i] != v {
			t.Errorf("离 inauspicious[%d] = %d, want %d", i, inaus[i], v)
		}
	}
}

func TestEightMansionPatterns_QianGua(t *testing.T) {
	// 乾6: 生气=兑7, 天医=艮8, 延年=坤2, 伏位=乾6
	//      祸害=震3, 五鬼=坎1, 六煞=巽4, 绝命=离9
	aus, inaus := eightMansionDirs(6)
	wantAus := [4]int{7, 8, 2, 6}
	wantInaus := [4]int{3, 1, 4, 9}
	for i, v := range wantAus {
		if aus[i] != v {
			t.Errorf("乾 auspicious[%d] = %d, want %d", i, aus[i], v)
		}
	}
	for i, v := range wantInaus {
		if inaus[i] != v {
			t.Errorf("乾 inauspicious[%d] = %d, want %d", i, inaus[i], v)
		}
	}
}

// =============================================================================
// 八宅大游年 — 方向名验证
// =============================================================================

func TestBaZhaiDirectionsForGua_DirectionNames(t *testing.T) {
	// 坎1宫: 伏位=北(1), 生气=东南(4), 天医=东(3), 延年=南(9)
	//        绝命=西南(2), 五鬼=东北(8), 六煞=西北(6), 祸害=西(7)
	dirs := baZhaiDirectionsForGua(1)

	if dirs.FuWei[0] != "北" {
		t.Errorf("坎伏位 = %s, want 北", dirs.FuWei[0])
	}
	if dirs.ShengQi[0] != "东南" {
		t.Errorf("坎生气 = %s, want 东南", dirs.ShengQi[0])
	}
	if dirs.TianYi[0] != "东" {
		t.Errorf("坎天医 = %s, want 东", dirs.TianYi[0])
	}
	if dirs.YanNian[0] != "南" {
		t.Errorf("坎延年 = %s, want 南", dirs.YanNian[0])
	}
	if dirs.JueMing[0] != "西南" {
		t.Errorf("坎绝命 = %s, want 西南", dirs.JueMing[0])
	}
	if dirs.WuGui[0] != "东北" {
		t.Errorf("坎五鬼 = %s, want 东北", dirs.WuGui[0])
	}
	if dirs.LiuSha[0] != "西北" {
		t.Errorf("坎六煞 = %s, want 西北", dirs.LiuSha[0])
	}
	if dirs.HuoHai[0] != "西" {
		t.Errorf("坎祸害 = %s, want 西", dirs.HuoHai[0])
	}
}

// =============================================================================
// 八宅大游年 — 东/西四命分组方位验证
// =============================================================================

func TestBaZhaiDirectionsForGua_DongXiGroups(t *testing.T) {
	// 东四命(1,3,4,9)的吉方位都是东四宅方向
	// 西四命(2,6,7,8)的吉方位都是西四宅方向
	// 东四宅: 1(北),3(东),4(东南),9(南)
	// 西四宅: 2(西南),6(西北),7(西),8(东北)
	dongSiHouses := map[int]bool{1: true, 3: true, 4: true, 9: true}
	xiSiHouses := map[int]bool{2: true, 6: true, 7: true, 8: true}

	for _, num := range []int{1, 3, 4, 9} {
		aus, _ := eightMansionDirs(num)
		for _, pn := range aus {
			if !dongSiHouses[pn] {
				t.Errorf("东四命%d: 吉方%d不在东四宅", num, pn)
			}
		}
	}

	for _, num := range []int{2, 6, 7, 8} {
		aus, _ := eightMansionDirs(num)
		for _, pn := range aus {
			if !xiSiHouses[pn] {
				t.Errorf("西四命%d: 吉方%d不在西四宅", num, pn)
			}
		}
	}
}

// =============================================================================
// 八宅大游年 — 无效卦号
// =============================================================================

func TestEightMansionDirs_InvalidGua(t *testing.T) {
	// Invalid gua numbers return zero-valued arrays (not empty slices)
	for _, n := range []int{0, 5, 10, -1, 100} {
		aus, inaus := eightMansionDirs(n)
		for _, v := range aus {
			if v != 0 {
				t.Errorf("gua %d: expected all zero aus, got %v", n, aus)
				break
			}
		}
		for _, v := range inaus {
			if v != 0 {
				t.Errorf("gua %d: expected all zero inaus, got %v", n, inaus)
				break
			}
		}
	}
}

// =============================================================================
// guaTable — 洛书数完整性
// =============================================================================

func TestGuaTable_Completeness(t *testing.T) {
	// guaTable indices 1-9 with index 0 empty
	if guaTable[0].Index != 0 {
		t.Error("guaTable[0] should be zero value")
	}
	for i := 1; i <= 9; i++ {
		if guaTable[i].Index != i {
			t.Errorf("guaTable[%d].Index = %d, want %d", i, guaTable[i].Index, i)
		}
		if guaTable[i].Name == "" {
			t.Errorf("guaTable[%d].Name is empty", i)
		}
	}
}

// =============================================================================
// 纳甲 — 柱纳甲
// =============================================================================

func TestZhuNaJia(t *testing.T) {
	// zhuNaJia wraps ganNaJia — 柱的纳甲取决于天干
	p := ganzhi.Zhu{Gan: ganzhi.GanJia, Zhi: ganzhi.ZhiZi}
	g := zhuNaJia(p)
	if g.Name != "乾" {
		t.Errorf("zhuNaJia(甲子) = %s, want 乾", g.Name)
	}

	p2 := ganzhi.Zhu{Gan: ganzhi.GanYi, Zhi: ganzhi.ZhiChou}
	g2 := zhuNaJia(p2)
	if g2.Name != "坤" {
		t.Errorf("zhuNaJia(乙丑) = %s, want 坤", g2.Name)
	}
}

// =============================================================================
// 命卦 — 东/西四命完整分组
// =============================================================================

