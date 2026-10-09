package tianwen

import (
	"encoding/json"
	"math"
	"strings"
	"time"

	"liki-engine/internal/engine/ganzhi"
)

// GregorianTime is the Gregorian calendar time.
type GregorianTime time.Time

func (g GregorianTime) Time() time.Time { return time.Time(g) }

// MarshalJSON marshals GregorianTime as RFC3339 string.
func (g GregorianTime) MarshalJSON() ([]byte, error) {
	return json.Marshal(time.Time(g))
}

// UnmarshalJSON unmarshals GregorianTime from YYYY-MM-DD or RFC3339 string.
func (g *GregorianTime) UnmarshalJSON(b []byte) error {
	s := strings.Trim(string(b), `"`)
	if len(s) == 10 {
		if t, err := time.Parse("2006-01-02", s); err == nil {
			*g = GregorianTime(t)
			return nil
		}
	}
	var t time.Time
	if err := json.Unmarshal(b, &t); err != nil {
		return err
	}
	*g = GregorianTime(t)
	return nil
}

// SolarTime is the absolute true-solar time for a birth moment.
type SolarTime time.Time

func (s SolarTime) Time() time.Time  { return time.Time(s) }
func (s SolarTime) Minutes() float64 { return float64(s.Time().Hour()*60 + s.Time().Minute()) }

// MarshalJSON marshals SolarTime as RFC3339 string.
func (s SolarTime) MarshalJSON() ([]byte, error) {
	return json.Marshal(time.Time(s))
}

// UnmarshalJSON unmarshals SolarTime from RFC3339 string.
func (s *SolarTime) UnmarshalJSON(b []byte) error {
	var t time.Time
	if err := json.Unmarshal(b, &t); err != nil {
		return err
	}
	*s = SolarTime(t)
	return nil
}

// equationOfTime computes the Equation of Time in minutes using the
// full Meeus algorithm (Astronomical Algorithms, Ch.28). Accuracy < 0.01 min.
func equationOfTime(t time.Time) float64 {
	year, month, day := t.Date()
	hour, minute, _ := t.Clock()

	// Julian Date at the given UT moment
	jd := julianDateFromTime(year, int(month), day, hour, minute)
	T := (jd - 2451545.0) / 36525.0

	// Mean longitude of Sun (degrees)
	L0 := 280.46646 + 36000.76983*T + 0.0003032*T*T
	// Mean anomaly of Sun (degrees)
	M := 357.52911 + 35999.05029*T - 0.0001537*T*T
	// Eccentricity of Earth's orbit
	e := 0.016708634 - 0.000042037*T - 0.0000001267*T*T
	// Mean obliquity of the ecliptic (degrees)
	eps0 := 23.0 + 26.0/60.0 + 21.448/3600.0 -
		46.8150/3600.0*T - 0.00059/3600.0*T*T + 0.001813/3600.0*T*T

	// y = tan²(ε/2)
	y := math.Pow(math.Tan((eps0/2)*(math.Pi/180.0)), 2)

	// Meeus eq. 28.3: result in radians
	E := y*sinDeg(2*L0) -
		2*e*sinDeg(M) +
		4*e*y*sinDeg(M)*cosDegrees(2*L0) -
		0.5*y*y*sinDeg(4*L0) -
		1.25*e*e*sinDeg(2*M)

	// Convert radians → degrees, then × 4 → minutes of time
	return 4 * E * (180.0 / math.Pi)
}

func cosDegrees(d float64) float64 { return math.Cos(d * math.Pi / 180.0) }

func julianDateFromTime(year, month, day, hour, minute int) float64 {
	if month <= 2 {
		year--
		month += 12
	}
	A := year / 100
	B := 2 - A + A/4
	jd := float64(int(365.25*float64(year+4716))) +
		float64(int(30.6001*float64(month+1))) +
		float64(day) + float64(B) - 1524.5
	return jd + (float64(hour)+float64(minute)/60.0)/24.0
}

// computeSolarTime returns true solar time in minutes and day offset.
// timezone is in hours (e.g. 8 for UTC+8).
func computeSolarTime(t time.Time, longitude, timezone float64) (float64, int) {
	hour, minute := t.Hour(), t.Minute()
	lst := float64(hour*60 + minute)
	lonOffset := 4.0 * (longitude - timezone*15)
	eot := equationOfTime(t)
	raw := lst + lonOffset + eot
	dayOffset := 0
	if raw < 0 {
		dayOffset = -1
	} else if raw >= 1440 {
		dayOffset = 1
	}
	ast := math.Mod(raw, 1440)
	if ast < 0 {
		ast += 1440
	}
	return ast, dayOffset
}

// GregorianToSolar returns the absolute true solar time as SolarTime.
// timezone is in hours (e.g. 8 for UTC+8), longitude is in degrees.
func GregorianToSolar(t time.Time, longitude, timezone float64) SolarTime {
	ast, dayOffset := computeSolarTime(t, longitude, timezone)
	loc := time.FixedZone("birth", int(timezone*3600))
	astHour := int(ast) / 60
	astMin := int(ast) % 60
	return SolarTime(time.Date(t.Year(), t.Month(), t.Day()+dayOffset, astHour, astMin, 0, 0, loc))
}

// HourBranchFromSolarTime converts solar time (minutes) to the earthly zhi of the hour.
func hourZhiFromSolarTime(astMinutes float64) ganzhi.Zhi {
	idx := (int(astMinutes+60) / 120) % 12
	return ganzhi.Zhi(idx + 1)
}
