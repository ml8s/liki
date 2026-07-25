package ziwei

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

func TestGoldenComputeDaXian(t *testing.T) {
	st := tianwen.GregorianToSolar(
		time.Date(1984, 2, 15, 8, 0, 0, 0, time.FixedZone("CST", 8*3600)),
		116.4, 8,
	)
	chart := ComputeChart(st, ganzhi.Male)
	daxian := ComputeDaXian(chart)

	got, err := json.MarshalIndent(daxian, "", "  ")
	if err != nil {
		t.Fatalf("marshal: %v", err)
	}

	golden := filepath.Join("testdata", "daxian_golden.json")
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
		t.Errorf("daxian output differs from golden file.\nGot:\n%s\n\nWant:\n%s", got, want)
	}
}

func TestGoldenComputeLiuNian(t *testing.T) {
	st := tianwen.GregorianToSolar(
		time.Date(1984, 2, 15, 8, 0, 0, 0, time.FixedZone("CST", 8*3600)),
		116.4, 8,
	)
	chart := ComputeChart(st, ganzhi.Male)
	ln := ComputeLiuNian(chart, 2026)

	got, err := json.MarshalIndent(ln, "", "  ")
	if err != nil {
		t.Fatalf("marshal: %v", err)
	}

	golden := filepath.Join("testdata", "liunian_golden.json")
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
		t.Errorf("liunian output differs from golden file.\nGot:\n%s\n\nWant:\n%s", got, want)
	}
}
