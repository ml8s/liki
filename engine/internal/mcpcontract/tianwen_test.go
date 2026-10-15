package mcpcontract

import (
	"context"
	"testing"
)

// China observed DST in 1986-1991. A wall-clock birth time inside the DST
// window may be recorded as +08:00 or +09:00; the engine must expose this
// ambiguity instead of silently treating the supplied offset as standard time.
func TestTianwenTimeFlagsHistoricalChinaDST(t *testing.T) {
	output, err := TianwenTime(context.Background(), TianwenTimeInput{
		Time:      "1990-06-01T14:00:00+09:00",
		Longitude: 116.4,
	})
	if err != nil {
		t.Fatalf("tianwen.time: %v", err)
	}
	if len(output.HistoricalTimeHints) == 0 {
		t.Fatal("expected historical China DST hint, got none")
	}
	hint := output.HistoricalTimeHints[0]
	if hint.Kind != "china_dst" || !hint.RequiresConfirm {
		t.Fatalf("unexpected hint: %+v", hint)
	}
	if hint.RecordedTime == "" || hint.StandardTime == "" || hint.DaylightTime == "" {
		t.Fatalf("DST hint must include all candidate interpretations: %+v", hint)
	}
}

func TestTianwenTimeInvalid(t *testing.T) {
	for _, input := range []TianwenTimeInput{
		{Time: "", Longitude: 116.4},
		{Time: "1984-02-15", Longitude: 116.4},
		{Time: "1984-02-15T08:00:00", Longitude: 116.4},
	} {
		if _, err := TianwenTime(context.Background(), input); err == nil {
			t.Errorf("expected invalid time error for %#v", input)
		}
	}
}
