package bazi

import (
	"fmt"

	"liki-engine/internal/engine/ganzhi"
)

// GanRelation describes a single stem-to-stem relationship.
type GanRelation struct {
	GanA     ganzhi.Gan `json:"gan_a"`
	GanB     ganzhi.Gan `json:"gan_b"`
	Type     string     `json:"type"`
	Relation string     `json:"relation"`
}

// ZhiRelation describes a single branch-to-branch relationship.
type ZhiRelation struct {
	ZhiA   ganzhi.Zhi `json:"zhi_a"`
	ZhiB   ganzhi.Zhi `json:"zhi_b"`
	Type   string     `json:"type"`
	Detail string     `json:"detail"`
}

// Relation type constants for stem and branch interactions.
const (
	relSame     = "相同"
	relSheng    = "相生"
	relKe       = "相克"
	relGanHe    = "天干五合"
	relNone     = "无"
	relLiuHe    = "六合"
	relSanHe    = "三合"
	relSanHui   = "三会"
	relLiuChong = "六冲"
	relXing     = "相刑"
	relLiuHai   = "六害"
	relAnHe     = "暗合"
	relPo       = "破"
)

// zhuInteraction holds stem and branch relations for one pillar against the bazi chart.
type zhuInteraction struct {
	ZhuLabel string        `json:"pillar_label"`
	GanRels  []GanRelation `json:"gan_rels"`
	ZhiRels  []ZhiRelation `json:"zhi_rels"`
}

// analyzeGanRelation checks the relationship between two stems.
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

// analyzeZhiRelation checks all relationship types between two branches.
// Priority: 六合 > 三合 > 六冲 > 六害 > 三会 > 相刑 > 暗合 > 破.
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

	for _, th := range ganzhi.TripleHeList {
		if containsPair(th.Branches, a, b) {
			r.Type = relSanHe
			r.Detail = fmt.Sprintf("%s%s三合%s局", ganzhi.ZhiName(a), ganzhi.ZhiName(b), th.Element.String())
			return r
		}
	}

	for _, p := range ganzhi.ChongPairs {
		if (a == p.A && b == p.B) || (a == p.B && b == p.A) {
			r.Type = relLiuChong
			r.Detail = fmt.Sprintf("%s%s相冲", ganzhi.ZhiName(a), ganzhi.ZhiName(b))
			return r
		}
	}

	// 相刑：与 ganzhi.IsXing 同口径——自刑（辰午酉亥）仅同支相见为刑，
	// 组内不同支（如午亥、辰酉）不构成刑。此前用 containsPair 会把午亥等误判为自刑。
	// 优先级：六合/三合/六冲 > 相刑 > 六害（寅巳相刑显著于六害；寅申冲/丑未冲显著于刑）。
	if ganzhi.IsXing(a, b) {
		r.Type = relXing
		xType := "刑"
		for _, x := range ganzhi.XingGroups {
			if containsPair(x.Branches, a, b) {
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

	for _, th := range ganzhi.TripleHuiList {
		if containsPair(th.Branches, a, b) {
			r.Type = relSanHui
			r.Detail = fmt.Sprintf("%s%s三会%s方", ganzhi.ZhiName(a), ganzhi.ZhiName(b), th.Element.String())
			return r
		}
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
	stemRels := make([]GanRelation, 4)
	branchRels := make([]ZhiRelation, 4)
	for i, np := range bz.Slice() {
		stemRels[i] = analyzeGanRelation(zhu.Gan, np.Gan)
		branchRels[i] = analyzeZhiRelation(zhu.Zhi, np.Zhi)
	}
	return stemRels, branchRels
}
