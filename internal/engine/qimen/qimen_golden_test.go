package qimen

import (
	"encoding/json"
	"os"
	"path/filepath"
	"testing"
	"time"

	"liki-engine/internal/engine/tianwen"
)

var updateGolden = os.Getenv("UPDATE_GOLDEN") == "1"

func TestGoldenComputeChart(t *testing.T) {
	st := tianwen.GregorianToSolar(
		time.Date(2026, 6, 28, 12, 0, 0, 0, time.FixedZone("CST", 8*3600)),
		116.4, 8,
	)
	chart := ComputeChart(st, "时家")

	// 命理锚点断言（独立于 golden 文件——UPDATE_GOLDEN=1 时同样执行）。
	assertChartAnchors(t, chart)

	got, err := json.MarshalIndent(chart, "", "  ")
	if err != nil {
		t.Fatalf("marshal: %v", err)
	}

	golden := filepath.Join("testdata", "chart_golden.json")
	if updateGolden {
		if err := os.MkdirAll("testdata", 0755); err != nil {
			t.Fatalf("mkdir testdata: %v", err)
		}
		if err := os.WriteFile(golden, got, 0644); err != nil {
			t.Fatalf("write golden: %v", err)
		}
		return
	}

	want, err := os.ReadFile(golden)
	if err != nil {
		t.Fatalf("read golden: %v — run with -update to regenerate", err)
	}
	if string(got) != string(want) {
		t.Errorf("chart output differs from golden file.\nGot:\n%s\n\nWant:\n%s", got, want)
	}
}

// assertChartAnchors 校验命理关键字段（2026-06-28 12:00 CST 时家）：
// 日干落坎、时干落坤；坤土克坎水；空亡（子丑寅）与马星（未）命中日/时干宫。
func assertChartAnchors(t *testing.T, chart Chart) {
	t.Helper()
	if chart.RiGanPalace != GongKan {
		t.Errorf("ri_gan_gong = %d, want 坎(1)", chart.RiGanPalace)
	}
	if chart.ShiGanPalace != GongKun {
		t.Errorf("shi_gan_gong = %d, want 坤(2)", chart.ShiGanPalace)
	}
	if chart.RiShiShengKe != "时干(2宫)克日干(1宫)" {
		t.Errorf("ri_shi_sheng_ke = %q, want 时干(2宫)克日干(1宫)", chart.RiShiShengKe)
	}
	if !chart.KongWangAffected {
		t.Error("kong_wang_affected = false, want true")
	}
	if !chart.MaXingAffected {
		t.Error("ma_xing_affected = false, want true")
	}
	if chart.Pan.Jushu <= 0 {
		t.Errorf("jushu = %d, want > 0", chart.Pan.Jushu)
	}
	if len(chart.Patterns) == 0 {
		t.Error("patterns empty")
	}
}
