package qimen

import "time"

type BirthDatePrecision string

const (
	BirthDateMoment BirthDatePrecision = "moment"
)

// BirthDate is the optional 年命 input. The wire layer supplies an exact
// instant, so the 立春 year boundary is resolved without a date-only fallback.
type BirthDate struct {
	Time      time.Time
	Precision BirthDatePrecision
	Has       bool
}
