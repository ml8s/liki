package liuyao

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
	// Use fixed [6]int for deterministic output.
	st := tianwen.GregorianToSolar(
		time.Date(1984, 2, 15, 8, 0, 0, 0, time.FixedZone("CST", 8*3600)),
		116.4, 8,
	)
	chart := ComputeChart(st, YongGuanGui, [6]int{7, 7, 7, 7, 7, 7})

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

// assertChartAnchors 校验命理关键字段（1984-02-15 08:00 CST，寅月，乾为天静卦）：
// 全少阳无动爻；九五甲申 被 寅月 冲 → 月破。
func assertChartAnchors(t *testing.T, chart Chart) {
	t.Helper()
	if chart.Name != "乾为天" {
		t.Errorf("name = %s, want 乾为天", chart.Name)
	}
	if len(chart.DongYao) != 0 {
		t.Errorf("dong_yao = %v, want empty（全少阳静卦）", chart.DongYao)
	}
	// 九五甲申：寅月冲申 → 月破；静卦无发动/生克
	if !chart.Lines[4].YuePo {
		t.Errorf("line5 甲申在寅月应月破(yue_po=true)")
	}
	if chart.Lines[4].DongSelf || chart.Lines[4].DongSheng || chart.Lines[4].DongKe {
		t.Errorf("line5 静卦不应有发动/生克标记")
	}
	if chart.YongShen.Name != "官鬼" {
		t.Errorf("yong_shen = %s, want 官鬼（yong_shen 参数）", chart.YongShen.Name)
	}
}
