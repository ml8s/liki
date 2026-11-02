package mcpcontract

import (
	"context"
	"encoding/json"
	"fmt"
	"time"

	"liki-engine/internal/engine/bazi"
	"liki-engine/internal/engine/tianwen"
)

type BaziPillar struct {
	Gan   LiuyaoGanName `json:"gan" jsonschema:"天干。"`
	Zhi   LiuyaoZhiName `json:"zhi" jsonschema:"地支。"`
	NaYin string        `json:"na_yin" jsonschema:"干支纳音。"`
}

type BaziDaYunRootRef struct {
	Source string `json:"source" jsonschema:"通根来源。"`
	Pillar string `json:"pillar,omitempty" jsonschema:"相关柱。"`
	Stem   string `json:"stem" jsonschema:"天干。"`
	Branch string `json:"branch" jsonschema:"地支。"`
}

type BaziDaYunStep struct {
	Gan       LiuyaoGanName      `json:"gan" jsonschema:"大运天干。"`
	Zhi       LiuyaoZhiName      `json:"zhi" jsonschema:"大运地支。"`
	StartDate DateString         `json:"start_date" jsonschema:"起始公历日 YYYY-MM-DD。"`
	EndDate   DateString         `json:"end_date" jsonschema:"结束公历日 YYYY-MM-DD。"`
	StartYear int                `json:"start_year" jsonschema:"起始公历年。"`
	EndYear   int                `json:"end_year" jsonschema:"结束公历年。"`
	Name      string             `json:"name" jsonschema:"干支名称。"`
	Element   ElementName        `json:"wuxing" jsonschema:"大运纳音五行。"`
	ShiShen   string             `json:"shi_shen" jsonschema:"相对日主的十神运。"`
	Rooted    *bool              `json:"rooted,omitempty" jsonschema:"大运干是否通根。"`
	RootRefs  []BaziDaYunRootRef `json:"root_refs,omitempty" jsonschema:"通根依据。"`
}

type BaziDaYun struct {
	StartDate       DateString      `json:"start_date" jsonschema:"起运公历日 YYYY-MM-DD。"`
	StartYearAfter  int             `json:"start_year_after" jsonschema:"出生后整年数。"`
	StartMonthAfter int             `json:"start_month_after" jsonschema:"出生后余月数。"`
	StartDayAfter   int             `json:"start_day_after" jsonschema:"出生后余日数。"`
	Direction       string          `json:"direction" jsonschema:"顺排或逆排。"`
	Steps           []BaziDaYunStep `json:"steps" jsonschema:"大运步骤。"`
}

// BaziChartInput consumes a resolved natal subject. The chart domain derives
// year/month from solar-term events and day/hour from LAST + 子时.
type BaziChartInput struct {
	Subject NatalSubject `json:"subject" jsonschema:"已归一化的出生主体。"`
}

// BaziChartCompute returns the canonical complete natal chart.
func BaziChartCompute(_ context.Context, input BaziChartInput) (json.RawMessage, error) {
	chart, err := input.Subject.baziChart()
	if err != nil {
		return nil, fmt.Errorf("bazi.chart: %w", err)
	}
	return marshalRawBaziChart(chart)
}

// ── typed handlers for the remaining bazi tools ──

type baziChartJSON struct {
	*bazi.Chart
	TiaoHou   baziTiaoHouJSON `json:"tiao_hou"`
	ZiShiRule string          `json:"zi_shi_rule"`
}

type baziTiaoHouJSON struct {
	PrimaryWuxing      string                      `json:"primary_wuxing"`
	SecondaryWuxing    string                      `json:"secondary_wuxing"`
	Season             string                      `json:"season"`
	TiaoHouExplanation string                      `json:"tiao_hou_explanation"`
	TiaoHouModelID     string                      `json:"tiao_hou_model_id"`
	Primary            bazi.TiaoHouAvailability    `json:"primary"`
	Secondary          requiredTiaoHouAvailability `json:"secondary"`
	SecondaryCondition string                      `json:"secondary_condition,omitempty"`
}

type requiredTiaoHouAvailability struct {
	value *bazi.TiaoHouAvailability
}

func (availability requiredTiaoHouAvailability) MarshalJSON() ([]byte, error) {
	if availability.value == nil {
		// The wire DTO models secondary as a required object with nullable
		// facts. Preserve that explicit closure instead of omitting the field.
		return []byte(`{"guan_xi_refs":null,"hidden":null,"occurrences":null,"stem":null,"transparent":null}`), nil
	}
	return json.Marshal(availability.value)
}

func marshalRawBaziChart(chart bazi.Chart) (json.RawMessage, error) {
	value := baziChartJSON{
		Chart: &chart,
		TiaoHou: baziTiaoHouJSON{
			PrimaryWuxing:      chart.TiaoHou.PrimaryWuxing,
			SecondaryWuxing:    chart.TiaoHou.SecondaryWuxing,
			Season:             chart.TiaoHou.Season,
			TiaoHouExplanation: chart.TiaoHou.TiaoHouExplanation,
			TiaoHouModelID:     chart.TiaoHou.ModelID,
			Primary:            chart.TiaoHou.Primary,
			Secondary:          requiredTiaoHouAvailability{value: chart.TiaoHou.Secondary},
			SecondaryCondition: chart.TiaoHou.SecondaryCondition,
		},
		ZiShiRule: chart.ZiShiRule,
	}
	raw, err := json.Marshal(value)
	if err != nil {
		return nil, fmt.Errorf("bazi.chart: marshal output: %w", err)
	}
	return raw, nil
}

// BaziBond returns raw cross-chart facts for two bazi charts.
func BaziBond(_ context.Context, input BaziBondInput) (json.RawMessage, error) {
	coreA, err := input.A.Subject.baziNatalCore()
	if err != nil {
		return nil, fmt.Errorf("bazi.bond: chart a: %w", err)
	}
	coreB, err := input.B.Subject.baziNatalCore()
	if err != nil {
		return nil, fmt.Errorf("bazi.bond: chart b: %w", err)
	}
	raw, err := marshalRawOutput("bazi.bond", bazi.ComputeBond(coreA, coreB))
	if err != nil {
		return nil, fmt.Errorf("bazi.bond: %w", err)
	}
	return raw, nil
}

// BaziLiuNian computes the year pillar and its interactions with a chart.
func BaziLiuNian(_ context.Context, input BaziLiuNianInput) (BaziLiuNianSeries, error) {
	items, err := runSeries(input, seriesCountValue(input.LiuNianCount), MaxSeriesCount,

		func(in BaziLiuNianInput) (BaziLiuNianOutput, error) {
			return computeBaziLiuNian(in)
		},
		func(in BaziLiuNianInput) (BaziLiuNianInput, error) {
			resolved, _, _, loc, err := parseBaziLiuNianTime(in)
			if err != nil {
				return in, fmt.Errorf("bazi.liunian: %w", err)
			}
			next, err := bazi.NextLiuNianTimeAt(resolved, loc)
			if err != nil {
				return in, fmt.Errorf("bazi.liunian: %w", err)
			}
			in.LiuNianStartTime = standardTimeValue(next.StartInstant.In(loc))
			return in, nil
		},
	)
	if err != nil {
		return BaziLiuNianSeries{}, err
	}
	return BaziLiuNianSeries{LiuNian: items}, nil
}

func parseBaziLiuNianTime(input BaziLiuNianInput) (bazi.LiuNianTime, tianwen.CivilDate, time.Time, *time.Location, error) {
	standard, loc, err := parseStandardTime(input.LiuNianStartTime)
	if err != nil {
		return bazi.LiuNianTime{}, tianwen.CivilDate{}, time.Time{}, nil, err
	}
	date := tianwen.StandardDateAt(standard, loc)
	resolved, err := bazi.ResolveLiuNianTimeAt(standard, loc)
	return resolved, date, standard, loc, err
}

func computeBaziLiuNian(input BaziLiuNianInput) (BaziLiuNianOutput, error) {
	resolved, date, at, _, err := parseBaziLiuNianTime(input)
	if err != nil {
		return BaziLiuNianOutput{}, fmt.Errorf("bazi.liunian: %w", err)
	}
	core, err := input.Subject.baziNatalCore()
	if err != nil {
		return BaziLiuNianOutput{}, fmt.Errorf("bazi.liunian: %w", err)
	}
	result, err := bazi.ComputeLiuNianAtTime(core, resolved)
	if err != nil {
		return BaziLiuNianOutput{}, fmt.Errorf("bazi.liunian: %w", err)
	}
	if result == nil {
		return BaziLiuNianOutput{}, fmt.Errorf("bazi.liunian: empty result for %s", date)
	}
	output, err := wireFromDomain[BaziLiuNianOutput](*result)
	if err != nil {
		return BaziLiuNianOutput{}, fmt.Errorf("bazi.liunian: %w", err)
	}
	output.NianJie = BaziLiuNianResolved{
		LiuNianStartTime: standardTimeValue(at),
		LiuNianStartDate: civilDateString(date),
		LichunYear:       resolved.LichunYear,
		LichunStart:      civilDateString(resolved.Start),
		LichunEnd:        civilDateString(resolved.End),
		BoundaryDay:      resolved.BoundaryDay,
	}
	return output, nil
}

// BaziLiuYue computes the month pillar and its interactions with a chart.
func BaziLiuYue(_ context.Context, input BaziLiuYueInput) (BaziLiuYueSeries, error) {
	items, err := runSeries(input, seriesCountValue(input.LiuYueCount), MaxSeriesCount,

		func(in BaziLiuYueInput) (BaziLiuYueOutput, error) {
			return computeBaziLiuYue(in)
		},
		func(in BaziLiuYueInput) (BaziLiuYueInput, error) {
			resolved, _, _, loc, err := parseBaziLiuYueTime(in)
			if err != nil {
				return in, fmt.Errorf("bazi.liuyue: %w", err)
			}
			next, err := bazi.NextLiuYueTimeAt(resolved, loc)
			if err != nil {
				return in, fmt.Errorf("bazi.liuyue: %w", err)
			}
			in.LiuYueStartTime = standardTimeValue(next.StartInstant.In(loc))
			return in, nil
		},
	)
	if err != nil {
		return BaziLiuYueSeries{}, err
	}
	return BaziLiuYueSeries{LiuYue: items}, nil
}

func parseBaziLiuYueTime(input BaziLiuYueInput) (bazi.LiuYueTime, tianwen.CivilDate, time.Time, *time.Location, error) {
	standard, loc, err := parseStandardTime(input.LiuYueStartTime)
	if err != nil {
		return bazi.LiuYueTime{}, tianwen.CivilDate{}, time.Time{}, nil, err
	}
	date := tianwen.StandardDateAt(standard, loc)
	resolved, err := bazi.ResolveLiuYueTimeAt(standard, loc)
	return resolved, date, standard, loc, err
}

func computeBaziLiuYue(input BaziLiuYueInput) (BaziLiuYueOutput, error) {
	resolved, date, at, _, err := parseBaziLiuYueTime(input)
	if err != nil {
		return BaziLiuYueOutput{}, fmt.Errorf("bazi.liuyue: %w", err)
	}
	core, err := input.Subject.baziNatalCore()
	if err != nil {
		return BaziLiuYueOutput{}, fmt.Errorf("bazi.liuyue: %w", err)
	}
	result, err := bazi.ComputeLiuYueAtTime(core, resolved)
	if err != nil {
		return BaziLiuYueOutput{}, fmt.Errorf("bazi.liuyue: %w", err)
	}
	if result == nil {
		return BaziLiuYueOutput{}, fmt.Errorf("bazi.liuyue: empty result for %s", date)
	}
	output, err := wireFromDomain[BaziLiuYueOutput](*result)
	if err != nil {
		return BaziLiuYueOutput{}, fmt.Errorf("bazi.liuyue: %w", err)
	}
	output.JieQi = BaziLiuYueJieQi{
		LiuYueStartTime: standardTimeValue(at),
		LiuYueStartDate: civilDateString(date),
		JieqiStart:      civilDateString(resolved.Start),
		JieqiEnd:        civilDateString(resolved.End),
		BoundaryDay:     resolved.BoundaryDay,
	}
	return output, nil
}

// BaziLiuRi computes the day pillar and its interactions with a chart.
func BaziLiuRi(_ context.Context, input BaziLiuRiInput) (BaziLiuRiSeries, error) {
	items, err := runSeries(input, seriesCountValue(input.LiuRiCount), MaxSeriesCount,

		func(in BaziLiuRiInput) (BaziLiuRiOutput, error) {
			return computeBaziLiuRi(in)
		},
		func(in BaziLiuRiInput) (BaziLiuRiInput, error) {
			resolved, loc, err := parseBaziLiuRiTime(in)
			if err != nil {
				return in, fmt.Errorf("bazi.liuri: %w", err)
			}
			next, err := bazi.NextLiuRiTime(resolved)
			if err != nil {
				return in, fmt.Errorf("bazi.liuri: %w", err)
			}
			in.LiuRiStartTime = standardTimeValue(next.Standard.In(loc))
			return in, nil
		},
	)
	if err != nil {
		return BaziLiuRiSeries{}, err
	}
	return BaziLiuRiSeries{LiuRi: items}, nil
}

func parseBaziLiuRiTime(input BaziLiuRiInput) (bazi.LiuRiTime, *time.Location, error) {
	standard, loc, err := parseStandardTime(input.LiuRiStartTime)
	if err != nil {
		return bazi.LiuRiTime{}, nil, err
	}
	resolved, err := bazi.ResolveLiuRiTimeAt(standard, loc, input.LiuRiLongitude)
	return resolved, loc, err
}

func computeBaziLiuRi(input BaziLiuRiInput) (BaziLiuRiOutput, error) {
	core, err := input.Subject.baziNatalCore()
	if err != nil {
		return BaziLiuRiOutput{}, fmt.Errorf("bazi.liuri: %w", err)
	}
	resolved, loc, err := parseBaziLiuRiTime(input)
	if err != nil {
		return BaziLiuRiOutput{}, fmt.Errorf("bazi.liuri: %w", err)
	}
	result, err := bazi.ComputeLiuRi(core, resolved)
	if err != nil {
		return BaziLiuRiOutput{}, fmt.Errorf("bazi.liuri: %w", err)
	}
	if result == nil {
		return BaziLiuRiOutput{}, fmt.Errorf("bazi.liuri: empty result for %s", resolved.Date)
	}
	output, err := wireFromDomain[BaziLiuRiOutput](*result)
	if err != nil {
		return BaziLiuRiOutput{}, fmt.Errorf("bazi.liuri: %w", err)
	}
	output.RiJie = BaziLiuRiJie{
		LiuRiStartTime: standardTimeValue(resolved.Standard.In(loc)),
		LiuRiLocalDate: civilDateString(resolved.LocalDate),
		LiuRiSolarDate: civilDateString(resolved.Date),
		BeijingDate:    civilDateString(resolved.BeijingDate),
	}
	return output, nil
}

// BaziLiuShi computes the hour pillar and its interactions with a chart.
func BaziLiuShi(_ context.Context, input BaziLiuShiInput) (BaziLiuShiSeries, error) {
	items, err := runSeries(input, seriesCountValue(input.LiuShiCount), MaxSeriesCount,

		func(in BaziLiuShiInput) (BaziLiuShiOutput, error) {
			return computeBaziLiuShi(in)
		},
		func(in BaziLiuShiInput) (BaziLiuShiInput, error) {
			t, loc, err := parseBaziLiuShiTime(in)
			if err != nil {
				return in, fmt.Errorf("bazi.liushi: %w", err)
			}
			next, err := bazi.NextLiuShiTime(t)
			if err != nil {
				return in, fmt.Errorf("bazi.liushi: %w", err)
			}
			in.LiuShiStartTime = standardTimeValue(next.Standard.In(loc))
			return in, nil
		},
	)
	if err != nil {
		return BaziLiuShiSeries{}, err
	}
	return BaziLiuShiSeries{LiuShi: items}, nil
}

func parseBaziLiuShiTime(input BaziLiuShiInput) (bazi.LiuShiTime, *time.Location, error) {
	standard, loc, err := parseStandardTime(input.LiuShiStartTime)
	if err != nil {
		return bazi.LiuShiTime{}, nil, err
	}
	resolved, err := bazi.ResolveLiuShiTime(standard, loc, input.LiuShiLongitude)
	return resolved, loc, err
}

func computeBaziLiuShi(input BaziLiuShiInput) (BaziLiuShiOutput, error) {
	core, err := input.Subject.baziNatalCore()
	if err != nil {
		return BaziLiuShiOutput{}, fmt.Errorf("bazi.liushi: %w", err)
	}
	resolved, loc, err := parseBaziLiuShiTime(input)
	if err != nil {
		return BaziLiuShiOutput{}, fmt.Errorf("bazi.liushi: %w", err)
	}
	result, err := bazi.ComputeLiuShi(core, resolved)
	if err != nil {
		return BaziLiuShiOutput{}, fmt.Errorf("bazi.liushi: %w", err)
	}
	if result == nil {
		return BaziLiuShiOutput{}, fmt.Errorf("bazi.liushi: empty result for %s", resolved.Standard)
	}
	output, err := wireFromDomain[BaziLiuShiOutput](*result)
	if err != nil {
		return BaziLiuShiOutput{}, fmt.Errorf("bazi.liushi: %w", err)
	}
	output.ShiChen = BaziLiuShiChen{
		LiuShiStartTime: standardTimeValue(resolved.Standard.In(loc)),
		LiuShiLocalDate: civilDateString(resolved.LocalDate),
		Shichen:         resolved.Shichen.String(),
	}
	return output, nil
}

// BaziXiaoYun returns the 小运 year list for a chart.
func BaziXiaoYun(_ context.Context, input BaziXiaoYunInput) (BaziXiaoYunOutput, error) {
	n := seriesCountValue(input.XiaoYunCount)
	core, err := input.Subject.baziNatalCore()
	if err != nil {
		return BaziXiaoYunOutput{}, fmt.Errorf("bazi.xiaoyun: %w", err)
	}
	items, err := wireFromDomain[[]BaziXiaoYunItem](bazi.ComputeXiaoYun(core, n))
	if err != nil {
		return BaziXiaoYunOutput{}, fmt.Errorf("bazi.xiaoyun: %w", err)
	}
	return BaziXiaoYunOutput{XiaoYun: items}, nil
}
