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
		Ri: ganzhi.Zhu{Gan: ganzhi.GanJia, Zhi: ganzhi.ZhiZi}, // 甲日子
	}
	zhus := computeXiaoYun(bz, ganzhi.Male, 3)
	if len(zhus) != 3 {
		t.Fatalf("len = %d, want 3", len(zhus))
	}
	// 男起丙寅
	if zhus[0].Gan != ganzhi.GanBing || zhus[0].Zhi != ganzhi.ZhiYin {
		t.Errorf("首岁 = %s%s, want 丙寅", zhus[0].Gan, zhus[0].Zhi)
	}
	// 顺行：丙寅→丁卯→戊辰
	want := []string{"丙寅", "丁卯", "戊辰"}
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
		Ri: ganzhi.Zhu{Gan: ganzhi.GanJia, Zhi: ganzhi.ZhiZi},
	}
	zhus := computeXiaoYun(bz, ganzhi.Female, 3)
	// 女起壬申
	if zhus[0].Gan != ganzhi.GanRen || zhus[0].Zhi != ganzhi.ZhiShen {
		t.Errorf("首岁 = %s%s, want 壬申", zhus[0].Gan, zhus[0].Zhi)
	}
	// 逆行：壬申→辛未→庚午
	want := []string{"壬申", "辛未", "庚午"}
	for i, w := range want {
		got := zhus[i].Gan.String() + zhus[i].Zhi.String()
		if got != w {
			t.Errorf("岁%d = %s, want %s", i+1, got, w)
		}
	}
}

func TestXiaoYun_ShiShen(t *testing.T) {
	// 甲日主，丙寅岁 → 丙=食神
	bz := ganzhi.Bazi{
		Ri: ganzhi.Zhu{Gan: ganzhi.GanJia, Zhi: ganzhi.ZhiZi},
	}
	zhus := computeXiaoYun(bz, ganzhi.Male, 1)
	if zhus[0].ShiShen != "食神" {
		t.Errorf("甲日 丙岁 十神 = %s, want 食神", zhus[0].ShiShen)
	}
	// 甲日主，壬申岁（女）→ 壬=偏印
	zhus2 := computeXiaoYun(bz, ganzhi.Female, 1)
	if zhus2[0].ShiShen != "偏印" {
		t.Errorf("甲日 壬岁 十神 = %s, want 偏印", zhus2[0].ShiShen)
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
