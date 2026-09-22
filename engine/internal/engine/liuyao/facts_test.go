package liuyao

import (
	"testing"
	"time"

	"liki-engine/internal/engine/ganzhi"
	"liki-engine/internal/engine/tianwen"
)

func factsChart(t *testing.T, yaos [6]int) Chart {
	t.Helper()
	st := tianwen.GregorianToSolar(
		time.Date(2026, 9, 8, 12, 0, 0, 0, time.FixedZone("CST", 8*3600)),
		116.4, 8,
	)
	return ComputeChart(st, YongShiYao, yaos)
}

func TestFacts_AllMovingLinesHaveTransformations(t *testing.T) {
	chart := factsChart(t, [6]int{9, 6, 9, 7, 8, 7})
	if len(chart.MovingTransformations) != 3 {
		t.Fatalf("transformations=%d, want 3", len(chart.MovingTransformations))
	}
	for _, item := range chart.MovingTransformations {
		if item.Position == 0 || item.FromBranch == "" || item.ToBranch == "" {
			t.Fatalf("incomplete transformation: %+v", item)
		}
	}
}

func TestFacts_YongShenCandidatesMarkSelection(t *testing.T) {
	chart := factsChart(t, [6]int{7, 7, 7, 7, 7, 7})
	selected := 0
	for _, candidate := range chart.YongShenCandidates {
		if candidate.Selected {
			selected++
		}
	}
	if selected != 1 {
		t.Fatalf("selected candidates=%d, want 1 (%+v)", selected, chart.YongShenCandidates)
	}
	if chart.ForceChain == nil || chart.ForceChain.Position == 0 || len(chart.ForceChain.Entries) < 3 {
		t.Fatalf("force chain incomplete: %+v", chart.ForceChain)
	}
}

func TestFacts_DayClashClassifiesStrongAndWeak(t *testing.T) {
	chart := factsChart(t, [6]int{7, 7, 7, 7, 7, 7})
	kinds := map[string]bool{}
	for _, fact := range chart.DayClashFacts {
		kinds[fact.Kind] = true
		if fact.Condition == "" || len(fact.Basis) == 0 {
			t.Fatalf("incomplete day clash fact: %+v", fact)
		}
	}
	// Exercise the full static/moving/void classification directly so the test
	// is independent of the limited branches available in a hexagram fixture.
	synthetic := func(wang ganzhi.WangShuai, moving, xunKong bool) Chart {
		lines := [6]Line{}
		lines[1] = Line{Position: 2, Zhi: ganzhi.ZhiWu, Wuxing: ganzhi.WxHuo, Type: ShaoYang, DongSelf: moving, XunKong: xunKong}
		return Chart{
			Lines:     lines,
			WangShuai: [6]ganzhi.WangShuai{ganzhi.WSWang, wang, ganzhi.WSWang, ganzhi.WSWang, ganzhi.WSWang, ganzhi.WSWang},
			RiZhi:     ganzhi.ZhiZi,
		}
	}
	syntheticChart := synthetic(ganzhi.WSWang, false, false)
	facts := computeDayClashFacts(&syntheticChart)
	if len(facts) != 1 || facts[0].Kind != "暗动" {
		t.Fatalf("strong fixture = %+v, want 暗动", facts)
	}
	syntheticChart = synthetic(ganzhi.WSSi, false, false)
	facts = computeDayClashFacts(&syntheticChart)
	if len(facts) != 1 || facts[0].Kind != "日破" {
		t.Fatalf("weak fixture = %+v, want 日破", facts)
	}
	syntheticChart = synthetic(ganzhi.WSWang, true, false)
	facts = computeDayClashFacts(&syntheticChart)
	if len(facts) != 1 || facts[0].Kind != "冲散" {
		t.Fatalf("moving fixture = %+v, want 冲散", facts)
	}
	syntheticChart = synthetic(ganzhi.WSWang, false, true)
	facts = computeDayClashFacts(&syntheticChart)
	if len(facts) != 1 || facts[0].Kind != "冲空" {
		t.Fatalf("void fixture = %+v, want 冲空", facts)
	}
}

func TestFacts_SanHeRequiresThreeDistinctBranchesForComplete(t *testing.T) {
	for i := 0; i < 4096; i++ {
		yaos := [6]int{}
		value := i
		for j := 5; j >= 0; j-- {
			yaos[j] = 6 + value%4
			value /= 4
		}
		chart := factsChart(t, yaos)
		for _, candidate := range chart.SanHeCandidates {
			if candidate.Complete && !candidate.LineComplete &&
				func() bool {
					distinct := map[string]bool{}
					for _, branch := range candidate.Branches {
						distinct[branch] = true
					}
					if candidate.DayPresent {
						distinct[ganzhi.ZhiName(chart.RiZhi)] = true
					}
					if candidate.MonthPresent {
						distinct[ganzhi.ZhiName(chart.YueZhi)] = true
					}
					return len(distinct) < 3
				}() {
				t.Fatalf("yaos=%v invalid complete candidate: %+v", yaos, candidate)
			}
			if candidate.ConclusionScope == "" {
				t.Fatalf("yaos=%v candidate lacks conclusion scope", yaos)
			}
		}
	}
}
