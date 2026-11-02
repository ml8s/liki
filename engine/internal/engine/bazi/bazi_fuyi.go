package bazi

import "liki-engine/internal/engine/ganzhi"

// computeFuYi computes the FuYi (扶抑) yongshen analysis from pillars.
// Uses data-driven lookup tables for strength, congGe, and yong/xi/ji.
func FuYi(
	bz ganzhi.Bazi,
	counts WuxingCounts,
	seasonStrength SeasonStrength,
	strength DayMasterStrength,
) FuYiResult {
	riYuan := bz.Ri.Gan
	basis := buildFuYiBasis(bz, strength.RootType, strength.YinBiCount)

	// Try congGe rules first (qualitative rule chain, not percentages).
	pat, congYong, congXi, congJi := lookupCongGe(bz)
	if pat != "" {
		return FuYiResult{
			WuxingCount: counts.StringCount(),
			WangShuai:   seasonStrength,
			Yong:        candidateString(congYong),
			Xi:          candidateString(congXi),
			Ji:          candidateString(congJi),
			Strength:    strength.Label,
			Pattern:     pat,
			ModelID:     "support_control_with_day_master_strength",
			Basis:       basis,
		}
	}

	// Normal扶抑 (qualitative, no element count comparison).
	yong, xi, ji := computeNormalYongJi(riYuan, strength.Label)

	return FuYiResult{
		WuxingCount: counts.StringCount(),
		WangShuai:   seasonStrength,
		Yong:        yong,
		Xi:          xi,
		Ji:          ji,
		Strength:    strength.Label,
		ModelID:     "support_control_with_day_master_strength",
		Basis:       basis,
	}
}

// computeNormalYongJi determines yong/xi/ji by strength (qualitative, no scoring).
func computeNormalYongJi(riYuan ganzhi.Gan, strengthLabel string) (yongShen, xiShen, jiShen string) {
	dmElem := ganzhi.GanWuxing(riYuan)
	rule, ok := yongJiRules[strengthLabel]
	if !ok {
		return NoCandidateWuxing, NoCandidateWuxing, NoCandidateWuxing
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
	return candidateString(resolve(rule.Yong).String()),
		candidateString(resolve(rule.Xi).String()),
		candidateString(resolve(rule.Ji).String())
}
