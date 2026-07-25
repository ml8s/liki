package liuyao

import (
	"testing"
)

func TestZhouYi_All64(t *testing.T) {
	if len(zhouyiTable) != 64 {
		t.Fatalf("zhouyi entries = %d, want 64", len(zhouyiTable))
	}
	for i, g := range zhouyiTable {
		if g.Name == "" {
			t.Errorf("gua[%d]: empty name", i)
		}
		if g.GuaCi == "" {
			t.Errorf("gua[%d]: empty 卦辞", i)
		}
		for j, y := range g.YaoCi {
			if y == "" {
				t.Errorf("gua[%d] 爻[%d]: empty", i, j)
			}
		}
	}
}

func TestZhouYi_GetGuaCi_Bounds(t *testing.T) {
	if _, err := GetGuaCi(-1); err == nil {
		t.Error("expected error for index -1")
	}
	if _, err := GetGuaCi(64); err == nil {
		t.Error("expected error for index 64")
	}
	if _, err := GetGuaCi(0); err != nil {
		t.Errorf("index 0: %v", err)
	}
	if _, err := GetGuaCi(63); err != nil {
		t.Errorf("index 63: %v", err)
	}
}

func TestZhouYi_WellKnownEntries(t *testing.T) {
	// 京房八宫序 (NOT 周易序卦序)
	tests := []struct {
		idx      int
		name     string
		wantGua  string
		wantYao0 string
		wantYao5 string
	}{
		{idx: 0, name: "乾", wantGua: "元亨利贞", wantYao0: "潜龙勿用", wantYao5: "亢龙有悔"},
		{idx: 56, name: "坤", wantGua: "元亨，利牝马之贞。君子有攸往，先迷后得主，利",
			wantYao0: "履霜，坚冰至", wantYao5: "龙战于野，其血玄黄"},
		{idx: 47, name: "师", wantGua: "贞丈人吉，无咎", wantYao0: "师出以律，否臧凶"},
		{idx: 63, name: "比", wantGua: "吉。原筮元永贞，无咎。不宁方来，后夫凶"},
		{idx: 42, name: "屯", wantGua: "元亨利贞。勿用有攸往，利建侯"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			g, err := GetGuaCi(tt.idx)
			if err != nil {
				t.Fatal(err)
			}
			if g.Name != tt.name {
				t.Errorf("Name = %q, want %q", g.Name, tt.name)
			}
			if g.GuaCi != tt.wantGua {
				t.Errorf("卦辞 = %q, want %q", g.GuaCi, tt.wantGua)
			}
			if tt.wantYao0 != "" && g.YaoCi[0] != tt.wantYao0 {
				t.Errorf("初爻 = %q, want %q", g.YaoCi[0], tt.wantYao0)
			}
			if tt.wantYao5 != "" && g.YaoCi[5] != tt.wantYao5 {
				t.Errorf("上爻 = %q, want %q", g.YaoCi[5], tt.wantYao5)
			}
		})
	}
}
