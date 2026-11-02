package ziwei

type Bond struct {
	MingGong BondPalacePair `json:"ming_gong"`
	FuQiGong BondPalacePair `json:"fu_qi_gong"`
	ZiNvGong BondPalacePair `json:"zi_nv_gong"`
}

type BondPalacePair struct {
	A      BondPalaceFacts `json:"a" jsonschema:"甲方。"`
	B      BondPalaceFacts `json:"b" jsonschema:"乙方。"`
	Status string          `json:"status" jsonschema:"主星状态说明。"`
}

// ComputeBond returns side-by-side natal palace facts. It deliberately does
// not project one chart's stars or SiHua into the other chart: such cross-chart
// placement is not treated here as a stable classical Ziwei rule.
func ComputeBond(a, b BondSideFacts) Bond {
	return Bond{
		MingGong: bondPalacePair(a.MingGong, b.MingGong, "命宫"),
		FuQiGong: bondPalacePair(a.FuQiGong, b.FuQiGong, "夫妻"),
		ZiNvGong: bondPalacePair(a.ZiNvGong, b.ZiNvGong, "子女"),
	}
}

func bondPalacePair(aPalace, bPalace BondPalaceFacts, kind string) BondPalacePair {
	return BondPalacePair{
		A:      aPalace,
		B:      bPalace,
		Status: pairStatus(kind, aPalace.MajorStars, bPalace.MajorStars),
	}
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
