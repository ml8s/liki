package bazi

import (
	"encoding/json"
	"strings"
	"testing"
	"time"

	"liki-engine/internal/engine/ganzhi"
	"liki-engine/internal/engine/tianwen"
)

// 外部评审 ⑨：chart 输出子时换日规则说明（zi_shi_rule）。
func TestZiShiRule(t *testing.T) {
	st := tianwen.GregorianToSolar(
		time.Date(1981, 8, 26, 23, 30, 0, 0, time.FixedZone("CST", 8*3600)), 130.7, 8)
	chart := ComputeChart(st, ganzhi.Male)
	if !strings.Contains(chart.ZiShiRule, "晚子时") {
		t.Errorf("zi_shi_rule = %q, want 含'晚子时'", chart.ZiShiRule)
	}
	if b, _ := json.Marshal(chart); !strings.Contains(string(b), "zi_shi_rule") {
		t.Error("chart JSON 应含 zi_shi_rule 键")
	}
}

// 外部评审 ⑧ 重构锚点：起运公历日 + 首步日期段（1984-02-15 男 = lunar 6年5月20日 → 起运 1990-08-04）。
func TestDaYun_StartDate_Anchor(t *testing.T) {
	st := tianwen.GregorianToSolar(
		time.Date(1984, 2, 15, 8, 0, 0, 0, time.FixedZone("CST", 8*3600)), 116.4, 8)
	chart := ComputeChart(st, ganzhi.Male)
	if chart.DaYun.StartDate != "1990-08-04" {
		t.Errorf("起运 start_date = %q, want 1990-08-04（lunar 6年5月20日）", chart.DaYun.StartDate)
	}
	s0 := chart.DaYun.Steps[0]
	if s0.StartDate != "1990-08-04" || s0.EndDate != "2000-08-03" {
		t.Errorf("steps[0] = %s~%s, want 1990-08-04~2000-08-03（10年-1天）", s0.StartDate, s0.EndDate)
	}
	if s0.StartYear != 1990 || s0.EndYear != 2000 {
		t.Errorf("steps[0] 公历年 = %d-%d, want 1990-2000", s0.StartYear, s0.EndYear)
	}
}

// 外部评审 ⑦：无 gong_jia/san_qi 的命例 → JSON 缺键（omitempty 未命中缺席）。
func TestOmitEmpty_Optional(t *testing.T) {
	st := tianwen.GregorianToSolar(
		time.Date(1981, 8, 26, 12, 0, 0, 0, time.FixedZone("CST", 8*3600)), 130.7, 8)
	full := ComputeFullChart(ComputeChart(st, ganzhi.Male))
	b, _ := json.Marshal(full)
	// 该命例若无拱夹/三奇 → 键缺席（omitempty）；若有则键存在且非 null——断言"不会是 null"
	if strings.Contains(string(b), `"gong_jia":null`) {
		t.Error("gong_jia 不应为 null（omitempty：未命中缺席）")
	}
	if strings.Contains(string(b), `"san_qi_name":""`) {
		t.Error("san_qi_name 不应为 \"\"（omitempty：未命中缺席）")
	}
}
