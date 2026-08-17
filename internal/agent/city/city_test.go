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
	if r.Country != "China" {
		t.Errorf("country = %q, want China", r.Country)
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
	SetHTTPClient(&mockDoer{resp: mockResp(`[{"name":"北京","lon":"116.4074","lat":"39.9042","address":{"country":"中国"}}]`)})

	out, err := SearchCoords(context.Background(), []byte(`{"city":"北京"}`))
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
	if r.Name != "北京" {
		t.Errorf("name = %s, want 北京", r.Name)
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
