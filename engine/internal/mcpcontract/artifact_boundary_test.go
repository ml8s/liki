package mcpcontract

import (
	"context"
	"encoding/json"
	"testing"
)

func TestEngineChartOutputsDoNotCarryDigests(t *testing.T) {
	registry := testEngine()
	for _, method := range []string{"bazi_chart", "ziwei_chart"} {
		t.Run(method, func(t *testing.T) {
			subject := map[string]any{
				"standard_time": "1990-05-20T12:49:00+08:00", "longitude": 116.4, "gender": "male",
			}
			raw, err := registry.Call(context.Background(), method, mustJSON(t, map[string]any{"subject": subject}))
			if err != nil {
				t.Fatal(err)
			}
			var output map[string]any
			decode(t, raw, &output)
			for _, field := range []string{"source_chart_digest", "chart_digest"} {
				if _, exists := output[field]; exists {
					t.Fatalf("%s carried %s", method, field)
				}
			}
		})
	}
}

func TestSeriesSubjectOmitsSourceDigest(t *testing.T) {
	registry := testEngine()
	input := map[string]any{"subject": map[string]any{"standard_time": "1990-05-20T12:49:00+08:00", "longitude": 116.4, "gender": "male"}, "scope": "liu_nian", "standard_time": "2026-06-15T12:00:00+08:00"}
	raw, err := registry.Call(context.Background(), "bazi_liupan", mustJSON(t, input))
	if err != nil {
		t.Fatal(err)
	}
	var output struct {
		LiuNian []map[string]any `json:"liu_nian"`
	}
	if err := json.Unmarshal(raw, &output); err != nil {
		t.Fatal(err)
	}
	if len(output.LiuNian) != 1 {
		t.Fatalf("omitted count items = %d, want 1", len(output.LiuNian))
	}
	if _, exists := output.LiuNian[0]["source_chart_digest"]; exists {
		t.Fatal("derived series carried source digest")
	}

	input["count"] = 0
	if _, err := registry.Call(context.Background(), "bazi_liupan", mustJSON(t, input)); err == nil {
		t.Fatal("explicit series count zero was accepted")
	}
}
