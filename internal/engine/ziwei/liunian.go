package ziwei

import "liki-engine/internal/engine/ganzhi"

// yearStemBranch returns the stem and branch for a Gregorian year.
func yearStemBranch(year int) (Gan, Zhi) {
	g := Gan(((year-4)%10+10)%10 + 1)
	z := Zhi(((year-4)%12+12)%12 + 1)
	return g, z
}

// liuNianMingGong returns the annual ming gong index.
func liuNianMingGong(liuYear, birthYear int) palaceIndex {
	xuSui := liuYear - birthYear + 1
	return palaceIndex((xuSui - 1) % 12)
}

// liuNianSiHua computes the annual four transformations.
func liuNianSiHua(liuYear int) siHuaResult {
	liuGan, _ := yearStemBranch(liuYear)
	return computeSiHua(liuGan)
}

// liuNianMinors computes the annual minor stars (zhi-1 values).
// Caller must convert to palaceIndex via zhiToPalace using the flow year 命宫.
func liuNianMinors(yearZhu ganzhi.Zhu, hourZhi Zhi) map[starIndex]int {
	return map[starIndex]int{
		QingYang: qingYangPos(yearZhu.Gan),
		TuoLuo:   tuoLuoPos(yearZhu.Gan),
		HuoXing:  huoXingIndex(yearZhu.Zhi, hourZhi),
		LingXing: lingXingIndex(yearZhu.Zhi, hourZhi),
	}
}

// ComputeLiuNian assembles the full annual analysis.
func ComputeLiuNian(chart Chart, liuYear int) LiuNian {
	mingGong := liuNianMingGong(liuYear, chart.BirthYear)
	siHua := liuNianSiHua(liuYear)
	siHuaPalace := make(map[starIndex]palaceIndex)
	for _, p := range chart.Palaces {
		for _, s := range p.Stars {
			if _, ok := siHua[s.Star]; ok {
				siHuaPalace[s.Star] = p.Index
			}
		}
	}
	liuYearGan, liuYearZhi := yearStemBranch(liuYear)
	minorStars := liuNianMinors(ganzhi.Zhu{Gan: liuYearGan, Zhi: liuYearZhi}, chart.HourZhi)
	return LiuNian{
		MingGong:     mingGong,
		MingGongName: PalaceNames[mingGong],
		SiHua:        siHua,
		SiHuaPalace:  siHuaPalace,
		MinorStars:   minorStars,
	}
}
