package mcpcontract

import (
	"encoding/json"
	"fmt"
	"time"
)

// MaxSeriesCount is the unified upper bound for batch time-series queries.
const MaxSeriesCount = 50

// SeriesResult is the batch output wrapper. Items length equals count.
type SeriesResult[T any] struct {
	Items []T `json:"items" jsonschema:"批量结果数组，长度 = count。"`
}

// validateCount normalizes a user-provided series count. 0 and 1 both mean
// "single item" (backward compatible); negative or > max is rejected.
func validateCount(count, max int) (int, error) {
	if count < 0 {
		return 0, fmt.Errorf("count must be non-negative, got %d", count)
	}
	if count > max {
		return 0, fmt.Errorf("count %d exceeds maximum %d", count, max)
	}
	if count == 0 {
		return 1, nil
	}
	return count, nil
}

// runSeries is the stateless generic time-series executor.
// compute: pure function from input to a single result.
// advance: pure function from input to the next input (time step).
// count<=1 returns compute(input) directly (single-item format).
// count>1 returns SeriesResult[Out]{Items: [...]}.
func runSeries[In any, Out any](
	input In, count, max int,
	compute func(In) (Out, error),
	advance func(In) (In, error),
) (any, error) {
	n, err := validateCount(count, max)
	if err != nil {
		return nil, err
	}
	if n <= 1 {
		return compute(input)
	}
	items := make([]Out, 0, n)
	cur := input
	for i := 0; i < n; i++ {
		out, err := compute(cur)
		if err != nil {
			return nil, err
		}
		items = append(items, out)
		cur, err = advance(cur)
		if err != nil {
			return nil, err
		}
	}
	return SeriesResult[Out]{Items: items}, nil
}

// runSeriesPrepared is the stateless generic time-series executor with a
// one-time prepare step. Use this when compute needs decoded shared state
// (e.g. a chart) that is invariant across the series: prepare decodes it
// once, then compute receives both the current input and the decoded context.
func runSeriesPrepared[In any, Ctx any, Out any](
	input In, count, max int,
	prepare func(In) (Ctx, error),
	compute func(In, Ctx) (Out, error),
	advance func(In) (In, error),
) (any, error) {
	n, err := validateCount(count, max)
	if err != nil {
		return nil, err
	}
	if n <= 1 {
		ctx, err := prepare(input)
		if err != nil {
			return nil, err
		}
		return compute(input, ctx)
	}
	ctx, err := prepare(input)
	if err != nil {
		return nil, err
	}
	items := make([]Out, 0, n)
	cur := input
	for i := 0; i < n; i++ {
		out, err := compute(cur, ctx)
		if err != nil {
			return nil, err
		}
		items = append(items, out)
		cur, err = advance(cur)
		if err != nil {
			return nil, err
		}
	}
	return SeriesResult[Out]{Items: items}, nil
}

// advanceGregorianDay steps a Gregorian date forward by one day using
// time.Date for correct month/year rollover.
func advanceGregorianDay(year, month, day int) (int, int, int) {
	t := time.Date(year, time.Month(month), day, 0, 0, 0, 0, time.UTC)
	next := t.AddDate(0, 0, 1)
	return next.Year(), int(next.Month()), next.Day()
}

// advanceGregorianMonth steps a Gregorian year/month forward by one month
// with year rollover (month 12 → next year month 1).
func advanceGregorianMonth(year, month int) (int, int) {
	if month >= 12 {
		return year + 1, 1
	}
	return year, month + 1
}

// advanceGregorianHour steps a Gregorian date+hour forward by one hour
// with day/month/year rollover.
func advanceGregorianHour(year, month, day, hour int) (int, int, int, int) {
	t := time.Date(year, time.Month(month), day, hour, 0, 0, 0, time.UTC)
	next := t.Add(time.Hour)
	return next.Year(), int(next.Month()), next.Day(), next.Hour()
}

// seriesOutputSchema builds a JSON Schema for a series tool's output:
// either the single Out format or the batch {items: [Out]} format.
// The singleSchema parameter is the marshalled JSON Schema for Out.
func seriesOutputSchema(singleSchema json.RawMessage) (json.RawMessage, error) {
	wrapper, err := json.Marshal(map[string]any{
		"type": "object",
		"properties": map[string]any{
			"items": map[string]any{
				"type":     "array",
				"items":    json.RawMessage(singleSchema),
				"minItems": 2,
			},
		},
		"required": []string{"items"},
	})
	if err != nil {
		return nil, err
	}
	combined, err := json.Marshal(map[string]any{
		"anyOf": []json.RawMessage{
			json.RawMessage(singleSchema),
			json.RawMessage(wrapper),
		},
	})
	return combined, err
}
