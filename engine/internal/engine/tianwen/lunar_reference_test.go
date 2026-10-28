package tianwen

import (
	"testing"
	"time"
)

func mustLoadLocation(t *testing.T, name string) *time.Location {
	t.Helper()
	loc, err := time.LoadLocation(name)
	if err != nil {
		t.Fatalf("load %s: %v", name, err)
	}
	return loc
}

func TestLunarReferenceShanghaiMatchesLegacy(t *testing.T) {
	ref := ShanghaiLunarReference()
	start := time.Date(2024, 1, 1, 12, 0, 0, 0, cstLocation)
	end := time.Date(2027, 12, 31, 12, 0, 0, 0, cstLocation)

	for at := start; !at.After(end); at = at.AddDate(0, 0, 1) {
		got := ref.LunarDateAt(at)
		want := LunarDateAt(at)
		if got != want {
			t.Fatalf("%s: reference = %+v, legacy = %+v", at.Format("2006-01-02"), got, want)
		}

		date, err := ref.CivilDateFromLunar(got)
		if err != nil {
			t.Fatalf("%s: round trip %+v: %v", at.Format("2006-01-02"), got, err)
		}
		if date != CivilDateOf(at) {
			t.Fatalf("%s: round trip date = %s, want %s", at.Format("2006-01-02"), date, CivilDateOf(at))
		}
	}
}

func TestLunarReferenceLocalDateAssignment(t *testing.T) {
	ny := mustLoadLocation(t, "America/New_York")
	cst := ShanghaiLunarReference()
	local := NewLunarReference(ny)
	date := NewCivilDate(2024, 2, 9)

	if got, want := cst.SolarDateToLunar(date), (LunarDate{Year: 2023, Month: 12, Day: 30}); got != want {
		t.Fatalf("CST 2024-02-09 = %+v, want %+v", got, want)
	}
	if got, want := local.SolarDateToLunar(date), (LunarDate{Year: 2024, Month: 1, Day: 1}); got != want {
		t.Fatalf("NY 2024-02-09 = %+v, want %+v", got, want)
	}

	at := time.Date(2024, 2, 9, 23, 30, 0, 0, ny)
	if got, want := local.LunarDateAt(at), (LunarDate{Year: 2024, Month: 1, Day: 1}); got != want {
		t.Fatalf("NY 2024-02-09 23:30 = %+v, want %+v", got, want)
	}
	if got, want := cst.LunarDateAt(at), (LunarDate{Year: 2024, Month: 1, Day: 1}); got != want {
		t.Fatalf("CST projection of NY 2024-02-09 23:30 = %+v, want %+v", got, want)
	}
}

func TestLunarReferenceLeapMonthDependsOnReference(t *testing.T) {
	ny := mustLoadLocation(t, "America/New_York")
	cst := ShanghaiLunarReference()
	local := NewLunarReference(ny)

	cstLeap, err := cst.LunarMonthMeta(2025, 6, true)
	if err != nil {
		t.Fatalf("CST 2025 leap 6: %v", err)
	}
	if cstLeap.DayCount != 29 {
		t.Fatalf("CST 2025 leap 6 day count = %d, want 29", cstLeap.DayCount)
	}
	if _, err := cst.LunarMonthMeta(2025, 7, true); err == nil {
		t.Fatal("CST 2025 leap 7 should not exist")
	}

	localLeap, err := local.LunarMonthMeta(2025, 7, true)
	if err != nil {
		t.Fatalf("NY 2025 leap 7: %v", err)
	}
	if localLeap.DayCount != 29 {
		t.Fatalf("NY 2025 leap 7 day count = %d, want 29", localLeap.DayCount)
	}
	if _, err := local.LunarMonthMeta(2025, 6, true); err == nil {
		t.Fatal("NY 2025 leap 6 should not exist")
	}

	next, err := cst.NextLunarMonth(2025, 6, false)
	if err != nil {
		t.Fatalf("CST next after 2025-6: %v", err)
	}
	if !next.Leap || next.Month != 6 {
		t.Fatalf("CST next after 2025-6 = %+v, want leap 6", next)
	}

	if got, want := cst.SolarDateToLunar(NewCivilDate(2025, 7, 25)), (LunarDate{Year: 2025, Month: 6, Day: 1, Leap: true}); got != want {
		t.Fatalf("CST 2025-07-25 = %+v, want %+v", got, want)
	}
	if got, want := local.SolarDateToLunar(NewCivilDate(2025, 7, 25)), (LunarDate{Year: 2025, Month: 7, Day: 2}); got != want {
		t.Fatalf("NY 2025-07-25 = %+v, want %+v", got, want)
	}
}

func TestLunarReferenceLocalRoundTrip(t *testing.T) {
	ny := mustLoadLocation(t, "America/New_York")
	ref := NewLunarReference(ny)
	start := time.Date(2024, 1, 1, 12, 0, 0, 0, ny)
	end := time.Date(2025, 12, 31, 12, 0, 0, 0, ny)

	for at := start; !at.After(end); at = at.AddDate(0, 0, 1) {
		lunar := ref.LunarDateAt(at)
		date, err := ref.CivilDateFromLunar(lunar)
		if err != nil {
			t.Fatalf("%s: %+v: %v", at.Format("2006-01-02"), lunar, err)
		}
		if date != CivilDateOf(at) {
			t.Fatalf("%s: %+v -> %s", at.Format("2006-01-02"), lunar, date)
		}
	}
}

func TestLunarReferenceDateSerialAcrossDST(t *testing.T) {
	ny := mustLoadLocation(t, "America/New_York")
	ref := NewLunarReference(ny)

	for _, date := range []CivilDate{
		NewCivilDate(2024, 3, 9),
		NewCivilDate(2024, 3, 10),
		NewCivilDate(2024, 3, 11),
		NewCivilDate(2024, 11, 2),
		NewCivilDate(2024, 11, 3),
		NewCivilDate(2024, 11, 4),
	} {
		if diff := ref.dateToJD(date.AddDays(1)) - ref.dateToJD(date); diff != 1 {
			t.Fatalf("%s -> next: date serial diff = %v, want 1", date, diff)
		}
	}
}

func TestLunarReferenceFixedLegacyWrappers(t *testing.T) {
	ref := fixedLunarReference(defaultTZ)
	if ref != ShanghaiLunarReference() {
		t.Fatalf("fixed UTC+8 reference = %p, want Shanghai", ref)
	}
	for year := 2024; year <= 2027; year++ {
		if got, want := getMonth11K(year, defaultTZ), ref.month11K(year); got != want {
			t.Fatalf("%d month 11 k = %v, want %v", year, got, want)
		}
	}
	start := ref.month11K(2024)
	end := ref.month11K(2025)
	if got, want := getLeapMonthK(start, end, defaultTZ), ref.leapMonthK(start, end); got != want {
		t.Fatalf("leap month k = %v, want %v", got, want)
	}
}

func TestLunarReferenceZhongQiDates(t *testing.T) {
	ref := ShanghaiLunarReference()
	got := ref.zhongQiDates(2024)
	want := []CivilDate{
		NewCivilDate(2024, 1, 20),
		NewCivilDate(2024, 2, 19),
		NewCivilDate(2024, 3, 20),
		NewCivilDate(2024, 4, 19),
		NewCivilDate(2024, 5, 20),
		NewCivilDate(2024, 6, 21),
		NewCivilDate(2024, 7, 22),
		NewCivilDate(2024, 8, 22),
		NewCivilDate(2024, 9, 22),
		NewCivilDate(2024, 10, 23),
		NewCivilDate(2024, 11, 22),
		NewCivilDate(2024, 12, 21),
		NewCivilDate(2025, 1, 20),
		NewCivilDate(2025, 2, 18),
		NewCivilDate(2025, 3, 20),
		NewCivilDate(2025, 4, 20),
		NewCivilDate(2025, 5, 21),
		NewCivilDate(2025, 6, 21),
		NewCivilDate(2025, 7, 22),
		NewCivilDate(2025, 8, 23),
		NewCivilDate(2025, 9, 23),
		NewCivilDate(2025, 10, 23),
		NewCivilDate(2025, 11, 22),
		NewCivilDate(2025, 12, 21),
	}
	if len(got) != len(want) {
		t.Fatalf("zhongqi count = %d, want %d", len(got), len(want))
	}
	for i := range want {
		if got[i] != want[i] {
			t.Fatalf("zhongqi[%d] = %s, want %s", i, got[i], want[i])
		}
	}
}

func TestLunarReferenceExactLeapMonths(t *testing.T) {
	ref := ShanghaiLunarReference()
	cases := []struct {
		year  int
		month int
	}{
		{2023, 2},
		{2025, 6},
		{2028, 5},
		{2031, 3},
		{2033, 11},
		{2036, 6},
	}
	for _, tc := range cases {
		meta, err := ref.LunarMonthMeta(tc.year, tc.month, true)
		if err != nil {
			t.Fatalf("%d leap %d: %v", tc.year, tc.month, err)
		}
		if meta.DayCount < 29 || meta.DayCount > 30 {
			t.Fatalf("%d leap %d day count = %d", tc.year, tc.month, meta.DayCount)
		}
		next, err := ref.NextLunarMonth(tc.year, tc.month, false)
		if err != nil {
			t.Fatalf("%d next after month %d: %v", tc.year, tc.month, err)
		}
		if !next.Leap || next.Month != tc.month {
			t.Fatalf("%d next after month %d = %+v, want leap %d", tc.year, tc.month, next, tc.month)
		}

		leapCount := 0
		for month := 1; month <= 12; month++ {
			if _, err := ref.LunarMonthMeta(tc.year, month, true); err == nil {
				leapCount++
				if month != tc.month {
					t.Fatalf("%d unexpected leap month %d", tc.year, month)
				}
			}
		}
		if leapCount != 1 {
			t.Fatalf("%d leap month count = %d, want 1", tc.year, leapCount)
		}
	}
}

func TestMonthHasZhongQiBoundaries(t *testing.T) {
	ref := ShanghaiLunarReference()
	zhongQi := ref.zhongQiDates(2024)

	if !monthHasZhongQi(NewCivilDate(2025, 6, 25), NewCivilDate(2025, 7, 25), zhongQi) {
		t.Fatal("2025 month 6 should contain 大暑 on 2025-07-22")
	}
	if monthHasZhongQi(NewCivilDate(2025, 7, 25), NewCivilDate(2025, 8, 23), zhongQi) {
		t.Fatal("2025 leap 6 must not contain 处暑 on its next-month boundary")
	}
}

func TestLunarReferenceFixedUTC8MatchesOfficial(t *testing.T) {
	fixedLoc := time.FixedZone("UTC+8", 8*3600)
	fixed := NewLunarReference(fixedLoc)
	official := ShanghaiLunarReference()

	for date := NewCivilDate(2024, 1, 1); !date.After(NewCivilDate(2027, 12, 31)); date = date.AddDays(1) {
		got := fixed.SolarDateToLunar(date)
		want := official.SolarDateToLunar(date)
		if got != want {
			t.Fatalf("%s: fixed UTC+8 = %+v, official = %+v", date, got, want)
		}
	}

	projector, err := NewLunarProjector(LunarBasisLocalAstronomical, fixedLoc)
	if err != nil {
		t.Fatal(err)
	}
	for _, at := range []time.Time{
		time.Date(2024, 2, 9, 23, 30, 0, 0, fixedLoc),
		time.Date(2025, 7, 25, 0, 0, 0, 0, fixedLoc),
		time.Date(2026, 12, 21, 12, 0, 0, 0, fixedLoc),
	} {
		if got, want := projector.LunarAt(at), LunarDateAt(at); got != want {
			t.Fatalf("%s: local_astronomical fixed UTC+8 = %+v, official = %+v", at, got, want)
		}
	}
}
