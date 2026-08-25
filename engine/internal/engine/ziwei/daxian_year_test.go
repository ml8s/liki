package ziwei

import (
	"testing"
	"time"

	"liki-engine/internal/engine/ganzhi"
	"liki-engine/internal/engine/tianwen"
)

// 外部评审 ⑧：大限公历年段（start_year = birthYear + QiSui − 1，虚岁→公历年）。
func TestDaXian_StartYear(t *testing.T) {
	st := tianwen.GregorianToSolar(
		time.Date(1981, 8, 26, 12, 0, 0, 0, time.FixedZone("CST", 8*3600)), 116.4, 8)
	lt := tianwen.SolarToLunar(tianwen.GregorianTime(st.Time()))
	chart := ComputeChart(lt, ganzhi.Male)
	steps := ComputeDaXian(chart)
	if len(steps) == 0 {
		t.Fatal("大限为空")
	}
	qiSui := steps[0].QiSui
	wantStart := chart.BirthYear + qiSui - 1
	if steps[0].StartYear != wantStart {
		t.Errorf("大限[0] start_year = %d, want %d（birthYear %d + QiSui %d − 1）",
			steps[0].StartYear, wantStart, chart.BirthYear, qiSui)
	}
	// 每步 10 年、连续
	for i := 1; i < len(steps); i++ {
		if steps[i].StartYear != steps[i-1].StartYear+10 {
			t.Errorf("大限[%d] start_year = %d, want 上步+10（%d）", i, steps[i].StartYear, steps[i-1].StartYear+10)
		}
		if steps[i].EndYear != steps[i].StartYear+9 {
			t.Errorf("大限[%d] end_year = %d, want start+9（%d）", i, steps[i].EndYear, steps[i].StartYear+9)
		}
	}
}
