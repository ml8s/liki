package huangli

import (
	"time"

	"liki-engine/internal/engine/ganzhi"
)

type BondDay struct {
	Day
	GanRelation    string `json:"gan_relation"`
	ZhiRelation    string `json:"zhi_relation"`
	TaiSuiRelation string `json:"tai_sui_relation"`
}

type BondMonth struct {
	Month  string    `json:"month"`
	Stem   string    `json:"gan"`
	Branch string    `json:"zhi"`
	Days   []BondDay `json:"days"`
}

func computeBondDay(bz ganzhi.Bazi, eventType string, dateStr string) (BondDay, error) {
	riYuan, dayBranch := bz.Ri.Gan, bz.Ri.Zhi
	dayEntry, err := QueryDate(dateStr)
	if err != nil {
		return BondDay{}, err
	}
	dayStem := dayEntry.RiZhu.Gan
	riZhi := dayEntry.RiZhu.Zhi
	t, _ := time.Parse("2006-01-02", dateStr) //nolint:errcheck
	taiSui := taiSui(t.Year())
	relZhi, _, _ := evaluateZhi(riZhi, dayBranch, "日柱")
	relTS, _, _ := evaluateZhi(riZhi, taiSui, "太岁")
	return BondDay{Day: dayEntry, GanRelation: ganzhi.ShiShenFromGan(riYuan, dayStem).String(), ZhiRelation: relZhi, TaiSuiRelation: relTS}, nil
}

func computeBondMonth(bz ganzhi.Bazi, eventType string, yearMonth string) (BondMonth, error) {
	riYuan, dayBranch := bz.Ri.Gan, bz.Ri.Zhi
	m, err := QueryMonth(yearMonth)
	if err != nil {
		return BondMonth{}, err
	}
	t, _ := time.Parse("2006-01", yearMonth) //nolint:errcheck
	taiSui := taiSui(t.Year())
	r := BondMonth{Month: m.Month, Stem: m.Stem, Branch: m.Branch, Days: make([]BondDay, len(m.Days))}
	for i, e := range m.Days {
		ds := e.RiZhu.Gan
		dz := e.RiZhu.Zhi
		r.Days[i] = BondDay{Day: e}
		r.Days[i].GanRelation = ganzhi.ShiShenFromGan(riYuan, ds).String()
		r.Days[i].ZhiRelation, _, _ = evaluateZhi(dz, dayBranch, "日柱")
		r.Days[i].TaiSuiRelation, _, _ = evaluateZhi(dz, taiSui, "太岁")
	}
	return r, nil
}
