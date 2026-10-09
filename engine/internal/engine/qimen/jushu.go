package qimen

import (
	"math"
	"time"

	"liki-engine/internal/engine/ganzhi"
	"liki-engine/internal/engine/tianwen"
)

type solarTermInfo struct {
	Name  string
	Index int
	Time  time.Time
}

func determineJuShu(t time.Time, bz ganzhi.Bazi, method Method) juShu {
	actual := currentSolarTermInfo(t)
	switch scopeMethods[method.Scope].MethodSource {
	case "quarter_rule":
		variant := quarterRules[method.QuarterRule]
		if method.QuarterRule == QuarterTwelveMinuteTenDivision {
			base := determineJuShu(t, bz, Method{
				Scope: variant.BaseScope, School: variant.BaseSchool, Dingju: method.BaseDingju,
			})
			quarter := newTwelveMinuteQuarterInfo(t, bz.Shi, !base.YinDun, method)
			ju := twelveMinuteShiftJu(base.Number, quarter.Index, quarter.HourYangDun)
			return juShu{
				Method: method, Number: ju, YinDun: !quarter.HourYangDun,
				Yuan: base.Yuan, JieQi: actual.Name, YongJuJieQi: base.YongJuJieQi,
				Quarter: &quarter,
			}
		}
		quarter := newTenMinuteQuarterInfo(t, bz.Shi)
		ju := quarter.Sanyuan.YangJu
		if !quarter.HourYangDun {
			ju = quarter.Sanyuan.YinJu
		}
		return juShu{
			Method: method, Number: ju, YinDun: !quarter.HourYangDun,
			Yuan: quarter.Sanyuan.Yuan, JieQi: actual.Name,
			YongJuJieQi: variant.YongJuJieQi, Quarter: &quarter,
		}
	case "month_cycle":
		group := monthGroups[bz.Nian.Zhi]
		monthSeq := floorMod(int(bz.Yue.Zhi)-int(monthFirstBranch), 12)
		ju := floorMod(group.LeadingJu-1-monthSeq, 9) + 1
		return juShu{
			Method: method, Number: ju, YinDun: true, Yuan: group.Yuan,
			JieQi: actual.Name, YongJuJieQi: "月家定局",
		}
	case "year_cycle":
		cycleYear := actual.Time.Year()
		if actual.Index < 3 {
			cycleYear--
		}
		years := cycleYear - yearAnchor
		element := floorMod((years-floorMod(years, yearCycle))/yearCycle, 3)
		yuan := yuanName(element)
		return juShu{
			Method: method, Number: yearJuByYuan[yuan], YinDun: true, Yuan: yuan,
			JieQi: actual.Name, YongJuJieQi: "年家定局",
		}
	default:
		yongJu := actual.Name
		state := ""
		if method.Dingju == DingjuZhiRun {
			yongJu, state = determineZhiRunTerm(t, actual)
		}
		yuan := determineYuan(bz.Ri)
		if method.Dingju == DingjuMaoShan {
			yuan = determineMaoShanYuan(t, actual)
		}
		entry := solarTermJuTable[yongJu]
		return juShu{
			Method: method,
			Number: entry[yuan], YinDun: entry[3] == 0,
			Yuan: yuanName(yuan), JieQi: actual.Name,
			YongJuJieQi: yongJu, ZhiRunState: state,
		}
	}
}

func determineMaoShanYuan(t time.Time, actual solarTermInfo) int {
	anchor := actual.Time.Truncate(time.Hour)
	if actual.Time.Hour()%2 == 0 {
		anchor = anchor.Add(-time.Hour)
	}
	elapsed := t.Sub(anchor)
	if elapsed < 0 {
		elapsed = 0
	}
	shichen := int(elapsed / (2 * time.Hour))
	yuan := shichen / maoShanShichenPerYuan
	if yuan >= maoShanMaxYuan {
		return maoShanMaxYuan - 1
	}
	return yuan
}

func currentSolarTermInfo(t time.Time) solarTermInfo {
	result := solarTermInfo{Name: "冬至"}
	var latest time.Time
	if previousWinter := tianwen.SolarTermTime(t.Year()-1, 270); !t.Before(previousWinter) {
		result = solarTermInfoAt(t.Year()-1, 0)
		latest = previousWinter
	}
	for name, longitude := range solarTermLongitudes {
		term := tianwen.SolarTermTime(t.Year(), longitude)
		if !t.Before(term) && term.After(latest) {
			result = solarTermInfo{
				Name: name, Index: solarTermCanonicalIndex(longitude), Time: term,
			}
			latest = term
		}
	}
	return result
}

func solarTermInfoAt(year, index int) solarTermInfo {
	name := solarTermOrder[index]
	return solarTermInfo{
		Name: name, Index: index,
		Time: tianwen.SolarTermTime(year, canonicalSolarLongitude(index)),
	}
}

func solarTermCanonicalIndex(longitude float64) int {
	if longitude < 0 || longitude >= 360 || math.Mod(longitude, 15) != 0 {
		return -1
	}
	return int(math.Mod(longitude-270+360, 360) / 15)
}

func canonicalSolarLongitude(index int) float64 {
	return math.Mod(float64(index)*15+270, 360)
}

func solarTermSequence(info solarTermInfo) int {
	year := info.Time.Year()
	if info.Index == 0 {
		year++
	}
	return year*24 + info.Index
}

func determineZhiRunTerm(t time.Time, actual solarTermInfo) (string, string) {
	day := qimenDayNumber(t)
	leader := day - floorMod(day, zhiRunLeaderDays)
	upcomingLongitude := 90.0
	if actual.Index >= 12 {
		upcomingLongitude = 270
	}
	upcoming := nextSolarTerm(t, upcomingLongitude)
	currentLongitude := 270.0
	if upcomingLongitude == 270 {
		currentLongitude = 90
	}
	current := latestSolarTerm(t, currentLongitude)
	previous := solarTermInfoAt(current.Time.Year()-1, 12-current.Index)

	for _, anchor := range []solarTermInfo{upcoming, current, previous} {
		anchorLeader := adoptingZhiRunLeader(anchor, t.Location())
		if anchorLeader > leader {
			continue
		}
		rawInterval := (leader - anchorLeader) / zhiRunLeaderDays
		interval := min(rawInterval, zhiRunMaxIntervals-1)
		workingSequence := solarTermSequence(anchor) + interval
		workingIndex := floorMod(workingSequence, 24)
		working := solarTermOrder[workingIndex]
		return working, zhiRunState(t, actual, working, upcoming, anchor, leader, rawInterval, workingSequence)
	}
	return actual.Name, zhiRunStateNames["behind"]
}

func nextSolarTerm(t time.Time, longitude float64) solarTermInfo {
	year := t.Year()
	term := tianwen.SolarTermTime(year, longitude)
	if term.Before(t) {
		year++
		term = tianwen.SolarTermTime(year, longitude)
	}
	return solarTermInfo{
		Name:  solarTermOrder[solarTermCanonicalIndex(longitude)],
		Index: solarTermCanonicalIndex(longitude), Time: term,
	}
}

func latestSolarTerm(t time.Time, longitude float64) solarTermInfo {
	year := t.Year()
	term := tianwen.SolarTermTime(year, longitude)
	if term.After(t) {
		year--
		term = tianwen.SolarTermTime(year, longitude)
	}
	return solarTermInfo{
		Name:  solarTermOrder[solarTermCanonicalIndex(longitude)],
		Index: solarTermCanonicalIndex(longitude), Time: term,
	}
}

func adoptingZhiRunLeader(term solarTermInfo, location *time.Location) int {
	day := qimenDayNumber(term.Time.In(location))
	lead := floorMod(day, zhiRunLeaderDays)
	leader := day - lead
	if lead > zhiRunThreshold {
		leader += zhiRunLeaderDays
	}
	return leader
}

func zhiRunState(t time.Time, actual solarTermInfo, working string, upcoming, anchor solarTermInfo, leader, rawInterval, workingSequence int) string {
	workingTerm := solarTermInfoFromSequence(workingSequence)
	workingDay := qimenDayNumber(workingTerm.Time.In(t.Location()))
	actualSequence := solarTermSequence(actual)
	if workingDay == leader && workingSequence == actualSequence && !workingTerm.Time.After(t) {
		return zhiRunStateNames["exact"]
	}
	if qimenDayNumber(upcoming.Time.In(t.Location())) == qimenDayNumber(t) {
		return zhiRunStateNames["behind"]
	}
	if rawInterval > zhiRunMaxIntervals-1 && zhiRunRepeatTerms[working] &&
		actual.Name == working && qimenDayNumber(anchor.Time.In(t.Location())) != qimenDayNumber(t) {
		return zhiRunStateNames["repeat"]
	}
	if workingSequence > actualSequence {
		return zhiRunStateNames["ahead"]
	}
	return zhiRunStateNames["behind"]
}

func solarTermInfoFromSequence(sequence int) solarTermInfo {
	sequenceYear := floorDiv(sequence, 24)
	index := floorMod(sequence, 24)
	occurrenceYear := sequenceYear
	if index == 0 {
		occurrenceYear--
	}
	return solarTermInfoAt(occurrenceYear, index)
}

func qimenDayNumber(t time.Time) int {
	year, month, day := t.Date()
	unixDay := int(time.Date(year, month, day, 0, 0, 0, 0, time.UTC).Unix() / 86400)
	result := unixDay - 10963
	if t.Hour() >= 23 {
		result++
	}
	return result
}

func floorMod(value, modulus int) int {
	return ((value % modulus) + modulus) % modulus
}

func floorDiv(value, modulus int) int {
	return (value - floorMod(value, modulus)) / modulus
}

func determineYuan(dayZhu ganzhi.Zhu) int {
	dayIdx := ganzhi.SixtyCycleIndex(dayZhu.Gan, dayZhu.Zhi)
	fuTou := ganzhi.SixtyToZhu(dayIdx - dayIdx%5)
	return fuTouYuanTable[fuTou.Zhi]
}

func yuanName(yuan int) string {
	return yuanNames[yuan]
}

func fuTou(dayZhu ganzhi.Zhu) ganzhi.Zhu {
	dayIdx := ganzhi.SixtyCycleIndex(dayZhu.Gan, dayZhu.Zhi)
	return ganzhi.SixtyToZhu(dayIdx - dayIdx%5)
}
