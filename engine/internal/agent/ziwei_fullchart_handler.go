package agent

import (
	"context"
	"encoding/json"
	"fmt"

	"liki-engine/internal/engine/ziwei"
)

func ziweiFullChartHandler(ctx context.Context, raw json.RawMessage) (json.RawMessage, error) {
	var p struct {
		Chart json.RawMessage `json:"chart"`
	}
	if err := json.Unmarshal(raw, &p); err != nil {
		return nil, fmt.Errorf("ziwei.fullchart: %w", err)
	}
	chart, err := parseChart(p.Chart)
	if err != nil {
		return nil, fmt.Errorf("ziwei.fullchart: %w", err)
	}
	result := ziwei.ComputeFullChart(chart)
	return wrapResult("ziwei_fullchart", result)
}
