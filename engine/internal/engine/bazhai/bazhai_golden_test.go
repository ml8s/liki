package bazhai

import (
	"encoding/json"
	"os"
	"path/filepath"
	"testing"
	"time"

	"liki-engine/internal/engine/ganzhi"
	"liki-engine/internal/engine/tianwen"
)

var updateGolden = os.Getenv("UPDATE_GOLDEN") == "1"

func TestGoldenComputeChart(t *testing.T) {
	st := tianwen.GregorianToSolar(
		time.Date(1984, 2, 15, 8, 0, 0, 0, time.FixedZone("CST", 8*3600)),
		116.4, 8,
	)
	chart := ComputeChart(st, ganzhi.Male)

	// 命理锚点断言（独立于 golden 文件——UPDATE_GOLDEN=1 时同样执行，
	// 防止错误输出被锁进 golden 后测试自证）。
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

// assertChartAnchors 校验命理关键字段（1984-02-15 08:00 男）：
// 男命 1984 → 艮卦/西四命；四吉方之首生气=西南；流年紫白 1984 七赤入中。
func assertChartAnchors(t *testing.T, chart Chart) {
	t.Helper()
	// 1984 男：命卦公式（《八宅明镜》2000 前）男 (100-84)%9=7 → 兑，西四命。
	if chart.MingGua.Gua.Name != "兑" {
		t.Errorf("ming_gua = %s, want 兑（(100-84)%%9=7）", chart.MingGua.Gua.Name)
	}
	if chart.MingGua.Group != "西四命" {
		t.Errorf("group = %s, want 西四命", chart.MingGua.Group)
	}
	// 兑命（西四）：生气在西北。
	if len(chart.BaZhaiDirs.ShengQi) == 0 || chart.BaZhaiDirs.ShengQi[0] != "西北" {
		t.Errorf("sheng_qi = %v, want 西北 居首", chart.BaZhaiDirs.ShengQi)
	}
	if chart.YearStars.Year != 1984 {
		t.Errorf("liu_nian_xing.year = %d, want 1984", chart.YearStars.Year)
	}
	if chart.YearStars.RuZhong != "七赤破军" {
		t.Errorf("liu_nian_xing.ru_zhong = %s, want 七赤破军（下元甲子）", chart.YearStars.RuZhong)
	}
	if len(chart.ZhuBagua) != 4 {
		t.Errorf("zhu_bagua len = %d, want 4", len(chart.ZhuBagua))
	}
}
