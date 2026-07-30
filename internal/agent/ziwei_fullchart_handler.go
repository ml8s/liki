package agent

import (
	"context"
	"encoding/json"
	"fmt"

	"liki-engine/internal/engine/ziwei"
)

func ziweiFullChartHandler(ctx context.Context, raw json.RawMessage) (json.RawMessage, error) {
	var p struct {
		Chart  json.RawMessage `json:"chart"`
		RiGan  int             `json:"ri_gan,omitempty"`
		RiZhi  int             `json:"ri_zhi,omitempty"`
	}
	if err := json.Unmarshal(raw, &p); err != nil {
		return nil, fmt.Errorf("ziwei.fullchart: %w", err)
	}
	var chart ziwei.Chart
	if err := json.Unmarshal(p.Chart, &chart); err != nil {
		return nil, fmt.Errorf("ziwei.fullchart: parse chart: %w", err)
	}
	result := ziwei.ComputeFullChart(chart, p.RiGan, p.RiZhi)
	return wrapResult("ziwei_fullchart", result)
}
