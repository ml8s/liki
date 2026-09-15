package liuyao

import (
	"encoding/json"
	"os"
	"strconv"
	"testing"

	"liki-engine/internal/engine/ganzhi"
)

func TestCore64Golden_JingFangEightPalaces(t *testing.T) {
	raw, err := os.ReadFile("testdata/core_64_golden.json")
	if err != nil {
		t.Fatalf("read core-64 golden: %v", err)
	}
	var doc struct {
		Version int `json:"version"`
		Policy  struct {
			PalaceSequence []string `json:"palace_sequence"`
			WorldResponse  string   `json:"world_response"`
		} `json:"policy"`
		Scope struct {
			Palaces            int `json:"palaces"`
			HexagramsPerPalace int `json:"hexagrams_per_palace"`
			CaseCount          int `json:"case_count"`
		} `json:"scope"`
		Cases []struct {
			Index  int    `json:"index"`
			Name   string `json:"name"`
			Palace string `json:"palace"`
			ShiPos int    `json:"shi_pos"`
		} `json:"cases"`
	}
	if err := json.Unmarshal(raw, &doc); err != nil {
		t.Fatalf("decode core-64 golden: %v", err)
	}
	if doc.Version != 1 || doc.Scope.Palaces != 8 || doc.Scope.HexagramsPerPalace != 8 ||
		len(doc.Cases) != 64 || len(doc.Policy.PalaceSequence) != 8 ||
		doc.Policy.WorldResponse != "world_plus_two_positions" {
		t.Fatalf("invalid fixture metadata: %+v", doc)
	}

	seen := make(map[string]bool)
	for _, golden := range doc.Cases {
		if golden.Index < 0 || golden.Index >= 64 {
			t.Fatalf("invalid index %d", golden.Index)
		}
		if seen[golden.Name] {
			t.Fatalf("duplicate hexagram %q", golden.Name)
		}
		seen[golden.Name] = true
		t.Run(strconv.Itoa(golden.Index)+"-"+golden.Name, func(t *testing.T) {
			got := guaTable[golden.Index]
			palaceIndex := -1
			for i, palace := range doc.Policy.PalaceSequence {
				if palace == golden.Palace {
					palaceIndex = i
					break
				}
			}
			if palaceIndex < 0 {
				t.Fatalf("invalid palace %q", golden.Palace)
			}
			if got.Name != golden.Name || got.PalaceIdx != palaceIndex {
				t.Fatalf("index %d = %s/%s, want %s/%s", golden.Index, got.Name, doc.Policy.PalaceSequence[got.PalaceIdx], golden.Name, golden.Palace)
			}
			if got.ShiPos != golden.ShiPos {
				t.Fatalf("%s shi = %d, want %d", golden.Name, got.ShiPos, golden.ShiPos)
			}
			binary := -1
			for candidate, index := range binaryToGuaTable {
				if int(index) == golden.Index {
					binary = candidate
					break
				}
			}
			if binary < 0 {
				t.Fatalf("hexagram %q has no binary encoding", golden.Name)
			}
			yaos := [6]YaoType{}
			for position := 0; position < 6; position++ {
				if binary&(1<<position) != 0 {
					yaos[position] = ShaoYang
				} else {
					yaos[position] = ShaoYin
				}
			}
			chart := computeGuaPan(yaos, ganzhi.Zhu{Gan: ganzhi.GanJia, Zhi: ganzhi.ZhiZi})
			if chart.BenGua != guaIndex(golden.Index) {
				t.Fatalf("binary %d produced %q, want %q", binary, chart.Name, golden.Name)
			}
			wantYing := (golden.ShiPos+2)%6 + 1
			for _, line := range chart.Lines {
				if line.Position != golden.ShiPos && line.Position != wantYing {
					continue
				}
				want := ""
				if line.Position == golden.ShiPos {
					want = "世"
				} else {
					want = "应"
				}
				if line.ShiYing != want {
					t.Fatalf("%s position %d shi_ying = %q, want %q", golden.Name, line.Position, line.ShiYing, want)
				}
			}
		})
	}
}
