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
func ComputeBond(a, b Chart) Bond {
	return Bond{
		AMingGong: a.GongWei[int(a.MingGong)].Name,
		BMingGong: b.GongWei[int(b.MingGong)].Name,
		FuQiGong:  gongWeiDuiZhao(a, b, 2, "夫妻"),
		ZiNvGong:  gongWeiDuiZhao(a, b, 3, "子女"),
	}
}

func gongWeiDuiZhao(a, b Chart, idx int, label string) *PairRef {
	return &PairRef{
		AGongName: a.GongWei[idx].Name,
		BGongName: b.GongWei[idx].Name,
		AZhuXing:  majorList(a.GongWei[idx].Stars),
		BZhuXing:  majorList(b.GongWei[idx].Stars),
		Status:    pairStatus(label, a.GongWei[idx], b.GongWei[idx]),
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

func pairStatus(label string, a, b gong) string {
	switch label {
	case "夫妻":
		return spouseStatus(majorList(a.Stars), majorList(b.Stars))
	case "子女":
		return childStatus(majorList(a.Stars), majorList(b.Stars))
	}
	return ""
}

func spouseStatus(a, b []string) string {
	switch {
	case len(a) == 0 && len(b) == 0:
		return "双方夫妻宫皆无主星，按紫微通例借对宫主星观察"
	case len(b) == 0:
		return "A方夫妻宫有主星，B方夫妻宫无主星，按通例借对宫"
	case len(a) == 0:
		return "A方夫妻宫无主星，按通例借对宫；B方夫妻宫有主星"
	default:
		return "双方夫妻宫皆有主星"
	}
}

func childStatus(a, b []string) string {
	switch {
	case len(a) == 0 && len(b) == 0:
		return "双方子女宫皆无主星，按紫微通例借对宫主星观察"
	case len(b) == 0:
		return "A方子女宫有主星，B方子女宫无主星，按通例借对宫"
	case len(a) == 0:
		return "A方子女宫无主星，按通例借对宫；B方子女宫有主星"
	default:
		return "双方子女宫皆有主星"
	}
}
