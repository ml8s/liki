// Package bazhai provides八宅风水 computation.
//
// Types
//
//	Chart, MingGua, gua
//
// Functions
//
//	ComputeChart(st SolarTime, gender Gender) → Chart
//	ComputeMingGua(gender Gender, birthYear int) → MingGua
package bazhai

import (
	"liki-engine/internal/engine/ganzhi"
	"liki-engine/internal/engine/tianwen"
)

// ComputeChart computes a complete八宅合参 from solar time and gender.
func ComputeChart(st tianwen.SolarTime, gender ganzhi.Gender) Chart {
	bz := tianwen.ComputeBazi(st)
	return computeChart(bz, gender, st.Time().Year())
}
