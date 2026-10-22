// Wire DTOs for the migrated bazi/ziwei/qimen engine tools.
//
// The shapes follow the legacy raw-JSON handlers and the captured engine
// payloads under testdata/actual/. The file started as a generated
// projection; it is ordinary source now, so edit it directly and keep
// TestEngineContractWireFidelity green when a shape changes.

package mcpcontract

// ── ziwei.chart input ──

// ZiweiChartInput takes the canonical {utc, timezone} pair (Zone is IANA) and the
// birth/query longitude. The boundary derives LocalCivilTime and LAST; the
// chart domain consumes those domain scales (ChineseCalendarDate (basis configured by the caller) + LAST shichen; no separate time.Time).
type ZiweiChartInput struct {
	Instant   Instant    `json:"instant" jsonschema:"出生瞬间 RFC3339；接受 UTC Z 或 numeric offset，服务端归一化为 UTC。"`
	Timezone  string     `json:"timezone" jsonschema:"出生地 IANA 时区。"`
	Longitude float64    `json:"longitude" jsonschema:"出生地经度（-180..180）；真太阳时校正用于时辰。"`
	Gender    GenderName `json:"gender" jsonschema:"性别"`
}

// ── ziwei.chart output ──

type ZiweiChartOutput struct {
	School          ZiweiChartSchool            `json:"school" jsonschema:"流派。"`
	GongWei         []ZiweiChartGongWeiItem     `json:"gong_wei" jsonschema:"12宫信息"`
	MingGong        *ZiweiPalaceName            `json:"ming_gong,omitempty" jsonschema:"命宫。"`
	SiHua           *map[string]SiHuaName       `json:"si_hua,omitempty" jsonschema:"四化: {星名: 化名}"`
	PalaceFacts     []ZiweiChartPalaceFactsItem `json:"palace_facts" jsonschema:"宫位原子事实；Python 因子层只读"`
	ShenGong        *ZiweiPalaceName            `json:"shen_gong,omitempty" jsonschema:"身宫。"`
	JuShu           *JuShuName                  `json:"ju_shu,omitempty" jsonschema:"局数。"`
	JuShuName       *string                     `json:"ju_shu_name,omitempty" jsonschema:"ju_shu_name。"`
	MingZhu         *string                     `json:"ming_zhu,omitempty" jsonschema:"命主"`
	ShenZhu         *string                     `json:"shen_zhu,omitempty" jsonschema:"身主"`
	NianGan         *LiuyaoGanName              `json:"nian_gan,omitempty" jsonschema:"年干"`
	NianZhi         *LiuyaoZhiName              `json:"nian_zhi,omitempty" jsonschema:"年支"`
	ShiZhi          *LiuyaoZhiName              `json:"shi_zhi,omitempty" jsonschema:"时支"`
	BirthYear       *int                        `json:"birth_year,omitempty" jsonschema:"出生公历年。"`
	ZiweiPos        *ZiweiPalaceName            `json:"ziwei_pos,omitempty" jsonschema:"紫微落宫。"`
	Gender          *GenderName                 `json:"gender,omitempty" jsonschema:"性别。"`
	BirthLunarMonth *int                        `json:"birth_lunar_month,omitempty" jsonschema:"出生农历月"`
	BirthIsLeap     *bool                       `json:"birth_is_leap,omitempty" jsonschema:"出生农历月是否闰月"`
	LunarMonth      *int                        `json:"lunar_month,omitempty" jsonschema:"排盘用农历月（闰月后半月按下月）"`
	LunarDay        *int                        `json:"lunar_day,omitempty" jsonschema:"排盘用农历日"`
	KongGong        *[]ZiweiChartKongGongItem   `json:"kong_gong,omitempty" jsonschema:"空宫借星（无主星宫位借对宫主星）"`
	Patterns        *[]ZiweiChartPatternsItem   `json:"patterns,omitempty" jsonschema:"命盘格局（如辅弼拱主/禄马交驰）"`
	SanFang         *[]ZiweiChartSanFangItem    `json:"san_fang,omitempty" jsonschema:"三方四正（命/财帛/官禄/迁移）"`
}

type ZiweiChartSchool struct {
	LeapMonth SchoolLeapMonthName `json:"leap_month" jsonschema:"枚举值：first_fifteen_days_current_month_next_fifteen_days_following_month。"`
	Source    SchoolSourceName    `json:"source" jsonschema:"来源。"`
}

type ZiweiChartGongWeiItem struct {
	Index      ZiweiPalaceName                    `json:"index" jsonschema:"宫名"`
	Name       ZiweiPalaceName                    `json:"name" jsonschema:"宫名"`
	Gan        LiuyaoGanName                      `json:"gan" jsonschema:"宫干"`
	Zhi        LiuyaoZhiName                      `json:"zhi" jsonschema:"宫支"`
	IsShenGong *bool                              `json:"is_shen_gong,omitempty" jsonschema:"是否身宫"`
	IsYuanGong *bool                              `json:"is_yuan_gong,omitempty" jsonschema:"是否来因宫"`
	XingYao    []ZiweiChartGongWeiItemXingYaoItem `json:"xing_yao" jsonschema:"星曜列表"`
	ZiWei      *string                            `json:"zi_wei,omitempty" jsonschema:"紫微星（本宫）"`
}

type ZiweiChartGongWeiItemXingYaoItem struct {
	Xing      string     `json:"xing" jsonschema:"星名"`
	Name      string     `json:"name" jsonschema:"全名"`
	IsZhuXing bool       `json:"is_zhu_xing" jsonschema:"是否主星"`
	SiHua     *SiHuaName `json:"si_hua,omitempty" jsonschema:"四化(禄/权/科/忌)"`
	LiangDu   *string    `json:"liang_du,omitempty" jsonschema:"亮度(庙/旺/利/平/陷)"`
}

type ZiweiChartPalaceFactsItem struct {
	Palace ZiweiPalaceName     `json:"palace" jsonschema:"宫位。"`
	Kind   PalaceFactsKindName `json:"kind" jsonschema:"类别。"`
	Target string              `json:"target" jsonschema:"target。"`
	Star   *string             `json:"star,omitempty" jsonschema:"star。"`
	Value  *string             `json:"value,omitempty" jsonschema:"取值。"`
}

type ZiweiChartKongGongItem struct {
	GongName ZiweiPalaceName `json:"gong_name" jsonschema:"gong_name。"`
	JieXing  []string        `json:"jie_xing" jsonschema:"jie_xing列表。"`
	DuiGong  ZiweiPalaceName `json:"dui_gong" jsonschema:"dui_gong。"`
}

type ZiweiChartPatternsItem struct {
	Name        string `json:"name" jsonschema:"格局名"`
	Description string `json:"description" jsonschema:"格局结构注解（非吉凶结论）"`
}

type ZiweiChartSanFangItem struct {
	Name    ZiweiPalaceName `json:"name" jsonschema:"名称。"`
	ZhuXing []string        `json:"zhu_xing" jsonschema:"主星。"`
	FuXing  []string        `json:"fu_xing" jsonschema:"辅星。"`
	SiHua   []SiHuaName     `json:"si_hua" jsonschema:"本宫全部四化，按星曜顺序输出"`
}

// ── ziwei.daxian input ──

type ZiweiDaXianInput struct {
	Chart ZiweiChartOutput `json:"chart" jsonschema:"ziwei.chart 返回的完整 chart 对象"`
}

// ── ziwei.daxian output ──

type ZiweiDaXianOutput []ZiweiDaXianItem

type ZiweiDaXianItem struct {
	StartYear *int            `json:"start_year,omitempty" jsonschema:"本限起始公历年"`
	EndYear   *int            `json:"end_year,omitempty" jsonschema:"本限结束公历年"`
	QiSui     int             `json:"qi_sui" jsonschema:"qi_sui。"`
	ZhiSui    int             `json:"zhi_sui" jsonschema:"zhi_sui。"`
	Gong      ZiweiPalaceName `json:"gong" jsonschema:"宫位。"`
	Name      ZiweiPalaceName `json:"name" jsonschema:"名称。"`
}

// ── ziwei.liunian input ──

type ZiweiLiuNianInput struct {
	LunarYear int              `json:"lunar_year" jsonschema:"农历年份（以正月初一为年界，2026：丙午年）"`
	Chart     ZiweiChartOutput `json:"chart" jsonschema:"ziwei.chart 返回的完整 chart 对象"`
}

// ── ziwei.liunian output ──

type ZiweiLiuNianOutput struct {
	MingGong     ZiweiPalaceName             `json:"ming_gong" jsonschema:"命宫。"`
	MingGongName *ZiweiPalaceName            `json:"ming_gong_name,omitempty" jsonschema:"ming_gong_name。"`
	Zhi          LiuyaoZhiName               `json:"zhi" jsonschema:"流年地支"`
	SiHua        map[string]SiHuaName        `json:"si_hua" jsonschema:"四化：{星名: 禄/权/科/忌}"`
	SiHuaGong    *map[string]ZiweiPalaceName `json:"si_hua_gong,omitempty" jsonschema:"四化落宫：{星名: 宫名}"`
	FuXing       *map[string]LiuyaoZhiName   `json:"fu_xing,omitempty" jsonschema:"流年辅星"`
	GongWei      *[]ZiweiLiuNianGongWeiItem  `json:"gong_wei,omitempty" jsonschema:"流年十二宫盘"`
}

type ZiweiLiuNianGongWeiItem struct {
	Zhi       LiuyaoZhiName   `json:"zhi" jsonschema:"地支"`
	Name      ZiweiPalaceName `json:"name" jsonschema:"宫名"`
	XingYao   *[]string       `json:"xing_yao,omitempty" jsonschema:"流耀"`
	IsLiuMing *bool           `json:"is_liu_ming,omitempty" jsonschema:"是否流年命宫。"`
}

// ── ziwei.liuyue input ──

type ZiweiLiuYueInput struct {
	TargetLunar ZiweiTargetLunarInput `json:"target_lunar" jsonschema:"target_lunar。"`
	Chart       ZiweiChartOutput      `json:"chart" jsonschema:"ziwei.chart 返回的完整 chart 对象"`
}

// ── ziwei.liuyue output ──

type ZiweiLiuYueOutput struct {
	TargetLunar    ZiweiLiuYueTargetLunar    `json:"target_lunar" jsonschema:"target_lunar。"`
	ResolvedPeriod ZiweiLiuYueResolvedPeriod `json:"resolved_period" jsonschema:"resolved_period。"`
	MingGong       ZiweiPalaceName           `json:"ming_gong" jsonschema:"命宫。"`
	MingGongName   *string                   `json:"ming_gong_name,omitempty" jsonschema:"ming_gong_name。"`
	Zhi            LiuyaoZhiName             `json:"zhi" jsonschema:"流月地支"`
	SiHua          map[string]SiHuaName      `json:"si_hua" jsonschema:"四化：{星名: 禄/权/科/忌}"`
	XingYao        *map[string]LiuyaoZhiName `json:"xing_yao,omitempty" jsonschema:"月星: {星名: zhiIdx}"`
	GongWei        *[]ZiweiLiuYueGongWeiItem `json:"gong_wei,omitempty" jsonschema:"流月十二宫盘"`
}

type ZiweiLiuYueTargetLunar struct {
	Year  int  `json:"year" jsonschema:"年。"`
	Month int  `json:"month" jsonschema:"月。"`
	Day   int  `json:"day" jsonschema:"日。"`
	Leap  bool `json:"leap" jsonschema:"是否闰月。"`
}

type ZiweiLiuYueResolvedPeriod struct {
	CalendarPeriod ZiweiLiuYueResolvedPeriodCalendarPeriod `json:"calendar_period" jsonschema:"calendar_period。"`
	FlowMonth      ZiweiLiuYueResolvedPeriodFlowMonth      `json:"flow_month" jsonschema:"flow_month。"`
}

type ZiweiLiuYueResolvedPeriodCalendarPeriod struct {
	Year     int                    `json:"year" jsonschema:"年。"`
	Month    int                    `json:"month" jsonschema:"月。"`
	Leap     bool                   `json:"leap" jsonschema:"是否闰月。"`
	Half     CalendarPeriodHalfName `json:"half" jsonschema:"枚举值：first、second、whole。"`
	DayStart int                    `json:"day_start" jsonschema:"day_start。"`
	DayEnd   int                    `json:"day_end" jsonschema:"day_end。"`
}

type ZiweiLiuYueResolvedPeriodFlowMonth struct {
	Year  int  `json:"year" jsonschema:"年。"`
	Month int  `json:"month" jsonschema:"月。"`
	Leap  bool `json:"leap" jsonschema:"是否闰月。"`
}

type ZiweiLiuYueGongWeiItem struct {
	Zhi       LiuyaoZhiName   `json:"zhi" jsonschema:"地支。"`
	Name      ZiweiPalaceName `json:"name" jsonschema:"名称。"`
	XingYao   []string        `json:"xing_yao" jsonschema:"星曜。"`
	IsLiuMing bool            `json:"is_liu_ming" jsonschema:"是否流年命宫。"`
}

// ── ziwei.liuri input ──

type ZiweiLiuRiInput struct {
	TargetLunar ZiweiTargetLunarInput `json:"target_lunar" jsonschema:"target_lunar。"`
	Chart       ZiweiChartOutput      `json:"chart" jsonschema:"ziwei.chart 返回的完整 chart 对象"`
}

// ── ziwei.liuri output ──

type ZiweiLiuRiOutput struct {
	MingGong     ZiweiPalaceName           `json:"ming_gong" jsonschema:"命宫。"`
	MingGongName *string                   `json:"ming_gong_name,omitempty" jsonschema:"ming_gong_name。"`
	Zhi          LiuyaoZhiName             `json:"zhi" jsonschema:"流日地支"`
	SiHua        map[string]SiHuaName      `json:"si_hua" jsonschema:"四化：{星名: 禄/权/科/忌}"`
	XingYao      *map[string]LiuyaoZhiName `json:"xing_yao,omitempty" jsonschema:"日星: {星名: zhiIdx}"`
	GongWei      *[]ZiweiLiuRiGongWeiItem  `json:"gong_wei,omitempty" jsonschema:"流日十二宫盘"`
}

type ZiweiLiuRiGongWeiItem struct {
	Zhi       LiuyaoZhiName   `json:"zhi" jsonschema:"地支。"`
	Name      ZiweiPalaceName `json:"name" jsonschema:"名称。"`
	XingYao   []string        `json:"xing_yao" jsonschema:"星曜。"`
	IsLiuMing bool            `json:"is_liu_ming" jsonschema:"是否流年命宫。"`
}

// ── ziwei.liushi input ──

type ZiweiLiuShiInput struct {
	TargetLunar ZiweiTargetLunarInput `json:"target_lunar" jsonschema:"target_lunar。"`
	ShiZhi      LiuyaoZhiName         `json:"shi_zhi" jsonschema:"时辰"`
	Chart       ZiweiChartOutput      `json:"chart" jsonschema:"ziwei.chart 返回的完整 chart 对象"`
}

// ── ziwei.liushi output ──

type ZiweiLiuShiOutput struct {
	MingGong     ZiweiPalaceName           `json:"ming_gong" jsonschema:"命宫。"`
	MingGongName *string                   `json:"ming_gong_name,omitempty" jsonschema:"ming_gong_name。"`
	Zhi          LiuyaoZhiName             `json:"zhi" jsonschema:"流时地支"`
	SiHua        map[string]SiHuaName      `json:"si_hua" jsonschema:"四化：{星名: 禄/权/科/忌}"`
	XingYao      *map[string]LiuyaoZhiName `json:"xing_yao,omitempty" jsonschema:"时星: {星名: zhiIdx}"`
	GongWei      *[]ZiweiLiuShiGongWeiItem `json:"gong_wei,omitempty" jsonschema:"流时十二宫盘"`
}

type ZiweiLiuShiGongWeiItem struct {
	Zhi       LiuyaoZhiName   `json:"zhi" jsonschema:"地支。"`
	Name      ZiweiPalaceName `json:"name" jsonschema:"名称。"`
	XingYao   []string        `json:"xing_yao" jsonschema:"星曜。"`
	IsLiuMing bool            `json:"is_liu_ming" jsonschema:"是否流年命宫。"`
}

// ── ziwei.bond input ──

type ZiweiBondSide struct {
	Chart ZiweiChartOutput `json:"chart" jsonschema:"紫微盘（ziwei.chart 返回的完整对象）"`
}

type ZiweiBondInput struct {
	A ZiweiBondSide `json:"a" jsonschema:"甲方：{ chart: 紫微盘 }，与 bazi.bond 同形"`
	B ZiweiBondSide `json:"b" jsonschema:"乙方：{ chart: 紫微盘 }，与 bazi.bond 同形"`
}

// ── ziwei.bond output ──

type ZiweiBondOutput struct {
	AMingGong string            `json:"a_ming_gong" jsonschema:"a_ming_gong。"`
	BMingGong string            `json:"b_ming_gong" jsonschema:"b_ming_gong。"`
	FuQiGong  ZiweiBondFuQiGong `json:"fu_qi_gong" jsonschema:"fu_qi_gong。"`
	ZiNvGong  ZiweiBondZiNvGong `json:"zi_nv_gong" jsonschema:"zi_nv_gong。"`
}

type ZiweiBondFuQiGong struct {
	AGongName     string   `json:"a_gong_name" jsonschema:"a_gong_name。"`
	BGongName     string   `json:"b_gong_name" jsonschema:"b_gong_name。"`
	AZhuXing      []string `json:"a_zhu_xing" jsonschema:"a_zhu_xing列表。"`
	BZhuXing      []string `json:"b_zhu_xing" jsonschema:"b_zhu_xing列表。"`
	ZhuXingStatus string   `json:"zhu_xing_status" jsonschema:"zhu_xing_status。"`
}

type ZiweiBondZiNvGong struct {
	AGongName     string   `json:"a_gong_name" jsonschema:"a_gong_name。"`
	BGongName     string   `json:"b_gong_name" jsonschema:"b_gong_name。"`
	AZhuXing      []string `json:"a_zhu_xing" jsonschema:"a_zhu_xing列表。"`
	BZhuXing      []string `json:"b_zhu_xing" jsonschema:"b_zhu_xing列表。"`
	ZhuXingStatus string   `json:"zhu_xing_status" jsonschema:"zhu_xing_status。"`
}

// ── ziwei.fullchart input ──

type ZiweiFullChartInput struct {
	Chart ZiweiChartOutput `json:"chart" jsonschema:"chart。"`
}

// ── ziwei.fullchart output ──

type ZiweiFullChartOutput struct {
	School          ZiweiFullChartSchool             `json:"school" jsonschema:"流派。"`
	GongWei         []ZiweiFullChartGongWeiItem      `json:"gong_wei" jsonschema:"12宫含杂曜/长生/博士/小限/将前/岁前"`
	MingGong        *ZiweiPalaceName                 `json:"ming_gong,omitempty" jsonschema:"命宫。"`
	SiHua           *map[string]SiHuaName            `json:"si_hua,omitempty" jsonschema:"四化：{星名: 禄/权/科/忌}"`
	PalaceFacts     *[]ZiweiFullChartPalaceFactsItem `json:"palace_facts,omitempty" jsonschema:"宫位原子事实；Python 因子层只读"`
	ShenGong        *ZiweiPalaceName                 `json:"shen_gong,omitempty" jsonschema:"身宫。"`
	JuShu           *JuShuName                       `json:"ju_shu,omitempty" jsonschema:"局数。"`
	MingZhu         *string                          `json:"ming_zhu,omitempty" jsonschema:"命主。"`
	ShenZhu         *string                          `json:"shen_zhu,omitempty" jsonschema:"身主。"`
	NianGan         *LiuyaoGanName                   `json:"nian_gan,omitempty" jsonschema:"年干"`
	NianZhi         *LiuyaoZhiName                   `json:"nian_zhi,omitempty" jsonschema:"年支"`
	ShiZhi          *LiuyaoZhiName                   `json:"shi_zhi,omitempty" jsonschema:"时支"`
	BirthYear       *int                             `json:"birth_year,omitempty" jsonschema:"出生公历年。"`
	BirthLunarMonth *int                             `json:"birth_lunar_month,omitempty" jsonschema:"出生农历月"`
	LunarMonth      *int                             `json:"lunar_month,omitempty" jsonschema:"排盘用农历月"`
	LunarDay        *int                             `json:"lunar_day,omitempty" jsonschema:"排盘用农历日"`
	JuShuName       *string                          `json:"ju_shu_name,omitempty" jsonschema:"ju_shu_name。"`
	Gender          *GenderName                      `json:"gender,omitempty" jsonschema:"性别。"`
	ZiweiPos        *ZiweiPalaceName                 `json:"ziwei_pos,omitempty" jsonschema:"紫微落宫。"`
	KongGong        *[]ZiweiFullChartKongGongItem    `json:"kong_gong,omitempty" jsonschema:"空宫借星（无主星宫位借对宫主星）"`
	Patterns        *[]ZiweiFullChartPatternsItem    `json:"patterns,omitempty" jsonschema:"命盘格局"`
	SanFang         *[]ZiweiFullChartSanFangItem     `json:"san_fang,omitempty" jsonschema:"三方四正（命/财帛/官禄/迁移）"`
}

type ZiweiFullChartSchool struct {
	LeapMonth SchoolLeapMonthName `json:"leap_month" jsonschema:"枚举值：first_fifteen_days_current_month_next_fifteen_days_following_month。"`
	Source    SchoolSourceName    `json:"source" jsonschema:"来源。"`
}

type ZiweiFullChartGongWeiItem struct {
	Index      ZiweiPalaceName                        `json:"index" jsonschema:"宫名"`
	Name       ZiweiPalaceName                        `json:"name" jsonschema:"宫名"`
	Gan        LiuyaoGanName                          `json:"gan" jsonschema:"宫干"`
	Zhi        LiuyaoZhiName                          `json:"zhi" jsonschema:"宫支"`
	IsShenGong *bool                                  `json:"is_shen_gong,omitempty" jsonschema:"是否身宫"`
	IsYuanGong *bool                                  `json:"is_yuan_gong,omitempty" jsonschema:"是否来因宫"`
	XingYao    []ZiweiFullChartGongWeiItemXingYaoItem `json:"xing_yao" jsonschema:"星曜列表"`
	ZiWei      *string                                `json:"zi_wei,omitempty" jsonschema:"紫微星（本宫）"`
	Ages       *[]int                                 `json:"ages,omitempty" jsonschema:"小限年龄"`
	ChangSheng *ChangShengStageName                   `json:"chang_sheng,omitempty" jsonschema:"长生"`
	BoShi      *string                                `json:"bo_shi,omitempty" jsonschema:"博士"`
	JiangQian  *string                                `json:"jiang_qian,omitempty" jsonschema:"将前"`
	SuiQian    *string                                `json:"sui_qian,omitempty" jsonschema:"岁前"`
	ZaYao      *[]string                              `json:"za_yao,omitempty" jsonschema:"杂曜"`
}

type ZiweiFullChartGongWeiItemXingYaoItem struct {
	IsZhuXing *bool              `json:"is_zhu_xing" jsonschema:"是否主星。"`
	LiangDu   *string            `json:"liang_du,omitempty" jsonschema:"亮度。"`
	Name      *string            `json:"name" jsonschema:"名称。"`
	SiHua     *SiHuaNameNullable `json:"si_hua,omitempty" jsonschema:"四化。"`
	Xing      *string            `json:"xing" jsonschema:"xing。"`
}

type ZiweiFullChartPalaceFactsItem struct {
	Palace ZiweiPalaceName `json:"palace" jsonschema:"宫位。"`
	Kind   string          `json:"kind" jsonschema:"类别。"`
	Target string          `json:"target" jsonschema:"target。"`
	Star   *string         `json:"star,omitempty" jsonschema:"star。"`
	Value  *string         `json:"value,omitempty" jsonschema:"取值。"`
}

type ZiweiFullChartKongGongItem struct {
	GongName ZiweiPalaceName `json:"gong_name" jsonschema:"gong_name。"`
	JieXing  []string        `json:"jie_xing" jsonschema:"jie_xing列表。"`
	DuiGong  ZiweiPalaceName `json:"dui_gong" jsonschema:"dui_gong。"`
}

type ZiweiFullChartPatternsItem struct {
	Description *string `json:"description" jsonschema:"说明。"`
	Name        *string `json:"name" jsonschema:"名称。"`
}

type ZiweiFullChartSanFangItem struct {
	Name    ZiweiPalaceName `json:"name" jsonschema:"名称。"`
	ZhuXing []string        `json:"zhu_xing" jsonschema:"主星。"`
	FuXing  []string        `json:"fu_xing" jsonschema:"辅星。"`
	SiHua   []SiHuaName     `json:"si_hua" jsonschema:"本宫全部四化，按星曜顺序输出"`
}
