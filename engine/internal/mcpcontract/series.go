package mcpcontract

import (
	"encoding/json"
	"fmt"

	"github.com/google/jsonschema-go/jsonschema"
)

// MaxSeriesCount is the unified upper bound for batch time-series queries.
const MaxSeriesCount = 50

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

// runSeries returns domain-owned items; the caller wraps them in a named
// series root (liu_nian/liu_yue/...). There is no generic "items" wire field.
func runSeries[In any, Out any](
	input In, count, max int,
	compute func(In) (Out, error),
	advance func(In) (In, error),
) ([]Out, error) {
	n, err := validateSeriesCount(count, max)
	if err != nil {
		return nil, err
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
	return items, nil
}

// runSeriesPrepared is runSeries with a one-time shared-state prepare step.
func runSeriesPrepared[In any, Ctx any, Out any](
	input In, count, max int,
	prepare func(In) (Ctx, error),
	compute func(In, Ctx) (Out, error),
	advance func(In) (In, error),
) ([]Out, error) {
	n, err := validateSeriesCount(count, max)
	if err != nil {
		return nil, err
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
	return items, nil
}

// seriesOutputSchema publishes a named domain root. The MCP root remains an
// object for every negotiated protocol version, but the property is domain
// vocabulary rather than a generic transport word.
func seriesOutputSchema[Out any](options *jsonschema.ForOptions, root string) (json.RawMessage, error) {
	schema, err := jsonschema.For[Out](options)
	if err != nil {
		return nil, err
	}
	return json.Marshal(map[string]any{
		"type": "object",
		"properties": map[string]any{
			root: map[string]any{
				"type":     "array",
				"items":    mustJSONSchemaRaw(schema),
				"minItems": 1,
			},
		},
		"required":             []string{root},
		"additionalProperties": false,
	})
}

func mustJSONSchemaRaw(schema *jsonschema.Schema) json.RawMessage {
	raw, err := schema.MarshalJSON()
	if err != nil {
		panic(err)
	}
	return raw
}
