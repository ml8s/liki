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

func TestSearchCity_Valid(t *testing.T) {
	orig := httpClient
	SetHTTPClient(&http.Client{
		Transport: &mockSearchTransport{
			body: `[{"lat":"39.9042","lon":"116.4074","name":"Beijing","address":{"country":"China","country_code":"CN"}}]`,
		},
	})
	defer func() { SetHTTPClient(orig) }()

	args := json.RawMessage(`{"city":"Beijing"}`)
	result, err := SearchCity(context.Background(), args)
	if err != nil {
		t.Fatalf("SearchCity: %v", err)
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

func TestSearchCity_EmptyCityName(t *testing.T) {
	args := json.RawMessage(`{"city":""}`)
	_, err := SearchCity(context.Background(), args)
	if err == nil {
		t.Fatal("expected error for empty city")
	}
	if !strings.Contains(err.Error(), "city is required") {
		t.Errorf("error = %v, want 'city is required'", err)
	}
}

func TestSearchCity_InvalidJSON(t *testing.T) {
	args := json.RawMessage(`not-json`)
	_, err := SearchCity(context.Background(), args)
	if err == nil {
		t.Fatal("expected error for invalid JSON")
	}
}

func TestSearchCity_HTTPError(t *testing.T) {
	orig := httpClient
	SetHTTPClient(&http.Client{
		Transport: &mockSearchTransport{
			status: http.StatusInternalServerError,
			body:   "server error",
		},
	})
	defer func() { SetHTTPClient(orig) }()

	args := json.RawMessage(`{"city":"Nowhere"}`)
	_, err := SearchCity(context.Background(), args)
	if err == nil {
		t.Fatal("expected error for search failure")
	}
}

func TestSearchCity_EmptyResults(t *testing.T) {
	orig := httpClient
	SetHTTPClient(&http.Client{
		Transport: &mockSearchTransport{
			body: `[]`,
		},
	})
	defer func() { SetHTTPClient(orig) }()

	args := json.RawMessage(`{"city":"Xyzzy"}`)
	_, err := SearchCity(context.Background(), args)
	if err == nil {
		t.Fatal("expected error for empty results")
	}
}

func TestSearchCity_MalformedJSON(t *testing.T) {
	orig := httpClient
	SetHTTPClient(&http.Client{
		Transport: &mockSearchTransport{
			body: `not json`,
		},
	})
	defer func() { SetHTTPClient(orig) }()

	args := json.RawMessage(`{"city":"X"}`)
	_, err := SearchCity(context.Background(), args)
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

// TestSearchCity_Schema verifies city 返回结构（name/longitude/latitude/country）
// 与 Result schema 声明一致（schema: tools_other.go city）
func TestSearchCity_Schema(t *testing.T) {
	old := HttpClient()
	defer SetHTTPClient(old)
	SetHTTPClient(&mockDoer{resp: mockResp(`[{"name":"北京","lon":"116.4074","lat":"39.9042","address":{"country":"中国"}}]`)})

	out, err := SearchCity(context.Background(), []byte(`{"city":"北京"}`))
	if err != nil {
		t.Fatalf("SearchCity: %v", err)
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
