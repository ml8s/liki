package mcpcontract

import (
	"context"
	"encoding/json"
	"testing"
	"time"
)

func TestTimeNowRequiresTimezone(t *testing.T) {
	if _, err := testEngine().Call(context.Background(), "time_now", nil); err == nil {
		t.Fatal("time.now without timezone must fail")
	}
}

func TestTimeNowCST(t *testing.T) {
	raw, err := testEngine().Call(context.Background(), "time_now", json.RawMessage(`{"timezone":"Asia/Shanghai"}`))
	if err != nil {
		t.Fatalf("call: %v", err)
	}
	if err := testEngine().ValidateOutput("time_now", raw); err != nil {
		t.Fatalf("output contract: %v", err)
	}
	var out TimeNowOutput
	if err := json.Unmarshal(raw, &out); err != nil {
		t.Fatal(err)
	}
	standard, err := time.Parse(time.RFC3339Nano, out.LocalStandardTime)
	if err != nil {
		t.Fatalf("local_standard_time: %q %v", out.LocalStandardTime, err)
	}
	if time.Since(standard).Abs() > 2*time.Second {
		t.Fatalf("local_standard_time %s is too far from now", out.LocalStandardTime)
	}
}

func TestTimeNowNewYork(t *testing.T) {
	raw, err := testEngine().Call(context.Background(), "time_now", json.RawMessage(`{"timezone":"America/New_York"}`))
	if err != nil {
		t.Fatalf("call: %v", err)
	}
	if err := testEngine().ValidateOutput("time_now", raw); err != nil {
		t.Fatalf("output contract: %v", err)
	}
	var out TimeNowOutput
	if err := json.Unmarshal(raw, &out); err != nil {
		t.Fatal(err)
	}
}

func TestTianwenSubjectCandidatesOutputContract(t *testing.T) {
	registry := testEngine()
	raw, err := registry.Call(context.Background(), "tianwen_subject_candidates", mustJSON(t, civilResolveInput(t, "1990-05-20T03:49:00Z", "Asia/Shanghai", 116.4)))
	if err != nil {
		t.Fatalf("tianwen_subject_candidates: %v", err)
	}
	if err := registry.ValidateOutput("tianwen_subject_candidates", raw); err != nil {
		t.Fatalf("tianwen_subject_candidates output contract: %v", err)
	}
}
