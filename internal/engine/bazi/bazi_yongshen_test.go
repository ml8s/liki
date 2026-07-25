package bazi

import (
	"testing"
	"time"

	"liki-engine/internal/engine/ganzhi"
	"liki-engine/internal/engine/tianwen"
)

func TestComputeYongShen_AllSchoolsPresent(t *testing.T) {
	st := tianwen.GregorianToSolar(
		time.Date(1984, 2, 15, 8, 0, 0, 0, time.FixedZone("CST", 8*3600)),
		116.4, 8,
	)
	chart := ComputeChart(st, ganzhi.Male)
	result := ComputeYongShen(chart)

	validStrengths := map[string]bool{"身强": true, "身弱": true, "中和": true}
	if !validStrengths[result.FuYi.Strength] {
		t.Errorf("FuYi.Strength = %q", result.FuYi.Strength)
	}
	// wuxing_count and wang_shuai should be present
	if len(result.FuYi.WuxingCount) == 0 {
		t.Error("FuYi.WuxingCount is empty")
	}
	if len(result.FuYi.WangShuai) != 5 {
		t.Errorf("FuYi.WangShuai has %d elements, want 5", len(result.FuYi.WangShuai))
	}
	if result.FuYi.Strength != "中和" {
		for _, field := range []struct{ name, val string }{
			{"FuYi.Yong", result.FuYi.Yong},
			{"FuYi.Xi", result.FuYi.Xi},
			{"FuYi.Ji", result.FuYi.Ji},
		} {
			if field.val == "" {
				t.Errorf("%s is empty", field.name)
			}
		}
	}
	for _, field := range []struct{ name, val string }{
		{"TiaoHou.Yong", result.TiaoHou.Yong},
		{"TiaoHou.Xi", result.TiaoHou.Xi},
		{"TiaoHou.Ji", result.TiaoHou.Ji},
		{"TiaoHou.Season", result.TiaoHou.Season},
		{"GeJu.Yong", result.GeJu.Yong},
		{"GeJu.Xi", result.GeJu.Xi},
		{"GeJu.Ji", result.GeJu.Ji},
		{"GeJu.Pattern", result.GeJu.Pattern},
		{"GeJu.Usage", result.GeJu.Usage},
	} {
		if field.val == "" {
			t.Errorf("%s is empty", field.name)
		}
	}

	validUsages := map[string]bool{"顺用": true, "逆用": true}
	if !validUsages[result.GeJu.Usage] {
		t.Errorf("GeJu.Usage = %q, want 顺用 or 逆用", result.GeJu.Usage)
	}
}

func TestComputeGeJuYongShen_ShunYong(t *testing.T) {
	cb := Chart{
		Ri: zhuInfo{
			Zhu: ganzhi.Zhu{Gan: ganzhi.GanJia, Zhi: ganzhi.ZhiYin},
		},
		Yue: zhuInfo{
			Zhu: ganzhi.Zhu{Gan: ganzhi.GanXin, Zhi: ganzhi.ZhiYou},
			},
		Nian: zhuInfo{Zhu: ganzhi.Zhu{Gan: ganzhi.GanGui, Zhi: ganzhi.ZhiSi}},
		Shi:  zhuInfo{Zhu: ganzhi.Zhu{Gan: ganzhi.GanRen, Zhi: ganzhi.ZhiShen}},
	}
	wc := map[ganzhi.Wuxing]int{ganzhi.WxMu: 1, ganzhi.WxJin: 2}
	result := computeGeJu(cb, wc)
	if result.Usage != "顺用" {
		t.Errorf("expected 顺用, got %s", result.Usage)
	}
	if result.Pattern != "正官格" {
		t.Errorf("expected 正官格, got %s", result.Pattern)
	}
}

func TestComputeGeJuYongShen_NiYong(t *testing.T) {
	cb := Chart{
		Ri: zhuInfo{
			Zhu: ganzhi.Zhu{Gan: ganzhi.GanJia, Zhi: ganzhi.ZhiYin},
		},
		Yue: zhuInfo{
			Zhu: ganzhi.Zhu{Gan: ganzhi.GanGeng, Zhi: ganzhi.ZhiShen},
			},
		Nian: zhuInfo{Zhu: ganzhi.Zhu{Gan: ganzhi.GanGui, Zhi: ganzhi.ZhiSi}},
		Shi:  zhuInfo{Zhu: ganzhi.Zhu{Gan: ganzhi.GanRen, Zhi: ganzhi.ZhiShen}},
	}
	wc := map[ganzhi.Wuxing]int{ganzhi.WxMu: 1, ganzhi.WxJin: 3, ganzhi.WxShui: 1}
	result := computeGeJu(cb, wc)
	if result.Usage != "逆用" {
		t.Errorf("expected 逆用, got %s", result.Usage)
	}
	if result.Pattern != "七杀格" {
		t.Errorf("expected 七杀格, got %s", result.Pattern)
	}
}
