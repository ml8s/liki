// Package xuankong provides玄空风水 computation.
//
// Types
//   Chart, SanYuanYun
//
// Functions
//   ComputeChart(st SolarTime, sitMountain int, faceMountain int) → Chart
//   ComputeSanYuanYun(year int) → SanYuanYun
package xuankong

import "liki-engine/internal/engine/tianwen"

// ComputeChart computes the 玄空飞星盘 for a given坐向.
func ComputeChart(st tianwen.SolarTime, sitMountain, faceMountain int) Chart {
	return computeChart(sitMountain, faceMountain, st.Time().Year())
}
