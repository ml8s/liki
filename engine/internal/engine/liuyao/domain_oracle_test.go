package liuyao

import (
	"encoding/json"
	"os"
	"testing"

	"liki-engine/internal/engine/ganzhi"
)

func TestDomainOracle_LiuyaoNaJiaAndWorld(t *testing.T) {
	raw, err := os.ReadFile("../../../../tests/fixtures/domain_oracle/liuyao_core.json")
	if err != nil {
		t.Fatalf("read liuyao oracle: %v", err)
	}
	var doc struct {
		NaJia []struct {
			Palace    string   `json:"palace"`
			UpperStem string   `json:"upper_stem"`
			LowerStem string   `json:"lower_stem"`
			Branches  []string `json:"branches"`
		} `json:"na_jia_all_8"`
		World []struct {
			Sequence int `json:"sequence"`
			World    int `json:"world"`
			Other    int `json:"other"`
		} `json:"world_by_palace_sequence"`
		VisibleYongShen []struct {
			Hexagram         string `json:"hexagram"`
			YongShen         string `json:"yong_shen"`
			ExpectedPosition int    `json:"expected_position"`
		} `json:"visible_yong_shen_positions"`
	}
	if err := json.Unmarshal(raw, &doc); err != nil {
		t.Fatalf("decode liuyao oracle: %v", err)
	}
	for _, tc := range doc.NaJia {
		pi := -1
		for i, name := range palaceNames {
			if name == tc.Palace {
				pi = i
				break
			}
		}
		if pi < 0 {
			t.Fatalf("unknown palace %s", tc.Palace)
		}
		if ganzhi.GanName(naGanTable[pi][0]) != tc.LowerStem || ganzhi.GanName(naGanTable[pi][1]) != tc.UpperStem {
			t.Fatalf("%s na gan = %s/%s, want %s/%s", tc.Palace, ganzhi.GanName(naGanTable[pi][0]), ganzhi.GanName(naGanTable[pi][1]), tc.LowerStem, tc.UpperStem)
		}
		for i, branch := range tc.Branches {
			if got := ganzhi.ZhiName(naZhiTable[pi][i]); got != branch {
				t.Errorf("%s branch %d = %s, want %s", tc.Palace, i+1, got, branch)
			}
		}
	}
	for _, tc := range doc.World {
		for palace := 0; palace < 8; palace++ {
			meta := guaTable[palace*8+tc.Sequence]
			if meta.ShiPos != tc.World {
				t.Errorf("%s sequence %d world = %d, want %d", meta.Name, tc.Sequence, meta.ShiPos, tc.World)
			}
		}
		wantOther := (tc.World+2)%6 + 1
		if wantOther != tc.Other {
			t.Fatalf("oracle other invalid for sequence %d", tc.Sequence)
		}
	}
	for _, tc := range doc.VisibleYongShen {
		t.Run(tc.Hexagram+"_"+tc.YongShen, func(t *testing.T) {
			yong, err := ParseYongShen(tc.YongShen)
			if err != nil {
				t.Fatal(err)
			}
			chart := computeGuaPan(
				[6]YaoType{ShaoYang, ShaoYang, ShaoYang, ShaoYang, ShaoYang, ShaoYang},
				ganzhi.Zhu{Gan: ganzhi.GanJia, Zhi: ganzhi.ZhiZi},
			)
			if chart.Name != tc.Hexagram {
				t.Fatalf("chart name = %s, want %s", chart.Name, tc.Hexagram)
			}
			if got := chart.findYongShen(yong); got != tc.ExpectedPosition {
				t.Fatalf("yong_shen position = %d, want %d", got, tc.ExpectedPosition)
			}
		})
	}
}
