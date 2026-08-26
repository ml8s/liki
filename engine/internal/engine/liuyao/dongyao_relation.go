package liuyao

import "liki-engine/internal/engine/ganzhi"

// DongYaoRelationType 动爻与用神的关系类型（枚举，正交化）。
type DongYaoRelationType string

const (
	RelationShengYong DongYaoRelationType = "生用"  // 动爻生用神 → 吉
	RelationKeYong    DongYaoRelationType = "克用"  // 动爻克用神 → 凶
	RelationBiHe      DongYaoRelationType = "比和"  // 动爻与用神五行相同 → 助
	RelationChongYong DongYaoRelationType = "冲用"  // 动爻地支冲用神 → 散
	RelationShengYuan DongYaoRelationType = "生原神" // 动爻生原神 → 间接吉
	RelationKeYuan    DongYaoRelationType = "克原神" // 动爻克原神 → 用神失助
	RelationShengJi   DongYaoRelationType = "生忌神" // 动爻生忌神 → 间接凶
	RelationKeJi      DongYaoRelationType = "克忌神" // 动爻克忌神 → 间接吉
	RelationNone      DongYaoRelationType = "无动爻" // 静卦
)

// DongYaoRelation 一个动爻与用神的关系。
type DongYaoRelation struct {
	Position int                 `json:"position"` // 动爻位置 1-6
	Relation DongYaoRelationType `json:"relation"` // 关系类型
}

// computeDongYaoRelations 计算每个动爻与用神的关系（9 种枚举）。
// 原神 = 生用神五行的五行；忌神 = 克用神五行的五行。
func computeDongYaoRelations(p *Chart, yongShenType YongShen) []DongYaoRelation {
	yongPos, _ := p.findYongShen(yongShenType)
	if yongPos == 0 {
		return nil // 用神不现
	}
	yongLine := p.Lines[yongPos-1]
	yWuxing := yongLine.Wuxing

	// 原神五行 = 生用神者；忌神五行 = 克用神者
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

		rel := DongYaoRelation{Position: dpos}

		// 直接关系
		if ganzhi.Sheng(dWuxing, yWuxing) {
			rel.Relation = RelationShengYong
		} else if ganzhi.Ke(dWuxing, yWuxing) {
			rel.Relation = RelationKeYong
		} else if dWuxing == yWuxing {
			rel.Relation = RelationBiHe
		}
		// 冲用
		if ganzhi.IsLiuChong(dLine.Zhi, yongLine.Zhi) {
			rel.Relation = RelationChongYong
		}
		// 通过原神/忌神
		if yuanWuxing != 0 {
			if ganzhi.Sheng(dWuxing, yuanWuxing) {
				rel.Relation = RelationShengYuan
			} else if ganzhi.Ke(dWuxing, yuanWuxing) {
				rel.Relation = RelationKeYuan
			}
		}
		if jiWuxing != 0 {
			if ganzhi.Sheng(dWuxing, jiWuxing) {
				rel.Relation = RelationShengJi
			} else if ganzhi.Ke(dWuxing, jiWuxing) {
				rel.Relation = RelationKeJi
			}
		}

		if rel.Relation != "" {
			relations = append(relations, rel)
		}
	}
	return relations
}
