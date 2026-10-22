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

func TestSearchCoords_CanDisableExternalGeocoder(t *testing.T) {
	t.Setenv("LIKI_EXTERNAL_GEOCODING", "off")
	_, err := SearchCoords(
		context.Background(),
		json.RawMessage(`{"city":"definitely-not-in-the-builtin-table"}`),
	)
	if err == nil {
		t.Fatal("expected external geocoding to be disabled")
	}
	if !strings.Contains(err.Error(), "外部地理编码已禁用") {
		t.Errorf("unexpected error: %v", err)
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
// 与 city_coords 的 Result schema 声明一致（wire DTO: mcpcontract/city.go）
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

	out, err := SearchCoords(context.Background(), []byte(`{"city":"Springfield"}`))
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
	if got, want := r.Longitude, 134.294501; got != want {
		t.Errorf("longitude = %v, want %v", got, want)
	}
	if got, want := r.Latitude, 48.364707; got != want {
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
	if got, want := r.Longitude, 134.294501; got != want {
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
	if got, want := r.Longitude, 130.361634; got != want {
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
	check := func(name string, records []geoRecord) {
		for _, e := range records {
			if e.Lng < -180 || e.Lng > 180 || e.Lat < -90 || e.Lat > 90 {
				t.Errorf("%s[%s]: invalid coords %v,%v", name, e.Name, e.Lng, e.Lat)
			}
			if e.Region == "" {
				t.Errorf("%s[%s]: empty region", name, e.Name)
			}
		}
	}
	check("counties", geo.Counties)
	check("cities", geo.Cities)
	check("world", geo.World)

	// 佳木斯下辖关键县级必须独立存在（抚远/富锦/同江）
	for _, c := range []string{"抚远市", "富锦市", "同江市", "桦南县"} {
		if _, ok := geoIndex[c]; !ok {
			t.Errorf("missing county %q", c)
		}
	}

	// 跨省同名区县必须全量存在（曾被 name 键覆盖：北京朝阳区、天津和平区、杭州西湖区……）。
	// 期望条数与民政部区划全表一致，新增同名区县时同步此表。
	expectedDuplicates := map[string]int{
		"朝阳区": 2, "通州区": 2, "和平区": 2, "河东区": 2, "普陀区": 2, "宝山区": 2,
		"白云区": 2, "西湖区": 2, "鼓楼区": 4, "长安区": 2, "新城区": 2, "江北区": 2,
		"城关区": 2, "市中区": 4, "青山区": 2, "新华区": 3, "桥西区": 2, "铁西区": 3,
		"城区": 3, "郊区": 3, "向阳区": 2, "南山区": 2, "城中区": 2, "永定区": 2,
		"海州区": 2, "西安区": 2, "铁东区": 2, "龙华区": 2,
	}
	for name, want := range expectedDuplicates {
		if got := len(geoIndex[name]); got != want {
			t.Errorf("duplicate district %q: %d records, want %d", name, got, want)
		}
	}

	// (name, region, city) 唯一：同名同省不同市（南京/徐州鼓楼区）必须靠 city 区分。
	seen := map[string]bool{}
	for _, record := range geo.Counties {
		key := record.Name + "|" + record.Region + "|" + record.City
		if seen[key] {
			t.Errorf("duplicate county record: %s", key)
		}
		seen[key] = true
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
	if got, want := r.Longitude, 134.294501; got != want {
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

func TestSearchCoords_AmbiguousNameFailsClosed(t *testing.T) {
	// "朝阳" 在内置表中同时对应 朝阳区 与 朝阳县：必须报歧义，不得猜一个。
	_, err := SearchCoords(context.Background(), json.RawMessage(`{"city":"朝阳"}`))
	if err == nil || !strings.Contains(err.Error(), "歧义") {
		t.Fatalf("expected ambiguity error, got %v", err)
	}
}

func TestBuiltin_AmbiguousDistrictFailsClosed(t *testing.T) {
	// "朝阳区" 跨省同名（北京/长春）：必须报歧义并列出省限定，不得静默返回吉林坐标。
	_, err := SearchCoords(context.Background(), json.RawMessage(`{"city":"朝阳区"}`))
	if err == nil {
		t.Fatal("跨省同名区县应返回歧义错误")
	}
	for _, want := range []string{"歧义", "北京市朝阳区", "吉林省"} {
		if !strings.Contains(err.Error(), want) {
			t.Errorf("error %q should contain %q", err, want)
		}
	}
}

func TestBuiltin_CompoundRegionResolves(t *testing.T) {
	// "省/市 + 名称" 复合形式应消歧到正确省份与经度区间。
	cases := []struct {
		query          string
		country        string
		lngMin, lngMax float64
	}{
		{"北京朝阳区", "北京市", 116.3, 116.6},
		{"天津和平区", "天津市", 117.1, 117.4},
		{"杭州西湖区", "浙江省", 120.0, 120.3},
		{"南京鼓楼区", "江苏省", 118.6, 118.9},
		{"徐州鼓楼区", "江苏省", 117.0, 117.4},
		{"石家庄桥西区", "河北省", 114.3, 114.6},
	}
	for _, c := range cases {
		result, err := SearchCoords(context.Background(), json.RawMessage(`{"city":"`+c.query+`"}`))
		if err != nil {
			t.Errorf("%s: %v", c.query, err)
			continue
		}
		var r searchResult
		if err := json.Unmarshal(result, &r); err != nil {
			t.Errorf("%s: unmarshal: %v", c.query, err)
			continue
		}
		if r.Country != c.country {
			t.Errorf("%s: country = %q, want %q", c.query, r.Country, c.country)
		}
		if r.Longitude < c.lngMin || r.Longitude > c.lngMax {
			t.Errorf("%s: longitude = %v, want in [%v,%v]", c.query, r.Longitude, c.lngMin, c.lngMax)
		}
	}
}

func TestResolveCandidates_SuffixMatchesCityAndCounty(t *testing.T) {
	candidates, err := ResolveCandidates(context.Background(), "长沙")
	if err != nil {
		t.Fatalf("ResolveCandidates: %v", err)
	}
	if len(candidates) < 2 {
		t.Fatalf("expected at least 长沙市/长沙县, got %+v", candidates)
	}
	if candidates[0].Name != "长沙市" {
		t.Errorf("city should rank first, got %q", candidates[0].Name)
	}
	if candidates[0].Level != "city" {
		t.Errorf("长沙市 level = %q, want city", candidates[0].Level)
	}
	var county *Candidate
	for i := range candidates {
		if candidates[i].Name == "长沙县" {
			county = &candidates[i]
		}
	}
	if county == nil {
		t.Fatalf("长沙县 missing from candidates: %+v", candidates)
	}
	if county.Level != "county" || county.Region != "湖南省" {
		t.Errorf("长沙县 = %+v, want county/湖南省", *county)
	}
}

func TestResolveCandidates_AmbiguousDistrict(t *testing.T) {
	candidates, err := ResolveCandidates(context.Background(), "朝阳区")
	if err != nil {
		t.Fatalf("ResolveCandidates: %v", err)
	}
	regions := make(map[string]bool)
	for _, candidate := range candidates {
		if candidate.Name != "朝阳区" {
			continue
		}
		regions[candidate.Region] = true
	}
	if !regions["北京市"] || !regions["吉林省"] {
		t.Fatalf("expected 北京市 and 吉林省 朝阳区, got %+v", candidates)
	}
}

func TestResolveCandidates_CompoundDisambiguates(t *testing.T) {
	cases := []struct {
		query  string
		name   string
		region string
		city   string
	}{
		{"北京朝阳区", "朝阳区", "北京市", ""},
		{"湖南长沙", "长沙市", "湖南省", ""},
		{"杭州西湖区", "西湖区", "浙江省", "杭州市"},
	}
	for _, c := range cases {
		candidates, err := ResolveCandidates(context.Background(), c.query)
		if err != nil {
			t.Errorf("%s: %v", c.query, err)
			continue
		}
		if len(candidates) == 0 {
			t.Errorf("%s: no candidates", c.query)
			continue
		}
		top := candidates[0]
		if top.Name != c.name || top.Region != c.region || top.City != c.city {
			t.Errorf("%s: top = %+v, want %s/%s/%s", c.query, top, c.name, c.region, c.city)
		}
	}
}

func TestResolveCandidates_ExactFirst(t *testing.T) {
	candidates, err := ResolveCandidates(context.Background(), "长沙市")
	if err != nil {
		t.Fatalf("ResolveCandidates: %v", err)
	}
	if len(candidates) == 0 || candidates[0].Name != "长沙市" {
		t.Fatalf("exact name should rank first, got %+v", candidates)
	}
}

func TestResolveCandidates_UnknownAndEmpty(t *testing.T) {
	for _, query := range []string{"", "   ", "不存在的地方xyz"} {
		candidates, err := ResolveCandidates(context.Background(), query)
		if err != nil {
			t.Fatalf("%q: %v", query, err)
		}
		if len(candidates) != 0 {
			t.Errorf("%q: want empty candidates, got %+v", query, candidates)
		}
	}
}

func TestResolveCandidates_WorldCity(t *testing.T) {
	candidates, err := ResolveCandidates(context.Background(), "纽约")
	if err != nil {
		t.Fatalf("ResolveCandidates: %v", err)
	}
	if len(candidates) == 0 || candidates[0].Name != "纽约" || candidates[0].Level != "world" {
		t.Fatalf("expected world 纽约, got %+v", candidates)
	}
}

func TestResolveCandidates_LimitAndDedup(t *testing.T) {
	candidates, err := ResolveCandidates(context.Background(), "州")
	if err != nil {
		t.Fatalf("ResolveCandidates: %v", err)
	}
	if len(candidates) > candidateLimit {
		t.Fatalf("candidates = %d, want <= %d", len(candidates), candidateLimit)
	}
	seen := make(map[string]bool, len(candidates))
	for _, candidate := range candidates {
		key := candidate.Name + "|" + candidate.Region + "|" + candidate.City
		if seen[key] {
			t.Errorf("duplicate candidate %q", key)
		}
		seen[key] = true
	}
}

func TestSearchCoords_AllCityNamesResolve(t *testing.T) {
	geoOnce.Do(loadGeo)
	for _, record := range geo.Cities {
		if _, ok, err := searchBuiltin(record.Name); err != nil {
			t.Fatalf("%s: %v", record.Name, err)
		} else if !ok {
			t.Fatalf("%s: not found", record.Name)
		}
	}
}

func TestBuiltinChinaRecordsHaveTimezones(t *testing.T) {
	geoOnce.Do(loadGeo)
	for _, table := range [][]geoRecord{geo.Counties, geo.Cities} {
		for _, record := range table {
			if record.TZ != "Asia/Shanghai" {
				t.Fatalf("%s (%s) tz=%q, want Asia/Shanghai", record.Name, record.Region, record.TZ)
			}
		}
	}
}

func TestResolveWorldAlias(t *testing.T) {
	candidates, err := ResolveCandidates(context.Background(), "New York")
	if err != nil {
		t.Fatalf("ResolveCandidates: %v", err)
	}
	if len(candidates) == 0 || candidates[0].Name != "纽约" {
		t.Fatalf("New York alias: got %+v", candidates)
	}
	if candidates[0].Timezone != "America/New_York" {
		t.Fatalf("New York timezone: %q", candidates[0].Timezone)
	}
	if candidates[0].BBox == nil {
		t.Fatalf("New York bbox missing")
	}
}
