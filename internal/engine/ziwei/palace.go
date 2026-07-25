package ziwei

// --- 命宫/身宫 (0.1) ---

func computeMingShen(lunarMonth int, hourZhi Zhi) (mingZhi, shenZhi Zhi) {
	h := int(hourZhi)
	mingZhi = Zhi(((lunarMonth-h+2)%12+12)%12 + 1)
	shenZhi = Zhi(((lunarMonth+h)%12+12)%12 + 1)
	return
}

func arrangePalaceZhis(mingZhi Zhi) [12]Zhi {
	var zhis [12]Zhi
	for i := 0; i < 12; i++ {
		zhis[i] = Zhi(((int(mingZhi)-1-i)%12+12)%12 + 1)
	}
	return zhis
}

func findShenGongIndex(palaceZhis [12]Zhi, shenZhi Zhi) palaceIndex {
	for i, z := range palaceZhis {
		if z == shenZhi {
			return palaceIndex(i)
		}
	}
	return 0
}

// --- 十二宫天干 (0.2) ---

func yinGan(yearGan Gan) Gan {
	g := ((int(yearGan)-1)%5)*2 + 3
	return Gan(((g-1)%10+10)%10 + 1)
}

func arrangePalaceGans(yearGan Gan, mingZhi Zhi) (mingGan Gan, gans [12]Gan) {
	yg := yinGan(yearGan)
	mingGan = Gan(((int(yg)+int(mingZhi)-3-1)%10+10)%10 + 1)
	for i := 0; i < 12; i++ {
		gans[i] = Gan(((int(mingGan)-1-i)%10+10)%10 + 1)
	}
	return
}
