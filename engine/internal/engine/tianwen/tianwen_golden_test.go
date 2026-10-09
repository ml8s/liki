package tianwen

import (
	"encoding/json"
	"os"
	"path/filepath"
	"testing"
	"time"
)

var updateGolden = os.Getenv("UPDATE_GOLDEN") == "1"

func TestGoldenGregorianToSolar(t *testing.T) {
	tests := []struct {
		name string
		t    time.Time
		lon  float64
		tz   float64
	}{
		{"beijing_noon", time.Date(2026, 6, 28, 12, 0, 0, 0, time.FixedZone("CST", 8*3600)), 116.4, 8},
		{"tokyo_morning", time.Date(2024, 1, 1, 9, 0, 0, 0, time.FixedZone("JST", 9*3600)), 139.7, 9},
		{"london_midnight", time.Date(2023, 12, 25, 0, 0, 0, 0, time.UTC), -0.1, 0},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			st := GregorianToSolar(tt.t, tt.lon, tt.tz)
			got, err := json.MarshalIndent(st, "", "  ")
			if err != nil {
				t.Fatalf("marshal: %v", err)
			}

			golden := filepath.Join("testdata", "solar_"+tt.name+"_golden.json")
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
				t.Errorf("solar %s output differs from golden file.\nGot:\n%s\n\nWant:\n%s", tt.name, got, want)
			}
		})
	}
}
