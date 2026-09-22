package ziwei

var jiangQianNames = [12]string{
	"将星", "攀鞍", "岁驿", "息神", "华盖", "劫煞",
	"灾煞", "天煞", "指背", "咸池", "月煞", "亡神",
}

var suiQianNames = [12]string{
	"岁建", "晦气", "丧门", "贯索", "官符", "小耗",
	"大耗", "龙德", "白虎", "天德", "吊客", "病符",
}

// computeJiangQian returns the 将前12神. iztro: 年支定起始→顺时针display序→zhiIdxToPalaceIndex.
func computeJiangQian(nianZhi Zhi, mingZhi Zhi) [12]string {
	var jqStartIdx int
	switch nianZhi {
	case 3, 7, 11: // 寅午戌→午
		jqStartIdx = 4
	case 9, 1, 5: // 申子辰→子
		jqStartIdx = 10
	case 6, 10, 2: // 巳酉丑→酉(anXingIdx=7)
		jqStartIdx = 7
	default: // 亥卯未→卯
		jqStartIdx = 1
	}
	result := [12]string{}
	for i := 0; i < 12; i++ {
		anXingIdx := (jqStartIdx + i) % 12
		zhiIdx := (anXingIdx + 2) % 12
		likiPalace := zhiIdxToPalaceIndex(zhiToZhiIdx(mingZhi), zhiIdx)
		result[likiPalace] = jiangQianNames[i]
	}
	return result
}

// computeSuiQian returns the 岁前12神. iztro: 年支起顺时针display序→zhiIdxToPalaceIndex.
func computeSuiQian(nianZhi Zhi, mingZhi Zhi) [12]string {
	nianAnXingIdx := (int(nianZhi) - 3 + 12) % 12 // nianZhi→iztro 安星索引
	result := [12]string{}
	for i := 0; i < 12; i++ {
		anXingIdx := (nianAnXingIdx + i) % 12
		zhiIdx := (anXingIdx + 2) % 12
		likiPalace := zhiIdxToPalaceIndex(zhiToZhiIdx(mingZhi), zhiIdx)
		result[likiPalace] = suiQianNames[i]
	}
	return result
}
