package bazi

import (
	"sort"
	"strings"

	"liki-engine/internal/engine/ganzhi"
)

// BranchCombination is a complete or partial branch group involving the flow
// year. Kind is san_he / san_hui / xing / half_he.
type BranchCombination struct {
	Kind         string   `json:"kind"`
	Group        string   `json:"group"`
	Branches     []string `json:"branches"`
	IncludesYear bool     `json:"includes_year"`
}

// LiuNianAtomicFacts are deterministic annual facts owned by the engine. The
// Python factor layer may resolve semantic targets but must not recompute the
// underlying five-element or branch rules.
type LiuNianAtomicFacts struct {
	ControlsElements          []string            `json:"controls_elements"`
	ControlsTargets           map[string]bool     `json:"controls_targets"`
	UnfavorableGan            bool                `json:"unfavorable_gan"`
	UnfavorableBranch         bool                `json:"unfavorable_branch"`
	WealthBreaksSeal          bool                `json:"wealth_breaks_seal"`
	Combinations              []BranchCombination `json:"combinations"`
	YearBranchRelations       map[string]string   `json:"year_branch_relations"`
	YearBranchControlledBy    []string            `json:"year_branch_controlled_by"`
	YearGanControlsDayGan     bool                `json:"year_gan_controls_day_gan"`
	DayGanControlsYearGan     bool                `json:"day_gan_controls_year_gan"`
	GanCombines               bool                `json:"gan_combines"`
	DayunGanCombines          bool                `json:"dayun_gan_combines"`
	DayunZhiClashesYear       bool                `json:"dayun_zhi_clashes_year"`
	DayVoidBranches           []string            `json:"day_void_branches"`
	YearEqualsDayPillar       bool                `json:"year_equals_day_pillar"`
	DayunEqualsYearPillar     bool                `json:"dayun_equals_year_pillar"`
	YearGanEqualsNatalYearGan bool                `json:"year_gan_equals_natal_year_gan"`
}

func computeLiuNianAtomicFacts(bz ganzhi.Bazi, yearGan ganzhi.Gan, yearZhi ganzhi.Zhi, yearShiShen ganzhi.ShiShen, currentDaYun *DaYunStep) LiuNianAtomicFacts {
	facts := LiuNianAtomicFacts{
		ControlsElements:       []string{},
		Combinations:           []BranchCombination{},
		YearBranchRelations:    map[string]string{},
		YearBranchControlledBy: []string{},
		DayVoidBranches:        []string{},
	}

	controls := map[ganzhi.Wuxing]bool{}
	for _, source := range []ganzhi.Wuxing{ganzhi.GanWuxing(yearGan), ganzhi.ZhiWuxing(yearZhi)} {
		for _, target := range []ganzhi.Wuxing{ganzhi.WxMu, ganzhi.WxHuo, ganzhi.WxTu, ganzhi.WxJin, ganzhi.WxShui} {
			if source != 0 && ganzhi.Ke(source, target) {
				controls[target] = true
			}
		}
	}
	for _, element := range []ganzhi.Wuxing{ganzhi.WxMu, ganzhi.WxHuo, ganzhi.WxTu, ganzhi.WxJin, ganzhi.WxShui} {
		if controls[element] {
			facts.ControlsElements = append(facts.ControlsElements, element.String())
		}
	}

	facts.ControlsTargets = controlledTargets(bz, controls)

	// ComputeYongShen is deterministic from the lean chart and keeps the
	// unfavorable-element rule in the engine rather than Python factors.
	chart := Chart{Nian: zhuInfo{Zhu: bz.Nian}, Yue: zhuInfo{Zhu: bz.Yue}, Ri: zhuInfo{Zhu: bz.Ri}, Shi: zhuInfo{Zhu: bz.Shi}}
	unfavorable := parseWuxingLabel(ComputeYongShen(chart).FuYi.Ji)
	if unfavorable != 0 {
		facts.UnfavorableGan = ganzhi.GanWuxing(yearGan) == unfavorable
		facts.UnfavorableBranch = ganzhi.ZhiWuxing(yearZhi) == unfavorable
	}

	if yearShiShen == ganzhi.ShiShenZhengYin || yearShiShen == ganzhi.ShiShenPianYin {
		facts.WealthBreaksSeal = ganzhi.Ke(ganzhi.ZhiWuxing(yearZhi), ganzhi.GanWuxing(yearGan))
	}

	dayElement := ganzhi.GanWuxing(bz.Ri.Gan)
	yearElement := ganzhi.GanWuxing(yearGan)
	facts.YearGanControlsDayGan = ganzhi.Ke(yearElement, dayElement)
	facts.DayGanControlsYearGan = ganzhi.Ke(dayElement, yearElement)
	facts.GanCombines = analyzeGanRelation(yearGan, bz.Ri.Gan).Type == relGanHe
	if currentDaYun != nil {
		facts.DayunGanCombines = analyzeGanRelation(currentDaYun.Gan, yearGan).Type == relGanHe
		facts.DayunZhiClashesYear = analyzeZhiRelation(currentDaYun.Zhi, yearZhi).Type == relLiuChong
	}
	facts.DayVoidBranches = dayVoidBranches(bz.Ri)
	facts.YearEqualsDayPillar = yearGan == bz.Ri.Gan && yearZhi == bz.Ri.Zhi
	if currentDaYun != nil {
		facts.DayunEqualsYearPillar = currentDaYun.Gan == yearGan && currentDaYun.Zhi == yearZhi
	}
	facts.YearGanEqualsNatalYearGan = yearGan == bz.Nian.Gan

	facts.Combinations = computeAnnualBranchCombinations(bz, yearZhi)
	facts.YearBranchRelations = annualBranchRelations(bz, yearZhi)
	facts.YearBranchControlledBy = annualBranchControlledBy(chart, yearZhi)
	return facts
}

func parseWuxingLabel(label string) ganzhi.Wuxing {
	for _, element := range []ganzhi.Wuxing{ganzhi.WxMu, ganzhi.WxHuo, ganzhi.WxTu, ganzhi.WxJin, ganzhi.WxShui} {
		if element.String() == label {
			return element
		}
	}
	return 0
}

func computeAnnualBranchCombinations(bz ganzhi.Bazi, yearZhi ganzhi.Zhi) []BranchCombination {
	natal := bz.Slice()
	counts := map[ganzhi.Zhi]int{}
	for _, pillar := range natal {
		counts[pillar.Zhi]++
	}
	counts[yearZhi]++
	containsYear := func(group []ganzhi.Zhi) bool {
		for _, branch := range group {
			if branch == yearZhi {
				return true
			}
		}
		return false
	}
	names := func(group []ganzhi.Zhi) []string {
		result := make([]string, 0, len(group))
		for _, branch := range group {
			result = append(result, ganzhi.ZhiName(branch))
		}
		return result
	}
	label := func(prefix string, group []ganzhi.Zhi) string {
		return prefix + strings.Join(names(group), "")
	}

	result := []BranchCombination{}
	appendCombination := func(kind, group string, branches []ganzhi.Zhi) {
		result = append(result, BranchCombination{
			Kind: kind, Group: group, Branches: names(branches), IncludesYear: containsYear(branches),
		})
	}
	for _, group := range ganzhi.TripleHeList {
		if groupComplete(counts, group.Zhi) {
			appendCombination("san_he", label("三合", group.Zhi), group.Zhi)
		} else if halfHeComplete(counts, group.Zhi) {
			// Preserve the canonical full-group label while listing only the
			// branches actually present in natal chart plus flow year.
			present := presentBranches(counts, group.Zhi)
			appendCombination("half_he", label("三合", group.Zhi), present)
		}
	}
	for _, group := range ganzhi.TripleHuiList {
		if groupComplete(counts, group.Zhi) {
			appendCombination("san_hui", label("三会", group.Zhi), group.Zhi)
		}
	}
	for _, group := range ganzhi.XingGroups {
		// Self-punishment groups contain four candidate branches; any one
		// branch appearing twice forms its own punishment, not all four.
		if len(group.Zhi) == 4 { // 辰午酉亥自刑组：任一同字成双即成立。
			for _, branch := range group.Zhi {
				if counts[branch] >= 2 {
					appendCombination("xing", label("三刑", []ganzhi.Zhi{branch}), []ganzhi.Zhi{branch, branch})
				}
			}
			continue
		}
		if groupComplete(counts, group.Zhi) {
			appendCombination("xing", label("三刑", group.Zhi), group.Zhi)
		}
	}
	sort.Slice(result, func(i, j int) bool {
		if result[i].Kind != result[j].Kind {
			return result[i].Kind < result[j].Kind
		}
		return result[i].Group < result[j].Group
	})
	return result
}

func groupComplete(counts map[ganzhi.Zhi]int, group []ganzhi.Zhi) bool {
	for _, branch := range group {
		if counts[branch] == 0 {
			return false
		}
	}
	return true
}

func halfHeComplete(counts map[ganzhi.Zhi]int, group []ganzhi.Zhi) bool {
	// A conventional half combination must contain the imperial branch
	// (子 / 午 / 卯 / 酉) plus one other member of the same trine.
	imperial := map[ganzhi.Zhi]bool{ganzhi.ZhiZi: true, ganzhi.ZhiWu: true, ganzhi.ZhiMao: true, ganzhi.ZhiYou: true}
	hasImperial := false
	present := 0
	for _, branch := range group {
		if counts[branch] > 0 {
			present++
			hasImperial = hasImperial || imperial[branch]
		}
	}
	return hasImperial && present >= 2
}

func presentBranches(counts map[ganzhi.Zhi]int, group []ganzhi.Zhi) []ganzhi.Zhi {
	var result []ganzhi.Zhi
	for _, branch := range group {
		if counts[branch] > 0 {
			result = append(result, branch)
		}
	}
	return result
}

func annualBranchRelations(bz ganzhi.Bazi, yearZhi ganzhi.Zhi) map[string]string {
	result := map[string]string{}
	for _, pillar := range bz.Slice() {
		relation := analyzeZhiRelation(yearZhi, pillar.Zhi)
		if relation.Type == relNone || relation.Type == relSame {
			continue
		}
		label := ""
		switch relation.Type {
		case relLiuHe:
			label = "liu_he"
		case relSanHe:
			label = "san_he"
		case relSanHui:
			label = "san_hui"
		case relLiuChong:
			label = "liu_chong"
		case relXing:
			label = "xing"
		case relLiuHai:
			label = "liu_hai"
		}
		if label != "" {
			result[ganzhi.ZhiName(pillar.Zhi)] = label
		}
	}
	return result
}

func annualBranchControlledBy(chart Chart, yearZhi ganzhi.Zhi) []string {
	target := ganzhi.ZhiWuxing(yearZhi)
	if target == 0 {
		return []string{}
	}
	states := ComputeYongShen(chart).FuYi.WangShuai
	result := make([]string, 0, 2)
	for _, element := range []ganzhi.Wuxing{ganzhi.WxMu, ganzhi.WxHuo, ganzhi.WxTu, ganzhi.WxJin, ganzhi.WxShui} {
		state := states[element.String()]
		if (state == ganzhi.WSWang.String() || state == ganzhi.WSXiang.String()) && ganzhi.Ke(element, target) {
			result = append(result, element.String())
		}
	}
	return result
}

func dayVoidBranches(day ganzhi.Zhu) []string {
	branches := ganzhi.XunKong(day.Gan, day.Zhi)
	return []string{ganzhi.ZhiName(branches[0]), ganzhi.ZhiName(branches[1])}
}

func controlledTargets(bz ganzhi.Bazi, controls map[ganzhi.Wuxing]bool) map[string]bool {
	day := ganzhi.GanWuxing(bz.Ri.Gan)
	return map[string]bool{
		"day_master":      controls[day],
		"wealth_star":     controls[controlledElement(day)],
		"officer_killing": controls[controllingElement(day)],
		"seal_star":       controls[generatingElement(day)],
		"food_injury":     controls[generatedElement(day)],
	}
}

func controlledElement(day ganzhi.Wuxing) ganzhi.Wuxing {
	for _, target := range []ganzhi.Wuxing{ganzhi.WxMu, ganzhi.WxHuo, ganzhi.WxTu, ganzhi.WxJin, ganzhi.WxShui} {
		if ganzhi.Ke(day, target) {
			return target
		}
	}
	return 0
}

func controllingElement(day ganzhi.Wuxing) ganzhi.Wuxing {
	for _, source := range []ganzhi.Wuxing{ganzhi.WxMu, ganzhi.WxHuo, ganzhi.WxTu, ganzhi.WxJin, ganzhi.WxShui} {
		if ganzhi.Ke(source, day) {
			return source
		}
	}
	return 0
}

func generatingElement(day ganzhi.Wuxing) ganzhi.Wuxing {
	for _, source := range []ganzhi.Wuxing{ganzhi.WxMu, ganzhi.WxHuo, ganzhi.WxTu, ganzhi.WxJin, ganzhi.WxShui} {
		if ganzhi.Sheng(source, day) {
			return source
		}
	}
	return 0
}

func generatedElement(day ganzhi.Wuxing) ganzhi.Wuxing {
	for _, target := range []ganzhi.Wuxing{ganzhi.WxMu, ganzhi.WxHuo, ganzhi.WxTu, ganzhi.WxJin, ganzhi.WxShui} {
		if ganzhi.Sheng(day, target) {
			return target
		}
	}
	return 0
}
