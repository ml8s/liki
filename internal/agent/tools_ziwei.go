package agent

import (
	"context"
	"encoding/json"
	"fmt"

	"liki-engine/internal/engine/ganzhi"
	"liki-engine/internal/engine/ziwei"
)

func ziweiChartHandler(ctx context.Context, raw json.RawMessage) (json.RawMessage, error) {
	var p struct {
		SolarTime string         `json:"solar_time"`
		Gender    ganzhi.Gender `json:"gender"`
	}
	if err := json.Unmarshal(raw, &p); err != nil {
		return nil, fmt.Errorf("compute_ziwei: %w", err)
	}
	if err := validateGender(p.Gender); err != nil {
		return nil, fmt.Errorf("compute_ziwei: %w", err)
	}
	st, err := parseSolarTime(p.SolarTime)
	if err != nil {
		return nil, fmt.Errorf("compute_ziwei: %w", err)
	}
	result := ziwei.ComputeChart(st, p.Gender)
	return wrapResult("ziwei", result)
}

func ziweiDaxianHandler(ctx context.Context, raw json.RawMessage) (json.RawMessage, error) {
	var p struct {
		Chart json.RawMessage `json:"chart"`
	}
	if err := json.Unmarshal(raw, &p); err != nil {
		return nil, fmt.Errorf("compute_ziwei_daxian: %w", err)
	}
	var chart ziwei.Chart
	if err := json.Unmarshal(p.Chart, &chart); err != nil {
		return nil, fmt.Errorf("compute_ziwei_daxian: parse chart: %w", err)
	}
	result := ziwei.ComputeDaXian(chart)
	return wrapResult("ziwei_daxian", result)
}

func ziweiLiunianHandler(ctx context.Context, raw json.RawMessage) (json.RawMessage, error) {
	var p struct {
		LiuYear int             `json:"liu_year"`
		Chart   json.RawMessage `json:"chart"`
	}
	if err := json.Unmarshal(raw, &p); err != nil {
		return nil, fmt.Errorf("compute_ziwei_liunian: %w", err)
	}
	var chart ziwei.Chart
	if err := json.Unmarshal(p.Chart, &chart); err != nil {
		return nil, fmt.Errorf("compute_ziwei_liunian: parse chart: %w", err)
	}
	result := ziwei.ComputeLiuNian(chart, p.LiuYear)
	return wrapResult("ziwei_liunian", result)
}

func ziweiLiuyueHandler(ctx context.Context, raw json.RawMessage) (json.RawMessage, error) {
	var p struct {
		LiuYear    int             `json:"liu_year"`
		LunarMonth int             `json:"lunar_month"`
		Chart      json.RawMessage `json:"chart"`
	}
	if err := json.Unmarshal(raw, &p); err != nil {
		return nil, fmt.Errorf("compute_ziwei_liuyue: %w", err)
	}
	var chart ziwei.Chart
	if err := json.Unmarshal(p.Chart, &chart); err != nil {
		return nil, fmt.Errorf("compute_ziwei_liuyue: parse chart: %w", err)
	}
	result := ziwei.ComputeLiuYue(chart, p.LiuYear, p.LunarMonth)
	return wrapResult("ziwei_liuyue", result)
}

func ziweiLiuriHandler(ctx context.Context, raw json.RawMessage) (json.RawMessage, error) {
	var p struct {
		LiuYear    int             `json:"liu_year"`
		LunarMonth int             `json:"lunar_month"`
		LunarDay   int             `json:"lunar_day"`
		Chart      json.RawMessage `json:"chart"`
	}
	if err := json.Unmarshal(raw, &p); err != nil {
		return nil, fmt.Errorf("compute_ziwei_liuri: %w", err)
	}
	var chart ziwei.Chart
	if err := json.Unmarshal(p.Chart, &chart); err != nil {
		return nil, fmt.Errorf("compute_ziwei_liuri: parse chart: %w", err)
	}
	result := ziwei.ComputeLiuRi(chart, p.LiuYear, p.LunarMonth, p.LunarDay)
	return wrapResult("ziwei_liuri", result)
}

func ziweiJudgmentHandler(ctx context.Context, raw json.RawMessage) (json.RawMessage, error) {
	var p struct {
		Chart json.RawMessage `json:"chart"`
	}
	if err := json.Unmarshal(raw, &p); err != nil {
		return nil, fmt.Errorf("ziwei.judgment: %w", err)
	}
	var chart ziwei.Chart
	if err := json.Unmarshal(p.Chart, &chart); err != nil {
		return nil, fmt.Errorf("ziwei.judgment: parse chart: %w", err)
	}
	result := ziwei.ComputeJudgment(chart)
	return wrapResult("ziwei_judgment", result)
}

func ziweiBondHandler(ctx context.Context, raw json.RawMessage) (json.RawMessage, error) {
	var p struct {
		A json.RawMessage `json:"a"`
		B json.RawMessage `json:"b"`
	}
	if err := json.Unmarshal(raw, &p); err != nil {
		return nil, fmt.Errorf("compute_ziwei_bond: %w", err)
	}
	var chartA, chartB ziwei.Chart
	if err := json.Unmarshal(p.A, &chartA); err != nil {
		return nil, fmt.Errorf("compute_ziwei_bond: parse chart a: %w", err)
	}
	if err := json.Unmarshal(p.B, &chartB); err != nil {
		return nil, fmt.Errorf("compute_ziwei_bond: parse chart b: %w", err)
	}
	if chartA.BirthYear == 0 {
		return nil, fmt.Errorf("compute_ziwei_bond: chart a is empty")
	}
	if chartB.BirthYear == 0 {
		return nil, fmt.Errorf("compute_ziwei_bond: chart b is empty")
	}
	result := ziwei.ComputeBond(chartA, chartB)
	return wrapResult("ziwei_bond", result)
}

var ziweiMethods = []RPCMethod{
	{
		Name: "ziwei.chart", Description: "紫微斗数排盘。返回十二宫星曜分布、亮度、四化。",
		Params: mustSchema(`{"type":"object","properties":{"solar_time":` + schemaSolarTime + `,"gender":` + schemaGender + `},"required":["solar_time","gender"]}`), Handler: ziweiChartHandler,
		Result: envelopeSchema(`{"type":"object","properties":{"palaces":{"type":"array"},"ming_gong":{"type":"integer"},"si_hua":{"type":"object"},"shen_gong":{"type":"integer"},"ju_shu":{"type":"integer"}},"required":["palaces","ming_gong","si_hua"]}`),
	},
	{
		Name: "ziwei.daxian", Description: "紫微斗数大限。返回十年大限各宫吉凶。chart 为 ziwei.chart 返回的完整 chart 对象。",
		Params: mustSchema(`{"type":"object","properties":{"chart":{"type":"object","description":"ziwei.chart 返回的完整 chart 对象"}},"required":["chart"]}`),
		Handler: ziweiDaxianHandler,
		Result:  envelopeSchema(`{"type":"array","items":{"type":"object","properties":{"start_age":{"type":"integer"},"end_age":{"type":"integer"},"palace":{"type":"integer"},"name":{"type":"string"}},"required":["start_age","end_age","palace","name"]}}`),
	},
	{
		Name: "ziwei.liunian", Description: "紫微流年。返回流年命盘及各宫变化。",
		Params: mustSchema(`{"type":"object","properties":{"liu_year":{"type":"integer","description":"流年年份"},"chart":{"type":"object","description":"ziwei.chart 返回的完整 chart 对象"}},"required":["liu_year","chart"]}`),
		Handler: ziweiLiunianHandler,
		Result:  envelopeSchema(`{"type":"object","properties":{"ming_gong":{"type":"integer"},"ming_gong_name":{"type":"string"},"si_hua":{"type":"object"}},"required":["ming_gong","si_hua"]}`),
	},
	{
		Name: "ziwei.liuyue", Description: "紫微流月。返回流月命盘及各宫变化。",
		Params: mustSchema(`{"type":"object","properties":{"liu_year":{"type":"integer","description":"流年年份"},"lunar_month":{"type":"integer","minimum":1,"maximum":12,"description":"农历月份"},"chart":{"type":"object","description":"ziwei.chart 返回的完整 chart 对象"}},"required":["liu_year","lunar_month","chart"]}`),
		Handler: ziweiLiuyueHandler,
		Result:  envelopeSchema(`{"type":"object","properties":{"ming_gong":{"type":"integer"},"ming_gong_name":{"type":"string"},"si_hua":{"type":"object"}},"required":["ming_gong","si_hua"]}`),
	},
	{
		Name: "ziwei.liuri", Description: "紫微流日。返回流日命盘及各宫变化。",
		Params: mustSchema(`{"type":"object","properties":{"liu_year":{"type":"integer","description":"流年年份"},"lunar_month":{"type":"integer","minimum":1,"maximum":12,"description":"农历月份"},"lunar_day":{"type":"integer","minimum":1,"maximum":30,"description":"农历日期"},"chart":{"type":"object","description":"ziwei.chart 返回的完整 chart 对象"}},"required":["liu_year","lunar_month","lunar_day","chart"]}`),
		Handler: ziweiLiuriHandler,
		Result:  envelopeSchema(`{"type":"object","properties":{"ming_gong":{"type":"integer"},"ming_gong_name":{"type":"string"},"si_hua":{"type":"object"}},"required":["ming_gong","si_hua"]}`),
	},
	{
		Name: "ziwei.judgment", Description: "紫微综合盘论断。返回格局、四化、三方四正、综合评级。",
		Params: mustSchema(`{"type":"object","properties":{"chart":{"type":"object","description":"ziwei.chart 返回的完整 chart 对象"}},"required":["chart"]}`),
		Handler: ziweiJudgmentHandler,
		Result:  envelopeSchema(`{"type":"object","properties":{"patterns":{"type":"array"},"si_hua":{"type":"array","description":"四化列表:[{star_id,star_name,type(禄/权/科/忌)}]"},"san_fang":{"type":"array"},"rating":{"type":"string"},"summary":{"type":"string"}},"required":["patterns","rating","summary"]}`),
	},
	{
		Name: "ziwei.bond", Description: "紫微合盘。返回双方命盘交互分析。",
		Params: mustSchema(`{"type":"object","properties":{"a":{"type":"object","description":"甲方紫微盘（ziwei.chart 返回的完整对象）"},"b":{"type":"object","description":"乙方紫微盘（ziwei.chart 返回的完整对象）"}},"required":["a","b"]}`),
		Handler: ziweiBondHandler,
		Result:  envelopeSchema(`{"type":"object","properties":{"a_into_b":{"type":"integer"},"b_into_a":{"type":"integer"},"star_cross":{"type":"array"}},"required":["a_into_b","b_into_a","star_cross"]}`),
	},
}
