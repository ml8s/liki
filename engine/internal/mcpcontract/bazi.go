package mcpcontract

import (
	"context"
	"encoding/json"
	"fmt"

	"liki-engine/internal/engine/bazi"
	"liki-engine/internal/engine/ganzhi"
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

// BaziChartInput takes the canonical {instant, timezone} pair (Zone is IANA) and the
// birth/query longitude. The boundary derives LocalCivilTime and LAST; the
// chart domain consumes those domain scales (year/month = solar terms (civil/instant); day/hour = LAST + 子时; no Gregorian/offset transit).
type BaziChartInput struct {
	Instant   Instant    `json:"instant" jsonschema:"出生瞬间 UTC（RFC3339，Z）。"`
	Timezone  string     `json:"timezone" jsonschema:"出生地 IANA 时区，如 Asia/Shanghai。"`
	Longitude float64    `json:"longitude" jsonschema:"出生地经度（-180..180）；真太阳时校正用于日柱/时柱。"`
	Gender    GenderName `json:"gender" jsonschema:"性别。"`
}

type BaziChart struct {
	Nian      BaziPillar `json:"nian" jsonschema:"年柱。"`
	Yue       BaziPillar `json:"yue" jsonschema:"月柱。"`
	Ri        BaziPillar `json:"ri" jsonschema:"日柱。"`
	Shi       BaziPillar `json:"shi" jsonschema:"时柱。"`
	DaYun     *BaziDaYun `json:"da_yun" jsonschema:"大运。"`
	Gender    GenderName `json:"gender" jsonschema:"性别。"`
	BirthYear int        `json:"birth_year" jsonschema:"出生公历年。"`
	ZiShiRule string     `json:"zi_shi_rule" jsonschema:"子时换日规则。"`
}

func BaziChartCompute(_ context.Context, input BaziChartInput) (BaziChart, error) {
	at, err := parseInstant(string(input.Instant))
	if err != nil {
		return BaziChart{}, fmt.Errorf("bazi.chart: %w", err)
	}

	loc, err := loadTimezone(input.Timezone)
	if err != nil {
		return BaziChart{}, fmt.Errorf("bazi.chart: %w", err)
	}

	if input.Longitude < -180 || input.Longitude > 180 {
		return BaziChart{}, fmt.Errorf("bazi.chart: longitude must be within [-180,180], got %v", input.Longitude)
	}

	chart := bazi.ComputeChartAt(at, loc, input.Longitude, ganzhi.Gender(input.Gender))
	return baziChartWire(chart), nil
}

func (chart BaziChart) domain() (bazi.Chart, error) {
	raw, err := json.Marshal(chart)
	if err != nil {
		return bazi.Chart{}, err
	}
	var out bazi.Chart
	if err := json.Unmarshal(raw, &out); err != nil {
		return bazi.Chart{}, fmt.Errorf("decode bazi chart: %w", err)
	}
	if out.BirthYear == 0 || out.Gender == "" {
		return bazi.Chart{}, fmt.Errorf("bazi chart is empty")
	}
	return out, nil
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

// validateFullChartInput ports the legacy bazi.fullchart chart validation so the
// typed tool keeps rejecting charts that the domain cannot expand.
func validateFullChartInput(chart bazi.Chart) error {
	pillars := []struct {
		name  string
		gan   ganzhi.Gan
		zhi   ganzhi.Zhi
		naYin string
	}{
		{"nian", chart.Nian.Gan, chart.Nian.Zhi, chart.Nian.NaYin},
		{"yue", chart.Yue.Gan, chart.Yue.Zhi, chart.Yue.NaYin},
		{"ri", chart.Ri.Gan, chart.Ri.Zhi, chart.Ri.NaYin},
		{"shi", chart.Shi.Gan, chart.Shi.Zhi, chart.Shi.NaYin},
	}
	for _, pillar := range pillars {
		if pillar.gan == 0 || pillar.zhi == 0 {
			return fmt.Errorf("bazi.fullchart: chart.%s requires non-empty gan and zhi", pillar.name)
		}
		if int(pillar.gan)%2 != int(pillar.zhi)%2 {
			return fmt.Errorf(
				"bazi.fullchart: chart.%s has invalid sexagenary pair %s%s",
				pillar.name, ganzhi.GanName(pillar.gan), ganzhi.ZhiName(pillar.zhi),
			)
		}
		if pillar.naYin == "" {
			return fmt.Errorf("bazi.fullchart: chart.%s.na_yin is required", pillar.name)
		}
		if expected := ganzhi.NayinLabel(pillar.gan, pillar.zhi); pillar.naYin != expected {
			return fmt.Errorf(
				"bazi.fullchart: chart.%s.na_yin is %q, want %q for %s%s",
				pillar.name, pillar.naYin, expected,
				ganzhi.GanName(pillar.gan), ganzhi.ZhiName(pillar.zhi),
			)
		}
	}
	if chart.Gender != ganzhi.Male && chart.Gender != ganzhi.Female {
		return fmt.Errorf("bazi.fullchart: chart.gender must be male or female")
	}
	return nil
}

// BaziFullChart expands a bazi.chart wire chart into the full chart.
func BaziFullChart(_ context.Context, input BaziFullChartInput) (BaziFullChartOutput, error) {
	chart, err := input.Chart.domain()
	if err != nil {
		return BaziFullChartOutput{}, fmt.Errorf("bazi.fullchart: %w", err)
	}
	if err := validateFullChartInput(chart); err != nil {
		return BaziFullChartOutput{}, err
	}
	return wireFromDomain[BaziFullChartOutput](bazi.ComputeFullChart(chart))
}

// BaziBond returns raw cross-chart facts for two bazi charts.
func BaziBond(_ context.Context, input BaziBondInput) (BaziBondOutput, error) {
	chartA, err := input.A.Chart.domain()
	if err != nil {
		return BaziBondOutput{}, fmt.Errorf("bazi.bond: chart a: %w", err)
	}
	chartB, err := input.B.Chart.domain()
	if err != nil {
		return BaziBondOutput{}, fmt.Errorf("bazi.bond: chart b: %w", err)
	}
	return wireFromDomain[BaziBondOutput](bazi.ComputeBond(chartA, chartB))
}

// BaziLiuNian computes the year pillar and its interactions with a chart.
func BaziLiuNian(_ context.Context, input BaziLiuNianInput) (BaziLiuNianOutput, error) {
	if input.Year <= 0 {
		return BaziLiuNianOutput{}, fmt.Errorf("bazi.liunian: year must be positive, got %d", input.Year)
	}
	chart, err := input.Chart.domain()
	if err != nil {
		return BaziLiuNianOutput{}, fmt.Errorf("bazi.liunian: %w", err)
	}
	result, err := bazi.ComputeLiuNian(chart, input.Year)
	if err != nil {
		return BaziLiuNianOutput{}, fmt.Errorf("bazi.liunian: %w", err)
	}
	if result == nil {
		return BaziLiuNianOutput{}, fmt.Errorf("bazi.liunian: empty result for year %d", input.Year)
	}
	return wireFromDomain[BaziLiuNianOutput](*result)
}

// BaziLiuYue computes the month pillar and its interactions with a chart.
func BaziLiuYue(_ context.Context, input BaziLiuYueInput) (BaziLiuYueOutput, error) {
	chart, err := input.Chart.domain()
	if err != nil {
		return BaziLiuYueOutput{}, fmt.Errorf("bazi.liuyue: %w", err)
	}
	result, err := bazi.ComputeLiuYue(chart, input.Year, input.Month)
	if err != nil {
		return BaziLiuYueOutput{}, fmt.Errorf("bazi.liuyue: %w", err)
	}
	if result == nil {
		return BaziLiuYueOutput{}, fmt.Errorf("bazi.liuyue: empty result for %d-%d", input.Year, input.Month)
	}
	return wireFromDomain[BaziLiuYueOutput](*result)
}

// BaziLiuRi computes the day pillar and its interactions with a chart.
func BaziLiuRi(_ context.Context, input BaziLiuRiInput) (BaziLiuRiOutput, error) {
	chart, err := input.Chart.domain()
	if err != nil {
		return BaziLiuRiOutput{}, fmt.Errorf("bazi.liuri: %w", err)
	}
	result, err := bazi.ComputeLiuRi(chart, input.Year, input.Month, input.Day)
	if err != nil {
		return BaziLiuRiOutput{}, fmt.Errorf("bazi.liuri: %w", err)
	}
	if result == nil {
		return BaziLiuRiOutput{}, fmt.Errorf("bazi.liuri: empty result for %d-%d-%d", input.Year, input.Month, input.Day)
	}
	return wireFromDomain[BaziLiuRiOutput](*result)
}

// BaziLiuShi computes the hour pillar and its interactions with a chart.
func BaziLiuShi(_ context.Context, input BaziLiuShiInput) (BaziLiuShiOutput, error) {
	if input.Hour < 0 || input.Hour > 23 {
		return BaziLiuShiOutput{}, fmt.Errorf("bazi.liushi: hour must be 0-23, got %d", input.Hour)
	}
	chart, err := input.Chart.domain()
	if err != nil {
		return BaziLiuShiOutput{}, fmt.Errorf("bazi.liushi: %w", err)
	}
	result, err := bazi.ComputeLiuShi(chart, input.Year, input.Month, input.Day, input.Hour)
	if err != nil {
		return BaziLiuShiOutput{}, fmt.Errorf("bazi.liushi: %w", err)
	}
	if result == nil {
		return BaziLiuShiOutput{}, fmt.Errorf("bazi.liushi: empty result for %d-%d-%d %02d", input.Year, input.Month, input.Day, input.Hour)
	}
	return wireFromDomain[BaziLiuShiOutput](*result)
}

// BaziXiaoYun returns the 小运 year list for a chart.
func BaziXiaoYun(_ context.Context, input BaziXiaoYunInput) (BaziXiaoYunOutput, error) {
	chart, err := input.Chart.domain()
	if err != nil {
		return BaziXiaoYunOutput{}, fmt.Errorf("bazi.xiaoyun: %w", err)
	}
	return wireFromDomain[BaziXiaoYunOutput](bazi.ComputeXiaoYun(chart, input.Count))
}
