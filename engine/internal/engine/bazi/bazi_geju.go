package bazi

import "liki-engine/internal/engine/ganzhi"

// computeGeJu determines the month-command pattern candidate. It does not
// derive yong/xi/ji: pattern success, rescue, and final favorable elements
// require whole-chart review.
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
	yueZhi := c.Yue.Zhi

	// 建禄格/月刃格: month zhi is the day master's 临官(禄) or 帝旺(刃).
	if isLu, isRen := jianLuYueRenZhi(riYuan, yueZhi); isLu || isRen {
		return computeJianLuYueRen(c, isRen)
	}

	// 月令透干定格局: 本气→中气→余气, 第一个透干者定格.
	var patternGan ganzhi.Gan
	var patternShiShen ganzhi.ShiShen
	found := false
	patternSource := ""

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
				patternSource = source
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
				patternSource = sourceMainQi + "_not_transparent"
				break
			}
		}
	}

	patternElem := ganzhi.GanWuxing(patternGan)
	patternName := shiShenToPatternName(patternShiShen)

	var yongFa string

	// 顺逆用入表 geju_rules.json：吉格顺用（官财印食）、凶格逆用（杀伤）。
	if p, ok := geJuRules[patternShiShen.String()]; ok {
		yongFa = p.Usage
	} else {
		yongFa = "逆用"
		patternName = "杂格"
	}

	return GeJuResult{
		Pattern:          patternName,
		Usage:            yongFa,
		PatternGod:       ganzhi.GanName(patternGan),
		PatternGodTenGod: patternShiShen.String(),
		PatternGodSource: patternSource,
		Structure:        buildGeJuStructure(c, &patternGan, patternElem),
	}
}

func computeJianLuYueRen(c Chart, isYueRen bool) GeJuResult {
	dmElem := ganzhi.GanWuxing(c.Ri.Gan)
	var patternName string
	if isYueRen {
		patternName = "月刃格"
	} else {
		patternName = "建禄格"
	}
	return GeJuResult{
		Pattern:          patternName,
		Usage:            "逆用",
		PatternGodSource: map[bool]string{true: "month_blade", false: "month_lu"}[isYueRen],
		Structure:        buildGeJuStructure(c, nil, dmElem),
	}
}

// jianLuYueRenZhi checks if a zhi is the day master's 临官(禄) or 帝旺(刃).
func jianLuYueRenZhi(riGan ganzhi.Gan, yueZhi ganzhi.Zhi) (isLu, isYueRen bool) {
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
		// 《三命通会·论阳刃》：五阳干有刃，五阴干无刃。
		// 阴干帝旺位仍可作为十二长生事实，但不得命名为月刃格。
		if int(riGan)%2 != 1 {
			return false, false
		}
		return false, true
	}
	return false, false
}

// shiShenToPatternName converts a shishen to its pattern name.
// 格局格名入表 geju_rules.json（《子平真诠》官杀财印食伤各成一格）。
func shiShenToPatternName(ss ganzhi.ShiShen) string {
	if p, ok := geJuRules[ss.String()]; ok {
		return p.Name
	}
	return "杂格"
}
