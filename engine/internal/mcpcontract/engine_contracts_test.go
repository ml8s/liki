package mcpcontract

import (
	"context"
	"encoding/json"
	"os"
	"strings"
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

func TestEngineContractDefinitions(t *testing.T) {
	registry := NewEngine()
	want := []string{
		"time.now", "huangli.days", "tianwen.time", "city.coords",
		"bazi.chart",
		"bazhai.chart", "bazhai.layout",
		"xuankong.chart", "xuankong.liunian",
		"liuyao.qigua",
		"liuyao.chart",
		"bazi.fullchart", "bazi.bond",
		"bazi.liunian", "bazi.liuyue", "bazi.liuri", "bazi.liushi", "bazi.xiaoyun",
		"ziwei.chart", "ziwei.fullchart", "ziwei.daxian", "ziwei.liunian",
		"ziwei.liuyue", "ziwei.liuri", "ziwei.liushi", "ziwei.bond",
		"qimen.chart",
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
		definition, ok := registry.Definition(name)
		if !ok {
			t.Fatalf("%s is not registered", name)
		}
		if definition.Description == "" {
			t.Errorf("%s has no description", name)
		}
		if len(definition.InputSchema) == 0 {
			t.Errorf("%s has no input schema", name)
		}
	}

	// Object roots expose output schemas; array roots (huangli.days,
	// bazi.xiaoyun, ziwei.daxian) stay runtime-validated but are intentionally
	// undeclared, matching the MCP object-only outputSchema rule.
	arrayRoots := map[string]bool{
		"huangli.days": true, "bazi.xiaoyun": true, "ziwei.daxian": true,
	}
	for _, name := range want {
		definition, _ := registry.Definition(name)
		if arrayRoots[name] {
			if len(definition.OutputSchema) != 0 {
				t.Errorf("%s should not declare a non-object output schema", name)
			}
			continue
		}
		if len(definition.OutputSchema) == 0 {
			t.Errorf("%s should declare an object output schema", name)
		}
	}
}

func TestEngineContractInputConstraints(t *testing.T) {
	registry := NewEngine()

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
	decode(t, mustDefinition(t, registry, "city.coords").InputSchema, &cityInput)
	if cityInput.Type != "object" || cityInput.AdditionalProperties == nil || *cityInput.AdditionalProperties {
		t.Fatalf("city input is not a closed object: %#v", cityInput)
	}
	if len(cityInput.Required) != 1 || cityInput.Required[0] != "city" {
		t.Fatalf("city required = %#v", cityInput.Required)
	}
	if city, ok := cityInput.Properties["city"]; !ok || city.Type != "string" || city.MinLength != 1 || city.MaxLength != 128 {
		t.Fatalf("city schema = %#v", city)
	}

	var tianwenInput struct {
		Properties map[string]struct {
			Type   string `json:"type"`
			Format string `json:"format"`
		} `json:"properties"`
	}
	decode(t, mustDefinition(t, registry, "tianwen.time").InputSchema, &tianwenInput)
	if timeSchema, ok := tianwenInput.Properties["time"]; !ok || timeSchema.Type != "string" || timeSchema.Format != "date-time" {
		t.Fatalf("tianwen time schema = %#v", timeSchema)
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
	decode(t, mustDefinition(t, registry, "huangli.days").InputSchema, &huangliInput)
	if len(huangliInput.Required) != 1 || huangliInput.Required[0] != "start_date" {
		t.Fatalf("huangli required = %#v", huangliInput.Required)
	}
	if start := huangliInput.Properties["start_date"]; start.Format != "date" {
		t.Fatalf("huangli start_date schema = %#v", start)
	}
	if count := huangliInput.Properties["count"]; count.Minimum == nil || *count.Minimum != 1 || count.Maximum == nil || *count.Maximum != 30 {
		t.Fatalf("huangli count schema = %#v", count)
	}
	if event := huangliInput.Properties["event"]; len(event.Enum) != 15 {
		t.Fatalf("huangli event enum = %#v", event.Enum)
	}
}

func TestEngineContractSchemasSelfDescribe(t *testing.T) {
	registry := NewEngine()
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
	definition, ok := registry.Definition(name)
	if !ok {
		t.Fatalf("%s is not registered", name)
	}
	return definition
}

func TestEngineContractBaziTiaoHouAllowsEmptySecondaryWuxing(t *testing.T) {
	registry := NewEngine()

	// 1990-05-20 11:49 CST at 116.4E (the counsel baseline birth) has a 调候
	// primary element but no 辅 element, so the engine emits
	// tiao_hou.secondary_wuxing = "" - the domain allows exactly that empty
	// value (engine/internal/engine/bazi/bazi_correctness_test.go). Output
	// validation must accept the Go zero value, otherwise real charts fail with
	// an output contract violation.
	timeDoc := callToolDocument(t, registry, "tianwen.time", map[string]any{
		"time": "1990-05-20T11:49:00+08:00", "longitude": 116.4,
	})
	solar, _ := timeDoc["solar"].(string)
	if solar == "" {
		t.Fatalf("tianwen.time output = %#v", timeDoc)
	}
	chart := callToolDocument(t, registry, "bazi.chart", map[string]any{
		"solar_time": solar, "gender": "male",
	})
	full := callToolDocument(t, registry, "bazi.fullchart", map[string]any{"chart": chart})
	tiaoHou, _ := full["tiao_hou"].(map[string]any)
	if tiaoHou == nil {
		t.Fatalf("bazi.fullchart output has no tiao_hou: %#v", full)
	}
	if secondary, _ := tiaoHou["secondary_wuxing"].(string); secondary != "" {
		t.Fatalf("expected an empty secondary_wuxing for the baseline birth, got %q", secondary)
	}
}

func TestEngineContractCalls(t *testing.T) {
	registry := NewEngine()

	cityRaw, err := registry.Call(context.Background(), "city.coords", mustJSON(t, CityCoordsInput{City: "北京"}))
	if err != nil {
		t.Fatalf("city.coords: %v", err)
	}
	var cityOutput CityCoordsOutput
	if err := json.Unmarshal(cityRaw, &cityOutput); err != nil {
		t.Fatal(err)
	}
	if cityOutput.Name == "" || cityOutput.Longitude == 0 || cityOutput.Latitude == 0 || cityOutput.Country == "" {
		t.Fatalf("incomplete city output: %#v", cityOutput)
	}

	huangliRaw, err := registry.Call(context.Background(), "huangli.days", mustJSON(t, HuangliDaysInput{
		StartDate: "2024-06-15",
		Count:     jsonschema.Ptr(HuangliCount(1)),
		Event:     "sign",
	}))
	if err != nil {
		t.Fatalf("huangli.days: %v", err)
	}
	var huangliOutput []HuangliDayOutput
	if err := json.Unmarshal(huangliRaw, &huangliOutput); err != nil {
		t.Fatal(err)
	}
	if len(huangliOutput) != 1 {
		t.Fatalf("huangli days = %d, want 1", len(huangliOutput))
	}
	day := huangliOutput[0]
	if day.Event != "sign" || day.EventLabel != "签约" || day.Suitability != "recommended" {
		t.Fatalf("huangli classification = %#v", day)
	}
	if len(day.ShiChen) == 0 || day.Mansion.Name == "" || day.RiZhu.Gan == "" {
		t.Fatalf("huangli wire DTO lost nested facts: %#v", day)
	}

	bazhaiRaw, err := registry.Call(context.Background(), "bazhai.chart", mustJSON(t, BazhaiChartInput{
		BirthYear: 1984,
		Gender:    "male",
	}))
	if err != nil {
		t.Fatalf("bazhai.chart: %v", err)
	}
	var bazhaiOutput BazhaiChartOutput
	if err := json.Unmarshal(bazhaiRaw, &bazhaiOutput); err != nil {
		t.Fatal(err)
	}
	if bazhaiOutput.MingGua.Gua.Name == "" || bazhaiOutput.MingGua.Group == "" ||
		len(bazhaiOutput.BaZhaiDirs.ShengQi) != 1 || len(bazhaiOutput.YearStars.Palaces) != 9 {
		t.Fatalf("incomplete bazhai chart: %#v", bazhaiOutput)
	}

	layoutRaw, err := registry.Call(context.Background(), "bazhai.layout", mustJSON(t, BazhaiLayoutInput{
		MingGua:   "乾",
		DoorGua:   "坤",
		MasterGua: "艮",
		StoveGua:  "兑",
	}))
	if err != nil {
		t.Fatalf("bazhai.layout: %v", err)
	}
	var layoutOutput BazhaiLayoutOutput
	if err := json.Unmarshal(layoutRaw, &layoutOutput); err != nil {
		t.Fatal(err)
	}
	if layoutOutput.Group != "西四命" || layoutOutput.Door.GuaName != "坤" || layoutOutput.Door.Match != "吉" {
		t.Fatalf("unexpected bazhai layout: %#v", layoutOutput)
	}

	chartRaw, err := registry.Call(context.Background(), "xuankong.chart", mustJSON(t, XuankongChartInput{
		PeriodDate: "2026-07-31",
		ZuoShan:    2,
		XiangShan:  14,
	}))
	if err != nil {
		t.Fatalf("xuankong.chart: %v", err)
	}
	var chart XuankongChart
	if err := json.Unmarshal(chartRaw, &chart); err != nil {
		t.Fatal(err)
	}
	if chart.ChartDigest == "" || len(chart.GongWei) != 9 || chart.ZuoShanName == "" || chart.XiangShanName == "" {
		t.Fatalf("incomplete xuankong chart: %#v", chart)
	}

	liunianRaw, err := registry.Call(context.Background(), "xuankong.liunian", mustJSON(t, XuankongLiunianInput{
		Chart: &chart,
		Year:  2026,
	}))
	if err != nil {
		t.Fatalf("xuankong.liunian: %v", err)
	}
	var liunian XuankongLiunianOutput
	if err := json.Unmarshal(liunianRaw, &liunian); err != nil {
		t.Fatal(err)
	}
	if len(liunian.GongWei) != 9 || len(liunian.HouseOverlay) != 9 {
		t.Fatalf("incomplete xuankong overlay: %#v", liunian)
	}

	tampered := chart
	tampered.ZuoShanName = "壬"
	if _, err := registry.Call(context.Background(), "xuankong.liunian", mustJSON(t, XuankongLiunianInput{
		Chart: &tampered,
		Year:  2026,
	})); err == nil || !strings.Contains(err.Error(), "digest mismatch") {
		t.Fatalf("expected digest mismatch, got %v", err)
	}

	qiguaRaw, err := registry.Call(context.Background(), "liuyao.qigua", mustJSON(t, LiuyaoQiguaInput{
		Mode: "yaos",
		Yaos: LiuyaoYaoValuesInput{7, 8, 7, 8, 7, 8},
	}))
	if err != nil {
		t.Fatalf("liuyao.qigua: %v", err)
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

	if _, err := registry.Call(context.Background(), "liuyao.qigua", mustJSON(t, LiuyaoQiguaInput{
		Yaos: LiuyaoYaoValuesInput{7, 7, 7, 7, 7, 7},
	})); err == nil || !strings.Contains(err.Error(), "auto mode does not accept") {
		t.Fatalf("expected auto-mode extra input rejection, got %v", err)
	}

	liuyaoChartRaw, err := registry.Call(context.Background(), "liuyao.chart", mustJSON(t, LiuyaoChartInput{
		SolarTime: "2026-07-31T10:00:00+08:00",
		Casting:   &qigua.Casting,
	}))
	if err != nil {
		t.Fatalf("liuyao.chart: %v", err)
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
	if _, err := registry.Call(context.Background(), "liuyao.chart", mustJSON(t, LiuyaoChartInput{
		SolarTime: "2026-07-31T10:00:00+08:00",
		Casting:   &tamperedQigua.Casting,
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
