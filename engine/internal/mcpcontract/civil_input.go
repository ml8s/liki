package mcpcontract

import (
	"fmt"
	"strings"
	"time"

	"liki-engine/internal/engine/tianwen"
)

type resolvedCivil struct {
	At       time.Time
	Loc      *time.Location
	Standard time.Time
	Civil    time.Time
}

func parseCivilTime(value CivilTime) (year int, month time.Month, day, hour, minute, second, nanosecond int, err error) {
	raw := strings.TrimSpace(string(value))
	for _, layout := range []string{"2006-01-02T15:04:05.999999999", "2006-01-02T15:04"} {
		parsed, parseErr := time.Parse(layout, raw)
		if parseErr != nil {
			continue
		}
		return parsed.Year(), parsed.Month(), parsed.Day(), parsed.Hour(), parsed.Minute(), parsed.Second(), parsed.Nanosecond(), nil
	}
	return 0, 0, 0, 0, 0, 0, 0, fmt.Errorf("civil_time must be local YYYY-MM-DDTHH:MM[:SS], got %q", value)
}

func resolveCivilTime(value CivilTime, timezone string, fold FoldName) (resolvedCivil, error) {
	year, month, day, hour, minute, second, nanosecond, err := parseCivilTime(value)
	if err != nil {
		return resolvedCivil{}, err
	}
	loc, err := loadTimezone(timezone)
	if err != nil {
		return resolvedCivil{}, err
	}
	instants, err := tianwen.CivilInstants(year, month, day, hour, minute, second, nanosecond, loc)
	if err != nil {
		return resolvedCivil{}, err
	}
	switch len(instants) {
	case 0:
		return resolvedCivil{}, fmt.Errorf("civil_time %q does not exist in %s", value, timezone)
	case 1:
	case 2:
		switch fold {
		case "earlier":
			instants = instants[:1]
		case "later":
			instants = instants[1:]
		case "":
			return resolvedCivil{}, fmt.Errorf(
				"civil_time %q is ambiguous in %s; candidates: %s and %s; pass fold=earlier or fold=later",
				value, timezone,
				standardTimeString(instants[0], loc), standardTimeString(instants[1], loc),
			)
		default:
			return resolvedCivil{}, fmt.Errorf("fold must be earlier or later, got %q", fold)
		}
	default:
		return resolvedCivil{}, fmt.Errorf("civil_time %q has %d candidates; refusing to guess", value, len(instants))
	}
	at := instants[0]
	return resolvedCivil{
		At:       at,
		Loc:      loc,
		Standard: tianwen.StandardLocalTime(at, loc),
		Civil:    at.In(loc),
	}, nil
}

func standardTimeString(at time.Time, loc *time.Location) string {
	return tianwen.StandardLocalTime(at, loc).Format(time.RFC3339Nano)
}

func civilTimeString(at time.Time, loc *time.Location) string {
	return at.In(loc).Format(time.RFC3339Nano)
}

func civilInputString(at time.Time, loc *time.Location) CivilTime {
	return CivilTime(at.In(loc).Format("2006-01-02T15:04:05.999999999"))
}

func parseStandardTime(value StandardTime) (time.Time, *time.Location, error) {
	at, err := time.Parse(time.RFC3339Nano, string(value))
	if err != nil {
		return time.Time{}, nil, fmt.Errorf("standard_time must be RFC3339 with offset, got %q: %w", value, err)
	}
	return at, at.Location(), nil
}

func standardTimeValue(at time.Time) StandardTime {
	return StandardTime(at.Format(time.RFC3339Nano))
}
