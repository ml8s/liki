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
// TiaoHouAt is the explicit domain entry for climate-adjustment evidence.
func TiaoHouAt(bz ganzhi.Bazi, relations []NatalRelationFact) TiaoHouResult {
	return computeTiaoHou(bz, relations)
}

func computeTiaoHou(bz ganzhi.Bazi, relations []NatalRelationFact) TiaoHouResult {
	riYuan, yueZhi := bz.Ri.Gan, bz.Yue.Zhi
	th, _ := queryTiaoHou(riYuan, yueZhi)
	entry, hasEntry := lookupTiaohou[tiaohouKey{int(riYuan), int(yueZhi)}]
	if !hasEntry {
		return TiaoHouResult{
			Season:  th.Season,
			ModelID: "qiongtong_primary_secondary_table",
			Primary: TiaoHouAvailability{Occurrences: []StemOccurrence{}, GuanXiLinks: []RelationLink{}},
		}
	}
	return TiaoHouResult{
		PrimaryWuxing:      th.PrimaryWuxing,
		SecondaryWuxing:    th.SecondaryWuxing,
		Season:             th.Season,
		TiaoHouExplanation: th.Detail,
		ModelID:            "qiongtong_primary_secondary_table",
		Primary:            buildTiaoHouAvailability(bz, entry.primary, "primary", relations),
		Secondary:          tiaoHouSecondary(bz, entry.secondary, relations),
		SecondaryCondition: th.SecondaryCondition,
	}
}

func tiaoHouSecondary(bz ganzhi.Bazi, secondary ganzhi.Gan, relations []NatalRelationFact) *TiaoHouAvailability {
	if secondary == 0 {
		return nil
	}
	availability := buildTiaoHouAvailability(bz, secondary, "secondary", relations)
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

	if e.secondary != 0 {
		secondaryWuxing := secondaryElem.String()
		return TiaoHou{
			Season:             season,
			PrimaryWuxing:      primaryElem.String(),
			SecondaryWuxing:    secondaryWuxing,
			Detail:             detail,
			SecondaryCondition: e.secondaryCondition,
		}, true
	}
	return TiaoHou{
		Season:             season,
		PrimaryWuxing:      primaryElem.String(),
		SecondaryWuxing:    NoSecondaryWuxing,
		Detail:             detail,
		SecondaryCondition: e.secondaryCondition,
	}, true
}
