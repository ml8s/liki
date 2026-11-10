package mcpcontract

// This file is the engine's typed tool catalogue: the wire vocabulary the
// schemas derive from (forOptions), the registration point (NewEngine), and the
// helper the registrations share. The registry mechanism itself lives in
// registry.go, and each tool's DTOs live beside the tool they describe.

import (
	"encoding/json"
	"fmt"
	"reflect"
	"sync"

	"github.com/google/jsonschema-go/jsonschema"
	"liki-engine/internal/engine/liuyao"
)

// buildStaticSchemaOptions builds wire-level constraints that Go struct tags
// cannot express through this library. Field-level descriptions still come from
// the DTO's `jsonschema` tags.
func buildStaticSchemaOptions() map[reflect.Type]*jsonschema.Schema {
	return map[reflect.Type]*jsonschema.Schema{
		// A few legacy wire fields carry more than one JSON type (a pillar 六破
		// reference is a name in some charts and an index in others). Those fields
		// keep their exact JSON value instead of being reshaped by a Go type.
		reflect.TypeFor[json.RawMessage](): {},
		reflect.TypeFor[StandardTime]():    {Type: "string", Format: "date-time"},
		reflect.TypeFor[DateString]():      {Type: "string", Format: "date"},
		reflect.TypeFor[Longitude]():       {Type: "number", Minimum: jsonschema.Ptr(float64(-180)), Maximum: jsonschema.Ptr(float64(180))},
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
			Type: "integer", Minimum: jsonschema.Ptr(float64(1)), Maximum: jsonschema.Ptr(float64(MaxSeriesCount)),
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
		reflect.TypeFor[GanName](): {
			Type: "string", Enum: []any{"甲", "乙", "丙", "丁", "戊", "己", "庚", "辛", "壬", "癸"},
		},
		reflect.TypeFor[ZhiName](): {
			Type: "string", Enum: []any{"子", "丑", "寅", "卯", "辰", "巳", "午", "未", "申", "酉", "戌", "亥"},
		},
		reflect.TypeFor[WangShuaiName](): {
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
	}
}

var (
	schemaOptionsOnce sync.Once
	schemaOptions     *jsonschema.ForOptions
	schemaOptionsErr  error
)

func cachedSchemaOptions() (*jsonschema.ForOptions, error) {
	schemaOptionsOnce.Do(func() {
		if err := liuyao.Load(); err != nil {
			schemaOptionsErr = fmt.Errorf("load liuyao tables for MCP contracts: %w", err)
			return
		}
		schemaOptions = buildSchemaOptions(liuyao.HexagramNames())
	})
	return schemaOptions, schemaOptionsErr
}

func buildSchemaOptions(liuyaoHexagrams []string) *jsonschema.ForOptions {
	options := &jsonschema.ForOptions{TypeSchemas: buildStaticSchemaOptions()}
	registerWireEnums(options)
	registerSixItemArray[LiuyaoCoinRoundInput, LiuyaoCoinRoundsInput](options)
	registerSixItemArray[LiuyaoYaoValue, LiuyaoYaoValuesInput](options)
	registerSixItemArray[LiuyaoCoinRoundOutput, LiuyaoCoinRoundsOutput](options)
	registerSixItemArray[LiuyaoYaoValue, LiuyaoYaoValuesOutput](options)
	daXianItem, err := jsonschema.For[ZiweiDaXianItem](options)
	if err != nil {
		panic(fmt.Sprintf("derive ZiweiDaXianItem schema: %v", err))
	}
	options.TypeSchemas[reflect.TypeFor[ZiweiDaXianItems]()] = &jsonschema.Schema{
		Type:     "array",
		Items:    daXianItem,
		MinItems: jsonschema.Ptr(12),
		MaxItems: jsonschema.Ptr(12),
	}
	hexagramEnum := make([]any, 0, len(liuyaoHexagrams))
	for _, name := range liuyaoHexagrams {
		hexagramEnum = append(hexagramEnum, name)
	}
	options.TypeSchemas[reflect.TypeFor[LiuyaoHexagramName]()] = &jsonschema.Schema{
		Type: "string", Enum: hexagramEnum,
	}
	return options
}

// registerSixItemArray registers a fixed-length six-element array schema for a
// DTO slice type, deriving the item contract from the element DTO. Go slices
// cannot express the length constraint on their own, and the item schema stays
// DTO-owned instead of being spelled out here.
func registerSixItemArray[Item, Slice any](options *jsonschema.ForOptions) {
	items, err := jsonschema.For[Item](options)
	if err != nil {
		panic(fmt.Sprintf("derive %T item schema: %v", *new(Item), err))
	}
	options.TypeSchemas[reflect.TypeFor[Slice]()] = &jsonschema.Schema{
		Type: "array", MinItems: jsonschema.Ptr(6), MaxItems: jsonschema.Ptr(6), Items: items,
	}
}

func mustRegister[In, Out any](registry *Registry, tool Tool[In, Out]) {
	if err := register(registry, tool); err != nil {
		panic(fmt.Sprintf("register engine MCP contract %s: %v", tool.Name, err))
	}
}

// mustSeriesSchema publishes the named domain series root. Out is one item,
// not the transport wrapper.
func mustSeriesSchema[Out any](options *jsonschema.ForOptions, root string) json.RawMessage {
	raw, err := seriesOutputSchema[Out](options, root)
	if err != nil {
		panic(fmt.Sprintf("build %s series output schema: %v", root, err))
	}
	return raw
}

// NewEngine returns the engine's typed MCP contract registry.
func NewEngine() *Registry {
	options, err := cachedSchemaOptions()
	if err != nil {
		panic(err)
	}
	registry := New(options)
	mustRegister(registry, Tool[TimeNowInput, TimeNowOutput]{
		Name:        "time_now",
		Description: "按 IANA 时区返回服务端当前标准时间（无 DST）。所有「现在/今天/当前年份」等运算必须以本返回值为时间基准，禁止使用模型自拟或猜测的时间；用户出生时间等事实输入不属于本工具职责。",
		Handler:     TimeNow,
	})
	mustRegister(registry, Tool[HuangliDaysInput, HuangliDaysOutput]{
		Name:        "huangli_days",
		Description: "黄历择日查日。用户询问某事项宜忌日期、某日吉凶或黄历宜忌事实时调用；需要日期范围与事项名称。返回连续日期的黄历事实，指定事项时附适配结论与判定依据。命理合参（八字/紫微本命、流年）不走本工具，改用对应排盘与查询工具。",
		Handler:     HuangliDays,
	})
	resolveSchema, err := tianwenResolveFormsSchema(options)
	if err != nil {
		panic("tianwen.subject_candidates input schema: " + err.Error())
	}
	mustRegister(registry, Tool[TianwenSubjectCandidatesInput, TianwenSubjectCandidatesOutput]{
		Name:        "subject_candidates",
		InputSchema: resolveSchema,
		Description: "把民用时间事实归一化为 canonical subject 候选：接受精确时间点、当地壁钟（公历/农历，分钟或秒精度）、时辰、同日区间，可带地点经度范围。时间明确且无歧义时，客户端可按 subject schema 直接组装排盘输入；时间模糊或触及时辰、日界、交节、夏令时边界时，先调用本工具、向用户展示 reasons 并等待确认。不得静默选择候选，也不得编造缺失时间或地点。",
		Handler:     TianwenSubjectCandidates,
	})
	mustRegister(registry, Tool[CityInfoInput, CityInfoOutput]{
		Name:        "city_info",
		Description: "根据城市名返回城市信息候选（内置表，不访问外链）：每项含名称、行政区、层级、经纬度、IANA 时区与可选经度范围。长度 1 可直接使用；多于 1（同名区县等）必须由用户确认，不得静默取第一个。未命中时请询问附近大城市，或让用户直接提供经纬度和时区。",
		Handler:     CityInfo,
	})
	mustRegisterRaw(registry, RawTool[BaziChartInput, BaziChartOutput]{
		Name:        "bazi_chart",
		Description: "排八字确定性命盘。subject 可由用户明确事实组装，也可取自用户确认后的 subject_candidates 候选；时间模糊或多候选时先归一化并确认。用户询问本命（性格、婚姻、事业、财运、健康、学业、六亲等）时调用。返回四柱、纳音、大运、十神、藏干、神煞、合会冲刑、扶抑候选、调候候选、格局结构证据与命盘关系事实。",
		Handler:     BaziChartCompute,
	})
	mustRegister(registry, Tool[BazhaiChartInput, BazhaiChartOutput]{
		Name:        "bazhai_chart",
		Description: "八宅排盘：由已校准 subject 推导出生年与性别，排命卦、四吉四凶方与出生年紫白飞星。住宅风水评估、命卦方位吉凶类问题调用；后续门/主/灶方位配合改用 bazhai_layout。八宅不需要出生时辰，也不与八字四柱合参。",
		Handler:     BazhaiChart,
	})
	mustRegister(registry, Tool[BazhaiLayoutInput, BazhaiLayoutOutput]{
		Name:        "bazhai_layout",
		Description: "八宅门主灶配合：输入命卦与门/主/灶卦，输出各自方向、游年九星与吉凶结论。已由 bazhai_chart 排出命卦、需评估大门/主卧/灶位方位配合时调用；仅覆盖八宅法，玄空飞星另用 xuankong 工具。",
		Handler:     BazhaiLayout,
	})
	mustRegister(registry, Tool[XuankongChartInput, XuankongChart]{
		Name:        "xuankong_chart",
		Description: "玄空飞星起盘：按宅体事实日期（建成、入住或改宅）与二十四山坐向，排九宫飞星盘。评估阳宅固定风水格局时调用；流年飞星变化改用 xuankong_liunian。注意起盘日期不是命主出生时间。",
		Handler:     XuankongChartCompute,
	})
	mustRegister(registry, Tool[XuankongLiunianInput, XuankongLiuNianSeries]{
		Name:         "xuankong_liunian",
		Description:  "玄空流年飞星：在宅运盘基础上叠加指定年份的流年飞星，count=1（默认）返回单年，count>1 返回批量。评估阳宅某一年风水变化时调用；尚未起宅运盘的先调用 xuankong_chart。",
		Handler:      XuankongLiunian,
		OutputSchema: mustSeriesSchema[XuankongLiunianOutput](options, "liu_nian"),
	})
	qiguaSchema, err := liuyaoQiguaFormsSchema(options)
	if err != nil {
		panic("liuyao_qigua input schema: " + err.Error())
	}
	mustRegister(registry, Tool[LiuyaoQiguaInput, LiuyaoQiguaOutput]{
		Name:        "liuyao_qigua",
		Description: "六爻起卦。mode=auto 时用安全随机三枚铜钱起六次，每次调用都是新占卜事件；mode=coins/yaos 可复现既有卦。返回可审计起卦收据。",
		InputSchema: qiguaSchema,
		Handler:     LiuyaoQigua,
	})
	mustRegister(registry, Tool[LiuyaoChartInput, LiuyaoChartOutput]{
		Name:        "liuyao_chart",
		Description: "六爻装卦。在 liuyao_qigua 起卦后、需要解读卦象吉凶或判断应期时调用。只接受 liuyao_qigua 返回的完整起卦收据；装卦并分析：纳甲、六亲、六兽、用神、旺衰、应期。",
		Handler:     LiuyaoChart,
	})
	mustRegisterRaw(registry, RawTool[BaziBondInput, BaziBondOutput]{
		Name:        "bazi_bond",
		Description: "八字合盘原始事实。用户提供双方出生信息、询问关系配适度（合婚、合作、亲子等）时调用。返回双方日主、四柱干支关系、十神视角、纳音、实盘五行与配偶星/夫妻宫事实；不做合婚评级。需要紫微宫位主星对照时改用 ziwei_bond，两者常配合使用。",
		Handler:     BaziBond,
	})
	baziLiupanInput, err := baziLiupanFormsSchema(options)
	if err != nil {
		panic("bazi.liupan input schema: " + err.Error())
	}
	baziLiupanOutput, err := baziLiupanOutputSchema(options)
	if err != nil {
		panic("bazi.liupan output schema: " + err.Error())
	}
	mustRegisterRaw(registry, RawTool[BaziLiupanInput, BaziLiupanOutput]{
		Name:         "bazi_liupan",
		Description:  "八字流盘。scope 显式选择层级：liu_nian 返回 year/shi_shen/gan/zhi/gan_wuxing/na_yin/sheng_ri/ke_ri/shen_sha/fu_yin_fan_yin/da_yun_interactions/natal_interactions/nian_jie；liu_yue 返回 shi_shen/gan/zhi/gan_wuxing/sheng_ri/ke_ri/shen_sha/gan_rels/zhi_rels/jie_qi；liu_ri 返回 date/shi_shen/gan/zhi/na_yin/shen_sha/gan_rels/zhi_rels/da_yun_rels/liu_nian_rels/ri_jie；liu_shi 返回 hour_range/shi_shen/gan/zhi/gan_rels/zhi_rels/shi_chen；xiao_yun 返回 method/basis/zhus。standard_time 为当地标准时间，liu_ri/liu_shi 必须带经度；count=1（默认）返回对应根数组，count>1 返回连续批量。",
		InputSchema:  baziLiupanInput,
		OutputSchema: baziLiupanOutput,
		Handler:      BaziLiupan,
	})
	mustRegisterRaw(registry, RawTool[ZiweiChartInput, ZiweiChartOutput]{
		Name:        "ziwei_chart",
		Description: "紫微斗数确定性命盘。subject 可由用户明确事实组装，也可取自用户确认后的 subject_candidates 候选；时间模糊或多候选时先归一化并确认。需要十二宫细节（宫位主星、亮度、四化、大限等）时调用。返回十二宫星曜、亮度、四化、杂曜、长生、博士、小限、将前、岁前与大限。",
		Handler:     ZiweiChartCompute,
	})
	ziweiLiupanInput, err := ziweiLiupanFormsSchema(options)
	if err != nil {
		panic("ziwei.liupan input schema: " + err.Error())
	}
	ziweiLiupanOutput, err := ziweiLiupanOutputSchema(options)
	if err != nil {
		panic("ziwei.liupan output schema: " + err.Error())
	}
	mustRegisterRaw(registry, RawTool[ZiweiLiupanInput, ZiweiLiupanOutput]{
		Name:         "ziwei_liupan",
		Description:  "紫微流盘。scope 显式选择层级：liu_nian 返回 ming_gong/zhi/si_hua/gong_wei/nian_jie；liu_yue 返回 yue_jie/ming_gong/zhi/si_hua/gong_wei；liu_ri 返回 ri_jie/ming_gong/zhi/si_hua/gong_wei；liu_shi 返回 ming_gong/zhi/si_hua/gong_wei/shi_chen。standard_time 为当地标准时间，liu_shi 必须带经度；count=1（默认）返回对应根数组，count>1 返回连续批量。",
		InputSchema:  ziweiLiupanInput,
		OutputSchema: ziweiLiupanOutput,
		Handler:      ZiweiLiupan,
	})
	mustRegisterRaw(registry, RawTool[ZiweiBondInput, ZiweiBondOutput]{
		Name:        "ziwei_bond",
		Description: "紫微合盘原始事实。用户询问两人关系并需要紫微侧证据（命宫、夫妻宫、子女宫主星对比）时调用；八字干支与十神视角改用 bazi_bond，两者常配合使用。并排返回双方命宫、夫妻宫、子女宫主星；不做跨盘投射和合婚评级。",
		Handler:     ZiweiBond,
	})
	qimenSchema, err := qimenChartFormsSchema(options)
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
