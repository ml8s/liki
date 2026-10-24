// Wire enum types used by the migrated engine tool contracts.
//
// The value sets come from the legacy schemas and the captured payloads under
// testdata/actual/. Edit them directly when the engine grows a value;
// TestEngineContractWireFidelity fails if a captured payload no longer fits.

package mcpcontract

import (
	"reflect"

	"github.com/google/jsonschema-go/jsonschema"
)

// registerWireEnum binds a generated wire enum to its JSON Schema values.
func registerWireEnum[T ~string](values ...string) {
	enum := make([]any, 0, len(values))
	for _, value := range values {
		enum = append(enum, value)
	}
	forOptions.TypeSchemas[reflect.TypeFor[T]()] = &jsonschema.Schema{Type: "string", Enum: enum}
}

// registerWireNullableEnum allows an explicit JSON null in addition to the
// enum values, because the legacy engine emitted null for real fields such
// as a pillar 藏干 slot that has no middle or minor stem.
func registerWireNullableEnum[T ~string](values ...string) {
	enum := make([]any, 0, len(values))
	for _, value := range values {
		enum = append(enum, value)
	}
	forOptions.TypeSchemas[reflect.TypeFor[T]()] = &jsonschema.Schema{
		Types: []string{"null", "string"},
		AnyOf: []*jsonschema.Schema{
			{Type: "null"},
			{Type: "string", Enum: enum},
		},
	}
}

func init() {
	registerWireEnum[CalendarPeriodHalfName]("first", "second", "whole")
	registerWireEnum[CandidatesTypeName]("duty_door", "duty_star", "kong_wang", "ma_xing")
	registerWireEnum[ChangShengStageName]("", "临官", "养", "冠带", "墓", "帝旺", "死", "沐浴", "病", "绝", "胎", "衰", "长生")
	registerWireNullableEnum[ChangShengStageNameNullable]("", "临官", "养", "冠带", "墓", "帝旺", "死", "沐浴", "病", "绝", "胎", "衰", "长生")
	registerWireEnum[DatesMatchName]("opposite", "same")
	registerWireEnum[DayBranchTypeName]("", "墓库", "桃花", "驿马")
	registerWireEnum[DayFuTouGanName]("己", "甲")
	registerWireEnum[DayFuTouName]("己丑", "己亥", "己卯", "己巳", "己未", "己酉", "甲午", "甲子", "甲寅", "甲戌", "甲申", "甲辰")
	registerWireEnum[DiPanGanName]("丁", "丙", "乙", "壬", "己", "庚", "戊", "癸", "辛")
	registerWireEnum[DualReferenceName]("劫煞", "华盖", "天乙贵人", "将星", "桃花", "灾煞", "驿马")
	registerWireEnum[DunSourceName]("hour_branch", "solar_term")
	registerWireEnum[CandidateWuxingName]("土", "木", "水", "火", "金", "无候选")
	registerWireEnum[SecondaryWuxingName]("土", "木", "水", "火", "金", "无辅调候")
	registerWireEnum[GanChongPositionName]("adjacent", "remote", "separated")
	registerWireEnum[GongJiaTypeName]("三会拱", "三合拱")
	registerWireEnum[HourBoundaryName]("late_zi", "zi_zheng")
	registerWireEnum[JuShuName]("土五局", "木三局", "水二局", "火六局", "金四局")
	registerWireEnum[LayerName]("earth", "heaven")
	registerWireEnum[LeadXunShouGanName]("甲")
	registerWireEnum[LeadXunShouName]("甲午", "甲子", "甲寅", "甲戌", "甲申", "甲辰")
	registerWireEnum[LiuYiName]("壬", "己", "庚", "戊", "癸", "辛")
	registerWireNullableEnum[LiuyaoGanNameNullable]("丁", "丙", "乙", "壬", "己", "庚", "戊", "甲", "癸", "辛")
	registerWireNullableEnum[LiuyaoZhiNameNullable]("丑", "亥", "午", "卯", "子", "寅", "巳", "戌", "未", "申", "辰", "酉")
	registerWireEnum[MethodDayBoundaryName]("late_zi_rolls_day_pillar")
	registerWireEnum[MethodDingjuMethodName]("chaibu", "maoshan", "none", "zhirun")
	registerWireEnum[MethodDingjuMethodName2]("拆补定局", "无定局", "置闰定局", "茅山定局")
	registerWireEnum[MethodDoorModeName]("center_projects_kun", "eight_doors_real_center_step", "nine_doors_with_center_forward", "three_day_eight_door_rotation")
	registerWireEnum[MethodDunSourceName]("winter_solstice_yang_summer_solstice_yin")
	registerWireEnum[MethodGeometryName]("luoshu_ring_rotation", "mingfa_star_door_forward_flight", "palace_number_shift", "solstice_day_star_flight")
	registerWireEnum[MethodJieQiName]("冬至", "处暑", "夏至", "大寒", "大暑", "大雪", "寒露", "小寒", "小暑", "小满", "小雪", "惊蛰", "春分", "清明", "白露", "秋分", "立冬", "立夏", "立春", "立秋", "芒种", "谷雨", "雨水", "霜降")
	registerWireEnum[MethodMethodSourceName]("month_cycle", "quarter_rule", "solar_term", "solstice_day_cycle", "year_cycle")
	registerWireEnum[MethodQuarterRuleName]("ten_minute_sanyuan", "twelve_minute_ten_division")
	registerWireEnum[MethodQuarterRuleName2]("十二分钟十分局刻家", "十分钟三元刻家")
	registerWireEnum[MethodSchoolName]("jinhan_yujing", "luoshu_feipan", "mingfa_feipan", "zhuanpan")
	registerWireEnum[MethodSchoolName2]("洛书飞盘法", "转盘法", "金函玉镜", "鸣法飞盘")
	registerWireEnum[MethodScopeName]("day", "hour", "month", "quarter", "year")
	registerWireEnum[MethodScopeName2]("刻家奇门", "年家周期奇门", "日家奇门", "时家奇门", "月家周期奇门")
	registerWireEnum[MethodSpiritFlightName]("solstice_yang_forward_yin_reverse", "yang_forward_yin_reverse")
	registerWireEnum[MethodSpiritModeName]("day_gan_twelve_spirits", "eight_spirits_yang_gouchen_zhuque_yin_baihu_xuanwu", "nine_spirits_baihu_taichang_xuanwu", "nine_spirits_gouchen_taichang_zhuque")
	registerWireEnum[MethodStarModeName]("nine_stars_forward_with_center", "star_gan_shift_together", "taiyi_nine_stars_day_flight", "tian_qin_lodges_kun")
	registerWireEnum[MethodYuanName]("上元", "下元", "中元")
	registerWireEnum[MethodZhiRunStateName]("接气", "正授", "置闰", "超神")
	registerWireEnum[PalaceFactsKindName]("brightness", "si_hua", "special", "star")
	registerWireEnum[PatternsName]("三奇入墓", "三奇得使", "人遁", "伏吟", "六仪击刑", "反吟", "地遁", "天遁", "太白入荧", "玉女守门", "荧入太白", "青龙返首", "飞鸟跌穴")
	registerWireEnum[PillarName]("nian", "ri", "shi", "yue")
	registerWireEnum[QimenDoorName]("中门", "休门", "伤门", "开门", "惊门", "景门", "杜门", "死门", "生门")
	registerWireEnum[QimenEightGongName]("乾", "兑", "坎", "坤", "巽", "离", "艮", "震")
	registerWireEnum[QimenGongName]("中", "乾", "兑", "坎", "坤", "巽", "离", "艮", "震")
	registerWireEnum[QimenSpiritName]("九地", "九天", "值符", "六合", "勾陈", "太常", "太阴", "朱雀", "玄武", "白虎", "螣蛇")
	registerWireEnum[QimenStarName]("天任", "天冲", "天心", "天柱", "天禽", "天芮", "天英", "天蓬", "天辅")
	registerWireEnum[RelatedToRoleName]("duty", "lead", "pillar", "yong_shen")
	registerWireEnum[RelationGroupsFieldName]("an_he", "gan_chong", "gan_chong_candidate", "gan_he", "gan_he_candidate", "liu_chong", "liu_hai", "liu_po", "liu_xing", "san_he", "san_he_partial", "san_hui", "zhi_liu_he")
	registerWireEnum[RiShiRelationObjectName]("shi_gan_gong")
	registerWireEnum[RiShiRelationRelationName]("ri_controls_shi", "ri_generates_shi", "same", "shi_controls_ri", "shi_generates_ri")
	registerWireEnum[RiShiRelationSubjectName]("ri_gan_gong")
	registerWireEnum[RootRefsSourceName]("natal_hidden_stem", "sitting_branch_main_qi")
	registerWireEnum[SanHePartialTypeName]("半合")
	registerWireEnum[SanHeTypeName]("三合")
	registerWireEnum[SanHuiTypeName]("三会")
	registerWireEnum[SchoolLeapMonthName]("first_fifteen_days_current_month_next_fifteen_days_following_month")
	registerWireEnum[SchoolSourceName]("iztro_compatible")
	registerWireEnum[ShenShaCategoryName]("中性", "凶", "吉")
	registerWireNullableEnum[ShenShaCategoryNameNullable]("中性", "凶", "吉")
	registerWireEnum[ShenShaSchoolPolicyName]("union_of_year_and_day_references")
	registerWireEnum[ShenshaName]("丧门", "劫煞", "文昌", "华盖", "吊客", "大耗", "天乙贵人", "天喜", "桃花", "灾煞", "病符", "红鸾", "羊刃", "驿马")
	registerWireEnum[SiHuaName]("忌", "权", "禄", "科")
	registerWireNullableEnum[SiHuaNameNullable]("忌", "权", "禄", "科")
	registerWireEnum[SpousePalaceStateName]("冲", "刑", "合", "害", "静")
	registerWireEnum[StrengthName]("neutral", "strong", "weak")
	registerWireNullableEnum[StrengthNameNullable]("neutral", "strong", "weak")
	registerWireEnum[TenGodName]("七杀", "伤官", "偏印", "偏财", "劫财", "正印", "正官", "正财", "比肩", "食神")
	registerWireNullableEnum[TenGodNameNullable]("七杀", "伤官", "偏印", "偏财", "劫财", "正印", "正官", "正财", "比肩", "食神")
	registerWireEnum[TianQinFollowsName]("天芮")
	registerWireEnum[TianQinStarName]("天禽")
	registerWireEnum[TianwangContextKindName]("地罗遮蔽", "天网四张")
	registerWireEnum[XingGongWuXingRelationName]("gong_controls_xing", "gong_generates_xing", "same", "xing_controls_gong", "xing_generates_gong")
	registerWireEnum[XingName]("咸池", "天乙", "天符", "太乙", "太阴", "招摇", "摄提", "轩辕", "青龙")
	registerWireEnum[YearBranchRelationsName]("half_he", "liu_chong", "liu_hai", "liu_he", "san_he", "san_he_gong", "san_hui", "san_hui_gong", "xing")
	registerWireEnum[YearStemTenGodName]("", "七杀", "伤官", "偏印", "偏财", "劫财", "正印", "正官", "正财", "比肩", "食神")
	registerWireEnum[ZiweiPalaceName]("仆役", "兄弟", "命宫", "夫妻", "子女", "官禄", "父母", "田宅", "疾厄", "福德", "财帛", "迁移")
}

type CalendarPeriodHalfName string

type CandidatesTypeName string

type ChangShengStageName string

type ChangShengStageNameNullable string

type DatesMatchName string

type DayBranchTypeName string

type DayFuTouGanName string

type DayFuTouName string

type DiPanGanName string

type DualReferenceName string

type DunSourceName string

type CandidateWuxingName string

type SecondaryWuxingName string

type GanChongPositionName string

type GongJiaTypeName string

type HourBoundaryName string

type JuShuName string

type LayerName string

type LeadXunShouGanName string

type LeadXunShouName string

type LiuYiName string

type LiuyaoGanNameNullable string

type LiuyaoZhiNameNullable string

type MethodDayBoundaryName string

type MethodDingjuMethodName string

type MethodDingjuMethodName2 string

type MethodDoorModeName string

type MethodDunSourceName string

type MethodGeometryName string

type MethodJieQiName string

type MethodMethodSourceName string

type MethodQuarterRuleName string

type MethodQuarterRuleName2 string

type MethodSchoolName string

type MethodSchoolName2 string

type MethodScopeName string

type MethodScopeName2 string

type MethodSpiritFlightName string

type MethodSpiritModeName string

type MethodStarModeName string

type MethodYuanName string

type MethodZhiRunStateName string

type PalaceFactsKindName string

type PatternsName string

type PillarName string

type QimenDoorName string

type QimenEightGongName string

type QimenGongName string

type QimenSpiritName string

type QimenStarName string

type RelatedToRoleName string

type RelationGroupsFieldName string

type RiShiRelationObjectName string

type RiShiRelationRelationName string

type RiShiRelationSubjectName string

type RootRefsSourceName string

type SanHePartialTypeName string

type SanHeTypeName string

type SanHuiTypeName string

type SchoolLeapMonthName string

type SchoolSourceName string

type ShenShaCategoryName string

type ShenShaCategoryNameNullable string

type ShenShaSchoolPolicyName string

type ShenshaName string

type SiHuaName string

type SiHuaNameNullable string

type SpousePalaceStateName string

type StrengthName string

type StrengthNameNullable string

type TenGodName string

type TenGodNameNullable string

type TianQinFollowsName string

type TianQinStarName string

type TianwangContextKindName string

type XingGongWuXingRelationName string

type XingName string

type YearBranchRelationsName string

type YearStemTenGodName string

type ZiweiPalaceName string
