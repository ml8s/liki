package mcpcontract

import (
	"context"
	"encoding/json"
	"strings"
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
	var out TimeNowOutput
	if err := json.Unmarshal(raw, &out); err != nil {
		t.Fatal(err)
	}
	if out.Timezone != "Asia/Shanghai" {
		t.Fatalf("unexpected output: %+v", out)
	}
	instant, err := time.Parse(time.RFC3339, out.Instant)
	if err != nil || !strings.HasSuffix(out.Instant, "Z") {
		t.Fatalf("instant: %q %v", out.Instant, err)
	}
	if time.Since(instant).Abs() > 2*time.Second {
		t.Fatalf("instant %s is too far from now", out.Instant)
	}
}

func TestTimeNowNewYork(t *testing.T) {
	raw, err := testEngine().Call(context.Background(), "time_now", json.RawMessage(`{"timezone":"America/New_York"}`))
	if err != nil {
		t.Fatalf("call: %v", err)
	}
	var out TimeNowOutput
	if err := json.Unmarshal(raw, &out); err != nil {
		t.Fatal(err)
	}
	if out.Timezone != "America/New_York" {
		t.Fatalf("unexpected output: %+v", out)
	}
}
