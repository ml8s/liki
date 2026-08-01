// Package qimen provides 奇门遁甲 computation.
//
// Types
//
//	Chart, Gong
//	GongIndex, StarIndex, DoorIndex, SpiritIndex
//	GanInteraction, MenInteraction, XingInteraction
//	WangShuai, Pattern, YingQi
//	ChartKind, EventKind
//
// Constants
//
//	GongKan .. GongLi  (九宫)
//	StarTianPeng .. StarTianYing  (九星)
//	DoorXiu .. DoorKai  (八门)
//	SpiritZhiFu .. SpiritJiuTian  (八神)
//	ShiQiMen, RiQiMen, YueQiMen, NianQiMen  (奇门种类)
//	EventGeneral, EventCareer, EventWealth, EventRelationship, EventStudy, EventTravel, EventHealth, EventLegal  (事件类型)
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

// EventKind represents the type of event for Qimen judgment/selection.
type EventKind string

const (
	EventGeneral      EventKind = "general"
	EventCareer       EventKind = "career"
	EventWealth       EventKind = "wealth"
	EventRelationship EventKind = "relationship"
	EventStudy        EventKind = "study"
	EventTravel       EventKind = "travel"
	EventHealth       EventKind = "health"
	EventLegal        EventKind = "legal"
)

// ParseEventKind parses an event kind string.
func ParseEventKind(s string) (EventKind, error) {
	switch s {
	case "general":
		return EventGeneral, nil
	case "career":
		return EventCareer, nil
	case "wealth":
		return EventWealth, nil
	case "relationship":
		return EventRelationship, nil
	case "study":
		return EventStudy, nil
	case "travel":
		return EventTravel, nil
	case "health":
		return EventHealth, nil
	case "legal":
		return EventLegal, nil
	}
	return "", fmt.Errorf("invalid event kind %q", s)
}

// ComputeChart computes a complete奇门盘 with all analyses.
func ComputeChart(st tianwen.SolarTime, kind ChartKind) Chart {
	bz := tianwen.ComputeBazi(st)
	t := st.Time()
	y, m, d := t.Date()
	return computeChart(bz, kind, y, int(m), d)
}
