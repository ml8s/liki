package qimen

import (
	"time"

	"liki-engine/internal/engine/tianwen"
)

func civilSolar(
	at time.Time,
	loc *time.Location,
	longitude float64,
) (time.Time, time.Time) {
	civil := at.In(loc)
	solar := tianwen.ApparentSolarCivil(at, loc, longitude)
	return time.Time(civil), time.Time(solar)
}

// ComputeChartAt is the explicit-time entry for the default chart.
// ComputeChartWithMethodAt is the explicit-time entry for a catalog method.
func ComputeChartWithMethodAt(
	at time.Time,
	loc *time.Location,
	longitude float64,
	method Method,
) (Chart, error) {
	gt, st := civilSolar(at, loc, longitude)
	return ComputeChartWithMethod(gt, st, method)
}

// ComputeChartWithYongShenAndMethodAt is the explicit-time entry for the
// yong-shen analysis.
func ComputeChartWithYongShenAndMethodAt(
	at time.Time,
	loc *time.Location,
	longitude float64,
	symbols []YongShenSymbol,
	birthDate BirthDate,
	method Method,
) (Chart, error) {
	gt, st := civilSolar(at, loc, longitude)
	return ComputeChartWithYongShenAndMethod(gt, st, symbols, birthDate, method)
}

// ComputeJinhanChartAt is the explicit-time entry for the Golden Mirror chart.
func ComputeJinhanChartAt(
	at time.Time,
	loc *time.Location,
	longitude float64,
) JinhanChart {
	gt, st := civilSolar(at, loc, longitude)
	return ComputeJinhanChart(gt, st)
}
