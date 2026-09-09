package liuyao

import (
	"fmt"

	"liki-engine/internal/engine/ganzhi"
)

// TimingCandidate is one auditable trigger hypothesis. It is never a promise
// that an event must happen on a particular date.
type TimingCandidate struct {
	ID            string   `json:"id"`
	Mechanism     string   `json:"mechanism"`
	Position      int      `json:"position,omitempty"`
	TriggerBranch string   `json:"trigger_branch,omitempty"`
	Window        string   `json:"window,omitempty"`
	Basis         []string `json:"basis,omitempty"`
	Condition     string   `json:"condition,omitempty"`
	Confidence    string   `json:"confidence"`
}

func zhiHePartner(z ganzhi.Zhi) ganzhi.Zhi {
	for candidate := ganzhi.Zhi(1); candidate <= 12; candidate++ {
		if candidate != z && ganzhi.IsZhiHe(z, candidate) {
			return candidate
		}
	}
	return 0
}

func candidateID(prefix string, position int, branch ganzhi.Zhi) string {
	return fmt.Sprintf("%s-%d-%s", prefix, position, ganzhi.ZhiName(branch))
}

func appendCandidate(candidates []TimingCandidate, candidate TimingCandidate) []TimingCandidate {
	for _, existing := range candidates {
		if existing.ID == candidate.ID && existing.Mechanism == candidate.Mechanism {
			return candidates
		}
	}
	candidate.Confidence = "candidate"
	return append(candidates, candidate)
}

func appendBranchTrigger(
	candidates []TimingCandidate, idPrefix, mechanism string, position int,
	branch ganzhi.Zhi, basis []string, condition string,
) []TimingCandidate {
	if branch < 1 || branch > 12 {
		return candidates
	}
	return appendCandidate(candidates, TimingCandidate{
		ID:            candidateID(idPrefix, position, branch),
		Mechanism:     mechanism,
		Position:      position,
		TriggerBranch: ganzhi.ZhiName(branch),
		Window:        "day_or_month",
		Basis:         basis,
		Condition:     condition,
	})
}

// computeTimingCandidates derives trigger hypotheses from the already computed
// YongShen state. It intentionally does not assign a probability or a final
// date; the caller must judge which candidates apply to the concrete question.
func computeTimingCandidates(p *Chart, typ YongShen) []TimingCandidate {
	var candidates []TimingCandidate
	yongPos, isBian := p.findYongShen(typ)
	if yongPos == 0 {
		if fs := p.findFuShen(typ); fs != nil && fs.Position >= 1 && fs.Position <= 6 {
			flying := p.Lines[fs.Position-1]
			candidates = appendBranchTrigger(
				candidates,
				"hidden-release",
				"冲飞出伏",
				fs.Position,
				chongZhi(flying.Zhi),
				[]string{"yong_shen.is_hidden=true", fmt.Sprintf("flying.position=%d", flying.Position)},
				"飞神受冲且伏神有气时，方可作为出伏候选",
			)
		}
		return candidates
	}

	var yao Line
	layer := "ben"
	if isBian {
		yao = p.BianLines[yongPos-1]
		layer = "bian"
	} else {
		yao = p.Lines[yongPos-1]
	}
	yong := p.YongShen
	moving := yao.Type.IsChanging()

	if moving {
		candidates = appendBranchTrigger(
			candidates,
			"moving-value",
			"动爻逢值",
			yongPos,
			yao.Zhi,
			[]string{fmt.Sprintf("%s_yao.type=%d", layer, int(yao.Type)), "yong_shen.position=" + fmt.Sprint(yongPos)},
			"动爻逢值仅为候选，须先确认动变未失用",
		)
		if partner := zhiHePartner(yao.Zhi); partner != 0 {
			candidates = appendBranchTrigger(
				candidates,
				"moving-harmony",
				"动爻逢合",
				yongPos,
				partner,
				[]string{fmt.Sprintf("%s_yao.type=%d", layer, int(yao.Type)), "branch_relation=六合"},
				"旺相之动可作应期候选；休囚被合须先辨合绊",
			)
		}
	}

	if yong.XunKong {
		candidates = appendBranchTrigger(
			candidates,
			"void-fill",
			"旬空填实",
			yongPos,
			yao.Zhi,
			[]string{"yong_shen.xun_kong=true"},
			"真假空须先判定；真空不作成事候选",
		)
		candidates = appendBranchTrigger(
			candidates,
			"void-clash",
			"冲空",
			yongPos,
			chongZhi(yao.Zhi),
			[]string{"yong_shen.xun_kong=true"},
			"旺相动空可冲实；休囚静空则多为冲脱",
		)
	}

	if yong.YuePo {
		candidates = appendBranchTrigger(
			candidates,
			"month-break-value",
			"月破逢值",
			yongPos,
			yao.Zhi,
			[]string{"yong_shen.yue_po=true"},
			"真破不作成事候选；假破可作出月后候选",
		)
		if partner := zhiHePartner(yao.Zhi); partner != 0 {
			candidates = appendBranchTrigger(
				candidates,
				"month-break-harmony",
				"月破逢合",
				yongPos,
				partner,
				[]string{"yong_shen.yue_po=true"},
				"假破有合可为补救候选；真破不单凭逢合论成",
			)
		}
		candidates = appendBranchTrigger(
			candidates,
			"month-break-exit",
			"出月令",
			yongPos,
			chongZhi(p.YueZhi),
			[]string{"yong_shen.yue_po=true", "month_branch=" + ganzhi.ZhiName(p.YueZhi)},
			"以离开当前月令作为解除月破的环境候选",
		)
	}

	if !moving && !isBian && ganzhi.IsLiuChong(yao.Zhi, p.RiZhi) {
		candidates = appendBranchTrigger(
			candidates,
			"static-clash",
			"静爻逢冲",
			yongPos,
			chongZhi(yao.Zhi),
			[]string{"yong_shen.position=" + fmt.Sprint(yongPos), "line.dong_self=false"},
			"旺相静爻可为暗动候选；休囚则须辨日破",
		)
	}

	return candidates
}
