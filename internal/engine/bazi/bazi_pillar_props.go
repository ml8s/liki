package bazi

import "liki-engine/internal/engine/ganzhi"

// isSelfHe returns true if the pillar's stem and its branch's main hidden stem
// form a 天干五合 (stem-he harmony within a single pillar — 干支自合).
//
// Traditional pairs:
//
//	甲午 (甲+午中己→甲己合), 乙巳 (乙+巳中庚→乙庚合), 丙戌 (丙+戌中辛→丙辛合),
//	丁亥 (丁+亥中壬→丁壬合), 戊子 (戊+子中癸→戊癸合),
//	庚辰 (庚+辰中乙→乙庚合), 辛巳 (辛+巳中丙→丙辛合),
//	壬戌 (壬+戌中丁→丁壬合), 癸巳 (癸+巳中戊→戊癸合).
func isSelfHe(p ganzhi.Zhu) bool {
	hs := ganzhi.CangGanForZhi(p.Zhi)
	for _, h := range hs.Slice() {
		if h != nil && isGanHePair(p.Gan, *h) {
			return true
		}
	}
	return false
}

// selfHeName returns the 干支自合 description string (e.g. "甲己合").
func selfHeName(p ganzhi.Zhu) string {
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

// sanQiType checks if the four-pillar stem set contains a 三奇贵人 pattern.
// Returns "天上" (甲戊庚), "地下" (乙丙丁), or "人中" (壬癸辛), or empty.
func sanQiType(bz ganzhi.Bazi) string {
	zhus := bz.Slice()
	stemSet := [11]bool{}
	for _, p := range zhus {
		if s := int(p.Gan); s >= 1 && s <= 10 {
			stemSet[s] = true
		}
	}

	// 天上三奇: 甲(1)+戊(5)+庚(7)
	if stemSet[1] && stemSet[5] && stemSet[7] {
		return "天上"
	}
	// 地下三奇: 乙(2)+丙(3)+丁(4)
	if stemSet[2] && stemSet[3] && stemSet[4] {
		return "地下"
	}
	// 人中三奇: 壬(9)+癸(10)+辛(8)
	if stemSet[9] && stemSet[10] && stemSet[8] {
		return "人中"
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
