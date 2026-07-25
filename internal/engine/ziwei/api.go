// Package ziwei provides紫微斗数 computation.
//
// Types
//   Chart, Bond, DaXianStep, LiuNian, LiuYue, LiuRi
//
// Functions
//   ComputeChart(st SolarTime, gender Gender) → Chart
//   ComputeDaXian(chart Chart) → []DaXianStep
//   ComputeLiuNian(chart Chart, liuYear int) → LiuNian
//   ComputeLiuYue(chart Chart, liuYear int, lunarMonth int) → LiuYue
//   ComputeLiuRi(chart Chart, liuYear int, lunarMonth int, lunarDay int) → LiuRi
//   ComputeBond(a, b Chart) → Bond
package ziwei

import (
	"liki-engine/internal/engine/ganzhi"
	"liki-engine/internal/engine/tianwen"
)

// ComputeChart computes a complete紫微命盘 from solar birth time and gender.
func ComputeChart(st tianwen.SolarTime, gender ganzhi.Gender) Chart {
	bz := tianwen.ComputeBazi(st)
	lt := tianwen.SolarToLunar(tianwen.GregorianTime(st.Time()))
	y, _, _ := st.Time().Date()
	chart := computeChart(bz, lt)
	chart.BirthYear = y
	chart.Gender = gender
	chart = buildChartDetail(chart)
	return chart
}
