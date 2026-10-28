package tianwen

import (
	"fmt"
	"time"
)

// JieqiMonthWindow is the solar-term month (节气月) containing a civil date.
// A month starts at a 节 (立春, 惊蛰, ...) and ends at the next 节. Start is
// inclusive and End is exclusive. Index is 0=寅 .. 11=丑.
type JieqiMonthWindow struct {
	Index        int
	Start        CivilDate
	End          CivilDate
	StartInstant time.Time
	EndInstant   time.Time
	BoundaryDay  bool
}

// JieqiMonthOf resolves the solar-term month containing date. The boundary
// date belongs to the new month; BoundaryDay reports that truncation. Selection
// compares civil dates, not instants, so a 节 that occurs later on its civil
// date still starts the new month for the whole date.
func JieqiMonthOf(date CivilDate) (JieqiMonthWindow, error) {
	if err := date.Validate(); err != nil {
		return JieqiMonthWindow{}, err
	}

	terms := solarTermsAround(date.Time())
	start := -1
	end := -1
	for index, term := range terms {
		if term.Index%2 != 0 {
			continue
		}
		termDate := CivilDateOf(term.At.In(cstLocation))
		if !date.Before(termDate) {
			start = index
			continue
		}
		if start >= 0 {
			end = index
			break
		}
	}
	if start < 0 {
		return JieqiMonthWindow{}, fmt.Errorf("jieqi month: no 节 before %s", date)
	}
	if end < 0 {
		return JieqiMonthWindow{}, fmt.Errorf("jieqi month: no 节 after %s", date)
	}

	startDate := CivilDateOf(terms[start].At.In(cstLocation))
	return JieqiMonthWindow{
		Index:        terms[start].Index / 2,
		Start:        startDate,
		End:          CivilDateOf(terms[end].At.In(cstLocation)),
		StartInstant: terms[start].At,
		EndInstant:   terms[end].At,
		BoundaryDay:  date == startDate,
	}, nil
}

// JieqiMonthAt resolves the solar-term month containing an absolute instant.
// The 节 boundary is compared as an instant; Start/End are local civil dates.
func JieqiMonthAt(instant time.Time, loc *time.Location) (JieqiMonthWindow, error) {
	if loc == nil {
		loc = cstLocation
	}
	terms := solarTermsAround(instant)
	start, end := -1, -1
	for index, term := range terms {
		if term.Index%2 != 0 {
			continue
		}
		if !term.At.After(instant) {
			start = index
			continue
		}
		if start >= 0 {
			end = index
			break
		}
	}
	if start < 0 || end < 0 {
		return JieqiMonthWindow{}, fmt.Errorf("jieqi month: no 节 around %s", instant)
	}
	startDate := StandardDateAt(terms[start].At, loc)
	return JieqiMonthWindow{
		Index:        terms[start].Index / 2,
		Start:        startDate,
		End:          StandardDateAt(terms[end].At, loc),
		StartInstant: terms[start].At,
		EndInstant:   terms[end].At,
		BoundaryDay:  StandardDateAt(instant, loc) == startDate,
	}, nil
}
