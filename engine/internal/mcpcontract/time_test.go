package mcpcontract

import (
	"context"
	"encoding/json"
	"strings"
	"testing"
	"time"
)

// time.now's CST field must represent the same instant as UTC, with the
// real Beijing offset.
func TestTimeNowCSTIsRealBeijingTime(t *testing.T) {
	raw, err := NewEngine().Call(context.Background(), "time.now", nil)
	if err != nil {
		t.Fatalf("call: %v", err)
	}
	var output TimeNowOutput
	if err := json.Unmarshal(raw, &output); err != nil {
		t.Fatalf("unmarshal: %v", err)
	}
	if !strings.HasSuffix(output.CST, "+08:00") {
		t.Fatalf("cst offset should be +08:00, got %q", output.CST)
	}
	cst, err := time.Parse(time.RFC3339, output.CST)
	if err != nil {
		t.Fatalf("parse cst: %v", err)
	}
	want := time.Now().In(time.FixedZone("CST", 8*3600))
	if diff := cst.Sub(want); diff < -2*time.Second || diff > 2*time.Second {
		t.Fatalf("cst %s differs from real Beijing time by %v (>2s)", output.CST, diff)
	}
}
