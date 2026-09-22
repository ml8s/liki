package qimen

import (
	"testing"
	"time"

	"liki-engine/internal/engine/tianwen"
)

func TestMaoShanCatalogAndResolver(t *testing.T) {
	if DingjuMaoShan.String() != "maoshan" {
		t.Fatalf("Mao Shan dingju method = %s", DingjuMaoShan)
	}
	got, err := ResolveMethod("hour", "zhuanpan", "maoshan", "")
	if err != nil {
		t.Fatal(err)
	}
	want := Method{Scope: ScopeHour, School: SchoolZhuanPan, Dingju: DingjuMaoShan}
	if got != want {
		t.Fatalf("Mao Shan method = %+v, want %+v", got, want)
	}
	invalid := [][3]string{
		{"day", "zhuanpan", "maoshan"},
		{"month", "zhuanpan", "maoshan"},
		{"year", "zhuanpan", "maoshan"},
		{"hour", "luoshu_feipan", "maoshan"},
		{"hour", "mingfa_feipan", "maoshan"},
	}
	for _, item := range invalid {
		if _, err := ResolveMethod(item[0], item[1], item[2], ""); err == nil {
			t.Fatalf("ResolveMethod(%+v) should reject", item)
		}
	}
}

type maoShanVector struct {
	Provenance  string `json:"provenance"`
	When        string `json:"when"`
	ActualJieQi string `json:"actual_jie_qi"`
	YongJuJieQi string `json:"yong_ju_jie_qi"`
	YinDun      bool   `json:"yin_dun"`
	Ju          int    `json:"ju"`
	Yuan        string `json:"yuan"`
}

type maoShanFixture struct {
	Source  goldenSource    `json:"source"`
	Rule    map[string]any  `json:"rule"`
	Vectors []maoShanVector `json:"vectors"`
}

func TestMaoShanDingjuAnchors(t *testing.T) {
	var fixture maoShanFixture
	loadGoldenFixture(t, "maoshan_golden.json", &fixture)
	wantSource := goldenSource{
		Name: "deminzhang/qimen-go", URL: "https://github.com/deminzhang/qimen-go",
		Commit:     "4d3f58fa0f401b5b3a337f119138e99e90685dda",
		SourceFile: "xuan/qimen.go", SourceSHA256: "93ba61a632af3d402bfd65754d1adeb74b937f87b40e62fc50adc15a1d4bd350",
		License: "MIT", LicenseSHA256: "cfdb13c1a086ecc3f498b6a75a65d1dab3d93072c58efd26d383db776749b996",
	}
	if fixture.Source != wantSource || len(fixture.Vectors) != 7 {
		t.Fatalf("Mao Shan fixture = %+v/%d", fixture.Source, len(fixture.Vectors))
	}
	for _, anchor := range fixture.Vectors {
		t.Run(anchor.When, func(t *testing.T) {
			when, err := time.ParseInLocation("2006-01-02 15:04", anchor.When, time.FixedZone("CST", 8*3600))
			if err != nil {
				t.Fatal(err)
			}
			chart, err := ComputeChartWithMethod(
				tianwen.SolarTime(when),
				Method{Scope: ScopeHour, School: SchoolZhuanPan, Dingju: DingjuMaoShan},
			)
			if err != nil {
				t.Fatal(err)
			}
			if chart.Method.DingjuMethod != "maoshan" || chart.Method.DingjuMethodName != "茅山定局" ||
				chart.Method.JieQi != anchor.ActualJieQi || chart.Method.YongJuJieQi != anchor.YongJuJieQi ||
				chart.Pan.YinDun != anchor.YinDun || chart.Pan.Jushu != anchor.Ju || chart.Method.Yuan != anchor.Yuan {
				t.Fatalf("dingju = %s/%s %s/%s %v/%d/%s; want %+v",
					chart.Method.DingjuMethod, chart.Method.DingjuMethodName, chart.Method.JieQi,
					chart.Method.YongJuJieQi, chart.Pan.YinDun, chart.Pan.Jushu, chart.Method.Yuan, anchor)
			}
		})
	}
}
