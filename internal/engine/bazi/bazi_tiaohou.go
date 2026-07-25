package bazi

import (
	"liki-engine/internal/engine/ganzhi"
)

// TiaoHou is the internal 穷通宝鉴 climate-adjustment result.
type TiaoHou struct {
	Season string
	Yong   string
	Xi     string
	Ji     string // empty when no clear 忌神
	Detail string
}

// tiaohouKey is the internal compound key for the lookup table.
type tiaohouKey struct {
	stem   int
	branch int
}

// computeTiaoHou returns the TiaoHou (调候) yongshen analysis for the given
// day-master and month-branch. Based on 穷通宝鉴.
func computeTiaoHou(riYuan ganzhi.Gan, monthBranch ganzhi.Zhi) TiaoHouResult {
	th, _ := queryTiaoHou(riYuan, monthBranch)
	return TiaoHouResult{
		Yong:   th.Yong,
		Xi:     th.Xi,
		Ji:     th.Ji,
		Season: th.Season,
		Detail: th.Detail,
	}
}

// queryTiaoHou returns the 穷通宝鉴 climate-adjustment result for a given
// day-master and month-branch. Returns (TiaoHou, true) on match, or
// (zero, false) if no entry exists.
func queryTiaoHou(riYuan ganzhi.Gan, monthBranch ganzhi.Zhi) (TiaoHou, bool) {
	e, ok := lookupTiaohou[tiaohouKey{int(riYuan), int(monthBranch)}]
	if !ok {
		return TiaoHou{}, false
	}

	yongElem := ganzhi.GanWuxing(e.primary)
	var xiElem ganzhi.Wuxing
	if e.secondary != 0 {
		xiElem = ganzhi.GanWuxing(e.secondary)
	}

	jiElem, hasJi := pickJiElement(ganzhi.GanWuxing(riYuan), e.primary, e.secondary)

	season := ganzhi.ZhiSeasonLabel(monthBranch)

	detail := ganzhi.ZhiName(monthBranch) + "月" + ganzhi.GanName(riYuan) + ganzhi.GanWuxing(riYuan).String()
	detail += "，用" + ganzhi.GanName(e.primary) + "调候"
	if e.secondary != 0 {
		detail += "，" + ganzhi.GanName(e.secondary) + "辅之"
	}

	jiStr := ""
	if hasJi {
		jiStr = jiElem.String()
	}

	xiStr := ""
	if e.secondary != 0 {
		xiStr = xiElem.String()
	}

	return TiaoHou{
		Season: season,
		Yong:   yongElem.String(),
		Xi:     xiStr,
		Ji:     jiStr,
		Detail: detail,
	}, true
}

// pickJiElement returns the Ji (忌神) element for the TiaoHou result.
// Ji is the element that controls (克) the day master. If the controlling
// element collides with yong/xi at the WUXING level — and the drain fallback
// also collides — returns false to signal no clear 忌神.
func pickJiElement(dmElem ganzhi.Wuxing, yong, xi ganzhi.Gan) (ganzhi.Wuxing, bool) {
	ctrlElem := elementThatControls(dmElem)
	yongWx := ganzhi.GanWuxing(yong)
	xiWx := ganzhi.GanWuxing(xi)

	if ctrlElem != yongWx && ctrlElem != xiWx {
		return ctrlElem, true
	}

	drain := elementThatDrains(dmElem)
	if drain != yongWx && drain != xiWx {
		return drain, true
	}

	// Both ctrl and drain conflict with yong/xi — no clear 忌神.
	return 0, false
}