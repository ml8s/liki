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
)

// ComputeChart computes the 玄空飞星盘 for a given坐向.
func ComputeChart(st time.Time, sitMountain, faceMountain int) Chart {
	return computeChart(sitMountain, faceMountain, st.Year())
}
