package mcpcontract

import (
	"context"
	"fmt"

	"liki-engine/internal/engine/qimen"
)

// QimenYongShenSymbols is the 用神 symbol combination accepted by qimen.chart.
// It is not a plain []string because the legacy params schema declared
// `minItems: 1` and `uniqueItems: true`; the named type carries those
// constraints through the registry's TypeSchemas (an empty or duplicated list
// used to be a schema error, and silently ignoring an empty list would drop the
// 用神 analysis the caller asked for).
type QimenYongShenSymbols []string

// QimenChart computes a qimen chart for the requested scope and school.
// scope/school/dingju knobs keep the legacy defaults when they are omitted.
func QimenChart(_ context.Context, input QimenChartInput) (QimenChartOutput, error) {
	at, err := parseInstant(string(input.Instant))
	if err != nil {
		return QimenChartOutput{}, fmt.Errorf("qimen.chart: %w", err)
	}
	loc, err := loadTimezone(input.Timezone)
	if err != nil {
		return QimenChartOutput{}, fmt.Errorf("qimen.chart: %w", err)
	}
	if input.Longitude < -180 || input.Longitude > 180 {
		return QimenChartOutput{}, fmt.Errorf("qimen.chart: longitude must be within [-180,180], got %v", input.Longitude)
	}
	method, err := qimen.ResolveMethodWithOptions(
		string(input.Scope), string(input.School), string(input.DingjuMethod),
		string(input.QuarterRule), string(input.BaseDingjuMethod),
		string(input.DunSource), string(input.HourBoundary),
	)
	if err != nil {
		return QimenChartOutput{}, fmt.Errorf("qimen.chart: %w", err)
	}
	if method.School == qimen.SchoolJinhanYuJing {
		if len(input.YongShen) > 0 || input.BirthDate != "" {
			return QimenChartOutput{}, fmt.Errorf("qimen.chart: jinhan_yujing does not support yong_shen or birth_date")
		}
		return wireFromDomain[QimenChartOutput](qimen.ComputeJinhanChartAt(at, loc, input.Longitude))
	}
	var birthDate qimen.BirthDate
	if input.BirthDate != "" {
		birthDate, err = qimen.ParseBirthDate(input.BirthDate)
		if err != nil {
			return QimenChartOutput{}, fmt.Errorf("qimen.chart: %w", err)
		}
	}
	if len(input.YongShen) > 0 || birthDate.Has {
		symbols := make([]qimen.YongShenSymbol, 0, len(input.YongShen))
		for _, name := range input.YongShen {
			symbol, err := qimen.ParseYongShen(name)
			if err != nil {
				return QimenChartOutput{}, fmt.Errorf("qimen.chart: %w", err)
			}
			symbols = append(symbols, symbol)
		}
		result, err := qimen.ComputeChartWithYongShenAndMethodAt(at, loc, input.Longitude, symbols, birthDate, method)
		if err != nil {
			return QimenChartOutput{}, fmt.Errorf("qimen.chart: %w", err)
		}
		return wireFromDomain[QimenChartOutput](result)
	}
	result, err := qimen.ComputeChartWithMethodAt(at, loc, input.Longitude, method)
	if err != nil {
		return QimenChartOutput{}, fmt.Errorf("qimen.chart: %w", err)
	}
	return wireFromDomain[QimenChartOutput](result)
}
