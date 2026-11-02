package bazi

import (
	"testing"

	"liki-engine/internal/engine/ganzhi"
)

func TestChartExtra_SanYuan(t *testing.T) {
	// 甲子年 丙寅月 戊辰日 庚申时 → known chart
	ch := Chart{
		Nian: Pillar{Zhu: zh("甲", "子")},
		Yue:  Pillar{Zhu: zh("丙", "寅")},
		Ri:   Pillar{Zhu: zh("戊", "辰")},
		Shi:  Pillar{Zhu: zh("庚", "申")},
	}

	chart := ComputeChartFromCore(ch.NatalCore())

	// 胎元: 月柱丙寅 → 干+1=丁, 支+3=巳 → 丁巳
	if chart.SanYuan.TaiYuan.Gan.String() != "丁" {
		t.Errorf("taiyuan gan = %s, want 丁", chart.SanYuan.TaiYuan.Gan)
	}
	if chart.SanYuan.TaiYuan.Zhi.String() != "巳" {
		t.Errorf("taiyuan zhi = %s, want 巳", chart.SanYuan.TaiYuan.Zhi)
	}
	// 命宫 and 身宫 should be non-zero
	if chart.SanYuan.MingGong.Gan == 0 || chart.SanYuan.MingGong.Zhi == 0 {
		t.Error("minggong should not be zero")
	}
	if chart.SanYuan.ShenGong.Gan == 0 || chart.SanYuan.ShenGong.Zhi == 0 {
		t.Error("shengong should not be zero")
	}
}

func TestChartExtra_ChangSheng(t *testing.T) {
	ch := Chart{
		Nian: Pillar{Zhu: zh("甲", "子")},
		Yue:  Pillar{Zhu: zh("丙", "寅")},
		Ri:   Pillar{Zhu: zh("戊", "辰")},
		Shi:  Pillar{Zhu: zh("庚", "申")},
	}

	chart := ComputeChartFromCore(ch.NatalCore())

	if len(chart.ChangSheng) != 12 {
		t.Fatalf("chang_sheng len = %d, want 12", len(chart.ChangSheng))
	}
	if chart.ChangSheng[0].Name != "长生" {
		t.Errorf("stage 0 = %s, want 长生", chart.ChangSheng[0].Name)
	}
}

func TestChartExtra_NayinRel(t *testing.T) {
	ch := Chart{
		Nian: Pillar{Zhu: zh("甲", "子")},
		Yue:  Pillar{Zhu: zh("丙", "寅")},
		Ri:   Pillar{Zhu: zh("戊", "辰")},
		Shi:  Pillar{Zhu: zh("庚", "申")},
	}

	chart := ComputeChartFromCore(ch.NatalCore())

	if len(chart.NayinRel) == 0 {
		t.Error("nayin_rel should not be empty")
	}
	for _, nr := range chart.NayinRel {
		if nr.A == "" || nr.B == "" || nr.Relation == "" {
			t.Errorf("nayin rel entry has empty field: A=%q B=%q Rel=%q", nr.A, nr.B, nr.Relation)
		}
	}
}

func TestChartExtra_AllFieldsPresent(t *testing.T) {
	ch := Chart{
		Nian: Pillar{Zhu: zh("甲", "子")},
		Yue:  Pillar{Zhu: zh("丙", "寅")},
		Ri:   Pillar{Zhu: zh("戊", "辰")},
		Shi:  Pillar{Zhu: zh("庚", "申")},
	}

	chart := ComputeChartFromCore(ch.NatalCore())

	// SanYuan should have all three components
	if chart.SanYuan.TaiYuan.Gan == 0 {
		t.Error("tai_yuan missing")
	}
	if chart.SanYuan.MingGong.Gan == 0 {
		t.Error("ming_gong missing")
	}
	if chart.SanYuan.ShenGong.Gan == 0 {
		t.Error("shen_gong missing")
	}
	// ChangSheng should have 12 stages
	if len(chart.ChangSheng) != 12 {
		t.Errorf("chang_sheng len = %d", len(chart.ChangSheng))
	}
	// NayinRel should have pairs
	if len(chart.NayinRel) == 0 {
		t.Error("nayin_rel empty")
	}
	// GongJia and SanQiName may be empty for some charts, that's fine
}

//nolint:errcheck
func zh(g, z string) ganzhi.Zhu {
	gan, _ := ganzhi.ParseGan(g)
	zhi, _ := ganzhi.ParseZhi(z)
	return ganzhi.Zhu{Gan: gan, Zhi: zhi}
}
