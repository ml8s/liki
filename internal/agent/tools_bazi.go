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
		SolarTime string        `json:"solar_time"`
		Gender    ganzhi.Gender `json:"gender"`
		Longitude *float64      `json:"longitude,omitempty"`
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
	// 真太阳时校正：有 longitude 时，先用经度校正时刻再排盘（所有命例都应校正）
	if p.Longitude != nil {
		ts := tianwen.ComputeTimeset(tianwen.GregorianTime(st.Time()), *p.Longitude)
		st = ts.Solar
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
	var p struct {
		Chart json.RawMessage `json:"chart"`
	}
	if err := json.Unmarshal(raw, &p); err != nil {
		return nil, fmt.Errorf("compute_yongshen: %w", err)
	}
	var c bazi.Chart
	if err := json.Unmarshal(p.Chart, &c); err != nil {
		return nil, fmt.Errorf("parse chart: %w", err)
	}
	if c.Ri.Gan == 0 {
		return nil, fmt.Errorf("chart has empty day stem")
	}
	result := bazi.ComputeYongShen(c)
	return wrapResult("yongshen", result)
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
	var coreA bazi.Chart
	if err := json.Unmarshal(p.A.Chart, &coreA); err != nil {
		return nil, fmt.Errorf("parse chart: %w", err)
	}
	if coreA.Ri.Gan == 0 {
		return nil, fmt.Errorf("chart has empty day stem")
	}
	var coreB bazi.Chart
	if err := json.Unmarshal(p.B.Chart, &coreB); err != nil {
		return nil, fmt.Errorf("parse chart: %w", err)
	}
	if coreB.Ri.Gan == 0 {
		return nil, fmt.Errorf("chart has empty day stem")
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
	var core bazi.Chart
	if err := json.Unmarshal(p.Chart, &core); err != nil {
		return nil, fmt.Errorf("parse chart: %w", err)
	}
	if core.Ri.Gan == 0 {
		return nil, fmt.Errorf("chart has empty day stem")
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
	var core bazi.Chart
	if err := json.Unmarshal(p.Chart, &core); err != nil {
		return nil, fmt.Errorf("parse chart: %w", err)
	}
	if core.Ri.Gan == 0 {
		return nil, fmt.Errorf("chart has empty day stem")
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
	var core bazi.Chart
	if err := json.Unmarshal(p.Chart, &core); err != nil {
		return nil, fmt.Errorf("parse chart: %w", err)
	}
	if core.Ri.Gan == 0 {
		return nil, fmt.Errorf("chart has empty day stem")
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
	var core bazi.Chart
	if err := json.Unmarshal(p.Chart, &core); err != nil {
		return nil, fmt.Errorf("parse chart: %w", err)
	}
	if core.Ri.Gan == 0 {
		return nil, fmt.Errorf("chart has empty day stem")
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
	var core bazi.Chart
	if err := json.Unmarshal(p.Chart, &core); err != nil {
		return nil, fmt.Errorf("parse chart: %w", err)
	}
	if core.Ri.Gan == 0 {
		return nil, fmt.Errorf("chart has empty day stem")
	}
	result := bazi.ComputeXiaoYun(core, p.Count)
	return wrapResult("xiaoyun", result)
}

func baziXiaoXianHandler(ctx context.Context, raw json.RawMessage) (json.RawMessage, error) {
	var p struct {
		Gender ganzhi.Gender `json:"gender"`
		Count  int           `json:"count"`
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
		Name: "bazi.fullchart", Description: "完整命盘。传入 bazi.chart 返回的最小命盘，补全十神/藏干/神煞/长生/空亡/自合/魁罡/三元/拱夹/纳音生克/长生十二宫/三奇贵人/合会冲刑。",
		Params:  mustSchema(`{"type":"object","properties":{"chart":{"type":"object","description":"bazi.chart 返回的最小命盘"}},"required":["chart"]}`),
		Handler: baziFullChartHandler,
		Result:  envelopeSchema(`{"type":"object","properties":{"nian":{"type":"object","description":"年柱: gan/zhi/na_yin","properties":{"gan":{"type":"string","enum":["甲","乙","丙","丁","戊","己","庚","辛","壬","癸"]},"zhi":{"type":"string","enum":["子","丑","寅","卯","辰","巳","午","未","申","酉","戌","亥"]},"na_yin":{"type":"string","cang_gan":{"type":"object","description":"藏干"},"shi_shens":{"type":"array","description":"十神"},"chang_sheng":{"type":"array","description":"长生"},"shen_sha":{"type":"array","description":"神煞"},"is_void":{"type":"boolean","description":"旬空"},"is_self_he":{"type":"boolean","description":"自合"},"is_kui_gang":{"type":"boolean","description":"魁罡"},"self_he_name":{"type":"string","description":"自合名"}}},"required":["gan","zhi","na_yin"]},"yue":{"type":"object","description":"月柱: gan/zhi/na_yin","properties":{"gan":{"type":"string","enum":["甲","乙","丙","丁","戊","己","庚","辛","壬","癸"]},"zhi":{"type":"string","enum":["子","丑","寅","卯","辰","巳","午","未","申","酉","戌","亥"]},"na_yin":{"type":"string","cang_gan":{"type":"object","description":"藏干"},"shi_shens":{"type":"array","description":"十神"},"chang_sheng":{"type":"array","description":"长生"},"shen_sha":{"type":"array","description":"神煞"},"is_void":{"type":"boolean","description":"旬空"},"is_self_he":{"type":"boolean","description":"自合"},"is_kui_gang":{"type":"boolean","description":"魁罡"},"self_he_name":{"type":"string","description":"自合名"}}},"required":["gan","zhi","na_yin"]},"ri":{"type":"object","description":"日柱: gan/zhi/na_yin","properties":{"gan":{"type":"string","enum":["甲","乙","丙","丁","戊","己","庚","辛","壬","癸"]},"zhi":{"type":"string","enum":["子","丑","寅","卯","辰","巳","午","未","申","酉","戌","亥"]},"na_yin":{"type":"string","cang_gan":{"type":"object","description":"藏干"},"shi_shens":{"type":"array","description":"十神"},"chang_sheng":{"type":"array","description":"长生"},"shen_sha":{"type":"array","description":"神煞"},"is_void":{"type":"boolean","description":"旬空"},"is_self_he":{"type":"boolean","description":"自合"},"is_kui_gang":{"type":"boolean","description":"魁罡"},"self_he_name":{"type":"string","description":"自合名"}}},"required":["gan","zhi","na_yin"]},"shi":{"type":"object","description":"时柱: gan/zhi/na_yin","properties":{"gan":{"type":"string","enum":["甲","乙","丙","丁","戊","己","庚","辛","壬","癸"]},"zhi":{"type":"string","enum":["子","丑","寅","卯","辰","巳","午","未","申","酉","戌","亥"]},"na_yin":{"type":"string","cang_gan":{"type":"object","description":"藏干"},"shi_shens":{"type":"array","description":"十神"},"chang_sheng":{"type":"array","description":"长生"},"shen_sha":{"type":"array","description":"神煞"},"is_void":{"type":"boolean","description":"旬空"},"is_self_he":{"type":"boolean","description":"自合"},"is_kui_gang":{"type":"boolean","description":"魁罡"},"self_he_name":{"type":"string","description":"自合名"}}},"required":["gan","zhi","na_yin"]},"da_yun":{"type":"object","description":"大运。start_year_after/start_month_after/start_day_after 为出生后起运精确时间；direction 顺排/逆排；steps 每步含 gan/zhi/qi_sui/zhi_sui/name/wuxing/shi_shen（十年一运）；current_step_index 当前大运索引"},"gender":{"type":"string"},"birth_year":{"type":"integer","description":"出生公历年份"},"san_yuan":{"type":"object","description":"三元: 胎元/命宫/身宫"},"gong_jia":{"type":"array","description":"拱夹"},"nayin_rel":{"type":"array","description":"纳音生克"},"chang_sheng":{"type":"array","description":"长生十二宫"},"san_qi_name":{"type":"string","description":"三奇贵人"},"gan_he":{"type":"array","description":"天干五合"},"zhi_liu_he":{"type":"array","description":"地支六合"},"san_he":{"type":"array","description":"三合局"},"san_hui":{"type":"array","description":"三会方"},"liu_chong":{"type":"array","description":"六冲"},"liu_hai":{"type":"array","description":"六害"},"liu_xing":{"type":"array","description":"相刑"}},"required":["nian","yue","ri","shi","da_yun","gender"]}`),
	},
	{
		Name: "bazi.chart", Description: "排八字命盘。返回最小命盘（四柱+纳音+大运+性别）。如需十神/藏干/神煞/长生/空亡等完整信息，请将结果传入 bazi.fullchart。用神需另行调用 bazi.yongshen。",
		Params: mustSchema(`{"type":"object","properties":{"solar_time":` + schemaSolarTime + `,"gender":` + schemaGender + `,"longitude":{"type":"number","description":"出生地经度（度），用于真太阳时校正。缺省用东八区 120。所有命例都应传实际经度，海外/西部命例必传否则时辰错"}},"required":["solar_time","gender"]}`), Handler: baziChartHandler,
		Result: envelopeSchema(`{"type":"object","properties":{"nian":{"type":"object","description":"年柱: gan/zhi/na_yin","properties":{"gan":{"type":"string","enum":["甲","乙","丙","丁","戊","己","庚","辛","壬","癸"]},"zhi":{"type":"string","enum":["子","丑","寅","卯","辰","巳","午","未","申","酉","戌","亥"]},"na_yin":{"type":"string"}},"required":["gan","zhi","na_yin"]},"yue":{"type":"object","description":"月柱: gan/zhi/na_yin","properties":{"gan":{"type":"string","enum":["甲","乙","丙","丁","戊","己","庚","辛","壬","癸"]},"zhi":{"type":"string","enum":["子","丑","寅","卯","辰","巳","午","未","申","酉","戌","亥"]},"na_yin":{"type":"string"}},"required":["gan","zhi","na_yin"]},"ri":{"type":"object","description":"日柱: gan/zhi/na_yin","properties":{"gan":{"type":"string","enum":["甲","乙","丙","丁","戊","己","庚","辛","壬","癸"]},"zhi":{"type":"string","enum":["子","丑","寅","卯","辰","巳","午","未","申","酉","戌","亥"]},"na_yin":{"type":"string"}},"required":["gan","zhi","na_yin"]},"shi":{"type":"object","description":"时柱: gan/zhi/na_yin","properties":{"gan":{"type":"string","enum":["甲","乙","丙","丁","戊","己","庚","辛","壬","癸"]},"zhi":{"type":"string","enum":["子","丑","寅","卯","辰","巳","午","未","申","酉","戌","亥"]},"na_yin":{"type":"string"}},"required":["gan","zhi","na_yin"]},"da_yun":{"type":"object","description":"大运。start_year_after/start_month_after/start_day_after 为出生后起运精确时间（3天=1年）；direction 顺排/逆排；steps 每步含 gan/zhi/qi_sui/zhi_sui/name/wuxing/shi_shen（十年一运）；current_step_index 为当前所处大运索引（-1 表示未起运或已过完所有大运）"},"gender":{"type":"string"},"birth_year":{"type":"integer","description":"出生公历年份（供 bazi.liunian/liuri 定位当年大运）"}},"required":["nian","yue","ri","shi","da_yun","gender"]}`),
	},
	{
		Name: "bazi.yongshen", Description: "八字用神分析。基于扶抑（旺衰）、调候（穷通宝鉴）、格局（子平）三派计算用神/喜神/忌神。返回五行计数、旺相休囚死、三派用神。LLM 综合三派结果判断最终用神。注意：三派结果分别在 fu_yi / tiao_hou / ge_ju 子对象中，顶层 shen / yongshen / xishen / jishen 字段不存在，请直接引用子对象内的 yong / xi / ji。",
		Params:  mustSchema(`{"type":"object","properties":{"chart":{"type":"object","description":"八字命盘（由 bazi.chart 返回的 data）"}},"required":["chart"]}`),
		Handler: baziYongShenHandler,
		Result:  envelopeSchema(`{"type":"object","properties":{"fu_yi":{"type":"object","properties":{"wuxing_count":{"type":"object"},"wang_shuai":{"type":"object"},"yong":{"type":"string"},"xi":{"type":"string"},"ji":{"type":"string"},"qiangruo":{"type":"string","enum":["身强","身弱","中和"]}},"required":["wuxing_count","wang_shuai","yong","xi","ji","qiangruo"]},"tiao_hou":{"type":"object","properties":{"yong":{"type":"string"},"xi":{"type":"string"},"ji":{"type":"string"},"season":{"type":"string"},"detail":{"type":"string"}},"required":["yong","xi","ji","season"]},"ge_ju":{"type":"object","properties":{"yong":{"type":"string"},"xi":{"type":"string"},"ji":{"type":"string"},"ge_ju":{"type":"string"},"yong_fa":{"type":"string"}},"required":["yong","xi","ji","ge_ju","yong_fa"]}},"required":["fu_yi","tiao_hou","ge_ju"]}`),
	},
	{
		Name: "bazi.bond", Description: "八字合盘。返回双方日主、天干关系（合/生/克）、地支关系（六合/三合/六冲）、纳音配合、五行互补。",
		Params:  mustSchema(`{"type":"object","properties":{"a":{"type":"object","properties":{"chart":{"type":"object","description":"八字命盘（由 bazi.chart 返回的最小命盘，不需 bazi.fullchart）"}},"required":["chart"]},"b":{"type":"object","properties":{"chart":{"type":"object","description":"八字命盘（由 bazi.chart 返回的最小命盘，不需 bazi.fullchart）"}},"required":["chart"]}},"required":["a","b"]}`),
		Handler: baziBondHandler,
		Result:  envelopeSchema(`{"type":"object","properties":{"zhu_cross":{"type":"object","description":"16对四柱互配结果，含jia_zhu/yi_zhu/jia_gan/yi_gan/gan_guan_xi/zhi_guan_xi等"},"shi_shen_cross":{"type":"object","description":"双方日主看对方四柱的十神映射: jia_dui_yi/yi_dui_jia"},"structure":{"type":"object","description":"大运交叉+旬宫判断，含da_yun/xun_gong"},"nayin_cross":{"type":"object","description":"纳音配合"},"shensha_cross":{"type":"object","description":"神煞交叉"}},"required":["zhu_cross","shi_shen_cross","structure"]}`),
	},
	{
		Name: "bazi.liunian", Description: "八字流年运势。返回流年干支与命局的十神、神煞、伏吟反吟。",
		Params:  mustSchema(`{"type":"object","properties":{"year":{"type":"integer","description":"目标年份"},"chart":{"type":"object","description":"八字命盘（由 bazi.chart 返回的最小命盘，不需 bazi.fullchart）"}},"required":["year","chart"]}`),
		Handler: baziLiunianHandler,
		Result:  envelopeSchema(`{"type":"object","properties":{"year":{"type":"integer"},"nian_name":{"type":"string"},"shi_shen":{"type":"string","enum":["正官","偏官","正印","偏印","正财","偏财","比肩","劫财","食神","伤官"]},"nian_gan":{"type":"string","description":"流年天干"},"nian_zhi":{"type":"string","description":"流年地支"},"wuxing":{"type":"string"},"na_yin":{"type":"string"},"sheng":{"type":"integer","description":"生数"},"ke":{"type":"integer","description":"克数"},"shensha":{"type":"array","description":"流年神煞"},"fuyin_fanyin":{"type":"array","description":"伏吟反吟"},"dayun_interactions":{"type":"array","description":"查询年份所在大运与流年交互（chart 需含 birth_year；未起运/大运过完/缺 birth_year 时为空数组）"},"natal_interactions":{"type":"array","description":"流年与命局交互"}},"required":["year","nian_name","shi_shen"]}`),
	},
	{
		Name: "bazi.liuyue", Description: "流月运势。返回流月干支与命局的十神、神煞。",
		Params:  mustSchema(`{"type":"object","properties":{"year":{"type":"integer","description":"目标年份"},"month":{"type":"integer","minimum":1,"maximum":12,"description":"目标月份"},"chart":{"type":"object","description":"八字命盘（由 bazi.chart 返回的最小命盘，不需 bazi.fullchart）"}},"required":["year","month","chart"]}`),
		Handler: baziLiuyueHandler,
		Result:  envelopeSchema(`{"type":"object","properties":{"year":{"type":"integer"},"month":{"type":"integer"},"month_name":{"type":"string"},"shi_shen":{"type":"string","enum":["正官","偏官","正印","偏印","正财","偏财","比肩","劫财","食神","伤官"]},"yue_gan":{"type":"string","description":"流月天干"},"yue_zhi":{"type":"string","description":"流月地支"},"wuxing":{"type":"string"},"sheng":{"type":"integer"},"ke":{"type":"integer"},"shensha":{"type":"array","description":"流月神煞"},"gan_rels":{"type":"array","description":"天干关系"},"zhi_rels":{"type":"array","description":"地支关系"}},"required":["year","month","month_name","shi_shen"]}`),
	},
	{
		Name: "bazi.liuri", Description: "流日运势。返回流日干支、十神、纳音。",
		Params:  mustSchema(`{"type":"object","properties":{"year":{"type":"integer","description":"目标年份"},"month":{"type":"integer","minimum":1,"maximum":12,"description":"目标月份"},"day":{"type":"integer","minimum":1,"maximum":31,"description":"目标日期"},"chart":{"type":"object","description":"八字命盘（由 bazi.chart 返回的最小命盘，不需 bazi.fullchart）"}},"required":["year","month","day","chart"]}`),
		Handler: baziLiuriHandler,
		Result:  envelopeSchema(`{"type":"object","properties":{"date":{"type":"string"},"day_name":{"type":"string"},"shi_shen":{"type":"string","enum":["正官","偏官","正印","偏印","正财","偏财","比肩","劫财","食神","伤官"]},"ri_gan":{"type":"string","description":"流日天干"},"ri_zhi":{"type":"string","description":"流日地支"},"day_nayin":{"type":"string"},"shensha":{"type":"array","description":"流日神煞"},"gan_rels":{"type":"array","description":"天干关系"},"zhi_rels":{"type":"array","description":"地支关系"},"dayun_rels":{"type":"array","description":"与大运关系"},"liunian_rels":{"type":"array","description":"与流年关系"}},"required":["date","day_name","shi_shen"]}`),
	},
	{
		Name: "bazi.liushi", Description: "流时运势。返回流时干支、十神。hour 为时辰（0-23）。",
		Params:  mustSchema(`{"type":"object","properties":{"year":{"type":"integer","description":"目标年份"},"month":{"type":"integer","minimum":1,"maximum":12,"description":"目标月份"},"day":{"type":"integer","minimum":1,"maximum":31,"description":"目标日期"},"hour":{"type":"integer","minimum":0,"maximum":23,"description":"时辰"},"chart":{"type":"object","description":"八字命盘（由 bazi.chart 返回的最小命盘，不需 bazi.fullchart）"}},"required":["year","month","day","hour","chart"]}`),
		Handler: baziLiushiHandler,
		Result:  envelopeSchema(`{"type":"object","properties":{"time":{"type":"string"},"hour_name":{"type":"string"},"shi_shen":{"type":"string","enum":["正官","偏官","正印","偏印","正财","偏财","比肩","劫财","食神","伤官"]},"shi_gan":{"type":"string","description":"流时天干"},"shi_zhi":{"type":"string","description":"流时地支"},"gan_rels":{"type":"array","description":"天干关系"},"zhi_rels":{"type":"array","description":"地支关系"}},"required":["time","hour_name","shi_shen"]}`),
	},
	{
		Name: "bazi.xiaoyun", Description: "小运。返回小运流年列表。count 默认 12。",
		Params:  mustSchema(`{"type":"object","properties":{"chart":{"type":"object","description":"八字命盘（由 bazi.chart 返回的最小命盘，不需 bazi.fullchart）"},"count":{"type":"integer","description":"返回年数，默认 12"}},"required":["chart"]}`),
		Handler: baziXiaoYunHandler,
		Result:  envelopeSchema(`{"type":"array","items":{"type":"object","properties":{"age":{"type":"integer"},"gan":{"type":"string","enum":["甲","乙","丙","丁","戊","己","庚","辛","壬","癸"]},"zhi":{"type":"string","enum":["子","丑","寅","卯","辰","巳","午","未","申","酉","戌","亥"]},"name":{"type":"string"}},"required":["age","gan","zhi","name"]}}`),
	},
	{
		Name: "bazi.xiaoxian", Description: "小限。返回小限列表。count 默认 12。",
		Params:  mustSchema(`{"type":"object","properties":{"gender":` + schemaGender + `,"count":{"type":"integer","description":"返回年数，默认 12"}},"required":["gender"]}`),
		Handler: baziXiaoXianHandler,
		Result:  envelopeSchema(`{"type":"array","items":{"type":"object","properties":{"age":{"type":"integer"},"zhi":{"type":"string"}},"required":["age","zhi"]}}`),
	},
}
