package ziwei

import (
	"testing"

	"liki-engine/internal/engine/ganzhi"
	"liki-engine/internal/engine/tianwen"
)

func flowDate(year, month, day int, leap bool) tianwen.LunarDate {
	return tianwen.LunarDate{Year: year, Month: month, Day: day, Leap: leap}
}

func TestComputeLiuRiAndLiuShiValidateLeapDate(t *testing.T) {
	chart := liuYueTestChart(t)
	shiZhi := ganzhi.Zhi(1)

	valid := flowDate(2025, 6, 20, true)
	if _, err := ComputeLiuRi(chart, valid); err != nil {
		t.Fatalf("valid leap flow day: %v", err)
	}
	if _, err := ComputeLiuShi(chart, valid, shiZhi); err != nil {
		t.Fatalf("valid leap flow hour: %v", err)
	}

	missingLeap := flowDate(2024, 6, 20, true)
	if _, err := ComputeLiuRi(chart, missingLeap); err == nil {
		t.Fatal("nonexistent leap flow day succeeded, want error")
	}
	if _, err := ComputeLiuShi(chart, missingLeap, shiZhi); err == nil {
		t.Fatal("nonexistent leap flow hour succeeded, want error")
	}

	beyondMonth := flowDate(2025, 6, 31, true)
	if _, err := ComputeLiuRi(chart, beyondMonth); err == nil {
		t.Fatal("date beyond lunar month succeeded, want error")
	}
	if _, err := ComputeLiuShi(chart, beyondMonth, shiZhi); err == nil {
		t.Fatal("date beyond lunar month succeeded, want error")
	}
}

func TestComputeChartAcceptsLocalAstronomicalLeapMonth(t *testing.T) {
	// ComputeChart is lunar-native: the upstream reference decides whether a
	// leap month exists. 2025 leap 7 is real in the local astronomical calendar
	// for America/New_York even though the official UTC+8 table has no leap 7.
	lt := tianwen.LunarTime{Year: 2025, Month: 7, Day: 2, Leap: true, Shichen: ganzhi.ZhiChen}
	if _, err := ComputeChart(lt, ganzhi.Male); err != nil {
		t.Fatalf("local astronomical leap month rejected: %v", err)
	}
}
