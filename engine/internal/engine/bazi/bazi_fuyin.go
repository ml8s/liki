package bazi

import "liki-engine/internal/engine/ganzhi"

// FuYinFanYin describes a 伏吟 or 反吟 hit between a flow pillar and a bazi pillar.
type FuYinFanYin struct {
	NatalIndex int    `json:"natal_index"` // 0-3 (年/月/日/时)
	Type       string `json:"type"`        // "伏吟" or "反吟"
	Detail     string `json:"detail"`
}

// computeFuYinFanYin checks a flow pillar (e.g., liunian or dayun) against each
// bazi pillar for 伏吟 (same gan+zhi) and 反吟 (gan clash + zhi clash —
// 天克地冲).
func computeFuYinFanYin(flow ganzhi.Zhu, bz ganzhi.Bazi) []FuYinFanYin {
	bazi := bz.Slice()
	var entries []FuYinFanYin

	for i, np := range bazi {
		sameGan := flow.Gan == np.Gan
		sameZhi := flow.Zhi == np.Zhi

		if sameGan && sameZhi {
			entries = append(entries, FuYinFanYin{
				NatalIndex: i,
				Type:       "伏吟",
				Detail:     ganzhi.GanName(flow.Gan) + ganzhi.ZhiName(flow.Zhi) + "伏吟",
			})
		} else if sameZhi && !sameGan {
			entries = append(entries, FuYinFanYin{
				NatalIndex: i,
				Type:       "伏吟",
				Detail:     ganzhi.ZhiName(flow.Zhi) + "地支伏吟",
			})
		}

		// 反吟: 天克地冲 (gan clash AND zhi clash)
		sr := analyzeGanRelation(flow.Gan, np.Gan)
		br := analyzeZhiRelation(flow.Zhi, np.Zhi)
		if sr.Type == relKe && br.Type == relLiuChong {
			entries = append(entries, FuYinFanYin{
				NatalIndex: i,
				Type:       "反吟",
				Detail: ganzhi.GanName(flow.Gan) + ganzhi.ZhiName(flow.Zhi) +
					"与" + ganzhi.GanName(np.Gan) + ganzhi.ZhiName(np.Zhi) + "天克地冲",
			})
		}
	}

	if entries == nil {
		return []FuYinFanYin{}
	}
	return entries
}
