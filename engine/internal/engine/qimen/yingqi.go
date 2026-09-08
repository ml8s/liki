package qimen

import (
	"strings"
	"time"

	"liki-engine/internal/engine/ganzhi"
	"liki-engine/internal/engine/tianwen"
)

type RelatedSymbol struct {
	Symbol string `json:"symbol"`
	Role   string `json:"role"`
}

type YingQiCandidate struct {
	Type        string          `json:"type"`
	Mechanism   string          `json:"mechanism"`
	Branch      string          `json:"branch"`
	Gong        GongIndex       `json:"gong"`
	RelatedTo   []RelatedSymbol `json:"related_to"`
	Dates       []YingQiDate    `json:"dates"`
	Description string          `json:"description"`
}

type YingQiDate struct {
	Date   string `json:"date"`
	Branch string `json:"branch"`
	Match  string `json:"match"`
	Reason string `json:"reason"`
}

type YingQi struct {
	Candidates []YingQiCandidate `json:"candidates"`
	Summary    string            `json:"summary"`
}

func computeYingQi(chart Chart, chartTime time.Time, focuses map[GongIndex][]RelatedSymbol) YingQi {
	var result YingQi
	days := yingQiDayWindow(chart, chartTime)
	appendCandidate := func(ruleType string, branchPalace BranchPalace) {
		rule := yingqiRuleTable[ruleType]
		result.Candidates = append(result.Candidates, YingQiCandidate{
			Type: ruleType, Mechanism: rule.Mechanism,
			Branch: ganzhi.ZhiName(branchPalace.Branch), Gong: branchPalace.Gong,
			RelatedTo:   relatedSymbols(branchPalace.Gong, focuses),
			Dates:       yingQiDates(ruleType, branchPalace.Branch, days),
			Description: rule.Description,
		})
	}
	appendCandidate("ma_xing", chart.Pan.MaXing)
	for _, voidBranch := range chart.Pan.KongWang {
		if voidBranch.Branch == 0 {
			continue
		}
		appendCandidate("kong_wang", voidBranch)
	}
	result.Summary = yingqiSummary
	return result
}

type yingQiDay struct {
	Date   time.Time
	Branch ganzhi.Zhi
}

func yingQiDayWindow(chart Chart, chartTime time.Time) []yingQiDay {
	anchor := yingQiAnchorDate(chart, chartTime)
	result := make([]yingQiDay, 0, yingqiHorizonDays)
	for offset := 0; offset < yingqiHorizonDays; offset++ {
		date := anchor.AddDate(0, 0, offset)
		result = append(result, yingQiDay{
			Date:   date,
			Branch: tianwen.RiZhu(tianwen.GregorianTime(date)).Zhi,
		})
	}
	return result
}

func yingQiDates(ruleType string, target ganzhi.Zhi, days []yingQiDay) []YingQiDate {
	matches := yingqiDateMatches[ruleType]
	if len(matches) == 0 {
		return []YingQiDate{}
	}
	result := []YingQiDate{}
	for _, day := range days {
		branch := day.Branch
		targetName := ganzhi.ZhiName(target)
		branchName := ganzhi.ZhiName(branch)
		for _, match := range matches {
			matched := (match.Match == "same" && branch == target) ||
				(match.Match == "opposite" && branch == oppositeZhi(target))
			if !matched {
				continue
			}
			result = append(result, YingQiDate{
				Date:   day.Date.Format("2006-01-02"),
				Branch: branchName,
				Match:  match.Match,
				Reason: formatYingQiReason(match.Reason, branchName, targetName),
			})
		}
	}
	return result
}

func yingQiAnchorDate(chart Chart, chartTime time.Time) time.Time {
	base := civilDate(chartTime)
	for _, offset := range []int{0, 1, -1} {
		date := base.AddDate(0, 0, offset)
		pillar := tianwen.RiZhu(tianwen.GregorianTime(date))
		if pillar.Gan == chart.Pan.RiGan && pillar.Zhi == chart.Pan.RiZhi {
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

func formatYingQiReason(template, branch, target string) string {
	return strings.NewReplacer("{branch}", branch, "{target}", target).Replace(template)
}

func baseYingQiFocuses(chart Chart) map[GongIndex][]RelatedSymbol {
	result := map[GongIndex][]RelatedSymbol{}
	addFocus(result, chart.RiGanPalace, RelatedSymbol{Symbol: "日干", Role: "pillar"})
	addFocus(result, chart.ShiGanPalace, RelatedSymbol{Symbol: "时干", Role: "pillar"})
	addFocus(result, leadPillarPalace(chart.Pan), RelatedSymbol{Symbol: "主柱", Role: "lead"})
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
