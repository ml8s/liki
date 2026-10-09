package tianwen

// Timeset packs three calendar representations of a moment.
type Timeset struct {
	Gregorian GregorianTime `json:"gregorian"`
	Solar     SolarTime     `json:"solar"`
	Lunar     LunarTime     `json:"lunar"`
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
		Gregorian: gt,
		Solar:     st,
		Lunar:     lt,
	}
}
