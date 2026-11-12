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
	return append(candidates, candidate)
}

func appendBranchTrigger(
	candidates []TimingCandidate, idPrefix, mechanism string, position int,
	branch ganzhi.Zhi, basis []string,
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
	})
}

// staticHarmonySources names the actual sources that can lift a static line.
// A six-harmony relation alone is insufficient: 《增删卜易·六合章》 defines
// 静而逢合 as 合起 when the partner comes from the day, month, or a moving line.
func staticHarmonySources(p *Chart, partner ganzhi.Zhi) []string {
	var sources []string
	if partner == p.RiZhi {
		sources = append(sources, "day_branch")
	}
	if partner == p.YueZhi {
		sources = append(sources, "month_branch")
	}
	for _, pos := range p.DongYao {
		if pos >= 1 && pos <= len(p.Lines) && p.Lines[pos-1].Zhi == partner {
			sources = append(sources, fmt.Sprintf("moving_yao.position=%d", pos))
		}
	}
	return sources
}

// computeTimingCandidates derives trigger hypotheses from the already computed
// YongShen state. It intentionally does not assign a probability or a final
// date; the caller must judge which candidates apply to the concrete question.
func computeTimingCandidates(p *Chart, typ YongShen) []TimingCandidate {
	var candidates []TimingCandidate
	yongPos := p.findYongShen(typ)
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
			)
		}
		return candidates
	}

	yao := p.Lines[yongPos-1]
	layer := "ben"
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
		)
		if partner := zhiHePartner(yao.Zhi); partner != 0 {
			candidates = appendBranchTrigger(
				candidates,
				"moving-harmony",
				"动爻逢合",
				yongPos,
				partner,
				[]string{fmt.Sprintf("%s_yao.type=%d", layer, int(yao.Type)), "branch_relation=六合"},
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
		)
		candidates = appendBranchTrigger(
			candidates,
			"void-clash",
			"冲空",
			yongPos,
			chongZhi(yao.Zhi),
			[]string{"yong_shen.xun_kong=true"},
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
		)
		if partner := zhiHePartner(yao.Zhi); partner != 0 {
			candidates = appendBranchTrigger(
				candidates,
				"month-break-harmony",
				"月破逢合",
				yongPos,
				partner,
				[]string{"yong_shen.yue_po=true"},
			)
		}
		candidates = appendBranchTrigger(
			candidates,
			"month-break-exit",
			"出月令",
			yongPos,
			chongZhi(p.YueZhi),
			[]string{"yong_shen.yue_po=true", "month_branch=" + ganzhi.ZhiName(p.YueZhi)},
		)
	}

	if !moving && ganzhi.IsLiuChong(yao.Zhi, p.RiZhi) {
		candidates = appendBranchTrigger(
			candidates,
			"static-clash",
			"静爻逢冲",
			yongPos,
			chongZhi(yao.Zhi),
			[]string{"yong_shen.position=" + fmt.Sprint(yongPos), "line.dong_self=false"},
		)
	}

	// Keep this separate from void/break handling. Those states already have
	// dedicated release paths, and a weak static line still can be lifted by
	// harmony when a day/month/moving source actually hits it.
	if !moving && !yong.XunKong && !yong.YuePo {
		if partner := zhiHePartner(yao.Zhi); partner != 0 {
			sources := staticHarmonySources(p, partner)
			if len(sources) > 0 {
				basis := append([]string{
					"yong_shen.position=" + fmt.Sprint(yongPos),
					"line.dong_self=false",
					"branch_relation=六合",
					"classical_scope=静而逢合谓之合起",
				}, sources...)
				candidates = appendBranchTrigger(
					candidates,
					"static-harmony",
					"静爻逢合",
					yongPos,
					partner,
					basis,
				)
			}
		}
	}

	return candidates
}
