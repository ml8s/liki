package bazi

import (
	"fmt"
	"time"

	"liki-engine/internal/engine/ganzhi"
	"liki-engine/internal/engine/tianwen"
)

// LiuYue holds the monthly (流月) analysis output.
type LiuYue struct {
	Year      int            `json:"year"`
	Month     int            `json:"month"`
	MonthGan  ganzhi.Gan            `json:"yue_gan"`
	MonthZhi  ganzhi.Zhi            `json:"yue_zhi"`
	MonthName string         `json:"month_name"`
	Element   string         `json:"wuxing"`
	ShiShen    string         `json:"shi_shen"`
	Generates int            `json:"sheng"`
	Restrains int            `json:"ke"`
	GanRels   []GanRelation  `json:"gan_rels"`
	ZhiRels   []ZhiRelation  `json:"zhi_rels"`
	ShenSha   []shenShaEntry `json:"shensha"`
}

// ComputeLiuYue computes the month pillar for a given year+month and analyzes
// its interactions with the bazi chart.
func computeLiuYue(bz ganzhi.Bazi, year, month int) (*LiuYue, error) {
	if month < 1 || month > 12 {
		return nil, fmt.Errorf("compute liuyue: invalid month %d", month)
	}
	riYuan := bz.Ri.Gan
	bazi := bz.Slice()
	// Use mid-month (15th) for stable solar term calculation.
	birthTime := time.Date(year, time.Month(month), 15, 12, 0, 0, 0, time.UTC)

	mp := tianwen.YueZhu(tianwen.GregorianTime(birthTime))

	dmElem := ganzhi.GanWuxing(riYuan)
	monthElem := ganzhi.GanWuxing(mp.Gan)

	tgName := ganzhi.ShiShenFromGan(riYuan, mp.Gan)

	gen, rest := countGenRest(monthElem, dmElem)

	// Month vs bazi: all 4 pillars, consistent with liunian.
	stemRels, branchRels := analyzeZhuWithBazi(mp, bz)

	shensha := computeDynamicShenSha(mp.Zhi, bazi[0].Zhi, riYuan)

	monthElemStr := monthElem.String()

	return &LiuYue{
		Year:      year,
		Month:     month,
		MonthGan:  mp.Gan,
		MonthZhi:  mp.Zhi,
		MonthName: ganzhi.ZhiName(mp.Zhi) + "月",
		Element:   monthElemStr,
		ShiShen:    tgName.String(),
		Generates: gen,
		Restrains: rest,
		GanRels:   stemRels,
		ZhiRels:   branchRels,
		ShenSha:   shensha,
	}, nil
}
