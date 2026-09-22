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

func TestComputeChartRejectsInvalidLunarDate(t *testing.T) {
	lt := tianwen.LunarTime{Year: 2024, Month: 6, Day: 1, Leap: true, Shichen: ganzhi.ZhiChen}
	if _, err := ComputeChart(lt, ganzhi.Male); err == nil {
		t.Fatal("chart for nonexistent leap month succeeded, want error")
	}

	lt = tianwen.LunarTime{Year: 2024, Month: 13, Day: 1, Shichen: ganzhi.ZhiChen}
	if _, err := ComputeChart(lt, ganzhi.Male); err == nil {
		t.Fatal("chart for lunar month 13 succeeded, want error")
	}
}
