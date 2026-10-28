package tianwen

import (
	"math"
	"sync"
	"time"

	"liki-engine/internal/engine/ganzhi"
)

var JieQiLongitudes = [24]float64{315, 330, 345, 0, 15, 30, 45, 60, 75, 90, 105, 120, 135, 150, 165, 180, 195, 210, 225, 240, 255, 270, 285, 300}

var solarTermLongitudes = func() [12]float64 {
	var a [12]float64
	for i := 0; i < 12; i++ {
		a[i] = JieQiLongitudes[i*2]
	}
	return a
}()

// SolarTermTime and the direct longitude projection can differ by a small
// interval around a boundary. Keep that region on the legacy projection so the
// public JianYue semantics remain byte-for-byte chart-equivalent.
const jianYueTermFallbackWindow = 3 * time.Minute

// JianYue returns the solar-term month zhi (建月) for a given date.
// 寅=立春..丑=小寒.
func JianYue(gt time.Time) ganzhi.Zhi {
	if zhi, ok := jianYueFromTermIntervals(gt); ok {
		return zhi
	}
	return jianYueFromSolarLongitude(gt)
}

// jianYueFromTermIntervals locates the month by comparing the instant with
// memoized solar-term instants. Away from boundaries, repeated chart work
// in one year avoids the full VSOP87 longitude evaluation.
func jianYueFromTermIntervals(gt time.Time) (ganzhi.Zhi, bool) {
	year := gt.Year()
	for i := 0; i < 12; i++ {
		startYears := [2]int{year, year}
		startCount := 1
		// 大雪 (index 10) starts the 子月 interval in December and ends at
		// 小寒 in January. A January date can therefore belong to the interval
		// started in the previous December or the one started this December.
		if i == 10 {
			startYears = [2]int{year - 1, year}
			startCount = 2
		}
		for _, startYear := range startYears[:startCount] {
			endYear := startYear
			if i == 10 {
				endYear++
			}
			start := SolarTermTime(startYear, solarTermLongitudes[i])
			end := SolarTermTime(endYear, solarTermLongitudes[(i+1)%12])
			// Near a term boundary, retain the original continuous-longitude
			// semantics; normal chart instants remain on the fast path.
			if nearSolarTermInstant(gt, start) || nearSolarTermInstant(gt, end) {
				return ganzhi.Zhi(0), false
			}
			if !gt.Before(start) && gt.Before(end) {
				return ganzhi.Zhi((i+2)%12 + 1), true
			}
		}
	}
	return ganzhi.Zhi(0), false
}

func nearSolarTermInstant(at, boundary time.Time) bool {
	distance := at.Sub(boundary)
	if distance < 0 {
		distance = -distance
	}
	return distance <= jianYueTermFallbackWindow
}

func jianYueFromSolarLongitude(gt time.Time) ganzhi.Zhi {
	t := gt
	lon := solarLongitude(t)
	for i := 0; i < 12; i++ {
		if angleInRange(lon, solarTermLongitudes[i], solarTermLongitudes[(i+1)%12]) {
			return ganzhi.Zhi((i+2)%12 + 1)
		}
	}
	return ganzhi.ZhiYin
}

// julianDay returns the Julian Day Number with fractional day (时刻精度).
func julianDay(year, month, day int) float64 {
	return julianDayHMS(year, month, day, 0, 0, 0)
}

// julianDayHMS returns the Julian Day with hour/minute/second precision.
func julianDayHMS(year, month, day, hour, min, sec int) float64 {
	if month <= 2 {
		year--
		month += 12
	}
	A := year / 100
	jd := float64(int(365.25*float64(year+4716)) + int(30.6001*float64(month+1)) + day + (2 - A + A/4) - 1524)
	// 加时刻（UTC 当天的时分秒转小数天，减 0.5 对齐儒略日中午起算）
	jd += (float64(hour)-12)/24.0 + float64(min)/1440.0 + float64(sec)/86400.0
	return jd
}

func julianDayTime(t time.Time) float64 {
	u := t.UTC()
	return julianDayHMS(u.Year(), int(u.Month()), u.Day(), u.Hour(), u.Minute(), u.Second()) +
		float64(u.Nanosecond())/1e9/86400.0
}

func normalizeDegrees(deg float64) float64 {
	return math.Mod(math.Mod(deg, 360.0)+360.0, 360.0)
}

func solarLongitude(t time.Time) float64 {
	jdUT := julianDayTime(t)
	jdTT := jdUT + dtT(jdUT-2451545.0)
	tt := (jdTT - 2451545.0) / 36525.0
	return normalizeDegrees(saLon(tt, -1) * 180.0 / math.Pi)
}

func angleInRange(lon, cur, next float64) bool {
	if cur <= next {
		return lon >= cur && lon < next
	}
	return lon >= cur || lon < next
}

// SolarTermTime returns the instant the sun's apparent longitude reaches
// targetLon degrees (a solar term / 节气) in the given year (Beijing time).
var solarTermCache sync.Map

type solarTermKey struct {
	year int
	lon  float64
}

// SolarTermTime returns the cached instant the sun's apparent longitude
// reaches targetLon degrees. targetLon 为固定枚举（节气黄经），缓存有界。
func SolarTermTime(year int, targetLon float64) time.Time {
	key := solarTermKey{year: year, lon: targetLon}
	if cached, ok := solarTermCache.Load(key); ok {
		return cached.(time.Time)
	}
	t := computeSolarTermTime(year, targetLon)
	solarTermCache.Store(key, t)
	return t
}

func computeSolarTermTime(year int, targetLon float64) time.Time {
	// 初始猜测落在目标年内：从 1 月 1 日黄经向东推到 targetLon。
	// 对 270°（冬至）这样黄经小于年初黄经的目标，加 360° 后落在 12 月，
	// 不会回退到上一年。
	jan1 := time.Date(year, 1, 1, 12, 0, 0, 0, time.UTC)
	diff0 := normalizeDegrees(targetLon - solarLongitude(jan1))
	t := jan1.Add(time.Duration(diff0 / 0.9856 * 24 * float64(time.Hour)))

	for iter := 0; iter < 100; iter++ {
		lon := solarLongitude(t)
		diff := targetLon - lon
		if diff > 180 {
			diff -= 360
		} else if diff < -180 {
			diff += 360
		}
		if math.Abs(diff) < 1e-7 {
			break
		}
		step := diff / 0.9856
		if step > 15 {
			step = 15
		} else if step < -15 {
			step = -15
		}
		t = t.Add(time.Duration(step * 24 * float64(time.Hour)))
	}
	return t
}

// AllSolarTerms returns all 24 solar term times for the given year, ordered
// from冬至(0) through大雪(23).
func AllSolarTerms(year int) [24]time.Time {
	// 24 节气顺序：从 year-1 冬至 到 year 大雪。
	// 冬至(270°)在 year-1 12月，小寒(285°)在 year 1月，其余在 year 年内。
	var terms [24]time.Time
	terms[0] = SolarTermTime(year-1, JieQiLongitudes[21]) // 冬至 year-1
	terms[1] = SolarTermTime(year, JieQiLongitudes[22])   // 小寒 year
	terms[2] = SolarTermTime(year, JieQiLongitudes[23])   // 大寒 year
	for i := 3; i < 24; i++ {
		terms[i] = SolarTermTime(year, JieQiLongitudes[i-3])
	}
	return terms
}
