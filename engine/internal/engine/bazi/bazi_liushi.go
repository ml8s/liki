package bazi

import (
	"liki-engine/internal/engine/ganzhi"
)

// LiuShi holds the hourly flow (流时) analysis: the current two-hour period's
// pillar and its interactions with the bazi chart.
type LiuShi struct {
	Time     string        `json:"time"`
	ShiGan   ganzhi.Gan    `json:"shi_gan"`
	ShiZhi   ganzhi.Zhi    `json:"shi_zhi"`
	HourName string        `json:"hour_name"`
	ShiShen  string        `json:"shi_shen"`
	GanRels  []GanRelation `json:"gan_rels"`
	ZhiRels  []ZhiRelation `json:"zhi_rels"`
}

// computeLiuShi analyzes a resolved LAST hour pillar against the bazi chart.
func computeLiuShi(bz ganzhi.Bazi, t LiuShiTime) (*LiuShi, error) {
	riYuan := bz.Ri.Gan
	shiGan := t.ShiGan
	shiZhi := t.ShiZhi

	tgName := ganzhi.ShiShenFromGan(riYuan, shiGan)
	hourName := ganzhi.GanName(shiGan) + ganzhi.ZhiName(shiZhi)

	// Hour vs bazi: all 4 pillars, consistent with liunian.
	ganRels, zhiRels := analyzeZhuWithBazi(ganzhi.Zhu{Gan: shiGan, Zhi: shiZhi}, bz)

	return &LiuShi{
		Time:     ganzhi.HourRanges[int(shiZhi)-1],
		ShiGan:   shiGan,
		ShiZhi:   shiZhi,
		HourName: hourName,
		ShiShen:  tgName.String(),
		GanRels:  ganRels,
		ZhiRels:  zhiRels,
	}, nil
}
