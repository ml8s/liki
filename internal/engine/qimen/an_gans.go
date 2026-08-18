package qimen

import "liki-engine/internal/engine/ganzhi"

// eightStems is the 8 stems for hidden stem arrangement.
var eightStems = [8]ganzhi.Gan{
	ganzhi.GanYi, ganzhi.GanBing, ganzhi.GanDing, ganzhi.GanWu,
	ganzhi.GanJi, ganzhi.GanGeng, ganzhi.GanXin, ganzhi.GanRen,
}

// placeAnGan arranges hidden stems (暗干) on the 9 palaces.
func placeAnGan(driveZhu ganzhi.Zhu, dutyDoorPalace int) [9]ganzhi.Gan {
	var angans [9]ganzhi.Gan

	// 甲遁于旬首.
	searchGan := driveZhu.Gan
	if driveZhu.Gan == ganzhi.GanJia {
		searchGan = findXunShou(driveZhu)
	}

	startIdx := 0
	for i, s := range eightStems {
		if s == searchGan {
			startIdx = i
			break
		}
	}
	for i, si := 0, 0; si < 8; i++ {
		pos := (dutyDoorPalace + i) % 9
		if pos == 4 {
			continue
		}
		angans[pos] = eightStems[(startIdx+si)%8]
		si++
	}
	return angans
}

// maXingZhi returns the 马星 branch for a given 时支（三合局马星口诀）。
func maXingZhi(driveZhi ganzhi.Zhi) ganzhi.Zhi {
	switch int(driveZhi) {
	case 1, 5, 9: // 子, 辰, 申 → 马在寅
		return ganzhi.ZhiYin
	case 3, 7, 11: // 寅, 午, 戌 → 马在申
		return ganzhi.ZhiShen
	case 6, 10, 2: // 巳, 酉, 丑 → 马在亥
		return ganzhi.ZhiHai
	case 12, 4, 8: // 亥, 卯, 未 → 马在巳
		return ganzhi.ZhiSi
	}
	return ganzhi.ZhiZi
}

// findMaXing returns the 马星 gong for a given branch.
func findMaXing(driveZhi ganzhi.Zhi) GongIndex {
	return zhiPalace(maXingZhi(driveZhi))
}

// kongWangZhi returns the two 空亡 branches of the driving pillar's 旬.
func kongWangZhi(driveZhu ganzhi.Zhu) [2]ganzhi.Zhi {
	idx := ganzhi.SixtyCycleIndex(driveZhu.Gan, driveZhu.Zhi) // 0-59
	xunIdx := idx / 10                                       // 0-5
	return [6][2]ganzhi.Zhi{
		{11, 12}, // 甲子旬: 戌亥
		{9, 10},  // 甲戌旬: 申酉
		{7, 8},   // 甲申旬: 午未
		{5, 6},   // 甲午旬: 辰巳
		{3, 4},   // 甲辰旬: 寅卯
		{1, 2},   // 甲寅旬: 子丑
	}[xunIdx]
}

// findKongWang returns the two 空亡 palaces.
func findKongWang(driveZhu ganzhi.Zhu) [2]GongIndex {
	z := kongWangZhi(driveZhu)
	return [2]GongIndex{zhiPalace(z[0]), zhiPalace(z[1])}
}
