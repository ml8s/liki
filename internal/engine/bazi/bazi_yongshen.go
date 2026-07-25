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
}

// TiaoHouResult is the TiaoHou (调候) yongshen analysis based on climate.
type TiaoHouResult struct {
	Yong   string `json:"yong"`
	Xi     string `json:"xi"`
	Ji     string `json:"ji"`
	Season string `json:"season"`
	Detail string `json:"detail"`
}

// GeJuResult is the GeJu (格局) yongshen analysis based on chart pattern.
type GeJuResult struct {
	Yong    string `json:"yong"`
	Xi      string `json:"xi"`
	Ji      string `json:"ji"`
	Pattern string `json:"ge_ju"`
	Usage   string `json:"yong_fa"` // "顺用" or "逆用"
}

// YongShenResult holds the three-school yongshen analysis.
type YongShenResult struct {
	FuYi    FuYiResult    `json:"fu_yi"`
	TiaoHou TiaoHouResult `json:"tiao_hou"`
	GeJu    GeJuResult    `json:"ge_ju"`
}

// ComputeYongShen computes yongshen from a Chart alone.
func ComputeYongShen(c Chart) YongShenResult {
	wc := computeElementCount(c.ToBazi(), computeCangGan(c.ToBazi()))
	ws := computeWangShuaiMap(c)

	return YongShenResult{
		FuYi:    computeFuYi(c, wc, ws),
		TiaoHou: computeTiaoHou(c.Ri.Gan, c.Yue.Zhi),
		GeJu:    computeGeJu(c, wc),
	}
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
