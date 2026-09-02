package ziwei

import (
	"testing"
	"time"

	"liki-engine/internal/engine/ganzhi"
	"liki-engine/internal/engine/tianwen"
)

// 无主星宫位借对宫主星。
// 1981-08-26T12:00 男命（SolarToLunar 路径实测）：命宫空宫 → 借对宫主星（紫微/贪狼）。
func TestKongGong_BorrowFromDuiGong(t *testing.T) {
	st := tianwen.GregorianToSolar(
		time.Date(1981, 8, 26, 12, 0, 0, 0, time.FixedZone("CST", 8*3600)), 116.4, 8)
	lt := tianwen.SolarToLunar(tianwen.GregorianTime(st.Time()))
	chart := ComputeChart(lt, ganzhi.Male)

	// 命宫确认为空宫（无主星）——测试前置
	mg := chart.MingGong
	if hasMajorStar(chart.GongWei[mg]) {
		t.Skip("此命例命宫有主星，换锚点")
	}

	// kong_gong 输出必须包含命宫，且借星=对宫主星
	found := false
	for _, kg := range chart.KongGong {
		if kg.GongName != gongLabels[mg] {
			continue
		}
		found = true
		if len(kg.JieXing) == 0 {
			t.Errorf("命宫空宫借星为空（对宫 %s 应有主星）", kg.DuiGong)
		}
		// 借星 = 对宫主星名
		dui := (int(mg) + 6) % 12
		var want []string
		for _, s := range chart.GongWei[dui].Stars {
			if s.IsMajor {
				want = append(want, s.Name)
			}
		}
		if len(kg.JieXing) != len(want) {
			t.Errorf("借星 %v, want 对宫主星 %v", kg.JieXing, want)
		}
	}
	if !found {
		t.Errorf("kong_gong 未包含命宫（空宫应记录）: %+v", chart.KongGong)
	}
}
