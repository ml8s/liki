package bazi

import (
	"reflect"
	"testing"
	"time"

	"liki-engine/internal/engine/ganzhi"
	"liki-engine/internal/engine/tianwen"
)

func strengthTestChart() Chart {
	at, err := time.Parse(time.RFC3339Nano, "1990-05-20T04:00:00Z")
	if err != nil {
		panic(err)
	}
	loc, err := time.LoadLocation("Asia/Shanghai")
	if err != nil {
		panic(err)
	}
	solar := tianwen.ApparentSolarCivil(at, loc, 116.407526)
	pillars := tianwen.ComputeBazi(time.Time(at), time.Time(solar))
	ny := [4]string{}
	for i, pillar := range pillars.Slice() {
		ny[i] = ganzhi.NayinLabel(pillar.Gan, pillar.Zhi)
	}
	makeInfo := func(i int) zhuInfo {
		return zhuInfo{Zhu: pillars.Slice()[i], NaYin: ny[i]}
	}
	return Chart{
		Nian:   makeInfo(0),
		Yue:    makeInfo(1),
		Ri:     makeInfo(2),
		Shi:    makeInfo(3),
		Gender: ganzhi.Male,
	}
}

func TestWuxingCountsAtMatchesStructuralProjection(t *testing.T) {
	chart := strengthTestChart()
	analysis := AnalyzePillars(chart.ToBazi(), chart.Gender)
	want := computeElementCount(chart.ToBazi(), analysis.HiddenStems)
	got := WuxingCountsAt(chart, analysis.HiddenStems)
	if !reflect.DeepEqual(map[ganzhi.Wuxing]int(got), want) {
		t.Fatalf("counts = %+v, want %+v", got, want)
	}
}

func TestAnalyzeDayMasterStrengthHasCompleteEvidence(t *testing.T) {
	chart := strengthTestChart()
	analysis := AnalyzePillars(chart.ToBazi(), chart.Gender)
	strength := AnalyzeDayMasterStrength(chart, analysis.HiddenStems)

	if strength.RootType == "" || strength.Season == "" || strength.Label == "" {
		t.Fatalf("incomplete strength evidence: %+v", strength)
	}
	if strength.YinBiCount < 0 || strength.YinBiCount > 3 {
		t.Fatalf("yin-bi count out of range: %d", strength.YinBiCount)
	}
}
