package bazi

import (
	"testing"
	"time"

	"liki-engine/internal/engine/ganzhi"
	"liki-engine/internal/engine/tianwen"
)

// TestDaYun_LongitudeIndependent：起运只取决于出生瞬时到节气瞬时的距离，
// 不应随经度变化（经度只影响日柱/时柱的读数）。此前 computeDaYunSteps 用
// 真太阳时壁钟与节气瞬时比较，属跨帧，起运日期会随经度漂移。
func TestDaYun_LongitudeIndependent(t *testing.T) {
	cst := time.FixedZone("CST", 8*3600)
	when := time.Date(1990, 5, 20, 12, 0, 0, 0, cst)
	gt := time.Time(when)

	var ref *DaYun
	for _, lon := range []float64{75.0, 116.4, 135.0} {
		st := tianwen.ApparentSolarFromCivil(when, lon, 8)
		chart := ComputeChart(gt, st, ganzhi.Male)
		if chart.DaYun == nil {
			t.Fatalf("lon=%v: da_yun is nil", lon)
		}
		if ref == nil {
			ref = chart.DaYun
			continue
		}
		if chart.DaYun.StartDate != ref.StartDate ||
			chart.DaYun.StartYearAfter != ref.StartYearAfter ||
			chart.DaYun.StartMonthAfter != ref.StartMonthAfter ||
			chart.DaYun.StartDayAfter != ref.StartDayAfter {
			t.Errorf("lon=%v 起运=%s (%d年%d月%d日), want %s (%d年%d月%d日)（不应随经度变化）",
				lon, chart.DaYun.StartDate, chart.DaYun.StartYearAfter, chart.DaYun.StartMonthAfter, chart.DaYun.StartDayAfter,
				ref.StartDate, ref.StartYearAfter, ref.StartMonthAfter, ref.StartDayAfter)
		}
		if len(chart.DaYun.Steps) != len(ref.Steps) {
			t.Fatalf("lon=%v steps=%d, want %d", lon, len(chart.DaYun.Steps), len(ref.Steps))
		}
		for i := range ref.Steps {
			if chart.DaYun.Steps[i].StartDate != ref.Steps[i].StartDate ||
				chart.DaYun.Steps[i].Gan != ref.Steps[i].Gan ||
				chart.DaYun.Steps[i].Zhi != ref.Steps[i].Zhi {
				t.Errorf("lon=%v step %d=%s%s@%s, want %s%s@%s",
					lon, i, chart.DaYun.Steps[i].Gan, chart.DaYun.Steps[i].Zhi, chart.DaYun.Steps[i].StartDate,
					ref.Steps[i].Gan, ref.Steps[i].Zhi, ref.Steps[i].StartDate)
			}
		}
	}
}

// TestDaYun_TermBoundaryUsesCivilInstant：立春前后的同一时辰窗口，
// 四柱年/月柱与起运都应按 Gregorian time判定（与经度无关）。
func TestDaYun_TermBoundaryUsesCivilInstant(t *testing.T) {
	cst := time.FixedZone("CST", 8*3600)
	after := time.Date(2024, 2, 4, 17, 0, 0, 0, cst) // 立春（≈16:26）后
	for _, lon := range []float64{75.0, 135.0} {
		gt := time.Time(after)
		st := tianwen.ApparentSolarFromCivil(after, lon, 8)
		chart := ComputeChart(gt, st, ganzhi.Male)
		if chart.Nian.Gan.String() != "甲" || chart.Nian.Zhi.String() != "辰" {
			t.Errorf("lon=%v 年柱=%s%s, want 甲辰", lon, chart.Nian.Gan, chart.Nian.Zhi)
		}
		if chart.Yue.Zhi.String() != "寅" {
			t.Errorf("lon=%v 月柱地支=%s, want 寅", lon, chart.Yue.Zhi)
		}
		if chart.DaYun == nil {
			t.Fatalf("lon=%v: da_yun is nil", lon)
		}
	}
}
