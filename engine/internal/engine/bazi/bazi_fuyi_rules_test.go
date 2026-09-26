package bazi

import (
	"testing"

	"liki-engine/internal/engine/ganzhi"
)

// 扶抑喜忌（computeNormalYongJi）固化测试。
// 入表前锁定当前命理输出：身强用克我者（官杀）、财生官杀为喜、印生我为忌；
// 身弱用生我者（印）、比劫帮身为喜、官杀克我为忌；中和不扶抑。
// 《子平真诠》扶抑用神。

func TestNormalYongJi_StrongUsesController(t *testing.T) {
	// 身强：用=克我者（官杀），喜=生克我者（财），忌=生我者（印）
	cases := []struct {
		gan      ganzhi.Gan
		wantYong string
		wantXi   string
		wantJi   string
	}{
		{ganzhi.GanYi, "金", "土", "水"},   // 乙木：克者金、财土生金、印水生木
		{ganzhi.GanBing, "水", "金", "木"}, // 丙火：克者水、财金生水、印木生火
		{ganzhi.GanXin, "火", "木", "土"},  // 辛金：克者火（官杀）、财木生火、印土生金
	}
	for _, tc := range cases {
		yong, xi, ji := computeNormalYongJi(tc.gan, "身强")
		if yong != tc.wantYong || xi != tc.wantXi || ji != tc.wantJi {
			t.Errorf("干 %d 身强: yong=%s xi=%s ji=%s, want %s/%s/%s",
				tc.gan, yong, xi, ji, tc.wantYong, tc.wantXi, tc.wantJi)
		}
	}
}

func TestNormalYongJi_WeakUsesGenerator(t *testing.T) {
	// 身弱：用=生我者（印），喜=同我者（比劫），忌=克我者（官杀）
	cases := []struct {
		gan      ganzhi.Gan
		wantYong string
		wantXi   string
		wantJi   string
	}{
		{ganzhi.GanYi, "水", "木", "金"},   // 乙木：印水生木、比劫木帮、官杀金克
		{ganzhi.GanBing, "木", "火", "水"}, // 丙火：印木生火、比劫火帮、官杀水克
		{ganzhi.GanXin, "土", "金", "火"},  // 辛金：印土生金、比劫金帮、官杀火克
	}
	for _, tc := range cases {
		yong, xi, ji := computeNormalYongJi(tc.gan, "身弱")
		if yong != tc.wantYong || xi != tc.wantXi || ji != tc.wantJi {
			t.Errorf("干 %d 身弱: yong=%s xi=%s ji=%s, want %s/%s/%s",
				tc.gan, yong, xi, ji, tc.wantYong, tc.wantXi, tc.wantJi)
		}
	}
}

func TestNormalYongJi_NeutralNoAdjust(t *testing.T) {
	yong, xi, ji := computeNormalYongJi(ganzhi.GanYi, "中和")
	if yong != "" || xi != "" || ji != "" {
		t.Errorf("中和不应扶抑: yong=%s xi=%s ji=%s", yong, xi, ji)
	}
}
