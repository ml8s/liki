package tianwen

import (
	"liki-engine/internal/engine/ganzhi"
	"time"
)

// RiZhu computes the day pillar for a given date using the Julian Day method.
// 1900-01-01 = 甲戌 (0-based index=10).
func RiZhu(gt time.Time) ganzhi.Zhu {
	t := gt
	year, month, day := t.Date()
	jd := int(julianDay(year, int(month), day))
	baseJD := int(julianDay(1900, 1, 1))
	diff := jd - baseJD
	gzIndex := (10 + diff) % 60
	if gzIndex < 0 {
		gzIndex += 60
	}
	return ganzhi.Zhu{Gan: ganzhi.Gan(gzIndex%10 + 1), Zhi: ganzhi.Zhi(gzIndex%12 + 1)}
}

// NianZhu computes the year pillar for a given date, accounting for 立春 boundary.
// If the date is before 立春 (exact solar-term time), the year gan/zhi is
// based on (year-1). Uses the precise 立春 moment, not just the calendar day,
// so births on 立春 day before the exact moment are still the previous year.
func NianZhu(gt time.Time) ganzhi.Zhu {
	t := gt
	year, _, _ := t.Date()
	lc := SolarTermTime(year, 315) // 立春精确时刻（UTC）
	if t.Before(lc) {
		year--
	}
	s := (year - 3) % 10
	if s <= 0 {
		s += 10
	}
	b := (year - 3) % 12
	if b <= 0 {
		b += 12
	}
	return ganzhi.Zhu{Gan: ganzhi.Gan(s), Zhi: ganzhi.Zhi(b)}
}

// YueZhu computes the month pillar from the given time, deriving the year gan via NianZhu internally.
func YueZhu(gt time.Time) ganzhi.Zhu {
	t := gt.UTC()
	jz := JianYue(time.Time(t))
	zhi := jz
	monthNum := (int(jz)+9)%12 + 1 // 1=寅月..12=丑月
	yp := NianZhu(time.Time(t))
	gan := ganzhi.Gan(((int(yp.Gan)*2 + monthNum) % 10))
	if gan == 0 {
		gan = 10
	}
	return ganzhi.Zhu{Gan: gan, Zhi: zhi}
}

// ShiZhu computes the hour pillar from solar time.
func ShiZhu(st time.Time) ganzhi.Zhu {
	solarMinutes := apparentSolarMinutes(st)
	zhi := shichenFromApparentSolarMinutes(solarMinutes)
	// 晚子时（23:00 后）：日柱不变，但时柱按次日日干起（lunar 约定）
	dayT := st
	if solarMinutes >= 1380 {
		dayT = dayT.AddDate(0, 0, 1)
	}
	dp := RiZhu(time.Time(dayT))
	gan := ganzhi.Gan(((int(dp.Gan)*2 + int(zhi) - 2) % 10))
	if gan == 0 {
		gan = 10
	}
	return ganzhi.Zhu{Gan: gan, Zhi: zhi}
}

// ShichenOf returns the hour branch of a true-solar moment.
func ShichenOf(st time.Time) ganzhi.Zhi {
	return shichenFromApparentSolarMinutes(apparentSolarMinutes(st))
}

// ComputeBazi derives the 四柱（年/月/日/时）.
//
// 时间基约定（“八字公历立春界”）：
//   - 年柱/月柱按 Gregorian time判定——节气时刻是绝对瞬时，与出生 Gregorian time直接比较，
//     与经度无关；
//   - 日柱/时柱按真太阳时判定——子时换日与时辰均以真太阳时为准。
func ComputeBazi(gt time.Time, st time.Time) ganzhi.Bazi {
	civil := gt
	solar := st
	civilYear, _, _ := civil.Date()
	beforeLichun := civil.Before(SolarTermTime(civilYear, 315))
	monthBranch := JianYue(time.Time(civil))
	shichen := ShichenOf(st)

	return ganzhi.Pillars(
		civil,
		solar,
		beforeLichun,
		monthBranch,
		shichen,
	)
}
