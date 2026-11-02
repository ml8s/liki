package bazi

import "liki-engine/internal/engine/ganzhi"

// Candidate closures replace nullable domain facts with explicit states.
const (
	NoCandidateWuxing = "无候选"
	NoSecondaryWuxing = "无辅调候"
	NoSelfHeName      = "无自合"
)

// FuYiResult is the FuYi (扶抑) yongshen analysis based on day-master strength.
// Pattern is non-empty when a special pattern is detected (e.g. 从旺格、从杀格).
type FuYiResult struct {
	WuxingCount map[string]int    `json:"wuxing_count"`
	WangShuai   map[string]string `json:"wang_shuai"`
	Yong        string            `json:"yong"`
	Xi          string            `json:"xi"`
	Ji          string            `json:"ji"`
	Strength    string            `json:"qiang_ruo"`
	Pattern     string            `json:"pattern,omitempty"`
	ModelID     string            `json:"fu_yi_model_id"`
	Basis       FuYiBasis         `json:"basis"`
}

// TiaoHouResult is the TiaoHou (调候) yongshen analysis based on climate.
type TiaoHouResult struct {
	PrimaryWuxing      string               `json:"primary_wuxing"`
	SecondaryWuxing    string               `json:"secondary_wuxing"`
	Season             string               `json:"season"`
	TiaoHouExplanation string               `json:"tiao_hou_explanation"`
	ModelID            string               `json:"tiao_hou_model_id"`
	Primary            TiaoHouAvailability  `json:"primary"`
	Secondary          *TiaoHouAvailability `json:"secondary,omitempty"`
	SecondaryCondition string               `json:"secondary_condition,omitempty"`
}

// GeJuResult is the GeJu (格局) structural candidate based on the month
// command. It deliberately does not emit yong/xi/ji: pattern success,
// rescue, and final favorable elements require whole-chart review.
type GeJuResult struct {
	Pattern string `json:"ge_ju"`
	// GeShenTou is the pattern-god transparency conclusion.
	GeShenTou bool   `json:"ge_shen_tou"`
	Usage     string `json:"yong_fa"` // "顺用" or "逆用"
	// 结构事实：月令格神与来源。当前模型未完整实现相神、成格、败格、
	// 救应，因此 Pattern 是格局候选，不是子平真诠完整成格结论。
	PatternGod       string        `json:"pattern_god,omitempty"`
	PatternGodTenGod string        `json:"pattern_god_ten_god,omitempty"`
	PatternGodSource string        `json:"pattern_god_source"`
	Structure        GeJuStructure `json:"structure"`
}

// GuanSha is the officer/killing mixture conclusion.
type GuanSha struct {
	QuQing bool `json:"qu_qing"`
}

// CaiKu is the wealth-tomb conclusion.
type CaiKu struct {
	Xian     bool `json:"xian"`
	XingRuKu bool `json:"xing_ru_ku"`
}

// ComputeYongshenEvidence computes the FuYi/TiaoHou candidates and the GeJu
// structural evidence. GeJu is not a Yongshen candidate.
func computeYongShenEvidenceFromAnalysis(
	core Core, analysis PillarAnalysis, relations []NatalRelationFact,
) (FuYiResult, TiaoHouResult, GeJuResult) {
	bz := core.ToBazi()
	counts := WuxingCountsAt(bz, analysis.HiddenStems)
	seasonStrength := SeasonStrengthAt(bz.Yue.Zhi)
	strength := AnalyzeDayMasterStrength(bz, analysis.HiddenStems)

	return FuYi(bz, counts, seasonStrength, strength, relations),
		TiaoHouAt(bz, relations),
		GeJuAt(bz, analysis, relations)
}

// candidateString models an absent domain candidate as an explicit closure value.
func candidateString(value string) string {
	if value == "" {
		return NoCandidateWuxing
	}
	return value
}

// computeWangShuaiMap returns the 旺相休囚死 for all five elements.
type SeasonStrength map[string]string

func SeasonStrengthAt(yueZhi ganzhi.Zhi) SeasonStrength {
	return map[string]string{
		ganzhi.WxMu.String():   ganzhi.WangShuaiOf(ganzhi.WxMu, yueZhi).String(),
		ganzhi.WxHuo.String():  ganzhi.WangShuaiOf(ganzhi.WxHuo, yueZhi).String(),
		ganzhi.WxTu.String():   ganzhi.WangShuaiOf(ganzhi.WxTu, yueZhi).String(),
		ganzhi.WxJin.String():  ganzhi.WangShuaiOf(ganzhi.WxJin, yueZhi).String(),
		ganzhi.WxShui.String(): ganzhi.WangShuaiOf(ganzhi.WxShui, yueZhi).String(),
	}
}
