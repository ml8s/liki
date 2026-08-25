package bazi

import (
	"testing"
	"time"

	"liki-engine/internal/engine/ganzhi"
	"liki-engine/internal/engine/tianwen"
)

type strengthRef struct {
	name                    string
	year, month, day, hour  int
	gender                  ganzhi.Gender
	longitude               float64
	timezone                float64
	expectStrength          string
}

func TestFuYi_Strength_Reference(t *testing.T) {
	tests := []strengthRef{
		{name: "己土寅月有根印比足→中和", year: 1984, month: 2, day: 15, hour: 8,
			gender: ganzhi.Male, longitude: 116.4, timezone: 8, expectStrength: "中和"},
		{name: "庚金辰月印比重→身强", year: 1980, month: 4, day: 15, hour: 12,
			gender: ganzhi.Male, longitude: 116.4, timezone: 8, expectStrength: "身强"},
		{name: "丙火申月无根仅月干丙→身弱", year: 1981, month: 8, day: 26, hour: 0,
			gender: ganzhi.Male, longitude: 130.7, timezone: 8, expectStrength: "身弱"},
		{name: "壬水午月无根→身弱", year: 1990, month: 6, day: 15, hour: 12,
			gender: ganzhi.Male, longitude: 116.4, timezone: 8, expectStrength: "身弱"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			loc := time.FixedZone("CST", int(tt.timezone)*3600)
			birth := time.Date(tt.year, time.Month(tt.month), tt.day, tt.hour, 0, 0, 0, loc)
			st := tianwen.GregorianToSolar(birth, tt.longitude, tt.timezone)
			chart := ComputeChart(st, tt.gender)
			result := ComputeYongShen(chart)
			if result.FuYi.Strength != tt.expectStrength {
				t.Errorf("Strength=%q, want=%q (%s%s %s%s %s%s %s%s)",
					result.FuYi.Strength, tt.expectStrength,
					chart.Nian.Gan, chart.Nian.Zhi,
					chart.Yue.Gan, chart.Yue.Zhi,
					chart.Ri.Gan, chart.Ri.Zhi,
					chart.Shi.Gan, chart.Shi.Zhi)
			}
		})
	}
}

func TestRegress_YinBi_ExcludesSelf(t *testing.T) {
	loc := time.FixedZone("CST", 8*3600)
	birth := time.Date(1981, 8, 26, 0, 15, 0, 0, loc)
	st := tianwen.GregorianToSolar(birth, 130.7, 8)
	chart := ComputeChart(st, ganzhi.Male)
	yinBi := countYinBi(chart)
	if yinBi >= 2 {
		t.Errorf("印比=%d(含日主)! 应=1, 日主不计入帮身", yinBi)
	}
}

func TestRegress_YinBi_HandCounted(t *testing.T) {
	tests := []struct {
		name      string
		chart     Chart
		wantYinBi int
	}{
		{name: "年印+月比劫+时食伤=2",
			chart: Chart{
				Ri: zhuInfo{Zhu: ganzhi.Zhu{Gan: ganzhi.GanJia}}, Nian: zhuInfo{Zhu: ganzhi.Zhu{Gan: ganzhi.GanGui}},
				Yue: zhuInfo{Zhu: ganzhi.Zhu{Gan: ganzhi.GanJia}}, Shi: zhuInfo{Zhu: ganzhi.Zhu{Gan: ganzhi.GanBing}}},
			wantYinBi: 2},
		{name: "三干皆印比=3",
			chart: Chart{
				Ri: zhuInfo{Zhu: ganzhi.Zhu{Gan: ganzhi.GanJia}}, Nian: zhuInfo{Zhu: ganzhi.Zhu{Gan: ganzhi.GanYi}},
				Yue: zhuInfo{Zhu: ganzhi.Zhu{Gan: ganzhi.GanGui}}, Shi: zhuInfo{Zhu: ganzhi.Zhu{Gan: ganzhi.GanYi}}},
			wantYinBi: 3},
		{name: "三干皆官杀=0",
			chart: Chart{
				Ri: zhuInfo{Zhu: ganzhi.Zhu{Gan: ganzhi.GanJia}}, Nian: zhuInfo{Zhu: ganzhi.Zhu{Gan: ganzhi.GanGeng}},
				Yue: zhuInfo{Zhu: ganzhi.Zhu{Gan: ganzhi.GanXin}}, Shi: zhuInfo{Zhu: ganzhi.Zhu{Gan: ganzhi.GanGeng}}},
			wantYinBi: 0},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := countYinBi(tt.chart)
			if got != tt.wantYinBi {
				t.Errorf("countYinBi=%d, want=%d", got, tt.wantYinBi)
			}
		})
	}
}

func TestRegress_CountStems_ExcludesSelf(t *testing.T) {
	chart := Chart{
		Ri: zhuInfo{Zhu: ganzhi.Zhu{Gan: ganzhi.GanJia}},
		Nian: zhuInfo{Zhu: ganzhi.Zhu{Gan: ganzhi.GanYi}},
		Yue: zhuInfo{Zhu: ganzhi.Zhu{Gan: ganzhi.GanGui}},
		Shi: zhuInfo{Zhu: ganzhi.Zhu{Gan: ganzhi.GanGeng}},
	}
	sc := countStems(chart)
	if sc.biBi != 1 || sc.yin != 1 || sc.guanSha != 1 {
		t.Errorf("countStems=(biBi=%d,yin=%d,guan=%d), want(1,1,1)", sc.biBi, sc.yin, sc.guanSha)
	}
}

