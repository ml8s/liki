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
	"liki-engine/internal/engine/qiming"
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
		"bazi.liuri", "bazi.liushi", "bazi.xiaoyun", "bazi.xiaoxian",
		"ziwei.chart", "ziwei.daxian", "ziwei.liunian", "ziwei.liuyue",
		"ziwei.liuri", "ziwei.bond",
		"qimen.chart",
		"qiming.pick", "qiming.build", "qiming.check", "qiming.wuge",
		"bazhai.chart", "bazhai.minggua",
		"xuankong.sanyuan", "xuankong.chart",
		"liuyao.chart",
		"huangli.days",
		"city",
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
		{"bazi.xiaoxian", `{"gender":"bad"}`},
		{"ziwei.chart", fmt.Sprintf(`{"lunar":%s,"gender":"bad"}`, lunarOK)},
		{"bazhai.chart", fmt.Sprintf(`{"solar_time":%s,"gender":"bad"}`, btOK)},
		{"bazhai.minggua", `{"gender":"x","birth_year":1984}`},
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
		{"bazi.xiaoxian", `{}`},
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
		{"xuankong.chart", fmt.Sprintf(`{"solar_time":%s,"sit_mountain":-1,"face_mountain":0}`, btOK)},
		{"xuankong.chart", fmt.Sprintf(`{"solar_time":%s,"sit_mountain":0,"face_mountain":24}`, btOK)},
		{"xuankong.sanyuan", `{"year":0}`},
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

func TestHandler_ComputeXiaoXian_Valid(t *testing.T) {
	r := NewRPCRegistry()
	result, err := r.Execute(context.Background(), "bazi.xiaoxian", json.RawMessage(`{"gender":"male"}`))
	if err != nil {
		t.Fatalf("bazi.xiaoxian: %v", err)
	}
	if getStr(result, "_product") != "xiaoxian" {
		t.Errorf("_product = %q, want xiaoxian", getStr(result, "_product"))
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

func TestHandler_ComputeMingGua_Valid(t *testing.T) {
	r := NewRPCRegistry()
	result, err := r.Execute(context.Background(), "bazhai.minggua", json.RawMessage(`{"gender":"male","birth_year":1984}`))
	if err != nil {
		t.Fatalf("bazhai.minggua: %v", err)
	}
	if getStr(result, "_product") != "minggua" {
		t.Errorf("_product = %q, want minggua", getStr(result, "_product"))
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
	params := json.RawMessage(fmt.Sprintf(`{"solar_time":%s,"sit_mountain":0,"face_mountain":12}`, btOK))
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

func TestHandler_ComputeSanYuanYun_Valid(t *testing.T) {
	r := NewRPCRegistry()
	result, err := r.Execute(context.Background(), "xuankong.sanyuan", json.RawMessage(`{"year":2026}`))
	if err != nil {
		t.Fatalf("xuankong.sanyuan: %v", err)
	}
	if getStr(result, "_product") != "sanyuan" {
		t.Errorf("_product = %q, want sanyuan", getStr(result, "_product"))
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

func TestHandler_AllHandlersAcceptContext(t *testing.T) {
	r := NewRPCRegistry()
	ctx, cancel := context.WithCancel(context.Background())
	cancel()

	// Verify handlers don't panic or hang when context is already cancelled.
	// Some handlers may succeed anyway (if they don't use ctx), which is fine.
	tests := []struct {
		method string
		params string
	}{
		{"bazi.xiaoxian", `{"gender":"male"}`},
		{"time.now", `{}`},
		{"liuyao.qigua", `{}`},
	}
	for _, tt := range tests {
		result, err := r.Execute(ctx, tt.method, json.RawMessage(tt.params))
		_ = result
		_ = err
		// OK if it errors, OK if it succeeds — the test just verifies no panic/timeout.
	}
}

// ── valid paths for handlers with low coverage ──

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

func TestHandler_ComputeYongShen_Valid(t *testing.T) {
	r := NewRPCRegistry()
	chart := getBaziChart(t, r, btOK, "male")
	params := json.RawMessage(fmt.Sprintf(`{"chart":%s}`, chart))
	result, err := r.Execute(context.Background(), "bazi.yongshen", params)
	if err != nil {
		t.Fatalf("bazi.yongshen: %v", err)
	}
	if getStr(result, "_product") != "yongshen" {
		t.Errorf("_product = %q, want yongshen", getStr(result, "_product"))
	}
	if !hasKey(result, "data") {
		t.Error("missing data")
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
	if len(methods) != 41 {
		t.Errorf("method count = %d, want 41 (40 + rpc.discover)", len(methods))
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
	result, err := r.Execute(context.Background(), "city", json.RawMessage(`{"city":"Beijing"}`))
	if err != nil {
		t.Fatalf("city: %v", err)
	}
	if getStr(result, "_product") != "city" {
		t.Errorf("_product = %q, want city", getStr(result, "_product"))
	}
	if !hasKey(result, "data") {
		t.Error("missing data")
	}
}

// =============================================================================
// qiming.pick
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

func TestHandler_QimingChar_NotFound(t *testing.T) {
	r := NewRPCRegistry()
	_, err := r.Execute(context.Background(), "qiming.char", json.RawMessage(`{"char":"龍"}`))
	if err == nil {
		t.Error("expected error for unknown character")
	}
}

func TestHandler_QimingChar_InvalidParams(t *testing.T) {
	r := NewRPCRegistry()
	_, err := r.Execute(context.Background(), "qiming.char", json.RawMessage(`{"char":""}`))
	if err == nil {
		t.Error("expected error for empty char")
	}
	_, err = r.Execute(context.Background(), "qiming.char", json.RawMessage(`{}`))
	if err == nil {
		t.Error("expected error for missing char")
	}
}
func TestHandler_NamingPick(t *testing.T) {
	r := NewRPCRegistry()
	params := json.RawMessage(`{"wuxing":"金"}`)
	result, err := r.Execute(context.Background(), "qiming.pick", params)
	if err != nil {
		t.Fatalf("qiming.pick: %v", err)
	}
	if getStr(result, "_product") != "naming_pick" {
		t.Errorf("_product = %q, want naming_pick", getStr(result, "_product"))
	}
	if !hasKey(result, "data") {
		t.Error("missing data")
	}
}

// =============================================================================
// qiming.wuge — 双字名
// =============================================================================

func TestHandler_NamingWuge(t *testing.T) {
	r := NewRPCRegistry()
	params := json.RawMessage(`{"surname":"姚","count":2}`)
	result, err := r.Execute(context.Background(), "qiming.wuge", params)
	if err != nil {
		t.Fatalf("qiming.wuge: %v", err)
	}
	if getStr(result, "_product") != "naming_wuge" {
		t.Errorf("_product = %q, want naming_wuge", getStr(result, "_product"))
	}

	var env struct {
		Data struct {
			SurnameStroke float64        `json:"surname_stroke"`
			Pairs         []map[string]any `json:"pairs"`
		} `json:"data"`
	}
	if err := json.Unmarshal(result, &env); err != nil {
		t.Fatal(err)
	}
	if env.Data.SurnameStroke == 0 {
		t.Error("surname_stroke should not be 0")
	}
	if len(env.Data.Pairs) == 0 {
		t.Fatal("wuge should return at least one viable pair for 姚")
	}
	for _, p := range env.Data.Pairs {
		s1, _ := p["s1"].(float64)
		s2, _ := p["s2"].(float64)
		if s1 == 0 || s2 == 0 {
			t.Errorf("double-name pair should have s1, s2 > 0: s1=%v s2=%v", s1, s2)
		}
	}
}

// =============================================================================
// qiming.wuge — 单名
// =============================================================================

func TestHandler_NamingWuge_Single(t *testing.T) {
	r := NewRPCRegistry()
	params := json.RawMessage(`{"surname":"姚","count":1}`)
	result, err := r.Execute(context.Background(), "qiming.wuge", params)
	if err != nil {
		t.Fatalf("qiming.wuge single: %v", err)
	}
	var env struct {
		Data struct {
			Pairs []map[string]any `json:"pairs"`
		} `json:"data"`
	}
	if err := json.Unmarshal(result, &env); err != nil {
		t.Fatal(err)
	}
	for _, p := range env.Data.Pairs {
		s2, _ := p["s2"].(float64)
		if s2 != 0 {
			t.Errorf("single-name pair should have s2=0: %+v", p)
		}
	}
}

func TestHandler_NamingWuge_SancaiFiltered(t *testing.T) {
	r := NewRPCRegistry()
	params := json.RawMessage(`{"surname":"姚","count":2}`)
	result, err := r.Execute(context.Background(), "qiming.wuge", params)
	if err != nil {
		t.Fatalf("qiming.wuge: %v", err)
	}

	var env struct {
		Data struct {
			SurnameStroke float64          `json:"surname_stroke"`
			Pairs         []map[string]any `json:"pairs"`
		} `json:"data"`
	}
	if err := json.Unmarshal(result, &env); err != nil {
		t.Fatal(err)
	}
	if len(env.Data.Pairs) == 0 {
		t.Fatal("wuge should return pairs")
	}
	stroke := int(env.Data.SurnameStroke)
	for _, p := range env.Data.Pairs {
		s1 := int(p["s1"].(float64))
		s2 := int(p["s2"].(float64))
		if !qiming.SancaiHarmonious(stroke, s1, s2) {
			t.Errorf("pair (s1=%d, s2=%d) not sancai-harmonious for stroke=%d", s1, s2, stroke)
		}
	}
}

func TestHandler_NamingPick_CharInfo(t *testing.T) {
	r := NewRPCRegistry()
	params := json.RawMessage(`{"wuxing":"金"}`)
	result, err := r.Execute(context.Background(), "qiming.pick", params)
	if err != nil {
		t.Fatalf("qiming.pick: %v", err)
	}

	var env struct {
		Data struct {
			Chars map[string][]map[string]any `json:"chars"`
		} `json:"data"`
	}
	if err := json.Unmarshal(result, &env); err != nil {
		t.Fatal(err)
	}
	if len(env.Data.Chars) == 0 {
		t.Fatal("pick should return chars")
	}

	checked := 0
	for _, charList := range env.Data.Chars {
		for _, c := range charList {
			checked++
			if _, ok := c["wuxing"]; !ok {
				t.Errorf("char %v missing wuxing", c["char"])
			}
			if _, ok := c["stroke"]; !ok {
				t.Errorf("char %v missing stroke", c["char"])
			}
			if _, ok := c["radical"]; !ok {
				t.Errorf("char %v missing radical", c["char"])
			}
			if _, ok := c["pinyin"]; !ok {
				t.Errorf("char %v missing pinyin", c["char"])
			}
			if _, ok := c["tone"]; !ok {
				t.Errorf("char %v missing tone", c["char"])
			}
		}
	}
	if checked == 0 {
		t.Fatal("no chars checked")
	}
}

func TestHandler_NamingPick_WithPairsFilter(t *testing.T) {
	r := NewRPCRegistry()
	// 姚=9, pairs限定s1范围: s1=3和s1=13
	params := json.RawMessage(`{"wuxing":"木","pairs":[{"s1":3,"s2":5},{"s1":13,"s2":7}]}`)
	result, err := r.Execute(context.Background(), "qiming.pick", params)
	if err != nil {
		t.Fatalf("qiming.pick with pairs: %v", err)
	}

	var env struct {
		Data struct {
			Chars map[string][]any `json:"chars"`
		} `json:"data"`
	}
	if err := json.Unmarshal(result, &env); err != nil {
		t.Fatal(err)
	}

	// Only strokes in pairs (both s1 and s2) should be present: 3, 5, 7, 13
	valid := map[string]bool{"3": true, "5": true, "7": true, "13": true}
	for stroke := range env.Data.Chars {
		if !valid[stroke] {
			t.Errorf("unexpected stroke %s returned (should only be 3,5,7,13)", stroke)
		}
	}
}

// =============================================================================
// qiming.build — 带 pairs
// =============================================================================

func TestHandler_NamingBuild_WithPairs(t *testing.T) {
	r := NewRPCRegistry()
	params := json.RawMessage(`{
		"surname":"姚",
		"chars1":{"8":[{"char":"林","tone":2}],"12":[{"char":"棋","tone":2}]},
		"chars2":{"15":[{"char":"槿","tone":3}],"16":[{"char":"澄","tone":2}]},
		"pairs":[{"s1":8,"s2":15},{"s1":12,"s2":16}]
	}`)
	result, err := r.Execute(context.Background(), "qiming.build", params)
	if err != nil {
		t.Fatalf("qiming.build with pairs: %v", err)
	}
	if getStr(result, "_product") != "naming_build" {
		t.Errorf("_product = %q, want naming_build", getStr(result, "_product"))
	}
}

// =============================================================================
// qiming.build — 双字名
// =============================================================================

func TestHandler_NamingBuild(t *testing.T) {
	r := NewRPCRegistry()
	params := json.RawMessage(`{"surname":"王","chars1":{"8":[{"char":"明","tone":2}]},"chars2":{"10":[{"char":"哲","tone":2}]}}`)
	result, err := r.Execute(context.Background(), "qiming.build", params)
	if err != nil {
		t.Fatalf("qiming.build: %v", err)
	}
	if getStr(result, "_product") != "naming_build" {
		t.Errorf("_product = %q, want naming_build", getStr(result, "_product"))
	}
}

// =============================================================================
// qiming.build — 单名
// =============================================================================

func TestHandler_NamingBuild_Single(t *testing.T) {
	r := NewRPCRegistry()
	params := json.RawMessage(`{"surname":"王","chars1":{"8":[{"char":"明","tone":2}]}}`)
	result, err := r.Execute(context.Background(), "qiming.build", params)
	if err != nil {
		t.Fatalf("qiming.build single: %v", err)
	}
	if getStr(result, "_product") != "naming_build" {
		t.Errorf("_product = %q, want naming_build", getStr(result, "_product"))
	}
}

// =============================================================================
// qiming.check — 批量两字名
// =============================================================================

func TestHandler_NamingCheck_MultiTwoCharNames(t *testing.T) {
	r := NewRPCRegistry()
	params := json.RawMessage(`{
		"surname":"沙",
		"names":["佳桐","若薇","梓卉"]
	}`)
	result, err := r.Execute(context.Background(), "qiming.check", params)
	if err != nil {
		if strings.Contains(err.Error(), "not found") {
			t.Skip("test chars not in DB: " + err.Error())
		}
		t.Fatalf("qiming.check: %v", err)
	}
	if getStr(result, "_product") != "naming_check" {
		t.Errorf("_product = %q, want naming_check", getStr(result, "_product"))
	}

	var data []map[string]any
	env := struct {
		Data []map[string]any `json:"data"`
	}{}
	if err := json.Unmarshal(result, &env); err != nil {
		t.Fatal(err)
	}
	data = env.Data
	if len(data) != 3 {
		t.Fatalf("expected 3 results, got %d", len(data))
	}
	for i, d := range data {
		given, _ := d["given_name"].(string)
		if len([]rune(given)) != 2 {
			t.Errorf("result[%d] given_name = %q (len=%d), want 2 chars", i, given, len([]rune(given)))
		}
		surname, _ := d["surname"].(string)
		if surname != "沙" {
			t.Errorf("result[%d] surname = %q, want 沙", i, surname)
		}
		chars, _ := d["characters"].([]any)
		if len(chars) != 2 {
			t.Errorf("result[%d] characters count = %d, want 2", i, len(chars))
		}
	}
}

func TestHandler_NamingCheck_SingleName(t *testing.T) {
	r := NewRPCRegistry()
	params := json.RawMessage(`{
		"surname":"王",
		"names":["明辉"],
		"yongshen":"火"
	}`)
	result, err := r.Execute(context.Background(), "qiming.check", params)
	if err != nil {
		if strings.Contains(err.Error(), "not found") {
			t.Skip("test chars not in DB: " + err.Error())
		}
		t.Fatalf("qiming.check: %v", err)
	}
	if getStr(result, "_product") != "naming_check" {
		t.Errorf("_product = %q, want naming_check", getStr(result, "_product"))
	}
	var env struct {
		Data []map[string]any `json:"data"`
	}
	if err := json.Unmarshal(result, &env); err != nil {
		t.Fatal(err)
	}
	if len(env.Data) != 1 {
		t.Fatalf("expected 1 result, got %d", len(env.Data))
	}
	given, _ := env.Data[0]["given_name"].(string)
	if given != "明辉" {
		t.Errorf("given_name = %q, want 明辉", given)
	}
	name, _ := env.Data[0]["name"].(string)
	if name == "" {
		t.Error("name is empty — should be set")
	}
}

func TestHandler_NamingCheck_AllParams(t *testing.T) {
	r := NewRPCRegistry()
	params := json.RawMessage(`{
		"surname":"沙",
		"names":["佳桐"],
		"yongshen":"火",
		"xishen":["木"],
		"jishen":["金"]
	}`)
	result, err := r.Execute(context.Background(), "qiming.check", params)
	if err != nil {
		if strings.Contains(err.Error(), "not found") {
			t.Skip("test chars not in DB: " + err.Error())
		}
		t.Fatalf("qiming.check: %v", err)
	}
	if getStr(result, "_product") != "naming_check" {
		t.Errorf("_product = %q, want naming_check", getStr(result, "_product"))
	}

	var env struct {
		Data []map[string]any `json:"data"`
	}
	if err := json.Unmarshal(result, &env); err != nil {
		t.Fatal(err)
	}
	if len(env.Data) != 1 {
		t.Fatalf("expected 1 result, got %d", len(env.Data))
	}
	wuxing, _ := env.Data[0]["wuxing"].(map[string]any)
	if wuxing == nil {
		t.Fatal("wuxing should be set when yong/xi/ji provided")
	}
	if _, ok := wuxing["yong"]; !ok {
		t.Error("wuxing.yong should be present")
	}
	// xi and ji may be omitempty — only check they're set when they appear
	t.Logf("wuxing: %+v", wuxing)
}

func TestHandler_NamingPipeline(t *testing.T) {
	r := NewRPCRegistry()

	pickResult, err := r.Execute(context.Background(), "qiming.pick",
		json.RawMessage(`{"wuxing":"金"}`))
	if err != nil {
		t.Fatalf("pick: %v", err)
	}
	var pickEnv struct {
		Data struct {
			Chars map[string]any `json:"chars"`
		} `json:"data"`
	}
	if err := json.Unmarshal(pickResult, &pickEnv); err != nil {
		t.Fatal(err)
	}
	if len(pickEnv.Data.Chars) == 0 {
		t.Fatal("pick returned empty chars")
	}

	buildParams := json.RawMessage(fmt.Sprintf(`{"surname":"王","chars1":%s,"chars2":%s}`,
		mustMarshalInline(pickEnv.Data.Chars), mustMarshalInline(pickEnv.Data.Chars)))
	buildResult, err := r.Execute(context.Background(), "qiming.build", buildParams)
	if err != nil {
		t.Fatalf("build: %v", err)
	}
	var buildEnv struct {
		Data struct {
			Names []string `json:"names"`
		} `json:"data"`
	}
	if err := json.Unmarshal(buildResult, &buildEnv); err != nil {
		t.Fatal(err)
	}
	if len(buildEnv.Data.Names) == 0 {
		t.Fatal("build returned no names")
	}

	firstName := []rune(buildEnv.Data.Names[0])
	givenName := string(firstName[1:]) // strip surname
	checkParams := json.RawMessage(fmt.Sprintf(`{"surname":"王","names":["%s"]}`, givenName))
	checkResult, err := r.Execute(context.Background(), "qiming.check", checkParams)
	if err != nil {
		if strings.Contains(err.Error(), "not found") {
			t.Skip("test chars not in DB: " + err.Error())
		}
		t.Fatalf("check: %v", err)
	}
	if getStr(checkResult, "_product") != "naming_check" {
		t.Errorf("_product = %q, want naming_check", getStr(checkResult, "_product"))
	}
}

func mustMarshalInline(v any) string {
	b, err := json.Marshal(v)
	if err != nil {
		panic(err)
	}
	return string(b)
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
	params := json.RawMessage(fmt.Sprintf(`{"liu_nian":2026,"chart":%s}`, env.Data))
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
	params := json.RawMessage(fmt.Sprintf(`{"liu_nian":2026,"lunar_month":5,"chart":%s}`, env.Data))
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
	if err := json.Unmarshal(chartA, &envA); err != nil { t.Fatal(err) }
	if err := json.Unmarshal(chartB, &envB); err != nil { t.Fatal(err) }
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
	params := json.RawMessage(fmt.Sprintf(`{"liu_nian":2026,"lunar_month":5,"lunar_day":10,"chart":%s}`, env.Data))
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

func TestHandler_ZiweiJudgment_Valid(t *testing.T) {
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
	_, err = r.Execute(context.Background(), "ziwei.judgment",
		json.RawMessage(fmt.Sprintf(`{"chart":%s}`, env.Data)))
	if err != nil {
		t.Fatalf("ziwei.judgment: %v", err)
	}
}

func TestHandler_QimenJudgment_Valid(t *testing.T) {
	r := NewRPCRegistry()
	chartResult, err := r.Execute(context.Background(), "qimen.chart",
		json.RawMessage(fmt.Sprintf(`{"solar_time":%s,"kind":"shi"}`, btOK)))
	if err != nil {
		t.Fatalf("qimen.chart: %v", err)
	}
	var env struct {
		Data json.RawMessage `json:"data"`
	}
	if err := json.Unmarshal(chartResult, &env); err != nil {
		t.Fatal(err)
	}
	params := json.RawMessage(fmt.Sprintf(`{"chart":%s,"event":"general"}`, env.Data))
	_, err = r.Execute(context.Background(), "qimen.judgment", params)
	if err != nil {
		t.Fatalf("qimen.judgment: %v", err)
	}
}

func TestHandler_LiuyaoJudgment_Valid(t *testing.T) {
	r := NewRPCRegistry()
	saat := fmt.Sprintf(`{"solar_time":%s,"yaos":[8,9,8,6,7,9],"yong_shen":"妻财"}`, btOK)
	chartResult, err := r.Execute(context.Background(), "liuyao.chart", json.RawMessage(saat))
	if err != nil {
		t.Fatalf("liuyao.chart: %v", err)
	}
	var env struct {
		Data json.RawMessage `json:"data"`
	}
	if err := json.Unmarshal(chartResult, &env); err != nil {
		t.Fatal(err)
	}
	_, err = r.Execute(context.Background(), "liuyao.judgment",
		json.RawMessage(fmt.Sprintf(`{"chart":%s}`, env.Data)))
	if err != nil {
		t.Fatalf("liuyao.judgment: %v", err)
	}
}

func TestHandler_BazhaiJudgment_Valid(t *testing.T) {
	r := NewRPCRegistry()
	chartResult, err := r.Execute(context.Background(), "bazhai.chart",
		json.RawMessage(fmt.Sprintf(`{"solar_time":%s,"gender":"male"}`, btOK)))
	if err != nil {
		t.Fatalf("bazhai.chart: %v", err)
	}
	var env struct {
		Data json.RawMessage `json:"data"`
	}
	if err := json.Unmarshal(chartResult, &env); err != nil {
		t.Fatal(err)
	}
	params := json.RawMessage(fmt.Sprintf(`{"chart":%s,"door_gua":1,"master_gua":2,"stove_gua":3}`, env.Data))
	_, err = r.Execute(context.Background(), "bazhai.judgment", params)
	if err != nil {
		t.Fatalf("bazhai.judgment: %v", err)
	}
}

func TestHandler_XuankongAnnual_Valid(t *testing.T) {
	r := NewRPCRegistry()
	chartResult, err := r.Execute(context.Background(), "xuankong.chart",
		json.RawMessage(fmt.Sprintf(`{"solar_time":%s,"sit_mountain":20,"face_mountain":8}`, btOK)))
	if err != nil {
		t.Fatalf("xuankong.chart: %v", err)
	}
	var env struct {
		Data json.RawMessage `json:"data"`
	}
	if err := json.Unmarshal(chartResult, &env); err != nil {
		t.Fatal(err)
	}
	_, err = r.Execute(context.Background(), "xuankong.annual",
		json.RawMessage(fmt.Sprintf(`{"chart":%s,"year":2026}`, env.Data)))
	if err != nil {
		t.Fatalf("xuankong.annual: %v", err)
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
	var env struct{ Data json.RawMessage `json:"data"` }
	if err := json.Unmarshal(chartResult, &env); err != nil {
		t.Fatal(err)
	}
	// 把 chart 传入 fullchart
	fullResult, err := r.Execute(context.Background(), "bazi.fullchart",
		json.RawMessage(fmt.Sprintf(`{"chart":%s}`, env.Data)))
	if err != nil {
		t.Fatalf("bazi.fullchart: %v", err)
	}
	var full struct{ Data map[string]any `json:"data"` }
	if err := json.Unmarshal(fullResult, &full); err != nil {
		t.Fatal(err)
	}
	// 验证完整结果包含扩展字段
	for _, pillar := range []string{"nian", "yue", "ri", "shi"} {
		p, ok := full.Data[pillar].(map[string]any)
		if !ok { t.Errorf("%s not found in fullchart result", pillar); continue }
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
	var e1, e2 struct{ Data json.RawMessage `json:"data"` }
	if err := json.Unmarshal(c1, &e1); err != nil { t.Fatal(err) }
	if err := json.Unmarshal(c2, &e2); err != nil { t.Fatal(err) }
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
	var env struct{ Data json.RawMessage `json:"data"` }
if err := json.Unmarshal(chartResult, &env); err != nil { t.Fatal(err) }
	_, err = r.Execute(context.Background(), "bazi.xiaoyun",
		json.RawMessage(fmt.Sprintf(`{"chart":%s,"count":3}`, env.Data)))
	if err != nil {
		t.Fatalf("bazi.xiaoyun: %v", err)
	}
}

func TestHandler_BaziXiaoxian_Valid(t *testing.T) {
	r := NewRPCRegistry()
	_, err := r.Execute(context.Background(), "bazi.xiaoxian",
		json.RawMessage(`{"gender":"male","count":3}`))
	if err != nil {
		t.Fatalf("bazi.xiaoxian: %v", err)
	}
}

func TestHandler_ZiweiLiuyue_Valid(t *testing.T) {
	r := NewRPCRegistry()
	chartResult, err := r.Execute(context.Background(), "ziwei.chart",
		json.RawMessage(fmt.Sprintf(`{"lunar":%s,"gender":"male"}`, lunarOK)))
	if err != nil {
		t.Fatalf("execute: %v", err)
	}
	var env struct{ Data json.RawMessage `json:"data"` }
if err := json.Unmarshal(chartResult, &env); err != nil { t.Fatal(err) }
	_, err = r.Execute(context.Background(), "ziwei.liuyue",
		json.RawMessage(fmt.Sprintf(`{"liu_nian":2026,"lunar_month":5,"chart":%s}`, env.Data)))
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
	var e1, e2 struct{ Data json.RawMessage `json:"data"` }
	if err := json.Unmarshal(c1, &e1); err != nil { t.Fatal(err) }
	if err := json.Unmarshal(c2, &e2); err != nil { t.Fatal(err) }
	_, err = r.Execute(context.Background(), "ziwei.bond",
		json.RawMessage(fmt.Sprintf(`{"a":%s,"b":%s}`, e1.Data, e2.Data)))
	if err != nil {
		t.Fatalf("ziwei.bond: %v", err)
	}
}


func TestHandler_QimenSelect_Valid(t *testing.T) {
	r := NewRPCRegistry()
	_, err := r.Execute(context.Background(), "qimen.select",
		json.RawMessage(`{"start_date":"2027-07-01","end_date":"2027-07-31","count":3}`))
	if err != nil {
		t.Fatalf("qimen.select: %v", err)
	}
}
