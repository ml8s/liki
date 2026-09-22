package bazi

import (
	"testing"

	"liki-engine/internal/engine/ganzhi"
)

func TestComputeChartExtra_SanYuan(t *testing.T) {
	// 甲子年 丙寅月 戊辰日 庚申时 → known chart
	ch := Chart{
		Nian: zhuInfo{Zhu: zh("甲", "子")},
		Yue:  zhuInfo{Zhu: zh("丙", "寅")},
		Ri:   zhuInfo{Zhu: zh("戊", "辰")},
		Shi:  zhuInfo{Zhu: zh("庚", "申")},
	}

	extra := ComputeChartExtra(ch)

	// 胎元: 月柱丙寅 → 干+1=丁, 支+3=巳 → 丁巳
	if extra.SanYuan.TaiYuan.Gan.String() != "丁" {
		t.Errorf("taiyuan gan = %s, want 丁", extra.SanYuan.TaiYuan.Gan)
	}
	if extra.SanYuan.TaiYuan.Zhi.String() != "巳" {
		t.Errorf("taiyuan zhi = %s, want 巳", extra.SanYuan.TaiYuan.Zhi)
	}
	// 命宫 and 身宫 should be non-zero
	if extra.SanYuan.MingGong.Gan == 0 || extra.SanYuan.MingGong.Zhi == 0 {
		t.Error("minggong should not be zero")
	}
	if extra.SanYuan.ShenGong.Gan == 0 || extra.SanYuan.ShenGong.Zhi == 0 {
		t.Error("shengong should not be zero")
	}
}

func TestComputeChartExtra_ChangSheng(t *testing.T) {
	ch := Chart{
		Nian: zhuInfo{Zhu: zh("甲", "子")},
		Yue:  zhuInfo{Zhu: zh("丙", "寅")},
		Ri:   zhuInfo{Zhu: zh("戊", "辰")},
		Shi:  zhuInfo{Zhu: zh("庚", "申")},
	}

	extra := ComputeChartExtra(ch)

	if len(extra.ChangSheng) != 12 {
		t.Fatalf("chang_sheng len = %d, want 12", len(extra.ChangSheng))
	}
	if extra.ChangSheng[0].Name != "长生" {
		t.Errorf("stage 0 = %s, want 长生", extra.ChangSheng[0].Name)
	}
}

func TestComputeChartExtra_NayinRel(t *testing.T) {
	ch := Chart{
		Nian: zhuInfo{Zhu: zh("甲", "子")},
		Yue:  zhuInfo{Zhu: zh("丙", "寅")},
		Ri:   zhuInfo{Zhu: zh("戊", "辰")},
		Shi:  zhuInfo{Zhu: zh("庚", "申")},
	}

	extra := ComputeChartExtra(ch)

	if len(extra.NayinRel) == 0 {
		t.Error("nayin_rel should not be empty")
	}
	for _, nr := range extra.NayinRel {
		if nr.A == "" || nr.B == "" || nr.Relation == "" {
			t.Errorf("nayin rel entry has empty field: A=%q B=%q Rel=%q", nr.A, nr.B, nr.Relation)
		}
	}
}

func TestComputeChartExtra_AllFieldsPresent(t *testing.T) {
	ch := Chart{
		Nian: zhuInfo{Zhu: zh("甲", "子")},
		Yue:  zhuInfo{Zhu: zh("丙", "寅")},
		Ri:   zhuInfo{Zhu: zh("戊", "辰")},
		Shi:  zhuInfo{Zhu: zh("庚", "申")},
	}

	extra := ComputeChartExtra(ch)

	// SanYuan should have all three components
	if extra.SanYuan.TaiYuan.Gan == 0 {
		t.Error("tai_yuan missing")
	}
	if extra.SanYuan.MingGong.Gan == 0 {
		t.Error("ming_gong missing")
	}
	if extra.SanYuan.ShenGong.Gan == 0 {
		t.Error("shen_gong missing")
	}
	// ChangSheng should have 12 stages
	if len(extra.ChangSheng) != 12 {
		t.Errorf("chang_sheng len = %d", len(extra.ChangSheng))
	}
	// NayinRel should have pairs
	if len(extra.NayinRel) == 0 {
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
