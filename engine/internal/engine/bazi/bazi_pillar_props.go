package bazi

import "liki-engine/internal/engine/ganzhi"

// selfHePillars 干支自合（12 组，现代命理共识：天干与地支藏干构成五合）：
//
//	甲午（甲+午中己→甲己合）、乙巳（乙+巳中庚→乙庚合）、丙戌（丙+戌中辛→丙辛合）、
//	丁亥（丁+亥中壬→丁壬合）、戊子（戊+子中癸→戊癸合）、己亥（己+亥中甲→甲己合）、
//	庚辰（庚+辰中乙→乙庚合）、辛巳（辛+巳中丙→丙辛合）、壬午（壬+午中丁→丁壬合）、
//	壬戌（壬+戌中丁→丁壬合）、癸巳（癸+巳中戊→戊癸合）、戊辰（戊+辰中癸→戊癸合）。
//
// 含《三命通会》9 组（甲午乙巳丙戌丁亥戊子庚辰辛巳壬戌癸巳）及主流扩展
// 己亥、壬午、戊辰；排除仅中余气暗合的非主流组合（如甲丑、戊丑、丙丑、甲未、乙申）。
var selfHePillars = map[ganzhi.Gan][]ganzhi.Zhi{
	ganzhi.GanJia:  {ganzhi.ZhiWu},                 // 甲午
	ganzhi.GanYi:   {ganzhi.ZhiSi},                 // 乙巳
	ganzhi.GanBing: {ganzhi.ZhiXu},                 // 丙戌
	ganzhi.GanDing: {ganzhi.ZhiHai},                // 丁亥
	ganzhi.GanWu:   {ganzhi.ZhiZi, ganzhi.ZhiChen}, // 戊子、戊辰
	ganzhi.GanJi:   {ganzhi.ZhiHai},                // 己亥
	ganzhi.GanGeng: {ganzhi.ZhiChen},               // 庚辰
	ganzhi.GanXin:  {ganzhi.ZhiSi},                 // 辛巳
	ganzhi.GanRen:  {ganzhi.ZhiWu, ganzhi.ZhiXu},   // 壬午、壬戌
	ganzhi.GanGui:  {ganzhi.ZhiSi},                 // 癸巳
}

// isSelfHe returns true if the pillar is one of the 干支自合 pillars.
func isSelfHe(p ganzhi.Zhu) bool {
	zhis, ok := selfHePillars[p.Gan]
	if !ok {
		return false
	}
	for _, z := range zhis {
		if z == p.Zhi {
			return true
		}
	}
	return false
}

// selfHeName returns the 干支自合 description string (e.g. "甲己合").
func selfHeName(p ganzhi.Zhu) string {
	if !isSelfHe(p) {
		return ""
	}
	hs := ganzhi.CangGanForZhi(p.Zhi)
	for _, h := range hs.Slice() {
		if h != nil && isGanHePair(p.Gan, *h) {
			return ganzhi.GanName(p.Gan) + ganzhi.GanName(*h) + "合"
		}
	}
	return ""
}

func isGanHePair(a, b ganzhi.Gan) bool { return ganzhi.IsGanHe(a, b) }

// isKuiGang checks if the pillar is a 魁罡 day pillar.
// 魁罡: 庚辰, 庚戌, 壬辰, 戊戌.
func isKuiGang(p ganzhi.Zhu) bool {
	s, b := int(p.Gan), int(p.Zhi)
	return (s == 7 && b == 5) || // 庚辰
		(s == 7 && b == 11) || // 庚戌
		(s == 9 && b == 5) || // 壬辰
		(s == 5 && b == 11) // 戊戌
}

// sanQiType checks if the four-pillar gan sequence contains a 三奇贵人 pattern.
// 三奇要求三干在年-月-日-时（或时-日-月-年）中顺次/逆次连续排列：
// 天上三奇（甲戊庚）、地下三奇（乙丙丁）、人中三奇（壬癸辛）。
// 仅"集齐"而不连续（如甲年庚月戊日）不构成三奇。
func sanQiType(bz ganzhi.Bazi) string {
	zhus := bz.Slice()
	gan := [4]int{}
	for i, p := range zhus {
		if s := int(p.Gan); s >= 1 && s <= 10 {
			gan[i] = s
		}
	}
	patterns := []struct {
		name    string
		a, b, c int
	}{
		{"天上", 1, 5, 7},  // 甲戊庚
		{"地下", 2, 3, 4},  // 乙丙丁
		{"人中", 9, 10, 8}, // 壬癸辛
	}
	for _, pat := range patterns {
		for start := 0; start <= 1; start++ { // 年-月-日 或 月-日-时 两个三元组
			s0, s1, s2 := gan[start], gan[start+1], gan[start+2]
			if (s0 == pat.a && s1 == pat.b && s2 == pat.c) || // 顺次
				(s0 == pat.c && s1 == pat.b && s2 == pat.a) { // 逆次
				return pat.name
			}
		}
	}
	return ""
}

// sanQiName returns the full name for a sanqi type code.
func sanQiName(typ string) string {
	switch typ {
	case "天上":
		return "天上三奇（甲戊庚）"
	case "地下":
		return "地下三奇（乙丙丁）"
	case "人中":
		return "人中三奇（壬癸辛）"
	}
	return ""
}
