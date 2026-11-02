package bazi

import (
	"testing"

	"liki-engine/internal/engine/ganzhi"
)

// natalCoreFromChart is a test-only adapter for tests that still construct a
// partial chart literal. Production code constructs NatalCore directly.
func natalCoreFromChart(chart Chart) NatalCore {
	return NatalCore{
		Pillars: [4]ganzhi.Zhu{
			chart.Nian.Zhu, chart.Yue.Zhu, chart.Ri.Zhu, chart.Shi.Zhu,
		},
		NaYin:     chart.NaYinArray(),
		DaYun:     chart.DaYun,
		Gender:    chart.Gender,
		BirthYear: chart.BirthYear,
	}
}

func computeYongshenEvidence(core NatalCore) (FuYiResult, TiaoHouResult, GeJuResult) {
	return computeYongShenEvidenceFromAnalysis(core, AnalyzePillars(core.ToBazi(), core.Gender))
}

func atomicCoreFromStrings(t *testing.T, pillars []string) NatalCore {
	t.Helper()
	if len(pillars) != 4 {
		t.Fatalf("pillars = %d, want 4", len(pillars))
	}
	var core NatalCore
	for index, pillar := range pillars {
		runes := []rune(pillar)
		gan, err := ganzhi.ParseGan(string(runes[0]))
		if err != nil {
			t.Fatal(err)
		}
		zhi, err := ganzhi.ParseZhi(string(runes[1]))
		if err != nil {
			t.Fatal(err)
		}
		core.Pillars[index] = ganzhi.Zhu{Gan: gan, Zhi: zhi}
		core.NaYin[index] = ganzhi.NayinLabel(gan, zhi)
	}
	core.Gender = ganzhi.Male
	return core
}

func assertStrings(t *testing.T, got, want []string) {
	t.Helper()
	if len(got) != len(want) {
		t.Fatalf("strings = %v, want %v", got, want)
	}
	for index := range want {
		if got[index] != want[index] {
			t.Fatalf("strings = %v, want %v", got, want)
		}
	}
}
