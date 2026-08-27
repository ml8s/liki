package ziwei

import "liki-engine/internal/engine/ganzhi"

// --- 紫微定位 (iztro 六五四三二) ---

// findZiwei computes紫微 IZTRO index (0=寅). Caller converts to Liki gong.
func findZiwei(ju juShu, lunarDay int) int {
	lunarDay4 := lunarDay
	added := 0
	for (lunarDay4+added)%int(ju) != 0 {
		added++
	}
	lunarDay4 += added
	quotient := lunarDay4 / int(ju)
	quotient %= 12
	anXingIdx := quotient - 1 // 寅=1-1=0
	if added%2 == 1 {
		anXingIdx = (anXingIdx - added + 12) % 12 // odd→reverse
	} else {
		anXingIdx = (anXingIdx + added) % 12 // even→forward
	}
	return anXingIdx
}

// anXingIdxToPalace converts iztro 寅=0 index to Liki gong index.
func anXingIdxToPalace(anXingIdx int, mingZhi Zhi) gongIndex {
	anXingIdx = anXingIdx % 12
	if anXingIdx < 0 {
		anXingIdx += 12
	}
	zhiIdx := (anXingIdx + 2) % 12 // 0=寅→zhi-1=2
	return zhiIdxToPalaceIndex(zhiToZhiIdx(mingZhi), zhiIdx)
}

// --- 主星安星 (0.5) ---

var ziweiOffsets = []struct {
	star   starIndex
	offset int
}{
	{ZiWei, 0}, {TianJi, 1}, {TaiYang, 3}, {WuQu, 4}, {TianTong, 5}, {LianZhen, 8},
}

var tianfuOffsets = []struct {
	star   starIndex
	offset int
}{
	{TianFu, 0}, {TaiYin, 1}, {TanLang, 2}, {JuMen, 3},
	{TianXiang, 4}, {TianLiang, 5}, {QiSha, 6}, {PoJun, 10},
}

func placeMainStars(ziweiAnXingIdx, tianfuAnXingIdx int, _ Zhi) map[int][]starIndex {
	// 返回 zhiIdx 坐标（0=子..11=亥）：安星序(寅=0)(寅=0 安星序) → zhiIdx 固定映射
	m := make(map[int][]starIndex)
	for _, e := range ziweiOffsets {
		anXingIdx := (ziweiAnXingIdx - e.offset + 12) % 12
		zm1 := anXingIdxToZhiIdx(anXingIdx)
		m[zm1] = append(m[zm1], e.star)
	}
	for _, e := range tianfuOffsets {
		anXingIdx := (tianfuAnXingIdx + e.offset) % 12
		zm1 := anXingIdxToZhiIdx(anXingIdx)
		m[zm1] = append(m[zm1], e.star)
	}
	return m
}

// --- 辅星安星 (0.6) — 每颗一个函数 + 装配 ---

// The following functions return zhi-1 values (0=子..11=亥), NOT gong indices.
// Callers must convert via zhiIdxToPalaceIndex when placing into a chart.

func luCunPos(nianGan Gan) int {
	if yg := int(nianGan); yg >= 1 && yg <= 10 {
		return luCunTable[yg-1]
	}
	return 0
}

func tianKuiPos(nianGan Gan) int {
	// iztro公式：fixEarthlyBranchIndex(某支)
	m := map[Gan]int{1: 11, 2: 10, 3: 9, 4: 9, 5: 11, 6: 10, 7: 11, 8: 4, 9: 1, 10: 1}
	return (m[nianGan] + 2) % 12
}

func tianYuePos(nianGan Gan) int {
	// iztro独立表：甲未乙申丙酉丁酉戊未己申庚未辛寅壬巳癸巳
	m := map[Gan]int{1: 5, 2: 6, 3: 7, 4: 7, 5: 5, 6: 6, 7: 5, 8: 0, 9: 3, 10: 3}
	return (m[nianGan] + 2) % 12
}

func qingYangPos(nianGan Gan) int { return (luCunPos(nianGan) + 1) % 12 }
func tuoLuoPos(nianGan Gan) int   { return (luCunPos(nianGan) - 1 + 12) % 12 }

func tianMaPos(nianZhi Zhi) int {
	if yz := int(nianZhi); yz >= 1 && yz <= 12 {
		return tianMaTable[yz-1]
	}
	return 0
}

func zuoFuPos(lunarMonth int) int { return (lunarMonth + 3) % 12 }
func youBiPos(lunarMonth int) int { return (11 - lunarMonth + 12) % 12 }
func wenChangPos(shiZhi Zhi) int  { return (11 - int(shiZhi) + 12) % 12 }
func wenQuPos(shiZhi Zhi) int     { return (int(shiZhi) + 3) % 12 }
func diKongPos(shiZhi Zhi) int    { return (12 - int(shiZhi) + 12) % 12 }
func diJiePos(shiZhi Zhi) int     { return (int(shiZhi) + 10) % 12 }

func huoXingIndex(nianZhi, shiZhi Zhi) int {
	ti := (int(shiZhi) - 1 + 12) % 12
	var base int
	switch {
	case inGroup(nianZhi, 3, 7, 11):
		base = 11
	case inGroup(nianZhi, 9, 1, 5):
		base = 0
	case inGroup(nianZhi, 6, 10, 2):
		base = 1
	case inGroup(nianZhi, 12, 4, 8):
		base = 7
	}
	return (base + ti + 2) % 12
}

func lingXingIndex(nianZhi, shiZhi Zhi) int {
	ti := (int(shiZhi) - 1 + 12) % 12
	var base int
	switch {
	case inGroup(nianZhi, 3, 7, 11):
		base = 1
	case inGroup(nianZhi, 9, 1, 5):
		base = 8
	case inGroup(nianZhi, 6, 10, 2):
		base = 8
	case inGroup(nianZhi, 12, 4, 8):
		base = 8
	}
	return (base + ti + 2) % 12
}

// zhiIdxToPalaceIndex converts an absolute zhi position (zhi-1: 0=子..11=亥) to
// the gong index whose zhi matches, given the 命宫 zhi.
func inGroup(zhi, a, b, c Zhi) bool { return zhi == a || zhi == b || zhi == c }

// placeMinorStars collects all 14 minor star placements.
// 返回 zhiIdx 坐标（0=子..11=亥）。
func placeMinorStars(yearZhu ganzhi.Zhu, lunarMonth int, shiZhi, _ Zhi) map[int][]starIndex {
	m := make(map[int][]starIndex)
	add := func(zm1 int, s starIndex) {
		m[zm1] = append(m[zm1], s)
	}
	tk := tianKuiPos(yearZhu.Gan)
	add(luCunPos(yearZhu.Gan), LuCun)
	add(tk, TianKui)
	add(tianYuePos(yearZhu.Gan), TianYue)
	add(qingYangPos(yearZhu.Gan), QingYang)
	add(tuoLuoPos(yearZhu.Gan), TuoLuo)
	add(tianMaPos(yearZhu.Zhi), TianMa)
	add(zuoFuPos(lunarMonth), ZuoFu)
	add(youBiPos(lunarMonth), YouBi)
	add(wenChangPos(shiZhi), WenChang)
	add(wenQuPos(shiZhi), WenQu)
	add(diKongPos(shiZhi), DiKong)
	add(diJiePos(shiZhi), DiJie)
	add(huoXingIndex(yearZhu.Zhi, shiZhi), HuoXing)
	add(lingXingIndex(yearZhu.Zhi, shiZhi), LingXing)
	return m
}
