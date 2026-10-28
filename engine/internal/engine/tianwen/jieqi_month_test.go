package tianwen

import "testing"

func TestJieqiMonthOfLichunBoundary(t *testing.T) {
	lichun := SolarTermCivilDate(2026, 315)
	if got := lichun.String(); got != "2026-02-04" {
		t.Fatalf("立春 date = %s, want 2026-02-04", got)
	}

	before, err := JieqiMonthOf(lichun.AddDays(-1))
	if err != nil {
		t.Fatal(err)
	}
	if before.Index != 11 || before.BoundaryDay {
		t.Fatalf("day before 立春 = %+v, want 丑月 (11) without boundary", before)
	}

	boundary, err := JieqiMonthOf(lichun)
	if err != nil {
		t.Fatal(err)
	}
	if boundary.Index != 0 || !boundary.BoundaryDay {
		t.Fatalf("立春 day = %+v, want 寅月 (0) on boundary", boundary)
	}
	if boundary.Start != lichun {
		t.Fatalf("start = %s, want %s", boundary.Start, lichun)
	}

	after, err := JieqiMonthOf(lichun.AddDays(1))
	if err != nil {
		t.Fatal(err)
	}
	if after.Index != 0 || after.BoundaryDay {
		t.Fatalf("day after 立春 = %+v, want 寅月 without boundary", after)
	}
}

func TestJieqiMonthOfEndsAtNextJie(t *testing.T) {
	date := NewCivilDate(2026, 6, 15)
	window, err := JieqiMonthOf(date)
	if err != nil {
		t.Fatal(err)
	}
	if !window.Start.Before(window.End) {
		t.Fatalf("window [%s, %s) is not ordered", window.Start, window.End)
	}
	next, err := JieqiMonthOf(window.End)
	if err != nil {
		t.Fatal(err)
	}
	if next.Index != (window.Index+1)%12 {
		t.Fatalf("next month index = %d, want %d", next.Index, (window.Index+1)%12)
	}
	if !next.BoundaryDay {
		t.Fatal("next 节 date must be a boundary day")
	}
}

func TestYueZhuFromYearGanMatchesYueZhu(t *testing.T) {
	date := NewCivilDate(2026, 6, 15)
	window, err := JieqiMonthOf(date)
	if err != nil {
		t.Fatal(err)
	}
	nian := NianZhu(date.Time())
	want := YueZhu(date.Time())
	got := YueZhuFromYearGan(nian.Gan, window.Index)
	if got != want {
		t.Fatalf("YueZhuFromYearGan = %+v, want %+v", got, want)
	}
}
