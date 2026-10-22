package mcpcontract

import (
	"context"
	"testing"
	"time"

	"liki-engine/internal/engine/qimen"
)

// These tests port the behaviour coverage that lived in the legacy raw-JSON
// handler tests (engine/internal/agent) before the tools moved to the typed
// registry. They assert handler-level semantics the wire DTOs cannot express:
// input validation, engine defaults, and school-specific chart shapes.

// TestMigratedInputValidation keeps the legacy rejection behaviour: invalid
// JSON, missing or malformed enum fields, and out-of-range values must fail
// before any engine work runs.
func TestMigratedInputValidation(t *testing.T) {
	registry := testEngine()
	chart := testBaziChart(t, registry, "1984-02-15T00:00:00Z", "Asia/Shanghai", 116.4, "male")
	ziweiChart := callToolDocument(t, registry, "ziwei_chart", map[string]any{
		"instant": "2000-09-20T00:00:00Z", "timezone": "Asia/Shanghai", "longitude": 116.4, "gender": "male",
	})

	cases := []struct {
		name   string
		tool   string
		input  any
		expect string
	}{
		{"bazi.chart missing gender", "bazi_chart", map[string]any{"instant": "1984-02-15T00:00:00Z", "timezone": "Asia/Shanghai"}, ""},
		{"bazi.chart bad gender", "bazi_chart", map[string]any{"instant": "1984-02-15T00:00:00Z", "timezone": "Asia/Shanghai", "longitude": 116.4, "gender": "other"}, ""},
		{"ziwei.chart bad gender", "ziwei_chart", map[string]any{
			"instant": "2000-09-20T00:00:00Z", "timezone": "Asia/Shanghai", "longitude": 116.4, "gender": "bad"}, ""},
		{"ziwei.daxian missing chart", "ziwei_daxian", map[string]any{}, ""},
		{"bazi.bond missing b", "bazi_bond", map[string]any{"a": map[string]any{"chart": chart}}, ""},
		{"bazi.liunian missing chart", "bazi_liunian", map[string]any{"year": 2026}, ""},
		{"bazi.liunian year below range", "bazi_liunian", map[string]any{"year": 0, "chart": chart}, ""},
		{"bazi.liuyue missing chart", "bazi_liuyue", map[string]any{"year": 2026, "month": 6}, ""},
		{"bazi.liushi hour below range", "bazi_liushi", map[string]any{
			"year": 2026, "month": 6, "day": 15, "hour": -1, "chart": chart}, ""},
		{"bazi.liushi hour above range", "bazi_liushi", map[string]any{
			"year": 2026, "month": 6, "day": 15, "hour": 24, "chart": chart}, ""},
		{"qimen.chart bad scope", "qimen_chart", map[string]any{
			"instant": "1984-02-15T00:00:00Z", "timezone": "Asia/Shanghai", "longitude": 116.4,
			"method": map[string]any{"scope": "ke", "school": "zhuanpan", "dingju_method": "chaibu"},
		}, ""},
		{"qimen.chart unknown yong_shen", "qimen_chart", map[string]any{
			"instant": "1984-02-15T00:00:00Z", "timezone": "Asia/Shanghai", "longitude": 116.4,
			"method":    map[string]any{"scope": "hour", "school": "zhuanpan", "dingju_method": "chaibu"},
			"yong_shen": []string{"不存在的符号"}}, ""},
		{"qimen.chart unknown field", "qimen_chart", map[string]any{
			"instant": "1984-02-15T00:00:00Z", "timezone": "Asia/Shanghai", "longitude": 116.4,
			"method":     map[string]any{"scope": "hour", "school": "zhuanpan", "dingju_method": "chaibu"},
			"birth_year": 1984,
		}, ""},
		{"qimen.chart invalid utc", "qimen_chart", map[string]any{
			"instant": "not-a-time", "method": map[string]any{
				"scope": "hour", "school": "zhuanpan", "dingju_method": "chaibu",
			},
		}, ""},
		{"ziwei.liuri missing target_lunar", "ziwei_liuri", map[string]any{"chart": ziweiChart}, ""},
		{"ziwei.bond missing b", "ziwei_bond", map[string]any{"a": map[string]any{"chart": ziweiChart}}, ""},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			assertToolError(t, registry, tc.tool, tc.input, tc.expect)
		})
	}

	// Malformed JSON must not panic or be silently accepted.
	for _, tool := range []string{
		"bazi_fullchart", "bazi_bond", "bazi_liunian", "bazi_liuyue", "bazi_liuri",
		"bazi_liushi", "bazi_xiaoyun", "ziwei_chart", "ziwei_fullchart", "ziwei_daxian",
		"ziwei_liunian", "ziwei_liuyue", "ziwei_liuri", "ziwei_liushi", "ziwei_bond", "qimen_chart",
	} {
		if _, err := registry.Call(context.Background(), tool, []byte("{bad")); err == nil {
			t.Errorf("%s accepted malformed JSON", tool)
		}
	}
}

// TestMigratedBaziFullChartExtras keeps the legacy "fullchart is the minimal
// chart plus extras" assertion.
func TestMigratedBaziFullChartExtras(t *testing.T) {
	registry := testEngine()
	chart := testBaziChart(t, registry, "1984-02-15T00:00:00Z", "Asia/Shanghai", 116.4, "male")
	full := callToolDocument(t, registry, "bazi_fullchart", map[string]any{"chart": chart})

	for _, field := range []string{
		"san_yuan", "gan_he", "zhi_liu_he", "ten_god_states", "element_states", "da_yun",
	} {
		if value, ok := full[field]; !ok || value == nil {
			t.Errorf("bazi.fullchart missing %s", field)
		}
	}
	for _, pillar := range []string{"nian", "yue", "ri", "shi"} {
		item, ok := full[pillar].(map[string]any)
		if !ok {
			t.Fatalf("bazi.fullchart %s = %#v", pillar, full[pillar])
		}
		if item["na_yin"] == nil || item["na_yin"] == "" {
			t.Errorf("bazi.fullchart %s.na_yin is empty", pillar)
		}
	}

	relations, ok := full["nayin_rel"].([]any)
	if !ok || len(relations) != 6 {
		t.Fatalf("nayin_rel = %#v, want 6 entries", full["nayin_rel"])
	}
	sameOnly := true
	for _, raw := range relations {
		if raw.(map[string]any)["relation"] != "相同" {
			sameOnly = false
			break
		}
	}
	if sameOnly {
		t.Fatalf("nayin_rel are all 相同: %#v", relations)
	}
}

// TestMigratedBaziLiuNianShiShen asserts the 十神 closure is populated for
// flowing years.
func TestMigratedBaziLiuNianShiShen(t *testing.T) {
	registry := testEngine()
	chart := testBaziChart(t, registry, "1984-02-15T00:00:00Z", "Asia/Shanghai", 116.4, "male")
	for _, year := range []int{2024, 2025, 2026} {
		document := callToolDocument(t, registry, "bazi_liunian", map[string]any{
			"year": year, "chart": chart,
		})
		if shiShen, _ := document["shi_shen"].(string); shiShen == "" {
			t.Errorf("bazi.liunian(%d) shi_shen is empty", year)
		}
	}
}

// TestMigratedZiweiDependentTools keeps the legacy valid-path coverage for the
// chart-dependent ziwei tools.
func TestMigratedZiweiDependentTools(t *testing.T) {
	registry := testEngine()
	chart := callToolDocument(t, registry, "ziwei_chart", map[string]any{
		"instant": "2000-09-20T00:00:00Z", "timezone": "Asia/Shanghai", "longitude": 116.4, "gender": "male",
	})
	chartB := callToolDocument(t, registry, "ziwei_chart", map[string]any{
		"instant": "2000-07-16T04:00:00Z", "timezone": "Asia/Shanghai", "longitude": 116.4, "gender": "female",
	})

	// ziwei.daxian keeps its array wire root.
	daxianRaw, err := registry.Call(context.Background(), "ziwei_daxian", mustJSON(t, map[string]any{"chart": chart}))
	if err != nil {
		t.Fatalf("ziwei.daxian: %v", err)
	}
	var daxian []any
	decode(t, daxianRaw, &daxian)
	if len(daxian) == 0 {
		t.Fatal("ziwei.daxian returned an empty list")
	}
	liunian := callToolDocument(t, registry, "ziwei_liunian", map[string]any{"lunar_year": 2026, "chart": chart})
	if len(liunian) == 0 {
		t.Fatal("ziwei.liunian returned an empty document")
	}
	liuyue := callToolDocument(t, registry, "ziwei_liuyue", map[string]any{
		"target_lunar": map[string]any{"year": 2026, "month": 5, "day": 10, "leap": false}, "chart": chart,
	})
	if len(liuyue) == 0 {
		t.Fatal("ziwei.liuyue returned an empty document")
	}
	liuri := callToolDocument(t, registry, "ziwei_liuri", map[string]any{
		"target_lunar": map[string]any{"year": 2026, "month": 5, "day": 10, "leap": false}, "chart": chart,
	})
	if len(liuri) == 0 {
		t.Fatal("ziwei.liuri returned an empty document")
	}
	liushi := callToolDocument(t, registry, "ziwei_liushi", map[string]any{
		"target_lunar": map[string]any{"year": 2026, "month": 5, "day": 10, "leap": false},
		"shi_zhi":      "午", "chart": chart,
	})
	if len(liushi) == 0 {
		t.Fatal("ziwei.liushi returned an empty document")
	}
	bond := callToolDocument(t, registry, "ziwei_bond", map[string]any{"a": map[string]any{"chart": chart}, "b": map[string]any{"chart": chartB}})
	if len(bond) == 0 {
		t.Fatal("ziwei.bond returned an empty document")
	}
	full := callToolDocument(t, registry, "ziwei_fullchart", map[string]any{"chart": chart})
	if len(full) == 0 {
		t.Fatal("ziwei.fullchart returned an empty document")
	}
}

// TestMigratedBaziBondAndXiaoYunValid keeps the legacy valid paths for the
// remaining bazi tools.
func TestMigratedBaziBondAndXiaoYunValid(t *testing.T) {
	registry := testEngine()
	chartA := testBaziChart(t, registry, "1984-02-15T00:00:00Z", "Asia/Shanghai", 116.4, "male")
	chartB := testBaziChart(t, registry, canonicalInstantB, "Asia/Shanghai", 116.4, "female")

	bond := callToolDocument(t, registry, "bazi_bond", map[string]any{
		"a": map[string]any{"chart": chartA}, "b": map[string]any{"chart": chartB},
	})
	if len(bond) == 0 {
		t.Fatal("bazi.bond returned an empty document")
	}
	liuyue := callToolDocument(t, registry, "bazi_liuyue", map[string]any{"year": 2026, "month": 6, "chart": chartA})
	if len(liuyue) == 0 {
		t.Fatal("bazi.liuyue returned an empty document")
	}
	liuri := callToolDocument(t, registry, "bazi_liuri", map[string]any{"year": 2026, "month": 6, "day": 15, "chart": chartA})
	if len(liuri) == 0 {
		t.Fatal("bazi.liuri returned an empty document")
	}
	liushi := callToolDocument(t, registry, "bazi_liushi", map[string]any{
		"year": 2026, "month": 6, "day": 15, "hour": 12, "chart": chartA,
	})
	if len(liushi) == 0 {
		t.Fatal("bazi.liushi returned an empty document")
	}
}

// TestMigratedQimenMaoShanContract ports the legacy 茅山定局 contract.
func TestMigratedQimenMaoShanContract(t *testing.T) {
	registry := testEngine()
	document := callQimenChart(t, registry, map[string]any{
		"instant": "2024-06-20T22:00:00Z", "timezone": "Asia/Shanghai", "longitude": 116.4, "scope": "hour",
		"school": "zhuanpan", "dingju_method": "maoshan",
	})
	method := document["method"].(map[string]any)
	if method["dingju_method"] != "maoshan" || method["dingju_method_name"] != "茅山定局" ||
		method["jie_qi"] != "夏至" || method["yong_ju_jie_qi"] != "夏至" {
		t.Fatalf("茅山定局 method = %+v", method)
	}
	assertJiaStaysHidden(t, document)

	for _, input := range []map[string]any{
		{"instant": "2024-06-20T22:00:00Z", "timezone": "Asia/Shanghai", "longitude": 116.4, "scope": "day", "school": "zhuanpan", "dingju_method": "maoshan"},
		{"instant": "2024-06-20T22:00:00Z", "timezone": "Asia/Shanghai", "longitude": 116.4, "scope": "hour", "school": "luoshu_feipan", "dingju_method": "maoshan"},
		{"instant": "2024-06-20T22:00:00Z", "timezone": "Asia/Shanghai", "longitude": 116.4, "scope": "hour", "school": "mingfa_feipan", "dingju_method": "maoshan"},
	} {
		assertQimenChartError(t, registry, input, "")
	}
}

// TestMigratedQimenQuarterRuleContract ports the legacy 刻家奇门 contract for
// both the ten-minute and twelve-minute rules.
func TestMigratedQimenQuarterRuleContract(t *testing.T) {
	registry := testEngine()
	document := callQimenChart(t, registry, map[string]any{
		"instant": "2026-09-06T16:00:00Z", "timezone": "Asia/Shanghai", "longitude": 116.4, "scope": "quarter", "school": "zhuanpan",
	})
	method := document["method"].(map[string]any)
	if method["scope"] != "quarter" || method["scope_name"] != "刻家奇门" ||
		method["school"] != "zhuanpan" || method["quarter_rule"] != "ten_minute_sanyuan" ||
		method["quarter_rule_name"] != "十分钟三元刻家" || method["method_source"] != "quarter_rule" {
		t.Fatalf("ten-minute quarter method = %+v", method)
	}
	quarter := method["quarter"].(map[string]any)
	if quarter["index"] != float64(5) || quarter["minutes_per_quarter"] != float64(10) {
		t.Fatalf("quarter metadata = %+v", quarter)
	}
	if lead := method["lead_pillar"].(map[string]any)["name"]; lead != "戊辰" {
		t.Fatalf("quarter lead = %v, want 戊辰", lead)
	}
	if mode := quarter["lead_pillar_mode"]; mode != "quarter" {
		t.Fatalf("quarter lead mode = %v, want quarter", mode)
	}
	assertJiaStaysHidden(t, document)

	twelve := callQimenChart(t, registry, map[string]any{
		"instant": "2002-10-08T10:27:00Z", "timezone": "Asia/Shanghai", "longitude": 116.4, "scope": "quarter",
		"school": "zhuanpan", "quarter_rule": "twelve_minute_ten_division",
	})
	twelveMethod := twelve["method"].(map[string]any)
	if twelveMethod["quarter_rule_name"] != "十二分钟十分局刻家" {
		t.Fatalf("twelve-minute quarter_rule_name = %v", twelveMethod["quarter_rule_name"])
	}
	assertJiaStaysHidden(t, twelve)
}

// TestMigratedQimenJinhanContract ports the legacy 金函玉镜 contract: a
// day-only dedicated chart that omits the standard board/analysis surface.
func TestMigratedQimenJinhanContract(t *testing.T) {
	registry := testEngine()
	document := callQimenChart(t, registry, map[string]any{
		"instant": "2024-06-29T04:00:00Z", "timezone": "Asia/Shanghai", "longitude": 116.4, "scope": "day", "school": "jinhan_yujing",
	})
	method := document["method"].(map[string]any)
	if method["scope"] != "day" || method["school"] != "jinhan_yujing" ||
		method["dingju_method"] != "none" || method["dingju_method_name"] != "无定局" ||
		method["dun_source"] != "winter_solstice_yang_summer_solstice_yin" {
		t.Fatalf("金函玉镜 method = %+v", method)
	}
	board := document["board"].(map[string]any)
	if board["yin_dun"] != true || board["ri_gan"] != "甲" || board["ri_zhi"] != "子" {
		t.Fatalf("金函玉镜 board header = %+v", board)
	}
	for _, field := range []string{
		"jushu", "zhi_fu_xing", "zhi_shi_men", "kong_wang", "ma_xing", "wu_bu_yu_shi",
	} {
		if _, exists := board[field]; exists {
			t.Errorf("金函玉镜 board must not expose standard field %q", field)
		}
	}
	if _, exists := board["day_spirits"]; !exists {
		t.Error("金函玉镜 board lacks day_spirits")
	}
	for _, field := range []string{
		"gan_interaction", "men_interaction", "xing_interaction", "xing_gong_wu_xing",
		"men_po", "men_zhi", "patterns", "ying_qi", "ri_gan_gong", "shi_gan_gong",
		"kong_wang_affected", "ma_xing_affected", "zhi_fu_xing_gong", "zhi_shi_men_gong",
		"yong_shen",
	} {
		if _, exists := document[field]; exists {
			t.Errorf("金函玉镜 chart must not expose standard analysis field %q", field)
		}
	}

	for _, input := range []map[string]any{
		{"instant": "2024-06-29T04:00:00Z", "timezone": "Asia/Shanghai", "longitude": 116.4, "scope": "hour", "school": "jinhan_yujing"},
		{"instant": "2024-06-29T04:00:00Z", "timezone": "Asia/Shanghai", "longitude": 116.4, "scope": "day", "school": "jinhan_yujing", "dingju_method": "none"},
		{"instant": "2024-06-29T04:00:00Z", "timezone": "Asia/Shanghai", "longitude": 116.4, "scope": "day", "school": "jinhan_yujing", "yong_shen": []string{"太乙"}},
	} {
		assertQimenChartError(t, registry, input, "")
	}
}

// TestMigratedQimenBoundaryDates sweeps every public method combination across
// Lichun/solstice boundaries. The registry validates each output, so a run
// without error is the 17-combination result-contract assertion the legacy
// schema test made.
func TestMigratedQimenBoundaryDates(t *testing.T) {
	registry := testEngine()
	methods := []map[string]any{
		{"scope": "hour", "school": "zhuanpan", "dingju_method": "chaibu"},
		{"scope": "hour", "school": "zhuanpan", "dingju_method": "zhirun"},
		{"scope": "hour", "school": "luoshu_feipan", "dingju_method": "chaibu"},
		{"scope": "hour", "school": "luoshu_feipan", "dingju_method": "zhirun"},
		{"scope": "hour", "school": "zhuanpan", "dingju_method": "maoshan"},
		{"scope": "hour", "school": "mingfa_feipan"},
		{"scope": "day", "school": "zhuanpan", "dingju_method": "chaibu"},
		{"scope": "day", "school": "zhuanpan", "dingju_method": "zhirun"},
		{"scope": "day", "school": "luoshu_feipan", "dingju_method": "chaibu"},
		{"scope": "day", "school": "luoshu_feipan", "dingju_method": "zhirun"},
		{"scope": "day", "school": "jinhan_yujing"},
		{"scope": "quarter", "school": "zhuanpan", "quarter_rule": "ten_minute_sanyuan"},
		{"scope": "quarter", "school": "zhuanpan", "quarter_rule": "twelve_minute_ten_division"},
		{"scope": "month", "school": "zhuanpan"},
		{"scope": "month", "school": "luoshu_feipan"},
		{"scope": "year", "school": "zhuanpan"},
		{"scope": "year", "school": "luoshu_feipan"},
	}
	if len(methods) != 17 {
		t.Fatalf("method coverage = %d, want all 17 public combinations", len(methods))
	}

	for day := 1; day <= 365; day += 15 {
		stamp := time.Date(2026, 1, 1, 0, 0, 0, 0, time.FixedZone("CST", 8*3600)).
			AddDate(0, 0, day-1).UTC().Format(time.RFC3339)
		for _, method := range methods {
			input := map[string]any{"instant": stamp, "timezone": "Asia/Shanghai", "longitude": 116.4}
			for key, value := range method {
				input[key] = value
			}
			document := callQimenChart(t, registry, input)
			assertJiaStaysHidden(t, document)
		}
	}
}

// TestMigratedQimenConsumesTianwenGregorian keeps the cross-tool contract that
// qimen.chart accepts the Gregorian candidate produced by tianwen.resolve.
func TestMigratedQimenConsumesTianwenGregorian(t *testing.T) {
	resolved, err := TianwenResolve(context.Background(), TianwenResolveInput{
		Instant: "2026-03-02T10:30:00Z", Timezone: "Asia/Shanghai", Longitude: 116.4,
	})
	if err != nil {
		t.Fatalf("tianwen.resolve: %v", err)
	}
	if len(resolved.Candidates) == 0 || resolved.Candidates[0].Instant == "" {
		t.Fatalf("tianwen.resolve candidates = %+v", resolved.Candidates)
	}
	registry := testEngine()
	document := callQimenChart(t, registry, map[string]any{
		"instant": string(resolved.Candidates[0].Instant), "timezone": resolved.Candidates[0].Timezone, "longitude": 116.4,
	})
	if document["method"] == nil || document["board"] == nil {
		t.Fatalf("qimen.chart from tianwen.resolve = %#v", document)
	}
}

// TestMigratedQimenInvalidMethodCombinations keeps the legacy rejection matrix.
func TestMigratedQimenInvalidMethodCombinations(t *testing.T) {
	registry := testEngine()
	const stamp = "2026-03-02T10:30:00Z"
	for _, input := range []map[string]any{
		{"scope": "year", "dingju_method": "chaibu"},
		{"scope": "day", "school": "mingfa_feipan"},
		{"scope": "hour", "school": "mingfa_feipan", "dingju_method": "zhirun"},
		{"kind": "hour"},
		{"plate": "zhuanpan"},
		{"ju_method": "chaibu"},
		{"yong_shen": []string{"甲"}},
		{"birth_year": 1984},
		{"dingju_method": "luoshu_feipan"},
	} {
		input["instant"] = stamp
		input["longitude"] = 116.4
		assertQimenChartError(t, registry, input, "")
	}

	// The legacy 立春 boundary rule: a bare date cannot resolve a Lichun-crossing
	// birth date, an explicit instant can.
	if _, err := qimen.ParseBirthDate("2024-02-04"); err == nil {
		t.Error("date-only Lichun boundary should require a time")
	}
	if _, err := qimen.ParseBirthDate("2024-02-04T02:00:00Z"); err != nil {
		t.Fatalf("RFC3339 Lichun boundary should parse: %v", err)
	}
	assertQimenChartError(t, registry, map[string]any{
		"instant": stamp, "timezone": "Asia/Shanghai", "longitude": 116.4, "birth_date": "2024-02-04", "yong_shen": []string{"生门"},
	}, "")
}
