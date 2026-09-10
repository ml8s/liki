package qimen

import (
	"encoding/json"
	"os"
	"reflect"
	"testing"

	"liki-engine/internal/engine/ganzhi"
)

func oracleGong(t *testing.T, name string) GongIndex {
	t.Helper()
	var gong GongIndex
	if err := gong.UnmarshalJSON([]byte(`"` + name + `"`)); err != nil {
		t.Fatal(err)
	}
	return gong
}

func oracleGan(t *testing.T, name string) ganzhi.Gan {
	t.Helper()
	gan, err := ganzhi.ParseGan(name)
	if err != nil {
		t.Fatal(err)
	}
	return gan
}

func TestDomainOracle_QimenSpecializedContext(t *testing.T) {
	raw, err := os.ReadFile("../../../../tests/fixtures/domain_oracle/qimen_specialized.json")
	if err != nil {
		t.Fatalf("read oracle: %v", err)
	}
	var doc struct {
		GengGe []struct {
			ID              string            `json:"id"`
			Pillars         map[string]string `json:"pillars"`
			GanInteractions []struct {
				Gong       string `json:"gong"`
				DiPanGan   string `json:"di_pan_gan"`
				TianPanGan string `json:"tian_pan_gan"`
			} `json:"gan_interactions"`
			Expected []struct {
				Level    string `json:"level"`
				Gong     string `json:"gong"`
				EarthGan string `json:"earth_gan"`
			} `json:"expected"`
		} `json:"geng_ge_cases"`
		Lost []struct {
			ID         string `json:"id"`
			ShiGanGong string `json:"shi_gan_gong"`
			Door       string `json:"door"`
			WangShuai  string `json:"wang_shuai"`
			Expected   []struct {
				Gong      string `json:"gong"`
				Door      string `json:"door"`
				WangShuai string `json:"wang_shuai"`
			} `json:"expected"`
		} `json:"lost_cases"`
		Thief []struct {
			ID               string  `json:"id"`
			Gong             string  `json:"gong"`
			Star             *string `json:"star"`
			Spirit           *string `json:"spirit"`
			TraditionalLabel string  `json:"traditional_label"`
			Patterns         []struct {
				Name       string `json:"name"`
				Gong       string `json:"gong"`
				Auspicious bool   `json:"auspicious"`
			} `json:"patterns"`
			Expected []ThiefFact `json:"expected"`
		} `json:"thief_cases"`
		Tianwang []struct {
			ID              string `json:"id"`
			HourGong        string `json:"hour_gong"`
			GanInteractions []struct {
				Gong       string `json:"gong"`
				TianPanGan string `json:"tian_pan_gan"`
			} `json:"gan_interactions"`
			Expected []TianwangFact `json:"expected"`
		} `json:"tianwang_cases"`
	}
	if err := json.Unmarshal(raw, &doc); err != nil {
		t.Fatalf("decode oracle: %v", err)
	}
	if len(doc.GengGe) != 2 || len(doc.Lost) != 1 || len(doc.Thief) != 2 || len(doc.Tianwang) != 1 {
		t.Fatalf("oracle coverage incomplete: geng=%d lost=%d thief=%d tianwang=%d", len(doc.GengGe), len(doc.Lost), len(doc.Thief), len(doc.Tianwang))
	}

	for _, tc := range doc.GengGe {
		t.Run("geng/"+tc.ID, func(t *testing.T) {
			pillarAt := func(prefix string) ganzhi.Zhu {
				name := tc.Pillars[prefix]
				return ganzhi.Zhu{
					Gan: oracleGan(t, string([]rune(name)[0])),
					Zhi: mustOracleZhi(t, string([]rune(name)[1])),
				}
			}
			bz := ganzhi.Bazi{
				Nian: pillarAt("nian"), Yue: pillarAt("yue"),
				Ri: pillarAt("ri"), Shi: pillarAt("shi"),
			}
			chart := Chart{}
			for _, row := range tc.GanInteractions {
				chart.GanInteractions = append(chart.GanInteractions, GanInteraction{
					Gong:       oracleGong(t, row.Gong),
					DiPanGan:   oracleGan(t, row.DiPanGan),
					TianPanGan: oracleGan(t, row.TianPanGan),
				})
			}
			got := computeSpecializedContext(chart, bz).GengGe
			if len(got) != len(tc.Expected) {
				t.Fatalf("geng facts = %#v, want %d rows", got, len(tc.Expected))
			}
			for index, want := range tc.Expected {
				fact := got[index]
				if fact.Level != want.Level || fact.Gong.String() != want.Gong || fact.EarthGan.String() != want.EarthGan {
					t.Fatalf("row %d = %#v, want %#v", index, fact, want)
				}
			}
		})
	}

	for _, tc := range doc.Lost {
		t.Run("lost/"+tc.ID, func(t *testing.T) {
			chart := Chart{ShiGanPalace: oracleGong(t, tc.ShiGanGong)}
			chart.ShiGanGongWangShuai.WangShuaiName = tc.WangShuai
			palace := oracleGong(t, tc.ShiGanGong)
			chart.ShiGanGongWangShuai.WangShuai = oracleWangShuai(t, tc.WangShuai)
			chart.Pan.GongWei[palace-1].DoorSet = true
			chart.Pan.GongWei[palace-1].Door = oracleDoor(t, tc.Door)
			got := computeSpecializedContext(chart, ganzhi.Bazi{}).LostContext
			if len(got) != 1 || got[0].Gong.String() != tc.Expected[0].Gong ||
				got[0].Door != tc.Expected[0].Door || got[0].WangShuai != tc.Expected[0].WangShuai {
				t.Fatalf("lost = %#v, want %#v", got, tc.Expected)
			}
		})
	}

	for _, tc := range doc.Thief {
		t.Run("thief/"+tc.ID, func(t *testing.T) {
			gong := oracleGong(t, tc.Gong)
			chart := Chart{}
			chart.Pan.GongWei[gong-1].Gong = PalaceIdentity{Name: gong.String()}
			if tc.Star != nil && *tc.Star == "天蓬" {
				chart.Pan.GongWei[gong-1].TianPan = append(chart.Pan.GongWei[gong-1].TianPan, TianPanSymbol{Star: StarTianPeng})
				chart.XingGongWuXing = append(chart.XingGongWuXing, XingGongWuXing{
					Star: StarTianPeng, Gong: gong, TraditionalLabel: tc.TraditionalLabel,
				})
			}
			if tc.Spirit != nil && *tc.Spirit == "朱雀" {
				chart.Pan.GongWei[gong-1].SpiritSet = true
				chart.Pan.GongWei[gong-1].Spirit = SpiritZhuQue
			}
			for _, pattern := range tc.Patterns {
				chart.Patterns = append(chart.Patterns, Pattern{
					Name: pattern.Name, Auspicious: pattern.Auspicious,
					GongWei: []GongIndex{gong},
				})
			}
			got := computeSpecializedContext(chart, ganzhi.Bazi{}).Thief
			if !reflect.DeepEqual(got, tc.Expected) {
				t.Fatalf("thief = %#v, want %#v", got, tc.Expected)
			}
		})
	}

	for _, tc := range doc.Tianwang {
		t.Run("tianwang/"+tc.ID, func(t *testing.T) {
			chart := Chart{ShiGanPalace: oracleGong(t, tc.HourGong)}
			for _, row := range tc.GanInteractions {
				chart.GanInteractions = append(chart.GanInteractions, GanInteraction{
					Gong: oracleGong(t, row.Gong), TianPanGan: oracleGan(t, row.TianPanGan),
				})
			}
			got := computeSpecializedContext(chart, ganzhi.Bazi{}).Tianwang
			if len(got) != len(tc.Expected) {
				t.Fatalf("tianwang rows = %d, want %d", len(got), len(tc.Expected))
			}
			for index, want := range tc.Expected {
				if got[index].Gong != want.Gong ||
					got[index].HourGong != want.HourGong ||
					got[index].RelationToHour != want.RelationToHour {
					t.Fatalf("row %d = %#v, want %#v", index, got[index], want)
				}
			}
		})
	}
}

func mustOracleZhi(t *testing.T, name string) ganzhi.Zhi {
	t.Helper()
	zhi, err := ganzhi.ParseZhi(name)
	if err != nil {
		t.Fatal(err)
	}
	return zhi
}

func oracleWangShuai(t *testing.T, name string) ganzhi.WangShuai {
	t.Helper()
	value, err := ganzhi.ParseWangShuai(name)
	if err != nil {
		t.Fatal(err)
	}
	return value
}

func oracleDoor(t *testing.T, name string) DoorIndex {
	t.Helper()
	for index := DoorXiu; index <= DoorZhong; index++ {
		if index.String() == name {
			return index
		}
	}
	t.Fatalf("unknown door %q", name)
	return 0
}
