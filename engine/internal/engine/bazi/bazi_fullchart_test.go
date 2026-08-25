package bazi

import (
	"encoding/json"
	"testing"
	"time"

	"liki-engine/internal/engine/ganzhi"
	"liki-engine/internal/engine/tianwen"
)

func TestChart_Lean_NoExtraFields(t *testing.T) {
	// 验证 bazi.chart（computeChartCore）返回的 zhuInfo 只含 Zhu + NaYin
	st := tianwen.GregorianToSolar(
		time.Date(1990, 6, 15, 10, 0, 0, 0, time.FixedZone("CST", 8*3600)),
		116.4, 8,
	)
	chart := ComputeChart(st, ganzhi.Male)

	got, err := json.Marshal(chart)
	if err != nil {
		t.Fatal(err)
	}
	// 反序列化检查是否有多余字段
	var raw map[string]any
	if err := json.Unmarshal(got, &raw); err != nil {
		t.Fatal(err)
	}
	for _, key := range []string{"nian", "yue", "ri", "shi"} {
		zhu, ok := raw[key].(map[string]any)
		if !ok {
			t.Fatalf("%s not an object", key)
		}
		// 应有 gan, zhi, na_yin
		for _, field := range []string{"gan", "zhi", "na_yin"} {
			if _, exists := zhu[field]; !exists {
				t.Errorf("%s.%s missing", key, field)
			}
		}
		// 不应有 cang_gan, shi_shens, shen_sha 等全量字段
		for _, field := range []string{"cang_gan", "shi_shens", "shen_sha", "chang_sheng", "is_void", "is_self_he", "is_kui_gang", "self_he_name"} {
			if _, exists := zhu[field]; exists {
				t.Errorf("%s.%s should not be in lean chart (got %v)", key, field, zhu[field])
			}
		}
	}
}

func TestComputeFullChart_HasAllFields(t *testing.T) {
	// 验证 fullchart 补全了所有字段
	st := tianwen.GregorianToSolar(
		time.Date(1990, 6, 15, 10, 0, 0, 0, time.FixedZone("CST", 8*3600)),
		116.4, 8,
	)
	chart := ComputeChart(st, ganzhi.Male)
	full := ComputeFullChart(chart)

	// 每柱应有全量字段
	for _, zhu := range []fullZhuInfo{full.Nian, full.Yue, full.Ri, full.Shi} {
		if zhu.NaYin == "" {
			t.Error("fullZhuInfo.na_yin is empty")
		}
		if len(zhu.ShiShens) == 0 {
			t.Error("fullZhuInfo.shi_shens is empty")
		}
		if zhu.CangGan.Main == 0 {
			t.Error("fullZhuInfo.cang_gan.main is empty")
		}
		if len(zhu.ChangSheng) == 0 {
			t.Error("fullZhuInfo.chang_sheng is empty")
		}
	}

	// fullchart 的 DaYun 应与 chart 一致
	if (full.DaYun == nil) != (chart.DaYun == nil) {
		t.Error("fullchart and chart DaYun mismatch")
	}
}

func TestComputeFullChart_SiHuaFromChart(t *testing.T) {
	// 验证 fullchart 的十神计算正确
	st := tianwen.GregorianToSolar(
		time.Date(1990, 6, 15, 10, 0, 0, 0, time.FixedZone("CST", 8*3600)),
		116.4, 8,
	)
	chart := ComputeChart(st, ganzhi.Male)
	full := ComputeFullChart(chart)

	// 日柱应有十神（对日主的关系）
	for _, entry := range full.Ri.ShiShens {
		if entry.Gan == full.Ri.Gan {
			// 日主自身应是比肩或劫财
			if entry.ShiShen.String() != "比肩" && entry.ShiShen.String() != "劫财" {
				t.Errorf("日柱自身十神=%s, 应为比肩或劫财", entry.Name)
			}
		}
	}
}

func TestComputeFullChart_ShenShaNotEmpty(t *testing.T) {
	// 验证 fullchart 有神煞
	st := tianwen.GregorianToSolar(
		time.Date(1990, 6, 15, 10, 0, 0, 0, time.FixedZone("CST", 8*3600)),
		116.4, 8,
	)
	chart := ComputeChart(st, ganzhi.Male)
	full := ComputeFullChart(chart)

	// 至少某一柱有神煞
	hasShenSha := false
	for _, zhu := range []fullZhuInfo{full.Nian, full.Yue, full.Ri, full.Shi} {
		if len(zhu.ShenSha) > 0 {
			hasShenSha = true
			break
		}
	}
	if !hasShenSha {
		t.Error("全柱无神煞")
	}
}

func TestComputeFullChart_InputIsLeanChart(t *testing.T) {
	// 验证 ComputeFullChart 接受 lean chart（不含藏干）
	// 即使 chart 是瘦的，fullchart 也能正确补全
	st := tianwen.GregorianToSolar(
		time.Date(1990, 6, 15, 10, 0, 0, 0, time.FixedZone("CST", 8*3600)),
		116.4, 8,
	)
	chart := ComputeChart(st, ganzhi.Male)

	// 验证 chart 是瘦的
	// chart is lean: can't access CangGan directly
	_ = chart.Nian.NaYin

	full := ComputeFullChart(chart)
	if full.Nian.CangGan.Main == 0 {
		t.Error("fullchart.cang_gan.main is still empty")
	}
}
