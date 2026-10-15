package agent

import (
	"context"
	"encoding/json"
	"fmt"
	"strings"
	"testing"

	"liki-engine/internal/engine/ganzhi"
	"liki-engine/internal/engine/qimen"
)

const btOK = `"2000-06-15T12:00:00+08:00"`
const btOK2 = `"2000-09-20T08:30:00+08:00"`
const lunarOK = `{"year":2000,"month":6,"day":15,"shichen":"午"}`
const lunarOK2 = `{"year":2000,"month":8,"day":23,"shichen":"辰"}`

// ── helpers ──

func getBaziChart(t *testing.T, r *RPCRegistry, solarTime, gender string) json.RawMessage {
	t.Helper()
	params := json.RawMessage(fmt.Sprintf(`{"solar_time":%s,"gender":%q}`, solarTime, gender))
	result, err := r.Execute(context.Background(), "bazi.chart", params)
	if err != nil {
		t.Fatalf("bazi.chart: %v", err)
	}
	var env struct {
		Data json.RawMessage `json:"data"`
	}
	if err := json.Unmarshal(result, &env); err != nil {
		t.Fatal(err)
	}
	return env.Data
}

func hasKey(raw json.RawMessage, key string) bool {
	var m map[string]any
	if err := json.Unmarshal(raw, &m); err != nil {
		return false
	}
	_, ok := m[key]
	return ok
}

func hasPath(raw json.RawMessage, path string) bool {
	var m map[string]any
	if err := json.Unmarshal(raw, &m); err != nil {
		return false
	}
	parts := strings.Split(path, ".")
	for i, part := range parts {
		v, ok := m[part]
		if !ok {
			return false
		}
		if i == len(parts)-1 {
			return true
		}
		if vm, ok := v.(map[string]any); ok {
			m = vm
		} else {
			return false
		}
	}
	return false
}

func getStr(raw json.RawMessage, path ...string) string {
	var m map[string]any
	if err := json.Unmarshal(raw, &m); err != nil {
		return ""
	}
	cur := any(m)
	for _, k := range path {
		cm, ok := cur.(map[string]any)
		if !ok {
			return ""
		}
		cur = cm[k]
	}
	s, _ := cur.(string)
	return s
}

func assertError(t *testing.T, r *RPCRegistry, method string, params json.RawMessage) {
	t.Helper()
	_, err := r.Execute(context.Background(), method, params)
	if err == nil {
		t.Errorf("%s: expected error", method)
	} else if strings.HasPrefix(err.Error(), "method not found") {
		t.Errorf("%s: method not found (spelling mistake?): %v", method, err)
	}
}

// ── gender validation ──

func TestValidateGender(t *testing.T) {
	tests := []struct {
		name    string
		gender  string
		wantErr bool
	}{
		{"male valid", "male", false},
		{"female valid", "female", false},
		{"空字符串", "", true},
		{"无效值 x", "x", true},
		{"无效值 other", "other", true},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := validateGender(ganzhi.Gender(tt.gender))
			if (err != nil) != tt.wantErr {
				t.Errorf("validateGender(%q) err=%v, wantErr=%v", tt.gender, err, tt.wantErr)
			}
		})
	}
}

// ── handler error paths ──

func TestHandler_InvalidJSON(t *testing.T) {
	r := NewRPCRegistry()
	handlers := []string{
		"bazi.chart",
		"bazi.bond", "bazi.liunian", "bazi.liuyue",
		"bazi.liuri", "bazi.liushi", "bazi.xiaoyun",
		"ziwei.chart", "ziwei.daxian", "ziwei.liunian", "ziwei.liuyue",
		"ziwei.liuri", "ziwei.bond",
		"qimen.chart",
		"liuyao.qigua", "liuyao.chart",
	}
	badJSON := json.RawMessage(`{bad`)
	for _, name := range handlers {
		t.Run(name, func(t *testing.T) {
			assertError(t, r, name, badJSON)
		})
	}
}

func TestHandler_MissingGender(t *testing.T) {
	r := NewRPCRegistry()
	handlers := []string{
		"bazi.chart", "ziwei.chart",
	}
	noGender := json.RawMessage(fmt.Sprintf(`{"solar_time":%s}`, btOK))
	for _, name := range handlers {
		t.Run(name, func(t *testing.T) {
			assertError(t, r, name, noGender)
		})
	}
}

func TestHandler_BadGender(t *testing.T) {
	r := NewRPCRegistry()
	handlers := []struct {
		name   string
		params string
	}{
		{"bazi.chart", fmt.Sprintf(`{"solar_time":%s,"gender":"other"}`, btOK)},
		{"ziwei.chart", fmt.Sprintf(`{"lunar":%s,"gender":"bad"}`, lunarOK)},
	}
	for _, tt := range handlers {
		t.Run(tt.name, func(t *testing.T) {
			assertError(t, r, tt.name, json.RawMessage(tt.params))
		})
	}
}

func TestHandler_MissingRequiredFields(t *testing.T) {
	r := NewRPCRegistry()
	tests := []struct {
		name   string
		params string
	}{
		{"ziwei.daxian", `{}`},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			assertError(t, r, tt.name, json.RawMessage(tt.params))
		})
	}
	// chart-dependent methods — need chart from bazi.chart
	t.Run("bazi.bond", func(t *testing.T) {
		chart := getBaziChart(t, r, btOK, "male")
		params := json.RawMessage(fmt.Sprintf(`{"a":{"chart":%s}}`, chart))
		assertError(t, r, "bazi.bond", params)
	})
	t.Run("bazi.liunian", func(t *testing.T) {
		assertError(t, r, "bazi.liunian", json.RawMessage(`{"year":2026}`))
	})
	t.Run("bazi.liuyue", func(t *testing.T) {
		assertError(t, r, "bazi.liuyue", json.RawMessage(`{"year":2026,"month":6}`))
	})
}

func TestHandler_RangeValidation(t *testing.T) {
	r := NewRPCRegistry()
	tests := []struct {
		name   string
		params string
	}{
		{"qimen.chart", fmt.Sprintf(`{"solar_time":%s,"yong_shen":["甲"]}`, btOK)},
		{"qimen.chart", fmt.Sprintf(`{"solar_time":%s,"scope":"ke"}`, btOK)},
		{"qimen.chart", fmt.Sprintf(`{"solar_time":%s,"birth_year":1984}`, btOK)},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			assertError(t, r, tt.name, json.RawMessage(tt.params))
		})
	}
	// chart-dependent methods
	t.Run("bazi.liunian", func(t *testing.T) {
		chart := getBaziChart(t, r, btOK, "male")
		params := json.RawMessage(fmt.Sprintf(`{"year":0,"chart":%s}`, chart))
		assertError(t, r, "bazi.liunian", params)
	})
	t.Run("bazi.liushi", func(t *testing.T) {
		chart := getBaziChart(t, r, btOK, "male")
		for _, hour := range []int{-1, 24} {
			params := json.RawMessage(fmt.Sprintf(`{"year":2026,"month":6,"day":15,"hour":%d,"chart":%s}`, hour, chart))
			assertError(t, r, "bazi.liushi", params)
		}
	})
}

func TestHandler_QimenDefault(t *testing.T) {
	r := NewRPCRegistry()
	params := json.RawMessage(fmt.Sprintf(`{"solar_time":%s}`, btOK))
	result := executeAndDecode(t, r, "qimen.chart", params)
	if result["_product"] != "qimen" || result["data"] == nil {
		t.Fatalf("default envelope = %+v", result)
	}
	method := result["data"].(map[string]any)["method"].(map[string]any)
	if method["scope"] != "hour" || method["school"] != "zhuanpan" || method["dingju_method"] != "chaibu" {
		t.Fatalf("default method = %v/%v/%v", method["scope"], method["school"], method["dingju_method"])
	}
}

func TestHandler_QimenDingjuMethod(t *testing.T) {
	r := NewRPCRegistry()
	params := json.RawMessage(`{"solar_time":"2026-01-05T13:00:00+08:00","dingju_method":"zhirun"}`)
	result := executeAndDecode(t, r, "qimen.chart", params)
	data := result["data"].(map[string]any)
	method := data["method"].(map[string]any)
	if method["dingju_method"] != "zhirun" || method["dingju_method_name"] != "置闰定局" {
		t.Fatalf("dingju method = %v/%v, want zhirun/置闰定局", method["dingju_method"], method["dingju_method_name"])
	}
	if method["jie_qi"] != "冬至" || method["yong_ju_jie_qi"] != "小寒" {
		t.Fatalf("terms = %v/%v, want 冬至/小寒", method["jie_qi"], method["yong_ju_jie_qi"])
	}
	assertError(t, r, "qimen.chart", json.RawMessage(`{"solar_time":"2026-01-05T13:00:00+08:00","dingju_method":"luoshu_feipan"}`))
}

func TestHandler_QimenMethodMatrix(t *testing.T) {
	r := NewRPCRegistry()
	params := json.RawMessage(`{
		"solar_time":"2026-01-14T01:03:00+08:00",
		"scope":"hour","school":"luoshu_feipan","dingju_method":"zhirun"
	}`)
	result := executeAndDecode(t, r, "qimen.chart", params)
	data := result["data"].(map[string]any)
	method := data["method"].(map[string]any)
	if method["scope"] != "hour" || method["school"] != "luoshu_feipan" || method["dingju_method"] != "zhirun" {
		t.Fatalf("method = %v/%v/%v", method["scope"], method["school"], method["dingju_method"])
	}
	if _, exists := data["pan"].(map[string]any)["school"]; exists {
		t.Fatal("pan school duplicates method.school")
	}
	invalid := []string{
		`{"solar_time":"2026-01-14T01:03:00+08:00","scope":"ke"}`,
		`{"solar_time":"2026-01-14T01:03:00+08:00","school":"yinpan"}`,
		`{"solar_time":"2026-01-14T01:03:00+08:00","scope":"year","dingju_method":"chaibu"}`,
		`{"solar_time":"2026-01-14T01:03:00+08:00","scope":"day","school":"mingfa_feipan"}`,
		`{"solar_time":"2026-01-14T01:03:00+08:00","scope":"hour","school":"mingfa_feipan","dingju_method":"zhirun"}`,
		`{"solar_time":"2026-01-14T01:03:00+08:00","kind":"hour"}`,
		`{"solar_time":"2026-01-14T01:03:00+08:00","plate":"zhuanpan"}`,
		`{"solar_time":"2026-01-14T01:03:00+08:00","ju_method":"chaibu"}`,
	}
	for _, params := range invalid {
		assertError(t, r, "qimen.chart", json.RawMessage(params))
	}
}

func TestHandler_QimenBirthDateBoundary(t *testing.T) {
	if _, err := qimen.ParseBirthDate("2024-02-04"); err == nil {
		t.Error("date-only Lichun boundary should require a time")
	}
	if _, err := qimen.ParseBirthDate("2024-02-04T02:00:00Z"); err != nil {
		t.Fatalf("RFC3339 Lichun boundary should parse: %v", err)
	}
}

func liuyaoCastingJSON(t *testing.T, r *RPCRegistry, values [6]int) json.RawMessage {
	t.Helper()
	valuesJSON, err := json.Marshal(values)
	if err != nil {
		t.Fatal(err)
	}
	result, err := r.Execute(context.Background(), "liuyao.qigua", json.RawMessage(
		[]byte(`{"mode":"yaos","yaos":`+string(valuesJSON)+`}`),
	))
	if err != nil {
		t.Fatalf("liuyao.qigua yaos: %v", err)
	}
	var env struct {
		Data struct {
			Casting json.RawMessage `json:"casting"`
		} `json:"data"`
	}
	if err := json.Unmarshal(result, &env); err != nil {
		t.Fatal(err)
	}
	if len(env.Data.Casting) == 0 {
		t.Fatal("liuyao.qigua returned no casting")
	}
	return env.Data.Casting
}

func TestHandler_LiuyaoYongShenDefault(t *testing.T) {
	r := NewRPCRegistry()
	params := json.RawMessage(fmt.Sprintf(`{"solar_time":%s,"casting":%s}`, btOK, liuyaoCastingJSON(t, r, [6]int{7, 7, 7, 7, 7, 7})))
	result, err := r.Execute(context.Background(), "liuyao.chart", params)
	if err != nil {
		t.Fatalf("liuyao.chart (default yong_shen): %v", err)
	}
	if getStr(result, "_product") != "liuyao" {
		t.Errorf("_product = %q, want liuyao", getStr(result, "_product"))
	}
}

// ── handler valid paths (envelope + data) ──

func TestHandler_ComputeChart_Valid(t *testing.T) {
	r := NewRPCRegistry()
	params := json.RawMessage(fmt.Sprintf(`{"solar_time":%s,"gender":"male"}`, btOK))
	result, err := r.Execute(context.Background(), "bazi.chart", params)
	if err != nil {
		t.Fatalf("bazi.chart: %v", err)
	}
	if getStr(result, "_product") != "chart" {
		t.Errorf("_product = %q, want chart", getStr(result, "_product"))
	}
	if !hasKey(result, "data") {
		t.Error("missing data")
	}
}

func TestHandler_ComputeBond_Valid(t *testing.T) {
	r := NewRPCRegistry()
	chartA := getBaziChart(t, r, btOK, "male")
	chartB := getBaziChart(t, r, btOK2, "female")
	params := json.RawMessage(fmt.Sprintf(`{"a":{"chart":%s},"b":{"chart":%s}}`, chartA, chartB))
	result, err := r.Execute(context.Background(), "bazi.bond", params)
	if err != nil {
		t.Fatalf("bazi.bond: %v", err)
	}
	if getStr(result, "_product") != "bond" {
		t.Errorf("_product = %q, want bond", getStr(result, "_product"))
	}
}

func TestHandler_ComputeLiunian_Valid(t *testing.T) {
	r := NewRPCRegistry()
	chart := getBaziChart(t, r, btOK, "male")
	params := json.RawMessage(fmt.Sprintf(`{"year":2026,"chart":%s}`, chart))
	result, err := r.Execute(context.Background(), "bazi.liunian", params)
	if err != nil {
		t.Fatalf("bazi.liunian: %v", err)
	}
	if getStr(result, "_product") != "liunian" {
		t.Errorf("_product = %q, want liunian", getStr(result, "_product"))
	}
}

func TestHandler_ComputeZiwei_Valid(t *testing.T) {
	r := NewRPCRegistry()
	params := json.RawMessage(fmt.Sprintf(`{"lunar":%s,"gender":"male"}`, lunarOK))
	result, err := r.Execute(context.Background(), "ziwei.chart", params)
	if err != nil {
		t.Fatalf("ziwei.chart: %v", err)
	}
	if getStr(result, "_product") != "ziwei" {
		t.Errorf("_product = %q, want ziwei", getStr(result, "_product"))
	}
}

func TestHandler_ComputeXiaoYun_Valid(t *testing.T) {
	r := NewRPCRegistry()
	chart := getBaziChart(t, r, btOK, "male")
	params := json.RawMessage(fmt.Sprintf(`{"chart":%s,"count":5}`, chart))
	result, err := r.Execute(context.Background(), "bazi.xiaoyun", params)
	if err != nil {
		t.Fatalf("bazi.xiaoyun: %v", err)
	}
	if getStr(result, "_product") != "xiaoyun" {
		t.Errorf("_product = %q, want xiaoyun", getStr(result, "_product"))
	}
	if !hasKey(result, "data") {
		t.Error("missing data")
	}
}

func TestHandler_ComputeXiaoYun_DefaultCount(t *testing.T) {
	r := NewRPCRegistry()
	chart := getBaziChart(t, r, btOK, "male")
	params := json.RawMessage(fmt.Sprintf(`{"chart":%s}`, chart))
	result, err := r.Execute(context.Background(), "bazi.xiaoyun", params)
	if err != nil {
		t.Fatalf("bazi.xiaoyun (default count): %v", err)
	}
	if getStr(result, "_product") != "xiaoyun" {
		t.Errorf("_product = %q, want xiaoyun", getStr(result, "_product"))
	}
}

func TestHandler_ComputeLiushi_Valid(t *testing.T) {
	r := NewRPCRegistry()
	chart := getBaziChart(t, r, btOK, "male")
	params := json.RawMessage(fmt.Sprintf(
		`{"year":2026,"month":6,"day":15,"hour":12,"chart":%s}`, chart))
	result, err := r.Execute(context.Background(), "bazi.liushi", params)
	if err != nil {
		t.Fatalf("bazi.liushi: %v", err)
	}
	if getStr(result, "_product") != "liushi" {
		t.Errorf("_product = %q, want liushi", getStr(result, "_product"))
	}
	if !hasKey(result, "data") {
		t.Error("missing data")
	}
}

func TestHandler_ComputeLiuri_Valid(t *testing.T) {
	r := NewRPCRegistry()
	chart := getBaziChart(t, r, btOK, "male")
	params := json.RawMessage(fmt.Sprintf(
		`{"year":2026,"month":6,"day":15,"chart":%s}`, chart))
	result, err := r.Execute(context.Background(), "bazi.liuri", params)
	if err != nil {
		t.Fatalf("bazi.liuri: %v", err)
	}
	if getStr(result, "_product") != "liuri" {
		t.Errorf("_product = %q, want liuri", getStr(result, "_product"))
	}
	if !hasKey(result, "data") {
		t.Error("missing data")
	}
}

func TestHandler_FullChartIncludesExtra(t *testing.T) {
	r := NewRPCRegistry()
	chart := getBaziChart(t, r, btOK, "male")
	params := json.RawMessage(fmt.Sprintf(`{"chart":%s}`, chart))
	result, err := r.Execute(context.Background(), "bazi.fullchart", params)
	if err != nil {
		t.Fatalf("bazi.fullchart: %v", err)
	}
	if !hasKey(result, "data") {
		t.Error("missing data")
	}
	if !hasPath(result, "data.san_yuan") {
		t.Error("fullchart missing san_yuan (from chart_extra)")
	}
	if !hasPath(result, "data.gan_he") {
		t.Error("fullchart missing gan_he (from hehui)")
	}
	if !hasPath(result, "data.zhi_liu_he") {
		t.Error("fullchart missing zhi_liu_he (from hehui)")
	}
	if !hasPath(result, "data.ten_god_states") {
		t.Error("fullchart missing ten_god_states")
	}
	if !hasPath(result, "data.element_states") {
		t.Error("fullchart missing element_states")
	}
	if !hasPath(result, "data.da_yun") {
		t.Error("fullchart missing da_yun")
	}
	for _, pillar := range []string{"nian", "yue", "ri", "shi"} {
		path := "data." + pillar + ".na_yin"
		if !hasPath(result, path) {
			t.Errorf("fullchart missing %s", path)
		}
	}
	var full struct {
		Data struct {
			NayinRel []map[string]string `json:"nayin_rel"`
		} `json:"data"`
	}
	if err := json.Unmarshal(result, &full); err != nil {
		t.Fatal(err)
	}
	if len(full.Data.NayinRel) != 6 {
		t.Fatalf("nayin_rel len = %d, want 6", len(full.Data.NayinRel))
	}
	hasElementRelation := false
	for _, relation := range full.Data.NayinRel {
		if relation["relation"] != "相同" {
			hasElementRelation = true
			break
		}
	}
	if !hasElementRelation {
		t.Fatalf("nayin_rel are all 相同: %#v", full.Data.NayinRel)
	}
}

func TestHandler_BaziFullChartRejectsMissingOrWrongNaYin(t *testing.T) {
	r := NewRPCRegistry()
	raw := getBaziChart(t, r, btOK, "male")
	var chart map[string]any
	if err := json.Unmarshal(raw, &chart); err != nil {
		t.Fatal(err)
	}
	nian, _ := chart["nian"].(map[string]any)
	original, _ := nian["na_yin"].(string)
	if original == "" {
		t.Fatal("bazi.chart returned empty nian.na_yin")
	}

	nian["na_yin"] = ""
	payload, err := json.Marshal(map[string]any{"chart": chart})
	if err != nil {
		t.Fatal(err)
	}
	_, err = r.Execute(context.Background(), "bazi.fullchart", payload)
	if err == nil || !strings.Contains(err.Error(), "chart.nian.na_yin is required") {
		t.Fatalf("missing na_yin error = %v, want explicit contract failure", err)
	}

	nian["na_yin"] = "错误纳音"
	payload, err = json.Marshal(map[string]any{"chart": chart})
	if err != nil {
		t.Fatal(err)
	}
	_, err = r.Execute(context.Background(), "bazi.fullchart", payload)
	if err == nil || !strings.Contains(err.Error(), original) {
		t.Fatalf("wrong na_yin error = %v, want mismatch with %q", err, original)
	}
}

func TestHandler_ComputeLiuyue_Valid(t *testing.T) {
	r := NewRPCRegistry()
	chart := getBaziChart(t, r, btOK, "male")
	params := json.RawMessage(fmt.Sprintf(`{"year":2026,"month":6,"chart":%s}`, chart))
	result, err := r.Execute(context.Background(), "bazi.liuyue", params)
	if err != nil {
		t.Fatalf("bazi.liuyue: %v", err)
	}
	if getStr(result, "_product") != "liuyue" {
		t.Errorf("_product = %q, want liuyue", getStr(result, "_product"))
	}
	if !hasKey(result, "data") {
		t.Error("missing data")
	}
}

func TestHandler_ComputeLiuNian_ShiShen_Correct(t *testing.T) {
	r := NewRPCRegistry()
	chart := getBaziChart(t, r, btOK, "male")
	for _, year := range []int{2024, 2025, 2026} {
		params := json.RawMessage(fmt.Sprintf(`{"year":%d,"chart":%s}`, year, string(chart)))
		result, err := r.Execute(context.Background(), "bazi.liunian", params)
		if err != nil {
			t.Fatalf("bazi.liunian(%d): %v", year, err)
		}
		shiShen := getStr(result, "data", "shi_shen")
		if shiShen == "" {
			t.Errorf("year=%d: shi_shen empty", year)
		}
	}
}

// ── wrapResult ──

func TestWrapResult(t *testing.T) {
	result, err := wrapResult("test", map[string]any{"key": "value"})
	if err != nil {
		t.Fatal(err)
	}
	if getStr(result, "_product") != "test" {
		t.Errorf("_product = %q, want test", getStr(result, "_product"))
	}
	if !hasKey(result, "data") {
		t.Error("missing data field")
	}
}

// ── liuyao.qigua ──

func TestHandler_LiuyaoQigua(t *testing.T) {
	r := NewRPCRegistry()
	result, err := r.Execute(context.Background(), "liuyao.qigua", json.RawMessage(`{}`))
	if err != nil {
		t.Fatalf("liuyao.qigua: %v", err)
	}
	if getStr(result, "_product") != "liuyao_qigua" {
		t.Errorf("_product = %q, want liuyao_qigua", getStr(result, "_product"))
	}
	var env struct {
		Data map[string]any `json:"data"`
	}
	if err := json.Unmarshal(result, &env); err != nil {
		t.Fatal(err)
	}
	if len(env.Data) != 1 {
		t.Fatalf("qigua data keys = %v, want only casting", env.Data)
	}
	casting, ok := env.Data["casting"].(map[string]any)
	if !ok || casting["schema_version"] != "liuyao-cast-v1" {
		t.Fatalf("casting receipt = %#v, want liuyao-cast-v1", env.Data["casting"])
	}
}

func TestHandler_LiuyaoQigua_ManualCoins(t *testing.T) {
	r := NewRPCRegistry()
	rounds := []string{
		`["正","反","反"]`, `["正","正","反"]`, `["正","正","正"]`,
		`["反","反","反"]`, `["正","正","反"]`, `["正","反","反"]`,
	}
	params := `{"mode":"coins","rounds":[` + strings.Join(rounds, ",") + `]}`
	result, err := r.Execute(context.Background(), "liuyao.qigua", json.RawMessage(params))
	if err != nil {
		t.Fatalf("liuyao.qigua coins: %v", err)
	}
	var env struct {
		Data struct {
			Casting struct {
				Mode string `json:"mode"`
				Yaos [6]int `json:"yaos"`
			} `json:"casting"`
		} `json:"data"`
	}
	if err := json.Unmarshal(result, &env); err != nil {
		t.Fatal(err)
	}
	want := [6]int{7, 8, 9, 6, 8, 7}
	if env.Data.Casting.Yaos != want || env.Data.Casting.Mode != "coins" {
		t.Fatalf("got=%+v, want casting yaos=%v mode=coins", env.Data.Casting, want)
	}
}

func TestHandler_LiuyaoChart_RejectsLegacyYaosInput(t *testing.T) {
	r := NewRPCRegistry()
	params := fmt.Sprintf(`{"solar_time":%s,"yaos":[7,7,7,7,7,7]}`, btOK)
	_, err := r.Execute(context.Background(), "liuyao.chart", json.RawMessage(params))
	if err == nil || !strings.Contains(err.Error(), "yaos") {
		t.Fatalf("error = %v, want legacy yaos rejected", err)
	}
}

// ── liuyao.chart — casting required ──

func TestHandler_LiuyaoChart_InvalidYongShen(t *testing.T) {
	r := NewRPCRegistry()
	params := json.RawMessage(fmt.Sprintf(
		`{"solar_time":%s,"yong_shen":"invalid","casting":%s}`,
		btOK, liuyaoCastingJSON(t, r, [6]int{7, 7, 7, 7, 7, 7}),
	))
	_, err := r.Execute(context.Background(), "liuyao.chart", params)
	if err == nil {
		t.Error("expected error for invalid yong_shen")
	}
}

func TestHandler_LiuyaoChart_MissingInput(t *testing.T) {
	r := NewRPCRegistry()
	params := json.RawMessage(fmt.Sprintf(`{"solar_time":%s}`, btOK))
	_, err := r.Execute(context.Background(), "liuyao.chart", params)
	if err == nil {
		t.Fatal("expected error for missing casting")
	}
	if !strings.Contains(err.Error(), "casting") {
		t.Errorf("error = %q, want 'casting is required'", err.Error())
	}
}

// =============================================================================
// ziwei dependent methods — need chart first
// =============================================================================

func TestHandler_ZiweiDaXian(t *testing.T) {
	r := NewRPCRegistry()
	// Get chart first
	chartResult, err := r.Execute(context.Background(), "ziwei.chart",
		json.RawMessage(fmt.Sprintf(`{"lunar":%s,"gender":"male"}`, lunarOK)))
	if err != nil {
		t.Fatalf("ziwei.chart: %v", err)
	}
	var env struct {
		Data json.RawMessage `json:"data"`
	}
	if err := json.Unmarshal(chartResult, &env); err != nil {
		t.Fatal(err)
	}
	params := json.RawMessage(fmt.Sprintf(`{"chart":%s}`, env.Data))
	result, err := r.Execute(context.Background(), "ziwei.daxian", params)
	if err != nil {
		t.Fatalf("ziwei.daxian: %v", err)
	}
	if getStr(result, "_product") != "ziwei_daxian" {
		t.Errorf("_product = %q, want ziwei_daxian", getStr(result, "_product"))
	}
}

func TestHandler_ZiweiLiunian(t *testing.T) {
	r := NewRPCRegistry()
	chartResult, err := r.Execute(context.Background(), "ziwei.chart",
		json.RawMessage(fmt.Sprintf(`{"lunar":%s,"gender":"male"}`, lunarOK)))
	if err != nil {
		t.Fatalf("ziwei.chart: %v", err)
	}
	var env struct {
		Data json.RawMessage `json:"data"`
	}
	if err := json.Unmarshal(chartResult, &env); err != nil {
		t.Fatal(err)
	}
	params := json.RawMessage(fmt.Sprintf(`{"lunar_year":2026,"chart":%s}`, env.Data))
	result, err := r.Execute(context.Background(), "ziwei.liunian", params)
	if err != nil {
		t.Fatalf("ziwei.liunian: %v", err)
	}
	if getStr(result, "_product") != "ziwei_liunian" {
		t.Errorf("_product = %q, want ziwei_liunian", getStr(result, "_product"))
	}
}

func TestHandler_ZiweiLiuyue(t *testing.T) {
	r := NewRPCRegistry()
	chartResult, err := r.Execute(context.Background(), "ziwei.chart",
		json.RawMessage(fmt.Sprintf(`{"lunar":%s,"gender":"male"}`, lunarOK)))
	if err != nil {
		t.Fatalf("ziwei.chart: %v", err)
	}
	var env struct {
		Data json.RawMessage `json:"data"`
	}
	if err := json.Unmarshal(chartResult, &env); err != nil {
		t.Fatal(err)
	}
	params := json.RawMessage(fmt.Sprintf(`{"target_lunar":{"year":2026,"month":5,"day":10,"leap":false},"chart":%s}`, env.Data))
	result, err := r.Execute(context.Background(), "ziwei.liuyue", params)
	if err != nil {
		t.Fatalf("ziwei.liuyue: %v", err)
	}
	if getStr(result, "_product") != "ziwei_liuyue" {
		t.Errorf("_product = %q, want ziwei_liuyue", getStr(result, "_product"))
	}
}

func TestHandler_ZiweiBond(t *testing.T) {
	r := NewRPCRegistry()
	chartA, err := r.Execute(context.Background(), "ziwei.chart",
		json.RawMessage(fmt.Sprintf(`{"lunar":%s,"gender":"male"}`, lunarOK)))
	if err != nil {
		t.Fatalf("ziwei.chart A: %v", err)
	}
	chartB, err := r.Execute(context.Background(), "ziwei.chart",
		json.RawMessage(fmt.Sprintf(`{"lunar":%s,"gender":"female"}`, lunarOK2)))
	if err != nil {
		t.Fatalf("ziwei.chart B: %v", err)
	}
	var envA, envB struct {
		Data json.RawMessage `json:"data"`
	}
	if err := json.Unmarshal(chartA, &envA); err != nil {
		t.Fatal(err)
	}
	if err := json.Unmarshal(chartB, &envB); err != nil {
		t.Fatal(err)
	}
	params := json.RawMessage(fmt.Sprintf(`{"a":%s,"b":%s}`, envA.Data, envB.Data))
	result, err := r.Execute(context.Background(), "ziwei.bond", params)
	if err != nil {
		t.Fatalf("ziwei.bond: %v", err)
	}
	if getStr(result, "_product") != "ziwei_bond" {
		t.Errorf("_product = %q, want ziwei_bond", getStr(result, "_product"))
	}
}

func TestHandler_ZiweiLiuri_Valid(t *testing.T) {
	r := NewRPCRegistry()
	chartResult, err := r.Execute(context.Background(), "ziwei.chart",
		json.RawMessage(fmt.Sprintf(`{"lunar":%s,"gender":"male"}`, lunarOK)))
	if err != nil {
		t.Fatalf("ziwei.chart: %v", err)
	}
	var env struct {
		Data json.RawMessage `json:"data"`
	}
	if err := json.Unmarshal(chartResult, &env); err != nil {
		t.Fatal(err)
	}
	params := json.RawMessage(fmt.Sprintf(`{"target_lunar":{"year":2026,"month":5,"day":10,"leap":false},"chart":%s}`, env.Data))
	result, err := r.Execute(context.Background(), "ziwei.liuri", params)
	if err != nil {
		t.Fatalf("ziwei.liuri: %v", err)
	}
	if getStr(result, "_product") != "ziwei_liuri" {
		t.Errorf("_product = %q, want ziwei_liuri", getStr(result, "_product"))
	}
	if !hasKey(result, "data") {
		t.Error("missing data")
	}
}

func TestHandler_BaziFullChart_Valid(t *testing.T) {
	r := NewRPCRegistry()
	// 先拿 bazi.chart
	chartResult, err := r.Execute(context.Background(), "bazi.chart",
		json.RawMessage(fmt.Sprintf(`{"solar_time":%s,"gender":"male"}`, btOK)))
	if err != nil {
		t.Fatalf("bazi.chart: %v", err)
	}
	var env struct {
		Data json.RawMessage `json:"data"`
	}
	if err := json.Unmarshal(chartResult, &env); err != nil {
		t.Fatal(err)
	}
	// 把 chart 传入 fullchart
	fullResult, err := r.Execute(context.Background(), "bazi.fullchart",
		json.RawMessage(fmt.Sprintf(`{"chart":%s}`, env.Data)))
	if err != nil {
		t.Fatalf("bazi.fullchart: %v", err)
	}
	var full struct {
		Data map[string]any `json:"data"`
	}
	if err := json.Unmarshal(fullResult, &full); err != nil {
		t.Fatal(err)
	}
	// 验证完整结果包含扩展字段
	for _, pillar := range []string{"nian", "yue", "ri", "shi"} {
		p, ok := full.Data[pillar].(map[string]any)
		if !ok {
			t.Errorf("%s not found in fullchart result", pillar)
			continue
		}
		// 应有 cang_gan（扩展字段）
		if _, exists := p["cang_gan"]; !exists {
			t.Errorf("%s.cang_gan missing in fullchart", pillar)
		}
		if _, exists := p["shi_shens"]; !exists {
			t.Errorf("%s.shi_shens missing in fullchart", pillar)
		}
		// 不应有额外字段
		for _, field := range []string{"is_void", "day_master_trend", "xun", "xun_kong"} {
			if _, exists := p[field]; !exists {
				t.Errorf("%s.%s missing in fullchart", pillar, field)
			}
		}
	}
}

// ── Remaining untested methods ──

func TestHandler_BaziBond_Valid(t *testing.T) {
	r := NewRPCRegistry()
	c1, err := r.Execute(context.Background(), "bazi.chart",
		json.RawMessage(fmt.Sprintf(`{"solar_time":%s,"gender":"male"}`, btOK)))
	if err != nil {
		t.Fatalf("execute: %v", err)
	}
	c2, err := r.Execute(context.Background(), "bazi.chart",
		json.RawMessage(fmt.Sprintf(`{"solar_time":%s,"gender":"female"}`, btOK)))
	if err != nil {
		t.Fatalf("execute: %v", err)
	}
	var e1, e2 struct {
		Data json.RawMessage `json:"data"`
	}
	if err := json.Unmarshal(c1, &e1); err != nil {
		t.Fatal(err)
	}
	if err := json.Unmarshal(c2, &e2); err != nil {
		t.Fatal(err)
	}
	_, err = r.Execute(context.Background(), "bazi.bond",
		json.RawMessage(fmt.Sprintf(`{"a":{"chart":%s},"b":{"chart":%s}}`, e1.Data, e2.Data)))
	if err != nil {
		t.Fatalf("bazi.bond: %v", err)
	}
}

func TestHandler_BaziXiaoyun_Valid(t *testing.T) {
	r := NewRPCRegistry()
	chartResult, err := r.Execute(context.Background(), "bazi.chart",
		json.RawMessage(fmt.Sprintf(`{"solar_time":%s,"gender":"male"}`, btOK)))
	if err != nil {
		t.Fatalf("execute: %v", err)
	}
	var env struct {
		Data json.RawMessage `json:"data"`
	}
	if err := json.Unmarshal(chartResult, &env); err != nil {
		t.Fatal(err)
	}
	_, err = r.Execute(context.Background(), "bazi.xiaoyun",
		json.RawMessage(fmt.Sprintf(`{"chart":%s,"count":3}`, env.Data)))
	if err != nil {
		t.Fatalf("bazi.xiaoyun: %v", err)
	}
}

func TestHandler_ZiweiLiuyue_Valid(t *testing.T) {
	r := NewRPCRegistry()
	chartResult, err := r.Execute(context.Background(), "ziwei.chart",
		json.RawMessage(fmt.Sprintf(`{"lunar":%s,"gender":"male"}`, lunarOK)))
	if err != nil {
		t.Fatalf("execute: %v", err)
	}
	var env struct {
		Data json.RawMessage `json:"data"`
	}
	if err := json.Unmarshal(chartResult, &env); err != nil {
		t.Fatal(err)
	}
	_, err = r.Execute(context.Background(), "ziwei.liuyue",
		json.RawMessage(fmt.Sprintf(`{"target_lunar":{"year":2026,"month":5,"day":10,"leap":false},"chart":%s}`, env.Data)))
	if err != nil {
		t.Fatalf("ziwei.liuyue: %v", err)
	}
}

func TestHandler_ZiweiBond_Valid(t *testing.T) {
	r := NewRPCRegistry()
	c1, err := r.Execute(context.Background(), "ziwei.chart",
		json.RawMessage(fmt.Sprintf(`{"lunar":%s,"gender":"male"}`, lunarOK)))
	if err != nil {
		t.Fatalf("execute: %v", err)
	}
	c2, err := r.Execute(context.Background(), "ziwei.chart",
		json.RawMessage(fmt.Sprintf(`{"lunar":%s,"gender":"female"}`, lunarOK)))
	if err != nil {
		t.Fatalf("execute: %v", err)
	}
	var e1, e2 struct {
		Data json.RawMessage `json:"data"`
	}
	if err := json.Unmarshal(c1, &e1); err != nil {
		t.Fatal(err)
	}
	if err := json.Unmarshal(c2, &e2); err != nil {
		t.Fatal(err)
	}
	_, err = r.Execute(context.Background(), "ziwei.bond",
		json.RawMessage(fmt.Sprintf(`{"a":%s,"b":%s}`, e1.Data, e2.Data)))
	if err != nil {
		t.Fatalf("ziwei.bond: %v", err)
	}
}
