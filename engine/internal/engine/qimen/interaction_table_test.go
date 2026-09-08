package qimen

import (
	"encoding/json"
	"os"
	"strings"
	"testing"

	"liki-engine/internal/engine/ganzhi"
)

func TestInteractionTablesDeclareProvenance(t *testing.T) {
	for _, path := range []string{
		"data/gan_interaction.json",
		"data/men_interaction.json",
		"data/xing_interaction.json",
	} {
		raw, err := os.ReadFile(path)
		if err != nil {
			t.Fatal(err)
		}
		var document struct {
			Provenance string            `json:"provenance"`
			Entries    []json.RawMessage `json:"entries"`
		}
		if err := json.Unmarshal(raw, &document); err != nil {
			t.Fatalf("%s: %v", path, err)
		}
		if document.Provenance != "curated" || len(document.Entries) == 0 {
			t.Fatalf("%s provenance/entries = %q/%d, want curated/non-empty", path, document.Provenance, len(document.Entries))
		}
	}
}

func TestGanInteractionTableContract(t *testing.T) {
	if len(ganInteractionTable) != 72 {
		t.Fatalf("gan interaction entries = %d, want 72", len(ganInteractionTable))
	}
	var document struct {
		Entries []map[string]json.RawMessage `json:"entries"`
	}
	if err := json.Unmarshal(ganInteractionJSON, &document); err != nil {
		t.Fatal(err)
	}
	for i, row := range document.Entries {
		if _, exists := row["name"]; exists {
			t.Fatalf("gan interaction row %d stores mechanically derived name", i)
		}
	}
	seen := map[[2]ganzhi.Gan]bool{}
	for key, entry := range ganInteractionTable {
		if seen[key] {
			t.Fatalf("duplicate gan interaction key %+v", key)
		}
		seen[key] = true
		if key[0] == ganzhi.GanJia || key[1] == ganzhi.GanJia {
			t.Fatalf("gan interaction contains visible Jia %+v", key)
		}
		if entry.PatternName == "" || entry.Meaning == "" {
			t.Fatalf("gan interaction %+v lacks pattern or meaning", key)
		}
	}
	wantMissing := [][2]string{
		{"己", "壬"}, {"己", "癸"}, {"庚", "己"}, {"庚", "壬"},
		{"壬", "癸"}, {"癸", "辛"}, {"丁", "己"}, {"乙", "己"}, {"乙", "壬"},
	}
	for _, pair := range wantMissing {
		earth, err := ganzhi.ParseGan(pair[0])
		if err != nil {
			t.Fatal(err)
		}
		heaven, err := ganzhi.ParseGan(pair[1])
		if err != nil {
			t.Fatal(err)
		}
		key := [2]ganzhi.Gan{earth, heaven}
		if seen[key] {
			t.Fatalf("gan interaction %+v is marked absent but exists", key)
		}
	}
}

func TestGanInteractionDisplayNameIsMechanical(t *testing.T) {
	earth, err := ganzhi.ParseGan("戊")
	if err != nil {
		t.Fatal(err)
	}
	heaven, err := ganzhi.ParseGan("庚")
	if err != nil {
		t.Fatal(err)
	}
	p := pan{GongWei: [9]Gong{{
		DiPanGan: earth,
		TianPan:  []TianPanSymbol{{Gan: heaven, Star: StarTianPeng}},
	}}}
	got := computeGanInteractions(p)
	if len(got) != 1 || got[0].Name != "庚+戊" {
		t.Fatalf("gan interaction display name = %+v, want 庚+戊", got)
	}
}

func TestGanInteractionDoesNotInventMissingRule(t *testing.T) {
	earth, err := ganzhi.ParseGan("己")
	if err != nil {
		t.Fatal(err)
	}
	heaven, err := ganzhi.ParseGan("壬")
	if err != nil {
		t.Fatal(err)
	}
	p := pan{GongWei: [9]Gong{{
		DiPanGan: earth,
		TianPan:  []TianPanSymbol{{Gan: heaven, Star: StarTianPeng}},
	}}}
	if got := computeGanInteractions(p); len(got) != 0 {
		t.Fatalf("missing gan rule invented output %+v", got)
	}
}

func TestMenInteractionTableContract(t *testing.T) {
	if len(menGongTable) != 33 {
		t.Fatalf("men interaction entries = %d, want 33", len(menGongTable))
	}
	wantCounts := map[DoorIndex]int{
		DoorKai: 8, DoorXiu: 5, DoorSheng: 5, DoorShang: 3,
		DoorDu: 3, DoorJing: 3, DoorSi: 3, DoorJingMen: 3,
	}
	counts := map[DoorIndex]int{}
	for key, entry := range menGongTable {
		if key[1] == int(GongZhong)-1 {
			t.Fatalf("men interaction places a door in the center: %+v", key)
		}
		counts[DoorIndex(key[0])]++
		if entry.Name == "" || entry.Meaning == "" {
			t.Fatalf("men interaction %+v lacks name or meaning", key)
		}
		if !strings.Contains(entry.Name, GongIndex(key[1]+1).String()) {
			t.Fatalf("men interaction %+v name %q lacks gong", key, entry.Name)
		}
	}
	for door, want := range wantCounts {
		if counts[door] != want {
			t.Fatalf("door %s interaction count = %d, want %d", door, counts[door], want)
		}
	}
}

func TestMenInteractionDoesNotInventMissingRule(t *testing.T) {
	p := pan{GongWei: [9]Gong{}}
	p.GongWei[int(GongZhen)-1].Door = DoorXiu
	if got := computeMenInteractions(p); len(got) != 0 {
		t.Fatalf("missing men rule invented output %+v", got)
	}
}

func TestXingInteractionTableContract(t *testing.T) {
	if len(xingGongTable) != 30 {
		t.Fatalf("xing interaction entries = %d, want 30", len(xingGongTable))
	}
	var document struct {
		Entries []map[string]json.RawMessage `json:"entries"`
	}
	if err := json.Unmarshal(xingInteractionJSON, &document); err != nil {
		t.Fatal(err)
	}
	for i, row := range document.Entries {
		if _, exists := row["name"]; exists {
			t.Fatalf("xing interaction row %d stores mechanically derived name", i)
		}
	}
	wantCounts := map[StarIndex]int{
		StarTianPeng: 5, StarTianRui: 4, StarTianChong: 3, StarTianFu: 3,
		StarTianQin: 3, StarTianXin: 3, StarTianZhu: 3, StarTianRen: 3, StarTianYing: 3,
	}
	counts := map[StarIndex]int{}
	for key, entry := range xingGongTable {
		counts[StarIndex(key[0])]++
		if entry.Meaning == "" {
			t.Fatalf("xing interaction %+v lacks meaning", key)
		}
	}
	for star, want := range wantCounts {
		if counts[star] != want {
			t.Fatalf("star %s interaction count = %d, want %d", star, counts[star], want)
		}
	}
}

func TestXingInteractionDisplayNameIsMechanical(t *testing.T) {
	p := pan{GongWei: [9]Gong{}}
	p.GongWei[int(GongLi)-1].TianPan = []TianPanSymbol{{Gan: ganzhi.GanWu, Star: StarTianPeng}}
	got := computeXingInteractions(p)
	if len(got) != 1 || got[0].Name != "水星入火宫" {
		t.Fatalf("xing interaction display name = %+v, want 水星入火宫", got)
	}
}

func TestXingInteractionDoesNotInventMissingRule(t *testing.T) {
	p := pan{GongWei: [9]Gong{}}
	p.GongWei[int(GongXun)-1].TianPan = []TianPanSymbol{{Gan: ganzhi.GanWu, Star: StarTianPeng}}
	if got := computeXingInteractions(p); len(got) != 0 {
		t.Fatalf("missing xing rule invented output %+v", got)
	}
}
