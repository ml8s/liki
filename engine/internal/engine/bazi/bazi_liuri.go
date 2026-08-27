package bazi

import (
	"fmt"
	"time"

	"liki-engine/internal/engine/ganzhi"
	"liki-engine/internal/engine/tianwen"
)

// LiuRi holds the daily flow (流日) analysis: today's pillar
// and its interactions with the bazi chart, dayun, and liunian pillars.
type LiuRi struct {
	Date        string         `json:"date"`
	RiGan       ganzhi.Gan     `json:"ri_gan"`
	RiZhi       ganzhi.Zhi     `json:"ri_zhi"`
	DayName     string         `json:"day_name"`
	DayNaYin    string         `json:"day_nayin"`
	ShiShen     string         `json:"shi_shen"`
	GanRels     []GanRelation  `json:"gan_rels"`
	ZhiRels     []ZhiRelation  `json:"zhi_rels"`
	DaYunRels   []ZhiRelation  `json:"dayun_rels"`
	LiuNianRels []ZhiRelation  `json:"liunian_rels"`
	ShenSha     []shenShaEntry `json:"shensha"`
}

// ComputeLiuRi computes the day pillar for the given date and its full
// interactions with the bazi chart, current dayun, and current liunian.
func computeLiuRi(bz ganzhi.Bazi, year, month, day int, daYunZhu *ganzhi.Zhu, liuNianZhu *ganzhi.Zhu) (*LiuRi, error) {
	riYuan := bz.Ri.Gan
	bazi := bz.Slice()

	dp := tianwen.RiZhu(tianwen.GregorianTime(time.Date(year, time.Month(month), day, 0, 0, 0, 0, time.UTC)))
	tgName := ganzhi.ShiShenFromGan(riYuan, dp.Gan)

	dayName := ganzhi.GanName(dp.Gan) + ganzhi.ZhiName(dp.Zhi)

	// Day vs bazi (gan + zhi relations) — all 4 pillars, consistent with liunian.
	ganRels, zhiRels := analyzeZhuWithBazi(dp, bz)

	// Day vs dayun.
	DaYunRels := make([]ZhiRelation, 0)
	if daYunZhu != nil {
		DaYunRels = append(DaYunRels, analyzeZhiRelation(dp.Zhi, daYunZhu.Zhi))
	}

	// Day vs liunian.
	liunianRels := make([]ZhiRelation, 0)
	if liuNianZhu != nil {
		liunianRels = append(liunianRels, analyzeZhiRelation(dp.Zhi, liuNianZhu.Zhi))
	}

	// Na yin.
	naYin := ganzhi.NayinLabel(dp.Gan, dp.Zhi)

	// Daily shensha: day gan/zhi vs bazi.
	var shensha []shenShaEntry
	// 天乙贵人 on day gan.
	if targets, ok := tianYiLookup[dp.Gan]; ok {
		for _, tb := range targets {
			for _, np := range bazi {
				if np.Zhi == tb {
					shensha = append(shensha, shenShaEntry{Name: "天乙贵人", Category: catJi, Description: "流日天乙贵人日"})
				}
			}
		}
	}
	// 文昌 on day gan.
	if targets, ok := wenChangLookup[dp.Gan]; ok {
		for _, tb := range targets {
			for _, np := range bazi {
				if np.Zhi == tb {
					shensha = append(shensha, shenShaEntry{Name: "文昌", Category: catJi, Description: "流日文昌日，利学业文书"})
				}
			}
		}
	}
	// 驿马/桃花/华盖 from year zhi triad → day zhi check.
	yBranch := bazi[0].Zhi
	triadMaps := []struct {
		m    map[ganzhi.Zhi]ganzhi.Zhi
		name string
		cat  string
		desc string
	}{
		{yimaZhiMap, "驿马", catZhongXing, "流日驿马，动象"},
		{taohuaZhiMap, "桃花", catZhongXing, "流日桃花，异性缘佳"},
		{huagaiZhiMap, "华盖", catZhongXing, "流日华盖，宜静思"},
	}
	for _, tm := range triadMaps {
		if tb, ok := tm.m[yBranch]; ok && dp.Zhi == tb {
			shensha = append(shensha, shenShaEntry{Name: tm.name, Category: tm.cat, Description: tm.desc})
		}
	}

	if shensha == nil {
		shensha = []shenShaEntry{}
	}
	return &LiuRi{
		Date:        fmt.Sprintf("%04d-%02d-%02d", year, month, day),
		RiGan:       dp.Gan,
		RiZhi:       dp.Zhi,
		DayName:     dayName,
		DayNaYin:    naYin,
		ShiShen:     tgName.String(),
		GanRels:     ganRels,
		ZhiRels:     zhiRels,
		DaYunRels:   DaYunRels,
		LiuNianRels: liunianRels,
		ShenSha:     shensha,
	}, nil
}
