package bazi

import (
	"liki-engine/internal/engine/ganzhi"
	"testing"
)

func TestBanHe_19810826(t *testing.T) {
	bz := ganzhi.Bazi{
		Nian: ganzhi.Zhu{Gan: ganzhi.GanXin, Zhi: ganzhi.ZhiYou},
		Yue:  ganzhi.Zhu{Gan: ganzhi.GanBing, Zhi: ganzhi.ZhiShen},
		Ri:   ganzhi.Zhu{Gan: ganzhi.GanBing, Zhi: ganzhi.ZhiZi},
		Shi:  ganzhi.Zhu{Gan: ganzhi.GanJi, Zhi: ganzhi.ZhiChou},
	}
	chart := canonicalChart(bz, ganzhi.Male, 1981)
	hehui := ComputeHeHui(chart)

	// Oracle detects: 酉丑半合金局 (year+hour), 申子半合水局 (month+day)
	if len(hehui.SanHePartial) != 2 {
		t.Fatalf("san_he_partial count = %d, want 2 (酉丑半合金 + 申子半合水)", len(hehui.SanHePartial))
	}

	foundYouChou := false
	foundShenZi := false
	for _, ph := range hehui.SanHePartial {
		branches := ph.Branches
		if len(branches) == 2 {
			if (branches[0] == "酉" && branches[1] == "丑") || (branches[0] == "丑" && branches[1] == "酉") {
				foundYouChou = true
				if ph.Element != "金" {
					t.Errorf("酉丑半合 element = %s, want 金", ph.Element)
				}
			}
			if (branches[0] == "申" && branches[1] == "子") || (branches[0] == "子" && branches[1] == "申") {
				foundShenZi = true
				if ph.Element != "水" {
					t.Errorf("申子半合 element = %s, want 水", ph.Element)
				}
			}
		}
	}
	if !foundYouChou {
		t.Error("酉丑半合金 not detected")
	}
	if !foundShenZi {
		t.Error("申子半合水 not detected")
	}
	t.Logf("san_he_partial: %d entries", len(hehui.SanHePartial))
	for _, ph := range hehui.SanHePartial {
		t.Logf("  %s [%s] element=%s pillars=%v", ph.Name, ph.Branches, ph.Element, ph.Pillars)
	}
}

func TestBanHe_RequiresMiddleBranch(t *testing.T) {
	// 巳丑 missing 旺支酉；子辰 missing 旺支子。二者是拱合候选，不是半合。
	bz := ganzhi.Bazi{
		Nian: ganzhi.Zhu{Gan: ganzhi.GanJia, Zhi: ganzhi.ZhiChen},
		Yue:  ganzhi.Zhu{Gan: ganzhi.GanBing, Zhi: ganzhi.ZhiChou},
		Ri:   ganzhi.Zhu{Gan: ganzhi.GanWu, Zhi: ganzhi.ZhiSi},
		Shi:  ganzhi.Zhu{Gan: ganzhi.GanGeng, Zhi: ganzhi.ZhiWei},
	}
	hehui := ComputeHeHui(canonicalChart(bz, ganzhi.Male, 2000))
	for _, partial := range hehui.SanHePartial {
		if equalStrings(partial.Branches, []string{"巳", "丑"}) || equalStrings(partial.Branches, []string{"子", "辰"}) {
			t.Fatalf("missing-middle pair %v must not be 半合; got %#v", partial.Branches, hehui.SanHePartial)
		}
	}
}
