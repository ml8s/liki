package tianwen

import (
	"math"
	"testing"
	"time"
)

func TestSolarTermTimeKnownInstants2024(t *testing.T) {
	cases := []struct {
		name string
		lon  float64
		want time.Time
	}{
		{"春分", 0, time.Date(2024, 3, 20, 3, 6, 24, 0, time.UTC)},
		{"夏至", 90, time.Date(2024, 6, 20, 20, 51, 0, 0, time.UTC)},
		{"秋分", 180, time.Date(2024, 9, 22, 12, 43, 40, 0, time.UTC)},
		{"冬至", 270, time.Date(2024, 12, 21, 9, 20, 34, 0, time.UTC)},
		{"立春", 315, time.Date(2024, 2, 4, 8, 26, 53, 0, time.UTC)},
	}
	const tolerance = 30 * time.Second
	for _, tc := range cases {
		got := SolarTermTime(2024, tc.lon).UTC()
		if diff := got.Sub(tc.want); diff < -tolerance || diff > tolerance {
			t.Fatalf("%s = %s, want %s (±%s)", tc.name, got.Format(time.RFC3339), tc.want.Format(time.RFC3339), tolerance)
		}
	}
}

func TestNewMoonKnownInstant2024(t *testing.T) {
	k := math.Round(approximateK(2024, 2, 10, 0))
	td := getNewMoonJDTD(k)
	got := jdToUTCTime(td - deltaTDays(td)).UTC()
	want := time.Date(2024, 2, 9, 22, 59, 0, 0, time.UTC)
	if diff := got.Sub(want); diff < -2*time.Minute || diff > 2*time.Minute {
		t.Fatalf("2024 正月朔 = %s, want %s (±2m)", got.Format(time.RFC3339), want.Format(time.RFC3339))
	}
}
