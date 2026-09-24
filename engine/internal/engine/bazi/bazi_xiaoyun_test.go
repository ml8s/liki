package bazi

import (
	"testing"

	"liki-engine/internal/engine/ganzhi"
)

// 小运（XiaoYun）经典规则：
// 男：起丙寅顺行；女：起壬申逆行。
// 十神按日主对岁运天干。
func TestXiaoYun_MaleStartAndDirection(t *testing.T) {
	bz := ganzhi.Bazi{
		Ri:   ganzhi.Zhu{Gan: ganzhi.GanJia, Zhi: ganzhi.ZhiZi}, // 甲日子
		Shi:  ganzhi.Zhu{Gan: ganzhi.GanJia, Zhi: ganzhi.ZhiZi}, // 甲子时
		Nian: ganzhi.Zhu{Gan: ganzhi.GanJia, Zhi: ganzhi.ZhiZi}, // 甲年（阳）
	}
	sets := computeXiaoYun(bz, ganzhi.Male, 3)
	if len(sets) != 2 {
		t.Fatalf("流派数 = %d, want 2", len(sets))
	}
	// 流派一（三命通会）：男起丙寅顺行
	zhus := sets[0].Zhus
	if zhus[0].Gan != ganzhi.GanBing || zhus[0].Zhi != ganzhi.ZhiYin {
		t.Errorf("首岁 = %s%s, want 丙寅", zhus[0].Gan, zhus[0].Zhi)
	}
	want := []string{"丙寅", "丁卯", "戊辰"}
	for i, w := range want {
		got := zhus[i].Gan.String() + zhus[i].Zhi.String()
		if got != w {
			t.Errorf("岁%d = %s, want %s", i+1, got, w)
		}
	}
	// 流派二（星平会海）：阳男顺行，由时柱甲子起
	zhus2 := sets[1].Zhus
	if zhus2[0].Gan != ganzhi.GanJia || zhus2[0].Zhi != ganzhi.ZhiZi {
		t.Errorf("首岁 = %s%s, want 甲子", zhus2[0].Gan, zhus2[0].Zhi)
	}
	want2 := []string{"甲子", "乙丑", "丙寅"}
	for i, w := range want2 {
		got := zhus2[i].Gan.String() + zhus2[i].Zhi.String()
		if got != w {
			t.Errorf("岁%d = %s, want %s", i+1, got, w)
		}
	}
	// 年龄从 1 起
	if zhus[0].Age != 1 {
		t.Errorf("首岁 Age = %d, want 1", zhus[0].Age)
	}
}

func TestXiaoYun_FemaleStartAndDirection(t *testing.T) {
	bz := ganzhi.Bazi{
		Ri:   ganzhi.Zhu{Gan: ganzhi.GanJia, Zhi: ganzhi.ZhiZi},
		Shi:  ganzhi.Zhu{Gan: ganzhi.GanJia, Zhi: ganzhi.ZhiZi},
		Nian: ganzhi.Zhu{Gan: ganzhi.GanJia, Zhi: ganzhi.ZhiZi},
	}
	sets := computeXiaoYun(bz, ganzhi.Female, 3)
	// 流派一（三命通会）：女起壬申逆行
	zhus := sets[0].Zhus
	if zhus[0].Gan != ganzhi.GanRen || zhus[0].Zhi != ganzhi.ZhiShen {
		t.Errorf("首岁 = %s%s, want 壬申", zhus[0].Gan, zhus[0].Zhi)
	}
	want := []string{"壬申", "辛未", "庚午"}
	for i, w := range want {
		got := zhus[i].Gan.String() + zhus[i].Zhi.String()
		if got != w {
			t.Errorf("岁%d = %s, want %s", i+1, got, w)
		}
	}
	// 流派二（星平会海）：阳女逆行，由时柱甲子起
	zhus2 := sets[1].Zhus
	if zhus2[0].Gan != ganzhi.GanJia || zhus2[0].Zhi != ganzhi.ZhiZi {
		t.Errorf("首岁 = %s%s, want 甲子", zhus2[0].Gan, zhus2[0].Zhi)
	}
	want2 := []string{"甲子", "癸亥", "壬戌"}
	for i, w := range want2 {
		got := zhus2[i].Gan.String() + zhus2[i].Zhi.String()
		if got != w {
			t.Errorf("岁%d = %s, want %s", i+1, got, w)
		}
	}
}

func TestXiaoYun_ShiShen(t *testing.T) {
	// 甲日主，时柱丙寅起（阳男顺）→ 丙=食神
	bz := ganzhi.Bazi{
		Ri:   ganzhi.Zhu{Gan: ganzhi.GanJia, Zhi: ganzhi.ZhiZi},
		Shi:  ganzhi.Zhu{Gan: ganzhi.GanBing, Zhi: ganzhi.ZhiYin},
		Nian: ganzhi.Zhu{Gan: ganzhi.GanJia, Zhi: ganzhi.ZhiZi},
	}
	sets := computeXiaoYun(bz, ganzhi.Male, 1)
	if sets[0].Zhus[0].ShiShen != "食神" {
		t.Errorf("甲日 丙岁 十神 = %s, want 食神", sets[0].Zhus[0].ShiShen)
	}
	// 流派二（星平会海）：时柱丙寅起（阳男顺，首岁仍丙寅）→ 丙=食神
	if sets[1].Zhus[0].ShiShen != "食神" {
		t.Errorf("甲日 丙岁 十神 = %s, want 食神", sets[1].Zhus[0].ShiShen)
	}
}

func TestXiaoYun_MaxAgeDefault(t *testing.T) {
	bz := ganzhi.Bazi{Ri: ganzhi.Zhu{Gan: ganzhi.GanJia, Zhi: ganzhi.ZhiZi}}
	// maxAge<=0 → 默认 12
	sets := computeXiaoYun(bz, ganzhi.Male, 0)
	if len(sets) != 2 {
		t.Errorf("流派数 = %d, want 2", len(sets))
	}
	if len(sets[0].Zhus) != 12 {
		t.Errorf("三命通会段默认岁数 = %d, want 12", len(sets[0].Zhus))
	}
}

// TestXiaoYun_ShiZhuVariation：12 候选时辰——星平会海流派首岁随时辰变化，三命通会固定。
func TestXiaoYun_ShiZhuVariation(t *testing.T) {
	shichen := []struct{ gan, zhi ganzhi.Gan; ganzhi.Zhi }{}
	_ = shichen
	// 12 个候选时辰（子时→亥时，地支变化）
	candidates := []struct {
		gan ganzhi.Gan
		zhi ganzhi.Zhi
	}{
		{ganzhi.GanJia, ganzhi.ZhiZi}, {ganzhi.GanYi, ganzhi.ZhiChou},
		{ganzhi.GanBing, ganzhi.ZhiYin}, {ganzhi.GanDing, ganzhi.ZhiMao},
		{ganzhi.GanWu, ganzhi.ZhiChen}, {ganzhi.GanJi, ganzhi.ZhiSi},
		{ganzhi.GanGeng, ganzhi.ZhiWu}, {ganzhi.GanXin, ganzhi.ZhiWei},
		{ganzhi.GanRen, ganzhi.ZhiShen}, {ganzhi.GanGui, ganzhi.ZhiYou},
		{ganzhi.GanJia, ganzhi.ZhiXu}, {ganzhi.GanYi, ganzhi.ZhiHai},
	}
	firstXingPing := map[string]bool{}
	firstSanMing := map[string]bool{}
	for _, cand := range candidates {
		bz := ganzhi.Bazi{
			Ri:   ganzhi.Zhu{Gan: ganzhi.GanJia, Zhi: ganzhi.ZhiZi},
			Shi:  ganzhi.Zhu{Gan: cand.gan, Zhi: cand.zhi},
			Nian: ganzhi.Zhu{Gan: ganzhi.GanJia, Zhi: ganzhi.ZhiZi},
		}
		sets := computeXiaoYun(bz, ganzhi.Male, 1)
		sanMing := sets[0].Zhus[0].Gan.String() + sets[0].Zhus[0].Zhi.String()
		xingPing := sets[1].Zhus[0].Gan.String() + sets[1].Zhus[0].Zhi.String()
		firstSanMing[sanMing] = true
		firstXingPing[xingPing] = true
	}
	if len(firstSanMing) != 1 {
		t.Errorf("三命通会流派首岁应固定（男起丙寅），实际变化: %v", firstSanMing)
	}
	if len(firstXingPing) != len(candidates) {
		t.Errorf("星平会海流派首岁应随时柱变化（12 时辰各不同），实际 %d 种: %v", len(firstXingPing), firstXingPing)
	}
}
