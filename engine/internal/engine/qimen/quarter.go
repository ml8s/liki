package qimen

import (
	"time"

	"liki-engine/internal/engine/ganzhi"
)

type quarterInfo struct {
	Index          int
	Minutes        int
	LeadPillarMode string
	DunSource      string
	HourBoundary   string
	Pillar         ganzhi.Zhu
	HourYangDun    bool
	Sanyuan        quarterSanyuanEntry
}

// ChartQuarter describes the selected quarter. The lead pillar itself remains
// method.lead_pillar to avoid a duplicate API fact.
type ChartQuarter struct {
	Index             int    `json:"index"`
	MinutesPerQuarter int    `json:"minutes_per_quarter"`
	LeadPillarMode    string `json:"lead_pillar_mode"`
	DunSource         string `json:"dun_source,omitempty"`
	HourBoundary      string `json:"hour_boundary,omitempty"`
}

func newTenMinuteQuarterInfo(t time.Time, hourPillar ganzhi.Zhu) quarterInfo {
	rule := quarterRules[QuarterTenMinuteSanYuan]
	index := quarterIndexFor(t, rule, 1)
	hourIndex := ganzhi.SixtyCycleIndex(hourPillar.Gan, hourPillar.Zhi)
	fuTou := ganzhi.SixtyToZhu(hourIndex - hourIndex%quarterFuTouUnits)
	sanyuan := quarterSanyuanByBranch[fuTou.Zhi]
	return quarterInfo{
		Index:          index,
		Minutes:        rule.Minutes,
		LeadPillarMode: rule.LeadPillarMode,
		Pillar:         quarterPillarAt(hourPillar, index),
		HourYangDun:    quarterYangBranch[hourPillar.Zhi],
		Sanyuan:        sanyuan,
	}
}

func newTwelveMinuteQuarterInfo(t time.Time, hourPillar ganzhi.Zhu, baseYangDun bool, method Method) quarterInfo {
	rule := quarterRules[QuarterTwelveMinuteTenDivision]
	hourYangDun := quarterYangBranch[hourPillar.Zhi]
	if method.DunSource == "solar_term" {
		hourYangDun = baseYangDun
	}
	return quarterInfo{
		Index:          twelveMinuteQuarterIndexWithBoundary(t, method.HourBoundary),
		Minutes:        rule.Minutes,
		LeadPillarMode: rule.LeadPillarMode,
		DunSource:      method.DunSource,
		HourBoundary:   method.HourBoundary,
		Pillar:         hourPillar,
		HourYangDun:    hourYangDun,
	}
}

func quarterIndex(t time.Time) int {
	return quarterIndexFor(t, quarterRules[QuarterTenMinuteSanYuan], 1)
}

func twelveMinuteQuarterIndex(t time.Time) int {
	return twelveMinuteQuarterIndexWithBoundary(
		t, quarterRules[QuarterTwelveMinuteTenDivision].DefaultHourBoundary,
	)
}

func twelveMinuteQuarterIndexWithBoundary(t time.Time, boundary string) int {
	rule := quarterRules[QuarterTwelveMinuteTenDivision]
	return quarterIndexFor(t, rule, rule.HourBoundaries[boundary])
}

func quarterIndexFor(t time.Time, rule quarterVariant, hourStartShift int) int {
	sinceShichenStart := ((t.Hour()+hourStartShift)%2)*60 + t.Minute()
	return min(rule.Count, sinceShichenStart/rule.Minutes+1)
}

func quarterPillarAt(hourPillar ganzhi.Zhu, index int) ganzhi.Zhu {
	start := quarterStartIndex[hourPillar.Gan]
	return ganzhi.SixtyToZhu(start + index - 1)
}

func twelveMinuteShiftJu(baseJu, index int, yang bool) int {
	rule := quarterRules[QuarterTwelveMinuteTenDivision]
	step := rule.YinStep
	if yang {
		step = rule.YangStep
	}
	return floorMod(baseJu-1+step*(index-1), rule.JuCycle) + 1
}
