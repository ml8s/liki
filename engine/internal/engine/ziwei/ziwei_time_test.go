package ziwei

import (
	"testing"
	"time"

	"liki-engine/internal/engine/ganzhi"
	"liki-engine/internal/engine/tianwen"
)

func civilDateOfLunar(t *testing.T, lunar tianwen.LunarDate) tianwen.CivilDate {
	t.Helper()
	solar := tianwen.CivilTimeFromLunarDate(lunar)
	if solar.IsZero() {
		t.Fatalf("lunar %+v has no civil date", lunar)
	}
	return tianwen.CivilDateOf(solar)
}

func TestResolveLiuNianTimeLunarNewYearBoundary(t *testing.T) {
	start := civilDateOfLunar(t, tianwen.LunarDate{Year: 2026, Month: 1, Day: 1})

	before, err := ResolveLiuNianTime(start.AddDays(-1))
	if err != nil {
		t.Fatal(err)
	}
	if before.LunarYear != 2025 || before.BoundaryDay {
		t.Fatalf("day before 正月初一 = %+v, want 2025 without boundary", before)
	}

	boundary, err := ResolveLiuNianTime(start)
	if err != nil {
		t.Fatal(err)
	}
	if boundary.LunarYear != 2026 || !boundary.BoundaryDay {
		t.Fatalf("正月初一 = %+v, want 2026 on boundary", boundary)
	}
	if boundary.Start != start {
		t.Fatalf("start = %s, want %s", boundary.Start, start)
	}

	after, err := ResolveLiuNianTime(start.AddDays(1))
	if err != nil {
		t.Fatal(err)
	}
	if after.LunarYear != 2026 || after.BoundaryDay {
		t.Fatalf("day after 正月初一 = %+v, want 2026 without boundary", after)
	}
}

func TestResolveLiuYueTimeLeapMonth(t *testing.T) {
	target := tianwen.LunarDate{Year: 2025, Month: 6, Day: 15, Leap: true}
	date := civilDateOfLunar(t, target)

	resolved, err := ResolveLiuYueTime(date)
	if err != nil {
		t.Fatal(err)
	}
	if resolved.Lunar != target {
		t.Fatalf("lunar = %+v, want %+v", resolved.Lunar, target)
	}
	if !resolved.Lunar.Leap || resolved.Lunar.Month != 6 {
		t.Fatalf("leap month lost: %+v", resolved.Lunar)
	}
	if resolved.Start != civilDateOfLunar(t, tianwen.LunarDate{Year: 2025, Month: 6, Day: 1, Leap: true}) {
		t.Fatalf("start = %s", resolved.Start)
	}
	if resolved.End != civilDateOfLunar(t, tianwen.LunarDate{Year: 2025, Month: 7, Day: 1}) {
		t.Fatalf("end = %s", resolved.End)
	}
}

func TestNextLiuYueTimeCrossesLeapMonth(t *testing.T) {
	ordinary := civilDateOfLunar(t, tianwen.LunarDate{Year: 2025, Month: 6, Day: 15})
	current, err := ResolveLiuYueTime(ordinary)
	if err != nil {
		t.Fatal(err)
	}
	leap, err := NextLiuYueTime(current)
	if err != nil {
		t.Fatal(err)
	}
	if !leap.Lunar.Leap || leap.Lunar.Month != 6 {
		t.Fatalf("next month = %+v, want 闰六月", leap.Lunar)
	}

	seventh, err := NextLiuYueTime(leap)
	if err != nil {
		t.Fatal(err)
	}
	if seventh.Lunar.Leap || seventh.Lunar.Month != 7 {
		t.Fatalf("next month = %+v, want 七月", seventh.Lunar)
	}
}

func TestNextLiuRiTimeHandles29DayLeapMonth(t *testing.T) {
	endOfLeap := civilDateOfLunar(t, tianwen.LunarDate{Year: 2025, Month: 6, Day: 29, Leap: true})
	current, err := ResolveLiuRiTime(endOfLeap)
	if err != nil {
		t.Fatal(err)
	}
	next, err := NextLiuRiTime(current)
	if err != nil {
		t.Fatal(err)
	}
	want := tianwen.LunarDate{Year: 2025, Month: 7, Day: 1}
	if next.Lunar != want {
		t.Fatalf("next lunar = %+v, want %+v", next.Lunar, want)
	}
}

func TestResolveLiuRiTimeMatchesLunarProjection(t *testing.T) {
	date := tianwen.NewCivilDate(2026, 6, 1)
	got, err := ResolveLiuRiTime(date)
	if err != nil {
		t.Fatal(err)
	}
	want := tianwen.SolarDateToLunar(date.Year, date.Month, date.Day)
	if got.Lunar != want {
		t.Fatalf("lunar = %+v, want %+v", got.Lunar, want)
	}
}

func TestComputeChartAtDefaultsToLocalAstronomical(t *testing.T) {
	loc, err := time.LoadLocation("America/New_York")
	if err != nil {
		t.Fatal(err)
	}
	// 23:30 EDT on 2026-06-01 is 2026-06-02 in Beijing. The local astronomical
	// calendar assigns the new moon to the local date, so it differs from the
	// official UTC+8 table near this boundary.
	at := time.Date(2026, 6, 1, 23, 30, 0, 0, loc)
	chart, err := ComputeChartAt(tianwen.StandardLocalTime(at, loc), -74.0, ganzhi.Male)
	if err != nil {
		t.Fatal(err)
	}
	if chart.LunarMonth != 4 || chart.LunarDay != 17 || chart.BirthIsLeap {
		t.Fatalf("natal lunar = %d-%d leap=%v, want 4-17",
			chart.LunarMonth, chart.LunarDay, chart.BirthIsLeap)
	}
}

func TestLunarBasisSelectsZiweiProjection(t *testing.T) {
	loc, err := time.LoadLocation("America/New_York")
	if err != nil {
		t.Fatal(err)
	}
	at := time.Date(2024, 2, 9, 23, 30, 0, 0, loc)

	cases := []struct {
		basis     tianwen.LunarBasis
		wantMonth int
		wantDay   int
		wantYear  int
	}{
		{tianwen.LunarBasisLocalLabel, 12, 30, 2023},
		{tianwen.LunarBasisOfficial, 1, 1, 2024},
		{tianwen.LunarBasisLocalAstronomical, 1, 1, 2024},
	}
	for _, tc := range cases {
		t.Run(string(tc.basis), func(t *testing.T) {
			chart, err := ComputeChartAtBasis(tianwen.StandardLocalTime(at, loc), -74.0, ganzhi.Male, tc.basis)
			if err != nil {
				t.Fatal(err)
			}
			if chart.LunarMonth != tc.wantMonth || chart.LunarDay != tc.wantDay || chart.BirthYear != tc.wantYear {
				t.Fatalf("chart lunar = %d-%d birth_year=%d, want %d-%d %d",
					chart.LunarMonth, chart.LunarDay, chart.BirthYear, tc.wantMonth, tc.wantDay, tc.wantYear)
			}

			projector, err := tianwen.NewLunarProjector(tc.basis, loc)
			if err != nil {
				t.Fatal(err)
			}
			selector := projector.SelectorDate(at)
			liuri, err := ResolveLiuRiTimeAtWith(projector, at, loc, "America/New_York")
			if err != nil {
				t.Fatal(err)
			}
			if liuri.Lunar.Year != tc.wantYear || liuri.Lunar.Month != tc.wantMonth || liuri.Lunar.Day != tc.wantDay {
				t.Fatalf("liuri lunar = %+v, want %d-%d-%d", liuri.Lunar, tc.wantYear, tc.wantMonth, tc.wantDay)
			}

			liunian, err := ResolveLiuNianTimeWith(projector, selector)
			if err != nil {
				t.Fatal(err)
			}
			if liunian.LunarYear != tc.wantYear {
				t.Fatalf("liunian year = %d, want %d", liunian.LunarYear, tc.wantYear)
			}
			next, err := NextLiuNianTimeWith(projector, liunian)
			if err != nil {
				t.Fatal(err)
			}
			if next.LunarYear != tc.wantYear+1 || !next.BoundaryDay {
				t.Fatalf("next liunian = %+v, want %d on boundary", next, tc.wantYear+1)
			}
		})
	}
}

func TestComputeChartAtUsesStandardDateAcrossDST(t *testing.T) {
	ny, err := time.LoadLocation("America/New_York")
	if err != nil {
		t.Fatal(err)
	}
	// 00:30 EDT is 23:30 EST on the previous day.
	at := time.Date(2024, 7, 1, 4, 30, 0, 0, time.UTC)
	standard := tianwen.StandardLocalTime(at, ny)
	ref := tianwen.NewLunarReference(time.FixedZone("standard", -5*3600))
	want := ref.LunarDateAt(standard)

	chart, err := ComputeChartAt(standard, -74.0, ganzhi.Male)
	if err != nil {
		t.Fatal(err)
	}
	if chart.LunarMonth != want.Month || chart.LunarDay != want.Day || chart.BirthIsLeap != want.Leap {
		t.Fatalf("chart lunar = %d-%d leap=%v, want standard-date %d-%d leap=%v",
			chart.LunarMonth, chart.LunarDay, chart.BirthIsLeap,
			want.Month, want.Day, want.Leap)
	}
}

func TestResolveLiuRiTimeAtWithUsesStandardDate(t *testing.T) {
	ny, err := time.LoadLocation("America/New_York")
	if err != nil {
		t.Fatal(err)
	}
	at := time.Date(2024, 7, 1, 4, 30, 0, 0, time.UTC)
	standard := tianwen.StandardLocalTime(at, ny)
	projector, err := tianwen.NewLunarProjector(tianwen.LunarBasisLocalAstronomical, ny)
	if err != nil {
		t.Fatal(err)
	}
	got, err := ResolveLiuRiTimeAtWith(projector, standard, ny, "America/New_York")
	if err != nil {
		t.Fatal(err)
	}
	wantDate := tianwen.StandardDateAt(standard, ny)
	if got.LocalDate != wantDate || got.Date != wantDate {
		t.Fatalf("dates = %s/%s, want %s", got.LocalDate, got.Date, wantDate)
	}
	if want := projector.SolarDateToLunar(wantDate); got.Lunar != want {
		t.Fatalf("lunar = %+v, want %+v", got.Lunar, want)
	}
}
