package bazi

import (
	"testing"
	"time"

	"liki-engine/internal/engine/ganzhi"
	"liki-engine/internal/engine/tianwen"
)

// fullchart 透传用神三派（完整命盘也应含 yong_shen）。
func TestFullChart_YongShen_Transparent(t *testing.T) {
	st := tianwen.GregorianToSolar(
		time.Date(1990, 6, 1, 12, 0, 0, 0, time.FixedZone("CST", 8*3600)), 116.4, 8)
	chart := ComputeChart(st, ganzhi.Male)
	full := ComputeFullChart(chart)
	if full.YongShen.FuYi.Yong != chart.YongShen.FuYi.Yong ||
		full.YongShen.TiaoHou.Yong != chart.YongShen.TiaoHou.Yong ||
		full.YongShen.GeJu.Yong != chart.YongShen.GeJu.Yong {
		t.Errorf("fullchart.yong_shen 与 chart.yong_shen 不一致: %+v vs %+v",
			full.YongShen, chart.YongShen)
	}
}
