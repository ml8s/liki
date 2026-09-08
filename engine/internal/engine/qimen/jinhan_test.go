package qimen

import (
	"encoding/json"
	"strings"
	"testing"
	"time"

	"liki-engine/internal/engine/ganzhi"
	"liki-engine/internal/engine/tianwen"
)

func TestJinhanCatalogAndResolver(t *testing.T) {
	got, err := ResolveMethod("day", "jinhan_yujing", "", "")
	if err != nil {
		t.Fatal(err)
	}
	want := Method{Scope: ScopeDay, School: SchoolJinhanYuJing, Dingju: DingjuNone}
	if got != want {
		t.Fatalf("Jin Han method = %+v, want %+v", got, want)
	}

	invalid := [][3]string{
		{"hour", "jinhan_yujing", ""},
		{"month", "jinhan_yujing", ""},
		{"year", "jinhan_yujing", ""},
		{"day", "jinhan_yujing", "chaibu"},
		{"day", "jinhan_yujing", "zhirun"},
		{"day", "jinhan_yujing", "none"},
	}
	for _, item := range invalid {
		if _, err := ResolveMethod(item[0], item[1], item[2], ""); err == nil {
			t.Fatalf("ResolveMethod(%+v) should reject", item)
		}
	}

	if len(jinhanStarOrder) != 9 || len(jinhanYangStarStarts) != 9 || len(jinhanYinStarStarts) != 9 {
		t.Fatalf("Jin Han star tables = %d/%d/%d; want nine entries each",
			len(jinhanStarOrder), len(jinhanYangStarStarts), len(jinhanYinStarStarts))
	}
	if len(jinhanDoorStarts) != 8 || len(jinhanDoorRing) != 8 || len(jinhanDoorOrder) != 8 {
		t.Fatalf("Jin Han door tables = %d/%d/%d; want eight entries each",
			len(jinhanDoorStarts), len(jinhanDoorRing), len(jinhanDoorOrder))
	}
	if len(jinhanDaySpirits) != 10 {
		t.Fatalf("Jin Han day-spirit stems = %d; want 10", len(jinhanDaySpirits))
	}
	for gan, spirits := range jinhanDaySpirits {
		if len(spirits) != 12 {
			t.Fatalf("Jin Han spirits for %s = %d; want 12", gan, len(spirits))
		}
	}
}

type jinhanGoldenVector struct {
	ID         string            `json:"id"`
	When       string            `json:"when"`
	DayGan     string            `json:"day_gan"`
	DayZhi     string            `json:"day_zhi"`
	YinDun     bool              `json:"yin_dun"`
	Stars      map[string]string `json:"stars"`
	Doors      map[string]string `json:"doors"`
	DaySpirits map[string]string `json:"day_spirits"`
}

type jinhanGoldenFixture struct {
	Source       goldenSource         `json:"source"`
	DomainSource jinhanDomainSource   `json:"domain_source"`
	Vectors      []jinhanGoldenVector `json:"vectors"`
}

type jinhanDomainSource struct {
	Name    string   `json:"name"`
	URL     string   `json:"url"`
	License string   `json:"license"`
	Rules   []string `json:"rules"`
}

func TestJinhanGolden(t *testing.T) {
	var fixture jinhanGoldenFixture
	loadGoldenFixture(t, "jinhan_golden.json", &fixture)
	if fixture.Source.Name != "kentang2017/kinqimen" ||
		fixture.Source.Commit != "e6680ac4ca0b0da5ce3fe637e05f9fc32066ec5a" ||
		fixture.Source.SourceSHA256 != "d804df7b3d30a6f92470330d6adfd0c5ce320e36b1635a33c537065ee3d04682" ||
		len(fixture.Vectors) != 9 {
		t.Fatalf("invalid Jin Han fixture source/vectors: %+v/%d", fixture.Source, len(fixture.Vectors))
	}
	if fixture.DomainSource.Name != "《奇门遁甲秘笈大全·诸葛武侯行兵遁甲金函玉镜卷一》" ||
		len(fixture.DomainSource.Rules) < 4 {
		t.Fatalf("invalid Jin Han domain source: %+v", fixture.DomainSource)
	}

	for _, vector := range fixture.Vectors {
		t.Run(vector.ID, func(t *testing.T) {
			when, err := time.ParseInLocation("2006-01-02 15:04", vector.When, time.FixedZone("CST", 8*3600))
			if err != nil {
				t.Fatal(err)
			}
			chart := ComputeJinhanChart(tianwen.SolarTime(when))
			if chart.Method.Scope != ScopeDay || chart.Method.School != SchoolJinhanYuJing ||
				chart.Method.DingjuMethod != DingjuNone {
				t.Fatalf("method = %+v", chart.Method)
			}
			if chart.Pan.YinDun != vector.YinDun {
				t.Fatalf("yin dun = %v, want %v", chart.Pan.YinDun, vector.YinDun)
			}
			wantGan, err := ganzhi.ParseGan(vector.DayGan)
			if err != nil {
				t.Fatal(err)
			}
			wantZhi, err := ganzhi.ParseZhi(vector.DayZhi)
			if err != nil {
				t.Fatal(err)
			}
			if chart.Pan.RiGan != wantGan || chart.Pan.RiZhi != wantZhi {
				t.Fatalf("day pillar = %s%s, want %s%s", chart.Pan.RiGan, chart.Pan.RiZhi, vector.DayGan, vector.DayZhi)
			}
			if vector.ID == "jinhan-08" &&
				(vector.Stars["艮"] != "太乙" || vector.Stars["离"] != "摄提" || vector.Doors["坎"] != "休门") {
				t.Fatalf("classical yang Jia Zi anchor = %+v", vector)
			}
			if vector.ID == "jinhan-02" &&
				(vector.Stars["坤"] != "太乙" || vector.Stars["坎"] != "摄提" || vector.Doors["离"] != "休门") {
				t.Fatalf("classical yin Jia Zi anchor = %+v", vector)
			}
			for _, palace := range chart.Pan.GongWei {
				gong := palace.Gong.Name
				if got := palace.Xing; got != vector.Stars[gong] {
					t.Fatalf("star %s = %q, want %q", gong, got, vector.Stars[gong])
				}
				wantDoor := vector.Doors[gong]
				if got := palace.Men; got != wantDoor || palace.MenPresent != (wantDoor != "") {
					t.Fatalf("door %s = %q/%v, want %q", gong, got, palace.MenPresent, wantDoor)
				}
			}
			for _, spirit := range chart.Pan.DaySpirits {
				branch := spirit.Zhi.String()
				if got := spirit.Shen; got != vector.DaySpirits[branch] {
					t.Fatalf("day spirit %s = %q, want %q", branch, got, vector.DaySpirits[branch])
				}
			}
		})
	}
}

func TestJinhanChartDoesNotExposeStandardPlate(t *testing.T) {
	when := time.Date(2024, 6, 29, 12, 0, 0, 0, time.FixedZone("CST", 8*3600))
	data, err := json.Marshal(ComputeJinhanChart(tianwen.SolarTime(when)))
	if err != nil {
		t.Fatal(err)
	}
	var output map[string]any
	if err := json.Unmarshal(data, &output); err != nil {
		t.Fatal(err)
	}
	pan, _ := output["pan"].(map[string]any)
	if pan == nil {
		t.Fatalf("missing pan: %v", output)
	}
	for _, field := range []string{"jushu", "zhi_fu_xing", "zhi_shi_men", "tian_pan", "di_pan_gan", "shen"} {
		if _, exists := pan[field]; exists {
			t.Fatalf("Jin Han pan must not expose standard-plate field %q", field)
		}
	}
	if _, exists := pan["day_spirits"]; !exists {
		t.Fatal("Jin Han pan lacks day_spirits")
	}
}

func TestComputeChartWithYongShenRejectsJinhan(t *testing.T) {
	st := tianwen.GregorianToSolar(
		time.Date(2026, 9, 7, 12, 0, 0, 0, time.FixedZone("CST", 8*3600)),
		116.4, 8,
	)
	symbol, err := ParseYongShen("生门")
	if err != nil {
		t.Fatal(err)
	}
	_, err = ComputeChartWithYongShenAndMethod(
		st, []YongShenSymbol{symbol}, BirthDate{},
		Method{Scope: ScopeDay, School: SchoolJinhanYuJing},
	)
	if err == nil || !strings.Contains(err.Error(), "use ComputeJinhanChart") {
		t.Fatalf("error = %v", err)
	}
}

func TestComputeChartWithMethodRejectsJinhanStandardChart(t *testing.T) {
	st := tianwen.GregorianToSolar(
		time.Date(2026, 9, 7, 12, 0, 0, 0, time.FixedZone("CST", 8*3600)),
		116.4, 8,
	)
	_, err := ComputeChartWithMethod(st, Method{Scope: ScopeDay, School: SchoolJinhanYuJing})
	if err == nil || !strings.Contains(err.Error(), "ComputeJinhanChart") {
		t.Fatalf("error = %v, want Jinhan standard-chart rejection", err)
	}

	_, err = ComputeChartWithYongShenAndMethod(
		st, nil, BirthDate{}, Method{Scope: ScopeDay, School: SchoolJinhanYuJing},
	)
	if err == nil || !strings.Contains(err.Error(), "ComputeJinhanChart") {
		t.Fatalf("error = %v, want Jinhan standard-chart rejection", err)
	}
}
