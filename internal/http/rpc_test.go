package http

import (
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/http/httptest"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/santhosh-tekuri/jsonschema/v6"

	"liki-engine/internal/agent"
	"liki-engine/internal/agent/city"
)

// ── Protocol error handling ──────────────────────────────────

func TestRPC_Discover(t *testing.T) {
	reg := agent.NewRPCRegistry()
	body := `{"jsonrpc":"2.0","method":"rpc.discover","id":1}`
	r := httptest.NewRequest("POST", "/jsonrpc", strings.NewReader(body))
	w := httptest.NewRecorder()

	HandleRPC(reg)(w, r)

	if w.Code != http.StatusOK {
		t.Fatalf("status = %d", w.Code)
	}

	var resp rpcResponse
	if err := json.NewDecoder(w.Body).Decode(&resp); err != nil {
		t.Fatalf("decode: %v", err)
	}
	if resp.Error != nil {
		t.Fatalf("rpc.discover error: %+v", resp.Error)
	}
	if resp.JSONRPC != "2.0" {
		t.Errorf("jsonrpc = %q, want 2.0", resp.JSONRPC)
	}
	if resp.ID != float64(1) {
		t.Errorf("id = %v, want 1", resp.ID)
	}

	doc := resp.Result.(map[string]any)
	if doc["openrpc"] != "1.4.1" {
		t.Errorf("openrpc = %v, want 1.4.1", doc["openrpc"])
	}

	methods := doc["methods"].([]any)
	if len(methods) < 29 {
		t.Errorf("methods count = %d, want >= 29", len(methods))
	}

	first := methods[0].(map[string]any)
	if first["name"] != "rpc.discover" {
		t.Errorf("first method = %v, want rpc.discover", first["name"])
	}
}

func TestRPC_ProtocolErrors(t *testing.T) {
	tests := []struct {
		name     string
		body     string
		wantCode int
	}{
		{"unknown method", `{"jsonrpc":"2.0","method":"nonexistent","params":{},"id":1}`, -32601},
		{"invalid jsonrpc version", `{"jsonrpc":"1.0","method":"bazi.chart","id":1}`, -32600},
		{"parse error", `{bad`, -32700},
		{"missing id", `{"jsonrpc":"2.0","method":"bazi.chart"}`, -32600},
		{"positional params", `{"jsonrpc":"2.0","method":"bazi.chart","params":[],"id":1}`, -32600},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			reg := agent.NewRPCRegistry()
			r := httptest.NewRequest("POST", "/jsonrpc", strings.NewReader(tt.body))
			w := httptest.NewRecorder()
			HandleRPC(reg)(w, r)

			var resp rpcResponse
			if err := json.NewDecoder(w.Body).Decode(&resp); err != nil {
				t.Fatalf("decode: %v", err)
			}
			if resp.Error == nil || resp.Error.Code != tt.wantCode {
				t.Errorf("error = %+v, want code %d", resp.Error, tt.wantCode)
			}
		})
	}
}

func TestRPC_NullParams(t *testing.T) {
	// null params should be treated as {} (JSON-RPC 2.0 spec).
	reg := agent.NewRPCRegistry()
	body := `{"jsonrpc":"2.0","method":"rpc.discover","params":null,"id":1}`
	r := httptest.NewRequest("POST", "/jsonrpc", strings.NewReader(body))
	w := httptest.NewRecorder()

	HandleRPC(reg)(w, r)

	var resp rpcResponse
	if err := json.NewDecoder(w.Body).Decode(&resp); err != nil {
		t.Fatalf("decode: %v", err)
	}
	if resp.Error != nil {
		t.Fatalf("null params should be accepted: %+v", resp.Error)
	}
}

func TestRPC_ResponseEnvelope(t *testing.T) {
	// All responses must include "jsonrpc":"2.0" and echo the request id.
	reg := agent.NewRPCRegistry()
	body := `{"jsonrpc":"2.0","method":"rpc.discover","id":"req-42"}`
	r := httptest.NewRequest("POST", "/jsonrpc", strings.NewReader(body))
	w := httptest.NewRecorder()

	HandleRPC(reg)(w, r)

	var resp rpcResponse
	if err := json.NewDecoder(w.Body).Decode(&resp); err != nil {
		t.Fatalf("decode: %v", err)
	}
	if resp.JSONRPC != "2.0" {
		t.Errorf("jsonrpc = %q, want 2.0", resp.JSONRPC)
	}
	if resp.ID != "req-42" {
		t.Errorf("id = %v, want req-42", resp.ID)
	}
}

func TestRPC_MissingRequiredParams(t *testing.T) {
	// Missing required params should return -32602 (invalid params per JSON-RPC spec).
	reg := agent.NewRPCRegistry()
	body := `{"jsonrpc":"2.0","method":"bazi.chart","params":{},"id":1}`
	r := httptest.NewRequest("POST", "/jsonrpc", strings.NewReader(body))
	w := httptest.NewRecorder()

	HandleRPC(reg)(w, r)

	var resp rpcResponse
	if err := json.NewDecoder(w.Body).Decode(&resp); err != nil {
		t.Fatalf("decode: %v", err)
	}
	if resp.Error == nil {
		t.Fatal("expected error for missing required params")
	}
	if resp.Error.Code != -32602 {
		t.Errorf("error code = %d, want -32602", resp.Error.Code)
	}
}

// ── Infrastructure ───────────────────────────────────────────

func TestRPC_CORSHeader(t *testing.T) {
	reg := agent.NewRPCRegistry()
	body := `{"jsonrpc":"2.0","method":"rpc.discover","id":1}`
	r := httptest.NewRequest("POST", "/jsonrpc", strings.NewReader(body))
	w := httptest.NewRecorder()

	HandleRPC(reg)(w, r)

	if w.Header().Get("Access-Control-Allow-Origin") != "*" {
		t.Error("missing CORS header")
	}
}

func TestRPC_ContentType(t *testing.T) {
	reg := agent.NewRPCRegistry()
	body := `{"jsonrpc":"2.0","method":"rpc.discover","id":1}`
	r := httptest.NewRequest("POST", "/jsonrpc", strings.NewReader(body))
	w := httptest.NewRecorder()

	HandleRPC(reg)(w, r)

	if ct := w.Header().Get("Content-Type"); ct != "application/json; charset=utf-8" {
		t.Errorf("Content-Type = %q", ct)
	}
}

func TestRPC_OptionsPreflight(t *testing.T) {
	reg := agent.NewRPCRegistry()
	r := httptest.NewRequest("OPTIONS", "/jsonrpc", nil)
	w := httptest.NewRecorder()

	HandleRPC(reg)(w, r)

	if w.Code != http.StatusNoContent {
		t.Errorf("status = %d, want 204", w.Code)
	}
	if w.Header().Get("Access-Control-Allow-Origin") != "*" {
		t.Error("missing CORS header on OPTIONS")
	}
}

// ── Discovery ────────────────────────────────────────────────

func TestRPC_DiscoverContainsAllMethods(t *testing.T) {
	reg := agent.NewRPCRegistry()
	doc := reg.OpenRPCDocument()

	var parsed map[string]any
	if err := json.Unmarshal(doc, &parsed); err != nil {
		t.Fatal(err)
	}

	methods := parsed["methods"].([]any)
	names := make(map[string]bool)
	for _, m := range methods {
		names[m.(map[string]any)["name"].(string)] = true
	}

	expected := []string{
		"rpc.discover",
		"bazi.fullchart", "bazi.chart", "bazi.yongshen", "bazi.bond", "bazi.liunian", "bazi.liuyue", "bazi.liuri", "bazi.liushi", "bazi.xiaoyun", "bazi.xiaoxian",
		"ziwei.chart", "ziwei.fullchart", "ziwei.daxian", "ziwei.liunian", "ziwei.liuyue", "ziwei.liuri", "ziwei.liushi", "ziwei.judgment", "ziwei.bond",
		"qimen.chart","qimen.judgment","qimen.select",
		"qiming.char", "qiming.pick", "qiming.build", "qiming.check",
		"bazhai.judgment", "bazhai.chart", "bazhai.minggua",
		"xuankong.annual", "xuankong.sanyuan", "xuankong.chart",
		"liuyao.judgment", "liuyao.qigua", "liuyao.chart",
		"huangli.days",
		"city", "time.now", "tianwen.time",
	}
	for _, name := range expected {
		if !names[name] {
			t.Errorf("missing method: %s", name)
		}
	}
	if len(methods) != len(expected) {
		t.Errorf("method count = %d, want %d", len(methods), len(expected))
	}
}

func TestRPC_DiscoverAllMethodsHaveResultSchema(t *testing.T) {
	reg := agent.NewRPCRegistry()
	doc := reg.OpenRPCDocument()

	var parsed map[string]any
	if err := json.Unmarshal(doc, &parsed); err != nil {
		t.Fatal(err)
	}

	for _, m := range parsed["methods"].([]any) {
		mm := m.(map[string]any)
		if mm["result"] == nil {
			t.Errorf("missing result schema: %s", mm["name"])
		}
	}
}

// ── Data-driven dispatch tests ───────────────────────────────

func TestRPC_DataDriven(t *testing.T) {
	origSearch := city.HttpClient()
	city.SetHTTPClient(&http.Client{
		Transport: &mockSearchTransport{
			body: `[{"lat":"39.9042","lon":"116.4074","name":"Beijing","address":{"country":"China","country_code":"CN"}}]`,
		},
	})
	defer func() { city.SetHTTPClient(origSearch) }()

	reg := agent.NewRPCRegistry()
	doc := reg.OpenRPCDocument()

	var parsed map[string]any
	if err := json.Unmarshal(doc, &parsed); err != nil {
		t.Fatal(err)
	}

	params := loadParams(t)

	for _, m := range parsed["methods"].([]any) {
		method := m.(map[string]any)
		name := method["name"].(string)

		if name == "rpc.discover" {
			continue
		}

		fixture, ok := params[name]
		if !ok {
			t.Logf("skip %s: no fixture (inter-dependent or external API)", name)
			continue
		}

		resultSchema, hasSchema := method["result"].(map[string]any)
		if !hasSchema {
			t.Logf("skip %s: no result schema", name)
			continue
		}

		t.Run(name, func(t *testing.T) {
			retries := 1
			var resp rpcResponse
			for attempt := 0; attempt < retries; attempt++ {
				reqBody := fmt.Sprintf(`{"jsonrpc":"2.0","method":"%s","params":%s,"id":1}`, name, mustMarshal(fixture))
				req := httptest.NewRequest("POST", "/jsonrpc", strings.NewReader(reqBody))
				req.Header.Set("Content-Type", "application/json")
				w := httptest.NewRecorder()
				HandleRPC(reg)(w, req)

				if w.Code != http.StatusOK {
					t.Fatalf("status = %d", w.Code)
				}

				if err := json.NewDecoder(w.Body).Decode(&resp); err != nil {
					t.Fatalf("decode: %v", err)
				}
				if resp.Error == nil {
					break
				}
				if attempt == retries-1 {
					t.Fatalf("RPC error: %+v", resp.Error)
				}
			}
			if resp.JSONRPC != "2.0" {
				t.Errorf("jsonrpc = %q, want 2.0", resp.JSONRPC)
			}

			// Validate response against result schema.
			c := jsonschema.NewCompiler()
			if err := c.AddResource("schema.json", resultSchema); err != nil {
				t.Fatalf("add schema resource: %v", err)
			}
			sch, err := c.Compile("schema.json")
			if err != nil {
				t.Fatalf("compile result schema for %s: %v", name, err)
			}
			if err := sch.Validate(resp.Result); err != nil {
				t.Errorf("result validation: %v", err)
			}
		})
	}
}

// ── Inter-dependent method tests ─────────────────────────────

var ziweiBirth = `{"lunar":{"year":2000,"month":6,"day":15,"shichen":"午"},"gender":"male"}`

func getZiweiChart(t *testing.T, reg *agent.RPCRegistry) map[string]any {
	t.Helper()
	body := `{"jsonrpc":"2.0","method":"ziwei.chart","params":` + ziweiBirth + `,"id":99}`
	var chart map[string]any
	postRPC(t, reg, body, func(resp rpcResponse) {
		chart = resp.Result.(map[string]any)["data"].(map[string]any)
	})
	return chart
}

func getNamingFixture(t *testing.T, reg *agent.RPCRegistry) (chars1 any) {
	t.Helper()
	body := `{"jsonrpc":"2.0","method":"qiming.pick","params":{"surname":"王","wuxing1":"金","count":1},"id":99}`
	postRPC(t, reg, body, func(resp rpcResponse) {
		chars1 = resp.Result.(map[string]any)["data"].(map[string]any)["chars"]
	})
	return
}

func validateSchema(t *testing.T, method string, result any) {
	t.Helper()
	reg := agent.NewRPCRegistry()
	doc := reg.OpenRPCDocument()
	var parsed map[string]any
	if err := json.Unmarshal(doc, &parsed); err != nil {
		t.Fatalf("unmarshal OpenRPC doc: %v", err)
	}
	for _, m := range parsed["methods"].([]any) {
		mm := m.(map[string]any)
		if mm["name"].(string) == method {
			c := jsonschema.NewCompiler()
			if err := c.AddResource("schema.json", mm["result"]); err != nil {
			t.Fatalf("add resource for %s: %v", method, err)
		}
			sch, err := c.Compile("schema.json")
			if err != nil {
				t.Fatalf("compile result schema for %s: %v", method, err)
			}
			if err := sch.Validate(result); err != nil {
				t.Errorf("result validation: %v", err)
			}
			return
		}
	}
	t.Fatalf("method %s not found in discover", method)
}

func TestRPC_Dispatch_ZiweiDaXian(t *testing.T) {
	reg := agent.NewRPCRegistry()
	chart := getZiweiChart(t, reg)

	params := map[string]any{"chart": chart}
	body := fmt.Sprintf(`{"jsonrpc":"2.0","method":"ziwei.daxian","params":%s,"id":1}`, mustMarshal(params))
	postRPC(t, reg, body, func(resp rpcResponse) {
		assertEnvelope(t, resp, "ziwei_daxian")
		steps := resp.Result.(map[string]any)["data"].([]any)
		if len(steps) == 0 {
			t.Fatal("empty daxian steps")
		}
		first := steps[0].(map[string]any)
		assertNonNil(t, first, "qi_sui", "zhi_sui", "name")
		validateSchema(t, "ziwei.daxian", resp.Result)
	})
}

func TestRPC_Dispatch_ZiweiLiuNian(t *testing.T) {
	reg := agent.NewRPCRegistry()
	chart := getZiweiChart(t, reg)

	params := map[string]any{"liu_nian": 2026, "chart": chart}
	body := fmt.Sprintf(`{"jsonrpc":"2.0","method":"ziwei.liunian","params":%s,"id":1}`, mustMarshal(params))
	postRPC(t, reg, body, func(resp rpcResponse) {
		assertEnvelope(t, resp, "ziwei_liunian")
		data := resp.Result.(map[string]any)["data"].(map[string]any)
		assertNonNil(t, data, "ming_gong", "si_hua")
		validateSchema(t, "ziwei.liunian", resp.Result)
	})
}

func TestRPC_Dispatch_ZiweiLiuYue(t *testing.T) {
	reg := agent.NewRPCRegistry()
	chart := getZiweiChart(t, reg)

	params := map[string]any{"liu_nian": 2026, "lunar_month": 1, "chart": chart}
	body := fmt.Sprintf(`{"jsonrpc":"2.0","method":"ziwei.liuyue","params":%s,"id":1}`, mustMarshal(params))
	postRPC(t, reg, body, func(resp rpcResponse) {
		assertEnvelope(t, resp, "ziwei_liuyue")
		data := resp.Result.(map[string]any)["data"].(map[string]any)
		assertNonNil(t, data, "ming_gong", "si_hua")
		validateSchema(t, "ziwei.liuyue", resp.Result)
	})
}

func TestRPC_Dispatch_ZiweiLiuRi(t *testing.T) {
	reg := agent.NewRPCRegistry()
	chart := getZiweiChart(t, reg)

	params := map[string]any{"liu_nian": 2026, "lunar_month": 1, "lunar_day": 1, "chart": chart}
	body := fmt.Sprintf(`{"jsonrpc":"2.0","method":"ziwei.liuri","params":%s,"id":1}`, mustMarshal(params))
	postRPC(t, reg, body, func(resp rpcResponse) {
		assertEnvelope(t, resp, "ziwei_liuri")
		data := resp.Result.(map[string]any)["data"].(map[string]any)
		assertNonNil(t, data, "ming_gong", "si_hua")
		validateSchema(t, "ziwei.liuri", resp.Result)
	})
}

func TestRPC_Dispatch_ZiweiBond(t *testing.T) {
	reg := agent.NewRPCRegistry()

	chartA := getZiweiChart(t, reg)
	// second person: 1986-08-20, female
	chartBBody := `{"jsonrpc":"2.0","method":"ziwei.chart","params":{"lunar":{"year":2000,"month":8,"day":23,"shichen":"辰"},"gender":"female"},"id":99}`
	var chartB map[string]any
	postRPC(t, reg, chartBBody, func(resp rpcResponse) {
		chartB = resp.Result.(map[string]any)["data"].(map[string]any)
	})

	params := map[string]any{"a": chartA, "b": chartB}
	body := fmt.Sprintf(`{"jsonrpc":"2.0","method":"ziwei.bond","params":%s,"id":1}`, mustMarshal(params))
	postRPC(t, reg, body, func(resp rpcResponse) {
		assertEnvelope(t, resp, "ziwei_bond")
		data := resp.Result.(map[string]any)["data"].(map[string]any)
		assertNonNil(t, data, "ming_gong_hu_ru", "ji_xing", "sha_xing")
		validateSchema(t, "ziwei.bond", resp.Result)
	})
}



// ── helpers ──────────────────────────────────────────────────

func loadParams(t *testing.T) map[string]any {
	t.Helper()
	path := filepath.Join("testdata", "rpc", "params.json")
	data, err := os.ReadFile(path)
	if err != nil {
		t.Fatalf("read params fixture: %v", err)
	}
	var params map[string]any
	if err := json.Unmarshal(data, &params); err != nil {
		t.Fatalf("unmarshal params fixture: %v", err)
	}
	return params
}

func mustMarshal(v any) string {
	b, err := json.Marshal(v)
	if err != nil {
		panic(err)
	}
	return string(b)
}

func postRPC(t *testing.T, reg *agent.RPCRegistry, body string, fn func(rpcResponse)) {
	t.Helper()
	r := httptest.NewRequest("POST", "/jsonrpc", strings.NewReader(body))
	r.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()
	HandleRPC(reg)(w, r)

	if w.Code != http.StatusOK {
		t.Fatalf("status = %d", w.Code)
	}
	var resp rpcResponse
	if err := json.NewDecoder(w.Body).Decode(&resp); err != nil {
		t.Fatalf("decode: %v", err)
	}
	if resp.Error != nil {
		t.Fatalf("%+v", resp.Error)
	}
	fn(resp)
}

// assertEnvelope verifies the JSON-RPC envelope: jsonrpc field, id echo, _product, and that data exists.
func assertEnvelope(t *testing.T, resp rpcResponse, wantProduct string) {
	t.Helper()
	if resp.JSONRPC != "2.0" {
		t.Errorf("jsonrpc = %q, want 2.0", resp.JSONRPC)
	}
	if resp.ID == nil {
		t.Error("id is nil")
	}

	m, ok := resp.Result.(map[string]any)
	if !ok {
		t.Fatalf("result is %T, want map[string]any", resp.Result)
	}
	product, _ := m["_product"].(string)
	if product != wantProduct {
		t.Errorf("_product = %q, want %q", product, wantProduct)
	}
	if m["data"] == nil {
		t.Fatal("data is nil")
	}
}

// assertNonNil checks that all named keys exist and have non-nil values.
func assertNonNil(t *testing.T, m map[string]any, keys ...string) {
	t.Helper()
	for _, k := range keys {
		v, ok := m[k]
		if !ok {
			t.Errorf("missing key %q", k)
			continue
		}
		if v == nil {
			t.Errorf("key %q is nil", k)
		}
	}
}

type mockSearchTransport struct {
	body string
}

func (m *mockSearchTransport) RoundTrip(_ *http.Request) (*http.Response, error) {
	return &http.Response{
		StatusCode: http.StatusOK,
		Body:       io.NopCloser(strings.NewReader(m.body)),
		Header:     http.Header{"Content-Type": []string{"application/json"}},
	}, nil
}

func TestRPC_Dispatch_BaziChart(t *testing.T) {
	reg := agent.NewRPCRegistry()
	body := `{"jsonrpc":"2.0","method":"bazi.chart","params":{"solar_time":"1984-02-15T08:00:00+08:00","gender":"male"},"id":99}`
	postRPC(t, reg, body, func(resp rpcResponse) {
		assertEnvelope(t, resp, "chart")
		data := resp.Result.(map[string]any)["data"].(map[string]any)
		// 四柱
		for _, k := range []string{"nian", "yue", "ri", "shi"} {
			assertNonNil(t, data, k)
		}
		// 大运新字段（起运精确年/月/日）
		dy := data["da_yun"].(map[string]any)
		for _, k := range []string{"start_age", "start_year_after", "start_month_after", "start_day_after", "direction", "steps", "current_step_index"} {
			assertNonNil(t, dy, k)
		}
		// 精确起运值（1984-02-15 男 = lunar 6年5月20日）
		if dy["start_year_after"].(float64) != 6 || dy["start_month_after"].(float64) != 5 || dy["start_day_after"].(float64) != 20 {
			t.Errorf("起运精确值 = %v年%v月%v日, want 6年5月20日",
				dy["start_year_after"], dy["start_month_after"], dy["start_day_after"])
		}
		validateSchema(t, "bazi.chart", resp.Result)
	})
}

func TestRPC_Dispatch_BaziFullChart(t *testing.T) {
	reg := agent.NewRPCRegistry()
	// 先 chart 再 fullchart
	var chart map[string]any
	postRPC(t, reg, `{"jsonrpc":"2.0","method":"bazi.chart","params":{"solar_time":"1984-02-15T08:00:00+08:00","gender":"male"},"id":1}`, func(resp rpcResponse) {
		chart = resp.Result.(map[string]any)["data"].(map[string]any)
	})
	params := map[string]any{"chart": chart}
	body := fmt.Sprintf(`{"jsonrpc":"2.0","method":"bazi.fullchart","params":%s,"id":2}`, mustMarshal(params))
	postRPC(t, reg, body, func(resp rpcResponse) {
		assertEnvelope(t, resp, "bazi_fullchart")
		data := resp.Result.(map[string]any)["data"].(map[string]any)
		// 三元 + 合冲 + 大运
		// 结构性字段必有值
		for _, k := range []string{"san_yuan", "gong_jia", "nayin_rel", "chang_sheng", "san_qi_name"} {
			assertNonNil(t, data, k)
		}
		// 合冲刑字段可能为空（如命例无三合/六冲）——检查 key 存在即可
		for _, k := range []string{"gan_he", "zhi_liu_he", "san_he", "san_hui", "liu_chong", "liu_hai", "liu_xing"} {
			if _, ok := data[k]; !ok {
				t.Errorf("missing key %q", k)
			}
		}
		// 四柱含藏干/十神/神煞
		ri := data["ri"].(map[string]any)
		for _, k := range []string{"gan", "zhi", "na_yin", "cang_gan", "shi_shens", "chang_sheng", "shen_sha"} {
			assertNonNil(t, ri, k)
		}
		validateSchema(t, "bazi.fullchart", resp.Result)
	})
}

func getBaziChart(t *testing.T, reg *agent.RPCRegistry) map[string]any {
	t.Helper()
	var chart map[string]any
	postRPC(t, reg, `{"jsonrpc":"2.0","method":"bazi.chart","params":{"solar_time":"1984-02-15T08:00:00+08:00","gender":"male"},"id":1}`, func(resp rpcResponse) {
		chart = resp.Result.(map[string]any)["data"].(map[string]any)
	})
	return chart
}

func TestRPC_Dispatch_BaziYongShen(t *testing.T) {
	reg := agent.NewRPCRegistry()
	chart := getBaziChart(t, reg)
	params := map[string]any{"chart": chart}
	body := fmt.Sprintf(`{"jsonrpc":"2.0","method":"bazi.yongshen","params":%s,"id":2}`, mustMarshal(params))
	postRPC(t, reg, body, func(resp rpcResponse) {
		assertEnvelope(t, resp, "yongshen")
		data := resp.Result.(map[string]any)["data"].(map[string]any)
		// 三派用神
		for _, k := range []string{"fu_yi", "tiao_hou", "ge_ju"} {
			assertNonNil(t, data, k)
		}
		validateSchema(t, "bazi.yongshen", resp.Result)
	})
}

func TestRPC_Dispatch_BaziBond(t *testing.T) {
	reg := agent.NewRPCRegistry()
	chart := getBaziChart(t, reg)
	params := map[string]any{"a": map[string]any{"chart": chart}, "b": map[string]any{"chart": chart}}
	body := fmt.Sprintf(`{"jsonrpc":"2.0","method":"bazi.bond","params":%s,"id":2}`, mustMarshal(params))
	postRPC(t, reg, body, func(resp rpcResponse) {
		assertEnvelope(t, resp, "bond")
		data := resp.Result.(map[string]any)["data"].(map[string]any)
		for _, k := range []string{"zhu_cross", "shi_shen_cross", "structure"} {
			assertNonNil(t, data, k)
		}
		validateSchema(t, "bazi.bond", resp.Result)
	})
}

func TestRPC_Dispatch_BaziLiuNian(t *testing.T) {
	reg := agent.NewRPCRegistry()
	chart := getBaziChart(t, reg)
	params := map[string]any{"chart": chart, "year": 2026}
	body := fmt.Sprintf(`{"jsonrpc":"2.0","method":"bazi.liunian","params":%s,"id":2}`, mustMarshal(params))
	postRPC(t, reg, body, func(resp rpcResponse) {
		assertEnvelope(t, resp, "liunian")
		data := resp.Result.(map[string]any)["data"].(map[string]any)
		for _, k := range []string{"year", "nian_gan", "nian_zhi", "shi_shen"} {
			assertNonNil(t, data, k)
		}
		validateSchema(t, "bazi.liunian", resp.Result)
	})
}

func TestRPC_Dispatch_BaziLiuYue(t *testing.T) {
	reg := agent.NewRPCRegistry()
	chart := getBaziChart(t, reg)
	params := map[string]any{"chart": chart, "year": 2026, "month": 6}
	body := fmt.Sprintf(`{"jsonrpc":"2.0","method":"bazi.liuyue","params":%s,"id":2}`, mustMarshal(params))
	postRPC(t, reg, body, func(resp rpcResponse) {
		assertEnvelope(t, resp, "liuyue")
		data := resp.Result.(map[string]any)["data"].(map[string]any)
		for _, k := range []string{"year", "month", "yue_gan", "yue_zhi", "month_name", "wuxing", "shi_shen"} {
			assertNonNil(t, data, k)
		}
		validateSchema(t, "bazi.liuyue", resp.Result)
	})
}

func TestRPC_Dispatch_BaziLiuRi(t *testing.T) {
	reg := agent.NewRPCRegistry()
	chart := getBaziChart(t, reg)
	params := map[string]any{"chart": chart, "year": 2026, "month": 6, "day": 4}
	body := fmt.Sprintf(`{"jsonrpc":"2.0","method":"bazi.liuri","params":%s,"id":2}`, mustMarshal(params))
	postRPC(t, reg, body, func(resp rpcResponse) {
		assertEnvelope(t, resp, "liuri")
		data := resp.Result.(map[string]any)["data"].(map[string]any)
		for _, k := range []string{"date", "ri_gan", "ri_zhi", "day_name", "day_nayin", "shi_shen"} {
			assertNonNil(t, data, k)
		}
		validateSchema(t, "bazi.liuri", resp.Result)
	})
}

func TestRPC_Dispatch_BaziLiuShi(t *testing.T) {
	reg := agent.NewRPCRegistry()
	chart := getBaziChart(t, reg)
	params := map[string]any{"chart": chart, "year": 2026, "month": 6, "day": 4, "hour": 12}
	body := fmt.Sprintf(`{"jsonrpc":"2.0","method":"bazi.liushi","params":%s,"id":2}`, mustMarshal(params))
	postRPC(t, reg, body, func(resp rpcResponse) {
		assertEnvelope(t, resp, "liushi")
		data := resp.Result.(map[string]any)["data"].(map[string]any)
		for _, k := range []string{"time", "shi_gan", "shi_zhi", "hour_name", "shi_shen", "gan_rels", "zhi_rels"} {
			assertNonNil(t, data, k)
		}
		validateSchema(t, "bazi.liushi", resp.Result)
	})
}

func TestRPC_Dispatch_BaziXiaoYun(t *testing.T) {
	reg := agent.NewRPCRegistry()
	chart := getBaziChart(t, reg)
	params := map[string]any{"chart": chart, "max_age": 12}
	body := fmt.Sprintf(`{"jsonrpc":"2.0","method":"bazi.xiaoyun","params":%s,"id":2}`, mustMarshal(params))
	postRPC(t, reg, body, func(resp rpcResponse) {
		assertEnvelope(t, resp, "xiaoyun")
		data := resp.Result.(map[string]any)["data"].([]any)
		if len(data) == 0 {
			t.Fatal("empty xiaoyun")
		}
		first := data[0].(map[string]any)
		assertNonNil(t, first, "age", "gan", "zhi", "name", "shi_shen")
		validateSchema(t, "bazi.xiaoyun", resp.Result)
	})
}
