package mcpcontract

import (
	"encoding/json"
	"fmt"
)

// MaxSeriesCount is the unified upper bound for batch time-series queries.
const MaxSeriesCount = 50

// SeriesResult is the series output wrapper. Items length equals count.
type SeriesResult[T any] struct {
	Items []T `json:"items" jsonschema:"结果数组，长度 = count。"`
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
// Every count returns SeriesResult[Out]{Items: [...]}: the MCP outputSchema
// root must stay a JSON object on every negotiated protocol version (wire
// models through 2025-11-25 reject top-level unions), so single results are
// items arrays of length one.
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
		out, err := compute(input)
		if err != nil {
			return nil, err
		}
		return SeriesResult[Out]{Items: []Out{out}}, nil
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
		out, err := compute(input, ctx)
		if err != nil {
			return nil, err
		}
		return SeriesResult[Out]{Items: []Out{out}}, nil
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

// seriesOutputSchema builds a JSON Schema for a series tool's output:
// always the object-rooted {items: [Out]} format. The singleSchema parameter
// is the marshalled JSON Schema for Out.
func seriesOutputSchema(singleSchema json.RawMessage) (json.RawMessage, error) {
	return json.Marshal(map[string]any{
		"type": "object",
		"properties": map[string]any{
			"items": map[string]any{
				"type":     "array",
				"items":    json.RawMessage(singleSchema),
				"minItems": 1,
			},
		},
		"required":             []string{"items"},
		"additionalProperties": false,
	})
}
