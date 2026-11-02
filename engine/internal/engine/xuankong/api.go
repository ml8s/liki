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

// ComputeChartAtStandard derives the 立春 year from a standard time and
// computes the 玄空 chart.
func ComputeChartAtStandard(standard time.Time, sitMountain, faceMountain int) (Chart, error) {
	ly, err := tianwen.LichunYearAt(standard)
	if err != nil {
		return Chart{}, err
	}
	return computeChart(sitMountain, faceMountain, ly.Year), nil
}
