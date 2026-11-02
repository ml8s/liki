package bazi

import (
	"liki-engine/internal/engine/ganzhi"
)

// LiuYue holds the monthly (流月) analysis output.
type LiuYue struct {
	Year         int                 `json:"year"`
	Month        int                 `json:"month"`
	YueGan       ganzhi.Gan          `json:"gan"`
	YueZhi       ganzhi.Zhi          `json:"zhi"`
	YueGanWuxing string              `json:"gan_wuxing"`
	ShiShen      string              `json:"shi_shen"`
	ShengRi      int                 `json:"sheng_ri"`
	KeRi         int                 `json:"ke_ri"`
	GanRels      []GanRelation       `json:"gan_rels"`
	ZhiRels      []ZhiRelation       `json:"zhi_rels"`
	ShenSha      []ShenShaProjection `json:"shen_sha"`
}

// ComputeLiuYue computes the month pillar for a resolved solar-term month and
// analyzes its interactions with the bazi chart.
func computeLiuYue(bz ganzhi.Bazi, t LiuYueTime) (*LiuYue, error) {
	riYuan := bz.Ri.Gan
	bazi := bz.Slice()

	mp := ganzhi.Zhu{Gan: t.YueGan, Zhi: t.YueZhi}

	dmElem := ganzhi.GanWuxing(riYuan)
	monthElem := ganzhi.GanWuxing(mp.Gan)

	tgName := ganzhi.ShiShenFromGan(riYuan, mp.Gan)

	gen, rest := countGenRest(monthElem, dmElem)

	// Month vs bazi: all 4 pillars, consistent with liunian.
	ganRels, zhiRels := analyzeZhuWithBazi(mp, bz)

	shensha := computeDynamicShenSha(mp.Zhi, bazi[0].Zhi, bazi[2].Zhi, riYuan)

	monthElemStr := monthElem.String()

	return &LiuYue{
		Year:         t.Date.Year,
		Month:        t.Date.Month,
		YueGan:       mp.Gan,
		YueZhi:       mp.Zhi,
		YueGanWuxing: monthElemStr,
		ShiShen:      tgName.String(),
		ShengRi:      gen,
		KeRi:         rest,
		GanRels:      ganRels,
		ZhiRels:      zhiRels,
		ShenSha:      shensha,
	}, nil
}
