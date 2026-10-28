package bazi

import (
	"reflect"
	"testing"
	"time"

	"liki-engine/internal/engine/ganzhi"
	"liki-engine/internal/engine/tianwen"
)

func TestResolveLiuRiTimeMatchesAlmanac(t *testing.T) {
	date := tianwen.NewCivilDate(2026, 6, 4)
	got, err := ResolveLiuRiTime(date)
	if err != nil {
		t.Fatalf("ResolveLiuRiTime: %v", err)
	}
	want := tianwen.RiZhu(date.Time())
	if got.RiGan != want.Gan || got.RiZhi != want.Zhi {
		t.Fatalf("day pillar = %s%s, want %s%s",
			ganzhi.GanName(got.RiGan), ganzhi.ZhiName(got.RiZhi),
			ganzhi.GanName(want.Gan), ganzhi.ZhiName(want.Zhi))
	}
	if got.Date != date {
		t.Fatalf("date = %s, want %s", got.Date, date)
	}
}

func TestResolveLiuRiTimeRejectsInvalidDate(t *testing.T) {
	if _, err := ResolveLiuRiTime(tianwen.NewCivilDate(2026, 2, 30)); err == nil {
		t.Fatal("expected invalid date error")
	}
}

func TestNextLiuRiTimeAdvancesOneCivilDay(t *testing.T) {
	current, err := ResolveLiuRiTime(tianwen.NewCivilDate(2026, 6, 30))
	if err != nil {
		t.Fatal(err)
	}
	next, err := NextLiuRiTime(current)
	if err != nil {
		t.Fatal(err)
	}
	if next.Date != tianwen.NewCivilDate(2026, 7, 1) {
		t.Fatalf("next date = %s, want 2026-07-01", next.Date)
	}

	prevIdx := ganzhi.SixtyCycleIndex(current.RiGan, current.RiZhi)
	nextIdx := ganzhi.SixtyCycleIndex(next.RiGan, next.RiZhi)
	if nextIdx != (prevIdx+1)%60 {
		t.Fatalf("day cycle index = %d, want %d", nextIdx, (prevIdx+1)%60)
	}
}

func TestResolveLiuNianTimeLichunBoundary(t *testing.T) {
	lichun := tianwen.SolarTermCivilDate(2026, 315)

	before, err := ResolveLiuNianTime(lichun.AddDays(-1))
	if err != nil {
		t.Fatal(err)
	}
	if before.LichunYear != 2025 || before.BoundaryDay {
		t.Fatalf("day before 立春 = %+v, want 2025 without boundary", before)
	}

	boundary, err := ResolveLiuNianTime(lichun)
	if err != nil {
		t.Fatal(err)
	}
	if boundary.LichunYear != 2026 || !boundary.BoundaryDay {
		t.Fatalf("立春 day = %+v, want 2026 on boundary", boundary)
	}
	if boundary.Start != lichun {
		t.Fatalf("start = %s, want %s", boundary.Start, lichun)
	}

	after, err := ResolveLiuNianTime(lichun.AddDays(1))
	if err != nil {
		t.Fatal(err)
	}
	if after.LichunYear != 2026 || after.BoundaryDay {
		t.Fatalf("day after 立春 = %+v, want 2026 without boundary", after)
	}
}

func TestResolveLiuYueTimeJieqiBoundary(t *testing.T) {
	lichun := tianwen.SolarTermCivilDate(2026, 315)

	before, err := ResolveLiuYueTime(lichun.AddDays(-1))
	if err != nil {
		t.Fatal(err)
	}
	if before.JieqiMonth != 12 || before.BoundaryDay {
		t.Fatalf("day before 立春 = %+v, want 丑月 (12) without boundary", before)
	}

	boundary, err := ResolveLiuYueTime(lichun)
	if err != nil {
		t.Fatal(err)
	}
	if boundary.JieqiMonth != 1 || !boundary.BoundaryDay {
		t.Fatalf("立春 day = %+v, want 寅月 (1) on boundary", boundary)
	}
	if boundary.Start != lichun {
		t.Fatalf("start = %s, want %s", boundary.Start, lichun)
	}

	after, err := ResolveLiuYueTime(lichun.AddDays(1))
	if err != nil {
		t.Fatal(err)
	}
	if after.JieqiMonth != 1 || after.BoundaryDay {
		t.Fatalf("day after 立春 = %+v, want 寅月 without boundary", after)
	}
}

func TestNextLiuNianTimeAdvancesToNextLichun(t *testing.T) {
	current, err := ResolveLiuNianTime(tianwen.NewCivilDate(2026, 6, 15))
	if err != nil {
		t.Fatal(err)
	}
	next, err := NextLiuNianTime(current)
	if err != nil {
		t.Fatal(err)
	}
	if next.LichunYear != current.LichunYear+1 {
		t.Fatalf("next year = %d, want %d", next.LichunYear, current.LichunYear+1)
	}
	if next.Start != current.End {
		t.Fatalf("next start = %s, want %s", next.Start, current.End)
	}
}

func TestNextLiuYueTimeAdvancesOneJieqiMonth(t *testing.T) {
	current, err := ResolveLiuYueTime(tianwen.NewCivilDate(2026, 6, 15))
	if err != nil {
		t.Fatal(err)
	}
	next, err := NextLiuYueTime(current)
	if err != nil {
		t.Fatal(err)
	}
	if next.JieqiMonth != current.JieqiMonth%12+1 {
		t.Fatalf("next month = %d, want %d", next.JieqiMonth, current.JieqiMonth%12+1)
	}
	if next.Start != current.End {
		t.Fatalf("next start = %s, want %s", next.Start, current.End)
	}
}

func TestResolveLiuNianTimeAtUsesExactLichun(t *testing.T) {
	loc, err := time.LoadLocation("Asia/Shanghai")
	if err != nil {
		t.Fatal(err)
	}
	lichun := tianwen.SolarTermTime(2024, 315)

	before, err := ResolveLiuNianTimeAt(lichun.Add(-time.Second), loc)
	if err != nil {
		t.Fatal(err)
	}
	if before.LichunYear != 2023 {
		t.Fatalf("before 立春 year = %d, want 2023", before.LichunYear)
	}

	after, err := ResolveLiuNianTimeAt(lichun, loc)
	if err != nil {
		t.Fatal(err)
	}
	if after.LichunYear != 2024 {
		t.Fatalf("at 立春 year = %d, want 2024", after.LichunYear)
	}
	if want := tianwen.CivilDateOf(lichun.In(loc)); after.Start != want {
		t.Fatalf("start = %s, want %s", after.Start, want)
	}
}

func TestResolveLiuYueTimeAtUsesExactJie(t *testing.T) {
	loc, err := time.LoadLocation("Asia/Shanghai")
	if err != nil {
		t.Fatal(err)
	}
	jingzhe := tianwen.SolarTermTime(2024, 345)

	before, err := ResolveLiuYueTimeAt(jingzhe.Add(-time.Second), loc)
	if err != nil {
		t.Fatal(err)
	}
	after, err := ResolveLiuYueTimeAt(jingzhe, loc)
	if err != nil {
		t.Fatal(err)
	}
	if before.JieqiMonth != 1 || after.JieqiMonth != 2 {
		t.Fatalf("jieqi months = %d/%d, want 1/2", before.JieqiMonth, after.JieqiMonth)
	}
	if want := tianwen.CivilDateOf(jingzhe.In(loc)); after.Start != want {
		t.Fatalf("start = %s, want %s", after.Start, want)
	}
}

func TestResolveLiuRiTimeAtUsesStandardDateAcrossDST(t *testing.T) {
	ny, err := time.LoadLocation("America/New_York")
	if err != nil {
		t.Fatal(err)
	}
	// 00:30 EDT is 23:30 EST on the previous day.
	at := time.Date(2024, 7, 1, 4, 30, 0, 0, time.UTC)
	got, err := ResolveLiuRiTimeAt(at, ny, -74)
	if err != nil {
		t.Fatal(err)
	}
	standard := tianwen.StandardDateAt(at, ny)
	last := tianwen.ApparentSolarCivil(at, ny, -74)
	lastDate := tianwen.CivilDateOf(last)
	if got.LocalDate != standard {
		t.Fatalf("local date = %s, want standard date %s", got.LocalDate, standard)
	}
	if got.Date != lastDate {
		t.Fatalf("ri date = %s, want LAST date %s", got.Date, lastDate)
	}
	zhu := tianwen.RiZhu(lastDate.Time())
	if got.RiGan != zhu.Gan || got.RiZhi != zhu.Zhi {
		t.Fatalf("ri zhu = %s%s, want %s%s", got.RiGan, got.RiZhi, zhu.Gan, zhu.Zhi)
	}
}

func TestComputeChartAtUsesStandardTimeAndLast(t *testing.T) {
	ny, err := time.LoadLocation("America/New_York")
	if err != nil {
		t.Fatal(err)
	}
	at := time.Date(2024, 7, 1, 4, 30, 0, 0, time.UTC)
	standard := tianwen.StandardLocalTime(at, ny)
	last := tianwen.ApparentSolarFromStandard(standard, -74)
	want := ComputeChart(standard, last, ganzhi.Male)
	got := ComputeChartAt(standard, -74, ganzhi.Male)
	if !reflect.DeepEqual(got, want) {
		t.Fatalf("ComputeChartAt differs from standard/LAST composition")
	}
}
