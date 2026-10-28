package tianwen

import (
	"fmt"
	"math"
	"sort"
	"time"
)

// LunarReference assigns astronomical instants to civil dates for the Chinese
// lunisolar calendar. It is the only timezone-bearing layer in the calendar:
// new moons, solar terms, and winter solstices stay absolute instants.
type LunarReference struct {
	id  string
	loc *time.Location
}

var shanghaiLunarReference = &LunarReference{id: "UTC+8", loc: cstLocation}

// ShanghaiLunarReference returns the GB/T 33661 reference: Chinese calendar
// dates are UTC+8 civil dates. The fixed zone is deliberate; Asia/Shanghai
// would introduce historical DST and diverge from the official table.
func ShanghaiLunarReference() *LunarReference {
	return shanghaiLunarReference
}

// NewLunarReference returns a reference that assigns each astronomical event
// to the civil date in loc. A nil loc falls back to the official UTC+8 frame.
func NewLunarReference(loc *time.Location) *LunarReference {
	if loc == nil {
		return ShanghaiLunarReference()
	}
	return &LunarReference{id: "loc:" + loc.String(), loc: loc}
}

func fixedLunarReference(tz float64) *LunarReference {
	if math.Abs(tz-defaultTZ) < 1e-9 {
		return ShanghaiLunarReference()
	}
	seconds := int(math.Round(tz * 3600.0))
	return &LunarReference{
		id:  fmt.Sprintf("fixed:%.6f", tz),
		loc: time.FixedZone(fmt.Sprintf("UTC%+g", tz), seconds),
	}
}

func (r *LunarReference) dateOf(at time.Time) CivilDate {
	return StandardDateAt(at, r.loc)
}

func (r *LunarReference) offsetHours(date CivilDate) float64 {
	return float64(StandardOffsetSeconds(date.TimeIn(r.loc), r.loc)) / 3600.0
}

// dateToJD is a civil-date serial, not a local instant: both endpoints use the
// same reference, so DST transitions cannot make day arithmetic non-integral.
func (r *LunarReference) dateToJD(date CivilDate) float64 {
	return gregorianToJD(date.Year, date.Month, date.Day, 0)
}

func (r *LunarReference) dateOfJD(jd float64) CivilDate {
	return StandardDateAt(jdToUTCTime(jd), r.loc)
}

func (r *LunarReference) approximateK(date CivilDate) float64 {
	return approximateK(date.Year, date.Month, date.Day, r.offsetHours(date))
}

func (r *LunarReference) newMoonDate(k float64) CivilDate {
	td := getNewMoonJDTD(k)
	ut := td - deltaTDays(td)
	return r.dateOfJD(ut)
}

func (r *LunarReference) newMoonDateJD(k float64) float64 {
	return r.dateToJD(r.newMoonDate(k))
}

func (r *LunarReference) month11K(year int) float64 {
	key := month11Key{year: year, ref: r.id}
	if cached, ok := month11KMemo.get(key); ok {
		return cached
	}

	wDate := r.dateOf(SolarTermTime(year, 270))
	k := math.Round(r.approximateK(wDate))
	mDate := r.newMoonDate(k)

	result := k
	if mDate.After(wDate) {
		result = k - 1
	}
	month11KMemo.put(key, result)
	return result
}

func (r *LunarReference) zhongQiDates(anchorYear int) []CivilDate {
	dates := make([]CivilDate, 0, 26)
	for year := anchorYear; year <= anchorYear+1; year++ {
		for i := 1; i < len(JieQiLongitudes); i += 2 {
			dates = append(dates, r.dateOf(SolarTermTime(year, JieQiLongitudes[i])))
		}
	}
	sort.Slice(dates, func(i, j int) bool { return dates[i].Before(dates[j]) })
	return dates
}

func monthHasZhongQi(start, next CivilDate, zhongQi []CivilDate) bool {
	for _, date := range zhongQi {
		if !date.Before(start) && date.Before(next) {
			return true
		}
	}
	return false
}

func (r *LunarReference) leapMonthK(currentYearMonth11K, nextYearMonth11K float64) float64 {
	key := leapMonthKey{current: currentYearMonth11K, next: nextYearMonth11K, ref: r.id}
	if cached, ok := leapMonthKMemo.get(key); ok {
		return cached
	}
	if nextYearMonth11K-currentYearMonth11K <= 12.0 {
		leapMonthKMemo.put(key, -1)
		return -1
	}

	anchorYear := r.newMoonDate(currentYearMonth11K).Year
	zhongQi := r.zhongQiDates(anchorYear)

	for i := 1.0; i <= 12.0; i++ {
		k := currentYearMonth11K + i
		start := r.newMoonDate(k)
		next := r.newMoonDate(k + 1)
		if !monthHasZhongQi(start, next, zhongQi) {
			leapMonthKMemo.put(key, k)
			return k
		}
	}
	leapMonthKMemo.put(key, -1)
	return -1
}

// LunarDateAt converts an instant to the Chinese lunar date whose civil date is
// assigned by this reference.
func (r *LunarReference) LunarDateAt(at time.Time) LunarDate {
	return r.SolarDateToLunar(r.dateOf(at))
}

// SolarDateToLunar converts a civil solar date to the Chinese lunar date. The
// date is taken as given; month/leap boundaries follow this reference.
func (r *LunarReference) SolarDateToLunar(date CivilDate) LunarDate {
	gDayJD := r.dateToJD(date)
	targetK := math.Round(r.approximateK(date))

	dayOneJD := r.newMoonDateJD(targetK)
	nextDayOneJD := r.newMoonDateJD(targetK + 1)

	if gDayJD < dayOneJD {
		targetK--
		dayOneJD = r.newMoonDateJD(targetK)
	} else if gDayJD >= nextDayOneJD {
		targetK++
		dayOneJD = nextDayOneJD
	}

	m11ThisYear := r.month11K(date.Year)
	var startM11K, endM11K float64
	var baseYear int

	if targetK >= m11ThisYear {
		startM11K = m11ThisYear
		endM11K = r.month11K(date.Year + 1)
		baseYear = date.Year
	} else {
		startM11K = r.month11K(date.Year - 1)
		endM11K = m11ThisYear
		baseYear = date.Year - 1
	}

	leapMonthK := r.leapMonthK(startM11K, endM11K)
	lunarMonth := 11
	leap := false
	lunarYear := baseYear

	for k := startM11K + 1; k <= targetK; k++ {
		if k == leapMonthK {
			leap = true
		} else {
			leap = false
			lunarMonth++
			if lunarMonth == 13 {
				lunarMonth = 1
				lunarYear++
			}
		}
	}

	day := int(math.Round(gDayJD-dayOneJD)) + 1
	return LunarDate{Year: lunarYear, Month: lunarMonth, Day: day, Leap: leap}
}

// CivilDateFromLunar converts a Chinese lunar date to the civil date assigned
// by this reference.
func (r *LunarReference) CivilDateFromLunar(ld LunarDate) (CivilDate, error) {
	targetK, ok := r.findLunarMonthInLunarYear(ld.Year, ld.Month, ld.Leap)
	if !ok {
		label := "month"
		if ld.Leap {
			label = "leap month"
		}
		return CivilDate{}, fmt.Errorf("lunar %s %d-%02d does not exist", label, ld.Year, ld.Month)
	}

	start := r.newMoonDate(targetK)
	next := r.newMoonDate(targetK + 1)
	monthSize := int(math.Round(r.dateToJD(next) - r.dateToJD(start)))

	if ld.Day < 1 || ld.Day > monthSize {
		return CivilDate{}, fmt.Errorf(
			"lunar day %d is outside %d-%02d%s (1-%d)",
			ld.Day, ld.Year, ld.Month, leapSuffix(ld.Leap), monthSize,
		)
	}
	return start.AddDays(ld.Day - 1), nil
}
