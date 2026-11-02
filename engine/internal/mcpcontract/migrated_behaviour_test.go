package mcpcontract

import (
	"context"
	"fmt"
	"testing"
	"time"
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
	subject := testBaziChart(t, "1984-02-15T00:00:00Z", "Asia/Shanghai", 116.4, "male")
	ziweiSubject := map[string]any{
		"birth_time":      "2000-09-20T08:00:00+08:00",
		"birth_longitude": 116.4,
		"gender":          "male",
	}

	cases := []struct {
		name   string
		tool   string
		input  any
		expect string
	}{
		{"bazi.chart missing gender", "bazi_chart", map[string]any{"subject": map[string]any{"birth_time": "1984-02-15T08:00:00+08:00", "birth_longitude": 116.4}}, ""},
		{"bazi.chart bad gender", "bazi_chart", map[string]any{"subject": map[string]any{"birth_time": "1984-02-15T08:00:00+08:00", "birth_longitude": 116.4, "gender": "other"}}, ""},
		{"ziwei.chart bad gender", "ziwei_chart", map[string]any{"subject": map[string]any{
			"birth_time": "2000-09-20T08:00:00+08:00", "birth_longitude": 116.4, "gender": "bad"}}, ""},
		{"bazi.bond missing b", "bazi_bond", map[string]any{"a": map[string]any{"subject": subject}}, ""},
		{"bazi.liunian missing subject", "bazi_liunian", map[string]any{"liu_nian_start_time": "2026-06-15T12:00:00+08:00"}, ""},
		{"bazi.liunian year below range", "bazi_liunian", map[string]any{"liu_nian_start_time": "0000-01-01T08:00:00", "subject": subject}, ""},
		{"bazi.liuyue missing subject", "bazi_liuyue", map[string]any{"liu_yue_start_time": "2026-06-15T12:00:00+08:00"}, ""},
		{"bazi.liushi invalid start_time", "bazi_liushi", map[string]any{
			"liu_shi_start_time": "not-a-time", "qi_ju_longitude": 116.4, "subject": subject}, ""},
		{"bazi.liushi invalid longitude", "bazi_liushi", map[string]any{
			"liu_shi_start_time": "2026-06-15T12:00:00+08:00", "liu_shi_longitude": 999, "subject": subject}, ""},
		{"qimen.chart bad scope", "qimen_chart", map[string]any{
			"qi_ju":  map[string]any{"time": "1984-02-15T08:00:00+08:00", "longitude": 116.4},
			"method": map[string]any{"scope": "ke", "school": "zhuanpan", "dingju_method": "chaibu"},
		}, ""},
		{"qimen.chart unknown yong_shen", "qimen_chart", map[string]any{
			"qi_ju":     map[string]any{"time": "1984-02-15T08:00:00+08:00", "longitude": 116.4},
			"method":    map[string]any{"scope": "hour", "school": "zhuanpan", "dingju_method": "chaibu"},
			"yong_shen": []string{"不存在的符号"}}, ""},
		{"qimen.chart unknown field", "qimen_chart", map[string]any{
			"qi_ju":      map[string]any{"time": "1984-02-15T08:00:00+08:00", "longitude": 116.4},
			"method":     map[string]any{"scope": "hour", "school": "zhuanpan", "dingju_method": "chaibu"},
			"birth_year": 1984,
		}, ""},
		{"qimen.chart invalid utc", "qimen_chart", map[string]any{
			"qi_ju": map[string]any{"time": "not-a-time"}, "method": map[string]any{
				"scope": "hour", "school": "zhuanpan", "dingju_method": "chaibu",
			},
		}, ""},
		{"ziwei.liuri missing start_time", "ziwei_liuri", map[string]any{"subject": ziweiSubject}, ""},
		{"ziwei.bond missing b", "ziwei_bond", map[string]any{"a": map[string]any{"subject": ziweiSubject}}, ""},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			assertToolError(t, registry, tc.tool, tc.input, tc.expect)
		})
	}

	// Malformed JSON must not panic or be silently accepted.
	for _, tool := range []string{
		"bazi_chart", "bazi_bond", "bazi_liunian", "bazi_liuyue", "bazi_liuri",
		"bazi_liushi", "bazi_xiaoyun", "ziwei_chart",
		"ziwei_liunian", "ziwei_liuyue", "ziwei_liuri", "ziwei_liushi", "ziwei_bond", "qimen_chart",
	} {
		if _, err := registry.Call(context.Background(), tool, []byte("{bad")); err == nil {
			t.Errorf("%s accepted malformed JSON", tool)
		}
	}
}

// TestMigratedBaziChartCompleteFacts asserts the canonical chart is complete.
func TestMigratedBaziChartCompleteFacts(t *testing.T) {
	registry := testEngine()
	subject := testBaziChart(t, "1984-02-15T00:00:00Z", "Asia/Shanghai", 116.4, "male")
	chart := callToolDocument(t, registry, "bazi_chart", map[string]any{"subject": subject})

	for _, field := range []string{
		"san_yuan", "guan_xi", "ten_god_states", "element_states", "da_yun",
	} {
		if value, ok := chart[field]; !ok || value == nil {
			t.Errorf("bazi.chart missing %s", field)
		}
	}
	for _, pillar := range []string{"nian", "yue", "ri", "shi"} {
		item, ok := chart[pillar].(map[string]any)
		if !ok {
			t.Fatalf("bazi.chart %s = %#v", pillar, chart[pillar])
		}
		if item["na_yin"] == nil || item["na_yin"] == "" {
			t.Errorf("bazi.chart %s.na_yin is empty", pillar)
		}
	}

	relations, ok := chart["na_yin_rel"].([]any)
	if !ok || len(relations) != 6 {
		t.Fatalf("na_yin_rel = %#v, want 6 entries", chart["na_yin_rel"])
	}
	sameOnly := true
	for _, raw := range relations {
		if raw.(map[string]any)["relation"] != "相同" {
			sameOnly = false
			break
		}
	}
	if sameOnly {
		t.Fatalf("na_yin_rel are all 相同: %#v", relations)
	}
}

// TestMigratedBaziLiuNianShiShen asserts the 十神 closure is populated for
// flowing years.
func TestMigratedBaziLiuNianShiShen(t *testing.T) {
	registry := testEngine()
	chart := testBaziChart(t, "1984-02-15T00:00:00Z", "Asia/Shanghai", 116.4, "male")
	for _, year := range []int{2024, 2025, 2026} {
		document := callToolDocument(t, registry, "bazi_liunian", map[string]any{
			"liu_nian_start_time": fmt.Sprintf("%04d-06-15T12:00:00+08:00", year), "subject": chart,
		})
		item := firstSeriesItem(t, document)
		if shiShen, _ := item["shi_shen"].(string); shiShen == "" {
			t.Errorf("bazi.liunian(%d) shi_shen is empty", year)
		}
	}
}

// TestMigratedZiweiDependentTools covers subject-dependent ziwei tools.
func TestMigratedZiweiDependentTools(t *testing.T) {
	registry := testEngine()
	chart := map[string]any{
		"birth_time": "2000-09-20T08:00:00+08:00", "birth_longitude": 116.4, "gender": "male",
	}
	chartB := map[string]any{
		"birth_time": "2000-07-16T12:00:00+08:00", "birth_longitude": 116.4, "gender": "female",
	}

	liunian := callToolDocument(t, registry, "ziwei_liunian", map[string]any{"liu_nian_start_time": "2026-06-15T12:00:00+08:00", "subject": chart})
	if len(liunian) == 0 {
		t.Fatal("ziwei.liunian returned an empty document")
	}
	liuyue := callToolDocument(t, registry, "ziwei_liuyue", map[string]any{
		"liu_yue_start_time": civilDateForLunar(t, 2026, 5, 10, false) + "T12:00:00+08:00", "subject": chart,
	})
	if len(liuyue) == 0 {
		t.Fatal("ziwei.liuyue returned an empty document")
	}
	liuri := callToolDocument(t, registry, "ziwei_liuri", map[string]any{
		"liu_ri_start_time": civilDateForLunar(t, 2026, 5, 10, false) + "T12:00:00+08:00", "subject": chart,
	})
	if len(liuri) == 0 {
		t.Fatal("ziwei.liuri returned an empty document")
	}
	liushi := callToolDocument(t, registry, "ziwei_liushi", map[string]any{
		"liu_shi_start_time": "2026-06-24T11:00:00+08:00",
		"liu_shi_longitude":  116.4, "subject": chart,
	})
	if len(liushi) == 0 {
		t.Fatal("ziwei.liushi returned an empty document")
	}
	bond := callToolDocument(t, registry, "ziwei_bond", map[string]any{"a": map[string]any{"subject": chart}, "b": map[string]any{"subject": chartB}})
	if len(bond) == 0 {
		t.Fatal("ziwei.bond returned an empty document")
	}
	ziweiChart := callToolDocument(t, registry, "ziwei_chart", map[string]any{"subject": chart})
	if len(ziweiChart) == 0 {
		t.Fatal("ziwei.chart returned an empty document")
	}
}

// TestMigratedBaziBondAndXiaoYunValid keeps the legacy valid paths for the
// remaining bazi tools.
func TestMigratedBaziBondAndXiaoYunValid(t *testing.T) {
	registry := testEngine()
	chartA := testBaziChart(t, "1984-02-15T00:00:00Z", "Asia/Shanghai", 116.4, "male")
	chartB := testBaziChart(t, canonicalInstantB, "Asia/Shanghai", 116.4, "female")

	bond := callToolDocument(t, registry, "bazi_bond", map[string]any{
		"a": map[string]any{"subject": chartA}, "b": map[string]any{"subject": chartB},
	})
	if len(bond) == 0 {
		t.Fatal("bazi.bond returned an empty document")
	}
	liuyue := callToolDocument(t, registry, "bazi_liuyue", map[string]any{"liu_yue_start_time": "2026-06-15T12:00:00+08:00", "subject": chartA})
	if len(liuyue) == 0 {
		t.Fatal("bazi.liuyue returned an empty document")
	}
	liuri := callToolDocument(t, registry, "bazi_liuri", map[string]any{"liu_ri_start_time": "2026-06-15T12:00:00+08:00", "liu_ri_longitude": 116.4, "subject": chartA})
	if len(liuri) == 0 {
		t.Fatal("bazi.liuri returned an empty document")
	}
	liushi := callToolDocument(t, registry, "bazi_liushi", map[string]any{
		"liu_shi_start_time": "2026-06-15T12:00:00+08:00",
		"liu_shi_longitude":  116.4, "subject": chartA,
	})
	if len(liushi) == 0 {
		t.Fatal("bazi.liushi returned an empty document")
	}
}

// TestMigratedQimenMaoShanContract ports the legacy 茅山定局 contract.
func TestMigratedQimenMaoShanContract(t *testing.T) {
	registry := testEngine()
	document := callQimenChart(t, registry, map[string]any{
		"qi_ju_time": "2024-06-21T06:00:00+08:00", "qi_ju_longitude": 116.4, "scope": "hour",
		"school": "zhuanpan", "dingju_method": "maoshan",
	})
	method := document["method"].(map[string]any)
	if method["dingju_method"] != "maoshan" || method["dingju_method_name"] != "茅山定局" ||
		method["jie_qi"] != "夏至" || method["yong_ju_jie_qi"] != "夏至" {
		t.Fatalf("茅山定局 method = %+v", method)
	}
	assertJiaStaysHidden(t, document)

	for _, input := range []map[string]any{
		{"qi_ju_time": "2024-06-21T06:00:00+08:00", "qi_ju_longitude": 116.4, "scope": "day", "school": "zhuanpan", "dingju_method": "maoshan"},
		{"qi_ju_time": "2024-06-21T06:00:00+08:00", "qi_ju_longitude": 116.4, "scope": "hour", "school": "luoshu_feipan", "dingju_method": "maoshan"},
		{"qi_ju_time": "2024-06-21T06:00:00+08:00", "qi_ju_longitude": 116.4, "scope": "hour", "school": "mingfa_feipan", "dingju_method": "maoshan"},
	} {
		assertQimenChartError(t, registry, input, "")
	}
}

// TestMigratedQimenQuarterRuleContract ports the legacy 刻家奇门 contract for
// both the ten-minute and twelve-minute rules.
func TestMigratedQimenQuarterRuleContract(t *testing.T) {
	registry := testEngine()
	document := callQimenChart(t, registry, map[string]any{
		"qi_ju_time": "2026-09-07T00:00:00+08:00", "qi_ju_longitude": 116.4, "scope": "quarter", "school": "zhuanpan",
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
		"qi_ju_time": "2002-10-08T18:27:00+08:00", "qi_ju_longitude": 116.4, "scope": "quarter",
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
		"qi_ju_time": "2024-06-29T12:00:00+08:00", "qi_ju_longitude": 116.4, "scope": "day", "school": "jinhan_yujing",
	})
	method := document["method"].(map[string]any)
	if method["scope"] != "day" || method["school"] != "jinhan_yujing" ||
		method["dingju_method"] != "none" || method["dingju_method_name"] != "无定局" ||
		method["dun_source"] != "winter_solstice_yang_summer_solstice_yin" {
		t.Fatalf("金函玉镜 method = %+v", method)
	}
	board := document["pan"].(map[string]any)
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
		"men_po", "men_zhi", "ge_ju", "ying_qi", "ri_gan_gong", "shi_gan_gong",
		"kong_wang_affected", "ma_xing_affected", "zhi_fu_xing_gong", "zhi_shi_men_gong",
		"yong_shen",
	} {
		if _, exists := document[field]; exists {
			t.Errorf("金函玉镜 chart must not expose standard analysis field %q", field)
		}
	}

	for _, input := range []map[string]any{
		{"qi_ju_time": "2024-06-29T12:00:00+08:00", "qi_ju_longitude": 116.4, "scope": "hour", "school": "jinhan_yujing"},
		{"qi_ju_time": "2024-06-29T12:00:00+08:00", "qi_ju_longitude": 116.4, "scope": "day", "school": "jinhan_yujing", "dingju_method": "none"},
		{"qi_ju_time": "2024-06-29T12:00:00+08:00", "qi_ju_longitude": 116.4, "scope": "day", "school": "jinhan_yujing", "yong_shen": []string{"太乙"}},
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
			input := map[string]any{"qi_ju_time": standardInputFromInstant(t, stamp, "Asia/Shanghai"), "qi_ju_longitude": 116.4}
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
	resolved, err := TianwenSubjectCandidates(context.Background(), civilResolveInput(t, "2026-03-02T10:30:00Z", "Asia/Shanghai", 116.4))
	if err != nil {
		t.Fatalf("tianwen.resolve: %v", err)
	}
	if len(resolved.Candidates) == 0 || resolved.Candidates[0].Subject.BirthTime == "" {
		t.Fatalf("tianwen.resolve candidates = %+v", resolved.Candidates)
	}
	registry := testEngine()
	document := callQimenChart(t, registry, map[string]any{
		"qi_ju_time": standardInputFromInstant(t, "2026-03-02T10:30:00Z", "Asia/Shanghai"), "qi_ju_longitude": 116.4,
	})
	if document["method"] == nil || document["pan"] == nil {
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
		input["qi_ju_time"] = standardInputFromInstant(t, stamp, "Asia/Shanghai")
		input["qi_ju_longitude"] = 116.4
		assertQimenChartError(t, registry, input, "")
	}

}
