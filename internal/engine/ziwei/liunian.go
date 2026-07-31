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

// liuNianMinors computes the annual minor stars (zhi-1 values).
// Caller must convert to palaceIndex via zhiToPalace using the flow year 命宫.
func liuNianMinors(yearZhu ganzhi.Zhu, hourZhi Zhi) map[starIndex]Zhi {
	// zhiMinus1(0-11) → Zhi(1-12)，输出地支名
	toZhi := func(zhiMinus1 int) Zhi { return Zhi(zhiMinus1 + 1) }
	return map[starIndex]Zhi{
		QingYang: toZhi(qingYangPos(yearZhu.Gan)),
		TuoLuo:   toZhi(tuoLuoPos(yearZhu.Gan)),
		HuoXing:  toZhi(huoXingIndex(yearZhu.Zhi, hourZhi)),
		LingXing: toZhi(lingXingIndex(yearZhu.Zhi, hourZhi)),
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
	return LiuNian{
		MingGong:     0,
		MingGongName: "命宫",
		Zhi:          liuYearZhi,
		SiHua:        siHua,
		SiHuaPalace:  siHuaPalace,
		FuXing:   minorStars,
	}
}
