// Wire DTOs for the migrated bazi/ziwei/qimen engine tools.
//
// The shapes follow the legacy raw-JSON handlers and the captured engine
// payloads under testdata/actual/. The file started as a generated
// projection; it is ordinary source now, so edit it directly and keep
// TestEngineContractWireFidelity green when a shape changes.

package mcpcontract

import "encoding/json"

// ── qimen.chart input ──

// QimenMethodInput is the closed method fact consumed by qimen.chart. Its
// published schema is narrowed by qimenChartFormsSchema to the catalog's valid
// combinations; the domain resolver remains the final authority.
type QimenMethodInput struct {
	Scope            MethodScopeName        `json:"scope" jsonschema:"时间层级；与 school 及定局/刻家参数共同构成一个方法分支。"`
	School           MethodSchoolName       `json:"school" jsonschema:"盘式流派；与 scope 组合决定可用参数。"`
	DingjuMethod     MethodDingjuMethodName `json:"dingju_method,omitempty" jsonschema:"定局法；仅 hour/day 分支携带。"`
	QuarterRule      MethodQuarterRuleName  `json:"quarter_rule,omitempty" jsonschema:"刻家规则；仅 quarter 分支携带。"`
	BaseDingjuMethod MethodDingjuMethodName `json:"base_dingju_method,omitempty" jsonschema:"基准时家定局法；仅十二分钟十分局分支携带。"`
	DunSource        DunSourceName          `json:"dun_source,omitempty" jsonschema:"分遁口径；仅十二分钟十分局分支携带。"`
	HourBoundary     HourBoundaryName       `json:"hour_boundary,omitempty" jsonschema:"时辰界；仅十二分钟十分局分支携带。"`
}

// QimenChartInput takes the canonical {instant, timezone} pair, longitude, and
// a closed method object. Time/location and method selection stay orthogonal;
// yong_shen/birth_date are chart-analysis inputs, not method knobs.
type QimenChartInput struct {
	Instant   Instant              `json:"instant" jsonschema:"占问瞬间 RFC3339；接受 UTC Z 或 numeric offset，服务端归一化为 UTC；节气换局/置闰按此时刻判定"`
	Timezone  string               `json:"timezone" jsonschema:"占问地 IANA 时区"`
	Longitude float64              `json:"longitude" jsonschema:"占问地经度（-180..180）；真太阳时校正用于时柱与分钟分法"`
	Method    QimenMethodInput     `json:"method" jsonschema:"奇门方法事实；scope、盘式流派和条件参数按目录组合传入。"`
	YongShen  QimenYongShenSymbols `json:"yong_shen,omitempty" jsonschema:"用神符号组合；甲不露盘，不能作为独立符号"`
	BirthDate string               `json:"birth_date,omitempty" jsonschema:"出生日期 YYYY-MM-DD 或 RFC3339；按精确立春取年命干，跨立春日必须给时刻"`
}

type qimenHourZhuanpanMethodForm struct {
	Scope        MethodScopeName        `json:"scope" jsonschema:"时间层级。"`
	School       MethodSchoolName       `json:"school" jsonschema:"盘式流派。"`
	DingjuMethod MethodDingjuMethodName `json:"dingju_method" jsonschema:"定局法。"`
}

type qimenHourLuoshuMethodForm struct {
	Scope        MethodScopeName        `json:"scope" jsonschema:"时间层级。"`
	School       MethodSchoolName       `json:"school" jsonschema:"盘式流派。"`
	DingjuMethod MethodDingjuMethodName `json:"dingju_method" jsonschema:"定局法。"`
}

type qimenHourMingfaMethodForm struct {
	Scope  MethodScopeName  `json:"scope" jsonschema:"时间层级。"`
	School MethodSchoolName `json:"school" jsonschema:"盘式流派。"`
}

type qimenDayZhuanpanMethodForm struct {
	Scope        MethodScopeName        `json:"scope" jsonschema:"时间层级。"`
	School       MethodSchoolName       `json:"school" jsonschema:"盘式流派。"`
	DingjuMethod MethodDingjuMethodName `json:"dingju_method" jsonschema:"定局法。"`
}

type qimenDayLuoshuMethodForm struct {
	Scope        MethodScopeName        `json:"scope" jsonschema:"时间层级。"`
	School       MethodSchoolName       `json:"school" jsonschema:"盘式流派。"`
	DingjuMethod MethodDingjuMethodName `json:"dingju_method" jsonschema:"定局法。"`
}

type qimenPeriodMethodForm struct {
	Scope  MethodScopeName  `json:"scope" jsonschema:"时间层级。"`
	School MethodSchoolName `json:"school" jsonschema:"盘式流派。"`
}

type qimenQuarterTenMethodForm struct {
	Scope       MethodScopeName       `json:"scope" jsonschema:"时间层级。"`
	School      MethodSchoolName      `json:"school" jsonschema:"盘式流派。"`
	QuarterRule MethodQuarterRuleName `json:"quarter_rule" jsonschema:"刻家规则。"`
}

type qimenQuarterTwelveMethodForm struct {
	Scope            MethodScopeName        `json:"scope" jsonschema:"时间层级。"`
	School           MethodSchoolName       `json:"school" jsonschema:"盘式流派。"`
	QuarterRule      MethodQuarterRuleName  `json:"quarter_rule" jsonschema:"刻家规则。"`
	BaseDingjuMethod MethodDingjuMethodName `json:"base_dingju_method" jsonschema:"基准时家定局法。"`
	DunSource        DunSourceName          `json:"dun_source" jsonschema:"分遁口径。"`
	HourBoundary     HourBoundaryName       `json:"hour_boundary" jsonschema:"时辰界。"`
}

type qimenJinhanMethodForm struct {
	Scope  MethodScopeName  `json:"scope" jsonschema:"时间层级。"`
	School MethodSchoolName `json:"school" jsonschema:"盘式流派。"`
}

type qimenChartStandardForm struct {
	Instant   Instant              `json:"instant" jsonschema:"占问瞬间。"`
	Timezone  string               `json:"timezone" jsonschema:"占问地 IANA 时区。"`
	Longitude float64              `json:"longitude" jsonschema:"占问地经度。"`
	Method    QimenMethodInput     `json:"method" jsonschema:"非金函玉镜奇门方法事实。"`
	YongShen  QimenYongShenSymbols `json:"yong_shen,omitempty" jsonschema:"用神符号组合。"`
	BirthDate string               `json:"birth_date,omitempty" jsonschema:"出生日期或时刻。"`
}

type qimenChartJinhanForm struct {
	Instant   Instant               `json:"instant" jsonschema:"占问瞬间。"`
	Timezone  string                `json:"timezone" jsonschema:"占问地 IANA 时区。"`
	Longitude float64               `json:"longitude" jsonschema:"占问地经度。"`
	Method    qimenJinhanMethodForm `json:"method" jsonschema:"金函玉镜方法事实。"`
}

func qimenMethodFormSchema[T any](
	constraints map[string]string, enums map[string][]string,
) (json.RawMessage, error) {
	_, raw, err := inputSchemaFor[T]()
	if err != nil {
		return nil, err
	}
	for field, value := range constraints {
		raw, err = withSchemaConst(raw, field, value)
		if err != nil {
			return nil, err
		}
	}
	for field, values := range enums {
		raw, err = withSchemaEnum(raw, field, values...)
		if err != nil {
			return nil, err
		}
	}
	return raw, nil
}

func qimenMethodFormsSchema() (json.RawMessage, error) {
	builders := []func() (json.RawMessage, error){
		func() (json.RawMessage, error) {
			return qimenMethodFormSchema[qimenHourZhuanpanMethodForm](
				map[string]string{"scope": "hour", "school": "zhuanpan"},
				map[string][]string{"dingju_method": {"chaibu", "zhirun", "maoshan"}},
			)
		},
		func() (json.RawMessage, error) {
			return qimenMethodFormSchema[qimenHourLuoshuMethodForm](
				map[string]string{"scope": "hour", "school": "luoshu_feipan"},
				map[string][]string{"dingju_method": {"chaibu", "zhirun"}},
			)
		},
		func() (json.RawMessage, error) {
			return qimenMethodFormSchema[qimenHourMingfaMethodForm](
				map[string]string{"scope": "hour", "school": "mingfa_feipan"}, nil,
			)
		},
		func() (json.RawMessage, error) {
			return qimenMethodFormSchema[qimenDayZhuanpanMethodForm](
				map[string]string{"scope": "day", "school": "zhuanpan"},
				map[string][]string{"dingju_method": {"chaibu", "zhirun"}},
			)
		},
		func() (json.RawMessage, error) {
			return qimenMethodFormSchema[qimenDayLuoshuMethodForm](
				map[string]string{"scope": "day", "school": "luoshu_feipan"},
				map[string][]string{"dingju_method": {"chaibu", "zhirun"}},
			)
		},
		func() (json.RawMessage, error) {
			return qimenMethodFormSchema[qimenPeriodMethodForm](
				map[string]string{"scope": "month"},
				map[string][]string{"school": {"zhuanpan", "luoshu_feipan"}},
			)
		},
		func() (json.RawMessage, error) {
			return qimenMethodFormSchema[qimenPeriodMethodForm](
				map[string]string{"scope": "year"},
				map[string][]string{"school": {"zhuanpan", "luoshu_feipan"}},
			)
		},
		func() (json.RawMessage, error) {
			return qimenMethodFormSchema[qimenQuarterTenMethodForm](
				map[string]string{
					"scope": "quarter", "school": "zhuanpan",
					"quarter_rule": "ten_minute_sanyuan",
				}, nil,
			)
		},
		func() (json.RawMessage, error) {
			return qimenMethodFormSchema[qimenQuarterTwelveMethodForm](
				map[string]string{
					"scope": "quarter", "school": "zhuanpan",
					"quarter_rule": "twelve_minute_ten_division",
				},
				map[string][]string{"base_dingju_method": {"chaibu", "zhirun", "maoshan"}},
			)
		},
	}
	forms := make([]json.RawMessage, 0, len(builders))
	for _, build := range builders {
		raw, err := build()
		if err != nil {
			return nil, err
		}
		forms = append(forms, raw)
	}
	return json.Marshal(map[string]any{
		"type": "object", "oneOf": forms,
	})
}

func qimenJinhanMethodSchema() (json.RawMessage, error) {
	return qimenMethodFormSchema[qimenJinhanMethodForm](
		map[string]string{"scope": "day", "school": "jinhan_yujing"}, nil,
	)
}

func qimenChartFormsSchema() (json.RawMessage, error) {
	methodForms, err := qimenMethodFormsSchema()
	if err != nil {
		return nil, err
	}
	_, standard, err := inputSchemaFor[qimenChartStandardForm]()
	if err != nil {
		return nil, err
	}
	standard, err = withSchemaProperty(standard, "method", methodForms)
	if err != nil {
		return nil, err
	}
	standard, err = withSchemaNumberBounds(standard, "longitude", -180, 180)
	if err != nil {
		return nil, err
	}

	jinhanMethod, err := qimenJinhanMethodSchema()
	if err != nil {
		return nil, err
	}
	_, jinhan, err := inputSchemaFor[qimenChartJinhanForm]()
	if err != nil {
		return nil, err
	}
	jinhan, err = withSchemaProperty(jinhan, "method", jinhanMethod)
	if err != nil {
		return nil, err
	}
	jinhan, err = withSchemaNumberBounds(jinhan, "longitude", -180, 180)
	if err != nil {
		return nil, err
	}
	return json.Marshal(map[string]any{
		"type":  "object",
		"oneOf": []json.RawMessage{standard, jinhan},
	})
}

// ── qimen.chart output ──

type QimenChartOutput struct {
	Method              QimenChartMethod                   `json:"method" jsonschema:"排盘方法元数据"`
	Board               QimenChartBoard                    `json:"board" jsonschema:"盘面。"`
	Patterns            *[]QimenChartPatternsItem          `json:"patterns,omitempty" jsonschema:"格局。"`
	GanInteraction      *[]QimenChartGanInteractionItem    `json:"gan_interaction,omitempty" jsonschema:"gan_interaction列表。"`
	MenInteraction      *[]QimenChartMenInteractionItem    `json:"men_interaction,omitempty" jsonschema:"men_interaction列表。"`
	XingInteraction     *[]QimenChartXingInteractionItem   `json:"xing_interaction,omitempty" jsonschema:"xing_interaction列表。"`
	XingGongWuXing      *[]QimenChartXingGongWuXingItem    `json:"xing_gong_wu_xing,omitempty" jsonschema:"星与宫的五行关系；不是完整季节旺衰"`
	MenPo               *[]QimenGongName                   `json:"men_po,omitempty" jsonschema:"men_po列表。"`
	MenZhi              *[]QimenGongName                   `json:"men_zhi,omitempty" jsonschema:"men_zhi列表。"`
	YingQi              *QimenChartYingQi                  `json:"ying_qi,omitempty" jsonschema:"ying_qi。"`
	RiGanGong           *QimenGongName                     `json:"ri_gan_gong,omitempty" jsonschema:"枚举值。"`
	ShiGanGong          *QimenGongName                     `json:"shi_gan_gong,omitempty" jsonschema:"枚举值。"`
	RiShiRelation       *QimenChartRiShiRelation           `json:"ri_shi_relation,omitempty" jsonschema:"ri_shi_relation。"`
	KongWangAffected    *[]QimenChartKongWangAffectedItem  `json:"kong_wang_affected,omitempty" jsonschema:"kong_wang_affected列表。"`
	MaXingAffected      *[]QimenChartMaXingAffectedItem    `json:"ma_xing_affected,omitempty" jsonschema:"ma_xing_affected列表。"`
	PalaceWangShuai     *[]QimenChartPalaceWangShuaiItem   `json:"palace_wang_shuai,omitempty" jsonschema:"palace_wang_shuai列表。"`
	ShiGanGongWangShuai *QimenChartShiGanGongWangShuai     `json:"shi_gan_gong_wang_shuai,omitempty" jsonschema:"shi_gan_gong_wang_shuai。"`
	ZhiFuXingGong       *QimenGongName                     `json:"zhi_fu_xing_gong,omitempty" jsonschema:"枚举值。"`
	ZhiShiMenGong       *QimenGongName                     `json:"zhi_shi_men_gong,omitempty" jsonschema:"枚举值。"`
	YongShen            *QimenChartYongShen                `json:"yong_shen,omitempty" jsonschema:"yong_shen。"`
	Specialized         *QimenChartSpecialized             `json:"specialized,omitempty" jsonschema:"专占领域原子事实；Python 只做 projection"`
	RiGanPalaceFacts    *[]QimenChartRiGanPalaceFactsItem  `json:"ri_gan_palace_facts,omitempty" jsonschema:"ri_gan_palace_facts列表。"`
	ShiGanPalaceFacts   *[]QimenChartShiGanPalaceFactsItem `json:"shi_gan_palace_facts,omitempty" jsonschema:"shi_gan_palace_facts列表。"`
}

type QimenChartMethod struct {
	Scope                MethodScopeName              `json:"scope" jsonschema:"时间层级"`
	ScopeName            MethodScopeName2             `json:"scope_name" jsonschema:"时间层级名称"`
	School               MethodSchoolName             `json:"school" jsonschema:"盘式流派"`
	SchoolName           MethodSchoolName2            `json:"school_name" jsonschema:"盘式流派名称"`
	DingjuMethod         *MethodDingjuMethodName      `json:"dingju_method,omitempty" jsonschema:"定局法；quarter 输出 quarter_rule，不输出本字段"`
	DingjuMethodName     *MethodDingjuMethodName2     `json:"dingju_method_name,omitempty" jsonschema:"定局法名称"`
	QuarterRule          *MethodQuarterRuleName       `json:"quarter_rule,omitempty" jsonschema:"刻家规则；仅 scope：quarter 输出"`
	QuarterRuleName      *MethodQuarterRuleName2      `json:"quarter_rule_name,omitempty" jsonschema:"刻家规则名称"`
	BaseDingjuMethod     *MethodDingjuMethodName      `json:"base_dingju_method,omitempty" jsonschema:"十二分钟十分局的基准时家定局法"`
	BaseDingjuMethodName *MethodDingjuMethodName2     `json:"base_dingju_method_name,omitempty" jsonschema:"基准时家定局法名称"`
	MethodSource         MethodMethodSourceName       `json:"method_source" jsonschema:"方法来源：节气定局、刻家规则、月周期或年周期"`
	DayBoundary          MethodDayBoundaryName        `json:"day_boundary" jsonschema:"枚举值：late_zi_rolls_day_pillar。"`
	SpiritMode           MethodSpiritModeName         `json:"spirit_mode" jsonschema:"枚举值：day_gan_twelve_spirits、eight_spirits_yang_gouchen_zhuque_yin_baihu_xuanwu、nine_spirits_baihu_taichang_xuanwu、nine_spirits_gouchen_taichang_zhuque。"`
	Geometry             MethodGeometryName           `json:"geometry" jsonschema:"枚举值：luoshu_ring_rotation、mingfa_star_door_forward_flight、palace_number_shift、solstice_day_star_flight。"`
	StarMode             MethodStarModeName           `json:"star_mode" jsonschema:"枚举值：nine_stars_forward_with_center、star_gan_shift_together、taiyi_nine_stars_day_flight、tian_qin_lodges_kun。"`
	DoorMode             MethodDoorModeName           `json:"door_mode" jsonschema:"枚举值：center_projects_kun、eight_doors_real_center_step、nine_doors_with_center_forward、three_day_eight_door_rotation。"`
	SpiritFlight         MethodSpiritFlightName       `json:"spirit_flight" jsonschema:"枚举值：solstice_yang_forward_yin_reverse、yang_forward_yin_reverse。"`
	TianQin              *QimenChartMethodTianQin     `json:"tian_qin,omitempty" jsonschema:"排盘方法元数据 中的 tian_qin。"`
	JieQi                MethodJieQiName              `json:"jie_qi" jsonschema:"枚举值：冬至、处暑、夏至、大寒、大暑、大雪、寒露、小寒、小暑、小满、小雪、惊蛰、春分、清明、白露、秋分、立冬、立夏、立春、立秋、芒种、谷雨、雨水、霜降。"`
	YongJuJieQi          *string                      `json:"yong_ju_jie_qi,omitempty" jsonschema:"排盘方法元数据 中的 yong_ju_jie_qi。"`
	Yuan                 *MethodYuanName              `json:"yuan,omitempty" jsonschema:"枚举值：上元、下元、中元。"`
	ZhiRunState          *MethodZhiRunStateName       `json:"zhi_run_state,omitempty" jsonschema:"枚举值：接气、正授、置闰、超神。"`
	LeadPillar           *QimenChartMethodLeadPillar  `json:"lead_pillar,omitempty" jsonschema:"排盘方法元数据 中的 lead_pillar。"`
	LeadXunShou          *QimenChartMethodLeadXunShou `json:"lead_xun_shou,omitempty" jsonschema:"排盘方法元数据 中的 lead_xun_shou。"`
	LeadXunShouGong      *QimenGongName               `json:"lead_xun_shou_gong,omitempty" jsonschema:"排盘方法元数据。"`
	DayFuTou             *QimenChartMethodDayFuTou    `json:"day_fu_tou,omitempty" jsonschema:"排盘方法元数据 中的 day_fu_tou。"`
	Quarter              *QimenChartMethodQuarter     `json:"quarter,omitempty" jsonschema:"刻家序数与主柱口径；干支不重复输出，以 lead_pillar 为准"`
	DunSource            *MethodDunSourceName         `json:"dun_source,omitempty" jsonschema:"枚举值：winter_solstice_yang_summer_solstice_yin。"`
	Basis                *string                      `json:"basis,omitempty" jsonschema:"依据。"`
}

type QimenChartMethodTianQin struct {
	Star                  TianQinStarName    `json:"star" jsonschema:"枚举值：天禽。"`
	HomeGong              QimenGongName      `json:"home_gong" jsonschema:"枚举值。"`
	LodgingGong           QimenEightGongName `json:"lodging_gong" jsonschema:"枚举值。"`
	Follows               TianQinFollowsName `json:"follows" jsonschema:"枚举值：天芮。"`
	CarriesCenterEarthGan bool               `json:"carries_center_earth_gan" jsonschema:"carries_center_earth_gan。"`
	LodgingRule           string             `json:"lodging_rule" jsonschema:"lodging_rule。"`
}

type QimenChartMethodLeadPillar struct {
	Name string        `json:"name" jsonschema:"名称。"`
	Gan  LiuyaoGanName `json:"gan" jsonschema:"天干。"`
	Zhi  LiuyaoZhiName `json:"zhi" jsonschema:"地支。"`
}

type QimenChartMethodLeadXunShou struct {
	Name  LeadXunShouName    `json:"name" jsonschema:"名称。"`
	Gan   LeadXunShouGanName `json:"gan" jsonschema:"天干。"`
	Zhi   LiuyaoZhiName      `json:"zhi" jsonschema:"地支。"`
	LiuYi LiuYiName          `json:"liu_yi" jsonschema:"枚举值：壬、己、庚、戊、癸、辛。"`
}

type QimenChartMethodDayFuTou struct {
	Name DayFuTouName    `json:"name" jsonschema:"名称。"`
	Gan  DayFuTouGanName `json:"gan" jsonschema:"天干。"`
	Zhi  LiuyaoZhiName   `json:"zhi" jsonschema:"地支。"`
}

type QimenChartMethodQuarter struct {
	Index             int               `json:"index" jsonschema:"序号。"`
	MinutesPerQuarter int               `json:"minutes_per_quarter" jsonschema:"刻家序数与主柱口径；干支不重复输出，以 lead_pillar 为准 中的 minutes_per_quarter。"`
	LeadPillarMode    MethodScopeName   `json:"lead_pillar_mode" jsonschema:"quarter：刻柱主盘；hour：十二分钟十分局刻家的时柱主盘"`
	DunSource         *DunSourceName    `json:"dun_source,omitempty" jsonschema:"hour_branch：时支分遁；solar_term：沿基准时家节气分遁"`
	HourBoundary      *HourBoundaryName `json:"hour_boundary,omitempty" jsonschema:"late_zi：23点起子时；zi_zheng：0点起子时"`
}

type QimenChartBoard struct {
	Jushu      *int                             `json:"jushu,omitempty" jsonschema:"局数。"`
	YinDun     bool                             `json:"yin_dun" jsonschema:"是否阴遁。"`
	RiGan      LiuyaoGanName                    `json:"ri_gan" jsonschema:"日干。"`
	RiZhi      LiuyaoZhiName                    `json:"ri_zhi" jsonschema:"日支。"`
	NianGan    *LiuyaoGanName                   `json:"nian_gan,omitempty" jsonschema:"年干。"`
	NianZhi    *LiuyaoZhiName                   `json:"nian_zhi,omitempty" jsonschema:"年支。"`
	YueGan     *LiuyaoGanName                   `json:"yue_gan,omitempty" jsonschema:"月干。"`
	YueZhi     *LiuyaoZhiName                   `json:"yue_zhi,omitempty" jsonschema:"月支。"`
	ZhiFuXing  *QimenStarName                   `json:"zhi_fu_xing,omitempty" jsonschema:"值符星。"`
	ZhiShiMen  *QimenDoorName                   `json:"zhi_shi_men,omitempty" jsonschema:"值使门。"`
	GongWei    []QimenChartBoardGongWeiItem     `json:"gong_wei" jsonschema:"十二宫。"`
	ShiGan     *LiuyaoGanName                   `json:"shi_gan,omitempty" jsonschema:"时干。"`
	ShiZhi     *LiuyaoZhiName                   `json:"shi_zhi,omitempty" jsonschema:"时支。"`
	MaXing     *QimenChartBoardMaXing           `json:"ma_xing,omitempty" jsonschema:"马星。"`
	KongWang   *[]QimenChartBoardKongWangItem   `json:"kong_wang,omitempty" jsonschema:"空亡。"`
	WuBuYuShi  *bool                            `json:"wu_bu_yu_shi,omitempty" jsonschema:"五不遇时。"`
	DaySpirits *[]QimenChartBoardDaySpiritsItem `json:"day_spirits,omitempty" jsonschema:"day_spirits列表。"`
}

type QimenChartBoardGongWeiItem struct {
	Gong        QimenChartBoardGongWeiItemGong           `json:"gong" jsonschema:"宫位。"`
	DiPanGan    *DiPanGanName                            `json:"di_pan_gan,omitempty" jsonschema:"地盘干。"`
	AnGan       *DiPanGanName                            `json:"an_gan,omitempty" jsonschema:"暗干。"`
	TianPanGan  *DiPanGanName                            `json:"tian_pan_gan,omitempty" jsonschema:"天盘干。"`
	TianPan     *[]QimenChartBoardGongWeiItemTianPanItem `json:"tian_pan,omitempty" jsonschema:"天盘。"`
	Men         *QimenDoorName                           `json:"men,omitempty" jsonschema:"八门。"`
	MenPresent  bool                                     `json:"men_present" jsonschema:"门存在集合。"`
	Shen        *QimenSpiritName                         `json:"shen,omitempty" jsonschema:"八神。"`
	ShenPresent *bool                                    `json:"shen_present,omitempty" jsonschema:"神存在集合。"`
	AnGanZhi    *string                                  `json:"an_gan_zhi,omitempty" jsonschema:"鸣法暗干支；仅鸣法飞盘输出，其他方法省略"`
	Xing        *XingName                                `json:"xing,omitempty" jsonschema:"枚举值：咸池、天乙、天符、太乙、太阴、招摇、摄提、轩辕、青龙。"`
}

type QimenChartBoardGongWeiItemGong struct {
	Name      QimenGongName `json:"name" jsonschema:"名称。"`
	Luoshu    int           `json:"luoshu" jsonschema:"luoshu。"`
	RingIndex *int          `json:"ring_index,omitempty" jsonschema:"ring_index。"`
}

type QimenChartBoardGongWeiItemTianPanItem struct {
	Gan  DiPanGanName  `json:"gan" jsonschema:"天干。"`
	Xing QimenStarName `json:"xing" jsonschema:"枚举值。"`
}

type QimenChartBoardMaXing struct {
	Branch LiuyaoZhiName      `json:"branch" jsonschema:"地支。"`
	Gong   QimenEightGongName `json:"gong" jsonschema:"宫位。"`
}

type QimenChartBoardKongWangItem struct {
	Branch LiuyaoZhiName      `json:"branch" jsonschema:"地支。"`
	Gong   QimenEightGongName `json:"gong" jsonschema:"宫位。"`
}

type QimenChartBoardDaySpiritsItem struct {
	Zhi  LiuyaoZhiName `json:"zhi" jsonschema:"地支。"`
	Shen string        `json:"shen" jsonschema:"八神。"`
}

type QimenChartPatternsItem struct {
	Name        PatternsName    `json:"name" jsonschema:"名称。"`
	Description string          `json:"description" jsonschema:"说明。"`
	Basis       string          `json:"basis" jsonschema:"依据。"`
	Auspicious  bool            `json:"auspicious" jsonschema:"auspicious。"`
	GongWei     []QimenGongName `json:"gong_wei" jsonschema:"十二宫。"`
}

type QimenChartGanInteractionItem struct {
	Gong       QimenGongName `json:"gong" jsonschema:"宫位。"`
	DiPanGan   DiPanGanName  `json:"di_pan_gan" jsonschema:"地盘干。"`
	TianPanGan DiPanGanName  `json:"tian_pan_gan" jsonschema:"天盘干。"`
	Name       string        `json:"name" jsonschema:"名称。"`
	Meaning    string        `json:"meaning" jsonschema:"meaning。"`
	Auspicious bool          `json:"auspicious" jsonschema:"auspicious。"`
}

type QimenChartMenInteractionItem struct {
	Door    QimenDoorName `json:"door" jsonschema:"枚举值。"`
	Gong    QimenGongName `json:"gong" jsonschema:"宫位。"`
	Name    string        `json:"name" jsonschema:"名称。"`
	Meaning string        `json:"meaning" jsonschema:"meaning。"`
}

type QimenChartXingInteractionItem struct {
	Xing       QimenStarName `json:"xing" jsonschema:"枚举值。"`
	Gong       QimenGongName `json:"gong" jsonschema:"宫位。"`
	Name       string        `json:"name" jsonschema:"名称。"`
	Meaning    string        `json:"meaning" jsonschema:"meaning。"`
	Auspicious bool          `json:"auspicious" jsonschema:"auspicious。"`
}

type QimenChartXingGongWuXingItem struct {
	Xing             QimenStarName              `json:"xing" jsonschema:"枚举值。"`
	Gong             QimenGongName              `json:"gong" jsonschema:"宫位。"`
	Relation         XingGongWuXingRelationName `json:"relation" jsonschema:"关系。"`
	RelationName     string                     `json:"relation_name" jsonschema:"relation_name。"`
	TraditionalLabel string                     `json:"traditional_label" jsonschema:"traditional_label。"`
}

type QimenChartYingQi struct {
	Candidates []QimenChartYingQiCandidatesItem `json:"candidates" jsonschema:"candidates列表。"`
	Summary    string                           `json:"summary" jsonschema:"summary。"`
}

type QimenChartYingQiCandidatesItem struct {
	Type        CandidatesTypeName                            `json:"type" jsonschema:"类型。"`
	Mechanism   string                                        `json:"mechanism" jsonschema:"mechanism。"`
	Branch      LiuyaoZhiName                                 `json:"branch" jsonschema:"地支。"`
	Gong        QimenEightGongName                            `json:"gong" jsonschema:"宫位。"`
	RelatedTo   []QimenChartYingQiCandidatesItemRelatedToItem `json:"related_to" jsonschema:"related_to列表。"`
	Description string                                        `json:"description" jsonschema:"说明。"`
	Dates       []QimenChartYingQiCandidatesItemDatesItem     `json:"dates" jsonschema:"dates列表。"`
}

type QimenChartYingQiCandidatesItemRelatedToItem struct {
	Symbol string            `json:"symbol" jsonschema:"符号。"`
	Role   RelatedToRoleName `json:"role" jsonschema:"角色。"`
}

type QimenChartYingQiCandidatesItemDatesItem struct {
	Date   DateString     `json:"date" jsonschema:"date。"`
	Branch LiuyaoZhiName  `json:"branch" jsonschema:"地支。"`
	Match  DatesMatchName `json:"match" jsonschema:"枚举值：opposite、same。"`
	Reason string         `json:"reason" jsonschema:"reason。"`
}

type QimenChartRiShiRelation struct {
	Subject  RiShiRelationSubjectName  `json:"subject" jsonschema:"主体。"`
	Object   RiShiRelationObjectName   `json:"object" jsonschema:"客体。"`
	Relation RiShiRelationRelationName `json:"relation" jsonschema:"关系。"`
	Name     string                    `json:"name" jsonschema:"名称。"`
}

type QimenChartKongWangAffectedItem struct {
	Symbol string             `json:"symbol" jsonschema:"符号。"`
	Branch LiuyaoZhiName      `json:"branch" jsonschema:"地支。"`
	Gong   QimenEightGongName `json:"gong" jsonschema:"宫位。"`
	Role   RelatedToRoleName  `json:"role" jsonschema:"角色。"`
}

type QimenChartMaXingAffectedItem struct {
	Symbol *string             `json:"symbol,omitempty" jsonschema:"符号。"`
	Branch *LiuyaoZhiName      `json:"branch,omitempty" jsonschema:"地支。"`
	Gong   *QimenEightGongName `json:"gong,omitempty" jsonschema:"宫位。"`
	Role   *RelatedToRoleName  `json:"role,omitempty" jsonschema:"角色。"`
}

type QimenChartPalaceWangShuaiItem struct {
	Gong          QimenGongName   `json:"gong" jsonschema:"宫位。"`
	Wuxing        ElementName     `json:"wuxing" jsonschema:"五行。"`
	WangShuai     LiuyaoWangShuai `json:"wang_shuai" jsonschema:"枚举值。"`
	WangShuaiName LiuyaoWangShuai `json:"wang_shuai_name" jsonschema:"枚举值。"`
}

type QimenChartShiGanGongWangShuai struct {
	Gong          QimenGongName   `json:"gong" jsonschema:"宫位。"`
	Wuxing        ElementName     `json:"wuxing" jsonschema:"五行。"`
	WangShuai     LiuyaoWangShuai `json:"wang_shuai" jsonschema:"枚举值。"`
	WangShuaiName LiuyaoWangShuai `json:"wang_shuai_name" jsonschema:"枚举值。"`
}

type QimenChartYongShen struct {
	NianGanGong        *QimenGongName                              `json:"nian_gan_gong,omitempty" jsonschema:"枚举值。"`
	Symbols            []QimenChartYongShenSymbolsItem             `json:"symbols" jsonschema:"symbols列表。"`
	NianGanPalaceFacts *[]QimenChartYongShenNianGanPalaceFactsItem `json:"nian_gan_palace_facts,omitempty" jsonschema:"nian_gan_palace_facts列表。"`
}

type QimenChartYongShenSymbolsItem struct {
	Symbol           string          `json:"symbol" jsonschema:"符号。"`
	Palace           QimenGongName   `json:"palace" jsonschema:"宫位。"`
	TianGan          []DiPanGanName  `json:"tian_gan" jsonschema:"tian_gan列表。"`
	KongWangBranches []LiuyaoZhiName `json:"kong_wang_branches" jsonschema:"kong_wang_branches列表。"`
	MaXingBranch     *LiuyaoZhiName  `json:"ma_xing_branch,omitempty" jsonschema:"枚举值。"`
}

type QimenChartYongShenNianGanPalaceFactsItem struct {
	Palace QimenGongName `json:"palace" jsonschema:"宫位。"`
	Layer  LayerName     `json:"layer" jsonschema:"层次。"`
}

type QimenChartSpecialized struct {
	GengGe          []QimenChartSpecializedGengGeItem          `json:"geng_ge" jsonschema:"专占领域原子事实；Python 只做 projection列表。"`
	LostContext     []QimenChartSpecializedLostContextItem     `json:"lost_context" jsonschema:"专占领域原子事实；Python 只做 projection列表。"`
	ThiefContext    []QimenChartSpecializedThiefContextItem    `json:"thief_context" jsonschema:"专占领域原子事实；Python 只做 projection列表。"`
	TianwangContext []QimenChartSpecializedTianwangContextItem `json:"tianwang_context" jsonschema:"专占领域原子事实；Python 只做 projection列表。"`
}

type QimenChartSpecializedGengGeItem struct {
	Level    string         `json:"level" jsonschema:"等级。"`
	Gong     QimenGongName  `json:"gong" jsonschema:"宫位。"`
	Door     *QimenDoorName `json:"door,omitempty" jsonschema:"door。"`
	EarthGan string         `json:"earth_gan" jsonschema:"earth_gan。"`
}

type QimenChartSpecializedLostContextItem struct {
	Gong      QimenGongName   `json:"gong" jsonschema:"宫位。"`
	Door      *QimenDoorName  `json:"door,omitempty" jsonschema:"door。"`
	WangShuai LiuyaoWangShuai `json:"wang_shuai" jsonschema:"wang_shuai。"`
}

type QimenChartSpecializedThiefContextItem struct {
	Role              string        `json:"role" jsonschema:"角色。"`
	Symbol            string        `json:"symbol" jsonschema:"符号。"`
	Gong              QimenGongName `json:"gong" jsonschema:"宫位。"`
	WangShuaiLabel    string        `json:"wang_shuai_label" jsonschema:"wang_shuai_label。"`
	LuckyPatternNames []string      `json:"lucky_pattern_names" jsonschema:"lucky_pattern_names列表。"`
}

type QimenChartSpecializedTianwangContextItem struct {
	Kind           *TianwangContextKindName `json:"kind,omitempty" jsonschema:"类别。"`
	Gan            *LiuYiName               `json:"gan,omitempty" jsonschema:"天干。"`
	Gong           *string                  `json:"gong,omitempty" jsonschema:"宫位。"`
	HourGong       *string                  `json:"hour_gong,omitempty" jsonschema:"hour_gong。"`
	RelationToHour *string                  `json:"relation_to_hour,omitempty" jsonschema:"relation_to_hour。"`
}

type QimenChartRiGanPalaceFactsItem struct {
	Palace QimenGongName `json:"palace" jsonschema:"宫位。"`
	Layer  LayerName     `json:"layer" jsonschema:"层次。"`
}

type QimenChartShiGanPalaceFactsItem struct {
	Palace QimenGongName `json:"palace" jsonschema:"宫位。"`
	Layer  LayerName     `json:"layer" jsonschema:"层次。"`
}
