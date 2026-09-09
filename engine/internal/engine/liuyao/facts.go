package liuyao

import (
	"fmt"
	"strings"

	"liki-engine/internal/engine/ganzhi"
)

// DayClashFact distinguishes the effect of a day clash. A clash is not always
// "bad": a strong static line may become hidden-moving, while a weak one may
// be day-broken and a moving line may be scattered.
type DayClashFact struct {
	Position  int      `json:"position"`
	Kind      string   `json:"kind"`
	Basis     []string `json:"basis,omitempty"`
	Condition string   `json:"condition,omitempty"`
}

// MovingTransformation records all mechanical consequences of one moving line
// changing into its corresponding changed line.
type MovingTransformation struct {
	Position       int      `json:"position"`
	FromBranch     string   `json:"from_branch"`
	ToBranch       string   `json:"to_branch"`
	ReturnRelation string   `json:"return_relation,omitempty"`
	AdvanceRetreat string   `json:"advance_retreat,omitempty"`
	ChangedVoid    bool     `json:"changed_void,omitempty"`
	ChangedBreak   bool     `json:"changed_break,omitempty"`
	Basis          []string `json:"basis,omitempty"`
}

// SanHeCandidate records a possible three-harmony group. It does not assert
// that the group is effective; callers must consider moving lines, void/break
// states, month/day support and the concrete target.
type SanHeCandidate struct {
	Positions       []int    `json:"positions"`
	Branches        []string `json:"branches"`
	Element         string   `json:"element"`
	Complete        bool     `json:"complete"`
	MovingCount     int      `json:"moving_count"`
	Qualification   string   `json:"qualification"`
	LineComplete    bool     `json:"line_complete"`
	DayPresent      bool     `json:"day_present"`
	MonthPresent    bool     `json:"month_present"`
	Missing         []string `json:"missing,omitempty"`
	StaticCount     int      `json:"static_count"`
	VoidCount       int      `json:"void_count"`
	BreakCount      int      `json:"break_count"`
	Activated       bool     `json:"activated"`
	Targets         []string `json:"targets,omitempty"`
	ConclusionScope string   `json:"conclusion_scope"`
}

// HiddenLine is one palace-base relative not visible in the original hexagram.
type HiddenLine struct {
	Position       int    `json:"position"`
	LiuQin         string `json:"liu_qin"`
	Gan            string `json:"gan"`
	Zhi            string `json:"zhi"`
	Wuxing         string `json:"wuxing"`
	FlyingPosition int    `json:"flying_position"`
	FlyingLiuQin   string `json:"flying_liu_qin"`
	FlyingGan      string `json:"flying_gan"`
	FlyingZhi      string `json:"flying_zhi"`
	FlyingWuxing   string `json:"flying_wuxing"`
	FeiFuRelation  string `json:"fei_fu_relation"`
	XunKong        bool   `json:"xun_kong,omitempty"`
	YuePo          bool   `json:"yue_po,omitempty"`
	MuKu           bool   `json:"mu_ku,omitempty"`
	ExitCondition  string `json:"exit_condition,omitempty"`
}

// BranchRelationFact records a direct branch relation between two visible lines.
type BranchRelationFact struct {
	LeftPosition  int      `json:"left_position"`
	RightPosition int      `json:"right_position"`
	Relations     []string `json:"relations"`
	Targets       []string `json:"targets,omitempty"`
	Basis         []string `json:"basis,omitempty"`
}

type ForceRole string

const (
	RoleYongShen ForceRole = "用神"
	RoleYuanShen ForceRole = "原神"
	RoleJiShen   ForceRole = "忌神"
	RoleChouShen ForceRole = "仇神"
)

// ForceEntry describes one role in the support/opposition chain.
type ForceEntry struct {
	Role      ForceRole `json:"role"`
	Element   string    `json:"element"`
	Positions []int     `json:"positions,omitempty"`
	Branches  []string  `json:"branches,omitempty"`
	Relations []string  `json:"relations,omitempty"`
	States    []string  `json:"states,omitempty"`
	Note      string    `json:"note,omitempty"`
}

// ForceChain is the structured support/opposition map around the YongShen.
type ForceChain struct {
	YongShen    string       `json:"yong_shen"`
	YongElement string       `json:"yong_element"`
	Position    int          `json:"position"`
	IsHidden    bool         `json:"is_hidden"`
	Entries     []ForceEntry `json:"entries"`
}

// YongShenCandidate makes every visible/hidden candidate auditable. Selected
// is the position currently used by deterministic engine functions.
type YongShenCandidate struct {
	Position  int      `json:"position"`
	Branch    string   `json:"branch"`
	IsHidden  bool     `json:"is_hidden"`
	Selected  bool     `json:"selected"`
	WangShuai string   `json:"wang_shuai,omitempty"`
	Reason    string   `json:"reason"`
	Basis     []string `json:"basis,omitempty"`
}

func forceEntry(chart *Chart, role ForceRole, element ganzhi.Wuxing, note string) ForceEntry {
	entry := ForceEntry{Role: role, Element: element.String(), Note: note}
	for _, line := range chart.Lines {
		if line.Wuxing != element {
			continue
		}
		entry.Positions = append(entry.Positions, line.Position)
		entry.Branches = append(entry.Branches, ganzhi.ZhiName(line.Zhi))
		relation := dayInteraction(line.Zhi, chart.RiZhi)
		entry.Relations = append(entry.Relations, fmt.Sprintf("日%s", relation.Relation))
		state := []string{chart.WangShuai[line.Position-1].String()}
		if line.DongSelf {
			state = append(state, "发动")
		} else {
			state = append(state, "安静")
		}
		if line.XunKong {
			state = append(state, "旬空")
		}
		if line.YuePo {
			state = append(state, "月破")
		}
		if line.MuKu {
			state = append(state, "入墓")
		}
		entry.States = append(entry.States, strings.Join(state, "/"))
	}
	return entry
}

func computeForceChain(chart *Chart, typ YongShen) *ForceChain {
	position, isBian := chart.findYongShen(typ)
	if position == 0 {
		return nil
	}
	// The force chain is evaluated on the original visible line layer. Changed
	// layers are handled by MovingTransformation; mixing them here would make
	// role positions ambiguous.
	_ = isBian
	yong := chart.Lines[position-1]
	yongElement := yong.Wuxing
	var yuan, ji, chou ganzhi.Wuxing
	for _, element := range []ganzhi.Wuxing{ganzhi.WxMu, ganzhi.WxHuo, ganzhi.WxTu, ganzhi.WxJin, ganzhi.WxShui} {
		if ganzhi.Sheng(element, yongElement) {
			yuan = element
		}
		if ganzhi.Ke(element, yongElement) {
			ji = element
		}
	}
	for _, element := range []ganzhi.Wuxing{ganzhi.WxMu, ganzhi.WxHuo, ganzhi.WxTu, ganzhi.WxJin, ganzhi.WxShui} {
		if yuan != 0 && ganzhi.Ke(element, yuan) {
			chou = element
		}
	}

	chain := &ForceChain{
		YongShen:    typ.String(),
		YongElement: yongElement.String(),
		Position:    position,
		IsHidden:    false,
	}
	chain.Entries = append(chain.Entries,
		forceEntry(chart, RoleYongShen, yongElement, "所问事项主线"),
		forceEntry(chart, RoleYuanShen, yuan, "生用神者"),
		forceEntry(chart, RoleJiShen, ji, "克用神者"),
	)
	if chou != 0 {
		chain.Entries = append(chain.Entries, forceEntry(chart, RoleChouShen, chou, "克原神者"))
	}
	return chain
}

func computeDayClashFacts(chart *Chart) []DayClashFact {
	var facts []DayClashFact
	for i := range chart.Lines {
		line := chart.Lines[i]
		if !ganzhi.IsLiuChong(line.Zhi, chart.RiZhi) {
			continue
		}
		ws := chart.WangShuai[i]
		strong := ws == ganzhi.WSWang || ws == ganzhi.WSXiang
		kind := ""
		condition := ""
		switch {
		case line.XunKong:
			kind = "冲空"
			condition = "动空可为冲实候选；休囚静空多作冲脱"
		case line.DongSelf:
			kind = "冲散"
			condition = "动爻被日冲，须查有无合救与其他生扶"
		case strong:
			kind = "暗动"
			condition = "旺相静爻逢冲，可作暗中发动候选"
		default:
			kind = "日破"
			condition = "休囚静爻逢冲，当下失用；出旬出月或得救应另论"
		}
		basis := []string{
			fmt.Sprintf("line.zhi=%s", ganzhi.ZhiName(line.Zhi)),
			fmt.Sprintf("ri_zhi=%s", ganzhi.ZhiName(chart.RiZhi)),
			fmt.Sprintf("wang_shuai=%s", ws.String()),
			fmt.Sprintf("dong_self=%t", line.DongSelf),
			fmt.Sprintf("xun_kong=%t", line.XunKong),
		}
		facts = append(facts, DayClashFact{
			Position:  line.Position,
			Kind:      kind,
			Basis:     basis,
			Condition: condition,
		})
	}
	return facts
}

func computeMovingTransformations(chart *Chart) []MovingTransformation {
	var facts []MovingTransformation
	for _, position := range chart.DongYao {
		if position < 1 || position > 6 {
			continue
		}
		from := chart.Lines[position-1]
		to := chart.BianLines[position-1]
		item := MovingTransformation{
			Position:   position,
			FromBranch: ganzhi.ZhiName(from.Zhi),
			ToBranch:   ganzhi.ZhiName(to.Zhi),
			Basis: []string{
				fmt.Sprintf("from_element=%s", from.Wuxing.String()),
				fmt.Sprintf("to_element=%s", to.Wuxing.String()),
			},
		}
		switch {
		case ganzhi.Sheng(to.Wuxing, from.Wuxing):
			item.ReturnRelation = "回头生"
		case ganzhi.Ke(to.Wuxing, from.Wuxing):
			item.ReturnRelation = "回头克"
		case to.Wuxing == from.Wuxing:
			item.ReturnRelation = "比和"
		}
		if next, ok := jinShenTable[from.Zhi]; ok && next == to.Zhi {
			item.AdvanceRetreat = "化进"
		}
		if prev, ok := tuiShenTable[from.Zhi]; ok && prev == to.Zhi {
			item.AdvanceRetreat = "化退"
		}
		item.ChangedVoid = to.XunKong
		item.ChangedBreak = to.YuePo
		if item.ChangedVoid {
			item.Basis = append(item.Basis, "changed.xun_kong=true")
		}
		if item.ChangedBreak {
			item.Basis = append(item.Basis, "changed.yue_po=true")
		}
		facts = append(facts, item)
	}
	return facts
}

func computeSanHeCandidates(chart *Chart) []SanHeCandidate {
	var facts []SanHeCandidate
	for _, group := range ganzhi.TripleHeList {
		linePositions := make([]int, 0, len(group.Zhi))
		lineBranches := make([]ganzhi.Zhi, 0, len(group.Zhi))
		seen := map[ganzhi.Zhi]bool{}
		var moving, static, voidCount, breakCount int
		for _, line := range chart.Lines {
			if !containsZhi(group.Zhi, line.Zhi) {
				continue
			}
			if !seen[line.Zhi] {
				seen[line.Zhi] = true
				linePositions = append(linePositions, line.Position)
				lineBranches = append(lineBranches, line.Zhi)
			}
			if line.DongSelf {
				moving++
			} else {
				static++
			}
			if line.XunKong {
				voidCount++
			}
			if line.YuePo {
				breakCount++
			}
		}
		if len(linePositions) == 0 {
			continue
		}

		dayPresent := containsZhi(group.Zhi, chart.RiZhi)
		monthPresent := containsZhi(group.Zhi, chart.YueZhi)
		var missing []string
		for _, zhi := range group.Zhi {
			if !containsZhi(lineBranches, zhi) {
				missing = append(missing, ganzhi.ZhiName(zhi))
			}
		}
		lineComplete := len(linePositions) == len(group.Zhi)
		complete := lineComplete || (len(linePositions)+boolToInt(dayPresent)+boolToInt(monthPresent) >= len(group.Zhi))
		targets := make([]string, 0, 3)
		yongPos := chart.YongShen.Position
		if yongPos > 0 {
			for _, pos := range linePositions {
				if pos == yongPos {
					targets = append(targets, "yong_shen")
					break
				}
			}
		}
		for _, line := range chart.Lines {
			for _, pos := range linePositions {
				if line.Position != pos {
					continue
				}
				switch line.ShiYing {
				case "世":
					targets = append(targets, "world")
				case "应":
					targets = append(targets, "other")
				}
			}
		}

		qualification := "卦爻支数不足三，只作半合 / 聚气候选，不作成局结论"
		activated := false
		switch {
		case lineComplete && moving > 0 && voidCount == 0 && breakCount == 0:
			activated = true
			qualification = "卦中三支齐且有发动，未见空破；可作成局候选，仍须核月日支持"
		case lineComplete:
			qualification = "卦中三支齐但无发动或见空破，作潜在聚气候选，不直接论成局"
		case complete:
			qualification = "卦爻与日 / 月合补三支，只作聚气候选；不以静支直接论成局"
		}

		facts = append(facts, SanHeCandidate{
			Positions:       linePositions,
			Branches:        namedBranches(lineBranches),
			Element:         group.Element.String(),
			Complete:        complete,
			MovingCount:     moving,
			Qualification:   qualification,
			LineComplete:    lineComplete,
			DayPresent:      dayPresent,
			MonthPresent:    monthPresent,
			Missing:         missing,
			StaticCount:     static,
			VoidCount:       voidCount,
			BreakCount:      breakCount,
			Activated:       activated,
			Targets:         targets,
			ConclusionScope: "structure_candidate_only_not_outcome",
		})
	}
	return facts
}

func boolToInt(value bool) int {
	if value {
		return 1
	}
	return 0
}

func namedBranches(branches []ganzhi.Zhi) []string {
	out := make([]string, 0, len(branches))
	for _, branch := range branches {
		out = append(out, ganzhi.ZhiName(branch))
	}
	return out
}

func containsZhi(list []ganzhi.Zhi, target ganzhi.Zhi) bool {
	for _, zhi := range list {
		if zhi == target {
			return true
		}
	}
	return false
}

func computeYongShenCandidates(chart *Chart, typ YongShen) []YongShenCandidate {
	if typ == YongShiYao {
		position := chart.findShiYao()
		if position == 0 {
			return nil
		}
		line := chart.Lines[position-1]
		return []YongShenCandidate{{
			Position:  position,
			Branch:    ganzhi.ZhiName(line.Zhi),
			Selected:  true,
			WangShuai: chart.WangShuai[position-1].String(),
			Reason:    "问自身，取世爻",
			Basis:     []string{"yong_shen=世爻"},
		}}
	}
	target := yongShenToLiuQin(typ)
	if target < 0 {
		return nil
	}
	var candidates []YongShenCandidate
	for i := range chart.Lines {
		line := chart.Lines[i]
		if line.LiuQin != target {
			continue
		}
		candidate := YongShenCandidate{
			Position:  line.Position,
			Branch:    ganzhi.ZhiName(line.Zhi),
			WangShuai: chart.WangShuai[i].String(),
		}
		if line.DongSelf {
			candidate.Basis = append(candidate.Basis, "dong_self=true")
		}
		if line.XunKong {
			candidate.Basis = append(candidate.Basis, "xun_kong=true")
		}
		if line.YuePo {
			candidate.Basis = append(candidate.Basis, "yue_po=true")
		}
		candidates = append(candidates, candidate)
	}
	selected, _ := chart.findYongShen(typ)
	for i := range candidates {
		if candidates[i].Position == selected {
			candidates[i].Selected = true
			if len(candidates) > 1 {
				candidates[i].Reason = "用神两现，按旺衰与空破取舍"
			} else {
				candidates[i].Reason = "卦中唯现"
			}
		} else if len(candidates) > 1 {
			candidates[i].Reason = "未选；列出以供复核"
		}
	}
	if selected == 0 {
		if hidden := chart.findFuShen(typ); hidden != nil {
			candidates = append(candidates, YongShenCandidate{
				Position: hidden.Position,
				Branch:   hidden.Zhi,
				IsHidden: true,
				Selected: true,
				Reason:   "本卦不现，取本宫伏神",
			})
		}
	}
	return candidates
}
