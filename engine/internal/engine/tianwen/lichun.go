package tianwen

import (
	"fmt"
	"time"
)

// LichunYear is the 立春 year containing a civil date. The 立春 boundary date
// belongs to the new year; BoundaryDay reports that date-level truncation.
type LichunYear struct {
	Year        int
	Start       CivilDate
	End         CivilDate
	BoundaryDay bool
}

// LichunYearOf returns the 立春 year containing date. The 立春 boundary date
// belongs to the new year.
func LichunYearOf(date CivilDate) (LichunYear, error) {
	if err := date.Validate(); err != nil {
		return LichunYear{}, err
	}
	lichunThisYear := SolarTermCivilDate(date.Year, 315)
	year := date.Year
	if date.Before(lichunThisYear) {
		year--
	}
	if year < 1 {
		return LichunYear{}, fmt.Errorf("lichun year: date %s precedes the supported calendar", date)
	}
	start := SolarTermCivilDate(year, 315)
	end := SolarTermCivilDate(year+1, 315)
	return LichunYear{
		Year:        year,
		Start:       start,
		End:         end,
		BoundaryDay: date == start,
	}, nil
}

// LichunYearInstant is the 立春 year containing an absolute instant, with the
// exact boundary instants. BoundaryDay reports whether the instant's CST date
// is the 立春 date.
type LichunYearInstant struct {
	Year        int
	Start       time.Time
	End         time.Time
	BoundaryDay bool
}

// LichunYearAt returns the 立春 year containing instant. The comparison uses
// absolute instants, so it is independent of the observer's timezone.
func LichunYearAt(instant time.Time) (LichunYearInstant, error) {
	cst := instant.In(cstLocation)
	year := cst.Year()
	lichun := SolarTermTime(year, 315)
	if instant.Before(lichun) {
		year--
	}
	if year < 1 {
		return LichunYearInstant{}, fmt.Errorf("lichun year: instant %s precedes the supported calendar", instant)
	}
	start := SolarTermTime(year, 315)
	end := SolarTermTime(year+1, 315)
	return LichunYearInstant{
		Year:        year,
		Start:       start,
		End:         end,
		BoundaryDay: CivilDateOf(cst) == CivilDateOf(start.In(cstLocation)),
	}, nil
}
