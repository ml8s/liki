package ziwei

import "liki-engine/internal/engine/ganzhi"

// --- 紫微定位 (0.4) ---

func findZiwei(ju juShu, lunarDay int) palaceIndex {
	start, ok := ziweiStartPos[ju]
	if !ok {
		return 0
	}
	n := (lunarDay + int(ju) - 1) / int(ju)
	pos := (start - (n - 1)) % 12
	if pos < 0 {
		pos += 12
	}
	return palaceIndex(pos)
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

func placeMainStars(ziweiPos palaceIndex) map[palaceIndex][]starIndex {
	tianfuPos := palaceIndex((int(ziweiPos) + 2) % 12)
	m := make(map[palaceIndex][]starIndex)
	for _, e := range ziweiOffsets {
		pos := palaceIndex(((int(ziweiPos) - e.offset) + 12) % 12)
		m[pos] = append(m[pos], e.star)
	}
	for _, e := range tianfuOffsets {
		pos := palaceIndex((int(tianfuPos) + e.offset) % 12)
		m[pos] = append(m[pos], e.star)
	}
	return m
}

// --- 辅星安星 (0.6) — 每颗一个函数 + 装配 ---

// The following functions return zhi-1 values (0=子..11=亥), NOT palace indices.
// Callers must convert via zhiToPalace when placing into a chart.

func luCunPos(yearGan Gan) int {
	if yg := int(yearGan); yg >= 1 && yg <= 10 {
		return luCunTable[yg-1]
	}
	return 0
}

func tianKuiPos(yearGan Gan) int {
	if yg := int(yearGan); yg >= 1 && yg <= 10 {
		return tianKuiTable[yg-1]
	}
	return 0
}

func tianYuePos(tk int) int { return (tk + 6) % 12 }

func qingYangPos(yearGan Gan) int { return (luCunPos(yearGan) + 1) % 12 }
func tuoLuoPos(yearGan Gan) int  { return (luCunPos(yearGan) - 1 + 12) % 12 }

func tianMaPos(yearZhi Zhi) int {
	if yz := int(yearZhi); yz >= 1 && yz <= 12 {
		return tianMaTable[yz-1]
	}
	return 0
}

func zuoFuPos(lunarMonth int) int  { return (lunarMonth + 2) % 12 }
func youBiPos(lunarMonth int) int   { return (11 - lunarMonth + 12) % 12 }
func wenChangPos(hourZhi Zhi) int   { return (11 - int(hourZhi) + 12) % 12 }
func wenQuPos(hourZhi Zhi) int      { return (int(hourZhi) + 3) % 12 }
func diKongPos(hourZhi Zhi) int     { return (12 - int(hourZhi) + 12) % 12 }
func diJiePos(hourZhi Zhi) int      { return (int(hourZhi) + 10) % 12 }

func huoXingPos(yearZhi, hourZhi Zhi) int { return huoXingIndex(yearZhi, hourZhi) }
func lingXingPos(yearZhi, hourZhi Zhi) int { return lingXingIndex(yearZhi, hourZhi) }

func huoXingIndex(yearZhi, hourZhi Zhi) int {
	switch {
	case inGroup(yearZhi, 3, 7, 11):  // 寅午戌: 丑宫起子时
		return int((hourZhi + 1) % 12)
	case inGroup(yearZhi, 9, 1, 5):   // 申子辰: 寅宫起子时
		return int((hourZhi + 2) % 12)
	case inGroup(yearZhi, 6, 10, 2):  // 巳酉丑: 卯宫起子时
		return int((hourZhi + 3) % 12)
	case inGroup(yearZhi, 12, 4, 8):  // 亥卯未: 酉宫起子时
		return int((hourZhi + 9) % 12)
	}
	return 0
}

func lingXingIndex(yearZhi, hourZhi Zhi) int {
	switch {
	case inGroup(yearZhi, 3, 7, 11):  // 寅午戌: 卯宫起子时
		return int((hourZhi + 3) % 12)
	case inGroup(yearZhi, 9, 1, 5):   // 申子辰: 戌宫起子时
		return int((hourZhi + 10) % 12)
	case inGroup(yearZhi, 6, 10, 2):  // 巳酉丑: 戌宫起子时
		return int((hourZhi + 10) % 12)
	case inGroup(yearZhi, 12, 4, 8):  // 亥卯未: 戌宫起子时
		return int((hourZhi + 10) % 12)
	}
	return 0
}

func inGroup(zhi, a, b, c Zhi) bool { return zhi == a || zhi == b || zhi == c }

// zhiToPalace converts an absolute branch position (zhi-1: 0=子..11=亥) to
// the palace index whose branch matches, given the 命宫 branch.
func zhiToPalace(zhiMinus1 int, mingZhi Zhi) palaceIndex {
	targetZhi := zhiMinus1 + 1
	return palaceIndex((int(mingZhi) - targetZhi + 12) % 12)
}

// placeMinorStars collects all 14 minor star placements.
func placeMinorStars(yearZhu ganzhi.Zhu, lunarMonth int, hourZhi, mingZhi Zhi) map[palaceIndex][]starIndex {
	m := make(map[palaceIndex][]starIndex)
	add := func(pos palaceIndex, s starIndex) {
		m[pos] = append(m[pos], s)
	}
	tk := tianKuiPos(yearZhu.Gan)
	add(zhiToPalace(luCunPos(yearZhu.Gan), mingZhi), LuCun)
	add(zhiToPalace(tk, mingZhi), TianKui)
	add(zhiToPalace(tianYuePos(tk), mingZhi), TianYue)
	add(zhiToPalace(qingYangPos(yearZhu.Gan), mingZhi), QingYang)
	add(zhiToPalace(tuoLuoPos(yearZhu.Gan), mingZhi), TuoLuo)
	add(zhiToPalace(tianMaPos(yearZhu.Zhi), mingZhi), TianMa)
	add(zhiToPalace(zuoFuPos(lunarMonth), mingZhi), ZuoFu)
	add(zhiToPalace(youBiPos(lunarMonth), mingZhi), YouBi)
	add(zhiToPalace(wenChangPos(hourZhi), mingZhi), WenChang)
	add(zhiToPalace(wenQuPos(hourZhi), mingZhi), WenQu)
	add(zhiToPalace(diKongPos(hourZhi), mingZhi), DiKong)
	add(zhiToPalace(diJiePos(hourZhi), mingZhi), DiJie)
	add(zhiToPalace(huoXingIndex(yearZhu.Zhi, hourZhi), mingZhi), HuoXing)
	add(zhiToPalace(lingXingIndex(yearZhu.Zhi, hourZhi), mingZhi), LingXing)
	return m
}
