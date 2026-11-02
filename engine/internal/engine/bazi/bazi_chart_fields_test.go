package bazi

import (
	"testing"
	"time"

	"liki-engine/internal/engine/ganzhi"
	"liki-engine/internal/engine/tianwen"
)

func TestComputeChart_HasAllFields(t *testing.T) {
	// 验证领域 Chart 直接承载所有柱级字段
	stgt := time.Time(time.Date(1990, 6, 15, 10, 0, 0, 0, time.FixedZone("CST", 8*3600)))
	st := tianwen.ApparentSolarFromCivil(
		time.Date(1990, 6, 15, 10, 0, 0, 0, time.FixedZone("CST", 8*3600)),
		116.4, 8,
	)
	chart := ComputeChart(stgt, st, ganzhi.Male)

	// 每柱应有全量字段
	for _, zhu := range []Pillar{chart.Nian, chart.Yue, chart.Ri, chart.Shi} {
		if zhu.NaYin == "" {
			t.Error("Pillar.na_yin is empty")
		}
		if len(zhu.ShiShens) == 0 {
			t.Error("Pillar.shi_shens is empty")
		}
		if zhu.CangGan.Main == 0 {
			t.Error("Pillar.cang_gan.main is empty")
		}
		if zhu.SelfSitting == "" {
			t.Error("Pillar.self_sitting is empty")
		}
		if zhu.DayMasterTrend == "" {
			t.Error("Pillar.day_master_trend is empty")
		}
		if zhu.Xun == "" || zhu.XunKong == "" {
			t.Errorf("Pillar xun/xun_kong is empty: %q/%q", zhu.Xun, zhu.XunKong)
		}
	}

}

func TestComputeChart_SiHuaFromChart(t *testing.T) {
	// 验证 Chart 的十神计算正确
	stgt := time.Time(time.Date(1990, 6, 15, 10, 0, 0, 0, time.FixedZone("CST", 8*3600)))
	st := tianwen.ApparentSolarFromCivil(
		time.Date(1990, 6, 15, 10, 0, 0, 0, time.FixedZone("CST", 8*3600)),
		116.4, 8,
	)
	chart := ComputeChart(stgt, st, ganzhi.Male)

	// 日柱应有十神（对日主的关系）
	for _, entry := range chart.Ri.ShiShens {
		if entry.Gan == chart.Ri.Gan {
			// 日主自身应是比肩或劫财
			if entry.ShiShen.String() != "比肩" && entry.ShiShen.String() != "劫财" {
				t.Errorf("日柱自身十神=%s, 应为比肩或劫财", entry.Name)
			}
		}
	}
}

func TestComputeChart_ShenShaNotEmpty(t *testing.T) {
	// 验证 Chart 有神煞
	stgt := time.Time(time.Date(1990, 6, 15, 10, 0, 0, 0, time.FixedZone("CST", 8*3600)))
	st := tianwen.ApparentSolarFromCivil(
		time.Date(1990, 6, 15, 10, 0, 0, 0, time.FixedZone("CST", 8*3600)),
		116.4, 8,
	)
	chart := ComputeChart(stgt, st, ganzhi.Male)

	// 至少某一柱有神煞
	hasShenSha := false
	for _, zhu := range []Pillar{chart.Nian, chart.Yue, chart.Ri, chart.Shi} {
		if len(zhu.ShenSha) > 0 {
			hasShenSha = true
			break
		}
	}
	if !hasShenSha {
		t.Error("全柱无神煞")
	}
}
