// Package qimen provides 奇门遁甲 computation.
//
// Types
//
//	Chart, Gong
//	GongIndex, StarIndex, DoorIndex, SpiritIndex
//	GanInteraction, MenInteraction, XingInteraction
//	WangShuai, Pattern, YingQi
//	ChartKind
//
// Constants
//
//	GongKan .. GongLi  (九宫)
//	StarTianPeng .. StarTianYing  (九星)
//	DoorXiu .. DoorKai  (八门)
//	SpiritZhiFu .. SpiritJiuTian  (八神)
//	ShiQiMen, RiQiMen, YueQiMen, NianQiMen  (奇门种类)
package qimen

import (
	"fmt"

	"liki-engine/internal/engine/tianwen"
)

// ChartKind represents the type of Qimen chart (时/日/月/年).
type ChartKind string

const (
	ShiQiMen  ChartKind = "shi"
	RiQiMen   ChartKind = "ri"
	YueQiMen  ChartKind = "yue"
	NianQiMen ChartKind = "nian"
)

// ParseChartKind parses a chart kind string.
func ParseChartKind(s string) (ChartKind, error) {
	switch s {
	case "shi":
		return ShiQiMen, nil
	case "ri":
		return RiQiMen, nil
	case "yue":
		return YueQiMen, nil
	case "nian":
		return NianQiMen, nil
	}
	return "", fmt.Errorf("invalid chart kind %q: must be shi/ri/yue/nian", s)
}

// ComputeChart computes a complete奇门盘 with all analyses.
func ComputeChart(st tianwen.SolarTime, kind ChartKind) Chart {
	bz := tianwen.ComputeBazi(st)
	t := st.Time()
	y, m, d := t.Date()
	return computeChart(bz, kind, y, int(m), d)
}

// ComputeChartWithYongShen computes a奇门盘 and aggregates the用神
// (求测人 + 用神符号组合落宫) based on the given用神符号.
func ComputeChartWithYongShen(st tianwen.SolarTime, kind ChartKind, syms []YongShenSymbol, birthYear int) Chart {
	chart := ComputeChart(st, kind)
	ys := ComputeYongShen(chart, syms)
	if birthYear > 0 {
		ys = ComputeYongShenWithBirth(chart, syms, birthYear)
	}
	chart.YongShen = &ys
	return chart
}
