package city

import (
	"context"
	"encoding/json"
	"io"
	"net/http"
	"strings"
	"testing"
)

func TestParseFloat(t *testing.T) {
	tests := []struct {
		input   string
		want    float64
		wantErr bool
	}{
		{"37.7749", 37.7749, false},
		{"0", 0, false},
		{"-122.4194", -122.4194, false},
		{"", 0, true},
		{"abc", 0, true},
		{"3.14", 3.14, false},
	}
	for _, tc := range tests {
		got, err := parseFloat(tc.input)
		if tc.wantErr && err == nil {
			t.Errorf("parseFloat(%q): want error, got nil", tc.input)
		}
		if !tc.wantErr && err != nil {
			t.Errorf("parseFloat(%q): %v", tc.input, err)
		}
		if got != tc.want {
			t.Errorf("parseFloat(%q)=%f, want %f", tc.input, got, tc.want)
		}
	}
}

func TestSearchCoords_Valid(t *testing.T) {
	orig := httpClient
	SetHTTPClient(&http.Client{
		Transport: &mockSearchTransport{
			body: `[{"lat":"39.9042","lon":"116.4074","name":"Beijing","address":{"country":"China","country_code":"CN"}}]`,
		},
	})
	defer func() { SetHTTPClient(orig) }()

	args := json.RawMessage(`{"city":"Beijing"}`)
	result, err := SearchCoords(context.Background(), args)
	if err != nil {
		t.Fatalf("SearchCoords: %v", err)
	}
	var r searchResult
	if err := json.Unmarshal(result, &r); err != nil {
		t.Fatalf("unmarshal result: %v", err)
	}
	if r.Name != "Beijing" {
		t.Errorf("name = %q, want Beijing", r.Name)
	}
	if r.Longitude != 116.4074 {
		t.Errorf("longitude = %f, want 116.4074", r.Longitude)
	}
	if r.Latitude != 39.9042 {
		t.Errorf("latitude = %f, want 39.9042", r.Latitude)
	}
	if r.Country != "中国" {
		t.Errorf("country = %q, want 中国", r.Country)
	}
}

func TestSearchCoords_EmptyCityName(t *testing.T) {
	args := json.RawMessage(`{"city":""}`)
	_, err := SearchCoords(context.Background(), args)
	if err == nil {
		t.Fatal("expected error for empty city")
	}
	if !strings.Contains(err.Error(), "city is required") {
		t.Errorf("error = %v, want 'city is required'", err)
	}
}

func TestSearchCoords_InvalidJSON(t *testing.T) {
	args := json.RawMessage(`not-json`)
	_, err := SearchCoords(context.Background(), args)
	if err == nil {
		t.Fatal("expected error for invalid JSON")
	}
}

func TestSearchCoords_HTTPError(t *testing.T) {
	orig := httpClient
	SetHTTPClient(&http.Client{
		Transport: &mockSearchTransport{
			status: http.StatusInternalServerError,
			body:   "server error",
		},
	})
	defer func() { SetHTTPClient(orig) }()

	args := json.RawMessage(`{"city":"Nowhere"}`)
	_, err := SearchCoords(context.Background(), args)
	if err == nil {
		t.Fatal("expected error for search failure")
	}
}

func TestSearchCoords_EmptyResults(t *testing.T) {
	orig := httpClient
	SetHTTPClient(&http.Client{
		Transport: &mockSearchTransport{
			body: `[]`,
		},
	})
	defer func() { SetHTTPClient(orig) }()

	args := json.RawMessage(`{"city":"Xyzzy"}`)
	_, err := SearchCoords(context.Background(), args)
	if err == nil {
		t.Fatal("expected error for empty results")
	}
}

func TestSearchCoords_MalformedJSON(t *testing.T) {
	orig := httpClient
	SetHTTPClient(&http.Client{
		Transport: &mockSearchTransport{
			body: `not json`,
		},
	})
	defer func() { SetHTTPClient(orig) }()

	args := json.RawMessage(`{"city":"X"}`)
	_, err := SearchCoords(context.Background(), args)
	if err == nil {
		t.Fatal("expected error for malformed response")
	}
}

type mockSearchTransport struct {
	status int
	body   string
}

func (m *mockSearchTransport) RoundTrip(_ *http.Request) (*http.Response, error) {
	if m.status == 0 {
		m.status = http.StatusOK
	}
	return &http.Response{
		StatusCode: m.status,
		Body:       io.NopCloser(strings.NewReader(m.body)),
		Header:     http.Header{"Content-Type": []string{"application/json"}},
	}, nil
}

type mockDoer struct{ resp *http.Response }

func (m *mockDoer) Do(req *http.Request) (*http.Response, error) { return m.resp, nil }

func mockResp(body string) *http.Response {
	return &http.Response{
		StatusCode: http.StatusOK,
		Body:       io.NopCloser(strings.NewReader(body)),
		Header:     http.Header{"Content-Type": {"application/json"}},
	}
}

// TestSearchCoords_Schema verifies city 返回结构（name/longitude/latitude/country）
// 与 Result schema 声明一致（schema: tools_other.go city）
func TestSearchCoords_Schema(t *testing.T) {
	old := HttpClient()
	defer SetHTTPClient(old)
	// "北京" 现命中内置表；用内置表不存在的名字走 OSM 兜底路径验证 schema
	SetHTTPClient(&mockDoer{resp: mockResp(`[{"name":"赛博坦","lon":"116.4074","lat":"39.9042","address":{"country":"中国"}}]`)})

	out, err := SearchCoords(context.Background(), []byte(`{"city":"赛博坦"}`))
	if err != nil {
		t.Fatalf("SearchCoords: %v", err)
	}
	var r struct {
		Name      string  `json:"name"`
		Longitude float64 `json:"longitude"`
		Latitude  float64 `json:"latitude"`
		Country   string  `json:"country"`
	}
	if err := json.Unmarshal(out, &r); err != nil {
		t.Fatalf("unmarshal: %v (输出: %s)", err, out)
	}
	if r.Name != "赛博坦" {
		t.Errorf("name = %s, want 赛博坦", r.Name)
	}
	if r.Country != "中国" {
		t.Errorf("country = %s, want 中国", r.Country)
	}
}

// TestSearchCoords_GlobalSearch 海外城市能正确返回经纬度
func TestSearchCoords_GlobalSearch(t *testing.T) {
	old := HttpClient()
	defer SetHTTPClient(old)
	SetHTTPClient(&mockDoer{resp: mockResp(`[{"name":"New York","lon":"-74.006","lat":"40.7128","address":{"country":"United States","city":"New York"}}]`)})

	out, err := SearchCoords(context.Background(), []byte(`{"city":"New York"}`))
	if err != nil {
		t.Fatalf("SearchCoords: %v", err)
	}
	var r searchResult
	if err := json.Unmarshal(out, &r); err != nil {
		t.Fatalf("unmarshal: %v", err)
	}
	if r.Name != "New York" {
		t.Errorf("name = %q, want New York", r.Name)
	}
	if r.Longitude != -74.006 {
		t.Errorf("longitude = %f, want -74.006", r.Longitude)
	}
	if r.Latitude != 40.7128 {
		t.Errorf("latitude = %f, want 40.7128", r.Latitude)
	}
}

// TestSearchCoords_AdministrativePriority 行政级别优先（county/city/state 优先于 POI）
func TestSearchCoords_AdministrativePriority(t *testing.T) {
	old := HttpClient()
	defer SetHTTPClient(old)
	// 返回顺序：POI 在前，行政区域在后
	SetHTTPClient(&mockDoer{resp: mockResp(`[
		{"name":"Beijing Road","lon":"116.4","lat":"39.9","type":"poi","address":{"country":"China"}},
		{"name":"Beijing","lon":"116.4074","lat":"39.9042","type":"city","address":{"country":"China","city":"Beijing"}}
	]`)})

	out, err := SearchCoords(context.Background(), []byte(`{"city":"Beijing"}`))
	if err != nil {
		t.Fatalf("SearchCoords: %v", err)
	}
	var r searchResult
	if err := json.Unmarshal(out, &r); err != nil {
		t.Fatalf("unmarshal: %v", err)
	}
	// 应该选择 city 类型，而非 poi
	if r.Name != "Beijing" {
		t.Errorf("name = %q, want Beijing (should prefer city over poi)", r.Name)
	}
}

func TestBuiltin_CountyExact(t *testing.T) {
	// 抚远市（佳木斯下辖，中国最东端）应精确命中内置表
	result, err := SearchCoords(context.Background(), json.RawMessage(`{"city":"抚远市"}`))
	if err != nil {
		t.Fatalf("SearchCoords: %v", err)
	}
	var r searchResult
	if err := json.Unmarshal(result, &r); err != nil {
		t.Fatalf("unmarshal: %v", err)
	}
	if got, want := r.Longitude, 134.300135; got != want {
		t.Errorf("longitude = %v, want %v", got, want)
	}
	if got, want := r.Latitude, 48.362461; got != want {
		t.Errorf("latitude = %v, want %v", got, want)
	}
	if r.Country == "" {
		t.Error("country should be populated from embedded region")
	}
}

func TestBuiltin_CountySuffixMatch(t *testing.T) {
	// 不带"市"后缀应回退命中县级条目
	result, err := SearchCoords(context.Background(), json.RawMessage(`{"city":"抚远"}`))
	if err != nil {
		t.Fatalf("SearchCoords: %v", err)
	}
	var r searchResult
	if err := json.Unmarshal(result, &r); err != nil {
		t.Fatalf("unmarshal: %v", err)
	}
	if got, want := r.Longitude, 134.300135; got != want {
		t.Errorf("longitude = %v, want %v", got, want)
	}
}

func TestBuiltin_PrefectureFallback(t *testing.T) {
	// 只给地级市"佳木斯" → 命中 cities 代表坐标（向阳区）
	result, err := SearchCoords(context.Background(), json.RawMessage(`{"city":"佳木斯"}`))
	if err != nil {
		t.Fatalf("SearchCoords: %v", err)
	}
	var r searchResult
	if err := json.Unmarshal(result, &r); err != nil {
		t.Fatalf("unmarshal: %v", err)
	}
	if got, want := r.Longitude, 130.357562; got != want {
		t.Errorf("longitude = %v, want %v", got, want)
	}
}

func TestBuiltin_WorldCity(t *testing.T) {
	result, err := SearchCoords(context.Background(), json.RawMessage(`{"city":"东京"}`))
	if err != nil {
		t.Fatalf("SearchCoords: %v", err)
	}
	var r searchResult
	if err := json.Unmarshal(result, &r); err != nil {
		t.Fatalf("unmarshal: %v", err)
	}
	if got, want := r.Longitude, 139.692; got != want {
		t.Errorf("longitude = %v, want %v", got, want)
	}
	if got, want := r.Country, "日本"; got != want {
		t.Errorf("country = %q, want %q", got, want)
	}
}

func TestBuiltin_OSMFallback(t *testing.T) {
	// 内置表未命中（长尾海外地名）→ 走 OSM 兜底
	orig := httpClient
	SetHTTPClient(&http.Client{
		Transport: &mockSearchTransport{
			body: `[{"lat":"51.5074","lon":"-0.1278","name":"London","address":{"country":"United Kingdom","country_code":"GB"}}]`,
		},
	})
	defer func() { SetHTTPClient(orig) }()

	result, err := SearchCoords(context.Background(), json.RawMessage(`{"city":"Hammersmith"}`))
	if err != nil {
		t.Fatalf("SearchCoords: %v", err)
	}
	var r searchResult
	if err := json.Unmarshal(result, &r); err != nil {
		t.Fatalf("unmarshal: %v", err)
	}
	if got, want := r.Longitude, -0.1278; got != want {
		t.Errorf("longitude = %v, want %v", got, want)
	}
}

func TestBuiltin_DataIntegrity(t *testing.T) {
	// 校验内嵌数据完整性：条目数、经纬度合法、region 非空
	geoOnce.Do(loadGeo)
	if len(geo.Counties) < 2800 {
		t.Errorf("counties = %d, want >= 2800", len(geo.Counties))
	}
	if len(geo.Cities) < 300 {
		t.Errorf("cities = %d, want >= 300", len(geo.Cities))
	}
	if len(geo.World) < 50 {
		t.Errorf("world = %d, want >= 50", len(geo.World))
	}
	check := func(name string, m map[string]geoEntry) {
		for k, e := range m {
			if e.Lng < -180 || e.Lng > 180 || e.Lat < -90 || e.Lat > 90 {
				t.Errorf("%s[%s]: invalid coords %v,%v", name, k, e.Lng, e.Lat)
			}
			if e.Region == "" {
				t.Errorf("%s[%s]: empty region", name, k)
			}
		}
	}
	check("counties", geo.Counties)
	check("cities", geo.Cities)
	check("world", geo.World)

	// 佳木斯下辖关键县级必须独立存在（抚远/富锦/同江）
	for _, c := range []string{"抚远市", "富锦市", "同江市", "桦南县"} {
		if _, ok := geo.Counties[c]; !ok {
			t.Errorf("missing county %q", c)
		}
	}
}

func TestBuiltin_ConflictingSuffix(t *testing.T) {
	// 输入带错误后缀（表里是"抚远市"，输入"抚远区"）→ 去后缀回退应命中
	result, err := SearchCoords(context.Background(), json.RawMessage(`{"city":"抚远区"}`))
	if err != nil {
		t.Fatalf("SearchCoords: %v", err)
	}
	var r searchResult
	if err := json.Unmarshal(result, &r); err != nil {
		t.Fatalf("unmarshal: %v", err)
	}
	if got, want := r.Longitude, 134.300135; got != want {
		t.Errorf("longitude = %v, want %v", got, want)
	}
}

func TestBuiltin_Municipality(t *testing.T) {
	// 直辖市："北京" → 应命中 cities["北京市"]
	result, err := SearchCoords(context.Background(), json.RawMessage(`{"city":"北京"}`))
	if err != nil {
		t.Fatalf("SearchCoords: %v", err)
	}
	var r searchResult
	if err := json.Unmarshal(result, &r); err != nil {
		t.Fatalf("unmarshal: %v", err)
	}
	if got, want := r.Longitude, 116.4074; got < want-0.5 || got > want+0.5 {
		t.Errorf("longitude = %v, want near %v (北京)", got, want)
	}
}

func TestCountryToZh(t *testing.T) {
	cases := map[string]string{
		"China":            "中国",
		"Taiwan":           "中国台湾",
		"Hong Kong":        "中国香港",
		"Macao":            "中国澳门",
		"United Kingdom":   "英国",
		"United States":    "美国",
		"Japan":            "日本",
		"SomeUnknownPlace": "SomeUnknownPlace",
		"":                 "",
	}
	for in, want := range cases {
		if got := countryToZh(in); got != want {
			t.Errorf("countryToZh(%q) = %q, want %q", in, got, want)
		}
	}
}

func TestBuiltin_OSMFallbackCountryZh(t *testing.T) {
	// OSM 兜底返回英文国家名 → country 字段应统一为中文
	orig := httpClient
	SetHTTPClient(&http.Client{
		Transport: &mockSearchTransport{
			body: `[{"lat":"51.5074","lon":"-0.1278","name":"London","address":{"country":"United Kingdom","country_code":"GB"}}]`,
		},
	})
	defer func() { SetHTTPClient(orig) }()

	result, err := SearchCoords(context.Background(), json.RawMessage(`{"city":"Camden Town"}`))
	if err != nil {
		t.Fatalf("SearchCoords: %v", err)
	}
	var r searchResult
	if err := json.Unmarshal(result, &r); err != nil {
		t.Fatalf("unmarshal: %v", err)
	}
	if got, want := r.Country, "英国"; got != want {
		t.Errorf("country = %q, want %q", got, want)
	}
}

// TestSearchCoords_AmbiguousSameName：同名行政区跨省歧义——应提示省/市限定，不静默取错省。
func TestSearchCoords_AmbiguousSameName(t *testing.T) {
	orig := httpClient
	httpClient = &http.Client{
		Transport: &mockSearchTransport{
			status: 200,
			body: `[
				{"lat":"41.56","lon":"120.45","name":"北山县","type":"county",
				 "address":{"country":"中国","state":"辽宁省","county":"北山县"}},
				{"lat":"34.25","lon":"119.60","name":"北山县","type":"county",
				 "address":{"country":"中国","state":"江苏省","county":"北山县"}}
			]`,
		},
	}
	defer func() { httpClient = orig }()
	_, err := SearchCoords(context.Background(), json.RawMessage(`{"city":"北山县"}`))
	if err == nil {
		t.Fatal("同名行政区歧义应返回错误提示，got nil")
	}
	if !strings.Contains(err.Error(), "同名行政区跨省歧义") {
		t.Errorf("错误应提示同名歧义，got: %v", err)
	}
}
