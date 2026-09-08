package qimen

import (
	"strings"
	"testing"
	"time"

	"liki-engine/internal/engine/ganzhi"
	"liki-engine/internal/engine/tianwen"
)

func TestPlateRingTables(t *testing.T) {
	for i, palace := range outerRing {
		if palace == GongZhong {
			t.Fatalf("outer ring contains central palace at %d", i)
		}
		if starHomePalace(starOrder8[i]) != int(palace)-1 {
			t.Errorf("ring %d: star %s does not originate from %s", i, starOrder8[i], palace)
		}
		if int(doorOrder[i]) == 0 {
			t.Errorf("ring %d: empty door", i)
		}
	}
	for i, palace := range zhiPalaceTable {
		if palace == 0 || palace == GongZhong {
			t.Errorf("branch %d has invalid palace %d", i+1, palace)
		}
	}
	for i, ma := range maXingTable {
		if ma == 0 {
			t.Errorf("branch %d has no horse", i+1)
		}
	}
	for i := 1; i <= 9; i++ {
		if palaceWuxing(GongIndex(i)) == 0 || starWuxing(StarIndex(i)) == 0 {
			t.Errorf("palace/star %d has no five-element", i)
		}
	}
	for i := 1; i <= 8; i++ {
		if doorWuxing(DoorIndex(i)) == 0 {
			t.Errorf("door %d has no five-element", i)
		}
	}
}

func TestPatternTableContract(t *testing.T) {
	names := map[string]bool{}
	for _, rule := range patternRules {
		if names[rule.Name] {
			t.Fatalf("duplicate pattern rule %q", rule.Name)
		}
		names[rule.Name] = true
		if rule.Description == "" || rule.Basis == "" {
			t.Errorf("%s lacks description or basis", rule.Name)
		}
		if rule.DutyStarPosition == "" {
			if len(rule.Conditions) == 0 {
				t.Errorf("%s lacks symbol conditions", rule.Name)
			}
			for _, condition := range rule.Conditions {
				if len(condition.HeavenGan) == 0 && len(condition.EarthGan) == 0 &&
					len(condition.Doors) == 0 && len(condition.Spirits) == 0 &&
					len(condition.Palaces) == 0 {
					t.Errorf("%s has an empty condition", rule.Name)
				}
			}
		}
	}
}

func TestPatternTableLoaderRejectsMalformedRules(t *testing.T) {
	tests := []struct {
		name string
		raw  string
		want string
	}{
		{
			name: "duty star position cannot carry ignored conditions",
			raw: `[
				{"name":"bad","description":"x","basis":"x","auspicious":false,
				 "duty_star_position":"home",
				 "conditions":[{"heaven_gan":["乙"],"palaces":["坎"]}]}
			]`,
			want: "must not define conditions",
		},
		{
			name: "standard rule requires conditions",
			raw: `[
				{"name":"bad","description":"x","basis":"x","auspicious":false}
			]`,
			want: "requires conditions",
		},
		{
			name: "empty condition is rejected",
			raw: `[
				{"name":"bad","description":"x","basis":"x","auspicious":false,
				 "conditions":[{}]}
			]`,
			want: "empty condition",
		},
		{
			name: "duplicate condition values are rejected",
			raw: `[
				{"name":"bad","description":"x","basis":"x","auspicious":false,
				 "conditions":[{"heaven_gan":["乙","乙"]}]}
			]`,
			want: "duplicate heaven_gan",
		},
		{
			name: "center palace is rejected",
			raw: `[
				{"name":"bad","description":"x","basis":"x","auspicious":false,
				 "conditions":[{"heaven_gan":["乙"],"palaces":["中"]}]}
			]`,
			want: "center palace",
		},
	}
	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			if _, err := decodePatternTable([]byte(test.raw)); err == nil || !strings.Contains(err.Error(), test.want) {
				t.Fatalf("error = %v, want %q", err, test.want)
			}
		})
	}
}

func TestStandardRotatingPlateAnchor(t *testing.T) {
	// Independent anchor: Yang-6 school, Geng-Chen hour, Jia-Xu/ Ji cycle head.
	ju := juShu{Number: 6, YinDun: false}
	hour := ganzhi.Zhu{Gan: ganzhi.GanGeng, Zhi: ganzhi.ZhiChen}
	ju.Method = Method{Scope: ScopeHour, School: SchoolZhuanPan, Dingju: DingjuChaiBu}
	bz := ganzhi.Bazi{Ri: ganzhi.Zhu{Gan: ganzhi.GanYi}, Shi: hour}
	chart := computePan(ju, bz)

	wantEarth := [9]ganzhi.Gan{
		ganzhi.GanRen, ganzhi.GanGui, ganzhi.GanDing, ganzhi.GanBing,
		ganzhi.GanYi, ganzhi.GanWu, ganzhi.GanJi, ganzhi.GanGeng, ganzhi.GanXin,
	}
	for i, want := range wantEarth {
		if chart.GongWei[i].DiPanGan != want {
			t.Fatalf("earth palace %d = %s, want %s", i+1, chart.GongWei[i].DiPanGan, want)
		}
	}
	if chart.DutyStar != StarTianZhu || chart.DutyDoor != DoorJingMen {
		t.Fatalf("duty = %s/%s, want 天柱/惊门", chart.DutyStar, chart.DutyDoor)
	}

	wantHeaven := []struct {
		palace GongIndex
		gan    ganzhi.Gan
		star   StarIndex
	}{
		{GongGen, ganzhi.GanJi, StarTianZhu},
		{GongZhen, ganzhi.GanWu, StarTianXin},
		{GongXun, ganzhi.GanRen, StarTianPeng},
		{GongLi, ganzhi.GanGeng, StarTianRen},
		{GongKun, ganzhi.GanDing, StarTianChong},
		{GongDui, ganzhi.GanBing, StarTianFu},
		{GongQian, ganzhi.GanXin, StarTianYing},
		{GongKan, ganzhi.GanGui, StarTianRui},
		{GongKan, ganzhi.GanYi, StarTianQin},
	}
	for _, want := range wantHeaven {
		found := false
		for _, item := range chart.GongWei[int(want.palace)-1].TianPan {
			if item.Gan == want.gan && item.Star == want.star {
				found = true
			}
		}
		if !found {
			t.Errorf("%s lacks heaven %s/%s", want.palace, want.gan, want.star)
		}
	}

	wantDoors := map[GongIndex]DoorIndex{
		GongXun: DoorJingMen, GongLi: DoorKai, GongKun: DoorXiu,
		GongDui: DoorSheng, GongQian: DoorShang, GongKan: DoorDu,
		GongGen: DoorJing, GongZhen: DoorSi,
	}
	for palace, want := range wantDoors {
		if got := chart.GongWei[int(palace)-1].Door; got != want {
			t.Errorf("%s door = %s, want %s", palace, got, want)
		}
	}
	if got := chart.GongWei[int(GongGen)-1].Spirit; got != SpiritZhiFu {
		t.Errorf("Gen spirit = %s, want 值符", got.YangName())
	}
}

func TestHourJiaUsesCycleHeadLiuYi(t *testing.T) {
	ju := juShu{Number: 1, YinDun: false}
	hour := ganzhi.Zhu{Gan: ganzhi.GanJia, Zhi: ganzhi.ZhiXu}
	ju.Method = Method{Scope: ScopeHour, School: SchoolZhuanPan, Dingju: DingjuChaiBu}
	bz := ganzhi.Bazi{Ri: ganzhi.Zhu{Gan: ganzhi.GanWu}, Shi: hour}
	chart := computePan(ju, bz)
	if chart.DutyStar != StarTianRui {
		t.Fatalf("duty star = %s, want 天芮 by Jia-Xu→Ji", chart.DutyStar)
	}
	if got := findGanPalaceIdx(chart, ganzhi.GanJi); got != GongKun {
		t.Errorf("Jia-Xu hidden Ji palace = %s, want 坤", got)
	}
}

func TestCentralCycleHeadMetadata(t *testing.T) {
	bz := ganzhi.Bazi{
		Ri:  ganzhi.Zhu{Gan: ganzhi.GanWu, Zhi: ganzhi.ZhiChen},
		Shi: ganzhi.Zhu{Gan: ganzhi.GanJia, Zhi: ganzhi.ZhiChen},
	}
	ju := juShu{Number: 1, Method: Method{Scope: ScopeHour, School: SchoolZhuanPan, Dingju: DingjuChaiBu}}
	method := buildChartMethod(ju, bz)
	if method.LeadXunShou.Name != "甲辰" || method.LeadXunShou.LiuYi != "壬" {
		t.Errorf("cycle head = %+v, want 甲辰/壬", method.LeadXunShou)
	}
	if method.LeadXunShouGong != GongZhong {
		t.Errorf("cycle-head palace = %s, want 中", method.LeadXunShouGong)
	}
	if method.TianQin.Star != "天禽" || method.TianQin.HomeGong != "中" ||
		method.TianQin.LodgingGong != "坤" || method.TianQin.Follows != "天芮" ||
		!method.TianQin.CarriesCenterEarthGan {
		t.Errorf("TianQin rule = %+v", method.TianQin)
	}
}

func TestExactSolarTermBoundary(t *testing.T) {
	boundaries := []struct {
		name      string
		longitude float64
		afterName string
	}{
		{"lichun", 315, "立春"},
		{"winter-solstice", 270, "冬至"},
		{"summer-solstice", 90, "夏至"},
	}
	for _, boundary := range boundaries {
		t.Run(boundary.name, func(t *testing.T) {
			exact := tianwen.SolarTermTime(2024, boundary.longitude)
			if got := currentSolarTermInfo(exact.Add(-time.Hour)).Name; got == boundary.afterName {
				t.Errorf("one hour before boundary entered %s", boundary.afterName)
			}
			if got := currentSolarTermInfo(exact.Add(time.Hour)).Name; got != boundary.afterName {
				t.Errorf("one hour after boundary term = %s, want %s", got, boundary.afterName)
			}
			if got := currentSolarTermInfo(exact).Name; got != boundary.afterName {
				t.Errorf("at boundary term = %s, want %s", got, boundary.afterName)
			}
		})
	}
}

func TestPatternConditionsRequireSamePalace(t *testing.T) {
	scattered := pan{
		DutyDoor: DoorSheng,
		GongWei: [9]Gong{
			0: {DiPanGan: ganzhi.GanWu, TianPan: []TianPanSymbol{{Gan: ganzhi.GanBing}}},
			1: {DiPanGan: ganzhi.GanDing, Door: DoorSheng},
			2: {DiPanGan: ganzhi.GanWu, TianPan: []TianPanSymbol{{Gan: ganzhi.GanBing}}, Door: DoorSheng},
		},
	}
	if names := patternNames(findPatterns(scattered)); names["天遁"] {
		t.Fatalf("scattered symbols must not form 天遁: %v", names)
	}
	scattered.GongWei[1] = Gong{
		DiPanGan: ganzhi.GanDing,
		TianPan:  []TianPanSymbol{{Gan: ganzhi.GanBing}},
		Door:     DoorSheng,
	}
	if !patternNames(findPatterns(scattered))["天遁"] {
		t.Fatal("same-palace 丙/生门/丁 must form 天遁")
	}
}

func TestTianQinSpecialPositionFollowsPlateGeometry(t *testing.T) {
	rotate := computeMethodChartForTest(
		t, "2024-12-25 12:00",
		Method{Scope: ScopeDay, School: SchoolZhuanPan, Dingju: DingjuZhiRun},
	)
	fly := computeMethodChartForTest(
		t, "2024-12-25 12:00",
		Method{Scope: ScopeDay, School: SchoolLuoShuFeiPan, Dingju: DingjuZhiRun},
	)
	if rotate.Pan.DutyStar != StarTianQin || fly.Pan.DutyStar != StarTianQin {
		t.Fatalf("duty stars = %s/%s, want 天禽/天禽", rotate.Pan.DutyStar, fly.Pan.DutyStar)
	}
	if got := patternPalaceByName(rotate.Patterns, "伏吟"); got != GongKun {
		t.Fatalf("rotate TianQin FuYin palace = %s, want effective 坤", got)
	}
	if got := patternPalaceByName(fly.Patterns, "伏吟"); got != GongZhong {
		t.Fatalf("fly TianQin FuYin palace = %s, want real 中", got)
	}
}

func TestRotatingCenterHeavenGanJoinsGanInteraction(t *testing.T) {
	chart := computeMethodChartForTest(
		t, "2024-12-26 12:00",
		Method{Scope: ScopeDay, School: SchoolZhuanPan, Dingju: DingjuChaiBu},
	)
	center := chart.Pan.GongWei[int(GongZhong)-1]
	if center.TianPanGan == nil {
		t.Fatal("rotating center lacks its heaven-stem fact")
	}
	found := false
	for _, interaction := range chart.GanInteractions {
		if interaction.Gong == GongZhong && interaction.DiPanGan == *center.TianPanGan &&
			interaction.TianPanGan == *center.TianPanGan {
			found = true
			break
		}
	}
	if !found {
		t.Fatalf("center %s+%s has no table interaction", center.DiPanGan, *center.TianPanGan)
	}
}

func patternPalaceByName(patterns []Pattern, name string) GongIndex {
	for _, pattern := range patterns {
		if pattern.Name != name {
			continue
		}
		if len(pattern.GongWei) == 0 {
			return 0
		}
		return pattern.GongWei[0]
	}
	return 0
}

func patternNames(patterns []Pattern) map[string]bool {
	result := make(map[string]bool, len(patterns))
	for _, pattern := range patterns {
		result[pattern.Name] = true
	}
	return result
}
