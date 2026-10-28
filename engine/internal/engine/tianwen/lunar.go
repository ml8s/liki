package tianwen

import (
	"container/list"
	"math"
	"sync"
	"time"

	"liki-engine/internal/engine/ganzhi"
)

// LunarTime is a Chinese lunisolar calendar date plus a shichen. The date is
// a calendar date (not a time scale); the shichen is a separate fact on the
// local apparent solar time (LAST) scale and must not be conflated with it.
//
// Date basis is a domain convention and must be explicit at the boundary:
//   - GB/T 33661 defines the Chinese calendar in UTC+8 civil dates;
//   - the ziwei path uses the apparent-solar date before converting (a school
//     convention), then the standard month sequence + leap rule, with the
//     LAST shichen.
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

type month11Key struct {
	year int
	ref  string
}

type leapMonthKey struct {
	current float64
	next    float64
	ref     string
}

type boundedMemoEntry[K comparable, V any] struct {
	key   K
	value V
}

// boundedMemo is a small, write-ordered cache for pure calendar functions.
// Lookups do not promote entries: the calendar working set is small and keeping
// lookup reads lock-free is preferable to strict LRU semantics here.
type boundedMemo[K comparable, V any] struct {
	mu      sync.RWMutex
	entries map[K]*list.Element
	order   *list.List
	cap     int
}

func newBoundedMemo[K comparable, V any](capacity int) *boundedMemo[K, V] {
	return &boundedMemo[K, V]{
		entries: make(map[K]*list.Element, capacity),
		order:   list.New(),
		cap:     capacity,
	}
}

func (m *boundedMemo[K, V]) get(key K) (V, bool) {
	m.mu.RLock()
	element, exists := m.entries[key]
	var value V
	if exists {
		value = element.Value.(*boundedMemoEntry[K, V]).value
	}
	m.mu.RUnlock()
	return value, exists
}

func (m *boundedMemo[K, V]) put(key K, value V) {
	m.mu.Lock()
	defer m.mu.Unlock()
	if element, exists := m.entries[key]; exists {
		element.Value.(*boundedMemoEntry[K, V]).value = value
		return
	}
	if m.order.Len() == m.cap {
		oldest := m.order.Back()
		if oldest != nil {
			entry := m.order.Remove(oldest).(*boundedMemoEntry[K, V])
			delete(m.entries, entry.key)
		}
	}
	m.entries[key] = m.order.PushFront(&boundedMemoEntry[K, V]{key: key, value: value})
}

var (
	newMoonJDTDMemo = newBoundedMemo[float64, float64](4096)
	month11KMemo    = newBoundedMemo[month11Key, float64](1024)
	leapMonthKMemo  = newBoundedMemo[leapMonthKey, float64](1024)
)

// ── Trigonometry ──

func sinDeg(deg float64) float64 { return math.Sin(deg * math.Pi / 180.0) }

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

func jdToUTCTime(jd float64) time.Time {
	z := math.Floor(jd + 0.5)
	f := jd + 0.5 - z
	y, m, d := jdToGregorian(jd, 0)
	return time.Date(y, time.Month(m), d, 0, 0, 0, 0, time.UTC).
		Add(time.Duration(f * float64(24*time.Hour)))
}

// deltaTDays returns TT-UT in days using the ShouXing ΔT table/model.
func deltaTDays(jdTD float64) float64 {
	ut := jdTD
	for range 3 {
		ut = jdTD - dtT(ut-2451545.0)
	}
	return dtT(ut - 2451545.0)
}

// ── New moon (Meeus Chapter 47) ──

func getNewMoonJDTD(k float64) float64 {
	if cached, ok := newMoonJDTDMemo.get(k); ok {
		return cached
	}
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

	result := JDE + corr
	newMoonJDTDMemo.put(k, result)
	return result
}

// approximateK returns the lunation number nearest the given civil date.
func approximateK(year, month, day int, tz float64) float64 {
	jd := gregorianToJD(year, month, day, tz)
	return (jd - 2451550.09765) / 29.530588853
}

// ── Month 11 and leap month ──

func getMonth11K(year int, tz float64) float64 {
	return fixedLunarReference(tz).month11K(year)
}

func getLeapMonthK(currentYearMonth11K, nextYearMonth11K, tz float64) float64 {
	return fixedLunarReference(tz).leapMonthK(currentYearMonth11K, nextYearMonth11K)
}

// ── Public API ──

// LunarDateAt converts a Gregorian instant to the Chinese lunar date in fixed
// UTC+8. It is the official compatibility wrapper; use NewLunarProjector when a
// timezone is available and the intent is local astronomical alignment.
func LunarDateAt(gt time.Time) LunarDate {
	return ShanghaiLunarReference().LunarDateAt(gt)
}

// SolarDateToLunar converts a civil solar date to the Chinese lunar date in the
// official UTC+8 reference. It is a legacy compatibility wrapper; use
// LunarProjector for a timezone-aware projection.
func SolarDateToLunar(solarYear, solarMonth, solarDay int) LunarDate {
	return ShanghaiLunarReference().SolarDateToLunar(NewCivilDate(solarYear, solarMonth, solarDay))
}

// CivilTimeFromLunarDate converts a Chinese lunar date to the official UTC+8
// civil date. It is a legacy compatibility wrapper; use LunarProjector when a
// timezone is available.
func CivilTimeFromLunarDate(ld LunarDate) time.Time {
	date, err := ShanghaiLunarReference().CivilDateFromLunar(ld)
	if err != nil {
		return time.Time{}
	}
	return time.Date(date.Year, time.Month(date.Month), date.Day, 0, 0, 0, 0, cstLocation)
}
