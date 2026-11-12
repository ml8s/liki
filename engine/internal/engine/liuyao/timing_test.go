package liuyao

import (
	"strings"
	"testing"
	"time"

	"liki-engine/internal/engine/ganzhi"
	"liki-engine/internal/engine/tianwen"
)

func fixedChart(t *testing.T, yong YongShen, yaos [6]int) Chart {
	t.Helper()
	stgt := time.Time(time.Date(1984, 2, 15, 8, 0, 0, 0, time.FixedZone("CST", 8*3600)))
	st := tianwen.ApparentSolarFromCivil(
		time.Date(1984, 2, 15, 8, 0, 0, 0, time.FixedZone("CST", 8*3600)),
		116.4, 8,
	)
	return ComputeChart(stgt, st, yong, yaos)
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

func staticHarmonyChart(day, month ganzhi.Zhi, movingPosition int, movingBranch ganzhi.Zhi) Chart {
	chart := Chart{
		RiZhi:  day,
		YueZhi: month,
	}
	chart.Lines[0] = Line{
		Position: 1,
		Type:     ShaoYin,
		Zhi:      ganzhi.ZhiZi,
		ShiYing:  "世",
	}
	for position := range chart.Lines {
		chart.Lines[position].Position = position + 1
	}
	if movingPosition != 0 {
		chart.Lines[movingPosition-1].Zhi = movingBranch
		chart.DongYao = []int{movingPosition}
	}
	chart.YongShen = YongShenResult{
		Name:     "世爻",
		Position: 1,
	}
	return chart
}

func findTimingCandidate(candidates []TimingCandidate, mechanism string) (TimingCandidate, bool) {
	for _, candidate := range candidates {
		if candidate.Mechanism == mechanism {
			return candidate, true
		}
	}
	return TimingCandidate{}, false
}

func TestTimingCandidates_StaticHarmonyRequiresSource(t *testing.T) {
	tests := []struct {
		name          string
		day           ganzhi.Zhi
		month         ganzhi.Zhi
		movingPos     int
		movingBranch  ganzhi.Zhi
		wantSource    string
		unwantSources []string
	}{
		{
			name:          "without day month or moving source",
			day:           ganzhi.ZhiYin,
			month:         ganzhi.ZhiMao,
			unwantSources: []string{"day_branch", "month_branch", "moving_yao.position="},
		},
		{
			name:       "from day branch",
			day:        ganzhi.ZhiChou,
			month:      ganzhi.ZhiYin,
			wantSource: "day_branch",
		},
		{
			name:       "from month branch",
			day:        ganzhi.ZhiYin,
			month:      ganzhi.ZhiChou,
			wantSource: "month_branch",
		},
		{
			name:         "from moving yao branch",
			day:          ganzhi.ZhiYin,
			month:        ganzhi.ZhiMao,
			movingPos:    2,
			movingBranch: ganzhi.ZhiChou,
			wantSource:   "moving_yao.position=2",
		},
	}
	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			chart := staticHarmonyChart(tc.day, tc.month, tc.movingPos, tc.movingBranch)
			candidates := computeTimingCandidates(&chart, YongShiYao)
			got, found := findTimingCandidate(candidates, "静爻逢合")
			if tc.wantSource == "" {
				if found {
					t.Fatalf("unexpected static harmony: %+v", got)
				}
				return
			}
			if !found {
				t.Fatalf("missing static harmony in %+v", candidates)
			}
			if got.ID != "static-harmony-1-丑" || got.TriggerBranch != "丑" {
				t.Fatalf("candidate identity = id %q branch %q", got.ID, got.TriggerBranch)
			}
			if !containsString(got.Basis, tc.wantSource) {
				t.Fatalf("basis missing %q: %+v", tc.wantSource, got)
			}
			for _, source := range tc.unwantSources {
				if hasPrefixedString(got.Basis, source) {
					t.Fatalf("basis has unwanted source %q: %+v", source, got)
				}
			}
		})
	}
}

func TestTimingCandidates_MovingYaoDoesNotUseStaticHarmony(t *testing.T) {
	chart := staticHarmonyChart(ganzhi.ZhiYin, ganzhi.ZhiMao, 0, 0)
	chart.Lines[0].Type = LaoYang
	candidates := computeTimingCandidates(&chart, YongShiYao)
	if _, found := findTimingCandidate(candidates, "静爻逢合"); found {
		t.Fatalf("static harmony used for moving yao: %+v", candidates)
	}
	if _, found := findTimingCandidate(candidates, "动爻逢值"); !found {
		t.Fatalf("moving value missing: %+v", candidates)
	}
}

func TestTimingCandidates_VoidOrBreakDoesNotDuplicateStaticHarmony(t *testing.T) {
	for _, state := range []func(*YongShenResult){
		func(y *YongShenResult) { y.XunKong = true },
		func(y *YongShenResult) { y.YuePo = true },
	} {
		chart := staticHarmonyChart(ganzhi.ZhiChou, ganzhi.ZhiYin, 0, 0)
		state(&chart.YongShen)
		candidates := computeTimingCandidates(&chart, YongShiYao)
		if _, found := findTimingCandidate(candidates, "静爻逢合"); found {
			t.Fatalf("unexpected static harmony duplicate for state %+v: %+v", chart.YongShen, candidates)
		}
	}
}

func containsString(values []string, want string) bool {
	for _, value := range values {
		if value == want {
			return true
		}
	}
	return false
}

func hasPrefixedString(values []string, prefix string) bool {
	for _, value := range values {
		if strings.HasPrefix(value, prefix) {
			return true
		}
	}
	return false
}
