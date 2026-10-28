package bazi

import (
	"liki-engine/internal/engine/ganzhi"
)

// LiuYue holds the monthly (流月) analysis output.
type LiuYue struct {
	Year      int                 `json:"year"`
	Month     int                 `json:"month"`
	YueGan    ganzhi.Gan          `json:"yue_gan"`
	YueZhi    ganzhi.Zhi          `json:"yue_zhi"`
	MonthName string              `json:"month_name"`
	Element   string              `json:"wuxing"`
	ShiShen   string              `json:"shi_shen"`
	Generates int                 `json:"sheng"`
	Restrains int                 `json:"ke"`
	GanRels   []GanRelation       `json:"gan_rels"`
	ZhiRels   []ZhiRelation       `json:"zhi_rels"`
	ShenSha   []ShenShaProjection `json:"shensha"`
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
		Year:      t.Date.Year,
		Month:     t.Date.Month,
		YueGan:    mp.Gan,
		YueZhi:    mp.Zhi,
		MonthName: ganzhi.ZhiName(mp.Zhi) + "月",
		Element:   monthElemStr,
		ShiShen:   tgName.String(),
		Generates: gen,
		Restrains: rest,
		GanRels:   ganRels,
		ZhiRels:   zhiRels,
		ShenSha:   shensha,
	}, nil
}
