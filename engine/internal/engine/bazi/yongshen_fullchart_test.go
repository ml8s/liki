package bazi

import (
	"encoding/json"
	"strings"
	"testing"
	"time"

	"liki-engine/internal/engine/ganzhi"
	"liki-engine/internal/engine/tianwen"
)

// 用神候选与格局证据归完整命盘（bazi.fullchart 承载）；chart 纯排盘不含 yong_shen。
func TestFullChart_YongShen(t *testing.T) {
	stgt := time.Time(time.Date(1990, 6, 1, 12, 0, 0, 0, time.FixedZone("CST", 8*3600)))
	st := tianwen.ApparentSolarFromCivil(
		time.Date(1990, 6, 1, 12, 0, 0, 0, time.FixedZone("CST", 8*3600)), 116.4, 8)
	chart := ComputeChart(stgt, st, ganzhi.Male)
	if b, _ := json.Marshal(chart); strings.Contains(string(b), "yong_shen") {
		t.Error("chart 不应含用神（纯排盘，yong_shen 归 fullchart）")
	}
	full := ComputeFullChart(chart)
	if full.FuYi.Yong == "" || full.FuYi.Yong == NoCandidateWuxing || full.TiaoHou.PrimaryWuxing == "" || full.GeJu.Pattern == "" {
		t.Errorf("fullchart.yong_shen 应非空: fu_yi=%q tiao_hou=%q ge_ju=%q",
			full.FuYi.Yong, full.TiaoHou.PrimaryWuxing, full.GeJu.Pattern)
	}
}
