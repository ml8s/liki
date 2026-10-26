package ganzhi

import (
	"time"
)

// yearPillar derives the year pillar from the civil Gregorian year and the
// 立春 boundary decision. The calendar year is incremented by the caller when
// the instant precedes that year's exact 立春.
func yearPillar(civilYear int, beforeLichun bool) Zhu {
	if beforeLichun {
		civilYear--
	}

	stem := (civilYear - 3) % 10
	if stem <= 0 {
		stem += 10
	}

	branch := (civilYear - 3) % 12
	if branch <= 0 {
		branch += 12
	}

	return Zhu{Gan: Gan(stem), Zhi: Zhi(branch)}
}

// monthPillar derives the month pillar from the solar-term month branch and
// the year pillar. The month stem follows the 五虎遁 rule.
func monthPillar(yearPillar Zhu, monthBranch Zhi) Zhu {
	monthNumber := (int(monthBranch)+9)%12 + 1 // 1=寅月 .. 12=丑月
	stem := (int(yearPillar.Gan)*2 + monthNumber) % 10
	if stem == 0 {
		stem = 10
	}
	return Zhu{Gan: Gan(stem), Zhi: monthBranch}
}

// DayPillar derives the day pillar from the true-solar calendar date.
func DayPillar(solarDate time.Time) Zhu {
	year, month, day := solarDate.Date()
	solarJulianDay := julianDayNumber(year, int(month), day)
	baseJulianDay := julianDayNumber(1900, 1, 1)
	sexagenaryIndex := (10 + solarJulianDay - baseJulianDay) % 60
	if sexagenaryIndex < 0 {
		sexagenaryIndex += 60
	}
	return Zhu{
		Gan: Gan(sexagenaryIndex%10 + 1),
		Zhi: Zhi(sexagenaryIndex%12 + 1),
	}
}

// julianDayNumber returns the Julian Day Number at noon for the civil date.
func julianDayNumber(year, month, day int) int {
	if month <= 2 {
		year--
		month += 12
	}

	gregorianCorrection := year/100 - year/400
	return 365*year + year/4 - gregorianCorrection +
		(153*(month-3)+2)/5 + day + 1721119
}

// HourPillar derives the hour pillar from the true-solar clock. Late 子时
// advances the day pillar for the hour stem, but does not change the chart's
// day pillar.
func HourPillar(solarClock time.Time, shichen Zhi) Zhu {
	solarMinutes := solarClock.Hour()*60 + solarClock.Minute()
	dayForHourStem := solarClock
	if solarMinutes >= 23*60 {
		dayForHourStem = dayForHourStem.AddDate(0, 0, 1)
	}

	dayPillar := DayPillar(dayForHourStem)
	stem := (int(dayPillar.Gan)*2 + int(shichen) - 2) % 10
	if stem == 0 {
		stem = 10
	}
	return Zhu{Gan: Gan(stem), Zhi: shichen}
}

// Pillars derives four pillars from already-resolved time facts. civilAt is
// the legal local wall-clock projection; solarAt is the true-solar projection.
func Pillars(
	civilAt time.Time,
	solarAt time.Time,
	beforeLichun bool,
	monthBranch Zhi,
	shichen Zhi,
) Bazi {
	civilYear, _, _ := civilAt.Date()
	nian := yearPillar(civilYear, beforeLichun)
	yue := monthPillar(nian, monthBranch)
	dayPillar := DayPillar(solarAt)
	hourPillar := HourPillar(solarAt, shichen)
	return Bazi{
		Nian: nian,
		Yue:  yue,
		Ri:   dayPillar,
		Shi:  hourPillar,
	}
}
