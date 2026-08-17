package city

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"net/url"
	"time"
)

// httpDoer is the interface for HTTP clients, allowing test injection.
type httpDoer interface {
	Do(req *http.Request) (*http.Response, error)
}

var httpClient httpDoer = &http.Client{Timeout: 15 * time.Second}

// SetHTTPClient replaces the HTTP client used for Nominatim queries. Call from
// tests to inject a mock transport.
func SetHTTPClient(c httpDoer) { httpClient = c }

// HttpClient returns the current HTTP client. Useful for save/restore in tests.
func HttpClient() httpDoer { return httpClient }

type searchResult struct {
	Name      string  `json:"name"`
	Longitude float64 `json:"longitude"`
	Latitude  float64 `json:"latitude"`
	Country   string  `json:"country"`
}

// SearchCoords resolves a city name to coordinates using Nominatim.
func SearchCoords(ctx context.Context, raw json.RawMessage) (json.RawMessage, error) {
	var args struct {
		City string `json:"city"`
	}
	if err := json.Unmarshal(raw, &args); err != nil {
		return nil, fmt.Errorf("city: search: %w", err)
	}
	if args.City == "" {
		return nil, fmt.Errorf("city is required")
	}

	result, err := searchNominatim(ctx, args.City)
	if err != nil {
		return nil, fmt.Errorf("未找到城市 '%s'，请尝试附近大城市或直接提供经纬度和时区: %w", args.City, err)
	}
	return json.Marshal(result)
}

func searchNominatim(ctx context.Context, query string) (searchResult, error) {
	// 两轮查询：先中国范围（避免县级地名被 POI 抢占），无结果时 fallback 全球范围。
	vals := url.Values{
		"q":               {query},
		"format":          {"json"},
		"limit":           {"5"},
		"accept-language": {"zh"},
		"addressdetails":  {"1"},
	}
	// 第一轮：中国范围
	vals.Set("countrycodes", "cn")
	uCN := "https://nominatim.openstreetmap.org/search?" + vals.Encode()

	req, err := http.NewRequestWithContext(ctx, "GET", uCN, nil)
	if err != nil {
		return searchResult{}, fmt.Errorf("search: new request: %w", err)
	}
	req.Header.Set("User-Agent", "Liki/1.0 (liki.app)")

	resp, err := httpClient.Do(req)
	if err != nil {
		return searchResult{}, fmt.Errorf("search: get: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return searchResult{}, fmt.Errorf("search: status %d", resp.StatusCode)
	}

	var results []struct {
		Lat     string `json:"lat"`
		Lon     string `json:"lon"`
		Name    string `json:"name"`
		Type    string `json:"type"`
		Address struct {
			Country     string `json:"country"`
			CountryCode string `json:"country_code"`
			County      string `json:"county"`
			City        string `json:"city"`
			State       string `json:"state"`
			Village     string `json:"village"`
			Town        string `json:"town"`
		} `json:"address"`
	}
	if err := json.NewDecoder(resp.Body).Decode(&results); err != nil {
		return searchResult{}, fmt.Errorf("search: decode: %w", err)
	}

	// 第二轮：无结果时 fallback 全球范围
	if len(results) == 0 {
		vals.Del("countrycodes")
		uGlobal := "https://nominatim.openstreetmap.org/search?" + vals.Encode()
		req2, err := http.NewRequestWithContext(ctx, "GET", uGlobal, nil)
		if err != nil {
			return searchResult{}, fmt.Errorf("search: new request: %w", err)
		}
		req2.Header.Set("User-Agent", "Liki/1.0 (liki.app)")
		resp2, err := httpClient.Do(req2)
		if err != nil {
			return searchResult{}, fmt.Errorf("search: get: %w", err)
		}
		defer resp2.Body.Close()
		if resp2.StatusCode != http.StatusOK {
			return searchResult{}, fmt.Errorf("search: status %d", resp2.StatusCode)
		}
		if err := json.NewDecoder(resp2.Body).Decode(&results); err != nil {
			return searchResult{}, fmt.Errorf("search: decode: %w", err)
		}
	}

	if len(results) == 0 {
		return searchResult{}, fmt.Errorf("search: no results for %s", query)
	}

	// 优先行政级别结果（county/city/state 地址或行政类型），排除 POI/街道抢占（外部评审 ③）。
	r := results[0]
	for _, cand := range results {
		if cand.Address.County != "" || cand.Address.City != "" || cand.Address.State != "" ||
			cand.Type == "administrative" || cand.Type == "county" || cand.Type == "city" {
			r = cand
			break
		}
	}
	lon, err := parseFloat(r.Lon)
	if err != nil {
		return searchResult{}, fmt.Errorf("search: parse lon: %w", err)
	}
	lat, err := parseFloat(r.Lat)
	if err != nil {
		return searchResult{}, fmt.Errorf("search: parse lat: %w", err)
	}
	return searchResult{
		Name:      r.Name,
		Longitude: lon,
		Latitude:  lat,
		Country:   r.Address.Country,
	}, nil
}

func parseFloat(s string) (float64, error) {
	var f float64
	if n, err := fmt.Sscanf(s, "%f", &f); n != 1 || err != nil {
		return 0, fmt.Errorf("parseFloat: %q: %w", s, err)
	}
	return f, nil
}
