package qimen

import (
	"regexp"
	"testing"
)

func TestQimenCatalogContract(t *testing.T) {
	wantScopes := map[Scope]scopeEntry{
		ScopeHour: {
			Scope: "hour", Name: "时家奇门", LeadPillar: "hour",
			MethodSource: "solar_term", DingjuMethods: []DingjuMethod{DingjuChaiBu, DingjuZhiRun, DingjuMaoShan},
			DefaultDingju: DingjuChaiBu, Features: []string{"wu_bu_yu_shi"},
		},
		ScopeDay: {
			Scope: "day", Name: "日家奇门", LeadPillar: "day",
			MethodSource: "solar_term", DingjuMethods: []DingjuMethod{DingjuChaiBu, DingjuZhiRun},
			DefaultDingju:   DingjuChaiBu,
			NoDingjuSchools: []School{SchoolJinhanYuJing},
		},
		ScopeQuarter: {
			Scope: "quarter", Name: "刻家奇门", LeadPillar: "quarter",
			MethodSource:       "quarter_rule",
			QuarterRules:       []QuarterRule{QuarterTenMinuteSanYuan, QuarterTwelveMinuteTenDivision},
			DefaultQuarterRule: QuarterTenMinuteSanYuan,
		},
		ScopeMonth: {Scope: "month", Name: "月家周期奇门", LeadPillar: "month", MethodSource: "month_cycle"},
		ScopeYear:  {Scope: "year", Name: "年家周期奇门", LeadPillar: "year", MethodSource: "year_cycle"},
	}
	if len(scopeMethods) != len(wantScopes) {
		t.Fatalf("scope catalog size = %d, want %d", len(scopeMethods), len(wantScopes))
	}
	for scope, want := range wantScopes {
		got := scopeMethods[scope]
		if got.Scope != want.Scope || got.Name != want.Name || got.LeadPillar != want.LeadPillar ||
			got.MethodSource != want.MethodSource || got.DefaultDingju != want.DefaultDingju ||
			got.DefaultQuarterRule != want.DefaultQuarterRule ||
			len(got.DingjuMethods) != len(want.DingjuMethods) ||
			len(got.QuarterRules) != len(want.QuarterRules) ||
			len(got.Features) != len(want.Features) ||
			len(got.NoDingjuSchools) != len(want.NoDingjuSchools) {
			t.Fatalf("scope %s = %+v, want %+v", scope, got, want)
		}
		for i, method := range want.DingjuMethods {
			if got.DingjuMethods[i] != method || !got.allowsDingju(method) {
				t.Fatalf("scope %s has invalid dingju method %d", scope, i)
			}
		}
		for i, rule := range want.QuarterRules {
			if got.QuarterRules[i] != rule || !got.allowsQuarterRule(rule) {
				t.Fatalf("scope %s has invalid quarter rule %d", scope, i)
			}
		}
		for i, school := range want.NoDingjuSchools {
			if got.NoDingjuSchools[i] != school || !got.allowsNoDingjuSchool(school) {
				t.Fatalf("scope %s has invalid no-dingju school %d", scope, i)
			}
		}
	}

	wantSchools := map[School]schoolEntry{
		SchoolZhuanPan: {
			School: "zhuanpan", Name: "转盘法",
			AllowedScopes:        []Scope{ScopeHour, ScopeDay, ScopeQuarter, ScopeMonth, ScopeYear},
			AllowedDingjuMethods: []DingjuMethod{DingjuChaiBu, DingjuZhiRun, DingjuMaoShan},
			AllowedQuarterRules:  []QuarterRule{QuarterTenMinuteSanYuan, QuarterTwelveMinuteTenDivision},
			SpiritMode:           "eight_spirits_yang_gouchen_zhuque_yin_baihu_xuanwu",
			Geometry:             "luoshu_ring_rotation", StarMode: "tian_qin_lodges_kun",
			DoorMode: "center_projects_kun", SpiritFlight: "yang_forward_yin_reverse",
		},
		SchoolLuoShuFeiPan: {
			School: "luoshu_feipan", Name: "洛书飞盘法",
			AllowedScopes:        []Scope{ScopeHour, ScopeDay, ScopeMonth, ScopeYear},
			AllowedDingjuMethods: []DingjuMethod{DingjuChaiBu, DingjuZhiRun},
			SpiritMode:           "nine_spirits_baihu_taichang_xuanwu",
			Geometry:             "palace_number_shift", StarMode: "star_gan_shift_together",
			DoorMode: "eight_doors_real_center_step", SpiritFlight: "yang_forward_yin_reverse",
		},
		SchoolMingFaFeiPan: {
			School: "mingfa_feipan", Name: "鸣法飞盘",
			AllowedScopes: []Scope{ScopeHour}, AllowedDingjuMethods: []DingjuMethod{DingjuChaiBu},
			SpiritMode: "nine_spirits_gouchen_taichang_zhuque",
			Geometry:   "mingfa_star_door_forward_flight", StarMode: "nine_stars_forward_with_center",
			DoorMode: "nine_doors_with_center_forward", SpiritFlight: "yang_forward_yin_reverse",
		},
		SchoolJinhanYuJing: {
			School: "jinhan_yujing", Name: "金函玉镜",
			AllowedScopes: []Scope{ScopeDay}, AllowedDingjuMethods: []DingjuMethod{DingjuNone},
			SpiritMode: "day_gan_twelve_spirits", Geometry: "solstice_day_star_flight",
			StarMode: "taiyi_nine_stars_day_flight", DoorMode: "three_day_eight_door_rotation",
			SpiritFlight: "solstice_yang_forward_yin_reverse",
		},
	}
	if len(schoolMethods) != len(wantSchools) {
		t.Fatalf("school catalog size = %d, want %d", len(schoolMethods), len(wantSchools))
	}
	for school, want := range wantSchools {
		got := schoolMethods[school]
		if got.School != want.School || got.Name != want.Name || got.SpiritMode != want.SpiritMode ||
			got.Geometry != want.Geometry || got.StarMode != want.StarMode || got.DoorMode != want.DoorMode ||
			got.SpiritFlight != want.SpiritFlight || len(got.AllowedScopes) != len(want.AllowedScopes) ||
			len(got.AllowedDingjuMethods) != len(want.AllowedDingjuMethods) ||
			len(got.AllowedQuarterRules) != len(want.AllowedQuarterRules) {
			t.Fatalf("school %s = %+v, want %+v", school, got, want)
		}
		for i, scope := range want.AllowedScopes {
			if got.AllowedScopes[i] != scope || !got.allowsScope(scope) {
				t.Fatalf("school %s has invalid scope %d", school, i)
			}
		}
		for i, method := range want.AllowedDingjuMethods {
			if got.AllowedDingjuMethods[i] != method || !got.allowsDingju(method) {
				t.Fatalf("school %s has invalid dingju method %d", school, i)
			}
		}
		for i, rule := range want.AllowedQuarterRules {
			if got.AllowedQuarterRules[i] != rule || !got.allowsQuarterRule(rule) {
				t.Fatalf("school %s has invalid quarter rule %d", school, i)
			}
		}
	}

	if dingjuMethodNames[DingjuChaiBu] != "拆补定局" || dingjuMethodNames[DingjuZhiRun] != "置闰定局" ||
		dingjuMethodNames[DingjuMaoShan] != "茅山定局" || dingjuMethodNames[DingjuNone] != "无定局" {
		t.Fatalf("dingju method names = %+v", dingjuMethodNames)
	}
	if quarterRuleNames[QuarterTenMinuteSanYuan] != "十分钟三元刻家" ||
		quarterRuleNames[QuarterTwelveMinuteTenDivision] != "十二分钟十分局刻家" {
		t.Fatalf("quarter rule names = %+v", quarterRuleNames)
	}
	if defaultScope != ScopeHour || defaultSchool != SchoolZhuanPan {
		t.Fatalf("defaults = %s/%s, want hour/zhuanpan", defaultScope, defaultSchool)
	}
	if len(monthGroups) != 12 || yearCycle != 60 || yearAnchor != 1864 {
		t.Fatalf("fixed calendars = month %d, year %d/%d", len(monthGroups), yearCycle, yearAnchor)
	}
}

func TestQimenClosedMethodCatalogCombinations(t *testing.T) {
	valid := [][4]string{
		{"hour", "zhuanpan", "chaibu", ""},
		{"hour", "zhuanpan", "zhirun", ""},
		{"hour", "zhuanpan", "maoshan", ""},
		{"hour", "luoshu_feipan", "chaibu", ""},
		{"hour", "luoshu_feipan", "zhirun", ""},
		{"hour", "mingfa_feipan", "chaibu", ""},
		{"day", "zhuanpan", "chaibu", ""},
		{"day", "zhuanpan", "zhirun", ""},
		{"day", "luoshu_feipan", "chaibu", ""},
		{"day", "luoshu_feipan", "zhirun", ""},
		{"day", "jinhan_yujing", "", ""},
		{"quarter", "zhuanpan", "", "ten_minute_sanyuan"},
		{"quarter", "zhuanpan", "", "twelve_minute_ten_division"},
		{"month", "zhuanpan", "", ""},
		{"month", "luoshu_feipan", "", ""},
		{"year", "zhuanpan", "", ""},
		{"year", "luoshu_feipan", "", ""},
	}
	seen := make(map[Method]bool, len(valid))
	for _, item := range valid {
		method, err := ResolveMethod(item[0], item[1], item[2], item[3])
		if err != nil {
			t.Fatalf("ResolveMethod(%+v): %v", item, err)
		}
		if seen[method] {
			t.Fatalf("duplicate method %+v", method)
		}
		seen[method] = true
	}
	if len(seen) != 17 {
		t.Fatalf("closed method catalog size = %d, want 17", len(seen))
	}
}

func TestPublicQimenDomainAxesDoNotContainProjectNames(t *testing.T) {
	projectNames := regexp.MustCompile(`(?i)horosa|kinqimen|atopx|potuo|mingyu|deminzhang|大师奇门`)
	for scope, entry := range scopeMethods {
		if projectNames.MatchString(scope.String()) || projectNames.MatchString(entry.Name) {
			t.Fatalf("scope %q contains a project name", scope)
		}
		for _, method := range entry.DingjuMethods {
			if projectNames.MatchString(method.String()) || projectNames.MatchString(dingjuMethodNames[method]) {
				t.Fatalf("public dingju method %q contains a project name", method)
			}
		}
		for _, rule := range entry.QuarterRules {
			if projectNames.MatchString(rule.String()) || projectNames.MatchString(quarterRuleNames[rule]) {
				t.Fatalf("public quarter rule %q contains a project name", rule)
			}
		}
	}
	for school, entry := range schoolMethods {
		if projectNames.MatchString(school.String()) || projectNames.MatchString(entry.Name) {
			t.Fatalf("school %q contains a project name", school)
		}
	}
}
