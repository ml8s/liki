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
	Instant   Instant    `json:"instant" jsonschema:"出生瞬间 RFC3339；接受 UTC Z 或 numeric offset，服务端归一化为 UTC。"`
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
	if chart.BirthYear == 0 || chart.Gender == "" {
		return bazi.Chart{}, fmt.Errorf("bazi chart is empty")
	}
	gender, err := ganzhi.ParseGender(string(chart.Gender))
	if err != nil {
		return bazi.Chart{}, fmt.Errorf("decode bazi chart: %w", err)
	}
	pillars, naYin, err := baziPillarsDomain(chart)
	if err != nil {
		return bazi.Chart{}, fmt.Errorf("decode bazi chart: %w", err)
	}
	daYun, err := baziDaYunDomain(chart.DaYun)
	if err != nil {
		return bazi.Chart{}, fmt.Errorf("decode bazi chart: %w", err)
	}
	return bazi.NewChartFromPillars(
		pillars, naYin, daYun, gender, chart.BirthYear, chart.ZiShiRule,
	), nil
}

func baziPillarsDomain(chart BaziChart) ([4]ganzhi.Zhu, [4]string, error) {
	var pillars [4]ganzhi.Zhu
	var naYin [4]string
	for index, pillar := range [4]BaziPillar{chart.Nian, chart.Yue, chart.Ri, chart.Shi} {
		gan, err := ganzhi.ParseGan(string(pillar.Gan))
		if err != nil {
			return pillars, naYin, err
		}
		zhi, err := ganzhi.ParseZhi(string(pillar.Zhi))
		if err != nil {
			return pillars, naYin, err
		}
		pillars[index] = ganzhi.Zhu{Gan: gan, Zhi: zhi}
		naYin[index] = pillar.NaYin
	}
	return pillars, naYin, nil
}

func baziDaYunDomain(dayun *BaziDaYun) (*bazi.DaYun, error) {
	if dayun == nil {
		return nil, nil
	}
	out := &bazi.DaYun{
		StartDate:       string(dayun.StartDate),
		StartYearAfter:  dayun.StartYearAfter,
		StartMonthAfter: dayun.StartMonthAfter,
		StartDayAfter:   dayun.StartDayAfter,
		Direction:       dayun.Direction,
	}
	out.Steps = make([]bazi.DaYunStep, len(dayun.Steps))
	for index, step := range dayun.Steps {
		gan, err := ganzhi.ParseGan(string(step.Gan))
		if err != nil {
			return nil, err
		}
		zhi, err := ganzhi.ParseZhi(string(step.Zhi))
		if err != nil {
			return nil, err
		}
		out.Steps[index] = bazi.DaYunStep{
			Gan: gan, Zhi: zhi,
			StartDate: string(step.StartDate), EndDate: string(step.EndDate),
			StartYear: step.StartYear, EndYear: step.EndYear,
			Name: step.Name, Element: string(step.Element), ShiShen: step.ShiShen,
			Rooted: step.Rooted,
		}
		if len(step.RootRefs) > 0 {
			refs := make([]bazi.DaYunRootRef, len(step.RootRefs))
			for refIndex, ref := range step.RootRefs {
				refs[refIndex] = bazi.DaYunRootRef{
					Source: ref.Source, Pillar: ref.Pillar,
					Stem: ref.Stem, Branch: ref.Branch,
				}
			}
			out.Steps[index].RootRefs = refs
		}
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
func BaziFullChart(_ context.Context, input BaziFullChartInput) (json.RawMessage, error) {
	chart, err := input.Chart.domain()
	if err != nil {
		return nil, fmt.Errorf("bazi.fullchart: %w", err)
	}
	if err := validateFullChartInput(chart); err != nil {
		return nil, err
	}
	return marshalRawBaziFullChart(bazi.ComputeFullChart(chart))
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
	chartA, err := input.A.Chart.domain()
	if err != nil {
		return nil, fmt.Errorf("bazi.bond: chart a: %w", err)
	}
	chartB, err := input.B.Chart.domain()
	if err != nil {
		return nil, fmt.Errorf("bazi.bond: chart b: %w", err)
	}
	return marshalRawOutput("bazi.bond", bazi.ComputeBond(chartA, chartB))
}

// BaziLiuNian computes the year pillar and its interactions with a chart.
func BaziLiuNian(_ context.Context, input BaziLiuNianInput) (any, error) {
	return runSeries(input, int(input.Count), MaxSeriesCount,
		func(in BaziLiuNianInput) (BaziLiuNianOutput, error) {
			return computeBaziLiuNian(in)
		},
		func(in BaziLiuNianInput) (BaziLiuNianInput, error) {
			in.Year++
			return in, nil
		},
	)
}

func computeBaziLiuNian(input BaziLiuNianInput) (BaziLiuNianOutput, error) {
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
func BaziLiuYue(_ context.Context, input BaziLiuYueInput) (any, error) {
	return runSeries(input, int(input.Count), MaxSeriesCount,
		func(in BaziLiuYueInput) (BaziLiuYueOutput, error) {
			return computeBaziLiuYue(in)
		},
		func(in BaziLiuYueInput) (BaziLiuYueInput, error) {
			in.Year, in.Month = advanceGregorianMonth(in.Year, in.Month)
			return in, nil
		},
	)
}

func computeBaziLiuYue(input BaziLiuYueInput) (BaziLiuYueOutput, error) {
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
func BaziLiuRi(_ context.Context, input BaziLiuRiInput) (any, error) {
	return runSeries(input, int(input.Count), MaxSeriesCount,
		func(in BaziLiuRiInput) (BaziLiuRiOutput, error) {
			return computeBaziLiuRi(in)
		},
		func(in BaziLiuRiInput) (BaziLiuRiInput, error) {
			in.Year, in.Month, in.Day = advanceGregorianDay(in.Year, in.Month, in.Day)
			return in, nil
		},
	)
}

func computeBaziLiuRi(input BaziLiuRiInput) (BaziLiuRiOutput, error) {
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
func BaziLiuShi(_ context.Context, input BaziLiuShiInput) (any, error) {
	return runSeries(input, int(input.Count), MaxSeriesCount,
		func(in BaziLiuShiInput) (BaziLiuShiOutput, error) {
			return computeBaziLiuShi(in)
		},
		func(in BaziLiuShiInput) (BaziLiuShiInput, error) {
			in.Year, in.Month, in.Day, in.Hour = advanceGregorianHour(in.Year, in.Month, in.Day, in.Hour)
			return in, nil
		},
	)
}

func computeBaziLiuShi(input BaziLiuShiInput) (BaziLiuShiOutput, error) {
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
func BaziXiaoYun(_ context.Context, input BaziXiaoYunInput) (any, error) {
	n, err := validateCount(int(input.Count), MaxSeriesCount)
	if err != nil {
		return nil, fmt.Errorf("bazi.xiaoyun: %w", err)
	}
	chart, err := input.Chart.domain()
	if err != nil {
		return nil, fmt.Errorf("bazi.xiaoyun: %w", err)
	}
	return wireFromDomain[BaziXiaoYunOutput](bazi.ComputeXiaoYun(chart, n))
}
