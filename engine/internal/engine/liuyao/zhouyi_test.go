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

// 八宫世应权威锚点（张全海《易经的预测》京房卦官世位表 + 甘雨泽《易经与人生决策》）：
// 八纯卦世六、一世初、二世二、三世三、四世四、五世五、游魂四、归魂三（坤宫同此模式）。
func TestZhouYi_ShiYingAnchors(t *testing.T) {
	want := map[string]int{
		// 乾宫
		"乾为天": 6, "天风姤": 1, "天山遁": 2, "天地否": 3, "风地观": 4, "山地剥": 5, "火地晋": 4, "火天大有": 3,
		// 坤宫（同八纯世六模式）
		"坤为地": 6, "地雷复": 1, "地泽临": 2, "地天泰": 3, "雷天大壮": 4, "泽天夬": 5, "水天需": 4, "水地比": 3,
		// 兑宫
		"兑为泽": 6, "泽水困": 1, "泽地萃": 2, "泽山咸": 3, "水山蹇": 4, "地山谦": 5, "雷山小过": 4, "雷泽归妹": 3,
		// 离宫
		"离为火": 6, "火山旅": 1, "火风鼎": 2, "火水未济": 3, "山水蒙": 4, "风水涣": 5, "天水讼": 4, "天火同人": 3,
		// 震宫
		"震为雷": 6, "雷地豫": 1, "雷水解": 2, "雷风恒": 3, "地风升": 4, "水风井": 5, "泽风大过": 4, "泽雷随": 3,
		// 巽宫
		"巽为风": 6, "风天小畜": 1, "风火家人": 2, "风雷益": 3, "天雷无妄": 4, "火雷噬嗑": 5, "山雷颐": 4, "山风蛊": 3,
		// 坎宫
		"坎为水": 6, "水泽节": 1, "水雷屯": 2, "水火既济": 3, "泽火革": 4, "雷火丰": 5, "地火明夷": 4, "地水师": 3,
		// 艮宫
		"艮为山": 6, "山火贲": 1, "山天大畜": 2, "山泽损": 3, "火泽睽": 4, "天泽履": 5, "风泽中孚": 4, "风山渐": 3,
	}
	for i, g := range guaTable {
		if g.Name == "" {
			continue
		}
		w, ok := want[g.Name]
		if !ok {
			t.Errorf("gua %d (%s) 不在锚点表", i, g.Name)
			continue
		}
		if g.ShiPos != w {
			t.Errorf("%s: shi_pos = %d, want %d（京房世应）", g.Name, g.ShiPos, w)
		}
	}
}
