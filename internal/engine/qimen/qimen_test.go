package qimen

import (
	"testing"

	"liki-engine/internal/engine/ganzhi"
)

// =============================================================================
// 地盘 (Earth Plate) — 三奇六仪排列
// =============================================================================

func TestPlaceDiPan_YangDunAll(t *testing.T) {
	tests := []struct {
		ju   int
		wuAt GongIndex
	}{
		{1, GongKan}, {2, GongKun}, {3, GongZhen},
		{4, GongXun}, {5, GongZhong}, {6, GongQian},
		{7, GongDui}, {8, GongGen}, {9, GongLi},
	}
	for _, tt := range tests {
		dipan := placeDiPan(tt.ju, false)
		pos := int(tt.wuAt) - 1
		if dipan[pos] != ganzhi.GanWu {
			t.Errorf("阳遁%d局: 戊应在%s, got %s", tt.ju, tt.wuAt, ganzhi.GanName(dipan[pos]))
		}
		allStems := map[ganzhi.Gan]bool{}
		for _, g := range dipan {
			allStems[g] = true
		}
		for _, g := range sanQiLiuYi {
			if !allStems[g] {
				t.Errorf("阳遁%d局: %s missing from dipan", tt.ju, ganzhi.GanName(g))
			}
		}
		// 验证顺排
		start := int(tt.wuAt) - 1
		for i := 0; i < 8; i++ {
			pos := (start + i) % 9
			next := (pos + 1) % 9
			if dipan[pos] != sanQiLiuYi[i] || dipan[next] != sanQiLiuYi[i+1] {
				t.Errorf("阳遁%d局: 顺排错误 pos%d=%s -> pos%d=%s",
					tt.ju, pos, ganzhi.GanName(dipan[pos]), next, ganzhi.GanName(dipan[next]))
				break
			}
		}
	}
}

func TestPlaceDiPan_YinDunAll(t *testing.T) {
	tests := []struct {
		ju   int
		wuAt GongIndex
	}{
		{1, GongKan}, {2, GongKun}, {3, GongZhen},
		{4, GongXun}, {5, GongZhong}, {6, GongQian},
		{7, GongDui}, {8, GongGen}, {9, GongLi},
	}
	for _, tt := range tests {
		dipan := placeDiPan(tt.ju, true)
		pos := int(tt.wuAt) - 1
		if dipan[pos] != ganzhi.GanWu {
			t.Errorf("阴遁%d局: 戊应在%s, got %s", tt.ju, tt.wuAt, ganzhi.GanName(dipan[pos]))
		}
		start := int(tt.wuAt) - 1
		for i := 0; i < 9; i++ {
			pos := (start - i + 9) % 9
			if dipan[pos] != sanQiLiuYi[i] {
				t.Errorf("阴遁%d局 pos%d: want %s, got %s",
					tt.ju, pos, ganzhi.GanName(sanQiLiuYi[i]), ganzhi.GanName(dipan[pos]))
			}
		}
	}
}

// =============================================================================
// 旬首与值符值使
// =============================================================================

func TestFindXunShou_AllSixtyDays(t *testing.T) {
	expectedXunShou := [6]ganzhi.Gan{
		ganzhi.GanWu, ganzhi.GanJi, ganzhi.GanGeng,
		ganzhi.GanXin, ganzhi.GanRen, ganzhi.GanGui,
	}
	for dayIdx := 0; dayIdx < 60; dayIdx++ {
		g := ganzhi.Gan(dayIdx%10 + 1)
		z := ganzhi.Zhi(dayIdx%12 + 1)
		want := expectedXunShou[dayIdx/10]
		got := findXunShou(ganzhi.Zhu{Gan: g, Zhi: z})
		if got != want {
			t.Errorf("day %d (%s%s): findXunShou = %s, want %s",
				dayIdx, ganzhi.GanName(g), ganzhi.ZhiName(z),
				ganzhi.GanName(got), ganzhi.GanName(want))
		}
	}
}

func TestFindDuty_KnownCases(t *testing.T) {
	dipan := placeDiPan(1, false) // 阳遁1局: 戊1,己2,庚3,辛4,壬5,癸6,丁7,丙8,乙9

	tests := []struct {
		name     string
		driveGan ganzhi.Gan
		driveZhi ganzhi.Zhi
		wantStar StarIndex
		wantDoor DoorIndex
	}{
		// 甲子旬(0-9): 旬首戊在坎1 → 天蓬/休门
		{"甲子", ganzhi.GanJia, ganzhi.ZhiZi, StarTianPeng, DoorXiu},
		{"乙丑", ganzhi.GanYi, ganzhi.ZhiChou, StarTianPeng, DoorXiu},
		{"癸酉", ganzhi.GanGui, ganzhi.ZhiYou, StarTianPeng, DoorXiu},
		// 甲戌旬(10-19): 旬首己在坤2 → 天芮/死门
		{"甲戌", ganzhi.GanJia, ganzhi.ZhiXu, StarTianRui, DoorSi},
		{"乙亥", ganzhi.GanYi, ganzhi.ZhiHai, StarTianRui, DoorSi},
		// 甲申旬(20-29): 旬首庚在震3 → 天冲/伤门
		{"甲申", ganzhi.GanJia, ganzhi.ZhiShen, StarTianChong, DoorShang},
		// 甲午旬(30-39): 旬首辛在巽4 → 天辅/杜门
		{"甲午", ganzhi.GanJia, ganzhi.ZhiWu, StarTianFu, DoorDu},
		// 甲辰旬(40-49): 旬首壬在中5 → 天禽/死门(寄坤)
		{"甲辰", ganzhi.GanJia, ganzhi.ZhiChen, StarTianQin, DoorSi},
		// 甲寅旬(50-59): 旬首癸在乾6 → 天心/开门
		{"甲寅", ganzhi.GanJia, ganzhi.ZhiYin, StarTianXin, DoorKai},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			d := findDuty(ganzhi.Zhu{Gan: tt.driveGan, Zhi: tt.driveZhi}, dipan)
			if d.Star != tt.wantStar {
				t.Errorf("star = %s, want %s", d.Star, tt.wantStar)
			}
			if d.Door != tt.wantDoor {
				t.Errorf("door = %s, want %s", d.Door, tt.wantDoor)
			}
		})
	}
}

// =============================================================================
// 天盘 (Heaven Plate) — 九星飞布
// =============================================================================

func TestPlaceTianPan_Yang1YiChou(t *testing.T) {
	// 阳遁1局 乙丑时: 时干乙在离9(pos8), 值符天蓬(starOrder8[0])加临之
	dipan := placeDiPan(1, false)
	duty := findDuty(ganzhi.Zhu{Gan: ganzhi.GanYi, Zhi: ganzhi.ZhiChou}, dipan)
	stars, stems := placeTianPan(ganzhi.Zhu{Gan: ganzhi.GanYi, Zhi: ganzhi.ZhiChou}, duty.Star, dipan)

	// 8 星（不含天禽，天禽寄坤2）飞布，跳过中5：
	// pos8: starOrder8[0]=天蓬(stem=dipan[0]=戊)
	// pos0: starOrder8[1]=天芮(stem=dipan[1]=己)
	// pos1: starOrder8[2]=天冲(stem=dipan[2]=庚)
	// pos2: starOrder8[3]=天辅(stem=dipan[3]=辛)
	// pos3: starOrder8[4]=天心(stem=dipan[5]=癸)
	// pos4: 中5虚空
	// pos5: starOrder8[5]=天柱(stem=dipan[6]=丁)
	// pos6: starOrder8[6]=天任(stem=dipan[7]=丙)
	// pos7: starOrder8[7]=天英(stem=dipan[8]=乙)
	expectedStars := [9]StarIndex{
		StarTianRui, StarTianChong, StarTianFu,
		StarTianXin, 0, StarTianZhu,
		StarTianRen, StarTianYing, StarTianPeng,
	}
	expectedStems := [9]ganzhi.Gan{
		ganzhi.GanJi, ganzhi.GanGeng, ganzhi.GanXin,
		ganzhi.GanGui, 0, ganzhi.GanDing,
		ganzhi.GanBing, ganzhi.GanYi, ganzhi.GanWu,
	}

	for i := 0; i < 9; i++ {
		if stars[i] != expectedStars[i] {
			t.Errorf("gong %d(%s): star = %s, want %s",
				i, GongIndex(i+1), stars[i], expectedStars[i])
		}
		if stems[i] != expectedStems[i] {
			t.Errorf("gong %d(%s): heaven stem = %s, want %s",
				i, GongIndex(i+1), ganzhi.GanName(stems[i]), ganzhi.GanName(expectedStems[i]))
		}
	}
}

func TestPlaceTianPan_Yang1BingYin(t *testing.T) {
	// 阳遁1局 丙寅时: 时干丙在艮8(pos7), 值符天蓬加临之
	dipan := placeDiPan(1, false)
	duty := findDuty(ganzhi.Zhu{Gan: ganzhi.GanBing, Zhi: ganzhi.ZhiYin}, dipan)
	stars, stems := placeTianPan(ganzhi.Zhu{Gan: ganzhi.GanBing, Zhi: ganzhi.ZhiYin}, duty.Star, dipan)

	// 8 星（不含天禽，天禽寄坤2）飞布，跳过中5：
	// pos7: starOrder8[0]=天蓬(stem=dipan[0]=戊)
	// pos8: starOrder8[1]=天芮(stem=dipan[1]=己)
	// pos0: starOrder8[2]=天冲(stem=dipan[2]=庚)
	// pos1: starOrder8[3]=天辅(stem=dipan[3]=辛)
	// pos2: starOrder8[4]=天心(stem=dipan[5]=癸)
	// pos3: starOrder8[5]=天柱(stem=dipan[6]=丁)
	// pos4: 中5虚空
	// pos5: starOrder8[6]=天任(stem=dipan[7]=丙)
	// pos6: starOrder8[7]=天英(stem=dipan[8]=乙)
	expectedStars := [9]StarIndex{
		StarTianChong, StarTianFu, StarTianXin,
		StarTianZhu, 0, StarTianRen,
		StarTianYing, StarTianPeng, StarTianRui,
	}
	expectedStems := [9]ganzhi.Gan{
		ganzhi.GanGeng, ganzhi.GanXin, ganzhi.GanGui,
		ganzhi.GanDing, 0, ganzhi.GanBing,
		ganzhi.GanYi, ganzhi.GanWu, ganzhi.GanJi,
	}

	for i := 0; i < 9; i++ {
		if stars[i] != expectedStars[i] {
			t.Errorf("gong %d(%s): star = %s, want %s",
				i, GongIndex(i+1), stars[i], expectedStars[i])
		}
		if stems[i] != expectedStems[i] {
			t.Errorf("gong %d(%s): heaven stem = %s, want %s",
				i, GongIndex(i+1), ganzhi.GanName(stems[i]), ganzhi.GanName(expectedStems[i]))
		}
	}
}

// TestPlaceTianPan_JiaDun verifies 甲遁 handling.
// When driveGan is 甲, it's not in the earth plate (甲遁于旬首).
// The correct 奇门 behavior: 甲遁于旬首仪, 值符应加临旬首所在宫.
// This test documents the current behavior vs expected.
func TestPlaceTianPan_JiaDun(t *testing.T) {
	// 阳遁1局 甲午时: driveGan=甲, 甲午旬旬首=辛(在巽4 pos3)
	// 甲遁于旬首辛, 值符天辅应加临辛所在宫(巽4 pos3).
	dipan := placeDiPan(1, false)
	duty := findDuty(ganzhi.Zhu{Gan: ganzhi.GanJia, Zhi: ganzhi.ZhiWu}, dipan)
	stars, _ := placeTianPan(ganzhi.Zhu{Gan: ganzhi.GanJia, Zhi: ganzhi.ZhiWu}, duty.Star, dipan)

	if duty.Star != StarTianFu {
		t.Fatalf("甲午旬 duty star = %s, want 天辅", duty.Star)
	}
	if stars[3] != StarTianFu {
		t.Errorf("甲遁未处理: 值符 %s 应在巽4(pos3=旬首辛位置), 实际在坎1(pos0)",
			duty.Star)
	}
	// 8 星（不含天禽，天禽寄坤2）顺时针飞布，跳过中5。
	// dutyIdx for 天辅 = 3, drivePalace = 3 (旬首辛在巽4)
	// pos3=天辅, pos5=天心, pos6=天柱, pos7=天任, pos8=天英, pos0=天蓬, pos1=天芮, pos2=天冲；中5虚空无星
	expected := [9]StarIndex{
		StarTianPeng, StarTianRui, StarTianChong, StarTianFu, 0,
		StarTianXin, StarTianZhu, StarTianRen, StarTianYing,
	}
	for i := 0; i < 9; i++ {
		if stars[i] != expected[i] {
			t.Errorf("gong %d(%s): got %s, want %s", i, GongIndex(i+1), stars[i], expected[i])
		}
	}
}

// TestPlaceTianPan_Zhong5JiKun2 中5寄坤2：时干（或旬首）落中5时，值符星应寄于坤2，
// 而非从乾6起排。阳遁1局甲辰时：旬首壬在中5→值符天禽（寄坤2随天芮），值符落坤2。
func TestPlaceTianPan_Zhong5JiKun2(t *testing.T) {
	dipan := placeDiPan(1, false) // 阳遁1局：壬在中5
	duty := findDuty(ganzhi.Zhu{Gan: ganzhi.GanJia, Zhi: ganzhi.ZhiChen}, dipan)
	if duty.Star != StarTianQin {
		t.Fatalf("甲辰时值符星 = %s, want 天禽（旬首壬在中5）", duty.Star)
	}
	stars, _ := placeTianPan(ganzhi.Zhu{Gan: ganzhi.GanJia, Zhi: ganzhi.ZhiChen}, duty.Star, dipan)
	// 值符天禽寄坤2随天芮，落坤2；8星从坤2顺排，跳过中5。
	expected := [9]StarIndex{
		StarTianPeng, StarTianRui, StarTianChong, StarTianFu, 0,
		StarTianXin, StarTianZhu, StarTianRen, StarTianYing,
	}
	for i := 0; i < 9; i++ {
		if stars[i] != expected[i] {
			t.Errorf("gong %d(%s): got %s, want %s", i, GongIndex(i+1), stars[i], expected[i])
		}
	}
}

// TestComputePan_ZhiFuShen_QinJiKun2 值符星为天禽（旬首在中5）时，
// 值符神应落值符星寄宫坤2（天禽寄坤2随天芮），而非默认坎1。
func TestComputePan_ZhiFuShen_QinJiKun2(t *testing.T) {
	ju := juShu{Number: 1, YinDun: false}
	p := computePan(ju, ganzhi.Zhu{Gan: ganzhi.GanJia, Zhi: ganzhi.ZhiChen}, 0)
	if p.DutyStar != StarTianQin {
		t.Fatalf("甲辰时值符星 = %s, want 天禽（旬首壬在中5）", p.DutyStar)
	}
	// 值符神应落坤2（值符星天禽寄坤2随天芮）。
	if p.GongWei[1].Spirit != SpiritZhiFu {
		t.Errorf("值符神应落坤2，got %s", p.GongWei[1].Spirit.YangName())
	}
}

// =============================================================================
// 人盘 (Human Plate) — 八门飞布
// =============================================================================

func TestPlaceRenPan_Yang1YiChou(t *testing.T) {
	// 阳遁1局 乙丑时: 时支丑→艮8(pos7), 值使休门(doorOrder[0])加临
	dipan := placeDiPan(1, false)
	duty := findDuty(ganzhi.Zhu{Gan: ganzhi.GanYi, Zhi: ganzhi.ZhiChou}, dipan)
	doors := placeRenPan(ganzhi.ZhiChou, duty.Door, false)

	// pos7: 休, pos8: 生, pos0: 伤, pos1: 杜, pos2: 景, pos3: 死, pos5: 惊, pos6: 开
	expected := [9]DoorIndex{
		DoorShang, DoorDu, DoorJing,
		DoorSi, 0, DoorJingMen,
		DoorKai, DoorXiu, DoorSheng,
	}

	for i := 0; i < 9; i++ {
		if doors[i] != expected[i] {
			t.Errorf("gong %d(%s): door = %s(%d), want %s(%d)",
				i, GongIndex(i+1), doors[i], doors[i], expected[i], expected[i])
		}
	}
}

func TestPlaceRenPan_Yang1JiaZi(t *testing.T) {
	// 阳遁1局 甲子时: 时支子→坎1(pos0), 值使休门加临
	dipan := placeDiPan(1, false)
	duty := findDuty(ganzhi.Zhu{Gan: ganzhi.GanJia, Zhi: ganzhi.ZhiZi}, dipan)
	doors := placeRenPan(ganzhi.ZhiZi, duty.Door, false)

	// 八门从坎1开始顺时针: 休0,生1,伤2,杜3,空4,景5,死6,惊7,开8
	expected := [9]DoorIndex{
		DoorXiu, DoorSheng, DoorShang,
		DoorDu, 0, DoorJing,
		DoorSi, DoorJingMen, DoorKai,
	}

	for i := 0; i < 9; i++ {
		if doors[i] != expected[i] {
			t.Errorf("gong %d(%s): door = %s(%d), want %s(%d)",
				i, GongIndex(i+1), doors[i], doors[i], expected[i], expected[i])
		}
	}
}

// TestPlaceRenPan_YinNi 阴遁八门逆排：值使惊门落时支午宫（离9），阴遁逆时针铺开。
func TestPlaceRenPan_YinNi(t *testing.T) {
	// 阴遁：值使惊门，时支午→离9。
	doors := placeRenPan(ganzhi.ZhiWu, DoorJingMen, true)
	// 离9=惊，逆时针：艮8开、兑7休、乾6生、巽4伤、震3杜、坤2景、坎1死。
	expected := [9]DoorIndex{
		DoorSi, DoorJing, DoorDu,
		DoorShang, 0, DoorSheng,
		DoorXiu, DoorKai, DoorJingMen,
	}
	for i := 0; i < 9; i++ {
		if doors[i] != expected[i] {
			t.Errorf("gong %d(%s): door = %s(%d), want %s(%d)（阴遁逆排）",
				i, GongIndex(i+1), doors[i], doors[i], expected[i], expected[i])
		}
	}
}

// =============================================================================
// 神盘 (Spirit Plate) — 八神飞布
// =============================================================================

func TestPlaceShenPan_YangDun(t *testing.T) {
	// 值符星在离9 → 八神从离9(pos8)顺时针排
	spirits := placeShenPan(false, GongLi)
	expected := [9]SpiritIndex{
		SpiritTengShe, SpiritTaiYin, SpiritLiuHe,
		SpiritGouChen, 0, SpiritZhuQue,
		SpiritJiuDi, SpiritJiuTian, SpiritZhiFu,
	}

	for i := 0; i < 9; i++ {
		if spirits[i] != expected[i] {
			t.Errorf("gong %d(%s): spirit = %d, want %d",
				i, GongIndex(i+1), spirits[i], expected[i])
		}
	}
}

func TestPlaceShenPan_YinDun(t *testing.T) {
	// 值符星在离9 → 八神从离9(pos8)逆时针排
	spirits := placeShenPan(true, GongLi)
	expected := [9]SpiritIndex{
		SpiritJiuTian, SpiritJiuDi, SpiritZhuQue,
		SpiritGouChen, 0, SpiritLiuHe,
		SpiritTaiYin, SpiritTengShe, SpiritZhiFu,
	}

	for i := 0; i < 9; i++ {
		if spirits[i] != expected[i] {
			t.Errorf("gong %d(%s): spirit = %d, want %d",
				i, GongIndex(i+1), spirits[i], expected[i])
		}
	}
}

func TestPlaceShenPan_KanPalace(t *testing.T) {
	// 值符星在坎1 → 验证从不同起始宫的排列
	spirits := placeShenPan(false, GongKan)
	expected := [9]SpiritIndex{
		SpiritZhiFu, SpiritTengShe, SpiritTaiYin,
		SpiritLiuHe, 0, SpiritGouChen,
		SpiritZhuQue, SpiritJiuDi, SpiritJiuTian,
	}

	for i := 0; i < 9; i++ {
		if spirits[i] != expected[i] {
			t.Errorf("gong %d(%s): spirit = %d, want %d",
				i, GongIndex(i+1), spirits[i], expected[i])
		}
	}
}

// =============================================================================
// 暗干 (Hidden Stems)
// =============================================================================

func TestPlaceAnGan_Yang1YiChou(t *testing.T) {
	// 阳遁1局 乙丑时: 时干乙, 值使休门在艮8宫(pos7)
	dipan := placeDiPan(1, false)
	duty := findDuty(ganzhi.Zhu{Gan: ganzhi.GanYi, Zhi: ganzhi.ZhiChou}, dipan)
	doors := placeRenPan(ganzhi.ZhiChou, duty.Door, false)

	dutyDoorPalace := 0
	for i, d := range doors {
		if d == duty.Door {
			dutyDoorPalace = i
			break
		}
	}

	angans := placeAnGan(ganzhi.Zhu{Gan: ganzhi.GanYi, Zhi: ganzhi.ZhiChou}, dutyDoorPalace)

	// 暗干序列=戊己庚辛壬癸丁丙乙（含癸）。时干乙在序列末(8)，从pos7(艮8)顺排8干（跳过中5）：
	// pos7=乙, pos8=戊, pos0=己, pos1=庚, pos2=辛, pos3=壬, pos5=癸, pos6=丁
	expected := [9]ganzhi.Gan{
		ganzhi.GanJi, ganzhi.GanGeng, ganzhi.GanXin,
		ganzhi.GanRen, 0, ganzhi.GanGui,
		ganzhi.GanDing, ganzhi.GanYi, ganzhi.GanWu,
	}

	for i := 0; i < 9; i++ {
		if angans[i] != expected[i] {
			t.Errorf("gong %d(%s): anGan = %s(%d), want %s(%d)",
				i, GongIndex(i+1), ganzhi.GanName(angans[i]), angans[i],
				ganzhi.GanName(expected[i]), expected[i])
		}
	}
}

// TestPlaceAnGan_XunShouLocatable 六甲旬首均能在暗干序列中定位（含癸），
// 防止甲寅旬（旬首癸）暗干起点错误。
func TestPlaceAnGan_XunShouLocatable(t *testing.T) {
	// 六甲旬首对应的六仪
	xunShouStems := []ganzhi.Gan{
		ganzhi.GanWu,  // 甲子→戊
		ganzhi.GanJi,  // 甲戌→己
		ganzhi.GanGeng, // 甲申→庚
		ganzhi.GanXin, // 甲午→辛
		ganzhi.GanRen, // 甲辰→壬
		ganzhi.GanGui, // 甲寅→癸
	}
	for _, g := range xunShouStems {
		found := false
		for _, s := range eightStems {
			if s == g {
				found = true
				break
			}
		}
		if !found {
			t.Errorf("暗干序列缺少六仪 %s（旬首不能定位）", ganzhi.GanName(g))
		}
	}
}

// =============================================================================
// 马星 (MaXing)
// =============================================================================

func TestFindMaXing_AllBranches(t *testing.T) {
	tests := []struct {
		zhi  ganzhi.Zhi
		want GongIndex
	}{
		{ganzhi.ZhiZi, GongGen},    // 子→寅(艮) — 申子辰马在寅
		{ganzhi.ZhiChou, GongQian}, // 丑→亥(乾) — 巳酉丑马在亥
		{ganzhi.ZhiYin, GongKun},   // 寅→申(坤) — 寅午戌马在申
		{ganzhi.ZhiMao, GongXun},   // 卯→巳(巽) — 亥卯未马在巳
		{ganzhi.ZhiChen, GongGen},  // 辰→寅(艮)
		{ganzhi.ZhiSi, GongQian},   // 巳→亥(乾)
		{ganzhi.ZhiWu, GongKun},    // 午→申(坤)
		{ganzhi.ZhiWei, GongXun},   // 未→巳(巽)
		{ganzhi.ZhiShen, GongGen},  // 申→寅(艮)
		{ganzhi.ZhiYou, GongQian},  // 酉→亥(乾)
		{ganzhi.ZhiXu, GongKun},    // 戌→申(坤)
		{ganzhi.ZhiHai, GongXun},   // 亥→巳(巽)
	}
	for _, tt := range tests {
		t.Run(ganzhi.ZhiName(tt.zhi), func(t *testing.T) {
			got := findMaXing(tt.zhi)
			if got != tt.want {
				t.Errorf("findMaXing(%s) = %s, want %s",
					ganzhi.ZhiName(tt.zhi), got, tt.want)
			}
		})
	}
}

// =============================================================================
// 空亡 (KongWang)
// =============================================================================

func TestFindKongWang_AllXun(t *testing.T) {
	// 甲子旬(0-9): 空戌亥 → 乾,乾
	// 甲戌旬(10-19): 空申酉 → 坤,兑
	// 甲申旬(20-29): 空午未 → 离,坤
	// 甲午旬(30-39): 空辰巳 → 巽,巽
	// 甲辰旬(40-49): 空寅卯 → 艮,震
	// 甲寅旬(50-59): 空子丑 → 坎,艮
	tests := []struct {
		gan      ganzhi.Gan
		zhi      ganzhi.Zhi
		wantPal1 GongIndex
		wantPal2 GongIndex
	}{
		{ganzhi.GanJia, ganzhi.ZhiZi, GongQian, GongQian},
		{ganzhi.GanGui, ganzhi.ZhiYou, GongQian, GongQian},
		{ganzhi.GanJia, ganzhi.ZhiXu, GongKun, GongDui},
		{ganzhi.GanJia, ganzhi.ZhiShen, GongLi, GongKun},
		{ganzhi.GanJia, ganzhi.ZhiWu, GongXun, GongXun},
		{ganzhi.GanJia, ganzhi.ZhiChen, GongGen, GongZhen},
		{ganzhi.GanJia, ganzhi.ZhiYin, GongKan, GongGen},
	}
	for _, tt := range tests {
		name := ganzhi.GanName(tt.gan) + ganzhi.ZhiName(tt.zhi)
		t.Run(name, func(t *testing.T) {
			kw := findKongWang(ganzhi.Zhu{Gan: tt.gan, Zhi: tt.zhi})
			if kw[0] != tt.wantPal1 {
				t.Errorf("kongWang[0] = %s, want %s", kw[0], tt.wantPal1)
			}
			if kw[1] != tt.wantPal2 {
				t.Errorf("kongWang[1] = %s, want %s", kw[1], tt.wantPal2)
			}
		})
	}
}

// =============================================================================
// 局数确定 (JuShu)
// =============================================================================

func TestDetermineYuan_AllPositions(t *testing.T) {
	// 三元符头（拆补法）：(idx%15)/5 → 0-4 上元、5-9 中元、10-14 下元（60 日循环 4 组）。
	for dayIdx := 0; dayIdx < 60; dayIdx++ {
		g := int(ganzhi.Gan(dayIdx%10 + 1))
		z := int(ganzhi.Zhi(dayIdx%12 + 1))
		got := determineYuan(ganzhi.Zhu{Gan: ganzhi.Gan(g), Zhi: ganzhi.Zhi(z)})

		want := (dayIdx % 15) / 5

		if got != want {
			t.Errorf("dayIdx=%d (%s%s): determineYuan=%d, want %d",
				dayIdx, ganzhi.GanName(ganzhi.Gan(g)), ganzhi.ZhiName(ganzhi.Zhi(z)), got, want)
		}
	}
}

// =============================================================================
// 完整排盘 (Full Pan) — 集成测试
// =============================================================================

func TestComputePan_Yang1YiChou(t *testing.T) {
	// 阳遁1局 乙丑时: 逐宫验证全部6层
	ju := juShu{Number: 1, YinDun: false}
	p := computePan(ju, ganzhi.Zhu{Gan: ganzhi.GanYi, Zhi: ganzhi.ZhiChou}, 0)

	if p.Jushu != 1 {
		t.Errorf("jushu = %d, want 1", p.Jushu)
	}
	if p.YinDun {
		t.Error("yinDun should be false")
	}
	if p.DutyStar != StarTianPeng {
		t.Errorf("dutyStar = %s, want 天蓬", p.DutyStar)
	}
	if p.DutyDoor != DoorXiu {
		t.Errorf("dutyDoor = %s, want 休", p.DutyDoor)
	}
	if p.MaXing != GongQian {
		t.Errorf("maXing = %s, want 乾 (丑→亥)", p.MaXing)
	}
	if p.KongWang[0] != GongQian || p.KongWang[1] != GongQian {
		t.Errorf("kongWang = [%s,%s], want [乾,乾] (乙丑→甲子旬空戌亥)", p.KongWang[0], p.KongWang[1])
	}

	type wantPalace struct {
		earth, heaven ganzhi.Gan
		star          StarIndex
		door          DoorIndex
		spirit        SpiritIndex
		hidden        ganzhi.Gan
	}
	want := [9]wantPalace{
		{ganzhi.GanWu, ganzhi.GanJi, StarTianRui, DoorShang, SpiritTengShe, ganzhi.GanJi},      // 坎1
		{ganzhi.GanJi, ganzhi.GanGeng, StarTianChong, DoorDu, SpiritTaiYin, ganzhi.GanGeng},    // 坤2
		{ganzhi.GanGeng, ganzhi.GanXin, StarTianFu, DoorJing, SpiritLiuHe, ganzhi.GanXin},      // 震3
		{ganzhi.GanXin, ganzhi.GanGui, StarTianXin, DoorSi, SpiritGouChen, ganzhi.GanRen},      // 巽4（天禽寄坤2，此宫天心）
		{ganzhi.GanRen, 0, 0, 0, 0, 0},                                                          // 中5（虚空）
		{ganzhi.GanGui, ganzhi.GanDing, StarTianZhu, DoorJingMen, SpiritZhuQue, ganzhi.GanGui}, // 乾6
		{ganzhi.GanDing, ganzhi.GanBing, StarTianRen, DoorKai, SpiritJiuDi, ganzhi.GanDing},    // 兑7
		{ganzhi.GanBing, ganzhi.GanYi, StarTianYing, DoorXiu, SpiritJiuTian, ganzhi.GanYi},     // 艮8
		{ganzhi.GanYi, ganzhi.GanWu, StarTianPeng, DoorSheng, SpiritZhiFu, ganzhi.GanWu},       // 离9
	}

	for i, w := range want {
		pal := p.GongWei[i]
		if pal.EarthStem != w.earth {
			t.Errorf("gong %d(%s) earth: %s, want %s",
				i, GongIndex(i+1), ganzhi.GanName(pal.EarthStem), ganzhi.GanName(w.earth))
		}
		if pal.HeavenStem != w.heaven {
			t.Errorf("gong %d(%s) heaven: %s, want %s",
				i, GongIndex(i+1), ganzhi.GanName(pal.HeavenStem), ganzhi.GanName(w.heaven))
		}
		if pal.Star != w.star {
			t.Errorf("gong %d(%s) star: %s, want %s",
				i, GongIndex(i+1), pal.Star, w.star)
		}
		if pal.Door != w.door {
			t.Errorf("gong %d(%s) door: %s, want %s",
				i, GongIndex(i+1), pal.Door, w.door)
		}
		if pal.Spirit != w.spirit {
			t.Errorf("gong %d(%s) spirit: %d, want %d",
				i, GongIndex(i+1), pal.Spirit, w.spirit)
		}
		if pal.HiddenStem != w.hidden {
			t.Errorf("gong %d(%s) hidden: %s(%d), want %s(%d)",
				i, GongIndex(i+1), ganzhi.GanName(pal.HiddenStem), pal.HiddenStem,
				ganzhi.GanName(w.hidden), w.hidden)
		}
	}
}

func TestComputePan_Yang1BingYin(t *testing.T) {
	// 阳遁1局 丙寅时: 丙在艮8, 寅→休门加临艮8
	ju := juShu{Number: 1, YinDun: false}
	p := computePan(ju, ganzhi.Zhu{Gan: ganzhi.GanBing, Zhi: ganzhi.ZhiYin}, 0)

	if p.DutyStar != StarTianPeng {
		t.Errorf("dutyStar = %s, want 天蓬", p.DutyStar)
	}
	if p.DutyDoor != DoorXiu {
		t.Errorf("dutyDoor = %s, want 休", p.DutyDoor)
	}
	if p.MaXing != GongKun {
		t.Errorf("maXing = %s, want 坤 (寅→申)", p.MaXing)
	}
	if p.KongWang[0] != GongQian || p.KongWang[1] != GongQian {
		t.Errorf("kongWang = [%s,%s], want [乾,乾] (丙寅→甲子旬空戌亥)", p.KongWang[0], p.KongWang[1])
	}

	type wantPalace struct {
		star   StarIndex
		door   DoorIndex
		heaven ganzhi.Gan
	}
	want := [9]wantPalace{
		{StarTianChong, DoorShang, ganzhi.GanGeng}, // 坎1
		{StarTianFu, DoorDu, ganzhi.GanXin},        // 坤2
		{StarTianXin, DoorJing, ganzhi.GanGui},     // 震3（天禽寄坤2，此宫天心）
		{StarTianZhu, DoorSi, ganzhi.GanDing},      // 巽4
		{0, 0, 0},                                  // 中5（虚空）
		{StarTianRen, DoorJingMen, ganzhi.GanBing}, // 乾6
		{StarTianYing, DoorKai, ganzhi.GanYi},      // 兑7
		{StarTianPeng, DoorXiu, ganzhi.GanWu},      // 艮8
		{StarTianRui, DoorSheng, ganzhi.GanJi},     // 离9
	}

	for i, w := range want {
		pal := p.GongWei[i]
		if pal.Star != w.star {
			t.Errorf("gong %d(%s) star: %s, want %s",
				i, GongIndex(i+1), pal.Star, w.star)
		}
		if pal.Door != w.door {
			t.Errorf("gong %d(%s) door: %s, want %s",
				i, GongIndex(i+1), pal.Door, w.door)
		}
		if pal.HeavenStem != w.heaven {
			t.Errorf("gong %d(%s) heaven: %s, want %s",
				i, GongIndex(i+1), ganzhi.GanName(pal.HeavenStem), ganzhi.GanName(w.heaven))
		}
	}
}

func TestComputePan_Yin9JiaWu(t *testing.T) {
	// 阴遁9局 甲午时
	// 地盘: 戊起离9逆排 → 戊9,己8,庚7,辛6,壬5,癸4,丁3,丙2,乙1
	// 甲午 idx=30, 旬首辛(甲午旬), 辛在乾6 → 值符天心, 值使开门
	ju := juShu{Number: 9, YinDun: true}
	p := computePan(ju, ganzhi.Zhu{Gan: ganzhi.GanJia, Zhi: ganzhi.ZhiWu}, 0)

	if p.Jushu != 9 {
		t.Errorf("jushu = %d, want 9", p.Jushu)
	}
	if !p.YinDun {
		t.Error("yinDun should be true")
	}
	if p.DutyStar != StarTianXin {
		t.Errorf("dutyStar = %s, want 天心", p.DutyStar)
	}
	if p.DutyDoor != DoorKai {
		t.Errorf("dutyDoor = %s, want 开", p.DutyDoor)
	}
	if p.MaXing != GongKun {
		t.Errorf("maXing = %s, want 坤 (午→申)", p.MaXing)
	}
	// 空亡: 甲午旬空辰巳 → 巽,巽
	if p.KongWang[0] != GongXun || p.KongWang[1] != GongXun {
		t.Errorf("kongWang = [%s,%s], want [巽,巽]", p.KongWang[0], p.KongWang[1])
	}
}

// =============================================================================
// 克应 (Stem Interactions)
// =============================================================================

func TestComputeGanInteractions_KnownPairs(t *testing.T) {
	p := pan{Jushu: 1, YinDun: false}
	// name/pattern 按标准"天盘+地盘"（X加Y）：青龙返首=戊加丙=天盘戊地盘丙=[地丙,天戊]
	p.GongWei[0] = Gong{EarthStem: ganzhi.GanBing, HeavenStem: ganzhi.GanWu}   // 戊加丙=青龙返首(吉)
	p.GongWei[1] = Gong{EarthStem: ganzhi.GanBing, HeavenStem: ganzhi.GanGeng} // 庚加丙=太白入荧(凶)
	p.GongWei[2] = Gong{EarthStem: ganzhi.GanWu, HeavenStem: ganzhi.GanBing}   // 丙加戊=飞鸟跌穴(吉)
	p.GongWei[3] = Gong{EarthStem: ganzhi.GanYi, HeavenStem: ganzhi.GanXin}    // 辛加乙=白虎猖狂(凶)
	p.GongWei[4] = Gong{EarthStem: ganzhi.GanWu, HeavenStem: ganzhi.GanRen}    // 壬加戊=小蛇化龙(吉)

	result := computeGanInteractions(p)

	if !result[0].Auspicious {
		t.Error("戊加丙 should be auspicious (青龙返首)")
	}
	if result[1].Auspicious {
		t.Error("庚加丙 should be inauspicious (太白入荧)")
	}
	if !result[2].Auspicious {
		t.Error("丙加戊 should be auspicious (飞鸟跌穴)")
	}
	if result[3].Auspicious {
		t.Error("辛加乙 should be inauspicious (白虎猖狂)")
	}
	if !result[4].Auspicious {
		t.Error("壬加戊 should be auspicious (小蛇化龙)")
	}
	if result[0].Name != "戊+丙" {
		t.Errorf("name = %q, want 戊+丙", result[0].Name)
	}
}

// =============================================================================
// 旺衰
// =============================================================================

func TestWangShuai_KnownStates(t *testing.T) {
	// 星五行入宫五行 → 旺相休囚废
	// 同→旺, 宫生星→相, 星生宫→休, 星克宫→囚, 宫克星→废
	tests := []struct {
		starElem, palElem ganzhi.Wuxing
		want              string
	}{
		{ganzhi.WxShui, ganzhi.WxShui, "旺"}, // 水入水
		{ganzhi.WxShui, ganzhi.WxJin, "相"},  // 水入金 (金生水)
		{ganzhi.WxShui, ganzhi.WxMu, "休"},   // 水入木 (水生木)
		{ganzhi.WxShui, ganzhi.WxHuo, "囚"},  // 水入火 (水克火)
		{ganzhi.WxShui, ganzhi.WxTu, "废"},   // 水入土 (土克水)
		{ganzhi.WxMu, ganzhi.WxMu, "旺"},     // 木入木
		{ganzhi.WxMu, ganzhi.WxShui, "相"},   // 木入水 (水生木)
		{ganzhi.WxMu, ganzhi.WxHuo, "休"},    // 木入火 (木生火)
		{ganzhi.WxMu, ganzhi.WxTu, "囚"},     // 木入土 (木克土)
		{ganzhi.WxMu, ganzhi.WxJin, "废"},    // 木入金 (金克木)
	}
	for _, tt := range tests {
		if s := wuxingState(tt.starElem, tt.palElem); s != tt.want {
			t.Errorf("starElem=%d palElem=%d: got %q, want %q", tt.starElem, tt.palElem, s, tt.want)
		}
	}
}

// =============================================================================
// 门迫/门制 (MenPo/MenZhi)
// =============================================================================

func TestMenPo(t *testing.T) {
	// 门迫 = 门克宫
	if !menPo(DoorXiu, GongLi) {
		t.Error("休门(水)+离宫(火) should be 门迫 (水克火)")
	}
	if !menPo(DoorKai, GongZhen) {
		t.Error("开门(金)+震宫(木) should be 门迫 (金克木)")
	}
	if !menPo(DoorSi, GongKan) {
		t.Error("死门(土)+坎宫(水) should be 门迫 (土克水)")
	}
	// 非门迫
	if menPo(DoorXiu, GongKun) {
		t.Error("休门(水)+坤宫(土) should NOT be 门迫 (土克水=门制)")
	}
	if menPo(DoorKai, GongQian) {
		t.Error("开门(金)+乾宫(金) should NOT be 门迫 (比和)")
	}
}

func TestMenZhi(t *testing.T) {
	// 门制 = 宫克门
	if !menZhi(DoorXiu, GongKun) {
		t.Error("休门(水)+坤宫(土) should be 门制 (土克水)")
	}
	if !menZhi(DoorSheng, GongZhen) {
		t.Error("生门(土)+震宫(木) should be 门制 (木克土)")
	}
	if !menZhi(DoorKai, GongLi) {
		t.Error("开门(金)+离宫(火) should be 门制 (火克金)")
	}
	if menZhi(DoorXiu, GongKan) {
		t.Error("休门(水)+坎宫(水) should NOT be 门制 (比和)")
	}
}

// =============================================================================
// 六甲旬首 → 六仪 (LiuJiaLiuYi consistency)
// =============================================================================

func TestLiuJiaLiuYi_Consistency(t *testing.T) {
	liuJiaZhi := [6]ganzhi.Zhi{ganzhi.ZhiZi, ganzhi.ZhiXu, ganzhi.ZhiShen, ganzhi.ZhiWu, ganzhi.ZhiChen, ganzhi.ZhiYin}
	for i, z := range liuJiaZhi {
		xunShou := liuJiaLiuYi[i]
		got := findXunShou(ganzhi.Zhu{Gan: ganzhi.GanJia, Zhi: z})
		if got != xunShou {
			t.Errorf("甲%s旬: 旬首仪 = %s, want %s", ganzhi.ZhiName(z), ganzhi.GanName(got), ganzhi.GanName(xunShou))
		}
	}
}

// =============================================================================
// Gong utility consistency
// =============================================================================

func TestZhiPalace_Bidirectional(t *testing.T) {
	allZhi := []ganzhi.Zhi{
		ganzhi.ZhiZi, ganzhi.ZhiChou, ganzhi.ZhiYin, ganzhi.ZhiMao,
		ganzhi.ZhiChen, ganzhi.ZhiSi, ganzhi.ZhiWu, ganzhi.ZhiWei,
		ganzhi.ZhiShen, ganzhi.ZhiYou, ganzhi.ZhiXu, ganzhi.ZhiHai,
	}
	for _, z := range allZhi {
		pal := zhiPalace(z)
		principalZhi := palaceZhi(pal)
		pal2 := zhiPalace(principalZhi)
		if pal != pal2 {
			t.Errorf("zhiPalace(%s)=%s, palaceZhi(%s)=%s, zhiPalace(%s)=%s — mismatch",
				ganzhi.ZhiName(z), pal, pal, ganzhi.ZhiName(principalZhi), ganzhi.ZhiName(principalZhi), pal2)
		}
	}
}

// TestZhiPalace_All 地支→宫位直接锚定（独立于 palaceZhi，非往返自证）。
func TestZhiPalace_All(t *testing.T) {
	cases := []struct {
		zhi  ganzhi.Zhi
		want GongIndex
	}{
		{ganzhi.ZhiZi, GongKan},    // 子→坎
		{ganzhi.ZhiChou, GongGen},  // 丑→艮
		{ganzhi.ZhiYin, GongGen},   // 寅→艮
		{ganzhi.ZhiMao, GongZhen},  // 卯→震
		{ganzhi.ZhiChen, GongXun},  // 辰→巽
		{ganzhi.ZhiSi, GongXun},    // 巳→巽
		{ganzhi.ZhiWu, GongLi},     // 午→离
		{ganzhi.ZhiWei, GongKun},   // 未→坤
		{ganzhi.ZhiShen, GongKun},  // 申→坤
		{ganzhi.ZhiYou, GongDui},   // 酉→兑
		{ganzhi.ZhiXu, GongQian},   // 戌→乾
		{ganzhi.ZhiHai, GongQian},  // 亥→乾
	}
	for _, c := range cases {
		if got := zhiPalace(c.zhi); got != c.want {
			t.Errorf("zhiPalace(%s) = %s, want %s", ganzhi.ZhiName(c.zhi), got, c.want)
		}
	}
}

func TestStarHomePalace_RoundTrip(t *testing.T) {
	for i := 0; i < 9; i++ {
		star := palaceStar[i]
		home := starHomePalace(star)
		if home != i {
			t.Errorf("starHomePalace(%s) = %d, want %d", star, home, i)
		}
	}
}

// =============================================================================
// Stringers
// =============================================================================

func TestPalaceIndex_String(t *testing.T) {
	tests := []struct {
		p    GongIndex
		want string
	}{
		{GongKan, "坎"}, {GongKun, "坤"}, {GongZhen, "震"},
		{GongXun, "巽"}, {GongZhong, "中"}, {GongQian, "乾"},
		{GongDui, "兑"}, {GongGen, "艮"}, {GongLi, "离"},
		{GongIndex(0), "?"}, {GongIndex(10), "?"}, {GongIndex(-1), "?"},
	}
	for _, tt := range tests {
		if got := tt.p.String(); got != tt.want {
			t.Errorf("GongIndex(%d).String() = %s, want %s", tt.p, got, tt.want)
		}
	}
}

func TestDoorIndex_String(t *testing.T) {
	tests := []struct {
		d    DoorIndex
		want string
	}{
		{DoorXiu, "休"}, {DoorSheng, "生"}, {DoorShang, "伤"}, {DoorDu, "杜"},
		{DoorJing, "景"}, {DoorSi, "死"}, {DoorJingMen, "惊"}, {DoorKai, "开"},
		{DoorIndex(0), "?"}, {DoorIndex(9), "?"}, {DoorIndex(-1), "?"},
	}
	for _, tt := range tests {
		if got := tt.d.String(); got != tt.want {
			t.Errorf("DoorIndex(%d).String() = %s, want %s", tt.d, got, tt.want)
		}
	}
}

func TestStarIndex_String(t *testing.T) {
	if got := StarTianPeng.String(); got != "天蓬" {
		t.Errorf("StarTianPeng.String() = %s, want 天蓬", got)
	}
	if got := StarIndex(0).String(); got != "?" {
		t.Errorf("StarIndex(0).String() = %s, want ?", got)
	}
	if got := StarIndex(10).String(); got != "?" {
		t.Errorf("StarIndex(10).String() = %s, want ?", got)
	}
}

func TestSpiritIndex_YangName(t *testing.T) {
	tests := []struct {
		s    SpiritIndex
		want string
	}{
		{SpiritZhiFu, "值符"}, {SpiritTengShe, "螣蛇"}, {SpiritTaiYin, "太阴"},
		{SpiritLiuHe, "六合"}, {SpiritGouChen, "勾陈"}, {SpiritZhuQue, "朱雀"},
		{SpiritJiuDi, "九地"}, {SpiritJiuTian, "九天"},
		{SpiritIndex(0), "?"}, {SpiritIndex(9), "?"},
	}
	for _, tt := range tests {
		if got := tt.s.YangName(); got != tt.want {
			t.Errorf("SpiritIndex(%d).YangName() = %s, want %s", tt.s, got, tt.want)
		}
	}
}

func TestSpiritIndex_YinName(t *testing.T) {
	tests := []struct {
		s    SpiritIndex
		want string
	}{
		{SpiritZhiFu, "值符"}, {SpiritTengShe, "螣蛇"}, {SpiritTaiYin, "太阴"},
		{SpiritLiuHe, "六合"}, {SpiritGouChen, "白虎"}, {SpiritZhuQue, "玄武"},
		{SpiritJiuDi, "九地"}, {SpiritJiuTian, "九天"},
		{SpiritIndex(0), "?"}, {SpiritIndex(9), "?"},
	}
	for _, tt := range tests {
		if got := tt.s.YinName(); got != tt.want {
			t.Errorf("SpiritIndex(%d).YinName() = %s, want %s", tt.s, got, tt.want)
		}
	}
}

// =============================================================================
// computeMenInteractions / doorAuspicious / findMenPo / findMenZhi / doorWuxing
// =============================================================================

func buildSamplePan() pan {
	var p pan
	p.Jushu = 1
	p.YinDun = false
	p.DutyStar = StarTianPeng
	p.DutyDoor = DoorXiu
	p.MaXing = GongQian
	p.DriveGan = ganzhi.GanYi
	p.DriveZhi = ganzhi.ZhiChou
	p.KongWang = [2]GongIndex{GongQian, GongQian}
	// Fill palaces with known data (阳遁1局 乙丑时).
	p.GongWei = [9]Gong{
		{EarthStem: ganzhi.GanWu, HeavenStem: ganzhi.GanJi, Star: StarTianRui, Door: DoorShang, Spirit: SpiritTengShe},
		{EarthStem: ganzhi.GanJi, HeavenStem: ganzhi.GanGeng, Star: StarTianChong, Door: DoorDu, Spirit: SpiritTaiYin},
		{EarthStem: ganzhi.GanGeng, HeavenStem: ganzhi.GanXin, Star: StarTianFu, Door: DoorJing, Spirit: SpiritLiuHe},
		{EarthStem: ganzhi.GanXin, HeavenStem: ganzhi.GanRen, Star: StarTianQin, Door: DoorSi, Spirit: SpiritGouChen},
		{EarthStem: ganzhi.GanRen, HeavenStem: ganzhi.GanGui, Star: StarTianXin, Door: 0, Spirit: 0}, // 中5
		{EarthStem: ganzhi.GanGui, HeavenStem: ganzhi.GanDing, Star: StarTianZhu, Door: DoorJingMen, Spirit: SpiritZhuQue},
		{EarthStem: ganzhi.GanDing, HeavenStem: ganzhi.GanBing, Star: StarTianRen, Door: DoorKai, Spirit: SpiritJiuDi},
		{EarthStem: ganzhi.GanBing, HeavenStem: ganzhi.GanYi, Star: StarTianYing, Door: DoorXiu, Spirit: SpiritJiuTian},
		{EarthStem: ganzhi.GanYi, HeavenStem: ganzhi.GanWu, Star: StarTianPeng, Door: DoorSheng, Spirit: SpiritZhiFu},
	}
	return p
}

func TestComputeMenInteractions(t *testing.T) {
	p := buildSamplePan() // 阳遁1局 乙丑时
	result := computeMenInteractions(p)

	// 独立锚定：门互克表条目（buildSamplePan 的门分布）。
	// 宫1伤门加坎、宫2杜门加坤、宫6惊门加乾、宫7开门加兑、宫8休门加艮、宫9生门加离。
	anchors := []struct {
		idx   int
		door  string
		name  string
		meaning string
	}{
		{0, "伤门", "伤门加坎", "道路之伤，水厄之灾"},
		{1, "杜门", "杜门加坤", "土木之阻，田宅有损"},
		{5, "惊门", "惊门加乾", "尊长不安，口舌失财"},
		{6, "开门", "开门加兑", "说合之利，口舌得财"},
		{7, "休门", "休门加艮", "求财有财，出行有喜"},
		{8, "生门", "生门加离", "文书之喜，或火食之喜"},
	}
	for _, a := range anchors {
		r := result[a.idx]
		if r.Name != a.name {
			t.Errorf("宫%d(%s) 门互克名 = %q, want %q", a.idx+1, GongIndex(a.idx+1), r.Name, a.name)
		}
		if r.Meaning != a.meaning {
			t.Errorf("宫%d(%s) 门互克意 = %q, want %q", a.idx+1, GongIndex(a.idx+1), r.Meaning, a.meaning)
		}
	}
}

func TestDoorAuspicious(t *testing.T) {
	tests := []struct {
		d    DoorIndex
		want string
	}{
		{DoorXiu, "吉门得地，谋事可成"},
		{DoorSheng, "吉门得地，谋事可成"},
		{DoorKai, "吉门得地，谋事可成"},
		{DoorDu, "中平之门，需择时而行"},
		{DoorJing, "中平之门，需择时而行"},
		{DoorShang, "凶门当位，行事多阻"},
		{DoorSi, "凶门当位，行事多阻"},
		{DoorJingMen, "凶门当位，行事多阻"},
		{DoorIndex(0), ""},
		{DoorIndex(9), ""},
	}
	for _, tt := range tests {
		if got := doorAuspicious(tt.d); got != tt.want {
			t.Errorf("doorAuspicious(%d) = %s, want %s", tt.d, got, tt.want)
		}
	}
}

func TestDoorWuxing(t *testing.T) {
	tests := []struct {
		d    DoorIndex
		want ganzhi.Wuxing
	}{
		{DoorXiu, ganzhi.WxShui},
		{DoorSheng, ganzhi.WxTu}, {DoorSi, ganzhi.WxTu},
		{DoorShang, ganzhi.WxMu}, {DoorDu, ganzhi.WxMu},
		{DoorJing, ganzhi.WxHuo},
		{DoorJingMen, ganzhi.WxJin}, {DoorKai, ganzhi.WxJin},
		{DoorIndex(0), 0}, {DoorIndex(9), 0},
	}
	for _, tt := range tests {
		if got := doorWuxing(tt.d); got != tt.want {
			t.Errorf("doorWuxing(%d) = %d, want %d", tt.d, got, tt.want)
		}
	}
}

func TestFindMenPo(t *testing.T) {
	// 休门(水)在离(火)=门迫(水克火), 死门(土)在坎(水)=门迫
	p := pan{}
	p.GongWei[0] = Gong{Door: DoorShang} // 震宫, 伤门(木) — 比和, 不迫
	p.GongWei[8] = Gong{Door: DoorXiu}   // 离宫, 休门(水) — 水克火=迫
	p.GongWei[1] = Gong{Door: DoorSi}    // 坤宫, 死门(土) — 比和
	// 死门(土)在坎(水)=土克水=迫
	p.GongWei[0] = Gong{Door: DoorSi, EarthStem: ganzhi.GanWu} // pos 0=坎

	result := findMenPo(p)
	// 门迫宫位应含坎1（死门土克坎水）与离9（休门水克离火）。
	hasKan, hasLi := false, false
	for _, g := range result {
		if g == GongKan {
			hasKan = true
		}
		if g == GongLi {
			hasLi = true
		}
	}
	if !hasKan {
		t.Error("门迫应含坎1（死门土克坎水）")
	}
	if !hasLi {
		t.Error("门迫应含离9（休门水克离火）")
	}
}

func TestFindMenZhi(t *testing.T) {
	// 休门(水)在坤(土)=门制(土克水)
	p := pan{}
	p.GongWei[1] = Gong{Door: DoorXiu} // pos1=坤, 休门(水) — 土克水=制
	// 生门(土)在震(木)=门制(木克土) — pos2=震
	p.GongWei[2] = Gong{Door: DoorSheng}

	result := findMenZhi(p)
	// 门制宫位应含坤2（休门水被坤土制）与震3（生门土被震木制）。
	hasKun, hasZhen := false, false
	for _, g := range result {
		if g == GongKun {
			hasKun = true
		}
		if g == GongZhen {
			hasZhen = true
		}
	}
	if !hasKun {
		t.Error("门制应含坤2（休门水被坤土制）")
	}
	if !hasZhen {
		t.Error("门制应含震3（生门土被震木制）")
	}
}

// =============================================================================
// computeXingInteractions / starNature / isAuspiciousStar / starWuxing
// =============================================================================

func TestComputeXingInteractions(t *testing.T) {
	p := buildSamplePan() // 阳遁1局 乙丑时
	result := computeXingInteractions(p)

	// 独立锚定：星入宫五行克应。
	// 宫1坎天芮(土)入水宫、宫2坤天冲(木)入土宫、宫3震天辅(木)入木宫、宫9离天蓬(水)入火宫。
	anchors := []struct {
		idx   int
		name  string
		meaning string
	}{
		{0, "土星入水宫", "土能克水，事宜稳重"},
		{1, "木星入土宫", "木能克土，宜田宅之事"},
		{2, "木星入木宫", "兄弟同心，合作得利"},
		{8, "水星入火宫", "水火相激，多事之秋"},
	}
	for _, a := range anchors {
		r := result[a.idx]
		if r.Name != a.name {
			t.Errorf("宫%d(%s) 星互克名 = %q, want %q", a.idx+1, GongIndex(a.idx+1), r.Name, a.name)
		}
		if r.Meaning != a.meaning {
			t.Errorf("宫%d(%s) 星互克意 = %q, want %q", a.idx+1, GongIndex(a.idx+1), r.Meaning, a.meaning)
		}
	}
}

func TestStarNature(t *testing.T) {
	tests := []struct {
		s    StarIndex
		want string
	}{
		{StarTianPeng, "水性之精"}, {StarTianRui, "土性之精"},
		{StarTianChong, "木性之精"}, {StarTianFu, "木性文明"},
		{StarTianQin, "土性中和"}, {StarTianXin, "金性肃杀"},
		{StarTianZhu, "金性锐利"}, {StarTianRen, "土性厚重"},
		{StarTianYing, "火性光明"}, {StarIndex(0), ""}, {StarIndex(10), ""},
	}
	for _, tt := range tests {
		if got := starNature(tt.s); got != tt.want {
			t.Errorf("starNature(%d) = %s, want %s", tt.s, got, tt.want)
		}
	}
}

func TestIsAuspiciousStar(t *testing.T) {
	auspicious := []StarIndex{StarTianFu, StarTianQin, StarTianXin, StarTianRen}
	for _, s := range []StarIndex{StarTianPeng, StarTianRui, StarTianChong, StarTianZhu, StarTianYing} {
		if isAuspiciousStar(s) {
			t.Errorf("%s should not be auspicious", s)
		}
	}
	for _, s := range auspicious {
		if !isAuspiciousStar(s) {
			t.Errorf("%s should be auspicious", s)
		}
	}
}

func TestStarWuxing(t *testing.T) {
	tests := []struct {
		s    StarIndex
		want ganzhi.Wuxing
	}{
		{StarTianPeng, ganzhi.WxShui},
		{StarTianRui, ganzhi.WxTu}, {StarTianQin, ganzhi.WxTu}, {StarTianRen, ganzhi.WxTu},
		{StarTianChong, ganzhi.WxMu}, {StarTianFu, ganzhi.WxMu},
		{StarTianXin, ganzhi.WxJin}, {StarTianZhu, ganzhi.WxJin},
		{StarTianYing, ganzhi.WxHuo},
		{StarIndex(0), 0}, {StarIndex(10), 0},
	}
	for _, tt := range tests {
		if got := starWuxing(tt.s); got != tt.want {
			t.Errorf("starWuxing(%d) = %d, want %d", tt.s, got, tt.want)
		}
	}
}

// =============================================================================
// computeWangShuai
// =============================================================================

func TestComputeWangShuai(t *testing.T) {
	p := buildSamplePan()
	result := computeWangShuai(p)

	for i := 0; i < 9; i++ {
		if p.GongWei[i].Star != 0 && result[i].State == "" {
			t.Errorf("gong %d: star present but no wangshuai state", i)
		}
	}
}

// =============================================================================
// findPatterns + helpers (dutyDoorPalace, hasStem, hasStemAtPalace, hasDoor, hasSpirit)
// =============================================================================

func TestDutyDoorPalace(t *testing.T) {
	p := buildSamplePan()
	// DutyDoor is 休门, which sits at pos7 (艮8→index 7)
	got := dutyDoorPalace(p)
	if got != GongGen { // pos7+1=8=艮
		t.Errorf("dutyDoorPalace = %s, want 艮", got)
	}
}

func TestHasStem(t *testing.T) {
	p := buildSamplePan()
	if !hasStem(p, ganzhi.GanWu) {
		t.Error("戊 should be present (earth坎, heaven离)")
	}
	if hasStem(p, 0) {
		t.Error("invalid stem 0 should not be found")
	}
}

func TestHasStemAtPalace(t *testing.T) {
	p := buildSamplePan()
	if !hasStemAtPalace(p, ganzhi.GanWu, GongKan) {
		t.Error("戊 should be at 坎宫 (earth stem)")
	}
	if !hasStemAtPalace(p, ganzhi.GanWu, GongLi) {
		t.Log("戊 also appears as heaven stem at 离宫")
	}
	if hasStemAtPalace(p, ganzhi.GanWu, GongIndex(0)) {
		t.Error("should return false for invalid gong 0")
	}
	if hasStemAtPalace(p, ganzhi.GanWu, GongIndex(10)) {
		t.Error("should return false for invalid gong 10")
	}
}

func TestHasDoor(t *testing.T) {
	p := buildSamplePan()
	if !hasDoor(p, DoorXiu) {
		t.Error("休门 should be present")
	}
	if !hasDoor(p, DoorShang) {
		t.Error("伤门 should be present")
	}
	// Door 0 IS present (中5宫 has no door→value=0), so don't test false for 0.
	// Test with a door that's definitely not present.
	allPresent := map[DoorIndex]bool{}
	for _, pp := range p.GongWei {
		if pp.Door != 0 {
			allPresent[pp.Door] = true
		}
	}
	if hasDoor(p, DoorKai) != allPresent[DoorKai] {
		t.Error("hasDoor for 开 inconsistent")
	}
}

func TestHasSpirit(t *testing.T) {
	p := buildSamplePan()
	if !hasSpirit(p, SpiritZhiFu) {
		t.Error("值符 should be present")
	}
	if !hasSpirit(p, SpiritTengShe) {
		t.Error("螣蛇 should be present")
	}
	// Test "not found": create a pan without a specific spirit.
	empty := pan{}
	if hasSpirit(empty, SpiritZhiFu) {
		t.Error("值符 should not be in empty pan")
	}
}

// TestPalaceZhi_All 宫→主支直接锚定（独立于 zhiPalace，非往返自证）。
func TestPalaceZhi_All(t *testing.T) {
	cases := []struct {
		gong GongIndex
		want ganzhi.Zhi
	}{
		{GongKan, ganzhi.ZhiZi},  // 坎→子
		{GongKun, ganzhi.ZhiWei}, // 坤→未
		{GongZhen, ganzhi.ZhiMao}, // 震→卯
		{GongXun, ganzhi.ZhiSi},  // 巽→巳
		{GongQian, ganzhi.ZhiXu}, // 乾→戌
		{GongDui, ganzhi.ZhiYou}, // 兑→酉
		{GongGen, ganzhi.ZhiYin}, // 艮→寅
		{GongLi, ganzhi.ZhiWu},   // 离→午
	}
	for _, c := range cases {
		if got := palaceZhi(c.gong); got != c.want {
			t.Errorf("palaceZhi(%s) = %s, want %s", c.gong, ganzhi.ZhiName(got), ganzhi.ZhiName(c.want))
		}
	}
}

func TestPalaceZhi_Invalid(t *testing.T) {
	if got := palaceZhi(0); got != ganzhi.ZhiZi {
		t.Errorf("palaceZhi(0) = %s, want 子", ganzhi.ZhiName(got))
	}
	if got := palaceZhi(10); got != ganzhi.ZhiZi {
		t.Errorf("palaceZhi(10) = %s, want 子", ganzhi.ZhiName(got))
	}
}

func TestFindPatterns(t *testing.T) {
	// Use a pan with specific conditions:
	// 丙 in heaven stem, 生门, 丁 → 天遁
	p := pan{
		DutyStar: StarTianPeng,
		DutyDoor: DoorXiu,
		GongWei: [9]Gong{
			{HeavenStem: ganzhi.GanBing, Door: DoorSheng, Spirit: SpiritZhiFu}, // 天遁条件: 丙
			{HeavenStem: ganzhi.GanDing},                                       // 天遁也需要丁
			{}, {}, {}, {}, {}, {}, {},
		},
	}
	patterns := findPatterns(p)
	// Should find 天遁 (丙+生门+丁)
	found := false
	for _, pt := range patterns {
		if pt.Name == "天遁" {
			found = true
			break
		}
	}
	if !found {
		t.Error("天遁 pattern 未触发（pan 含丙+生门+丁，应触发）")
	}
}

func TestFindPatterns_FuYin(t *testing.T) {
	// 伏吟: 值符归位 (duty star in its home gong)
	p := pan{
		DutyStar: StarTianPeng, // home=坎(pos0)
		GongWei: [9]Gong{
			{Star: StarTianPeng}, // in坎(pos0) — 归位
			{}, {}, {}, {}, {}, {}, {}, {},
		},
	}
	patterns := findPatterns(p)
	found := false
	for _, pt := range patterns {
		if pt.Name == "伏吟" {
			found = true
			break
		}
	}
	if !found {
		t.Error("expected pattern '伏吟' not found")
	}
}

// TestFindPatterns_FuYin_Qin 值符星为天禽（中5寄坤2）时，伏吟判断按天芮（天禽寄坤2随天芮）。
// 天芮在坤2本位→伏吟。
func TestFindPatterns_FuYin_Qin(t *testing.T) {
	p := pan{
		DutyStar: StarTianQin, // 天禽寄坤2随天芮
		GongWei: [9]Gong{
			{},                            // 坎1(pos0)
			{Star: StarTianRui},            // 坤2(pos1) 天芮本位=归位
			{}, {}, {}, {}, {}, {}, {},
		},
	}
	patterns := findPatterns(p)
	found := false
	for _, pt := range patterns {
		if pt.Name == "伏吟" {
			found = true
			break
		}
	}
	if !found {
		t.Error("expected pattern '伏吟' not found（值符天禽按天芮判断本位坤2）")
	}
}

// =============================================================================
// computeYingQi / chongBranch
// =============================================================================

func TestComputeYingQi(t *testing.T) {
	p := buildSamplePan() // 阳遁1局 乙丑时
	yq := computeYingQi(p)

	// 具体文案锚定：丑→巳酉丑马在亥（冲巳）；乙丑甲子旬空戌亥。
	if yq.MaXing != "马星在亥，冲则动，应期在巳（年月日时）" {
		t.Errorf("MaXing = %q, want 马星在亥（丑→巳酉丑马在亥）", yq.MaXing)
	}
	if yq.KongWang != "空亡在戌 亥，填实或冲空之时应事" {
		t.Errorf("KongWang = %q, want 空亡在戌亥（乙丑甲子旬）", yq.KongWang)
	}
	if yq.DutyMove == "" {
		t.Error("DutyMove is empty")
	}
	if yq.Summary == "" {
		t.Error("Summary is empty")
	}
}

func TestChongBranch(t *testing.T) {
	tests := []struct {
		z    ganzhi.Zhi
		want ganzhi.Zhi
	}{
		{ganzhi.ZhiZi, ganzhi.ZhiWu},
		{ganzhi.ZhiWu, ganzhi.ZhiZi},
		{ganzhi.ZhiChou, ganzhi.ZhiWei},
		{ganzhi.ZhiYin, ganzhi.ZhiShen},
		{ganzhi.ZhiMao, ganzhi.ZhiYou},
		{ganzhi.ZhiChen, ganzhi.ZhiXu},
		{ganzhi.ZhiSi, ganzhi.ZhiHai},
		{ganzhi.ZhiHai, ganzhi.ZhiSi},
	}
	for _, tt := range tests {
		got := chongBranch(tt.z)
		if got != tt.want {
			t.Errorf("chongBranch(%s) = %s, want %s",
				ganzhi.ZhiName(tt.z), ganzhi.ZhiName(got), ganzhi.ZhiName(tt.want))
		}
	}
}

// =============================================================================
// genericGanInteraction — all branches
// =============================================================================

func TestGenericGanInteraction_AllRelations(t *testing.T) {
	// 己(土)+甲(木): 土 earth, 木 heaven
	// 木克土: heaven overcomes earth → 上克下, auspicious=false
	got := genericGanInteraction(ganzhi.GanJi, ganzhi.GanJia) // earth=己(土), heaven=甲(木)
	if got.Name != "甲+己" {
		t.Errorf("Name = %s, want 甲+己（天盘+地盘）", got.Name)
	}
	// 甲(木)+己(土): 木 earth, 土 heaven
	// 木克土: earth overcomes heaven → 下克上, auspicious=true
	got2 := genericGanInteraction(ganzhi.GanJia, ganzhi.GanJi)
	if !got2.Auspicious {
		t.Error("甲+己 (木克土, 下克上) should be auspicious")
	}

	// 甲(木)+乙(木): 比和
	got3 := genericGanInteraction(ganzhi.GanJia, ganzhi.GanYi)
	if got3.Auspicious {
		t.Error("甲+乙 (比和) should NOT be auspicious")
	}
	if got3.Meaning != "比和，静守为宜" {
		t.Errorf("meaning = %s, want 比和，静守为宜", got3.Meaning)
	}

	// 甲(木)+壬(水): 壬水(heaven)生甲木(earth) → 上生下, auspicious=true
	got4 := genericGanInteraction(ganzhi.GanJia, ganzhi.GanRen)
	if !got4.Auspicious {
		t.Error("甲+壬 (上生下) should be auspicious")
	}

	// 甲(木)+丙(火): 甲木(earth)生丙火(heaven) → 下生上, auspicious=false
	got5 := genericGanInteraction(ganzhi.GanJia, ganzhi.GanBing)
	if got5.Auspicious {
		t.Error("甲+丙 (下生上) should NOT be auspicious")
	}
	if got5.Meaning != "下生上，耗损有忧" {
		t.Errorf("meaning = %s, want 下生上，耗损有忧", got5.Meaning)
	}
}

func TestComputeXingInteractions_Generic(t *testing.T) {
	// Test star-gong pair NOT in starPalaceTable (triggers generic).
	p := pan{GongWei: [9]Gong{
		{Star: StarTianPeng}, // gong 0 (坎) — 天蓬 in 坎 IS in table
	}}
	// 天蓬 in 巽 (gong index 3, not in table) → generic
	p2 := pan{GongWei: [9]Gong{
		{}, {}, {}, {Star: StarTianPeng}, // pos3=巽, not in table → generic
	}}
	result := computeXingInteractions(p2)
	if result[3].Name == "" {
		t.Error("generic star interaction should have a name")
	}
	if result[3].Name != "天蓬加巽" {
		t.Errorf("generic name = %s, want 天蓬加巽", result[3].Name)
	}
	// 天蓬在坎(pos0): known entry → should produce interaction name.
	known := computeXingInteractions(p)
	if known[0].Name == "" {
		t.Error("gong 0 (天蓬加坎): known entry should have interaction name")
	}
}

func TestComputeWangShuai_Full(t *testing.T) {
	// Ensure all 5 states are exercised across different star-gong combos.
	// Use various stars in various palaces.
	tests := []struct {
		name  string
		star  StarIndex
		palIx int // 0-based gong index
		want  string
	}{
		{"天蓬入坎(同=旺)", StarTianPeng, 0, "旺"}, // 水入水
		{"天辅入坎(水=相)", StarTianFu, 0, "相"},   // 木入水 (水生木=相→star gets相)
		// Wait: 天辅(木) in 坎(水): sw=mu(1), pw=水(5)
		// starElem=1, palElem=5
		// 1==5? No. 1==(5%5)+1=1? Yes → 相!
		{"天蓬入震(水=生)", StarTianPeng, 2, "休"}, // 水入木 (水生木=休)
		{"天芮入坎(土=囚)", StarTianRui, 0, "囚"},  // 土入水 (土克水=囚)
		{"天英入坎(火=废)", StarTianYing, 0, "废"}, // 火入水 (水克火=废)
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			sw := starWuxing(tt.star)
			pw := palaceWuxing(GongIndex(tt.palIx + 1))
			got := wuxingState(sw, pw)
			if got != tt.want {
				t.Errorf("starWuxing=%d palaceWuxing=%d: wuxingState = %s, want %s",
					sw, pw, got, tt.want)
			}
		})
	}
}

// =============================================================================
// determineJuShu
// =============================================================================

func TestDetermineJuShu_KnownDates(t *testing.T) {
	// 2024年6月22日 (夏至之后，2024年夏至为6月21日)
	ju := determineJuShu(2024, 6, 22, ganzhi.GanBing, ganzhi.ZhiWu)
	if ju.Number < 1 || ju.Number > 9 {
		t.Errorf("jushu out of range: %d", ju.Number)
	}
	if ju.Yuan == "" {
		t.Error("Yuan is empty")
	}

	// 冬至
	ju2 := determineJuShu(2024, 12, 22, ganzhi.GanJia, ganzhi.ZhiZi)
	if ju2.YinDun == ju.YinDun {
		t.Errorf("冬至 and 夏至 should have different yin/yang: %v vs %v", ju.YinDun, ju2.YinDun)
	}
}

// =============================================================================
// palaceWuxing — edge cases
// =============================================================================

func TestPalaceWuxing_All(t *testing.T) {
	tests := []struct {
		p    GongIndex
		want ganzhi.Wuxing
	}{
		{GongKan, ganzhi.WxShui},
		{GongKun, ganzhi.WxTu}, {GongZhong, ganzhi.WxTu}, {GongGen, ganzhi.WxTu},
		{GongZhen, ganzhi.WxMu}, {GongXun, ganzhi.WxMu},
		{GongQian, ganzhi.WxJin}, {GongDui, ganzhi.WxJin},
		{GongLi, ganzhi.WxHuo},
		{GongIndex(0), ganzhi.WxTu}, {GongIndex(10), ganzhi.WxTu},
	}
	for _, tt := range tests {
		if got := palaceWuxing(tt.p); got != tt.want {
			t.Errorf("palaceWuxing(%d) = %d, want %d", tt.p, got, tt.want)
		}
	}
}

// =============================================================================
// zhiPalace / findMaXing — edge cases
// =============================================================================

func TestZhiPalace_Invalid(t *testing.T) {
	if got := zhiPalace(0); got != GongKan {
		t.Errorf("zhiPalace(0) = %s, want 坎", got)
	}
}

func TestFindMaXing_Invalid(t *testing.T) {
	// All valid 12 zhi are covered. Just verify no panic for invalid.
	got := findMaXing(0)
	if got != GongKan {
		t.Logf("findMaXing(0) = %s", got)
	}
	got2 := findMaXing(13)
	if got2 != GongKan {
		t.Logf("findMaXing(13) = %s", got2)
	}
}

// =============================================================================
// starHomePalace — 无效
// =============================================================================

func TestStarHomePalace_Invalid(t *testing.T) {
	if got := starHomePalace(0); got != 4 {
		t.Errorf("starHomePalace(0) = %d, want 4 (default 中)", got)
	}
	if got := starHomePalace(10); got != 4 {
		t.Errorf("starHomePalace(10) = %d, want 4 (default 中)", got)
	}
}

// ── 十干克应权威锚点（《奇门遁甲秘笈大全》方向：name/pattern=天盘+地盘）──
// 曾发现字段 di/tian 存反导致查询方向错误（65 条全反），修复后以此防回归。
func TestGanInteraction_AuthoritativeAnchors(t *testing.T) {
	// key = [Earth(地盘), Heaven(天盘)] → 期望格名（含关键子串）+ 吉凶
	anchors := []struct {
		earth, heaven ganzhi.Gan
		want          string
		auspicious    bool
	}{
		{ganzhi.GanBing, ganzhi.GanWu, "青龙返首", true},    // 戊加丙
		{ganzhi.GanWu, ganzhi.GanBing, "飞鸟跌穴", true},    // 丙加戊
		{ganzhi.GanGeng, ganzhi.GanWu, "值符飞宫", false},   // 戊加庚
		{ganzhi.GanWu, ganzhi.GanGeng, "天乙伏宫", false},   // 庚加戊
		{ganzhi.GanXin, ganzhi.GanYi, "青龙逃走", false},    // 乙加辛
		{ganzhi.GanYi, ganzhi.GanXin, "白虎猖狂", false},    // 辛加乙
		{ganzhi.GanGui, ganzhi.GanDing, "朱雀投江", false},  // 丁加癸
		{ganzhi.GanDing, ganzhi.GanGui, "螣蛇夭矫", false},  // 癸加丁
		{ganzhi.GanBing, ganzhi.GanGeng, "太白入荧", false}, // 庚加丙
		{ganzhi.GanGeng, ganzhi.GanBing, "荧入太白", false}, // 丙加庚
		{ganzhi.GanGui, ganzhi.GanGeng, "大格", false},    // 庚加癸（天盘庚地盘癸）
		{ganzhi.GanGeng, ganzhi.GanGui, "太白入网", false},  // 癸加庚（天盘癸地盘庚）
		{ganzhi.GanRen, ganzhi.GanGeng, "移荡格", false},   // 庚加壬（上格/小格）
		{ganzhi.GanJi, ganzhi.GanGeng, "官符刑格", false},   // 庚加己（刑格）
		{ganzhi.GanXin, ganzhi.GanGeng, "白虎干格", false},  // 庚加辛（干格）
		{ganzhi.GanYi, ganzhi.GanGeng, "太白逢星", false},   // 庚加乙（合格）
	}
	for _, a := range anchors {
		entry, ok := ganInteractionTable[[2]ganzhi.Gan{a.earth, a.heaven}]
		if !ok {
			t.Errorf("表内缺失: 地盘%s天盘%s", ganzhi.GanName(a.earth), ganzhi.GanName(a.heaven))
			continue
		}
		if !containsStr(entry.PatternName, a.want) && !containsStr(entry.Name, a.want) {
			t.Errorf("地盘%s天盘%s（%s加%s）: got %q, want 含%q",
				ganzhi.GanName(a.earth), ganzhi.GanName(a.heaven),
				ganzhi.GanName(a.heaven), ganzhi.GanName(a.earth),
				entry.PatternName, a.want)
		}
		if entry.Auspicious != a.auspicious {
			t.Errorf("地盘%s天盘%s: 吉=%v, want %v", ganzhi.GanName(a.earth), ganzhi.GanName(a.heaven), entry.Auspicious, a.auspicious)
		}
	}
}

func containsStr(s, sub string) bool {
	return len(sub) > 0 && (len(s) >= len(sub)) && func() bool {
		for i := 0; i+len(sub) <= len(s); i++ {
			if s[i:i+len(sub)] == sub {
				return true
			}
		}
		return false
	}()
}
