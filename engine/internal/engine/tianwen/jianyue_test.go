package tianwen

import (
	"testing"
	"time"

	"liki-engine/internal/engine/ganzhi"
)

func jianYueReference(gt time.Time) ganzhi.Zhi {
	lon := solarLongitude(gt)
	for i := 0; i < 12; i++ {
		if angleInRange(lon, solarTermLongitudes[i], solarTermLongitudes[(i+1)%12]) {
			return ganzhi.Zhi((i+2)%12 + 1)
		}
	}
	return ganzhi.ZhiYin
}

func TestJianYueTermIntervalsMatchSolarLongitude(t *testing.T) {
	for _, year := range []int{1900, 1984, 2023, 2024, 2025, 2026, 2027, 2100} {
		for _, longitude := range solarTermLongitudes {
			boundary := SolarTermTime(year, longitude)
			for _, offset := range []time.Duration{
				-15 * time.Minute, -3 * time.Second, -time.Second,
				time.Second, 3 * time.Second, 15 * time.Minute,
			} {
				at := boundary.Add(offset)
				got, want := JianYue(at), jianYueReference(at)
				if got != want {
					t.Errorf("JianYue(%s, longitude %.0f, offset %s) = %v, want %v lon=%.12f", at, longitude, offset, got, want, solarLongitude(at))
				}
			}
		}

		for at := time.Date(year, 1, 1, 12, 0, 0, 0, time.UTC); at.Year() == year; at = at.AddDate(0, 0, 1) {
			got, want := JianYue(at), jianYueReference(at)
			if got != want {
				t.Fatalf("JianYue(%s) = %v, want %v", at, got, want)
			}
		}
	}
}
