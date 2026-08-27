package bazi

import (
	"time"

	"liki-engine/internal/engine/ganzhi"
	"liki-engine/internal/engine/tianwen"
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

// liushiZhiIdx maps date hour to traditional "时辰" zhi index (0-11, 0=子).
// 23:00-00:59 → 0, 01:00-02:59 → 1, etc.
func liushiZhiIdx(hour int) int {
	switch {
	case hour >= 23 || hour < 1:
		return 0
	default:
		return (hour-1)/2 + 1
	}
}

// ComputeLiuShi computes the hour pillar for the given day and hour, and its
// interactions with the bazi chart.
func computeLiuShi(bz ganzhi.Bazi, year, month, day, hour int) (*LiuShi, error) {
	riYuan := bz.Ri.Gan

	dp := tianwen.RiZhu(tianwen.GregorianTime(time.Date(year, time.Month(month), day, 0, 0, 0, 0, time.UTC)))
	hbi := liushiZhiIdx(hour)
	shiZhi := ganzhi.Zhi(hbi + 1)
	shiGan := ganzhi.Gan(((int(dp.Gan)*2 + int(shiZhi) - 2) % 10))
	if shiGan == 0 {
		shiGan = 10
	}

	tgName := ganzhi.ShiShenFromGan(riYuan, shiGan)

	hourName := ganzhi.GanName(shiGan) + ganzhi.ZhiName(shiZhi)

	// Hour vs bazi: all 4 pillars, consistent with liunian.
	ganRels, zhiRels := analyzeZhuWithBazi(ganzhi.Zhu{Gan: shiGan, Zhi: shiZhi}, bz)

	return &LiuShi{
		Time:     ganzhi.HourRanges[hbi],
		ShiGan:   shiGan,
		ShiZhi:   shiZhi,
		HourName: hourName,
		ShiShen:  tgName.String(),
		GanRels:  ganRels,
		ZhiRels:  zhiRels,
	}, nil
}
