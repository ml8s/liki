package bazi

import "testing"

// 大运步骤已携带公历年段（start_year/end_year）——ComputeCurrentStepIndex 直接按年份判断。
func TestComputeCurrentStepIndex(t *testing.T) {
	dy := &DaYun{
		StartDate: "2026-03-15",
		Direction: "顺排",
		Steps: []DaYunStep{
			{StartYear: 2026, EndYear: 2035}, // index 0
			{StartYear: 2036, EndYear: 2045}, // index 1
			{StartYear: 2046, EndYear: 2055}, // index 2
			{StartYear: 2056, EndYear: 2065}, // index 3
			{StartYear: 2066, EndYear: 2075}, // index 4
			{StartYear: 2076, EndYear: 2085}, // index 5
			{StartYear: 2086, EndYear: 2095}, // index 6
			{StartYear: 2096, EndYear: 2105}, // index 7
		},
	}
	tests := []struct {
		currentYr int
		want      int
		desc      string
	}{
		{2025, -1, "未起运"},
		{2026, 0, "首步起始年"},
		{2035, 0, "首步末年"},
		{2036, 1, "第二步起始年"},
		{2060, 3, "中段"},
		{2096, 7, "末步起始年"},
		{2105, 7, "末步末年"},
		{2106, -1, "已过完"},
	}
	for _, tt := range tests {
		t.Run(tt.desc, func(t *testing.T) {
			got := ComputeCurrentStepIndex(dy, tt.currentYr)
			if got != tt.want {
				t.Errorf("ComputeCurrentStepIndex(%d) = %d; want %d (%s)",
					tt.currentYr, got, tt.want, tt.desc)
			}
		})
	}
}
