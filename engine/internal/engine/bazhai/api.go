// Package bazhai provides八宅风水 computation.
//
// Types
//
//	Chart, MingGua, gua
//
// Functions
//
//	ComputeChart(gender Gender, birthYear int) → Chart
//	ComputeMingGua(gender Gender, birthYear int) → MingGua
package bazhai

import (
	"time"

	"liki-engine/internal/engine/ganzhi"
	"liki-engine/internal/engine/tianwen"
)

// ComputeChart computes八宅命盘 from a 立春 birth year and gender.
func ComputeChart(gender ganzhi.Gender, birthYear int) Chart {
	mg := ComputeMingGua(gender, birthYear)
	return Chart{
		MingGua:    mg,
		BaZhaiDirs: baZhaiDirectionsForGua(mg.Gua.Index),
		YearStars:  computeYearStars(birthYear),
	}
}

// ComputeChartAtDate derives the 立春 year from a birth date and computes the
// 八宅 chart. The 立春 boundary date belongs to the new year.
func ComputeChartAtDate(gender ganzhi.Gender, date tianwen.CivilDate) (Chart, error) {
	ly, err := tianwen.LichunYearOf(date)
	if err != nil {
		return Chart{}, err
	}
	return ComputeChart(gender, ly.Year), nil
}

// ComputeChartAtInstant derives the 立春 year from an absolute instant and
// computes the 八宅 chart. The 立春 boundary is an absolute instant, so the
// observer's timezone does not change the year.
func ComputeChartAtInstant(gender ganzhi.Gender, instant time.Time) (Chart, error) {
	ly, err := tianwen.LichunYearAt(instant)
	if err != nil {
		return Chart{}, err
	}
	return ComputeChart(gender, ly.Year), nil
}
