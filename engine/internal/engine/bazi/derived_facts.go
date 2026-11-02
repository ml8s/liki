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
	ID    string `json:"id"`
	Field string `json:"relation_kind"`
	Group string `json:"name"`
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
	Season             string   `json:"season"`
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

func computeTenGodStates(c Chart) ([]TenGodState, map[string]ElementState) {
	pillars := [4]Pillar{c.Nian, c.Yue, c.Ri, c.Shi}
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
		// 十神强度按 ten_god_strength_rules 表判定（《子平真诠》旺衰强弱：
		// 得令且透/根为强，失令且不透不根为弱；旺而不强判 neutral）。
		strength := "neutral"
		seasonStrong := element.SeasonStrength == "strong"
		for _, rule := range tenGodStrengthRules {
			if rule.Season != nil && *rule.Season != seasonStrong {
				continue
			}
			if rule.Transparent != nil && *rule.Transparent != accumulator.transparent {
				continue
			}
			if rule.Rooted != nil && *rule.Rooted != element.Rooted {
				continue
			}
			strength = rule.Kind
			break
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

func computeElementStates(c Chart, transparent, rooted map[ganzhi.Wuxing]bool) map[string]ElementState {
	all := []ganzhi.Wuxing{ganzhi.WxMu, ganzhi.WxHuo, ganzhi.WxTu, ganzhi.WxJin, ganzhi.WxShui}
	result := make(map[string]ElementState, len(all))
	for _, element := range all {
		state := ElementState{
			Wuxing:       element.String(),
			Season:       c.FuYi.WangShuai[element.String()],
			Transparent:  transparent[element],
			Rooted:       rooted[element],
			Controls:     elementControlledBy(element).String(),
			ControlledBy: elementThatControls(element).String(),
			Reasons:      []string{},
		}
		switch c.FuYi.WangShuai[element.String()] {
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

func computeLuRoots(c Chart) []LuRoot {
	roots := make([]LuRoot, 0, 4)
	pillars := [4]Pillar{c.Nian, c.Yue, c.Ri, c.Shi}
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

// ChartConclusions are the small set of natal conclusions used to fill the
// published GuanSha, CaiKu, and GeJu transparency fields.
type ChartConclusions struct {
	PatternGodTransparent bool
	OfficerKillingCleaned bool
	WealthTombPresent     bool
	WealthStarInTomb      bool
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

func computeChartConclusions(c Chart) ChartConclusions {
	var conclusions ChartConclusions
	pillars := [4]Pillar{c.Nian, c.Yue, c.Ri, c.Shi}

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
			conclusions.WealthTombPresent = true
			for _, item := range pillar.ShiShens {
				if item.ShiShen == ganzhi.ShiShenZhengCai || item.ShiShen == ganzhi.ShiShenPianCai {
					conclusions.WealthStarInTomb = true
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
			conclusions.OfficerKillingCleaned = hasOfficer && hasKilling
			if conclusions.OfficerKillingCleaned {
				break
			}
		}
	}

	conclusions.PatternGodTransparent = patternGodTransparent(c)
	return conclusions
}

func computeGuanXi(c Chart) []RelationGroup {
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
		groupModel := RelationGroup{ID: field + ":" + group, Field: field, Group: group}
		if group != "" && !seen[groupModel] {
			groups = append(groups, groupModel)
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
	for _, relation := range c.GanChong {
		if group := ganChongGroupLabel(relation.GanA, relation.GanB); group != "" {
			field := "gan_chong_candidate"
			if relation.Position == "adjacent" {
				field = "gan_chong"
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
			continue
		}
		if zhiCounts[group.Zhi[1]] > 0 {
			present := []ganzhi.Zhi{}
			for _, branch := range group.Zhi {
				if zhiCounts[branch] > 0 {
					present = append(present, branch)
				}
			}
			if len(present) == 2 {
				add("san_he_partial", branchGroupLabel(present))
			}
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
	for _, relation := range ganzhi.PoPairs {
		if zhiCounts[relation.A] > 0 && zhiCounts[relation.B] > 0 {
			add("liu_po", ganzhi.ZhiName(relation.A)+ganzhi.ZhiName(relation.B))
		}
	}
	for _, relation := range ganzhi.AnHePairs {
		if zhiCounts[relation.A] > 0 && zhiCounts[relation.B] > 0 {
			add("an_he", ganzhi.ZhiName(relation.A)+ganzhi.ZhiName(relation.B))
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

func ganChongGroupLabel(first, second string) string {
	a, errA := ganzhi.ParseGan(first)
	b, errB := ganzhi.ParseGan(second)
	if errA != nil || errB != nil {
		return ""
	}
	for _, relation := range ganzhi.GanChongPairs {
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

func patternGodTransparent(c Chart) bool {
	target, ok := patternTenGods[c.GeJu.Pattern]
	if !ok {
		return false
	}
	for _, pillar := range []Pillar{c.Nian, c.Yue, c.Ri, c.Shi} {
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
