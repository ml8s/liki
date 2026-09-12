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
	result, err := bazhai.ComputeLayout(p.MingGua, p.DoorGua, p.MasterGua, p.StoveGua)
	if err != nil {
		return nil, fmt.Errorf("bazhai.layout: %w", err)
	}
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
	if (*p.SitMountain+12)%24 != *p.FaceMountain {
		return nil, fmt.Errorf("xuankong.chart: xiang_shan must be opposite zuo_shan by 12 mountains, got %d/%d",
			*p.SitMountain, *p.FaceMountain)
	}
	periodDate, err := time.Parse("2006-01-02", p.PeriodDate)
	if err != nil {
		return nil, fmt.Errorf("xuankong.chart: invalid period_date %q: %w", p.PeriodDate, err)
	}
	if periodDate.Year() < 1864 || periodDate.Year() > 2200 {
		return nil, fmt.Errorf("xuankong.chart: period_date year %d out of range (1864-2200)", periodDate.Year())
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
		if err := c.ValidateChartDigest(); err != nil {
			return nil, fmt.Errorf("xuankong.liunian: %w", err)
		}
		chart = &c
	}
	result := xuankong.ComputeLiuNian(p.Year, chart)
	return wrapResult("xuankong_liunian", result)
}

// ── liuyao ──

// schemaLiuyaoCasting is the complete receipt contract shared by qigua output
// and chart input. Keeping it self-contained avoids a loose generic object.
const schemaLiuyaoCasting = `{"type":"object","additionalProperties":false,"properties":{"schema_version":{"type":"string","const":"liuyao-cast-v1"},"mode":{"type":"string","enum":["coins","yaos"]},"order":{"type":"string","const":"bottom_up"},"coin_convention":{"type":"object","additionalProperties":false,"properties":{"正":{"type":"integer","const":3},"反":{"type":"integer","const":2}},"required":["正","反"]},"rounds":{"type":"array","minItems":6,"maxItems":6,"items":{"type":"object","additionalProperties":false,"properties":{"position":{"type":"integer","minimum":1,"maximum":6},"coins":{"type":"array","minItems":3,"maxItems":3,"items":{"type":"string","enum":["正","反"]}},"value":{"type":"integer","minimum":6,"maximum":9},"label":{"type":"string","enum":["老阴","少阳","少阴","老阳"]},"changing":{"type":"boolean"}},"required":["position","coins","value","label","changing"]}},"yaos":{"type":"array","minItems":6,"maxItems":6,"items":{"type":"integer","minimum":6,"maximum":9}},"dong_yao":{"type":"array","items":{"type":"integer","minimum":1,"maximum":6}},"casting_id":{"type":"string","pattern":"^[a-f0-9]{64}$"}},"required":["schema_version","mode","order","yaos","dong_yao","casting_id"]}`

func liuyaoQiguaHandler(ctx context.Context, raw json.RawMessage) (json.RawMessage, error) {
	var p struct {
		Mode   string     `json:"mode,omitempty"`
		Rounds [][]string `json:"rounds,omitempty"`
		Yaos   []int      `json:"yaos,omitempty"`
	}
	if err := json.Unmarshal(raw, &p); err != nil {
		return nil, fmt.Errorf("liuyao.qigua: %w", err)
	}

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
		casting, err = liuyao.SecureQigua()
		if err != nil {
			return nil, fmt.Errorf("liuyao.qigua: %w", err)
		}
	case "coins":
		if p.Yaos != nil {
			return nil, fmt.Errorf("liuyao.qigua: coins mode accepts rounds only")
		}
		casting, err = liuyao.NewCoinsCasting(p.Rounds)
		if err != nil {
			return nil, fmt.Errorf("liuyao.qigua: %w", err)
		}
	case "yaos":
		if p.Rounds != nil {
			return nil, fmt.Errorf("liuyao.qigua: yaos mode accepts yaos only")
		}
		var values [6]int
		copy(values[:], p.Yaos)
		casting, err = liuyao.NewValuesCasting(values)
		if err != nil {
			return nil, fmt.Errorf("liuyao.qigua: %w", err)
		}
	default:
		return nil, fmt.Errorf("liuyao.qigua: unsupported mode %q", mode)
	}

	return wrapResult("liuyao_qigua", struct {
		Casting liuyao.Casting `json:"casting"`
	}{Casting: casting})
}

func liuyaoChartHandler(ctx context.Context, raw json.RawMessage) (json.RawMessage, error) {
	var p struct {
		SolarTime string          `json:"solar_time"`
		YongShen  string          `json:"yong_shen"`
		Casting   *liuyao.Casting `json:"casting"`
	}
	if err := json.Unmarshal(raw, &p); err != nil {
		return nil, fmt.Errorf("liuyao.chart: %w", err)
	}
	if p.Casting == nil {
		return nil, fmt.Errorf("liuyao.chart: casting is required")
	}
	if err := p.Casting.Validate(); err != nil {
		return nil, fmt.Errorf("liuyao.chart: invalid casting: %w", err)
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
	result := liuyao.ComputeChart(st, ys, p.Casting.Yaos)
	result.Casting = p.Casting
	return wrapResult("liuyao", result)
}

// ── huangli ──

func huangliDaysHandler(ctx context.Context, raw json.RawMessage) (json.RawMessage, error) {
	var p struct {
		StartDate string `json:"start_date"`
		Count     int    `json:"count"`
		Event     string `json:"event"`
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
		entry, err := huangli.QueryDate(dateStr, p.Event)
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

const schemaDoorStoveItem = `{"type":"object","additionalProperties":false,"properties":{"direction":{"type":"string","enum":["北","东北","东","东南","南","西南","西","西北"],"description":"方位"},"gua_name":{"type":"string","enum":["坎","坤","震","巽","乾","兑","艮","离"]},"wuxing":{"type":"string","enum":["水","土","木","金","火"]},"youxing":{"type":"string","enum":["生气","天医","延年","伏位","祸害","五鬼","六煞","绝命"]},"rating":{"type":"string","enum":["大吉","吉","平","凶","大凶"]},"group":{"type":"string","enum":["东四卦","西四卦"]},"match":{"type":"string","enum":["吉","凶"]}},"required":["direction","gua_name","wuxing","youxing","rating","group","match"]}`

const schemaDirectionArray = `{"type":"array","minItems":1,"maxItems":1,"uniqueItems":true,"items":{"type":"string","enum":["北","东北","东","东南","南","西南","西","西北"]}}`

const schemaFlyingStarNameEnum = `{"type":"string","enum":["一白贪狼","二黑巨门","三碧禄存","四绿文曲","五黄廉贞","六白武曲","七赤破军","八白左辅","九紫右弼"]}`

const schemaChartDigest = `{"type":"string","pattern":"^[a-f0-9]{64}$","description":"canonical SHA-256 chart digest"}`

const schemaMountainName = `{"type":"string","enum":["子","癸","丑","艮","寅","甲","卯","乙","辰","巽","巳","丙","午","丁","未","坤","申","庚","酉","辛","戌","乾","亥","壬"]}`

const schemaShouShanAssessment = `{"type":"string","enum":["收山得宜、拨水入零堂，理气合局（完整收山出煞须验实际砂水）","收山得宜（坐宫山星当令）；拨水入零堂未得，向首零神不见水","拨水入零堂得宜（向宫向星零神）；收山未得，坐宫山星不当令","收山、拨水入零堂俱未得，宜择时改向并验实际砂水"]}`

const schemaFlyingStar = `{"type":"object","additionalProperties":false,"properties":{"number":{"type":"integer","minimum":1,"maximum":9},"color":{"type":"string","enum":["白","黑","碧","绿","黄","赤","紫"]},"name":{"type":"string","enum":["一白贪狼","二黑巨门","三碧禄存","四绿文曲","五黄廉贞","六白武曲","七赤破军","八白左辅","九紫右弼"]},"wuxing":{"type":"string","enum":["金","木","水","火","土"]}},"required":["number","color","name","wuxing"]}`

const schemaAnnualFlyingStar = `{"type":"object","additionalProperties":false,"properties":{"gong_num":{"type":"integer","minimum":1,"maximum":9},"xing":{"type":"integer","minimum":1,"maximum":9},"xing_name":{"type":"string","enum":["一白贪狼","二黑巨门","三碧禄存","四绿文曲","五黄廉贞","六白武曲","七赤破军","八白左辅","九紫右弼"]},"wuxing":{"type":"string","enum":["金","木","水","火","土"]},"ru_zhong":{"type":"boolean"}},"required":["gong_num","xing","xing_name","wuxing","ru_zhong"]}`

const schemaXuankongPalace = `{"type":"object","additionalProperties":false,"properties":{"gong_num":{"type":"integer","minimum":1,"maximum":9},"yun_xing":` + schemaFlyingStar + `,"shan_xing":` + schemaFlyingStar + `,"xiang_xing":` + schemaFlyingStar + `},"required":["gong_num","yun_xing","shan_xing","xiang_xing"]}`

const schemaXuankongChartData = `{"type":"object","description":"xuankong.chart 返回的完整宅盘；传回 xuankong.liunian 时 digest 必须匹配","additionalProperties":false,"properties":{"yun":{"type":"object","additionalProperties":false,"properties":{"year":{"type":"integer","minimum":1864,"maximum":2200},"year_boundary":{"type":"string","enum":["gregorian_calendar_year"],"description":"当前实现按公历日历年换运；立春年界属另一流派，不在本盘隐含"},"yuan":{"type":"string","enum":["上元","中元","下元"]},"yun_number":{"type":"integer","minimum":1,"maximum":9},"yun_name":{"type":"string","enum":["一运","二运","三运","四运","五运","六运","七运","八运","九运"]},"start_year":{"type":"integer"},"end_year":{"type":"integer"}},"required":["year","year_boundary","yuan","yun_number","yun_name","start_year","end_year"]},"gong_wei":{"type":"array","minItems":9,"maxItems":9,"items":` + schemaXuankongPalace + `},"four_situation":{"type":"object","additionalProperties":false,"properties":{"name":{"type":"string","enum":["旺山旺向","双星会坐","双星会向","上山下水","未入四大局"]},"sit_mountain_timely":{"type":"boolean"},"face_facing_timely":{"type":"boolean"},"sit_facing_timely":{"type":"boolean"},"face_mountain_timely":{"type":"boolean"}},"required":["name","sit_mountain_timely","face_facing_timely","sit_facing_timely","face_mountain_timely"]},"fu_yin_layers":{"type":"array","items":{"type":"string","enum":["运盘伏吟","山星伏吟","向星伏吟","全盘伏吟"]}},"fan_yin_layers":{"type":"array","items":{"type":"string","enum":["运盘反吟","山星反吟","向星反吟"]}},"xing_jia_hui":{"type":"array","minItems":9,"maxItems":9,"description":"星加会","items":{"type":"object","additionalProperties":false,"properties":{"shan_num":{"type":"integer","minimum":1,"maximum":9},"xiang_num":{"type":"integer","minimum":1,"maximum":9},"name":{"type":"string"},"meaning":{"type":"string"}},"required":["shan_num","xiang_num","name","meaning"]}},"chart_digest":` + schemaChartDigest + `,"zuo_shan":{"type":"integer","minimum":0,"maximum":23,"description":"坐山(0-23)"},"zuo_shan_name":` + schemaMountainName + `,"xiang_shan":{"type":"integer","minimum":0,"maximum":23,"description":"向山(0-23)"},"xiang_shan_name":` + schemaMountainName + `,"shou_shan_chu_sha":{"type":"object","additionalProperties":false,"properties":{"zheng_shen":{"type":"integer","minimum":1,"maximum":9},"ling_shen":{"type":"integer","minimum":1,"maximum":9},"shou_shan":{"type":"boolean"},"chu_sha":{"type":"boolean"},"assessment":` + schemaShouShanAssessment + `},"required":["zheng_shen","ling_shen","shou_shan","chu_sha","assessment"]}},"required":["yun","gong_wei","four_situation","fu_yin_layers","fan_yin_layers","xing_jia_hui","chart_digest","zuo_shan","zuo_shan_name","xiang_shan","xiang_shan_name","shou_shan_chu_sha"]}`

var otherMethods = []RPCMethod{
	{
		Name: "qimen.chart", Description: "奇门排盘：时/刻/日/月/年家 × 转盘/飞盘/鸣法；时家与日家支持拆补/置闰；刻家包含十分钟三元与十二分钟十分局；金函玉镜为 day 专用盘。yong_shen 可显式定位用神，事象路由由上层 skill 完成；ying_qi 返回日期窗口。solar_time 必须来自 tianwen.time 的真太阳时。",
		Params:  qimenSchemas.Params,
		Handler: qimenChartHandler,
		Result:  envelopeSchema(string(qimenSchemas.Result)),
	},
	{
		Name: "bazhai.chart", Description: "八宅排盘。按出生年份与性别排命卦 + 四吉四凶方 + 出生年紫白飞星。八宅不需要出生时辰，也不与八字四柱合参。",
		Params:  mustSchema(`{"type":"object","additionalProperties":false,"properties":{"birth_year":{"type":"integer","minimum":1900,"maximum":2100,"description":"出生公历年份"},"gender":{"type":"string","enum":["male","female"]}},"required":["birth_year","gender"]}`),
		Handler: bazhaiChartHandler,
		Result:  envelopeSchema(`{"type":"object","additionalProperties":false,"properties":{"ming_gua":{"type":"object","additionalProperties":false,"properties":{"gua":{"type":"object","additionalProperties":false,"properties":{"index":{"type":"integer","minimum":1,"maximum":9},"name":{"type":"string","enum":["坎","坤","震","巽","乾","兑","艮","离"]},"wuxing":{"type":"string","enum":["水","土","木","金","火"]},"yin_yang":{"type":"string","enum":["阳","阴"]}},"required":["index","name","wuxing","yin_yang"]},"group":{"type":"string","enum":["东四命","西四命"]},"year_boundary":{"type":"string","enum":["gregorian_calendar_year"]}},"required":["gua","group","year_boundary"]},"ba_zhai_dirs":{"type":"object","additionalProperties":false,"properties":{"sheng_qi":` + schemaDirectionArray + `,"tian_yi":` + schemaDirectionArray + `,"yan_nian":` + schemaDirectionArray + `,"fu_wei":` + schemaDirectionArray + `,"huo_hai":` + schemaDirectionArray + `,"wu_gui":` + schemaDirectionArray + `,"liu_sha":` + schemaDirectionArray + `,"jue_ming":` + schemaDirectionArray + `},"required":["sheng_qi","tian_yi","yan_nian","fu_wei","huo_hai","wu_gui","liu_sha","jue_ming"]},"liu_nian_xing":{"type":"object","additionalProperties":false,"properties":{"year":{"type":"integer","minimum":1900,"maximum":2100},"year_boundary":{"type":"string","enum":["gregorian_calendar_year"]},"ru_zhong":` + schemaFlyingStarNameEnum + `,"gong_wei":{"type":"array","minItems":9,"maxItems":9,"items":` + schemaAnnualFlyingStar + `}},"required":["year","year_boundary","ru_zhong","gong_wei"]}},"required":["ming_gua","ba_zhai_dirs","liu_nian_xing"]}`),
	},
	{
		Name: "bazhai.layout", Description: "八宅门主灶配合。命卦 + 门/主/灶卦 → 方向、游年九星、吉凶。确定性计算。",
		Params:  mustSchema(`{"type":"object","additionalProperties":false,"properties":{"ming_gua":{"type":"string","enum":["坎","坤","震","巽","乾","兑","艮","离"],"description":"命卦"},"door_gua":{"type":"string","enum":["坎","坤","震","巽","乾","兑","艮","离"],"description":"门卦"},"master_gua":{"type":"string","enum":["坎","坤","震","巽","乾","兑","艮","离"],"description":"主卧卦"},"stove_gua":{"type":"string","enum":["坎","坤","震","巽","乾","兑","艮","离"],"description":"灶卦"}},"required":["ming_gua","door_gua","master_gua","stove_gua"]}`),
		Handler: bazhaiLayoutHandler,
		Result:  envelopeSchema(`{"type":"object","additionalProperties":false,"properties":{"group":{"type":"string","enum":["东四命","西四命"]},"ming_gua":{"type":"string","enum":["坎","坤","震","巽","乾","兑","艮","离"]},"door":` + schemaDoorStoveItem + `,"master":` + schemaDoorStoveItem + `,"stove":` + schemaDoorStoveItem + `},"required":["group","ming_gua","door","master","stove"]}`),
	},
	{
		Name: "xuankong.chart", Description: "玄空飞星。period_date 为宅运起盘日期（如建成/入住/改宅日期），不是命主出生时间。zuo_shan/xiang_shan 为坐向（0-23）。",
		Params:  mustSchema(`{"type":"object","additionalProperties":false,"properties":{"period_date":{"type":"string","format":"date","description":"宅运起盘日期，YYYY-MM-DD；年份 1864-2200"},"zuo_shan":{"type":"integer","minimum":0,"maximum":23,"description":"坐山，二十四山 index：0=子,1=癸,2=丑,3=艮,4=寅,5=甲,6=卯,7=乙,8=辰,9=巽,10=巳,11=丙,12=午,13=丁,14=未,15=坤,16=申,17=庚,18=酉,19=辛,20=戌,21=乾,22=亥,23=壬"},"xiang_shan":{"type":"integer","minimum":0,"maximum":23,"description":"朝向；必须等于坐山后 12 山（180°相对）"}},"required":["period_date","zuo_shan","xiang_shan"]}`),
		Handler: xuankongChartHandler,
		Result:  envelopeSchema(schemaXuankongChartData),
	},
	{
		Name: "xuankong.liunian", Description: "玄空流年飞星。chart（可选）+ year → 流年飞星盘 + 宅盘凶星落宫对照（确定性计算）。",
		Params:  mustSchema(`{"type":"object","additionalProperties":false,"properties":{"chart":` + schemaXuankongChartData + `,"year":{"type":"integer","minimum":1864,"maximum":2200,"description":"目标年份"}},"required":["year"]}`),
		Handler: xuankongLiunianHandler,
		Result:  envelopeSchema(`{"type":"object","additionalProperties":false,"properties":{"year":{"type":"integer","minimum":1864,"maximum":2200},"year_boundary":{"type":"string","enum":["gregorian_calendar_year"]},"ru_zhong":` + schemaFlyingStarNameEnum + `,"gong_wei":{"type":"array","minItems":9,"maxItems":9,"items":` + schemaAnnualFlyingStar + `},"house_overlay":{"type":"array","minItems":9,"maxItems":9,"description":"全部流年星落宫对照宅盘该宫三星；不预设星曜固有吉凶","items":{"type":"object","additionalProperties":false,"properties":{"gong_num":{"type":"integer","minimum":1,"maximum":9},"star":` + schemaFlyingStarNameEnum + `,"palace_stars":{"type":"string","description":"运星/山星/向星"}},"required":["gong_num","star","palace_stars"]}}},"required":["year","ru_zhong","gong_wei"]}`),
	},

	{
		Name: "liuyao.qigua", Description: "六爻起卦。mode=auto 时用安全随机三枚铜钱起六次；mode=coins 时传入六组正/反记录；mode=yaos 时传入六个爻值。返回可审计起卦收据。",
		Params:  mustSchema(`{"type":"object","additionalProperties":false,"properties":{"mode":{"type":"string","enum":["auto","coins","yaos"],"default":"auto"},"rounds":{"type":"array","minItems":6,"maxItems":6,"description":"coins 模式：六组三枚硬币，初爻到上爻","items":{"type":"array","minItems":3,"maxItems":3,"items":{"type":"string","enum":["正","反"]}}},"yaos":{"type":"array","minItems":6,"maxItems":6,"items":{"type":"integer","minimum":6,"maximum":9},"description":"yaos 模式：六爻值，初爻到上爻"}},"required":[]}`),
		Handler: liuyaoQiguaHandler,
		Result:  envelopeSchema(`{"type":"object","additionalProperties":false,"properties":{"casting":` + schemaLiuyaoCasting + `},"required":["casting"]}`),
	},
	{
		Name: "liuyao.chart", Description: "六爻装卦。只接受 liuyao.qigua 返回的完整 casting 收据；装卦并分析：纳甲、六亲、六兽、用神、旺衰、应期。lines.liu_qin: 0=父母 1=兄弟 2=官鬼 3=妻财 4=子孙；lines.liu_shou: 0=青龙 1=朱雀 2=勾陈 3=螣蛇 4=白虎 5=玄武；wang_shuai: 0=旺 1=相 2=休 3=囚 4=死",
		Params:  mustSchema(`{"type":"object","additionalProperties":false,"properties":{"solar_time":` + schemaSolarTime + `,"yong_shen":{"type":"string","enum":["父母","兄弟","官鬼","妻财","子孙","世爻","应爻"],"description":"用神；默认世爻"},"casting":` + schemaLiuyaoCasting + `},"required":["solar_time","casting"]}`),
		Handler: liuyaoChartHandler,
		Result:  envelopeSchema(`{"type":"object","properties":{"name":{"type":"string"},"ben_gua":{"type":"string","enum":["乾","姤","遁","否","观","晋","大有","剥","复","颐","屯","益","震","噬嗑","随","无妄","明夷","贲","既济","家人","丰","离","革","同人","临","损","节","中孚","归妹","睽","兑","履","泰","大畜","需","小畜","大壮","大有","夬","乾","姤","遁","否","观","晋","大有","剥","复","颐","屯","益","震","噬嗑","随","无妄","明夷","贲","既济","家人","丰","离","革","同人"]},"lines":{"type":"array","description":"每爻：六亲/六神/世应 + 确定性状态（yue_po 月破/dong_self 发动/dong_sheng 动爻生/dong_ke 动爻克）","items":{"type":"object","properties":{"position":{"type":"integer"},"type":{"type":"integer"},"gan":{"type":"string"},"zhi":{"type":"string"},"wuxing":{"type":"string"},"liu_qin":{"type":"string","enum":["父母","兄弟","官鬼","妻财","子孙"]},"shi_ying":{"type":"string","description":"世/应"},"liu_shou":{"type":"string","enum":["青龙","朱雀","勾陈","螣蛇","白虎","玄武"]},"yue_po":{"type":"boolean","description":"月破"},"dong_self":{"type":"boolean","description":"本爻发动"},"dong_sheng":{"type":"boolean","description":"有动爻生此爻"},"dong_ke":{"type":"boolean","description":"有动爻克此爻"},"xun_kong":{"type":"boolean","description":"该爻地支值日柱旬空"},"mu_ku_branch":{"type":"string"},"mu_ku_element":{"type":"string"},"mu_ku_types":{"type":"array","items":{"type":"string","enum":["day","moving","transformed"]}},"chang_sheng_yue":{"type":"string"}},"required":["position","type","gan","zhi","wuxing","liu_qin","shi_ying","liu_shou"]}},"yong_shen":{"type":"object","description":"用神结果","properties":{"name":{"type":"string","description":"用神六亲"},"position":{"type":"integer","description":"爻位1-6（0=本卦不现）"},"is_hidden":{"type":"boolean","description":"本卦不现，取本宫伏神"},"wang_shuai":{"type":"string","description":"用神旺衰"},"yue_po":{"type":"boolean","description":"用神月破"},"xun_kong":{"type":"boolean","description":"用神旬空"},"mu_ku":{"type":"boolean","description":"用神入墓"},"liu_shou":{"type":"string","description":"用神临的六神","enum":["青龙","朱雀","勾陈","螣蛇","白虎","玄武"]},"fu_shen":{"type":"object","description":"飞伏（用神不现时）","properties":{"position":{"type":"integer","description":"爻位"},"liu_qin":{"type":"string","description":"伏神六亲"},"zhi":{"type":"string","description":"伏神地支"}},"required":["position","liu_qin","zhi"]},"fu_shen_candidates":{"type":"array","description":"本宫全部同六亲伏神候选","items":{"type":"object","properties":{"position":{"type":"integer"},"liu_qin":{"type":"string"},"zhi":{"type":"string"}},"required":["position","liu_qin","zhi"]}},"chang_sheng":{"type":"string"}},"required":["name","position"]},"wang_shuai":{"type":"array"},"yue_jian_zhi":{"type":"string","enum":["子","丑","寅","卯","辰","巳","午","未","申","酉","戌","亥"]},"yue_jian_gan":{"type":"string","enum":["甲","乙","丙","丁","戊","己","庚","辛","壬","癸"]},"bian_gua":{"type":"string","description":"变卦名"},"bian_yao":{"type":"array","description":"变爻"},"gong":{"type":"string","description":"八宫"},"gua_ci":{"type":"object","description":"卦辞爻辞"},"gong_wuxing":{"type":"string","description":"宫五行"},"ri_chen_gan":{"type":"string","description":"日辰天干"},"ri_chen_zhi":{"type":"string","description":"日辰地支"},"ri_chen_relations":{"type":"array","description":"日辰与爻关系集合；冲合可与生扶克并存","items":{"type":"object","additionalProperties":false,"properties":{"relations":{"type":"array","items":{"type":"string","enum":["冲","合","生","扶","克","平"]}}},"required":["relations"]}},"xun_kong":{"type":"array","items":{"type":"string","enum":["子","丑","寅","卯","辰","巳","午","未","申","酉","戌","亥"]},"description":"日柱旬空地支（甲子旬空戌亥…）"},"dong_yao_relations":{"type":"array","description":"动爻与用神的关系集合；直接与间接作用可并列，静卦为空数组","items":{"type":"object","properties":{"position":{"type":"integer","description":"动爻位置"},"relations":{"type":"array","items":{"type":"string","enum":["生用","克用","比和","冲用","生原神","克原神","生忌神","克忌神"]}}},"required":["position","relations"]}},"patterns":{"type":"array","description":"特殊格局","items":{"type":"object","properties":{"type":{"type":"string","description":"格局类型"},"sub_type":{"type":"string","description":"子类型"},"position":{"type":"integer"},"is_true":{"type":"boolean","description":"结构是否有实质效力；空亡/月破中 true=真空/真破，false=假空/假破"},"assessment":{"type":"string"}},"required":["type","assessment"]}},"casting":{"type":"object","description":"完整起卦收据"},"timing_candidates":{"type":"array","description":"应期触发机制候选；非确定日期","items":{"type":"object","properties":{"id":{"type":"string"},"mechanism":{"type":"string"},"position":{"type":"integer"},"trigger_branch":{"type":"string","enum":["子","丑","寅","卯","辰","巳","午","未","申","酉","戌","亥"]},"window":{"type":"string"},"basis":{"type":"array","items":{"type":"string"}},"condition":{"type":"string"},"confidence":{"type":"string"}},"required":["id","mechanism","confidence"]}},"day_clash_facts":{"type":"array","description":"日冲细分事实","items":{"type":"object","properties":{"position":{"type":"integer"},"kind":{"type":"string"},"basis":{"type":"array","items":{"type":"string"}},"condition":{"type":"string"}},"required":["position","kind"]}},"moving_transformations":{"type":"array","description":"动爻到变爻的机械变化","items":{"type":"object","properties":{"position":{"type":"integer"},"from_branch":{"type":"string"},"to_branch":{"type":"string"},"return_relation":{"type":"string"},"advance_retreat":{"type":"string"},"changed_void":{"type":"boolean"},"changed_break":{"type":"boolean"},"basis":{"type":"array","items":{"type":"string"}}},"required":["position","from_branch","to_branch"]}},"force_chain":{"type":"object","properties":{"yong_shen":{"type":"string"},"yong_element":{"type":"string"},"position":{"type":"integer"},"is_hidden":{"type":"boolean"},"entries":{"type":"array","items":{"type":"object","properties":{"role":{"type":"string"},"element":{"type":"string"},"positions":{"type":"array","items":{"type":"integer"}},"branches":{"type":"array","items":{"type":"string"}},"relations":{"type":"array","items":{"type":"string"}},"states":{"type":"array","items":{"type":"string"}},"note":{"type":"string"}},"required":["role","element"]}}},"required":["yong_shen","yong_element","position","entries"]},"yong_shen_candidates":{"type":"array","description":"用神候选与当前确定性选择","items":{"type":"object","properties":{"position":{"type":"integer"},"branch":{"type":"string"},"is_hidden":{"type":"boolean"},"selected":{"type":"boolean"},"wang_shuai":{"type":"string"},"reason":{"type":"string"},"basis":{"type":"array","items":{"type":"string"}}},"required":["position","branch","is_hidden","selected","reason"]}},"san_he_candidates":{"type":"array","properties":{"moving_count":{"type":"integer"},"static_count":{"type":"integer"},"void_count":{"type":"integer"},"break_count":{"type":"integer"},"positions":{"type":"array","items":{"type":"integer"}},"branches":{"type":"array","items":{"type":"string"}},"element":{"type":"string"},"complete":{"type":"boolean"},"qualification":{"type":"string"},"line_complete":{"type":"boolean"},"day_present":{"type":"boolean"},"month_present":{"type":"boolean"},"missing":{"type":"array","items":{"type":"string"}},"activated":{"type":"boolean"},"targets":{"type":"array","items":{"type":"string"}},"conclusion_scope":{"type":"string"}},"required":["positions","branches","element","complete","moving_count","qualification","line_complete","activated","conclusion_scope"],"description":"三合候选；complete=true 只是三支齐，成局仍须按发动与空破月日复核","items":{"type":"object","properties":{"moving_count":{"type":"integer"},"static_count":{"type":"integer"},"void_count":{"type":"integer"},"break_count":{"type":"integer"},"positions":{"type":"array","items":{"type":"integer"}},"branches":{"type":"array","items":{"type":"string"}},"element":{"type":"string"},"complete":{"type":"boolean"},"qualification":{"type":"string"},"line_complete":{"type":"boolean"},"day_present":{"type":"boolean"},"month_present":{"type":"boolean"},"missing":{"type":"array","items":{"type":"string"}},"activated":{"type":"boolean"},"targets":{"type":"array","items":{"type":"string"}},"conclusion_scope":{"type":"string"}},"required":["positions","branches","element","complete","moving_count","qualification","line_complete","activated","conclusion_scope"]}},"tomb_school":{"type":"object","additionalProperties":false,"properties":{"earth_branch":{"type":"string","enum":["辰"]},"school":{"type":"string","enum":["engine_default_chen"]}},"required":["earth_branch","school"]},"hidden_lines":{"type":"array","description":"本宫全部伏神","items":{"type":"object","properties":{"position":{"type":"integer"},"liu_qin":{"type":"string"},"gan":{"type":"string"},"zhi":{"type":"string"},"wuxing":{"type":"string"},"flying_position":{"type":"integer"},"flying_liu_qin":{"type":"string"},"flying_gan":{"type":"string"},"flying_zhi":{"type":"string"},"flying_wuxing":{"type":"string"},"fei_fu_relation":{"type":"string"},"xun_kong":{"type":"boolean"},"yue_po":{"type":"boolean"},"mu_ku":{"type":"boolean"},"exit_condition":{"type":"string"}},"required":["position","liu_qin","gan","zhi","wuxing","flying_position","flying_liu_qin","fei_fu_relation"]}},"branch_relation_facts":{"type":"array","description":"爻间地支关系","items":{"type":"object","properties":{"left_position":{"type":"integer"},"right_position":{"type":"integer"},"relations":{"type":"array","items":{"type":"string"}},"targets":{"type":"array","items":{"type":"string"}},"basis":{"type":"array","items":{"type":"string"}}},"required":["left_position","right_position","relations"]}},"conflicts":{"type":"array","description":"engine 已判定的并列/矛盾信号；解释必须覆盖","items":{"type":"object","properties":{"id":{"type":"string","enum":["moving-support-opposition","yuanshen-support-opposition","strong-but-void","strong-but-break"]},"reason":{"type":"string"},"facts":{"type":"array","items":{"type":"string"}}},"required":["id","reason"]}}},"required":["name","ben_gua","lines","yong_shen","tomb_school","casting"]}`),
	},
	{
		Name: "huangli.days", Description: "黄历查日。返回连续N天的黄历信息；event 给定时按建除事项规则输出 suitability。",
		Params:  mustSchema(`{"type":"object","properties":{"start_date":{"type":"string","format":"date","description":"起始日期 YYYY-MM-DD"},"count":{"type":"integer","minimum":1,"maximum":30,"description":"查几天，默认3，最多30"},"event":{"type":"string","enum":["wedding","engage","opening","sign","move","travel","build","exam","medical","sacrifice","cleaning","renovation","bed_install","income","funeral"],"description":"事项类型；缺省仅列黄历事实"}},"required":["start_date"]}`),
		Handler: huangliDaysHandler,
		Result:  envelopeSchema(`{"type":"array","items":{"type":"object","properties":{"date":{"type":"string"},"ri_zhu":{"type":"object","properties":{"gan":{"type":"string","enum":["甲","乙","丙","丁","戊","己","庚","辛","壬","癸"]},"zhi":{"type":"string","enum":["子","丑","寅","卯","辰","巳","午","未","申","酉","戌","亥"]},"nayin":{"type":"string"}},"required":["gan","zhi"]},"nayin":{"type":"string"},"wuxing":{"type":"string","enum":["金","木","水","火","土"]},"jian_chu":{"type":"string","enum":["建","除","满","平","定","执","破","危","成","收","开","闭"]},"event":{"type":"string"},"event_label":{"type":"string"},"suitability":{"type":"string","enum":["unspecified","recommended","unsuitable","possible"]},"reason":{"type":"string"},"warnings":{"type":"array","items":{"type":"string"}},"huangdao":{"type":"object","properties":{"name":{"type":"string"},"path":{"type":"string"}},"required":["name","path"]},"xi_shen":{"type":"string"},"cai_shen":{"type":"string"},"fu_shen":{"type":"string"},"gan_ji":{"type":"string"},"zhi_ji":{"type":"string"},"mansion":{"type":"object","properties":{"name":{"type":"string"},"index":{"type":"integer"},"animal":{"type":"string"},"wuxing":{"type":"string","enum":["金","木","水","火","土","日","月"]},"group":{"type":"string"}},"required":["name","index"]}},"required":["date","ri_zhu"]}}`),
	},
	{
		Name: "time.now", Description: "获取服务端当前时间。返回 UTC、本地、北京时间，用于 AI agent 获取准确时间避免幻觉。",
		Params:  mustSchema(`{"type":"object","additionalProperties":false,"properties":{},"required":[]}`),
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
