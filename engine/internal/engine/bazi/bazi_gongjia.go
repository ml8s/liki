package bazi

import (
	"liki-engine/internal/engine/ganzhi"
)

// GongJia describes a recognized 三合拱 / 三会拱 between two bazi pillars.
type GongJia struct {
	ZhuA    int        `json:"pillar_a"` // index 0-3 of first pillar
	ZhuB    int        `json:"pillar_b"` // index 0-3 of second pillar
	Type    string     `json:"type"`     // "三合拱" / "三会拱"
	Element string     `json:"wuxing"`   // 拱局五行
	Zhi     ganzhi.Zhi `json:"zhi"`      // the hidden zhi between them
}

// computeGongJia detects only the classical 八拱：三合局首墓二支拱旺支，
// 三会方首尾二支拱方中支。任意隔一支的“子寅拱丑”不冒充领域拱局。
func computeGongJia(bz ganzhi.Bazi) []GongJia {
	var results []GongJia
	classical := map[struct{ a, b, mid ganzhi.Zhi }]struct {
		kind    string
		element ganzhi.Wuxing
	}{}
	for _, group := range ganzhi.TripleHeList {
		key := struct{ a, b, mid ganzhi.Zhi }{group.Zhi[0], group.Zhi[2], group.Zhi[1]}
		classical[key] = struct {
			kind    string
			element ganzhi.Wuxing
		}{"三合拱", group.Element}
	}
	for _, group := range ganzhi.TripleHuiList {
		key := struct{ a, b, mid ganzhi.Zhi }{group.Zhi[0], group.Zhi[2], group.Zhi[1]}
		classical[key] = struct {
			kind    string
			element ganzhi.Wuxing
		}{"三会拱", group.Element}
	}
	bs := zhiSet(bz)
	for item, meta := range classical {
		if !bs[item.a] || !bs[item.b] || bs[item.mid] {
			continue
		}
		pA, pB := zhuIndexForZhi(bz, int(item.a)), zhuIndexForZhi(bz, int(item.b))
		if pA >= 0 && pB >= 0 {
			results = append(results, GongJia{
				ZhuA: pA, ZhuB: pB, Type: meta.kind,
				Element: meta.element.String(), Zhi: item.mid,
			})
		}
	}
	return results
}

func zhuIndexForZhi(bz ganzhi.Bazi, b int) int {
	zhus := bz.Slice()
	for i, p := range zhus {
		if int(p.Zhi) == b {
			return i
		}
	}
	return -1
}
