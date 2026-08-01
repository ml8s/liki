package tianwen

import (

	"liki-engine/internal/engine/ganzhi"
)

// RiZhu computes the day pillar for a given date using the Julian Day method.
// 1900-01-01 = 甲戌 (0-based index=10).
func RiZhu(gt GregorianTime) ganzhi.Zhu {
	t := gt.Time()
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
// If the date is before 立春 (exact solar-term time), the year stem/branch is
// based on (year-1). Uses the precise 立春 moment, not just the calendar day,
// so births on 立春 day before the exact moment are still the previous year.
func NianZhu(gt GregorianTime) ganzhi.Zhu {
	t := gt.Time()
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

// YueZhu computes the month pillar from the given time, deriving the year stem via NianZhu internally.
func YueZhu(gt GregorianTime) ganzhi.Zhu {
	t := gt.Time().UTC()
	jz := JianYue(GregorianTime(t))
	zhi := jz
	monthNum := (int(jz)+9)%12 + 1 // 1=寅月..12=丑月
	yp := NianZhu(GregorianTime(t))
	gan := ganzhi.Gan(((int(yp.Gan)*2 + monthNum) % 10))
	if gan == 0 {
		gan = 10
	}
	return ganzhi.Zhu{Gan: gan, Zhi: zhi}
}

// ShiZhu computes the hour pillar from solar time.
func ShiZhu(st SolarTime) ganzhi.Zhu {
	solarMinutes := st.Minutes()
	zhi := hourZhiFromSolarTime(solarMinutes)
	// 晚子时（23:00 后）：日柱不变，但时柱按次日日干起（lunar 约定）
	dayT := st.Time()
	if solarMinutes >= 1380 {
		dayT = dayT.AddDate(0, 0, 1)
	}
	dp := RiZhu(GregorianTime(dayT))
	gan := ganzhi.Gan(((int(dp.Gan)*2 + int(zhi) - 2) % 10))
	if gan == 0 {
		gan = 10
	}
	return ganzhi.Zhu{Gan: gan, Zhi: zhi}
}

func ComputeBazi(st SolarTime) ganzhi.Bazi {
	t := st.Time()
	yp := NianZhu(GregorianTime(t))
	mp := YueZhu(GregorianTime(t.UTC()))
	dp := RiZhu(GregorianTime(t))
	hp := ShiZhu(st)
	return ganzhi.Bazi{Nian: yp, Yue: mp, Ri: dp, Shi: hp}
}
