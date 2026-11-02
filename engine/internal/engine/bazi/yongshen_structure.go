package bazi

import (
	"sort"
	"strings"

	"liki-engine/internal/engine/ganzhi"
)

// StemOccurrence records one exact stem occurrence in a palace. For a visible
// stem source is "gan"; for a hidden stem it is main_qi / mid_qi / minor_qi.
type StemOccurrence struct {
	Pillar string `json:"pillar"`
	Branch string `json:"branch"`
	Stem   string `json:"stem"`
	Source string `json:"source"`
}

// ElementStructure is an atomic element projection used by the FuYi, TiaoHou, and GeJu projections. It never asserts final pattern success or failure.
type ElementStructure struct {
	Wuxing             string           `json:"wuxing"`
	TransparentPillars []string         `json:"transparent_pillars"`
	Roots              []StemOccurrence `json:"roots"`
	Season             string           `json:"season"`
	Timely             bool             `json:"timely"`
	Strength           string           `json:"strength"`
}

// RelationLink points to one canonical guan_xi occurrence and records why the
// calling evidence projection selected it.
type RelationLink struct {
	ID      string   `json:"id"`
	Targets []string `json:"targets"`
}

// RelationFact is an internal targeted projection used while building links.
// Targets are supplied by the calling evidence projection (for example
// pattern/controller/generator).
type RelationFact struct {
	Field    string   `json:"field"`
	Group    string   `json:"group"`
	Pillars  []string `json:"pillars"`
	Branches []string `json:"branches"`
	Stems    []string `json:"-"`
	Targets  []string `json:"targets"`
}

// GeJuStructure records the month-command element, its controller and its
// generator. These are structural inputs only; classical success, failure and
// rescue conclusions remain outside this candidate model.
type GeJuStructure struct {
	Pattern     ElementStructure `json:"pattern"`
	Controller  ElementStructure `json:"controller"`
	Generator   ElementStructure `json:"generator"`
	GuanXiLinks []RelationLink   `json:"guan_xi_links"`
}

// FuYiBasis exposes the actual inputs of the qualitative strength lookup and
// the day-master-relevant relation projection.
type FuYiBasis struct {
	RiGanGen      string           `json:"ri_gan_gen"`
	YinBiShuLiang int              `json:"yin_bi_shu_liang"`
	RiGanTongGen  []StemOccurrence `json:"ri_gan_tong_gen"`
	GuanXiLinks   []RelationLink   `json:"guan_xi_links"`
}

// TiaoHouAvailability records whether an exact tabulated stem is visible,
// hidden, or absent, and which branch relations touch its containing palace.
type TiaoHouAvailability struct {
	Stem        string           `json:"stem"`
	Transparent bool             `json:"transparent"`
	Hidden      bool             `json:"hidden"`
	Occurrences []StemOccurrence `json:"occurrences"`
	GuanXiLinks []RelationLink   `json:"guan_xi_links"`
}

func buildGeJuStructure(
	bz ganzhi.Bazi, patternGan *ganzhi.Gan, patternElem ganzhi.Wuxing, relations []NatalRelationFact,
) GeJuStructure {
	controller := elementThatControls(patternElem)
	generator := elementThatGenerates(patternElem)
	return GeJuStructure{
		Pattern:    buildElementStructure(bz, patternElem, patternGan),
		Controller: buildElementStructure(bz, controller, nil),
		Generator:  buildElementStructure(bz, generator, nil),
		GuanXiLinks: relationLinks(buildRelationFacts(relations,
			func(pillar string, gan ganzhi.Gan) []string {
				return elementTargets(ganzhi.GanWuxing(gan), map[string]string{
					patternElem.String(): "pattern",
					controller.String():  "controller",
					generator.String():   "generator",
				})
			},
			func(pillar string) []string {
				return branchElementTargets(bz, pillarIndex(pillar), map[string]string{
					patternElem.String(): "pattern",
					controller.String():  "controller",
					generator.String():   "generator",
				})
			},
		)),
	}
}

func buildFuYiBasis(
	bz ganzhi.Bazi, rootType string, yinBi int, relations []NatalRelationFact,
) FuYiBasis {
	dayMaster := ganzhi.GanWuxing(bz.Ri.Gan)
	support := elementThatGenerates(dayMaster)
	return FuYiBasis{
		RiGanGen:      rootType,
		YinBiShuLiang: yinBi,
		RiGanTongGen:  hiddenStemOccurrences(bz, dayMaster),
		GuanXiLinks: relationLinks(buildRelationFacts(relations,
			func(pillar string, stem ganzhi.Gan) []string {
				targets := make([]string, 0, 2)
				if pillar == zhuLabels[zhuRi] {
					targets = append(targets, "day_master")
				}
				elem := ganzhi.GanWuxing(stem)
				if pillar != zhuLabels[zhuRi] && (elem == dayMaster || elem == support) {
					targets = append(targets, "yin_bi")
				}
				return targets
			},
			func(pillar string) []string {
				targets := make([]string, 0, 2)
				if pillar == zhuLabels[zhuRi] {
					targets = append(targets, "day_branch")
				}
				if len(branchElementTargets(bz, pillarIndex(pillar), map[string]string{dayMaster.String(): "day_master_root"})) > 0 {
					targets = append(targets, "day_master_root")
				}
				return targets
			},
		)),
	}
}

func buildTiaoHouAvailability(
	bz ganzhi.Bazi, stem ganzhi.Gan, target string, relations []NatalRelationFact,
) TiaoHouAvailability {
	availability := TiaoHouAvailability{
		Stem:        ganzhi.GanName(stem),
		Occurrences: make([]StemOccurrence, 0),
		GuanXiLinks: make([]RelationLink, 0),
	}
	pillars := [4]struct {
		index int
		gan   ganzhi.Gan
		zhi   ganzhi.Zhi
	}{
		{zhuNian, bz.Nian.Gan, bz.Nian.Zhi},
		{zhuYue, bz.Yue.Gan, bz.Yue.Zhi},
		{zhuRi, bz.Ri.Gan, bz.Ri.Zhi},
		{zhuShi, bz.Shi.Gan, bz.Shi.Zhi},
	}
	cangGan := HiddenStemFacts(bz)
	targetPillars := map[int]bool{}
	for _, pillar := range pillars {
		if pillar.gan == stem {
			availability.Transparent = true
			targetPillars[pillar.index] = true
			availability.Occurrences = append(availability.Occurrences, StemOccurrence{
				Pillar: zhuLabels[pillar.index], Branch: ganzhi.ZhiName(pillar.zhi),
				Stem: ganzhi.GanName(stem), Source: sourceGan,
			})
		}
		hidden := []struct {
			gan    ganzhi.Gan
			source string
		}{
			{cangGan[pillar.index].Main, sourceMainQi},
		}
		if cangGan[pillar.index].Mid != nil {
			hidden = append(hidden, struct {
				gan    ganzhi.Gan
				source string
			}{*cangGan[pillar.index].Mid, sourceMidQi})
		}
		if cangGan[pillar.index].Minor != nil {
			hidden = append(hidden, struct {
				gan    ganzhi.Gan
				source string
			}{*cangGan[pillar.index].Minor, sourceMinQi})
		}
		for _, item := range hidden {
			if item.gan == stem {
				availability.Hidden = true
				targetPillars[pillar.index] = true
				availability.Occurrences = append(availability.Occurrences, StemOccurrence{
					Pillar: zhuLabels[pillar.index], Branch: ganzhi.ZhiName(pillar.zhi),
					Stem: ganzhi.GanName(stem), Source: item.source,
				})
			}
		}
	}
	availability.GuanXiLinks = relationLinks(buildRelationFacts(relations,
		func(pillar string, relationStem ganzhi.Gan) []string {
			if targetPillars[pillarIndex(pillar)] && relationStem == stem {
				return []string{target}
			}
			return nil
		},
		func(pillar string) []string {
			if targetPillars[pillarIndex(pillar)] {
				return []string{target}
			}
			return nil
		},
	))
	return availability
}

func buildElementStructure(bz ganzhi.Bazi, elem ganzhi.Wuxing, exact *ganzhi.Gan) ElementStructure {
	structure := ElementStructure{
		Wuxing:             elem.String(),
		TransparentPillars: transparentPillars(bz, elem, exact),
		Roots:              hiddenStemOccurrences(bz, elem),
	}
	structure.Season = ganzhi.WangShuaiOf(elem, bz.Yue.Zhi).String()
	structure.Timely = structure.Season == ganzhi.WSWang.String() || structure.Season == ganzhi.WSXiang.String()
	switch {
	case structure.Timely:
		structure.Strength = "strong"
	case len(structure.TransparentPillars) > 0 && len(structure.Roots) > 0:
		structure.Strength = "strong"
	case isDeadOrTrappedSeason(structure.Season) && len(structure.TransparentPillars) == 0 && len(structure.Roots) == 0:
		structure.Strength = "weak"
	default:
		structure.Strength = "neutral"
	}
	return structure
}

func transparentPillars(bz ganzhi.Bazi, elem ganzhi.Wuxing, exact *ganzhi.Gan) []string {
	result := make([]string, 0, 4)
	for index, gan := range []ganzhi.Gan{bz.Nian.Gan, bz.Yue.Gan, bz.Ri.Gan, bz.Shi.Gan} {
		if exact != nil {
			if gan == *exact {
				result = append(result, zhuLabels[index])
			}
			continue
		}
		if ganzhi.GanWuxing(gan) == elem {
			result = append(result, zhuLabels[index])
		}
	}
	return result
}

func hiddenStemOccurrences(bz ganzhi.Bazi, elem ganzhi.Wuxing) []StemOccurrence {
	result := make([]StemOccurrence, 0, 8)
	cangGan := HiddenStemFacts(bz)
	branches := []ganzhi.Zhi{bz.Nian.Zhi, bz.Yue.Zhi, bz.Ri.Zhi, bz.Shi.Zhi}
	for index, hidden := range cangGan {
		items := []struct {
			gan    ganzhi.Gan
			source string
		}{{hidden.Main, sourceMainQi}}
		if hidden.Mid != nil {
			items = append(items, struct {
				gan    ganzhi.Gan
				source string
			}{*hidden.Mid, sourceMidQi})
		}
		if hidden.Minor != nil {
			items = append(items, struct {
				gan    ganzhi.Gan
				source string
			}{*hidden.Minor, sourceMinQi})
		}
		for _, item := range items {
			if ganzhi.GanWuxing(item.gan) == elem {
				result = append(result, StemOccurrence{
					Pillar: zhuLabels[index], Branch: ganzhi.ZhiName(branches[index]),
					Stem: ganzhi.GanName(item.gan), Source: item.source,
				})
			}
		}
	}
	return result
}

func buildRelationFacts(
	relations []NatalRelationFact,
	stemTargets func(pillar string, stem ganzhi.Gan) []string,
	branchTargets func(pillar string) []string,
) []RelationFact {
	branchMemo := map[string][]string{}
	memoBranchTargets := func(pillar string) []string {
		if branchTargets == nil {
			return nil
		}
		if value, ok := branchMemo[pillar]; ok {
			return value
		}
		value := branchTargets(pillar)
		branchMemo[pillar] = value
		return value
	}

	facts := make([]RelationFact, 0, len(relations))
	for _, relation := range relations {
		targets := make([]string, 0, 4)
		stems := make([]string, 0, len(relation.Members))
		branches := make([]string, 0, len(relation.Members))
		pillars := make([]string, 0, len(relation.Members))
		for _, member := range relation.Members {
			pillars = append(pillars, member.Pillar)
			if member.Gan != "" {
				stem, err := ganzhi.ParseGan(member.Gan)
				if err != nil {
					continue
				}
				stems = append(stems, member.Gan)
				targets = appendUniqueStrings(targets, stemTargets(member.Pillar, stem)...)
			}
			if member.Zhi != "" {
				if _, err := ganzhi.ParseZhi(member.Zhi); err != nil {
					continue
				}
				branches = append(branches, member.Zhi)
				targets = appendUniqueStrings(targets, memoBranchTargets(member.Pillar)...)
			}
		}
		if len(targets) == 0 || len(stems)+len(branches) != len(relation.Members) {
			continue
		}
		if len(stems) > 0 && len(branches) > 0 {
			continue
		}
		facts = append(facts, RelationFact{
			Field: relation.Kind, Group: relation.Group, Pillars: pillars,
			Branches: branches, Stems: stems, Targets: targets,
		})
	}
	return dedupeRelationFacts(facts)
}

func elementTargets(elem ganzhi.Wuxing, labels map[string]string) []string {
	if label, ok := labels[elem.String()]; ok {
		return []string{label}
	}
	return nil
}

func branchElementTargets(bz ganzhi.Bazi, pillar int, labels map[string]string) []string {
	var branch ganzhi.Zhi
	switch pillar {
	case zhuNian:
		branch = bz.Nian.Zhi
	case zhuYue:
		branch = bz.Yue.Zhi
	case zhuRi:
		branch = bz.Ri.Zhi
	case zhuShi:
		branch = bz.Shi.Zhi
	}
	if branch == 0 {
		return nil
	}
	hidden := HiddenStemFacts(bz)[pillar]
	stems := []ganzhi.Gan{hidden.Main}
	if hidden.Mid != nil {
		stems = append(stems, *hidden.Mid)
	}
	if hidden.Minor != nil {
		stems = append(stems, *hidden.Minor)
	}
	targets := make([]string, 0, len(labels))
	for _, stem := range stems {
		if label, ok := labels[ganzhi.GanWuxing(stem).String()]; ok {
			targets = appendUniqueStrings(targets, label)
		}
	}
	return targets
}

func pillarNames(indices ...int) []string {
	names := make([]string, 0, len(indices))
	for _, index := range indices {
		names = append(names, zhuLabels[index])
	}
	return names
}

func appendUniqueStrings(values []string, additions ...string) []string {
	for _, addition := range additions {
		found := false
		for _, value := range values {
			if value == addition {
				found = true
				break
			}
		}
		if !found {
			values = append(values, addition)
		}
	}
	return values
}

func dedupeRelationFacts(values []RelationFact) []RelationFact {
	result := make([]RelationFact, 0, len(values))
	seen := map[string]bool{}
	for _, value := range values {
		key := strings.Join([]string{value.Field, value.Group, strings.Join(value.Pillars, ","), strings.Join(value.Branches, ","), strings.Join(value.Targets, ",")}, "\x00")
		if seen[key] {
			continue
		}
		seen[key] = true
		result = append(result, value)
	}
	return result
}

// relationLinks converts school-targeted relation facts into links to the
// chart's canonical guan_xi occurrences while retaining projection targets.
func relationLinks(values []RelationFact) []RelationLink {
	links := map[string]map[string]bool{}
	for _, value := range values {
		fact, ok := canonicalRelationReference(value)
		if !ok {
			continue
		}
		for _, target := range value.Targets {
			if links[fact.ID] == nil {
				links[fact.ID] = map[string]bool{}
			}
			links[fact.ID][target] = true
		}
	}
	result := make([]RelationLink, 0, len(links))
	for id, targetSet := range links {
		targets := make([]string, 0, len(targetSet))
		for target := range targetSet {
			targets = append(targets, target)
		}
		sort.Strings(targets)
		result = append(result, RelationLink{ID: id, Targets: targets})
	}
	sort.Slice(result, func(i, j int) bool { return result[i].ID < result[j].ID })
	return result
}

func canonicalRelationReference(value RelationFact) (NatalRelationFact, bool) {
	if value.Group == "" || len(value.Pillars) == 0 {
		return NatalRelationFact{}, false
	}
	fact := NatalRelationFact{Kind: value.Field}
	switch value.Field {
	case "gan_he", "gan_he_candidate", "gan_chong", "gan_chong_candidate":
		if len(value.Stems) != 2 || len(value.Pillars) != 2 {
			return NatalRelationFact{}, false
		}
		if value.Field == "gan_he" || value.Field == "gan_he_candidate" {
			fact.Group = ganHeGroupLabel(value.Stems[0], value.Stems[1])
			if pillarDistance(value.Pillars[0], value.Pillars[1]) != 1 {
				fact.Kind = "gan_he_candidate"
			}
		} else {
			fact.Group = ganChongGroupLabel(value.Stems[0], value.Stems[1])
		}
		fact.Members = []NatalRelationMember{
			{Pillar: value.Pillars[0], Gan: value.Stems[0]},
			{Pillar: value.Pillars[1], Gan: value.Stems[1]},
		}
	case "zhi_liu_he", "liu_chong", "liu_hai", "liu_po", "an_he":
		if len(value.Branches) != 2 || len(value.Pillars) != 2 {
			return NatalRelationFact{}, false
		}
		fact.Group = canonicalZhiPairGroup(value.Field, value.Branches[0], value.Branches[1])
		fact.Members = zhiMembers(value.Pillars, value.Branches)
	case "liu_xing":
		if len(value.Branches) != 2 || len(value.Pillars) != 2 {
			return NatalRelationFact{}, false
		}
		fact.Group = canonicalXingGroup(value.Branches[0], value.Branches[1])
		fact.Members = zhiMembers(value.Pillars, value.Branches)
	case "san_he", "san_he_partial", "san_hui":
		fact.Group = branchGroupLabel(parseZhiNames(value.Branches))
		fact.Members = zhiMembers(value.Pillars, value.Branches)
	default:
		return NatalRelationFact{}, false
	}
	if fact.Group == "" || len(fact.Members) == 0 {
		return NatalRelationFact{}, false
	}
	fact.ID = natalRelationFactID(fact)
	return fact, true
}

func zhiMembers(pillars, branches []string) []NatalRelationMember {
	if len(pillars) != len(branches) {
		return nil
	}
	members := make([]NatalRelationMember, 0, len(branches))
	for index, branch := range branches {
		members = append(members, NatalRelationMember{Pillar: pillars[index], Zhi: branch})
	}
	return members
}

func pillarDistance(first, second string) int {
	order := map[string]int{"nian": 0, "yue": 1, "ri": 2, "shi": 3}
	a, okA := order[first]
	b, okB := order[second]
	if !okA || !okB {
		return 0
	}
	distance := a - b
	if distance < 0 {
		distance = -distance
	}
	return distance
}

func isDeadOrTrappedSeason(season string) bool {
	return season == ganzhi.WSQiu.String() || season == ganzhi.WSSi.String()
}
