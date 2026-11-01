package ziwei

import (
	"testing"
	"time"

	"liki-engine/internal/engine/ganzhi"
	"liki-engine/internal/engine/tianwen"
)

// 无主星宫位借对宫主星。
// 1981-08-26T12:00 男命（LunarDateAt 路径实测）：命宫空宫 → 借对宫主星（紫微/贪狼）。
func TestKongGong_BorrowFromDuiGong(t *testing.T) {
	st := tianwen.ApparentSolarFromCivil(
		time.Date(1981, 8, 26, 12, 0, 0, 0, time.FixedZone("CST", 8*3600)), 116.4, 8)
	lt := lunarTimeFromSolar(st)
	chart, err := ComputeChart(lt, ganzhi.Male)
	if err != nil {
		t.Fatalf("ComputeChart: %v", err)
	}

	// 命宫确认为空宫（无主星）——测试前置
	mg := chart.MingGong
	if hasMajorStar(chart.GongWei[mg]) {
		t.Skip("此命例命宫有主星，换锚点")
	}

	// kong_gong 只表达借星关系；借入星曜仍由来源宫唯一承载。
	found := false
	for _, kg := range chart.KongGong {
		if kg.KongGong != mg {
			continue
		}
		found = true
		dui := (int(mg) + 6) % 12
		if kg.JieGong != gongIndex(dui) {
			t.Fatalf("借星宫 = %s, want %s", gongLabels[kg.JieGong], gongLabels[dui])
		}
		if !hasMajorStar(chart.GongWei[kg.JieGong]) {
			t.Errorf("借星宫 %s 无主星", gongLabels[kg.JieGong])
		}
	}
	if !found {
		t.Errorf("kong_gong 未包含命宫（空宫应记录）: %+v", chart.KongGong)
	}
}
