package mcpcontract

import (
	"context"
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"reflect"
	"slices"
	"sort"
	"strings"
	"testing"

	"liki-engine/internal/engine/qimen"
)

// These tests replace the legacy raw-JSON handler tests for the tools that
// moved to the typed registry. They cover behaviour that the wire DTOs cannot
// express: handler-level validation, engine defaults, and enum closure against
// the engine's own tables.

const (
	canonicalInstantA = "2000-06-15T04:00:00Z"
	canonicalInstantB = "2000-09-20T00:30:00Z"
)

// callToolDocument calls a tool and returns the decoded JSON document.
func callToolDocument(t *testing.T, registry *Registry, name string, input any) map[string]any {
	t.Helper()
	raw, err := registry.Call(context.Background(), name, mustJSON(t, input))
	if err != nil {
		t.Fatalf("%s: %v", name, err)
	}
	if err := registry.ValidateOutput(name, raw); err != nil {
		t.Fatalf("%s output contract: %v", name, err)
	}
	var document map[string]any
	decode(t, raw, &document)
	return document
}

// assertToolError asserts a tool rejects the input, optionally with a message
// fragment.
func assertToolError(t *testing.T, registry *Registry, name string, input any, contains string) {
	t.Helper()
	_, err := registry.Call(context.Background(), name, mustJSON(t, input))
	if err == nil {
		t.Errorf("%s: expected error for %v", name, input)
		return
	}
	if contains != "" && !strings.Contains(err.Error(), contains) {
		t.Errorf("%s: error %q does not mention %q", name, err, contains)
	}
}

// testBaziChart returns a bazi.chart result document for the migrated handlers.
func testBaziChart(t *testing.T, registry *Registry, utc, timezone string, longitude float64, gender string) map[string]any {
	t.Helper()
	return callToolDocument(t, registry, "bazi_chart", map[string]any{
		"instant": utc, "timezone": timezone, "longitude": longitude, "gender": gender,
	})
}

func TestMigratedToolsRejectUnknownArguments(t *testing.T) {
	registry := testEngine()
	solarTime := canonicalInstantA

	// Every migrated tool validates against its closed DTO schema: unknown
	// arguments are the caller's mistake and must surface as a tool error, not
	// silently pass. LLM consumers learn the contract from inputSchema.
	if _, err := registry.Call(context.Background(), "bazi_chart", mustJSON(t, map[string]any{
		"instant": solarTime, "timezone": "Asia/Shanghai", "longitude": 116.4, "gender": "male", "legacy_extra": true,
	})); err == nil {
		t.Fatalf("bazi.chart: unknown arguments must be rejected")
	}

	// Flow tools consume the ziwei.chart output object, exactly as documented.
	ziweiChart := callToolDocument(t, registry, "ziwei_chart", map[string]any{
		"instant": "1984-02-16T04:00:00Z", "timezone": "Asia/Shanghai", "longitude": 116.4, "gender": "male",
	})
	if _, err := registry.Call(context.Background(), "ziwei_liunian", mustJSON(t, map[string]any{"chart": ziweiChart, "lunar_year": 2030})); err != nil {
		t.Fatalf("ziwei.liunian: %v", err)
	}
	if _, err := registry.Call(context.Background(), "ziwei_daxian", mustJSON(t, map[string]any{"chart": ziweiChart})); err != nil {
		t.Fatalf("ziwei.daxian: %v", err)
	}

	// qimen.chart's legacy schema did close the root object, so unknown root
	// arguments stay rejected there.
	assertQimenChartError(t, registry, map[string]any{
		"instant": solarTime, "timezone": "Asia/Shanghai", "longitude": 116.4, "kind": "hour",
	}, "")
}

func TestMigratedBaziFullChartRejectsMissingOrWrongNaYin(t *testing.T) {
	registry := testEngine()
	chart := testBaziChart(t, registry, "1984-02-15T00:00:00Z", "Asia/Shanghai", 116.4, "male")
	nian, ok := chart["nian"].(map[string]any)
	if !ok {
		t.Fatalf("chart.nian = %#v", chart["nian"])
	}
	original, _ := nian["na_yin"].(string)
	if original == "" {
		t.Fatal("bazi.chart returned empty nian.na_yin")
	}

	nian["na_yin"] = ""
	assertToolError(t, registry, "bazi_fullchart", map[string]any{"chart": chart}, "chart.nian.na_yin is required")

	nian["na_yin"] = "错误纳音"
	assertToolError(t, registry, "bazi_fullchart", map[string]any{"chart": chart}, original)

	nian["na_yin"] = original
	full := callToolDocument(t, registry, "bazi_fullchart", map[string]any{"chart": chart})
	if len(full["nian"].(map[string]any)) == 0 {
		t.Fatalf("bazi.fullchart returned an empty nian pillar: %#v", full["nian"])
	}
}

func TestMigratedBaziXiaoYunDefaultCount(t *testing.T) {
	registry := testEngine()
	chart := testBaziChart(t, registry, "1984-02-15T00:00:00Z", "Asia/Shanghai", 116.4, "male")

	// bazi.xiaoyun returns one entry per 小运 school; every entry carries the
	// same number of 柱 as the requested count (12 by default).
	raw, err := registry.Call(context.Background(), "bazi_xiaoyun", mustJSON(t, map[string]any{"chart": chart}))
	if err != nil {
		t.Fatalf("bazi.xiaoyun: %v", err)
	}
	var sets []any
	decode(t, raw, &sets)
	if err := registry.ValidateOutput("bazi_xiaoyun", raw); err != nil {
		t.Fatalf("bazi.xiaoyun output contract: %v", err)
	}
	if len(sets) == 0 {
		t.Fatal("bazi.xiaoyun returned no schools")
	}
	for i, entry := range sets {
		set, _ := entry.(map[string]any)
		zhus, _ := set["zhus"].([]any)
		if len(zhus) != 1 {
			t.Fatalf("school %d default xiaoyun count = %d, want 1 (unified series default)", i, len(zhus))
		}
	}

	raw, err = registry.Call(context.Background(), "bazi_xiaoyun", mustJSON(t, map[string]any{"chart": chart, "count": 5}))
	if err != nil {
		t.Fatalf("bazi.xiaoyun(count=5): %v", err)
	}
	decode(t, raw, &sets)
	if err := registry.ValidateOutput("bazi_xiaoyun", raw); err != nil {
		t.Fatalf("bazi.xiaoyun(count=5) output contract: %v", err)
	}
	for i, entry := range sets {
		set, _ := entry.(map[string]any)
		zhus, _ := set["zhus"].([]any)
		if len(zhus) != 5 {
			t.Fatalf("school %d xiaoyun count = %d, want 5", i, len(zhus))
		}
	}
}

func TestMigratedQimenCanonicalMethodAndDingju(t *testing.T) {
	registry := testEngine()

	method := callQimenChart(t, registry, map[string]any{
		"instant": "1984-02-15T00:00:00Z", "timezone": "Asia/Shanghai", "longitude": 116.4,
	})["method"].(map[string]any)
	if method["scope"] != "hour" || method["school"] != "zhuanpan" || method["dingju_method"] != "chaibu" {
		t.Fatalf("canonical method = %v/%v/%v, want hour/zhuanpan/chaibu",
			method["scope"], method["school"], method["dingju_method"])
	}
	assertToolError(t, registry, "qimen_chart", map[string]any{
		"instant": "1984-02-15T00:00:00Z", "timezone": "Asia/Shanghai", "longitude": 116.4,
	}, "method")

	method = callQimenChart(t, registry, map[string]any{
		"instant": "2026-01-05T05:00:00Z", "timezone": "Asia/Shanghai", "longitude": 116.4, "dingju_method": "zhirun",
	})["method"].(map[string]any)
	if method["dingju_method"] != "zhirun" || method["dingju_method_name"] != "置闰定局" {
		t.Fatalf("dingju method = %v/%v, want zhirun/置闰定局",
			method["dingju_method"], method["dingju_method_name"])
	}
	if method["jie_qi"] != "冬至" || method["yong_ju_jie_qi"] != "小寒" {
		t.Fatalf("terms = %v/%v, want 冬至/小寒", method["jie_qi"], method["yong_ju_jie_qi"])
	}

	assertQimenChartError(t, registry, map[string]any{
		"instant": "2026-01-05T05:00:00Z", "timezone": "Asia/Shanghai", "longitude": 116.4, "dingju_method": "luoshu_feipan",
	}, "")
}

func TestMigratedQimenMethodMatrix(t *testing.T) {
	registry := testEngine()
	const solarTime = "2026-01-13T17:03:00Z"

	document := callQimenChart(t, registry, map[string]any{
		"instant": solarTime, "timezone": "Asia/Shanghai", "longitude": 116.4, "scope": "hour", "school": "luoshu_feipan", "dingju_method": "zhirun",
	})
	method := document["method"].(map[string]any)
	if method["scope"] != "hour" || method["school"] != "luoshu_feipan" || method["dingju_method"] != "zhirun" {
		t.Fatalf("method = %v/%v/%v", method["scope"], method["school"], method["dingju_method"])
	}
	if _, exists := document["board"].(map[string]any)["school"]; exists {
		t.Fatal("board school duplicates method.school")
	}

	invalid := []map[string]any{
		{"scope": "ke"},
		{"school": "yinpan"},
		{"scope": "year", "dingju_method": "chaibu"},
		{"scope": "day", "school": "mingfa_feipan"},
		{"scope": "hour", "school": "mingfa_feipan", "dingju_method": "zhirun"},
		{"kind": "hour"},
		{"plate": "zhuanpan"},
		{"ju_method": "chaibu"},
		{"yong_shen": []string{"甲"}},
		{"yong_shen": []string{}},
		{"yong_shen": []string{"生门", "生门"}},
		{"birth_year": 1984},
	}
	for _, params := range invalid {
		params["instant"] = "2024-06-15T04:00:00Z"
		params["timezone"] = "Asia/Shanghai"
		params["longitude"] = 116.4
		assertQimenChartError(t, registry, params, "")
	}
}

func TestMigratedQimenBirthDateAndYongShenPaths(t *testing.T) {
	if _, err := qimen.ParseBirthDate("2024-02-04"); err == nil {
		t.Error("date-only Lichun boundary should require a time")
	}
	if _, err := qimen.ParseBirthDate("2024-02-04T02:00:00Z"); err != nil {
		t.Fatalf("RFC3339 Lichun boundary should parse: %v", err)
	}

	registry := testEngine()
	document := callQimenChart(t, registry, map[string]any{
		"instant": "1984-02-15T00:00:00Z", "timezone": "Asia/Shanghai", "longitude": 116.4, "yong_shen": []string{"生门", "戊"}, "birth_date": "1984-02-15",
	})
	if document["yong_shen"] == nil {
		t.Fatal("explicit yong_shen was not resolved")
	}
	assertJiaStaysHidden(t, document)

	assertQimenChartError(t, registry, map[string]any{
		"instant": "1984-02-15T00:00:00Z", "timezone": "Asia/Shanghai", "longitude": 116.4, "yong_shen": []string{"不存在的符号"},
	}, "")
}

// assertJiaStaysHidden keeps 甲 off the chart: 甲 never exposes itself as a
// 天盘/地盘/暗干 symbol, it only appears through its 值符 role.
func assertJiaStaysHidden(t *testing.T, data map[string]any) {
	t.Helper()
	board, ok := data["board"].(map[string]any)
	if !ok {
		t.Fatalf("board = %#v", data["board"])
	}
	for _, raw := range board["gong_wei"].([]any) {
		palace := raw.(map[string]any)
		for _, field := range []string{"di_pan_gan", "an_gan", "tian_pan_gan"} {
			if palace[field] == "甲" {
				t.Fatalf("%s exposes 甲", field)
			}
		}
		if tianPan, ok := palace["tian_pan"].([]any); ok {
			for _, raw := range tianPan {
				if raw.(map[string]any)["gan"] == "甲" {
					t.Fatal("tian_pan exposes 甲")
				}
			}
		}
	}
	if interactions, ok := data["gan_interaction"].([]any); ok {
		for _, raw := range interactions {
			interaction := raw.(map[string]any)
			if interaction["di_pan_gan"] == "甲" || interaction["tian_pan_gan"] == "甲" {
				t.Fatal("gan_interaction exposes 甲")
			}
		}
	}
	if yongShen, ok := data["yong_shen"].(map[string]any); ok {
		for _, raw := range yongShen["symbols"].([]any) {
			for _, gan := range raw.(map[string]any)["tian_gan"].([]any) {
				if gan == "甲" {
					t.Fatal("yong_shen tian_gan exposes 甲")
				}
			}
		}
	}
}

// TestMigratedShenShaEnumCoversRealOutput pins the declared shensha enum to
// the engine's own lookup tables. The 流日/流月/流年 shensha names depend on the
// flow pillar's ganzhi (流日 文昌/天乙贵人 are 日干 lookups), so a single sample
// date is not enough: the sweep below walks a full 十干 cycle of days and every
// month, which is the smallest input set that reaches every producer.
func TestMigratedShenShaEnumCoversRealOutput(t *testing.T) {
	registry := testEngine()
	chart := testBaziChart(t, registry, "1984-02-15T00:00:00Z", "Asia/Shanghai", 116.4, "male")

	type call struct {
		name   string
		params map[string]any
	}
	calls := []call{}
	for _, year := range []int{2025, 2026, 2027} {
		calls = append(calls, call{"bazi_liunian", map[string]any{"chart": chart, "year": year}})
	}
	for month := 1; month <= 12; month++ {
		calls = append(calls, call{"bazi_liuyue", map[string]any{"chart": chart, "year": 2026, "month": month}})
	}
	for day := 1; day <= 30; day++ {
		calls = append(calls, call{"bazi_liuri", map[string]any{"chart": chart, "year": 2026, "month": 6, "day": day}})
	}

	for _, item := range calls {
		allowed := enumAt(t, outputProperties(t, registry, item.name), "shensha", "[]", "name")
		document := callToolDocument(t, registry, item.name, item.params)
		items, _ := document["shensha"].([]any)
		for _, raw := range items {
			entry, ok := raw.(map[string]any)
			if !ok {
				continue
			}
			shenSha, _ := entry["name"].(string)
			if !slices.Contains(allowed, shenSha) {
				t.Errorf("%s %v returned shensha %q outside the declared enum %v",
					item.name, item.params, shenSha, allowed)
			}
		}
	}
}

func TestMigratedShiShenEnumMatchesTenGodClosure(t *testing.T) {
	registry := testEngine()
	want := []string{
		"比肩", "劫财", "食神", "伤官", "正财",
		"偏财", "七杀", "正官", "偏印", "正印",
	}
	for _, name := range []string{"bazi_liunian", "bazi_liuyue", "bazi_liuri", "bazi_liushi"} {
		allowed := enumAt(t, outputProperties(t, registry, name), "shi_shen")
		if len(allowed) != len(want) {
			t.Fatalf("%s: shi_shen enum = %v, want the %d-value 十神 closure", name, allowed, len(want))
		}
		for _, value := range want {
			if !slices.Contains(allowed, value) {
				t.Errorf("%s: shi_shen enum is missing %q (%v)", name, value, allowed)
			}
		}
	}
}

func TestMigratedQimenPatternEnumMatchesTable(t *testing.T) {
	rawTable, err := os.ReadFile(filepath.Join(
		"..", "engine", "qimen", "data", "patterns.json",
	))
	if err != nil {
		t.Fatal(err)
	}
	var table []struct {
		Name string `json:"name"`
	}
	if err := json.Unmarshal(rawTable, &table); err != nil {
		t.Fatal(err)
	}
	want := make([]string, 0, len(table))
	for _, rule := range table {
		want = append(want, rule.Name)
	}
	sort.Strings(want)

	registry := testEngine()
	got := enumAt(t, outputProperties(t, registry, "qimen_chart"), "patterns", "[]", "name")
	sort.Strings(got)
	if !reflect.DeepEqual(got, want) {
		t.Fatalf("patterns name enum = %v, want engine table %v", got, want)
	}
}

func TestMigratedZiweiLiuYueTargetLunarContract(t *testing.T) {
	registry := testEngine()
	chart := callToolDocument(t, registry, "ziwei_chart", map[string]any{
		"instant": "2000-09-20T00:00:00Z", "timezone": "Asia/Shanghai", "longitude": 116.4,
		"gender": "male",
	})

	document := callToolDocument(t, registry, "ziwei_liuyue", map[string]any{
		"chart": chart, "target_lunar": map[string]any{"year": 2025, "month": 6, "day": 16, "leap": true},
	})
	period, ok := document["resolved_period"].(map[string]any)
	if !ok {
		t.Fatalf("resolved_period = %#v", document["resolved_period"])
	}
	flowMonth := period["flow_month"].(map[string]any)
	if flowMonth["year"] != float64(2025) || flowMonth["month"] != float64(7) || flowMonth["leap"] != false {
		t.Fatalf("flow_month = %#v, want 2025 month 7 non-leap", flowMonth)
	}

	rejected := []map[string]any{
		// The old bare month input is gone: the exact lunar day is required.
		{"chart": chart, "lunar_year": 2025, "lunar_month": 6},
		{"chart": chart, "target_lunar": map[string]any{"year": 2025, "month": 6, "leap": true}},
		{"chart": chart, "target_lunar": map[string]any{"year": 2024, "month": 6, "day": 10, "leap": true}},
		{"chart": chart, "target_lunar": map[string]any{"year": 2025, "month": 6, "day": 31, "leap": true}},
	}
	for _, params := range rejected {
		assertToolError(t, registry, "ziwei_liuyue", params, "")
	}
}

func TestBondInputsShareShape(t *testing.T) {
	registry := NewEngine()
	for _, tool := range []string{"bazi_bond", "ziwei_bond"} {
		definition, err := registry.Definition(tool)
		if err != nil {
			t.Fatalf("%s: missing", tool)
		}
		var schema struct {
			Properties map[string]struct {
				Properties map[string]any `json:"properties"`
				Required   []string       `json:"required"`
			} `json:"properties"`
			Required []string `json:"required"`
		}
		if err := json.Unmarshal(definition.InputSchema, &schema); err != nil {
			t.Fatalf("%s: decode schema: %v", tool, err)
		}
		for _, side := range []string{"a", "b"} {
			entry, ok := schema.Properties[side]
			if !ok {
				t.Fatalf("%s: missing side %s", tool, side)
			}
			if _, ok := entry.Properties["chart"]; !ok {
				t.Errorf("%s.%s: must wrap the chart as {chart}", tool, side)
			}
			if len(entry.Required) != 1 || entry.Required[0] != "chart" {
				t.Errorf("%s.%s required = %v, want [chart]", tool, side, entry.Required)
			}
		}
	}
}

func TestMigratedChartsRejectInvalidLongitude(t *testing.T) {
	registry := testEngine()
	for _, tool := range []string{"bazi_chart", "ziwei_chart"} {
		assertToolError(t, registry, tool, map[string]any{
			"instant": "2024-06-15T04:00:00Z", "timezone": "Asia/Shanghai",
			"longitude": 500,
			"gender":    "male",
		}, "longitude")
	}
	assertQimenChartError(t, registry, map[string]any{
		"instant": "2024-06-15T04:00:00Z", "timezone": "Asia/Shanghai",
		"longitude": 500,
	}, "longitude")
}

func TestChartsNormalizeEquivalentOffsetForms(t *testing.T) {
	registry := testEngine()
	canonical := testBaziChart(t, registry, "1981-08-25T16:20:00Z", "Asia/Shanghai", 132.044137, "male")
	offset := testBaziChart(t, registry, "1981-08-26T00:20:00+08:00", "Asia/Shanghai", 132.044137, "male")

	// The two wire forms identify the same instant. Offset is a serialization
	// detail, never a timezone, so the full domain documents must be identical.
	if !reflect.DeepEqual(canonical, offset) {
		t.Fatalf("offset form changed chart result: canonical=%#v offset=%#v", canonical, offset)
	}
}

func TestMigratedBaziChartTrueSolarChain(t *testing.T) {
	registry := testEngine()
	resolved, err := TianwenResolve(context.Background(), TianwenResolveInput{
		Instant: "1981-08-25T16:20:00Z", Timezone: "Asia/Shanghai",
		Longitude: 132.044137,
	})
	if err != nil {
		t.Fatalf("tianwen.resolve: %v", err)
	}
	if len(resolved.Candidates) == 0 {
		t.Fatal("tianwen.resolve returned no candidates")
	}
	candidate := resolved.Candidates[0]
	if len(candidate.Reasons) != 1 || candidate.Reasons[0].Domain != "input" {
		t.Fatalf("input candidate reasons = %+v", candidate.Reasons)
	}

	chart := callToolDocument(t, registry, "bazi_chart", map[string]any{
		"instant": string(candidate.Instant), "timezone": candidate.Timezone, "longitude": 132.044137, "gender": "male",
	})
	want := []string{"辛酉", "丙申", "丙子", "己丑"}
	for index, pillar := range []string{"nian", "yue", "ri", "shi"} {
		entry := chart[pillar].(map[string]any)
		if got := fmt.Sprintf("%v%v", entry["gan"], entry["zhi"]); got != want[index] {
			t.Fatalf("%s = %s, want %s", pillar, got, want[index])
		}
	}
}

// outputProperties returns a tool's declared output schema properties.
func outputProperties(t *testing.T, registry *Registry, name string) map[string]any {
	t.Helper()
	definition, err := registry.Definition(name)
	if err != nil {
		t.Fatalf("%s: %v", name, err)
	}
	if len(definition.OutputSchema) == 0 {
		t.Fatalf("%s declares no output schema", name)
	}
	var schema map[string]any
	decode(t, definition.OutputSchema, &schema)
	// Series tools wrap the single-output schema in a oneOf/anyOf union.
	for _, key := range []string{"anyOf", "oneOf"} {
		if branches, ok := schema[key].([]any); ok && len(branches) > 0 {
			if first, ok := branches[0].(map[string]any); ok {
				schema = first
				break
			}
		}
	}
	properties, _ := schema["properties"].(map[string]any)
	if properties == nil {
		t.Fatalf("%s output schema has no properties", name)
	}
	return properties
}

// enumAt walks a property path and returns the declared enum values. The "[ ]"
// segment descends into array items.
func enumAt(t *testing.T, properties map[string]any, path ...string) []string {
	t.Helper()
	var node any = properties
	for _, segment := range path {
		current, ok := node.(map[string]any)
		if !ok {
			t.Fatalf("enumAt(%v): %q parent is not an object", path, segment)
		}
		if segment == "[]" {
			node = current["items"]
			continue
		}
		// The first segment indexes the caller-supplied properties map; nested
		// segments must descend through the schema's own "properties".
		if child, ok := current[segment]; ok {
			node = child
			continue
		}
		nested, _ := current["properties"].(map[string]any)
		node = nested[segment]
	}
	schema, ok := node.(map[string]any)
	if !ok {
		t.Fatalf("enumAt(%v): path does not resolve to a schema", path)
	}
	values, _ := schema["enum"].([]any)
	if len(values) == 0 {
		t.Fatalf("enumAt(%v): no enum declared", path)
	}
	out := make([]string, 0, len(values))
	for _, value := range values {
		text, ok := value.(string)
		if !ok {
			t.Fatalf("enumAt(%v): non-string enum value %#v", path, value)
		}
		out = append(out, text)
	}
	return out
}
