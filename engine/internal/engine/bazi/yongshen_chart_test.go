package bazi

import (
	"encoding/json"
	"strings"
	"testing"
	"time"

	"liki-engine/internal/engine/ganzhi"
	"liki-engine/internal/engine/tianwen"
)

// FuYi/TiaoHou 候选与 GeJu 结构证据由领域 Chart 直接承载。
func TestChart_YongShen(t *testing.T) {
	stgt := time.Time(time.Date(1990, 6, 1, 12, 0, 0, 0, time.FixedZone("CST", 8*3600)))
	st := tianwen.ApparentSolarFromCivil(
		time.Date(1990, 6, 1, 12, 0, 0, 0, time.FixedZone("CST", 8*3600)), 116.4, 8)
	chart := ComputeChart(stgt, st, ganzhi.Male)
	if b, _ := json.Marshal(chart); strings.Contains(string(b), "yong_shen") {
		t.Error("最终裁决字段 yong_shen 不属于 Chart 输出")
	}
	if chart.FuYi.Yong == "" || chart.FuYi.Yong == NoCandidateWuxing || chart.TiaoHou.PrimaryWuxing == "" || chart.GeJu.Pattern == "" {
		t.Errorf("FuYi/TiaoHou/GeJu 证据应非空: fu_yi=%q tiao_hou=%q ge_ju=%q",
			chart.FuYi.Yong, chart.TiaoHou.PrimaryWuxing, chart.GeJu.Pattern)
	}
}
