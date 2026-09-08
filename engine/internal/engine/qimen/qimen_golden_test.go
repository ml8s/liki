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
	chart := ComputeChart(st)

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
		t.Fatalf("read golden: %v — run with UPDATE_GOLDEN=1 to regenerate", err)
	}
	if string(got) != string(want) {
		t.Errorf("chart output differs from golden file.\nGot:\n%s\n\nWant:\n%s", got, want)
	}
}

// assertChartAnchors 校验命理关键字段（2026-06-28 12:00 CST 时家）：
// 日干落震、时干落离；震木生离火；空亡（子丑→坎艮）未命中日/时干宫；马星（申→坤）未命中日/时干宫。
func assertChartAnchors(t *testing.T, chart Chart) {
	t.Helper()
	// 2026-06-28 午时：夏至中元（癸酉日）→ 阴遁3局。日干癸天盘落震、时干戊天盘落兑（用神落宫以天盘为核心）；
	// 时柱戊午（甲寅旬）空子丑（坎艮）；午→寅午戌马在申（坤）。
	if chart.Pan.Jushu != 3 {
		t.Errorf("jushu = %d, want 3（夏至中元 阴遁3局）", chart.Pan.Jushu)
	}
	if !chart.Pan.YinDun {
		t.Error("yin_dun = false, want true（夏至后阴遁）")
	}
	if chart.RiGanPalace != GongZhen {
		t.Errorf("ri_gan_gong = %d, want 震(3)", chart.RiGanPalace)
	}
	if chart.ShiGanPalace != GongDui {
		t.Errorf("shi_gan_gong = %d, want 兑(7)", chart.ShiGanPalace)
	}
	if chart.RiShiRelation.Relation != "shi_controls_ri" {
		t.Errorf("ri_shi_relation = %+v, want shi_controls_ri", chart.RiShiRelation)
	}
	if chart.Method.Scope != "hour" || chart.Method.School != "zhuanpan" || chart.Method.DingjuMethod != "chaibu" {
		t.Errorf("method = %+v, want 时家/转盘/拆补", chart.Method)
	}
	if chart.Method.JieQi != "夏至" || chart.Method.Yuan != "中元" {
		t.Errorf("solar-term method = %s/%s, want 夏至/中元", chart.Method.JieQi, chart.Method.Yuan)
	}
	if len(chart.KongWangAffected) != 0 {
		t.Errorf("kong_wang_affected = %+v, want empty", chart.KongWangAffected)
	}
	if len(chart.MaXingAffected) != 0 {
		t.Errorf("ma_xing_affected = %+v, want empty", chart.MaXingAffected)
	}
	if len(chart.Patterns) == 0 {
		t.Error("patterns empty")
	}
}
