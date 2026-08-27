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
	}
}

// computeNormalYongJi determines yong/xi/ji by strength (qualitative, no scoring).
func computeNormalYongJi(riYuan ganzhi.Gan, strengthLabel string) (yongShen, xiShen, jiShen string) {
	dmElem := ganzhi.GanWuxing(riYuan)

	switch strengthLabel {
	case "身强":
		ctrlElem := elementThatControls(dmElem)
		yongShen = ctrlElem.String()                     // 官杀克身
		xiShen = elementThatGenerates(ctrlElem).String() // 财生官杀
		jiShen = elementThatGenerates(dmElem).String()   // 印生日主(引发过旺)

	case "身弱":
		genElem := elementThatGenerates(dmElem)
		yongShen = genElem.String()                   // 印生日主
		xiShen = dmElem.String()                      // 比劫帮身
		jiShen = elementThatControls(dmElem).String() // 官杀克身

	case "中和":
		// 中和者无太过不及, 不应扶抑.
	}
	return
}
