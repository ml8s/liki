package mcpcontract

import (
	"context"
	"encoding/json"
	"fmt"
	"time"

	"liki-engine/internal/engine/tianwen"
	"liki-engine/internal/engine/ziwei"
)

func ziweiLunarDate(lunar tianwen.LunarDate) ZiweiLunarDate {
	return ZiweiLunarDate{
		Year:  lunar.Year,
		Month: lunar.Month,
		Day:   lunar.Day,
		Leap:  lunar.Leap,
	}
}

func ziweiLunarProjector(loc *time.Location) (tianwen.LunarProjector, error) {
	return tianwen.NewLunarProjector(tianwen.LunarBasisLocalAstronomical, loc)
}

// domain decodes a ziwei.chart wire chart back into the domain chart. The wire
// DTO is a faithful projection of the domain JSON shape, so the round trip is
// lossless; the school/birth year checks keep the "empty chart" failure mode.

// ZiweiChart computes the twelve-palace chart from lunar birth data.
// ZiweiChart computes the canonical complete twelve-palace chart.
func ZiweiChartCompute(_ context.Context, input ZiweiChartInput) (json.RawMessage, error) {
	chart, err := input.Subject.ziweiChart()
	if err != nil {
		return nil, fmt.Errorf("ziwei.chart: %w", err)
	}
	return marshalRawOutput("ziwei.chart", chart)
}

// ZiweiLiupan computes one selected flow layer from the calibrated natal
// subject. The natal chart is prepared once and reused across the requested
// batch.
func ZiweiLiupan(_ context.Context, input ZiweiLiupanInput) (json.RawMessage, error) {
	if err := validateZiweiLiupan(input); err != nil {
		return nil, fmt.Errorf("ziwei.liupan: %w", err)
	}
	count := seriesCountValue(input.Count)
	chart, err := input.Subject.ziweiChart()
	if err != nil {
		return nil, fmt.Errorf("ziwei.liupan: %w", err)
	}
	current := input

	switch input.Scope {
	case "liu_nian":
		items := make([]ZiweiLiuNianOutput, 0, count)
		for range count {
			item, err := computeZiweiLiuNian(current, chart)
			if err != nil {
				return nil, err
			}
			items = append(items, item)
			current, err = advanceZiweiLiuNian(current)
			if err != nil {
				return nil, err
			}
		}
		return marshalRawOutput("ziwei.liupan", ZiweiLiuNianSeries{LiuNian: items})
	case "liu_yue":
		items := make([]ZiweiLiuYueOutput, 0, count)
		for range count {
			item, err := computeZiweiLiuYue(current, chart)
			if err != nil {
				return nil, err
			}
			items = append(items, item)
			current, err = advanceZiweiLiuYue(current)
			if err != nil {
				return nil, err
			}
		}
		return marshalRawOutput("ziwei.liupan", ZiweiLiuYueSeries{LiuYue: items})
	case "liu_ri":
		items := make([]ZiweiLiuRiOutput, 0, count)
		for range count {
			item, err := computeZiweiLiuRi(current, chart)
			if err != nil {
				return nil, err
			}
			items = append(items, item)
			current, err = advanceZiweiLiuRi(current)
			if err != nil {
				return nil, err
			}
		}
		return marshalRawOutput("ziwei.liupan", ZiweiLiuRiSeries{LiuRi: items})
	case "liu_shi":
		items := make([]ZiweiLiuShiOutput, 0, count)
		for range count {
			item, err := computeZiweiLiuShi(current, chart)
			if err != nil {
				return nil, err
			}
			items = append(items, item)
			current, err = advanceZiweiLiuShi(current)
			if err != nil {
				return nil, err
			}
		}
		return marshalRawOutput("ziwei.liupan", ZiweiLiuShiSeries{LiuShi: items})
	default:
		return nil, fmt.Errorf("ziwei.liupan: invalid scope %q", input.Scope)
	}
}

func validateZiweiLiupan(input ZiweiLiupanInput) error {
	if input.StandardTime == "" {
		return fmt.Errorf("standard_time is required")
	}
	switch input.Scope {
	case "liu_nian", "liu_yue", "liu_ri":
		if input.Longitude != nil {
			return fmt.Errorf("longitude is not allowed")
		}
	case "liu_shi":
		if input.Longitude == nil {
			return fmt.Errorf("longitude is required")
		}
		if *input.Longitude < -180 || *input.Longitude > 180 {
			return fmt.Errorf("longitude must be within [-180,180], got %v", *input.Longitude)
		}
	default:
		return fmt.Errorf("invalid scope %q", input.Scope)
	}
	return nil
}

func parseZiweiLiuNianTime(input ZiweiLiupanInput) (ziwei.LiuNianTime, tianwen.CivilDate, time.Time, *time.Location, tianwen.LunarProjector, error) {
	standard, loc, err := parseStandardTime(input.StandardTime)
	if err != nil {
		return ziwei.LiuNianTime{}, tianwen.CivilDate{}, time.Time{}, nil, tianwen.LunarProjector{}, err
	}
	projector, err := ziweiLunarProjector(loc)
	if err != nil {
		return ziwei.LiuNianTime{}, tianwen.CivilDate{}, time.Time{}, nil, tianwen.LunarProjector{}, err
	}
	date := projector.SelectorDate(standard)
	resolved, err := ziwei.ResolveLiuNianTimeWith(projector, date)
	return resolved, date, standard, loc, projector, err
}

func computeZiweiLiuNian(input ZiweiLiupanInput, chart ziwei.Chart) (ZiweiLiuNianOutput, error) {
	resolved, date, at, _, _, err := parseZiweiLiuNianTime(input)
	if err != nil {
		return ZiweiLiuNianOutput{}, fmt.Errorf("ziwei.liupan liu_nian: %w", err)
	}
	result := ziwei.ComputeLiuNian(chart.FlowContext(), resolved.LunarYear)
	output, err := wireFromDomain[ZiweiLiuNianOutput](result)
	if err != nil {
		return ZiweiLiuNianOutput{}, fmt.Errorf("ziwei.liupan liu_nian: %w", err)
	}
	output.NianJie = ZiweiLiuNianResolved{
		LiuNianStandardTime: standardTimeValue(at),
		LiuNianStartDate:    civilDateString(date),
		LunarYear:           resolved.LunarYear,
		LunarYearStart:      civilDateString(resolved.Start),
		LunarYearEnd:        civilDateString(resolved.End),
		BoundaryDay:         resolved.BoundaryDay,
	}
	return output, nil
}

func advanceZiweiLiuNian(input ZiweiLiupanInput) (ZiweiLiupanInput, error) {
	resolved, _, _, _, projector, err := parseZiweiLiuNianTime(input)
	if err != nil {
		return input, fmt.Errorf("ziwei.liupan liu_nian: %w", err)
	}
	next, err := ziwei.NextLiuNianTimeWith(projector, resolved)
	if err != nil {
		return input, fmt.Errorf("ziwei.liupan liu_nian: %w", err)
	}
	input.StandardTime = standardTimeValue(projector.SelectorInstant(next.Start))
	return input, nil
}

func parseZiweiLiuYueTime(input ZiweiLiupanInput) (ziwei.LiuYueTime, tianwen.CivilDate, time.Time, *time.Location, tianwen.LunarProjector, error) {
	standard, loc, err := parseStandardTime(input.StandardTime)
	if err != nil {
		return ziwei.LiuYueTime{}, tianwen.CivilDate{}, time.Time{}, nil, tianwen.LunarProjector{}, err
	}
	projector, err := ziweiLunarProjector(loc)
	if err != nil {
		return ziwei.LiuYueTime{}, tianwen.CivilDate{}, time.Time{}, nil, tianwen.LunarProjector{}, err
	}
	date := projector.SelectorDate(standard)
	resolved, err := ziwei.ResolveLiuYueTimeWith(projector, date)
	return resolved, date, standard, loc, projector, err
}

func computeZiweiLiuYue(input ZiweiLiupanInput, chart ziwei.Chart) (ZiweiLiuYueOutput, error) {
	resolved, date, at, _, projector, err := parseZiweiLiuYueTime(input)
	if err != nil {
		return ZiweiLiuYueOutput{}, fmt.Errorf("ziwei.liupan liu_yue: %w", err)
	}
	result, err := ziwei.ComputeLiuYueWith(chart.FlowContext(), resolved.Lunar, projector)
	if err != nil {
		return ZiweiLiuYueOutput{}, fmt.Errorf("ziwei.liupan liu_yue: %w", err)
	}
	output, err := wireFromDomain[ZiweiLiuYueOutput](result)
	if err != nil {
		return ZiweiLiuYueOutput{}, fmt.Errorf("ziwei.liupan liu_yue: %w", err)
	}
	calendar, err := wireFromDomain[ZiweiLiuYueResolvedPeriodCalendarPeriod](result.ResolvedPeriod.CalendarPeriod)
	if err != nil {
		return ZiweiLiuYueOutput{}, fmt.Errorf("ziwei.liupan liu_yue: %w", err)
	}
	flowMonth, err := wireFromDomain[ZiweiLiuYueResolvedPeriodFlowMonth](result.ResolvedPeriod.FlowMonth)
	if err != nil {
		return ZiweiLiuYueOutput{}, fmt.Errorf("ziwei.liupan liu_yue: %w", err)
	}
	output.YueJie = ZiweiYueJie{
		LiuYueStandardTime: standardTimeValue(at),
		LiuYueStartDate:    civilDateString(date),
		Lunar:              ziweiLunarDate(resolved.Lunar),
		LunarMonthStart:    civilDateString(resolved.Start),
		LunarMonthEnd:      civilDateString(resolved.End),
		BoundaryDay:        resolved.BoundaryDay,
		CalendarPeriod:     calendar,
		FlowMonth:          flowMonth,
	}
	return output, nil
}

func advanceZiweiLiuYue(input ZiweiLiupanInput) (ZiweiLiupanInput, error) {
	resolved, _, _, _, projector, err := parseZiweiLiuYueTime(input)
	if err != nil {
		return input, fmt.Errorf("ziwei.liupan liu_yue: %w", err)
	}
	next, err := ziwei.NextLiuYueTimeWith(projector, resolved)
	if err != nil {
		return input, fmt.Errorf("ziwei.liupan liu_yue: %w", err)
	}
	input.StandardTime = standardTimeValue(projector.SelectorInstant(next.Date))
	return input, nil
}

func parseZiweiLiuRiTime(input ZiweiLiupanInput) (ziwei.LiuRiTime, tianwen.LunarProjector, time.Time, *time.Location, error) {
	standard, loc, err := parseStandardTime(input.StandardTime)
	if err != nil {
		return ziwei.LiuRiTime{}, tianwen.LunarProjector{}, time.Time{}, nil, err
	}
	projector, err := ziweiLunarProjector(loc)
	if err != nil {
		return ziwei.LiuRiTime{}, tianwen.LunarProjector{}, time.Time{}, nil, err
	}
	resolved, err := ziwei.ResolveLiuRiTimeAtWith(projector, standard, loc)
	return resolved, projector, standard, loc, err
}

func computeZiweiLiuRi(input ZiweiLiupanInput, chart ziwei.Chart) (ZiweiLiuRiOutput, error) {
	resolved, projector, _, loc, err := parseZiweiLiuRiTime(input)
	if err != nil {
		return ZiweiLiuRiOutput{}, fmt.Errorf("ziwei.liupan liu_ri: %w", err)
	}
	result, err := ziwei.ComputeLiuRiWith(chart.FlowContext(), resolved.Lunar, projector)
	if err != nil {
		return ZiweiLiuRiOutput{}, fmt.Errorf("ziwei.liupan liu_ri: %w", err)
	}
	output, err := wireFromDomain[ZiweiLiuRiOutput](result)
	if err != nil {
		return ZiweiLiuRiOutput{}, fmt.Errorf("ziwei.liupan liu_ri: %w", err)
	}
	output.RiJie = ZiweiLiuRiJie{
		LiuRiStandardTime: standardTimeValue(resolved.Standard.In(loc)),
		LiuRiLocalDate:    civilDateString(resolved.LocalDate),
		Lunar:             ziweiLunarDate(resolved.Lunar),
	}
	return output, nil
}

func advanceZiweiLiuRi(input ZiweiLiupanInput) (ZiweiLiupanInput, error) {
	resolved, projector, _, loc, err := parseZiweiLiuRiTime(input)
	if err != nil {
		return input, fmt.Errorf("ziwei.liupan liu_ri: %w", err)
	}
	next, err := ziwei.NextLiuRiTimeWith(projector, resolved)
	if err != nil {
		return input, fmt.Errorf("ziwei.liupan liu_ri: %w", err)
	}
	input.StandardTime = standardTimeValue(next.Standard.In(loc))
	return input, nil
}

func parseZiweiLiuShiTime(input ZiweiLiupanInput) (ziwei.LiuShiTime, tianwen.LunarProjector, time.Time, *time.Location, error) {
	if input.Longitude == nil {
		return ziwei.LiuShiTime{}, tianwen.LunarProjector{}, time.Time{}, nil, fmt.Errorf("longitude is required")
	}
	standard, loc, err := parseStandardTime(input.StandardTime)
	if err != nil {
		return ziwei.LiuShiTime{}, tianwen.LunarProjector{}, time.Time{}, nil, err
	}
	projector, err := ziweiLunarProjector(loc)
	if err != nil {
		return ziwei.LiuShiTime{}, tianwen.LunarProjector{}, time.Time{}, nil, err
	}
	resolved, err := ziwei.ResolveLiuShiTimeWith(projector, standard, loc, float64(*input.Longitude))
	return resolved, projector, standard, loc, err
}

func computeZiweiLiuShi(input ZiweiLiupanInput, chart ziwei.Chart) (ZiweiLiuShiOutput, error) {
	resolved, projector, _, loc, err := parseZiweiLiuShiTime(input)
	if err != nil {
		return ZiweiLiuShiOutput{}, fmt.Errorf("ziwei.liupan liu_shi: %w", err)
	}
	result, err := ziwei.ComputeLiuShiAtTimeWith(chart.FlowContext(), resolved, projector)
	if err != nil {
		return ZiweiLiuShiOutput{}, fmt.Errorf("ziwei.liupan liu_shi: %w", err)
	}
	output, err := wireFromDomain[ZiweiLiuShiOutput](result)
	if err != nil {
		return ZiweiLiuShiOutput{}, fmt.Errorf("ziwei.liupan liu_shi: %w", err)
	}
	output.ShiChen = ZiweiLiuShiChen{
		LiuShiStandardTime: standardTimeValue(resolved.Standard.In(loc)),
		LiuShiLocalDate:    civilDateString(resolved.LocalDate),
		Lunar:              ziweiLunarDate(resolved.Lunar),
		Shichen:            ZhiName(resolved.Shichen.String()),
	}
	return output, nil
}

func advanceZiweiLiuShi(input ZiweiLiupanInput) (ZiweiLiupanInput, error) {
	resolved, projector, _, loc, err := parseZiweiLiuShiTime(input)
	if err != nil {
		return input, fmt.Errorf("ziwei.liupan liu_shi: %w", err)
	}
	next, err := ziwei.NextLiuShiTimeWith(projector, resolved)
	if err != nil {
		return input, fmt.Errorf("ziwei.liupan liu_shi: %w", err)
	}
	input.StandardTime = standardTimeValue(next.Standard.In(loc))
	return input, nil
}

// ZiweiBond returns raw cross-chart facts for two charts.
func ZiweiBond(_ context.Context, input ZiweiBondInput) (json.RawMessage, error) {
	chartA, err := input.A.Subject.ziweiChart()
	if err != nil {
		return nil, fmt.Errorf("ziwei.bond: chart a: %w", err)
	}
	chartB, err := input.B.Subject.ziweiChart()
	if err != nil {
		return nil, fmt.Errorf("ziwei.bond: chart b: %w", err)
	}
	raw, err := marshalRawOutput("ziwei.bond", ziwei.ComputeBond(chartA.BondSideFacts(), chartB.BondSideFacts()))
	if err != nil {
		return nil, fmt.Errorf("ziwei.bond: %w", err)
	}
	return raw, nil
}
