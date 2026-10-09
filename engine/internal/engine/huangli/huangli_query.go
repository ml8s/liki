package huangli

import (
	"fmt"
	"time"

	"liki-engine/internal/engine/ganzhi"
)

// Day is the huangli query result for one day.
type ShiChenFortune struct {
	Zhi         string `json:"zhi"`
	Time        string `json:"time"` // e.g. "23:00-01:00"
	HuangDaoStr string `json:"huangdao"`
	JianChu     string `json:"jian_chu"`
	Suitable    bool   `json:"suitable"`
}

type Day struct {
	Date        string           `json:"date"`
	RiZhu       riZhuInfo        `json:"ri_zhu"`
	NaYin       string           `json:"na_yin"`
	Wuxing      string           `json:"wuxing"`
	JianChu     string           `json:"jian_chu"`
	HuangDao    huangDaoStar     `json:"huangdao"`
	XiShen      string           `json:"xi_shen"`
	CaiShen     string           `json:"cai_shen"`
	FuShen      string           `json:"fu_shen"`
	StemTaboo   string           `json:"gan_ji"`
	BranchTaboo string           `json:"zhi_ji"`
	Mansion     dayMansion       `json:"mansion"`
	JieQi       string           `json:"jie_qi"`
	JieQiDays   int              `json:"jie_qi_days"`
	RenYuan     string           `json:"ren_yuan"`
	ShiChen     []ShiChenFortune `json:"shi_chen,omitempty"`
	Event       string           `json:"event,omitempty"`
	EventLabel  string           `json:"event_label,omitempty"`
	Suitability string           `json:"suitability,omitempty"`
	Reason      string           `json:"reason,omitempty"`
	Warnings    []string         `json:"warnings,omitempty"`
}

// Month holds monthly huangli data.
type Month struct {
	Month string `json:"month"`
	Gan   string `json:"gan"`
	Zhi   string `json:"zhi"`
	Days  []Day  `json:"days"`
}

func renYuanName(ry renYuanSiLing) string {
	if ry.Current == nil {
		return ""
	}
	return ry.Current.GanName
}

// QueryDate returns huangli info for a single date.
func QueryDate(dateStr string, event ...string) (Day, error) {
	t, err := time.Parse("2006-01-02", dateStr)
	if err != nil {
		return Day{}, fmt.Errorf("huangli: parse date %s: %w", dateStr, err)
	}

	dpi := lookupRiZhu(t)
	yueZhi := yueZhuForDate(t).Zhi
	jq := computeJieQiDepth(t.Year(), int(t.Month()), t.Day())
	ry := computeRenYuanSiLingForDate(yueZhi, jq.DaysIn)

	entry := Day{
		Date:        dateStr,
		RiZhu:       dpi,
		JianChu:     lookupJianChu(t),
		HuangDao:    huangDaoForDay(yueZhi, dpi.Zhi),
		XiShen:      xiShenDirection(dpi.Gan),
		CaiShen:     caiShenDirection(dpi.Gan),
		FuShen:      fuShenDirection(dpi.Gan),
		StemTaboo:   pengZuGanTaboo(dpi.Gan),
		BranchTaboo: pengZuZhiTaboo(dpi.Zhi),
		NaYin:       ganzhi.NayinLabel(dpi.Gan, dpi.Zhi),
		Wuxing:      ganzhi.ZhiWuxing(dpi.Zhi).String(),
		Mansion:     mansionForDay(ganzhi.Zhu{Gan: dpi.Gan, Zhi: dpi.Zhi}),
		JieQi:       jq.TermName,
		JieQiDays:   jq.DaysIn,
		RenYuan:     renYuanName(ry),
	}

	entry.ShiChen = computeShiChen(dpi.Zhi, yueZhi, entry.JianChu)
	if len(event) > 0 {
		if err := applyEventRule(&entry, event[0]); err != nil {
			return Day{}, err
		}
	}

	return entry, nil
}

func applyEventRule(entry *Day, event string) error {
	if event == "" {
		entry.Suitability = "unspecified"
		entry.Reason = "未指定事项，仅列黄历事实。"
		return nil
	}
	rule, ok := jianChuCfg.EventRules[event]
	if !ok {
		return fmt.Errorf("huangli: unknown event %q", event)
	}
	entry.Event = event
	entry.EventLabel = rule.Label
	switch {
	case containsJianChu(rule.Suitable, entry.JianChu):
		entry.Suitability = "recommended"
		entry.Reason = "建除「" + entry.JianChu + "」适合" + rule.Label + "。"
	case containsJianChu(rule.Forbidden, entry.JianChu):
		entry.Suitability = "unsuitable"
		entry.Reason = "建除「" + entry.JianChu + "」忌" + rule.Label + "。"
	default:
		entry.Suitability = "possible"
		entry.Reason = "建除「" + entry.JianChu + "」无明确适配或冲突。"
	}
	return nil
}

func containsJianChu(values []string, target string) bool {
	for _, value := range values {
		if value == target {
			return true
		}
	}
	return false
}

// QueryMonth returns huangli entries for every day in the given month.
func QueryMonth(yearMonth string) (Month, error) {
	t, err := time.Parse("2006-01", yearMonth)
	if err != nil {
		return Month{}, fmt.Errorf("huangli: parse year-month %s: %w", yearMonth, err)
	}

	year, month := t.Year(), int(t.Month())
	daysInMonth := time.Date(year, time.Month(month+1), 0, 0, 0, 0, 0, time.UTC).Day()

	var days []Day
	for d := 1; d <= daysInMonth; d++ {
		dateStr := fmt.Sprintf("%04d-%02d-%02d", year, month, d)
		entry, err := QueryDate(dateStr)
		if err != nil {
			return Month{}, err
		}
		days = append(days, entry)
	}
	mp := yueZhuForDate(t)
	return Month{
		Month: yearMonth,
		Gan:   ganzhi.GanName(mp.Gan),
		Zhi:   ganzhi.ZhiName(mp.Zhi),
		Days:  days,
	}, nil
}
