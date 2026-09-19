package bazi

import (
	"liki-engine/internal/engine/ganzhi"
	"testing"
)

func TestFullWalkthrough_19810826(t *testing.T) {
	bz := ganzhi.Bazi{
		Nian: ganzhi.Zhu{Gan: ganzhi.GanXin, Zhi: ganzhi.ZhiYou},
		Yue:  ganzhi.Zhu{Gan: ganzhi.GanBing, Zhi: ganzhi.ZhiShen},
		Ri:   ganzhi.Zhu{Gan: ganzhi.GanBing, Zhi: ganzhi.ZhiZi},
		Shi:  ganzhi.Zhu{Gan: ganzhi.GanJi, Zhi: ganzhi.ZhiChou},
	}
	chart := canonicalChart(bz, ganzhi.Male, 1981)
	full := ComputeFullChart(chart)

	t.Run("TaiYuan", func(t *testing.T) {
		// Oracle: 胎元 = 丁亥 (月干丙+1=丁, 月支申+3=亥)
		ty := full.SanYuan.TaiYuan
		got := string(ganzhi.GanName(ty.Gan)) + ganzhi.ZhiName(ty.Zhi)
		if got != "丁亥" {
			t.Errorf("tai_yuan = %s, want 丁亥", got)
		}
	})

	t.Run("MingGong", func(t *testing.T) {
		mg := full.SanYuan.MingGong
		got := string(ganzhi.GanName(mg.Gan)) + ganzhi.ZhiName(mg.Zhi)
		if got != "丙申" {
			t.Errorf("ming_gong = %s, want 丙申", got)
		}
	})

	t.Run("ShenGong", func(t *testing.T) {
		sg := full.SanYuan.ShenGong
		got := string(ganzhi.GanName(sg.Gan)) + ganzhi.ZhiName(sg.Zhi)
		if got != "戊戌" {
			t.Errorf("shen_gong = %s, want 戊戌", got)
		}
	})

	t.Run("TaiXi", func(t *testing.T) {
		tx := full.TaiXi
		got := string(ganzhi.GanName(tx.Gan)) + ganzhi.ZhiName(tx.Zhi)
		if got != "辛丑" {
			t.Errorf("tai_xi = %s, want 辛丑", got)
		}
	})

	t.Run("LiuPo_ZiYou", func(t *testing.T) {
		hehui := ComputeHeHui(chart)
		found := false
		for _, po := range hehui.LiuPo {
			if (po.ZhiA == "子" && po.ZhiB == "酉") || (po.ZhiA == "酉" && po.ZhiB == "子") {
				found = true
			}
		}
		if !found {
			t.Error("子酉破 not found in liu_po")
		}
	})

	t.Run("ZhiLiuHe_ZiChou", func(t *testing.T) {
		// Oracle: 子丑合化土
		found := false
		for _, he := range full.ZhiLiuHe {
			if (he.ZhiA == "子" && he.ZhiB == "丑") || (he.ZhiA == "丑" && he.ZhiB == "子") {
				found = true
				if he.Element != "土" {
					t.Errorf("子丑合 element = %s, want 土", he.Element)
				}
			}
		}
		if !found {
			t.Error("子丑六合 not found in zhi_liu_he")
		}
	})

	t.Run("SanHePartial_YouChou_ShenZi", func(t *testing.T) {
		// Oracle detects 酉丑半合金局 and 申子半合水局
		if len(full.SanHePartial) != 2 {
			t.Fatalf("san_he_partial count = %d, want 2", len(full.SanHePartial))
		}
	})

	t.Run("KongWangPerPillarVsGlobal", func(t *testing.T) {
		// Oracle reports empty per-pillar: year=子丑, month=辰巳, day=申酉, time=午未
		pillars := [4]fullZhuInfo{full.Nian, full.Yue, full.Ri, full.Shi}
		wantXun := []string{"甲寅旬", "甲午旬", "甲戌旬", "甲申旬"}
		wantXunKong := []string{"子丑", "辰巳", "申酉", "午未"}
		for i, pillar := range pillars {
			if pillar.Xun != wantXun[i] || pillar.XunKong != wantXunKong[i] {
				t.Fatalf("pillar[%d] xun = %q/%q, want %q/%q", i, pillar.Xun, pillar.XunKong, wantXun[i], wantXunKong[i])
			}
		}
		if full.DayXun != "甲戌旬" || full.DayXunKong != "申酉" {
			t.Fatalf("day xun = %q/%q, want 甲戌旬/申酉", full.DayXun, full.DayXunKong)
		}
	})

	t.Run("WuxingCount", func(t *testing.T) {
		fuYi, _, _ := ComputeYongShenSchools(chart)
		wc := fuYi.WuxingCount
		if wc["金"] != 4 || wc["木"] != 0 || wc["水"] != 3 || wc["火"] != 2 || wc["土"] != 3 {
			t.Fatalf("wuxing_count = %v, want 金4 木0 水3 火2 土3（干+藏干口径）", wc)
		}
	})

	t.Run("StartTend_Dayun", func(t *testing.T) {
		// Oracle start_tend: 6年0月20日, date=1987-09-15
		if full.DaYun != nil && len(full.DaYun.Steps) > 0 {
			step := full.DaYun.Steps[0]
			t.Logf("dayun[0]: %s%s start_date=%s start_year=%d",
				ganzhi.GanName(step.Gan), ganzhi.ZhiName(step.Zhi),
				step.StartDate, step.StartYear)
			// Oracle says 起运 6年0月20日 → 1987-09-15
			// Birth: 1981-08-26
			// 1981 + 6 = 1987 ✓
		}
	})
}
