package liuyao

import "liki-engine/internal/engine/ganzhi"

// tombOf maps an element to its conventional tomb branch. Earth is a known
// school difference; 辰 is the mechanical default and interpretation may state
// the chosen school explicitly.
func tombOf(element ganzhi.Wuxing) ganzhi.Zhi {
	switch element {
	case ganzhi.WxMu:
		return ganzhi.ZhiWei
	case ganzhi.WxHuo:
		return ganzhi.ZhiXu
	case ganzhi.WxJin:
		return ganzhi.ZhiChou
	case ganzhi.WxShui:
		return ganzhi.ZhiChen
	case ganzhi.WxTu:
		return ganzhi.ZhiChen
	default:
		return 0
	}
}

// yangStemOfElement chooses the yang stem representing an element for the
// twelve-life-cycle table: 木甲、火丙、土戊、金庚、水壬。
func yangStemOfElement(element ganzhi.Wuxing) ganzhi.Gan {
	switch element {
	case ganzhi.WxMu:
		return ganzhi.GanJia
	case ganzhi.WxHuo:
		return ganzhi.GanBing
	case ganzhi.WxTu:
		return ganzhi.GanWu
	case ganzhi.WxJin:
		return ganzhi.GanGeng
	case ganzhi.WxShui:
		return ganzhi.GanRen
	default:
		return 0
	}
}

// lifeStageOf returns the twelve-life-cycle stage of element at branch.
func lifeStageOf(element ganzhi.Wuxing, branch ganzhi.Zhi) string {
	stem := yangStemOfElement(element)
	stages, ok := ganzhi.ChangShengTable[stem]
	if !ok {
		return ""
	}
	for i, zhi := range stages {
		if zhi == branch && i < len(ganzhi.StageNamesZH) {
			return ganzhi.StageNamesZH[i]
		}
	}
	return ""
}
