package tianwen

import (
	"errors"
	"time"

	"liki-engine/internal/engine/ganzhi"
)

var (
	ErrCivilTimeGap = errors.New("tianwen: civil date-time does not exist (DST gap)")
)

// ResolveCivilCandidates resolves a civil wall time in an IANA location.
// DST folds return both candidates; DST gaps return ErrCivilTimeGap.
// SolarTerm is one astronomical solar-term boundary.
type SolarTerm struct {
	Index int
	Name  string
	At    time.Time
}

// SolarTermState is the solar-term window containing an instant.
// Current is the term whose boundary started the current window; Next is the
// boundary that ends it.
type SolarTermState struct {
	Previous SolarTerm
	Current  SolarTerm
	Next     SolarTerm
}

// ApparentSolarOffset returns the correction from local civil time to local
// apparent solar time (LAST) for the given longitude.
func apparentSolarMinutes(t time.Time) float64 {
	return float64(t.Hour()*60 + t.Minute())
}

// ApparentSolarCivil returns the apparent solar wall-clock projection. The
// returned time.Time is a labeled read; it is not an absolute instant.
func ApparentSolarCivil(
	at time.Time,
	loc *time.Location,
	longitude float64,
) time.Time {
	civil := at.In(loc)
	_, offsetSeconds := civil.Zone()
	return ApparentSolarFromCivil(
		civil,
		longitude,
		float64(offsetSeconds)/3600,
	)
}

// ShichenAt derives the hour branch from apparent solar time.
func ShichenAt(
	at time.Time,
	loc *time.Location,
	longitude float64,
) ganzhi.Zhi {
	return ShichenOf(time.Time(ApparentSolarCivil(at, loc, longitude)))
}

// SolarTermStateAt returns the previous, current, and next solar-term
// boundaries around at. Boundary comparisons use absolute instants.
