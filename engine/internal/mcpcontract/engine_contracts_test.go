package mcpcontract

import (
	"context"
	"encoding/json"
	"os"
	"reflect"
	"slices"
	"strings"
	"sync"
	"testing"

	"github.com/google/jsonschema-go/jsonschema"

	"liki-engine/internal/engine/tables"
)

func TestMain(m *testing.M) {
	if err := tables.Load(); err != nil {
		panic(err)
	}
	os.Exit(m.Run())
}

// testEngine derives the full typed registry once. Registry is read-only after
// construction, and deriving every JSON Schema in every test dominates the
// package's race/coverage runtime.
var testEngine = sync.OnceValue(NewEngine)

func TestEngineContractDefinitions(t *testing.T) {
	registry := testEngine()
	want := []string{
		"time_now", "huangli_days", "tianwen_resolve", "city_coords",
		"bazi_chart",
		"bazhai_chart", "bazhai_layout",
		"xuankong_chart", "xuankong_liunian",
		"liuyao_qigua",
		"liuyao_chart",
		"bazi_fullchart", "bazi_bond",
		"bazi_liunian", "bazi_liuyue", "bazi_liuri", "bazi_liushi", "bazi_xiaoyun",
		"ziwei_chart", "ziwei_fullchart", "ziwei_daxian", "ziwei_liunian",
		"ziwei_liuyue", "ziwei_liuri", "ziwei_liushi", "ziwei_bond",
		"qimen_chart",
	}
	got := registry.Names()
	if len(got) != len(want) {
		t.Fatalf("contracts = %v, want %v", got, want)
	}
	for i := range want {
		if got[i] != want[i] {
			t.Fatalf("contracts = %v, want %v", got, want)
		}
	}

	for _, name := range want {
		definition, err := registry.Definition(name)
		if err != nil {
			t.Fatalf("%s is not registered", name)
		}
		if definition.Description == "" {
			t.Errorf("%s has no description", name)
		}
		if len(definition.InputSchema) == 0 {
			t.Errorf("%s has no input schema", name)
		}
	}

	// Every output root is declared: object, array, scalar, or composition.
	for _, name := range want {
		definition, err := registry.Definition(name)
		if err != nil {
			t.Fatalf("%s: %v", name, err)
		}
		if len(definition.OutputSchema) == 0 {
			t.Errorf("%s should declare an output schema", name)
		}
	}
}

func TestEngineContractInputConstraints(t *testing.T) {
	registry := testEngine()

	var cityInput struct {
		Type                 string   `json:"type"`
		AdditionalProperties *bool    `json:"additionalProperties"`
		Required             []string `json:"required"`
		Properties           map[string]struct {
			Type      string `json:"type"`
			MinLength int    `json:"minLength"`
			MaxLength int    `json:"maxLength"`
		} `json:"properties"`
	}
	decode(t, mustDefinition(t, registry, "city_coords").InputSchema, &cityInput)
	if cityInput.Type != "object" || cityInput.AdditionalProperties == nil || *cityInput.AdditionalProperties {
		t.Fatalf("city input is not a closed object: %#v", cityInput)
	}
	if len(cityInput.Required) != 1 || cityInput.Required[0] != "city" {
		t.Fatalf("city required = %#v", cityInput.Required)
	}
	if city, ok := cityInput.Properties["city"]; !ok || city.Type != "string" || city.MinLength != 1 || city.MaxLength != 128 {
		t.Fatalf("city schema = %#v", city)
	}

	var resolveInput struct {
		OneOf []struct {
			Required             []string `json:"required"`
			AdditionalProperties *bool    `json:"additionalProperties"`
			Properties           map[string]struct {
				Type   any    `json:"type"`
				Format string `json:"format"`
			} `json:"properties"`
		} `json:"oneOf"`
	}
	decode(t, mustDefinition(t, registry, "tianwen_resolve").InputSchema, &resolveInput)
	if len(resolveInput.OneOf) != 4 {
		t.Fatalf("tianwen.resolve oneOf forms = %d, want 4", len(resolveInput.OneOf))
	}
	for _, form := range resolveInput.OneOf {
		if !slices.Contains(form.Required, "longitude") {
			t.Fatalf("tianwen.resolve form must require longitude: %#v", form.Required)
		}
		if form.AdditionalProperties == nil || *form.AdditionalProperties {
			t.Fatalf("tianwen.resolve form must be closed: %#v", form.Required)
		}
	}

	var huangliInput struct {
		Required   []string `json:"required"`
		Properties map[string]struct {
			Format  string   `json:"format"`
			Minimum *float64 `json:"minimum"`
			Maximum *float64 `json:"maximum"`
			Enum    []any    `json:"enum"`
		} `json:"properties"`
	}
	decode(t, mustDefinition(t, registry, "huangli_days").InputSchema, &huangliInput)
	if len(huangliInput.Required) != 1 || huangliInput.Required[0] != "standard_time" {
		t.Fatalf("huangli required = %#v", huangliInput.Required)
	}
	if _, ok := huangliInput.Properties["standard_time"]; !ok {
		t.Fatalf("huangli civil_time schema missing")
	}
	if count := huangliInput.Properties["count"]; count.Minimum == nil || *count.Minimum != 1 || count.Maximum == nil || *count.Maximum != 30 {
		t.Fatalf("huangli count schema = %#v", count)
	}
	if event := huangliInput.Properties["event"]; len(event.Enum) != 15 {
		t.Fatalf("huangli event enum = %#v", event.Enum)
	}
}

func TestLiuyaoQiguaSchemaUsesMutuallyExclusiveForms(t *testing.T) {
	registry := testEngine()

	var input struct {
		AdditionalProperties *bool               `json:"additionalProperties"`
		Properties           map[string]struct{} `json:"properties"`
		OneOf                []struct {
			Required             []string `json:"required"`
			AdditionalProperties *bool    `json:"additionalProperties"`
			Properties           map[string]struct {
				Const string `json:"const"`
				Enum  []any  `json:"enum"`
			} `json:"properties"`
		} `json:"oneOf"`
	}
	decode(t, mustDefinition(t, registry, "liuyao_qigua").InputSchema, &input)
	if input.AdditionalProperties != nil && *input.AdditionalProperties {
		t.Fatalf("liuyao.qigua composition root must let closed branches own properties: %#v", input.AdditionalProperties)
	}
	if len(input.Properties) != 0 {
		t.Fatalf("liuyao.qigua root must not add union fields: %#v", input.Properties)
	}
	if len(input.OneOf) != 3 {
		t.Fatalf("liuyao.qigua forms = %d, want 3", len(input.OneOf))
	}

	wantRequired := map[string][]string{
		"auto":  {"mode"},
		"coins": {"mode", "rounds"},
		"yaos":  {"mode", "yaos"},
	}
	seen := make(map[string]bool, len(input.OneOf))
	for _, form := range input.OneOf {
		mode := form.Properties["mode"].Const
		if mode == "" || form.Properties["mode"].Enum != nil {
			t.Fatalf("liuyao.qigua mode must be a branch constant: %#v", form.Properties["mode"])
		}
		if seen[mode] {
			t.Fatalf("duplicate liuyao.qigua form: %s", mode)
		}
		seen[mode] = true
		want := wantRequired[mode]
		if len(form.Required) != len(want) {
			t.Fatalf("%s required = %#v, want %#v", mode, form.Required, want)
		}
		for _, field := range want {
			if !slices.Contains(form.Required, field) {
				t.Fatalf("%s required = %#v, want %s", mode, form.Required, field)
			}
		}
		if _, unexpected := form.Properties["rounds"]; mode != "coins" && unexpected {
			t.Fatalf("%s form must not expose rounds", mode)
		}
		if _, unexpected := form.Properties["yaos"]; mode != "yaos" && unexpected {
			t.Fatalf("%s form must not expose yaos", mode)
		}
	}
	for mode := range wantRequired {
		if !seen[mode] {
			t.Fatalf("liuyao.qigua is missing the %s form", mode)
		}
	}
}

func TestQimenChartSchemaUsesMethodForms(t *testing.T) {
	registry := testEngine()

	type leaf struct {
		Const   string   `json:"const"`
		Enum    []string `json:"enum"`
		Minimum *float64 `json:"minimum"`
		Maximum *float64 `json:"maximum"`
	}
	type methodForm struct {
		Required             []string        `json:"required"`
		AdditionalProperties *bool           `json:"additionalProperties"`
		Properties           map[string]leaf `json:"properties"`
	}
	type outerProperty struct {
		leaf
		Properties map[string]leaf `json:"properties"`
		OneOf      []methodForm    `json:"oneOf"`
	}
	type rootForm struct {
		Required             []string                 `json:"required"`
		AdditionalProperties *bool                    `json:"additionalProperties"`
		Properties           map[string]outerProperty `json:"properties"`
	}
	type root struct {
		AdditionalProperties *bool      `json:"additionalProperties"`
		OneOf                []rootForm `json:"oneOf"`
	}
	var input root
	decode(t, mustDefinition(t, registry, "qimen_chart").InputSchema, &input)
	if input.AdditionalProperties != nil && *input.AdditionalProperties {
		t.Fatalf("qimen.chart composition root must let closed branches own properties: %#v", input.AdditionalProperties)
	}
	if len(input.OneOf) != 2 {
		t.Fatalf("qimen.chart root forms = %d, want standard and jinhan", len(input.OneOf))
	}
	standard := input.OneOf[0]
	jinhan := input.OneOf[1]
	if _, exists := standard.Properties["yong_shen"]; !exists {
		t.Fatal("standard qimen form lacks yong_shen")
	}
	if _, exists := standard.Properties["birth_standard_time"]; !exists {
		t.Fatal("standard qimen form lacks birth_standard_time")
	}
	for _, field := range []string{"yong_shen", "birth_standard_time", "birth_timezone"} {
		if _, exists := jinhan.Properties[field]; exists {
			t.Fatalf("jinhan form must not expose %s", field)
		}
	}
	for _, branch := range []rootForm{standard, jinhan} {
		longitude := branch.Properties["longitude"]
		if longitude.Minimum == nil || *longitude.Minimum != -180 ||
			longitude.Maximum == nil || *longitude.Maximum != 180 {
			t.Fatalf("qimen longitude bounds = %#v", longitude)
		}
	}
	if len(standard.Properties["method"].OneOf) != 9 {
		t.Fatalf("qimen method forms = %d, want 9", len(standard.Properties["method"].OneOf))
	}
	scopes := map[string]int{}
	for _, method := range standard.Properties["method"].OneOf {
		scope := method.Properties["scope"].Const
		if scope == "" {
			t.Fatalf("qimen method form lacks a scope const: %#v", method.Properties)
		}
		scopes[scope]++
		if method.AdditionalProperties == nil || !*method.AdditionalProperties {
			// JSON Schema false is decoded into the typed pointer as false.
		} else {
			t.Fatalf("qimen %s method form is not closed", scope)
		}
		if scope != "quarter" {
			if _, exists := method.Properties["quarter_rule"]; exists {
				t.Fatalf("qimen %s form must not expose quarter_rule", scope)
			}
		}
		if scope != "quarter" || method.Properties["quarter_rule"].Const != "twelve_minute_ten_division" {
			for _, field := range []string{"base_dingju_method", "dun_source", "hour_boundary"} {
				if _, exists := method.Properties[field]; exists {
					t.Fatalf("qimen %s form must not expose %s", scope, field)
				}
			}
		}
	}
	wantScopes := map[string]int{"hour": 3, "day": 2, "quarter": 2, "month": 1, "year": 1}
	if !reflect.DeepEqual(scopes, wantScopes) {
		t.Fatalf("qimen method scope coverage = %#v, want %#v", scopes, wantScopes)
	}
	if jinhan.Properties["method"].Properties["school"].Const != "jinhan_yujing" ||
		jinhan.Properties["method"].Properties["scope"].Const != "day" {
		t.Fatalf("jinhan method = %#v", jinhan.Properties["method"].Properties)
	}
}

func TestEngineContractSchemasSelfDescribe(t *testing.T) {
	registry := testEngine()
	for _, name := range registry.Names() {
		definition := mustDefinition(t, registry, name)
		for _, raw := range []json.RawMessage{definition.InputSchema, definition.OutputSchema} {
			if len(raw) == 0 {
				continue
			}
			var schema map[string]any
			decode(t, raw, &schema)
			assertSchemaPropertiesSelfDescribe(t, name, schema)
		}
	}
}

func assertSchemaPropertiesSelfDescribe(t *testing.T, name string, schema map[string]any) {
	t.Helper()
	properties, _ := schema["properties"].(map[string]any)
	if properties == nil {
		return
	}
	if schema["additionalProperties"] != false {
		t.Errorf("%s object schema does not close properties", name)
	}
	for field, raw := range properties {
		property, ok := raw.(map[string]any)
		if !ok {
			t.Fatalf("%s.%s: property is not an object", name, field)
		}
		if field == "schema_version" || field == "contract_version" {
			t.Errorf("%s.%s: wire contracts must not publish a version field", name, field)
		}
		if property["description"] == "" {
			t.Errorf("%s.%s: missing description", name, field)
		}
		assertSchemaPropertiesSelfDescribe(t, name+"."+field, property)
		if items, ok := property["items"].(map[string]any); ok {
			assertSchemaPropertiesSelfDescribe(t, name+"."+field+"[]", items)
		}
	}
	definitions, hasDefinitions := schema["$defs"].(map[string]any)
	if !hasDefinitions {
		return
	}
	for _, raw := range definitions {
		definition, ok := raw.(map[string]any)
		if !ok {
			t.Fatalf("%s $defs contains a non-object schema", name)
		}
		assertSchemaPropertiesSelfDescribe(t, name+".$defs", definition)
	}
}

func mustDefinition(t *testing.T, registry *Registry, name string) Definition {
	t.Helper()
	definition, err := registry.Definition(name)
	if err != nil {
		t.Fatalf("%s: %v", name, err)
	}
	return definition
}

func TestEngineContractBaziTiaoHouUsesExplicitNoSecondary(t *testing.T) {
	registry := testEngine()

	// 1990-05-20 11:49 CST at 116.4E (the counsel baseline birth) has a 调候
	// primary element but no 辅 element. Missing candidates use an explicit
	// closure value; neither JSON null nor an empty-string sentinel is allowed.
	chart := callToolDocument(t, registry, "bazi_chart", map[string]any{
		"standard_time": "1990-05-20T12:49:00+08:00", "longitude": 116.4, "gender": "male",
	})
	full := callToolDocument(t, registry, "bazi_fullchart", map[string]any{"chart": chart})
	tiaoHou, _ := full["tiao_hou"].(map[string]any)
	if tiaoHou == nil {
		t.Fatalf("bazi.fullchart output has no tiao_hou: %#v", full)
	}
	secondary, ok := tiaoHou["secondary_wuxing"]
	if !ok {
		t.Fatalf("bazi.fullchart output omitted secondary_wuxing: %#v", tiaoHou)
	}
	if secondary != "无辅调候" {
		t.Fatalf("expected explicit 无辅调候 secondary_wuxing, got %#v", secondary)
	}

	var schema map[string]any
	decode(t, mustDefinition(t, registry, "bazi_fullchart").OutputSchema, &schema)
	properties := schema["properties"].(map[string]any)
	fuYi := properties["fu_yi"].(map[string]any)["properties"].(map[string]any)
	tiaoHouSchema := properties["tiao_hou"].(map[string]any)["properties"].(map[string]any)
	for name, field := range map[string]map[string]any{
		"fu_yi.yong":                fuYi["yong"].(map[string]any),
		"fu_yi.xi":                  fuYi["xi"].(map[string]any),
		"fu_yi.ji":                  fuYi["ji"].(map[string]any),
		"tiao_hou.secondary_wuxing": tiaoHouSchema["secondary_wuxing"].(map[string]any),
	} {
		if field["type"] != "string" {
			t.Errorf("%s schema is not a string: %#v", name, field)
		}
		values, ok := field["enum"].([]any)
		if !ok {
			t.Fatalf("%s schema has no enum: %#v", name, field)
		}
		want := []any{"土", "木", "水", "火", "金", "无候选"}
		if name == "tiao_hou.secondary_wuxing" {
			want = []any{"土", "木", "水", "火", "金", "无辅调候"}
		}
		if !slices.Equal(values, want) {
			t.Errorf("%s enum = %#v, want %v", name, values, want)
		}
		if slices.Contains(values, nil) {
			t.Errorf("%s enum still allows null", name)
		}
	}

}

func TestEngineContractCalls(t *testing.T) {
	registry := testEngine()

	cityRaw, err := registry.Call(context.Background(), "city_coords", mustJSON(t, CityCoordsInput{City: "北京"}))
	if err != nil {
		t.Fatalf("city.coords: %v", err)
	}
	if err := registry.ValidateOutput("city_coords", cityRaw); err != nil {
		t.Fatalf("city.coords output contract: %v", err)
	}
	var cityOutput CityCoordsOutput
	if err := json.Unmarshal(cityRaw, &cityOutput); err != nil {
		t.Fatal(err)
	}
	if len(cityOutput.Candidates) == 0 {
		t.Fatalf("city output has no candidates: %#v", cityOutput)
	}
	best := cityOutput.Candidates[0]
	if best.Name == "" || best.Longitude == 0 || best.Latitude == 0 || best.Country == "" || best.Timezone == "" {
		t.Fatalf("incomplete city candidate: %#v", best)
	}

	huangliRaw, err := registry.Call(context.Background(), "huangli_days", mustJSON(t, HuangliDaysInput{
		StandardTime: standardInputFromInstant(t, "2024-06-15T04:00:00Z", "Asia/Shanghai"),
		Count:        jsonschema.Ptr(HuangliCount(1)),
		Event:        "sign",
	}))
	if err != nil {
		t.Fatalf("huangli.days: %v", err)
	}
	if err := registry.ValidateOutput("huangli_days", huangliRaw); err != nil {
		t.Fatalf("huangli.days output contract: %v", err)
	}
	var huangliOutput HuangliDaysOutput
	if err := json.Unmarshal(huangliRaw, &huangliOutput); err != nil {
		t.Fatal(err)
	}
	if len(huangliOutput.Items) != 1 {
		t.Fatalf("huangli days = %d, want 1", len(huangliOutput.Items))
	}
	day := huangliOutput.Items[0]
	if day.Event != "sign" || day.EventLabel != "签约" || day.Suitability != "recommended" {
		t.Fatalf("huangli classification = %#v", day)
	}
	if len(day.ShiChen) == 0 || day.Mansion.Name == "" || day.RiZhu.Gan == "" {
		t.Fatalf("huangli wire DTO lost nested facts: %#v", day)
	}

	bazhaiRaw, err := registry.Call(context.Background(), "bazhai_chart", mustJSON(t, BazhaiChartInput{
		StandardTime: standardInputFromInstant(t, "1984-06-15T04:00:00Z", "Asia/Shanghai"),
		Gender:       "male",
	}))
	if err != nil {
		t.Fatalf("bazhai.chart: %v", err)
	}
	if err := registry.ValidateOutput("bazhai_chart", bazhaiRaw); err != nil {
		t.Fatalf("bazhai.chart output contract: %v", err)
	}
	var bazhaiOutput BazhaiChartOutput
	if err := json.Unmarshal(bazhaiRaw, &bazhaiOutput); err != nil {
		t.Fatal(err)
	}
	if bazhaiOutput.MingGua.Gua.Name == "" || bazhaiOutput.MingGua.Group == "" ||
		string(bazhaiOutput.BaZhaiDirs.ShengQi) == "" || len(bazhaiOutput.YearStars.Palaces) != 9 {
		t.Fatalf("incomplete bazhai chart: %#v", bazhaiOutput)
	}

	layoutRaw, err := registry.Call(context.Background(), "bazhai_layout", mustJSON(t, BazhaiLayoutInput{
		MingGua:   "乾",
		DoorGua:   "坤",
		MasterGua: "艮",
		StoveGua:  "兑",
	}))
	if err != nil {
		t.Fatalf("bazhai.layout: %v", err)
	}
	if err := registry.ValidateOutput("bazhai_layout", layoutRaw); err != nil {
		t.Fatalf("bazhai.layout output contract: %v", err)
	}
	var layoutOutput BazhaiLayoutOutput
	if err := json.Unmarshal(layoutRaw, &layoutOutput); err != nil {
		t.Fatal(err)
	}
	if layoutOutput.Group != "西四命" || layoutOutput.Door.GuaName != "坤" || layoutOutput.Door.Match != "吉" {
		t.Fatalf("unexpected bazhai layout: %#v", layoutOutput)
	}

	chartRaw, err := registry.Call(context.Background(), "xuankong_chart", mustJSON(t, XuankongChartInput{
		StandardTime: standardInputFromInstant(t, "2026-07-31T04:00:00Z", "Asia/Shanghai"),
		ZuoShan:      2,
		XiangShan:    14,
	}))
	if err != nil {
		t.Fatalf("xuankong.chart: %v", err)
	}
	if err := registry.ValidateOutput("xuankong_chart", chartRaw); err != nil {
		t.Fatalf("xuankong.chart output contract: %v", err)
	}
	var chart XuankongChart
	if err := json.Unmarshal(chartRaw, &chart); err != nil {
		t.Fatal(err)
	}
	if chart.ChartDigest == "" || len(chart.GongWei) != 9 || chart.ZuoShanName == "" || chart.XiangShanName == "" {
		t.Fatalf("incomplete xuankong chart: %#v", chart)
	}

	liunianRaw, err := registry.Call(context.Background(), "xuankong_liunian", mustJSON(t, XuankongLiunianInput{
		Chart:        &chart,
		StandardTime: standardInputFromInstant(t, "2026-06-15T04:00:00Z", "Asia/Shanghai"),
	}))
	if err != nil {
		t.Fatalf("xuankong.liunian: %v", err)
	}
	if err := registry.ValidateOutput("xuankong_liunian", liunianRaw); err != nil {
		t.Fatalf("xuankong.liunian output contract: %v", err)
	}
	var liunianDocument struct {
		Items []XuankongLiunianOutput `json:"items"`
	}
	if err := json.Unmarshal(liunianRaw, &liunianDocument); err != nil {
		t.Fatal(err)
	}
	if len(liunianDocument.Items) != 1 ||
		len(liunianDocument.Items[0].GongWei) != 9 ||
		len(liunianDocument.Items[0].HouseOverlay) != 9 {
		t.Fatalf("incomplete xuankong overlay: %#v", liunianDocument)
	}

	tampered := chart
	tampered.ZuoShanName = "壬"
	if _, err := registry.Call(context.Background(), "xuankong_liunian", mustJSON(t, XuankongLiunianInput{
		Chart:        &tampered,
		StandardTime: standardInputFromInstant(t, "2026-06-15T04:00:00Z", "Asia/Shanghai"),
	})); err == nil || !strings.Contains(err.Error(), "digest mismatch") {
		t.Fatalf("expected digest mismatch, got %v", err)
	}

	qiguaRaw, err := registry.Call(context.Background(), "liuyao_qigua", mustJSON(t, LiuyaoQiguaInput{
		Mode: "yaos",
		Yaos: LiuyaoYaoValuesInput{7, 8, 7, 8, 7, 8},
	}))
	if err != nil {
		t.Fatalf("liuyao.qigua: %v", err)
	}
	if err := registry.ValidateOutput("liuyao_qigua", qiguaRaw); err != nil {
		t.Fatalf("liuyao.qigua output contract: %v", err)
	}
	var qigua LiuyaoQiguaOutput
	if err := json.Unmarshal(qiguaRaw, &qigua); err != nil {
		t.Fatal(err)
	}
	if qigua.Casting.Mode != "yaos" || len(qigua.Casting.Yaos) != 6 ||
		len(qigua.Casting.DongYao) != 0 || qigua.Casting.CastingID == "" {
		t.Fatalf("incomplete liuyao casting: %#v", qigua.Casting)
	}
	if qigua.Casting.CoinConvention != nil || qigua.Casting.Rounds != nil {
		t.Fatalf("yaos casting should omit coin-only audit fields: %#v", qigua.Casting)
	}

	if _, err := registry.Call(context.Background(), "liuyao_qigua", mustJSON(t, LiuyaoQiguaInput{
		Yaos: LiuyaoYaoValuesInput{7, 7, 7, 7, 7, 7},
	})); err == nil || !strings.Contains(err.Error(), "auto mode does not accept") {
		t.Fatalf("expected auto-mode extra input rejection, got %v", err)
	}

	liuyaoChartRaw, err := registry.Call(context.Background(), "liuyao_chart", mustJSON(t, LiuyaoChartInput{
		StandardTime: standardInputFromInstant(t, "2026-07-31T02:00:00Z", "Asia/Shanghai"),
		Longitude:    116.4,
		Casting:      &qigua.Casting,
	}))
	if err != nil {
		t.Fatalf("liuyao.chart: %v", err)
	}
	if err := registry.ValidateOutput("liuyao_chart", liuyaoChartRaw); err != nil {
		t.Fatalf("liuyao.chart output contract: %v", err)
	}
	var liuyaoChart LiuyaoChartOutput
	if err := json.Unmarshal(liuyaoChartRaw, &liuyaoChart); err != nil {
		t.Fatal(err)
	}
	if liuyaoChart.Name == "" || len(liuyaoChart.Lines) != 6 || len(liuyaoChart.BianLines) != 6 ||
		liuyaoChart.Casting.CastingID == "" || liuyaoChart.TombSchool.School == "" {
		t.Fatalf("incomplete liuyao chart: %#v", liuyaoChart)
	}

	tamperedQigua := qigua
	tamperedQigua.Casting.Yaos[0] = 8
	if _, err := registry.Call(context.Background(), "liuyao_chart", mustJSON(t, LiuyaoChartInput{
		StandardTime: standardInputFromInstant(t, "2026-07-31T02:00:00Z", "Asia/Shanghai"),
		Longitude:    116.4,
		Casting:      &tamperedQigua.Casting,
	})); err == nil || (!strings.Contains(err.Error(), "casting_id mismatch") &&
		!strings.Contains(err.Error(), "dong_yao mismatch")) {
		t.Fatalf("expected tampered casting rejection, got %v", err)
	}
}

func mustJSON(t *testing.T, value any) json.RawMessage {
	t.Helper()
	raw, err := json.Marshal(value)
	if err != nil {
		t.Fatal(err)
	}
	return raw
}
