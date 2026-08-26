package tianwen

import (
	"math"
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

// JianYue returns the solar-term month branch (建月) for a given date.
// 寅=立春..丑=小寒.
func JianYue(gt GregorianTime) ganzhi.Zhi {
	t := gt.Time()
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

func solarLongitude(t time.Time) float64 {
	// 儒略日基于 UTC——先用 UTC 字段（避免本地时区偏移）
	u := t.UTC()
	jd := julianDayHMS(u.Year(), int(u.Month()), u.Day(), u.Hour(), u.Minute(), u.Second())
	return solarLongitudeShouXing(jd - 2451545.0)
}

func angleInRange(lon, cur, next float64) bool {
	if cur <= next {
		return lon >= cur && lon < next
	}
	return lon >= cur || lon < next
}

func SolarTermTime(year int, targetLon float64) time.Time {
	// 初始猜测：按黄经基准（太阳 1月1日黄经≈280°，每天 0.9856°）。
	// 归一化到当年：先算 1月1日黄经，再推 targetLon 对应日。
	jan1Lon := solarLongitude(time.Date(year, 1, 1, 12, 0, 0, 0, time.UTC))
	diff0 := targetLon - jan1Lon
	if diff0 > 180 {
		diff0 -= 360
	} else if diff0 < -180 {
		diff0 += 360
	}
	t := time.Date(year, 1, 1, 12, 0, 0, 0, time.UTC).AddDate(0, 0, int(diff0/0.9856))
	for iter := 0; iter < 30; iter++ {
		lon := solarLongitude(t)
		diff := targetLon - lon
		if diff > 180 {
			diff -= 360
		} else if diff < -180 {
			diff += 360
		}
		if math.Abs(diff) < 0.01 {
			break
		}
		step := diff / 0.9856
		if step > 15 {
			step = 15
		} else if step < -15 {
			step = -15
		}
		t = t.Add(time.Duration(step*24*3600) * time.Second)
	}
	// ΔT 修正：solarLongitude 算的是力学时，转世界时（对齐 lunar qiHigh）
	jd := julianDayHMS(t.Year(), int(t.Month()), t.Day(), t.Hour(), t.Minute(), t.Second())
	tt := (jd - 2451545.0) / 36525.0
	dt := dtT(tt)
	t = t.Add(-time.Duration(dt*86400.0) * time.Second)

	// For non-节 targets (ti=0), the initial guess at Feb 5 can cause
	// backward convergence into the previous year (e.g. 处暑 150°, 秋分 180°).
	// Advance one solar year and refine if we landed before the target year.
	if t.Year() < year {
		t = t.AddDate(0, 0, 365)
		lon := solarLongitude(t)
		diff := targetLon - lon
		if diff > 180 {
			diff -= 360
		} else if diff < -180 {
			diff += 360
		}
		step := diff / 0.9856
		if step <= 15 && step >= -15 {
			t = t.Add(time.Duration(step*24*3600) * time.Second)
		}
	}
	return t
}

func SolarTermIndex(year, month, day int) int {
	lon := solarLongitude(time.Date(year, time.Month(month), day, 12, 0, 0, 0, time.UTC))
	for i := 0; i < 24; i++ {
		if angleInRange(lon, JieQiLongitudes[(i+21)%24], JieQiLongitudes[(i+22)%24]) {
			return i
		}
	}
	return 23
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
