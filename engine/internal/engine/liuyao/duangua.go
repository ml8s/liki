package liuyao

import "strings"

import "liki-engine/internal/engine/ganzhi"

// DayRelation describes 日建与爻的关系.
type DayRelation struct {
	Relations []string `json:"relations"` // 冲/合/生/扶/克/平；冲合可与其他关系并存
}

func dayInteraction(lineZhi ganzhi.Zhi, riZhi ganzhi.Zhi) DayRelation {
	le := ganzhi.ZhiWuxing(lineZhi)
	de := ganzhi.ZhiWuxing(riZhi)

	rel := DayRelation{Relations: []string{}}

	if ganzhi.IsLiuChong(lineZhi, riZhi) {
		rel.Relations = append(rel.Relations, "冲")
	}
	if ganzhi.IsZhiHe(lineZhi, riZhi) {
		rel.Relations = append(rel.Relations, "合")
	}
	if ganzhi.Sheng(de, le) {
		rel.Relations = append(rel.Relations, "生")
	}
	if le == de {
		rel.Relations = append(rel.Relations, "扶")
	}
	if ganzhi.Ke(de, le) {
		rel.Relations = append(rel.Relations, "克")
	}
	if len(rel.Relations) == 0 {
		rel.Relations = append(rel.Relations, "平")
	}
	return rel
}

func (p *Chart) dayRelationHas(position int, names ...string) bool {
	if position < 1 || position > 6 {
		return false
	}
	for _, relation := range p.DayRelations[position-1].Relations {
		for _, name := range names {
			if relation == name {
				return true
			}
		}
	}
	return false
}

func (r DayRelation) String() string {
	return strings.Join(r.Relations, "兼")
}

func ordinal(n int) string {
	names := [7]string{"", "初", "二", "三", "四", "五", "上"}
	if n >= 1 && n <= 6 {
		return names[n]
	}
	return "?"
}

func chongZhi(z ganzhi.Zhi) ganzhi.Zhi {
	return ganzhi.ChongZhi(z)
}
