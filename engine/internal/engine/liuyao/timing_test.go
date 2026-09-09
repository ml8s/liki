package liuyao

import (
	"strings"
	"testing"
	"time"

	"liki-engine/internal/engine/tianwen"
)

func fixedChart(t *testing.T, yong YongShen, yaos [6]int) Chart {
	t.Helper()
	st := tianwen.GregorianToSolar(
		time.Date(1984, 2, 15, 8, 0, 0, 0, time.FixedZone("CST", 8*3600)),
		116.4, 8,
	)
	return ComputeChart(st, yong, yaos)
}

func TestTimingCandidates_MovingLine(t *testing.T) {
	chart := fixedChart(t, YongQiCai, [6]int{7, 9, 7, 7, 7, 7})
	if chart.YongShen.Position == 0 {
		t.Skip("fixture has no wealth line")
	}
	found := false
	for _, candidate := range chart.TimingCandidates {
		if candidate.Mechanism == "动爻逢值" {
			found = true
			if candidate.Position == 0 || candidate.TriggerBranch == "" {
				t.Fatalf("incomplete moving candidate: %+v", candidate)
			}
		}
	}
	if !found {
		t.Fatalf("expected moving-value candidate, got %+v", chart.TimingCandidates)
	}
}

func TestTimingCandidates_XunKong(t *testing.T) {
	chart := fixedChart(t, YongXiongDi, [6]int{7, 7, 7, 7, 7, 7})
	if !chart.YongShen.XunKong {
		t.Fatalf("fixture expected xun-kong yong, got %+v", chart.YongShen)
	}
	mechanisms := map[string]bool{}
	for _, candidate := range chart.TimingCandidates {
		mechanisms[candidate.Mechanism] = true
	}
	for _, want := range []string{"旬空填实", "冲空"} {
		if !mechanisms[want] {
			t.Fatalf("missing %q in %+v", want, chart.TimingCandidates)
		}
	}
}

func TestTimingCandidates_StaticClashRequiresActualDayClash(t *testing.T) {
	chart := fixedChart(t, YongGuanGui, [6]int{7, 7, 7, 7, 7, 7})
	for _, candidate := range chart.TimingCandidates {
		if candidate.Mechanism == "静爻逢冲" {
			t.Fatalf("unexpected static-clash candidate without day clash: %+v", candidate)
		}
		if strings.Contains(string(candidate.ID), "static-clash") {
			t.Fatalf("unexpected static-clash id: %+v", candidate)
		}
	}
}
