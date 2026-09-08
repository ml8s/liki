package qimen

import (
	"testing"
	"time"

	"liki-engine/internal/engine/ganzhi"
	"liki-engine/internal/engine/tianwen"
)

// Method-matrix anchors come from atopx/qimen at commit
// 8eb06d007d4a5fcc5352d9054f81469e5f023f45 (MIT). Combinations absent from
// that project's checked-in golden table are fixed vectors generated from the
// same independent implementation. Six of those vectors are additionally
// cross-checked against independent implementations over the facts for which
// the projects share the same domain convention. The yin-dun day-fly zhirun
// core remains single-sourced; the yang repeat-state vector is partially checked
// over the facts shared with mingyu.

func TestParseMethodMatrix(t *testing.T) {
	validMatrix := []Method{
		{Scope: ScopeHour, School: SchoolZhuanPan, Dingju: DingjuChaiBu},
		{Scope: ScopeHour, School: SchoolZhuanPan, Dingju: DingjuZhiRun},
		{Scope: ScopeHour, School: SchoolLuoShuFeiPan, Dingju: DingjuChaiBu},
		{Scope: ScopeHour, School: SchoolLuoShuFeiPan, Dingju: DingjuZhiRun},
		{Scope: ScopeDay, School: SchoolZhuanPan, Dingju: DingjuChaiBu},
		{Scope: ScopeDay, School: SchoolZhuanPan, Dingju: DingjuZhiRun},
		{Scope: ScopeDay, School: SchoolLuoShuFeiPan, Dingju: DingjuChaiBu},
		{Scope: ScopeDay, School: SchoolLuoShuFeiPan, Dingju: DingjuZhiRun},
		{Scope: ScopeMonth, School: SchoolZhuanPan},
		{Scope: ScopeMonth, School: SchoolLuoShuFeiPan},
		{Scope: ScopeYear, School: SchoolZhuanPan},
		{Scope: ScopeYear, School: SchoolLuoShuFeiPan},
	}
	for _, method := range validMatrix {
		tc := method
		got, err := ParseMethod(tc.Scope.String(), tc.School.String(), tc.Dingju.String(), "")
		if err != nil {
			t.Fatalf("ParseMethod(%q,%q,%q): %v", tc.Scope, tc.School, tc.Dingju, err)
		}
		if got != tc {
			t.Fatalf("ParseMethod(%q,%q,%q) = %+v, want %+v", tc.Scope, tc.School, tc.Dingju, got, tc)
		}
	}
	invalid := [][3]string{
		{"ke", "zhuanpan", ""},
		{"hour", "yinpan", ""},
		{"year", "zhuanpan", "chaibu"},
		{"", "zhuanpan", ""},
	}
	for _, tc := range invalid {
		if _, err := ParseMethod(tc[0], tc[1], tc[2], ""); err == nil {
			t.Errorf("ParseMethod(%q,%q,%q) should reject", tc[0], tc[1], tc[2])
		}
	}
	if _, err := ComputeChartWithMethod(
		tianwen.SolarTime{}, Method{Scope: ScopeYear, School: SchoolZhuanPan, Dingju: DingjuChaiBu},
	); err == nil {
		t.Error("ComputeChartWithMethod should reject an unsupported method combination")
	}
}

func TestResolveMethodDefaults(t *testing.T) {
	got, err := ResolveMethod("", "", "", "")
	if err != nil {
		t.Fatal(err)
	}
	want := Method{Scope: ScopeHour, School: SchoolZhuanPan, Dingju: DingjuChaiBu}
	if got != want {
		t.Fatalf("default method = %+v, want %+v", got, want)
	}
	got, err = ResolveMethod("day", "luoshu_feipan", "", "")
	if err != nil {
		t.Fatal(err)
	}
	want = Method{Scope: ScopeDay, School: SchoolLuoShuFeiPan, Dingju: DingjuChaiBu}
	if got != want {
		t.Fatalf("day fly default = %+v, want %+v", got, want)
	}
	if _, err := ResolveMethod("month", "", "chaibu", ""); err == nil {
		t.Fatal("explicit dingju method on a fixed-calendar scope should reject")
	}
}

func TestFlyingPlateShiftsStarGanPairs(t *testing.T) {
	dipan := placeDiPan(1, true)
	lead := ganzhi.Zhu{Gan: ganzhi.GanJi, Zhi: ganzhi.ZhiSi}
	result, landing := placeFlyTianPan(lead, GongZhong, dipan)
	if landing != GongLi {
		t.Fatalf("duty-star landing = %d, want %d", landing, GongLi)
	}
	delta := flyDelta(GongZhong, landing)
	for home := 1; home <= 9; home++ {
		target := int(flyTo(GongIndex(home), delta)) - 1
		item := result[target]
		if len(item) != 1 || item[0].Gan != dipan[home-1] || item[0].Star != palaceStar[home-1] {
			t.Fatalf("home palace %d shifted to %+v; want star-gan pair %+v", home, item,
				TianPanSymbol{Gan: dipan[home-1], Star: palaceStar[home-1]})
		}
	}
	if got := placeFlyShenPan(false, landing); got[int(landing)-1] != flySpiritOrder[0] {
		t.Fatalf("yang flying spirits do not start at duty-star palace: %+v", got)
	}
	if got := placeFlyShenPan(true, landing); got[int(landing)-1] != flySpiritOrder[0] {
		t.Fatalf("yin flying spirits do not start at duty-star palace: %+v", got)
	}
}

type methodAnchor struct {
	ID              string                 `json:"id"`
	Name            string                 `json:"name"`
	Provenance      string                 `json:"provenance"`
	Method          goldenMethod           `json:"method"`
	When            string                 `json:"when"`
	ActualJieQi     string                 `json:"actual_jie_qi"`
	YongJuJieQi     string                 `json:"yong_ju_jie_qi"`
	YinDun          bool                   `json:"yin_dun"`
	Ju              int                    `json:"ju"`
	Yuan            string                 `json:"yuan"`
	ZhiRunState     string                 `json:"zhi_run_state,omitempty"`
	DayFuTou        string                 `json:"day_fu_tou"`
	Lead            string                 `json:"lead"`
	XunShou         string                 `json:"xun_shou"`
	KongWang        string                 `json:"kong_wang"`
	DutyStar        string                 `json:"duty_star"`
	DutySource      int                    `json:"duty_source"`
	DutyLand        int                    `json:"duty_land"`
	DutyDoor        string                 `json:"duty_door"`
	DoorLand        int                    `json:"door_land"`
	Palaces         [9]string              `json:"palaces"`
	CrossValidation *methodCrossValidation `json:"cross_validation,omitempty"`
}

type methodGoldenFixture struct {
	Source       goldenSource           `json:"source"`
	CrossSources []methodCrossSource    `json:"cross_sources"`
	Vectors      []methodAnchor         `json:"vectors"`
	CrossGaps    []methodCombinationGap `json:"cross_validation_gaps"`
}

type methodCrossSourceFile struct {
	Path   string `json:"path"`
	SHA256 string `json:"sha256"`
}

type methodCrossSource struct {
	ID            string                  `json:"id"`
	Name          string                  `json:"name"`
	URL           string                  `json:"url"`
	Commit        string                  `json:"commit"`
	License       string                  `json:"license"`
	Usage         string                  `json:"usage"`
	LicenseSHA256 string                  `json:"license_sha256"`
	SourceFiles   []methodCrossSourceFile `json:"source_files"`
}

type methodCrossExclusion struct {
	Fact   string `json:"fact"`
	Reason string `json:"reason"`
}

type methodCrossValidation struct {
	SourceID      string                 `json:"source_id"`
	Status        string                 `json:"status"`
	ComparedFacts []string               `json:"compared_facts"`
	ExcludedFacts []methodCrossExclusion `json:"excluded_facts"`
	YinDun        bool                   `json:"yin_dun"`
	Ju            int                    `json:"ju"`
	Lead          string                 `json:"lead"`
	XunShou       string                 `json:"xun_shou"`
	KongWang      string                 `json:"kong_wang"`
	DutyStar      string                 `json:"duty_star"`
	DutySource    int                    `json:"duty_source,omitempty"`
	DutyLand      int                    `json:"duty_land"`
	DutyDoor      string                 `json:"duty_door"`
	DoorLand      int                    `json:"door_land"`
	CorePalaces   [9]string              `json:"core_palaces"`
	OuterDoors    []string               `json:"outer_doors"`
	Spirits       []string               `json:"spirits,omitempty"`
}

type methodCombinationGap struct {
	ID                  string                 `json:"id"`
	Method              goldenMethod           `json:"method"`
	VectorIDs           []string               `json:"vector_ids"`
	RequiredConventions []string               `json:"required_conventions"`
	Status              string                 `json:"status"`
	Reason              string                 `json:"reason"`
	WantedFacts         []string               `json:"wanted_facts"`
	ExcludedFacts       []methodCrossExclusion `json:"excluded_facts"`
}

func loadMethodMatrixGoldenFixture(t *testing.T) methodGoldenFixture {
	t.Helper()
	var fixture methodGoldenFixture
	loadGoldenFixture(t, "method_matrix_golden.json", &fixture)
	if fixture.Source != atopxGoldenSource() || len(fixture.Vectors) != 12 {
		t.Fatalf("method golden source/vectors = %+v/%d", fixture.Source, len(fixture.Vectors))
	}
	seenVector := map[string]bool{}
	checkedIn, generated := 0, 0
	for _, vector := range fixture.Vectors {
		if seenVector[vector.ID] || vector.Name == "" {
			t.Fatalf("invalid method golden vector %+v", vector)
		}
		switch vector.Provenance {
		case "atopx_checked_in":
			checkedIn++
		case "generated_from_atopx_reference":
			generated++
		default:
			t.Fatalf("unknown method golden provenance %q", vector.Provenance)
		}
		if _, err := ParseMethod(vector.Method.Scope, vector.Method.School, vector.Method.Dingju, ""); err != nil {
			t.Fatalf("invalid method golden combination %+v: %v", vector.Method, err)
		}
		if vector.ActualJieQi == "" || vector.YongJuJieQi == "" || vector.Ju < 1 || vector.Ju > 9 ||
			vector.Yuan == "" || vector.DayFuTou == "" {
			t.Fatalf("method golden vector %q lacks dingju metadata", vector.ID)
		}
		if DingjuMethod(vector.Method.Dingju) == DingjuZhiRun && vector.ZhiRunState == "" {
			t.Fatalf("zhirun vector %q lacks state", vector.ID)
		}
		if DingjuMethod(vector.Method.Dingju) != DingjuZhiRun && vector.ZhiRunState != "" {
			t.Fatalf("non-zhirun vector %q cannot carry state", vector.ID)
		}
		switch vector.ID {
		case "method-06":
			if vector.ZhiRunState != zhiRunStateNames["behind"] {
				t.Fatalf("method-06 state = %q, want 接气", vector.ZhiRunState)
			}
		case "method-12":
			if vector.ZhiRunState != zhiRunStateNames["repeat"] {
				t.Fatalf("method-12 state = %q, want 置闰", vector.ZhiRunState)
			}
		}
		seenVector[vector.ID] = true
	}
	if checkedIn != 5 || generated != 7 {
		t.Fatalf("method golden provenance counts = checked-in %d, generated %d; want 5/7", checkedIn, generated)
	}
	assertMethodCrossSources(t, fixture.CrossSources)
	openVectors := assertMethodCombinationGaps(t, fixture)
	crossValidated := 0
	wantCrossValidated := map[string]bool{
		"method-03": true, "method-04": true, "method-05": true,
		"method-09": true, "method-11": true, "method-12": true,
	}
	for _, vector := range fixture.Vectors {
		if vector.CrossValidation == nil {
			if vector.Provenance == "generated_from_atopx_reference" && !openVectors[vector.ID] {
				t.Fatalf("generated vector %q lacks cross validation or a combination gap", vector.ID)
			}
			continue
		}
		if !wantCrossValidated[vector.ID] {
			t.Fatalf("unexpected cross-validated vector %q", vector.ID)
		}
		delete(wantCrossValidated, vector.ID)
		crossValidated++
	}
	if crossValidated != 6 || len(wantCrossValidated) != 0 {
		t.Fatalf("method golden cross-validated vectors = %d, missing %+v", crossValidated, wantCrossValidated)
	}
	return fixture
}

func assertMethodCombinationGaps(t *testing.T, fixture methodGoldenFixture) map[string]bool {
	t.Helper()
	if len(fixture.CrossGaps) != 1 {
		t.Fatalf("method golden combination gaps = %d, want 1", len(fixture.CrossGaps))
	}
	gap := fixture.CrossGaps[0]
	wantMethod := goldenMethod{Scope: "day", School: "luoshu_feipan", Dingju: "zhirun"}
	wantVectors := map[string]bool{"method-06": true}
	if gap.ID != "day-feipan-zhirun-yin-core" || gap.Method != wantMethod || gap.Status != "open" ||
		gap.Reason == "" || len(gap.VectorIDs) != len(wantVectors) {
		t.Fatalf("invalid method combination gap %+v", gap)
	}
	wantConventions := map[string]bool{
		"day_lead_pillar": true, "yin_dun": true, "palace_number_shift": true,
		"star_gan_shift_together": true, "eight_doors_real_center_step": true,
		"zhirun": true,
	}
	seenConventions := make(map[string]bool, len(gap.RequiredConventions))
	for _, convention := range gap.RequiredConventions {
		if !wantConventions[convention] || seenConventions[convention] {
			t.Fatalf("invalid gap convention %q", convention)
		}
		seenConventions[convention] = true
	}
	if len(seenConventions) != len(wantConventions) {
		t.Fatalf("gap conventions = %+v, want %+v", gap.RequiredConventions, wantConventions)
	}

	byID := make(map[string]methodAnchor, len(fixture.Vectors))
	for _, vector := range fixture.Vectors {
		byID[vector.ID] = vector
	}
	covered := make(map[string]bool, len(gap.VectorIDs))
	for _, id := range gap.VectorIDs {
		vector, exists := byID[id]
		if !exists || !wantVectors[id] || covered[id] || vector.Method != gap.Method {
			t.Fatalf("invalid gap vector %q in %+v", id, gap)
		}
		if vector.CrossValidation != nil {
			t.Fatalf("gap vector %q cannot also be cross-validated", id)
		}
		covered[id] = true
	}

	wantedFacts := map[string]bool{
		"yin_dun": true, "ju": true, "lead": true, "xun_shou": true,
		"kong_wang": true, "duty_star": true, "duty_source": true,
		"duty_land": true, "duty_door": true, "door_land": true,
		"heaven_stem": true, "earth_stem": true, "star": true,
		"outer_door": true,
	}
	for _, fact := range gap.WantedFacts {
		if !wantedFacts[fact] {
			t.Fatalf("invalid gap fact %q", fact)
		}
		delete(wantedFacts, fact)
	}
	if len(wantedFacts) != 0 {
		t.Fatalf("gap is missing facts %+v", wantedFacts)
	}

	allowedExclusions := map[string]bool{"hidden_stem": true, "spirit": true}
	seenExclusions := make(map[string]bool, len(gap.ExcludedFacts))
	for _, exclusion := range gap.ExcludedFacts {
		if !allowedExclusions[exclusion.Fact] || exclusion.Reason == "" ||
			seenExclusions[exclusion.Fact] {
			t.Fatalf("invalid gap exclusion %+v", exclusion)
		}
		seenExclusions[exclusion.Fact] = true
	}
	if len(seenExclusions) != len(allowedExclusions) {
		t.Fatalf("gap exclusions = %+v, want hidden_stem and spirit", gap.ExcludedFacts)
	}
	return covered
}

func assertMethodCrossSources(t *testing.T, sources []methodCrossSource) {
	t.Helper()
	if len(sources) != 2 {
		t.Fatalf("method golden cross sources = %d, want 2", len(sources))
	}
	want := map[string]methodCrossSource{
		"qimen_go": {
			ID: "qimen_go", Name: "deminzhang/qimen-go",
			URL:     "https://github.com/deminzhang/qimen-go",
			Commit:  "4d3f58fa0f401b5b3a337f119138e99e90685dda",
			License: "MIT", Usage: "independent_execution",
			LicenseSHA256: "cfdb13c1a086ecc3f498b6a75a65d1dab3d93072c58efd26d383db776749b996",
			SourceFiles: []methodCrossSourceFile{
				{Path: "xuan/qimen.go", SHA256: "93ba61a632af3d402bfd65754d1adeb74b937f87b40e62fc50adc15a1d4bd350"},
				{Path: "xuan/qimen_defs.go", SHA256: "bacb7684160cb0d4297e1a13fb3674c7eebf56183fea2e7a5c12bb18a7dd47ac"},
			},
		},
		"mingyu": {
			ID: "mingyu", Name: "Brhiza/mingyu",
			URL:     "https://github.com/Brhiza/mingyu",
			Commit:  "7823b99fe0263608a538ed894c40cb825f0b3145",
			License: "AGPL-3.0-only", Usage: "factual_cross_check_only",
			LicenseSHA256: "d6662a5292a32e7bd3220901fc34aa2630eef474ae784caedc6d404cbd59e1ed",
			SourceFiles: []methodCrossSourceFile{
				{Path: "packages/core/src/calendar/timeManager.ts", SHA256: "f1dfb5f713ce9596b96a64665ac09ae95fa1f6dfc25e1947d89b6ade0980a2f1"},
				{Path: "packages/core/src/divination/algorithms/qimen/helpers/jushu.ts", SHA256: "c8b9165084f2537985c7788854798e5251f2cd7fef9bf483bc180b8691afda88"},
				{Path: "packages/core/src/divination/algorithms/qimen/helpers/layout.ts", SHA256: "a47c8aa64d61866bfcaabe35641191dbc5894498b41e46ed829ca0123053ecfa"},
			},
		},
	}
	seen := map[string]bool{}
	for _, source := range sources {
		if seen[source.ID] {
			t.Fatalf("duplicate cross source %q", source.ID)
		}
		seen[source.ID] = true
		expected := want[source.ID]
		if source.ID != expected.ID || source.Name != expected.Name || source.URL != expected.URL ||
			source.Commit != expected.Commit || source.License != expected.License ||
			source.Usage != expected.Usage || source.LicenseSHA256 != expected.LicenseSHA256 ||
			len(source.SourceFiles) != len(expected.SourceFiles) {
			t.Fatalf("cross source = %+v, want %+v", source, want[source.ID])
		}
		for i, file := range source.SourceFiles {
			if file != expected.SourceFiles[i] {
				t.Fatalf("cross source file %d = %+v, want %+v", i, file, expected.SourceFiles[i])
			}
		}
	}
	for id := range want {
		if !seen[id] {
			t.Fatalf("missing cross source %q", id)
		}
	}
}

func TestMethodMatrixExternalAnchors(t *testing.T) {
	fixture := loadMethodMatrixGoldenFixture(t)
	for _, anchor := range fixture.Vectors {
		t.Run(anchor.Name, func(t *testing.T) {
			method, err := ParseMethod(anchor.Method.Scope, anchor.Method.School, anchor.Method.Dingju, "")
			if err != nil {
				t.Fatal(err)
			}
			chart := computeMethodChartForTest(t, anchor.When, method)
			if chart.Method.Scope != method.Scope.String() || chart.Method.School != method.School.String() {
				t.Fatalf("method = %s/%s", chart.Method.Scope, chart.Method.School)
			}
			if chart.Method.JieQi != anchor.ActualJieQi || chart.Method.YongJuJieQi != anchor.YongJuJieQi ||
				chart.Pan.YinDun != anchor.YinDun || chart.Pan.Jushu != anchor.Ju || chart.Method.Yuan != anchor.Yuan ||
				chart.Method.ZhiRunState != anchor.ZhiRunState || chart.Method.DayFuTou.Name != anchor.DayFuTou {
				t.Fatalf("dingju metadata = %+v, want %+v", chart.Method, anchor)
			}
			if got := ganzhi.GanName(chart.Pan.LeadGan) + ganzhi.ZhiName(chart.Pan.LeadZhi); got != anchor.Lead {
				t.Fatalf("lead pillar = %s, want %s", got, anchor.Lead)
			}
			if got := chart.Method.LeadPillar.Name; got != anchor.Lead {
				t.Fatalf("method lead pillar = %s, want %s", got, anchor.Lead)
			}
			if chart.Method.LeadXunShou.LiuYi != anchor.XunShou {
				t.Fatalf("xun shou = %s, want %s", chart.Method.LeadXunShou.LiuYi, anchor.XunShou)
			}
			if got := ganzhi.ZhiName(chart.Pan.KongWang[0].Branch) + ganzhi.ZhiName(chart.Pan.KongWang[1].Branch); got != anchor.KongWang {
				t.Fatalf("kong wang = %s, want %s", got, anchor.KongWang)
			}
			if got := chart.Pan.DutyStar.String(); got != anchor.DutyStar {
				t.Fatalf("duty star = %s, want %s", got, anchor.DutyStar)
			}
			if got := int(chart.Method.LeadXunShouGong); got != anchor.DutySource {
				t.Fatalf("duty source = %d, want %d", got, anchor.DutySource)
			}
			if got := int(chart.Pan.DutyStarPalace); got != anchor.DutyLand {
				t.Fatalf("duty landing = %d, want %d", got, anchor.DutyLand)
			}
			if got := chart.Pan.DutyDoor.String(); got != anchor.DutyDoor {
				t.Fatalf("duty door = %s, want %s", got, anchor.DutyDoor)
			}
			if got := int(chart.Pan.DutyDoorPalace); got != anchor.DoorLand {
				t.Fatalf("duty door landing = %d, want %d", got, anchor.DoorLand)
			}
			assertMethodPalaces(t, chart, anchor.Palaces)
			if anchor.CrossValidation != nil {
				assertMethodCrossValidation(t, fixture, chart, anchor, *anchor.CrossValidation)
			}
		})
	}
}

func assertMethodCrossValidation(
	t *testing.T,
	fixture methodGoldenFixture,
	chart Chart,
	anchor methodAnchor,
	validation methodCrossValidation,
) {
	t.Helper()
	sourceByID := map[string]methodCrossSource{}
	for _, source := range fixture.CrossSources {
		sourceByID[source.ID] = source
	}
	source, ok := sourceByID[validation.SourceID]
	if !ok {
		t.Fatalf("unknown cross source %q", validation.SourceID)
	}
	if source.Usage != "independent_execution" && source.Usage != "factual_cross_check_only" {
		t.Fatalf("invalid cross source usage %q", source.Usage)
	}
	if validation.Status != "partial_match" {
		t.Fatalf("invalid cross-validation status %q", validation.Status)
	}

	allowedFacts := map[string]bool{
		"yin_dun": true, "ju": true, "lead": true, "xun_shou": true,
		"kong_wang": true, "duty_star": true, "duty_source": true,
		"duty_land": true, "duty_door": true, "door_land": true,
		"heaven_stem": true, "earth_stem": true, "star": true,
		"outer_door": true, "spirit": true,
	}
	requiredFacts := map[string]bool{
		"yin_dun": true, "ju": true, "lead": true, "xun_shou": true,
		"kong_wang": true, "duty_star": true, "duty_door": true,
		"heaven_stem": true, "earth_stem": true, "star": true,
		"outer_door": true,
	}
	seenFacts := map[string]bool{}
	for _, fact := range validation.ComparedFacts {
		if fact == "" || seenFacts[fact] || !allowedFacts[fact] {
			t.Fatalf("invalid or duplicate compared fact %q", fact)
		}
		seenFacts[fact] = true
	}
	seenExclusions := map[string]bool{}
	for _, exclusion := range validation.ExcludedFacts {
		if exclusion.Fact == "" || exclusion.Reason == "" || seenExclusions[exclusion.Fact] {
			t.Fatalf("invalid or duplicate exclusion %+v", exclusion)
		}
		if seenFacts[exclusion.Fact] {
			t.Fatalf("fact %q is both compared and excluded", exclusion.Fact)
		}
		seenExclusions[exclusion.Fact] = true
	}
	delete(requiredFacts, "spirit")
	for fact := range seenExclusions {
		delete(requiredFacts, fact)
	}
	for fact := range requiredFacts {
		if !seenFacts[fact] {
			t.Fatalf("missing compared fact %q", fact)
		}
	}
	if seenFacts["spirit"] && len(validation.Spirits) == 0 {
		t.Fatal("a spirit cross-validation must contain spirit values")
	}
	if !seenFacts["spirit"] && len(validation.Spirits) != 0 {
		t.Fatal("spirits cannot be present when spirit is excluded")
	}

	if chart.Pan.YinDun != validation.YinDun || chart.Pan.Jushu != validation.Ju {
		t.Fatalf("cross-checked dun/ju = %v/%d, want %v/%d",
			chart.Pan.YinDun, chart.Pan.Jushu, validation.YinDun, validation.Ju)
	}
	if anchor.Lead != validation.Lead || anchor.XunShou != validation.XunShou ||
		anchor.KongWang != validation.KongWang || anchor.DutyStar != validation.DutyStar ||
		anchor.DutyDoor != validation.DutyDoor {
		t.Fatalf("anchor and cross-validation metadata differ: %+v vs %+v", anchor, validation)
	}
	if seenFacts["duty_source"] && anchor.DutySource != validation.DutySource {
		t.Fatalf("cross duty source = %d, want %d", validation.DutySource, anchor.DutySource)
	}
	if seenFacts["duty_land"] && anchor.DutyLand != validation.DutyLand {
		t.Fatalf("cross duty landing = %d, want %d", validation.DutyLand, anchor.DutyLand)
	}
	if seenFacts["door_land"] && anchor.DoorLand != validation.DoorLand {
		t.Fatalf("cross door landing = %d, want %d", validation.DoorLand, anchor.DoorLand)
	}

	nonEmptyCore := 0
	for i, want := range validation.CorePalaces {
		if want == "" {
			continue
		}
		runes := []rune(want)
		if len(runes) != 3 {
			t.Fatalf("cross core palace %d %q must have three runes", i+1, want)
		}
		nonEmptyCore++
		palace := chart.Pan.GongWei[i]
		if len(palace.TianPan) == 0 {
			t.Fatalf("cross core palace %d has no heaven layer", i+1)
		}
		got := palace.TianPan[0].Gan.String() + palace.DiPanGan.String() +
			string([]rune(palace.TianPan[0].Star.String())[1])
		if got != want {
			t.Fatalf("cross core palace %d = %s, want %s", i+1, got, want)
		}
	}
	if nonEmptyCore < 8 {
		t.Fatalf("cross core palaces = %d, want at least 8", nonEmptyCore)
	}

	if seenFacts["outer_door"] {
		doorIndex := 0
		for i, palace := range chart.Pan.GongWei {
			if !palace.DoorSet {
				continue
			}
			want := validation.OuterDoors[doorIndex]
			doorIndex++
			if string([]rune(palace.Door.String())[0]) != want {
				t.Fatalf("cross outer door %d = %v, want %s", i+1, palace.Door, want)
			}
		}
		if doorIndex != len(validation.OuterDoors) {
			t.Fatalf("cross outer door count = %d, want %d", doorIndex, len(validation.OuterDoors))
		}
	} else if len(validation.OuterDoors) != 0 {
		t.Fatal("outer doors cannot be present when outer_door is excluded")
	}

	if len(validation.Spirits) != 0 {
		spiritIndex := 0
		for i, palace := range chart.Pan.GongWei {
			got := spiritNameForChart(chart, palace.Spirit)
			if !palace.SpiritSet {
				continue
			}
			if spiritIndex >= len(validation.Spirits) || got != validation.Spirits[spiritIndex] {
				t.Fatalf("cross spirit %d = %s, want %v", i+1, got, validation.Spirits)
			}
			spiritIndex++
		}
		if spiritIndex != len(validation.Spirits) {
			t.Fatalf("cross spirit count = %d, want %d", spiritIndex, len(validation.Spirits))
		}
	}
}

func TestMethodMatrixCombinationInvariants(t *testing.T) {
	for _, method := range []Method{
		{Scope: ScopeHour, School: SchoolZhuanPan, Dingju: DingjuChaiBu},
		{Scope: ScopeHour, School: SchoolZhuanPan, Dingju: DingjuZhiRun},
		{Scope: ScopeHour, School: SchoolLuoShuFeiPan, Dingju: DingjuChaiBu},
		{Scope: ScopeHour, School: SchoolLuoShuFeiPan, Dingju: DingjuZhiRun},
		{Scope: ScopeDay, School: SchoolZhuanPan, Dingju: DingjuChaiBu},
		{Scope: ScopeDay, School: SchoolZhuanPan, Dingju: DingjuZhiRun},
		{Scope: ScopeDay, School: SchoolLuoShuFeiPan, Dingju: DingjuChaiBu},
		{Scope: ScopeDay, School: SchoolLuoShuFeiPan, Dingju: DingjuZhiRun},
		{Scope: ScopeMonth, School: SchoolZhuanPan},
		{Scope: ScopeMonth, School: SchoolLuoShuFeiPan},
		{Scope: ScopeYear, School: SchoolZhuanPan},
		{Scope: ScopeYear, School: SchoolLuoShuFeiPan},
	} {
		t.Run(method.Scope.String()+"/"+method.School.String()+"/"+method.Dingju.String(), func(t *testing.T) {
			chart := computeMethodChartForTest(t, "2026-03-02 18:30", method)
			st := methodSolarTimeForTest(t, "2026-03-02 18:30")
			bz := tianwen.ComputeBazi(st)
			wantLead := map[Scope]ganzhi.Zhu{
				ScopeHour: bz.Shi, ScopeDay: bz.Ri, ScopeMonth: bz.Yue, ScopeYear: bz.Nian,
			}[method.Scope]
			if chart.Pan.LeadGan != wantLead.Gan || chart.Pan.LeadZhi != wantLead.Zhi {
				t.Fatalf("lead pillar = %s%s, want %s%s", chart.Pan.LeadGan, chart.Pan.LeadZhi, wantLead.Gan, wantLead.Zhi)
			}
			stars, doors, spirits := 0, 0, 0
			for _, palace := range chart.Pan.GongWei {
				if palace.AnGan == 0 {
					t.Fatal("hidden stem missing")
				}
				stars += len(palace.TianPan)
				if palace.DoorSet {
					doors++
				}
				if palace.SpiritSet {
					spirits++
				}
			}
			if method.School == SchoolLuoShuFeiPan {
				if stars != 9 || doors != 8 || spirits != 9 {
					t.Fatalf("fly layer counts = %d/%d/%d, want 9/8/9", stars, doors, spirits)
				}
			} else if stars != 9 || doors != 8 || spirits != 8 {
				t.Fatalf("rotate layer counts = %d/%d/%d, want 9/8/8", stars, doors, spirits)
			}
			if chart.Method.Scope != method.Scope.String() || chart.Method.School != method.School.String() {
				t.Fatalf("method metadata = %s/%s", chart.Method.Scope, chart.Method.School)
			}
		})
	}
}

func assertMethodPalaces(t *testing.T, chart Chart, specs [9]string) {
	t.Helper()
	for i, spec := range specs {
		runes := []rune(spec)
		if len(runes) != 6 {
			t.Fatalf("palace %d spec %q must have six runes", i+1, spec)
		}
		palace := chart.Pan.GongWei[i]
		if got := palace.DiPanGan.String(); got != string(runes[1]) {
			t.Errorf("palace %d earth = %s, want %s", i+1, got, string(runes[1]))
		}
		if got := palace.AnGan.String(); got != string(runes[2]) {
			t.Errorf("palace %d hidden = %s, want %s", i+1, got, string(runes[2]))
		}
		starName := ""
		if len(palace.TianPan) != 0 {
			starName = palace.TianPan[0].Star.String()
		}
		assertMethodLayer(t, i+1, "star", runes[3], len(palace.TianPan) != 0, starName, starRuneNames)
		assertMethodLayer(t, i+1, "door", runes[4], palace.Door != 0, palace.Door.String(), doorRuneNames)
		assertMethodLayer(
			t, i+1, "spirit", runes[5], palace.Spirit != 0,
			spiritNameForChart(chart, palace.Spirit),
			map[rune]string{runes[5]: spiritExpectedNameForChart(chart, runes[5])},
		)
		heaven := ""
		if len(palace.TianPan) != 0 {
			heaven = palace.TianPan[0].Gan.String()
		} else if palace.TianPanGan != nil {
			heaven = palace.TianPanGan.String()
		}
		if heaven != string(runes[0]) {
			t.Errorf("palace %d heaven = %q, want %s", i+1, heaven, string(runes[0]))
		}
	}
}

var starRuneNames = map[rune]string{
	'蓬': "天蓬", '芮': "天芮", '冲': "天冲", '辅': "天辅",
	'禽': "天禽", '心': "天心", '柱': "天柱", '任': "天任", '英': "天英",
}

var doorRuneNames = map[rune]string{
	'休': "休门", '生': "生门", '伤': "伤门", '杜': "杜门",
	'景': "景门", '死': "死门", '惊': "惊门", '开': "开门",
}

var spiritRuneNames = map[rune]string{
	'符': "值符", '蛇': "螣蛇", '阴': "太阴", '合': "六合",
	'虎': "白虎", '常': "太常", '玄': "玄武", '地': "九地", '天': "九天",
}

func assertMethodLayer(t *testing.T, palace int, layer string, wantRune rune, present bool, value string, names map[rune]string) {
	t.Helper()
	if wantRune == '-' {
		if present {
			t.Errorf("palace %d %s = %s, want empty", palace, layer, value)
		}
		return
	}
	if !present {
		t.Errorf("palace %d %s empty, want %s", palace, layer, names[wantRune])
		return
	}
	if got := value; got != names[wantRune] {
		t.Errorf("palace %d %s = %s, want %s", palace, layer, got, names[wantRune])
	}
}

func spiritNameForChart(chart Chart, spirit SpiritIndex) string {
	return spiritDisplayName(spirit, chart.Pan.YinDun, chart.Pan.School)
}

func spiritExpectedNameForChart(chart Chart, wantRune rune) string {
	if chart.Pan.School == SchoolLuoShuFeiPan {
		return spiritRuneNames[wantRune]
	}
	switch wantRune {
	case '虎':
		if chart.Pan.YinDun {
			return SpiritGouChen.YinName()
		}
		return SpiritGouChen.YangName()
	case '玄':
		if chart.Pan.YinDun {
			return SpiritZhuQue.YinName()
		}
		return SpiritZhuQue.YangName()
	default:
		return spiritRuneNames[wantRune]
	}
}

func computeMethodChartForTest(t *testing.T, when string, method Method) Chart {
	t.Helper()
	civil, err := time.ParseInLocation("2006-01-02 15:04", when, time.FixedZone("CST", 8*3600))
	if err != nil {
		t.Fatal(err)
	}
	chart, err := ComputeChartWithMethod(tianwen.SolarTime(civil), method)
	if err != nil {
		t.Fatal(err)
	}
	return chart
}

func methodSolarTimeForTest(t *testing.T, when string) tianwen.SolarTime {
	t.Helper()
	civil, err := time.ParseInLocation("2006-01-02 15:04", when, time.FixedZone("CST", 8*3600))
	if err != nil {
		t.Fatal(err)
	}
	return tianwen.SolarTime(civil)
}
