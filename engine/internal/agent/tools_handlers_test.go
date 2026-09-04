package agent

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"strings"
	"testing"

	"liki-engine/internal/agent/city"
	"liki-engine/internal/engine/ganzhi"
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

// ── TimePoint.Timeset ──

func TestTimePoint_Timeset(t *testing.T) {
	tests := []struct {
		name    string
		time    string
		wantErr bool
	}{
		{"有效时间", "1984-02-15T08:00:00+08:00", false},
		{"无效格式", "1984-02-15", true},
		{"空字符串", "", true},
		{"无时区", "1984-02-15T08:00:00", true},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			tp := TimePoint{Time: tt.time, Longitude: 116.4}
			_, err := tp.Timeset()
			if (err != nil) != tt.wantErr {
				t.Errorf("Timeset(%q) err=%v, wantErr=%v", tt.time, err, tt.wantErr)
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
		"qiming.pick", "qiming.compose", "qiming.check",
		"bazhai.chart", "bazhai.layout",
		"xuankong.chart", "xuankong.liunian",
		"liuyao.qigua", "liuyao.chart",
		"huangli.days",
		"city.coords",
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
		"bazi.chart", "ziwei.chart", "bazhai.chart",
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
		{"bazhai.chart", fmt.Sprintf(`{"solar_time":%s,"gender":"bad"}`, btOK)},
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
		{"qiming.pick", `{}`},
		{"qiming.check", `{}`},
		{"xuankong.chart", fmt.Sprintf(`{"solar_time":%s}`, btOK)},
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
		{"qimen.chart", fmt.Sprintf(`{"solar_time":%s,"kind":"invalid"}`, btOK)},
		{"xuankong.chart", fmt.Sprintf(`{"solar_time":%s,"zuo_shan":-1,"xiang_shan":0}`, btOK)},
		{"xuankong.chart", fmt.Sprintf(`{"solar_time":%s,"zuo_shan":0,"xiang_shan":24}`, btOK)},
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

func TestHandler_QimenKindDefault(t *testing.T) {
	r := NewRPCRegistry()
	params := json.RawMessage(fmt.Sprintf(`{"solar_time":%s}`, btOK))
	result, err := r.Execute(context.Background(), "qimen.chart", params)
	if err != nil {
		t.Fatalf("qimen.chart (default kind): %v", err)
	}
	if !hasKey(result, "_product") || !hasKey(result, "data") {
		t.Error("expected envelope with _product and data")
	}
}

func TestHandler_QimenChart_Compatible(t *testing.T) {
	r := NewRPCRegistry()
	params := json.RawMessage(fmt.Sprintf(`{"solar_time":%s}`, btOK))
	result, err := r.Execute(context.Background(), "qimen.chart", params)
	if err != nil {
		t.Fatalf("qimen.chart: %v", err)
	}
	if getStr(result, "_product") != "qimen" {
		t.Errorf("_product = %q, want qimen", getStr(result, "_product"))
	}
	if !hasKey(result, "data") {
		t.Error("missing data")
	}
}

func TestHandler_LiuyaoYongShenDefault(t *testing.T) {
	r := NewRPCRegistry()
	params := json.RawMessage(fmt.Sprintf(`{"solar_time":%s,"yaos":[7,7,7,7,7,7]}`, btOK))
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

func TestHandler_ComputeBazhaiChart_Valid(t *testing.T) {
	r := NewRPCRegistry()
	params := json.RawMessage(fmt.Sprintf(`{"solar_time":%s,"gender":"male"}`, btOK))
	result, err := r.Execute(context.Background(), "bazhai.chart", params)
	if err != nil {
		t.Fatalf("bazhai.chart: %v", err)
	}
	if getStr(result, "_product") != "bazhai" {
		t.Errorf("_product = %q, want bazhai", getStr(result, "_product"))
	}
	if !hasKey(result, "data") {
		t.Error("missing data")
	}
}

func TestHandler_ComputeXuankongChart_Valid(t *testing.T) {
	r := NewRPCRegistry()
	params := json.RawMessage(fmt.Sprintf(`{"solar_time":%s,"zuo_shan":0,"xiang_shan":12}`, btOK))
	result, err := r.Execute(context.Background(), "xuankong.chart", params)
	if err != nil {
		t.Fatalf("xuankong.chart: %v", err)
	}
	if getStr(result, "_product") != "xuankong" {
		t.Errorf("_product = %q, want xuankong", getStr(result, "_product"))
	}
	if !hasKey(result, "data") {
		t.Error("missing data")
	}
}

func TestHandler_HuangliDays_Valid(t *testing.T) {
	r := NewRPCRegistry()
	result, err := r.Execute(context.Background(), "huangli.days", json.RawMessage(`{"start_date":"2026-06-26","count":3}`))
	if err != nil {
		t.Fatalf("huangli.days: %v", err)
	}
	if getStr(result, "_product") != "huangli_days" {
		t.Errorf("_product = %q, want huangli_days", getStr(result, "_product"))
	}
}

func TestHandler_HuangliDays_InvalidDate(t *testing.T) {
	r := NewRPCRegistry()
	_, err := r.Execute(context.Background(), "huangli.days", json.RawMessage(`{"start_date":"not-a-date"}`))
	if err == nil {
		t.Error("expected error for invalid date format")
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
	// fullchart 现在应包含 chart_extra 和 hehui 的数据
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

// ── OpenRPC document ──

func TestOpenRPCDocument(t *testing.T) {
	r := NewRPCRegistry()
	doc := r.OpenRPCDocument()

	var raw map[string]any
	if err := json.Unmarshal(doc, &raw); err != nil {
		t.Fatalf("OpenRPCDocument is not valid JSON: %v", err)
	}
	if raw["openrpc"] != "1.4.1" {
		t.Errorf("openrpc version = %v, want 1.4.1", raw["openrpc"])
	}
	methods, ok := raw["methods"].([]any)
	if !ok {
		t.Fatal("missing methods array")
	}
	if len(methods) != 32 {
		t.Errorf("method count = %d, want 32 (31 + rpc.discover)", len(methods))
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
	// Check yaos and dong_yao exist.
	var env struct {
		Data struct {
			Yaos    [6]int `json:"yaos"`
			DongYao []int  `json:"dong_yao"`
		} `json:"data"`
	}
	if err := json.Unmarshal(result, &env); err != nil {
		t.Fatal(err)
	}
}

// ── liuyao.chart — yaos required ──

func TestHandler_LiuyaoChart_InvalidYongShen(t *testing.T) {
	r := NewRPCRegistry()
	params := json.RawMessage(fmt.Sprintf(`{"solar_time":%s,"yong_shen":"invalid","yaos":[7,7,7,7,7,7]}`, btOK))
	_, err := r.Execute(context.Background(), "liuyao.chart", params)
	if err == nil {
		t.Error("expected error for invalid yong_shen")
	}
}

func TestHandler_LiuyaoChart_MissingYaos(t *testing.T) {
	r := NewRPCRegistry()
	params := json.RawMessage(fmt.Sprintf(`{"solar_time":%s}`, btOK))
	_, err := r.Execute(context.Background(), "liuyao.chart", params)
	if err == nil {
		t.Fatal("expected error for missing yaos")
	}
	if !strings.Contains(err.Error(), "missing property 'yaos'") {
		t.Errorf("error = %q, want 'missing property \\'yaos\\''", err.Error())
	}
}

// ── time.now ──

func TestHandler_TimeNow(t *testing.T) {
	r := NewRPCRegistry()
	result, err := r.Execute(context.Background(), "time.now", json.RawMessage(`{}`))
	if err != nil {
		t.Fatalf("time.now: %v", err)
	}
	if getStr(result, "_product") != "time_now" {
		t.Errorf("_product = %q, want time_now", getStr(result, "_product"))
	}
	// Check CST contains "+08:00" or "CST"
	var env struct {
		Data struct {
			UTC string `json:"utc"`
			CST string `json:"cst"`
		} `json:"data"`
	}
	if err := json.Unmarshal(result, &env); err != nil {
		t.Fatal(err)
	}
	if env.Data.UTC == "" {
		t.Error("utc is empty")
	}
	if env.Data.CST == "" {
		t.Error("cst is empty")
	}
}

func TestHandler_City_Valid(t *testing.T) {
	orig := city.HttpClient()
	city.SetHTTPClient(&http.Client{
		Transport: &mockNominatimTransport{body: `[{"lat":"39.9042","lon":"116.4074","name":"Beijing","address":{"country":"China","country_code":"CN"}}]`},
	})
	defer city.SetHTTPClient(orig)

	r := NewRPCRegistry()
	result, err := r.Execute(context.Background(), "city.coords", json.RawMessage(`{"city":"Beijing"}`))
	if err != nil {
		t.Fatalf("city.coords: %v", err)
	}
	if getStr(result, "_product") != "city_coords" {
		t.Errorf("_product = %q, want city_coords", getStr(result, "_product"))
	}
	if !hasKey(result, "data") {
		t.Error("missing data")
	}
}

// =============================================================================
// qiming
// =============================================================================

func TestHandler_QimingChar_Valid(t *testing.T) {
	r := NewRPCRegistry()
	result, err := r.Execute(context.Background(), "qiming.char", json.RawMessage(`{"char":"林"}`))
	if err != nil {
		t.Fatalf("qiming.char: %v", err)
	}
	if getStr(result, "_product") != "qiming_char" {
		t.Errorf("_product = %q, want qiming_char", getStr(result, "_product"))
	}
	if !hasKey(result, "data") {
		t.Error("missing data")
	}
}

func TestHandler_QimingChar_StrokeSemantics(t *testing.T) {
	r := NewRPCRegistry()
	result, err := r.Execute(context.Background(), "qiming.char", json.RawMessage(`{"char":"郑"}`))
	if err != nil {
		t.Fatalf("qiming.char: %v", err)
	}
	var env struct {
		Data struct {
			Stroke int `json:"stroke"`
		} `json:"data"`
	}
	if err := json.Unmarshal(result, &env); err != nil {
		t.Fatal(err)
	}
	if env.Data.Stroke != 8 {
		t.Fatalf("stroke data = %+v, want modern 8", env.Data)
	}
	var raw struct {
		Data map[string]any `json:"data"`
	}
	if err := json.Unmarshal(result, &raw); err != nil {
		t.Fatal(err)
	}
	for _, field := range []string{"traditional", "kangxi_stroke", "kangxi_form"} {
		if _, exists := raw.Data[field]; exists {
			t.Fatalf("qiming.char output should not contain %s", field)
		}
	}
}

func TestHandler_QimingChar_Errors(t *testing.T) {
	r := NewRPCRegistry()
	for name, params := range map[string]string{
		"unknown":       `{"char":"龍"}`,
		"empty":         `{"char":""}`,
		"missing":       `{}`,
		"unknown_field": `{"char":"林","disallowed":false}`,
	} {
		t.Run(name, func(t *testing.T) {
			if _, err := r.Execute(context.Background(), "qiming.char", json.RawMessage(params)); err == nil {
				t.Fatal("expected error")
			}
		})
	}
}

func TestHandler_QimingPick(t *testing.T) {
	r := NewRPCRegistry()
	result, err := r.Execute(context.Background(), "qiming.pick", json.RawMessage(`{"wuxing1":"木","wuxing2":"火","count":2}`))
	if err != nil {
		t.Fatalf("qiming.pick: %v", err)
	}
	if getStr(result, "_product") != "qiming_pick" {
		t.Errorf("_product = %q, want qiming_pick", getStr(result, "_product"))
	}
	var env struct {
		Data struct {
			Pools []struct {
				Slot  string   `json:"slot"`
				Chars []string `json:"chars"`
			} `json:"pools"`
		} `json:"data"`
	}
	if err := json.Unmarshal(result, &env); err != nil {
		t.Fatal(err)
	}
	if len(env.Data.Pools) != 2 {
		t.Fatalf("pick result = %+v, want two pools", env.Data)
	}
	if env.Data.Pools[0].Slot != "first" || env.Data.Pools[1].Slot != "second" {
		t.Fatalf("slots = %q, %q", env.Data.Pools[0].Slot, env.Data.Pools[1].Slot)
	}
	if len(env.Data.Pools[0].Chars) == 0 || len(env.Data.Pools[1].Chars) == 0 {
		t.Fatal("candidate pools are empty")
	}
}

func TestHandler_QimingPick_SingleName(t *testing.T) {
	r := NewRPCRegistry()
	result, err := r.Execute(context.Background(), "qiming.pick", json.RawMessage(`{"wuxing1":"木","count":1}`))
	if err != nil {
		t.Fatalf("qiming.pick: %v", err)
	}
	var env struct {
		Data struct {
			Pools []struct {
				Slot string `json:"slot"`
			} `json:"pools"`
		} `json:"data"`
	}
	if err := json.Unmarshal(result, &env); err != nil {
		t.Fatal(err)
	}
	if len(env.Data.Pools) != 1 || env.Data.Pools[0].Slot != "first" {
		t.Fatalf("pools = %+v, want one first pool", env.Data.Pools)
	}
}

func TestHandler_QimingPick_DefaultCount(t *testing.T) {
	r := NewRPCRegistry()
	result, err := r.Execute(context.Background(), "qiming.pick", json.RawMessage(`{"wuxing1":"木"}`))
	if err != nil {
		t.Fatalf("qiming.pick: %v", err)
	}
	var env struct {
		Data struct {
			Pools []struct {
				Slot string `json:"slot"`
			} `json:"pools"`
		} `json:"data"`
	}
	if err := json.Unmarshal(result, &env); err != nil {
		t.Fatal(err)
	}
	if len(env.Data.Pools) != 2 {
		t.Fatalf("pools = %+v, want first and second pools", env.Data.Pools)
	}
}

func TestHandler_QimingPick_RejectsInvalidParameters(t *testing.T) {
	r := NewRPCRegistry()
	for name, params := range map[string]string{
		"surname":                  `{"surname":"陈","wuxing1":"木"}`,
		"unknown":                  `{"wuxing1":"木","unknown":false}`,
		"wuxing2_with_single_name": `{"wuxing1":"木","wuxing2":"火","count":1}`,
	} {
		t.Run(name, func(t *testing.T) {
			if _, err := r.Execute(context.Background(), "qiming.pick", json.RawMessage(params)); err == nil {
				t.Fatal("expected error")
			}
		})
	}
}

func TestHandler_QimingCompose(t *testing.T) {
	r := NewRPCRegistry()
	result, err := r.Execute(context.Background(), "qiming.compose", json.RawMessage(`{"first":["林"],"second":["炎"],"max_names":10}`))
	if err != nil {
		t.Fatalf("qiming.compose: %v", err)
	}
	if getStr(result, "_product") != "qiming_compose" {
		t.Errorf("_product = %q, want qiming_compose", getStr(result, "_product"))
	}
	var env struct {
		Data struct {
			TotalPossible int      `json:"total_possible"`
			Names         []string `json:"names"`
		} `json:"data"`
	}
	if err := json.Unmarshal(result, &env); err != nil {
		t.Fatal(err)
	}
	if env.Data.TotalPossible != 1 || len(env.Data.Names) != 1 || env.Data.Names[0] != "林炎" {
		t.Fatalf("compose result = %+v", env.Data)
	}
}

func TestHandler_QimingCompose_SingleName(t *testing.T) {
	r := NewRPCRegistry()
	result, err := r.Execute(context.Background(), "qiming.compose", json.RawMessage(`{"first":["林","桐"],"max_names":10}`))
	if err != nil {
		t.Fatalf("qiming.compose: %v", err)
	}
	var env struct {
		Data struct {
			TotalPossible int      `json:"total_possible"`
			Names         []string `json:"names"`
		} `json:"data"`
	}
	if err := json.Unmarshal(result, &env); err != nil {
		t.Fatal(err)
	}
	if env.Data.TotalPossible != 2 || len(env.Data.Names) != 2 || env.Data.Names[0] != "林" || env.Data.Names[1] != "桐" {
		t.Fatalf("compose result = %+v", env.Data)
	}
}

func TestHandler_QimingCompose_Errors(t *testing.T) {
	r := NewRPCRegistry()
	tests := []struct {
		name   string
		params string
	}{
		{"unknown_char", `{"first":["龍"]}`},
		{"duplicate", `{"first":["林","林"]}`},
		{"unknown_parameter", `{"first":["林"],"count":2}`},
	}
	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			if _, err := r.Execute(context.Background(), "qiming.compose", json.RawMessage(test.params)); err == nil {
				t.Fatal("expected error")
			}
		})
	}
}

func TestHandler_QimingCheck(t *testing.T) {
	r := NewRPCRegistry()
	result, err := r.Execute(context.Background(), "qiming.check", json.RawMessage(`{"given_names":["林炎"],"yongshen":"火","xishen":["木"],"jishen":["水"]}`))
	if err != nil {
		t.Fatalf("qiming.check: %v", err)
	}
	if getStr(result, "_product") != "qiming_check" {
		t.Errorf("_product = %q, want qiming_check", getStr(result, "_product"))
	}
	var env struct {
		Data []struct {
			GivenName string `json:"given_name"`
			Valid     bool   `json:"valid"`
			Phonetic  struct {
				Tones string `json:"tones"`
			} `json:"phonetic"`
			Wuxing struct {
				Yong bool `json:"yong"`
				Xi   bool `json:"xi"`
				Ji   bool `json:"ji"`
			} `json:"wuxing"`
		} `json:"data"`
	}
	if err := json.Unmarshal(result, &env); err != nil {
		t.Fatal(err)
	}
	if len(env.Data) != 1 || !env.Data[0].Valid || env.Data[0].GivenName != "林炎" {
		t.Fatalf("check result = %+v", env.Data)
	}
	if env.Data[0].Phonetic.Tones != "2-2" {
		t.Errorf("tones = %q, want 2-2", env.Data[0].Phonetic.Tones)
	}
	if !env.Data[0].Wuxing.Yong || !env.Data[0].Wuxing.Xi || env.Data[0].Wuxing.Ji {
		t.Fatalf("wuxing = %+v", env.Data[0].Wuxing)
	}
}

func TestHandler_QimingCheck_UnknownChar(t *testing.T) {
	r := NewRPCRegistry()
	result, err := r.Execute(context.Background(), "qiming.check", json.RawMessage(`{"given_names":["龍明"]}`))
	if err != nil {
		t.Fatalf("qiming.check: %v", err)
	}
	var env struct {
		Data []struct {
			Valid  bool `json:"valid"`
			Errors []struct {
				Code string `json:"code"`
				Char string `json:"char"`
			} `json:"errors"`
		} `json:"data"`
	}
	if err := json.Unmarshal(result, &env); err != nil {
		t.Fatal(err)
	}
	if len(env.Data) != 1 || env.Data[0].Valid {
		t.Fatalf("check result = %+v", env.Data)
	}
	if len(env.Data[0].Errors) == 0 || env.Data[0].Errors[0].Code != "character_not_found" {
		t.Fatalf("errors = %+v", env.Data[0].Errors)
	}
}

func TestHandler_QimingCheck_WithoutWuxingConstraints(t *testing.T) {
	r := NewRPCRegistry()
	result, err := r.Execute(context.Background(), "qiming.check", json.RawMessage(`{"given_names":["林炎"]}`))
	if err != nil {
		t.Fatalf("qiming.check: %v", err)
	}
	var env struct {
		Data []struct {
			Valid  bool           `json:"valid"`
			Wuxing map[string]any `json:"wuxing"`
		} `json:"data"`
	}
	if err := json.Unmarshal(result, &env); err != nil {
		t.Fatal(err)
	}
	if len(env.Data) != 1 || !env.Data[0].Valid || env.Data[0].Wuxing != nil {
		t.Fatalf("check result = %+v", env.Data)
	}
}

func TestHandler_QimingCheck_RejectsUnknownParameter(t *testing.T) {
	r := NewRPCRegistry()
	if _, err := r.Execute(context.Background(), "qiming.check", json.RawMessage(`{"given_names":["林炎"],"unknown":false}`)); err == nil {
		t.Fatal("expected error for unknown parameter")
	}
}

// =============================================================================
// tianwen.time
// =============================================================================

func TestHandler_TianwenTime(t *testing.T) {
	r := NewRPCRegistry()
	params := json.RawMessage(`{"time":"2000-06-15T12:00:00+08:00","longitude":116.4}`)
	result, err := r.Execute(context.Background(), "tianwen.time", params)
	if err != nil {
		t.Fatalf("tianwen.time: %v", err)
	}
	var env struct {
		Data json.RawMessage `json:"data"`
	}
	if err := json.Unmarshal(result, &env); err != nil {
		t.Fatalf("unmarshal envelope: %v", err)
	}
	var ts struct {
		Solar     string `json:"solar"`
		Gregorian string `json:"gregorian"`
		Lunar     struct {
			Year  int `json:"year"`
			Month int `json:"month"`
			Day   int `json:"day"`
		} `json:"lunar"`
	}
	if err := json.Unmarshal(env.Data, &ts); err != nil {
		t.Fatalf("unmarshal: %v", err)
	}
	if ts.Solar == "" {
		t.Error("solar is empty")
	}
	if ts.Gregorian == "" {
		t.Error("gregorian is empty")
	}
	if ts.Lunar.Year == 0 {
		t.Error("lunar year is zero")
	}
}

func TestHandler_TianwenTime_InvalidTime(t *testing.T) {
	r := NewRPCRegistry()
	params := json.RawMessage(`{"time":"","longitude":116.4}`)
	assertError(t, r, "tianwen.time", params)
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
	params := json.RawMessage(fmt.Sprintf(`{"lunar_year":2026,"lunar_month":5,"chart":%s}`, env.Data))
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
	params := json.RawMessage(fmt.Sprintf(`{"lunar_year":2026,"lunar_month":5,"lunar_day":10,"chart":%s}`, env.Data))
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

// mockNominatimTransport simulates a Nominatim API response for city handler tests.
type mockNominatimTransport struct {
	body string
}

func (m *mockNominatimTransport) RoundTrip(_ *http.Request) (*http.Response, error) {
	return &http.Response{
		StatusCode: http.StatusOK,
		Body:       io.NopCloser(strings.NewReader(m.body)),
		Header:     http.Header{"Content-Type": []string{"application/json"}},
	}, nil
}

// ── Judgment / Annual handler tests ──

func TestHandler_XuankongLiunian_Valid(t *testing.T) {
	r := NewRPCRegistry()
	chartResult, err := r.Execute(context.Background(), "xuankong.chart",
		json.RawMessage(fmt.Sprintf(`{"solar_time":%s,"zuo_shan":20,"xiang_shan":8}`, btOK)))
	if err != nil {
		t.Fatalf("xuankong.chart: %v", err)
	}
	var env struct {
		Data json.RawMessage `json:"data"`
	}
	if err := json.Unmarshal(chartResult, &env); err != nil {
		t.Fatal(err)
	}
	result, err := r.Execute(context.Background(), "xuankong.liunian",
		json.RawMessage(fmt.Sprintf(`{"chart":%s,"year":2026}`, env.Data)))
	if err != nil {
		t.Fatalf("xuankong.liunian: %v", err)
	}
	if getStr(result, "_product") != "xuankong_liunian" {
		t.Errorf("_product = %q, want xuankong_liunian", getStr(result, "_product"))
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
		if _, exists := p["is_void"]; !exists {
			t.Errorf("%s.is_void missing in fullchart", pillar)
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
		json.RawMessage(fmt.Sprintf(`{"lunar_year":2026,"lunar_month":5,"chart":%s}`, env.Data)))
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
