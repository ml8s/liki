package tianwen

import (
	"testing"
	"time"
)

func TestCivilDateString(t *testing.T) {
	d := NewCivilDate(2026, 6, 1)
	if got := d.String(); got != "2026-06-01" {
		t.Fatalf("String() = %q", got)
	}
}

func TestCivilDateValidate(t *testing.T) {
	valid := []CivilDate{
		NewCivilDate(2024, 2, 29),
		NewCivilDate(2026, 6, 1),
		NewCivilDate(1, 1, 1),
	}
	for _, d := range valid {
		if err := d.Validate(); err != nil {
			t.Errorf("Validate(%s) = %v, want nil", d, err)
		}
	}

	invalid := []CivilDate{
		NewCivilDate(2026, 2, 30),
		NewCivilDate(2026, 13, 1),
		NewCivilDate(2026, 0, 1),
		NewCivilDate(2026, 1, 0),
		NewCivilDate(0, 1, 1),
	}
	for _, d := range invalid {
		if err := d.Validate(); err == nil {
			t.Errorf("Validate(%s) = nil, want error", d)
		}
	}
}

func TestCivilDateAddDays(t *testing.T) {
	cases := []struct {
		start CivilDate
		days  int
		want  CivilDate
	}{
		{NewCivilDate(2026, 6, 1), 1, NewCivilDate(2026, 6, 2)},
		{NewCivilDate(2026, 6, 30), 1, NewCivilDate(2026, 7, 1)},
		{NewCivilDate(2026, 12, 31), 1, NewCivilDate(2027, 1, 1)},
		{NewCivilDate(2024, 2, 28), 1, NewCivilDate(2024, 2, 29)},
		{NewCivilDate(2024, 2, 29), 1, NewCivilDate(2024, 3, 1)},
		{NewCivilDate(2026, 3, 1), -1, NewCivilDate(2026, 2, 28)},
	}
	for _, tc := range cases {
		if got := tc.start.AddDays(tc.days); got != tc.want {
			t.Errorf("%s AddDays(%d) = %s, want %s", tc.start, tc.days, got, tc.want)
		}
	}
}

func TestCivilDateTimeIsNoonCST(t *testing.T) {
	d := NewCivilDate(2026, 6, 1)
	got := d.Time()

	if got.Hour() != 12 || got.Minute() != 0 || got.Second() != 0 {
		t.Fatalf("Time() = %s, want 12:00:00", got.Format(time.RFC3339))
	}
	if year, month, day := got.Date(); year != 2026 || month != time.June || day != 1 {
		t.Fatalf("Time() date = %04d-%02d-%02d", year, month, day)
	}
	if _, offset := got.Zone(); offset != 8*3600 {
		t.Fatalf("Time() offset = %d, want CST", offset)
	}
}

func TestCivilDateOfUsesInputLocation(t *testing.T) {
	loc, err := time.LoadLocation("America/New_York")
	if err != nil {
		t.Fatal(err)
	}
	at := time.Date(2026, 6, 1, 23, 30, 0, 0, loc)
	if got := CivilDateOf(at); got != (CivilDate{Year: 2026, Month: 6, Day: 1}) {
		t.Fatalf("CivilDateOf = %s, want 2026-06-01", got)
	}
}

func TestCivilDateOrdering(t *testing.T) {
	earlier := NewCivilDate(2026, 6, 1)
	later := NewCivilDate(2026, 6, 2)

	if !earlier.Before(later) || earlier.After(later) {
		t.Fatalf("ordering mismatch for %s and %s", earlier, later)
	}
	if later.Before(earlier) || !later.After(earlier) {
		t.Fatalf("ordering mismatch for %s and %s", later, earlier)
	}
	if (CivilDate{}).IsZero() != true {
		t.Fatal("zero CivilDate must report IsZero")
	}
	if earlier.IsZero() {
		t.Fatal("non-zero CivilDate reported IsZero")
	}
}
