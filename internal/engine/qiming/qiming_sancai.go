package qiming

import (
	"liki-engine/internal/engine/ganzhi"
)

// SanCai holds the three-talent (三才) analysis.
type SanCai struct {
	Configuration string `json:"configuration"`
	Fortune       string `json:"fortune"`
	Description   string `json:"description"`
}

// computeSanCai returns the three-talent analysis.
func computeSanCai(tianElem, renElem, diElem string) SanCai {
	key := tianElem + renElem + diElem
	if v, ok := sanCaiCfg[key]; ok {
		return SanCai{Configuration: key, Fortune: v.Fortune, Description: v.Desc}
	}
	return SanCai{Configuration: key, Fortune: fortuneNeutral, Description: "三才配置中等，无大吉亦无大凶。"}
}

// strokeToWuxing maps the last digit of a wuge number to its wuxing element.
func strokeToWuxing(n int) ganzhi.Wuxing {
	switch n % 10 {
	case 1, 2:
		return ganzhi.WxMu
	case 3, 4:
		return ganzhi.WxHuo
	case 5, 6:
		return ganzhi.WxTu
	case 7, 8:
		return ganzhi.WxJin
	default:
		return ganzhi.WxShui
	}
}

// SancaiHarmonious returns true if the three cai elements are mutually generating.
func SancaiHarmonious(surnameStroke, s1, s2 int) bool {
	tian := surnameStroke + 1
	ren := surnameStroke + s1
	di := s1 + s2
	if s2 == 0 {
		di = s1 + 1
	}
	return ganzhi.Sheng(strokeToWuxing(tian), strokeToWuxing(ren)) &&
		ganzhi.Sheng(strokeToWuxing(ren), strokeToWuxing(di))
}

// FilterSancai filters stroke pairs to only those with harmonious sancai.
func FilterSancai(surnameStroke int, pairs []StrokePair) []StrokePair {
	if len(pairs) == 0 {
		return pairs
	}
	filtered := make([]StrokePair, 0, len(pairs))
	for _, p := range pairs {
		if SancaiHarmonious(surnameStroke, p.S1, p.S2) {
			filtered = append(filtered, p)
		}
	}
	return filtered
}
