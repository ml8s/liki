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
	reflect.TypeFor[StandardTime]():    {Type: "string", Format: "date-time"},
	reflect.TypeFor[DateString]():      {Type: "string", Format: "date"},
	reflect.TypeFor[GenderName]():      {Type: "string", Enum: []any{"male", "female"}},
	reflect.TypeFor[CityName]():        {Type: "string", MinLength: jsonschema.Ptr(1), MaxLength: jsonschema.Ptr(128)},
	reflect.TypeFor[CityLevelName]():   {Type: "string", Enum: []any{"city", "county", "world"}},
	reflect.TypeFor[HuangliEvent](): {
		Type: "string",
		Enum: []any{
			"wedding", "engage", "opening", "sign", "move", "travel", "build",
			"exam", "medical", "sacrifice", "cleaning", "renovation",
			"bed_install", "income", "funeral",
		},
	},
	reflect.TypeFor[HuangliCount](): {Type: "integer", Minimum: jsonschema.Ptr(float64(1)), Maximum: jsonschema.Ptr(float64(30))},
	reflect.TypeFor[SeriesCount](): {
		Type: "integer", Minimum: jsonschema.Ptr(float64(0)), Maximum: jsonschema.Ptr(float64(MaxSeriesCount)),
	},
	reflect.TypeFor[GuaNumber](): {Type: "integer", Minimum: jsonschema.Ptr(float64(1)), Maximum: jsonschema.Ptr(float64(9))},
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
		// Absent candidates are null, not an ambiguous empty-string sentinel.
		Type: "string", Enum: []any{"水", "土", "木", "金", "火"},
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
	reflect.TypeFor[YearBoundaryName](): {Type: "string", Enum: []any{"lichun"}},
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
	reflect.TypeFor[LiuyaoCastingID](): {Type: "string", Pattern: "^[a-f0-9]{64}$"},
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

// mustSeriesSchema marshals the single-output type schema and wraps it as a
// oneOf union with the batch {items:[...]} format for series tools.
func mustSeriesSchema[Out any]() json.RawMessage {
	schema, err := jsonschema.For[Out](forOptions)
	if err != nil {
		panic(fmt.Sprintf("derive series output schema for %T: %v", *new(Out), err))
	}
	single, err := schema.MarshalJSON()
	if err != nil {
		panic(fmt.Sprintf("marshal series output schema for %T: %v", *new(Out), err))
	}
	combined, err := seriesOutputSchema(single)
	if err != nil {
		panic(fmt.Sprintf("build series output schema for %T: %v", *new(Out), err))
	}
	return combined
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
		Name:        "time_now",
		Description: "获取服务端当前时间。返回 UTC、本地、北京时间，用于 AI agent 获取准确时间避免幻觉。",
		Handler:     TimeNow,
	})
	mustRegister(registry, Tool[HuangliDaysInput, []HuangliDayOutput]{
		Name:        "huangli_days",
		Description: "黄历查日。返回连续N天的黄历信息；event 给定时按建除事项规则输出 suitability。",
		Handler:     HuangliDays,
	})
	resolveSchema, err := tianwenResolveFormsSchema()
	if err != nil {
		panic("tianwen.resolve input schema: " + err.Error())
	}
	mustRegister(registry, Tool[TianwenResolveInput, TianwenResolveOutput]{
		Name:        "tianwen_resolve",
		InputSchema: resolveSchema,
		Description: "把时间事实归一化为考时候选：接受精确时间点、当地壁钟（公历/农历，分钟或秒精度）、时辰、同日区间，可带地点经度范围。输出永远包含一个 input 候选（输入的直接换算）；若时间/地点不确定性触及时辰、日界、交节或夏令时边界，追加候选并标注原因领域（时辰/日界/交节/时制）。下游排盘能力使用候选 instant + 经度排盘。",
		Handler:     TianwenResolve,
	})
	mustRegister(registry, Tool[CityCoordsInput, CityCoordsOutput]{
		Name:        "city_coords",
		Description: "根据城市名返回地点候选（内置表，不访问外链）：每项含经纬度、IANA 时区、地点东西经度范围。长度 1 可直接使用；多于 1（同名区县等）必须由用户确认，不得静默取第一个。未命中时请询问附近大城市，或让用户直接提供经纬度和时区。",
		Handler:     CityCoords,
	})
	mustRegister(registry, Tool[BaziChartInput, BaziChart]{
		Name:        "bazi_chart",
		Description: "排八字命盘。返回最小命盘（四柱+纳音+大运+性别）。",
		Handler:     BaziChartCompute,
	})
	mustRegister(registry, Tool[BazhaiChartInput, BazhaiChartOutput]{
		Name:        "bazhai_chart",
		Description: "八宅排盘。按出生年份与性别排命卦 + 四吉四凶方 + 出生年紫白飞星。八宅不需要出生时辰，也不与八字四柱合参。",
		Handler:     BazhaiChart,
	})
	mustRegister(registry, Tool[BazhaiLayoutInput, BazhaiLayoutOutput]{
		Name:        "bazhai_layout",
		Description: "八宅门主灶配合。命卦 + 门/主/灶卦 → 方向、游年九星、吉凶。确定性计算。",
		Handler:     BazhaiLayout,
	})
	mustRegister(registry, Tool[XuankongChartInput, XuankongChart]{
		Name:        "xuankong_chart",
		Description: "玄空飞星。period_date 为宅运起盘日期（如建成/入住/改宅日期），不是命主出生时间。zuo_shan/xiang_shan 为坐向（0-23）。",
		Handler:     XuankongChartCompute,
	})
	mustRegister(registry, Tool[XuankongLiunianInput, any]{
		Name:         "xuankong_liunian",
		Description:  "玄空流年飞星。count=1（默认）返回单年；count>1 返回 {items:[...]} 批量。",
		Handler:      XuankongLiunian,
		OutputSchema: mustSeriesSchema[XuankongLiunianOutput](),
	})
	qiguaSchema, err := liuyaoQiguaFormsSchema()
	if err != nil {
		panic("liuyao.qigua input schema: " + err.Error())
	}
	mustRegister(registry, Tool[LiuyaoQiguaInput, LiuyaoQiguaOutput]{
		Name:        "liuyao_qigua",
		Description: "六爻起卦。mode=auto 时用安全随机三枚铜钱起六次；mode=coins 时传入六组正/反记录；mode=yaos 时传入六个爻值。返回可审计起卦收据。",
		InputSchema: qiguaSchema,
		Handler:     LiuyaoQigua,
	})
	mustRegister(registry, Tool[LiuyaoChartInput, LiuyaoChartOutput]{
		Name:        "liuyao_chart",
		Description: "六爻装卦。只接受 liuyao.qigua 返回的完整 casting 收据；装卦并分析：纳甲、六亲、六兽、用神、旺衰、应期。",
		Handler:     LiuyaoChart,
	})
	mustRegisterRaw(registry, RawTool[BaziFullChartInput, BaziFullChartOutput]{
		Name:        "bazi_fullchart",
		Description: "完整命盘。传入 bazi.chart 返回的最小命盘，补全十神/藏干/神煞/自坐/空亡/自合/魁罡/三元/拱夹/纳音生克/长生十二宫/三奇贵人/合会冲刑/十干禄/命理原子事实。",
		Handler:     BaziFullChart,
	})
	mustRegisterRaw(registry, RawTool[BaziBondInput, BaziBondOutput]{
		Name:        "bazi_bond",
		Description: "八字合盘原始事实。返回双方日主、四柱干支关系、十神视角、纳音、实盘五行与配偶星/夫妻宫事实；不做合婚评级。",
		Handler:     BaziBond,
	})
	mustRegister(registry, Tool[BaziLiuNianInput, any]{
		Name:         "bazi_liunian",
		Description:  "八字流年运势。count=1（默认）返回单年；count>1 返回 {items:[...]} 批量。",
		Handler:      BaziLiuNian,
		OutputSchema: mustSeriesSchema[BaziLiuNianOutput](),
	})
	mustRegister(registry, Tool[BaziLiuYueInput, any]{
		Name:         "bazi_liuyue",
		Description:  "流月运势。count=1（默认）返回单月；count>1 返回 {items:[...]} 批量（跨年进位）。",
		Handler:      BaziLiuYue,
		OutputSchema: mustSeriesSchema[BaziLiuYueOutput](),
	})
	mustRegister(registry, Tool[BaziLiuRiInput, any]{
		Name:         "bazi_liuri",
		Description:  "流日运势。count=1（默认）返回单日；count>1 返回 {items:[...]} 批量（跨月跨年进位）。",
		Handler:      BaziLiuRi,
		OutputSchema: mustSeriesSchema[BaziLiuRiOutput](),
	})
	mustRegister(registry, Tool[BaziLiuShiInput, any]{
		Name:         "bazi_liushi",
		Description:  "流时运势。count=1（默认）返回单时辰；count>1 返回 {items:[...]} 批量（跨日跨月跨年进位）。",
		Handler:      BaziLiuShi,
		OutputSchema: mustSeriesSchema[BaziLiuShiOutput](),
	})
	mustRegister(registry, Tool[BaziXiaoYunInput, any]{
		Name:         "bazi_xiaoyun",
		Description:  "小运。返回小运流年列表。count 默认 1。",
		Handler:      BaziXiaoYun,
		OutputSchema: mustSeriesSchema[BaziXiaoYunOutput](),
	})
	mustRegisterRaw(registry, RawTool[ZiweiChartInput, ZiweiChartOutput]{
		Name:        "ziwei_chart",
		Description: "紫微斗数排盘。返回十二宫星曜分布、亮度、四化。",
		Handler:     ZiweiChartCompute,
	})
	mustRegisterRaw(registry, RawTool[ZiweiFullChartInput, ZiweiFullChartOutput]{
		Name:        "ziwei_fullchart",
		Description: "紫微全盘。扩展杂曜、长生、博士、小限、将前、岁前。",
		Handler:     ZiweiFullChart,
	})
	mustRegister(registry, Tool[ZiweiDaXianInput, ZiweiDaXianOutput]{
		Name:        "ziwei_daxian",
		Description: "紫微斗数大限。返回十年大限各宫吉凶。chart 为 ziwei.chart 返回的完整 chart 对象。",
		Handler:     ZiweiDaXian,
	})
	mustRegister(registry, Tool[ZiweiLiuNianInput, any]{
		Name:         "ziwei_liunian",
		Description:  "紫微流年。count=1（默认）返回单年；count>1 返回 {items:[...]} 批量。",
		Handler:      ZiweiLiuNian,
		OutputSchema: mustSeriesSchema[ZiweiLiuNianOutput](),
	})
	mustRegister(registry, Tool[ZiweiLiuYueInput, any]{
		Name:         "ziwei_liuyue",
		Description:  "紫微流月。count=1（默认）返回单月；count>1 返回 {items:[...]} 批量。",
		Handler:      ZiweiLiuYue,
		OutputSchema: mustSeriesSchema[ZiweiLiuYueOutput](),
	})
	mustRegister(registry, Tool[ZiweiLiuRiInput, any]{
		Name:         "ziwei_liuri",
		Description:  "紫微流日。count=1（默认）返回单日；count>1 返回 {items:[...]} 批量。",
		Handler:      ZiweiLiuRi,
		OutputSchema: mustSeriesSchema[ZiweiLiuRiOutput](),
	})
	mustRegister(registry, Tool[ZiweiLiuShiInput, any]{
		Name:         "ziwei_liushi",
		Description:  "紫微流时。count=1（默认）返回单时辰；count>1 返回 {items:[...]} 批量（12地支循环）。",
		Handler:      ZiweiLiuShi,
		OutputSchema: mustSeriesSchema[ZiweiLiuShiOutput](),
	})
	mustRegisterRaw(registry, RawTool[ZiweiBondInput, ZiweiBondOutput]{
		Name:        "ziwei_bond",
		Description: "紫微合盘原始事实。并排返回双方命宫、夫妻宫、子女宫主星；不做跨盘投射和合婚评级。",
		Handler:     ZiweiBond,
	})
	qimenSchema, err := qimenChartFormsSchema()
	if err != nil {
		panic("qimen.chart input schema: " + err.Error())
	}
	mustRegisterRaw(registry, RawTool[QimenChartInput, QimenChartOutput]{
		Name:        "qimen_chart",
		Description: "奇门排盘。method 显式选择时/刻/日/月/年家与转盘/飞盘/鸣法/金函玉镜的合法组合；时间、地点、用神保持直接参数。",
		InputSchema: qimenSchema,
		Handler:     QimenChart,
	})
	return registry
}
