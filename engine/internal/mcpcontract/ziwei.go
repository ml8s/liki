package mcpcontract

import (
	"context"
	"encoding/json"
	"fmt"

	"liki-engine/internal/engine/ganzhi"
	"liki-engine/internal/engine/tianwen"
	"liki-engine/internal/engine/ziwei"
)

// ZiweiTargetLunarInput is the resolved lunar date used by the flow tools.
type ZiweiTargetLunarInput struct {
	Year  int  `json:"year" jsonschema:"农历年份（1900-2100）。"`
	Month int  `json:"month" jsonschema:"农历月份（1-12）。"`
	Day   int  `json:"day" jsonschema:"农历日期（1-30）。"`
	Leap  bool `json:"leap" jsonschema:"是否闰月。"`
}

func (in ZiweiTargetLunarInput) domain() tianwen.LunarDate {
	return tianwen.LunarDate{Year: in.Year, Month: in.Month, Day: in.Day, Leap: in.Leap}
}

// domain decodes a ziwei.chart wire chart back into the domain chart. The wire
// DTO is a faithful projection of the domain JSON shape, so the round trip is
// lossless; the school/birth year checks keep the "empty chart" failure mode.
func (chart ZiweiChartOutput) domain() (ziwei.Chart, error) {
	raw, err := json.Marshal(chart)
	if err != nil {
		return ziwei.Chart{}, err
	}
	var out ziwei.Chart
	if err := json.Unmarshal(raw, &out); err != nil {
		return out, fmt.Errorf("decode ziwei chart: %w", err)
	}
	if out.School.LeapMonth == "" || out.School.Source == "" {
		return out, fmt.Errorf("chart is empty (use ziwei.chart output)")
	}
	if out.BirthYear == 0 {
		return out, fmt.Errorf("chart has no birth year (use ziwei.chart output)")
	}
	return out, nil
}

// ZiweiChart computes the twelve-palace chart from lunar birth data.
func ZiweiChartCompute(_ context.Context, input ZiweiChartInput) (ZiweiChartOutput, error) {
	at, err := parseInstant(string(input.Instant))
	if err != nil {
		return ZiweiChartOutput{}, fmt.Errorf("ziwei.chart: %w", err)
	}
	loc, err := loadTimezone(input.Timezone)
	if err != nil {
		return ZiweiChartOutput{}, fmt.Errorf("ziwei.chart: %w", err)
	}
	if input.Longitude < -180 || input.Longitude > 180 {
		return ZiweiChartOutput{}, fmt.Errorf("ziwei.chart: longitude must be within [-180,180], got %v", input.Longitude)
	}
	if _, err := ganzhi.ParseGender(string(input.Gender)); err != nil {
		return ZiweiChartOutput{}, fmt.Errorf("ziwei.chart: gender must be male or female, got %q", input.Gender)
	}
	chart, err := ziwei.ComputeChartAt(at, loc, input.Longitude, ganzhi.Gender(input.Gender))
	if err != nil {
		return ZiweiChartOutput{}, fmt.Errorf("ziwei.chart: %w", err)
	}
	return wireFromDomain[ZiweiChartOutput](chart)
}

// ZiweiFullChart expands a chart with 杂曜、长生、博士、小限、将前、岁前.
func ZiweiFullChart(_ context.Context, input ZiweiFullChartInput) (ZiweiFullChartOutput, error) {
	chart, err := input.Chart.domain()
	if err != nil {
		return ZiweiFullChartOutput{}, fmt.Errorf("ziwei.fullchart: %w", err)
	}
	return wireFromDomain[ZiweiFullChartOutput](ziwei.ComputeFullChart(chart))
}

// ZiweiDaXian returns the ten-year 大限 steps of a chart.
func ZiweiDaXian(_ context.Context, input ZiweiDaXianInput) (ZiweiDaXianOutput, error) {
	chart, err := input.Chart.domain()
	if err != nil {
		return ZiweiDaXianOutput{}, fmt.Errorf("ziwei.daxian: %w", err)
	}
	return wireFromDomain[ZiweiDaXianOutput](ziwei.ComputeDaXian(chart))
}

// ZiweiLiuNian returns the flow year palace and its 四化.
func ZiweiLiuNian(_ context.Context, input ZiweiLiuNianInput) (ZiweiLiuNianOutput, error) {
	chart, err := input.Chart.domain()
	if err != nil {
		return ZiweiLiuNianOutput{}, fmt.Errorf("ziwei.liunian: %w", err)
	}
	return wireFromDomain[ZiweiLiuNianOutput](ziwei.ComputeLiuNian(chart, input.LunarYear))
}

// ZiweiLiuYue returns the flow month palace and its 四化.
func ZiweiLiuYue(_ context.Context, input ZiweiLiuYueInput) (ZiweiLiuYueOutput, error) {
	chart, err := input.Chart.domain()
	if err != nil {
		return ZiweiLiuYueOutput{}, fmt.Errorf("ziwei.liuyue: %w", err)
	}
	result, err := ziwei.ComputeLiuYue(chart, input.TargetLunar.domain())
	if err != nil {
		return ZiweiLiuYueOutput{}, fmt.Errorf("ziwei.liuyue: %w", err)
	}
	return wireFromDomain[ZiweiLiuYueOutput](result)
}

// ZiweiLiuRi returns the flow day palace and its 四化.
func ZiweiLiuRi(_ context.Context, input ZiweiLiuRiInput) (ZiweiLiuRiOutput, error) {
	chart, err := input.Chart.domain()
	if err != nil {
		return ZiweiLiuRiOutput{}, fmt.Errorf("ziwei.liuri: %w", err)
	}
	result, err := ziwei.ComputeLiuRi(chart, input.TargetLunar.domain())
	if err != nil {
		return ZiweiLiuRiOutput{}, fmt.Errorf("ziwei.liuri: %w", err)
	}
	return wireFromDomain[ZiweiLiuRiOutput](result)
}

// ZiweiLiuShi returns the flow hour palace and its 四化.
func ZiweiLiuShi(_ context.Context, input ZiweiLiuShiInput) (ZiweiLiuShiOutput, error) {
	chart, err := input.Chart.domain()
	if err != nil {
		return ZiweiLiuShiOutput{}, fmt.Errorf("ziwei.liushi: %w", err)
	}
	shiZhi, err := ganzhi.ParseZhi(string(input.ShiZhi))
	if err != nil {
		return ZiweiLiuShiOutput{}, fmt.Errorf("ziwei.liushi: shi_zhi: %w", err)
	}
	result, err := ziwei.ComputeLiuShi(chart, input.TargetLunar.domain(), shiZhi)
	if err != nil {
		return ZiweiLiuShiOutput{}, fmt.Errorf("ziwei.liushi: %w", err)
	}
	return wireFromDomain[ZiweiLiuShiOutput](result)
}

// ZiweiBond returns raw cross-chart facts for two charts.
func ZiweiBond(_ context.Context, input ZiweiBondInput) (ZiweiBondOutput, error) {
	chartA, err := input.A.Chart.domain()
	if err != nil {
		return ZiweiBondOutput{}, fmt.Errorf("ziwei.bond: chart a: %w", err)
	}
	chartB, err := input.B.Chart.domain()
	if err != nil {
		return ZiweiBondOutput{}, fmt.Errorf("ziwei.bond: chart b: %w", err)
	}
	return wireFromDomain[ZiweiBondOutput](ziwei.ComputeBond(chartA, chartB))
}
