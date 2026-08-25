package qiming

import "testing"

func TestSancaiHarmonious_AllGenerating(t *testing.T) {
	// 姓=2 (this is NOT 笔画, but the surname stroke value used in computation)
	// s1=3, s2=5
	// 天格=2+1=3→火, 人格=2+3=5→土, 地格=3+5=8→金
	// 火→土(生) ✅, 土→金(生) ✅
	if !SancaiHarmonious(singleStrokes(2), 3, 5) {
		t.Error("expected true: 火→土→金 all generating")
	}
}

func TestSancaiHarmonious_Overcoming(t *testing.T) {
	// 姓=7, s1=3, s2=5 → 天格=8金, 人格=10水, 地格=8金 → 金水金
	// 查表 金水金 = da_ji → 应通过
	if !SancaiHarmonious(singleStrokes(7), 3, 5) {
		t.Error("expected true: 金水金 is da_ji in table")
	}
}

func TestSancaiHarmonious_SingleName(t *testing.T) {
	// 姓=2, s1=3, s2=0 → 天格=3火, 人格=5土, 地格=4火 → 火土火
	// 查表 火土火 = da_ji → 应通过
	if !SancaiHarmonious(singleStrokes(2), 3, 0) {
		t.Error("expected true: 火土火 is da_ji in table")
	}
}

func TestSancaiHarmonious_SingleHarmonious(t *testing.T) {
	// 姓=2, s1=5, s2=0 → 天格=3火, 人格=7金, 地格=6土 → 火金土
	// 查表 火金土 = da_ji → 应通过
	if !SancaiHarmonious(singleStrokes(2), 5, 0) {
		t.Error("expected true: 火金土 is da_ji in table")
	}
}

func TestSancaiHarmonious_TableLookup(t *testing.T) {
	// 查表语义：木木木（比和）在大吉表中应通过
	// 姓=0, s1=1, s2=1 → 天格=1木, 人格=1木, 地格=2木 → 木木木
	// 125表中 木木木 = da_ji → 应通过
	if !SancaiHarmonious(singleStrokes(0), 1, 1) {
		t.Error("expected true: 木木木 is da_ji in table")
	}
	// 木土水 = xiong → 应拒绝
	// 天格=1木, 人格=3火?? 需要构造 木土水
	// 天格=1木 → 姓=0；人格=6土 → s1=6；地格=10水 → s2=4
	if SancaiHarmonious(singleStrokes(0), 6, 4) {
		t.Error("expected false: 木土水 is xiong in table")
	}
}

func TestFilterSancai(t *testing.T) {
	// 姓=2 → pair (3,5) harmonious, pair (7,3) not
	// 天格=3→火, 人格=2+7=9→水, 地格=7+3=10→水
	// 火→水 ❌ (火生土)
	pairs := []StrokePair{
		{S1: 3, S2: 5},  // harmonious
		{S1: 7, S2: 3},  // not harmonious
	}
	result := FilterSancai(singleStrokes(2), pairs)
	if len(result) != 1 {
		t.Errorf("expected 1 harmonious pair, got %d", len(result))
	}
}

func TestFilterSancai_Empty(t *testing.T) {
	if r := FilterSancai(singleStrokes(2), nil); r != nil {
		t.Error("expected nil for nil input")
	}
	if r := FilterSancai(singleStrokes(2), []StrokePair{}); len(r) != 0 {
		t.Error("expected empty for empty input")
	}
}
