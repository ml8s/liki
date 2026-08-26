package bazi

import (
	"sort"

	"liki-engine/internal/engine/ganzhi"
)

// GongJia describes a 拱 between two bazi pillars.
type GongJia struct {
	ZhuA int        `json:"pillar_a"` // index 0-3 of first pillar
	ZhuB int        `json:"pillar_b"` // index 0-3 of second pillar
	Type string     `json:"type"`     // "拱"
	Zhi  ganzhi.Zhi `json:"zhi"`      // the hidden branch between them
}

// computeGongJia detects 拱 (gap=2) between branches of bazi pillars.
// When two pillar branches differ by 2 (mod 12), the midpoint = 拱.
// Adjacent branches (gap=1) have no hidden branch and are skipped.
func computeGongJia(bz ganzhi.Bazi) []GongJia {
	zhus := bz.Slice()
	bs := make([]int, 0, 4)
	seen := [13]bool{}
	for _, p := range zhus {
		b := int(p.Zhi)
		if b >= 1 && b <= 12 && !seen[b] {
			seen[b] = true
			bs = append(bs, b)
		}
	}
	sort.Ints(bs)

	var results []GongJia

	for i := 0; i < len(bs); i++ {
		for j := i + 1; j < len(bs); j++ {
			a, bb := bs[i], bs[j]
			forward := (bb - a + 12) % 12
			backward := (a - bb + 12) % 12
			gap := forward
			if backward < forward {
				gap = backward
			}

			if gap != 2 {
				continue
			}

			midB := a%12 + 1
			if backward < forward {
				midB = (a+10)%12 + 1
			}
			if midB > 12 {
				midB = 1
			}
			pA, pB := zhuIndexForBranch(bz, a), zhuIndexForBranch(bz, bb)
			if pA >= 0 && pB >= 0 {
				results = append(results, GongJia{
					ZhuA: pA,
					ZhuB: pB,
					Type: "拱",
					Zhi:  ganzhi.Zhi(midB),
				})
			}
		}
	}

	return results
}

func zhuIndexForBranch(bz ganzhi.Bazi, b int) int {
	zhus := bz.Slice()
	for i, p := range zhus {
		if int(p.Zhi) == b {
			return i
		}
	}
	return -1
}
