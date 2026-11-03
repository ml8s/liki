package mcpcontract

import (
	"context"
	"encoding/json"
	"fmt"
	"time"

	"liki-engine/internal/engine/bazi"
	"liki-engine/internal/engine/tianwen"
)

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
		return []byte(`{"guan_xi_links":null,"hidden":null,"occurrences":null,"stem":null,"transparent":null}`), nil
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
	coreA, err := input.A.Subject.baziCore()
	if err != nil {
		return nil, fmt.Errorf("bazi.bond: chart a: %w", err)
	}
	coreB, err := input.B.Subject.baziCore()
	if err != nil {
		return nil, fmt.Errorf("bazi.bond: chart b: %w", err)
	}
	raw, err := marshalRawOutput("bazi.bond", newBaziBondOutput(bazi.ComputeBond(coreA, coreB)))
	if err != nil {
		return nil, fmt.Errorf("bazi.bond: %w", err)
	}
	return raw, nil
}

// BaziLiupan computes one selected dynamic time layer from the calibrated natal
// subject. The five layers share one core preparation; their calendar advance
// rules remain domain-specific.
func BaziLiupan(_ context.Context, input BaziLiupanInput) (json.RawMessage, error) {
	if err := validateBaziLiupan(input); err != nil {
		return nil, fmt.Errorf("bazi.liupan: %w", err)
	}
	count := seriesCountValue(input.Count)
	core, err := input.Subject.baziCore()
	if err != nil {
		return nil, fmt.Errorf("bazi.liupan: %w", err)
	}
	current := input

	switch input.Scope {
	case "liu_nian":
		items := make([]BaziLiuNianOutput, 0, count)
		for range count {
			item, err := computeBaziLiuNian(current, core)
			if err != nil {
				return nil, err
			}
			items = append(items, item)
			current, err = advanceBaziLiuNian(current)
			if err != nil {
				return nil, err
			}
		}
		return marshalRawOutput("bazi.liupan", BaziLiuNianSeries{LiuNian: items})
	case "liu_yue":
		items := make([]BaziLiuYueOutput, 0, count)
		for range count {
			item, err := computeBaziLiuYue(current, core)
			if err != nil {
				return nil, err
			}
			items = append(items, item)
			current, err = advanceBaziLiuYue(current)
			if err != nil {
				return nil, err
			}
		}
		return marshalRawOutput("bazi.liupan", BaziLiuYueSeries{LiuYue: items})
	case "liu_ri":
		items := make([]BaziLiuRiOutput, 0, count)
		for range count {
			item, err := computeBaziLiuRi(current, core)
			if err != nil {
				return nil, err
			}
			items = append(items, item)
			current, err = advanceBaziLiuRi(current)
			if err != nil {
				return nil, err
			}
		}
		return marshalRawOutput("bazi.liupan", BaziLiuRiSeries{LiuRi: items})
	case "liu_shi":
		items := make([]BaziLiuShiOutput, 0, count)
		for range count {
			item, err := computeBaziLiuShi(current, core)
			if err != nil {
				return nil, err
			}
			items = append(items, item)
			current, err = advanceBaziLiuShi(current)
			if err != nil {
				return nil, err
			}
		}
		return marshalRawOutput("bazi.liupan", BaziLiuShiSeries{LiuShi: items})
	case "xiao_yun":
		items, err := wireFromDomain[[]BaziXiaoYunItem](bazi.ComputeXiaoYun(core, count))
		if err != nil {
			return nil, fmt.Errorf("bazi.liupan: %w", err)
		}
		return marshalRawOutput("bazi.liupan", BaziXiaoYunOutput{XiaoYun: items})
	default:
		return nil, fmt.Errorf("bazi.liupan: invalid scope %q", input.Scope)
	}
}

func validateBaziLiupan(input BaziLiupanInput) error {
	if input.Longitude != nil && (*input.Longitude < -180 || *input.Longitude > 180) {
		return fmt.Errorf("longitude must be within [-180,180], got %v", *input.Longitude)
	}
	switch input.Scope {
	case "liu_nian", "liu_yue", "liu_ri", "liu_shi":
		if input.StartTime == "" {
			return fmt.Errorf("start_time is required")
		}
		if input.Scope == "liu_ri" || input.Scope == "liu_shi" {
			if input.Longitude == nil {
				return fmt.Errorf("longitude is required")
			}
		} else if input.Longitude != nil {
			return fmt.Errorf("longitude is not allowed")
		}
	case "xiao_yun":
		if input.StartTime != "" {
			return fmt.Errorf("start_time is not allowed")
		}
		if input.Longitude != nil {
			return fmt.Errorf("longitude is not allowed")
		}
	default:
		return fmt.Errorf("invalid scope %q", input.Scope)
	}
	return nil
}

func parseBaziLiuNianTime(input BaziLiupanInput) (bazi.LiuNianTime, tianwen.CivilDate, time.Time, *time.Location, error) {
	standard, loc, err := parseStandardTime(input.StartTime)
	if err != nil {
		return bazi.LiuNianTime{}, tianwen.CivilDate{}, time.Time{}, nil, err
	}
	date := tianwen.StandardDateAt(standard, loc)
	resolved, err := bazi.ResolveLiuNianTimeAt(standard, loc)
	return resolved, date, standard, loc, err
}

func computeBaziLiuNian(input BaziLiupanInput, core bazi.Core) (BaziLiuNianOutput, error) {
	resolved, date, at, _, err := parseBaziLiuNianTime(input)
	if err != nil {
		return BaziLiuNianOutput{}, fmt.Errorf("bazi.liupan liu_nian: %w", err)
	}
	result, err := bazi.ComputeLiuNianAtTime(core, resolved)
	if err != nil {
		return BaziLiuNianOutput{}, fmt.Errorf("bazi.liupan liu_nian: %w", err)
	}
	if result == nil {
		return BaziLiuNianOutput{}, fmt.Errorf("bazi.liupan liu_nian: empty result for %s", date)
	}
	output, err := wireFromDomain[BaziLiuNianOutput](*result)
	if err != nil {
		return BaziLiuNianOutput{}, fmt.Errorf("bazi.liupan liu_nian: %w", err)
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

func advanceBaziLiuNian(input BaziLiupanInput) (BaziLiupanInput, error) {
	resolved, _, _, loc, err := parseBaziLiuNianTime(input)
	if err != nil {
		return input, fmt.Errorf("bazi.liupan liu_nian: %w", err)
	}
	next, err := bazi.NextLiuNianTimeAt(resolved, loc)
	if err != nil {
		return input, fmt.Errorf("bazi.liupan liu_nian: %w", err)
	}
	input.StartTime = standardTimeValue(next.StartInstant.In(loc))
	return input, nil
}

func parseBaziLiuYueTime(input BaziLiupanInput) (bazi.LiuYueTime, tianwen.CivilDate, time.Time, *time.Location, error) {
	standard, loc, err := parseStandardTime(input.StartTime)
	if err != nil {
		return bazi.LiuYueTime{}, tianwen.CivilDate{}, time.Time{}, nil, err
	}
	date := tianwen.StandardDateAt(standard, loc)
	resolved, err := bazi.ResolveLiuYueTimeAt(standard, loc)
	return resolved, date, standard, loc, err
}

func computeBaziLiuYue(input BaziLiupanInput, core bazi.Core) (BaziLiuYueOutput, error) {
	resolved, date, at, _, err := parseBaziLiuYueTime(input)
	if err != nil {
		return BaziLiuYueOutput{}, fmt.Errorf("bazi.liupan liu_yue: %w", err)
	}
	result, err := bazi.ComputeLiuYueAtTime(core, resolved)
	if err != nil {
		return BaziLiuYueOutput{}, fmt.Errorf("bazi.liupan liu_yue: %w", err)
	}
	if result == nil {
		return BaziLiuYueOutput{}, fmt.Errorf("bazi.liupan liu_yue: empty result for %s", date)
	}
	output, err := wireFromDomain[BaziLiuYueOutput](*result)
	if err != nil {
		return BaziLiuYueOutput{}, fmt.Errorf("bazi.liupan liu_yue: %w", err)
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

func advanceBaziLiuYue(input BaziLiupanInput) (BaziLiupanInput, error) {
	resolved, _, _, loc, err := parseBaziLiuYueTime(input)
	if err != nil {
		return input, fmt.Errorf("bazi.liupan liu_yue: %w", err)
	}
	next, err := bazi.NextLiuYueTimeAt(resolved, loc)
	if err != nil {
		return input, fmt.Errorf("bazi.liupan liu_yue: %w", err)
	}
	input.StartTime = standardTimeValue(next.StartInstant.In(loc))
	return input, nil
}

func parseBaziLiuRiTime(input BaziLiupanInput) (bazi.LiuRiTime, *time.Location, error) {
	if input.Longitude == nil {
		return bazi.LiuRiTime{}, nil, fmt.Errorf("longitude is required")
	}
	standard, loc, err := parseStandardTime(input.StartTime)
	if err != nil {
		return bazi.LiuRiTime{}, nil, err
	}
	resolved, err := bazi.ResolveLiuRiTimeAt(standard, loc, *input.Longitude)
	return resolved, loc, err
}

func computeBaziLiuRi(input BaziLiupanInput, core bazi.Core) (BaziLiuRiOutput, error) {
	resolved, loc, err := parseBaziLiuRiTime(input)
	if err != nil {
		return BaziLiuRiOutput{}, fmt.Errorf("bazi.liupan liu_ri: %w", err)
	}
	result, err := bazi.ComputeLiuRi(core, resolved)
	if err != nil {
		return BaziLiuRiOutput{}, fmt.Errorf("bazi.liupan liu_ri: %w", err)
	}
	if result == nil {
		return BaziLiuRiOutput{}, fmt.Errorf("bazi.liupan liu_ri: empty result for %s", resolved.Date)
	}
	output, err := wireFromDomain[BaziLiuRiOutput](*result)
	if err != nil {
		return BaziLiuRiOutput{}, fmt.Errorf("bazi.liupan liu_ri: %w", err)
	}
	output.RiJie = BaziLiuRiJie{
		LiuRiStartTime: standardTimeValue(resolved.Standard.In(loc)),
		LiuRiLocalDate: civilDateString(resolved.Date),
		LiuRiSolarDate: civilDateString(resolved.Date),
	}
	return output, nil
}

func advanceBaziLiuRi(input BaziLiupanInput) (BaziLiupanInput, error) {
	resolved, loc, err := parseBaziLiuRiTime(input)
	if err != nil {
		return input, fmt.Errorf("bazi.liupan liu_ri: %w", err)
	}
	next, err := bazi.NextLiuRiTime(resolved)
	if err != nil {
		return input, fmt.Errorf("bazi.liupan liu_ri: %w", err)
	}
	input.StartTime = standardTimeValue(next.Standard.In(loc))
	return input, nil
}

func parseBaziLiuShiTime(input BaziLiupanInput) (bazi.LiuShiTime, *time.Location, error) {
	if input.Longitude == nil {
		return bazi.LiuShiTime{}, nil, fmt.Errorf("longitude is required")
	}
	standard, loc, err := parseStandardTime(input.StartTime)
	if err != nil {
		return bazi.LiuShiTime{}, nil, err
	}
	resolved, err := bazi.ResolveLiuShiTime(standard, loc, *input.Longitude)
	return resolved, loc, err
}

func computeBaziLiuShi(input BaziLiupanInput, core bazi.Core) (BaziLiuShiOutput, error) {
	resolved, loc, err := parseBaziLiuShiTime(input)
	if err != nil {
		return BaziLiuShiOutput{}, fmt.Errorf("bazi.liupan liu_shi: %w", err)
	}
	result, err := bazi.ComputeLiuShi(core, resolved)
	if err != nil {
		return BaziLiuShiOutput{}, fmt.Errorf("bazi.liupan liu_shi: %w", err)
	}
	if result == nil {
		return BaziLiuShiOutput{}, fmt.Errorf("bazi.liupan liu_shi: empty result for %s", resolved.Standard)
	}
	output, err := wireFromDomain[BaziLiuShiOutput](*result)
	if err != nil {
		return BaziLiuShiOutput{}, fmt.Errorf("bazi.liupan liu_shi: %w", err)
	}
	output.ShiChen = BaziLiuShiChen{
		LiuShiStartTime: standardTimeValue(resolved.Standard.In(loc)),
		LiuShiLocalDate: civilDateString(resolved.LocalDate),
		Shichen:         ZhiName(resolved.Shichen.String()),
	}
	return output, nil
}

func advanceBaziLiuShi(input BaziLiupanInput) (BaziLiupanInput, error) {
	resolved, loc, err := parseBaziLiuShiTime(input)
	if err != nil {
		return input, fmt.Errorf("bazi.liupan liu_shi: %w", err)
	}
	next, err := bazi.NextLiuShiTime(resolved)
	if err != nil {
		return input, fmt.Errorf("bazi.liupan liu_shi: %w", err)
	}
	input.StartTime = standardTimeValue(next.Standard.In(loc))
	return input, nil
}
