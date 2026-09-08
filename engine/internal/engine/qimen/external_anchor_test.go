package qimen

import (
	"encoding/json"
	"os"
	"path/filepath"
	"testing"
	"time"

	"liki-engine/internal/engine/ganzhi"
	"liki-engine/internal/engine/tianwen"
)

type goldenMethod struct {
	Scope  string `json:"scope"`
	School string `json:"school"`
	Dingju string `json:"dingju_method"`
}

type goldenSource struct {
	Name          string `json:"name"`
	URL           string `json:"url"`
	Commit        string `json:"commit"`
	SourceFile    string `json:"source_file"`
	SourceSHA256  string `json:"source_sha256"`
	License       string `json:"license"`
	LicenseSHA256 string `json:"license_sha256,omitempty"`
}

type externalAnchor struct {
	ID          string       `json:"id"`
	Provenance  string       `json:"provenance"`
	Method      goldenMethod `json:"method"`
	When        string       `json:"when"`
	DayPillar   string       `json:"day_pillar"`
	HourPillar  string       `json:"hour_pillar"`
	YinDun      bool         `json:"yin_dun"`
	Ju          int          `json:"ju"`
	JieQi       string       `json:"jie_qi"`
	Yuan        string       `json:"yuan"`
	XunShou     string       `json:"xun_shou"`
	DutyStar    string       `json:"duty_star"`
	XunShouGong int          `json:"xun_shou_gong"`
	KongWang    string       `json:"kong_wang"`
	Palaces     [9]string    `json:"palaces"`
}

type externalGoldenFixture struct {
	Source  goldenSource     `json:"source"`
	Vectors []externalAnchor `json:"vectors"`
}

func loadExternalGoldenFixture(t *testing.T) externalGoldenFixture {
	t.Helper()
	var fixture externalGoldenFixture
	loadGoldenFixture(t, "external_rotating_golden.json", &fixture)
	wantSource := atopxGoldenSource()
	if fixture.Source != wantSource || len(fixture.Vectors) != 11 {
		t.Fatalf("external golden source/vectors = %+v/%d", fixture.Source, len(fixture.Vectors))
	}
	seen := map[string]bool{}
	defaultMethod := goldenMethod{Scope: "hour", School: "zhuanpan", Dingju: "chaibu"}
	for _, vector := range fixture.Vectors {
		if seen[vector.When] || vector.ID == "" || vector.Method != defaultMethod ||
			vector.Provenance != "atopx_checked_in_zhirun_vector_verified_as_chaibu" {
			t.Fatalf("invalid external golden vector %+v", vector)
		}
		seen[vector.When] = true
	}
	return fixture
}

func atopxGoldenSource() goldenSource {
	return goldenSource{
		Name: "atopx/qimen", URL: "https://github.com/atopx/qimen",
		Commit:       "8eb06d007d4a5fcc5352d9054f81469e5f023f45",
		SourceFile:   "qimen_golden_test.go",
		SourceSHA256: "fb0f27c58e62152f130a0026c713f985c36c42651d319f5601e32784fed0b026",
		License:      "MIT",
	}
}

func loadGoldenFixture(t *testing.T, name string, target any) {
	t.Helper()
	data, err := os.ReadFile(filepath.Join("testdata", name))
	if err != nil {
		t.Fatal(err)
	}
	if err := json.Unmarshal(data, target); err != nil {
		t.Fatalf("load golden fixture %s: %v", name, err)
	}
}

var externalAnchorStars = map[rune]StarIndex{
	'蓬': StarTianPeng, '芮': StarTianRui, '冲': StarTianChong,
	'辅': StarTianFu, '英': StarTianYing, '柱': StarTianZhu,
	'心': StarTianXin, '任': StarTianRen,
}

var externalAnchorSpirits = map[rune]SpiritIndex{
	'符': SpiritZhiFu, '蛇': SpiritTengShe, '阴': SpiritTaiYin,
	'合': SpiritLiuHe, '虎': SpiritGouChen, '玄': SpiritZhuQue,
	'地': SpiritJiuDi, '天': SpiritJiuTian,
}

func TestExternalRotatingPlateAnchors(t *testing.T) {
	for _, anchor := range loadExternalGoldenFixture(t).Vectors {
		t.Run(anchor.When, func(t *testing.T) {
			civil, err := time.ParseInLocation("2006-01-02 15:04", anchor.When, time.FixedZone("CST", 8*3600))
			if err != nil {
				t.Fatal(err)
			}
			chart := ComputeChart(tianwen.SolarTime(civil))
			assertExternalChart(t, chart, anchor)
			assertExternalPalaces(t, chart, anchor.Palaces)
		})
	}
}

func assertExternalChart(t *testing.T, chart Chart, anchor externalAnchor) {
	t.Helper()
	if got := ganzhi.GanName(chart.Pan.RiGan) + ganzhi.ZhiName(chart.Pan.RiZhi); got != anchor.DayPillar {
		t.Errorf("day pillar = %s, want %s", got, anchor.DayPillar)
	}
	if got := ganzhi.GanName(chart.Pan.HourGan) + ganzhi.ZhiName(chart.Pan.HourZhi); got != anchor.HourPillar {
		t.Errorf("hour pillar = %s, want %s", got, anchor.HourPillar)
	}
	if chart.Pan.YinDun != anchor.YinDun || chart.Pan.Jushu != anchor.Ju {
		t.Errorf("dun/ju = %v/%d, want %v/%d", chart.Pan.YinDun, chart.Pan.Jushu, anchor.YinDun, anchor.Ju)
	}
	if chart.Method.JieQi != anchor.JieQi || chart.Method.Yuan != anchor.Yuan {
		t.Errorf("jieqi/yuan = %s/%s, want %s/%s", chart.Method.JieQi, chart.Method.Yuan, anchor.JieQi, anchor.Yuan)
	}
	if chart.Method.LeadXunShou.LiuYi != anchor.XunShou {
		t.Errorf("xun shou = %s, want %s", chart.Method.LeadXunShou.LiuYi, anchor.XunShou)
	}
	if got := chart.Pan.DutyStar.String(); got != anchor.DutyStar {
		t.Errorf("duty star = %s, want %s", got, anchor.DutyStar)
	}
	if got := int(chart.Method.LeadXunShouGong); got != anchor.XunShouGong {
		t.Errorf("xun shou palace = %d, want %d", got, anchor.XunShouGong)
	}
	if got := ganzhi.ZhiName(chart.Pan.KongWang[0].Branch) + ganzhi.ZhiName(chart.Pan.KongWang[1].Branch); got != anchor.KongWang {
		t.Errorf("kong wang = %s, want %s", got, anchor.KongWang)
	}
}

func assertExternalPalaces(t *testing.T, chart Chart, specs [9]string) {
	t.Helper()
	centerEarthGan := chart.Pan.GongWei[int(GongZhong)-1].DiPanGan
	for i, spec := range specs {
		runes := []rune(spec)
		if len(runes) != 6 {
			t.Fatalf("palace %d spec %q must contain six runes", i+1, spec)
		}
		palace := chart.Pan.GongWei[i]
		if i == int(GongZhong)-1 {
			if palace.TianPanGan == nil || palace.TianPanGan.String() != string(runes[0]) {
				t.Errorf("palace %d heaven gan = %v, want %s", i+1, palace.TianPanGan, string(runes[0]))
			}
		}
		if got := palace.DiPanGan.String(); got != string(runes[1]) {
			t.Errorf("palace %d earth gan = %s, want %s", i+1, got, string(runes[1]))
		}
		if got := palace.AnGan.String(); got != string(runes[2]) {
			t.Errorf("palace %d hidden gan = %s, want %s", i+1, got, string(runes[2]))
		}
		if i == int(GongZhong)-1 {
			if len(palace.TianPan) != 0 || palace.Door != 0 || palace.Spirit != 0 {
				t.Errorf("central palace has ring layers: %+v", palace)
			}
			continue
		}
		assertExternalPalaceLayers(t, i+1, palace, runes, centerEarthGan)
	}
}

func assertExternalPalaceLayers(
	t *testing.T, palaceNumber int, palace Gong, runes []rune, centerEarthGan ganzhi.Gan,
) {
	t.Helper()
	wantStar := externalAnchorStars[runes[3]]
	foundStar, foundQin := false, false
	for _, item := range palace.TianPan {
		if item.Star == wantStar {
			foundStar = true
			if got := item.Gan.String(); got != string(runes[0]) {
				t.Errorf("palace %d %s heaven gan = %s, want %s", palaceNumber, wantStar, got, string(runes[0]))
			}
		}
		if wantStar == StarTianRui && item.Star == StarTianQin {
			foundQin = true
			if item.Gan != centerEarthGan {
				t.Errorf("palace %d TianQin gan = %s, want center earth gan %s", palaceNumber, item.Gan, centerEarthGan)
			}
		}
	}
	if !foundStar {
		t.Errorf("palace %d lacks star %s", palaceNumber, wantStar)
	}
	if wantStar == StarTianRui && !foundQin {
		t.Errorf("palace %d lacks explicit TianQin fact", palaceNumber)
	}
	wantDoor, err := ParseDoorIndex(string(runes[4]) + "门")
	if err != nil {
		t.Fatalf("palace %d door: %v", palaceNumber, err)
	}
	if palace.Door != wantDoor {
		t.Errorf("palace %d door = %s, want %s", palaceNumber, palace.Door, wantDoor)
	}
	if palace.Spirit != externalAnchorSpirits[runes[5]] {
		t.Errorf("palace %d spirit slot = %d, want %d", palaceNumber, palace.Spirit, externalAnchorSpirits[runes[5]])
	}
}
