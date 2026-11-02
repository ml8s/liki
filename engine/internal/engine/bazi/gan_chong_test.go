package bazi

import (
	"testing"

	"liki-engine/internal/engine/ganzhi"
)

func TestGanChong_ClassicPairsAndPositions(t *testing.T) {
	tests := []struct {
		name    string
		pillars []string
		want    []GanChongDetection
	}{
		{
			name:    "adjacent",
			pillars: []string{"甲子", "庚午", "丙寅", "壬申"},
			want: []GanChongDetection{
				{GanA: "甲", GanB: "庚", PillarA: 0, PillarB: 1, Position: "adjacent"},
				{GanA: "丙", GanB: "壬", PillarA: 2, PillarB: 3, Position: "adjacent"},
			},
		},
		{
			name:    "separated",
			pillars: []string{"甲子", "丙寅", "庚午", "壬申"},
			want: []GanChongDetection{
				{GanA: "甲", GanB: "庚", PillarA: 0, PillarB: 2, Position: "separated"},
				{GanA: "丙", GanB: "壬", PillarA: 1, PillarB: 3, Position: "separated"},
			},
		},
		{
			name:    "remote",
			pillars: []string{"甲子", "丙寅", "戊辰", "庚午"},
			want: []GanChongDetection{
				{GanA: "甲", GanB: "庚", PillarA: 0, PillarB: 3, Position: "remote"},
			},
		},
		{
			name:    "wu_ji_and_non_classic_pairs_are_not_chong",
			pillars: []string{"戊子", "甲午", "己丑", "乙未"},
			want:    []GanChongDetection{},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			bz := pillarsBaziFromStrings(t, tt.pillars)
			got := detectGanChong(bz)
			if len(got) != len(tt.want) {
				t.Fatalf("pairs = %#v, want %#v", got, tt.want)
			}
			for i := range got {
				if got[i] != tt.want[i] {
					t.Fatalf("pairs[%d] = %#v, want %#v", i, got[i], tt.want[i])
				}
			}
		})
	}
}

func TestGanChong_GuanXiDistinguishTightPairs(t *testing.T) {
	chart := pillarsCoreFromStrings(t, []string{"甲子", "丙寅", "庚午", "壬申"})
	full := ComputeChartFromCore(chart)
	type relationKey struct {
		Kind  string
		Group string
	}
	want := map[relationKey]bool{
		{Kind: "gan_chong_candidate", Group: "甲庚"}: true,
		{Kind: "liu_chong", Group: "子午"}:           true,
		{Kind: "san_he_partial", Group: "寅午"}:      true,
	}
	got := make(map[relationKey]bool, len(full.GuanXi))
	for _, group := range full.GuanXi {
		got[relationKey{Kind: group.Kind, Group: group.Group}] = true
	}
	for group := range want {
		if !got[group] {
			t.Fatalf("missing relation group %#v; got %#v", group, full.GuanXi)
		}
	}
	if got[relationKey{Kind: "gan_chong", Group: "甲庚"}] {
		t.Fatalf("separated 甲庚 must be candidate, not tight gan_chong")
	}
}

func pillarsBaziFromStrings(t *testing.T, pillars []string) ganzhi.Bazi {
	t.Helper()
	if len(pillars) != 4 {
		t.Fatalf("pillar count = %d, want 4", len(pillars))
	}
	chart := pillarsCoreFromStrings(t, pillars)
	return chart.ToBazi()
}
