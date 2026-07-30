package ziwei

import "liki-engine/internal/engine/ganzhi"

// --- 紫微定位 (iztro 六五四三二) ---

// findZiwei computes紫微 IZTRO index (0=寅). Caller converts to Liki palace.
func findZiwei(ju juShu, lunarDay int) int {
	lunarDay4 := lunarDay
	added := 0
	for (lunarDay4+added)%int(ju) != 0 {
		added++
	}
	lunarDay4 += added
	quotient := lunarDay4 / int(ju)
	quotient %= 12
	iztroIdx := quotient - 1 // 寅=1-1=0
	if added%2 == 1 {
		iztroIdx = (iztroIdx - added + 12) % 12 // odd→reverse
	} else {
		iztroIdx = (iztroIdx + added) % 12 // even→forward
	}
	return iztroIdx
}

// iztroIdxToPalace converts iztro 寅=0 index to Liki palace index.
func iztroIdxToPalace(iztroIdx int, mingZhi Zhi) palaceIndex {
	iztroIdx = iztroIdx % 12
	if iztroIdx < 0 { iztroIdx += 12 }
	zhiMinus1 := (iztroIdx + 2) % 12 // 0=寅→zhi-1=2
	return zhiToPalace(zhiMinus1, mingZhi)
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

func placeMainStars(iztroZW, iztroTF int, mingZhi Zhi) map[palaceIndex][]starIndex {
	m := make(map[palaceIndex][]starIndex)
	for _, e := range ziweiOffsets {
		izIdx := (iztroZW - e.offset + 12) % 12
		m[iztroIdxToPalace(izIdx, mingZhi)] = append(m[iztroIdxToPalace(izIdx, mingZhi)], e.star)
	}
	for _, e := range tianfuOffsets {
		izIdx := (iztroTF + e.offset) % 12
		m[iztroIdxToPalace(izIdx, mingZhi)] = append(m[iztroIdxToPalace(izIdx, mingZhi)], e.star)
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
	// iztro公式：fixEarthlyBranchIndex(某支)
	m := map[Gan]int{1:11,2:10,3:9,4:9,5:11,6:10,7:11,8:4,9:1,10:1}
	return (m[yearGan] + 2) % 12
}

func tianYuePos(yearGan Gan) int {
	// iztro独立表：甲未乙申丙酉丁酉戊未己申庚未辛寅壬巳癸巳
	m := map[Gan]int{1:5,2:6,3:7,4:7,5:5,6:6,7:5,8:0,9:3,10:3}
	return (m[yearGan] + 2) % 12
}

func qingYangPos(yearGan Gan) int { return (luCunPos(yearGan) + 1) % 12 }
func tuoLuoPos(yearGan Gan) int  { return (luCunPos(yearGan) - 1 + 12) % 12 }

func tianMaPos(yearZhi Zhi) int {
	if yz := int(yearZhi); yz >= 1 && yz <= 12 {
		return tianMaTable[yz-1]
	}
	return 0
}

func zuoFuPos(lunarMonth int) int  { return (lunarMonth + 3) % 12 }
func youBiPos(lunarMonth int) int   { return (11 - lunarMonth + 12) % 12 }
func wenChangPos(hourZhi Zhi) int   { return (11 - int(hourZhi) + 12) % 12 }
func wenQuPos(hourZhi Zhi) int      { return (int(hourZhi) + 3) % 12 }
func diKongPos(hourZhi Zhi) int     { return (12 - int(hourZhi) + 12) % 12 }
func diJiePos(hourZhi Zhi) int      { return (int(hourZhi) + 10) % 12 }

func huoXingPos(yearZhi, hourZhi Zhi) int { return huoXingIndex(yearZhi, hourZhi) }
func lingXingPos(yearZhi, hourZhi Zhi) int { return lingXingIndex(yearZhi, hourZhi) }

func huoXingIndex(yearZhi, hourZhi Zhi) int {
	ti := (int(hourZhi) - 1 + 12) % 12
	var base int
	switch {
	case inGroup(yearZhi, 3, 7, 11): base = 11
	case inGroup(yearZhi, 9, 1, 5):  base = 0
	case inGroup(yearZhi, 6, 10, 2): base = 1
	case inGroup(yearZhi, 12, 4, 8): base = 7
	}
	return (base + ti + 2) % 12
}

func lingXingIndex(yearZhi, hourZhi Zhi) int {
	ti := (int(hourZhi) - 1 + 12) % 12
	var base int
	switch {
	case inGroup(yearZhi, 3, 7, 11): base = 1
	case inGroup(yearZhi, 9, 1, 5):  base = 8
	case inGroup(yearZhi, 6, 10, 2): base = 8
	case inGroup(yearZhi, 12, 4, 8): base = 8
	}
	return (base + ti + 2) % 12
}

// zhiToPalace converts an absolute branch position (zhi-1: 0=子..11=亥) to
// the palace index whose branch matches, given the 命宫 branch.
func inGroup(zhi, a, b, c Zhi) bool { return zhi == a || zhi == b || zhi == c }

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
	add(zhiToPalace(tianYuePos(yearZhu.Gan), mingZhi), TianYue)
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
