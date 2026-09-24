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
	zhus := computeXiaoYun(bz, ganzhi.Male, 3)
	if len(zhus) != 3 {
		t.Fatalf("len = %d, want 3", len(zhus))
	}
	// 阳男顺行，由时柱甲子起
	if zhus[0].Gan != ganzhi.GanJia || zhus[0].Zhi != ganzhi.ZhiZi {
		t.Errorf("首岁 = %s%s, want 甲子", zhus[0].Gan, zhus[0].Zhi)
	}
	// 顺行：甲子→乙丑→丙寅
	want := []string{"甲子", "乙丑", "丙寅"}
	for i, w := range want {
		got := zhus[i].Gan.String() + zhus[i].Zhi.String()
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
	zhus := computeXiaoYun(bz, ganzhi.Female, 3)
	// 阳女逆行，由时柱甲子起
	if zhus[0].Gan != ganzhi.GanJia || zhus[0].Zhi != ganzhi.ZhiZi {
		t.Errorf("首岁 = %s%s, want 甲子", zhus[0].Gan, zhus[0].Zhi)
	}
	// 逆行：甲子→癸亥→壬戌
	want := []string{"甲子", "癸亥", "壬戌"}
	for i, w := range want {
		got := zhus[i].Gan.String() + zhus[i].Zhi.String()
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
	zhus := computeXiaoYun(bz, ganzhi.Male, 1)
	if zhus[0].ShiShen != "食神" {
		t.Errorf("甲日 丙岁 十神 = %s, want 食神", zhus[0].ShiShen)
	}
	// 甲日主，时柱丙寅起（阳女逆，首岁仍丙寅）→ 丙=食神
	zhus2 := computeXiaoYun(bz, ganzhi.Female, 1)
	if zhus2[0].ShiShen != "食神" {
		t.Errorf("甲日 丙岁 十神 = %s, want 食神", zhus2[0].ShiShen)
	}
}

func TestXiaoYun_MaxAgeDefault(t *testing.T) {
	bz := ganzhi.Bazi{Ri: ganzhi.Zhu{Gan: ganzhi.GanJia, Zhi: ganzhi.ZhiZi}}
	// maxAge<=0 → 默认 12
	zhus := computeXiaoYun(bz, ganzhi.Male, 0)
	if len(zhus) != 12 {
		t.Errorf("默认 len = %d, want 12", len(zhus))
	}
}
