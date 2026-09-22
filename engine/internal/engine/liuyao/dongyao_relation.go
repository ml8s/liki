package liuyao

import "liki-engine/internal/engine/ganzhi"

// DongYaoRelationType 动爻与用神的关系类型（枚举，正交化）。
type DongYaoRelationType string

const (
	RelationShengYong DongYaoRelationType = "生用"  // 动爻生用神
	RelationKeYong    DongYaoRelationType = "克用"  // 动爻克用神
	RelationBiHe      DongYaoRelationType = "比和"  // 动爻与用神五行相同
	RelationChongYong DongYaoRelationType = "冲用"  // 动爻地支冲用神
	RelationShengYuan DongYaoRelationType = "生原神" // 动爻生原神
	RelationKeYuan    DongYaoRelationType = "克原神" // 动爻克原神
	RelationShengJi   DongYaoRelationType = "生忌神" // 动爻生忌神
	RelationKeJi      DongYaoRelationType = "克忌神" // 动爻克忌神
)

// DongYaoRelation 一个动爻与用神的关系集合。
// 直接作用与间接作用可以并存；静卦没有记录，不使用伪关系填充。
type DongYaoRelation struct {
	Position  int                   `json:"position"` // 动爻位置 1-6
	Relations []DongYaoRelationType `json:"relations"`
}

// computeDongYaoRelations 计算每个动爻与用神的关系（8 种枚举）。
// 原神 = 生用神五行的五行；忌神 = 克用神五行的五行。
func computeDongYaoRelations(p *Chart, yongShenType YongShen) []DongYaoRelation {
	yongPos := p.findYongShen(yongShenType)
	var yWuxing ganzhi.Wuxing
	var yongZhi ganzhi.Zhi
	if yongPos > 0 {
		yongLine := p.Lines[yongPos-1]
		yWuxing = yongLine.Wuxing
		yongZhi = yongLine.Zhi
	} else if fuShen := p.findFuShen(yongShenType); fuShen != nil {
		yongZhi = fuShenZhi(fuShen)
		yWuxing = ganzhi.ZhiWuxing(yongZhi)
	} else {
		return nil // 本卦和本宫伏神都不现
	}

	var yuanWuxing, jiWuxing ganzhi.Wuxing
	for _, wx := range []ganzhi.Wuxing{ganzhi.WxMu, ganzhi.WxHuo, ganzhi.WxTu, ganzhi.WxJin, ganzhi.WxShui} {
		if ganzhi.Sheng(wx, yWuxing) {
			yuanWuxing = wx
		}
		if ganzhi.Ke(wx, yWuxing) {
			jiWuxing = wx
		}
	}

	var relations []DongYaoRelation
	for _, dpos := range p.DongYao {
		if dpos < 1 || dpos > 6 {
			continue
		}
		dLine := p.Lines[dpos-1]
		if dLine.Position == yongPos {
			continue // 动爻即用神本身
		}
		dWuxing := dLine.Wuxing
		rel := DongYaoRelation{Position: dpos, Relations: []DongYaoRelationType{}}
		appendRelation := func(item DongYaoRelationType) {
			for _, existing := range rel.Relations {
				if existing == item {
					return
				}
			}
			rel.Relations = append(rel.Relations, item)
		}

		if ganzhi.Sheng(dWuxing, yWuxing) {
			appendRelation(RelationShengYong)
		} else if ganzhi.Ke(dWuxing, yWuxing) {
			appendRelation(RelationKeYong)
		} else if dWuxing == yWuxing {
			appendRelation(RelationBiHe)
		}
		if yongZhi > 0 && ganzhi.IsLiuChong(dLine.Zhi, yongZhi) {
			appendRelation(RelationChongYong)
		}
		if ganzhi.Sheng(dWuxing, yuanWuxing) {
			appendRelation(RelationShengYuan)
		} else if ganzhi.Ke(dWuxing, yuanWuxing) {
			appendRelation(RelationKeYuan)
		}
		if ganzhi.Sheng(dWuxing, jiWuxing) {
			appendRelation(RelationShengJi)
		} else if ganzhi.Ke(dWuxing, jiWuxing) {
			appendRelation(RelationKeJi)
		}
		if len(rel.Relations) > 0 {
			relations = append(relations, rel)
		}
	}
	return relations
}
