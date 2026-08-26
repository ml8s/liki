package bazi

import (
	"testing"
	"time"

	"liki-engine/internal/engine/ganzhi"
	"liki-engine/internal/engine/tianwen"
)

// ── 三派对比测试 ──
// 三个学派独立输出，无交叉污染。验证各派均有返回。

func TestThreeSchools_Consistency(t *testing.T) {
	tests := []struct {
		name                   string
		year, month, day, hour int
		gender                 ganzhi.Gender
		wantFuYiEmpty          bool // 允许扶抑为中和→空
	}{
		{name: "己日寅月", year: 1984, month: 2, day: 15, hour: 8,
			gender: ganzhi.Male, wantFuYiEmpty: true},
		{name: "乙日申月", year: 2001, month: 8, day: 25, hour: 12,
			gender: ganzhi.Male, wantFuYiEmpty: false},
		{name: "丁日午月", year: 1966, month: 6, day: 15, hour: 12,
			gender: ganzhi.Male, wantFuYiEmpty: true},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			loc := time.FixedZone("CST", 8*3600)
			birth := time.Date(tt.year, time.Month(tt.month), tt.day, tt.hour, 0, 0, 0, loc)
			st := tianwen.GregorianToSolar(birth, 116.4, 8)
			chart := ComputeChart(st, tt.gender)
			result := ComputeYongShen(chart)

			t.Logf("%s%s %s%s %s%s %s%s",
				chart.Nian.Gan, chart.Nian.Zhi,
				chart.Yue.Gan, chart.Yue.Zhi,
				chart.Ri.Gan, chart.Ri.Zhi,
				chart.Shi.Gan, chart.Shi.Zhi)

			t.Logf("  扶抑: strength=%s yong=%s xi=%s ji=%s pattern=%s",
				result.FuYi.Strength, result.FuYi.Yong, result.FuYi.Xi, result.FuYi.Ji, result.FuYi.Pattern)
			t.Logf("  调候: yong=%s xi=%s ji=%s",
				result.TiaoHou.Yong, result.TiaoHou.Xi, result.TiaoHou.Ji)
			t.Logf("  格局: %s %s yong=%s xi=%s ji=%s",
				result.GeJu.Pattern, result.GeJu.Usage,
				result.GeJu.Yong, result.GeJu.Xi, result.GeJu.Ji)

			// 调候/格局必须有返回
			if result.TiaoHou.Yong == "" {
				t.Error("TiaoHou.Yong 为空")
			}
			if result.GeJu.Pattern == "" {
				t.Error("格局为空")
			}
			if !tt.wantFuYiEmpty && result.FuYi.Yong == "" {
				t.Errorf("FuYi.Yong 不应为空(strength=%s)", result.FuYi.Strength)
			}
		})
	}
}
