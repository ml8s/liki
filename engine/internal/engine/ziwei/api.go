package ziwei

import (
	"fmt"
	"time"

	"liki-engine/internal/engine/ganzhi"
	"liki-engine/internal/engine/tianwen"
)

// ComputeChart computes a complete紫微命盘 from lunar time and gender.
// The LunarTime should come from tianwen.resolve's "lunar" field.
func ComputeChart(lt tianwen.LunarTime, gender ganzhi.Gender) (Chart, error) {
	date := tianwen.LunarDate{Year: lt.Year, Month: lt.Month, Day: lt.Day, Leap: lt.Leap}
	if _, err := tianwen.ValidateLunarDate(date); err != nil {
		return Chart{}, fmt.Errorf("validate lunar date: %w", err)
	}
	return computeValidatedChart(lt, gender), nil
}

// computeValidatedChart assembles a chart after the lunar date has been
// validated, or after it was produced by the calendar projection itself.
// Keeping this path unexported makes the public boundary remain fail-closed.
func computeValidatedChart(lt tianwen.LunarTime, gender ganzhi.Gender) Chart {
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

// ComputeChartAt is the explicit-time entry for new callers. at must already
// be normalized to UTC; loc is the explicit IANA location. Ziwei lunar
// projection currently follows the tianwen UTC+8 convention.
func ComputeChartAt(
	at time.Time,
	loc *time.Location,
	longitude float64,
	gender ganzhi.Gender,
) (Chart, error) {
	lunarDate := tianwen.LunarDateAt(at)
	shichen := tianwen.ShichenAt(at, loc, longitude)
	lunar := tianwen.LunarTime{
		Year:    lunarDate.Year,
		Month:   lunarDate.Month,
		Day:     lunarDate.Day,
		Leap:    lunarDate.Leap,
		Shichen: shichen,
	}
	// LunarDateAt projects a real civil date into the lunisolar calendar, so
	// validating that projection again repeats expensive new-moon and leap-month
	// work. Public callers that construct LunarTime still go through ComputeChart.
	return computeValidatedChart(lunar, gender), nil
}
