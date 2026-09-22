package agent

import (
	"context"
	"encoding/json"
	"strings"
	"testing"
	"time"
)

// time.now 的 cst 字段必须表示同一时刻的北京时间。
func TestTimeNowCSTIsRealBeijingTime(t *testing.T) {
	raw, err := timeNowHandler(context.Background(), json.RawMessage("{}"))
	if err != nil {
		t.Fatalf("handler: %v", err)
	}
	var out struct {
		Data struct {
			UTC   string `json:"utc"`
			Local string `json:"local"`
			CST   string `json:"cst"`
		} `json:"data"`
	}
	if err := json.Unmarshal(raw, &out); err != nil {
		t.Fatalf("unmarshal: %v", err)
	}
	cstStr := out.Data.CST
	if !strings.HasSuffix(cstStr, "+08:00") {
		t.Fatalf("cst 偏移应为 +08:00，got %q", cstStr)
	}
	cst, err := time.Parse(time.RFC3339, cstStr)
	if err != nil {
		t.Fatalf("parse cst: %v", err)
	}
	want := time.Now().In(time.FixedZone("CST", 8*3600))
	if d := cst.Sub(want); d < -2*time.Second || d > 2*time.Second {
		t.Fatalf("cst %s 与真实北京时间 %s 偏差 %v（>2s）", cstStr, want.Format(time.RFC3339), d)
	}
}
