package bazi

import (
	"testing"
)

func TestComputeCurrentStepIndex(t *testing.T) {
	dy := &DaYun{
		StartAge:  6,
		Direction: "顺排",
		Steps: []DaYunStep{
			{AgeStart: 6, AgeEnd: 15},   // index 0
			{AgeStart: 16, AgeEnd: 25},  // index 1
			{AgeStart: 26, AgeEnd: 35},  // index 2
			{AgeStart: 36, AgeEnd: 45},  // index 3
			{AgeStart: 46, AgeEnd: 55},  // index 4
			{AgeStart: 56, AgeEnd: 65},  // index 5
			{AgeStart: 66, AgeEnd: 75},  // index 6
			{AgeStart: 76, AgeEnd: 85},  // index 7
		},
	}

	// The function computes: age = currentYear - birthYear,
	// then if currentYearDay < birthYearDay (birthday hasn't passed), age--.
	tests := []struct {
		name      string
		birthYear int
		currentYr int
		currentYD int
		birthYD   int
		want      int
		desc      string
	}{
		// Before 大运 starts (age < 6)
		{"before_大运_婴儿", 2020, 2021, 180, 200, -1, "age 1"},
		{"before_大运_5岁生日前", 2020, 2025, 180, 200, -1, "age=5-1=4"},
		{"before_大运_5岁生日后", 2020, 2025, 250, 200, -1, "age=5"},

		// First step (age 6-15)
		{"first_step_6岁生日前", 2020, 2026, 180, 200, -1, "age=6-1=5"},
		{"first_step_6岁生日后", 2020, 2026, 250, 200, 0, "age=6"},
		{"first_step_15岁生日前", 2020, 2035, 180, 200, 0, "age=15-1=14"},
		{"first_step_15岁生日后", 2020, 2035, 250, 200, 0, "age=15"},

		// Second step (age 16-25)
		{"second_step_16岁生日后", 2020, 2036, 250, 200, 1, "age=16"},

		// 1981年生人 + 1981→2026 (currentYD 180 < birthYD 238 → age=45-1=44)
		{"1981_44岁_壬辰运", 1981, 2025, 180, 238, 3, "age=44, 36-45=index3"},
		// 45岁生日后 → age=45
		{"1981_45岁生日后_壬辰运", 1981, 2026, 250, 238, 3, "age=45, 36-45=index3"},
		// 46岁生日后 → age=46
		{"1981_46岁_辛卯运", 1981, 2027, 250, 238, 4, "age=46, 46-55=index4"},

		// Last step (age 76-85)
		{"last_step_76岁生日后", 2020, 2096, 250, 200, 7, "age=76"},

		// Past all steps (age > 85)
		{"past_all_86岁生日后", 2020, 2106, 250, 200, -1, "age=86"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := ComputeCurrentStepIndex(dy, tt.birthYear, tt.currentYr, tt.currentYD, tt.birthYD)
			if got != tt.want {
				t.Errorf("ComputeCurrentStepIndex(%d,%d,%d,%d) = %d; want %d (%s)",
					tt.birthYear, tt.currentYr, tt.currentYD, tt.birthYD, got, tt.want, tt.desc)
			}
		})
	}

	// Edge cases: nil and empty DaYun.
	if got := ComputeCurrentStepIndex(nil, 2000, 2026, 180, 200); got != -1 {
		t.Errorf("nil DaYun should return -1, got %d", got)
	}
	empty := &DaYun{Steps: []DaYunStep{}}
	if got := ComputeCurrentStepIndex(empty, 2000, 2026, 180, 200); got != -1 {
		t.Errorf("empty DaYun should return -1, got %d", got)
	}
}
