// Package xuankong provides玄空风水 computation.
//
// Types
//
//	Chart, SanYuanYun
//
// Functions
//
//	ComputeChart(st time.Time, sitMountain int, faceMountain int) → Chart
//	ComputeSanYuanYun(year int) → SanYuanYun
package xuankong

import (
	"time"

	"liki-engine/internal/engine/tianwen"
)

// ComputeChart computes the 玄空飞星盘 for a given 坐向. The year is the
// 立春 year containing st, not the Gregorian calendar year.
func ComputeChart(st time.Time, sitMountain, faceMountain int) Chart {
	chart, err := ComputeChartAtStandard(st, sitMountain, faceMountain)
	if err != nil {
		return Chart{}
	}
	return chart
}

// ComputeChartAtStandard derives the 立春 year from a standard time and
// computes the 玄空 chart.
func ComputeChartAtStandard(instant time.Time, sitMountain, faceMountain int) (Chart, error) {
	ly, err := tianwen.LichunYearAt(instant)
	if err != nil {
		return Chart{}, err
	}
	return computeChart(sitMountain, faceMountain, ly.Year), nil
}
