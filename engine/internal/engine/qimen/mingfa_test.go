package qimen

import (
	"testing"
	"time"

	"liki-engine/internal/engine/ganzhi"
	"liki-engine/internal/engine/tianwen"
)

func TestMingFaCatalogAndResolver(t *testing.T) {
	if SchoolMingFaFeiPan.String() != "mingfa_feipan" {
		t.Fatalf("Ming Fa school = %s", SchoolMingFaFeiPan)
	}
	got, err := ResolveMethod("hour", "mingfa_feipan", "", "")
	if err != nil {
		t.Fatal(err)
	}
	want := Method{Scope: ScopeHour, School: SchoolMingFaFeiPan, Dingju: DingjuChaiBu}
	if got != want {
		t.Fatalf("Ming Fa method = %+v, want %+v", got, want)
	}
	invalid := [][3]string{
		{"day", "mingfa_feipan", "chaibu"},
		{"month", "mingfa_feipan", ""},
		{"year", "mingfa_feipan", ""},
		{"hour", "mingfa_feipan", "zhirun"},
	}
	for _, item := range invalid {
		if _, err := ResolveMethod(item[0], item[1], item[2], ""); err == nil {
			t.Fatalf("ResolveMethod(%+v) should reject", item)
		}
	}
}

func TestMingFaDoorIndex(t *testing.T) {
	if DoorZhong.String() != "中门" {
		t.Fatalf("center door = %s", DoorZhong)
	}
	parsed, err := ParseDoorIndex("中门")
	if err != nil || parsed != DoorZhong {
		t.Fatalf("ParseDoorIndex(中门) = %v, %v; want %v", parsed, err, DoorZhong)
	}
}

type mingfaPalaceWant struct {
	Heaven       string `json:"heaven"`
	Earth        string `json:"earth"`
	Star         string `json:"star"`
	Door         string `json:"door"`
	Spirit       string `json:"spirit"`
	HiddenPillar string `json:"hidden_pillar"`
}

type mingfaVector struct {
	Provenance  string             `json:"provenance"`
	When        string             `json:"when"`
	Method      goldenMethod       `json:"method"`
	ActualJieQi string             `json:"actual_jie_qi"`
	YongJuJieQi string             `json:"yong_ju_jie_qi"`
	YinDun      bool               `json:"yin_dun"`
	Ju          int                `json:"ju"`
	Yuan        string             `json:"yuan"`
	Lead        string             `json:"lead"`
	XunShou     string             `json:"xun_shou"`
	KongWang    string             `json:"kong_wang"`
	DutyStar    string             `json:"duty_star"`
	DutySource  int                `json:"duty_source"`
	DutyLand    int                `json:"duty_land"`
	DutyDoor    string             `json:"duty_door"`
	DoorLand    int                `json:"door_land"`
	Palaces     []mingfaPalaceWant `json:"palaces"`
}

type mingfaFixture struct {
	Source           goldenSource            `json:"source"`
	Vectors          []mingfaVector          `json:"vectors"`
	ReferenceAnchors []mingfaReferenceAnchor `json:"reference_anchors"`
}

func loadMingFaFixture(t *testing.T) mingfaFixture {
	t.Helper()
	var fixture mingfaFixture
	loadGoldenFixture(t, "mingfa_golden.json", &fixture)
	wantSource := goldenSource{
		Name: "potuo/feipan-qimen", URL: "https://github.com/potuo/feipan-qimen",
		Commit:       "e4fe69d45b91ff18b7de04ea3555a9ff9f69eafa",
		SourceFile:   "app/src/test/java/com/potuo/feipanqimen2/qimen/QimenCalculatorTest.kt",
		SourceSHA256: "52a0e06219bdaba06981563a810d21c2ebec894d6607af2f1f76b976a52a6644",
		License:      "MIT", LicenseSHA256: "dd127edca411f5d3d54ec2557cc18297abdf7c546ca76edf3a996a5f9cb279b7",
	}
	if fixture.Source != wantSource || len(fixture.Vectors) != 1 || fixture.Vectors[0].Provenance != "potuo_checked_in" {
		t.Fatalf("Ming Fa golden = %+v/%d", fixture.Source, len(fixture.Vectors))
	}
	if len(fixture.ReferenceAnchors) != 4 {
		t.Fatalf("Ming Fa reference anchors = %d, want 4", len(fixture.ReferenceAnchors))
	}
	return fixture
}

type mingfaReferenceAnchor struct {
	When             string   `json:"when"`
	YinDun           *bool    `json:"yin_dun,omitempty"`
	Ju               *int     `json:"ju,omitempty"`
	XunShou          *string  `json:"xun_shou,omitempty"`
	KongWang         *string  `json:"kong_wang,omitempty"`
	DutyStar         *string  `json:"duty_star,omitempty"`
	DutySource       *int     `json:"duty_source,omitempty"`
	DutyLand         *int     `json:"duty_land,omitempty"`
	DutyDoor         *string  `json:"duty_door,omitempty"`
	DoorLand         *int     `json:"door_land,omitempty"`
	CenterStarPalace *int     `json:"center_star_palace,omitempty"`
	CenterDoorPalace *int     `json:"center_door_palace,omitempty"`
	HiddenPillars    []string `json:"hidden_pillars,omitempty"`
}

func TestMingFaReferenceAnchors(t *testing.T) {
	for _, anchor := range loadMingFaFixture(t).ReferenceAnchors {
		t.Run(anchor.When, func(t *testing.T) {
			when, err := time.ParseInLocation("2006-01-02 15:04", anchor.When, time.FixedZone("CST", 8*3600))
			if err != nil {
				t.Fatal(err)
			}
			chart, err := ComputeChartWithMethod(
				tianwen.SolarTime(when),
				Method{Scope: ScopeHour, School: SchoolMingFaFeiPan, Dingju: DingjuChaiBu},
			)
			if err != nil {
				t.Fatal(err)
			}
			if anchor.YinDun != nil && chart.Pan.YinDun != *anchor.YinDun {
				t.Fatalf("yin dun = %v, want %v", chart.Pan.YinDun, *anchor.YinDun)
			}
			if anchor.Ju != nil && chart.Pan.Jushu != *anchor.Ju {
				t.Fatalf("ju = %d, want %d", chart.Pan.Jushu, *anchor.Ju)
			}
			if anchor.XunShou != nil && chart.Method.LeadXunShou.LiuYi != *anchor.XunShou {
				t.Fatalf("xun shou = %s, want %s", chart.Method.LeadXunShou.LiuYi, *anchor.XunShou)
			}
			if anchor.KongWang != nil {
				got := ganzhi.ZhiName(chart.Pan.KongWang[0].Branch) + ganzhi.ZhiName(chart.Pan.KongWang[1].Branch)
				if got != *anchor.KongWang {
					t.Fatalf("kong wang = %s, want %s", got, *anchor.KongWang)
				}
			}
			if anchor.DutyStar != nil && chart.Pan.DutyStar.String() != *anchor.DutyStar {
				t.Fatalf("duty star = %s, want %s", chart.Pan.DutyStar, *anchor.DutyStar)
			}
			if anchor.DutySource != nil && int(chart.Method.LeadXunShouGong) != *anchor.DutySource {
				t.Fatalf("duty source = %d, want %d", chart.Method.LeadXunShouGong, *anchor.DutySource)
			}
			if anchor.DutyLand != nil && int(chart.Pan.DutyStarPalace) != *anchor.DutyLand {
				t.Fatalf("duty land = %d, want %d", chart.Pan.DutyStarPalace, *anchor.DutyLand)
			}
			if anchor.DutyDoor != nil && chart.Pan.DutyDoor.String() != *anchor.DutyDoor {
				t.Fatalf("duty door = %s, want %s", chart.Pan.DutyDoor, *anchor.DutyDoor)
			}
			if anchor.DoorLand != nil && int(chart.Pan.DutyDoorPalace) != *anchor.DoorLand {
				t.Fatalf("door land = %d, want %d", chart.Pan.DutyDoorPalace, *anchor.DoorLand)
			}
			if anchor.CenterStarPalace != nil {
				palace := chart.Pan.GongWei[*anchor.CenterStarPalace-1]
				if len(palace.TianPan) != 1 || palace.TianPan[0].Star != StarTianQin {
					t.Fatalf("center star palace %d = %+v", *anchor.CenterStarPalace, palace)
				}
			}
			if anchor.CenterDoorPalace != nil {
				palace := chart.Pan.GongWei[*anchor.CenterDoorPalace-1]
				if !palace.DoorSet || palace.Door != DoorZhong {
					t.Fatalf("center door palace %d = %+v", *anchor.CenterDoorPalace, palace)
				}
			}
			for i, want := range anchor.HiddenPillars {
				if got := chart.Pan.GongWei[i].HiddenPillar; got != want {
					t.Fatalf("hidden pillar %d = %q, want %q", i+1, got, want)
				}
			}
		})
	}
}

func TestMingFaCheckedInAnchor(t *testing.T) {
	anchor := loadMingFaFixture(t).Vectors[0]
	when, err := time.ParseInLocation("2006-01-02 15:04", anchor.When, time.FixedZone("CST", 8*3600))
	if err != nil {
		t.Fatal(err)
	}
	method, err := ParseMethod(anchor.Method.Scope, anchor.Method.School, anchor.Method.Dingju, "")
	if err != nil {
		t.Fatal(err)
	}
	chart, err := ComputeChartWithMethod(tianwen.SolarTime(when), method)
	if err != nil {
		t.Fatal(err)
	}
	if chart.Method.Scope != "hour" || chart.Method.School != "mingfa_feipan" ||
		chart.Method.DingjuMethod != "chaibu" || chart.Method.SchoolName != "鸣法飞盘" {
		t.Fatalf("method = %+v", chart.Method)
	}
	if !chart.Pan.YinDun || chart.Pan.Jushu != 8 || chart.Method.Yuan != "下元" {
		t.Fatalf("dingju = %v/%d/%s", chart.Pan.YinDun, chart.Pan.Jushu, chart.Method.Yuan)
	}
	if got := ganzhi.GanName(chart.Pan.LeadGan) + ganzhi.ZhiName(chart.Pan.LeadZhi); got != "癸亥" {
		t.Fatalf("lead pillar = %s", got)
	}
	if chart.Method.LeadXunShou.LiuYi != "癸" || chart.Pan.KongWang[0].Branch == 0 ||
		chart.Pan.KongWang[1].Branch == 0 {
		t.Fatalf("xun/void = %+v/%+v", chart.Method.LeadXunShou, chart.Pan.KongWang)
	}
	if chart.Pan.DutyStar != StarTianChong || chart.Method.LeadXunShouGong != GongZhen ||
		chart.Pan.DutyStarPalace != GongZhen || chart.Pan.DutyDoor != DoorShang ||
		chart.Pan.DutyDoorPalace != GongZhen {
		t.Fatalf("duty = %s/%d %s/%d", chart.Pan.DutyStar, chart.Pan.DutyStarPalace,
			chart.Pan.DutyDoor, chart.Pan.DutyDoorPalace)
	}

	wants := [9]mingfaPalaceWant(anchor.Palaces)
	stars, doors, spirits := 0, 0, 0
	for i, want := range wants {
		palace := chart.Pan.GongWei[i]
		stars += len(palace.TianPan)
		if palace.DoorSet {
			doors++
		}
		if palace.SpiritSet {
			spirits++
		}
		if len(palace.TianPan) != 1 || palace.TianPan[0].Gan.String() != want.Heaven ||
			palace.TianPan[0].Star.String() != want.Star || palace.DiPanGan.String() != want.Earth ||
			palace.Door.String() != want.Door || spiritDisplayName(palace.Spirit, true, SchoolMingFaFeiPan) != want.Spirit ||
			palace.HiddenPillar != want.HiddenPillar || palace.AnGan.String() != string([]rune(want.HiddenPillar)[0]) {
			t.Fatalf("palace %d = %+v, want %+v", i+1, palace, want)
		}
	}
	if stars != 9 || doors != 9 || spirits != 9 {
		t.Fatalf("layer counts = %d/%d/%d, want 9/9/9", stars, doors, spirits)
	}
	if chart.Method.TianQin != nil {
		t.Fatal("Ming Fa must not expose rotating TianQin lodging metadata")
	}
}

func TestMingFaCenterDoorBehavior(t *testing.T) {
	if doorWuxing(DoorZhong) != ganzhi.WxTu {
		t.Fatalf("center door element = %v, want earth", doorWuxing(DoorZhong))
	}
	if !menPo(DoorZhong, GongKan) {
		t.Fatal("center door over water palace should be door-oppressed")
	}
	if !menZhi(DoorZhong, GongZhen) {
		t.Fatal("wood palace over center door should be door-controlled")
	}
	for key := range menGongTable {
		if key[0] == int(DoorZhong) {
			t.Fatal("center door must not receive an invented men-gong interpretation")
		}
	}
}
