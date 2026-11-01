package mcpcontract

import (
	"context"
	"encoding/json"
	"fmt"
	"time"

	"liki-engine/internal/engine/bazi"
	"liki-engine/internal/engine/ganzhi"
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

type BaziChart struct {
	ChartDigest ChartDigest `json:"chart_digest" jsonschema:"canonical SHA-256 chart digest。"`
	Nian        BaziPillar  `json:"nian" jsonschema:"年柱。"`
	Yue         BaziPillar  `json:"yue" jsonschema:"月柱。"`
	Ri          BaziPillar  `json:"ri" jsonschema:"日柱。"`
	Shi         BaziPillar  `json:"shi" jsonschema:"时柱。"`
	DaYun       *BaziDaYun  `json:"da_yun" jsonschema:"大运。"`
	Gender      GenderName  `json:"gender" jsonschema:"性别。"`
	BirthYear   int         `json:"birth_year" jsonschema:"出生公历年。"`
	ZiShiRule   string      `json:"zi_shi_rule" jsonschema:"子时换日规则。"`
}

func BaziChartCompute(_ context.Context, input BaziChartInput) (BaziChart, error) {
	chart, err := input.Subject.baziChart()
	if err != nil {
		return BaziChart{}, fmt.Errorf("bazi.chart: %w", err)
	}
	wire := baziChartWire(chart)
	digest, err := chartDigest(wire)
	if err != nil {
		return BaziChart{}, fmt.Errorf("bazi.chart: %w", err)
	}
	wire.ChartDigest = ChartDigest(digest)
	return wire, nil
}

func baziPillarWire(gan ganzhi.Gan, zhi ganzhi.Zhi, naYin string) BaziPillar {
	return BaziPillar{Gan: LiuyaoGanName(gan.String()), Zhi: LiuyaoZhiName(zhi.String()), NaYin: naYin}
}

func baziChartWire(chart bazi.Chart) BaziChart {
	out := BaziChart{
		Nian:   baziPillarWire(chart.Nian.Gan, chart.Nian.Zhi, chart.Nian.NaYin),
		Yue:    baziPillarWire(chart.Yue.Gan, chart.Yue.Zhi, chart.Yue.NaYin),
		Ri:     baziPillarWire(chart.Ri.Gan, chart.Ri.Zhi, chart.Ri.NaYin),
		Shi:    baziPillarWire(chart.Shi.Gan, chart.Shi.Zhi, chart.Shi.NaYin),
		Gender: GenderName(chart.Gender), BirthYear: chart.BirthYear, ZiShiRule: chart.ZiShiRule,
	}
	if chart.DaYun != nil {
		out.DaYun = baziDaYunWire(*chart.DaYun)
	}
	return out
}

func baziDaYunWire(dayun bazi.DaYun) *BaziDaYun {
	out := &BaziDaYun{
		StartDate: DateString(dayun.StartDate), StartYearAfter: dayun.StartYearAfter,
		StartMonthAfter: dayun.StartMonthAfter, StartDayAfter: dayun.StartDayAfter,
		Direction: dayun.Direction,
	}
	for _, step := range dayun.Steps {
		wire := BaziDaYunStep{
			Gan: LiuyaoGanName(step.Gan.String()), Zhi: LiuyaoZhiName(step.Zhi.String()),
			StartDate: DateString(step.StartDate), EndDate: DateString(step.EndDate),
			StartYear: step.StartYear, EndYear: step.EndYear, Name: step.Name,
			Element: ElementName(step.Element), ShiShen: step.ShiShen, Rooted: step.Rooted,
		}
		for _, ref := range step.RootRefs {
			wire.RootRefs = append(wire.RootRefs, BaziDaYunRootRef(ref))
		}
		out.Steps = append(out.Steps, wire)
	}
	return out
}

// ── typed handlers for the remaining bazi tools ──

// BaziFullChart expands a bazi.chart wire chart into the full chart.
func BaziFullChart(_ context.Context, input BaziFullChartInput) (json.RawMessage, error) {
	chart, err := input.Subject.baziChart()
	if err != nil {
		return nil, fmt.Errorf("bazi.fullchart: %w", err)
	}
	raw, err := marshalRawBaziFullChart(bazi.ComputeFullChart(chart))
	if err != nil {
		return nil, fmt.Errorf("bazi.fullchart: %w", err)
	}
	return raw, nil
}

type baziFullChartJSON struct {
	*bazi.FullChart
	TiaoHou baziTiaoHouJSON `json:"tiao_hou"`
}

type baziTiaoHouJSON struct {
	PrimaryWuxing      string                      `json:"primary_wuxing"`
	SecondaryWuxing    string                      `json:"secondary_wuxing"`
	Season             string                      `json:"season"`
	Detail             string                      `json:"detail"`
	Model              string                      `json:"model"`
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
		return []byte(`{"hidden":null,"occurrences":null,"relation_facts":null,"stem":null,"transparent":null}`), nil
	}
	return json.Marshal(availability.value)
}

func marshalRawBaziFullChart(full bazi.FullChart) (json.RawMessage, error) {
	value := baziFullChartJSON{
		FullChart: &full,
		TiaoHou: baziTiaoHouJSON{
			PrimaryWuxing:      full.TiaoHou.PrimaryWuxing,
			SecondaryWuxing:    full.TiaoHou.SecondaryWuxing,
			Season:             full.TiaoHou.Season,
			Detail:             full.TiaoHou.Detail,
			Model:              full.TiaoHou.Model,
			Primary:            full.TiaoHou.Primary,
			Secondary:          requiredTiaoHouAvailability{value: full.TiaoHou.Secondary},
			SecondaryCondition: full.TiaoHou.SecondaryCondition,
		},
	}
	raw, err := json.Marshal(value)
	if err != nil {
		return nil, fmt.Errorf("bazi.fullchart: marshal output: %w", err)
	}
	return raw, nil
}

// BaziBond returns raw cross-chart facts for two bazi charts.
func BaziBond(_ context.Context, input BaziBondInput) (json.RawMessage, error) {
	chartA, err := input.A.Subject.baziChart()
	if err != nil {
		return nil, fmt.Errorf("bazi.bond: chart a: %w", err)
	}
	chartB, err := input.B.Subject.baziChart()
	if err != nil {
		return nil, fmt.Errorf("bazi.bond: chart b: %w", err)
	}
	raw, err := marshalRawOutput("bazi.bond", bazi.ComputeBond(chartA, chartB))
	if err != nil {
		return nil, fmt.Errorf("bazi.bond: %w", err)
	}
	return raw, nil
}

// BaziLiuNian computes the year pillar and its interactions with a chart.
func BaziLiuNian(_ context.Context, input BaziLiuNianInput) (any, error) {
	return runSeries(input, seriesCountValue(input.Count), MaxSeriesCount,
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
			in.StartTime = standardTimeValue(next.StartInstant.In(loc))
			return in, nil
		},
	)
}

func parseBaziLiuNianTime(input BaziLiuNianInput) (bazi.LiuNianTime, tianwen.CivilDate, time.Time, *time.Location, error) {
	standard, loc, err := parseStandardTime(input.StartTime)
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
	chart, err := input.Subject.baziChart()
	if err != nil {
		return BaziLiuNianOutput{}, fmt.Errorf("bazi.liunian: %w", err)
	}
	result, err := bazi.ComputeLiuNianAtTime(chart, resolved)
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
	output.Resolved = BaziLiuNianResolved{
		SelectorStandardTime: standardTimeValue(at),
		SelectorDate:         civilDateString(date),
		LichunYear:           resolved.LichunYear,
		LichunStart:          civilDateString(resolved.Start),
		LichunEnd:            civilDateString(resolved.End),
		BoundaryDay:          resolved.BoundaryDay,
	}
	return output, nil
}

// BaziLiuYue computes the month pillar and its interactions with a chart.
func BaziLiuYue(_ context.Context, input BaziLiuYueInput) (any, error) {
	return runSeries(input, seriesCountValue(input.Count), MaxSeriesCount,
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
			in.StartTime = standardTimeValue(next.StartInstant.In(loc))
			return in, nil
		},
	)
}

func parseBaziLiuYueTime(input BaziLiuYueInput) (bazi.LiuYueTime, tianwen.CivilDate, time.Time, *time.Location, error) {
	standard, loc, err := parseStandardTime(input.StartTime)
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
	chart, err := input.Subject.baziChart()
	if err != nil {
		return BaziLiuYueOutput{}, fmt.Errorf("bazi.liuyue: %w", err)
	}
	result, err := bazi.ComputeLiuYueAtTime(chart, resolved)
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
	output.Resolved = BaziLiuYueResolved{
		SelectorStandardTime: standardTimeValue(at),
		SelectorDate:         civilDateString(date),
		JieqiStart:           civilDateString(resolved.Start),
		JieqiEnd:             civilDateString(resolved.End),
		BoundaryDay:          resolved.BoundaryDay,
	}
	return output, nil
}

// BaziLiuRi computes the day pillar and its interactions with a chart.
func BaziLiuRi(_ context.Context, input BaziLiuRiInput) (any, error) {
	return runSeries(input, seriesCountValue(input.Count), MaxSeriesCount,
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
			in.StartTime = standardTimeValue(next.Standard.In(loc))
			return in, nil
		},
	)
}

func parseBaziLiuRiTime(input BaziLiuRiInput) (bazi.LiuRiTime, *time.Location, error) {
	standard, loc, err := parseStandardTime(input.StartTime)
	if err != nil {
		return bazi.LiuRiTime{}, nil, err
	}
	resolved, err := bazi.ResolveLiuRiTimeAt(standard, loc, input.Longitude)
	return resolved, loc, err
}

func computeBaziLiuRi(input BaziLiuRiInput) (BaziLiuRiOutput, error) {
	chart, err := input.Subject.baziChart()
	if err != nil {
		return BaziLiuRiOutput{}, fmt.Errorf("bazi.liuri: %w", err)
	}
	resolved, loc, err := parseBaziLiuRiTime(input)
	if err != nil {
		return BaziLiuRiOutput{}, fmt.Errorf("bazi.liuri: %w", err)
	}
	result, err := bazi.ComputeLiuRi(chart, resolved)
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
	output.Resolved = BaziLiuRiResolved{
		StandardTime: standardTimeValue(resolved.Standard.In(loc)),
		LocalDate:    civilDateString(resolved.LocalDate),
		LastDate:     civilDateString(resolved.Date),
		BeijingDate:  civilDateString(resolved.BeijingDate),
	}
	return output, nil
}

// BaziLiuShi computes the hour pillar and its interactions with a chart.
func BaziLiuShi(_ context.Context, input BaziLiuShiInput) (any, error) {
	return runSeries(input, seriesCountValue(input.Count), MaxSeriesCount,
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
			in.StartTime = standardTimeValue(next.Standard.In(loc))
			return in, nil
		},
	)
}

func parseBaziLiuShiTime(input BaziLiuShiInput) (bazi.LiuShiTime, *time.Location, error) {
	standard, loc, err := parseStandardTime(input.StartTime)
	if err != nil {
		return bazi.LiuShiTime{}, nil, err
	}
	resolved, err := bazi.ResolveLiuShiTime(standard, loc, input.Longitude)
	return resolved, loc, err
}

func computeBaziLiuShi(input BaziLiuShiInput) (BaziLiuShiOutput, error) {
	chart, err := input.Subject.baziChart()
	if err != nil {
		return BaziLiuShiOutput{}, fmt.Errorf("bazi.liushi: %w", err)
	}
	resolved, loc, err := parseBaziLiuShiTime(input)
	if err != nil {
		return BaziLiuShiOutput{}, fmt.Errorf("bazi.liushi: %w", err)
	}
	result, err := bazi.ComputeLiuShi(chart, resolved)
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
	output.Resolved = BaziLiuShiResolved{
		StandardTime: standardTimeValue(resolved.Standard.In(loc)),
		LocalDate:    civilDateString(resolved.LocalDate),
		Shichen:      resolved.Shichen.String(),
	}
	return output, nil
}

// BaziXiaoYun returns the 小运 year list for a chart.
func BaziXiaoYun(_ context.Context, input BaziXiaoYunInput) (BaziXiaoYunOutput, error) {
	n := seriesCountValue(input.Count)
	chart, err := input.Subject.baziChart()
	if err != nil {
		return BaziXiaoYunOutput{}, fmt.Errorf("bazi.xiaoyun: %w", err)
	}
	items, err := wireFromDomain[[]BaziXiaoYunItem](bazi.ComputeXiaoYun(chart, n))
	if err != nil {
		return BaziXiaoYunOutput{}, fmt.Errorf("bazi.xiaoyun: %w", err)
	}
	return BaziXiaoYunOutput{Items: items}, nil
}
