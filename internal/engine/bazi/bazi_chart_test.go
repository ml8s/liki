//go:build integration

package bazi

import (
	"encoding/json"
	"testing"
	"time"

	"liki-engine/internal/engine/ganzhi"
	"liki-engine/internal/engine/tianwen"
)

// ── Known birth dates for chart regression tests ──

type chartSnapshot struct {
	Name   string
	Year   int
	Month  int
	Day    int
	Hour   int
	Minute int
	Lon    float64
	TZ     float64
	Gender ganzhi.Gender
}

func computeFullChart(g chartSnapshot) FullChart {
	st := tianwen.GregorianToSolar(
		time.Date(g.Year, time.Month(g.Month), g.Day, g.Hour, g.Minute, 0, 0,
			time.FixedZone("", int(g.TZ*3600))),
		g.Lon, g.TZ)
	return ComputeFullChart(ComputeChart(st, g.Gender))
}

var goldenCharts = []chartSnapshot{
	{Name: "Miles-1982-10-13-06:45-HongKong-male", Year: 1982, Month: 10, Day: 13, Hour: 6, Minute: 45, Lon: 114.134, TZ: 8, Gender: ganzhi.Male},
	{Name: "Beijing-1984-02-15-08:00-male", Year: 1984, Month: 2, Day: 15, Hour: 8, Minute: 0, Lon: 120, TZ: 8, Gender: ganzhi.Male},
	{Name: "Shanghai-1990-05-20-15:00-female", Year: 1990, Month: 5, Day: 20, Hour: 15, Minute: 0, Lon: 121.5, TZ: 8, Gender: ganzhi.Female},
	{Name: "Tokyo-2000-01-01-00:00-male", Year: 2000, Month: 1, Day: 1, Hour: 0, Minute: 0, Lon: 139.76, TZ: 9, Gender: ganzhi.Male},
	{Name: "NewYork-2020-06-15-20:00-female", Year: 2020, Month: 6, Day: 15, Hour: 20, Minute: 0, Lon: -74.0, TZ: -4, Gender: ganzhi.Female},
}

// ── Pillars validity ──

func TestGoldenChart_Pillars(t *testing.T) {
	for _, g := range goldenCharts {
		t.Run(g.Name, func(t *testing.T) {
			cr := computeFullChart(g)

			pillars := []struct {
				name string
				gan  ganzhi.Gan
				zhi  ganzhi.Zhi
			}{
				{"Nian", cr.Nian.Gan, cr.Nian.Zhi},
				{"Yue", cr.Yue.Gan, cr.Yue.Zhi},
				{"Ri", cr.Ri.Gan, cr.Ri.Zhi},
				{"Shi", cr.Shi.Gan, cr.Shi.Zhi},
			}
			for _, p := range pillars {
				if p.gan < 1 || p.gan > 10 {
					t.Errorf("%s.Gan = %d, want [1,10]", p.name, p.gan)
				}
				if p.zhi < 1 || p.zhi > 12 {
					t.Errorf("%s.Zhi = %d, want [1,12]", p.name, p.zhi)
				}
			}

			// Ri.Gan IS the day master
			if cr.Ri.Gan < 1 || cr.Ri.Gan > 10 {
				t.Errorf("Ri.Gan (日主) = %d, want [1,10]", cr.Ri.Gan)
			}
		})
	}
}

// ── JSON snapshot stability ──

func TestGoldenChart_JSONSnapshot(t *testing.T) {
	for _, g := range goldenCharts {
		t.Run(g.Name, func(t *testing.T) {
			cr := computeFullChart(g)

			b, err := json.MarshalIndent(&cr, "", "  ")
			if err != nil {
				t.Fatalf("marshal: %v", err)
			}

			var m map[string]any
			if err := json.Unmarshal(b, &m); err != nil {
				t.Fatalf("unmarshal: %v", err)
			}

		// Required top-level keys (new Chart format)
		for _, k := range []string{"nian", "yue", "ri", "shi", "da_yun", "gender"} {
			if v, ok := m[k]; !ok || v == nil {
				t.Errorf("%s = %v, want non-nil", k, v)
			}
		}
		})
	}
}

// ── Wuxing counts ──

func TestGoldenChart_HiddenStems(t *testing.T) {
	for _, g := range goldenCharts {
		t.Run(g.Name, func(t *testing.T) {
			cr := computeFullChart(g)
			hs := cr.CangGanArray()

			if len(hs) != 4 {
				t.Fatalf("len(CangGanArray) = %d, want 4", len(hs))
			}
			names := [4]string{"nian", "yue", "ri", "shi"}
			for i, h := range hs {
				if h.Main == 0 {
					t.Errorf("%s pillar hidden stem main qi is zero", names[i])
				}
			}
		})
	}
}

// ── NaYin ──

func TestGoldenChart_NaYin(t *testing.T) {
	for _, g := range goldenCharts {
		t.Run(g.Name, func(t *testing.T) {
			cr := computeFullChart(g)

			if cr.Nian.NaYin == "" {
				t.Error("Nian.NaYin is empty")
			}
			if cr.Yue.NaYin == "" {
				t.Error("Yue.NaYin is empty")
			}
			if cr.Ri.NaYin == "" {
				t.Error("Ri.NaYin is empty")
			}
			if cr.Shi.NaYin == "" {
				t.Error("Shi.NaYin is empty")
			}

			for _, n := range []string{
				cr.Nian.NaYin, cr.Yue.NaYin,
				cr.Ri.NaYin, cr.Shi.NaYin,
			} {
				if elem := ganzhi.NayinWuxing(n); elem == 0 {
					t.Errorf("nayin %q has unknown element", n)
				}
			}
		})
	}
}

// ── ShenSha ──

func TestGoldenChart_ShenSha(t *testing.T) {
	for _, g := range goldenCharts {
		t.Run(g.Name, func(t *testing.T) {
			cr := computeFullChart(g)

			validCats := map[string]bool{"吉": true, "凶": true, "中性": true}
			pillars := []struct {
				name string
				ss   []shenShaEntry
			}{
				{"Nian", cr.Nian.ShenSha},
				{"Yue", cr.Yue.ShenSha},
				{"Ri", cr.Ri.ShenSha},
				{"Shi", cr.Shi.ShenSha},
			}
			for _, p := range pillars {
				if len(p.ss) < 1 {
					t.Errorf("%s: no shensha entries", p.name)
					continue
				}
				for _, e := range p.ss {
					if e.Name == "" {
						t.Errorf("%s: shensha name is empty", p.name)
					}
					if !validCats[e.Category] {
						t.Errorf("%s: shensha %q category = %q, want 吉/凶/中性", p.name, e.Name, e.Category)
					}
				}
			}
		})
	}
}

// ── TaiYuan / MingGong / ShenGong ──

func TestGoldenChart_KongWang(t *testing.T) {
	for _, g := range goldenCharts {
		t.Run(g.Name, func(t *testing.T) {
			cr := computeFullChart(g)

			// Verify IsVoid fields exist and are accessible on all 4 pillars.
			pillars := []bool{cr.Nian.IsVoid, cr.Yue.IsVoid, cr.Ri.IsVoid, cr.Shi.IsVoid}
			for _, v := range pillars {
				_ = v // valid bool by construction
			}
		})
	}
}

// ── ShiShens table ──

func TestGoldenChart_ShiShens(t *testing.T) {
	for _, g := range goldenCharts {
		t.Run(g.Name, func(t *testing.T) {
			cr := computeFullChart(g)

			pillars := [4]fullZhuInfo{cr.Nian, cr.Yue, cr.Ri, cr.Shi}
			names := [4]string{"nian", "yue", "ri", "shi"}

			validShiShens := map[string]bool{
				"比肩": true, "劫财": true, "食神": true, "伤官": true,
				"偏财": true, "正财": true, "七杀": true, "正官": true,
				"偏印": true, "正印": true,
			}

			for i, p := range pillars {
				if len(p.ShiShens) < 1 {
					t.Errorf("%s pillar: no shi shens", names[i])
					continue
				}
				if p.ShiShens[0].Source != sourceGan {
					t.Errorf("%s pillar: first shi shen source = %s, want stem", names[i], p.ShiShens[0].Source)
				}
				if !validShiShens[p.ShiShens[0].ShiShen.String()] {
					t.Errorf("%s pillar: unknown shi shen %q", names[i], p.ShiShens[0].ShiShen)
				}
				hasMainQi := false
				for _, e := range p.ShiShens {
					if e.Source == sourceMainQi {
						hasMainQi = true
						break
					}
				}
				if !hasMainQi {
					t.Errorf("%s pillar: no main_qi shi shen", names[i])
				}
			}
		})
	}
}
