package mcpcontract

import (
	"encoding/json"
	"fmt"
)

// MaxSeriesCount is the unified upper bound for batch time-series queries.
const MaxSeriesCount = 50

// SeriesResult is the series output wrapper. Items length equals count.
type SeriesResult[T any] struct {
	SourceChartDigest *ChartDigest `json:"source_chart_digest,omitempty" jsonschema:"来源 chart 的 canonical SHA-256 digest；无 source chart 时省略。"`
	Items             []T          `json:"items" jsonschema:"结果数组，长度 = count。"`
}

func seriesCountValue(count *SeriesCount) int {
	if count == nil {
		return 1
	}
	return int(*count)
}

func validateSeriesCount(count, max int) (int, error) {
	if count < 1 {
		return 0, fmt.Errorf("count must be at least 1, got %d", count)
	}
	if count > max {
		return 0, fmt.Errorf("count %d exceeds maximum %d", count, max)
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
	sourceDigest, err := inputSourceDigest(input)
	if err != nil {
		return nil, err
	}
	n, err := validateSeriesCount(count, max)
	if err != nil {
		return nil, err
	}
	if n <= 1 {
		out, err := compute(input)
		if err != nil {
			return nil, err
		}
		return SeriesResult[Out]{SourceChartDigest: sourceDigest, Items: []Out{out}}, nil
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
	return SeriesResult[Out]{SourceChartDigest: sourceDigest, Items: items}, nil
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
	sourceDigest, err := inputSourceDigest(input)
	if err != nil {
		return nil, err
	}
	n, err := validateSeriesCount(count, max)
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
		return SeriesResult[Out]{SourceChartDigest: sourceDigest, Items: []Out{out}}, nil
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
	return SeriesResult[Out]{SourceChartDigest: sourceDigest, Items: items}, nil
}

// seriesOutputSchema builds a JSON Schema for a series tool's output:
// always the object-rooted {items: [Out]} format. The singleSchema parameter
// is the marshalled JSON Schema for Out.
func seriesOutputSchema(singleSchema json.RawMessage) (json.RawMessage, error) {
	return json.Marshal(map[string]any{
		"type": "object",
		"properties": map[string]any{
			"source_chart_digest": map[string]any{
				"anyOf": []any{
					map[string]any{"type": "string", "pattern": "^[a-f0-9]{64}$"},
					map[string]any{"type": "null"},
				},
				"description": "来源 chart 的 canonical SHA-256 digest；无 source chart 时省略。",
			},
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
