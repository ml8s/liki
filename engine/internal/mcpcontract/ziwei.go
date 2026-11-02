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

// ZiweiLiuNian returns the flow year palace and its 四化.
func ZiweiLiuNian(_ context.Context, input ZiweiLiuNianInput) (ZiweiLiuNianSeries, error) {
	items, err := runSeriesPrepared(input, seriesCountValue(input.LiuNianCount), MaxSeriesCount,

		func(in ZiweiLiuNianInput) (ziwei.Chart, error) {
			return in.Subject.ziweiChart()
		},
		func(in ZiweiLiuNianInput, chart ziwei.Chart) (ZiweiLiuNianOutput, error) {
			resolved, date, at, _, _, err := parseZiweiLiuNianTime(in)
			if err != nil {
				return ZiweiLiuNianOutput{}, fmt.Errorf("ziwei.liunian: %w", err)
			}
			output, err := wireFromDomain[ZiweiLiuNianOutput](
				ziwei.ComputeLiuNian(chart.FlowContext(), resolved.LunarYear))
			if err != nil {
				return ZiweiLiuNianOutput{}, fmt.Errorf("ziwei.liunian: %w", err)
			}
			output.NianJie = ZiweiLiuNianResolved{
				LiuNianStartTime: standardTimeValue(at),
				LiuNianStartDate: civilDateString(date),
				LunarYear:        resolved.LunarYear,
				LunarYearStart:   civilDateString(resolved.Start),
				LunarYearEnd:     civilDateString(resolved.End),
				BoundaryDay:      resolved.BoundaryDay,
			}
			return output, nil
		},
		func(in ZiweiLiuNianInput) (ZiweiLiuNianInput, error) {
			resolved, _, _, _, projector, err := parseZiweiLiuNianTime(in)
			if err != nil {
				return in, fmt.Errorf("ziwei.liunian: %w", err)
			}
			next, err := ziwei.NextLiuNianTimeWith(projector, resolved)
			if err != nil {
				return in, fmt.Errorf("ziwei.liunian: %w", err)
			}
			in.LiuNianStartTime = standardTimeValue(projector.SelectorInstant(next.Start))
			return in, nil
		},
	)
	if err != nil {
		return ZiweiLiuNianSeries{}, err
	}
	return ZiweiLiuNianSeries{LiuNian: items}, nil
}

func parseZiweiLiuNianTime(input ZiweiLiuNianInput) (ziwei.LiuNianTime, tianwen.CivilDate, time.Time, *time.Location, tianwen.LunarProjector, error) {
	standard, loc, err := parseStandardTime(input.LiuNianStartTime)
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

// ZiweiLiuYue returns the flow month palace and its 四化.
func ZiweiLiuYue(_ context.Context, input ZiweiLiuYueInput) (ZiweiLiuYueSeries, error) {
	items, err := runSeriesPrepared(input, seriesCountValue(input.LiuYueCount), MaxSeriesCount,

		func(in ZiweiLiuYueInput) (ziwei.Chart, error) {
			return in.Subject.ziweiChart()
		},
		func(in ZiweiLiuYueInput, chart ziwei.Chart) (ZiweiLiuYueOutput, error) {
			resolved, date, at, _, projector, err := parseZiweiLiuYueTime(in)
			if err != nil {
				return ZiweiLiuYueOutput{}, fmt.Errorf("ziwei.liuyue: %w", err)
			}
			result, err := ziwei.ComputeLiuYueWith(chart.FlowContext(), resolved.Lunar, projector)
			if err != nil {
				return ZiweiLiuYueOutput{}, fmt.Errorf("ziwei.liuyue: %w", err)
			}
			output, err := wireFromDomain[ZiweiLiuYueOutput](result)
			if err != nil {
				return ZiweiLiuYueOutput{}, fmt.Errorf("ziwei.liuyue: %w", err)
			}
			calendar, err := wireFromDomain[ZiweiLiuYueResolvedPeriodCalendarPeriod](result.ResolvedPeriod.CalendarPeriod)
			if err != nil {
				return ZiweiLiuYueOutput{}, fmt.Errorf("ziwei.liuyue: %w", err)
			}
			flowMonth, err := wireFromDomain[ZiweiLiuYueResolvedPeriodFlowMonth](result.ResolvedPeriod.FlowMonth)
			if err != nil {
				return ZiweiLiuYueOutput{}, fmt.Errorf("ziwei.liuyue: %w", err)
			}
			output.YueJie = ZiweiYueJie{
				LiuYueStartTime: standardTimeValue(at),
				LiuYueStartDate: civilDateString(date),
				Lunar:           ziweiLunarDate(resolved.Lunar),
				LunarMonthStart: civilDateString(resolved.Start),
				LunarMonthEnd:   civilDateString(resolved.End),
				BoundaryDay:     resolved.BoundaryDay,
				CalendarPeriod:  calendar,
				FlowMonth:       flowMonth,
			}
			return output, nil
		},
		func(in ZiweiLiuYueInput) (ZiweiLiuYueInput, error) {
			resolved, _, _, _, projector, err := parseZiweiLiuYueTime(in)
			if err != nil {
				return in, fmt.Errorf("ziwei.liuyue: %w", err)
			}
			next, err := ziwei.NextLiuYueTimeWith(projector, resolved)
			if err != nil {
				return in, fmt.Errorf("ziwei.liuyue: %w", err)
			}
			in.LiuYueStartTime = standardTimeValue(projector.SelectorInstant(next.Date))
			return in, nil
		},
	)
	if err != nil {
		return ZiweiLiuYueSeries{}, err
	}
	return ZiweiLiuYueSeries{LiuYue: items}, nil
}

func parseZiweiLiuYueTime(input ZiweiLiuYueInput) (ziwei.LiuYueTime, tianwen.CivilDate, time.Time, *time.Location, tianwen.LunarProjector, error) {
	standard, loc, err := parseStandardTime(input.LiuYueStartTime)
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

// ZiweiLiuRi returns the flow day palace and its 四化.
func ZiweiLiuRi(_ context.Context, input ZiweiLiuRiInput) (ZiweiLiuRiSeries, error) {
	items, err := runSeriesPrepared(input, seriesCountValue(input.LiuRiCount), MaxSeriesCount,

		func(in ZiweiLiuRiInput) (ziwei.Chart, error) {
			return in.Subject.ziweiChart()
		},
		func(in ZiweiLiuRiInput, chart ziwei.Chart) (ZiweiLiuRiOutput, error) {
			resolved, projector, _, loc, err := parseZiweiLiuRiTime(in)
			if err != nil {
				return ZiweiLiuRiOutput{}, fmt.Errorf("ziwei.liuri: %w", err)
			}
			result, err := ziwei.ComputeLiuRiWith(chart.FlowContext(), resolved.Lunar, projector)
			if err != nil {
				return ZiweiLiuRiOutput{}, fmt.Errorf("ziwei.liuri: %w", err)
			}
			output, err := wireFromDomain[ZiweiLiuRiOutput](result)
			if err != nil {
				return ZiweiLiuRiOutput{}, fmt.Errorf("ziwei.liuri: %w", err)
			}
			output.RiJie = ZiweiLiuRiJie{
				LiuRiStartTime: standardTimeValue(resolved.Standard.In(loc)),
				LiuRiLocalDate: civilDateString(resolved.LocalDate),
				BeijingDate:    civilDateString(resolved.BeijingDate),
				Lunar:          ziweiLunarDate(resolved.Lunar),
			}
			return output, nil
		},
		func(in ZiweiLiuRiInput) (ZiweiLiuRiInput, error) {
			resolved, projector, _, loc, err := parseZiweiLiuRiTime(in)
			if err != nil {
				return in, fmt.Errorf("ziwei.liuri: %w", err)
			}
			next, err := ziwei.NextLiuRiTimeWith(projector, resolved)
			if err != nil {
				return in, fmt.Errorf("ziwei.liuri: %w", err)
			}
			in.LiuRiStartTime = standardTimeValue(next.Standard.In(loc))
			return in, nil
		},
	)
	if err != nil {
		return ZiweiLiuRiSeries{}, err
	}
	return ZiweiLiuRiSeries{LiuRi: items}, nil
}

func parseZiweiLiuRiTime(input ZiweiLiuRiInput) (ziwei.LiuRiTime, tianwen.LunarProjector, time.Time, *time.Location, error) {
	standard, loc, err := parseStandardTime(input.LiuRiStartTime)
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

// ZiweiLiuShi returns the flow hour palace and its 四化.
func ZiweiLiuShi(_ context.Context, input ZiweiLiuShiInput) (ZiweiLiuShiSeries, error) {
	items, err := runSeriesPrepared(input, seriesCountValue(input.LiuShiCount), MaxSeriesCount,

		func(in ZiweiLiuShiInput) (ziwei.Chart, error) {
			return in.Subject.ziweiChart()
		},
		func(in ZiweiLiuShiInput, chart ziwei.Chart) (ZiweiLiuShiOutput, error) {
			resolved, projector, _, loc, err := parseZiweiLiuShiTime(in)
			if err != nil {
				return ZiweiLiuShiOutput{}, fmt.Errorf("ziwei.liushi: %w", err)
			}
			result, err := ziwei.ComputeLiuShiAtTimeWith(chart.FlowContext(), resolved, projector)
			if err != nil {
				return ZiweiLiuShiOutput{}, fmt.Errorf("ziwei.liushi: %w", err)
			}
			output, err := wireFromDomain[ZiweiLiuShiOutput](result)
			if err != nil {
				return ZiweiLiuShiOutput{}, fmt.Errorf("ziwei.liushi: %w", err)
			}
			output.ShiChen = ZiweiLiuShiChen{
				LiuShiStartTime: standardTimeValue(resolved.Standard.In(loc)),
				LiuShiLocalDate: civilDateString(resolved.LocalDate),
				Lunar:           ziweiLunarDate(resolved.Lunar),
				Shichen:         resolved.Shichen.String(),
			}
			return output, nil
		},
		func(in ZiweiLiuShiInput) (ZiweiLiuShiInput, error) {
			resolved, projector, _, loc, err := parseZiweiLiuShiTime(in)
			if err != nil {
				return in, fmt.Errorf("ziwei.liushi: %w", err)
			}
			next, err := ziwei.NextLiuShiTimeWith(projector, resolved)
			if err != nil {
				return in, fmt.Errorf("ziwei.liushi: %w", err)
			}
			in.LiuShiStartTime = standardTimeValue(next.Standard.In(loc))
			return in, nil
		},
	)
	if err != nil {
		return ZiweiLiuShiSeries{}, err
	}
	return ZiweiLiuShiSeries{LiuShi: items}, nil
}

func parseZiweiLiuShiTime(input ZiweiLiuShiInput) (ziwei.LiuShiTime, tianwen.LunarProjector, time.Time, *time.Location, error) {
	standard, loc, err := parseStandardTime(input.LiuShiStartTime)
	if err != nil {
		return ziwei.LiuShiTime{}, tianwen.LunarProjector{}, time.Time{}, nil, err
	}
	projector, err := ziweiLunarProjector(loc)
	if err != nil {
		return ziwei.LiuShiTime{}, tianwen.LunarProjector{}, time.Time{}, nil, err
	}
	resolved, err := ziwei.ResolveLiuShiTimeWith(projector, standard, loc, input.LiuShiLongitude)
	return resolved, projector, standard, loc, err
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
