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
		name  string
		gan   ganzhi.Gan
		zhi   ganzhi.Zhi
		naYin string
	}{
		{"nian", chart.Nian.Gan, chart.Nian.Zhi, chart.Nian.NaYin},
		{"yue", chart.Yue.Gan, chart.Yue.Zhi, chart.Yue.NaYin},
		{"ri", chart.Ri.Gan, chart.Ri.Zhi, chart.Ri.NaYin},
		{"shi", chart.Shi.Gan, chart.Shi.Zhi, chart.Shi.NaYin},
	}
	for _, pillar := range pillars {
		if pillar.gan == 0 || pillar.zhi == 0 {
			return fmt.Errorf("bazi.fullchart: chart.%s requires non-empty gan and zhi", pillar.name)
		}
		if int(pillar.gan)%2 != int(pillar.zhi)%2 {
			return fmt.Errorf(
				"bazi.fullchart: chart.%s has invalid sexagenary pair %s%s",
				pillar.name, ganzhi.GanName(pillar.gan), ganzhi.ZhiName(pillar.zhi),
			)
		}
		if pillar.naYin == "" {
			return fmt.Errorf("bazi.fullchart: chart.%s.na_yin is required", pillar.name)
		}
		if expected := ganzhi.NayinLabel(pillar.gan, pillar.zhi); pillar.naYin != expected {
			return fmt.Errorf(
				"bazi.fullchart: chart.%s.na_yin is %q, want %q for %s%s",
				pillar.name, pillar.naYin, expected,
				ganzhi.GanName(pillar.gan), ganzhi.ZhiName(pillar.zhi),
			)
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
		Name: "bazi.fullchart", Description: "完整命盘。传入 bazi.chart 返回的最小命盘，补全十神/藏干/神煞/自坐/空亡/自合/魁罡/三元/拱夹/纳音生克/长生十二宫/三奇贵人/合会冲刑/十干禄/命理原子事实。",
		Params:  mustSchema(`{"type":"object","properties":{"chart":{"type":"object","description":"bazi.chart 返回的最小命盘"}},"required":["chart"]}`),
		Handler: baziFullChartHandler,
		Result:  envelopeSchema(`{"type":"object","properties":{"nian":{"type":"object","description":"年柱: gan/zhi/na_yin","properties":{"gan":{"type":"string","enum":["甲","乙","丙","丁","戊","己","庚","辛","壬","癸"]},"zhi":{"type":"string","enum":["子","丑","寅","卯","辰","巳","午","未","申","酉","戌","亥"]},"na_yin":{"type":"string"},"cang_gan":{"type":"object","description":"藏干"},"shi_shens":{"type":"array","description":"十神"},"day_master_trend":{"type":"string","description":"日主在此柱地支的十二长生（地势）"},"xun":{"type":"string","description":"该柱所在旬"},"xun_kong":{"type":"string","description":"该柱所在旬的旬空二支；is_void 另指该柱值日柱旬空"},"self_sitting":{"type":"string","description":"自坐长生（该柱天干在自己的地支上的十二长生阶段）"},"shen_sha":{"type":"array","description":"神煞（天乙贵人/羊刃/桃花/驿马/华盖/劫煞/灾煞/红鸾/天喜/天罗地网/十恶大败等——按盘动态计算）"},"is_void":{"type":"boolean","description":"旬空"},"is_self_he":{"type":"boolean","description":"自合"},"is_kui_gang":{"type":"boolean","description":"魁罡"},"self_he_name":{"type":"string","description":"自合名"}},"required":["gan","zhi","day_master_trend","xun","xun_kong","na_yin","cang_gan","shi_shens","self_sitting","shen_sha","is_void","is_self_he","is_kui_gang","self_he_name"]},"yue":{"type":"object","description":"月柱: gan/zhi/na_yin","properties":{"gan":{"type":"string","enum":["甲","乙","丙","丁","戊","己","庚","辛","壬","癸"]},"zhi":{"type":"string","enum":["子","丑","寅","卯","辰","巳","午","未","申","酉","戌","亥"]},"na_yin":{"type":"string"},"cang_gan":{"type":"object","description":"藏干"},"shi_shens":{"type":"array","description":"十神"},"day_master_trend":{"type":"string","description":"日主在此柱地支的十二长生（地势）"},"xun":{"type":"string","description":"该柱所在旬"},"xun_kong":{"type":"string","description":"该柱所在旬的旬空二支；is_void 另指该柱值日柱旬空"},"self_sitting":{"type":"string","description":"自坐长生（该柱天干在自己的地支上的十二长生阶段）"},"shen_sha":{"type":"array","description":"神煞（天乙贵人/羊刃/桃花/驿马/华盖/劫煞/灾煞/红鸾/天喜/天罗地网/十恶大败等——按盘动态计算）"},"is_void":{"type":"boolean","description":"旬空"},"is_self_he":{"type":"boolean","description":"自合"},"is_kui_gang":{"type":"boolean","description":"魁罡"},"self_he_name":{"type":"string","description":"自合名"}},"required":["gan","zhi","day_master_trend","xun","xun_kong","na_yin","cang_gan","shi_shens","self_sitting","shen_sha","is_void","is_self_he","is_kui_gang","self_he_name"]},"ri":{"type":"object","description":"日柱: gan/zhi/na_yin","properties":{"gan":{"type":"string","enum":["甲","乙","丙","丁","戊","己","庚","辛","壬","癸"]},"zhi":{"type":"string","enum":["子","丑","寅","卯","辰","巳","午","未","申","酉","戌","亥"]},"na_yin":{"type":"string"},"cang_gan":{"type":"object","description":"藏干"},"shi_shens":{"type":"array","description":"十神"},"day_master_trend":{"type":"string","description":"日主在此柱地支的十二长生（地势）"},"xun":{"type":"string","description":"该柱所在旬"},"xun_kong":{"type":"string","description":"该柱所在旬的旬空二支；is_void 另指该柱值日柱旬空"},"self_sitting":{"type":"string","description":"自坐长生（该柱天干在自己的地支上的十二长生阶段）"},"shen_sha":{"type":"array","description":"神煞（天乙贵人/羊刃/桃花/驿马/华盖/劫煞/灾煞/红鸾/天喜/天罗地网/十恶大败等——按盘动态计算）"},"is_void":{"type":"boolean","description":"旬空"},"is_self_he":{"type":"boolean","description":"自合"},"is_kui_gang":{"type":"boolean","description":"魁罡"},"self_he_name":{"type":"string","description":"自合名"}},"required":["gan","zhi","day_master_trend","xun","xun_kong","na_yin","cang_gan","shi_shens","self_sitting","shen_sha","is_void","is_self_he","is_kui_gang","self_he_name"]},"shi":{"type":"object","description":"时柱: gan/zhi/na_yin","properties":{"gan":{"type":"string","enum":["甲","乙","丙","丁","戊","己","庚","辛","壬","癸"]},"zhi":{"type":"string","enum":["子","丑","寅","卯","辰","巳","午","未","申","酉","戌","亥"]},"na_yin":{"type":"string"},"cang_gan":{"type":"object","description":"藏干"},"shi_shens":{"type":"array","description":"十神"},"day_master_trend":{"type":"string","description":"日主在此柱地支的十二长生（地势）"},"xun":{"type":"string","description":"该柱所在旬"},"xun_kong":{"type":"string","description":"该柱所在旬的旬空二支；is_void 另指该柱值日柱旬空"},"self_sitting":{"type":"string","description":"自坐长生（该柱天干在自己的地支上的十二长生阶段）"},"shen_sha":{"type":"array","description":"神煞（天乙贵人/羊刃/桃花/驿马/华盖/劫煞/灾煞/红鸾/天喜/天罗地网/十恶大败等——按盘动态计算）"},"is_void":{"type":"boolean","description":"旬空"},"is_self_he":{"type":"boolean","description":"自合"},"is_kui_gang":{"type":"boolean","description":"魁罡"},"self_he_name":{"type":"string","description":"自合名"}},"required":["gan","zhi","day_master_trend","xun","xun_kong","na_yin","cang_gan","shi_shens","self_sitting","shen_sha","is_void","is_self_he","is_kui_gang","self_he_name"]},"da_yun":{"type":"object","description":"大运。fullchart 会为每步补 rooted/root_refs；root_refs 说明坐支本气与原局藏干通根。","properties":{"steps":{"type":"array","items":{"type":"object","properties":{"gan":{"type":"string"},"zhi":{"type":"string"},"name":{"type":"string"},"wuxing":{"type":"string","enum":["木","火","土","金","水"]},"shi_shen":{"type":"string"},"rooted":{"type":"boolean"},"root_refs":{"type":"array","items":{"type":"object","properties":{"source":{"type":"string","enum":["sitting_branch_main_qi","natal_hidden_stem"]},"pillar":{"type":"string","enum":["nian","yue","ri","shi"]},"stem":{"type":"string"},"branch":{"type":"string"}},"required":["source","stem","branch"]}}},"required":["gan","zhi","name","wuxing","shi_shen","rooted"]}},"current_step_index":{"type":"integer"}},"required":["steps","current_step_index"]},"gender":{"type":"string"},"birth_year":{"type":"integer","description":"出生公历年份"},"fu_yi":{"type":"object","description":"用神三派（扶抑/调候/格局，由 bazi.fullchart 计算）","properties":{"fu_yi":{"type":"object","description":"扶抑法候选；basis 输出强弱表输入与日主相关关系投影，动态再权衡仍须整体合参","properties":{"wuxing_count":{"type":"object"},"wang_shuai":{"type":"object"},"yong":{"type":"string"},"xi":{"type":"string"},"ji":{"type":"string"},"qiangruo":{"type":"string","enum":["身强","身弱","中和"]},"model":{"type":"string","enum":["support_control_with_day_master_strength"]},"basis":{"type":"object","additionalProperties":false,"properties":{"root_type":{"type":"string","enum":["month_main","month_mid","zhi_main","zhi_mid","none"]},"season":{"type":"string","enum":["旺","相","休","囚","死"]},"yin_bi_count":{"type":"integer","minimum":0,"maximum":3},"day_master_roots":{"type":"array","items":{"type":"object","additionalProperties":false,"properties":{"pillar":{"type":"string","enum":["nian","yue","ri","shi"]},"branch":{"type":"string","enum":["子","丑","寅","卯","辰","巳","午","未","申","酉","戌","亥"]},"stem":{"type":"string","enum":["甲","乙","丙","丁","戊","己","庚","辛","壬","癸"]},"source":{"type":"string","enum":["gan","main_qi","mid_qi","minor_qi"]}},"required":["pillar","branch","stem","source"]}},"relation_facts":{"type":"array","items":{"type":"object","additionalProperties":false,"properties":{"field":{"type":"string","enum":["gan_he","gan_chong","gan_chong_candidate","zhi_liu_he","san_he","san_he_partial","san_hui","liu_chong","liu_hai","liu_xing","liu_po","an_he"]},"group":{"type":"string"},"pillars":{"type":"array","items":{"type":"string","enum":["nian","yue","ri","shi"]}},"branches":{"type":"array","items":{"type":"string","enum":["子","丑","寅","卯","辰","巳","午","未","申","酉","戌","亥"]}},"targets":{"type":"array","items":{"type":"string","enum":["day_master","yin_bi","day_branch","day_master_root"]}}},"required":["field","group","pillars","branches","targets"]}}},"required":["root_type","season","yin_bi_count","day_master_roots","relation_facts"]}},"required":["wuxing_count","wang_shuai","yong","xi","ji","qiangruo","model","basis"]},"tiao_hou":{"type":"object","additionalProperties":false,"description":"穷通宝鉴调候表输出主辅用神及其原局显隐事实；忌神不在本表机械反推","properties":{"primary_wuxing":{"type":"string","description":"穷通宝鉴主用五行"},"secondary_wuxing":{"type":"string","description":"穷通宝鉴辅用五行候选；secondary_condition 非空时为条件候选"},"secondary_condition":{"type":"string","description":"辅用神的经典条件；非空时不得当作无条件喜神"},"season":{"type":"string"},"detail":{"type":"string"},"model":{"type":"string","enum":["qiongtong_primary_secondary_table"]},"primary":{"type":"object","additionalProperties":false,"properties":{"stem":{"type":"string","enum":["甲","乙","丙","丁","戊","己","庚","辛","壬","癸"]},"transparent":{"type":"boolean"},"hidden":{"type":"boolean"},"occurrences":{"type":"array","items":{"type":"object","additionalProperties":false,"properties":{"pillar":{"type":"string","enum":["nian","yue","ri","shi"]},"branch":{"type":"string","enum":["子","丑","寅","卯","辰","巳","午","未","申","酉","戌","亥"]},"stem":{"type":"string","enum":["甲","乙","丙","丁","戊","己","庚","辛","壬","癸"]},"source":{"type":"string","enum":["gan","main_qi","mid_qi","minor_qi"]}},"required":["pillar","branch","stem","source"]}},"relation_facts":{"type":"array","items":{"type":"object","additionalProperties":false,"properties":{"field":{"type":"string","enum":["gan_he","gan_chong","gan_chong_candidate","zhi_liu_he","san_he","san_he_partial","san_hui","liu_chong","liu_hai","liu_xing","liu_po","an_he"]},"group":{"type":"string"},"pillars":{"type":"array","items":{"type":"string","enum":["nian","yue","ri","shi"]}},"branches":{"type":"array","items":{"type":"string","enum":["子","丑","寅","卯","辰","巳","午","未","申","酉","戌","亥"]}},"targets":{"type":"array","items":{"type":"string","enum":["primary","secondary"]}}},"required":["field","group","pillars","branches","targets"]}}},"required":["stem","transparent","hidden","occurrences","relation_facts"]},"secondary":{"type":"object","additionalProperties":false,"properties":{"stem":{"type":"string","enum":["甲","乙","丙","丁","戊","己","庚","辛","壬","癸"]},"transparent":{"type":"boolean"},"hidden":{"type":"boolean"},"occurrences":{"type":"array","items":{"type":"object","additionalProperties":false,"properties":{"pillar":{"type":"string","enum":["nian","yue","ri","shi"]},"branch":{"type":"string","enum":["子","丑","寅","卯","辰","巳","午","未","申","酉","戌","亥"]},"stem":{"type":"string","enum":["甲","乙","丙","丁","戊","己","庚","辛","壬","癸"]},"source":{"type":"string","enum":["gan","main_qi","mid_qi","minor_qi"]}},"required":["pillar","branch","stem","source"]}},"relation_facts":{"type":"array","items":{"type":"object","additionalProperties":false,"properties":{"field":{"type":"string","enum":["gan_he","gan_chong","gan_chong_candidate","zhi_liu_he","san_he","san_he_partial","san_hui","liu_chong","liu_hai","liu_xing","liu_po","an_he"]},"group":{"type":"string"},"pillars":{"type":"array","items":{"type":"string","enum":["nian","yue","ri","shi"]}},"branches":{"type":"array","items":{"type":"string","enum":["子","丑","寅","卯","辰","巳","午","未","申","酉","戌","亥"]}},"targets":{"type":"array","items":{"type":"string","enum":["primary","secondary"]}}},"required":["field","group","pillars","branches","targets"]}}},"required":["stem","transparent","hidden","occurrences","relation_facts"]}},"required":["primary_wuxing","season","model","primary"]},"ge_ju":{"type":"object","additionalProperties":false,"description":"月令透干格局候选；structure 输出格神、克格神与生格神原子事实，不输出成败格终判","properties":{"ge_ju":{"type":"string"},"yong_fa":{"type":"string"},"pattern_god":{"type":"string"},"pattern_god_ten_god":{"type":"string"},"pattern_god_source":{"type":"string","enum":["main_qi","mid_qi","minor_qi","main_qi_not_transparent","month_lu","month_blade"]},"structure":{"type":"object","additionalProperties":false,"properties":{"pattern":{"type":"object","additionalProperties":false,"properties":{"wuxing":{"type":"string","enum":["木","火","土","金","水"]},"transparent_pillars":{"type":"array","items":{"type":"string","enum":["nian","yue","ri","shi"]}},"roots":{"type":"array","items":{"type":"object","additionalProperties":false,"properties":{"pillar":{"type":"string","enum":["nian","yue","ri","shi"]},"branch":{"type":"string","enum":["子","丑","寅","卯","辰","巳","午","未","申","酉","戌","亥"]},"stem":{"type":"string","enum":["甲","乙","丙","丁","戊","己","庚","辛","壬","癸"]},"source":{"type":"string","enum":["gan","main_qi","mid_qi","minor_qi"]}},"required":["pillar","branch","stem","source"]}},"season":{"type":"string","enum":["旺","相","休","囚","死"]},"timely":{"type":"boolean"},"strength":{"type":"string","enum":["strong","weak","neutral"]}},"required":["wuxing","transparent_pillars","roots","season","timely","strength"]},"controller":{"type":"object","additionalProperties":false,"properties":{"wuxing":{"type":"string","enum":["木","火","土","金","水"]},"transparent_pillars":{"type":"array","items":{"type":"string","enum":["nian","yue","ri","shi"]}},"roots":{"type":"array","items":{"type":"object","additionalProperties":false,"properties":{"pillar":{"type":"string","enum":["nian","yue","ri","shi"]},"branch":{"type":"string","enum":["子","丑","寅","卯","辰","巳","午","未","申","酉","戌","亥"]},"stem":{"type":"string","enum":["甲","乙","丙","丁","戊","己","庚","辛","壬","癸"]},"source":{"type":"string","enum":["gan","main_qi","mid_qi","minor_qi"]}},"required":["pillar","branch","stem","source"]}},"season":{"type":"string","enum":["旺","相","休","囚","死"]},"timely":{"type":"boolean"},"strength":{"type":"string","enum":["strong","weak","neutral"]}},"required":["wuxing","transparent_pillars","roots","season","timely","strength"]},"generator":{"type":"object","additionalProperties":false,"properties":{"wuxing":{"type":"string","enum":["木","火","土","金","水"]},"transparent_pillars":{"type":"array","items":{"type":"string","enum":["nian","yue","ri","shi"]}},"roots":{"type":"array","items":{"type":"object","additionalProperties":false,"properties":{"pillar":{"type":"string","enum":["nian","yue","ri","shi"]},"branch":{"type":"string","enum":["子","丑","寅","卯","辰","巳","午","未","申","酉","戌","亥"]},"stem":{"type":"string","enum":["甲","乙","丙","丁","戊","己","庚","辛","壬","癸"]},"source":{"type":"string","enum":["gan","main_qi","mid_qi","minor_qi"]}},"required":["pillar","branch","stem","source"]}},"season":{"type":"string","enum":["旺","相","休","囚","死"]},"timely":{"type":"boolean"},"strength":{"type":"string","enum":["strong","weak","neutral"]}},"required":["wuxing","transparent_pillars","roots","season","timely","strength"]},"relation_facts":{"type":"array","items":{"type":"object","additionalProperties":false,"properties":{"field":{"type":"string","enum":["gan_he","gan_chong","gan_chong_candidate","zhi_liu_he","san_he","san_he_partial","san_hui","liu_chong","liu_hai","liu_xing","liu_po","an_he"]},"group":{"type":"string"},"pillars":{"type":"array","items":{"type":"string","enum":["nian","yue","ri","shi"]}},"branches":{"type":"array","items":{"type":"string","enum":["子","丑","寅","卯","辰","巳","午","未","申","酉","戌","亥"]}},"targets":{"type":"array","items":{"type":"string","enum":["pattern","controller","generator"]}}},"required":["field","group","pillars","branches","targets"]}}},"required":["pattern","controller","generator","relation_facts"]}},"required":["ge_ju","yong_fa","pattern_god_source","structure"]}},"required":["fu_yi","tiao_hou","ge_ju"]},"san_yuan":{"type":"object","description":"三元: 胎元/命宫/身宫","properties":{"tai_yuan":{"type":"object","description":"胎元（月干+1 月支+3）","properties":{"gan":{"type":"string"},"zhi":{"type":"string"}},"required":["gan","zhi"]},"ming_gong":{"type":"object","description":"命宫","properties":{"gan":{"type":"string"},"zhi":{"type":"string"}},"required":["gan","zhi"]},"shen_gong":{"type":"object","description":"身宫","properties":{"gan":{"type":"string"},"zhi":{"type":"string"}},"required":["gan","zhi"]}},"required":["tai_yuan","ming_gong","shen_gong"]},"tai_xi":{"type":"object","description":"胎息（日干五合配、日支六合配）","properties":{"gan":{"type":"string"},"zhi":{"type":"string"}},"required":["gan","zhi"]},"gong_jia":{"type":"array","description":"经典拱局：三合拱 / 三会拱","items":{"type":"object","additionalProperties":false,"properties":{"pillar_a":{"type":"integer","minimum":0,"maximum":3},"pillar_b":{"type":"integer","minimum":0,"maximum":3},"type":{"type":"string","enum":["三合拱","三会拱"]},"wuxing":{"type":"string","enum":["木","火","土","金","水"]},"zhi":{"type":"string","enum":["子","丑","寅","卯","辰","巳","午","未","申","酉","戌","亥"]}},"required":["pillar_a","pillar_b","type","wuxing","zhi"]}},"nayin_rel":{"type":"array","description":"纳音生克"},"chang_sheng":{"type":"array","description":"日主十二宫全表（长生/沐浴/冠带/临官/帝旺/衰/病/死/墓/绝/胎/养在十二支的分布）"},"zodiac":{"type":"string","description":"生肖（年支对应动物，如酉→鸡）"},"san_qi_name":{"type":"string","description":"三奇贵人"},"gan_he":{"type":"array","description":"天干五合"},"gan_chong":{"type":"array","description":"天干相冲（甲庚/乙辛/丙壬/丁癸）","items":{"type":"object","additionalProperties":false,"properties":{"gan_a":{"type":"string","enum":["甲","乙","丙","丁","戊","己","庚","辛","壬","癸"]},"gan_b":{"type":"string","enum":["甲","乙","丙","丁","戊","己","庚","辛","壬","癸"]},"pillar_a":{"type":"integer","minimum":0,"maximum":3},"pillar_b":{"type":"integer","minimum":0,"maximum":3},"position":{"type":"string","enum":["adjacent","separated","remote"]}},"required":["gan_a","gan_b","pillar_a","pillar_b","position"]}},"zhi_liu_he":{"type":"array","description":"地支六合"},"san_he_partial":{"type":"array","description":"半合局（三合缺一支且必须含旺支）","items":{"type":"object","additionalProperties":false,"properties":{"type":{"type":"string","enum":["半合"]},"name":{"type":"string"},"wuxing":{"type":"string","enum":["木","火","土","金","水"]},"branches":{"type":"array","items":{"type":"string","enum":["子","丑","寅","卯","辰","巳","午","未","申","酉","戌","亥"]}},"pillars":{"type":"array","items":{"type":"string","enum":["nian","yue","ri","shi"]}}},"required":["type","name","wuxing","branches","pillars"]}},"san_he":{"type":"array","description":"三合局","items":{"type":"object","additionalProperties":false,"properties":{"type":{"type":"string","enum":["三合"]},"name":{"type":"string"},"wuxing":{"type":"string","enum":["木","火","土","金","水"]},"branches":{"type":"array","items":{"type":"string","enum":["子","丑","寅","卯","辰","巳","午","未","申","酉","戌","亥"]}},"pillars":{"type":"array","items":{"type":"string","enum":["nian","yue","ri","shi"]}}},"required":["type","name","wuxing","branches","pillars"]}},"san_hui":{"type":"array","description":"三会方","items":{"type":"object","additionalProperties":false,"properties":{"type":{"type":"string","enum":["三会"]},"name":{"type":"string"},"wuxing":{"type":"string","enum":["木","火","土","金","水"]},"branches":{"type":"array","items":{"type":"string","enum":["子","丑","寅","卯","辰","巳","午","未","申","酉","戌","亥"]}},"pillars":{"type":"array","items":{"type":"string","enum":["nian","yue","ri","shi"]}}},"required":["type","name","wuxing","branches","pillars"]}},"liu_chong":{"type":"array","description":"六冲"},"liu_hai":{"type":"array","description":"六害"},"liu_xing":{"type":"array","description":"相刑"},"liu_po":{"type":"array","description":"六破","items":{"type":"object","additionalProperties":false,"properties":{"zhi_a":{"type":"string","enum":["子","丑","寅","卯","辰","巳","午","未","申","酉","戌","亥"]},"zhi_b":{"type":"string","enum":["子","丑","寅","卯","辰","巳","午","未","申","酉","戌","亥"]},"pillar_a":{"type":"string","enum":["nian","yue","ri","shi"]},"pillar_b":{"type":"string","enum":["nian","yue","ri","shi"]},"element":{"type":"string"}},"required":["zhi_a","zhi_b","pillar_a","pillar_b","element"]}},"an_he":{"type":"array","description":"地支暗合","items":{"type":"object","additionalProperties":false,"properties":{"zhi_a":{"type":"string","enum":["子","丑","寅","卯","辰","巳","午","未","申","酉","戌","亥"]},"zhi_b":{"type":"string","enum":["子","丑","寅","卯","辰","巳","午","未","申","酉","戌","亥"]},"pillar_a":{"type":"string","enum":["nian","yue","ri","shi"]},"pillar_b":{"type":"string","enum":["nian","yue","ri","shi"]},"element":{"type":"string"}},"required":["zhi_a","zhi_b","pillar_a","pillar_b","element"]}},"shen_sha_school":{"type":"object","additionalProperties":false,"properties":{"dual_reference":{"type":"array","items":{"type":"string","enum":["天乙贵人","桃花","驿马","华盖","将星","劫煞","灾煞"]}},"policy":{"type":"string","enum":["union_of_year_and_day_references"]}},"required":["dual_reference","policy"]},"lu_roots":{"type":"array","description":"透干十神得十干禄","items":{"type":"object","properties":{"shi_shen":{"type":"string"},"gan":{"type":"string"},"zhi":{"type":"string"}},"required":["shi_shen","gan","zhi"]}},"relation_groups":{"type":"array","description":"完整合会冲刑组；Python 只读","items":{"type":"object","additionalProperties":false,"properties":{"field":{"type":"string","enum":["gan_he","gan_he_candidate","gan_chong","gan_chong_candidate","zhi_liu_he","san_he","san_he_partial","san_hui","liu_chong","liu_hai","liu_xing","liu_po","an_he"]},"group":{"type":"string"}},"required":["field","group"]}},"ten_god_states":{"type":"array","description":"engine 十神显隐 / 通根 / 得令 / 旺弱状态","items":{"type":"object","properties":{"shi_shen":{"type":"string","enum":["正官","七杀","正印","偏印","正财","偏财","比肩","劫财","食神","伤官"]},"wuxing":{"type":"string","enum":["木","火","土","金","水"]},"transparent":{"type":"boolean"},"hidden":{"type":"boolean"},"rooted":{"type":"boolean"},"timely":{"type":"boolean"},"count":{"type":"integer","minimum":0},"strength":{"type":"string","enum":["strong","weak","neutral"]}},"required":["shi_shen","wuxing","transparent","hidden","rooted","timely","count","strength"]}},"element_states":{"type":"array","description":"engine 五行季节旺弱、组合旺弱与生克方向","items":{"type":"object","properties":{"wuxing":{"type":"string","enum":["木","火","土","金","水"]},"season":{"type":"string","enum":["旺","相","休","囚","死"]},"season_strength":{"type":"string","enum":["strong","weak"]},"strength":{"type":"string","enum":["strong","weak","neutral"]},"transparent":{"type":"boolean"},"rooted":{"type":"boolean"},"controls":{"type":"string","enum":["木","火","土","金","水"]},"controlled_by":{"type":"string","enum":["木","火","土","金","水"]},"controller_strength":{"type":"string","enum":["strong","weak"]},"reasons":{"type":"array","items":{"type":"string"}}},"required":["wuxing","season","season_strength","strength","transparent","rooted","controls","controlled_by","controller_strength","reasons"]}},"atomic_facts":{"type":"object","description":"engine 确定性命理原子事实；Python 因子层只读","properties":{"day_master_element":{"type":"string","enum":["木","火","土","金","水"]},"day_master_stem":{"type":"string","enum":["甲","乙","丙","丁","戊","己","庚","辛","壬","癸"]},"day_branch":{"type":"string","enum":["子","丑","寅","卯","辰","巳","午","未","申","酉","戌","亥"]},"month_longevity":{"type":"string","enum":["长生","沐浴","冠带","临官","帝旺","衰","病","死","墓","绝","胎","养",""]},"year_stem_ten_god":{"type":"string","enum":["正官","七杀","正印","偏印","正财","偏财","比肩","劫财","食神","伤官",""]},"month_main_ten_god":{"type":"string","enum":["正官","七杀","正印","偏印","正财","偏财","比肩","劫财","食神","伤官",""]},"hour_stem_ten_god":{"type":"string","enum":["正官","七杀","正印","偏印","正财","偏财","比肩","劫财","食神","伤官",""]},"pattern_god_transparent":{"type":"boolean"},"pillar_punishments":{"type":"object","properties":{"nian":{"type":"boolean"},"yue":{"type":"boolean"},"ri":{"type":"boolean"},"shi":{"type":"boolean"}},"required":["nian","yue","ri","shi"]},"officer_killing_cleaned":{"type":"boolean"},"wealth_tomb_present":{"type":"boolean"},"wealth_star_in_tomb":{"type":"boolean"},"spouse_palace_state":{"type":"string","enum":["冲","合","刑","害","静"]},"day_branch_type":{"type":"string","enum":["桃花","驿马","墓库",""]},"year_officer_killing":{"type":"boolean"}},"required":["day_master_element","day_master_stem","day_branch","month_longevity","year_stem_ten_god","month_main_ten_god","hour_stem_ten_god","pattern_god_transparent","pillar_punishments","officer_killing_cleaned","wealth_tomb_present","wealth_star_in_tomb","spouse_palace_state","day_branch_type","year_officer_killing"]},"fu_yi":{"type":"object","description":"扶抑派用神：真正的用/喜/忌 + 身强弱","properties":{"wuxing_count":{"type":"object","description":"五行数量"},"wang_shuai":{"type":"object","description":"五行旺相休囚死"},"yong":{"type":"string","description":"用神五行"},"xi":{"type":"string","description":"喜神五行"},"ji":{"type":"string","description":"忌神五行"},"qiangruo":{"type":"string","description":"身强弱"},"pattern":{"type":"string","description":"从格标签"},"model":{"type":"string"},"basis":{"type":"object","description":"强弱表输入","properties":{"root_type":{"type":"string"},"season":{"type":"string"},"yin_bi_count":{"type":"integer"},"day_master_roots":{"type":"array"},"relation_facts":{"type":"array"},"day_branch":{"type":"string"}}}}},"tiao_hou":{"type":"object","description":"调候派：主辅调候五行候选 + 季节 + 条件","properties":{"primary_wuxing":{"type":"string"},"secondary_wuxing":{"type":"string"},"season":{"type":"string"},"detail":{"type":"string"},"model":{"type":"string"},"primary":{"type":"object","properties":{"stem":{"type":"string"},"transparent":{"type":"boolean"},"hidden":{"type":"boolean"},"occurrences":{"type":"array"},"relation_facts":{"type":"array"}}},"secondary":{"type":"object"},"secondary_condition":{"type":"string"}}},"ge_ju":{"type":"object","description":"格局派：月令格神 + 顺逆用 + 结构（非用神）","properties":{"ge_ju":{"type":"string"},"yong_fa":{"type":"string"},"pattern_god":{"type":"string"},"pattern_god_ten_god":{"type":"string"},"pattern_god_source":{"type":"string"},"structure":{"type":"object","properties":{"pattern":{"type":"object"},"controller":{"type":"object"},"generator":{"type":"object"},"relation_facts":{"type":"array"}}}}},"day_xun":{"type":"string","description":"日柱所在旬"},"day_xun_kong":{"type":"string","description":"日柱旬空二支；各柱 is_void 均按此判断"}},"required":["nian","yue","ri","shi","da_yun","gender","fu_yi","tiao_hou","ge_ju","shen_sha_school","lu_roots","relation_groups","ten_god_states","element_states","atomic_facts","day_xun","day_xun_kong"]}`),
	},
	{
		Name: "bazi.chart", Description: "排八字命盘。返回最小命盘（四柱+纳音+大运+性别）。如需十神/藏干/神煞/自坐/空亡/用神三派等完整信息，请将结果传入 bazi.fullchart（用神属完整命盘）。",
		Params: mustSchema(`{"type":"object","properties":{"solar_time":` + schemaSolarTime + `,"gender":` + schemaGender + `},"required":["solar_time","gender"]}`), Handler: baziChartHandler,
		Result: envelopeSchema(`{"type":"object","properties":{"nian":{"type":"object","description":"年柱: gan/zhi/na_yin","properties":{"gan":{"type":"string","enum":["甲","乙","丙","丁","戊","己","庚","辛","壬","癸"]},"zhi":{"type":"string","enum":["子","丑","寅","卯","辰","巳","午","未","申","酉","戌","亥"]},"na_yin":{"type":"string"}},"required":["gan","zhi","na_yin"]},"yue":{"type":"object","description":"月柱: gan/zhi/na_yin","properties":{"gan":{"type":"string","enum":["甲","乙","丙","丁","戊","己","庚","辛","壬","癸"]},"zhi":{"type":"string","enum":["子","丑","寅","卯","辰","巳","午","未","申","酉","戌","亥"]},"na_yin":{"type":"string"}},"required":["gan","zhi","na_yin"]},"ri":{"type":"object","description":"日柱: gan/zhi/na_yin","properties":{"gan":{"type":"string","enum":["甲","乙","丙","丁","戊","己","庚","辛","壬","癸"]},"zhi":{"type":"string","enum":["子","丑","寅","卯","辰","巳","午","未","申","酉","戌","亥"]},"na_yin":{"type":"string"}},"required":["gan","zhi","na_yin"]},"shi":{"type":"object","description":"时柱: gan/zhi/na_yin","properties":{"gan":{"type":"string","enum":["甲","乙","丙","丁","戊","己","庚","辛","壬","癸"]},"zhi":{"type":"string","enum":["子","丑","寅","卯","辰","巳","午","未","申","酉","戌","亥"]},"na_yin":{"type":"string"}},"required":["gan","zhi","na_yin"]},"da_yun":{"type":"object","description":"大运。start_year_after/start_month_after/start_day_after 为出生后起运精确时间（3天=1年）；direction 顺排/逆排；steps 每步含 gan/zhi/name/wuxing/shi_shen，与 start_date/end_date（公历日期段）/start_year/end_year（公历年）；current_step_index 为当前所处大运索引（-1 表示未起运或已过完所有大运）；next_step_in_years 距下一大运剩余年数（公历口径）；steps 每步含 start_date/end_date（公历日期段）与 start_year/end_year（公历年，免虚岁换算）"},"gender":{"type":"string"},"birth_year":{"type":"integer","description":"出生公历年份（供 bazi.liunian/liuri 定位当年大运）"},"zi_shi_rule":{"type":"string","description":"子时换日规则说明（晚子时日柱不变、时柱按次日日干起）"}},"required":["nian","yue","ri","shi","da_yun","gender"]}`),
	},
	{
		Name: "bazi.bond", Description: "八字合盘原始事实。返回双方日主、四柱干支关系、十神视角、纳音、实盘五行与配偶星/夫妻宫事实；不做合婚评级。",
		Params:  mustSchema(`{"type":"object","properties":{"a":{"type":"object","properties":{"chart":{"type":"object","description":"八字命盘（由 bazi.chart 返回的最小命盘，不需 bazi.fullchart）"}},"required":["chart"]},"b":{"type":"object","properties":{"chart":{"type":"object","description":"八字命盘（由 bazi.chart 返回的最小命盘，不需 bazi.fullchart）"}},"required":["chart"]}},"required":["a","b"]}`),
		Handler: baziBondHandler,
		Result:  envelopeSchema(`{"type":"object","properties":{"day_master":{"type":"object","properties":{"a":{"type":"object","properties":{"gan":{"type":"string"},"wuxing":{"type":"string","enum":["木","火","土","金","水"]},"gender":{"type":"string","enum":["male","female"]}},"required":["gan","wuxing","gender"],"additionalProperties":false},"b":{"type":"object","properties":{"gan":{"type":"string"},"wuxing":{"type":"string","enum":["木","火","土","金","水"]},"gender":{"type":"string","enum":["male","female"]}},"required":["gan","wuxing","gender"],"additionalProperties":false},"gan_guan_xi":{"type":"object","properties":{"gan_a":{"type":"string"},"gan_b":{"type":"string"},"type":{"type":"string"},"relation":{"type":"string"}},"required":["gan_a","gan_b","type","relation"],"additionalProperties":false}},"required":["a","b","gan_guan_xi"],"additionalProperties":false},"spouse_palace":{"type":"object","properties":{"a_zhi":{"type":"string"},"b_zhi":{"type":"string"},"relation":{"type":"object","properties":{"zhi_a":{"type":"string"},"zhi_b":{"type":"string"},"type":{"type":"string"},"detail":{"type":"string"}},"required":["zhi_a","zhi_b","type","detail"],"additionalProperties":false}},"required":["a_zhi","b_zhi","relation"],"additionalProperties":false},"zhu_cross":{"type":"object","properties":{"pairs":{"type":"array","items":{"type":"object","properties":{"a_zhu":{"type":"string","enum":["nian","yue","ri","shi"]},"b_zhu":{"type":"string","enum":["nian","yue","ri","shi"]},"a_gan":{"type":"string"},"b_gan":{"type":"string"},"a_zhi":{"type":"string"},"b_zhi":{"type":"string"},"gan_guan_xi":{"type":"object","properties":{"gan_a":{"type":"string"},"gan_b":{"type":"string"},"type":{"type":"string"},"relation":{"type":"string"}},"required":["gan_a","gan_b","type","relation"],"additionalProperties":false},"zhi_guan_xi":{"type":"object","properties":{"zhi_a":{"type":"string"},"zhi_b":{"type":"string"},"type":{"type":"string"},"detail":{"type":"string"}},"required":["zhi_a","zhi_b","type","detail"],"additionalProperties":false}},"required":["a_zhu","b_zhu","a_gan","b_gan","a_zhi","b_zhi","gan_guan_xi","zhi_guan_xi"],"additionalProperties":false},"minItems":16,"maxItems":16}},"required":["pairs"],"additionalProperties":false},"shi_shen_cross":{"type":"object","properties":{"a_to_b":{"type":"object","properties":{"nian_gan":{"type":"string","enum":["比肩","劫财","食神","伤官","正财","偏财","七杀","正官","偏印","正印"]},"yue_gan":{"type":"string","enum":["比肩","劫财","食神","伤官","正财","偏财","七杀","正官","偏印","正印"]},"ri_gan":{"type":"string","enum":["比肩","劫财","食神","伤官","正财","偏财","七杀","正官","偏印","正印"]},"shi_gan":{"type":"string","enum":["比肩","劫财","食神","伤官","正财","偏财","七杀","正官","偏印","正印"]}},"required":["nian_gan","yue_gan","ri_gan","shi_gan"],"additionalProperties":false},"b_to_a":{"type":"object","properties":{"nian_gan":{"type":"string","enum":["比肩","劫财","食神","伤官","正财","偏财","七杀","正官","偏印","正印"]},"yue_gan":{"type":"string","enum":["比肩","劫财","食神","伤官","正财","偏财","七杀","正官","偏印","正印"]},"ri_gan":{"type":"string","enum":["比肩","劫财","食神","伤官","正财","偏财","七杀","正官","偏印","正印"]},"shi_gan":{"type":"string","enum":["比肩","劫财","食神","伤官","正财","偏财","七杀","正官","偏印","正印"]}},"required":["nian_gan","yue_gan","ri_gan","shi_gan"],"additionalProperties":false}},"required":["a_to_b","b_to_a"],"additionalProperties":false},"wuxing_cross":{"type":"object","properties":{"a":{"type":"object","properties":{"木":{"type":"integer","minimum":0},"火":{"type":"integer","minimum":0},"土":{"type":"integer","minimum":0},"金":{"type":"integer","minimum":0},"水":{"type":"integer","minimum":0}},"required":["木","火","土","金","水"],"additionalProperties":false},"b":{"type":"object","properties":{"木":{"type":"integer","minimum":0},"火":{"type":"integer","minimum":0},"土":{"type":"integer","minimum":0},"金":{"type":"integer","minimum":0},"水":{"type":"integer","minimum":0}},"required":["木","火","土","金","水"],"additionalProperties":false},"combined":{"type":"object","properties":{"木":{"type":"integer","minimum":0},"火":{"type":"integer","minimum":0},"土":{"type":"integer","minimum":0},"金":{"type":"integer","minimum":0},"水":{"type":"integer","minimum":0}},"required":["木","火","土","金","水"],"additionalProperties":false},"fu_yi_in_other":{"type":"object","properties":{"a":{"type":"object","properties":{"model":{"type":"string"},"yong":{"type":"string"},"xi":{"type":"string"},"ji":{"type":"string"},"yong_in_other":{"type":"integer","minimum":0},"xi_in_other":{"type":"integer","minimum":0},"ji_in_other":{"type":"integer","minimum":0}},"required":["model","yong","xi","ji","yong_in_other","xi_in_other","ji_in_other"],"additionalProperties":false},"b":{"type":"object","properties":{"model":{"type":"string"},"yong":{"type":"string"},"xi":{"type":"string"},"ji":{"type":"string"},"yong_in_other":{"type":"integer","minimum":0},"xi_in_other":{"type":"integer","minimum":0},"ji_in_other":{"type":"integer","minimum":0}},"required":["model","yong","xi","ji","yong_in_other","xi_in_other","ji_in_other"],"additionalProperties":false}},"required":["a","b"],"additionalProperties":false}},"required":["a","b","combined","fu_yi_in_other"],"additionalProperties":false},"nayin_cross":{"type":"object","properties":{"pairs":{"type":"array","items":{"type":"object","properties":{"a_zhu":{"type":"string","enum":["nian","yue","ri","shi"]},"b_zhu":{"type":"string","enum":["nian","yue","ri","shi"]},"a_na_yin":{"type":"string"},"b_na_yin":{"type":"string"},"relation":{"type":"string"}},"required":["a_zhu","b_zhu","a_na_yin","b_na_yin","relation"],"additionalProperties":false},"minItems":16,"maxItems":16},"wuxing_counts":{"type":"object","properties":{"a":{"type":"object","additionalProperties":{"type":"integer","minimum":0}},"b":{"type":"object","additionalProperties":{"type":"integer","minimum":0}}},"required":["a","b"],"additionalProperties":false}},"required":["pairs","wuxing_counts"],"additionalProperties":false},"spouse_star":{"type":"object","properties":{"a":{"type":"object","properties":{"gender":{"type":"string","enum":["male","female"]},"primary":{"type":"object","properties":{"ten_god":{"type":"string"},"occurrences":{"type":"array","items":{"type":"object","properties":{"pillar":{"type":"string","enum":["nian","yue","ri","shi"]},"branch":{"type":"string"},"stem":{"type":"string"},"source":{"type":"string"},"is_void":{"type":"boolean"}},"required":["pillar","branch","stem","source","is_void"],"additionalProperties":false}}},"required":["ten_god","occurrences"],"additionalProperties":false},"secondary":{"type":"object","properties":{"ten_god":{"type":"string"},"occurrences":{"type":"array","items":{"type":"object","properties":{"pillar":{"type":"string","enum":["nian","yue","ri","shi"]},"branch":{"type":"string"},"stem":{"type":"string"},"source":{"type":"string"},"is_void":{"type":"boolean"}},"required":["pillar","branch","stem","source","is_void"],"additionalProperties":false}}},"required":["ten_god","occurrences"],"additionalProperties":false}},"required":["gender","primary","secondary"],"additionalProperties":false},"b":{"type":"object","properties":{"gender":{"type":"string","enum":["male","female"]},"primary":{"type":"object","properties":{"ten_god":{"type":"string"},"occurrences":{"type":"array","items":{"type":"object","properties":{"pillar":{"type":"string","enum":["nian","yue","ri","shi"]},"branch":{"type":"string"},"stem":{"type":"string"},"source":{"type":"string"},"is_void":{"type":"boolean"}},"required":["pillar","branch","stem","source","is_void"],"additionalProperties":false}}},"required":["ten_god","occurrences"],"additionalProperties":false},"secondary":{"type":"object","properties":{"ten_god":{"type":"string"},"occurrences":{"type":"array","items":{"type":"object","properties":{"pillar":{"type":"string","enum":["nian","yue","ri","shi"]},"branch":{"type":"string"},"stem":{"type":"string"},"source":{"type":"string"},"is_void":{"type":"boolean"}},"required":["pillar","branch","stem","source","is_void"],"additionalProperties":false}}},"required":["ten_god","occurrences"],"additionalProperties":false}},"required":["gender","primary","secondary"],"additionalProperties":false}},"required":["a","b"],"additionalProperties":false},"shensha_cross":{"type":"object","properties":{"tian_yi":{"type":"object","properties":{"a_present_in_b":{"type":"boolean"},"b_present_in_a":{"type":"boolean"}},"required":["a_present_in_b","b_present_in_a"],"additionalProperties":false},"lu":{"type":"object","properties":{"a_present_in_b":{"type":"boolean"},"b_present_in_a":{"type":"boolean"}},"required":["a_present_in_b","b_present_in_a"],"additionalProperties":false},"tao_hua":{"type":"object","properties":{"a_present_in_b":{"type":"boolean"},"b_present_in_a":{"type":"boolean"}},"required":["a_present_in_b","b_present_in_a"],"additionalProperties":false},"yi_ma":{"type":"object","properties":{"a_present_in_b":{"type":"boolean"},"b_present_in_a":{"type":"boolean"}},"required":["a_present_in_b","b_present_in_a"],"additionalProperties":false},"kong_wang":{"type":"object","properties":{"a_present_in_b":{"type":"boolean"},"b_present_in_a":{"type":"boolean"}},"required":["a_present_in_b","b_present_in_a"],"additionalProperties":false},"kui_gang":{"type":"object","properties":{"a_present_in_b":{"type":"boolean"},"b_present_in_a":{"type":"boolean"}},"required":["a_present_in_b","b_present_in_a"],"additionalProperties":false},"ri_de":{"type":"object","properties":{"a_present_in_b":{"type":"boolean"},"b_present_in_a":{"type":"boolean"}},"required":["a_present_in_b","b_present_in_a"],"additionalProperties":false},"ri_gui":{"type":"object","properties":{"a_present_in_b":{"type":"boolean"},"b_present_in_a":{"type":"boolean"}},"required":["a_present_in_b","b_present_in_a"],"additionalProperties":false}},"required":["tian_yi","lu","tao_hua","yi_ma","kong_wang","kui_gang","ri_de","ri_gui"],"additionalProperties":false}},"required":["day_master","spouse_palace","zhu_cross","shi_shen_cross","wuxing_cross","nayin_cross","spouse_star","shensha_cross"],"additionalProperties":false}`),
	},
	{
		Name: "bazi.liunian", Description: "八字流年运势。返回流年干支与命局的十神、神煞、伏吟反吟。",
		Params:  mustSchema(`{"type":"object","properties":{"year":{"type":"integer","minimum":1,"maximum":8000,"description":"目标年份（VSOP87D 天文精度范围 1-8000）"},"chart":{"type":"object","description":"八字命盘（由 bazi.chart 返回的最小命盘，不需 bazi.fullchart）"}},"required":["year","chart"]}`),
		Handler: baziLiunianHandler,
		Result:  envelopeSchema(`{"type":"object","properties":{"year":{"type":"integer"},"nian_name":{"type":"string"},"shi_shen":{"type":"string","enum":["正官","七杀","正印","偏印","正财","偏财","比肩","劫财","食神","伤官"]},"nian_gan":{"type":"string","description":"流年天干"},"nian_zhi":{"type":"string","description":"流年地支"},"wuxing":{"type":"string"},"na_yin":{"type":"string"},"sheng":{"type":"integer","description":"生数"},"ke":{"type":"integer","description":"克数"},"shensha":{"type":"array","description":"流年神煞（动态：桃花/驿马/华盖/劫煞/灾煞/红鸾/天喜/天乙贵人/羊刃——年支+日支双查；值年：病符/丧门/吊客/大耗——命局四柱逢煞支即应）","items":{"type":"object","properties":{"name":{"type":"string","enum":["桃花","驿马","华盖","劫煞","灾煞","红鸾","天喜","天乙贵人","羊刃","病符","丧门","吊客","大耗"]},"category":{"type":"string","enum":["吉","凶","中性"]},"description":{"type":"string"}}}},"fuyin_fanyin":{"type":"array","description":"伏吟反吟"},"dayun_interactions":{"type":"array","description":"查询年份所在大运与流年交互（chart 需含 birth_year；未起运/大运过完/缺 birth_year 时为空数组）"},"natal_interactions":{"type":"array","description":"流年与命局交互"},"atomic_facts":{"type":"object","description":"engine 流年原子事实","properties":{"controls_targets":{"type":"object","properties":{"day_master":{"type":"boolean"},"wealth_star":{"type":"boolean"},"officer_killing":{"type":"boolean"},"seal_star":{"type":"boolean"},"food_injury":{"type":"boolean"}},"additionalProperties":false,"required":["day_master","wealth_star","officer_killing","seal_star","food_injury"]},"controls_elements":{"type":"array","items":{"type":"string","enum":["木","火","土","金","水"]}},"unfavorable_gan":{"type":"boolean"},"unfavorable_branch":{"type":"boolean"},"wealth_breaks_seal":{"type":"boolean"},"year_branch_relations":{"type":"object","additionalProperties":{"type":"string","enum":["liu_he","half_he","san_he_gong","san_he","san_hui_gong","san_hui","liu_chong","xing","liu_hai"]}},"year_branch_controlled_by":{"type":"array","items":{"type":"string","enum":["木","火","土","金","水"]}},"year_gan_controls_day_gan":{"type":"boolean"},"day_gan_controls_year_gan":{"type":"boolean"},"gan_combines":{"type":"boolean"},"dayun_gan_combines":{"type":"boolean"},"dayun_zhi_clashes_year":{"type":"boolean"},"day_void_branches":{"type":"array","items":{"type":"string"}},"year_equals_day_pillar":{"type":"boolean"},"dayun_equals_year_pillar":{"type":"boolean"},"year_gan_equals_natal_year_gan":{"type":"boolean"},"combinations":{"type":"array","items":{"type":"object","properties":{"kind":{"type":"string","enum":["san_he","san_hui","xing","half_he"]},"group":{"type":"string"},"branches":{"type":"array","items":{"type":"string"}},"includes_year":{"type":"boolean"}},"required":["kind","group","branches","includes_year"]}}},"required":["controls_elements","controls_targets","unfavorable_gan","unfavorable_branch","wealth_breaks_seal","year_branch_relations","year_branch_controlled_by","year_gan_controls_day_gan","day_gan_controls_year_gan","gan_combines","dayun_gan_combines","dayun_zhi_clashes_year","day_void_branches","year_equals_day_pillar","dayun_equals_year_pillar","year_gan_equals_natal_year_gan","combinations"]}},"required":["year","nian_name","shi_shen","atomic_facts"]}`),
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
		Result:  envelopeSchema(`{"type":"array","items":{"type":"object","properties":{"method":{"type":"string","description":"流派标识"},"basis":{"type":"string","description":"起法依据"},"zhus":{"type":"array","items":{"type":"object","properties":{"age":{"type":"integer"},"gan":{"type":"string","enum":["甲","乙","丙","丁","戊","己","庚","辛","壬","癸"]},"zhi":{"type":"string","enum":["子","丑","寅","卯","辰","巳","午","未","申","酉","戌","亥"]},"name":{"type":"string"},"shi_shen":{"type":"string"}},"required":["age","gan","zhi","name"]}}},"required":["method","basis","zhus"]}}`),
	},
}
