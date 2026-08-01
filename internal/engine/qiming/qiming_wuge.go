package qiming

const (
	fortuneAuspicious    = "大吉"
	fortuneGood          = "吉"
	fortuneNeutral       = "半吉"
	fortuneInauspicious  = "凶"
)

// ge is one of the five-grid (五格) numbers.
type ge struct {
	Stroke      int    `json:"stroke"`
	Element     string `json:"wuxing"`
	Fortune     string `json:"ji_xiong"`
	Description string `json:"description"`
}

// WuGe holds the five-grid (五格) analysis.
type WuGe struct {
	TianGe ge `json:"tiange"`
	RenGe  ge `json:"renge"`
	DiGe   ge `json:"dige"`
	WaiGe  ge `json:"waige"`
	ZongGe ge `json:"zongge"`
}

// StrokePair is a viable (s1, s2) stroke combination.
type StrokePair struct {
	S1 int `json:"s1"`
	S2 int `json:"s2"`
}

// ListViableStrokes returns stroke pairs (s1, s2) where 人/地/外/总 are all auspicious.
// count=2 for double-name (s2 range 1..81), count=1 for single-name (s2=0).
func ListViableStrokes(surnameStroke, count int) []StrokePair {
	var pairs []StrokePair

	if count == 1 {
		for s1 := 1; s1 <= 81; s1++ {
			wg := computeWuGeFromStrokes(surnameStroke, s1, 0)
			if allWuGeAuspicious(wg) {
				pairs = append(pairs, StrokePair{S1: s1, S2: 0})
			}
		}
		return pairs
	}

	for s1 := 1; s1 <= 81; s1++ {
		for s2 := 1; s2 <= 81; s2++ {
			wg := computeWuGeFromStrokes(surnameStroke, s1, s2)
			if allWuGeAuspicious(wg) {
				pairs = append(pairs, StrokePair{S1: s1, S2: s2})
			}
		}
	}
	return pairs
}

func allWuGeAuspicious(wg WuGe) bool {
	return isAuspicious(wg.RenGe.Fortune) &&
		isAuspicious(wg.DiGe.Fortune) &&
		isAuspicious(wg.WaiGe.Fortune) &&
		isAuspicious(wg.ZongGe.Fortune)
}

func isAuspicious(fortune string) bool {
	return fortune == fortuneAuspicious || fortune == fortuneGood
}

func strokeResult(stroke int) ge {
	if stroke < 1 {
		stroke = 1
	}
	if stroke > 81 {
		stroke = ((stroke - 1) % 81) + 1
	}
	if v, ok := sanCaiNums[stroke]; ok {
		return ge{Stroke: stroke, Element: v.Element, Fortune: v.Fortune, Description: v.Desc}
	}
	return ge{Stroke: stroke, Element: "土", Fortune: fortuneNeutral, Description: ""}
}

func fortuneYAMLToChinese(f string) string {
	switch f {
	case "ji":
		return fortuneGood
	case "da_ji":
		return fortuneAuspicious
	case "ban_ji":
		return fortuneNeutral
	case "xiong":
		return fortuneInauspicious
	}
	return f
}
