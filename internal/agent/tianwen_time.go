package agent

import (
	"context"
	"encoding/json"
	"fmt"
)

func tianwenTimeHandler(ctx context.Context, raw json.RawMessage) (json.RawMessage, error) {
	var tp TimePoint
	if err := json.Unmarshal(raw, &tp); err != nil {
		return nil, fmt.Errorf("tianwen.time: %w", err)
	}
	ts, err := tp.Timeset()
	if err != nil {
		return nil, fmt.Errorf("tianwen.time: %w", err)
	}
	return wrapResult("tianwen_time", ts)
}
