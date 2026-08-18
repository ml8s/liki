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
// 日干落震、时干落艮；震木克艮土；空亡（子丑→坎艮）命中时干艮8；马星（申→坤）未命中日/时干宫。
func assertChartAnchors(t *testing.T, chart Chart) {
	t.Helper()
	// 2026-06-28 午时：夏至中元（癸酉日）→ 阴遁3局。日干癸天盘落震、时干戊天盘落艮（用神落宫以天盘为核心）；
	// 时柱戊午（甲寅旬）空子丑（坎艮）→ 时干艮8空亡；午→寅午戌马在申（坤）。
	if chart.Pan.Jushu != 3 {
		t.Errorf("jushu = %d, want 3（夏至中元 阴遁3局）", chart.Pan.Jushu)
	}
	if !chart.Pan.YinDun {
		t.Error("yin_dun = false, want true（夏至后阴遁）")
	}
	if chart.RiGanPalace != GongZhen {
		t.Errorf("ri_gan_gong = %d, want 震(3)", chart.RiGanPalace)
	}
	if chart.ShiGanPalace != GongGen {
		t.Errorf("shi_gan_gong = %d, want 艮(8)", chart.ShiGanPalace)
	}
	if chart.RiShiShengKe != "日干(3宫)克时干(8宫)" {
		t.Errorf("ri_shi_sheng_ke = %q, want 日干(3宫)克时干(8宫)", chart.RiShiShengKe)
	}
	if !chart.KongWangAffected {
		t.Error("kong_wang_affected = false, want true（时干艮8在空亡坎艮）")
	}
	if chart.MaXingAffected {
		t.Error("ma_xing_affected = true, want false（马星坤不在日震3时艮8）")
	}
	if len(chart.Patterns) == 0 {
		t.Error("patterns empty")
	}
}
