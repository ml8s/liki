package xuankong

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
	st := tianwen.ApparentSolarFromCivil(
		time.Date(2026, 6, 28, 12, 0, 0, 0, time.FixedZone("CST", 8*3600)),
		116.4, 8,
	)
	chart, err := ComputeChartAtStandard(st, 1, 13)
	if err != nil {
		t.Fatalf("compute chart: %v", err)
	}
	if chart.Yun.YunNumber != 9 || chart.Yun.YunName != "九运" {
		t.Fatalf("yun = %+v, want 2024-2043 九运", chart.Yun)
	}
	if chart.FourSituation.Name != "双星会坐" {
		t.Fatalf("four situation = %s, want 双星会坐", chart.FourSituation.Name)
	}
	if got := chart.Palaces[4]; got.PeriodStar.Number != 9 || got.MountainStar.Number != 5 || got.FacingStar.Number != 4 {
		t.Fatalf("center stars = %d/%d/%d, want 9/5/4", got.PeriodStar.Number, got.MountainStar.Number, got.FacingStar.Number)
	}

	got, marshalErr := json.MarshalIndent(chart, "", "  ")
	if err != nil {
		t.Fatalf("marshal: %v", marshalErr)
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
