package qimen

import (
	"encoding/json"
	"fmt"
	"os"
	"testing"
	"time"

	"liki-engine/internal/engine/ganzhi"
	"liki-engine/internal/engine/tianwen"
)

type quarterGoldenVector struct {
	ID           string `json:"id"`
	QuarterRule  string `json:"quarter_rule"`
	BaseDingju   string `json:"base_dingju_method"`
	When         string `json:"when"`
	Hour         string `json:"hour"`
	QuarterIndex int    `json:"quarter_index"`
	LeadMode     string `json:"lead_pillar_mode"`
	DunSource    string `json:"dun_source"`
	HourBoundary string `json:"hour_boundary"`
	BaseJu       int    `json:"base_ju"`
	Lead         string `json:"lead"`
	YinDun       bool   `json:"yin_dun"`
	Ju           int    `json:"ju"`
	Yuan         string `json:"yuan"`
	XunShou      string `json:"xun_shou"`
	DutyStar     string `json:"duty_star"`
	DutyDoor     string `json:"duty_door"`
}

type quarterGoldenFixture struct {
	Source struct {
		Name       string `json:"name"`
		URL        string `json:"url"`
		Artifact   string `json:"artifact"`
		SHA256     string `json:"sha256"`
		License    string `json:"license"`
		Usage      string `json:"usage"`
		RuleSource string `json:"rule_source"`
	} `json:"source"`
	Vectors []quarterGoldenVector `json:"vectors"`
}

func TestQuarterCatalogIsClosedToTwoMethods(t *testing.T) {
	method, err := ResolveMethod("quarter", "zhuanpan", "", "")
	if err != nil {
		t.Fatal(err)
	}
	want := Method{
		Scope: ScopeQuarter, School: SchoolZhuanPan,
		QuarterRule: QuarterTenMinuteSanYuan,
	}
	if method != want {
		t.Fatalf("quarter method = %+v, want %+v", method, want)
	}
	method, err = ResolveMethod(
		"quarter", "zhuanpan", "", "twelve_minute_ten_division",
	)
	if err != nil {
		t.Fatal(err)
	}
	want = Method{
		Scope: ScopeQuarter, School: SchoolZhuanPan,
		QuarterRule: QuarterTwelveMinuteTenDivision, BaseDingju: DingjuChaiBu,
		DunSource: "hour_branch", HourBoundary: "late_zi",
	}
	if method != want {
		t.Fatalf("十二分钟十分局 quarter method = %+v, want %+v", method, want)
	}
	for method, wantBase := range map[string]DingjuMethod{
		"zhirun": DingjuZhiRun, "maoshan": DingjuMaoShan,
	} {
		got, err := ParseMethodWithOptions(
			"quarter", "zhuanpan", "", "twelve_minute_ten_division", method, "", "",
		)
		if err != nil {
			t.Fatal(err)
		}
		want := Method{
			Scope: ScopeQuarter, School: SchoolZhuanPan,
			QuarterRule: QuarterTwelveMinuteTenDivision, BaseDingju: wantBase,
			DunSource: "hour_branch", HourBoundary: "late_zi",
		}
		if got != want {
			t.Fatalf("twelve-minute base %s method = %+v, want %+v", method, got, want)
		}
	}
	for _, item := range []struct {
		dun, boundary string
	}{
		{"hour_branch", "late_zi"},
		{"solar_term", "zi_zheng"},
	} {
		method, err := ParseMethodWithOptions(
			"quarter", "zhuanpan", "", "twelve_minute_ten_division",
			"maoshan", item.dun, item.boundary,
		)
		if err != nil {
			t.Fatal(err)
		}
		want := Method{
			Scope: ScopeQuarter, School: SchoolZhuanPan,
			QuarterRule: QuarterTwelveMinuteTenDivision, BaseDingju: DingjuMaoShan,
			DunSource: item.dun, HourBoundary: item.boundary,
		}
		if method != want {
			t.Fatalf("十二分钟十分局 options %s/%s = %+v, want %+v", item.dun, item.boundary, method, want)
		}
	}
	for _, item := range [][5]string{
		{"quarter", "zhuanpan", "", "twelve_minute_ten_division", "ten_minute_sanyuan"},
		{"hour", "zhuanpan", "chaibu", "", "zhirun"},
		{"quarter", "zhuanpan", "ten_minute_sanyuan", "", "chaibu"},
	} {
		if _, err := ParseMethodWithOptions(item[0], item[1], item[2], item[3], item[4], "", ""); err == nil {
			t.Errorf("ParseMethodWithOptions(%v) should reject", item)
		}
	}
	for _, item := range [][7]string{
		{"quarter", "zhuanpan", "", "twelve_minute_ten_division", "", "jieqi", ""},
		{"quarter", "zhuanpan", "", "twelve_minute_ten_division", "", "", "midnight"},
		{"hour", "zhuanpan", "chaibu", "", "", "solar_term", ""},
		{"hour", "zhuanpan", "chaibu", "", "", "", "zi_zheng"},
	} {
		if _, err := ParseMethodWithOptions(item[0], item[1], item[2], item[3], item[4], item[5], item[6]); err == nil {
			t.Errorf("ParseMethodWithOptions(%v) should reject", item)
		}
	}
	for _, item := range [][3]string{
		{"quarter", "zhuanpan", "chaibu"},
		{"quarter", "zhuanpan", "zhirun"},
		{"quarter", "zhuanpan", "maoshan"},
		{"quarter", "luoshu_feipan", "ten_minute_sanyuan"},
		{"quarter", "mingfa_feipan", "ten_minute_sanyuan"},
		{"quarter", "luoshu_feipan", "twelve_minute_ten_division"},
		{"quarter", "mingfa_feipan", "twelve_minute_ten_division"},
		{"quarter", "jinhan_yujing", ""},
		{"hour", "zhuanpan", "ten_minute_sanyuan"},
		{"hour", "zhuanpan", "twelve_minute_ten_division"},
	} {
		if _, err := ParseMethod(item[0], item[1], item[2], ""); err == nil {
			t.Errorf("ParseMethod(%q,%q,%q) should reject", item[0], item[1], item[2])
		}
	}
}

func TestTwelveMinuteTenDivisionQuarterBoundaries(t *testing.T) {
	cases := []struct {
		hour, minute int
		index        int
	}{
		{23, 0, 1},
		{23, 11, 1},
		{23, 12, 2},
		{0, 0, 6},
		{0, 59, 10},
		{23, 107, 9},
		{23, 108, 10},
		{23, 119, 10},
		{1, 0, 1},
	}
	for _, tc := range cases {
		when := time.Date(2026, 9, 7, tc.hour, tc.minute, 0, 0, time.FixedZone("CST", 8*3600))
		if got := twelveMinuteQuarterIndex(when); got != tc.index {
			t.Fatalf("twelveMinuteQuarterIndex(%02d:%02d) = %d, want %d", tc.hour, tc.minute, got, tc.index)
		}
	}
}

func TestTwelveMinuteTenDivisionHourBoundaryOptions(t *testing.T) {
	when := time.Date(2026, 9, 7, 0, 0, 0, 0, time.FixedZone("CST", 8*3600))
	if got := twelveMinuteQuarterIndexWithBoundary(when, "late_zi"); got != 6 {
		t.Fatalf("late-zi quarter = %d, want 6", got)
	}
	if got := twelveMinuteQuarterIndexWithBoundary(when, "zi_zheng"); got != 1 {
		t.Fatalf("zi-zheng quarter = %d, want 1", got)
	}
}

func TestTwelveMinuteTenDivisionDunSourceOptions(t *testing.T) {
	when := time.Date(2026, 9, 7, 0, 0, 0, 0, time.FixedZone("CST", 8*3600))
	tests := []struct {
		dun      string
		boundary string
		yin      bool
		ju       int
	}{
		{"hour_branch", "late_zi", false, 9},
		{"solar_term", "late_zi", true, 8},
		{"hour_branch", "zi_zheng", false, 4},
		{"solar_term", "zi_zheng", true, 4},
	}
	for _, want := range tests {
		method, err := ParseMethodWithOptions(
			"quarter", "zhuanpan", "", "twelve_minute_ten_division",
			"chaibu", want.dun, want.boundary,
		)
		if err != nil {
			t.Fatal(err)
		}
		chart, err := ComputeChartWithMethod(tianwen.SolarTime(when), method)
		if err != nil {
			t.Fatal(err)
		}
		if chart.Pan.YinDun != want.yin || chart.Pan.Jushu != want.ju ||
			chart.Method.Quarter.DunSource != want.dun ||
			chart.Method.Quarter.HourBoundary != want.boundary {
			t.Fatalf("options %s/%s = %d/%v/%+v", want.dun, want.boundary,
				chart.Pan.Jushu, chart.Pan.YinDun, chart.Method.Quarter)
		}
	}
}

func TestTwelveMinuteTenDivisionJuShift(t *testing.T) {
	wantYang := []int{1, 2, 3, 4, 5, 6, 7, 8, 9, 1}
	wantYin := []int{9, 8, 7, 6, 5, 4, 3, 2, 1, 9}
	for i := range wantYang {
		if got := twelveMinuteShiftJu(1, i+1, true); got != wantYang[i] {
			t.Fatalf("yang shift quarter %d = %d, want %d", i+1, got, wantYang[i])
		}
		if got := twelveMinuteShiftJu(9, i+1, false); got != wantYin[i] {
			t.Fatalf("yin shift quarter %d = %d, want %d", i+1, got, wantYin[i])
		}
	}
}

func TestQuarterPillarFiveRatEscape(t *testing.T) {
	cases := []struct {
		hour string
		want [12]string
	}{
		{"甲子", [12]string{"甲子", "乙丑", "丙寅", "丁卯", "戊辰", "己巳", "庚午", "辛未", "壬申", "癸酉", "甲戌", "乙亥"}},
		{"乙丑", [12]string{"丙子", "丁丑", "戊寅", "己卯", "庚辰", "辛巳", "壬午", "癸未", "甲申", "乙酉", "丙戌", "丁亥"}},
		{"癸酉", [12]string{"壬子", "癸丑", "甲寅", "乙卯", "丙辰", "丁巳", "戊午", "己未", "庚申", "辛酉", "壬戌", "癸亥"}},
	}
	for _, tc := range cases {
		hourPillar, err := parsePillarForTest(tc.hour)
		if err != nil {
			t.Fatal(err)
		}
		for i := range tc.want {
			pillar := quarterPillarAt(hourPillar, i+1)
			got := ganzhi.GanName(pillar.Gan) + ganzhi.ZhiName(pillar.Zhi)
			if got != tc.want[i] {
				t.Fatalf("hour %s quarter %d = %s, want %s", tc.hour, i+1, got, tc.want[i])
			}
		}
	}
}

func parsePillarForTest(value string) (ganzhi.Zhu, error) {
	runes := []rune(value)
	if len(runes) != 2 {
		return ganzhi.Zhu{}, fmt.Errorf("invalid pillar %q", value)
	}
	gan, err := ganzhi.ParseGan(string(runes[0]))
	if err != nil {
		return ganzhi.Zhu{}, err
	}
	zhi, err := ganzhi.ParseZhi(string(runes[1]))
	if err != nil {
		return ganzhi.Zhu{}, err
	}
	return ganzhi.Zhu{Gan: gan, Zhi: zhi}, nil
}

func TestQuarterBoundaries(t *testing.T) {
	cases := []struct {
		hour, minute int
		index        int
	}{
		{23, 0, 1},
		{23, 9, 1},
		{23, 10, 2},
		{0, 0, 7},
		{0, 50, 12},
		{0, 59, 12},
		{1, 0, 1},
		{1, 119, 12},
	}
	for _, tc := range cases {
		if got := quarterIndex(time.Date(2026, 9, 7, tc.hour, tc.minute, 0, 0, time.FixedZone("CST", 8*3600))); got != tc.index {
			t.Fatalf("quarterIndex(%02d:%02d) = %d, want %d", tc.hour, tc.minute, got, tc.index)
		}
	}
}

func TestTenMinuteQuarterGoldenVectors(t *testing.T) {
	var fixture quarterGoldenFixture
	data, err := os.ReadFile("testdata/quarter_golden.json")
	if err != nil {
		t.Fatal(err)
	}
	if err := json.Unmarshal(data, &fixture); err != nil {
		t.Fatal(err)
	}
	if fixture.Source.Name == "" || fixture.Source.SHA256 == "" || len(fixture.Vectors) == 0 {
		t.Fatal("quarter golden provenance is incomplete")
	}
	for _, want := range fixture.Vectors {
		t.Run(want.ID, func(t *testing.T) {
			when, err := time.Parse(time.RFC3339, want.When)
			if err != nil {
				t.Fatal(err)
			}
			chart, err := ComputeChartWithMethod(tianwen.SolarTime(when), Method{
				Scope: ScopeQuarter, School: SchoolZhuanPan, QuarterRule: QuarterTenMinuteSanYuan,
			})
			if err != nil {
				t.Fatal(err)
			}
			if got := ganzhi.GanName(chart.Pan.HourGan) + ganzhi.ZhiName(chart.Pan.HourZhi); got != want.Hour {
				t.Fatalf("hour pillar = %s, want %s", got, want.Hour)
			}
			if got := chart.Method.LeadPillar.Name; got != want.Lead {
				t.Fatalf("lead = %s, want %s", got, want.Lead)
			}
			if chart.Method.Quarter == nil || chart.Method.Quarter.Index != want.QuarterIndex ||
				chart.Method.Quarter.MinutesPerQuarter != 10 || chart.Method.Quarter.LeadPillarMode != "quarter" {
				t.Fatalf("quarter metadata = %+v, want index %d/10min", chart.Method.Quarter, want.QuarterIndex)
			}
			if chart.Pan.Jushu != want.Ju || chart.Pan.YinDun != want.YinDun || chart.Method.Yuan != want.Yuan {
				t.Fatalf("dingju = %d/%v/%s, want %d/%v/%s",
					chart.Pan.Jushu, chart.Pan.YinDun, chart.Method.Yuan,
					want.Ju, want.YinDun, want.Yuan)
			}
			if got := chart.Method.LeadXunShou.Name; got != want.XunShou {
				t.Fatalf("lead xun shou = %s, want %s", got, want.XunShou)
			}
			if got := chart.Pan.DutyStar.String(); got != want.DutyStar {
				t.Fatalf("duty star = %s, want %s", got, want.DutyStar)
			}
			if got := chart.Pan.DutyDoor.String(); got != want.DutyDoor {
				t.Fatalf("duty door = %s, want %s", got, want.DutyDoor)
			}
		})
	}
}

func TestTwelveMinuteTenDivisionQuarterGoldenVectors(t *testing.T) {
	var fixture quarterGoldenFixture
	data, err := os.ReadFile("testdata/twelve_minute_quarter_golden.json")
	if err != nil {
		t.Fatal(err)
	}
	if err := json.Unmarshal(data, &fixture); err != nil {
		t.Fatal(err)
	}
	if fixture.Source.Name == "" || fixture.Source.URL == "" || fixture.Source.SHA256 == "" ||
		fixture.Source.License != "AGPL-3.0; execution facts only" || len(fixture.Vectors) == 0 {
		t.Fatal("十二分钟十分局 quarter golden provenance is incomplete")
	}
	for _, want := range fixture.Vectors {
		t.Run(want.ID, func(t *testing.T) {
			when, err := time.Parse(time.RFC3339, want.When)
			if err != nil {
				t.Fatal(err)
			}
			base, err := ComputeChartWithMethod(tianwen.SolarTime(when), Method{
				Scope: ScopeHour, School: SchoolZhuanPan,
				Dingju: mustDingjuForTest(t, want.BaseDingju),
			})
			if err != nil {
				t.Fatal(err)
			}
			if base.Pan.Jushu != want.BaseJu {
				t.Fatalf("base hour ju = %d, want %d", base.Pan.Jushu, want.BaseJu)
			}
			chart, err := ComputeChartWithMethod(tianwen.SolarTime(when), Method{
				Scope: ScopeQuarter, School: SchoolZhuanPan,
				QuarterRule: QuarterTwelveMinuteTenDivision, BaseDingju: mustDingjuForTest(t, want.BaseDingju),
				DunSource: want.DunSource, HourBoundary: want.HourBoundary,
			})
			if err != nil {
				t.Fatal(err)
			}
			if got := ganzhi.GanName(chart.Pan.HourGan) + ganzhi.ZhiName(chart.Pan.HourZhi); got != want.Hour {
				t.Fatalf("hour pillar = %s, want %s", got, want.Hour)
			}
			if chart.Method.Quarter == nil || chart.Method.Quarter.Index != want.QuarterIndex ||
				chart.Method.Quarter.MinutesPerQuarter != 12 ||
				chart.Method.Quarter.LeadPillarMode != "hour" ||
				chart.Method.Quarter.DunSource != want.DunSource ||
				chart.Method.Quarter.HourBoundary != want.HourBoundary {
				t.Fatalf("quarter metadata = %+v, want index %d/12min/hour lead", chart.Method.Quarter, want.QuarterIndex)
			}
			if got := chart.Method.LeadPillar.Name; got != want.Hour || got != want.Lead {
				t.Fatalf("lead = %s, want 十二分钟十分局 hour anchor %s", got, want.Lead)
			}
			if chart.Pan.Jushu != want.Ju || chart.Pan.YinDun != want.YinDun || chart.Method.Yuan != want.Yuan {
				t.Fatalf("dingju = %d/%v/%s, want %d/%v/%s",
					chart.Pan.Jushu, chart.Pan.YinDun, chart.Method.Yuan,
					want.Ju, want.YinDun, want.Yuan)
			}
		})
	}
}

func mustDingjuForTest(t *testing.T, value string) DingjuMethod {
	t.Helper()
	method, err := ParseDingjuMethod(value)
	if err != nil {
		t.Fatal(err)
	}
	return method
}
