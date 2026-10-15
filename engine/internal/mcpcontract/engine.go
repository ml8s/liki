package mcpcontract

// This file is the engine's typed tool catalogue: the wire vocabulary the
// schemas derive from (forOptions), the registration point (NewEngine), and the
// helper the registrations share. The registry mechanism itself lives in
// registry.go, and each tool's DTOs live beside the tool they describe.

import (
	"encoding/json"
	"fmt"
	"reflect"

	"github.com/google/jsonschema-go/jsonschema"
	"liki-engine/internal/engine/liuyao"
)

// forOptions adds reusable wire-level constraints that Go struct tags cannot
// express through this library. Field-level descriptions still come from the
// DTO's `jsonschema` tags.
var forOptions = &jsonschema.ForOptions{TypeSchemas: map[reflect.Type]*jsonschema.Schema{
	// A few legacy wire fields carry more than one JSON type (a pillar 六破
	// reference is a name in some charts and an index in others). Those fields
	// keep their exact JSON value instead of being reshaped by a Go type.
	reflect.TypeFor[json.RawMessage](): {},
	reflect.TypeFor[RFC3339Time]():     {Type: "string", Format: "date-time"},
	reflect.TypeFor[DateString]():      {Type: "string", Format: "date"},
	reflect.TypeFor[GenderName]():      {Type: "string", Enum: []any{"male", "female"}},
	reflect.TypeFor[CityName]():        {Type: "string", MinLength: jsonschema.Ptr(1), MaxLength: jsonschema.Ptr(128)},
	reflect.TypeFor[HuangliEvent](): {
		Type: "string",
		Enum: []any{
			"wedding", "engage", "opening", "sign", "move", "travel", "build",
			"exam", "medical", "sacrifice", "cleaning", "renovation",
			"bed_install", "income", "funeral",
		},
	},
	reflect.TypeFor[HuangliCount](): {Type: "integer", Minimum: jsonschema.Ptr(float64(1)), Maximum: jsonschema.Ptr(float64(30))},
	reflect.TypeFor[GuaNumber]():    {Type: "integer", Minimum: jsonschema.Ptr(float64(1)), Maximum: jsonschema.Ptr(float64(9))},
	reflect.TypeFor[MountainIndex](): {
		Type: "integer", Minimum: jsonschema.Ptr(float64(0)), Maximum: jsonschema.Ptr(float64(23)),
	},
	reflect.TypeFor[XuankongYear](): {
		Type: "integer", Minimum: jsonschema.Ptr(float64(1864)), Maximum: jsonschema.Ptr(float64(2200)),
	},
	reflect.TypeFor[FlyingStarNumber](): {
		Type: "integer", Minimum: jsonschema.Ptr(float64(1)), Maximum: jsonschema.Ptr(float64(9)),
	},
	reflect.TypeFor[BazhaiBirthYear](): {
		Type: "integer", Minimum: jsonschema.Ptr(float64(1900)), Maximum: jsonschema.Ptr(float64(2100)),
	},
	reflect.TypeFor[EightGuaName](): {
		Type: "string", Enum: []any{"坎", "坤", "震", "巽", "乾", "兑", "艮", "离"},
	},
	reflect.TypeFor[ElementName](): {
		// The empty string is part of the wire contract for element fields
		// whose analysis has no candidate: the engine documents this for
		// bazi 调候 TiaoHouResult.SecondaryWuxing ("喜神 may be empty"), and the
		// JSON schema must accept everything the engine can emit.
		Type: "string", Enum: []any{"", "水", "土", "木", "金", "火"},
	},
	reflect.TypeFor[StarColorName](): {
		Type: "string", Enum: []any{"白", "黑", "碧", "绿", "黄", "赤", "紫"},
	},
	reflect.TypeFor[MountainName](): {
		Type: "string", Enum: []any{
			"子", "癸", "丑", "艮", "寅", "甲", "卯", "乙", "辰", "巽", "巳", "丙",
			"午", "丁", "未", "坤", "申", "庚", "酉", "辛", "戌", "乾", "亥", "壬",
		},
	},
	reflect.TypeFor[XuankongYuanName](): {Type: "string", Enum: []any{"上元", "中元", "下元"}},
	reflect.TypeFor[YunName](): {
		Type: "string", Enum: []any{
			"一运", "二运", "三运", "四运", "五运", "六运", "七运", "八运", "九运",
		},
	},
	reflect.TypeFor[XuankongSituationName](): {
		Type: "string", Enum: []any{
			"旺山旺向", "双星会坐", "双星会向", "上山下水", "未入四大局",
		},
	},
	reflect.TypeFor[XuankongFuYinLayer](): {
		Type: "string", Enum: []any{"运盘伏吟", "山星伏吟", "向星伏吟", "全盘伏吟"},
	},
	reflect.TypeFor[XuankongFanYinLayer](): {
		Type: "string", Enum: []any{"运盘反吟", "山星反吟", "向星反吟"},
	},
	reflect.TypeFor[XuankongAssessment](): {
		Type: "string", Enum: []any{
			"收山得宜、拨水入零堂，理气合局（完整收山出煞须验实际砂水）",
			"收山得宜（坐宫山星当令）；拨水入零堂未得，向首零神不见水",
			"拨水入零堂得宜（向宫向星零神）；收山未得，坐宫山星不当令",
			"收山、拨水入零堂俱未得，宜择时改向并验实际砂水",
		},
	},
	reflect.TypeFor[ChartDigest](): {Type: "string", Pattern: "^[a-f0-9]{64}$"},
	reflect.TypeFor[YinYangName](): {Type: "string", Enum: []any{"阳", "阴"}},
	reflect.TypeFor[DirectionName](): {
		Type: "string", Enum: []any{"北", "东北", "东", "东南", "南", "西南", "西", "西北"},
	},
	reflect.TypeFor[LifeGroupName]():    {Type: "string", Enum: []any{"东四命", "西四命"}},
	reflect.TypeFor[YearBoundaryName](): {Type: "string", Enum: []any{"gregorian_calendar_year"}},
	reflect.TypeFor[FlyingStarName](): {
		Type: "string", Enum: []any{
			"一白贪狼", "二黑巨门", "三碧禄存", "四绿文曲", "五黄廉贞",
			"六白武曲", "七赤破军", "八白左辅", "九紫右弼",
		},
	},
	reflect.TypeFor[YouxingName](): {
		Type: "string", Enum: []any{"生气", "天医", "延年", "伏位", "祸害", "五鬼", "六煞", "绝命"},
	},
	reflect.TypeFor[BazhaiRatingName](): {
		Type: "string", Enum: []any{"大吉", "吉", "平", "凶", "大凶"},
	},
	reflect.TypeFor[TrigramGroupName](): {Type: "string", Enum: []any{"东四卦", "西四卦"}},
	reflect.TypeFor[BazhaiMatchName]():  {Type: "string", Enum: []any{"吉", "凶"}},
	reflect.TypeFor[LiuyaoQiguaMode]():  {Type: "string", Enum: []any{"auto", "coins", "yaos"}},
	reflect.TypeFor[LiuyaoCastMode]():   {Type: "string", Enum: []any{"coins", "yaos"}},
	reflect.TypeFor[LiuyaoOrderName]():  {Type: "string", Enum: []any{"bottom_up"}},
	reflect.TypeFor[LiuyaoCoinSymbol](): {
		Type: "string", Enum: []any{"正", "反"},
	},
	reflect.TypeFor[LiuyaoHeadsValue](): {Type: "integer", Enum: []any{3}},
	reflect.TypeFor[LiuyaoTailsValue](): {Type: "integer", Enum: []any{2}},
	reflect.TypeFor[LiuyaoYaoValue](): {
		Type: "integer", Minimum: jsonschema.Ptr(float64(6)), Maximum: jsonschema.Ptr(float64(9)),
	},
	reflect.TypeFor[LiuyaoPosition](): {
		Type: "integer", Minimum: jsonschema.Ptr(float64(1)), Maximum: jsonschema.Ptr(float64(6)),
	},
	reflect.TypeFor[LiuyaoOptionalPosition](): {
		Type: "integer", Minimum: jsonschema.Ptr(float64(0)), Maximum: jsonschema.Ptr(float64(6)),
	},
	reflect.TypeFor[LiuyaoDongRelationType](): {
		Type: "string", Enum: []any{
			"生用", "克用", "比和", "冲用", "生原神", "克原神", "生忌神", "克忌神",
		},
	},
	reflect.TypeFor[LiuyaoYaoLabel](): {
		Type: "string", Enum: []any{"老阴", "少阳", "少阴", "老阳"},
	},
	reflect.TypeFor[LiuyaoSchemaVersion](): {Type: "string", Enum: []any{"liuyao-cast-v1"}},
	reflect.TypeFor[LiuyaoCastingID]():     {Type: "string", Pattern: "^[a-f0-9]{64}$"},
	reflect.TypeFor[LiuyaoGanName](): {
		Type: "string", Enum: []any{"甲", "乙", "丙", "丁", "戊", "己", "庚", "辛", "壬", "癸"},
	},
	reflect.TypeFor[LiuyaoZhiName](): {
		Type: "string", Enum: []any{"子", "丑", "寅", "卯", "辰", "巳", "午", "未", "申", "酉", "戌", "亥"},
	},
	reflect.TypeFor[LiuyaoWangShuai](): {
		Type: "string", Enum: []any{"旺", "相", "休", "囚", "死"},
	},
	reflect.TypeFor[LiuyaoShiYing](): {Type: "string", Enum: []any{"世", "应", ""}},
	reflect.TypeFor[LiuyaoLiuQinName](): {
		Type: "string", Enum: []any{"父母", "兄弟", "官鬼", "妻财", "子孙"},
	},
	reflect.TypeFor[LiuyaoLiuShouName](): {
		Type: "string", Enum: []any{"青龙", "朱雀", "勾陈", "螣蛇", "白虎", "玄武"},
	},
	reflect.TypeFor[LiuyaoMuKuType](): {
		Type: "string", Enum: []any{"day", "moving", "transformed"},
	},
	reflect.TypeFor[LiuyaoPatternType](): {
		Type: "string", Enum: []any{
			"旬空", "月破", "飞伏", "进退", "冲合", "反吟", "随鬼入墓", "独发", "独静", "两现",
		},
	},
	reflect.TypeFor[LiuyaoForceRole](): {
		Type: "string", Enum: []any{"用神", "原神", "忌神", "仇神"},
	},
	reflect.TypeFor[LiuyaoConflictID](): {
		Type: "string", Enum: []any{
			"moving-support-opposition", "yuanshen-support-opposition", "strong-but-void", "strong-but-break",
		},
	},
	reflect.TypeFor[LiuyaoTombBranch](): {Type: "string", Enum: []any{"辰"}},
	reflect.TypeFor[LiuyaoTombSchool](): {Type: "string", Enum: []any{"engine_default_chen"}},
	reflect.TypeFor[LiuyaoYongShenName](): {
		Type: "string", Enum: []any{"父母", "兄弟", "官鬼", "妻财", "子孙", "世爻", "应爻"},
	},
	reflect.TypeFor[QimenYongShenSymbols](): {
		Type:        "array",
		Items:       &jsonschema.Schema{Type: "string"},
		MinItems:    jsonschema.Ptr(1),
		UniqueItems: true,
	},
}}

// registerSixItemArray registers a fixed-length six-element array schema for a
// DTO slice type, deriving the item contract from the element DTO. Go slices
// cannot express the length constraint on their own, and the item schema stays
// DTO-owned instead of being spelled out here.
func registerSixItemArray[Item, Slice any]() {
	items, err := jsonschema.For[Item](forOptions)
	if err != nil {
		panic(fmt.Sprintf("derive %T item schema: %v", *new(Item), err))
	}
	forOptions.TypeSchemas[reflect.TypeFor[Slice]()] = &jsonschema.Schema{
		Type: "array", MinItems: jsonschema.Ptr(6), MaxItems: jsonschema.Ptr(6), Items: items,
	}
}

// Fixed-length six-coin and six-line arrays: input and output carry distinct
// DTOs for the same wire shape, so each type gets its own registration.
func init() {
	registerSixItemArray[LiuyaoCoinRoundInput, LiuyaoCoinRoundsInput]()
	registerSixItemArray[LiuyaoYaoValue, LiuyaoYaoValuesInput]()
	registerSixItemArray[LiuyaoCoinRoundOutput, LiuyaoCoinRoundsOutput]()
	registerSixItemArray[LiuyaoYaoValue, LiuyaoYaoValuesOutput]()
}

func mustRegister[In, Out any](registry *Registry, tool Tool[In, Out]) {
	if err := register(registry, tool); err != nil {
		panic(fmt.Sprintf("register engine MCP contract %s: %v", tool.Name, err))
	}
}

// NewEngine returns the engine's typed MCP contract registry.
func NewEngine() *Registry {
	registry := New()
	if err := liuyao.Load(); err != nil {
		panic(fmt.Sprintf("load liuyao tables for MCP contracts: %v", err))
	}
	hexagramNames := liuyao.HexagramNames()
	hexagramEnum := make([]any, 0, len(hexagramNames))
	for _, name := range hexagramNames {
		hexagramEnum = append(hexagramEnum, name)
	}
	forOptions.TypeSchemas[reflect.TypeFor[LiuyaoHexagramName]()] = &jsonschema.Schema{
		Type: "string", Enum: hexagramEnum,
	}
	mustRegister(registry, Tool[TimeNowInput, TimeNowOutput]{
		Name:        "time.now",
		Description: "获取服务端当前时间。返回 UTC、本地、北京时间，用于 AI agent 获取准确时间避免幻觉。",
		Handler:     TimeNow,
	})
	mustRegister(registry, Tool[HuangliDaysInput, []HuangliDayOutput]{
		Name:        "huangli.days",
		Description: "黄历查日。返回连续N天的黄历信息；event 给定时按建除事项规则输出 suitability。",
		Handler:     HuangliDays,
	})
	mustRegister(registry, Tool[TianwenTimeInput, TianwenTimeOutput]{
		Name:        "tianwen.time",
		Description: "根据时间和经度计算真太阳时，返回公历、真太阳时、农历三套时间。",
		Handler:     TianwenTime,
	})
	mustRegister(registry, Tool[CityCoordsInput, CityCoordsOutput]{
		Name:        "city.coords",
		Description: "根据城市名查询经纬度。优先使用内置城市表；未命中时可查询 Nominatim（OSM），部署方可用 LIKI_EXTERNAL_GEOCODING=off 禁用。",
		Handler:     CityCoords,
	})
	mustRegister(registry, Tool[BaziChartInput, BaziChart]{
		Name:                "bazi.chart",
		Description:         "排八字命盘。返回最小命盘（四柱+纳音+大运+性别）。",
		Handler:             BaziChartCompute,
		TolerateUnknownArgs: true,
	})
	mustRegister(registry, Tool[BazhaiChartInput, BazhaiChartOutput]{
		Name:        "bazhai.chart",
		Description: "八宅排盘。按出生年份与性别排命卦 + 四吉四凶方 + 出生年紫白飞星。八宅不需要出生时辰，也不与八字四柱合参。",
		Handler:     BazhaiChart,
	})
	mustRegister(registry, Tool[BazhaiLayoutInput, BazhaiLayoutOutput]{
		Name:        "bazhai.layout",
		Description: "八宅门主灶配合。命卦 + 门/主/灶卦 → 方向、游年九星、吉凶。确定性计算。",
		Handler:     BazhaiLayout,
	})
	mustRegister(registry, Tool[XuankongChartInput, XuankongChart]{
		Name:        "xuankong.chart",
		Description: "玄空飞星。period_date 为宅运起盘日期（如建成/入住/改宅日期），不是命主出生时间。zuo_shan/xiang_shan 为坐向（0-23）。",
		Handler:     XuankongChartCompute,
	})
	mustRegister(registry, Tool[XuankongLiunianInput, XuankongLiunianOutput]{
		Name:        "xuankong.liunian",
		Description: "玄空流年飞星。chart（可选）+ year → 流年飞星盘 + 宅盘凶星落宫对照（确定性计算）。",
		Handler:     XuankongLiunian,
	})
	mustRegister(registry, Tool[LiuyaoQiguaInput, LiuyaoQiguaOutput]{
		Name:        "liuyao.qigua",
		Description: "六爻起卦。mode=auto 时用安全随机三枚铜钱起六次；mode=coins 时传入六组正/反记录；mode=yaos 时传入六个爻值。返回可审计起卦收据。",
		Handler:     LiuyaoQigua,
	})
	mustRegister(registry, Tool[LiuyaoChartInput, LiuyaoChartOutput]{
		Name:        "liuyao.chart",
		Description: "六爻装卦。只接受 liuyao.qigua 返回的完整 casting 收据；装卦并分析：纳甲、六亲、六兽、用神、旺衰、应期。",
		Handler:     LiuyaoChart,
	})
	mustRegister(registry, Tool[BaziFullChartInput, BaziFullChartOutput]{
		Name:                "bazi.fullchart",
		Description:         "完整命盘。传入 bazi.chart 返回的最小命盘，补全十神/藏干/神煞/自坐/空亡/自合/魁罡/三元/拱夹/纳音生克/长生十二宫/三奇贵人/合会冲刑/十干禄/命理原子事实。",
		Handler:             BaziFullChart,
		TolerateUnknownArgs: true,
	})
	mustRegister(registry, Tool[BaziBondInput, BaziBondOutput]{
		Name:                "bazi.bond",
		Description:         "八字合盘原始事实。返回双方日主、四柱干支关系、十神视角、纳音、实盘五行与配偶星/夫妻宫事实；不做合婚评级。",
		Handler:             BaziBond,
		TolerateUnknownArgs: true,
	})
	mustRegister(registry, Tool[BaziLiuNianInput, BaziLiuNianOutput]{
		Name:                "bazi.liunian",
		Description:         "八字流年运势。返回流年干支与命局的十神、神煞、伏吟反吟。",
		Handler:             BaziLiuNian,
		TolerateUnknownArgs: true,
	})
	mustRegister(registry, Tool[BaziLiuYueInput, BaziLiuYueOutput]{
		Name:                "bazi.liuyue",
		Description:         "流月运势。返回流月干支与命局的十神、神煞。",
		Handler:             BaziLiuYue,
		TolerateUnknownArgs: true,
	})
	mustRegister(registry, Tool[BaziLiuRiInput, BaziLiuRiOutput]{
		Name:                "bazi.liuri",
		Description:         "流日运势。返回流日干支、十神、纳音。",
		Handler:             BaziLiuRi,
		TolerateUnknownArgs: true,
	})
	mustRegister(registry, Tool[BaziLiuShiInput, BaziLiuShiOutput]{
		Name:                "bazi.liushi",
		Description:         "流时运势。返回流时干支、十神。hour 为时辰（0-23）。",
		Handler:             BaziLiuShi,
		TolerateUnknownArgs: true,
	})
	mustRegister(registry, Tool[BaziXiaoYunInput, BaziXiaoYunOutput]{
		Name:                "bazi.xiaoyun",
		Description:         "小运。返回小运流年列表。count 默认 12。",
		Handler:             BaziXiaoYun,
		TolerateUnknownArgs: true,
	})
	mustRegister(registry, Tool[ZiweiChartInput, ZiweiChartOutput]{
		Name:                "ziwei.chart",
		Description:         "紫微斗数排盘。返回十二宫星曜分布、亮度、四化。",
		Handler:             ZiweiChartCompute,
		TolerateUnknownArgs: true,
	})
	mustRegister(registry, Tool[ZiweiFullChartInput, ZiweiFullChartOutput]{
		Name:                "ziwei.fullchart",
		Description:         "紫微全盘。扩展杂曜、长生、博士、小限、将前、岁前。",
		Handler:             ZiweiFullChart,
		TolerateUnknownArgs: true,
	})
	mustRegister(registry, Tool[ZiweiDaXianInput, ZiweiDaXianOutput]{
		Name:                "ziwei.daxian",
		Description:         "紫微斗数大限。返回十年大限各宫吉凶。chart 为 ziwei.chart 返回的完整 chart 对象。",
		Handler:             ZiweiDaXian,
		TolerateUnknownArgs: true,
	})
	mustRegister(registry, Tool[ZiweiLiuNianInput, ZiweiLiuNianOutput]{
		Name:                "ziwei.liunian",
		Description:         "紫微流年。返回流年命盘及各宫变化。",
		Handler:             ZiweiLiuNian,
		TolerateUnknownArgs: true,
	})
	mustRegister(registry, Tool[ZiweiLiuYueInput, ZiweiLiuYueOutput]{
		Name:                 "ziwei.liuyue",
		Description:          "紫微流月。target_lunar 必须是已解析的完整农历日期；闰月十六日起按下一流月。",
		Handler:              ZiweiLiuYue,
		TolerateUnknownArgs:  true,
		KeepClosedInputPaths: []string{"", "/properties/target_lunar"},
	})
	mustRegister(registry, Tool[ZiweiLiuRiInput, ZiweiLiuRiOutput]{
		Name:                 "ziwei.liuri",
		Description:          "紫微流日。target_lunar 必须是已解析的完整农历日期。",
		Handler:              ZiweiLiuRi,
		TolerateUnknownArgs:  true,
		KeepClosedInputPaths: []string{"", "/properties/target_lunar"},
	})
	mustRegister(registry, Tool[ZiweiLiuShiInput, ZiweiLiuShiOutput]{
		Name:                 "ziwei.liushi",
		Description:          "紫微流时。返回流时命宫及四化。",
		Handler:              ZiweiLiuShi,
		TolerateUnknownArgs:  true,
		KeepClosedInputPaths: []string{"", "/properties/target_lunar"},
	})
	mustRegister(registry, Tool[ZiweiBondInput, ZiweiBondOutput]{
		Name:                "ziwei.bond",
		Description:         "紫微合盘原始事实。并排返回双方命宫、夫妻宫、子女宫主星；不做跨盘投射和合婚评级。",
		Handler:             ZiweiBond,
		TolerateUnknownArgs: true,
	})
	mustRegister(registry, Tool[QimenChartInput, QimenChartOutput]{
		Name:                 "qimen.chart",
		Description:          "奇门排盘：时/刻/日/月/年家 × 转盘/飞盘/鸣法；时家与日家支持拆补/置闰；刻家包含十分钟三元与十二分钟十分局；金函玉镜为 day 专用盘。yong_shen 可显式定位用神，事象路由由上层 skill 完成；ying_qi 返回日期窗口。solar_time 必须来自 tianwen.time 的真太阳时。",
		Handler:              QimenChart,
		TolerateUnknownArgs:  true,
		KeepClosedInputPaths: []string{""},
	})
	return registry
}
