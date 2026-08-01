package bazi

import "liki-engine/internal/engine/ganzhi"

// computeGeJu determines yong/xi/ji based on pattern type (顺用/逆用).
// Pattern is determined by 月令支藏干透干法 (子平正法):
//   - 建禄/月刃 → 逆用 (prioritized)
//   - 本气透干 → 用该干十神定格局
//   - 中气透干 → 用该干十神定格局
//   - 余气透干 → 用该干十神定格局
//   - 都不透 → 月令本气虚格
func computeGeJu(c Chart, wc map[ganzhi.Wuxing]int) GeJuResult {
	bz := c.ToBazi()
	hs := computeCangGan(bz)
	shiShens := computeShiShensTable(bz, hs)
	riYuan := c.Ri.Gan
	dmElem := ganzhi.GanWuxing(riYuan)
	yueZhi := c.Yue.Zhi

	// 建禄格/月刃格: month branch is the day master's 临官(禄) or 帝旺(刃).
	if isLu, isRen := jianLuYueRenBranch(riYuan, yueZhi); isLu || isRen {
		return computeJianLuYueRen(dmElem, isRen)
	}

	// 月令透干定格局: 本气→中气→余气, 第一个透干者定格.
	var patternGan ganzhi.Gan
	var patternShiShen ganzhi.ShiShen
	found := false

	// 本气 → 中气 → 余气 遍历
	for _, source := range []string{sourceMainQi, sourceMidQi, sourceMinQi} {
		for _, ss := range shiShens[1] {
			if ss.Source != source {
				continue
			}
			if ss.Gan == c.Nian.Gan || ss.Gan == c.Yue.Gan ||
				ss.Gan == c.Ri.Gan || ss.Gan == c.Shi.Gan {
				patternGan = ss.Gan
				patternShiShen = ss.ShiShen
				found = true
				break
			}
		}
		if found {
			break
		}
	}

	// 不透干 → 月令本气虚格
	if !found {
		for _, ss := range shiShens[1] {
			if ss.Source == sourceMainQi {
				patternGan = ss.Gan
				patternShiShen = ss.ShiShen
				break
			}
		}
	}

	patternElem := ganzhi.GanWuxing(patternGan)
	patternName := shiShenToPatternName(patternShiShen)

	var yong, xi, ji ganzhi.Wuxing
	var yongFa string

	switch patternShiShen {
	case ganzhi.ShiShenZhengGuan, ganzhi.ShiShenZhengCai,
		ganzhi.ShiShenPianCai, ganzhi.ShiShenZhengYin,
		ganzhi.ShiShenPianYin, ganzhi.ShiShenShiShen:
		yongFa = "顺用"
		yong = elementThatGenerates(patternElem)
		ji = elementThatControls(patternElem)
		xi = elementThatControls(ji) // 制忌神者为喜神

	case ganzhi.ShiShenQiSha, ganzhi.ShiShenShangGuan:
		yongFa = "逆用"
		yong = elementThatControls(patternElem)
		xi = elementThatGenerates(yong)
		ji = elementThatGenerates(patternElem)

	default:
		yongFa = "逆用"
		patternName = "杂格"
		yong = elementThatControls(dmElem)
		xi = elementThatGenerates(yong)
		ji = elementThatGenerates(dmElem)
	}

	return GeJuResult{
		Yong:    yong.String(),
		Xi:      xi.String(),
		Ji:      ji.String(),
		Pattern: patternName,
		Usage:   yongFa,
	}
}

func computeJianLuYueRen(dmElem ganzhi.Wuxing, isYueRen bool) GeJuResult {
	var patternName string
	if isYueRen {
		patternName = "月刃格"
	} else {
		patternName = "建禄格"
	}
	yong := elementThatControls(dmElem)
	xi := elementThatGenerates(yong)
	ji := elementThatGenerates(dmElem)
	return GeJuResult{
		Yong:    yong.String(),
		Xi:      xi.String(),
		Ji:      ji.String(),
		Pattern: patternName,
		Usage:   "逆用",
	}
}

// jianLuYueRenBranch checks if a branch is the day master's 临官(禄) or 帝旺(刃).
func jianLuYueRenBranch(riGan ganzhi.Gan, yueZhi ganzhi.Zhi) (isLu, isYueRen bool) {
	stages := ganzhi.ChangShengTable[riGan]
	if len(stages) < 5 {
		return false, false
	}
	lu := stages[3]  // 临官 = 禄
	ren := stages[4] // 帝旺 = 刃

	switch yueZhi {
	case lu:
		return true, false
	case ren:
		return false, true
	}
	return false, false
}

// shiShenToPatternName converts a shishen to its pattern name.
func shiShenToPatternName(ss ganzhi.ShiShen) string {
	switch ss {
	case ganzhi.ShiShenZhengGuan:
		return "正官格"
	case ganzhi.ShiShenQiSha:
		return "七杀格"
	case ganzhi.ShiShenZhengCai:
		return "正财格"
	case ganzhi.ShiShenPianCai:
		return "偏财格"
	case ganzhi.ShiShenZhengYin:
		return "正印格"
	case ganzhi.ShiShenPianYin:
		return "偏印格"
	case ganzhi.ShiShenShiShen:
		return "食神格"
	case ganzhi.ShiShenShangGuan:
		return "伤官格"
	default:
		return "杂格"
	}
}
