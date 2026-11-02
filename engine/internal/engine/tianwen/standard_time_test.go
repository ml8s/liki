package tianwen

import (
	"fmt"
	"sort"
	"testing"
	"time"
)

func TestStandardLocalTimeRemovesDST(t *testing.T) {
	cases := []struct {
		name string
		zone string
		at   time.Time
		want time.Time
	}{
		{
			name: "new_york_summer",
			zone: "America/New_York",
			at:   time.Date(2024, 7, 1, 4, 30, 0, 0, time.UTC),
			want: time.Date(2024, 6, 30, 23, 30, 0, 0, time.UTC),
		},
		{
			name: "london_summer",
			zone: "Europe/London",
			at:   time.Date(2024, 7, 1, 12, 0, 0, 0, time.UTC),
			want: time.Date(2024, 7, 1, 12, 0, 0, 0, time.UTC),
		},
		{
			name: "sydney_summer",
			zone: "Australia/Sydney",
			at:   time.Date(2024, 1, 1, 12, 0, 0, 0, time.UTC),
			want: time.Date(2024, 1, 1, 22, 0, 0, 0, time.UTC),
		},
		{
			name: "shanghai_1990_dst",
			zone: "Asia/Shanghai",
			at:   time.Date(1990, 7, 1, 4, 0, 0, 0, time.UTC),
			want: time.Date(1990, 7, 1, 12, 0, 0, 0, time.UTC),
		},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			loc, err := time.LoadLocation(tc.zone)
			if err != nil {
				t.Fatal(err)
			}
			got := StandardLocalTime(tc.at, loc)
			if got.Year() != tc.want.Year() || got.Month() != tc.want.Month() ||
				got.Day() != tc.want.Day() || got.Hour() != tc.want.Hour() || got.Minute() != tc.want.Minute() {
				t.Fatalf("standard local = %s, want %s", got.Format("2006-01-02 15:04"), tc.want.Format("2006-01-02 15:04"))
			}
		})
	}
}

func TestStandardDateAtNoDST(t *testing.T) {
	ny, err := time.LoadLocation("America/New_York")
	if err != nil {
		t.Fatal(err)
	}
	// 00:30 EDT is 23:30 EST on the previous date.
	at := time.Date(2024, 7, 1, 4, 30, 0, 0, time.UTC)
	if got, want := StandardDateAt(at, ny), NewCivilDate(2024, 6, 30); got != want {
		t.Fatalf("standard date = %s, want %s", got, want)
	}
}

func civilInstants(year int, month time.Month, day, hour, minute, second, nanosecond int, zone *time.Location) ([]time.Time, error) {
	probe := time.Date(year, month, day, hour, minute, second, nanosecond, zone)
	offsets := []int{}
	seen := make(map[int]bool)
	for _, candidate := range []time.Time{
		probe.Add(-36 * time.Hour), probe.Add(-12 * time.Hour), probe,
		probe.Add(12 * time.Hour), probe.Add(36 * time.Hour),
	} {
		_, offset := candidate.In(zone).Zone()
		if seen[offset] {
			continue
		}
		seen[offset] = true
		fixed := time.FixedZone("", offset)
		wall := time.Date(year, month, day, hour, minute, second, nanosecond, fixed)
		back := wall.In(zone)
		if back.Year() == year && back.Month() == month && back.Day() == day &&
			back.Hour() == hour && back.Minute() == minute && back.Second() == second {
			offsets = append(offsets, offset)
		}
	}
	if len(offsets) == 0 {
		return nil, fmt.Errorf("the recorded wall clock does not exist in %s (DST gap); confirm the time", zone)
	}
	sort.Slice(offsets, func(i, j int) bool { return offsets[i] < offsets[j] })
	instants := make([]time.Time, 0, len(offsets))
	for _, offset := range offsets {
		fixed := time.FixedZone("", offset)
		instants = append(instants, time.Date(year, month, day, hour, minute, second, nanosecond, fixed).UTC())
	}
	sort.Slice(instants, func(i, j int) bool { return instants[i].Before(instants[j]) })
	return instants, nil
}

func TestCivilInstantsFoldAndGap(t *testing.T) {
	ny, err := time.LoadLocation("America/New_York")
	if err != nil {
		t.Fatal(err)
	}

	fold, err := civilInstants(2024, 11, 3, 1, 30, 0, 0, ny)
	if err != nil {
		t.Fatal(err)
	}
	if len(fold) != 2 {
		t.Fatalf("fold candidates = %d, want 2", len(fold))
	}
	if want := time.Date(2024, 11, 3, 5, 30, 0, 0, time.UTC); !fold[0].Equal(want) {
		t.Fatalf("earlier = %s, want %s", fold[0], want)
	}
	if want := time.Date(2024, 11, 3, 6, 30, 0, 0, time.UTC); !fold[1].Equal(want) {
		t.Fatalf("later = %s, want %s", fold[1], want)
	}

	if _, err := civilInstants(2024, 3, 10, 2, 30, 0, 0, ny); err == nil {
		t.Fatal("gap civil time should fail")
	}
}

func TestApparentSolarFromStandardCentralMeridian(t *testing.T) {
	cases := []struct {
		name   string
		offset time.Duration
		lon    float64
	}{
		{"utc+5:30", 5*time.Hour + 30*time.Minute, 82.5},
		{"utc+5:45", 5*time.Hour + 45*time.Minute, 86.25},
		{"utc+8", 8 * time.Hour, 120},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			standard := time.Date(2024, 6, 1, 12, 0, 0, 0, time.FixedZone(tc.name, int(tc.offset.Seconds())))
			last := ApparentSolarFromStandard(standard, tc.lon)
			diff := last.Sub(time.Date(2024, 6, 1, 12, 0, 0, 0, last.Location()))
			if diff < -10*time.Minute || diff > 10*time.Minute {
				t.Fatalf("LAST = %s, want within 10m of 12:00", last.Format("15:04:05"))
			}
		})
	}

	standard := time.Date(2024, 6, 1, 12, 0, 0, 0, time.FixedZone("UTC+8", 8*3600))
	base := ApparentSolarFromStandard(standard, 120)
	east := ApparentSolarFromStandard(standard, 135)
	if diff := east.Sub(base); diff != time.Hour {
		t.Fatalf("15° longitude shift = %s, want 1h", diff)
	}
}
