package agent

import (
	"context"
	"encoding/json"
	"fmt"

	"liki-engine/internal/engine/ganzhi"
	"liki-engine/internal/engine/tianwen"
	"liki-engine/internal/engine/ziwei"
)

// parseChart unmarshals and validates a chart from raw JSON.
func parseChart(raw json.RawMessage) (ziwei.Chart, error) {
	var chart ziwei.Chart
	if err := json.Unmarshal(raw, &chart); err != nil {
		return chart, fmt.Errorf("parse chart: %w", err)
	}
	return chart, nil
}

func ziweiChartHandler(ctx context.Context, raw json.RawMessage) (json.RawMessage, error) {
	var p struct {
		Lunar  json.RawMessage `json:"lunar"`
		Gender ganzhi.Gender  `json:"gender"`
	}
	if err := json.Unmarshal(raw, &p); err != nil {
		return nil, fmt.Errorf("compute_ziwei: %w", err)
	}
	if err := validateGender(p.Gender); err != nil {
		return nil, fmt.Errorf("compute_ziwei: %w", err)
	}
	var lt tianwen.LunarTime
	if err := json.Unmarshal(p.Lunar, &lt); err != nil {
		return nil, fmt.Errorf("compute_ziwei: parse lunar: %w", err)
	}
	result := ziwei.ComputeChart(lt, p.Gender)
	return wrapResult("ziwei", result)
}

func ziweiDaxianHandler(ctx context.Context, raw json.RawMessage) (json.RawMessage, error) {
	var p struct {
		Chart json.RawMessage `json:"chart"`
	}
	if err := json.Unmarshal(raw, &p); err != nil {
		return nil, fmt.Errorf("compute_ziwei_daxian: %w", err)
	}
	chart, err := parseChart(p.Chart)
	if err != nil {
		return nil, fmt.Errorf("compute_ziwei_daxian: %w", err)
	}
	result := ziwei.ComputeDaXian(chart)
	return wrapResult("ziwei_daxian", result)
}

func ziweiLiunianHandler(ctx context.Context, raw json.RawMessage) (json.RawMessage, error) {
	var p struct {
		LiuNian int             `json:"liu_nian"`
		Chart   json.RawMessage `json:"chart"`
	}
	if err := json.Unmarshal(raw, &p); err != nil {
		return nil, fmt.Errorf("compute_ziwei_liunian: %w", err)
	}
	var chart ziwei.Chart
	if err := json.Unmarshal(p.Chart, &chart); err != nil {
		return nil, fmt.Errorf("compute_ziwei_liunian: parse chart: %w", err)
	}
	result := ziwei.ComputeLiuNian(chart, p.LiuNian)
	return wrapResult("ziwei_liunian", result)
}

func ziweiLiuyueHandler(ctx context.Context, raw json.RawMessage) (json.RawMessage, error) {
	var p struct {
		LiuNian    int             `json:"liu_nian"`
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
	result := ziwei.ComputeLiuYue(chart, p.LiuNian, p.LunarMonth)
	return wrapResult("ziwei_liuyue", result)
}

func ziweiLiuriHandler(ctx context.Context, raw json.RawMessage) (json.RawMessage, error) {
	var p struct {
		LiuNian    int             `json:"liu_nian"`
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
	result := ziwei.ComputeLiuRi(chart, p.LiuNian, p.LunarMonth, p.LunarDay)
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
	chart, err := parseChart(p.Chart)
	if err != nil {
		return nil, fmt.Errorf("ziwei.judgment: %w", err)
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
		Params: mustSchema(`{"type":"object","properties":{"lunar":{"type":"object","properties":{"year":{"type":"integer"},"month":{"type":"integer"},"day":{"type":"integer"},"shichen":{"type":"string","enum":["子","丑","寅","卯","辰","巳","午","未","申","酉","戌","亥"],"description":"时辰（子丑寅卯辰巳午未申酉戌亥）"}},"required":["year","month","day","shichen"]},"gender":` + schemaGender + `},"required":["lunar","gender"]}`), Handler: ziweiChartHandler,
		Result: envelopeSchema(`{"type":"object","properties":{"gong_wei":{"type":"array","description":"12宫信息","items":{"type":"object","properties":{"index":{"type":"string","description":"宫名"},"name":{"type":"string","description":"宫名"},"gan":{"type":"string","description":"宫干"},"zhi":{"type":"string","description":"宫支"},"is_shen_gong":{"type":"boolean","description":"是否身宫"},"is_yuan_gong":{"type":"boolean","description":"是否来因宫"},"xing_yao":{"type":"array","description":"星曜列表"},"zi_wei":{"type":"string","description":"紫微星（本宫）"},"ages":{"type":"array","items":{"type":"integer"},"description":"小限年龄"},"chang_sheng":{"type":"string","description":"长生"},"bo_shi":{"type":"string","description":"博士"},"jiang_qian":{"type":"string","description":"将前"},"sui_qian":{"type":"string","description":"岁前"},"za_yao":{"type":"array","items":{"type":"string"},"description":"杂曜"}},"required":["index","name","gan","zhi","xing_yao"]}},"ming_gong":{"type":"string","enum":["命宫","兄弟","夫妻","子女","财帛","疾厄","迁移","仆役","官禄","田宅","福德","父母"]},"si_hua":{"type":"object","description":"四化: {星名: 化名}"},"shen_gong":{"type":"string","enum":["命宫","兄弟","夫妻","子女","财帛","疾厄","迁移","仆役","官禄","田宅","福德","父母"]},"ju_shu":{"type":"string","enum":["水二局","木三局","金四局","土五局","火六局"]},"ju_shu_name":{"type":"string"},"ming_zhu":{"type":"string","description":"命主"},"shen_zhu":{"type":"string","description":"身主"},"nian_gan":{"type":"string","enum":["甲","乙","丙","丁","戊","己","庚","辛","壬","癸"],"description":"年干"},"nian_zhi":{"type":"string","enum":["子","丑","寅","卯","辰","巳","午","未","申","酉","戌","亥"],"description":"年支"},"shi_zhi":{"type":"string","enum":["子","丑","寅","卯","辰","巳","午","未","申","酉","戌","亥"],"description":"时支"},"birth_year":{"type":"integer"},"ziwei_pos":{"type":"string","enum":["命宫","兄弟","夫妻","子女","财帛","疾厄","迁移","仆役","官禄","田宅","福德","父母"]},"gender":{"type":"string"},"birth_lunar_month":{"type":"integer","description":"出生农历月"},"lunar_month":{"type":"integer","description":"排盘用农历月（闰月后半月按下月）"},"lunar_day":{"type":"integer","description":"排盘用农历日"},"patterns":{"type":"array","items":{"type":"object","properties":{"name":{"type":"string","description":"格局名"},"description":{"type":"string"},"score":{"type":"integer"}},"required":["name","description","score"]},"description":"命盘格局（如辅弼拱主/禄马交驰）"}},"required":["gong_wei"]}`),
	},
	{
		Name: "ziwei.daxian", Description: "紫微斗数大限。返回十年大限各宫吉凶。chart 为 ziwei.chart 返回的完整 chart 对象。",
		Params: mustSchema(`{"type":"object","properties":{"chart":{"type":"object","description":"ziwei.chart 返回的完整 chart 对象"}},"required":["chart"]}`),
		Handler: ziweiDaxianHandler,
		Result:  envelopeSchema(`{"type":"array","items":{"type":"object","properties":{"qi_sui":{"type":"integer"},"zhi_sui":{"type":"integer"},"gong":{"type":"string","enum":["命宫","兄弟","夫妻","子女","财帛","疾厄","迁移","仆役","官禄","田宅","福德","父母"]},"name":{"type":"string"}},"required":["qi_sui","zhi_sui","gong","name"]}}`),
	},
	{
		Name: "ziwei.liunian", Description: "紫微流年。返回流年命盘及各宫变化。",
		Params: mustSchema(`{"type":"object","properties":{"liu_nian":{"type":"integer","description":"流年年份"},"chart":{"type":"object","description":"ziwei.chart 返回的完整 chart 对象"}},"required":["liu_nian","chart"]}`),
		Handler: ziweiLiunianHandler,
		Result:  envelopeSchema(`{"type":"object","properties":{"ming_gong":{"type":"string","enum":["命宫","兄弟","夫妻","子女","财帛","疾厄","迁移","仆役","官禄","田宅","福德","父母"]},"ming_gong_name":{"type":"string"},"zhi":{"type":"string","description":"流年地支","enum":["子","丑","寅","卯","辰","巳","午","未","申","酉","戌","亥"]},"si_hua":{"type":"object","additionalProperties":{"type":"string","enum":["禄","权","科","忌"]},"description":"四化：{星名: 禄/权/科/忌}"},"si_hua_gong":{"type":"object","description":"四化落宫"},"fu_xing":{"type":"object","description":"流年辅星"},"gong_wei":{"type":"array","description":"流年十二宫盘","items":{"type":"object","properties":{"zhi":{"type":"string","description":"地支"},"name":{"type":"string","description":"宫名"},"xing_yao":{"type":"array","description":"流耀"},"is_liu_ming":{"type":"boolean"}},"required":["zhi","name"]}}},"required":["ming_gong","zhi","si_hua"]}`),
	},
	{
		Name: "ziwei.liuyue", Description: "紫微流月。返回流月命盘及各宫变化。",
		Params: mustSchema(`{"type":"object","properties":{"liu_nian":{"type":"integer","description":"流年年份"},"lunar_month":{"type":"integer","minimum":1,"maximum":12,"description":"农历月份"},"chart":{"type":"object","description":"ziwei.chart 返回的完整 chart 对象"}},"required":["liu_nian","lunar_month","chart"]}`),
		Handler: ziweiLiuyueHandler,
		Result:  envelopeSchema(`{"type":"object","properties":{"ming_gong":{"type":"string","enum":["命宫","兄弟","夫妻","子女","财帛","疾厄","迁移","仆役","官禄","田宅","福德","父母"]},"ming_gong_name":{"type":"string"},"zhi":{"type":"string","description":"流月地支","enum":["子","丑","寅","卯","辰","巳","午","未","申","酉","戌","亥"]},"si_hua":{"type":"object","additionalProperties":{"type":"string","enum":["禄","权","科","忌"]},"description":"四化：{星名: 禄/权/科/忌}"},"xing_yao":{"type":"object","description":"月星: {星名: zhiIdx}"},"gong_wei":{"type":"array","description":"流月十二宫盘","items":{"type":"object","properties":{"zhi":{"type":"string"},"name":{"type":"string"},"xing_yao":{"type":"array"},"is_liu_ming":{"type":"boolean"}}}}},"required":["ming_gong","zhi","si_hua"]}`),
	},
	{
		Name: "ziwei.liuri", Description: "紫微流日。返回流日命盘及各宫变化。",
		Params: mustSchema(`{"type":"object","properties":{"liu_nian":{"type":"integer","description":"流年年份"},"lunar_month":{"type":"integer","minimum":1,"maximum":12,"description":"农历月份"},"lunar_day":{"type":"integer","minimum":1,"maximum":30,"description":"农历日期"},"chart":{"type":"object","description":"ziwei.chart 返回的完整 chart 对象"}},"required":["liu_nian","lunar_month","lunar_day","chart"]}`),
		Handler: ziweiLiuriHandler,
		Result:  envelopeSchema(`{"type":"object","properties":{"ming_gong":{"type":"string","enum":["命宫","兄弟","夫妻","子女","财帛","疾厄","迁移","仆役","官禄","田宅","福德","父母"]},"ming_gong_name":{"type":"string"},"zhi":{"type":"string","description":"流日地支","enum":["子","丑","寅","卯","辰","巳","午","未","申","酉","戌","亥"]},"si_hua":{"type":"object","additionalProperties":{"type":"string","enum":["禄","权","科","忌"]},"description":"四化：{星名: 禄/权/科/忌}"},"xing_yao":{"type":"object","description":"日星: {星名: zhiIdx}"},"gong_wei":{"type":"array","description":"流日十二宫盘","items":{"type":"object","properties":{"zhi":{"type":"string"},"name":{"type":"string"},"xing_yao":{"type":"array"},"is_liu_ming":{"type":"boolean"}}}}},"required":["ming_gong","zhi","si_hua"]}`),
	},
	{
		Name: "ziwei.judgment", Description: "紫微综合盘论断。返回格局、四化、三方四正、综合评级。",
		Params: mustSchema(`{"type":"object","properties":{"chart":{"type":"object","description":"ziwei.chart 返回的完整 chart 对象"}},"required":["chart"]}`),
		Handler: ziweiJudgmentHandler,
		Result:  envelopeSchema(`{"type":"object","properties":{"patterns":{"type":"array","items":{"type":"object","properties":{"name":{"type":"string"},"description":{"type":"string"},"score":{"type":"integer"}}}},"si_hua":{"type":"array","description":"四化列表:[{star_id,star_name,type(禄/权/科/忌)}]"},"san_fang":{"type":"array","items":{"type":"object","properties":{"name":{"type":"string"},"zhu_xing":{"type":"array"},"fu_xing":{"type":"array"},"si_hua":{"type":"string"}}}},"rating":{"type":"string"},"summary":{"type":"string"},"rule":{"type":"integer","description":"论断规则版本"}},"required":["patterns","rating","summary"]}`),
	},
	{
		Name: "ziwei.bond", Description: "紫微合盘。返回双方命盘交互分析。",
		Params: mustSchema(`{"type":"object","properties":{"a":{"type":"object","description":"甲方紫微盘（ziwei.chart 返回的完整对象）"},"b":{"type":"object","description":"乙方紫微盘（ziwei.chart 返回的完整对象）"}},"required":["a","b"]}`),
		Handler: ziweiBondHandler,
		Result:  envelopeSchema(`{"type":"object","properties":{"jia_gong_name":{"type":"string","description":"甲宫名"},"yi_gong_name":{"type":"string","description":"乙宫名"},"ming_gong_hu_ru":{"type":"object"},"ming_gong_hu_ru":{"type":"object"},"fu_qi_gong":{"type":"object"},"zi_nv_gong":{"type":"object"},"ji_xing":{"type":"array"},"sha_xing":{"type":"array"},"lu_ma_ru":{"type":"array","description":"禄马入"},"kong_wang":{"type":"array"},"si_hua_ru":{"type":"array","description":"四化入"},"wu_xing_sheng_ke":{"type":"string"},"summary":{"type":"string"}},"required":["ming_gong_hu_ru","wu_xing_sheng_ke"]}`),
	},
	{
		Name: "ziwei.fullchart", Description: "紫微全盘。扩展杂曜、长生、博士、小限、将前、岁前。",
		Handler: ziweiFullChartHandler,
		Params: mustSchema(`{"type":"object","properties":{"chart":{"type":"object"}},"required":["chart"]}`),
		Result:  envelopeSchema(`{"type":"object","properties":{"gong_wei":{"type":"array","description":"12宫含杂曜/长生/博士/小限/将前/岁前","items":{"type":"object","properties":{"index":{"type":"string","description":"宫名"},"name":{"type":"string","description":"宫名"},"gan":{"type":"string","description":"宫干"},"zhi":{"type":"string","description":"宫支"},"is_shen_gong":{"type":"boolean","description":"是否身宫"},"is_yuan_gong":{"type":"boolean","description":"是否来因宫"},"xing_yao":{"type":"array","description":"星曜列表"},"zi_wei":{"type":"string","description":"紫微星（本宫）"},"ages":{"type":"array","items":{"type":"integer"},"description":"小限年龄"},"chang_sheng":{"type":"string","description":"长生"},"bo_shi":{"type":"string","description":"博士"},"jiang_qian":{"type":"string","description":"将前"},"sui_qian":{"type":"string","description":"岁前"},"za_yao":{"type":"array","items":{"type":"string"},"description":"杂曜"}},"required":["index","name","gan","zhi","xing_yao"]}},"ming_gong":{"type":"string","enum":["命宫","兄弟","夫妻","子女","财帛","疾厄","迁移","仆役","官禄","田宅","福德","父母"]},"si_hua":{"type":"object","additionalProperties":{"type":"string","enum":["禄","权","科","忌"]},"description":"四化：{星名: 禄/权/科/忌}"},"shen_gong":{"type":"string","enum":["命宫","兄弟","夫妻","子女","财帛","疾厄","迁移","仆役","官禄","田宅","福德","父母"]},"ju_shu":{"type":"string","enum":["水二局","木三局","金四局","土五局","火六局"]},"ming_zhu":{"type":"string"},"shen_zhu":{"type":"string"},"nian_gan":{"type":"string","enum":["甲","乙","丙","丁","戊","己","庚","辛","壬","癸"],"description":"年干"},"nian_zhi":{"type":"string","enum":["子","丑","寅","卯","辰","巳","午","未","申","酉","戌","亥"],"description":"年支"},"shi_zhi":{"type":"string","enum":["子","丑","寅","卯","辰","巳","午","未","申","酉","戌","亥"],"description":"时支"},"birth_year":{"type":"integer"},"birth_lunar_month":{"type":"integer","description":"出生农历月"},"lunar_month":{"type":"integer","description":"排盘用农历月"},"lunar_day":{"type":"integer","description":"排盘用农历日"},"ju_shu_name":{"type":"string"},"gender":{"type":"string"},"ziwei_pos":{"type":"string"},"patterns":{"type":"array","description":"命盘格局"}},"required":["gong_wei"]}`),
	},
	{
		Name: "ziwei.liushi", Description: "紫微流时。返回流时命宫及四化。",
		Handler: ziweiLiuShiHandler,
		Params: mustSchema(`{"type":"object","properties":{"liu_nian":{"type":"integer"},"lunar_month":{"type":"integer"},"lunar_day":{"type":"integer"},"shi_zhi":{"type":"integer"},"chart":{"type":"object"}},"required":["liu_nian","lunar_month","lunar_day","shi_zhi","chart"]}`),
		Result:  envelopeSchema(`{"type":"object","properties":{"ming_gong":{"type":"string","enum":["命宫","兄弟","夫妻","子女","财帛","疾厄","迁移","仆役","官禄","田宅","福德","父母"]},"ming_gong_name":{"type":"string"},"zhi":{"type":"string","description":"流时地支","enum":["子","丑","寅","卯","辰","巳","午","未","申","酉","戌","亥"]},"si_hua":{"type":"object","additionalProperties":{"type":"string","enum":["禄","权","科","忌"]},"description":"四化：{星名: 禄/权/科/忌}"},"xing_yao":{"type":"object","description":"时星: {星名: zhiIdx}"},"gong_wei":{"type":"array","description":"流时十二宫盘","items":{"type":"object","properties":{"zhi":{"type":"string"},"name":{"type":"string"},"xing_yao":{"type":"array"},"is_liu_ming":{"type":"boolean"}}}}},"required":["ming_gong","zhi","si_hua"]}`),
	},
}

func ziweiLiuShiHandler(ctx context.Context, raw json.RawMessage) (json.RawMessage, error) {
	var p struct {
		LiuNian    int             `json:"liu_nian"`
		LunarMonth int             `json:"lunar_month"`
		LunarDay   int             `json:"lunar_day"`
		ShiZhi     int             `json:"shi_zhi"`
		Chart      json.RawMessage `json:"chart"`
	}
	if err := json.Unmarshal(raw, &p); err != nil {
		return nil, fmt.Errorf("ziwei.liushi: %w", err)
	}
	chart, err := parseChart(p.Chart)
	if err != nil {
		return nil, fmt.Errorf("ziwei.liushi: %w", err)
	}
	result := ziwei.ComputeLiuShi(chart, p.LiuNian, p.LunarMonth, p.LunarDay, ganzhi.Zhi(p.ShiZhi))
	return wrapResult("ziwei_liushi", result)
}
