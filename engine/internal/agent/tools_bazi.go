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

func validateFourPillarChart(chart bazi.Chart) error {
	pillars := []struct {
		name string
		gan  ganzhi.Gan
		zhi  ganzhi.Zhi
	}{
		{"nian", chart.Nian.Gan, chart.Nian.Zhi},
		{"yue", chart.Yue.Gan, chart.Yue.Zhi},
		{"ri", chart.Ri.Gan, chart.Ri.Zhi},
		{"shi", chart.Shi.Gan, chart.Shi.Zhi},
	}
	for _, pillar := range pillars {
		if pillar.gan == 0 || pillar.zhi == 0 {
			return fmt.Errorf("bazi.fullchart: chart.%s requires non-empty gan and zhi", pillar.name)
		}
	}
	if chart.Gender != ganzhi.Male && chart.Gender != ganzhi.Female {
		return fmt.Errorf("bazi.fullchart: chart.gender must be male or female")
	}
	return nil
}

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
	if err := validateFourPillarChart(chart); err != nil {
		return nil, err
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
		return nil, fmt.Errorf("bazi.chart: %w", err)
	}
	if err := validateGender(p.Gender); err != nil {
		return nil, fmt.Errorf("bazi.chart: %w", err)
	}
	st, err := parseSolarTime(p.SolarTime)
	if err != nil {
		return nil, fmt.Errorf("bazi.chart: %w", err)
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
		result.DaYun.CurrentStepIndex = bazi.ComputeCurrentStepIndex(result.DaYun, now.Year())
		// 大运临界：距下一大运剩余年数 = 当前步 EndYear − 当前公历年（日期段直判，免虚岁换算）。
		if idx := result.DaYun.CurrentStepIndex; idx >= 0 && idx < len(result.DaYun.Steps) {
			if end := result.DaYun.Steps[idx].EndYear; now.Year() <= end {
				next := end - now.Year()
				result.DaYun.NextStepInYears = &next
			}
		}
	}
	return wrapResult("chart", result)
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
		return nil, fmt.Errorf("bazi.bond: %w", err)
	}
	var coreA bazi.Chart
	if err := json.Unmarshal(p.A.Chart, &coreA); err != nil {
		return nil, fmt.Errorf("bazi.bond: parse chart: %w", err)
	}
	if coreA.Ri.Gan == 0 {
		return nil, fmt.Errorf("bazi.bond: chart has empty day gan")
	}
	var coreB bazi.Chart
	if err := json.Unmarshal(p.B.Chart, &coreB); err != nil {
		return nil, fmt.Errorf("bazi.bond: parse chart: %w", err)
	}
	if coreB.Ri.Gan == 0 {
		return nil, fmt.Errorf("bazi.bond: chart has empty day gan")
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
		return nil, fmt.Errorf("bazi.liunian: %w", err)
	}
	if p.Year <= 0 {
		return nil, fmt.Errorf("bazi.liunian: year must be positive, got %d", p.Year)
	}
	var core bazi.Chart
	if err := json.Unmarshal(p.Chart, &core); err != nil {
		return nil, fmt.Errorf("bazi.liunian: parse chart: %w", err)
	}
	if core.Ri.Gan == 0 {
		return nil, fmt.Errorf("bazi.liunian: chart has empty day gan")
	}
	result, err := bazi.ComputeLiuNian(core, p.Year)
	if err != nil {
		return nil, fmt.Errorf("bazi.liunian: %w", err)
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
		return nil, fmt.Errorf("bazi.liuyue: %w", err)
	}
	var core bazi.Chart
	if err := json.Unmarshal(p.Chart, &core); err != nil {
		return nil, fmt.Errorf("bazi.liuyue: parse chart: %w", err)
	}
	if core.Ri.Gan == 0 {
		return nil, fmt.Errorf("bazi.liuyue: chart has empty day gan")
	}
	result, err := bazi.ComputeLiuYue(core, p.Year, p.Month)
	if err != nil {
		return nil, fmt.Errorf("bazi.liuyue: %w", err)
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
		return nil, fmt.Errorf("bazi.liuri: %w", err)
	}
	var core bazi.Chart
	if err := json.Unmarshal(p.Chart, &core); err != nil {
		return nil, fmt.Errorf("bazi.liuri: parse chart: %w", err)
	}
	if core.Ri.Gan == 0 {
		return nil, fmt.Errorf("bazi.liuri: chart has empty day gan")
	}
	result, err := bazi.ComputeLiuRi(core, p.Year, p.Month, p.Day)
	if err != nil {
		return nil, fmt.Errorf("bazi.liuri: %w", err)
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
		return nil, fmt.Errorf("bazi.liushi: %w", err)
	}
	if p.Hour < 0 || p.Hour > 23 {
		return nil, fmt.Errorf("bazi.liushi: hour must be 0-23, got %d", p.Hour)
	}
	var core bazi.Chart
	if err := json.Unmarshal(p.Chart, &core); err != nil {
		return nil, fmt.Errorf("bazi.liushi: parse chart: %w", err)
	}
	if core.Ri.Gan == 0 {
		return nil, fmt.Errorf("bazi.liushi: chart has empty day gan")
	}
	result, err := bazi.ComputeLiuShi(core, p.Year, p.Month, p.Day, p.Hour)
	if err != nil {
		return nil, fmt.Errorf("bazi.liushi: %w", err)
	}
	return wrapResult("liushi", result)
}

func baziXiaoYunHandler(ctx context.Context, raw json.RawMessage) (json.RawMessage, error) {
	var p struct {
		Chart json.RawMessage `json:"chart"`
		Count int             `json:"count"`
	}
	if err := json.Unmarshal(raw, &p); err != nil {
		return nil, fmt.Errorf("bazi.xiaoyun: %w", err)
	}
	var core bazi.Chart
	if err := json.Unmarshal(p.Chart, &core); err != nil {
		return nil, fmt.Errorf("parse chart: %w", err)
	}
	if core.Ri.Gan == 0 {
		return nil, fmt.Errorf("chart has empty day gan")
	}
	result := bazi.ComputeXiaoYun(core, p.Count)
	return wrapResult("xiaoyun", result)
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
		Name: "bazi.fullchart", Description: "完整命盘。传入 bazi.chart 返回的最小命盘，补全十神/藏干/神煞/长生/空亡/自合/魁罡/三元/拱夹/纳音生克/长生十二宫/三奇贵人/合会冲刑/十干禄/命理原子事实。",
		Params:  mustSchema(`{"type":"object","properties":{"chart":{"type":"object","description":"bazi.chart 返回的最小命盘"}},"required":["chart"]}`),
		Handler: baziFullChartHandler,
		Result:  envelopeSchema(`{"type":"object","properties":{"nian":{"type":"object","description":"年柱: gan/zhi/na_yin","properties":{"gan":{"type":"string","enum":["甲","乙","丙","丁","戊","己","庚","辛","壬","癸"]},"zhi":{"type":"string","enum":["子","丑","寅","卯","辰","巳","午","未","申","酉","戌","亥"]},"na_yin":{"type":"string"},"cang_gan":{"type":"object","description":"藏干"},"shi_shens":{"type":"array","description":"十神"},"chang_sheng":{"type":"array","description":"长生"},"shen_sha":{"type":"array","description":"神煞（天乙贵人/羊刃/桃花/驿马/华盖/劫煞/灾煞/红鸾/天喜/天罗地网/十恶大败等——按盘动态计算）"},"is_void":{"type":"boolean","description":"旬空"},"is_self_he":{"type":"boolean","description":"自合"},"is_kui_gang":{"type":"boolean","description":"魁罡"},"self_he_name":{"type":"string","description":"自合名"}},"required":["gan","zhi","na_yin"]},"yue":{"type":"object","description":"月柱: gan/zhi/na_yin","properties":{"gan":{"type":"string","enum":["甲","乙","丙","丁","戊","己","庚","辛","壬","癸"]},"zhi":{"type":"string","enum":["子","丑","寅","卯","辰","巳","午","未","申","酉","戌","亥"]},"na_yin":{"type":"string"},"cang_gan":{"type":"object","description":"藏干"},"shi_shens":{"type":"array","description":"十神"},"chang_sheng":{"type":"array","description":"长生"},"shen_sha":{"type":"array","description":"神煞（天乙贵人/羊刃/桃花/驿马/华盖/劫煞/灾煞/红鸾/天喜/天罗地网/十恶大败等——按盘动态计算）"},"is_void":{"type":"boolean","description":"旬空"},"is_self_he":{"type":"boolean","description":"自合"},"is_kui_gang":{"type":"boolean","description":"魁罡"},"self_he_name":{"type":"string","description":"自合名"}},"required":["gan","zhi","na_yin"]},"ri":{"type":"object","description":"日柱: gan/zhi/na_yin","properties":{"gan":{"type":"string","enum":["甲","乙","丙","丁","戊","己","庚","辛","壬","癸"]},"zhi":{"type":"string","enum":["子","丑","寅","卯","辰","巳","午","未","申","酉","戌","亥"]},"na_yin":{"type":"string"},"cang_gan":{"type":"object","description":"藏干"},"shi_shens":{"type":"array","description":"十神"},"chang_sheng":{"type":"array","description":"长生"},"shen_sha":{"type":"array","description":"神煞（天乙贵人/羊刃/桃花/驿马/华盖/劫煞/灾煞/红鸾/天喜/天罗地网/十恶大败等——按盘动态计算）"},"is_void":{"type":"boolean","description":"旬空"},"is_self_he":{"type":"boolean","description":"自合"},"is_kui_gang":{"type":"boolean","description":"魁罡"},"self_he_name":{"type":"string","description":"自合名"}},"required":["gan","zhi","na_yin"]},"shi":{"type":"object","description":"时柱: gan/zhi/na_yin","properties":{"gan":{"type":"string","enum":["甲","乙","丙","丁","戊","己","庚","辛","壬","癸"]},"zhi":{"type":"string","enum":["子","丑","寅","卯","辰","巳","午","未","申","酉","戌","亥"]},"na_yin":{"type":"string"},"cang_gan":{"type":"object","description":"藏干"},"shi_shens":{"type":"array","description":"十神"},"chang_sheng":{"type":"array","description":"长生"},"shen_sha":{"type":"array","description":"神煞（天乙贵人/羊刃/桃花/驿马/华盖/劫煞/灾煞/红鸾/天喜/天罗地网/十恶大败等——按盘动态计算）"},"is_void":{"type":"boolean","description":"旬空"},"is_self_he":{"type":"boolean","description":"自合"},"is_kui_gang":{"type":"boolean","description":"魁罡"},"self_he_name":{"type":"string","description":"自合名"}},"required":["gan","zhi","na_yin"]},"da_yun":{"type":"object","description":"大运。fullchart 会为每步补 rooted/root_refs；root_refs 说明坐支本气与原局藏干通根。","properties":{"steps":{"type":"array","items":{"type":"object","properties":{"gan":{"type":"string"},"zhi":{"type":"string"},"name":{"type":"string"},"wuxing":{"type":"string","enum":["木","火","土","金","水"]},"shi_shen":{"type":"string"},"rooted":{"type":"boolean"},"root_refs":{"type":"array","items":{"type":"object","properties":{"source":{"type":"string","enum":["sitting_branch_main_qi","natal_hidden_stem"]},"pillar":{"type":"string","enum":["nian","yue","ri","shi"]},"stem":{"type":"string"},"branch":{"type":"string"}},"required":["source","stem","branch"]}}},"required":["gan","zhi","name","wuxing","shi_shen","rooted"]}},"current_step_index":{"type":"integer"}},"required":["steps","current_step_index"]},"gender":{"type":"string"},"birth_year":{"type":"integer","description":"出生公历年份"},"yong_shen":{"type":"object","description":"用神三派（透传自 bazi.chart）","properties":{"fu_yi":{"type":"object","properties":{"wuxing_count":{"type":"object"},"wang_shuai":{"type":"object"},"yong":{"type":"string"},"xi":{"type":"string"},"ji":{"type":"string"},"qiangruo":{"type":"string","enum":["身强","身弱","中和"]}},"required":["wuxing_count","wang_shuai","yong","xi","ji","qiangruo"]},"tiao_hou":{"type":"object","properties":{"yong":{"type":"string"},"xi":{"type":"string"},"ji":{"type":"string"},"season":{"type":"string"},"detail":{"type":"string"}},"required":["yong","xi","ji","season"]},"ge_ju":{"type":"object","properties":{"yong":{"type":"string"},"xi":{"type":"string"},"ji":{"type":"string"},"ge_ju":{"type":"string"},"yong_fa":{"type":"string"}},"required":["yong","xi","ji","ge_ju","yong_fa"]}},"required":["fu_yi","tiao_hou","ge_ju"]},"san_yuan":{"type":"object","description":"三元: 胎元/命宫/身宫"},"gong_jia":{"type":"array","description":"拱夹"},"nayin_rel":{"type":"array","description":"纳音生克"},"chang_sheng":{"type":"array","description":"长生十二宫"},"san_qi_name":{"type":"string","description":"三奇贵人"},"gan_he":{"type":"array","description":"天干五合"},"zhi_liu_he":{"type":"array","description":"地支六合"},"san_he":{"type":"array","description":"三合局"},"san_hui":{"type":"array","description":"三会方"},"liu_chong":{"type":"array","description":"六冲"},"liu_hai":{"type":"array","description":"六害"},"liu_xing":{"type":"array","description":"相刑"},"lu_roots":{"type":"array","description":"透干十神得十干禄","items":{"type":"object","properties":{"shi_shen":{"type":"string"},"gan":{"type":"string"},"zhi":{"type":"string"}},"required":["shi_shen","gan","zhi"]}},"relation_groups":{"type":"array","description":"完整合会冲刑组；Python 只读","items":{"type":"object","additionalProperties":false,"properties":{"field":{"type":"string","enum":["gan_he","zhi_liu_he","san_he","san_hui","liu_chong","liu_hai","liu_xing"]},"group":{"type":"string"}},"required":["field","group"]}},"ten_god_states":{"type":"array","description":"engine 十神显隐 / 通根 / 得令 / 旺弱状态","items":{"type":"object","properties":{"shi_shen":{"type":"string","enum":["正官","七杀","正印","偏印","正财","偏财","比肩","劫财","食神","伤官"]},"wuxing":{"type":"string","enum":["木","火","土","金","水"]},"transparent":{"type":"boolean"},"hidden":{"type":"boolean"},"rooted":{"type":"boolean"},"timely":{"type":"boolean"},"count":{"type":"integer","minimum":0},"strength":{"type":"string","enum":["strong","weak","neutral"]}},"required":["shi_shen","wuxing","transparent","hidden","rooted","timely","count","strength"]}},"element_states":{"type":"array","description":"engine 五行季节旺弱、组合旺弱与生克方向","items":{"type":"object","properties":{"wuxing":{"type":"string","enum":["木","火","土","金","水"]},"season_strength":{"type":"string","enum":["strong","weak"]},"strength":{"type":"string","enum":["strong","weak","neutral"]},"transparent":{"type":"boolean"},"rooted":{"type":"boolean"},"controls":{"type":"string","enum":["木","火","土","金","水"]},"controlled_by":{"type":"string","enum":["木","火","土","金","水"]},"controller_strength":{"type":"string","enum":["strong","weak"]},"reasons":{"type":"array","items":{"type":"string"}}},"required":["wuxing","season_strength","strength","transparent","rooted","controls","controlled_by","controller_strength","reasons"]}},"atomic_facts":{"type":"object","description":"engine 确定性命理原子事实；Python 因子层只读","properties":{"day_master_element":{"type":"string","enum":["木","火","土","金","水"]},"day_master_stem":{"type":"string","enum":["甲","乙","丙","丁","戊","己","庚","辛","壬","癸"]},"day_branch":{"type":"string","enum":["子","丑","寅","卯","辰","巳","午","未","申","酉","戌","亥"]},"month_longevity":{"type":"string","enum":["长生","沐浴","冠带","临官","帝旺","衰","病","死","墓","绝","胎","养",""]},"year_stem_ten_god":{"type":"string","enum":["正官","七杀","正印","偏印","正财","偏财","比肩","劫财","食神","伤官",""]},"month_main_ten_god":{"type":"string","enum":["正官","七杀","正印","偏印","正财","偏财","比肩","劫财","食神","伤官",""]},"hour_stem_ten_god":{"type":"string","enum":["正官","七杀","正印","偏印","正财","偏财","比肩","劫财","食神","伤官",""]},"pattern_god_transparent":{"type":"boolean"},"pillar_punishments":{"type":"object","properties":{"nian":{"type":"boolean"},"yue":{"type":"boolean"},"ri":{"type":"boolean"},"shi":{"type":"boolean"}},"required":["nian","yue","ri","shi"]},"officer_killing_cleaned":{"type":"boolean"},"wealth_tomb_present":{"type":"boolean"},"wealth_star_in_tomb":{"type":"boolean"},"spouse_palace_state":{"type":"string","enum":["冲","合","刑","害","静"]},"day_branch_type":{"type":"string","enum":["桃花","驿马","墓库",""]},"year_officer_killing":{"type":"boolean"}},"required":["day_master_element","day_master_stem","day_branch","month_longevity","year_stem_ten_god","month_main_ten_god","hour_stem_ten_god","pattern_god_transparent","pillar_punishments","officer_killing_cleaned","wealth_tomb_present","wealth_star_in_tomb","spouse_palace_state","day_branch_type","year_officer_killing"]},"xun_kong":{"type":"string","description":"旬空（按日柱所居之旬），如甲申旬空午未"}},"required":["nian","yue","ri","shi","da_yun","gender","lu_roots","relation_groups","ten_god_states","element_states","atomic_facts"]}`),
	},
	{
		Name: "bazi.chart", Description: "排八字命盘。返回最小命盘（四柱+纳音+大运+性别）。如需十神/藏干/神煞/长生/空亡/用神三派等完整信息，请将结果传入 bazi.fullchart（用神属完整命盘）。",
		Params: mustSchema(`{"type":"object","properties":{"solar_time":` + schemaSolarTime + `,"gender":` + schemaGender + `,"longitude":{"type":"number","description":"出生地经度（度），用于真太阳时校正。缺省用东八区 120。所有命例都应传实际经度，海外/西部命例必传否则时辰错"}},"required":["solar_time","gender"]}`), Handler: baziChartHandler,
		Result: envelopeSchema(`{"type":"object","properties":{"nian":{"type":"object","description":"年柱: gan/zhi/na_yin","properties":{"gan":{"type":"string","enum":["甲","乙","丙","丁","戊","己","庚","辛","壬","癸"]},"zhi":{"type":"string","enum":["子","丑","寅","卯","辰","巳","午","未","申","酉","戌","亥"]},"na_yin":{"type":"string"}},"required":["gan","zhi","na_yin"]},"yue":{"type":"object","description":"月柱: gan/zhi/na_yin","properties":{"gan":{"type":"string","enum":["甲","乙","丙","丁","戊","己","庚","辛","壬","癸"]},"zhi":{"type":"string","enum":["子","丑","寅","卯","辰","巳","午","未","申","酉","戌","亥"]},"na_yin":{"type":"string"}},"required":["gan","zhi","na_yin"]},"ri":{"type":"object","description":"日柱: gan/zhi/na_yin","properties":{"gan":{"type":"string","enum":["甲","乙","丙","丁","戊","己","庚","辛","壬","癸"]},"zhi":{"type":"string","enum":["子","丑","寅","卯","辰","巳","午","未","申","酉","戌","亥"]},"na_yin":{"type":"string"}},"required":["gan","zhi","na_yin"]},"shi":{"type":"object","description":"时柱: gan/zhi/na_yin","properties":{"gan":{"type":"string","enum":["甲","乙","丙","丁","戊","己","庚","辛","壬","癸"]},"zhi":{"type":"string","enum":["子","丑","寅","卯","辰","巳","午","未","申","酉","戌","亥"]},"na_yin":{"type":"string"}},"required":["gan","zhi","na_yin"]},"da_yun":{"type":"object","description":"大运。start_year_after/start_month_after/start_day_after 为出生后起运精确时间（3天=1年）；direction 顺排/逆排；steps 每步含 gan/zhi/name/wuxing/shi_shen，与 start_date/end_date（公历日期段）/start_year/end_year（公历年）；current_step_index 为当前所处大运索引（-1 表示未起运或已过完所有大运）；next_step_in_years 距下一大运剩余年数（公历口径）；steps 每步含 start_date/end_date（公历日期段）与 start_year/end_year（公历年，免虚岁换算）"},"gender":{"type":"string"},"birth_year":{"type":"integer","description":"出生公历年份（供 bazi.liunian/liuri 定位当年大运）"},"zi_shi_rule":{"type":"string","description":"子时换日规则说明（晚子时日柱不变、时柱按次日日干起）"}},"required":["nian","yue","ri","shi","da_yun","gender"]}`),
	},
	{
		Name: "bazi.bond", Description: "八字合盘。返回双方日主、天干关系（合/生/克）、地支关系（六合/三合/六冲）、纳音配合、五行互补。",
		Params:  mustSchema(`{"type":"object","properties":{"a":{"type":"object","properties":{"chart":{"type":"object","description":"八字命盘（由 bazi.chart 返回的最小命盘，不需 bazi.fullchart）"}},"required":["chart"]},"b":{"type":"object","properties":{"chart":{"type":"object","description":"八字命盘（由 bazi.chart 返回的最小命盘，不需 bazi.fullchart）"}},"required":["chart"]}},"required":["a","b"]}`),
		Handler: baziBondHandler,
		Result:  envelopeSchema(`{"type":"object","properties":{"zhu_cross":{"type":"object","description":"16对四柱互配结果，含jia_zhu/yi_zhu/jia_gan/yi_gan/gan_guan_xi/zhi_guan_xi等"},"shi_shen_cross":{"type":"object","description":"双方日主看对方四柱的十神映射: jia_dui_yi/yi_dui_jia"},"structure":{"type":"object","description":"大运交叉+旬宫判断，含da_yun/xun_gong"},"nayin_cross":{"type":"object","description":"纳音配合"},"shensha_cross":{"type":"object","description":"神煞交叉"}},"required":["zhu_cross","shi_shen_cross","structure"]}`),
	},
	{
		Name: "bazi.liunian", Description: "八字流年运势。返回流年干支与命局的十神、神煞、伏吟反吟。",
		Params:  mustSchema(`{"type":"object","properties":{"year":{"type":"integer","minimum":1,"maximum":8000,"description":"目标年份（VSOP87D 天文精度范围 1-8000）"},"chart":{"type":"object","description":"八字命盘（由 bazi.chart 返回的最小命盘，不需 bazi.fullchart）"}},"required":["year","chart"]}`),
		Handler: baziLiunianHandler,
		Result:  envelopeSchema(`{"type":"object","properties":{"year":{"type":"integer"},"nian_name":{"type":"string"},"shi_shen":{"type":"string","enum":["正官","七杀","正印","偏印","正财","偏财","比肩","劫财","食神","伤官"]},"nian_gan":{"type":"string","description":"流年天干"},"nian_zhi":{"type":"string","description":"流年地支"},"wuxing":{"type":"string"},"na_yin":{"type":"string"},"sheng":{"type":"integer","description":"生数"},"ke":{"type":"integer","description":"克数"},"shensha":{"type":"array","description":"流年神煞（动态：桃花/驿马/华盖/劫煞/灾煞/红鸾/天喜/天乙贵人/羊刃——年支+日支双查；值年：病符/丧门/吊客/大耗——命局四柱逢煞支即应）","items":{"type":"object","properties":{"name":{"type":"string","enum":["桃花","驿马","华盖","劫煞","灾煞","红鸾","天喜","天乙贵人","羊刃","病符","丧门","吊客","大耗"]},"category":{"type":"string","enum":["吉","凶","中性"]},"description":{"type":"string"}}}},"fuyin_fanyin":{"type":"array","description":"伏吟反吟"},"dayun_interactions":{"type":"array","description":"查询年份所在大运与流年交互（chart 需含 birth_year；未起运/大运过完/缺 birth_year 时为空数组）"},"natal_interactions":{"type":"array","description":"流年与命局交互"},"atomic_facts":{"type":"object","description":"engine 流年原子事实","properties":{"controls_targets":{"type":"object","properties":{"day_master":{"type":"boolean"},"wealth_star":{"type":"boolean"},"officer_killing":{"type":"boolean"},"seal_star":{"type":"boolean"},"food_injury":{"type":"boolean"}},"additionalProperties":false,"required":["day_master","wealth_star","officer_killing","seal_star","food_injury"]},"controls_elements":{"type":"array","items":{"type":"string","enum":["木","火","土","金","水"]}},"unfavorable_gan":{"type":"boolean"},"unfavorable_branch":{"type":"boolean"},"wealth_breaks_seal":{"type":"boolean"},"year_branch_relations":{"type":"object","additionalProperties":{"type":"string","enum":["liu_he","san_he","san_hui","liu_chong","xing","liu_hai"]}},"year_branch_controlled_by":{"type":"array","items":{"type":"string","enum":["木","火","土","金","水"]}},"year_gan_controls_day_gan":{"type":"boolean"},"day_gan_controls_year_gan":{"type":"boolean"},"gan_combines":{"type":"boolean"},"dayun_gan_combines":{"type":"boolean"},"dayun_zhi_clashes_year":{"type":"boolean"},"day_void_branches":{"type":"array","items":{"type":"string"}},"year_equals_day_pillar":{"type":"boolean"},"dayun_equals_year_pillar":{"type":"boolean"},"year_gan_equals_natal_year_gan":{"type":"boolean"},"combinations":{"type":"array","items":{"type":"object","properties":{"kind":{"type":"string","enum":["san_he","san_hui","xing","half_he"]},"group":{"type":"string"},"branches":{"type":"array","items":{"type":"string"}},"includes_year":{"type":"boolean"}},"required":["kind","group","branches","includes_year"]}}},"required":["controls_elements","controls_targets","unfavorable_gan","unfavorable_branch","wealth_breaks_seal","year_branch_relations","year_branch_controlled_by","year_gan_controls_day_gan","day_gan_controls_year_gan","gan_combines","dayun_gan_combines","dayun_zhi_clashes_year","day_void_branches","year_equals_day_pillar","dayun_equals_year_pillar","year_gan_equals_natal_year_gan","combinations"]}},"required":["year","nian_name","shi_shen","atomic_facts"]}`),
	},
	{
		Name: "bazi.liuyue", Description: "流月运势。返回流月干支与命局的十神、神煞。",
		Params:  mustSchema(`{"type":"object","properties":{"year":{"type":"integer","description":"目标年份"},"month":{"type":"integer","minimum":1,"maximum":12,"description":"目标月份"},"chart":{"type":"object","description":"八字命盘（由 bazi.chart 返回的最小命盘，不需 bazi.fullchart）"}},"required":["year","month","chart"]}`),
		Handler: baziLiuyueHandler,
		Result:  envelopeSchema(`{"type":"object","properties":{"year":{"type":"integer"},"month":{"type":"integer"},"month_name":{"type":"string"},"shi_shen":{"type":"string","enum":["正官","七杀","正印","偏印","正财","偏财","比肩","劫财","食神","伤官"]},"yue_gan":{"type":"string","description":"流月天干"},"yue_zhi":{"type":"string","description":"流月地支"},"wuxing":{"type":"string"},"sheng":{"type":"integer"},"ke":{"type":"integer"},"shensha":{"type":"array","description":"流月神煞（动态：桃花/驿马/华盖/劫煞/灾煞/红鸾/天喜/天乙贵人/羊刃——年支+日支双查）","items":{"type":"object","properties":{"name":{"type":"string","enum":["桃花","驿马","华盖","劫煞","灾煞","红鸾","天喜","天乙贵人","羊刃"]},"category":{"type":"string","enum":["吉","凶","中性"]},"description":{"type":"string"}}}},"gan_rels":{"type":"array","description":"天干关系"},"zhi_rels":{"type":"array","description":"地支关系"}},"required":["year","month","month_name","shi_shen"]}`),
	},
	{
		Name: "bazi.liuri", Description: "流日运势。返回流日干支、十神、纳音。",
		Params:  mustSchema(`{"type":"object","properties":{"year":{"type":"integer","description":"目标年份"},"month":{"type":"integer","minimum":1,"maximum":12,"description":"目标月份"},"day":{"type":"integer","minimum":1,"maximum":31,"description":"目标日期"},"chart":{"type":"object","description":"八字命盘（由 bazi.chart 返回的最小命盘，不需 bazi.fullchart）"}},"required":["year","month","day","chart"]}`),
		Handler: baziLiuriHandler,
		Result:  envelopeSchema(`{"type":"object","properties":{"date":{"type":"string"},"day_name":{"type":"string"},"shi_shen":{"type":"string","enum":["正官","七杀","正印","偏印","正财","偏财","比肩","劫财","食神","伤官"]},"ri_gan":{"type":"string","description":"流日天干"},"ri_zhi":{"type":"string","description":"流日地支"},"day_nayin":{"type":"string"},"shensha":{"type":"array","description":"流日神煞（动态：桃花/驿马/华盖/劫煞/灾煞/红鸾/天喜/天乙贵人/羊刃）","items":{"type":"object","properties":{"name":{"type":"string","enum":["桃花","驿马","华盖","劫煞","灾煞","红鸾","天喜","天乙贵人","羊刃"]},"category":{"type":"string","enum":["吉","凶","中性"]},"description":{"type":"string"}}}},"gan_rels":{"type":"array","description":"天干关系"},"zhi_rels":{"type":"array","description":"地支关系"},"dayun_rels":{"type":"array","description":"与大运关系"},"liunian_rels":{"type":"array","description":"与流年关系"}},"required":["date","day_name","shi_shen"]}`),
	},
	{
		Name: "bazi.liushi", Description: "流时运势。返回流时干支、十神。hour 为时辰（0-23）。",
		Params:  mustSchema(`{"type":"object","properties":{"year":{"type":"integer","description":"目标年份"},"month":{"type":"integer","minimum":1,"maximum":12,"description":"目标月份"},"day":{"type":"integer","minimum":1,"maximum":31,"description":"目标日期"},"hour":{"type":"integer","minimum":0,"maximum":23,"description":"时辰"},"chart":{"type":"object","description":"八字命盘（由 bazi.chart 返回的最小命盘，不需 bazi.fullchart）"}},"required":["year","month","day","hour","chart"]}`),
		Handler: baziLiushiHandler,
		Result:  envelopeSchema(`{"type":"object","properties":{"time":{"type":"string"},"hour_name":{"type":"string"},"shi_shen":{"type":"string","enum":["正官","七杀","正印","偏印","正财","偏财","比肩","劫财","食神","伤官"]},"shi_gan":{"type":"string","description":"流时天干"},"shi_zhi":{"type":"string","description":"流时地支"},"gan_rels":{"type":"array","description":"天干关系"},"zhi_rels":{"type":"array","description":"地支关系"}},"required":["time","hour_name","shi_shen"]}`),
	},
	{
		Name: "bazi.xiaoyun", Description: "小运。返回小运流年列表。count 默认 12。",
		Params:  mustSchema(`{"type":"object","properties":{"chart":{"type":"object","description":"八字命盘（由 bazi.chart 返回的最小命盘，不需 bazi.fullchart）"},"count":{"type":"integer","description":"返回年数，默认 12"}},"required":["chart"]}`),
		Handler: baziXiaoYunHandler,
		Result:  envelopeSchema(`{"type":"array","items":{"type":"object","properties":{"age":{"type":"integer"},"gan":{"type":"string","enum":["甲","乙","丙","丁","戊","己","庚","辛","壬","癸"]},"zhi":{"type":"string","enum":["子","丑","寅","卯","辰","巳","午","未","申","酉","戌","亥"]},"name":{"type":"string"}},"required":["age","gan","zhi","name"]}}`),
	},
}
