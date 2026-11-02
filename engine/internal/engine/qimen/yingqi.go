package qimen

import (
	"time"

	"liki-engine/internal/engine/ganzhi"
	"liki-engine/internal/engine/tianwen"
)

type RelatedSymbol struct {
	Symbol string `json:"symbol"`
	Role   string `json:"role"`
}

type YingQiCandidate struct {
	Type      string          `json:"ying_qi_kind"`
	Mechanism string          `json:"mechanism"`
	Branch    string          `json:"branch"`
	Gong      GongIndex       `json:"gong"`
	RelatedTo []RelatedSymbol `json:"related_to"`
	Dates     []YingQiDate    `json:"dates"`
}

type YingQiDate struct {
	Date   string `json:"trigger_date"`
	Branch string `json:"branch"`
	Match  string `json:"match"`
}

type YingQi struct {
	Candidates []YingQiCandidate `json:"candidates"`
}

func computeYingQi(chart Chart, chartTime time.Time, focuses map[GongIndex][]RelatedSymbol) YingQi {
	var result YingQi
	window := yingQiDayWindow(chart, chartTime)
	appendCandidate := func(ruleType string, branchPalace BranchPalace) {
		rule := yingqiRuleTable[ruleType]
		result.Candidates = append(result.Candidates, YingQiCandidate{
			Type: ruleType, Mechanism: rule.Mechanism,
			Branch: ganzhi.ZhiName(branchPalace.Branch), Gong: branchPalace.Gong,
			RelatedTo: relatedSymbols(branchPalace.Gong, focuses),
			Dates:     yingQiDates(ruleType, branchPalace.Branch, window),
		})
	}
	if chart.Board.MaXing.Branch != 0 {
		appendCandidate("ma_xing", chart.Board.MaXing)
	}
	if branch := palaceZhi(chart.DutyStarPalace); branch != 0 {
		appendCandidate("duty_star", BranchPalace{Branch: branch, Gong: chart.DutyStarPalace})
	}
	if branch := palaceZhi(chart.DutyDoorPalace); branch != 0 {
		appendCandidate("duty_door", BranchPalace{Branch: branch, Gong: chart.DutyDoorPalace})
	}
	for _, voidBranch := range chart.Board.KongWang {
		if voidBranch.Branch == 0 {
			continue
		}
		appendCandidate("kong_wang", voidBranch)
	}
	return result
}

type yingQiWindow struct {
	Anchor       time.Time
	AnchorBranch ganzhi.Zhi
	Horizon      int
}

func yingQiDayWindow(chart Chart, chartTime time.Time) yingQiWindow {
	anchor := yingQiAnchorDate(chart, chartTime)
	return yingQiWindow{
		Anchor:       anchor,
		AnchorBranch: tianwen.RiZhu(time.Time(anchor)).Zhi,
		Horizon:      yingqiHorizonDays,
	}
}

func yingQiDates(ruleType string, target ganzhi.Zhi, window yingQiWindow) []YingQiDate {
	matches := yingqiDateMatches[ruleType]
	if len(matches) == 0 {
		return []YingQiDate{}
	}
	sameOffset, oppositeOffset := -1, -1
	var sameMatch, oppositeMatch *yingqiDateMatch
	for index := range matches {
		match := &matches[index]
		wanted := target
		if match.Match == "opposite" {
			wanted = oppositeZhi(target)
		}
		offset := (int(wanted) - int(window.AnchorBranch) + 12) % 12
		if match.Match == "same" {
			sameOffset, sameMatch = offset, match
		} else {
			oppositeOffset, oppositeMatch = offset, match
		}
	}
	step := 12
	start := sameOffset
	if sameOffset < 0 {
		start = oppositeOffset
	}
	if sameOffset >= 0 && oppositeOffset >= 0 {
		step = 6
		start = min(sameOffset, oppositeOffset)
	}
	result := make([]YingQiDate, 0, window.Horizon/step+1)
	for offset := start; offset < window.Horizon; offset += step {
		match, wanted := sameMatch, target
		if oppositeMatch != nil && (sameMatch == nil || (offset-sameOffset+12)%12 != 0) {
			match, wanted = oppositeMatch, oppositeZhi(target)
		}
		date := window.Anchor.AddDate(0, 0, offset)
		result = append(result, YingQiDate{
			Date:   date.Format("2006-01-02"),
			Branch: ganzhi.ZhiName(wanted),
			Match:  match.Match,
		})
	}
	return result
}

func yingQiAnchorDate(chart Chart, chartTime time.Time) time.Time {
	base := civilDate(chartTime)
	for _, offset := range []int{0, 1, -1} {
		date := base.AddDate(0, 0, offset)
		pillar := tianwen.RiZhu(time.Time(date))
		if pillar.Gan == chart.Board.RiGan && pillar.Zhi == chart.Board.RiZhi {
			return date
		}
	}
	return base
}

func civilDate(value time.Time) time.Time {
	year, month, day := value.Date()
	return time.Date(year, month, day, 0, 0, 0, 0, value.Location())
}

func oppositeZhi(value ganzhi.Zhi) ganzhi.Zhi {
	return ganzhi.Zhi((int(value)+5)%12 + 1)
}

func baseYingQiFocuses(chart Chart) map[GongIndex][]RelatedSymbol {
	result := map[GongIndex][]RelatedSymbol{}
	addFocus(result, chart.DayStemPalace.Palace, RelatedSymbol{Symbol: "日干", Role: "pillar"})
	addFocus(result, chart.HourStemPalace.Palace, RelatedSymbol{Symbol: "时干", Role: "pillar"})
	addFocus(result, leadPillarPalace(chart.Board), RelatedSymbol{Symbol: "主柱", Role: "lead"})
	if chart.DutyStarPalace != 0 {
		addFocus(result, chart.DutyStarPalace, RelatedSymbol{Symbol: "值符星", Role: "duty"})
	}
	if chart.DutyDoorPalace != 0 {
		addFocus(result, chart.DutyDoorPalace, RelatedSymbol{Symbol: "值使门", Role: "duty"})
	}
	return result
}

func yingQiFocuses(chart Chart, yongShen YongShenResult) map[GongIndex][]RelatedSymbol {
	focuses := baseYingQiFocuses(chart)
	if yongShen.NianGanPalace != nil {
		addFocus(focuses, *yongShen.NianGanPalace, RelatedSymbol{Symbol: "年命干", Role: "pillar"})
	}
	for _, symbol := range yongShen.Symbols {
		if symbol.Palace == 0 {
			continue
		}
		addFocus(focuses, symbol.Palace, RelatedSymbol{Symbol: symbol.Symbol, Role: "yong_shen"})
	}
	return focuses
}

func addFocus(focuses map[GongIndex][]RelatedSymbol, palace GongIndex, symbol RelatedSymbol) {
	for _, existing := range focuses[palace] {
		if existing == symbol {
			return
		}
	}
	focuses[palace] = append(focuses[palace], symbol)
}

func relatedSymbols(palace GongIndex, focuses map[GongIndex][]RelatedSymbol) []RelatedSymbol {
	if len(focuses[palace]) == 0 {
		return []RelatedSymbol{}
	}
	return focuses[palace]
}
