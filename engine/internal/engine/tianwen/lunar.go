package tianwen

import (
	"math"
	"time"

	"liki-engine/internal/engine/ganzhi"
)

// LunarTime is a Chinese lunar calendar date with shichen.
type LunarTime struct {
	Year    int        `json:"year"`
	Month   int        `json:"month"`
	Day     int        `json:"day"`
	Leap    bool       `json:"leap"`
	Shichen ganzhi.Zhi `json:"shichen"`
}

// Default timezone for lunar calendar: UTC+8 (China Standard Time).
// The lunar day boundary follows local midnight, and traditionally the
// Chinese calendar uses China Standard Time.
const defaultTZ = 8.0

const secondsToDays = 1.0 / 86400.0

// ── Trigonometry ──

func sinDeg(deg float64) float64 { return math.Sin(deg * math.Pi / 180.0) }

func normalizeDeg(deg float64) float64 {
	return math.Mod(math.Mod(deg, 360.0)+360.0, 360.0)
}

// ── Julian Day (Meeus Chapter 7) ──

func gregorianToJD(year, month, day int, tz float64) float64 {
	y := float64(year)
	m := float64(month)
	d := float64(day) - tz/24.0 // midnight local → UT

	if m < 3 {
		y--
		m += 12
	}
	a := math.Floor(y / 100.0)
	b := 2.0 - a + math.Floor(a/4.0)
	return math.Floor(365.25*(y+4716.0)) +
		math.Floor(30.6001*(m+1.0)) +
		d + b - 1524.5
}

func jdToGregorian(jd, tz float64) (int, int, int) {
	jd += tz / 24.0
	z := math.Floor(jd + 0.5)
	f := jd + 0.5 - z
	alpha := math.Floor((z - 1867216.25) / 36524.25)
	a := z + 1.0 + alpha - math.Floor(alpha/4.0)
	b := a + 1524.0
	c := math.Floor((b - 122.1) / 365.25)
	d := math.Floor(365.25 * c)
	e := math.Floor((b - d) / 30.6001)

	month := e - 1
	if e >= 14 {
		month = e - 13
	}
	year := c - 4716
	if month <= 2 {
		year = c - 4715
	}
	day := int(b - d - math.Floor(30.6001*e) + f)
	return int(year), int(month), day
}

// ── Delta T (Meeus Chapter 9) ──

func getDeltaTSeconds(jd float64) float64 {
	return -15.0 + (jd-2382148.0)*(jd-2382148.0)/41048480.0
}

func getDeltaTSecondsTD(jdTD float64) float64 {
	ut := jdTD
	var prev float64
	for ut != prev {
		prev = ut
		ut = jdTD - getDeltaTSeconds(ut)*secondsToDays
	}
	return getDeltaTSeconds(ut)
}

// ── Sun true longitude (Meeus Chapter 24) ──

func getSunTrueLongitude(jdTD float64) float64 {
	T := (jdTD - 2451545.0) / 36525.0
	T2 := T * T
	L0 := 280.46645 + 36000.76983*T + 0.0003032*T2
	M := 357.52910 + 35999.05030*T + 0.0001559*T2 - 0.00000048*T2*T
	C := (1.914600-0.004817*T-0.000014*T2)*sinDeg(M) +
		(0.019993-0.000101*T)*sinDeg(2*M) +
		0.000290*sinDeg(3*M)
	return normalizeDeg(L0 + C)
}

// ── New moon (Meeus Chapter 47) ──

func getNewMoonJDTD(k float64) float64 {
	T := k / 1236.85
	T2 := T * T
	T3 := T2 * T
	T4 := T3 * T
	E := 1.0 - 0.002516*T - 0.0000074*T2

	JDE := 2451550.09765 +
		29.530588853*k +
		0.0001337*T2 -
		0.000000150*T3 +
		0.00000000073*T4

	M := 2.5534 + 29.10535669*k - 0.0000218*T2 - 0.00000011*T3
	MPrime := 201.5643 + 385.81693528*k + 0.0107438*T2 + 0.00001239*T3 - 0.000000058*T4
	F := 160.7108 + 390.67050274*k - 0.0016341*T2 - 0.00000227*T3 + 0.000000011*T4
	Omega := 124.7746 - 1.56375580*k + 0.0020691*T2 + 0.00000215*T3

	corr := -0.40720*sinDeg(MPrime) +
		0.17241*E*sinDeg(M) +
		0.01608*sinDeg(2*MPrime) +
		0.01039*sinDeg(2*F) +
		0.00739*E*sinDeg(MPrime-M) -
		0.00514*E*sinDeg(MPrime+M) +
		0.00208*E*E*sinDeg(2*M) -
		0.00111*sinDeg(MPrime-2*F) -
		0.00057*sinDeg(MPrime+2*F) +
		0.00056*E*sinDeg(2*MPrime+M) -
		0.00042*sinDeg(3*MPrime) +
		0.00042*E*sinDeg(M+2*F) +
		0.00038*E*sinDeg(M-2*F) -
		0.00024*E*sinDeg(2*MPrime-M) -
		0.00017*sinDeg(Omega) -
		0.00007*sinDeg(MPrime+2*M) +
		0.00004*sinDeg(2*MPrime-2*F) +
		0.00004*sinDeg(3*M) +
		0.00003*sinDeg(MPrime+M-2*F) +
		0.00003*sinDeg(2*MPrime+2*F) -
		0.00003*sinDeg(MPrime+M+2*F) +
		0.00003*sinDeg(MPrime-M+2*F) -
		0.00002*sinDeg(MPrime-M-2*F) -
		0.00002*sinDeg(3*MPrime+M) +
		0.00002*sinDeg(4*MPrime) +
		0.000325*sinDeg(299.77+0.107408*k-0.009173*T2) +
		0.000165*sinDeg(251.88+0.016321*k) +
		0.000164*sinDeg(251.83+26.651886*k) +
		0.000126*sinDeg(349.42+36.412478*k) +
		0.000110*sinDeg(84.66+18.206239*k) +
		0.000062*sinDeg(141.74+53.303771*k) +
		0.000060*sinDeg(207.14+2.453732*k) +
		0.000056*sinDeg(154.84+7.306860*k) +
		0.000047*sinDeg(34.52+27.261239*k) +
		0.000042*sinDeg(207.19+0.121824*k) +
		0.000040*sinDeg(291.34+1.844379*k) +
		0.000037*sinDeg(161.72+24.198154*k) +
		0.000035*sinDeg(239.56+25.513099*k) +
		0.000023*sinDeg(331.55+3.592518*k)

	return JDE + corr
}

// approximateK returns an approximate lunation number for the given date.
func approximateK(year, month, day int, tz float64) float64 {
	k := 2.0
	if year%400 == 0 || (year%4 == 0 && year%100 != 0) {
		k = 1.0
	}
	dayOfYear := math.Floor(275.0*float64(month)/9.0) -
		k*math.Floor((float64(month)+9.0)/12.0) +
		float64(day) - 30.0
	denom := 365.0
	if k == 1.0 {
		denom = 366.0
	}
	f := -tz / 24.0
	return ((dayOfYear-1.0+f)/denom + float64(year) - 2000.0) * 12.3685
}

// ── Winter solstice ──

func getWinterSolsticeJDTD(year int, tz float64) float64 {
	left := gregorianToJD(year, 12, 1, tz)
	right := left + 35.0

	var mid float64
	for range 100 {
		mid = (left + right) / 2.0
		lon := getSunTrueLongitude(mid)
		if math.Abs(lon-270.0) < 1e-9 {
			return mid
		}
		if lon < 270.0 {
			left = mid
		} else {
			right = mid
		}
	}
	return mid
}

// ── Month 11 and leap month ──

func compareDates(y1, m1, d1, y2, m2, d2 int) int {
	if y1 != y2 {
		return y1 - y2
	}
	if m1 != m2 {
		return m1 - m2
	}
	return d1 - d2
}

func getMonth11K(year int, tz float64) float64 {
	w := getWinterSolsticeJDTD(year, tz)
	wDeltaT := getDeltaTSeconds(w)
	wy, wm, wd := jdToGregorian(w-wDeltaT*secondsToDays, tz)

	k := math.Round(approximateK(wy, wm, wd, tz))
	m := getNewMoonJDTD(k)
	mDeltaT := getDeltaTSeconds(m)
	my, mm, md := jdToGregorian(m-mDeltaT*secondsToDays, tz)

	if compareDates(my, mm, md, wy, wm, wd) > 0 {
		return k - 1
	}
	return k
}

func getLeapMonthK(currentYearMonth11K, nextYearMonth11K, tz float64) float64 {
	if nextYearMonth11K-currentYearMonth11K <= 12.0 {
		return -1
	}

	for i := 1.0; i <= 12.0; i++ {
		curTD := getNewMoonJDTD(currentYearMonth11K + i)
		nextTD := getNewMoonJDTD(currentYearMonth11K + i + 1)

		curUT := curTD - getDeltaTSecondsTD(curTD)*secondsToDays
		y, m, d := jdToGregorian(curUT, tz)
		curMidJD := gregorianToJD(y, m, d, tz)
		curMidTD := curMidJD + getDeltaTSeconds(curMidJD)*secondsToDays
		lon1 := getSunTrueLongitude(curMidTD)

		nextUT := nextTD - getDeltaTSecondsTD(nextTD)*secondsToDays
		y2, m2, d2 := jdToGregorian(nextUT, tz)
		nextMidJD := gregorianToJD(y2, m2, d2, tz)
		nextMidTD := nextMidJD + getDeltaTSeconds(nextMidJD)*secondsToDays
		lon2 := getSunTrueLongitude(nextMidTD)

		if math.Floor(lon1/30.0) == math.Floor(lon2/30.0) {
			return currentYearMonth11K + i
		}
	}
	return -1
}

// ── Helpers ──

func newMoonMidnightJD(k, tz float64) float64 {
	td := getNewMoonJDTD(k)
	ut := td - getDeltaTSecondsTD(td)*secondsToDays
	y, m, d := jdToGregorian(ut, tz)
	return gregorianToJD(y, m, d, tz)
}

// ── Public API ──

// SolarToLunar converts a Gregorian date to Chinese lunar date.
// The date is interpreted at local midnight in UTC+8.
func SolarToLunar(gt GregorianTime) LunarTime {
	t := gt.Time()
	solarYear, solarMonth, solarDay := t.Date()
	tz := defaultTZ

	gDayJD := gregorianToJD(solarYear, int(solarMonth), solarDay, tz)
	targetK := math.Round(approximateK(solarYear, int(solarMonth), solarDay, tz))

	dayOneJD := newMoonMidnightJD(targetK, tz)
	nextDayOneJD := newMoonMidnightJD(targetK+1, tz)

	if gDayJD < dayOneJD {
		targetK--
		dayOneJD = newMoonMidnightJD(targetK, tz)
	} else if gDayJD >= nextDayOneJD {
		targetK++
		dayOneJD = nextDayOneJD
	}

	m11ThisYear := getMonth11K(solarYear, tz)
	var startM11K, endM11K float64
	var baseYear int

	if targetK >= m11ThisYear {
		startM11K = m11ThisYear
		endM11K = getMonth11K(solarYear+1, tz)
		baseYear = solarYear
	} else {
		startM11K = getMonth11K(solarYear-1, tz)
		endM11K = m11ThisYear
		baseYear = solarYear - 1
	}

	leapMonthK := getLeapMonthK(startM11K, endM11K, tz)
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

	day := int(gDayJD - dayOneJD + 1)
	return LunarTime{Year: lunarYear, Month: lunarMonth, Day: day, Leap: leap}
}

// LunarToGregorian converts a Chinese lunar date to Gregorian date.
// The result is in UTC+8 local time.
func LunarToGregorian(lt LunarTime) GregorianTime {
	tz := defaultTZ

	targetK, ok := findLunarMonthK(lt.Year, lt.Month, lt.Leap, tz)
	if !ok {
		return GregorianTime(time.Time{})
	}

	// Validate day is within the month.
	td := getNewMoonJDTD(targetK)
	nextTD := getNewMoonJDTD(targetK + 1)
	ut := td - getDeltaTSecondsTD(td)*secondsToDays
	nextUT := nextTD - getDeltaTSecondsTD(nextTD)*secondsToDays
	y, m, d := jdToGregorian(ut, tz)
	ny, nm, nd := jdToGregorian(nextUT, tz)
	midnightJD := gregorianToJD(y, m, d, tz)
	nextMidnightJD := gregorianToJD(ny, nm, nd, tz)
	monthSize := int(nextMidnightJD - midnightJD)

	if lt.Day < 1 || lt.Day > monthSize {
		return GregorianTime(time.Time{})
	}

	targetJD := midnightJD + float64(lt.Day-1)
	gy, gm, gd := jdToGregorian(targetJD, tz)
	return GregorianTime(time.Date(gy, time.Month(gm), gd, 0, 0, 0, 0, time.FixedZone("CST", int(defaultTZ*3600))))
}

// findLunarMonthK searches for the lunation number k that corresponds to
// the given lunar year, month, and leap flag.
func findLunarMonthK(lunarYear, lunarMonth int, leap bool, tz float64) (float64, bool) {
	// Primary anchor: the Gregorian year whose Month 11 starts this lunar year.
	anchorYear := lunarYear
	if lunarMonth < 11 {
		anchorYear = lunarYear - 1
	}

	if k, ok := searchMonthKInRange(anchorYear, lunarMonth, leap, tz); ok {
		return k, true
	}

	// Fallback for months < 11: the leap month may shift the year boundary.
	if lunarMonth < 11 {
		if k, ok := searchMonthKInRange(lunarYear, lunarMonth, leap, tz); ok {
			return k, true
		}
	}

	return 0, false
}

func searchMonthKInRange(anchorYear, lunarMonth int, leap bool, tz float64) (float64, bool) {
	startM11K := getMonth11K(anchorYear, tz)
	endM11K := getMonth11K(anchorYear+1, tz)
	leapMonthK := getLeapMonthK(startM11K, endM11K, tz)

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
