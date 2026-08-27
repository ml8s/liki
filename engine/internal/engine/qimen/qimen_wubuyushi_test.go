package qimen

import (
	"fmt"
	"testing"
	"time"

	"liki-engine/internal/engine/ganzhi"
	"liki-engine/internal/engine/tianwen"
)

// ── 五不遇时命理知识测试 ──
// 基于奇门经典理论: "五不遇时龙不精，号为日月损光明"
// 时干克日干，且阴阳相同

func TestWuBuYuShi_AllTenDayGan(t *testing.T) {
	// 命理: 10个日干各有1个五不遇时的时干
	tests := []struct {
		riGan  ganzhi.Gan // 日干
		shiGan ganzhi.Gan // 五不遇时的时干
	}{
		{riGan: ganzhi.GanJia, shiGan: ganzhi.GanGeng},  // 甲→庚(阳金克阳木)
		{riGan: ganzhi.GanYi, shiGan: ganzhi.GanXin},    // 乙→辛(阴金克阴木)
		{riGan: ganzhi.GanBing, shiGan: ganzhi.GanRen},  // 丙→壬(阳水克阳火)
		{riGan: ganzhi.GanDing, shiGan: ganzhi.GanGui},  // 丁→癸(阴水克阴火)
		{riGan: ganzhi.GanWu, shiGan: ganzhi.GanJia},    // 戊→甲(阳木克阳土)
		{riGan: ganzhi.GanJi, shiGan: ganzhi.GanYi},     // 己→乙(阴木克阴土)
		{riGan: ganzhi.GanGeng, shiGan: ganzhi.GanBing}, // 庚→丙(阳火克阳金)
		{riGan: ganzhi.GanXin, shiGan: ganzhi.GanDing},  // 辛→丁(阴火克阴金)
		{riGan: ganzhi.GanRen, shiGan: ganzhi.GanWu},    // 壬→戊(阳土克阳水)
		{riGan: ganzhi.GanGui, shiGan: ganzhi.GanJi},    // 癸→己(阴土克阴水)
	}
	for _, tt := range tests {
		name := tt.riGan.String() + "日" + tt.shiGan.String() + "时→五不遇"
		t.Run(name, func(t *testing.T) {
			got := isWuBuYuShi(tt.riGan, tt.shiGan)
			if !got {
				t.Errorf("isWuBuYuShi(%s,%s)=false, 命理应为true(五不遇时)",
					tt.riGan.String(), tt.shiGan.String())
			}
		})
	}
}

func TestWuBuYuShi_NonMatching_ReturnsFalse(t *testing.T) {
	// 命理: 非五不遇时的组合应返回false
	tests := []struct {
		riGan  ganzhi.Gan
		shiGan ganzhi.Gan
		desc   string
	}{
		{ganzhi.GanJia, ganzhi.GanBing, "甲丙: 丙克甲? 火克金? 不, 丙火生甲木→非克"},
		{ganzhi.GanJia, ganzhi.GanYi, "甲乙: 同木比和, 非克"},
		{ganzhi.GanJia, ganzhi.GanDing, "甲丁: 丁火生甲木(火生木→非克)"},
		{ganzhi.GanBing, ganzhi.GanJia, "丙甲: 甲木生丙火→非克"},
		{ganzhi.GanGeng, ganzhi.GanJia, "庚甲: 甲庚互换→非时干克日干"},
		{ganzhi.GanGui, ganzhi.GanRen, "癸壬: 同水比和, 非克"},
	}
	for _, tt := range tests {
		name := tt.riGan.String() + "日" + tt.shiGan.String() + "时→非"
		t.Run(name, func(t *testing.T) {
			got := isWuBuYuShi(tt.riGan, tt.shiGan)
			if got {
				t.Errorf("isWuBuYuShi(%s,%s)=true, 命理应为false(%s)",
					tt.riGan.String(), tt.shiGan.String(), tt.desc)
			}
		})
	}
}

func TestWuBuYuShi_SameYinYang_Required(t *testing.T) {
	// 命理: 时干克日干但阴阳不同 → 不是五不遇时
	// 戊日(阳): 庚时(阳金克阳木→不对, 庚克甲不克戊)
	// 实际查验: 庚克甲(金克木), 庚不克戊(金生水→不对, 庚金=土生金)

	// 验证最重要的: 丙日(阳火) → 壬(阳水克阳火=五不遇) ✅
	// 但癸(阴水克阳火=阴阳不同) → 不是五不遇时
	tests := []struct {
		riGan  ganzhi.Gan
		shiGan ganzhi.Gan
		desc   string
	}{
		{ganzhi.GanBing, ganzhi.GanGui, "丙癸: 癸水克丙火但阴克阳→非五不遇"},
		{ganzhi.GanDing, ganzhi.GanRen, "丁壬: 壬水克丁火但阳克阴→非五不遇"},
	}
	for _, tt := range tests {
		name := tt.riGan.String() + "日" + tt.shiGan.String() + "时"
		t.Run(name, func(t *testing.T) {
			got := isWuBuYuShi(tt.riGan, tt.shiGan)
			if got {
				t.Errorf("isWuBuYuShi(%s,%s)=true, 命理应为false(%s)",
					tt.riGan.String(), tt.shiGan.String(), tt.desc)
			}
		})
	}
}

func TestWuBuYuShi_ChartField_Present(t *testing.T) {
	// 验证所有chart都正确计算wu_bu_yu_shi字段
	dates := []string{"2026-01-01", "2026-06-15", "2026-12-25"}
	for _, d := range dates {
		t.Run(d, func(t *testing.T) {
			for hour := 0; hour < 24; hour += 2 {
				bt, err := time.Parse("2006-01-02 15", d+" "+fmt.Sprintf("%02d", hour))
				if err != nil {
					t.Fatal(err)
				}
				st := tianwen.GregorianToSolar(bt, 116.4, 8)
				chart := ComputeChart(st, ShiQiMen)
				// 验证isWuBuYuShi与chart字段一致
				riGan := chart.Pan.RiGan
				shiGan := chart.Pan.DriveGan
				expected := isWuBuYuShi(riGan, shiGan)
				if chart.Pan.WuBuYuShi != expected {
					t.Errorf("chart.Pan.WuBuYuShi=%v, isWuBuYuShi(%s,%s)=%v",
						chart.Pan.WuBuYuShi, riGan.String(), shiGan.String(), expected)
				}
			}
		})
	}
}
