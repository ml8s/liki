package tianwen

import (
	"errors"
	"sort"
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

// ApparentSolarAt returns the apparent solar wall-clock projection for an instant. The
// returned time.Time is a labeled read; it is not an absolute instant.
func ApparentSolarAt(
	at time.Time,
	loc *time.Location,
	longitude float64,
) time.Time {
	return ApparentSolarFromStandard(StandardLocalTime(at, loc), longitude)
}

// ShichenAt derives the hour branch from apparent solar time.
func ShichenAt(
	at time.Time,
	loc *time.Location,
	longitude float64,
) ganzhi.Zhi {
	return ShichenFromStandard(StandardLocalTime(at, loc), longitude)
}

// solarTermsAround returns the 24 solar terms for the years adjacent to at,
// sorted by boundary instant. The three-year window guarantees that the
// previous and next boundaries exist for any instant inside the middle year.
func solarTermsAround(at time.Time) []SolarTerm {
	terms := make([]SolarTerm, 0, 24*3)
	for year := at.Year() - 1; year <= at.Year()+1; year++ {
		for index, longitude := range JieQiLongitudes {
			terms = append(terms, SolarTerm{
				Index: index,
				Name:  JieQiNames[index],
				At:    SolarTermTime(year, longitude),
			})
		}
	}

	sort.Slice(terms, func(i, j int) bool {
		return terms[i].At.Before(terms[j].At)
	})
	return terms
}

// SolarTermStateAt returns the previous, current, and next solar-term
// boundaries around at. Boundary comparisons use absolute instants.
func SolarTermStateAt(at time.Time) SolarTermState {
	terms := solarTermsAround(at)

	current := sort.Search(len(terms), func(i int) bool {
		return terms[i].At.After(at)
	})
	if current > 0 {
		current--
	}
	// Guard the lower edge; a valid request always lands after the first term
	// in the constructed three-year window, but the zero time must not panic.
	if current == 0 {
		current = 1
	}

	return SolarTermState{
		Previous: terms[current-1],
		Current:  terms[current],
		Next:     terms[min(current+1, len(terms)-1)],
	}
}
