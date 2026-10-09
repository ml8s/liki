package bazi

import (
	"liki-engine/internal/engine/ganzhi"
)

// TiaoHou is the internal 穷通宝鉴 climate-adjustment result.
type TiaoHou struct {
	Season             string
	PrimaryWuxing      string
	SecondaryWuxing    string
	Detail             string
	SecondaryCondition string
}

// tiaohouKey is the internal compound key for the lookup table.
type tiaohouKey struct {
	gan int
	zhi int
}

// computeTiaoHou returns the TiaoHou (调候) yongshen analysis for a natal
// chart. The table is keyed by day-master and month-zhi; the availability
// projection is computed against all four palaces.
func computeTiaoHou(c Chart) TiaoHouResult {
	riYuan, yueZhi := c.Ri.Gan, c.Yue.Zhi
	th, _ := queryTiaoHou(riYuan, yueZhi)
	entry, hasEntry := lookupTiaohou[tiaohouKey{int(riYuan), int(yueZhi)}]
	if !hasEntry {
		return TiaoHouResult{
			Season:  th.Season,
			Model:   "qiongtong_primary_secondary_table",
			Primary: TiaoHouAvailability{Occurrences: []StemOccurrence{}, RelationFacts: []RelationFact{}},
		}
	}
	return TiaoHouResult{
		PrimaryWuxing:      th.PrimaryWuxing,
		SecondaryWuxing:    th.SecondaryWuxing,
		Season:             th.Season,
		Detail:             th.Detail,
		Model:              "qiongtong_primary_secondary_table",
		Primary:            buildTiaoHouAvailability(c, entry.primary, "primary"),
		Secondary:          tiaoHouSecondary(c, entry.secondary),
		SecondaryCondition: th.SecondaryCondition,
	}
}

func tiaoHouSecondary(c Chart, secondary ganzhi.Gan) *TiaoHouAvailability {
	if secondary == 0 {
		return nil
	}
	availability := buildTiaoHouAvailability(c, secondary, "secondary")
	return &availability
}

// queryTiaoHou returns the 穷通宝鉴 climate-adjustment result for a given
// day-master and month-zhi. Returns (TiaoHou, true) on match, or
// (zero, false) if no entry exists.
func queryTiaoHou(riYuan ganzhi.Gan, yueZhi ganzhi.Zhi) (TiaoHou, bool) {
	e, ok := lookupTiaohou[tiaohouKey{int(riYuan), int(yueZhi)}]
	if !ok {
		return TiaoHou{}, false
	}

	primaryElem := ganzhi.GanWuxing(e.primary)
	var secondaryElem ganzhi.Wuxing
	if e.secondary != 0 {
		secondaryElem = ganzhi.GanWuxing(e.secondary)
	}

	season := ganzhi.ZhiSeasonLabel(yueZhi)

	detail := ganzhi.ZhiName(yueZhi) + "月" + ganzhi.GanName(riYuan) + ganzhi.GanWuxing(riYuan).String()
	detail += "，用" + ganzhi.GanName(e.primary) + "调候"
	if e.secondary != 0 {
		if e.secondaryCondition != "" {
			detail += "，" + e.secondaryCondition
		} else {
			detail += "，" + ganzhi.GanName(e.secondary) + "辅之"
		}
	}

	secondaryWuxing := ""
	if e.secondary != 0 {
		secondaryWuxing = secondaryElem.String()
	}

	return TiaoHou{
		Season:             season,
		PrimaryWuxing:      primaryElem.String(),
		SecondaryWuxing:    secondaryWuxing,
		Detail:             detail,
		SecondaryCondition: e.secondaryCondition,
	}, true
}
