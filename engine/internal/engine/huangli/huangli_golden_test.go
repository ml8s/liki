package huangli

import (
	"encoding/json"
	"os"
	"path/filepath"
	"testing"
	"time"

	"liki-engine/internal/engine/tianwen"
)

var updateGolden = os.Getenv("UPDATE_GOLDEN") == "1"

func TestGoldenComputeBondDay(t *testing.T) {
	st := tianwen.GregorianToSolar(
		time.Date(2026, 6, 28, 12, 0, 0, 0, time.FixedZone("CST", 8*3600)),
		116.4, 8,
	)
	bd, err := ComputeBondDay(st, "嫁娶", "2026-06-28")
	if err != nil {
		t.Fatalf("ComputeBondDay: %v", err)
	}

	got, err := json.MarshalIndent(bd, "", "  ")
	if err != nil {
		t.Fatalf("marshal: %v", err)
	}

	golden := filepath.Join("testdata", "bond_day_golden.json")
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
		t.Errorf("bond day output differs from golden file.\nGot:\n%s\n\nWant:\n%s", got, want)
	}
}

func TestGoldenComputeBondMonth(t *testing.T) {
	st := tianwen.GregorianToSolar(
		time.Date(2026, 6, 28, 12, 0, 0, 0, time.FixedZone("CST", 8*3600)),
		116.4, 8,
	)
	bm, err := ComputeBondMonth(st, "入宅", "2026-06")
	if err != nil {
		t.Fatalf("ComputeBondMonth: %v", err)
	}

	got, err := json.MarshalIndent(bm, "", "  ")
	if err != nil {
		t.Fatalf("marshal: %v", err)
	}

	golden := filepath.Join("testdata", "bond_month_golden.json")
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
		t.Errorf("bond month output differs from golden file.\nGot:\n%s\n\nWant:\n%s", got, want)
	}
}
