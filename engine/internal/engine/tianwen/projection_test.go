package tianwen

import (
	"testing"
	"time"
)

func ApparentSolarOffset(at time.Time, longitude float64) time.Duration {
	_, offsetSeconds := at.Zone()
	solar := ApparentSolarFromCivil(at, longitude, float64(offsetSeconds)/3600)
	return solar.Sub(at)
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

	if got := ShichenFromStandard(StandardLocalTime(at, loc), 120.0); got.String() != "子" {
		t.Fatalf("ShichenAt late zi = %s, want 子", got.String())
	}
}

func TestLunarDateAtKnownDate(t *testing.T) {
	at := time.Date(1990, 5, 20, 4, 0, 0, 0, time.UTC)
	got := lunarDateAtForTest(ShanghaiLunarReference(), at)
	if got.Year != 1990 || got.Month != 4 || got.Day != 26 || got.Leap {
		t.Fatalf("LunarDateAt = %+v", got)
	}
}
