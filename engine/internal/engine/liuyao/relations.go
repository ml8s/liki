package liuyao

import (
	"fmt"

	"liki-engine/internal/engine/ganzhi"
)

func lineTargets(p *Chart, left, right int) []string {
	var targets []string
	yongPos := p.YongShen.Position
	if yongPos > 0 && (left == yongPos || right == yongPos) {
		targets = append(targets, "yong_shen")
	}
	for _, line := range p.Lines {
		if line.Position != left && line.Position != right {
			continue
		}
		switch line.ShiYing {
		case "世":
			targets = append(targets, "world")
		case "应":
			targets = append(targets, "other")
		}
		if line.DongSelf {
			targets = append(targets, "moving")
		}
	}
	return targets
}

func branchRelations(left, right ganzhi.Zhi) []string {
	var relations []string
	if ganzhi.IsLiuChong(left, right) {
		relations = append(relations, "六冲")
	}
	if ganzhi.IsZhiHe(left, right) {
		relations = append(relations, "六合")
	}
	if ganzhi.IsHai(left, right) {
		relations = append(relations, "六害")
	}
	for _, group := range ganzhi.XingGroups {
		if len(group.Zhi) == 2 && group.Type == "zi" &&
			left == right && containsZhi(group.Zhi, left) {
			relations = append(relations, "自刑")
			break
		}
		if len(group.Zhi) > 2 && left != right &&
			containsZhi(group.Zhi, left) && containsZhi(group.Zhi, right) {
			relations = append(relations, "刑（成组候选）")
			break
		}
	}
	return relations
}

func computeBranchRelationFacts(p *Chart) []BranchRelationFact {
	var facts []BranchRelationFact
	for leftIndex := 0; leftIndex < 6; leftIndex++ {
		for rightIndex := leftIndex + 1; rightIndex < 6; rightIndex++ {
			left, right := p.Lines[leftIndex], p.Lines[rightIndex]
			relations := branchRelations(left.Zhi, right.Zhi)
			if len(relations) == 0 {
				continue
			}
			facts = append(facts, BranchRelationFact{
				LeftPosition:  left.Position,
				RightPosition: right.Position,
				Relations:     relations,
				Targets:       lineTargets(p, left.Position, right.Position),
				Basis: []string{
					fmt.Sprintf("left=%s", ganzhi.ZhiName(left.Zhi)),
					fmt.Sprintf("right=%s", ganzhi.ZhiName(right.Zhi)),
				},
			})
		}
	}
	return facts
}
