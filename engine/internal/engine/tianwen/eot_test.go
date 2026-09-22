package tianwen

import (
	"math"
	"testing"
)

// cosDeg converts degrees to radians and applies cosine.
func cosDeg(d float64) float64 { return math.Cos(d * math.Pi / 180.0) }

// Meeus full EoT (Astronomical Algorithms, Ch.28)
func meeusEoT(jd float64) float64 {
	T := (jd - 2451545.0) / 36525.0

	L0 := 280.46646 + 36000.76983*T + 0.0003032*T*T
	M := 357.52911 + 35999.05029*T - 0.0001537*T*T
	e := 0.016708634 - 0.000042037*T - 0.0000001267*T*T
	eps0 := 23.0 + 26.0/60.0 + 21.448/3600.0 -
		46.8150/3600.0*T - 0.00059/3600.0*T*T + 0.001813/3600.0*T*T

	y := math.Pow(math.Tan((eps0/2)*(math.Pi/180.0)), 2)

	EoT := 4 * (y*sinDeg(2*L0) -
		2*e*sinDeg(M) +
		4*e*y*sinDeg(M)*cosDeg(2*L0) -
		0.5*y*y*sinDeg(4*L0) -
		1.25*e*e*sinDeg(2*M)) * (180.0 / math.Pi)

	return EoT
}

func julianDate(year, month, day int) float64 {
	if month <= 2 {
		year--
		month += 12
	}
	A := year / 100
	B := 2 - A + A/4
	return float64(int(365.25*float64(year+4716))) + float64(int(30.6001*float64(month+1))) +
		float64(day) + float64(B) - 1524.5
}

func TestMeeusEoTAnchors(t *testing.T) {
	anchors := []struct {
		name               string
		year, month, day   int
		wantMin, tolerance float64
	}{
		{"Feb 14 1992 (minimum)", 1992, 2, 14, -14.27, 0.1},
		{"Nov 3 1992 (maximum)", 1992, 11, 3, 16.42, 0.1},
		{"Sep 1 1992 (zero crossing)", 1992, 9, 1, -0.15, 0.5},
		{"Apr 15 1992 (zero crossing)", 1992, 4, 15, -0.08, 0.5},
		{"Jul 26 1992", 1992, 7, 26, -6.48, 0.1},
		{"May 15 1992", 1992, 5, 15, 3.66, 0.1},
		{"Aug 26 1981 (critical case)", 1981, 8, 26, -1.93, 0.1},
	}

	for _, a := range anchors {
		jd := julianDate(a.year, a.month, a.day)
		got := meeusEoT(jd)
		if math.Abs(got-a.wantMin) > a.tolerance {
			t.Errorf("%s: got %.4f min, want %.2f ± %.2f", a.name, got, a.wantMin, a.tolerance)
		} else {
			t.Logf("✓ %s: %.4f min (want %.2f)", a.name, got, a.wantMin)
		}
	}
}

func TestMeeusVsSpencer(t *testing.T) {
	maxDiff := 0.0
	worstDay := 0
	for doy := 1; doy <= 365; doy++ {
		B := 360.0 * float64(doy-81) / 365.0
		BRad := B * math.Pi / 180.0
		spencer := 9.87*math.Sin(2*BRad) - 7.53*math.Cos(BRad) - 1.5*math.Sin(BRad)

		jd := julianDate(1992, 1, doy) // doy -> actual date in 1992
		meeus := meeusEoT(jd)

		diff := math.Abs(spencer - meeus)
		if diff > maxDiff {
			maxDiff = diff
			worstDay = doy
		}
	}
	t.Logf("Max diff Spencer vs Meeus: %.4f min (day %d)", maxDiff, worstDay)
	if maxDiff > 1.5 {
		t.Errorf("Unexpected: max diff %.4f min exceeds 1.5 min", maxDiff)
	}
}
