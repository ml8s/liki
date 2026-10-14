package agent

import (
	"context"
	"encoding/json"
	"testing"
)

// China observed DST in 1986-1991. A wall-clock birth time inside the DST
// window may be recorded as +08:00 or +09:00; the engine must expose this
// ambiguity instead of silently treating the supplied offset as standard time.
func TestTianwenTimeFlagsHistoricalChinaDST(t *testing.T) {
	raw, err := tianwenTimeHandler(context.Background(), json.RawMessage(
		`{"time":"1990-06-01T14:00:00+09:00","longitude":116.4}`,
	))
	if err != nil {
		t.Fatalf("tianwen.time: %v", err)
	}

	var env struct {
		Data struct {
			HistoricalTimeHints []struct {
				Kind            string `json:"kind"`
				RequiresConfirm bool   `json:"requires_confirm"`
				RecordedTime    string `json:"recorded_time"`
				StandardTime    string `json:"standard_time"`
				DaylightTime    string `json:"daylight_time"`
			} `json:"historical_time_hints"`
		} `json:"data"`
	}
	if err := json.Unmarshal(raw, &env); err != nil {
		t.Fatalf("unmarshal: %v", err)
	}
	if len(env.Data.HistoricalTimeHints) == 0 {
		t.Fatal("expected historical China DST hint, got none")
	}
	hint := env.Data.HistoricalTimeHints[0]
	if hint.Kind != "china_dst" || !hint.RequiresConfirm {
		t.Fatalf("unexpected hint: %+v", hint)
	}
	if hint.RecordedTime == "" || hint.StandardTime == "" || hint.DaylightTime == "" {
		t.Fatalf("DST hint must include all candidate interpretations: %+v", hint)
	}
}
