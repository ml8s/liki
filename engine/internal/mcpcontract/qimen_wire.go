// Engine wire DTOs for the MCP contract layer.
//
// These types are ordinary source now, not generated artifacts. Edit them
// directly when the published engine payload changes.
package mcpcontract

import (
	"encoding/json"

	"github.com/google/jsonschema-go/jsonschema"
)

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

// QimenChartInput takes the Qimen event, a closed method object, and optional
// chart-analysis facts. Event, method, and focus stay orthogonal.
type QimenChartInput struct {
	StandardTime StandardTime         `json:"standard_time" jsonschema:"起局地标准时间（无 DST），RFC3339 带 offset；节气换局/置闰按此时刻判定。"`
	Longitude    Longitude            `json:"longitude" jsonschema:"起局地经度（-180..180）；真太阳时校正用于时柱与分钟分法。"`
	Method       QimenMethodInput     `json:"method" jsonschema:"奇门方法事实；scope、盘式流派和条件参数按目录组合传入。"`
	YongShen     QimenYongShenSymbols `json:"yong_shen,omitempty" jsonschema:"用神符号组合；甲不露盘，不能作为独立符号"`
	Subject      *NatalSubject        `json:"subject,omitempty" jsonschema:"可选命主；提供时用于年命。"`
}

type qimenScopeSchoolDingjuForm struct {
	Scope        MethodScopeName        `json:"scope" jsonschema:"时间层级。"`
	School       MethodSchoolName       `json:"school" jsonschema:"盘式流派。"`
	DingjuMethod MethodDingjuMethodName `json:"dingju_method" jsonschema:"定局法。"`
}

type qimenScopeSchoolForm struct {
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

type qimenChartStandardForm struct {
	StandardTime StandardTime         `json:"standard_time" jsonschema:"起局地标准时间。"`
	Longitude    Longitude            `json:"longitude" jsonschema:"起局地经度。"`
	Method       QimenMethodInput     `json:"method" jsonschema:"非金函玉镜奇门方法事实。"`
	YongShen     QimenYongShenSymbols `json:"yong_shen,omitempty" jsonschema:"用神符号组合。"`
	Subject      *NatalSubject        `json:"subject,omitempty" jsonschema:"可选命主；提供时用于年命。"`
}

type qimenChartJinhanForm struct {
	StandardTime StandardTime         `json:"standard_time" jsonschema:"起局地标准时间。"`
	Longitude    Longitude            `json:"longitude" jsonschema:"起局地经度。"`
	Method       qimenScopeSchoolForm `json:"method" jsonschema:"金函玉镜方法事实。"`
}

func qimenMethodFormSchema[T any](options *jsonschema.ForOptions,
	constraints map[string]string, enums map[string][]string,
) (json.RawMessage, error) {
	_, raw, err := inputSchemaFor[T](options)
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

func qimenMethodFormsSchema(options *jsonschema.ForOptions) (json.RawMessage, error) {
	builders := []func() (json.RawMessage, error){
		func() (json.RawMessage, error) {
			raw, err := qimenMethodFormSchema[qimenScopeSchoolDingjuForm](options,
				map[string]string{"scope": "hour", "school": "zhuanpan"},
				map[string][]string{"dingju_method": {"chaibu", "zhirun", "maoshan"}},
			)
			if err == nil {
				raw, err = withSchemaOptionalDefault(raw, "dingju_method", "chaibu")
			}
			return raw, err
		},
		func() (json.RawMessage, error) {
			raw, err := qimenMethodFormSchema[qimenScopeSchoolDingjuForm](options,
				map[string]string{"scope": "hour", "school": "luoshu_feipan"},
				map[string][]string{"dingju_method": {"chaibu", "zhirun"}},
			)
			if err == nil {
				raw, err = withSchemaOptionalDefault(raw, "dingju_method", "chaibu")
			}
			return raw, err
		},
		func() (json.RawMessage, error) {
			return qimenMethodFormSchema[qimenScopeSchoolForm](options,
				map[string]string{"scope": "hour", "school": "mingfa_feipan"}, nil,
			)
		},
		func() (json.RawMessage, error) {
			raw, err := qimenMethodFormSchema[qimenScopeSchoolDingjuForm](options,
				map[string]string{"scope": "day", "school": "zhuanpan"},
				map[string][]string{"dingju_method": {"chaibu", "zhirun"}},
			)
			if err == nil {
				raw, err = withSchemaOptionalDefault(raw, "dingju_method", "chaibu")
			}
			return raw, err
		},
		func() (json.RawMessage, error) {
			raw, err := qimenMethodFormSchema[qimenScopeSchoolDingjuForm](options,
				map[string]string{"scope": "day", "school": "luoshu_feipan"},
				map[string][]string{"dingju_method": {"chaibu", "zhirun"}},
			)
			if err == nil {
				raw, err = withSchemaOptionalDefault(raw, "dingju_method", "chaibu")
			}
			return raw, err
		},
		func() (json.RawMessage, error) {
			return qimenMethodFormSchema[qimenScopeSchoolForm](options,
				map[string]string{"scope": "month"},
				map[string][]string{"school": {"zhuanpan", "luoshu_feipan"}},
			)
		},
		func() (json.RawMessage, error) {
			return qimenMethodFormSchema[qimenScopeSchoolForm](options,
				map[string]string{"scope": "year"},
				map[string][]string{"school": {"zhuanpan", "luoshu_feipan"}},
			)
		},
		func() (json.RawMessage, error) {
			return qimenMethodFormSchema[qimenQuarterTenMethodForm](options,
				map[string]string{
					"scope": "quarter", "school": "zhuanpan",
					"quarter_rule": "ten_minute_sanyuan",
				}, nil,
			)
		},
		func() (json.RawMessage, error) {
			return qimenMethodFormSchema[qimenQuarterTwelveMethodForm](options,
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

func qimenJinhanMethodSchema(options *jsonschema.ForOptions) (json.RawMessage, error) {
	return qimenMethodFormSchema[qimenScopeSchoolForm](options,
		map[string]string{"scope": "day", "school": "jinhan_yujing"}, nil,
	)
}

func qimenChartFormsSchema(options *jsonschema.ForOptions) (json.RawMessage, error) {
	methodForms, err := qimenMethodFormsSchema(options)
	if err != nil {
		return nil, err
	}
	_, standard, err := inputSchemaFor[qimenChartStandardForm](options)
	if err != nil {
		return nil, err
	}
	standard, err = withSchemaNumberBounds(standard, "longitude", -180, 180)
	if err != nil {
		return nil, err
	}
	standard, err = withSchemaProperty(standard, "method", methodForms)
	if err != nil {
		return nil, err
	}

	jinhanMethod, err := qimenJinhanMethodSchema(options)
	if err != nil {
		return nil, err
	}
	_, jinhan, err := inputSchemaFor[qimenChartJinhanForm](options)
	if err != nil {
		return nil, err
	}
	jinhan, err = withSchemaNumberBounds(jinhan, "longitude", -180, 180)
	if err != nil {
		return nil, err
	}
	jinhan, err = withSchemaProperty(jinhan, "method", jinhanMethod)
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
	Method           QimenChartMethod                  `json:"method" jsonschema:"排盘方法元数据"`
	Board            QimenChartPan                     `json:"pan" jsonschema:"盘面。"`
	Patterns         *[]QimenChartGeJuItem             `json:"ge_ju,omitempty" jsonschema:"格局。"`
	GanInteraction   *[]QimenChartGanInteractionItem   `json:"gan_interaction,omitempty" jsonschema:"天盘干与地盘干克应列表。"`
	MenInteraction   *[]QimenChartMenInteractionItem   `json:"men_interaction,omitempty" jsonschema:"八门与宫位克应列表。"`
	XingInteraction  *[]QimenChartXingInteractionItem  `json:"xing_interaction,omitempty" jsonschema:"九星与宫位克应列表。"`
	XingGongWuXing   *[]QimenChartXingGongWuXingItem   `json:"xing_gong_wu_xing,omitempty" jsonschema:"星与宫的五行关系；不是完整季节旺衰"`
	MenPo            *[]QimenGongName                  `json:"men_po,omitempty" jsonschema:"门破宫位列表。"`
	MenZhi           *[]QimenGongName                  `json:"men_zhi,omitempty" jsonschema:"门落地支宫位列表。"`
	YingQi           *QimenChartYingQi                 `json:"ying_qi,omitempty" jsonschema:"应期候选。"`
	DayStemPalace    *QimenStemPalace                  `json:"day_stem_palace,omitempty" jsonschema:"日干落宫领域对象；金函玉镜不输出。"`
	HourStemPalace   *QimenStemPalace                  `json:"hour_stem_palace,omitempty" jsonschema:"时干落宫领域对象；金函玉镜不输出。"`
	RiShiRelation    *QimenChartRiShiRelation          `json:"ri_shi_relation,omitempty" jsonschema:"日干宫与时干宫关系。"`
	KongWangAffected *[]QimenChartKongWangAffectedItem `json:"kong_wang_affected,omitempty" jsonschema:"kong_wang_affected列表。"`
	MaXingAffected   *[]QimenChartMaXingAffectedItem   `json:"ma_xing_affected,omitempty" jsonschema:"ma_xing_affected列表。"`
	GongWangShuai    *[]QimenGongWangShuaiFact         `json:"gong_wang_shuai,omitempty" jsonschema:"gong_wang_shuai列表。"`
	ZhiFuXingGong    *QimenGongName                    `json:"zhi_fu_xing_gong,omitempty" jsonschema:"值符星落宫。"`
	ZhiShiMenGong    *QimenGongName                    `json:"zhi_shi_men_gong,omitempty" jsonschema:"值使门落宫。"`
	YongShen         *QimenChartYongShen               `json:"yong_shen,omitempty" jsonschema:"用神落宫与相关事实。"`
	Specialized      *QimenChartSpecialFacts           `json:"special_facts,omitempty" jsonschema:"专占领域原子事实。"`
}

type QimenChartMethod struct {
	Scope            MethodScopeName              `json:"scope" jsonschema:"时间层级"`
	School           MethodSchoolName             `json:"school" jsonschema:"盘式流派"`
	DingjuMethod     *MethodDingjuMethodName      `json:"dingju_method,omitempty" jsonschema:"定局法；quarter 输出 quarter_rule，不输出本字段"`
	QuarterRule      *MethodQuarterRuleName       `json:"quarter_rule,omitempty" jsonschema:"刻家规则；仅 scope：quarter 输出"`
	BaseDingjuMethod *MethodDingjuMethodName      `json:"base_dingju_method,omitempty" jsonschema:"十二分钟十分局的基准时家定局法"`
	MethodSource     MethodMethodSourceName       `json:"method_source" jsonschema:"方法来源：节气定局、刻家规则、月周期或年周期"`
	DayBoundary      MethodDayBoundaryName        `json:"day_boundary" jsonschema:"晚子时日柱口径。"`
	SpiritMode       MethodSpiritModeName         `json:"spirit_mode" jsonschema:"神煞排布模式。"`
	Geometry         MethodGeometryName           `json:"geometry" jsonschema:"盘面几何算法。"`
	StarMode         MethodStarModeName           `json:"star_mode" jsonschema:"九星排布模式。"`
	DoorMode         MethodDoorModeName           `json:"door_mode" jsonschema:"八门排布模式。"`
	SpiritFlight     MethodSpiritFlightName       `json:"spirit_flight" jsonschema:"神煞飞布方向。"`
	TianQin          *QimenChartMethodTianQin     `json:"tian_qin,omitempty" jsonschema:"排盘方法元数据 中的 tian_qin。"`
	JieQi            MethodJieQiName              `json:"jie_qi" jsonschema:"当前定局节气。"`
	YongJuJieQi      *string                      `json:"yong_ju_jie_qi,omitempty" jsonschema:"排盘方法元数据 中的 yong_ju_jie_qi。"`
	Yuan             *MethodYuanName              `json:"yuan,omitempty" jsonschema:"三元。"`
	ZhiRunState      *MethodZhiRunStateName       `json:"zhi_run_state,omitempty" jsonschema:"置闰状态。"`
	LeadPillar       *QimenChartMethodLeadPillar  `json:"lead_pillar,omitempty" jsonschema:"排盘方法元数据 中的 lead_pillar。"`
	LeadXunShou      *QimenChartMethodLeadXunShou `json:"lead_xun_shou,omitempty" jsonschema:"排盘方法元数据 中的 lead_xun_shou。"`
	LeadXunShouGong  *QimenGongName               `json:"lead_xun_shou_gong,omitempty" jsonschema:"排盘方法元数据。"`
	DayFuTou         *QimenChartMethodDayFuTou    `json:"day_fu_tou,omitempty" jsonschema:"排盘方法元数据 中的 day_fu_tou。"`
	Quarter          *QimenChartMethodQuarter     `json:"quarter,omitempty" jsonschema:"刻家序数与主柱口径；干支不重复输出，以 lead_pillar 为准"`
	DunSource        *MethodDunSourceName         `json:"dun_source,omitempty" jsonschema:"阴阳遁来源。"`
	MethodYiJu       *string                      `json:"method_yi_ju,omitempty" jsonschema:"金函方法依据。"`
}

type QimenChartMethodTianQin struct {
	Star                  TianQinStarName    `json:"star" jsonschema:"天禽星。"`
	HomeGong              QimenGongName      `json:"home_gong" jsonschema:"天禽本宫。"`
	LodgingGong           QimenEightGongName `json:"lodging_gong" jsonschema:"天禽寄宫。"`
	Follows               TianQinFollowsName `json:"follows" jsonschema:"所随星曜。"`
	CarriesCenterEarthGan bool               `json:"carries_center_earth_gan" jsonschema:"天禽是否携带中宫土干。"`
	LodgingRule           string             `json:"lodging_rule" jsonschema:"天禽寄宫规则说明。"`
}

type QimenChartMethodLeadPillar struct {
	Name string  `json:"name" jsonschema:"名称。"`
	Gan  GanName `json:"gan" jsonschema:"天干。"`
	Zhi  ZhiName `json:"zhi" jsonschema:"地支。"`
}

type QimenChartMethodLeadXunShou struct {
	Name  LeadXunShouName    `json:"name" jsonschema:"名称。"`
	Gan   LeadXunShouGanName `json:"gan" jsonschema:"天干。"`
	Zhi   ZhiName            `json:"zhi" jsonschema:"地支。"`
	LiuYi LiuYiName          `json:"liu_yi" jsonschema:"六仪。"`
}

type QimenChartMethodDayFuTou struct {
	Name DayFuTouName    `json:"name" jsonschema:"名称。"`
	Gan  DayFuTouGanName `json:"gan" jsonschema:"天干。"`
	Zhi  ZhiName         `json:"zhi" jsonschema:"地支。"`
}

type QimenChartMethodQuarter struct {
	Index             int               `json:"index" jsonschema:"序号。"`
	MinutesPerQuarter int               `json:"minutes_per_quarter" jsonschema:"刻家序数与主柱口径；干支不重复输出，以 lead_pillar 为准 中的 minutes_per_quarter。"`
	LeadPillarMode    MethodScopeName   `json:"lead_pillar_mode" jsonschema:"quarter：刻柱主盘；hour：十二分钟十分局刻家的时柱主盘"`
	DunSource         *DunSourceName    `json:"dun_source,omitempty" jsonschema:"hour_branch：时支分遁；solar_term：沿基准时家节气分遁"`
	HourBoundary      *HourBoundaryName `json:"hour_boundary,omitempty" jsonschema:"late_zi：23点起子时；zi_zheng：0点起子时"`
}

type QimenChartPan struct {
	Jushu      *int                           `json:"jushu,omitempty" jsonschema:"局数。"`
	YinDun     bool                           `json:"yin_dun" jsonschema:"是否阴遁。"`
	RiGan      GanName                        `json:"ri_gan" jsonschema:"日干。"`
	RiZhi      ZhiName                        `json:"ri_zhi" jsonschema:"日支。"`
	NianGan    *GanName                       `json:"nian_gan,omitempty" jsonschema:"年干。"`
	NianZhi    *ZhiName                       `json:"nian_zhi,omitempty" jsonschema:"年支。"`
	YueGan     *GanName                       `json:"yue_gan,omitempty" jsonschema:"月干。"`
	YueZhi     *ZhiName                       `json:"yue_zhi,omitempty" jsonschema:"月支。"`
	ZhiFuXing  *QimenStarName                 `json:"zhi_fu_xing,omitempty" jsonschema:"值符星。"`
	ZhiShiMen  *QimenDoorName                 `json:"zhi_shi_men,omitempty" jsonschema:"值使门。"`
	GongWei    []QimenChartPanGongWeiItem     `json:"gong_wei" jsonschema:"十二宫。"`
	ShiGan     *GanName                       `json:"shi_gan,omitempty" jsonschema:"时干。"`
	ShiZhi     *ZhiName                       `json:"shi_zhi,omitempty" jsonschema:"时支。"`
	MaXing     *QimenBranchPalaceFact         `json:"ma_xing,omitempty" jsonschema:"马星。"`
	KongWang   *[]QimenBranchPalaceFact       `json:"kong_wang,omitempty" jsonschema:"空亡。"`
	WuBuYuShi  *bool                          `json:"wu_bu_yu_shi,omitempty" jsonschema:"五不遇时。"`
	DaySpirits *[]QimenChartPanDaySpiritsItem `json:"day_spirits,omitempty" jsonschema:"金函玉镜日干十二神列表。"`
}

type QimenChartPanGongWeiItem struct {
	Gong        QimenChartPanGongWeiItemGong           `json:"gong" jsonschema:"宫位。"`
	DiPanGan    *DiPanGanName                          `json:"di_pan_gan,omitempty" jsonschema:"地盘干。"`
	AnGan       *DiPanGanName                          `json:"an_gan,omitempty" jsonschema:"暗干。"`
	TianPanGan  *DiPanGanName                          `json:"tian_pan_gan,omitempty" jsonschema:"天盘干。"`
	TianPan     *[]QimenChartPanGongWeiItemTianPanItem `json:"tian_pan,omitempty" jsonschema:"天盘。"`
	Men         *QimenDoorName                         `json:"men,omitempty" jsonschema:"八门。"`
	MenPresent  bool                                   `json:"men_present" jsonschema:"门存在集合。"`
	Shen        *QimenSpiritName                       `json:"shen,omitempty" jsonschema:"八神。"`
	ShenPresent *bool                                  `json:"shen_present,omitempty" jsonschema:"神存在集合。"`
	AnGanZhi    *string                                `json:"an_gan_zhi,omitempty" jsonschema:"鸣法暗干支；仅鸣法飞盘输出，其他方法省略"`
	Xing        *XingName                              `json:"xing,omitempty" jsonschema:"太乙九星。"`
}

type QimenChartPanGongWeiItemGong struct {
	Name      QimenGongName `json:"name" jsonschema:"名称。"`
	Luoshu    int           `json:"luoshu" jsonschema:"洛书宫数。"`
	RingIndex *int          `json:"ring_index,omitempty" jsonschema:"八方环序；中宫省略。"`
}

type QimenChartPanGongWeiItemTianPanItem struct {
	Gan  DiPanGanName  `json:"gan" jsonschema:"天干。"`
	Xing QimenStarName `json:"xing" jsonschema:"九星。"`
}

type QimenBranchPalaceFact struct {
	Branch ZhiName            `json:"branch" jsonschema:"地支。"`
	Gong   QimenEightGongName `json:"gong" jsonschema:"宫位。"`
}

type QimenChartPanDaySpiritsItem struct {
	Zhi  ZhiName `json:"zhi" jsonschema:"地支。"`
	Shen string  `json:"shen" jsonschema:"八神。"`
}

type QimenChartGeJuItem struct {
	Name        PatternsName    `json:"name" jsonschema:"名称。"`
	Description string          `json:"description" jsonschema:"说明。"`
	YiJu        string          `json:"yi_ju" jsonschema:"格局依据。"`
	Auspicious  bool            `json:"auspicious" jsonschema:"是否为吉象。"`
	GongWei     []QimenGongName `json:"gong_wei" jsonschema:"十二宫。"`
}

type QimenChartGanInteractionItem struct {
	Gong       QimenGongName `json:"gong" jsonschema:"宫位。"`
	DiPanGan   DiPanGanName  `json:"di_pan_gan" jsonschema:"地盘干。"`
	TianPanGan DiPanGanName  `json:"tian_pan_gan" jsonschema:"天盘干。"`
	Name       string        `json:"name" jsonschema:"名称。"`
	Meaning    string        `json:"meaning" jsonschema:"克应含义。"`
	Auspicious bool          `json:"auspicious" jsonschema:"是否为吉象。"`
}

type QimenChartMenInteractionItem struct {
	Door    QimenDoorName `json:"door" jsonschema:"宫位八门；中宫无门时省略。"`
	Gong    QimenGongName `json:"gong" jsonschema:"宫位。"`
	Name    string        `json:"name" jsonschema:"名称。"`
	Meaning string        `json:"meaning" jsonschema:"克应含义。"`
}

type QimenChartXingInteractionItem struct {
	Xing       QimenStarName `json:"xing" jsonschema:"九星。"`
	Gong       QimenGongName `json:"gong" jsonschema:"宫位。"`
	Name       string        `json:"name" jsonschema:"名称。"`
	Meaning    string        `json:"meaning" jsonschema:"克应含义。"`
	Auspicious bool          `json:"auspicious" jsonschema:"是否为吉象。"`
}

type QimenChartXingGongWuXingItem struct {
	Xing             QimenStarName              `json:"xing" jsonschema:"九星。"`
	Gong             QimenGongName              `json:"gong" jsonschema:"宫位。"`
	Relation         XingGongWuXingRelationName `json:"relation" jsonschema:"关系。"`
	RelationName     string                     `json:"relation_name" jsonschema:"星宫五行关系名称。"`
	TraditionalLabel string                     `json:"traditional_label" jsonschema:"传统旺衰标记。"`
}

type QimenChartYingQi struct {
	Candidates []QimenChartYingQiCandidatesItem `json:"candidates" jsonschema:"应期候选：马星、空亡、值符宫与值使宫给出的动期候选；候选不是吉凶结论。"`
}

type QimenChartYingQiCandidatesItem struct {
	Type      CandidatesTypeName                            `json:"ying_qi_kind" jsonschema:"应期候选类型。"`
	Mechanism string                                        `json:"mechanism" jsonschema:"应期触发机制。"`
	Branch    ZhiName                                       `json:"branch" jsonschema:"地支。"`
	Gong      QimenEightGongName                            `json:"gong" jsonschema:"宫位。"`
	RelatedTo []QimenChartYingQiCandidatesItemRelatedToItem `json:"related_to" jsonschema:"用神解释可关联的落宫事实；空列表表示不进入用神解释。"`
	Dates     []QimenChartYingQiCandidatesItemDatesItem     `json:"dates" jsonschema:"候选触发日期。"`
}

type QimenChartYingQiCandidatesItemRelatedToItem struct {
	Symbol string            `json:"symbol" jsonschema:"符号。"`
	Role   RelatedToRoleName `json:"role" jsonschema:"角色。"`
}

type QimenChartYingQiCandidatesItemDatesItem struct {
	Date   DateString     `json:"trigger_date" jsonschema:"应期日期。"`
	Branch ZhiName        `json:"branch" jsonschema:"地支。"`
	Match  DatesMatchName `json:"match" jsonschema:"应期触发方向。"`
}

type QimenChartRiShiRelation struct {
	Zhu      RiShiRelationSubjectName  `json:"zhu" jsonschema:"主。"`
	Ke       RiShiRelationObjectName   `json:"ke" jsonschema:"客。"`
	Relation RiShiRelationRelationName `json:"relation" jsonschema:"关系。"`
	Name     string                    `json:"name" jsonschema:"名称。"`
}

type QimenChartKongWangAffectedItem struct {
	Symbol string             `json:"symbol" jsonschema:"符号。"`
	Branch ZhiName            `json:"branch" jsonschema:"地支。"`
	Gong   QimenEightGongName `json:"gong" jsonschema:"宫位。"`
	Role   RelatedToRoleName  `json:"role" jsonschema:"角色。"`
}

type QimenChartMaXingAffectedItem struct {
	Symbol *string             `json:"symbol,omitempty" jsonschema:"符号。"`
	Branch *ZhiName            `json:"branch,omitempty" jsonschema:"地支。"`
	Gong   *QimenEightGongName `json:"gong,omitempty" jsonschema:"宫位。"`
	Role   *RelatedToRoleName  `json:"role,omitempty" jsonschema:"角色。"`
}

type QimenGongWangShuaiFact struct {
	Gong      QimenGongName `json:"gong" jsonschema:"宫位。"`
	Wuxing    ElementName   `json:"wuxing" jsonschema:"五行。"`
	WangShuai WangShuaiName `json:"wang_shuai" jsonschema:"旺相休囚死状态。"`
}

type QimenChartYongShen struct {
	NianGanGong      *QimenGongName                  `json:"nian_gan_gong,omitempty" jsonschema:"年干落宫。"`
	Symbols          []QimenChartYongShenSymbolsItem `json:"symbols" jsonschema:"用神符号及落宫事实。"`
	NianGanGongFacts *[]QimenStemPalaceFact          `json:"nian_gan_gong_facts,omitempty" jsonschema:"年命干落宫事实列表。"`
}

type QimenChartYongShenSymbolsItem struct {
	Symbol           string         `json:"symbol" jsonschema:"符号。"`
	Palace           QimenGongName  `json:"gong" jsonschema:"宫位。"`
	TianGan          []DiPanGanName `json:"tian_gan" jsonschema:"相关天盘干列表。"`
	KongWangBranches []ZhiName      `json:"kong_wang_branches" jsonschema:"相关旬空地支列表。"`
	MaXingBranch     *ZhiName       `json:"ma_xing_branch,omitempty" jsonschema:"驿马落支。"`
}

type QimenChartSpecialFacts struct {
	GengGe          []QimenChartSpecialFactsGengGeItem          `json:"geng_ge" jsonschema:"击刑格局事实列表。"`
	LostContext     []QimenChartSpecialFactsLostContextItem     `json:"lost_facts" jsonschema:"失物专占事实列表。"`
	ThiefContext    []QimenChartSpecialFactsThiefContextItem    `json:"thief_facts" jsonschema:"捕盗事实列表。"`
	TianwangContext []QimenChartSpecialFactsTianwangContextItem `json:"tianwang_facts" jsonschema:"天网事实列表。"`
}

type QimenChartSpecialFactsGengGeItem struct {
	Level    string         `json:"level" jsonschema:"等级。"`
	Gong     QimenGongName  `json:"gong" jsonschema:"宫位。"`
	Door     *QimenDoorName `json:"door,omitempty" jsonschema:"宫位门；无门时省略。"`
	EarthGan string         `json:"earth_gan" jsonschema:"地盘干。"`
}

type QimenChartSpecialFactsLostContextItem struct {
	Gong      QimenGongName  `json:"gong" jsonschema:"宫位。"`
	Door      *QimenDoorName `json:"door,omitempty" jsonschema:"宫位门；无门时省略。"`
	WangShuai WangShuaiName  `json:"wang_shuai" jsonschema:"旺相休囚死状态。"`
}

type QimenChartSpecialFactsThiefContextItem struct {
	Role              string        `json:"role" jsonschema:"角色。"`
	Symbol            string        `json:"symbol" jsonschema:"符号。"`
	Gong              QimenGongName `json:"gong" jsonschema:"宫位。"`
	WangShuaiLabel    string        `json:"wang_shuai" jsonschema:"落宫旺衰。"`
	LuckyPatternNames []string      `json:"lucky_ge_ju_names" jsonschema:"有利格局名称列表。"`
}

type QimenChartSpecialFactsTianwangContextItem struct {
	Kind           *TianwangContextKindName `json:"kind,omitempty" jsonschema:"类别。"`
	Gan            *LiuYiName               `json:"gan,omitempty" jsonschema:"天干。"`
	Gong           *string                  `json:"gong,omitempty" jsonschema:"宫位。"`
	HourGong       *string                  `json:"hour_gong,omitempty" jsonschema:"时干落宫。"`
	RelationToHour *string                  `json:"relation_to_hour,omitempty" jsonschema:"与时干宫的关系。"`
}

type QimenStemPalaceFact struct {
	Palace QimenGongName `json:"gong" jsonschema:"宫位。"`
	Layer  LayerName     `json:"layer" jsonschema:"层次。"`
}

type QimenStemPalace struct {
	Palace    QimenGongName          `json:"palace" jsonschema:"天盘优先落宫。"`
	Facts     []QimenStemPalaceFact  `json:"facts" jsonschema:"天盘/地盘落宫明细。"`
	WangShuai QimenGongWangShuaiFact `json:"wang_shuai" jsonschema:"落宫旺相休囚死。"`
}
