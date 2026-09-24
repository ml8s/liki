package city

import (
	"context"
	_ "embed"
	"encoding/json"
	"fmt"
	"net/http"
	"net/url"
	"strings"
	"sync"
	"time"
)

//go:embed cities_data.json
var citiesData []byte

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

type geoEntry struct {
	Lng    float64 `json:"lng"`
	Lat    float64 `json:"lat"`
	Region string  `json:"region"`
}

type geoDB struct {
	Counties map[string]geoEntry `json:"counties"`
	Cities   map[string]geoEntry `json:"cities"`
	World    map[string]geoEntry `json:"world"`
}

var (
	geoOnce sync.Once
	geo     geoDB
)

func loadGeo() {
	if err := json.Unmarshal(citiesData, &geo); err != nil {
		panic("city: load embedded data: " + err.Error())
	}
}

// regionSuffixes are administrative-division suffixes stripped for fuzzy match.
var regionSuffixes = []string{"市", "区", "县", "旗", "盟", "州"}

// countryZh maps OSM English/OSM-native country names to Chinese, so the
// country field stays consistent with the embedded table's Chinese regions.
var countryZh = map[string]string{
	"China": "中国", "Taiwan": "中国台湾", "Hong Kong": "中国香港", "Macau": "中国澳门", "Macao": "中国澳门",
	"Japan": "日本", "South Korea": "韩国", "Korea": "韩国", "Singapore": "新加坡", "Malaysia": "马来西亚",
	"Thailand": "泰国", "Indonesia": "印度尼西亚", "Philippines": "菲律宾", "Laos": "老挝", "Cambodia": "柬埔寨",
	"Myanmar": "缅甸", "Bangladesh": "孟加拉国", "Nepal": "尼泊尔", "Sri Lanka": "斯里兰卡", "India": "印度",
	"Vietnam": "越南", "Mongolia": "蒙古",
	"Australia": "澳大利亚", "New Zealand": "新西兰",
	"United States": "美国", "United States of America": "美国", "USA": "美国", "Canada": "加拿大",
	"United Kingdom": "英国", "England": "英国", "France": "法国", "Germany": "德国", "Netherlands": "荷兰",
	"Belgium": "比利时", "Austria": "奥地利", "Switzerland": "瑞士", "Italy": "意大利", "Spain": "西班牙",
	"Portugal": "葡萄牙", "Ireland": "爱尔兰", "Denmark": "丹麦", "Sweden": "瑞典", "Norway": "挪威",
	"Finland": "芬兰", "Poland": "波兰", "Czechia": "捷克", "Greece": "希腊", "Hungary": "匈牙利",
	"Russia": "俄罗斯", "Ukraine": "乌克兰", "Turkey": "土耳其", "Israel": "以色列", "United Arab Emirates": "阿联酋",
	"Saudi Arabia": "沙特阿拉伯", "Qatar": "卡塔尔", "Kuwait": "科威特", "Iran": "伊朗", "Iraq": "伊拉克",
	"Brazil": "巴西", "Mexico": "墨西哥", "Argentina": "阿根廷", "Chile": "智利", "Peru": "秘鲁",
	"Colombia": "哥伦比亚", "Venezuela": "委内瑞拉",
	"Egypt": "埃及", "South Africa": "南非", "Kenya": "肯尼亚", "Nigeria": "尼日利亚", "Morocco": "摩洛哥",
	"Ethiopia": "埃塞俄比亚", "Tanzania": "坦桑尼亚", "Ghana": "加纳",
	"Pakistan": "巴基斯坦", "Afghanistan": "阿富汗", "Uzbekistan": "乌兹别克斯坦", "Kazakhstan": "哈萨克斯坦",
	"France (Metropolitan)": "法国",
}

func countryToZh(name string) string {
	if zh, ok := countryZh[name]; ok {
		return zh
	}
	return name
}

// SearchCoords resolves a city name to coordinates. The embedded table
// (counties/cities/world, WGS-84) is checked first; Nominatim (OSM) is used
// as a fallback for long-tail places.
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

	result, ok := searchBuiltin(args.City)
	if !ok {
		var err error
		result, err = searchNominatim(ctx, args.City)
		if err != nil {
			return nil, fmt.Errorf("未找到城市 '%s'，请尝试附近大城市或直接提供经纬度和时区: %w", args.City, err)
		}
	}
	return json.Marshal(result)
}

// searchBuiltin resolves a place name against the embedded table.
func searchBuiltin(name string) (searchResult, bool) {
	geoOnce.Do(loadGeo)
	name = strings.TrimSpace(name)
	if name == "" {
		return searchResult{}, false
	}
	for _, m := range []map[string]geoEntry{geo.Counties, geo.Cities, geo.World} {
		if e, ok := m[name]; ok {
			return resultFrom(name, e), true
		}
	}
	for _, s := range regionSuffixes {
		for _, m := range []map[string]geoEntry{geo.Counties, geo.Cities, geo.World} {
			if e, ok := m[name+s]; ok {
				return resultFrom(name+s, e), true
			}
		}
	}
	if r, ok := matchByStrippedSuffix(name); ok {
		return r, true
	}
	return searchResult{}, false
}

func matchByStrippedSuffix(name string) (searchResult, bool) {
	base := stripRegionSuffix(name)
	if base == "" {
		return searchResult{}, false
	}
	for _, m := range []map[string]geoEntry{geo.Counties, geo.Cities, geo.World} {
		for key, e := range m {
			if stripRegionSuffix(key) == base {
				return resultFrom(key, e), true
			}
		}
	}
	return searchResult{}, false
}

// stripRegionSuffix removes trailing administrative-division suffixes.
func stripRegionSuffix(s string) string {
	for _, suf := range regionSuffixes {
		s = strings.TrimSuffix(s, suf)
	}
	return s
}

func resultFrom(name string, e geoEntry) searchResult {
	return searchResult{Name: name, Longitude: e.Lng, Latitude: e.Lat, Country: e.Region}
}

func searchNominatim(ctx context.Context, query string) (searchResult, error) {
	vals := url.Values{
		"q":               {query},
		"format":          {"json"},
		"limit":           {"5"},
		"accept-language": {"zh"},
		"addressdetails":  {"1"},
	}
	u := "https://nominatim.openstreetmap.org/search?" + vals.Encode()

	req, err := http.NewRequestWithContext(ctx, "GET", u, nil)
	if err != nil {
		return searchResult{}, fmt.Errorf("search: new request: %w", err)
	}
	req.Header.Set("User-Agent", "Liki-Engine/2026.08 (https://liki.hk; contact: api@liki.hk)")

	resp, err := httpClient.Do(req)
	if err != nil {
		return searchResult{}, fmt.Errorf("search: get: %w", err)
	}
	defer func() { _ = resp.Body.Close() }()

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
	if len(results) == 0 {
		return searchResult{}, fmt.Errorf("search: no results for %s", query)
	}

	// Prefer administrative results over POIs and streets.
	var admins []int
	for i, cand := range results {
		if cand.Address.County != "" || cand.Address.City != "" || cand.Address.State != "" ||
			cand.Type == "administrative" || cand.Type == "county" || cand.Type == "city" {
			admins = append(admins, i)
		}
	}
	if len(admins) > 1 && results[admins[0]].Name == results[admins[1]].Name {
		// 同名行政区跨省歧义：无法可靠消歧，明确提示（不静默取错省）。
		return searchResult{}, fmt.Errorf("同名行政区跨省歧义，请提供省/市限定（如 '辽宁朝阳'）：%s", query)
	}
	if len(admins) == 0 {
		admins = []int{0}
	}
	r := results[admins[0]]
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
		Country:   countryToZh(r.Address.Country),
	}, nil
}

func parseFloat(s string) (float64, error) {
	var f float64
	if n, err := fmt.Sscanf(s, "%f", &f); n != 1 || err != nil {
		return 0, fmt.Errorf("parseFloat: %q: %w", s, err)
	}
	return f, nil
}
