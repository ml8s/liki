package tianwen

import (
	"fmt"
	"math"
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

// LunarMonthMeta validates and returns information about the real month in the
// official UTC+8 reference. Use LunarProjector for a timezone-aware reference.
func LunarMonthMeta(year, month int, leap bool) (LunarMonthInfo, error) {
	return ShanghaiLunarReference().LunarMonthMeta(year, month, leap)
}

// LunarMonthMeta validates and returns information about the real month
// identified by the year/month/leap triple in this reference.
func (r *LunarReference) LunarMonthMeta(year, month int, leap bool) (LunarMonthInfo, error) {
	if month < 1 || month > 12 {
		return LunarMonthInfo{}, fmt.Errorf("lunar month must be 1-12, got %d", month)
	}
	k, ok := r.findLunarMonthInLunarYear(year, month, leap)
	if !ok {
		label := "month"
		if leap {
			label = "leap month"
		}
		return LunarMonthInfo{}, fmt.Errorf("lunar %s %d-%02d does not exist", label, year, month)
	}

	dayCount, err := r.lunarMonthDayCount(k)
	if err != nil {
		return LunarMonthInfo{}, err
	}
	return LunarMonthInfo{Year: year, Month: month, Leap: leap, DayCount: dayCount}, nil
}

// ValidateLunarDate checks date against the official UTC+8 reference. Use
// LunarProjector for a timezone-aware reference.
func ValidateLunarDate(date LunarDate) (LunarMonthInfo, error) {
	return ShanghaiLunarReference().ValidateLunarDate(date)
}

// ValidateLunarDate checks that date identifies one real day in this reference
// and returns the containing month's metadata.
func (r *LunarReference) ValidateLunarDate(date LunarDate) (LunarMonthInfo, error) {
	meta, err := r.LunarMonthMeta(date.Year, date.Month, date.Leap)
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

// NextLunarMonth returns the next lunation in the official UTC+8 reference.
// Use LunarProjector for a timezone-aware reference.
func NextLunarMonth(year, month int, leap bool) (LunarMonthInfo, error) {
	return ShanghaiLunarReference().NextLunarMonth(year, month, leap)
}

// NextLunarMonth returns the next real lunation after the identified month in
// this reference.
func (r *LunarReference) NextLunarMonth(year, month int, leap bool) (LunarMonthInfo, error) {
	if month < 1 || month > 12 {
		return LunarMonthInfo{}, fmt.Errorf("lunar month must be 1-12, got %d", month)
	}
	currentK, ok := r.findLunarMonthInLunarYear(year, month, leap)
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
		startM11K := r.month11K(anchorYear)
		endM11K := r.month11K(anchorYear + 1)
		leapK := r.leapMonthK(startM11K, endM11K)
		switch {
		case currentK+1 == leapK:
			nextLeap = true
		case month == 12:
			nextYear, nextMonth = year+1, 1
		default:
			nextMonth = month + 1
		}
	}

	meta, err := r.LunarMonthMeta(nextYear, nextMonth, nextLeap)
	if err != nil {
		return LunarMonthInfo{}, fmt.Errorf("find lunar month after %d-%02d%s: %w", year, month, leapSuffix(leap), err)
	}
	return meta, nil
}

func (r *LunarReference) findLunarMonthInLunarYear(lunarYear, lunarMonth int, leap bool) (float64, bool) {
	anchorYear := lunarYear
	if lunarMonth < 11 {
		anchorYear = lunarYear - 1
	}
	return r.searchMonthKInRange(anchorYear, lunarMonth, leap)
}

func (r *LunarReference) searchMonthKInRange(anchorYear, lunarMonth int, leap bool) (float64, bool) {
	startM11K := r.month11K(anchorYear)
	endM11K := r.month11K(anchorYear + 1)
	leapMonthK := r.leapMonthK(startM11K, endM11K)

	currentMonth := 11
	isLeap := false

	for i := 0.0; i <= 14.0; i++ {
		k := startM11K + i

		if i > 0 {
			if k == leapMonthK {
				isLeap = true
			} else {
				isLeap = false
				currentMonth++
				if currentMonth == 13 {
					currentMonth = 1
				}
			}
		}

		if currentMonth == lunarMonth && isLeap == leap {
			return k, true
		}
	}
	return 0, false
}

func (r *LunarReference) lunarMonthDayCount(k float64) (int, error) {
	current := r.dateToJD(r.newMoonDate(k))
	next := r.dateToJD(r.newMoonDate(k + 1))
	dayCount := int(math.Round(next - current))
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
