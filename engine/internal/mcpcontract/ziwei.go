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
func ZiweiChartCompute(_ context.Context, input ZiweiChartInput) (json.RawMessage, error) {
	chart, err := input.Subject.ziweiChart()
	if err != nil {
		return nil, fmt.Errorf("ziwei.chart: %w", err)
	}
	raw, err := marshalRawOutput("ziwei.chart", chart)
	if err != nil {
		return nil, fmt.Errorf("ziwei.chart: %w", err)
	}
	var item map[string]any
	if err := json.Unmarshal(raw, &item); err != nil {
		return nil, fmt.Errorf("ziwei.chart: decode chart: %w", err)
	}
	digest, err := chartDigest(item)
	if err != nil {
		return nil, fmt.Errorf("ziwei.chart: %w", err)
	}
	item["chart_digest"] = digest
	return json.Marshal(item)
}

// ZiweiFullChart expands a chart with 杂曜、长生、博士、小限、将前、岁前.
func ZiweiFullChart(_ context.Context, input ZiweiFullChartInput) (json.RawMessage, error) {
	chart, err := input.Subject.ziweiChart()
	if err != nil {
		return nil, fmt.Errorf("ziwei.fullchart: %w", err)
	}
	raw, err := marshalRawOutput("ziwei.fullchart", ziwei.ComputeFullChart(chart))
	if err != nil {
		return nil, fmt.Errorf("ziwei.fullchart: %w", err)
	}
	return raw, nil
}

// ZiweiDaXian returns the ten-year 大限 steps of a chart.
func ZiweiDaXian(_ context.Context, input ZiweiDaXianInput) (ZiweiDaXianOutput, error) {
	chart, err := input.Subject.ziweiChart()
	if err != nil {
		return ZiweiDaXianOutput{}, fmt.Errorf("ziwei.daxian: %w", err)
	}
	items, err := wireFromDomain[[]ZiweiDaXianItem](ziwei.ComputeDaXian(chart))
	if err != nil {
		return ZiweiDaXianOutput{}, fmt.Errorf("ziwei.daxian: %w", err)
	}
	return ZiweiDaXianOutput{Items: items}, nil
}

// ZiweiLiuNian returns the flow year palace and its 四化.
func ZiweiLiuNian(_ context.Context, input ZiweiLiuNianInput) (any, error) {
	return runSeriesPrepared(input, seriesCountValue(input.Count), MaxSeriesCount,
		func(in ZiweiLiuNianInput) (ziwei.Chart, error) {
			return in.Subject.ziweiChart()
		},
		func(in ZiweiLiuNianInput, chart ziwei.Chart) (ZiweiLiuNianOutput, error) {
			resolved, date, at, _, _, err := parseZiweiLiuNianTime(in)
			if err != nil {
				return ZiweiLiuNianOutput{}, fmt.Errorf("ziwei.liunian: %w", err)
			}
			output, err := wireFromDomain[ZiweiLiuNianOutput](
				ziwei.ComputeLiuNian(chart, resolved.LunarYear))
			if err != nil {
				return ZiweiLiuNianOutput{}, fmt.Errorf("ziwei.liunian: %w", err)
			}
			output.Resolved = ZiweiLiuNianResolved{
				SelectorStandardTime: standardTimeValue(at),
				SelectorDate:         civilDateString(date),
				LunarYear:            resolved.LunarYear,
				LunarYearStart:       civilDateString(resolved.Start),
				LunarYearEnd:         civilDateString(resolved.End),
				BoundaryDay:          resolved.BoundaryDay,
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
			in.StartTime = standardTimeValue(projector.SelectorInstant(next.Start))
			return in, nil
		},
	)
}

func parseZiweiLiuNianTime(input ZiweiLiuNianInput) (ziwei.LiuNianTime, tianwen.CivilDate, time.Time, *time.Location, tianwen.LunarProjector, error) {
	standard, loc, err := parseStandardTime(input.StartTime)
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
func ZiweiLiuYue(_ context.Context, input ZiweiLiuYueInput) (any, error) {
	return runSeriesPrepared(input, seriesCountValue(input.Count), MaxSeriesCount,
		func(in ZiweiLiuYueInput) (ziwei.Chart, error) {
			return in.Subject.ziweiChart()
		},
		func(in ZiweiLiuYueInput, chart ziwei.Chart) (ZiweiLiuYueOutput, error) {
			resolved, date, at, _, projector, err := parseZiweiLiuYueTime(in)
			if err != nil {
				return ZiweiLiuYueOutput{}, fmt.Errorf("ziwei.liuyue: %w", err)
			}
			result, err := ziwei.ComputeLiuYueWith(chart, resolved.Lunar, projector)
			if err != nil {
				return ZiweiLiuYueOutput{}, fmt.Errorf("ziwei.liuyue: %w", err)
			}
			output, err := wireFromDomain[ZiweiLiuYueOutput](result)
			if err != nil {
				return ZiweiLiuYueOutput{}, fmt.Errorf("ziwei.liuyue: %w", err)
			}
			output.Resolved = ZiweiLiuYueResolved{
				SelectorStandardTime: standardTimeValue(at),
				SelectorDate:         civilDateString(date),
				Lunar:                ziweiLunarDate(resolved.Lunar),
				LunarMonthStart:      civilDateString(resolved.Start),
				LunarMonthEnd:        civilDateString(resolved.End),
				BoundaryDay:          resolved.BoundaryDay,
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
			in.StartTime = standardTimeValue(projector.SelectorInstant(next.Date))
			return in, nil
		},
	)
}

func parseZiweiLiuYueTime(input ZiweiLiuYueInput) (ziwei.LiuYueTime, tianwen.CivilDate, time.Time, *time.Location, tianwen.LunarProjector, error) {
	standard, loc, err := parseStandardTime(input.StartTime)
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
func ZiweiLiuRi(_ context.Context, input ZiweiLiuRiInput) (any, error) {
	return runSeriesPrepared(input, seriesCountValue(input.Count), MaxSeriesCount,
		func(in ZiweiLiuRiInput) (ziwei.Chart, error) {
			return in.Subject.ziweiChart()
		},
		func(in ZiweiLiuRiInput, chart ziwei.Chart) (ZiweiLiuRiOutput, error) {
			resolved, projector, _, loc, err := parseZiweiLiuRiTime(in)
			if err != nil {
				return ZiweiLiuRiOutput{}, fmt.Errorf("ziwei.liuri: %w", err)
			}
			result, err := ziwei.ComputeLiuRiWith(chart, resolved.Lunar, projector)
			if err != nil {
				return ZiweiLiuRiOutput{}, fmt.Errorf("ziwei.liuri: %w", err)
			}
			output, err := wireFromDomain[ZiweiLiuRiOutput](result)
			if err != nil {
				return ZiweiLiuRiOutput{}, fmt.Errorf("ziwei.liuri: %w", err)
			}
			output.Lunar = ziweiLunarDate(resolved.Lunar)
			output.Resolved = ZiweiLiuRiResolved{
				StandardTime: standardTimeValue(resolved.Standard.In(loc)),
				LocalDate:    civilDateString(resolved.LocalDate),
				BeijingDate:  civilDateString(resolved.BeijingDate),
				Lunar:        ziweiLunarDate(resolved.Lunar),
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
			in.StartTime = standardTimeValue(next.Standard.In(loc))
			return in, nil
		},
	)
}

func parseZiweiLiuRiTime(input ZiweiLiuRiInput) (ziwei.LiuRiTime, tianwen.LunarProjector, time.Time, *time.Location, error) {
	standard, loc, err := parseStandardTime(input.StartTime)
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
func ZiweiLiuShi(_ context.Context, input ZiweiLiuShiInput) (any, error) {
	return runSeriesPrepared(input, seriesCountValue(input.Count), MaxSeriesCount,
		func(in ZiweiLiuShiInput) (ziwei.Chart, error) {
			return in.Subject.ziweiChart()
		},
		func(in ZiweiLiuShiInput, chart ziwei.Chart) (ZiweiLiuShiOutput, error) {
			resolved, projector, _, loc, err := parseZiweiLiuShiTime(in)
			if err != nil {
				return ZiweiLiuShiOutput{}, fmt.Errorf("ziwei.liushi: %w", err)
			}
			result, err := ziwei.ComputeLiuShiAtTimeWith(chart, resolved, projector)
			if err != nil {
				return ZiweiLiuShiOutput{}, fmt.Errorf("ziwei.liushi: %w", err)
			}
			output, err := wireFromDomain[ZiweiLiuShiOutput](result)
			if err != nil {
				return ZiweiLiuShiOutput{}, fmt.Errorf("ziwei.liushi: %w", err)
			}
			output.Resolved = ZiweiLiuShiResolved{
				StandardTime: standardTimeValue(resolved.Standard.In(loc)),
				LocalDate:    civilDateString(resolved.LocalDate),
				Lunar:        ziweiLunarDate(resolved.Lunar),
				Shichen:      resolved.Shichen.String(),
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
			in.StartTime = standardTimeValue(next.Standard.In(loc))
			return in, nil
		},
	)
}

func parseZiweiLiuShiTime(input ZiweiLiuShiInput) (ziwei.LiuShiTime, tianwen.LunarProjector, time.Time, *time.Location, error) {
	standard, loc, err := parseStandardTime(input.StartTime)
	if err != nil {
		return ziwei.LiuShiTime{}, tianwen.LunarProjector{}, time.Time{}, nil, err
	}
	projector, err := ziweiLunarProjector(loc)
	if err != nil {
		return ziwei.LiuShiTime{}, tianwen.LunarProjector{}, time.Time{}, nil, err
	}
	resolved, err := ziwei.ResolveLiuShiTimeWith(projector, standard, loc, input.Longitude)
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
	raw, err := marshalRawOutput("ziwei.bond", ziwei.ComputeBond(chartA, chartB))
	if err != nil {
		return nil, fmt.Errorf("ziwei.bond: %w", err)
	}
	return raw, nil
}
