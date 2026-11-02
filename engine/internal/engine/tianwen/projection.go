package tianwen

import (
	"errors"
	"sort"
	"time"
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

// ApparentSolarOffset returns the correction from local civil time to local
// apparent solar time (LAST) for the given longitude.
func apparentSolarMinutes(t time.Time) float64 {
	return float64(t.Hour()*60 + t.Minute())
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
