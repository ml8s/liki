package mcpcontract

import (
	"context"
	"encoding/json"
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

// QimenChart computes a qimen chart for an explicitly selected method.
func QimenChart(_ context.Context, input QimenChartInput) (json.RawMessage, error) {
	at, loc, err := parseStandardTime(input.StandardTime)
	if err != nil {
		return nil, fmt.Errorf("qimen.chart: %w", err)
	}
	if input.Longitude < -180 || input.Longitude > 180 {
		return nil, fmt.Errorf("qimen.chart: longitude must be within [-180,180], got %v", input.Longitude)
	}
	methodInput := input.Method
	method, err := qimen.ResolveMethodWithOptions(
		string(methodInput.Scope), string(methodInput.School), string(methodInput.DingjuMethod),
		string(methodInput.QuarterRule), string(methodInput.BaseDingjuMethod),
		string(methodInput.DunSource), string(methodInput.HourBoundary),
	)
	if err != nil {
		return nil, fmt.Errorf("qimen.chart: %w", err)
	}
	if method.School == qimen.SchoolJinhanYuJing {
		if len(input.YongShen) > 0 || input.BirthStandardTime != "" {
			return nil, fmt.Errorf("qimen.chart: jinhan_yujing does not support yong_shen or birth_instant")
		}
		return marshalRawOutput("qimen.chart", qimen.ComputeJinhanChartAt(at, loc, input.Longitude))
	}
	var birthDate qimen.BirthDate
	if input.BirthStandardTime != "" {
		birthAt, _, err := parseStandardTime(input.BirthStandardTime)
		if err != nil {
			return nil, fmt.Errorf("qimen.chart: birth: %w", err)
		}
		birthDate = qimen.BirthDate{Time: birthAt, Precision: qimen.BirthDateMoment, Has: true}
	}
	if len(input.YongShen) > 0 || birthDate.Has {
		symbols := make([]qimen.YongShenSymbol, 0, len(input.YongShen))
		for _, name := range input.YongShen {
			symbol, err := qimen.ParseYongShen(name)
			if err != nil {
				return nil, fmt.Errorf("qimen.chart: %w", err)
			}
			symbols = append(symbols, symbol)
		}
		result, err := qimen.ComputeChartWithYongShenAndMethodAt(at, loc, input.Longitude, symbols, birthDate, method)
		if err != nil {
			return nil, fmt.Errorf("qimen.chart: %w", err)
		}
		return marshalRawOutput("qimen.chart", result)
	}
	result, err := qimen.ComputeChartWithMethodAt(at, loc, input.Longitude, method)
	if err != nil {
		return nil, fmt.Errorf("qimen.chart: %w", err)
	}
	return marshalRawOutput("qimen.chart", result)
}
