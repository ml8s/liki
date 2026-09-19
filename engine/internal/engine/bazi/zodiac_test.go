package bazi

import (
	"liki-engine/internal/engine/ganzhi"
	"testing"
)

func TestZodiac(t *testing.T) {
	// 1981 = 辛酉年 → 酉 → 鸡
	c := canonicalChart(ganzhi.Bazi{
		Nian: ganzhi.Zhu{Gan: ganzhi.GanXin, Zhi: ganzhi.ZhiYou},
		Yue:  ganzhi.Zhu{Gan: ganzhi.GanBing, Zhi: ganzhi.ZhiShen},
		Ri:   ganzhi.Zhu{Gan: ganzhi.GanBing, Zhi: ganzhi.ZhiZi},
		Shi:  ganzhi.Zhu{Gan: ganzhi.GanJi, Zhi: ganzhi.ZhiChou},
	}, ganzhi.Male, 1981)
	full := ComputeFullChart(c)
	if full.Zodiac != "鸡" {
		t.Errorf("zodiac = %q, want 鸡 (辛酉年)", full.Zodiac)
	}
	t.Logf("zodiac = %s (辛酉年) ✓", full.Zodiac)
}
