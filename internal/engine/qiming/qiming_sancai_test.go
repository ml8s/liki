package qiming

import "testing"

func TestStrokeToWuxing(t *testing.T) {
	tests := []struct {
		n    int
		want string
	}{
		{1, "木"}, {2, "木"}, {3, "火"}, {4, "火"},
		{5, "土"}, {6, "土"}, {7, "金"}, {8, "金"},
		{9, "水"}, {10, "水"}, {11, "木"}, {81, "木"},
	}
	for _, tt := range tests {
		got := strokeToWuxing(tt.n).String()
		if got != tt.want {
			t.Errorf("strokeToWuxing(%d) = %q, want %q", tt.n, got, tt.want)
		}
	}
}

func TestSancaiHarmonious_AllGenerating(t *testing.T) {
	// 姓=2 (this is NOT 笔画, but the surname stroke value used in computation)
	// s1=3, s2=5
	// 天格=2+1=3→火, 人格=2+3=5→土, 地格=3+5=8→金
	// 火→土(生) ✅, 土→金(生) ✅
	if !SancaiHarmonious(2, 3, 5) {
		t.Error("expected true: 火→土→金 all generating")
	}
}

func TestSancaiHarmonious_Overcoming(t *testing.T) {
	// 姓=7, s1=3, s2=5
	// 天格=8→金, 人格=10→水, 地格=8→金
	// 金→水(生) ✅, 水→金 ❌ (水→木, not 金)
	if SancaiHarmonious(7, 3, 5) {
		t.Error("expected false: 水→金 is not generating")
	}
}

func TestSancaiHarmonious_SingleName(t *testing.T) {
	// 姓=2, s1=3, s2=0 (单名)
	// 天格=3→火, 人格=5→土, 地格=3+1=4→火
	// 火→土(生) ✅, 土→火 ❌ (土→金)
	if SancaiHarmonious(2, 3, 0) {
		t.Error("expected false: 土→火 not generating")
	}
}

func TestSancaiHarmonious_SingleHarmonious(t *testing.T) {
	// 姓=2, s1=5, s2=0
	// 天格=3→火, 人格=7→金, 地格=5+1=6→土
	// 火→金 ❌
	if SancaiHarmonious(2, 5, 0) {
		t.Error("expected false: 火→金 not generating")
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
	result := FilterSancai(2, pairs)
	if len(result) != 1 {
		t.Errorf("expected 1 harmonious pair, got %d", len(result))
	}
}

func TestFilterSancai_Empty(t *testing.T) {
	if r := FilterSancai(2, nil); r != nil {
		t.Error("expected nil for nil input")
	}
	if r := FilterSancai(2, []StrokePair{}); len(r) != 0 {
		t.Error("expected empty for empty input")
	}
}
