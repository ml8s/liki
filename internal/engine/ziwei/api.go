package ziwei

import (
	"liki-engine/internal/engine/ganzhi"
	"liki-engine/internal/engine/tianwen"
)

// ComputeChart computes a complete紫微命盘 from lunar time and gender.
// The LunarTime should come from tianwen.time's "lunar" field.
func ComputeChart(lt tianwen.LunarTime, gender ganzhi.Gender) Chart {
	nianGan := Gan(((lt.Year - 4) % 10 + 10) % 10 + 1)
	nianZhi := Zhi(((lt.Year - 4) % 12 + 12) % 12 + 1)
	shiZhi := Zhi(lt.Shichen)

	// Reconstruct minimal Bazi for computeChart
	bz := ganzhi.Bazi{
		Nian: ganzhi.Zhu{Gan: ganzhi.Gan(nianGan), Zhi: ganzhi.Zhi(nianZhi)},
		Shi:  ganzhi.Zhu{Zhi: ganzhi.Zhi(shiZhi)},
	}

	chart := computeChart(bz, lt)
	chart.BirthYear = lt.Year
	chart.Gender = gender
	chart.LunarMonth = lt.Month
	chart.LunarDay = lt.Day
	chart.NianZhi = nianZhi
	chart = buildChartDetail(chart)
	chart.MingZhu = soulStar(chart.Palaces[0].Zhi)
	chart.ShenZhu = bodyStar(nianZhi)
	return chart
}
