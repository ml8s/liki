package ziwei

import "liki-engine/internal/engine/ganzhi"

var boShiCycle = []string{
	"博士", "力士", "青龙", "小耗", "将军", "奏书",
	"飞廉", "喜神", "病符", "大耗", "伏兵", "官府",
}

// computeBoShi 博士十二神。iztro算法：从禄存起，年支阴阳=性别→顺行，≠→逆行。
func computeBoShi(ju juShu, mingZhi Zhi, nianGan Gan, gender ganzhi.Gender, nianZhi Zhi) [12]string {
	// 禄存在zhiIdx
	luCunZhiM1 := luCunTable[int(nianGan)-1]
	// zhiIdx → palace index via zhi matching
	palaceZhis := arrangePalaceZhis(mingZhi)
	luPalace := -1
	for i, z := range palaceZhis {
		if int(z)-1 == luCunZhiM1 {
			luPalace = i
			break
		}
	}
	if luPalace < 0 { return [12]string{} }

	// 年支阴阳 vs 性别 → 方向
	nianIsYang := int(nianZhi)%2 == 1
	isMale := gender == Male
	same := (nianIsYang && isMale) || (!nianIsYang && !isMale)
	dir := 1  // 相同→顺行
	if !same { dir = -1 }

	var result [12]string
	for i := 0; i < 12; i++ {
		idx := (luPalace - i*dir + 12) % 12
		result[idx] = boShiCycle[i]
	}
	return result
}
