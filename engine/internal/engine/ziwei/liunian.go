package ziwei

import "liki-engine/internal/engine/ganzhi"

// yearGanZhi returns the nian gan-zhi (年干支) for a lunar year number.
// e.g., yearGanZhi(2026) = 丙午. The formula uses (year-4)%10/%12 from the 1984 甲子 anchor.
func yearGanZhi(year int) (Gan, Zhi) {
	g := Gan(((year-4)%10+10)%10 + 1)
	z := Zhi(((year-4)%12+12)%12 + 1)
	return g, z
}

// liuNianSiHua computes the annual four transformations.
func liuNianSiHua(liuNian int) siHuaResult {
	liuGan, _ := yearGanZhi(liuNian)
	return computeSiHua(liuGan)
}

// liuNianMinors computes the annual minor stars (流耀) as Zhi (earth zhi).
// 10 颗流耀：流魁/流钺/流昌/流曲/流禄/流羊/流陀/流马/流鸾/流喜（iztro yearly 流耀）。
func liuNianMinors(yearZhu ganzhi.Zhu, _ Zhi) map[starIndex]Zhi {
	// zhiIdx(0-11) → Zhi(1-12)，输出地支名
	toZhi := func(zhiIdx int) Zhi { return Zhi((zhiIdx%12+12)%12 + 1) }
	changZhiIdx, quZhiIdx := liuChangQuByGan(yearZhu.Gan)
	hongLuanZhiIdx := hongLuanPos(yearZhu.Zhi)
	return map[starIndex]Zhi{
		TianKui:  toZhi(tianKuiPos(yearZhu.Gan)),   // 流魁
		TianYue:  toZhi(tianYuePos(yearZhu.Gan)),   // 流钺
		WenChang: toZhi(changZhiIdx),               // 流昌
		WenQu:    toZhi(quZhiIdx),                  // 流曲
		LuCun:    toZhi(luCunPos(yearZhu.Gan)),     // 流禄
		QingYang: toZhi(qingYangPos(yearZhu.Gan)),  // 流羊
		TuoLuo:   toZhi(tuoLuoPos(yearZhu.Gan)),    // 流陀
		TianMa:   toZhi(tianMaPos(yearZhu.Zhi)),    // 流马
		HongLuan: toZhi(hongLuanZhiIdx),            // 流鸾
		TianXi:   toZhi((hongLuanZhiIdx + 6) % 12), // 流喜
	}
}

// ComputeLiuNian assembles the full annual analysis.
func ComputeLiuNian(chart Chart, liuNian int) LiuNian {
	siHua := liuNianSiHua(liuNian)
	siHuaPalace := make(map[starIndex]gongIndex)
	for _, p := range chart.GongWei {
		for _, s := range p.Stars {
			if _, ok := siHua[s.Star]; ok {
				siHuaPalace[s.Star] = p.Index
			}
		}
	}
	liuYearGan, liuYearZhi := yearGanZhi(liuNian)
	minorStars := liuNianMinors(ganzhi.Zhu{Gan: liuYearGan, Zhi: liuYearZhi}, chart.ShiZhi)

	// 流年命宫 = 流年支所在本命宫（地支坐标）
	mingZhi := chart.GongWei[chart.MingGong].Zhi
	flowMingZhiIdx := zhiToZhiIdx(liuYearZhi)
	flowMing := zhiIdxToPalaceIndex(zhiToZhiIdx(mingZhi), flowMingZhiIdx)

	// 流年盘：流耀地支 → 安星索引（iztro 简称：流魁/流钺/流昌/流曲/流禄/流羊/流陀/流马/流鸾/流喜）
	flowShort := map[starIndex]string{
		TianKui: "流魁", TianYue: "流钺", WenChang: "流昌", WenQu: "流曲",
		LuCun: "流禄", QingYang: "流羊", TuoLuo: "流陀", TianMa: "流马",
		HongLuan: "流鸾", TianXi: "流喜",
	}
	starByAnXingIdx := make(map[int][]string)
	for sid, z := range minorStars {
		anXingIdx := zhiIdxToAnXingIdx(zhiToZhiIdx(z))
		name, ok := flowShort[sid]
		if !ok {
			name = starName(sid)
		}
		starByAnXingIdx[anXingIdx] = append(starByAnXingIdx[anXingIdx], name)
	}
	// 年解（流年附加星，iztro 年支定位）
	njAnXingIdx := zhiIdxToAnXingIdx(nianJiePos(liuYearZhi))
	starByAnXingIdx[njAnXingIdx] = append(starByAnXingIdx[njAnXingIdx], "年解")
	yearlyIndex := zhiIdxToAnXingIdx(zhiToZhiIdx(liuYearZhi))
	flowPalaces := buildFlowPalaces(yearlyIndex, starByAnXingIdx)

	return LiuNian{
		MingGong:     flowMing,
		MingGongName: chart.GongWei[flowMing].Name,
		Zhi:          liuYearZhi,
		SiHua:        siHua,
		SiHuaPalace:  siHuaPalace,
		FuXing:       minorStars,
		GongWei:      flowPalaces,
	}
}
