package tianwen

// Timeset packs three calendar representations of a moment.
type Timeset struct {
	Gregorian GregorianTime `json:"gregorian"`
	Solar     SolarTime     `json:"solar"`
	Lunar     LunarTime     `json:"lunar"`
}

// ComputeTimeset converts a Gregorian time into a full Timeset
// (Gregorian → Solar → Lunar). Longitude is in degrees.
func ComputeTimeset(gt GregorianTime, lon float64) Timeset {
	_, offset := gt.Time().Zone()
	tz := float64(offset) / 3600
	lon, _ = normGeo(lon, 0) // only normalize lon; tz from timestamp

	st := GregorianToSolar(gt.Time(), lon, tz)
	lt := SolarToLunar(GregorianTime(st.Time()))
	lt.Shichen = hourZhiFromSolarTime(st.Minutes())

	return Timeset{
		Gregorian: gt,
		Solar:     st,
		Lunar:     lt,
	}
}

// normGeo normalizes longitude and timezone defaults.
// Defaults: longitude 120 (Beijing), timezone UTC+8.
func normGeo(lon, tz float64) (float64, float64) {
	if lon == 0 {
		lon = 120
	}
	if tz == 0 {
		tz = 8
	}
	return lon, tz
}
