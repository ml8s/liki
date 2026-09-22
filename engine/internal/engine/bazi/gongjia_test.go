package bazi

import (
	"testing"

	"liki-engine/internal/engine/ganzhi"
)

func TestGongJia_ClassicOnly(t *testing.T) {
	tests := []struct {
		name    string
		pillars []string
		want    []GongJia
	}{
		{
			name:    "san-he-first-and-tomb-growth-center",
			pillars: []string{"甲申", "乙丑", "丙辰", "丁卯"},
			want: []GongJia{{
				ZhuA: 0, ZhuB: 2, Type: "三合拱", Element: "水", Zhi: ganzhi.ZhiZi,
			}},
		},
		{
			name:    "san-hui-first-and-last-middle",
			pillars: []string{"乙亥", "己丑", "丙辰", "庚戌"},
			want: []GongJia{{
				ZhuA: 0, ZhuB: 1, Type: "三会拱", Element: "水", Zhi: ganzhi.ZhiZi,
			}},
		},
		{
			name:    "arbitrary-two-branches-with-gap-is-not-gong",
			pillars: []string{"甲子", "乙丑", "丙寅", "丁卯"},
			want:    []GongJia{},
		},
		{
			name:    "complete-sanhe-is-not-gong",
			pillars: []string{"甲申", "丙子", "戊辰", "庚午"},
			want:    []GongJia{},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			chart := atomicChartFromStrings(t, tt.pillars)
			got := ComputeChartExtra(chart).GongJia
			if len(got) != len(tt.want) {
				t.Fatalf("gong = %#v, want %#v", got, tt.want)
			}
			for i := range got {
				if got[i] != tt.want[i] {
					t.Fatalf("gong[%d] = %#v, want %#v", i, got[i], tt.want[i])
				}
			}
		})
	}
}
