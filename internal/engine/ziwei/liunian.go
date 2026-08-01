package ziwei

import "liki-engine/internal/engine/ganzhi"

// yearStemBranch returns the stem and branch for a Gregorian year.
func yearStemBranch(year int) (Gan, Zhi) {
	g := Gan(((year-4)%10+10)%10 + 1)
	z := Zhi(((year-4)%12+12)%12 + 1)
	return g, z
}



// liuNianSiHua computes the annual four transformations.
func liuNianSiHua(liuNian int) siHuaResult {
	liuGan, _ := yearStemBranch(liuNian)
	return computeSiHua(liuGan)
}

// liuNianMinors computes the annual minor stars (流耀) as Zhi (earth branch).
// 10 颗流耀：流魁/流钺/流昌/流曲/流禄/流羊/流陀/流马/流鸾/流喜（iztro yearly 流耀）。
func liuNianMinors(yearZhu ganzhi.Zhu, _ Zhi) map[starIndex]Zhi {
	// zhiMinus1(0-11) → Zhi(1-12)，输出地支名
	toZhi := func(zhiMinus1 int) Zhi { return Zhi((zhiMinus1%12 + 12) % 12 + 1) }
	chZM1, quZM1 := liuChangQuByGan(yearZhu.Gan)
	hlZM1 := hongLuanPos(yearZhu.Zhi)
	return map[starIndex]Zhi{
		TianKui:  toZhi(tianKuiPos(yearZhu.Gan)),   // 流魁
		TianYue:  toZhi(tianYuePos(yearZhu.Gan)),   // 流钺
		WenChang: toZhi(chZM1),                     // 流昌
		WenQu:    toZhi(quZM1),                     // 流曲
		LuCun:    toZhi(luCunPos(yearZhu.Gan)),     // 流禄
		QingYang: toZhi(qingYangPos(yearZhu.Gan)),  // 流羊
		TuoLuo:   toZhi(tuoLuoPos(yearZhu.Gan)),    // 流陀
		TianMa:   toZhi(tianMaPos(yearZhu.Zhi)),    // 流马
		HongLuan: toZhi(hlZM1),                     // 流鸾
		TianXi:   toZhi((hlZM1 + 6) % 12),          // 流喜
	}
}

// ComputeLiuNian assembles the full annual analysis.
func ComputeLiuNian(chart Chart, liuNian int) LiuNian {
	siHua := liuNianSiHua(liuNian)
	siHuaPalace := make(map[starIndex]palaceIndex)
	for _, p := range chart.Palaces {
		for _, s := range p.Stars {
			if _, ok := siHua[s.Star]; ok {
				siHuaPalace[s.Star] = p.Index
			}
		}
	}
	liuYearGan, liuYearZhi := yearStemBranch(liuNian)
	minorStars := liuNianMinors(ganzhi.Zhu{Gan: liuYearGan, Zhi: liuYearZhi}, chart.ShiZhi)

	// 流年命宫 = 流年支所在本命宫（地支坐标）
	mingZhi := chart.Palaces[chart.MingGong].Zhi
	flowMingZM1 := zhiToZhiMinus1(liuYearZhi)
	flowMing := zhiMinus1ToPalace(flowMingZM1, mingZhi)

	// 流年盘：流耀地支 → display 索引（iztro 简称：流魁/流钺/流昌/流曲/流禄/流羊/流陀/流马/流鸾/流喜）
	flowShort := map[starIndex]string{
		TianKui: "流魁", TianYue: "流钺", WenChang: "流昌", WenQu: "流曲",
		LuCun: "流禄", QingYang: "流羊", TuoLuo: "流陀", TianMa: "流马",
		HongLuan: "流鸾", TianXi: "流喜",
	}
	starByDisplay := make(map[int][]string)
	for sid, z := range minorStars {
		disp := zhiMinus1ToDisplay(zhiToZhiMinus1(z))
		name, ok := flowShort[sid]
		if !ok { name = starName(sid) }
		starByDisplay[disp] = append(starByDisplay[disp], name)
	}
	// 年解（流年附加星，iztro 年支定位）
	njDisp := zhiMinus1ToDisplay(nianJiePos(liuYearZhi))
	starByDisplay[njDisp] = append(starByDisplay[njDisp], "年解")
	yearlyIndex := zhiMinus1ToDisplay(zhiToZhiMinus1(liuYearZhi))
	flowPalaces := buildFlowPalaces(zhiToZhiMinus1(chart.Palaces[chart.MingGong].Zhi), yearlyIndex, starByDisplay)

	return LiuNian{
		MingGong:     flowMing,
		MingGongName: chart.Palaces[flowMing].Name,
		Zhi:          liuYearZhi,
		SiHua:        siHua,
		SiHuaPalace:  siHuaPalace,
		FuXing:       minorStars,
		Palaces:      flowPalaces,
	}
}
