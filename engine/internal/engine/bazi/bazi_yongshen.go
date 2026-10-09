package bazi

import "liki-engine/internal/engine/ganzhi"

// FuYiResult is the FuYi (扶抑) yongshen analysis based on day-master strength.
// Pattern is non-empty when a special pattern is detected (e.g. 从旺格、从杀格).
type FuYiResult struct {
	WuxingCount map[string]int    `json:"wuxing_count"`
	WangShuai   map[string]string `json:"wang_shuai"`
	Yong        string            `json:"yong"`
	Xi          string            `json:"xi"`
	Ji          string            `json:"ji"`
	Strength    string            `json:"qiangruo"`
	Pattern     string            `json:"pattern,omitempty"`
	Model       string            `json:"model"`
	Basis       FuYiBasis         `json:"basis"`
}

// TiaoHouResult is the TiaoHou (调候) yongshen analysis based on climate.
type TiaoHouResult struct {
	PrimaryWuxing      string               `json:"primary_wuxing"`
	SecondaryWuxing    string               `json:"secondary_wuxing"`
	Season             string               `json:"season"`
	Detail             string               `json:"detail"`
	Model              string               `json:"model"`
	Primary            TiaoHouAvailability  `json:"primary"`
	Secondary          *TiaoHouAvailability `json:"secondary,omitempty"`
	SecondaryCondition string               `json:"secondary_condition,omitempty"`
}

// GeJuResult is the GeJu (格局) structural candidate based on the month
// command. It deliberately does not emit yong/xi/ji: pattern success,
// rescue, and final favorable elements require whole-chart review.
type GeJuResult struct {
	Pattern string `json:"ge_ju"`
	Usage   string `json:"yong_fa"` // "顺用" or "逆用"
	// 结构事实：月令格神与来源。当前模型未完整实现相神、成格、败格、
	// 救应，因此 Pattern 是格局候选，不是子平真诠完整成格结论。
	PatternGod       string        `json:"pattern_god,omitempty"`
	PatternGodTenGod string        `json:"pattern_god_ten_god,omitempty"`
	PatternGodSource string        `json:"pattern_god_source"`
	Structure        GeJuStructure `json:"structure"`
}

// ComputeYongShenSchools computes the three yongshen schools from a Chart.
// fu_yi is the actual yongshen (yong/xi/ji); tiao_hou and ge_ju are
// independent evidence sources that are NOT yongshen.
func ComputeYongShenSchools(c Chart) (FuYiResult, TiaoHouResult, GeJuResult) {
	wc := computeElementCount(c.ToBazi(), computeCangGan(c.ToBazi()))
	ws := computeWangShuaiMap(c)

	return computeFuYi(c, wc, ws), computeTiaoHou(c), computeGeJu(c, wc)
}

// computeWangShuaiMap returns the 旺相休囚死 for all five elements.
func computeWangShuaiMap(c Chart) map[string]string {
	return map[string]string{
		ganzhi.WxMu.String():   ganzhi.WangShuaiOf(ganzhi.WxMu, c.Yue.Zhi).String(),
		ganzhi.WxHuo.String():  ganzhi.WangShuaiOf(ganzhi.WxHuo, c.Yue.Zhi).String(),
		ganzhi.WxTu.String():   ganzhi.WangShuaiOf(ganzhi.WxTu, c.Yue.Zhi).String(),
		ganzhi.WxJin.String():  ganzhi.WangShuaiOf(ganzhi.WxJin, c.Yue.Zhi).String(),
		ganzhi.WxShui.String(): ganzhi.WangShuaiOf(ganzhi.WxShui, c.Yue.Zhi).String(),
	}
}
