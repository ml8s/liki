package ziwei

type Bond struct {
	AMingGong string `json:"a_ming_gong"`
	BMingGong string `json:"b_ming_gong"`

	FuQiGong *PairRef `json:"fu_qi_gong,omitempty"`
	ZiNvGong *PairRef `json:"zi_nv_gong,omitempty"`
}

type PairRef struct {
	AGongName string   `json:"a_gong_name"`
	BGongName string   `json:"b_gong_name"`
	AZhuXing  []string `json:"a_zhu_xing"`
	BZhuXing  []string `json:"b_zhu_xing"`
	Status    string   `json:"zhu_xing_status"`
}

// ComputeBond returns side-by-side natal palace facts. It deliberately does
// not project one chart's stars or SiHua into the other chart: such cross-chart
// placement is not treated here as a stable classical Ziwei rule.
func ComputeBond(a, b ZiweiBondSideFacts) Bond {
	return Bond{
		AMingGong: a.MingGong.Name,
		BMingGong: b.MingGong.Name,
		FuQiGong:  gongWeiDuiZhao(a, b, "夫妻"),
		ZiNvGong:  gongWeiDuiZhao(a, b, "子女"),
	}
}

func gongWeiDuiZhao(a, b ZiweiBondSideFacts, kind string) *PairRef {
	var aPalace, bPalace ZiweiBondPalaceFacts
	switch kind {
	case "夫妻":
		aPalace, bPalace = a.FuQiGong, b.FuQiGong
	case "子女":
		aPalace, bPalace = a.ZiNvGong, b.ZiNvGong
	}
	return &PairRef{
		AGongName: aPalace.Name,
		BGongName: bPalace.Name,
		AZhuXing:  aPalace.MajorStars,
		BZhuXing:  bPalace.MajorStars,
		Status:    pairStatus(kind, aPalace.MajorStars, bPalace.MajorStars),
	}
}

func majorList(stars []starInfo) []string {
	r := []string{}
	for _, s := range stars {
		if s.IsMajor {
			r = append(r, s.Name)
		}
	}
	return r
}

func pairStatus(label string, aStars, bStars []string) string {
	switch label {
	case "夫妻":
		return spouseStatus(aStars, bStars)
	case "子女":
		return childStatus(aStars, bStars)
	}
	return ""
}

func spouseStatus(aStars, bStars []string) string {
	switch {
	case len(aStars) == 0 && len(bStars) == 0:
		return "双方夫妻宫皆无主星，按紫微通例借对宫主星观察"
	case len(bStars) == 0:
		return "A方夫妻宫有主星，B方夫妻宫无主星，按通例借对宫"
	case len(aStars) == 0:
		return "A方夫妻宫无主星，按通例借对宫；B方夫妻宫有主星"
	default:
		return "双方夫妻宫皆有主星"
	}
}

func childStatus(aStars, bStars []string) string {
	switch {
	case len(aStars) == 0 && len(bStars) == 0:
		return "双方子女宫皆无主星，按紫微通例借对宫主星观察"
	case len(bStars) == 0:
		return "A方子女宫有主星，B方子女宫无主星，按通例借对宫"
	case len(aStars) == 0:
		return "A方子女宫无主星，按通例借对宫；B方子女宫有主星"
	default:
		return "双方子女宫皆有主星"
	}
}
