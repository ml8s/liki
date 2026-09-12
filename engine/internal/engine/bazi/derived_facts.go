package bazi

import (
	"sort"
	"strings"

	"liki-engine/internal/engine/ganzhi"
)

// LuRoot records a visible stem whose traditional Lu branch appears in one of
// the four palaces. Hidden stems are not promoted to Lu roots.
type LuRoot struct {
	ShiShen string `json:"shi_shen"`
	Gan     string `json:"gan"`
	Zhi     string `json:"zhi"`
}

// RelationGroup is a complete natal relation group. Pair relations are
// complete when present; triple groups require every member branch.
type RelationGroup struct {
	Field string `json:"field"`
	Group string `json:"group"`
}

// TenGodState aggregates all visible and hidden occurrences of one ten god.
// Rooted, timely and strength follow the underlying element; transparent and
// hidden describe this specific ten god.
type TenGodState struct {
	ShiShen     string `json:"shi_shen"`
	Wuxing      string `json:"wuxing"`
	Transparent bool   `json:"transparent"`
	Hidden      bool   `json:"hidden"`
	Rooted      bool   `json:"rooted"`
	Timely      bool   `json:"timely"`
	Count       int    `json:"count"`
	Strength    string `json:"strength"`
}

// ElementState records engine-owned seasonal, rooted and control facts for one
// element. Combined strength is strong when timely, or transparent and rooted;
// weak only when untimely, not transparent and rootless.
type ElementState struct {
	Wuxing             string   `json:"wuxing"`
	SeasonStrength     string   `json:"season_strength"`
	Strength           string   `json:"strength"`
	Transparent        bool     `json:"transparent"`
	Rooted             bool     `json:"rooted"`
	Controls           string   `json:"controls"`
	ControlledBy       string   `json:"controlled_by"`
	ControllerStrength string   `json:"controller_strength"`
	Reasons            []string `json:"reasons"`
}

type tenGodAccumulator struct {
	wuxing      ganzhi.Wuxing
	transparent bool
	hidden      bool
	count       int
}

func computeTenGodStates(c FullChart) ([]TenGodState, map[string]ElementState) {
	pillars := [4]fullZhuInfo{c.Nian, c.Yue, c.Ri, c.Shi}
	transparent := map[ganzhi.Wuxing]bool{}
	rooted := map[ganzhi.Wuxing]bool{}
	accumulators := map[ganzhi.ShiShen]*tenGodAccumulator{}

	for _, pillar := range pillars {
		transparent[ganzhi.GanWuxing(pillar.Gan)] = true
		hiddenStems := []ganzhi.Gan{pillar.CangGan.Main}
		if pillar.CangGan.Mid != nil {
			hiddenStems = append(hiddenStems, *pillar.CangGan.Mid)
		}
		if pillar.CangGan.Minor != nil {
			hiddenStems = append(hiddenStems, *pillar.CangGan.Minor)
		}
		for _, gan := range hiddenStems {
			if gan != 0 {
				rooted[ganzhi.GanWuxing(gan)] = true
			}
		}
		for _, item := range pillar.ShiShens {
			if item.Source != sourceGan && !hiddenTenGodSource(item.Source) {
				continue
			}
			state := accumulators[item.ShiShen]
			if state == nil {
				state = &tenGodAccumulator{wuxing: ganzhi.GanWuxing(item.Gan)}
				accumulators[item.ShiShen] = state
			}
			if item.Source == sourceGan {
				state.transparent = true
			} else {
				state.hidden = true
			}
			state.count++
		}
	}

	elements := computeElementStates(c, transparent, rooted)
	states := make([]TenGodState, 0, len(accumulators))
	for shiShen := ganzhi.ShiShen(0); shiShen <= ganzhi.ShiShenZhengYin; shiShen++ {
		accumulator := accumulators[shiShen]
		if accumulator == nil {
			continue
		}
		element := elements[accumulator.wuxing.String()]
		strength := "neutral"
		if element.SeasonStrength == "strong" {
			strength = "strong"
		} else if accumulator.transparent && element.Rooted {
			strength = "strong"
		} else if element.SeasonStrength == "weak" && !accumulator.transparent && !element.Rooted {
			strength = "weak"
		}
		states = append(states, TenGodState{
			ShiShen:     shiShen.String(),
			Wuxing:      accumulator.wuxing.String(),
			Transparent: accumulator.transparent,
			Hidden:      accumulator.hidden,
			Rooted:      element.Rooted,
			Timely:      element.SeasonStrength == "strong",
			Count:       accumulator.count,
			Strength:    strength,
		})
	}
	return states, elements
}

func hiddenTenGodSource(source string) bool {
	return source == sourceMainQi || source == sourceMidQi || source == sourceMinQi
}

func computeElementStates(c FullChart, transparent, rooted map[ganzhi.Wuxing]bool) map[string]ElementState {
	all := []ganzhi.Wuxing{ganzhi.WxMu, ganzhi.WxHuo, ganzhi.WxTu, ganzhi.WxJin, ganzhi.WxShui}
	result := make(map[string]ElementState, len(all))
	for _, element := range all {
		state := ElementState{
			Wuxing:       element.String(),
			Transparent:  transparent[element],
			Rooted:       rooted[element],
			Controls:     controlledElement(element).String(),
			ControlledBy: controllingElement(element).String(),
			Reasons:      []string{},
		}
		switch c.YongShen.FuYi.WangShuai[element.String()] {
		case ganzhi.WSWang.String(), ganzhi.WSXiang.String():
			state.SeasonStrength = "strong"
		case ganzhi.WSXiu.String(), ganzhi.WSQiu.String(), ganzhi.WSSi.String():
			state.SeasonStrength = "weak"
		}
		switch {
		case state.SeasonStrength == "strong":
			state.Strength = "strong"
			state.Reasons = append(state.Reasons, "timely")
		case state.Transparent && state.Rooted:
			state.Strength = "strong"
			state.Reasons = append(state.Reasons, "transparent_rooted")
		case state.SeasonStrength == "weak" && !state.Transparent && !state.Rooted:
			state.Strength = "weak"
			state.Reasons = append(state.Reasons, "untimely", "not_transparent", "not_rooted")
		default:
			state.Strength = "neutral"
		}
		result[state.Wuxing] = state
	}
	for _, element := range all {
		state := result[element.String()]
		state.ControllerStrength = result[state.ControlledBy].SeasonStrength
		result[state.Wuxing] = state
	}
	return result
}

func elementStateList(states map[string]ElementState) []ElementState {
	result := make([]ElementState, 0, len(states))
	for _, element := range []ganzhi.Wuxing{ganzhi.WxMu, ganzhi.WxHuo, ganzhi.WxTu, ganzhi.WxJin, ganzhi.WxShui} {
		result = append(result, states[element.String()])
	}
	return result
}

func computeLuRoots(c FullChart) []LuRoot {
	roots := make([]LuRoot, 0, 4)
	pillars := [4]fullZhuInfo{c.Nian, c.Yue, c.Ri, c.Shi}
	branches := map[ganzhi.Zhi]bool{}
	for _, pillar := range pillars {
		branches[pillar.Zhi] = true
	}
	for _, pillar := range pillars {
		for _, item := range pillar.ShiShens {
			if item.Source != sourceGan || !branches[ganzhi.LuZhi(item.Gan)] {
				continue
			}
			roots = append(roots, LuRoot{
				ShiShen: item.ShiShen.String(),
				Gan:     ganzhi.GanName(item.Gan),
				Zhi:     ganzhi.ZhiName(ganzhi.LuZhi(item.Gan)),
			})
		}
	}
	return roots
}

// AtomicFacts are deterministic natal facts consumed by the Python factor
// layer. They are engine-owned domain conclusions; Python must not recompute.
type AtomicFacts struct {
	DayMasterElement      string          `json:"day_master_element"`
	DayMasterStem         string          `json:"day_master_stem"`
	DayBranch             string          `json:"day_branch"`
	MonthLongevity        string          `json:"month_longevity"`
	YearStemTenGod        string          `json:"year_stem_ten_god"`
	MonthMainTenGod       string          `json:"month_main_ten_god"`
	HourStemTenGod        string          `json:"hour_stem_ten_god"`
	PatternGodTransparent bool            `json:"pattern_god_transparent"`
	PillarPunishments     map[string]bool `json:"pillar_punishments"`
	OfficerKillingCleaned bool            `json:"officer_killing_cleaned"`
	WealthTombPresent     bool            `json:"wealth_tomb_present"`
	WealthStarInTomb      bool            `json:"wealth_star_in_tomb"`
	SpousePalaceState     string          `json:"spouse_palace_state"`
	DayBranchType         string          `json:"day_branch_type"`
	YearOfficerKilling    bool            `json:"year_officer_killing"`
}

func wealthTomb(element ganzhi.Wuxing) ganzhi.Zhi {
	switch element {
	case ganzhi.WxJin:
		return ganzhi.ZhiChou
	case ganzhi.WxMu:
		return ganzhi.ZhiWei
	case ganzhi.WxShui:
		return ganzhi.ZhiChen
	case ganzhi.WxHuo:
		return ganzhi.ZhiXu
	case ganzhi.WxTu:
		return ganzhi.ZhiChen
	default:
		return 0
	}
}

func computeAtomicFacts(c FullChart) AtomicFacts {
	facts := AtomicFacts{
		DayMasterElement:      ganzhi.GanWuxing(c.Ri.Gan).String(),
		DayMasterStem:         ganzhi.GanName(c.Ri.Gan),
		DayBranch:             ganzhi.ZhiName(c.Ri.Zhi),
		MonthLongevity:        monthLongevity(c.Ri.Gan, c.Yue.Zhi),
		YearStemTenGod:        stemTenGod(c.Nian),
		MonthMainTenGod:       hiddenMainTenGod(c.Yue),
		HourStemTenGod:        stemTenGod(c.Shi),
		PatternGodTransparent: patternGodTransparent(c),
		PillarPunishments:     map[string]bool{},
		SpousePalaceState:     "静",
		DayBranchType:         dayBranchType(c.Ri.Zhi),
	}
	for _, pillar := range zhuLabels {
		facts.PillarPunishments[pillar] = false
	}
	pillars := [4]fullZhuInfo{c.Nian, c.Yue, c.Ri, c.Shi}

	dayElement := ganzhi.GanWuxing(c.Ri.Gan)
	wealthElement := ganzhi.Wuxing(0)
	for _, element := range []ganzhi.Wuxing{ganzhi.WxMu, ganzhi.WxHuo, ganzhi.WxTu, ganzhi.WxJin, ganzhi.WxShui} {
		if ganzhi.Ke(dayElement, element) {
			wealthElement = element
			break
		}
	}
	tomb := wealthTomb(wealthElement)
	if tomb != 0 {
		for _, pillar := range pillars {
			if pillar.Zhi != tomb {
				continue
			}
			facts.WealthTombPresent = true
			for _, item := range pillar.ShiShens {
				if item.ShiShen == ganzhi.ShiShenZhengCai || item.ShiShen == ganzhi.ShiShenPianCai {
					facts.WealthStarInTomb = true
					break
				}
			}
		}
	}

	hasOfficer := false
	hasKilling := false
	killingPillars := make(map[int]bool)
	for index, pillar := range pillars {
		for _, item := range pillar.ShiShens {
			if item.Source != sourceGan {
				continue
			}
			switch item.ShiShen {
			case ganzhi.ShiShenZhengGuan:
				hasOfficer = true
			case ganzhi.ShiShenQiSha:
				hasKilling = true
				killingPillars[index] = true
			}
		}
	}
	for _, relation := range append(append([]ZhiPairRel{}, c.ZhiLiuHe...), c.LiuChong...) {
		if killingPillars[relation.PillarA] || killingPillars[relation.PillarB] {
			facts.OfficerKillingCleaned = hasOfficer && hasKilling
			if facts.OfficerKillingCleaned {
				break
			}
		}
	}

	for _, item := range c.Nian.ShiShens {
		if item.ShiShen == ganzhi.ShiShenZhengGuan || item.ShiShen == ganzhi.ShiShenQiSha {
			facts.YearOfficerKilling = true
			break
		}
	}

	relationGroups := []struct {
		state string
		items []ZhiPairRel
	}{
		{"冲", c.LiuChong},
		{"合", c.ZhiLiuHe},
		{"刑", c.LiuXing},
		{"害", c.LiuHai},
	}
spousePriority:
	for _, group := range relationGroups {
		for _, relation := range group.items {
			if relation.PillarA == zhuRi || relation.PillarB == zhuRi {
				facts.SpousePalaceState = group.state
				break spousePriority
			}
		}
	}
	for _, relation := range c.LiuXing {
		for _, index := range []int{relation.PillarA, relation.PillarB} {
			if index >= 0 && index < len(zhuLabels) {
				facts.PillarPunishments[zhuLabels[index]] = true
			}
		}
	}
	return facts
}

func computeRelationGroups(c FullChart) []RelationGroup {
	bz := c.ToBazi()
	pillars := bz.Slice()
	ganCounts := map[ganzhi.Gan]int{}
	zhiCounts := map[ganzhi.Zhi]int{}
	for _, pillar := range pillars {
		ganCounts[pillar.Gan]++
		zhiCounts[pillar.Zhi]++
	}
	groups := make([]RelationGroup, 0, 8)
	seen := map[RelationGroup]bool{}
	add := func(field, group string) {
		groupModel := RelationGroup{Field: field, Group: group}
		if group != "" && !seen[groupModel] {
			groups = append(groups, RelationGroup{Field: field, Group: group})
			seen[groupModel] = true
		}
	}

	for _, relation := range c.GanHe {
		if group := ganHeGroupLabel(relation.GanA, relation.GanB); group != "" {
			field := "gan_he_candidate"
			if relation.Position == "adjacent" {
				field = "gan_he"
			}
			add(field, group)
		}
	}
	for _, relation := range ganzhi.ZhiHes {
		if zhiCounts[relation.A] > 0 && zhiCounts[relation.B] > 0 {
			add("zhi_liu_he", ganzhi.ZhiName(relation.A)+ganzhi.ZhiName(relation.B))
		}
	}
	for _, group := range ganzhi.TripleHeList {
		if allBranchesPresent(zhiCounts, group.Zhi) {
			add("san_he", branchGroupLabel(group.Zhi))
		}
	}
	for _, group := range ganzhi.TripleHuiList {
		if allBranchesPresent(zhiCounts, group.Zhi) {
			add("san_hui", branchGroupLabel(group.Zhi))
		}
	}
	for _, relation := range ganzhi.ChongPairs {
		if zhiCounts[relation.A] > 0 && zhiCounts[relation.B] > 0 {
			add("liu_chong", ganzhi.ZhiName(relation.A)+ganzhi.ZhiName(relation.B))
		}
	}
	for _, relation := range ganzhi.HaiPairs {
		if zhiCounts[relation.A] > 0 && zhiCounts[relation.B] > 0 {
			add("liu_hai", ganzhi.ZhiName(relation.A)+ganzhi.ZhiName(relation.B))
		}
	}
	for _, group := range ganzhi.XingGroups {
		if group.Type == "zi" {
			for _, branch := range group.Zhi {
				if zhiCounts[branch] >= 2 {
					add("liu_xing", ganzhi.ZhiName(branch)+ganzhi.ZhiName(branch))
				}
			}
			continue
		}
		if allBranchesPresent(zhiCounts, group.Zhi) {
			add("liu_xing", xingGroupLabel(group.Zhi))
		}
	}
	sort.Slice(groups, func(i, j int) bool {
		if groups[i].Field != groups[j].Field {
			return groups[i].Field < groups[j].Field
		}
		return groups[i].Group < groups[j].Group
	})
	return groups
}

// ganHeGroupLabel canonicalizes an observed adjacent stem pair to the table's
// stable group order. Non-adjacent stems never enter this function.
func ganHeGroupLabel(first, second string) string {
	a, errA := ganzhi.ParseGan(first)
	b, errB := ganzhi.ParseGan(second)
	if errA != nil || errB != nil {
		return ""
	}
	for _, relation := range ganzhi.GanHes {
		if (relation.A == a && relation.B == b) || (relation.A == b && relation.B == a) {
			return ganzhi.GanName(relation.A) + ganzhi.GanName(relation.B)
		}
	}
	return ""
}

func allBranchesPresent(counts map[ganzhi.Zhi]int, branches []ganzhi.Zhi) bool {
	for _, branch := range branches {
		if counts[branch] == 0 {
			return false
		}
	}
	return true
}

func branchGroupLabel(branches []ganzhi.Zhi) string {
	parts := make([]string, 0, len(branches))
	for _, branch := range branches {
		parts = append(parts, ganzhi.ZhiName(branch))
	}
	return strings.Join(parts, "")
}

func xingGroupLabel(branches []ganzhi.Zhi) string {
	present := map[ganzhi.Zhi]bool{}
	for _, branch := range branches {
		present[branch] = true
	}
	if present[ganzhi.ZhiChou] && present[ganzhi.ZhiXu] && present[ganzhi.ZhiWei] {
		return "丑戌未"
	}
	return branchGroupLabel(branches)
}

func monthLongevity(day ganzhi.Gan, month ganzhi.Zhi) string {
	for stage, branch := range ganzhi.ChangShengTable[day] {
		if branch == month {
			return ganzhi.StageNamesZH[stage]
		}
	}
	return ""
}

func stemTenGod(pillar fullZhuInfo) string {
	for _, item := range pillar.ShiShens {
		if item.Source == sourceGan {
			return item.ShiShen.String()
		}
	}
	return ""
}

func hiddenMainTenGod(pillar fullZhuInfo) string {
	for _, item := range pillar.ShiShens {
		if item.Source == sourceMainQi {
			return item.ShiShen.String()
		}
	}
	return ""
}

func patternGodTransparent(c FullChart) bool {
	target, ok := patternTenGods[c.YongShen.GeJu.Pattern]
	if !ok {
		return false
	}
	for _, pillar := range []fullZhuInfo{c.Nian, c.Yue, c.Ri, c.Shi} {
		for _, item := range pillar.ShiShens {
			if item.Source == sourceGan && item.ShiShen == target {
				return true
			}
		}
	}
	return false
}

var patternTenGods = map[string]ganzhi.ShiShen{
	"正官格": ganzhi.ShiShenZhengGuan,
	"七杀格": ganzhi.ShiShenQiSha,
	"正财格": ganzhi.ShiShenZhengCai,
	"偏财格": ganzhi.ShiShenPianCai,
	"正印格": ganzhi.ShiShenZhengYin,
	"偏印格": ganzhi.ShiShenPianYin,
	"食神格": ganzhi.ShiShenShiShen,
	"伤官格": ganzhi.ShiShenShangGuan,
}

func dayBranchType(zhi ganzhi.Zhi) string {
	switch zhi {
	case ganzhi.ZhiZi, ganzhi.ZhiMao, ganzhi.ZhiWu, ganzhi.ZhiYou:
		return "桃花"
	case ganzhi.ZhiYin, ganzhi.ZhiSi, ganzhi.ZhiShen, ganzhi.ZhiHai:
		return "驿马"
	case ganzhi.ZhiChou, ganzhi.ZhiChen, ganzhi.ZhiWei, ganzhi.ZhiXu:
		return "墓库"
	default:
		return ""
	}
}
