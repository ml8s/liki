package ziwei

import "liki-engine/internal/engine/ganzhi"

// allPalaceXiaoXian computes XiaoXian ages for all 12 palaces.
// mingZhi needed for iztro→Liki palace conversion.
func allPalaceXiaoXian(nianZhi Zhi, gender ganzhi.Gender, count int, mingZhi Zhi) [12][]int {
	return anXingXiaoXian(nianZhi, gender, count, mingZhi)
}

func anXingXiaoXian(nianZhi Zhi, gender ganzhi.Gender, count int, mingZhi Zhi) [12][]int {
	var ageIdx int
	switch nianZhi {
	case 3, 7, 11: ageIdx = 2  // 寅午戌→辰
	case 9, 1, 5:  ageIdx = 8  // 申子辰→戌
	case 6, 10, 2: ageIdx = 5  // 巳酉丑→未
	case 12, 4, 8: ageIdx = 11 // 亥卯未→丑
	}
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
