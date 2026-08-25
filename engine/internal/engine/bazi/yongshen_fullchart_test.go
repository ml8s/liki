package bazi

import (
	"encoding/json"
	"strings"
	"testing"
	"time"

	"liki-engine/internal/engine/ganzhi"
	"liki-engine/internal/engine/tianwen"
)

// 用神三派归完整命盘（bazi.fullchart 承载）；chart 纯排盘不含 yong_shen。
func TestFullChart_YongShen(t *testing.T) {
	st := tianwen.GregorianToSolar(
		time.Date(1990, 6, 1, 12, 0, 0, 0, time.FixedZone("CST", 8*3600)), 116.4, 8)
	chart := ComputeChart(st, ganzhi.Male)
	if b, _ := json.Marshal(chart); strings.Contains(string(b), "yong_shen") {
		t.Error("chart 不应含用神（纯排盘，yong_shen 归 fullchart）")
	}
	full := ComputeFullChart(chart)
	if full.YongShen.FuYi.Yong == "" || full.YongShen.TiaoHou.Yong == "" || full.YongShen.GeJu.Yong == "" {
		t.Errorf("fullchart.yong_shen 三派应非空: fu_yi=%q tiao_hou=%q ge_ju=%q",
			full.YongShen.FuYi.Yong, full.YongShen.TiaoHou.Yong, full.YongShen.GeJu.Yong)
	}
}
