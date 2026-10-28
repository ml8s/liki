package tianwen

import "time"

// StandardOffsetSeconds returns the zone's non-DST offset for the instant.
// It is the offset used by 命理 local time; DST is an input/display concern.
func StandardOffsetSeconds(at time.Time, loc *time.Location) int {
	if loc == nil {
		loc = time.UTC
	}
	local := at.In(loc)
	_, offset := local.Zone()
	if !local.IsDST() {
		return offset
	}
	for _, delta := range []time.Duration{
		-180 * 24 * time.Hour, 180 * 24 * time.Hour,
		-90 * 24 * time.Hour, 90 * 24 * time.Hour,
	} {
		probe := at.Add(delta).In(loc)
		if !probe.IsDST() {
			_, standard := probe.Zone()
			return standard
		}
	}
	return offset - 3600
}

// StandardLocalTime projects an instant to the zone's standard local time
// (the instant with DST removed). It is the 命理 local clock.
func StandardLocalTime(at time.Time, loc *time.Location) time.Time {
	if loc == nil {
		loc = time.UTC
	}
	return at.In(time.FixedZone("standard", StandardOffsetSeconds(at, loc)))
}

// StandardDateAt returns the standard local civil date of an instant.
func StandardDateAt(at time.Time, loc *time.Location) CivilDate {
	return CivilDateOf(StandardLocalTime(at, loc))
}

// StandardMidnight returns the standard local midnight of a date.
func StandardMidnight(date CivilDate, loc *time.Location) time.Time {
	if loc == nil {
		loc = time.UTC
	}
	noon := time.Date(date.Year, time.Month(date.Month), date.Day, 12, 0, 0, 0, loc)
	offset := StandardOffsetSeconds(noon, loc)
	return time.Date(date.Year, time.Month(date.Month), date.Day, 0, 0, 0, 0,
		time.FixedZone("standard", offset))
}

// StandardNoon returns the standard local noon of a date.
func StandardNoon(date CivilDate, loc *time.Location) time.Time {
	if loc == nil {
		loc = time.UTC
	}
	noon := time.Date(date.Year, time.Month(date.Month), date.Day, 12, 0, 0, 0, loc)
	offset := StandardOffsetSeconds(noon, loc)
	return time.Date(date.Year, time.Month(date.Month), date.Day, 12, 0, 0, 0,
		time.FixedZone("standard", offset))
}
