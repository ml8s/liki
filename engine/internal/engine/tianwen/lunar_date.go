package tianwen

import (
	"fmt"
)

// LunarDate is a date in the Chinese lunisolar calendar. It deliberately has
// no shichen: calendar validation and shichen selection are separate concerns.
type LunarDate struct {
	Year  int  `json:"year"`
	Month int  `json:"month"`
	Day   int  `json:"day"`
	Leap  bool `json:"leap"`
}

// LunarMonthInfo is one real month in a lunisolar year.
type LunarMonthInfo struct {
	Year     int
	Month    int
	Leap     bool
	DayCount int
}

// LunarMonthMeta validates and returns information about the real month
// identified by the year/month/leap triple. A missing leap month is an error,
// not a silent fallback to the ordinary month.
func LunarMonthMeta(year, month int, leap bool) (LunarMonthInfo, error) {
	if month < 1 || month > 12 {
		return LunarMonthInfo{}, fmt.Errorf("lunar month must be 1-12, got %d", month)
	}
	k, ok := findLunarMonthInLunarYear(year, month, leap)
	if !ok {
		label := "month"
		if leap {
			label = "leap month"
		}
		return LunarMonthInfo{}, fmt.Errorf("lunar %s %d-%02d does not exist", label, year, month)
	}

	dayCount, err := lunarMonthDayCount(k)
	if err != nil {
		return LunarMonthInfo{}, err
	}
	return LunarMonthInfo{Year: year, Month: month, Leap: leap, DayCount: dayCount}, nil
}

// ValidateLunarDate checks that date identifies one real day in the Chinese
// lunisolar calendar and returns the containing month's metadata.
func ValidateLunarDate(date LunarDate) (LunarMonthInfo, error) {
	meta, err := LunarMonthMeta(date.Year, date.Month, date.Leap)
	if err != nil {
		return LunarMonthInfo{}, err
	}
	if date.Day < 1 || date.Day > meta.DayCount {
		return LunarMonthInfo{}, fmt.Errorf(
			"lunar day %d is outside %d-%02d%s (1-%d)",
			date.Day, date.Year, date.Month, leapSuffix(date.Leap), meta.DayCount,
		)
	}
	return meta, nil
}

// NextLunarMonth returns the next real lunation after the identified month.
// Month transitions follow the calendar (including leap months and year wrap);
// callers must not implement month++ heuristics.
func NextLunarMonth(year, month int, leap bool) (LunarMonthInfo, error) {
	if month < 1 || month > 12 {
		return LunarMonthInfo{}, fmt.Errorf("lunar month must be 1-12, got %d", month)
	}
	currentK, ok := findLunarMonthInLunarYear(year, month, leap)
	if !ok {
		return LunarMonthInfo{}, fmt.Errorf("lunar month %d-%02d does not exist", year, month)
	}

	nextYear, nextMonth, nextLeap := year, month, false
	if leap {
		// A leap month follows its ordinary-month ordinal; the next lunation
		// therefore advances to the next ordinal. A leap twelfth month is
		// followed by the next year's first month.
		if month == 12 {
			nextYear, nextMonth = year+1, 1
		} else {
			nextMonth = month + 1
		}
	} else {
		anchorYear := year
		if month < 11 {
			anchorYear = year - 1
		}
		startM11K := getMonth11K(anchorYear, defaultTZ)
		endM11K := getMonth11K(anchorYear+1, defaultTZ)
		leapK := getLeapMonthK(startM11K, endM11K, defaultTZ)
		switch {
		case currentK+1 == leapK:
			nextLeap = true
		case month == 12:
			nextYear, nextMonth = year+1, 1
		default:
			nextMonth = month + 1
		}
	}

	meta, err := LunarMonthMeta(nextYear, nextMonth, nextLeap)
	if err != nil {
		return LunarMonthInfo{}, fmt.Errorf("find lunar month after %d-%02d%s: %w", year, month, leapSuffix(leap), err)
	}
	return meta, nil
}

func findLunarMonthInLunarYear(lunarYear, lunarMonth int, leap bool) (float64, bool) {
	anchorYear := lunarYear
	if lunarMonth < 11 {
		anchorYear = lunarYear - 1
	}
	return searchMonthKInRange(anchorYear, lunarMonth, leap, defaultTZ)
}

func lunarMonthDayCount(k float64) (int, error) {
	currentMidnight := newMoonMidnightJD(k, defaultTZ)
	nextMidnight := newMoonMidnightJD(k+1, defaultTZ)
	dayCount := int(nextMidnight - currentMidnight)
	if dayCount < 29 || dayCount > 30 {
		return 0, fmt.Errorf("invalid lunar month length %d at lunation %.3f", dayCount, k)
	}
	return dayCount, nil
}

func leapSuffix(leap bool) string {
	if leap {
		return " leap"
	}
	return ""
}
