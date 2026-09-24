package ziwei

import "liki-engine/internal/engine/ganzhi"

// xiaoXianStartIndex 小限起宫：生年支三合局定起宫索引（xiaoxian_rules.json）。
func xiaoXianStartIndex(nianZhi Zhi) int {
	if idx, ok := xiaoxianStartByBranch[nianZhi]; ok {
		return idx
	}
	return 0
}

// allPalaceXiaoXian computes XiaoXian ages for all 12 palaces.
// mingZhi needed for iztro→Liki gong conversion.
func allPalaceXiaoXian(nianZhi Zhi, gender ganzhi.Gender, count int, mingZhi Zhi) [12][]int {
	return anXingXiaoXian(nianZhi, gender, count, mingZhi)
}

func anXingXiaoXian(nianZhi Zhi, gender ganzhi.Gender, count int, mingZhi Zhi) [12][]int {
	ageIdx := xiaoXianStartIndex(nianZhi)
	var result [12][]int
	for i := 0; i < 12; i++ {
		ages := make([]int, count)
		for j := 0; j < count; j++ {
			ages[j] = 12*j + i + 1
		}
		// 通过anXingIdx（经ageIdx偏移）确定实际放置位置
		var anXingIdx int
		if gender == Male {
			anXingIdx = (ageIdx + i) % 12
		} else {
			anXingIdx = (ageIdx - i + 12) % 12
		}
		// display坐标i映射到Liki宫位：通过anXingIdx（经ageIdx偏移）确定实际放置位置
		placeIdx := (anXingIdx + 2) % 12
		likiPalace := zhiIdxToPalaceIndex(zhiToZhiIdx(mingZhi), placeIdx)
		result[likiPalace] = ages
	}
	return result
}
