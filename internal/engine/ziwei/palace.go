package ziwei

// --- 命宫/身宫 (0.1) ---

func computeMingShen(lunarMonth int, shiZhi Zhi) (mingZhi, shenZhi Zhi) {
	h := int(shiZhi)
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

func findShenGongIndex(palaceZhis [12]Zhi, shenZhi Zhi) gongIndex {
	for i, z := range palaceZhis {
		if z == shenZhi {
			return gongIndex(i)
		}
	}
	return 0
}

// --- 十二宫天干 (0.2) ---

func arrangePalaceGans(nianGan Gan, mingZhi Zhi, soulIzTroIdx int) (mingGan Gan, gans [12]Gan) {
	// 正月干（五虎遁）+ soulIndex（iztro算法）
	var zhengYueGan Gan
	switch nianGan {
	case 1, 6: zhengYueGan = 3  // 甲己丙
	case 2, 7: zhengYueGan = 5  // 乙庚戊
	case 3, 8: zhengYueGan = 7  // 丙辛庚
	case 4, 9: zhengYueGan = 9  // 丁壬壬
	case 5, 10: zhengYueGan = 1 // 戊癸甲
	}
	// iztro: mingGan = 正月干 + soulIndex(寅0系)
	mingGan = Gan(((int(zhengYueGan) - 1 + soulIzTroIdx) % 10 + 10) % 10 + 1)
	// iztro公式：gans[i] = fixIndex(HEAVENLY_STEMS.indexOf(mingGan) - soulIzTroIdx + i)
	// 其中soulIzTroIdx是命宫在display坐标中的索引(寅=0 安星序)
	for i := 0; i < 12; i++ {
		// display坐标中第i宫的地支 = (寅+i)
		// 对应天干 = mingGan - soulIzTroIdx + i
		gan := Gan(((int(mingGan) - 1 - soulIzTroIdx + i) % 10 + 10) % 10 + 1)
		// i是display坐标索引，需要映射到Liki gong order
		// display i → zhiIdx → Liki gong
		palaceZhiM1 := (i + 2) % 12
		likiIdx := zhiIdxToPalaceIndex(zhiToZhiIdx(mingZhi), palaceZhiM1)
		gans[likiIdx] = gan
	}
	return
}

// yinGan still needed by liuyue.go
func yinGan(nianGan Gan) Gan {
	g := ((int(nianGan)-1)%5)*2 + 3
	return Gan(((g-1)%10+10)%10 + 1)
}
