package ziwei

import (
	"time"

	"liki-engine/internal/engine/ganzhi"
	"liki-engine/internal/engine/tianwen"
)

// ComputeChart computes a complete紫微命盘 from an already-resolved lunar time
// and gender. Timezone, lunar reference, and date validity belong to the
// caller's projection; this entry does not re-derive the calendar. Use
// ComputeChartAtBasis when resolving a standard-time input.
func ComputeChart(lt tianwen.LunarTime, gender ganzhi.Gender) (Chart, error) {
	return computeChartFromLunarTime(lt, gender), nil
}

// computeChartFromLunarTime assembles a chart from a caller-resolved lunar
// time. It deliberately does not re-validate the calendar reference.
func computeChartFromLunarTime(lt tianwen.LunarTime, gender ganzhi.Gender) Chart {
	nianGan := Gan(((lt.Year-4)%10+10)%10 + 1)
	nianZhi := Zhi(((lt.Year-4)%12+12)%12 + 1)
	shiZhi := Zhi(lt.Shichen)

	// Reconstruct minimal Bazi for computeChart
	bz := ganzhi.Bazi{
		Nian: ganzhi.Zhu{Gan: ganzhi.Gan(nianGan), Zhi: ganzhi.Zhi(nianZhi)},
		Shi:  ganzhi.Zhu{Zhi: ganzhi.Zhi(shiZhi)},
	}

	chart := computeChart(bz, lt)
	chart.BirthYear = lt.Year
	chart.Gender = gender
	chart.School = ChartSchool{
		LeapMonth: "first_fifteen_days_current_month_next_fifteen_days_following_month",
		Source:    "iztro_compatible",
	}
	// LunarMonth/LunarDay 已在 computeChart 内设置（含闰月调整），不覆盖
	chart.NianZhi = nianZhi
	chart = buildChartDetail(chart)
	chart.MingZhu = soulStar(chart.GongWei[0].Zhi)
	chart.ShenZhu = bodyStar(nianZhi)
	chart.PalaceFacts = computePalaceFacts(chart)
	return chart
}

// ComputeChartAt is the standard-time entry for new callers. The default lunar
// basis is the local astronomical calendar, which aligns each locality.
func ComputeChartAt(
	standard time.Time,
	longitude float64,
	gender ganzhi.Gender,
) (Chart, error) {
	return ComputeChartAtBasis(standard, longitude, gender, tianwen.LunarBasisLocalAstronomical)
}

// ComputeChartAtBasis is ComputeChartAt with an explicit lunar basis.
func ComputeChartAtBasis(
	standard time.Time,
	longitude float64,
	gender ganzhi.Gender,
	basis tianwen.LunarBasis,
) (Chart, error) {
	loc := standard.Location()
	projector, err := tianwen.NewLunarProjector(basis, loc)
	if err != nil {
		return Chart{}, err
	}
	lunarDate := projector.LunarAt(standard)
	shichen := tianwen.ShichenFromStandard(standard, longitude)
	lunar := tianwen.LunarTime{
		Year:    lunarDate.Year,
		Month:   lunarDate.Month,
		Day:     lunarDate.Day,
		Leap:    lunarDate.Leap,
		Shichen: shichen,
	}
	// The projector already validated the civil-date projection, so validating
	// the resulting lunar date again would repeat expensive new-moon and
	// leap-month work.
	chart := computeChartFromLunarTime(lunar, gender)
	return chart, nil
}
