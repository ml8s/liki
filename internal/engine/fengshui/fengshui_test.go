package fengshui

import (
	"testing"

	"liki-engine/internal/engine/ganzhi"
)

// =============================================================================
// 24 山方位 — 角度与方向验证
// =============================================================================

func Test24Mountains_AngleConsistency(t *testing.T) {
	// 每山 15°, index*15 = angle
	for i, m := range Mountains24Table {
		if m.Index != i {
			t.Errorf("mountain %s: index=%d, want %d", m.Name, m.Index, i)
		}
		if m.Angle != i*15 {
			t.Errorf("mountain %s: angle=%d, want %d", m.Name, m.Angle, i*15)
		}
	}
}

func Test24Mountains_PrincipalDirections(t *testing.T) {
	// 四正方位: 子(正北0°), 午(正南180°), 卯(正东90°), 酉(正西270°)
	tests := []struct {
		index int
		name  string
		angle int
	}{
		{0, "子", 0},
		{6, "卯", 90},
		{12, "午", 180},
		{18, "酉", 270},
	}
	for _, tt := range tests {
		m := Mountains24Table[tt.index]
		if m.Name != tt.name || m.Angle != tt.angle {
			t.Errorf("index %d: got %s(%d°), want %s(%d°)",
				tt.index, m.Name, m.Angle, tt.name, tt.angle)
		}
	}
}

func Test24Mountains_OppositePairs(t *testing.T) {
	// 对宫相差180° = index相差12
	pairs := [][2]string{
		{"子", "午"}, {"癸", "丁"}, {"丑", "未"},
		{"艮", "坤"}, {"寅", "申"}, {"甲", "庚"},
		{"卯", "酉"}, {"乙", "辛"}, {"辰", "戌"},
		{"巽", "乾"}, {"巳", "亥"}, {"丙", "壬"},
	}
	for _, pair := range pairs {
		// Find both in the table
		var idxA, idxB int
		for _, m := range Mountains24Table {
			if m.Name == pair[0] {
				idxA = m.Index
			}
			if m.Name == pair[1] {
				idxB = m.Index
			}
		}
		if (idxA+12)%24 != idxB {
			t.Errorf("%s(index %d) → %s(index %d): not 180° apart (expected %d apart)",
				pair[0], idxA, pair[1], idxB, 12)
		}
	}
}

// =============================================================================
// 24 山 — 三元龙分类
// =============================================================================

func Test24Mountains_YuanLongCount(t *testing.T) {
	// 天地人三元各 8 山
	var tian, di, ren int
	for _, m := range Mountains24Table {
		switch m.YuanLong {
		case "天元龙":
			tian++
		case "地元龙":
			di++
		case "人元龙":
			ren++
		}
	}
	if tian != 8 {
		t.Errorf("天元龙 count = %d, want 8", tian)
	}
	if di != 8 {
		t.Errorf("地元龙 count = %d, want 8", di)
	}
	if ren != 8 {
		t.Errorf("人元龙 count = %d, want 8", ren)
	}
}

func Test24Mountains_TianYuanLong(t *testing.T) {
	// 天元龙: 子午卯酉乾坤艮巽 (四正+四隅的卦主)
	tianNames := map[string]bool{
		"子": true, "午": true, "卯": true, "酉": true,
		"乾": true, "坤": true, "艮": true, "巽": true,
	}
	for _, m := range Mountains24Table {
		if m.YuanLong == "天元龙" {
			if !tianNames[m.Name] {
				t.Errorf("%s is marked 天元龙 but should not be", m.Name)
			}
			delete(tianNames, m.Name)
		}
	}
	for n := range tianNames {
		t.Errorf("%s should be 天元龙 but is not", n)
	}
}

func Test24Mountains_TrigramAssignment(t *testing.T) {
	// 每卦三山: 壬子癸→坎, 丑艮寅→艮, 甲卯乙→震, 辰巽巳→巽,
	//           丙午丁→离, 未坤申→坤, 庚酉辛→兑, 戌乾亥→乾
	trigramMountains := map[string][]string{
		"坎": {"壬", "子", "癸"},
		"艮": {"丑", "艮", "寅"},
		"震": {"甲", "卯", "乙"},
		"巽": {"辰", "巽", "巳"},
		"离": {"丙", "午", "丁"},
		"坤": {"未", "坤", "申"},
		"兑": {"庚", "酉", "辛"},
		"乾": {"戌", "乾", "亥"},
	}
	for trigram, expected := range trigramMountains {
		for _, name := range expected {
			found := false
			for _, m := range Mountains24Table {
				if m.Name == name && m.Trigram == trigram {
					found = true
					break
				}
			}
			if !found {
				t.Errorf("%s should belong to %s卦", name, trigram)
			}
		}
	}
}

func Test24Mountains_TrigramYuanLong(t *testing.T) {
	// 每卦三山分属天地人三元: 顺时针为地元→天元→人元 或 逆时针
	// 四正卦(坎震离兑): 地元在左(顺时针方向),天元在中,人元在右
	// 四隅卦(乾坤艮巽): 地元在逆时针方向,天元在中,人元在顺时针
	// 实际排列: 壬(地)子(天)癸(人), 丑(地)艮(天)寅(人), ...
	expected := map[string]string{
		"壬": "地元龙", "子": "天元龙", "癸": "人元龙", // 坎
		"丑": "地元龙", "艮": "天元龙", "寅": "人元龙", // 艮
		"甲": "地元龙", "卯": "天元龙", "乙": "人元龙", // 震
		"辰": "地元龙", "巽": "天元龙", "巳": "人元龙", // 巽
		"丙": "地元龙", "午": "天元龙", "丁": "人元龙", // 离
		"未": "地元龙", "坤": "天元龙", "申": "人元龙", // 坤
		"庚": "地元龙", "酉": "天元龙", "辛": "人元龙", // 兑
		"戌": "地元龙", "乾": "天元龙", "亥": "人元龙", // 乾
	}
	for _, m := range Mountains24Table {
		if e, ok := expected[m.Name]; ok && m.YuanLong != e {
			t.Errorf("%s: yuanlong=%s, want %s", m.Name, m.YuanLong, e)
		}
	}
}

// =============================================================================
// 24 山 — 阴阳与五行
// =============================================================================

func Test24Mountains_WuxingAssignment(t *testing.T) {
	// 24山五行按地支/天干本气, 非卦气
	// 地支: 寅卯=木, 巳午=火, 申酉=金, 亥子=水, 辰戌丑未=土
	// 天干: 甲乙=木, 丙丁=火, 庚辛=金, 壬癸=水
	// 卦山: 乾=金, 坤=土, 艮=土, 巽=木
	type want struct {
		name   string
		wuxing ganzhi.Wuxing
	}
	ref := []want{
		{"子", ganzhi.WxShui}, {"癸", ganzhi.WxShui}, {"丑", ganzhi.WxTu},
		{"艮", ganzhi.WxTu}, {"寅", ganzhi.WxMu}, {"甲", ganzhi.WxMu},
		{"卯", ganzhi.WxMu}, {"乙", ganzhi.WxMu}, {"辰", ganzhi.WxTu},
		{"巽", ganzhi.WxMu}, {"巳", ganzhi.WxHuo}, {"丙", ganzhi.WxHuo},
		{"午", ganzhi.WxHuo}, {"丁", ganzhi.WxHuo}, {"未", ganzhi.WxTu},
		{"坤", ganzhi.WxTu}, {"申", ganzhi.WxJin}, {"庚", ganzhi.WxJin},
		{"酉", ganzhi.WxJin}, {"辛", ganzhi.WxJin}, {"戌", ganzhi.WxTu},
		{"乾", ganzhi.WxJin}, {"亥", ganzhi.WxShui}, {"壬", ganzhi.WxShui},
	}
	for _, w := range ref {
		found := false
		for _, m := range Mountains24Table {
			if m.Name == w.name {
				if m.Element != w.wuxing {
					t.Errorf("%s: wuxing=%s, want %s", m.Name, m.Element, w.wuxing)
				}
				found = true
				break
			}
		}
		if !found {
			t.Errorf("mountain %s not in table", w.name)
		}
	}
}

func Test24Mountains_DiTianRenYinYang(t *testing.T) {
	// 四正卦(坎震离兑): 天元=阴
	zhengGua := map[string]bool{"坎": true, "震": true, "离": true, "兑": true}
	for _, m := range Mountains24Table {
		if m.YuanLong != "天元龙" {
			continue
		}
		if zhengGua[m.Trigram] {
			if m.YinYang != "阴" {
				t.Errorf("%s(%s 天元): 四正卦天元应为阴, got %s",
					m.Name, m.Trigram, m.YinYang)
			}
		} else {
			if m.YinYang != "阳" {
				t.Errorf("%s(%s 天元): 四隅卦天元应为阳, got %s",
					m.Name, m.Trigram, m.YinYang)
			}
		}
	}
}

// =============================================================================
// 九宫 — 方位与五行
// =============================================================================

func TestPalaceByNumber_Range(t *testing.T) {
	// Valid range
	for i := 1; i <= 9; i++ {
		p := PalaceByNumber(i)
		if p.Number != i {
			t.Errorf("PalaceByNumber(%d).Number = %d", i, p.Number)
		}
		if p.Name == "" {
			t.Errorf("PalaceByNumber(%d): empty name", i)
		}
	}
	// Out of range
	for _, n := range []int{0, 10, -1, 100} {
		p := PalaceByNumber(n)
		if p.Number != 0 {
			t.Errorf("PalaceByNumber(%d) should return zero value, got number=%d", n, p.Number)
		}
	}
}

func TestPalaceTable_Directions(t *testing.T) {
	expected := map[int]string{
		1: "北", 2: "西南", 3: "东", 4: "东南", 5: "中",
		6: "西北", 7: "西", 8: "东北", 9: "南",
	}
	for i := 1; i <= 9; i++ {
		if PalaceTable[i].Direction != expected[i] {
			t.Errorf("palace %d(%s): direction=%s, want %s",
				i, PalaceTable[i].Name, PalaceTable[i].Direction, expected[i])
		}
	}
}

func TestPalaceTable_Wuxing(t *testing.T) {
	// 坎1水, 坤2土, 震3木, 巽4木, 中5土, 乾6金, 兑7金, 艮8土, 离9火
	expected := map[int]ganzhi.Wuxing{
		1: ganzhi.WxShui, 2: ganzhi.WxTu, 3: ganzhi.WxMu,
		4: ganzhi.WxMu, 5: ganzhi.WxTu, 6: ganzhi.WxJin,
		7: ganzhi.WxJin, 8: ganzhi.WxTu, 9: ganzhi.WxHuo,
	}
	for i := 1; i <= 9; i++ {
		if PalaceTable[i].Element != expected[i] {
			t.Errorf("palace %d(%s): wuxing=%s, want %s",
				i, PalaceTable[i].Name, PalaceTable[i].Element, expected[i])
		}
	}
}

// =============================================================================
// 紫白飞星 — 属性验证
// =============================================================================

func TestStarByNumber_Range(t *testing.T) {
	for i := 1; i <= 9; i++ {
		s := StarByNumber(i)
		if s.Number != i {
			t.Errorf("StarByNumber(%d).Number = %d", i, s.Number)
		}
	}
	for _, n := range []int{0, 10, -1, 100} {
		s := StarByNumber(n)
		if s.Number != 0 {
			t.Errorf("StarByNumber(%d) should return zero value, got number=%d", n, s.Number)
		}
	}
}

func TestStarTable_Colors(t *testing.T) {
	colors := map[int]string{
		1: "白", 2: "黑", 3: "碧", 4: "绿", 5: "黄",
		6: "白", 7: "赤", 8: "白", 9: "紫",
	}
	for i := 1; i <= 9; i++ {
		if StarTable[i].Color != colors[i] {
			t.Errorf("star %d(%s): color=%s, want %s",
				i, StarTable[i].Name, StarTable[i].Color, colors[i])
		}
	}
}

func TestStarTable_Wuxing(t *testing.T) {
	// 一白水, 二黑土, 三碧木, 四绿木, 五黄土, 六白金, 七赤金, 八白土, 九紫火
	expected := map[int]ganzhi.Wuxing{
		1: ganzhi.WxShui, 2: ganzhi.WxTu, 3: ganzhi.WxMu,
		4: ganzhi.WxMu, 5: ganzhi.WxTu, 6: ganzhi.WxJin,
		7: ganzhi.WxJin, 8: ganzhi.WxTu, 9: ganzhi.WxHuo,
	}
	for i := 1; i <= 9; i++ {
		if StarTable[i].Element != expected[i] {
			t.Errorf("star %d(%s): wuxing=%s, want %s",
				i, StarTable[i].Name, StarTable[i].Element, expected[i])
		}
	}
}

func TestStarTable_Auspicious(t *testing.T) {
	// 吉星: 一白,四绿,六白,八白,九紫 / 凶星: 二黑,三碧,五黄,七赤
	auspicious := map[int]bool{
		1: true, 2: false, 3: false, 4: true, 5: false,
		6: true, 7: false, 8: true, 9: true,
	}
	for i := 1; i <= 9; i++ {
		if StarTable[i].Auspicious != auspicious[i] {
			t.Errorf("star %d(%s): auspicious=%v, want %v",
				i, StarTable[i].Name, StarTable[i].Auspicious, auspicious[i])
		}
	}
}

// =============================================================================
// 洛书飞星顺序
// =============================================================================

func TestLuoshuFlyOrder_NoCenter(t *testing.T) {
	// 洛书飞星路线不含中宫(5), 顺序: 6→7→8→9→1→2→3→4
	for _, n := range LuoshuFlyOrder {
		if n == 5 || n < 1 || n > 9 {
			t.Errorf("LuoshuFlyOrder contains invalid value %d", n)
		}
	}
	// Verify all 8 non-center numbers present exactly once
	seen := make(map[int]bool)
	for _, n := range LuoshuFlyOrder {
		if seen[n] {
			t.Errorf("LuoshuFlyOrder: duplicate %d", n)
		}
		seen[n] = true
	}
	for i := 1; i <= 9; i++ {
		if i == 5 {
			continue
		}
		if !seen[i] {
			t.Errorf("LuoshuFlyOrder: missing %d", i)
		}
	}
}

// =============================================================================
// 24 山索引循环
// =============================================================================

func Test24Mountains_CircularIndex(t *testing.T) {
	// 验证山的索引循环: 壬(23)下一个是子(0)
	if Mountains24Table[23].Name != "壬" {
		t.Errorf("index 23 should be 壬, got %s", Mountains24Table[23].Name)
	}
	if Mountains24Table[0].Name != "子" {
		t.Errorf("index 0 should be 子, got %s", Mountains24Table[0].Name)
	}
}

// =============================================================================
// 24 山 — 地支山与卦山
// =============================================================================

func Test24Mountains_ZhiAndGanMountains(t *testing.T) {
	// 24山中: 12地支山 + 8天干山 + 4卦山
	zhiNames := []string{"子", "丑", "寅", "卯", "辰", "巳", "午", "未", "申", "酉", "戌", "亥"}
	ganMountains := []string{"甲", "乙", "丙", "丁", "庚", "辛", "壬", "癸"}
	guaMountains := []string{"乾", "坤", "艮", "巽"}

	type category int
	const (
		catZhi category = iota
		catGan
		catGua
	)

	classify := func(name string) category {
		for _, z := range zhiNames {
			if name == z {
				return catZhi
			}
		}
		for _, g := range ganMountains {
			if name == g {
				return catGan
			}
		}
		for _, g := range guaMountains {
			if name == g {
				return catGua
			}
		}
		return -1
	}

	catCount := map[category]int{}
	for _, m := range Mountains24Table {
		c := classify(m.Name)
		catCount[c]++
	}

	if catCount[catZhi] != 12 {
		t.Errorf("地支山 count = %d, want 12", catCount[catZhi])
	}
	if catCount[catGan] != 8 {
		t.Errorf("天干山 count = %d, want 8", catCount[catGan])
	}
	if catCount[catGua] != 4 {
		t.Errorf("卦山 count = %d, want 4", catCount[catGua])
	}
}
