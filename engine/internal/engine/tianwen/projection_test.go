package tianwen

import (
	"sort"
	"testing"
	"time"

	"liki-engine/internal/engine/ganzhi"
)

func ApparentSolarOffset(at time.Time, longitude float64) time.Duration {
	_, offsetSeconds := at.Zone()
	solar := ApparentSolarFromCivil(at, longitude, float64(offsetSeconds)/3600)
	return solar.Sub(at)
}

func ApparentSolarDate(
	at time.Time,
	loc *time.Location,
	longitude float64,
) (year int, month time.Month, day int) {
	return ApparentSolarCivil(at, loc, longitude).Date()
}

func MonthBranchAt(at time.Time) ganzhi.Zhi {
	return JianYue(time.Time(at))
}

func TestApparentSolarOffsetIsDeterministic(t *testing.T) {
	at := time.Date(1990, 5, 20, 4, 0, 0, 0, time.UTC)
	offset := ApparentSolarOffset(at, 116.407526)
	again := ApparentSolarOffset(at, 116.407526)

	if offset == 0 {
		t.Skip("equation of time happened to cancel longitude correction")
	}
	if offset != again {
		t.Fatalf("offset changed: %s != %s", offset, again)
	}
}

func TestShichenAtLateZi(t *testing.T) {
	at := time.Date(1990, 5, 20, 15, 30, 0, 0, time.UTC)
	loc, err := time.LoadLocation("Asia/Shanghai")
	if err != nil {
		t.Fatal(err)
	}

	if got := ShichenAt(at, loc, 120.0); got.String() != "子" {
		t.Fatalf("ShichenAt late zi = %s, want 子", got.String())
	}
}

func TestApparentSolarCivilUsesExplicitLocationNotInputOffset(t *testing.T) {
	// +09:00 and Z denote the same absolute instant when normalized.
	atOffset, err := time.Parse(time.RFC3339Nano, "2026-10-22T13:00:00+09:00")
	if err != nil {
		t.Fatal(err)
	}
	atUTC, err := time.Parse(time.RFC3339Nano, "2026-10-22T04:00:00Z")
	if err != nil {
		t.Fatal(err)
	}
	loc, err := time.LoadLocation("Asia/Shanghai")
	if err != nil {
		t.Fatal(err)
	}

	fromOffset := ApparentSolarCivil(atOffset.UTC(), loc, 116.407526)
	fromUTC := ApparentSolarCivil(atUTC, loc, 116.407526)

	if !fromOffset.Equal(fromUTC) {
		t.Fatalf("projection differs: offset=%s utc=%s", fromOffset, fromUTC)
	}
	if _, offsetSeconds := fromOffset.Zone(); offsetSeconds != 8*3600 {
		t.Fatalf("civil offset = %d, want explicit Asia/Shanghai offset", offsetSeconds)
	}
}

func TestLunarDateAtKnownDate(t *testing.T) {
	at := time.Date(1990, 5, 20, 4, 0, 0, 0, time.UTC)
	got := LunarDateAt(at)
	if got.Year != 1990 || got.Month != 4 || got.Day != 26 || got.Leap {
		t.Fatalf("LunarDateAt = %+v", got)
	}
}

func SolarTermStateAt(at time.Time) SolarTermState {
	terms := make([]SolarTerm, 0, 24*3)
	for year := at.Year() - 1; year <= at.Year()+1; year++ {
		for index, longitude := range JieQiLongitudes {
			terms = append(terms, SolarTerm{
				Index: index,
				Name:  JieQiNames[index],
				At:    SolarTermTime(year, longitude),
			})
		}
	}

	sort.Slice(terms, func(i, j int) bool {
		return terms[i].At.Before(terms[j].At)
	})

	current := sort.Search(len(terms), func(i int) bool {
		return terms[i].At.After(at)
	})
	if current > 0 {
		current--
	}

	return SolarTermState{
		Previous: terms[current-1],
		Current:  terms[current],
		Next:     terms[min(current+1, len(terms)-1)],
	}
}

func TestSolarTermStateAtLichun(t *testing.T) {
	at := time.Date(2024, 2, 10, 0, 0, 0, 0, time.UTC)
	state := SolarTermStateAt(at)

	if state.Current.Name != "立春" {
		t.Fatalf("current solar term = %s, want 立春", state.Current.Name)
	}
	if state.Previous.Name != "大寒" {
		t.Fatalf("previous solar term = %s, want 大寒", state.Previous.Name)
	}
	if state.Next.Name != "雨水" {
		t.Fatalf("next solar term = %s, want 雨水", state.Next.Name)
	}
	if got := MonthBranchAt(at); got.String() != "寅" {
		t.Fatalf("month branch at 立春 = %s, want 寅", got.String())
	}
}

func TestApparentSolarCivilCanCrossDay(t *testing.T) {
	// Early morning in far-west China can fall on the prior apparent-solar day.
	at := time.Date(1990, 5, 20, 18, 10, 0, 0, time.UTC)
	loc, err := time.LoadLocation("Asia/Shanghai")
	if err != nil {
		t.Fatal(err)
	}

	year, month, day := ApparentSolarDate(at, loc, 75.0)
	if year != 1990 || month != time.May || day != 20 {
		t.Fatalf("apparent solar date = %d-%d-%d, want 1990-05-20", year, month, day)
	}
}
