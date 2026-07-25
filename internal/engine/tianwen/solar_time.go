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

// computeSolarTime returns true solar time in minutes and day offset.
// timezone is in hours (e.g. 8 for UTC+8).
func computeSolarTime(t time.Time, longitude, timezone float64) (float64, int) {
	year, month, day := t.Date()
	hour, minute := t.Hour(), t.Minute()
	lst := float64(hour*60 + minute)
	lonOffset := 4.0 * (longitude - timezone*15)
	n := dayOfYear(year, int(month), day)
	B := 360.0 * float64(n-81) / 365.0
	BRad := B * math.Pi / 180.0
	eot := 9.87*math.Sin(2*BRad) - 7.53*math.Cos(BRad) - 1.5*math.Sin(BRad)
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

// HourBranchFromSolarTime converts solar time (minutes) to the earthly branch of the hour.
func hourZhiFromSolarTime(astMinutes float64) ganzhi.Zhi {
	idx := (int(astMinutes+60) / 120) % 12
	return ganzhi.Zhi(idx + 1)
}

func dayOfYear(year, month, day int) int {
	daysBefore := []int{0, 31, 59, 90, 120, 151, 181, 212, 243, 273, 304, 334}
	n := daysBefore[month-1] + day
	if month > 2 && isLeapYear(year) {
		n++
	}
	return n
}

func isLeapYear(y int) bool { return y%4 == 0 && (y%100 != 0 || y%400 == 0) }
