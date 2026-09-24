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

	// 流时时干按当日（0 点）日干起五鼠遁；晚子时(23:00-24:00)除外——
	// 与排盘 zi_shi_rule（lunar 约定）一致：晚子时日柱不变、时柱按次日日干起。
	dayForShiGan := time.Date(year, time.Month(month), day, 0, 0, 0, 0, time.UTC)
	if hour >= 23 {
		dayForShiGan = dayForShiGan.AddDate(0, 0, 1)
	}
	dp := tianwen.RiZhu(tianwen.GregorianTime(dayForShiGan))
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
