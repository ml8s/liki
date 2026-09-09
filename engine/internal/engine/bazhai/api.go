// Package bazhai provides八宅风水 computation.
//
// Types
//
//	Chart, MingGua, gua
//
// Functions
//
//	ComputeMingGuaChart(gender Gender, birthYear int) → Chart
//	ComputeMingGua(gender Gender, birthYear int) → MingGua
package bazhai

import (
	"liki-engine/internal/engine/ganzhi"
)

// ComputeMingGuaChart computes八宅命卦, auspicious/inauspicious directions and annual stars.
func ComputeMingGuaChart(gender ganzhi.Gender, birthYear int) Chart {
	mg := ComputeMingGua(gender, birthYear)
	return Chart{
		MingGua:    mg,
		BaZhaiDirs: baZhaiDirectionsForGua(mg.Gua.Index),
		YearStars:  computeYearStars(birthYear),
	}
}
