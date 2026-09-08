package qimen

import "liki-engine/internal/engine/ganzhi"

// XingInteraction holds star-gong interaction data.
type XingInteraction struct {
	Star       string `json:"xing"`
	Gong       string `json:"gong"`
	Name       string `json:"name"`
	Meaning    string `json:"meaning"`
	Auspicious bool   `json:"auspicious"`
}

// computeXingInteractions returns star-gong 克应 for each gong.
func computeXingInteractions(pan pan) []XingInteraction {
	result := []XingInteraction{}
	for i := 0; i < 9; i++ {
		p := pan.GongWei[i]
		for _, item := range p.TianPan {
			key := [2]int{int(item.Star), i}
			if entry, ok := xingGongTable[key]; ok {
				entry.Name = xingInteractionName(item.Star, GongIndex(i+1))
				result = append(result, entry)
			}
		}
	}
	return result
}

func xingInteractionName(star StarIndex, gong GongIndex) string {
	if gong == GongZhong {
		return starWuxing(star).String() + "星入中宫"
	}
	return starWuxing(star).String() + "星入" + palaceWuxing(gong).String() + "宫"
}

// XingGongWuXing represents 五行关系 of a star in a gong.
type XingGongWuXing struct {
	Star             StarIndex `json:"xing"`
	Gong             GongIndex `json:"gong"`
	Relation         string    `json:"relation"`
	RelationName     string    `json:"relation_name"`
	TraditionalLabel string    `json:"traditional_label"`
}

// computeXingGongWuXing computes the 五行关系 for each star in the pan.
func computeXingGongWuXing(pan pan) []XingGongWuXing {
	result := []XingGongWuXing{}
	for i, p := range pan.GongWei {
		for _, item := range p.TianPan {
			relation := xingGongRelation(item.Star, GongIndex(i+1))
			result = append(result, XingGongWuXing{
				Star: item.Star, Gong: GongIndex(i + 1),
				Relation: relation, RelationName: wuxingRelations[relation].Name,
				TraditionalLabel: wuxingRelations[relation].TraditionalLabel,
			})
		}
	}
	return result
}

// starWuxing returns the element of a star.
func starWuxing(s StarIndex) ganzhi.Wuxing {
	if s >= 1 && int(s) <= len(starWuxingTable) {
		return starWuxingTable[int(s)-1]
	}
	return 0
}

func xingGongRelation(star StarIndex, palace GongIndex) string {
	starElem := starWuxing(star)
	palElem := palaceWuxing(palace)
	if starElem == palElem {
		return "same"
	}
	switch {
	case ganzhi.Sheng(palElem, starElem):
		return "gong_generates_xing"
	case ganzhi.Sheng(starElem, palElem):
		return "xing_generates_gong"
	case ganzhi.Ke(starElem, palElem):
		return "xing_controls_gong"
	default:
		return "gong_controls_xing"
	}
}
