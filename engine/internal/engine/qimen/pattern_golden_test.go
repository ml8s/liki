package qimen

import (
	"testing"

	"liki-engine/internal/engine/ganzhi"
)

type patternGoldenPalace struct {
	Gong       string             `json:"gong"`
	DiPanGan   string             `json:"di_pan_gan"`
	TianPan    []goldenTianSymbol `json:"tian_pan"`
	TianPanGan string             `json:"tian_pan_gan"`
	Door       string             `json:"door"`
	Spirit     string             `json:"spirit"`
}

type goldenTianSymbol struct {
	Gan  string `json:"gan"`
	Star string `json:"star"`
}

type patternGoldenPanSpec struct {
	School   string                `json:"school"`
	YinDun   bool                  `json:"yin_dun"`
	DutyStar string                `json:"duty_star"`
	DutyDoor string                `json:"duty_door"`
	Palaces  []patternGoldenPalace `json:"palaces"`
}

type patternGoldenVector struct {
	ID              string               `json:"id"`
	ExpectedPattern string               `json:"expected_pattern"`
	ExpectedPalaces []string             `json:"expected_palaces"`
	SourceBasis     string               `json:"source_basis"`
	Pan             patternGoldenPanSpec `json:"pan"`
}

type patternGoldenFixture struct {
	Source struct {
		Name  string `json:"name"`
		Scope string `json:"scope"`
	} `json:"source"`
	Vectors []patternGoldenVector `json:"vectors"`
}

func TestPatternGoldenVectors(t *testing.T) {
	var fixture patternGoldenFixture
	loadGoldenFixture(t, "pattern_golden.json", &fixture)
	if fixture.Source.Name != "qimen classical pattern definitions" ||
		fixture.Source.Scope != "curated_definition_vectors" || len(fixture.Vectors) != 25 {
		t.Fatalf("pattern golden fixture = %+v/%d vectors", fixture.Source, len(fixture.Vectors))
	}
	seen := map[string]bool{}
	covered := map[string]bool{}
	vectorCounts := map[string]int{}
	sanQiDeShiPairs := map[string]bool{}
	bases := map[string]string{}
	for _, rule := range patternRules {
		bases[rule.Name] = rule.Basis
	}
	for _, vector := range fixture.Vectors {
		if seen[vector.ID] || len(vector.ExpectedPalaces) == 0 {
			t.Fatalf("invalid pattern golden vector %+v", vector)
		}
		if bases[vector.ExpectedPattern] != vector.SourceBasis {
			t.Fatalf("%s basis = %q, want pattern-table basis %q", vector.ID, vector.SourceBasis, bases[vector.ExpectedPattern])
		}
		seen[vector.ID] = true
		covered[vector.ExpectedPattern] = true
		vectorCounts[vector.ExpectedPattern]++
		if vector.ExpectedPattern == "三奇得使" && len(vector.Pan.Palaces) == 1 &&
			len(vector.Pan.Palaces[0].TianPan) == 1 {
			sanQiDeShiPairs[vector.Pan.Palaces[0].TianPan[0].Gan+"|"+vector.Pan.Palaces[0].DiPanGan] = true
		}
		assertPatternGoldenVector(t, vector)
	}
	for _, pair := range []string{"乙|己", "乙|辛", "丙|戊", "丙|庚", "丁|壬", "丁|癸"} {
		if !sanQiDeShiPairs[pair] {
			t.Fatalf("三奇得使 golden lacks pair %s", pair)
		}
	}
	for _, rule := range patternRules {
		if !covered[rule.Name] {
			t.Fatalf("pattern rule %s has no golden vector", rule.Name)
		}
		if rule.DutyStarPosition == "" && vectorCounts[rule.Name] < len(rule.Conditions) {
			t.Fatalf("pattern rule %s covers %d/%d condition branches", rule.Name, vectorCounts[rule.Name], len(rule.Conditions))
		}
	}
}

func TestGanInteractionPatternsUseSingleSymbolFact(t *testing.T) {
	if len(patternRules) != 13 {
		t.Fatalf("pattern rules = %d, want 13", len(patternRules))
	}
	for _, name := range []string{"青龙返首", "飞鸟跌穴", "荧入太白", "太白入荧"} {
		var rule *patternRule
		for i := range patternRules {
			if patternRules[i].Name == name {
				rule = &patternRules[i]
				break
			}
		}
		if rule == nil {
			t.Fatalf("pattern rule %s missing", name)
		}
		if len(rule.Conditions) != 1 || len(rule.Conditions[0].HeavenGan) != 1 ||
			len(rule.Conditions[0].EarthGan) != 1 || len(rule.Conditions[0].Doors) != 0 ||
			len(rule.Conditions[0].Spirits) != 0 {
			t.Fatalf("%s conditions = %+v, want one heaven/earth gan pair", name, rule.Conditions)
		}
		entry, ok := ganInteractionTable[[2]ganzhi.Gan{
			rule.Conditions[0].EarthGan[0], rule.Conditions[0].HeavenGan[0],
		}]
		if !ok || entry.PatternName != name || entry.Auspicious != rule.Auspicious {
			t.Fatalf("%s is not derived from its gan interaction fact %+v", name, entry)
		}
	}
}

func TestLiuYiJiXingUsesPalaceBoundConditions(t *testing.T) {
	var rule *patternRule
	for i := range patternRules {
		if patternRules[i].Name == "六仪击刑" {
			rule = &patternRules[i]
			break
		}
	}
	if rule == nil {
		t.Fatal("六仪击刑 pattern rule missing")
	}
	if rule.Auspicious {
		t.Fatal("六仪击刑 must be inauspicious")
	}
	want := []struct {
		heavenGan string
		palace    string
	}{
		{heavenGan: "戊", palace: "震"},
		{heavenGan: "己", palace: "坤"},
		{heavenGan: "庚", palace: "艮"},
		{heavenGan: "辛", palace: "离"},
		{heavenGan: "壬", palace: "巽"},
		{heavenGan: "癸", palace: "巽"},
	}
	if len(rule.Conditions) != len(want) {
		t.Fatalf("六仪击刑 conditions = %+v", rule.Conditions)
	}
	for i, expected := range want {
		condition := rule.Conditions[i]
		if len(condition.HeavenGan) != 1 || condition.HeavenGan[0].String() != expected.heavenGan ||
			len(condition.Palaces) != 1 || condition.Palaces[0].String() != expected.palace ||
			len(condition.EarthGan) != 0 || len(condition.Doors) != 0 || len(condition.Spirits) != 0 {
			t.Fatalf("六仪击刑 condition %d = %+v, want %s at %s", i, condition, expected.heavenGan, expected.palace)
		}
	}
}

func TestSanQiRuMuUsesPalaceBoundConditions(t *testing.T) {
	var rule *patternRule
	for i := range patternRules {
		if patternRules[i].Name == "三奇入墓" {
			rule = &patternRules[i]
			break
		}
	}
	if rule == nil {
		t.Fatal("三奇入墓 pattern rule missing")
	}
	if rule.Auspicious {
		t.Fatal("三奇入墓 must be inauspicious")
	}
	want := []struct {
		heavenGan string
		palace    string
	}{
		{heavenGan: "乙", palace: "坤"},
		{heavenGan: "丙", palace: "乾"},
		{heavenGan: "丁", palace: "艮"},
	}
	if len(rule.Conditions) != len(want) {
		t.Fatalf("三奇入墓 conditions = %+v", rule.Conditions)
	}
	for i, expected := range want {
		condition := rule.Conditions[i]
		if len(condition.HeavenGan) != 1 || condition.HeavenGan[0].String() != expected.heavenGan ||
			len(condition.Palaces) != 1 || condition.Palaces[0].String() != expected.palace ||
			len(condition.EarthGan) != 0 || len(condition.Doors) != 0 || len(condition.Spirits) != 0 {
			t.Fatalf("三奇入墓 condition %d = %+v, want %s at %s", i, condition, expected.heavenGan, expected.palace)
		}
	}
}

func TestSanQiRuMuRejectsUnboundPalace(t *testing.T) {
	chart := pan{School: SchoolZhuanPan}
	chart.GongWei[int(GongLi)-1].TianPan = []TianPanSymbol{
		{Gan: ganzhi.GanYi, Star: StarTianRui},
	}
	for _, pattern := range findPatterns(chart) {
		if pattern.Name == "三奇入墓" {
			t.Fatal("乙奇 outside 坤 must not be detected as 三奇入墓")
		}
	}
}

func assertPatternGoldenVector(t *testing.T, vector patternGoldenVector) {
	t.Helper()
	chart := patternGoldenPan(t, vector.Pan)
	patterns := findPatterns(chart)
	found := false
	for _, pattern := range patterns {
		if pattern.Name != vector.ExpectedPattern {
			continue
		}
		found = true
		if len(pattern.GongWei) != len(vector.ExpectedPalaces) {
			t.Fatalf("%s pattern palaces = %+v, want %+v", vector.ID, pattern.GongWei, vector.ExpectedPalaces)
		}
		for i, want := range vector.ExpectedPalaces {
			if pattern.GongWei[i].String() != want {
				t.Fatalf("%s palace %d = %s, want %s", vector.ID, i, pattern.GongWei[i], want)
			}
		}
	}
	if !found {
		t.Fatalf("%s patterns = %+v, want %s", vector.ID, patterns, vector.ExpectedPattern)
	}
}

func patternGoldenPan(t *testing.T, golden patternGoldenPanSpec) pan {
	t.Helper()
	school, err := ParseSchool(golden.School)
	if err != nil {
		t.Fatal(err)
	}
	dutyStar, err := ParseStarIndex(golden.DutyStar)
	if err != nil {
		t.Fatal(err)
	}
	dutyDoor, err := ParseDoorIndex(golden.DutyDoor)
	if err != nil {
		t.Fatal(err)
	}
	result := pan{School: school, YinDun: golden.YinDun, DutyStar: dutyStar, DutyDoor: dutyDoor}
	for _, goldenPalace := range golden.Palaces {
		palace, err := ParsePalaceIndex(goldenPalace.Gong)
		if err != nil {
			t.Fatal(err)
		}
		item := &result.GongWei[int(palace)-1]
		item.DiPanGan, err = ganzhi.ParseGan(goldenPalace.DiPanGan)
		if err != nil {
			t.Fatal(err)
		}
		for _, symbol := range goldenPalace.TianPan {
			gan, err := ganzhi.ParseGan(symbol.Gan)
			if err != nil {
				t.Fatal(err)
			}
			star, err := ParseStarIndex(symbol.Star)
			if err != nil {
				t.Fatal(err)
			}
			item.TianPan = append(item.TianPan, TianPanSymbol{Gan: gan, Star: star})
		}
		if goldenPalace.TianPanGan != "" {
			gan, err := ganzhi.ParseGan(goldenPalace.TianPanGan)
			if err != nil {
				t.Fatal(err)
			}
			item.TianPanGan = &gan
		}
		if goldenPalace.Door != "" {
			item.Door, err = ParseDoorIndex(goldenPalace.Door)
			if err != nil {
				t.Fatal(err)
			}
		}
		if goldenPalace.Spirit != "" {
			item.Spirit, err = ParseSpiritIndex(goldenPalace.Spirit)
			if err != nil {
				t.Fatal(err)
			}
		}
	}
	return result
}
