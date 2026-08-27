package ziwei

// soulStar returns the 命主 (soul star) name for a given ming gong zhi.
func soulStar(mingZhi Zhi) string {
	switch mingZhi {
	case 1:
		return "贪狼"
	case 2:
		return "巨门"
	case 3:
		return "禄存"
	case 4:
		return "文曲"
	case 5:
		return "廉贞"
	case 6:
		return "武曲"
	case 7:
		return "破军"
	case 8:
		return "武曲"
	case 9:
		return "廉贞"
	case 10:
		return "文曲"
	case 11:
		return "禄存"
	case 12:
		return "巨门"
	}
	return ""
}

// bodyStar returns the 身主 (body star) name for a given birth year zhi.
func bodyStar(nianZhi Zhi) string {
	switch nianZhi {
	case 1, 7:
		return "火星"
	case 2, 8:
		return "天相"
	case 3, 9:
		return "天梁"
	case 4, 10:
		return "天同"
	case 5, 11:
		return "文昌"
	case 6, 12:
		return "天机"
	}
	return ""
}

// yuanGongPalace returns the gong index of 来因宫 (original gong).
// iztro rule: the gong whose heavenly gan equals the birth year's heavenly gan.
func yuanGongPalace(palaces [12]gong, nianGan Gan) gongIndex {
	for i, p := range palaces {
		if p.Gan == nianGan && p.Zhi != 1 && p.Zhi != 2 { // 排除子丑
			return gongIndex(i)
		}
	}
	return 0
}
