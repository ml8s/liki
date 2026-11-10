// Engine wire DTOs for the MCP contract layer.
//
// These types are ordinary source now, not generated artifacts. Edit them
// directly when the published engine payload changes.
package mcpcontract

import "liki-engine/internal/engine/ziwei"

// ── ziwei.chart input ──

// ZiweiChartInput consumes a resolved natal subject. The chart domain derives
// the local astronomical lunar date and LAST shichen from the subject.
type ZiweiChartInput struct {
	Subject NatalSubject `json:"subject" jsonschema:"已归一化的出生主体。"`
}

type ZiweiDaXianItems []ZiweiDaXianItem

type ZiweiDaXianItem struct {
	StartYear int             `json:"start_year" jsonschema:"本限起始公历年"`
	EndYear   int             `json:"end_year" jsonschema:"本限结束公历年"`
	QiSui     int             `json:"qi_sui" jsonschema:"起运虚岁。"`
	ZhiSui    int             `json:"zhi_sui" jsonschema:"止运虚岁。"`
	MingGong  ZiweiPalaceName `json:"ming_gong" jsonschema:"大限命宫。"`
}

// ── ziwei.liupan input ──

type ZiweiLiupanInput struct {
	Scope        ZiweiLiupanScopeName `json:"scope" jsonschema:"流盘层级。"`
	Subject      NatalSubject         `json:"subject" jsonschema:"已归一化的出生主体。"`
	StandardTime StandardTime         `json:"standard_time" jsonschema:"流盘选择起点：local standard time，RFC3339 fixed offset。"`
	Count        *SeriesCount         `json:"count,omitempty" jsonschema:"连续查询数量。"`
	Longitude    *Longitude           `json:"longitude,omitempty" jsonschema:"地点经度；liu_shi 必填（-180..180），真太阳时校正用于时辰。"`
}

// ZiweiLiupanOutput is the raw-handler schema marker; the published contract is
// the four named domain roots as optional top-level properties.
type ZiweiLiupanOutput struct{}

// ── ziwei.liupan liu_nian output ──

type ZiweiLiuNianOutput struct {
	MingGong ZiweiPalaceName           `json:"ming_gong" jsonschema:"命宫。"`
	Zhi      ZhiName                   `json:"zhi" jsonschema:"流年地支"`
	SiHua    []ZiweiFlowTransformation `json:"si_hua" jsonschema:"四化事实：星、化、落宫。"`
	GongWei  []ZiweiFlowPalace         `json:"gong_wei" jsonschema:"流年十二宫盘。"`
	NianJie  ZiweiLiuNianResolved      `json:"nian_jie" jsonschema:"流年农历年界。"`
}

type ZiweiLiuNianResolved struct {
	LiuNianStandardTime StandardTime `json:"standard_time" jsonschema:"流年选择器时区标准时间（无 DST）。"`
	LiuNianStartDate    DateString   `json:"start_date" jsonschema:"按选定农历基准取的流年选择日期。"`
	LunarYear           int          `json:"lunar_year" jsonschema:"该日期所在的农历年。"`
	LunarYearStart      DateString   `json:"lunar_year_start" jsonschema:"正月初一公历日（含）。"`
	LunarYearEnd        DateString   `json:"lunar_year_end" jsonschema:"下一正月初一公历日（不含）。"`
	BoundaryDay         bool         `json:"boundary_day" jsonschema:"输入日期是否为正月初一边界日（归入新年）。"`
}

type ZiweiFlowTransformation struct {
	Star   string          `json:"star" jsonschema:"被化的本命星。"`
	Hua    SiHuaName       `json:"hua" jsonschema:"化禄/化权/化科/化忌。"`
	Palace ZiweiPalaceName `json:"palace" jsonschema:"本命星所在宫。"`
}

type ZiweiFlowPalace struct {
	Zhi       ZhiName         `json:"zhi" jsonschema:"地支。"`
	Name      ZiweiPalaceName `json:"name" jsonschema:"宫名。"`
	Stars     []string        `json:"stars" jsonschema:"流曜。"`
	IsLiuMing bool            `json:"is_liu_ming" jsonschema:"是否流盘命宫。"`
}

// ── ziwei.liupan liu_yue output ──

type ZiweiLunarDate struct {
	Year  int  `json:"year" jsonschema:"农历年。"`
	Month int  `json:"month" jsonschema:"农历月。"`
	Day   int  `json:"day" jsonschema:"农历日。"`
	Leap  bool `json:"leap" jsonschema:"是否闰月。"`
}

type ZiweiYueJie struct {
	LiuYueStandardTime StandardTime                            `json:"standard_time" jsonschema:"流月选择器时区标准时间（无 DST）。"`
	LiuYueStartDate    DateString                              `json:"start_date" jsonschema:"按选定农历基准取的流月选择日期。"`
	Lunar              ZiweiLunarDate                          `json:"lunar" jsonschema:"该日期对应的农历日期。"`
	LunarMonthStart    DateString                              `json:"lunar_month_start" jsonschema:"农历月初一公历日（含）。"`
	LunarMonthEnd      DateString                              `json:"lunar_month_end" jsonschema:"下一农历月初一公历日（不含）。"`
	BoundaryDay        bool                                    `json:"boundary_day" jsonschema:"输入日期是否为农历月初一边界日（归入新月）。"`
	CalendarPeriod     ZiweiLiuYueResolvedPeriodCalendarPeriod `json:"calendar_period" jsonschema:"流月所在日历区间。"`
	FlowMonth          ZiweiLiuYueResolvedPeriodFlowMonth      `json:"flow_month" jsonschema:"流月农历年月。"`
}

type ZiweiLiuYueOutput struct {
	YueJie   ZiweiYueJie               `json:"yue_jie" jsonschema:"流月农历月界与月令。"`
	MingGong ZiweiPalaceName           `json:"ming_gong" jsonschema:"命宫。"`
	Zhi      ZhiName                   `json:"zhi" jsonschema:"流月地支"`
	SiHua    []ZiweiFlowTransformation `json:"si_hua" jsonschema:"四化事实：星、化、落宫。"`
	GongWei  []ZiweiFlowPalace         `json:"gong_wei" jsonschema:"流月十二宫盘。"`
}

type ZiweiLiuYueTargetLunar struct {
	Year  int  `json:"year" jsonschema:"年。"`
	Month int  `json:"month" jsonschema:"月。"`
	Day   int  `json:"day" jsonschema:"日。"`
	Leap  bool `json:"leap" jsonschema:"是否闰月。"`
}

type ZiweiLiuYueResolvedPeriodCalendarPeriod struct {
	Year     int                    `json:"year" jsonschema:"年。"`
	Month    int                    `json:"month" jsonschema:"月。"`
	Leap     bool                   `json:"leap" jsonschema:"是否闰月。"`
	Half     CalendarPeriodHalfName `json:"half" jsonschema:"农历半月；whole 为整月，闰月按 first / second 拆分。"`
	DayStart int                    `json:"day_start" jsonschema:"流月日段起始农历日序。"`
	DayEnd   int                    `json:"day_end" jsonschema:"流月日段结束农历日序。"`
}

type ZiweiLiuYueResolvedPeriodFlowMonth struct {
	Year  int  `json:"year" jsonschema:"年。"`
	Month int  `json:"month" jsonschema:"月。"`
	Leap  bool `json:"leap" jsonschema:"是否闰月。"`
}

type ZiweiLiuRiJie struct {
	LiuRiStandardTime StandardTime   `json:"standard_time" jsonschema:"流日选择器时区标准时间（无 DST）。"`
	LiuRiLocalDate    DateString     `json:"local_date" jsonschema:"流日当地民用日。"`
	Lunar             ZiweiLunarDate `json:"lunar" jsonschema:"该日期对应的农历日期。"`
}

// ── ziwei.liupan liu_ri output ──

type ZiweiLiuRiOutput struct {
	RiJie    ZiweiLiuRiJie             `json:"ri_jie" jsonschema:"流日日界。"`
	MingGong ZiweiPalaceName           `json:"ming_gong" jsonschema:"命宫。"`
	Zhi      ZhiName                   `json:"zhi" jsonschema:"流日地支"`
	SiHua    []ZiweiFlowTransformation `json:"si_hua" jsonschema:"四化事实：星、化、落宫。"`
	GongWei  []ZiweiFlowPalace         `json:"gong_wei" jsonschema:"流日十二宫盘。"`
}

// ── ziwei.liupan liu_shi output ──

type ZiweiLiuShiChen struct {
	LiuShiStandardTime StandardTime   `json:"standard_time" jsonschema:"流时选择器时区标准时间（无 DST）。"`
	LiuShiLocalDate    DateString     `json:"local_date" jsonschema:"流时当地标准日。"`
	Lunar              ZiweiLunarDate `json:"lunar" jsonschema:"该日期对应的农历日期。"`
	Shichen            ZhiName        `json:"shichen" jsonschema:"真太阳时时辰。"`
}

type ZiweiLiuShiOutput struct {
	MingGong ZiweiPalaceName           `json:"ming_gong" jsonschema:"命宫。"`
	Zhi      ZhiName                   `json:"zhi" jsonschema:"流时地支"`
	SiHua    []ZiweiFlowTransformation `json:"si_hua" jsonschema:"四化事实：星、化、落宫。"`
	GongWei  []ZiweiFlowPalace         `json:"gong_wei" jsonschema:"流时十二宫盘。"`
	ShiChen  ZiweiLiuShiChen           `json:"shi_chen" jsonschema:"流时时辰定位。"`
}

// ── ziwei.bond input ──

type ZiweiBondSide struct {
	Subject NatalSubject `json:"subject" jsonschema:"已归一化的出生主体。"`
}

type ZiweiBondInput struct {
	A ZiweiBondSide `json:"a" jsonschema:"甲方：已归一化的出生主体。"`
	B ZiweiBondSide `json:"b" jsonschema:"乙方：已归一化的出生主体。"`
}

// ── ziwei.bond output ──

type ZiweiBondOutput struct {
	MingGong ZiweiBondPalacePair `json:"ming_gong" jsonschema:"双方本命宫。"`
	FuQiGong ZiweiBondPalacePair `json:"fu_qi_gong" jsonschema:"双方夫妻宫。"`
	ZiNvGong ZiweiBondPalacePair `json:"zi_nv_gong" jsonschema:"双方子女宫。"`
}

type ZiweiBondPalaceFacts = ziwei.BondPalaceFacts

type ZiweiBondPalacePair = ziwei.BondPalacePair

// ── ziwei.chart output ──

type ZiweiChartOutput struct {
	KouJing         ZiweiChartKouJing      `json:"kou_jing" jsonschema:"排盘口径。"`
	ShiErGong       [12]ZiweiChartGongItem `json:"shi_er_gong" jsonschema:"十二宫；从命宫起顺布。"`
	JuShu           JuShuCode              `json:"ju_shu" jsonschema:"局数机器码。"`
	JuShuName       string                 `json:"ju_shu_name" jsonschema:"局数名。"`
	KongGong        []ZiweiKongGongJieXing `json:"kong_gong,omitempty" jsonschema:"空宫借星；只表达宫位关系，星曜由 jie_gong 宫承载。"`
	NianGan         GanName                `json:"nian_gan" jsonschema:"年干。"`
	NianZhi         ZhiName                `json:"nian_zhi,omitempty" jsonschema:"年支。"`
	ShiZhi          ZhiName                `json:"shi_zhi" jsonschema:"时支。"`
	BirthYear       int                    `json:"birth_year" jsonschema:"出生公历年。"`
	LunarMonth      int                    `json:"lunar_month,omitempty" jsonschema:"排盘调整农历月。"`
	LunarDay        int                    `json:"lunar_day,omitempty" jsonschema:"排盘农历日。"`
	BirthLunarMonth int                    `json:"birth_lunar_month,omitempty" jsonschema:"出生原始农历月。"`
	BirthIsLeap     bool                   `json:"birth_is_leap,omitempty" jsonschema:"出生农历月是否闰月。"`
	MingZhu         string                 `json:"ming_zhu,omitempty" jsonschema:"命主。"`
	ShenZhu         string                 `json:"shen_zhu,omitempty" jsonschema:"身主。"`
	GeJu            []ZiweiGeJu            `json:"ge_ju,omitempty" jsonschema:"紫微格局候选。"`
	SanFangSiZheng  ZiweiSanFangSiZheng    `json:"san_fang_si_zheng" jsonschema:"命宫三方四正的宫位关系；星曜由所引宫位承载。"`
	DaXian          ZiweiDaXianItems       `json:"da_xian" jsonschema:"十年大限序列。"`
}

type ZiweiChartKouJing struct {
	LeapMonth SchoolLeapMonthName `json:"leap_month" jsonschema:"闰月取月口径。"`
	Source    SchoolSourceName    `json:"source" jsonschema:"排盘口径来源。"`
}

type ZiweiKongGongJieXing struct {
	KongGong ZiweiPalaceName `json:"kong_gong" jsonschema:"无主星宫。"`
	JieGong  ZiweiPalaceName `json:"jie_gong" jsonschema:"借星来源宫；其主星即空宫所借。"`
}

type ZiweiSanFangSiZheng struct {
	MingGong []ZiweiPalaceName `json:"ming_gong" jsonschema:"三方四正宫位：命宫、财帛、官禄、迁移。"`
}

type ZiweiGeJu struct {
	Description *string `json:"description,omitempty" jsonschema:"说明；展示文本由内容层管理。"`
	Name        string  `json:"name" jsonschema:"格局名。"`
}

type ZiweiChartGongItem struct {
	Ming       ZiweiPalaceName      `json:"ming" jsonschema:"宫名。"`
	Gan        GanName              `json:"gan" jsonschema:"宫干。"`
	Zhi        ZhiName              `json:"zhi" jsonschema:"宫支。"`
	ShenGong   *bool                `json:"shen_gong,omitempty" jsonschema:"是否身宫。"`
	LaiYinGong *bool                `json:"lai_yin_gong,omitempty" jsonschema:"是否来因宫。"`
	XingYao    []ZiweiStar          `json:"xing_yao" jsonschema:"星曜列表。"`
	XiaoXian   *[]int               `json:"xiao_xian,omitempty" jsonschema:"小限年龄。"`
	ChangSheng *ChangShengStageName `json:"chang_sheng,omitempty" jsonschema:"长生十二宫。"`
	BoShi      *string              `json:"bo_shi,omitempty" jsonschema:"博士十二神。"`
	JiangQian  *string              `json:"jiang_qian,omitempty" jsonschema:"将前十二神。"`
	SuiQian    *string              `json:"sui_qian,omitempty" jsonschema:"岁前十二神。"`
	ZaYao      *[]string            `json:"za_yao,omitempty" jsonschema:"杂曜。"`
}

type ZiweiStar struct {
	Lei     []string           `json:"lei,omitempty" jsonschema:"星曜类别：zhu_xing/fu_xing/liu_ji/wen_xing/sha_xing。"`
	LiangDu *string            `json:"liang_du,omitempty" jsonschema:"亮度。"`
	Ming    *string            `json:"ming" jsonschema:"星曜名。"`
	SiHua   *SiHuaNameNullable `json:"si_hua,omitempty" jsonschema:"四化。"`
}

// ── named natal series roots ──

type ZiweiLiuNianSeries struct {
	LiuNian []ZiweiLiuNianOutput `json:"liu_nian" jsonschema:"流年数组，长度 = count。"`
}

type ZiweiLiuYueSeries struct {
	LiuYue []ZiweiLiuYueOutput `json:"liu_yue" jsonschema:"流月数组，长度 = count。"`
}

type ZiweiLiuRiSeries struct {
	LiuRi []ZiweiLiuRiOutput `json:"liu_ri" jsonschema:"流日数组，长度 = count。"`
}

type ZiweiLiuShiSeries struct {
	LiuShi []ZiweiLiuShiOutput `json:"liu_shi" jsonschema:"流时数组，长度 = count。"`
}
