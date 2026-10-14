package tianwen

import (
	"time"
)

// Timeset packs three calendar representations of a moment.
type Timeset struct {
	Gregorian           GregorianTime        `json:"gregorian"`
	Solar               SolarTime            `json:"solar"`
	Lunar               LunarTime            `json:"lunar"`
	HistoricalTimeHints []HistoricalTimeHint `json:"historical_time_hints,omitempty"`
}

type HistoricalTimeHint struct {
	Kind            string `json:"kind"`
	Year            int    `json:"year"`
	RecordedTime    string `json:"recorded_time"`
	StandardTime    string `json:"standard_time"`
	DaylightTime    string `json:"daylight_time"`
	RequiresConfirm bool   `json:"requires_confirm"`
}

// ComputeTimeset converts a Gregorian time into a full Timeset
// (Gregorian → Solar → Lunar). Longitude is in degrees and is taken as given:
// 0° is a legitimate longitude (Greenwich) and is never silently replaced with
// a default city's longitude.
func ComputeTimeset(gt GregorianTime, lon float64) Timeset {
	_, offset := gt.Time().Zone()
	tz := float64(offset) / 3600

	st := GregorianToSolar(gt.Time(), lon, tz)
	lt := SolarToLunar(GregorianTime(st.Time()))
	lt.Shichen = hourZhiFromSolarTime(st.Minutes())

	return Timeset{
		Gregorian:           gt,
		Solar:               st,
		Lunar:               lt,
		HistoricalTimeHints: chinaDSTHints(gt),
	}
}

// ChinaDSTHints exposes the 1986-1991 China daylight-saving ambiguity instead
// of silently interpreting a supplied wall-clock offset as standard time.
func chinaDSTHints(gt GregorianTime) []HistoricalTimeHint {
	t := gt.Time()
	year := t.Year()
	if year < 1986 || year > 1991 {
		return nil
	}
	month, day := t.Month(), t.Day()
	inWindow := year == 1986 &&
		(month > 5 || (month == 5 && day >= 5)) &&
		(month < 9 || (month == 9 && day <= 14))
	if year >= 1987 {
		inWindow = (month > 4 || (month == 4 && day >= 15)) &&
			(month < 9 || (month == 9 && day <= 15))
	}
	if !inWindow {
		return nil
	}
	_, offset := t.Zone()
	if offset != 8*3600 && offset != 9*3600 {
		return nil
	}
	standard := t.In(time.FixedZone("CST", 8*3600))
	daylight := t.In(time.FixedZone("CDT", 9*3600))
	return []HistoricalTimeHint{{
		Kind:            "china_dst",
		Year:            year,
		RecordedTime:    t.Format(time.RFC3339),
		StandardTime:    standard.Format(time.RFC3339),
		DaylightTime:    daylight.Format(time.RFC3339),
		RequiresConfirm: true,
	}}
}
