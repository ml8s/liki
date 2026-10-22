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
func ResolveCivilCandidates(
	year, month, day, hour, minute, second, nanosecond int,
	loc *time.Location,
) ([]time.Time, error) {
	wall := time.Date(year, time.Month(month), day, hour, minute, second, nanosecond, time.UTC)

	_, offsetBefore := wall.Add(-26 * time.Hour).In(loc).Zone()
	_, offsetAfter := wall.Add(26 * time.Hour).In(loc).Zone()
	offsets := []int{offsetBefore}
	if offsetAfter != offsetBefore {
		offsets = append(offsets, offsetAfter)
	}

	candidates := make([]time.Time, 0, 2)
	for _, offset := range offsets {
		at := wall.Add(-time.Duration(offset) * time.Second)
		local := at.In(loc)
		gotYear, gotMonth, gotDay := local.Date()
		gotHour, gotMinute, gotSecond := local.Clock()
		if gotYear != year || int(gotMonth) != month || gotDay != day ||
			gotHour != hour || gotMinute != minute || gotSecond != second ||
			local.Nanosecond() != nanosecond {
			continue
		}
		candidates = append(candidates, at.UTC())
	}

	if len(candidates) == 0 {
		return nil, ErrCivilTimeGap
	}

	sort.Slice(candidates, func(i, j int) bool {
		return candidates[i].Before(candidates[j])
	})
	return candidates, nil
}

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

func ApparentSolarOffset(at time.Time, longitude float64) time.Duration {
	_, offsetSeconds := at.Zone()
	solar := ApparentSolarFromCivil(at, longitude, float64(offsetSeconds)/3600)
	return solar.Sub(at)
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

func ApparentSolarDate(
	at time.Time,
	loc *time.Location,
	longitude float64,
) (year int, month time.Month, day int) {
	return ApparentSolarCivil(at, loc, longitude).Date()
}

// ShichenAt derives the hour branch from apparent solar time.
func ShichenAt(
	at time.Time,
	loc *time.Location,
	longitude float64,
) ganzhi.Zhi {
	return ShichenOf(time.Time(ApparentSolarCivil(at, loc, longitude)))
}

func MonthBranchAt(at time.Time) ganzhi.Zhi {
	return JianYue(time.Time(at))
}

// SolarTermStateAt returns the previous, current, and next solar-term
// boundaries around at. Boundary comparisons use absolute instants.
func SolarTermStateAt(at time.Time) SolarTermState {
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

	current := sort.Search(len(terms), func(i int) bool {
		return terms[i].At.After(at)
	})
	if current > 0 {
		current--
	}

	return SolarTermState{
		Previous: terms[current-1],
		Current:  terms[current],
		Next:     terms[min(current+1, len(terms)-1)],
	}
}
