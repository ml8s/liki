package bazi

import "liki-engine/internal/engine/ganzhi"

// computeFuYi computes the FuYi (扶抑) yongshen analysis from a Chart.
// Uses data-driven lookup tables for strength, congGe, and yong/xi/ji.
func computeFuYi(c Chart, wc map[ganzhi.Wuxing]int, ws map[string]string) FuYiResult {
	riYuan := c.Ri.Gan
	cangGan := computeCangGan(c.ToBazi())
	yueZhi := c.Yue.Zhi

	// Use lookup-based strength (qualitative root+season+印比, not scoring).
	rootType := classifyRoot(riYuan, yueZhi, cangGan)
	season := classifySeason(riYuan, yueZhi)
	yinBi := countYinBi(c)
	strengthLabel := lookupStrength(rootType, season, yinBi)
	basis := buildFuYiBasis(c, rootType, season, yinBi)

	// Try congGe rules first (qualitative rule chain, not percentages).
	pat, yong, xi, ji := lookupCongGe(c)
	if pat != "" {
		wcStr := make(map[string]int, len(wc))
		for k, v := range wc {
			wcStr[k.String()] = v
		}
		return FuYiResult{
			WuxingCount: wcStr,
			WangShuai:   ws,
			Yong:        yong,
			Xi:          xi,
			Ji:          ji,
			Strength:    strengthLabel,
			Pattern:     pat,
			Model:       "support_control_with_day_master_strength",
			Basis:       basis,
		}
	}

	// Normal扶抑 (qualitative, no element count comparison).
	yong, xi, ji = computeNormalYongJi(riYuan, strengthLabel)

	wcStr := make(map[string]int, len(wc))
	for k, v := range wc {
		wcStr[k.String()] = v
	}

	return FuYiResult{
		WuxingCount: wcStr,
		WangShuai:   ws,
		Yong:        yong,
		Xi:          xi,
		Ji:          ji,
		Strength:    strengthLabel,
		Model:       "support_control_with_day_master_strength",
		Basis:       basis,
	}
}

// computeNormalYongJi determines yong/xi/ji by strength (qualitative, no scoring).
func computeNormalYongJi(riYuan ganzhi.Gan, strengthLabel string) (yongShen, xiShen, jiShen string) {
	dmElem := ganzhi.GanWuxing(riYuan)
	rule, ok := yongJiRules[strengthLabel]
	if !ok {
		return "", "", ""
	}
	// 五行关系 → 具体五行：克我者官杀、生我者印、同我者比劫、生克我者财生官杀。
	resolve := func(rel string) ganzhi.Wuxing {
		switch rel {
		case "克我者":
			return elementThatControls(dmElem)
		case "生我者":
			return elementThatGenerates(dmElem)
		case "同我者":
			return dmElem
		case "生克我者":
			return elementThatGenerates(elementThatControls(dmElem))
		}
		return 0
	}
	yongShen = resolve(rule.Yong).String()
	xiShen = resolve(rule.Xi).String()
	jiShen = resolve(rule.Ji).String()
	return
}
