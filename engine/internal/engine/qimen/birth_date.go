package qimen

import (
	"fmt"
	"time"

	"liki-engine/internal/engine/tianwen"
)

type BirthDatePrecision string

const (
	BirthDateOnly   BirthDatePrecision = "date_only"
	BirthDateMoment BirthDatePrecision = "moment"
)

type BirthDate struct {
	Time      time.Time
	Precision BirthDatePrecision
	Has       bool
}

func ParseBirthDate(value string) (BirthDate, error) {
	if len(value) == 10 {
		t, err := time.Parse("2006-01-02", value)
		if err != nil {
			return BirthDate{}, fmt.Errorf("invalid birth_date %q: %w", value, err)
		}
		if isLichunCivilDay(t) {
			return BirthDate{}, fmt.Errorf("birth_date is on the Lichun boundary day; provide RFC3339 time")
		}
		return BirthDate{Time: t, Precision: BirthDateOnly, Has: true}, nil
	}
	t, err := time.Parse(time.RFC3339, value)
	if err != nil {
		return BirthDate{}, fmt.Errorf("invalid birth_date %q: %w", value, err)
	}
	return BirthDate{Time: t, Precision: BirthDateMoment, Has: true}, nil
}

func isLichunCivilDay(dateOnly time.Time) bool {
	lichun := tianwen.SolarTermTime(dateOnly.Year(), 315).In(time.FixedZone("CST", 8*3600))
	return dateOnly.Year() == lichun.Year() &&
		dateOnly.Month() == lichun.Month() &&
		dateOnly.Day() == lichun.Day()
}
