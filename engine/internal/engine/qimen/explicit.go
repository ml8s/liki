package qimen

import (
	"time"

	"liki-engine/internal/engine/tianwen"
)

func civilSolar(standard time.Time, longitude float64) (time.Time, time.Time) {
	solar := tianwen.ApparentSolarFromStandard(standard, longitude)
	return standard, solar
}

// ComputeChartAt is the explicit-time entry for the default chart.
// ComputeChartWithMethodAt is the explicit-time entry for a catalog method.
func ComputeChartWithMethodAt(
	standard time.Time,
	longitude float64,
	method Method,
) (Chart, error) {
	gt, st := civilSolar(standard, longitude)
	return ComputeChartWithMethod(gt, st, method)
}

// ComputeChartWithYongShenAndMethodAt is the explicit-time entry for the
// yong-shen analysis.
func ComputeChartWithYongShenAndMethodAt(
	standard time.Time,
	longitude float64,
	symbols []YongShenSymbol,
	birthDate BirthDate,
	method Method,
) (Chart, error) {
	gt, st := civilSolar(standard, longitude)
	return ComputeChartWithYongShenAndMethod(gt, st, symbols, birthDate, method)
}

// ComputeJinhanChartAt is the explicit-time entry for the Golden Mirror chart.
func ComputeJinhanChartAt(
	standard time.Time,
	longitude float64,
) JinhanChart {
	gt, st := civilSolar(standard, longitude)
	return ComputeJinhanChart(gt, st)
}
