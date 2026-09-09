package agent

import (
	"context"
	_ "embed"
	"encoding/json"
	"fmt"
	"time"

	"liki-engine/internal/agent/city"
	"liki-engine/internal/engine/bazhai"
	"liki-engine/internal/engine/ganzhi"
	"liki-engine/internal/engine/huangli"
	"liki-engine/internal/engine/liuyao"
	"liki-engine/internal/engine/qimen"
	"liki-engine/internal/engine/tianwen"
	"liki-engine/internal/engine/xuankong"
)

// ── qimen ──

func qimenChartHandler(ctx context.Context, raw json.RawMessage) (json.RawMessage, error) {
	var p struct {
		SolarTime        string   `json:"solar_time"`
		Scope            string   `json:"scope"`
		School           string   `json:"school"`
		DingjuMethod     string   `json:"dingju_method"`
		QuarterRule      string   `json:"quarter_rule"`
		BaseDingjuMethod string   `json:"base_dingju_method"`
		DunSource        string   `json:"dun_source"`
		HourBoundary     string   `json:"hour_boundary"`
		YongShen         []string `json:"yong_shen"`
		BirthDate        string   `json:"birth_date"`
	}
	if err := json.Unmarshal(raw, &p); err != nil {
		return nil, fmt.Errorf("qimen.chart: %w", err)
	}
	st, err := parseSolarTime(p.SolarTime)
	if err != nil {
		return nil, fmt.Errorf("qimen.chart: %w", err)
	}
	method, err := qimen.ResolveMethodWithOptions(
		p.Scope, p.School, p.DingjuMethod, p.QuarterRule,
		p.BaseDingjuMethod, p.DunSource, p.HourBoundary,
	)
	if err != nil {
		return nil, fmt.Errorf("qimen.chart: %w", err)
	}
	if method.School == qimen.SchoolJinhanYuJing {
		if len(p.YongShen) > 0 || p.BirthDate != "" {
			return nil, fmt.Errorf("qimen.chart: jinhan_yujing does not support yong_shen or birth_date")
		}
		return wrapResult("qimen", qimen.ComputeJinhanChart(st))
	}
	var birthDate qimen.BirthDate
	if p.BirthDate != "" {
		birthDate, err = qimen.ParseBirthDate(p.BirthDate)
		if err != nil {
			return nil, fmt.Errorf("qimen.chart: %w", err)
		}
	}
	if len(p.YongShen) > 0 || birthDate.Has {
		syms := make([]qimen.YongShenSymbol, 0, len(p.YongShen))
		for _, name := range p.YongShen {
			sym, err := qimen.ParseYongShen(name)
			if err != nil {
				return nil, fmt.Errorf("qimen.chart: %w", err)
			}
			syms = append(syms, sym)
		}
		result, err := qimen.ComputeChartWithYongShenAndMethod(st, syms, birthDate, method)
		if err != nil {
			return nil, fmt.Errorf("qimen.chart: %w", err)
		}
		return wrapResult("qimen", result)
	}
	result, err := qimen.ComputeChartWithMethod(st, method)
	if err != nil {
		return nil, fmt.Errorf("qimen.chart: %w", err)
	}
	return wrapResult("qimen", result)
}

// ── bazhai ──

func bazhaiLayoutHandler(ctx context.Context, raw json.RawMessage) (json.RawMessage, error) {
	var p struct {
		MingGua   string `json:"ming_gua"`
		DoorGua   string `json:"door_gua"`
		MasterGua string `json:"master_gua"`
		StoveGua  string `json:"stove_gua"`
	}
	if err := json.Unmarshal(raw, &p); err != nil {
		return nil, fmt.Errorf("bazhai.layout: %w", err)
	}
	result := bazhai.ComputeLayout(p.MingGua, p.DoorGua, p.MasterGua, p.StoveGua)
	return wrapResult("bazhai_layout", result)
}

func bazhaiChartHandler(ctx context.Context, raw json.RawMessage) (json.RawMessage, error) {
	var p struct {
		BirthYear int           `json:"birth_year"`
		Gender    ganzhi.Gender `json:"gender"`
	}
	if err := json.Unmarshal(raw, &p); err != nil {
		return nil, fmt.Errorf("bazhai.chart: %w", err)
	}
	if err := validateGender(p.Gender); err != nil {
		return nil, fmt.Errorf("bazhai.chart: %w", err)
	}
	if p.BirthYear < 1900 || p.BirthYear > 2100 {
		return nil, fmt.Errorf("bazhai.chart: birth_year must be 1900-2100, got %d", p.BirthYear)
	}
	result := bazhai.ComputeChart(p.Gender, p.BirthYear)
	return wrapResult("bazhai", result)
}

// ── xuankong ──

func xuankongChartHandler(ctx context.Context, raw json.RawMessage) (json.RawMessage, error) {
	var p struct {
		PeriodDate   string `json:"period_date"`
		SitMountain  *int   `json:"zuo_shan"`
		FaceMountain *int   `json:"xiang_shan"`
	}
	if err := json.Unmarshal(raw, &p); err != nil {
		return nil, fmt.Errorf("xuankong.chart: %w", err)
	}
	if p.SitMountain == nil {
		return nil, fmt.Errorf("xuankong.chart: zuo_shan is required")
	}
	if p.FaceMountain == nil {
		return nil, fmt.Errorf("xuankong.chart: xiang_shan is required")
	}
	if *p.SitMountain < 0 || *p.SitMountain > 23 {
		return nil, fmt.Errorf("xuankong.chart: zuo_shan must be 0-23, got %d", *p.SitMountain)
	}
	if *p.FaceMountain < 0 || *p.FaceMountain > 23 {
		return nil, fmt.Errorf("xuankong.chart: xiang_shan must be 0-23, got %d", *p.FaceMountain)
	}
	periodDate, err := time.Parse("2006-01-02", p.PeriodDate)
	if err != nil {
		return nil, fmt.Errorf("xuankong.chart: invalid period_date %q: %w", p.PeriodDate, err)
	}
	result := xuankong.ComputeChart(tianwen.SolarTime(periodDate), *p.SitMountain, *p.FaceMountain)
	return wrapResult("xuankong", result)
}

func xuankongLiunianHandler(ctx context.Context, raw json.RawMessage) (json.RawMessage, error) {
	var p struct {
		Chart json.RawMessage `json:"chart"` // 可选：宅盘
		Year  int             `json:"year"`
	}
	if err := json.Unmarshal(raw, &p); err != nil {
		return nil, fmt.Errorf("xuankong.liunian: %w", err)
	}
	if p.Year < 1864 || p.Year > 2200 {
		return nil, fmt.Errorf("xuankong.liunian: year %d out of range (1864-2200)", p.Year)
	}
	var chart *xuankong.Chart
	if len(p.Chart) > 0 && string(p.Chart) != "null" {
		var c xuankong.Chart
		if err := json.Unmarshal(p.Chart, &c); err != nil {
			return nil, fmt.Errorf("xuankong.liunian: parse chart: %w", err)
		}
		chart = &c
	}
	result := xuankong.ComputeLiuNian(p.Year, chart)
	return wrapResult("xuankong_liunian", result)
}

// ── liuyao ──

func liuyaoQiguaHandler(ctx context.Context, raw json.RawMessage) (json.RawMessage, error) {
	var p struct {
		Seed   *int64     `json:"seed,omitempty"` // 可选：固定随机种子（测试用）
		Mode   string     `json:"mode,omitempty"`
		Rounds [][]string `json:"rounds,omitempty"`
		Yaos   []int      `json:"yaos,omitempty"`
	}
	if err := json.Unmarshal(raw, &p); err != nil {
		return nil, fmt.Errorf("liuyao.qigua: %w", err)
	}

	var result liuyao.QiguaResult
	var casting liuyao.Casting
	var err error
	mode := p.Mode
	if mode == "" {
		mode = "auto"
	}
	switch mode {
	case "auto":
		if p.Rounds != nil || p.Yaos != nil {
			return nil, fmt.Errorf("liuyao.qigua: auto mode does not accept rounds or yaos")
		}
		if p.Seed != nil {
			result = liuyao.QiguaWithSeed(*p.Seed)
			casting, err = liuyao.NewValuesCasting(result.Yaos, "yaos", "deterministic_seed")
			if err != nil {
				return nil, fmt.Errorf("liuyao.qigua: %w", err)
			}
		} else {
			casting, err = liuyao.SecureQigua()
			if err != nil {
				return nil, fmt.Errorf("liuyao.qigua: %w", err)
			}
			result.Yaos = casting.Yaos
			result.DongYao = casting.DongYao
		}
	case "coins":
		if p.Seed != nil || p.Yaos != nil {
			return nil, fmt.Errorf("liuyao.qigua: coins mode accepts rounds only")
		}
		casting, err = liuyao.NewCoinsCasting(p.Rounds, "external_manual")
		if err != nil {
			return nil, fmt.Errorf("liuyao.qigua: %w", err)
		}
		result.Yaos = casting.Yaos
		result.DongYao = casting.DongYao
	case "yaos":
		if p.Seed != nil || p.Rounds != nil {
			return nil, fmt.Errorf("liuyao.qigua: yaos mode accepts yaos only")
		}
		var values [6]int
		copy(values[:], p.Yaos)
		casting, err = liuyao.NewValuesCasting(values, "yaos", "external_user_values")
		if err != nil {
			return nil, fmt.Errorf("liuyao.qigua: %w", err)
		}
		result.Yaos = casting.Yaos
		result.DongYao = casting.DongYao
	default:
		return nil, fmt.Errorf("liuyao.qigua: unsupported mode %q", mode)
	}

	return wrapResult("liuyao_qigua", struct {
		liuyao.QiguaResult
		Casting liuyao.Casting `json:"casting"`
	}{
		QiguaResult: result,
		Casting:     casting,
	})
}

func liuyaoChartHandler(ctx context.Context, raw json.RawMessage) (json.RawMessage, error) {
	var p struct {
		SolarTime string          `json:"solar_time"`
		YongShen  string          `json:"yong_shen"`
		Yaos      [6]int          `json:"yaos"`
		Casting   *liuyao.Casting `json:"casting"`
	}
	if err := json.Unmarshal(raw, &p); err != nil {
		return nil, fmt.Errorf("liuyao.chart: %w", err)
	}
	var yaos [6]int
	castingMode := "legacy_yaos"
	if p.Casting != nil {
		if p.Yaos != ([6]int{}) {
			return nil, fmt.Errorf("liuyao.chart: casting and yaos are mutually exclusive")
		}
		if err := p.Casting.Validate(); err != nil {
			return nil, fmt.Errorf("liuyao.chart: invalid casting: %w", err)
		}
		yaos = p.Casting.Yaos
		castingMode = p.Casting.Mode
	} else {
		yaos = p.Yaos
		if yaos == ([6]int{}) {
			return nil, fmt.Errorf("liuyao.chart: casting or yaos is required")
		}
	}
	for i, v := range yaos {
		if v < 6 || v > 9 {
			return nil, fmt.Errorf("liuyao.chart: yao[%d] = %d, must be 6-9", i, v)
		}
	}
	st, err := parseSolarTime(p.SolarTime)
	if err != nil {
		return nil, fmt.Errorf("liuyao.chart: %w", err)
	}
	if p.YongShen == "" {
		p.YongShen = "世爻"
	}
	ys, err := liuyao.ParseYongShen(p.YongShen)
	if err != nil {
		return nil, fmt.Errorf("liuyao.chart: %w", err)
	}
	result := liuyao.ComputeChart(st, ys, yaos)
	result.CastingMode = castingMode
	if p.Casting != nil {
		result.Casting = p.Casting
	} else if legacy, castErr := liuyao.NewValuesCasting(yaos, "legacy_yaos", "legacy_rpc_yaos"); castErr == nil {
		result.Casting = &legacy
	}
	return wrapResult("liuyao", result)
}

// ── huangli ──

func huangliDaysHandler(ctx context.Context, raw json.RawMessage) (json.RawMessage, error) {
	var p struct {
		StartDate string `json:"start_date"`
		Count     int    `json:"count"`
	}
	if err := json.Unmarshal(raw, &p); err != nil {
		return nil, fmt.Errorf("huangli.days: %w", err)
	}
	start, err := time.Parse("2006-01-02", p.StartDate)
	if err != nil {
		return nil, fmt.Errorf("huangli.days: parse date: %w", err)
	}
	if p.Count < 1 {
		p.Count = 3
	}
	if p.Count > 30 {
		p.Count = 30
	}
	days := make([]huangli.Day, 0, p.Count)
	for i := 0; i < p.Count; i++ {
		dateStr := start.AddDate(0, 0, i).Format("2006-01-02")
		entry, err := huangli.QueryDate(dateStr)
		if err != nil {
			return nil, fmt.Errorf("huangli.days: %w", err)
		}
		days = append(days, entry)
	}
	return wrapResult("huangli_days", days)
}

// ── time / infra ──

func cityCoordsHandler(ctx context.Context, raw json.RawMessage) (json.RawMessage, error) {
	result, err := city.SearchCoords(ctx, raw)
	if err != nil {
		return nil, err
	}
	return wrapResult("city_coords", result)
}

func timeNowHandler(ctx context.Context, raw json.RawMessage) (json.RawMessage, error) {
	now := time.Now()
	result := struct {
		UTC   string `json:"utc"`
		Local string `json:"local"`
		CST   string `json:"cst"`
	}{
		UTC:   now.UTC().Format(time.RFC3339),
		Local: now.Format(time.RFC3339),
		CST:   now.In(time.FixedZone("CST", 8*3600)).Format(time.RFC3339),
	}
	return wrapResult("time_now", result)
}

func schemaTimePointParams() json.RawMessage {
	return mustSchema(`{"type":"object","properties":{"time":{"type":"string","format":"date-time","description":"出生时间（RFC3339），如 1984-02-04T06:00:00+08:00"},"longitude":{"type":"number","description":"出生地经度，用于真太阳时校正。北京≈116.4"}},"required":["time","longitude"]}`)
}

//go:embed qimen_schema.json
var qimenSchemaJSON []byte

var qimenSchemas = mustLoadQimenSchemas()

type qimenSchemaDocument struct {
	Params json.RawMessage `json:"params"`
	Result json.RawMessage `json:"result"`
}

func mustLoadQimenSchemas() qimenSchemaDocument {
	var doc qimenSchemaDocument
	if err := json.Unmarshal(qimenSchemaJSON, &doc); err != nil {
		panic("invalid qimen schema document: " + err.Error())
	}
	if len(doc.Params) == 0 || len(doc.Result) == 0 {
		panic("qimen schema document lacks params or result")
	}
	return doc
}

const schemaDoorStoveItem = `{"type":"object","properties":{"direction":{"type":"string","description":"方位，如西北"},"gua_name":{"type":"string","enum":["坎","坤","震","巽","乾","兑","艮","离"]},"wuxing":{"type":"string","enum":["水","土","木","金","火"]},"youxing":{"type":"string","enum":["生气","天医","延年","伏位","祸害","五鬼","六煞","绝命"]},"rating":{"type":"string","enum":["大吉","吉","平","凶","大凶"]},"group":{"type":"string","enum":["东四宅","西四宅"]},"match":{"type":"string","enum":["吉","凶"]}},"required":["direction","gua_name","wuxing","youxing","rating","group","match"]}`

var otherMethods = []RPCMethod{
	{
		Name: "qimen.chart", Description: "奇门排盘：时/刻/日/月/年家 × 转盘/飞盘/鸣法；时家与日家支持拆补/置闰；刻家包含十分钟三元与十二分钟十分局；金函玉镜为 day 专用盘。yong_shen 可显式定位用神，事象路由由上层 skill 完成；ying_qi 返回日期窗口。solar_time 必须来自 tianwen.time 的真太阳时。",
		Params:  qimenSchemas.Params,
		Handler: qimenChartHandler,
		Result:  envelopeSchema(string(qimenSchemas.Result)),
	},
	{
		Name: "bazhai.chart", Description: "八宅排盘。按出生年份与性别排命卦 + 四吉四凶方 + 流年紫白飞星。八宅不需要出生时辰，也不与八字四柱合参。",
		Params:  mustSchema(`{"type":"object","properties":{"birth_year":{"type":"integer","minimum":1900,"maximum":2100,"description":"出生公历年份"},"gender":{"type":"string","enum":["male","female"]}},"required":["birth_year","gender"]}`),
		Handler: bazhaiChartHandler,
		Result:  envelopeSchema(`{"type":"object","properties":{"ming_gua":{"type":"object"},"ba_zhai_dirs":{"type":"object"},"liu_nian_xing":{"type":"object","description":"流年紫白飞星（与玄空共用 schema：year/ru_zhong/gong_wei）"}},"required":["ming_gua","ba_zhai_dirs","liu_nian_xing"]}`),
	},
	{
		Name: "bazhai.layout", Description: "八宅门主灶配合。命卦 + 门/主/灶卦 → 方向、游年九星、吉凶。确定性计算。",
		Params:  mustSchema(`{"type":"object","properties":{"ming_gua":{"type":"string","enum":["坎","坤","震","巽","乾","兑","艮","离"],"description":"命卦"},"door_gua":{"type":"string","enum":["坎","坤","震","巽","乾","兑","艮","离"],"description":"门卦"},"master_gua":{"type":"string","enum":["坎","坤","震","巽","乾","兑","艮","离"],"description":"主卧卦"},"stove_gua":{"type":"string","enum":["坎","坤","震","巽","乾","兑","艮","离"],"description":"灶卦"}},"required":["ming_gua","door_gua","master_gua","stove_gua"]}`),
		Handler: bazhaiLayoutHandler,
		Result:  envelopeSchema(`{"type":"object","properties":{"group":{"type":"string","enum":["东四宅","西四宅"]},"ming_gua":{"type":"string","enum":["坎","坤","震","巽","乾","兑","艮","离"]},"door":` + schemaDoorStoveItem + `,"master":` + schemaDoorStoveItem + `,"stove":` + schemaDoorStoveItem + `},"required":["group","ming_gua","door","master","stove"]}`),
	},
	{
		Name: "xuankong.chart", Description: "玄空飞星。period_date 为宅运起盘日期（如建成/入住/改宅日期），不是命主出生时间。zuo_shan/xiang_shan 为坐向（0-23）。",
		Params:  mustSchema(`{"type":"object","properties":{"period_date":{"type":"string","format":"date","description":"宅运起盘日期，YYYY-MM-DD"},"zuo_shan":{"type":"integer","minimum":0,"maximum":23,"description":"山向"},"xiang_shan":{"type":"integer","minimum":0,"maximum":23,"description":"朝向"}},"required":["period_date","zuo_shan","xiang_shan"]}`),
		Handler: xuankongChartHandler,
		Result:  envelopeSchema(`{"type":"object","properties":{"yun":{"type":"object","description":"三元九运","properties":{"year":{"type":"integer","description":"当前年份"},"yuan":{"type":"string","description":"上元/中元/下元"},"yun_number":{"type":"integer","description":"运数1-9"},"yun_name":{"type":"string","description":"运名:一运/九运"},"start_year":{"type":"integer","description":"本运起始年"}},"required":["year","yuan","yun_number"]},"gong_wei":{"type":"array"},"wang_shan":{"type":"boolean"},"zuo_shan":{"type":"integer","description":"坐山(0-23)"},"xiang_shan":{"type":"integer","description":"向山(0-23)"},"shan_xing":{"type":"boolean","description":"双星会坐：坐宫山向星皆当令"},"wang_xiang":{"type":"boolean","description":"旺向：向宫向星=当令"},"xiang_xing":{"type":"boolean","description":"双星会向：向宫山向星皆当令"},"xing_jia_hui":{"type":"array","description":"星加会"},"shou_shan_chu_sha":{"type":"object","description":"收山出煞"},"fu_yin":{"type":"boolean","description":"伏吟（运盘）"},"fan_yin":{"type":"boolean","description":"反吟（运盘，恒false）"},"xia_shui":{"type":"boolean","description":"上山下水：向宫山星=当令且坐宫向星=当令"}},"required":["yun","gong_wei","wang_shan"]}`),
	},
	{
		Name: "xuankong.liunian", Description: "玄空流年飞星。chart（可选）+ year → 流年飞星盘 + 宅盘凶星落宫对照（确定性计算）。",
		Params:  mustSchema(`{"type":"object","properties":{"chart":{"type":"object","description":"可选：xuankong.chart 返回的宅盘，给则叠加凶星落宫对照"},"year":{"type":"integer","description":"年份"}},"required":["year"]}`),
		Handler: xuankongLiunianHandler,
		Result:  envelopeSchema(`{"type":"object","properties":{"year":{"type":"integer"},"ru_zhong":{"type":"string","description":"入中星名"},"gong_wei":{"type":"array","description":"流年飞星盘（gong_num/xing/xing_name/wuxing/rating/ru_zhong）"},"house_overlay":{"type":"array","description":"流年凶星落宫对照宅盘该宫三星"}},"required":["year","ru_zhong","gong_wei"]}`),
	},
	{
		Name: "liuyao.qigua", Description: "六爻起卦。mode=auto 时用安全随机三枚铜钱起六次；mode=coins 时传入六组正/反记录；mode=yaos 时传入六个爻值。返回原始爻值、动爻和可审计起卦收据。",
		Params:  mustSchema(`{"type":"object","properties":{"mode":{"type":"string","enum":["auto","coins","yaos"],"default":"auto"},"seed":{"type":"integer","description":"仅 auto 模式：固定随机种子（测试用）"},"rounds":{"type":"array","minItems":6,"maxItems":6,"description":"coins 模式：六组三枚硬币，初爻到上爻","items":{"type":"array","minItems":3,"maxItems":3,"items":{"type":"string","enum":["正","反"]}}},"yaos":{"type":"array","minItems":6,"maxItems":6,"items":{"type":"integer","minimum":6,"maximum":9},"description":"yaos 模式：六爻值，初爻到上爻"}},"required":[]}`),
		Handler: liuyaoQiguaHandler,
		Result:  envelopeSchema(`{"type":"object","properties":{"yaos":{"type":"array","items":{"type":"integer"}, "description":"六爻值 6-9"},"dong_yao":{"type":"array","items":{"type":"integer"},"description":"动爻位置 1-6"},"casting":{"type":"object","description":"完整起卦收据，可回传 liuyao.chart"}},"required":["yaos","dong_yao","casting"]}`),
	},
	{
		Name: "liuyao.chart", Description: "六爻装卦。优先传入 liuyao.qigua 返回的 casting；兼容旧 yaos 输入。装卦并分析：纳甲、六亲、六兽、用神、旺衰、应期。lines.liu_qin: 0=父母 1=兄弟 2=官鬼 3=妻财 4=子孙；lines.liu_shou: 0=青龙 1=朱雀 2=勾陈 3=螣蛇 4=白虎 5=玄武；wang_shuai: 0=旺 1=相 2=休 3=囚 4=死",
		Params:  mustSchema(`{"type":"object","properties":{"solar_time":` + schemaSolarTime + `,"yong_shen":{"type":"string","description":"用神六亲（如 妻财/官鬼/父母/兄弟/子孙/世爻），可选，默认世爻"},"yaos":{"type":"array","items":{"type":"integer"},"minItems":6,"maxItems":6,"description":"兼容旧输入：六爻值（6-9）；与 casting 互斥"},"casting":{"type":"object","description":"liuyao.qigua 返回的完整起卦收据；与 yaos 互斥"}},"required":["solar_time"]}`),
		Handler: liuyaoChartHandler,
		Result:  envelopeSchema(`{"type":"object","properties":{"name":{"type":"string"},"ben_gua":{"type":"string","enum":["乾","姤","遁","否","观","晋","大有","剥","复","颐","屯","益","震","噬嗑","随","无妄","明夷","贲","既济","家人","丰","离","革","同人","临","损","节","中孚","归妹","睽","兑","履","泰","大畜","需","小畜","大壮","大有","夬","乾","姤","遁","否","观","晋","大有","剥","复","颐","屯","益","震","噬嗑","随","无妄","明夷","贲","既济","家人","丰","离","革","同人"]},"lines":{"type":"array","description":"每爻：六亲/六神/世应 + 确定性状态（yue_po 月破/dong_self 发动/dong_sheng 动爻生/dong_ke 动爻克）","items":{"type":"object","properties":{"position":{"type":"integer"},"type":{"type":"integer"},"gan":{"type":"string"},"zhi":{"type":"string"},"wuxing":{"type":"string"},"liu_qin":{"type":"string","enum":["父母","兄弟","官鬼","妻财","子孙"]},"shi_ying":{"type":"string","description":"世/应"},"liu_shou":{"type":"string","enum":["青龙","朱雀","勾陈","螣蛇","白虎","玄武"]},"yue_po":{"type":"boolean","description":"月破"},"dong_self":{"type":"boolean","description":"本爻发动"},"dong_sheng":{"type":"boolean","description":"有动爻生此爻"},"dong_ke":{"type":"boolean","description":"有动爻克此爻"},"xun_kong":{"type":"boolean","description":"该爻地支值日柱旬空"},"mu_ku_branch":{"type":"string"},"mu_ku_element":{"type":"string"},"chang_sheng_yue":{"type":"string"}},"required":["position","type","gan","zhi","wuxing","liu_qin","shi_ying","liu_shou"]}},"yong_shen":{"type":"object","description":"用神结果","properties":{"name":{"type":"string","description":"用神六亲"},"position":{"type":"integer","description":"爻位1-6（0=未找到）"},"wang_shuai":{"type":"string","description":"用神旺衰"},"yue_po":{"type":"boolean","description":"用神月破"},"xun_kong":{"type":"boolean","description":"用神旬空"},"mu_ku":{"type":"boolean","description":"用神入墓"},"liu_shou":{"type":"string","description":"用神临的六神","enum":["青龙","朱雀","勾陈","螣蛇","白虎","玄武"]},"fu_shen":{"type":"object","description":"飞伏（用神不现时）","properties":{"position":{"type":"integer","description":"爻位"},"liu_qin":{"type":"string","description":"伏神六亲"},"zhi":{"type":"string","description":"伏神地支"}},"required":["position","liu_qin","zhi"]},"chang_sheng":{"type":"string"}},"required":["name","position"]},"wang_shuai":{"type":"array"},"yue_jian_zhi":{"type":"string","enum":["子","丑","寅","卯","辰","巳","午","未","申","酉","戌","亥"]},"yue_jian_gan":{"type":"string","enum":["甲","乙","丙","丁","戊","己","庚","辛","壬","癸"]},"ying_qi":{"type":"object","description":"应期判断结果","properties":{"yong_shen":{"type":"string","description":"用神"},"dong_yao_pos":{"type":"integer","description":"动爻位置"},"ying_time":{"type":"string","description":"应期描述"},"assessment":{"type":"string","description":"综合判断"}},"required":["yong_shen","assessment"]},"bian_gua":{"type":"string","description":"变卦名"},"bian_yao":{"type":"array","description":"变爻"},"dong_yao":{"type":"array","items":{"type":"integer"},"description":"动爻位置"},"gong":{"type":"string","description":"八宫"},"gua_ci":{"type":"object","description":"卦辞爻辞"},"gong_wuxing":{"type":"string","description":"宫五行"},"ri_chen_gan":{"type":"string","description":"日辰天干"},"ri_chen_zhi":{"type":"string","description":"日辰地支"},"ri_chen_relations":{"type":"array","description":"日辰与爻关系"},"xun_kong":{"type":"array","items":{"type":"string","enum":["子","丑","寅","卯","辰","巳","午","未","申","酉","戌","亥"]},"description":"日柱旬空地支（甲子旬空戌亥…）"},"dong_yao_relations":{"type":"array","description":"动爻与用神的关系","items":{"type":"object","properties":{"position":{"type":"integer","description":"动爻位置"},"relation":{"type":"string","description":"关系类型（生用/克用/比和/冲用/生原神/克原神/生忌神/克忌神）"}},"required":["position","relation"]}},"patterns":{"type":"array","description":"特殊格局","items":{"type":"object","properties":{"type":{"type":"string","description":"格局类型"},"sub_type":{"type":"string","description":"子类型"},"position":{"type":"integer"},"is_true":{"type":"boolean","description":"结构是否有实质效力；空亡/月破中 true=真空/真破，false=假空/假破"},"assessment":{"type":"string"}},"required":["type","assessment"]}},"casting":{"type":"object","description":"完整起卦收据"},"casting_mode":{"type":"string","description":"起卦输入模式"},"timing_candidates":{"type":"array","description":"应期触发机制候选；非确定日期","items":{"type":"object","properties":{"id":{"type":"string"},"mechanism":{"type":"string"},"position":{"type":"integer"},"trigger_branch":{"type":"string","enum":["子","丑","寅","卯","辰","巳","午","未","申","酉","戌","亥"]},"window":{"type":"string"},"basis":{"type":"array","items":{"type":"string"}},"condition":{"type":"string"},"confidence":{"type":"string"}},"required":["id","mechanism","confidence"]}},"day_clash_facts":{"type":"array","description":"日冲细分事实","items":{"type":"object","properties":{"position":{"type":"integer"},"kind":{"type":"string"},"basis":{"type":"array","items":{"type":"string"}},"condition":{"type":"string"}},"required":["position","kind"]}},"moving_transformations":{"type":"array","description":"动爻到变爻的机械变化","items":{"type":"object","properties":{"position":{"type":"integer"},"from_branch":{"type":"string"},"to_branch":{"type":"string"},"return_relation":{"type":"string"},"advance_retreat":{"type":"string"},"changed_void":{"type":"boolean"},"changed_break":{"type":"boolean"},"basis":{"type":"array","items":{"type":"string"}}},"required":["position","from_branch","to_branch"]}},"force_chain":{"type":"object","properties":{"yong_shen":{"type":"string"},"yong_element":{"type":"string"},"position":{"type":"integer"},"is_hidden":{"type":"boolean"},"entries":{"type":"array","items":{"type":"object","properties":{"role":{"type":"string"},"element":{"type":"string"},"positions":{"type":"array","items":{"type":"integer"}},"branches":{"type":"array","items":{"type":"string"}},"relations":{"type":"array","items":{"type":"string"}},"states":{"type":"array","items":{"type":"string"}},"note":{"type":"string"}},"required":["role","element"]}}},"required":["yong_shen","yong_element","position","entries"]},"yong_shen_candidates":{"type":"array","description":"用神候选与当前确定性选择","items":{"type":"object","properties":{"position":{"type":"integer"},"branch":{"type":"string"},"is_hidden":{"type":"boolean"},"selected":{"type":"boolean"},"wang_shuai":{"type":"string"},"reason":{"type":"string"},"basis":{"type":"array","items":{"type":"string"}}},"required":["position","branch","is_hidden","selected","reason"]}},"san_he_candidates":{"type":"array","properties":{"moving_count":{"type":"integer"},"static_count":{"type":"integer"},"void_count":{"type":"integer"},"break_count":{"type":"integer"},"positions":{"type":"array","items":{"type":"integer"}},"branches":{"type":"array","items":{"type":"string"}},"element":{"type":"string"},"complete":{"type":"boolean"},"qualification":{"type":"string"},"line_complete":{"type":"boolean"},"day_present":{"type":"boolean"},"month_present":{"type":"boolean"},"missing":{"type":"array","items":{"type":"string"}},"activated":{"type":"boolean"},"targets":{"type":"array","items":{"type":"string"}},"conclusion_scope":{"type":"string"}},"required":["positions","branches","element","complete","moving_count","qualification","line_complete","activated","conclusion_scope"],"description":"三合候选；complete=true 只是三支齐，成局仍须按发动与空破月日复核","items":{"type":"object","properties":{"moving_count":{"type":"integer"},"static_count":{"type":"integer"},"void_count":{"type":"integer"},"break_count":{"type":"integer"},"positions":{"type":"array","items":{"type":"integer"}},"branches":{"type":"array","items":{"type":"string"}},"element":{"type":"string"},"complete":{"type":"boolean"},"qualification":{"type":"string"},"line_complete":{"type":"boolean"},"day_present":{"type":"boolean"},"month_present":{"type":"boolean"},"missing":{"type":"array","items":{"type":"string"}},"activated":{"type":"boolean"},"targets":{"type":"array","items":{"type":"string"}},"conclusion_scope":{"type":"string"}},"required":["positions","branches","element","complete","moving_count","qualification","line_complete","activated","conclusion_scope"]}},"hidden_lines":{"type":"array","description":"本宫全部伏神","items":{"type":"object","properties":{"position":{"type":"integer"},"liu_qin":{"type":"string"},"gan":{"type":"string"},"zhi":{"type":"string"},"wuxing":{"type":"string"},"flying_position":{"type":"integer"},"flying_liu_qin":{"type":"string"},"flying_gan":{"type":"string"},"flying_zhi":{"type":"string"},"flying_wuxing":{"type":"string"},"fei_fu_relation":{"type":"string"},"xun_kong":{"type":"boolean"},"yue_po":{"type":"boolean"},"mu_ku":{"type":"boolean"},"exit_condition":{"type":"string"}},"required":["position","liu_qin","gan","zhi","wuxing","flying_position","flying_liu_qin","fei_fu_relation"]}},"branch_relation_facts":{"type":"array","description":"爻间地支关系","items":{"type":"object","properties":{"left_position":{"type":"integer"},"right_position":{"type":"integer"},"relations":{"type":"array","items":{"type":"string"}},"targets":{"type":"array","items":{"type":"string"}},"basis":{"type":"array","items":{"type":"string"}}},"required":["left_position","right_position","relations"]}}},"required":["name","ben_gua","lines","yong_shen"]}`),
	},
	{
		Name: "huangli.days", Description: "黄历查日。返回连续N天的黄历信息（建除、黄道、二十八宿、时辰吉凶等）。",
		Params:  mustSchema(`{"type":"object","properties":{"start_date":{"type":"string","description":"起始日期 YYYY-MM-DD"},"count":{"type":"integer","description":"查几天，默认3，最多30"}},"required":["start_date"]}`),
		Handler: huangliDaysHandler,
		Result:  envelopeSchema(`{"type":"array","items":{"type":"object","properties":{"date":{"type":"string"},"ri_zhu":{"type":"object","properties":{"gan":{"type":"string","enum":["甲","乙","丙","丁","戊","己","庚","辛","壬","癸"]},"zhi":{"type":"string","enum":["子","丑","寅","卯","辰","巳","午","未","申","酉","戌","亥"]},"nayin":{"type":"string"}},"required":["gan","zhi"]},"nayin":{"type":"string"},"wuxing":{"type":"string","enum":["金","木","水","火","土"]},"jian_chu":{"type":"string","enum":["建","除","满","平","定","执","破","危","成","收","开","闭"]},"huangdao":{"type":"object","properties":{"name":{"type":"string"},"path":{"type":"string"}},"required":["name","path"]},"xi_shen":{"type":"string"},"cai_shen":{"type":"string"},"fu_shen":{"type":"string"},"gan_ji":{"type":"string"},"zhi_ji":{"type":"string"},"mansion":{"type":"object","properties":{"name":{"type":"string"},"index":{"type":"integer"},"animal":{"type":"string"},"wuxing":{"type":"string","enum":["金","木","水","火","土","日","月"]},"group":{"type":"string"}},"required":["name","index"]}},"required":["date","ri_zhu"]}}`),
	},
	{
		Name: "time.now", Description: "获取服务端当前时间。返回 UTC、本地、北京时间，用于 AI agent 获取准确时间避免幻觉。",
		Params:  mustSchema(`{"type":"object","properties":{"seed":{"type":"integer","description":"可选：固定随机种子（测试用）"}},"required":[]}`),
		Handler: timeNowHandler,
		Result:  envelopeSchema(`{"type":"object","properties":{"utc":{"type":"string"},"local":{"type":"string"},"cst":{"type":"string"}},"required":["utc","cst"]}`),
	},
	{
		Name: "tianwen.time", Description: "根据时间和经度计算真太阳时，返回公历、真太阳时、农历三套时间。",
		Params:  schemaTimePointParams(),
		Handler: tianwenTimeHandler,
		Result:  envelopeSchema(`{"type":"object","properties":{"solar":{"type":"string","format":"date-time"},"gregorian":{"type":"string","format":"date-time"},"lunar":{"type":"object","description":"农历信息: year/month/day/shichen","properties":{"year":{"type":"integer"},"month":{"type":"integer"},"day":{"type":"integer"},"leap":{"type":"boolean"},"shichen":{"type":"string"}}}},"required":["solar","gregorian","lunar"]}`),
	},
	{
		Name: "city.coords", Description: "根据城市名查询经纬度。支持中英文城市名，全球范围搜索。基于 Nominatim 服务。",
		Params:  mustSchema(`{"type":"object","properties":{"city":{"type":"string","description":"城市名称（中英文均可）"}},"required":["city"]}`),
		Handler: cityCoordsHandler,
		Result:  envelopeSchema(`{"type":"object","properties":{"name":{"type":"string"},"longitude":{"type":"number"},"latitude":{"type":"number"},"country":{"type":"string"}},"required":["name","longitude","latitude","country"]}`),
	},
}
