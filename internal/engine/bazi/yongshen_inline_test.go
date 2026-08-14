package bazi

import (
	"testing"
	"time"

	"liki-engine/internal/engine/ganzhi"
	"liki-engine/internal/engine/tianwen"
)

// bazi.yongshen 合并进 chart 后的内联覆盖：
// ComputeChart 输出的 YongShen 三派必须非空，且与直接调用 ComputeYongShen 一致。
func TestChart_YongShen_Inline(t *testing.T) {
	cases := []struct {
		name   string
		year   int
		gender ganzhi.Gender
	}{
		{"1990男", 1990, ganzhi.Male},
		{"1984女", 1984, ganzhi.Female},
		{"2005男", 2005, ganzhi.Male},
		{"1977女", 1977, ganzhi.Female},
	}
	for _, c := range cases {
		t.Run(c.name, func(t *testing.T) {
			st := tianwen.GregorianToSolar(
				time.Date(c.year, 6, 1, 12, 0, 0, 0, time.FixedZone("CST", 8*3600)),
				116.4, 8)
			chart := ComputeChart(st, c.gender)
			// 三派非空
			if chart.YongShen.FuYi.Yong == "" || chart.YongShen.TiaoHou.Yong == "" || chart.YongShen.GeJu.Yong == "" {
				t.Errorf("yong_shen 三派用神为空: fu_yi=%q tiao_hou=%q ge_ju=%q",
					chart.YongShen.FuYi.Yong, chart.YongShen.TiaoHou.Yong, chart.YongShen.GeJu.Yong)
			}
			// 与直接调用一致
			direct := ComputeYongShen(chart)
			if direct.FuYi.Yong != chart.YongShen.FuYi.Yong {
				t.Errorf("chart.yong_shen.fu_yi.yong=%q, 直接调用=%q",
					chart.YongShen.FuYi.Yong, direct.FuYi.Yong)
			}
			// JSON 输出含 yong_shen 键（schema 一致性由 agent 层保证，这里锁值）
			got := chart.YongShen
			if got.FuYi.Strength == "" {
				t.Error("fu_yi.qiangruo 为空（身强弱）")
			}
		})
	}
}
