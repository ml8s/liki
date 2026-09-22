package bazi

import (
	"fmt"

	"liki-engine/internal/engine/ganzhi"
)

// GanRelation describes a single gan-to-gan relationship.
type GanRelation struct {
	GanA     ganzhi.Gan `json:"gan_a"`
	GanB     ganzhi.Gan `json:"gan_b"`
	Type     string     `json:"type"`
	Relation string     `json:"relation"`
}

// ZhiRelation describes a single zhi-to-zhi relationship.
type ZhiRelation struct {
	ZhiA   ganzhi.Zhi `json:"zhi_a"`
	ZhiB   ganzhi.Zhi `json:"zhi_b"`
	Type   string     `json:"type"`
	Detail string     `json:"detail"`
}

// Relation type constants for gan and zhi interactions.
const (
	relSame         = "相同"
	relSheng        = "相生"
	relKe           = "相克"
	relGanHe        = "天干五合"
	relNone         = "无"
	relLiuHe        = "六合"
	relSanHe        = "三合"
	relSanHePartial = "半合"
	relSanHeGong    = "三合拱"
	relSanHui       = "三会"
	relSanHuiGong   = "三会拱"
	relLiuChong     = "六冲"
	relXing         = "相刑"
	relLiuHai       = "六害"
	relAnHe         = "暗合"
	relPo           = "破"
)

// zhuInteraction holds gan and zhi relations for one pillar against the bazi chart.
type zhuInteraction struct {
	ZhuLabel string        `json:"pillar_label"`
	GanRels  []GanRelation `json:"gan_rels"`
	ZhiRels  []ZhiRelation `json:"zhi_rels"`
}

// analyzeGanRelation checks the relationship between two gan.
func analyzeGanRelation(a, b ganzhi.Gan) GanRelation {
	r := GanRelation{GanA: a, GanB: b}
	if a == b {
		r.Type = relSame
		r.Relation = fmt.Sprintf("%s%s同气", ganzhi.GanName(a), ganzhi.GanName(b))
		return r
	}

	for _, p := range ganzhi.GanHes {
		if (a == p.A && b == p.B) || (a == p.B && b == p.A) {
			r.Type = relGanHe
			r.Relation = fmt.Sprintf("%s%s合化%s", ganzhi.GanName(a), ganzhi.GanName(b), p.Result.String())
			return r
		}
	}

	aElem, bElem := ganzhi.GanWuxing(a), ganzhi.GanWuxing(b)
	if aElem == bElem {
		r.Type = relSame
		r.Relation = fmt.Sprintf("%s%s同气", ganzhi.GanName(a), ganzhi.GanName(b))
	} else if ganzhi.Sheng(aElem, bElem) {
		r.Type = relSheng
		r.Relation = fmt.Sprintf("%s生%s", ganzhi.GanName(a), ganzhi.GanName(b))
	} else if ganzhi.Sheng(bElem, aElem) {
		r.Type = relSheng
		r.Relation = fmt.Sprintf("%s生%s", ganzhi.GanName(b), ganzhi.GanName(a))
	} else if ganzhi.Ke(aElem, bElem) {
		r.Type = relKe
		r.Relation = fmt.Sprintf("%s克%s", ganzhi.GanName(a), ganzhi.GanName(b))
	} else if ganzhi.Ke(bElem, aElem) {
		r.Type = relKe
		r.Relation = fmt.Sprintf("%s克%s", ganzhi.GanName(b), ganzhi.GanName(a))
	} else {
		r.Type = relNone
		r.Relation = "无特殊关系"
	}
	return r
}

// analyzeZhiRelation checks relationship types between exactly two zhi.
// A pair alone can never be a complete 三合 / 三会: two members with the
// imperial branch are 半合; the first-and-last pair is a 拱 candidate.  Full
// groups must be detected from the whole chart by ComputeHeHui.
func analyzeZhiRelation(a, b ganzhi.Zhi) ZhiRelation {
	r := ZhiRelation{ZhiA: a, ZhiB: b}
	if a == b {
		// 同支：辰午酉亥 同支相见为自刑，其余为同气。
		if ganzhi.IsXing(a, b) {
			r.Type = relXing
			r.Detail = fmt.Sprintf("%s%s自刑", ganzhi.ZhiName(a), ganzhi.ZhiName(b))
			return r
		}
		r.Type = relSame
		r.Detail = fmt.Sprintf("%s%s同气", ganzhi.ZhiName(a), ganzhi.ZhiName(b))
		return r
	}

	for _, p := range ganzhi.ZhiHes {
		if (a == p.A && b == p.B) || (a == p.B && b == p.A) {
			r.Type = relLiuHe
			r.Detail = fmt.Sprintf("%s%s合化%s", ganzhi.ZhiName(a), ganzhi.ZhiName(b), p.Result.String())
			return r
		}
	}

	if rel, ok := pairTripleHeRelation(a, b); ok {
		return rel
	}

	for _, p := range ganzhi.ChongPairs {
		if (a == p.A && b == p.B) || (a == p.B && b == p.A) {
			r.Type = relLiuChong
			r.Detail = fmt.Sprintf("%s%s相冲", ganzhi.ZhiName(a), ganzhi.ZhiName(b))
			return r
		}
	}

	// 相刑：与 ganzhi.IsXing 同口径——自刑（辰午酉亥）仅同支相见为刑，
	// 组内不同支（如午亥、辰酉）不构成刑。
	// 优先级：六合/半合 / 拱 / 六冲 > 相刑 > 六害（寅巳相刑显著于六害；寅申冲/丑未冲显著于刑）。
	if ganzhi.IsXing(a, b) {
		r.Type = relXing
		xType := "刑"
		for _, x := range ganzhi.XingGroups {
			if containsPair(x.Zhi, a, b) {
				xType = xingTypeLabel(x.Type)
				break
			}
		}
		r.Detail = fmt.Sprintf("%s%s%s", ganzhi.ZhiName(a), ganzhi.ZhiName(b), xType)
		return r
	}

	for _, p := range ganzhi.HaiPairs {
		if (a == p.A && b == p.B) || (a == p.B && b == p.A) {
			r.Type = relLiuHai
			r.Detail = fmt.Sprintf("%s%s相害", ganzhi.ZhiName(a), ganzhi.ZhiName(b))
			return r
		}
	}

	if rel, ok := pairTripleHuiRelation(a, b); ok {
		return rel
	}

	if ganzhi.IsAnHe(a, b) {
		r.Type = relAnHe
		r.Detail = fmt.Sprintf("%s%s暗合", ganzhi.ZhiName(a), ganzhi.ZhiName(b))
		return r
	}

	if ganzhi.IsPo(a, b) {
		r.Type = relPo
		r.Detail = fmt.Sprintf("%s%s相破", ganzhi.ZhiName(a), ganzhi.ZhiName(b))
		return r
	}

	r.Type = relNone
	r.Detail = "无特殊关系"
	return r
}

// pairTripleHeRelation classifies two members of a trine. The trine's second
// branch is the imperial / prosperous branch; a half combination must include
// it. The first and last branches form the classical arch candidate only.
func pairTripleHeRelation(a, b ganzhi.Zhi) (ZhiRelation, bool) {
	r := ZhiRelation{ZhiA: a, ZhiB: b}
	for _, tr := range ganzhi.TripleHeList {
		if !containsPair(tr.Zhi, a, b) {
			continue
		}
		imperial := tr.Zhi[1]
		if a == imperial || b == imperial {
			var missing ganzhi.Zhi
			for _, z := range tr.Zhi {
				if z != a && z != b {
					missing = z
					break
				}
			}
			r.Type = relSanHePartial
			r.Detail = fmt.Sprintf("%s%s半合%s局，缺%s", ganzhi.ZhiName(a), ganzhi.ZhiName(b), tr.Element.String(), ganzhi.ZhiName(missing))
			return r, true
		}
		if (a == tr.Zhi[0] && b == tr.Zhi[2]) || (a == tr.Zhi[2] && b == tr.Zhi[0]) {
			r.Type = relSanHeGong
			r.Detail = fmt.Sprintf("%s%s拱%s，%s局候选", ganzhi.ZhiName(a), ganzhi.ZhiName(b), ganzhi.ZhiName(imperial), tr.Element.String())
			return r, true
		}
	}
	return ZhiRelation{}, false
}

// pairTripleHuiRelation classifies the first-and-last seasonal pair as an arch
// candidate. Other two-branch subsets are not a 三会 and return no relation.
func pairTripleHuiRelation(a, b ganzhi.Zhi) (ZhiRelation, bool) {
	r := ZhiRelation{ZhiA: a, ZhiB: b}
	for _, group := range ganzhi.TripleHuiList {
		if (a == group.Zhi[0] && b == group.Zhi[2]) || (a == group.Zhi[2] && b == group.Zhi[0]) {
			r.Type = relSanHuiGong
			r.Detail = fmt.Sprintf("%s%s拱%s，%s方候选", ganzhi.ZhiName(a), ganzhi.ZhiName(b), ganzhi.ZhiName(group.Zhi[1]), group.Element.String())
			return r, true
		}
	}
	return ZhiRelation{}, false
}

const (
	xingWuLi   = "无礼之刑"
	xingWuEn   = "无恩之刑"
	xingShiShi = "恃势之刑"
	xingZi     = "自刑"
)

func xingTypeLabel(t string) string {
	switch t {
	case "wuli":
		return xingWuLi
	case "wuen":
		return xingWuEn
	case "shishi":
		return xingShiShi
	case "zi":
		return xingZi
	}
	return "刑"
}

// analyzeZhuWithBazi analyzes one pillar against all 4 bazi chart pillars.
func analyzeZhuWithBazi(zhu ganzhi.Zhu, bz ganzhi.Bazi) ([]GanRelation, []ZhiRelation) {
	ganRels := make([]GanRelation, 4)
	zhiRels := make([]ZhiRelation, 4)
	for i, np := range bz.Slice() {
		ganRels[i] = analyzeGanRelation(zhu.Gan, np.Gan)
		zhiRels[i] = analyzeZhiRelation(zhu.Zhi, np.Zhi)
	}
	return ganRels, zhiRels
}
