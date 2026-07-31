package qiming

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

// SancaiHarmonious returns true if the three-cai configuration is auspicious
// (大吉/吉), per the 125-combination table (same standard as qiming.check).
// 五行取自 strokeResult（>81 先按 ((n-1)%81)+1 回绕），与五格评估同一来源。
func SancaiHarmonious(surnameStroke, s1, s2 int) bool {
	tian := surnameStroke + 1
	ren := surnameStroke + s1
	di := s1 + s2
	if s2 == 0 {
		di = s1 + 1
	}
	key := strokeResult(tian).Element + strokeResult(ren).Element + strokeResult(di).Element
	if v, ok := sanCaiCfg[key]; ok {
		return v.Fortune == fortuneAuspicious || v.Fortune == fortuneGood
	}
	return false
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
