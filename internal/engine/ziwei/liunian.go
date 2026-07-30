package ziwei

import "liki-engine/internal/engine/ganzhi"

// yearStemBranch returns the stem and branch for a Gregorian year.
func yearStemBranch(year int) (Gan, Zhi) {
	g := Gan(((year-4)%10+10)%10 + 1)
	z := Zhi(((year-4)%12+12)%12 + 1)
	return g, z
}

// liuNianMingGong returns flow year命宫 index.
// 古典规则：流年地支直指——哪个宫的地支等于流年地支，即为流年命宫。
func liuNianMingGong(liuYear int, chart Chart) palaceIndex {
	_, liuZhi := yearStemBranch(liuYear)
	for i, p := range chart.Palaces {
		if p.Zhi == liuZhi {
			return palaceIndex(i)
		}
	}
	return 0
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
		MingGong:     0,
		MingGongName: "命宫",
		Zhi:          liuYearZhi,
		SiHua:        siHua,
		SiHuaPalace:  siHuaPalace,
		MinorStars:   minorStars,
	}
}
