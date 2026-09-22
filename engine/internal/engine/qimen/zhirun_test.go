package qimen

import (
	"testing"
	"time"

	"liki-engine/internal/engine/ganzhi"
	"liki-engine/internal/engine/tianwen"
)

func TestParseDingjuMethod(t *testing.T) {
	valid := map[string]DingjuMethod{
		"chaibu": DingjuChaiBu,
		"zhirun": DingjuZhiRun,
	}
	for input, want := range valid {
		got, err := ParseDingjuMethod(input)
		if err != nil || got != want {
			t.Fatalf("ParseDingjuMethod(%q) = %v, %v; want %v", input, got, err, want)
		}
	}
	for _, input := range []string{"", "luoshu_feipan", "SHIRUN", "day"} {
		if _, err := ParseDingjuMethod(input); err == nil {
			t.Errorf("ParseDingjuMethod(%q) should reject", input)
		}
	}
}

func TestZhiRunTableContract(t *testing.T) {
	if dingjuMethodNames[DingjuZhiRun] != "置闰定局" {
		t.Fatalf("zhirun method name = %q", dingjuMethodNames[DingjuZhiRun])
	}
	if zhiRunLeaderDays != 15 || zhiRunThreshold != 7 || zhiRunMaxIntervals != 12 {
		t.Fatalf("grid = %d/%d/%d, want 15/7/12", zhiRunLeaderDays, zhiRunThreshold, zhiRunMaxIntervals)
	}
	if !zhiRunRepeatTerms["芒种"] || !zhiRunRepeatTerms["大雪"] {
		t.Fatal("only Mangzhong and Daxue may repeat")
	}
	for _, basis := range zhiRunBasis {
		if basis == "" {
			t.Fatal("zhirun table basis is incomplete")
		}
	}
	for _, name := range []string{"正授", "超神", "接气", "置闰"} {
		found := false
		for _, state := range zhiRunStateNames {
			if state == name {
				found = true
			}
		}
		if !found {
			t.Fatalf("state %q missing", name)
		}
	}
}

func TestComputeChartDefaultIsChaiBu(t *testing.T) {
	chart := zhiRunChartForTest(t, "2026-01-05 13:00", DingjuChaiBu)
	if chart.Method.DingjuMethod != "chaibu" || chart.Method.DingjuMethodName != "拆补定局" {
		t.Fatalf("default method = %s/%s, want chaibu/拆补定局", chart.Method.DingjuMethod, chart.Method.DingjuMethodName)
	}
	if chart.Method.JieQi != "冬至" || chart.Method.YongJuJieQi != "冬至" || chart.Pan.Jushu != 1 {
		t.Fatalf("chai bu method = %s/%s/%d, want 冬至/冬至/1", chart.Method.JieQi, chart.Method.YongJuJieQi, chart.Pan.Jushu)
	}
	if chart.Method.ZhiRunState != "" {
		t.Fatalf("chai bu zhi_run_state = %q, want empty", chart.Method.ZhiRunState)
	}
}

type zhiRunDingjuAnchor struct {
	Provenance  string `json:"provenance"`
	When        string `json:"when"`
	JieQi       string `json:"jie_qi"`
	YongJuJieQi string `json:"yong_ju_jie_qi"`
	YinDun      bool   `json:"yin_dun"`
	Ju          int    `json:"ju"`
	Yuan        string `json:"yuan"`
}

type zhiRunStateAnchor struct {
	Provenance string `json:"provenance"`
	When       string `json:"when"`
	State      string `json:"state"`
}

type zhiRunDayBoundaryAnchor struct {
	Provenance string `json:"provenance"`
	When       string `json:"when"`
	Day        string `json:"day"`
	FuTou      string `json:"fu_tou"`
}

type zhiRunGoldenFixture struct {
	Source             goldenSource              `json:"source"`
	DingjuVectors      []zhiRunDingjuAnchor      `json:"dingju_vectors"`
	StateVectors       []zhiRunStateAnchor       `json:"state_vectors"`
	DayBoundaryVectors []zhiRunDayBoundaryAnchor `json:"day_boundary_vectors"`
}

func loadZhiRunGoldenFixture(t *testing.T) zhiRunGoldenFixture {
	t.Helper()
	var fixture zhiRunGoldenFixture
	loadGoldenFixture(t, "zhirun_golden.json", &fixture)
	if fixture.Source != atopxGoldenSource() || len(fixture.DingjuVectors) != 16 ||
		len(fixture.StateVectors) != 6 || len(fixture.DayBoundaryVectors) != 2 {
		t.Fatalf("zhirun golden dimensions = %+v/%d/%d/%d", fixture.Source,
			len(fixture.DingjuVectors), len(fixture.StateVectors), len(fixture.DayBoundaryVectors))
	}
	for _, vector := range fixture.DingjuVectors {
		if vector.Provenance != "atopx_checked_in" {
			t.Fatalf("invalid zhirun dingju provenance %+v", vector)
		}
	}
	for _, vector := range fixture.StateVectors {
		if vector.Provenance != "atopx_reference_note" {
			t.Fatalf("invalid zhirun state provenance %+v", vector)
		}
	}
	for _, vector := range fixture.DayBoundaryVectors {
		if vector.Provenance != "atopx_checked_in" {
			t.Fatalf("invalid zhirun day-boundary provenance %+v", vector)
		}
	}
	return fixture
}

func TestZhiRunExternalDingjuAnchors(t *testing.T) {
	for _, anchor := range loadZhiRunGoldenFixture(t).DingjuVectors {
		t.Run(anchor.When, func(t *testing.T) {
			chart := zhiRunChartForTest(t, anchor.When, DingjuZhiRun)
			if chart.Method.DingjuMethod != "zhirun" || chart.Method.DingjuMethodName != "置闰定局" {
				t.Fatalf("method = %s/%s, want zhirun/置闰定局", chart.Method.DingjuMethod, chart.Method.DingjuMethodName)
			}
			if chart.Method.JieQi != anchor.JieQi {
				t.Errorf("actual term = %s, want %s", chart.Method.JieQi, anchor.JieQi)
			}
			if chart.Method.YongJuJieQi != anchor.YongJuJieQi {
				t.Errorf("dingju term = %s, want %s", chart.Method.YongJuJieQi, anchor.YongJuJieQi)
			}
			if chart.Pan.YinDun != anchor.YinDun || chart.Pan.Jushu != anchor.Ju || chart.Method.Yuan != anchor.Yuan {
				t.Errorf("dun/ju/yuan = %v/%d/%s, want %v/%d/%s",
					chart.Pan.YinDun, chart.Pan.Jushu, chart.Method.Yuan,
					anchor.YinDun, anchor.Ju, anchor.Yuan)
			}
		})
	}
}

func TestZhiRunStateAnchors(t *testing.T) {
	for _, anchor := range loadZhiRunGoldenFixture(t).StateVectors {
		t.Run(anchor.When+"/"+anchor.State, func(t *testing.T) {
			chart := zhiRunChartForTest(t, anchor.When, DingjuZhiRun)
			if chart.Method.ZhiRunState != anchor.State {
				t.Fatalf("state = %q, want %q", chart.Method.ZhiRunState, anchor.State)
			}
		})
	}
}

func TestLateZiRollsQimenDayPillar(t *testing.T) {
	for _, anchor := range loadZhiRunGoldenFixture(t).DayBoundaryVectors {
		t.Run(anchor.When, func(t *testing.T) {
			chart := zhiRunChartForTest(t, anchor.When, DingjuZhiRun)
			day := ganzhi.GanName(chart.Pan.RiGan) + ganzhi.ZhiName(chart.Pan.RiZhi)
			if day != anchor.Day {
				t.Fatalf("day pillar = %s, want %s", day, anchor.Day)
			}
			if chart.Method.DayFuTou.Name != anchor.FuTou {
				t.Fatalf("day fu tou = %s, want %s", chart.Method.DayFuTou.Name, anchor.FuTou)
			}
			if chart.Method.DayBoundary != "late_zi_rolls_day_pillar" {
				t.Fatalf("day boundary = %s", chart.Method.DayBoundary)
			}
		})
	}
}

func TestZhiRunExternalRotatingPlateAnchor(t *testing.T) {
	chart := zhiRunChartForTest(t, "2026-01-05 13:00", DingjuZhiRun)
	if chart.Pan.DutyStar != StarTianRui || chart.Pan.DutyDoor != DoorSi {
		t.Fatalf("duty = %s/%s, want 天芮/死门", chart.Pan.DutyStar, chart.Pan.DutyDoor)
	}
	if chart.DutyStarPalace != GongKun || chart.DutyDoorPalace != GongLi {
		t.Fatalf("duty palaces = %s/%s, want 坤/离", chart.DutyStarPalace, chart.DutyDoorPalace)
	}
	assertExternalPalaces(t, chart, [9]string{
		"乙乙丁蓬生合", "戊戊癸芮惊符", "己己庚冲杜玄",
		"庚庚丙辅景地", "辛辛辛---", "壬壬乙心休阴",
		"癸癸壬柱开蛇", "丁丁己任伤虎", "丙丙戊英死天",
	})
}

func TestComputeChartWithYongShenAndMethod(t *testing.T) {
	symbol, err := ParseYongShen("生门")
	if err != nil {
		t.Fatal(err)
	}
	chart, err := ComputeChartWithYongShenAndMethod(
		zhiRunTimeForTest(t, "2026-01-05 13:00"), []YongShenSymbol{symbol},
		BirthDate{}, Method{Scope: ScopeHour, School: SchoolZhuanPan, Dingju: DingjuZhiRun},
	)
	if err != nil {
		t.Fatal(err)
	}
	if chart.Method.DingjuMethod != "zhirun" || chart.YongShen == nil || len(chart.YongShen.Symbols) != 1 {
		t.Fatalf("chart method/yong shen = %s/%+v", chart.Method.DingjuMethod, chart.YongShen)
	}
}

func zhiRunChartForTest(t *testing.T, when string, method DingjuMethod) Chart {
	t.Helper()
	civil, err := time.ParseInLocation("2006-01-02 15:04", when, time.FixedZone("CST", 8*3600))
	if err != nil {
		t.Fatal(err)
	}
	chart, err := ComputeChartWithMethod(
		tianwen.SolarTime(civil),
		Method{Scope: ScopeHour, School: SchoolZhuanPan, Dingju: method},
	)
	if err != nil {
		t.Fatal(err)
	}
	return chart
}

func zhiRunTimeForTest(t *testing.T, when string) (st tianwen.SolarTime) {
	t.Helper()
	civil, err := time.ParseInLocation("2006-01-02 15:04", when, time.FixedZone("CST", 8*3600))
	if err != nil {
		t.Fatal(err)
	}
	return tianwen.SolarTime(civil)
}
