package bazi

import (
	"strings"

	"liki-engine/internal/engine/ganzhi"
)

// HeHuiResult holds the complete 合会冲刑 analysis for a bazi chart.
type HeHuiResult struct {
	GanHe    []GanHePair   `json:"gan_he"`
	ZhiLiuHe []ZhiPairRel  `json:"zhi_liu_he"`
	SanHe    []TripleGroup `json:"san_he"`
	SanHui   []TripleGroup `json:"san_hui"`
	LiuChong []ZhiPairRel  `json:"liu_chong"`
	LiuHai   []ZhiPairRel  `json:"liu_hai"`
	LiuXing  []ZhiPairRel  `json:"liu_xing"`
}

// GanHePair describes a 天干五合 pair. Adjacency is an engine-owned fact;
// non-adjacent pairs remain candidates and must not be promoted to tight combinations.
type GanHePair struct {
	GanA      string `json:"gan_a"`
	GanB      string `json:"gan_b"`
	PillarA   int    `json:"pillar_a"`
	PillarB   int    `json:"pillar_b"`
	Position  string `json:"position"`
	Contested bool   `json:"contested"`
	HeElement string `json:"he_element"`
}

// ZhiPairRel describes a paired zhi relationship (六合/六冲/六害/相刑) between two pillars.
type ZhiPairRel struct {
	ZhiA    string `json:"zhi_a"`
	ZhiB    string `json:"zhi_b"`
	PillarA int    `json:"pillar_a"`
	PillarB int    `json:"pillar_b"`
	Element string `json:"wuxing,omitempty"`
}

// TripleGroup describes a complete 三合局 or 三会方.
type TripleGroup struct {
	Type     string   `json:"type"`
	Name     string   `json:"name"`
	Element  string   `json:"wuxing"`
	Branches []string `json:"branches"`
	Pillars  []string `json:"pillars"`
}

// ComputeHeHui computes the full 合会冲刑 analysis from a Chart.
func ComputeHeHui(c Chart) HeHuiResult {
	bz := c.ToBazi()
	return HeHuiResult{
		GanHe:    detectGanHe(bz),
		ZhiLiuHe: detectZhiPairs(bz, ganzhi.IsZhiHe, true),
		SanHe:    detectTriple(bz, ganzhi.TripleHeList, relSanHe, "局"),
		SanHui:   detectTriple(bz, ganzhi.TripleHuiList, relSanHui, "方"),
		LiuChong: detectZhiPairs(bz, ganzhi.IsLiuChong, false),
		LiuHai:   detectZhiPairs(bz, ganzhi.IsHai, false),
		LiuXing:  detectZhiPairs(bz, ganzhi.IsXing, false),
	}
}

const (
	zhuNian = 0
	zhuYue  = 1
	zhuRi   = 2
	zhuShi  = 3
)

func detectGanHe(bz ganzhi.Bazi) []GanHePair {
	pairs := make([]GanHePair, 0, 5)
	zhus := bz.Slice()
	for i := 0; i < 3; i++ {
		for j := i + 1; j < 4; j++ {
			a, b := zhus[i].Gan, zhus[j].Gan
			if !ganzhi.IsGanHe(a, b) {
				continue
			}
			position := "separated"
			switch j - i {
			case 1:
				position = "adjacent"
			case 3:
				position = "remote"
			}
			pairs = append(pairs, GanHePair{
				GanA: ganzhi.GanName(a), GanB: ganzhi.GanName(b),
				PillarA: i, PillarB: j, Position: position,
				HeElement: ganHeResult(a, b).String(),
			})
		}
	}
	for i := range pairs {
		participations := 0
		for _, other := range pairs {
			if other.PillarA == pairs[i].PillarA || other.PillarA == pairs[i].PillarB ||
				other.PillarB == pairs[i].PillarA || other.PillarB == pairs[i].PillarB {
				participations++
			}
		}
		pairs[i].Contested = participations > 1
	}
	return pairs
}

func ganHeResult(a, b ganzhi.Gan) ganzhi.Wuxing {
	for _, p := range ganzhi.GanHes {
		if (a == p.A && b == p.B) || (a == p.B && b == p.A) {
			return p.Result
		}
	}
	return 0
}

// detectZhiPairs checks all pillar zhi pairs using check. When withElem is true,
// also computes the 六合 element produced by the pair.
func detectZhiPairs(bz ganzhi.Bazi, check func(ganzhi.Zhi, ganzhi.Zhi) bool, withElem bool) []ZhiPairRel {
	zhus := bz.Slice()
	rels := make([]ZhiPairRel, 0, 12)
	for i := 0; i < 3; i++ {
		for j := i + 1; j < 4; j++ {
			za, zb := zhus[i].Zhi, zhus[j].Zhi
			if check(za, zb) {
				var elem string
				if withElem {
					elem = zhiHeElement(za, zb)
				}
				rels = append(rels, ZhiPairRel{
					ZhiA:    ganzhi.ZhiName(za),
					ZhiB:    ganzhi.ZhiName(zb),
					PillarA: i,
					PillarB: j,
					Element: elem,
				})
			}
		}
	}
	return rels
}

func detectTriple(bz ganzhi.Bazi, list []ganzhi.SanHeHui, typ, suffix string) []TripleGroup {
	bs := zhiSet(bz)
	results := make([]TripleGroup, 0, 4)
	for _, tr := range list {
		if countZhi(bs, tr.Zhi...) == len(tr.Zhi) {
			results = append(results, TripleGroup{
				Type:     typ,
				Name:     tripleName(tr.Zhi, tr.Element, suffix),
				Element:  tr.Element.String(),
				Branches: zhiNames(tr.Zhi),
				Pillars:  pillarsWithAnyZhi(bz, tr.Zhi...),
			})
		}
	}
	return results
}

func zhiNames(items []ganzhi.Zhi) []string {
	names := make([]string, 0, len(items))
	for _, item := range items {
		names = append(names, ganzhi.ZhiName(item))
	}
	return names
}

func pillarsWithAnyZhi(bz ganzhi.Bazi, targets ...ganzhi.Zhi) []string {
	pillars := make([]string, 0, 4)
	for index, pillar := range bz.Slice() {
		for _, target := range targets {
			if pillar.Zhi == target {
				pillars = append(pillars, zhuLabels[index])
				break
			}
		}
	}
	return pillars
}

func zhiHeElement(a, b ganzhi.Zhi) string {
	for _, p := range ganzhi.ZhiHes {
		if (a == p.A && b == p.B) || (a == p.B && b == p.A) {
			return p.Result.String()
		}
	}
	return ""
}

func zhiSet(bz ganzhi.Bazi) [13]bool {
	zhus := bz.Slice()
	var bs [13]bool
	for _, p := range zhus {
		if b := int(p.Zhi); b >= 1 && b <= 12 {
			bs[b] = true
		}
	}
	return bs
}

func countZhi(bs [13]bool, targets ...ganzhi.Zhi) int {
	c := 0
	for _, t := range targets {
		if t >= 1 && t <= 12 && bs[int(t)] {
			c++
		}
	}
	return c
}

func tripleName(zhi []ganzhi.Zhi, element ganzhi.Wuxing, suffix string) string {
	parts := make([]string, len(zhi))
	for i, b := range zhi {
		parts[i] = ganzhi.ZhiName(b)
	}
	return strings.Join(parts, "") + element.String() + suffix
}

func containsPair(list []ganzhi.Zhi, a, b ganzhi.Zhi) bool {
	hasA, hasB := false, false
	for _, v := range list {
		if v == a {
			hasA = true
		}
		if v == b {
			hasB = true
		}
	}
	return hasA && hasB
}
