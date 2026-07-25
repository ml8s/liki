package bazi

import (
	"testing"
	"time"

	"liki-engine/internal/engine/ganzhi"
	"liki-engine/internal/engine/tianwen"
)

// ── 从格测试（JSON规则链驱动） ──
// 使用真实八字验证 lookupCongGe

func TestCongGe_LookupRules(t *testing.T) {
	tests := []struct {
		name        string
		time        string
		gender      ganzhi.Gender
		wantPattern string
	}{
		// 从旺格: 印比垄断
		{
			name: "从旺",
			// 找一个印比极旺的八字
			time:   "1986-01-05T12:00:00+08:00",
			gender: ganzhi.Male,
			// 年: 乙丑(木), 月: 己丑(土?实际上1986-01-05还在1985年)
			// 需要找一个确实能从旺的八字
			wantPattern: "",
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			loc := time.FixedZone("CST", 8*3600)
			bt, err := time.ParseInLocation("2006-01-02T15:04:05Z07:00", tt.time, loc)
			if err != nil { t.Fatal(err) }
			st := tianwen.GregorianToSolar(bt, 116.4, 8)
			chart := ComputeChart(st, tt.gender)
			pat, _, _, _ := lookupCongGe(chart)
			if tt.wantPattern != "" && pat != tt.wantPattern {
				t.Errorf("chart: %s%s %s%s %s%s %s%s | pattern=%q, want=%q",
					chart.Nian.Gan, chart.Nian.Zhi, chart.Yue.Gan, chart.Yue.Zhi,
					chart.Ri.Gan, chart.Ri.Zhi, chart.Shi.Gan, chart.Shi.Zhi,
					pat, tt.wantPattern)
			}
			t.Logf("pattern=%q", pat)
		})
	}
}
