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
	"liki-engine/internal/engine/ganzhi"
)

// ComputeChart computes八宅命盘 from birth year and gender.
func ComputeChart(gender ganzhi.Gender, birthYear int) Chart {
	mg := ComputeMingGua(gender, birthYear)
	return Chart{
		MingGua:    mg,
		BaZhaiDirs: baZhaiDirectionsForGua(mg.Gua.Index),
		YearStars:  computeYearStars(birthYear),
	}
}
