package agent

import (
	"context"
	"encoding/json"
	"fmt"
	"time"

	"liki-engine/internal/engine/bazi"
	"liki-engine/internal/engine/ganzhi"
	"liki-engine/internal/engine/tianwen"
)

func baziFullChartHandler(ctx context.Context, raw json.RawMessage) (json.RawMessage, error) {
	var p struct {
		Chart json.RawMessage `json:"chart"`
	}
	if err := json.Unmarshal(raw, &p); err != nil {
		return nil, fmt.Errorf("bazi.fullchart: %w", err)
	}
	var chart bazi.Chart
	if err := json.Unmarshal(p.Chart, &chart); err != nil {
		return nil, fmt.Errorf("bazi.fullchart: %w", err)
	}
	result := bazi.ComputeFullChart(chart)
	return wrapResult("bazi_fullchart", result)
}

func baziChartHandler(ctx context.Context, raw json.RawMessage) (json.RawMessage, error) {
	var p struct {
		SolarTime string         `json:"solar_time"`
		Gender    ganzhi.Gender `json:"gender"`
	}
	if err := json.Unmarshal(raw, &p); err != nil {
		return nil, fmt.Errorf("compute_chart: %w", err)
	}
	if err := validateGender(p.Gender); err != nil {
		return nil, fmt.Errorf("compute_chart: %w", err)
	}
	st, err := parseSolarTime(p.SolarTime)
	if err != nil {
		return nil, fmt.Errorf("compute_chart: %w", err)
	}
	result := bazi.ComputeChart(st, p.Gender)
	// Compute current step index from server time.
	if result.DaYun != nil {
		now := time.Now()
		birth := time.Time(st)
		result.DaYun.CurrentStepIndex = bazi.ComputeCurrentStepIndex(
			result.DaYun, birth.Year(), now.Year(), now.YearDay(), birth.YearDay(),
		)
	}
	return wrapResult("chart", result)
}

func baziYongShenHandler(ctx context.Context, raw json.RawMessage) (json.RawMessage, error) {
	c, err := parseChart(raw, "compute_yongshen")
	if err != nil {
		return nil, err
	}
	result := bazi.ComputeYongShen(c)
	return wrapResult("yongshen", result)
}

func baziHeHuiHandler(ctx context.Context, raw json.RawMessage) (json.RawMessage, error) {
	c, err := parseChart(raw, "compute_hehui")
	if err != nil {
		return nil, err
	}
	result := bazi.ComputeHeHui(c)
	return wrapResult("hehui", result)
}

func baziChartExtraHandler(ctx context.Context, raw json.RawMessage) (json.RawMessage, error) {
	c, err := parseChart(raw, "compute_chart_extra")
	if err != nil {
		return nil, err
	}
	result := bazi.ComputeChartExtra(c)
	return wrapResult("chart_extra", result)
}

func parseChart(raw json.RawMessage, method string) (bazi.Chart, error) {
	// Try direct format: chart JSON at top level, e.g. {"nian":{...},"yue":{...}}
	var chart bazi.Chart
	if json.Unmarshal(raw, &chart) == nil {
		// Validate: at minimum Gan should be non-zero
		if chart.Ri.Gan != 0 {
			return chart, nil
		}
	}
	// Fall back to wrapped format: {"chart":{"nian":{...}}}
	var p struct {
		Chart bazi.Chart `json:"chart"`
	}
	if err := json.Unmarshal(raw, &p); err != nil {
		return p.Chart, fmt.Errorf("%s: parse chart: %w", method, err)
	}
	return p.Chart, nil
}

func baziBondHandler(ctx context.Context, raw json.RawMessage) (json.RawMessage, error) {
	var p struct {
		A struct {
			Chart json.RawMessage `json:"chart"`
		} `json:"a"`
		B struct {
			Chart json.RawMessage `json:"chart"`
		} `json:"b"`
	}
	if err := json.Unmarshal(raw, &p); err != nil {
		return nil, fmt.Errorf("compute_bond: %w", err)
	}
	coreA, err := parseChart(p.A.Chart, "compute_bond")
	if err != nil {
		return nil, err
	}
	coreB, err := parseChart(p.B.Chart, "compute_bond")
	if err != nil {
		return nil, err
	}
	result := bazi.ComputeBond(coreA, coreB)
	return wrapResult("bond", result)
}

func baziLiunianHandler(ctx context.Context, raw json.RawMessage) (json.RawMessage, error) {
	var p struct {
		Year  int             `json:"year"`
		Chart json.RawMessage `json:"chart"`
	}
	if err := json.Unmarshal(raw, &p); err != nil {
		return nil, fmt.Errorf("compute_liunian: %w", err)
	}
	if p.Year <= 0 {
		return nil, fmt.Errorf("compute_liunian: year must be positive, got %d", p.Year)
	}
	core, err := parseChart(p.Chart, "compute_liunian")
	if err != nil {
		return nil, err
	}
	result, err := bazi.ComputeLiuNian(core, p.Year)
	if err != nil {
		return nil, fmt.Errorf("compute_liunian: %w", err)
	}
	return wrapResult("liunian", result)
}

func baziLiuyueHandler(ctx context.Context, raw json.RawMessage) (json.RawMessage, error) {
	var p struct {
		Year  int             `json:"year"`
		Month int             `json:"month"`
		Chart json.RawMessage `json:"chart"`
	}
	if err := json.Unmarshal(raw, &p); err != nil {
		return nil, fmt.Errorf("compute_liuyue: %w", err)
	}
	core, err := parseChart(p.Chart, "compute_liuyue")
	if err != nil {
		return nil, err
	}
	result, err := bazi.ComputeLiuYue(core, p.Year, p.Month)
	if err != nil {
		return nil, fmt.Errorf("compute_liuyue: %w", err)
	}
	return wrapResult("liuyue", result)
}

func baziLiuriHandler(ctx context.Context, raw json.RawMessage) (json.RawMessage, error) {
	var p struct {
		Year  int             `json:"year"`
		Month int             `json:"month"`
		Day   int             `json:"day"`
		Chart json.RawMessage `json:"chart"`
	}
	if err := json.Unmarshal(raw, &p); err != nil {
		return nil, fmt.Errorf("compute_liuri: %w", err)
	}
	core, err := parseChart(p.Chart, "compute_liuri")
	if err != nil {
		return nil, err
	}
	result, err := bazi.ComputeLiuRi(core, p.Year, p.Month, p.Day)
	if err != nil {
		return nil, fmt.Errorf("compute_liuri: %w", err)
	}
	return wrapResult("liuri", result)
}

func baziLiushiHandler(ctx context.Context, raw json.RawMessage) (json.RawMessage, error) {
	var p struct {
		Year  int             `json:"year"`
		Month int             `json:"month"`
		Day   int             `json:"day"`
		Hour  int             `json:"hour"`
		Chart json.RawMessage `json:"chart"`
	}
	if err := json.Unmarshal(raw, &p); err != nil {
		return nil, fmt.Errorf("compute_liushi: %w", err)
	}
	if p.Hour < 0 || p.Hour > 23 {
		return nil, fmt.Errorf("compute_liushi: hour must be 0-23, got %d", p.Hour)
	}
	core, err := parseChart(p.Chart, "compute_liushi")
	if err != nil {
		return nil, err
	}
	result, err := bazi.ComputeLiuShi(core, p.Year, p.Month, p.Day, p.Hour)
	if err != nil {
		return nil, fmt.Errorf("compute_liushi: %w", err)
	}
	return wrapResult("liushi", result)
}

func baziXiaoYunHandler(ctx context.Context, raw json.RawMessage) (json.RawMessage, error) {
	var p struct {
		Chart json.RawMessage `json:"chart"`
		Count int             `json:"count"`
	}
	if err := json.Unmarshal(raw, &p); err != nil {
		return nil, fmt.Errorf("compute_xiaoyun: %w", err)
	}
	core, err := parseChart(p.Chart, "compute_xiaoyun")
	if err != nil {
		return nil, err
	}
	result := bazi.ComputeXiaoYun(core, p.Count)
	return wrapResult("xiaoyun", result)
}

func baziXiaoXianHandler(ctx context.Context, raw json.RawMessage) (json.RawMessage, error) {
	var p struct {
		Gender ganzhi.Gender `json:"gender"`
		Count  int            `json:"count"`
	}
	if err := json.Unmarshal(raw, &p); err != nil {
		return nil, fmt.Errorf("compute_xiaoxian: %w", err)
	}
	if err := validateGender(p.Gender); err != nil {
		return nil, fmt.Errorf("compute_xiaoxian: %w", err)
	}
	result := bazi.ComputeXiaoXian(p.Gender, p.Count)
	return wrapResult("xiaoxian", result)
}

func parseSolarTime(s string) (tianwen.SolarTime, error) {
	t, err := time.Parse(time.RFC3339, s)
	if err != nil {
		return tianwen.SolarTime{}, fmt.Errorf("invalid solar_time %q: %w", s, err)
	}
	return tianwen.SolarTime(t), nil
}

var baziMethods = []RPCMethod{
	{
		Name: "bazi.fullchart", Description: "扩展命盘。传入 bazi.chart 返回的最小命盘，补全十神/藏干/神煞/长生/空亡/自合/魁罡等完整信息。",
		Params: mustSchema(`{"type":"object","properties":{"chart":{"type":"object","description":"bazi.chart 返回的最小命盘"}},"required":["chart"]}`),
		Handler: baziFullChartHandler,
		Result: envelopeSchema(`{"type":"object","properties":{"nian":{"type":"object","properties":{"gan":{"type":"string"},"zhi":{"type":"string"},"na_yin":{"type":"string"}},"required":["gan","zhi","na_yin"]},"yue":{"type":"object","properties":{"gan":{"type":"string"},"zhi":{"type":"string"},"na_yin":{"type":"string"}},"required":["gan","zhi","na_yin"]},"ri":{"type":"object","properties":{"gan":{"type":"string"},"zhi":{"type":"string"},"na_yin":{"type":"string"}},"required":["gan","zhi","na_yin"]},"shi":{"type":"object","properties":{"gan":{"type":"string"},"zhi":{"type":"string"},"na_yin":{"type":"string"}},"required":["gan","zhi","na_yin"]},"da_yun":{"type":"object"},"gender":{"type":"string"}},"required":["nian","yue","ri","shi","da_yun","gender"]}`),
	},
	{
		Name: "bazi.chart", Description: "排八字命盘。返回最小命盘（四柱+纳音+大运+性别）。如需十神/藏干/神煞/长生/空亡等完整信息，请将结果传入 bazi.fullchart。用神、合会冲刑、补充信息需另行调用 bazi.yongshen / bazi.hehui / bazi.chart_extra。",
		Params: mustSchema(`{"type":"object","properties":{"solar_time":` + schemaSolarTime + `,"gender":` + schemaGender + `},"required":["solar_time","gender"]}`), Handler: baziChartHandler,
		Result: envelopeSchema(`{"type":"object","properties":{"nian":{"type":"object","properties":{"gan":{"type":"string"},"zhi":{"type":"string"},"na_yin":{"type":"string"}},"required":["gan","zhi","na_yin"]},"yue":{"type":"object","properties":{"gan":{"type":"string"},"zhi":{"type":"string"},"na_yin":{"type":"string"}},"required":["gan","zhi","na_yin"]},"ri":{"type":"object","properties":{"gan":{"type":"string"},"zhi":{"type":"string"},"na_yin":{"type":"string"}},"required":["gan","zhi","na_yin"]},"shi":{"type":"object","properties":{"gan":{"type":"string"},"zhi":{"type":"string"},"na_yin":{"type":"string"}},"required":["gan","zhi","na_yin"]},"da_yun":{"type":"object","description":"大运。steps 每步含 gan/zhi/age_start/age_end/name/element/shi_shen（十年一运）。current_step_index 为当前所处大运索引（-1 表示未起运或已过完所有大运）"},"gender":{"type":"string"}},"required":["nian","yue","ri","shi","da_yun","gender"]}`),
	},
	{
		Name: "bazi.yongshen", Description: "八字用神分析。基于扶抑（旺衰）、调候（穷通宝鉴）、格局（子平）三派计算用神/喜神/忌神。返回五行计数、旺相休囚死、三派用神。LLM 综合三派结果判断最终用神。注意：三派结果分别在 fu_yi / tiao_hou / ge_ju 子对象中，顶层 shen / yongshen / xishen / jishen 字段不存在，请直接引用子对象内的 yong / xi / ji。",
		Params: mustSchema(`{"type":"object","properties":{"chart":{"type":"object","description":"八字命盘（由 bazi.chart 返回的 data）"}},"required":["chart"]}`),
		Handler: baziYongShenHandler,
		Result: envelopeSchema(`{"type":"object","properties":{"fu_yi":{"type":"object","properties":{"wuxing_count":{"type":"object"},"wang_shuai":{"type":"object"},"yong":{"type":"string"},"xi":{"type":"string"},"ji":{"type":"string"},"qiangruo":{"type":"string"}},"required":["wuxing_count","wang_shuai","yong","xi","ji","qiangruo"]},"tiao_hou":{"type":"object","properties":{"yong":{"type":"string"},"xi":{"type":"string"},"ji":{"type":"string"},"season":{"type":"string"},"detail":{"type":"string"}},"required":["yong","xi","ji","season"]},"ge_ju":{"type":"object","properties":{"yong":{"type":"string"},"xi":{"type":"string"},"ji":{"type":"string"},"ge_ju":{"type":"string"},"yong_fa":{"type":"string"}},"required":["yong","xi","ji","ge_ju","yong_fa"]}},"required":["fu_yi","tiao_hou","ge_ju"]}`),
	},
	{
		Name: "bazi.hehui", Description: "八字合会冲刑分析。返回天干五合、地支六合、三合局、三会方、六冲、六害、相刑。",
		Params: mustSchema(`{"type":"object","properties":{"chart":{"type":"object","description":"八字命盘（由 bazi.chart 返回的 data）"}},"required":["chart"]}`),
		Handler: baziHeHuiHandler,
		Result:  envelopeSchema(`{"type":"object","properties":{"gan_he":{"type":"array","description":"天干五合"},"zhi_liu_he":{"type":"array","description":"地支六合"},"san_he":{"type":"array","description":"完整三合局。注意：只返回完整三合（三支齐全），半合/拱合需 LLM 自行判断"},"san_hui":{"type":"array","description":"三会方"},"liu_chong":{"type":"array","description":"六冲"},"liu_hai":{"type":"array","description":"六害"},"liu_xing":{"type":"array","description":"相刑"}},"required":["gan_he","zhi_liu_he","san_he","san_hui","liu_chong","liu_hai","liu_xing"]}`),
	},
	{
		Name: "bazi.chart_extra", Description: "八字补充信息。返回三元（胎元/命宫/身宫）、拱夹、纳音生克、长生十二宫、三奇贵人。按需调用。",
		Params: mustSchema(`{"type":"object","properties":{"chart":{"type":"object","description":"八字命盘（由 bazi.chart 返回的 data）"}},"required":["chart"]}`),
		Handler: baziChartExtraHandler,
		Result:  envelopeSchema(`{"type":"object","properties":{"san_yuan":{"type":"object","description":"三元: 胎元/命宫/身宫，各含gan,zhi"},"gong_jia":{"type":"array"},"nayin_rel":{"type":"array"},"chang_sheng":{"type":"array"},"san_qi_name":{"type":"string"}},"required":["san_yuan","gong_jia","nayin_rel","chang_sheng","san_qi_name"]}`),
	},
	{
		Name: "bazi.bond", Description: "八字合盘。返回双方日主、天干关系（合/生/克）、地支关系（六合/三合/六冲）、纳音配合、五行互补。",
		Params: mustSchema(`{"type":"object","properties":{"a":{"type":"object","properties":{"chart":{"type":"object","description":"八字命盘（由 bazi.chart 返回的最小命盘，不需 bazi.fullchart）"}},"required":["chart"]},"b":{"type":"object","properties":{"chart":{"type":"object","description":"八字命盘（由 bazi.chart 返回的最小命盘，不需 bazi.fullchart）"}},"required":["chart"]}},"required":["a","b"]}`),
		Handler: baziBondHandler,
		Result:  envelopeSchema(`{"type":"object","properties":{"zhu_cross":{"type":"object","description":"16对四柱互配结果，含a_zhu/b_zhu/a_stem/b_stem/stem_rel/branch_rel等"},"shi_shen_cross":{"type":"object","description":"双方日主看对方四柱的十神映射: a_to_b/b_to_a"},"structure":{"type":"object","description":"大运交叉+旬宫判断，含da_yun/xun_gong"}},"required":["zhu_cross","shi_shen_cross","structure"]}`),
	},
	{
		Name: "bazi.liunian", Description: "八字流年运势。返回流年干支与命局的十神、神煞、伏吟反吟。",
		Params: mustSchema(`{"type":"object","properties":{"year":{"type":"integer","description":"目标年份"},"chart":{"type":"object","description":"八字命盘（由 bazi.chart 返回的最小命盘，不需 bazi.fullchart）"}},"required":["year","chart"]}`),
		Handler: baziLiunianHandler,
		Result:  envelopeSchema(`{"type":"object","properties":{"year":{"type":"integer"},"year_name":{"type":"string"},"shi_shen":{"type":"string"}},"required":["year","year_name","shi_shen"]}`),
	},
	{
		Name: "bazi.liuyue", Description: "流月运势。返回流月干支与命局的十神、神煞。",
		Params: mustSchema(`{"type":"object","properties":{"year":{"type":"integer","description":"目标年份"},"month":{"type":"integer","minimum":1,"maximum":12,"description":"目标月份"},"chart":{"type":"object","description":"八字命盘（由 bazi.chart 返回的最小命盘，不需 bazi.fullchart）"}},"required":["year","month","chart"]}`),
		Handler: baziLiuyueHandler,
		Result:  envelopeSchema(`{"type":"object","properties":{"year":{"type":"integer"},"month":{"type":"integer"},"month_name":{"type":"string"},"shi_shen":{"type":"string"}},"required":["year","month","month_name","shi_shen"]}`),
	},
	{
		Name: "bazi.liuri", Description: "流日运势。返回流日干支、十神、纳音。",
		Params: mustSchema(`{"type":"object","properties":{"year":{"type":"integer","description":"目标年份"},"month":{"type":"integer","minimum":1,"maximum":12,"description":"目标月份"},"day":{"type":"integer","minimum":1,"maximum":31,"description":"目标日期"},"chart":{"type":"object","description":"八字命盘（由 bazi.chart 返回的最小命盘，不需 bazi.fullchart）"}},"required":["year","month","day","chart"]}`),
		Handler: baziLiuriHandler,
		Result:  envelopeSchema(`{"type":"object","properties":{"date":{"type":"string"},"day_name":{"type":"string"},"shi_shen":{"type":"string"}},"required":["date","day_name","shi_shen"]}`),
	},
	{
		Name: "bazi.liushi", Description: "流时运势。返回流时干支、十神。hour 为时辰（0-23）。",
		Params: mustSchema(`{"type":"object","properties":{"year":{"type":"integer","description":"目标年份"},"month":{"type":"integer","minimum":1,"maximum":12,"description":"目标月份"},"day":{"type":"integer","minimum":1,"maximum":31,"description":"目标日期"},"hour":{"type":"integer","minimum":0,"maximum":23,"description":"时辰"},"chart":{"type":"object","description":"八字命盘（由 bazi.chart 返回的最小命盘，不需 bazi.fullchart）"}},"required":["year","month","day","hour","chart"]}`),
		Handler: baziLiushiHandler,
		Result:  envelopeSchema(`{"type":"object","properties":{"time":{"type":"string"},"hour_name":{"type":"string"},"shi_shen":{"type":"string"}},"required":["time","hour_name","shi_shen"]}`),
	},
	{
		Name: "bazi.xiaoyun", Description: "小运。返回小运流年列表。count 默认 12。",
		Params: mustSchema(`{"type":"object","properties":{"chart":{"type":"object","description":"八字命盘（由 bazi.chart 返回的最小命盘，不需 bazi.fullchart）"},"count":{"type":"integer","description":"返回年数，默认 12"}},"required":["chart"]}`),
		Handler: baziXiaoYunHandler,
		Result:  envelopeSchema(`{"type":"array","items":{"type":"object","properties":{"age":{"type":"integer"},"gan":{"type":"string"},"zhi":{"type":"string"},"name":{"type":"string"}},"required":["age","gan","zhi","name"]}}`),
	},
	{
		Name: "bazi.xiaoxian", Description: "小限。返回小限列表。count 默认 12。",
		Params: mustSchema(`{"type":"object","properties":{"gender":` + schemaGender + `,"count":{"type":"integer","description":"返回年数，默认 12"}},"required":["gender"]}`),
		Handler: baziXiaoXianHandler,
		Result:  envelopeSchema(`{"type":"array","items":{"type":"object","properties":{"age":{"type":"integer"},"branch":{"type":"string"}},"required":["age","branch"]}}`),
	},
}
