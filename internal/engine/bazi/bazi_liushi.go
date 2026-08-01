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
	HourGan  ganzhi.Gan           `json:"hour_stem"`
	HourZhi  ganzhi.Zhi           `json:"hour_branch"`
	HourName string        `json:"hour_name"`
	ShiShen   string        `json:"shi_shen"`
	GanRels  []GanRelation `json:"gan_rels"`
	ZhiRels  []ZhiRelation `json:"zhi_rels"`
}

// hourBranchIndex maps date hour to traditional "时辰" branch index (0-11, 0=子).
// 23:00-00:59 → 0, 01:00-02:59 → 1, etc.
func hourBranchIndex(hour int) int {
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
	hbi := hourBranchIndex(hour)
	hourBranch := ganzhi.Zhi(hbi + 1)
	hourStem := ganzhi.Gan(((int(dp.Gan)*2 + int(hourBranch) - 2) % 10))
	if hourStem == 0 {
		hourStem = 10
	}

	tgName := ganzhi.ShiShenFromGan(riYuan, hourStem)

	hourName := ganzhi.GanName(hourStem) + ganzhi.ZhiName(hourBranch)

	// Hour vs bazi: all 4 pillars, consistent with liunian.
	stemRels, branchRels := analyzeZhuWithBazi(ganzhi.Zhu{Gan: hourStem, Zhi: hourBranch}, bz)

	return &LiuShi{
		Time:     ganzhi.HourRanges[hbi],
		HourGan:  hourStem,
		HourZhi:  hourBranch,
		HourName: hourName,
		ShiShen:   tgName.String(),
		GanRels:  stemRels,
		ZhiRels:  branchRels,
	}, nil
}
