package qimen

import (
	"testing"
	"time"

	"liki-engine/internal/engine/tianwen"
)

// ── 奇门命理知识测试 ──

func TestQimen_YinYangDun_Seasonal(t *testing.T) {
	tests := []struct {
		name    string
		date    string
		wantYin bool
	}{
		{name: "冬至后(阳遁)", date: "2024-12-22", wantYin: false},
		{name: "夏至后(阴遁)", date: "2024-06-22", wantYin: true},
		{name: "立春后(阳遁)", date: "2024-02-05", wantYin: false},
		{name: "立秋后(阴遁)", date: "2024-08-08", wantYin: true},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			bt, err := time.Parse("2006-01-02", tt.date)
			if err != nil {
				t.Fatal(err)
			}
			st := tianwen.GregorianToSolar(bt, 116.4, 8)
			chart := ComputeChart(st, "时家")
			if chart.Pan.YinDun != tt.wantYin {
				t.Errorf("%s: yin=%v, want %v", tt.date, chart.Pan.YinDun, tt.wantYin)
			}
		})
	}
}

func TestQimen_JieQiDingJu_YangDun(t *testing.T) {
	// 冬至后应阳遁, 且局数在1-9范围内
	dates := []string{"2024-12-22", "2025-01-06", "2025-02-04"}
	for _, d := range dates {
		t.Run(d, func(t *testing.T) {
			bt, err := time.Parse("2006-01-02", d)
			if err != nil {
				t.Fatal(err)
			}
			st := tianwen.GregorianToSolar(bt, 116.4, 8)
			chart := ComputeChart(st, "时家")
			if chart.Pan.YinDun {
				t.Errorf("%s: 冬至后应为阳遁, got yin=%v", d, chart.Pan.YinDun)
			}
			if chart.Pan.Jushu < 1 || chart.Pan.Jushu > 9 {
				t.Errorf("%s: ju=%d out of range(1-9)", d, chart.Pan.Jushu)
			}
		})
	}
}

func TestQimen_ZhiFuZhiShi_NonEmpty(t *testing.T) {
	dates := []string{"2024-12-22", "2024-06-22", "2024-03-20", "2024-09-22"}
	for _, d := range dates {
		t.Run(d, func(t *testing.T) {
			bt, err := time.Parse("2006-01-02", d)
			if err != nil {
				t.Fatal(err)
			}
			st := tianwen.GregorianToSolar(bt, 116.4, 8)
			chart := ComputeChart(st, "时家")
			if chart.Pan.DutyStar < 1 || chart.Pan.DutyStar > 9 {
				t.Errorf("%s: dutyStar=%d(out of range)", d, chart.Pan.DutyStar)
			}
			if chart.Pan.DutyDoor < 1 || chart.Pan.DutyDoor > 8 {
				t.Errorf("%s: dutyDoor=%d(out of range)", d, chart.Pan.DutyDoor)
			}
		})
	}
}
