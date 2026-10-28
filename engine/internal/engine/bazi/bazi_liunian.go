package bazi

import (
	"liki-engine/internal/engine/ganzhi"
)

// LiuNian holds the annual (流年) analysis output.
type LiuNian struct {
	Year              int                 `json:"year"`
	NianGan           ganzhi.Gan          `json:"nian_gan"`
	NianZhi           ganzhi.Zhi          `json:"nian_zhi"`
	YearName          string              `json:"nian_name"`
	Element           string              `json:"wuxing"`
	NaYin             string              `json:"na_yin"`
	ShiShen           string              `json:"shi_shen"`
	Generates         int                 `json:"sheng"`
	Restrains         int                 `json:"ke"`
	NatalInteractions []zhuInteraction    `json:"natal_interactions"`
	DaYunInteractions []zhuInteraction    `json:"dayun_interactions"`
	ShenSha           []ShenShaProjection `json:"shensha"`
	FuYinFanYin       []FuYinFanYin       `json:"fuyin_fanyin"`
	AtomicFacts       LiuNianAtomicFacts  `json:"atomic_facts"`
}

// ComputeLiuNian computes the year pillar for a resolved 立春 year and
// analyzes its relationship to the day master.
func computeLiuNian(bz ganzhi.Bazi, t LiuNianTime, currentDaYun *DaYunStep) (*LiuNian, error) {
	riYuan := bz.Ri.Gan
	nianGan, nianZhi := t.NianGan, t.NianZhi

	dmElem := ganzhi.GanWuxing(riYuan)
	yearElem := ganzhi.GanWuxing(nianGan)

	tgName := ganzhi.ShiShenFromGan(riYuan, nianGan)

	gen, rest := countGenRest(yearElem, dmElem)

	naYin := ganzhi.NayinLabel(nianGan, nianZhi)

	r := &LiuNian{
		Year:      t.LichunYear,
		NianGan:   nianGan,
		NianZhi:   nianZhi,
		YearName:  ganzhi.GanName(nianGan) + ganzhi.ZhiName(nianZhi),
		Element:   yearElem.String(),
		NaYin:     naYin,
		ShiShen:   tgName.String(),
		Generates: gen,
		Restrains: rest,
	}

	// Three-layer analysis when bazi chart and current dayun are available.
	liuNianZhu := ganzhi.Zhu{Gan: nianGan, Zhi: nianZhi}
	r.NatalInteractions = make([]zhuInteraction, 1)
	ganRels, zhiRels := analyzeZhuWithBazi(liuNianZhu, bz)
	r.NatalInteractions[0] = zhuInteraction{
		ZhuLabel: r.YearName,
		GanRels:  ganRels,
		ZhiRels:  zhiRels,
	}

	if currentDaYun != nil {
		dyZhu := ganzhi.Zhu{Gan: currentDaYun.Gan, Zhi: currentDaYun.Zhi}
		dyGanRels, dyZhiRels := analyzeZhuWithBazi(dyZhu, bz)
		r.DaYunInteractions = []zhuInteraction{{
			ZhuLabel: currentDaYun.ShiShen + "(" + currentDaYun.Name + ")",
			GanRels:  dyGanRels,
			ZhiRels:  dyZhiRels,
		}}
	} else {
		r.DaYunInteractions = []zhuInteraction{}
	}

	r.ShenSha = computeDynamicShenSha(nianZhi, bz.Nian.Zhi, bz.Ri.Zhi, riYuan)
	r.ShenSha = append(r.ShenSha, computeAnnualShenSha(nianZhi, bz)...)
	r.FuYinFanYin = computeFuYinFanYin(liuNianZhu, bz)
	r.AtomicFacts = computeLiuNianAtomicFacts(bz, t.LichunYear, nianGan, nianZhi, tgName, currentDaYun)

	return r, nil
}

func countGenRest(elem, dmElem ganzhi.Wuxing) (gen, rest int) {
	if elem == dmElem {
		return 1, 0
	}
	if ganzhi.Sheng(elem, dmElem) {
		return 1, 0
	}
	if ganzhi.Sheng(dmElem, elem) {
		return 0, 1
	}
	if ganzhi.Ke(elem, dmElem) {
		return 0, 1
	}
	return 1, 0
}
