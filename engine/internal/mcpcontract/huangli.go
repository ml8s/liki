package mcpcontract

import (
	"context"
	"fmt"

	"liki-engine/internal/engine/huangli"
	"liki-engine/internal/engine/tianwen"
)

// HuangliDaysInput selects a contiguous date range and optional life event.
type HuangliDaysInput struct {
	StandardTime StandardTime  `json:"standard_time" jsonschema:"目标地标准时间（无 DST），RFC3339 带 offset。"`
	Count        *HuangliCount `json:"count,omitempty" jsonschema:"查几天，默认3，最多30。"`
	Event        HuangliEvent  `json:"event,omitempty" jsonschema:"事项类型；缺省仅列黄历事实。"`
}

type HuangliRiZhuOutput struct {
	Gan   string `json:"gan" jsonschema:"日干。"`
	Zhi   string `json:"zhi" jsonschema:"日支。"`
	NaYin string `json:"na_yin" jsonschema:"日柱纳音。"`
}

type HuangliHuangDaoOutput struct {
	Index    int    `json:"index" jsonschema:"黄道/黑道序号。"`
	Name     string `json:"name" jsonschema:"黄道/黑道神名。"`
	Path     string `json:"path" jsonschema:"黄道或黑道。"`
	Sequence int    `json:"sequence" jsonschema:"十二神循环位置。"`
}

type HuangliMansionOutput struct {
	Index    int    `json:"index" jsonschema:"二十八宿序号。"`
	Name     string `json:"name" jsonschema:"二十八宿名称。"`
	DongWu   string `json:"dong_wu" jsonschema:"七曜动物。"`
	Wuxing   string `json:"wuxing" jsonschema:"宿五行。"`
	Group    string `json:"group" jsonschema:"四象分组。"`
	GroupIdx int    `json:"group_idx" jsonschema:"四象内序号。"`
}

type HuangliShiChenOutput struct {
	Zhi      string `json:"zhi" jsonschema:"时辰地支。"`
	Time     string `json:"time" jsonschema:"时辰钟点区间。"`
	HuangDao string `json:"huangdao" jsonschema:"该时辰黄道/黑道神。"`
	JianChu  string `json:"jian_chu" jsonschema:"该时辰建除。"`
	Suitable bool   `json:"suitable" jsonschema:"该时辰是否宜。"`
}

// HuangliDayOutput is one day's almanac wire record.
type HuangliDayOutput struct {
	Date        string                 `json:"date" jsonschema:"日期 YYYY-MM-DD。"`
	RiZhu       HuangliRiZhuOutput     `json:"ri_zhu" jsonschema:"日柱信息。"`
	NaYin       string                 `json:"na_yin" jsonschema:"日柱纳音。"`
	Wuxing      string                 `json:"wuxing" jsonschema:"日支五行。"`
	JianChu     string                 `json:"jian_chu" jsonschema:"建除十二神。"`
	HuangDao    HuangliHuangDaoOutput  `json:"huangdao" jsonschema:"当日黄道/黑道。"`
	XiShen      string                 `json:"xi_shen" jsonschema:"喜神方位。"`
	CaiShen     string                 `json:"cai_shen" jsonschema:"财神方位。"`
	FuShen      string                 `json:"fu_shen" jsonschema:"福神方位。"`
	StemTaboo   string                 `json:"gan_ji" jsonschema:"彭祖日干忌。"`
	BranchTaboo string                 `json:"zhi_ji" jsonschema:"彭祖日支忌。"`
	Mansion     HuangliMansionOutput   `json:"mansion" jsonschema:"当日二十八宿。"`
	JieQi       string                 `json:"jie_qi" jsonschema:"节气。"`
	JieQiDays   int                    `json:"jie_qi_days" jsonschema:"进入节气天数。"`
	RenYuan     string                 `json:"ren_yuan" jsonschema:"人元司令天干。"`
	ShiChen     []HuangliShiChenOutput `json:"shi_chen,omitempty" jsonschema:"十二时辰宜忌。"`
	Event       string                 `json:"event,omitempty" jsonschema:"查询事项。"`
	EventLabel  string                 `json:"event_label,omitempty" jsonschema:"事项中文标签。"`
	Suitability string                 `json:"suitability,omitempty" jsonschema:"事项适配结论。"`
	Reason      string                 `json:"reason,omitempty" jsonschema:"事项适配依据。"`
}

// HuangliDays returns almanac facts for consecutive dates.
func HuangliDays(_ context.Context, input HuangliDaysInput) ([]HuangliDayOutput, error) {
	standard, loc, err := parseStandardTime(input.StandardTime)
	if err != nil {
		return nil, fmt.Errorf("huangli.days: %w", err)
	}
	start := tianwen.StandardDateAt(standard, loc)
	count := 3
	if input.Count != nil {
		count = int(*input.Count)
	}
	// 边界由 HuangliCount 的 schema 约束保证（1-30）；此处不再静默修正。

	days := make([]HuangliDayOutput, 0, count)
	for i := 0; i < count; i++ {
		date := start.AddDays(i).String()
		entry, err := huangli.QueryDate(date, string(input.Event))
		if err != nil {
			return nil, fmt.Errorf("huangli.days: %w", err)
		}
		days = append(days, huangliDayWire(entry))
	}
	return days, nil
}

func huangliDayWire(day huangli.Day) HuangliDayOutput {
	output := HuangliDayOutput{
		Date: day.Date,
		RiZhu: HuangliRiZhuOutput{
			Gan:   day.RiZhu.Gan.String(),
			Zhi:   day.RiZhu.Zhi.String(),
			NaYin: day.RiZhu.NaYin,
		},
		NaYin:       day.NaYin,
		Wuxing:      day.Wuxing,
		JianChu:     day.JianChu,
		HuangDao:    HuangliHuangDaoOutput(day.HuangDao),
		XiShen:      day.XiShen,
		CaiShen:     day.CaiShen,
		FuShen:      day.FuShen,
		StemTaboo:   day.StemTaboo,
		BranchTaboo: day.BranchTaboo,
		Mansion: HuangliMansionOutput{
			Index:    day.Mansion.Index,
			Name:     day.Mansion.Name,
			DongWu:   day.Mansion.DongWu,
			Wuxing:   day.Mansion.Element,
			Group:    day.Mansion.Group,
			GroupIdx: day.Mansion.GroupIdx,
		},
		JieQi:       day.JieQi,
		JieQiDays:   day.JieQiDays,
		RenYuan:     day.RenYuan,
		Event:       day.Event,
		EventLabel:  day.EventLabel,
		Suitability: day.Suitability,
		Reason:      day.Reason,
	}
	for _, item := range day.ShiChen {
		output.ShiChen = append(output.ShiChen, HuangliShiChenOutput{
			Zhi:      item.Zhi,
			Time:     item.Time,
			HuangDao: item.HuangDaoStr,
			JianChu:  item.JianChu,
			Suitable: item.Suitable,
		})
	}
	return output
}
