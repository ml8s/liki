package ziwei

import (
	"reflect"
	"sort"
	"testing"

	"liki-engine/internal/engine/ganzhi"
	"liki-engine/internal/engine/tianwen"
)

func liuYueTestChart(t *testing.T) Chart {
	t.Helper()
	shiZhi, err := ganzhi.ParseZhi("辰")
	if err != nil {
		t.Fatalf("parse shichen: %v", err)
	}
	chart, err := ComputeChart(tianwen.LunarTime{Year: 2000, Month: 8, Day: 23, Shichen: shiZhi}, Female)
	if err != nil {
		t.Fatalf("ComputeChart: %v", err)
	}
	return chart
}

func lunarDate(year, month, day int, leap bool) tianwen.LunarDate {
	return tianwen.LunarDate{Year: year, Month: month, Day: day, Leap: leap}
}

func TestResolveLiuYuePeriod(t *testing.T) {
	meta, err := tianwen.LunarMonthMeta(2025, 6, true)
	if err != nil {
		t.Fatalf("load leap-month oracle: %v", err)
	}

	tests := []struct {
		name      string
		target    tianwen.LunarDate
		wantHalf  string
		wantStart int
		wantEnd   int
		wantFlow  LunarMonth
	}{
		{
			name:      "ordinary month is whole",
			target:    lunarDate(2025, 6, 20, false),
			wantHalf:  "whole",
			wantStart: 1,
			wantEnd:   mustLunarMonthDayCount(t, 2025, 6, false),
			wantFlow:  LunarMonth{Year: 2025, Month: 6, Leap: false},
		},
		{
			name:      "leap month first half boundary",
			target:    lunarDate(2025, 6, 15, true),
			wantHalf:  "first",
			wantStart: 1,
			wantEnd:   15,
			wantFlow:  LunarMonth{Year: 2025, Month: 6, Leap: true},
		},
		{
			name:      "leap month second half boundary",
			target:    lunarDate(2025, 6, 16, true),
			wantHalf:  "second",
			wantStart: 16,
			wantEnd:   meta.DayCount,
			wantFlow:  LunarMonth{Year: 2025, Month: 7, Leap: false},
		},
		{
			name:      "leap month last day",
			target:    lunarDate(2025, 6, meta.DayCount, true),
			wantHalf:  "second",
			wantStart: 16,
			wantEnd:   meta.DayCount,
			wantFlow:  LunarMonth{Year: 2025, Month: 7, Leap: false},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := resolveLiuYuePeriod(tt.target)
			if err != nil {
				t.Fatalf("resolveLiuYuePeriod: %v", err)
			}
			calendar := got.CalendarPeriod
			if calendar.Year != tt.target.Year || calendar.Month != tt.target.Month || calendar.Leap != tt.target.Leap {
				t.Fatalf("calendar = %+v, want target month %+v", calendar, tt.target)
			}
			if calendar.Half != tt.wantHalf || calendar.DayStart != tt.wantStart || calendar.DayEnd != tt.wantEnd {
				t.Fatalf("calendar = %+v, want half=%s days=%d-%d", calendar, tt.wantHalf, tt.wantStart, tt.wantEnd)
			}
			if got.FlowMonth != tt.wantFlow {
				t.Fatalf("flow month = %+v, want %+v", got.FlowMonth, tt.wantFlow)
			}
		})
	}
}

func TestResolveLiuYuePeriod_Errors(t *testing.T) {
	tests := []struct {
		name   string
		target tianwen.LunarDate
	}{
		{name: "missing leap month", target: lunarDate(2024, 6, 10, true)},
		{name: "day before month", target: lunarDate(2025, 6, 0, true)},
		{name: "day after month", target: lunarDate(2025, 6, 31, true)},
		{name: "invalid month", target: lunarDate(2025, 13, 10, false)},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if _, err := resolveLiuYuePeriod(tt.target); err == nil {
				t.Fatal("resolveLiuYuePeriod succeeded, want error")
			}
		})
	}
}

func TestComputeLiuYue_LeapHalfMatchesFlowMonth(t *testing.T) {
	chart := liuYueTestChart(t)
	firstLeap, err := ComputeLiuYue(chart, lunarDate(2025, 6, 15, true))
	if err != nil {
		t.Fatalf("first leap half: %v", err)
	}
	ordinarySix, err := ComputeLiuYue(chart, lunarDate(2025, 6, 15, false))
	if err != nil {
		t.Fatalf("ordinary sixth month: %v", err)
	}
	assertSameLiuYueChart(t, firstLeap, ordinarySix)

	secondLeap, err := ComputeLiuYue(chart, lunarDate(2025, 6, 16, true))
	if err != nil {
		t.Fatalf("second leap half: %v", err)
	}
	ordinarySeven, err := ComputeLiuYue(chart, lunarDate(2025, 7, 16, false))
	if err != nil {
		t.Fatalf("ordinary seventh month: %v", err)
	}
	assertSameLiuYueChart(t, secondLeap, ordinarySeven)
}

func assertSameLiuYueChart(t *testing.T, got, want LiuYue) {
	t.Helper()
	got, want = liuYueWithSortedStars(got), liuYueWithSortedStars(want)
	if got.MingGong != want.MingGong || got.MingGongName != want.MingGongName || got.Zhi != want.Zhi {
		t.Fatalf("core = %d/%s/%s, want %d/%s/%s", got.MingGong, got.MingGongName, got.Zhi, want.MingGong, want.MingGongName, want.Zhi)
	}
	if !reflect.DeepEqual(got.SiHua, want.SiHua) {
		t.Fatalf("si hua = %#v, want %#v", got.SiHua, want.SiHua)
	}
	if !reflect.DeepEqual(got.Stars, want.Stars) {
		t.Fatalf("stars = %#v, want %#v", got.Stars, want.Stars)
	}
	if !reflect.DeepEqual(got.GongWei, want.GongWei) {
		t.Fatalf("gong wei = %#v, want %#v", got.GongWei, want.GongWei)
	}
}

func liuYueWithSortedStars(value LiuYue) LiuYue {
	for i := range value.GongWei {
		stars := append([]string(nil), value.GongWei[i].Stars...)
		sort.Strings(stars)
		value.GongWei[i].Stars = stars
	}
	return value
}

func mustLunarMonthDayCount(t *testing.T, year, month int, leap bool) int {
	t.Helper()
	meta, err := tianwen.LunarMonthMeta(year, month, leap)
	if err != nil {
		t.Fatalf("LunarMonthMeta(%d-%02d leap=%t): %v", year, month, leap, err)
	}
	return meta.DayCount
}
