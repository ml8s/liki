package tianwen

import (
	"fmt"
	"time"
)

// LunarBasis selects the date frame used to project an instant into the
// Chinese lunisolar calendar.
type LunarBasis string

const (
	// LunarBasisOfficial uses the Beijing civil date and the UTC+8 calendar.
	LunarBasisOfficial LunarBasis = "official"
	// LunarBasisLocalLabel uses the caller's local date label with the UTC+8
	// calendar table. This is the ziwei-compatible school convention.
	LunarBasisLocalLabel LunarBasis = "local_label"
	// LunarBasisLocalAstronomical uses the caller's local date and a calendar
	// whose new moons and 中气 are assigned in the caller's zone.
	LunarBasisLocalAstronomical LunarBasis = "local_astronomical"
)

// ParseLunarBasis validates a basis name. The empty string is the default
// local_astronomical basis, which is the globally aligned projection.
func ParseLunarBasis(basis string) (LunarBasis, error) {
	switch LunarBasis(basis) {
	case "", LunarBasisLocalAstronomical:
		return LunarBasisLocalAstronomical, nil
	case LunarBasisOfficial, LunarBasisLocalLabel:
		return LunarBasis(basis), nil
	default:
		return "", fmt.Errorf(
			"lunar basis must be official, local_label, or local_astronomical, got %q", basis,
		)
	}
}

// LunarProjector combines the basis choice with the date-assignment reference.
// It is the only place where a timezone participates in lunar conversion.
type LunarProjector struct {
	basis LunarBasis
	ref   *LunarReference
	loc   *time.Location
}

// NewLunarProjector builds a projector. A nil location falls back to UTC+8.
func NewLunarProjector(basis LunarBasis, loc *time.Location) (LunarProjector, error) {
	parsed, err := ParseLunarBasis(string(basis))
	if err != nil {
		return LunarProjector{}, err
	}
	if loc == nil {
		loc = cstLocation
	}
	ref := ShanghaiLunarReference()
	if parsed == LunarBasisLocalAstronomical {
		ref = NewLunarReference(loc)
	}
	return LunarProjector{basis: parsed, ref: ref, loc: loc}, nil
}

// DefaultLunarProjector returns the legacy local-label basis used by
// date-based callers that do not carry a location.
func DefaultLunarProjector() LunarProjector {
	return LunarProjector{
		basis: LunarBasisLocalLabel,
		ref:   ShanghaiLunarReference(),
		loc:   cstLocation,
	}
}

// Basis returns the normalized basis.
func (p LunarProjector) Basis() LunarBasis { return p.basis }

// OutputBasis is the echo value; the default local_astronomical basis is
// omitted.
func (p LunarProjector) OutputBasis() string {
	if p.basis == LunarBasisLocalAstronomical {
		return ""
	}
	return string(p.basis)
}

// SelectorDate is the civil date used to look up the lunar date.
func (p LunarProjector) SelectorDate(at time.Time) CivilDate {
	if p.basis == LunarBasisOfficial {
		return BeijingDateAt(at)
	}
	return StandardDateAt(at, p.loc)
}

// SelectorInstant is a representative instant for a selector date in the
// basis's date frame.
func (p LunarProjector) SelectorInstant(date CivilDate) time.Time {
	if p.basis == LunarBasisOfficial {
		return date.Time()
	}
	return StandardNoon(date, p.loc)
}

// DayStart is midnight on the civil date in the basis's date frame.
func (p LunarProjector) DayStart(date CivilDate) time.Time {
	if p.basis == LunarBasisOfficial {
		return time.Date(date.Year, time.Month(date.Month), date.Day, 0, 0, 0, 0, cstLocation)
	}
	return StandardMidnight(date, p.loc)
}

// SolarDateToLunar converts an already-selected civil date.
func (p LunarProjector) SolarDateToLunar(date CivilDate) LunarDate {
	return p.ref.SolarDateToLunar(date)
}

// LunarAt projects an instant to the lunar date under this basis.
func (p LunarProjector) LunarAt(at time.Time) LunarDate {
	return p.ref.SolarDateToLunar(p.SelectorDate(at))
}

// CivilDateFromLunar converts a lunar date under this basis.
func (p LunarProjector) CivilDateFromLunar(ld LunarDate) (CivilDate, error) {
	return p.ref.CivilDateFromLunar(ld)
}

// LunarMonthMeta returns month metadata under this basis.
func (p LunarProjector) LunarMonthMeta(year, month int, leap bool) (LunarMonthInfo, error) {
	return p.ref.LunarMonthMeta(year, month, leap)
}

// ValidateLunarDate validates a lunar date under this basis.
func (p LunarProjector) ValidateLunarDate(date LunarDate) (LunarMonthInfo, error) {
	return p.ref.ValidateLunarDate(date)
}

// NextLunarMonth returns the next lunation under this basis.
func (p LunarProjector) NextLunarMonth(year, month int, leap bool) (LunarMonthInfo, error) {
	return p.ref.NextLunarMonth(year, month, leap)
}
