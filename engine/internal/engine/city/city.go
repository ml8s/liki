package city

import (
	"context"
	_ "embed"
	"encoding/json"
	"fmt"
	"net/http"
	"net/url"
	"os"
	"sort"
	"strconv"
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

// Coords is the resolved geocoding result. It is the typed domain boundary;
// MCP applies its own public DTO on top of it.
type Coords struct {
	Name      string  `json:"name"`
	Longitude float64 `json:"longitude"`
	Latitude  float64 `json:"latitude"`
	Country   string  `json:"country"`
	Timezone  string  `json:"timezone,omitempty"`
	BBox      *BBox   `json:"bbox,omitempty"`
}

// BBox is the raw administrative extent stored with a place record.
type BBox struct {
	West  float64 `json:"west"`
	East  float64 `json:"east"`
	South float64 `json:"south"`
	North float64 `json:"north"`
}

// searchResult is retained as the internal spelling used by geocoding helpers.
type searchResult = Coords

// Candidate is one possible place match. Candidate mode never guesses: callers
// receive the ranked list and pick the intended place themselves.
type Candidate struct {
	Name      string  `json:"name"`
	Region    string  `json:"region"`
	City      string  `json:"city,omitempty"`
	Country   string  `json:"country"`
	Level     string  `json:"level"`
	Longitude float64 `json:"longitude"`
	Latitude  float64 `json:"latitude"`
	Timezone  string  `json:"timezone,omitempty"`
	BBox      *BBox   `json:"bbox,omitempty"`
}

// candidateLimit caps how many candidates a single query may return.
const candidateLimit = 20

// geoRecord 是内嵌地理表的一条记录。县级表存在跨省同名（如朝阳区：北京/长春），
// 因此名称不是唯一键：同名记录靠 Region（省）与 City（地级市，可空）消歧。
type geoRecord struct {
	Name    string   `json:"name"`
	Region  string   `json:"region"`
	City    string   `json:"city,omitempty"`
	Lng     float64  `json:"lng"`
	Lat     float64  `json:"lat"`
	TZ      string   `json:"tz,omitempty"`
	BBox    *BBox    `json:"bbox,omitempty"`
	Aliases []string `json:"aliases,omitempty"`
}

type geoDB struct {
	Counties []geoRecord `json:"counties"`
	Cities   []geoRecord `json:"cities"`
	World    []geoRecord `json:"world"`
}

var (
	geoOnce        sync.Once
	geo            geoDB
	geoIndex       map[string][]geoRecord
	nominatimCache = newExternalGeoCache(4096, 24*time.Hour)
)

type externalGeoEntry struct {
	result  Coords
	expires time.Time
}

type externalGeoCache struct {
	mu      sync.Mutex
	max     int
	ttl     time.Duration
	entries map[string]externalGeoEntry
}

func newExternalGeoCache(max int, ttl time.Duration) *externalGeoCache {
	return &externalGeoCache{max: max, ttl: ttl, entries: make(map[string]externalGeoEntry)}
}

func (c *externalGeoCache) get(query string) (Coords, bool) {
	c.mu.Lock()
	defer c.mu.Unlock()
	entry, ok := c.entries[query]
	if !ok || time.Now().After(entry.expires) {
		if ok {
			delete(c.entries, query)
		}
		return Coords{}, false
	}
	return entry.result, true
}

func (c *externalGeoCache) set(query string, result Coords) {
	c.mu.Lock()
	defer c.mu.Unlock()
	if len(c.entries) >= c.max {
		now := time.Now()
		for key, item := range c.entries {
			if now.After(item.expires) {
				delete(c.entries, key)
			}
		}
		if len(c.entries) >= c.max {
			c.entries = make(map[string]externalGeoEntry)
		}
	}
	c.entries[query] = externalGeoCacheEntry(result, c.ttl)
}

func externalGeoCacheEntry(result Coords, ttl time.Duration) externalGeoEntry {
	return externalGeoEntry{result: result, expires: time.Now().Add(ttl)}
}

func loadGeo() {
	if err := json.Unmarshal(citiesData, &geo); err != nil {
		panic("city: load embedded data: " + err.Error())
	}
	geoIndex = make(map[string][]geoRecord, len(geo.Counties)+len(geo.Cities)+len(geo.World))
	// cities 与 counties 在少数地级市上存在完全相同的记录（同名同省同坐标），
	// 合并索引时按记录身份去重，避免把同一地点报成歧义。
	seen := make(map[string]struct{}, len(geo.Counties)+len(geo.Cities)+len(geo.World))
	for _, table := range [][]geoRecord{geo.Counties, geo.Cities, geo.World} {
		for _, record := range table {
			key := fmt.Sprintf("%s|%s|%s|%.6f|%.6f", record.Name, record.Region, record.City, record.Lng, record.Lat)
			if _, ok := seen[key]; ok {
				continue
			}
			seen[key] = struct{}{}
			geoIndex[record.Name] = append(geoIndex[record.Name], record)
			for _, alias := range record.Aliases {
				if alias == "" || alias == record.Name {
					continue
				}
				geoIndex[alias] = append(geoIndex[alias], record)
			}
		}
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
	result, err := ResolveCoords(ctx, args.City)
	if err != nil {
		return nil, err
	}
	return json.Marshal(result)
}

// ResolveCoords resolves a city name to typed coordinates. The embedded table
// (counties/cities/world, WGS-84) is checked first; Nominatim (OSM) is used
// as a fallback for long-tail places.
func ResolveCoords(ctx context.Context, query string) (Coords, error) {
	if query == "" {
		return Coords{}, fmt.Errorf("city is required")
	}
	result, ok, err := searchBuiltin(query)
	if err != nil {
		return Coords{}, err
	}
	if ok {
		return result, nil
	}
	if cached, hit := nominatimCache.get(query); hit {
		return cached, nil
	}
	if !externalGeocodingEnabled() {
		return Coords{}, fmt.Errorf(
			"城市 '%s' 不在内置表且外部地理编码已禁用；请提供经纬度或开启 LIKI_EXTERNAL_GEOCODING",
			query,
		)
	}
	result, err = searchNominatim(ctx, query)
	if err != nil {
		return Coords{}, fmt.Errorf("未找到城市 '%s'，请尝试附近大城市或直接提供经纬度和时区: %w", query, err)
	}
	nominatimCache.set(query, result)
	return result, nil
}

func externalGeocodingEnabled() bool {
	switch strings.ToLower(strings.TrimSpace(os.Getenv("LIKI_EXTERNAL_GEOCODING"))) {
	case "0", "false", "off":
		return false
	default:
		return true
	}
}

// searchBuiltin resolves a place name against the embedded table.
//
// 唯一名直接命中；同名多条（跨省同名区县，如朝阳区：北京/长春）必须显式消歧：
// 先尝试“省/市 + 名称”复合形式（如“北京朝阳区”“南京鼓楼区”），无法消歧时
// fail-closed 列出候选，绝不静默返回某一个同名地点的坐标。
// 后缀回退（name+后缀、去后缀同名）同样按记录消歧。
func searchBuiltin(name string) (Coords, bool, error) {
	geoOnce.Do(loadGeo)
	name = strings.TrimSpace(name)
	if name == "" {
		return Coords{}, false, nil
	}
	if records := geoIndex[name]; len(records) > 0 {
		return resolveRecords(name, records)
	}
	if result, ok, err := searchCompound(name); ok || err != nil {
		return result, ok, err
	}
	candidates := suffixCandidates(name)
	if len(candidates) == 0 {
		return Coords{}, false, nil
	}
	return resolveRecords(name, candidates)
}

// ResolveCandidates returns the ranked built-in matches for query instead of
// failing on ambiguity. It never calls external geocoding: empty results mean
// the place is not in the embedded table.
//
// Ranking: exact name, region-qualified compound (北京朝阳区 / 湖南长沙),
// suffix-equivalent (长沙 → 长沙市/长沙县), name prefix, name substring; within
// one class prefectures rank before counties and world cities.
func ResolveCandidates(_ context.Context, query string) ([]Candidate, error) {
	geoOnce.Do(loadGeo)
	query = strings.TrimSpace(query)
	if query == "" {
		return []Candidate{}, nil
	}
	base := stripRegionSuffix(query)
	tables := []struct {
		level   string
		prio    int
		records []geoRecord
	}{
		{"city", 0, geo.Cities},
		{"county", 1, geo.Counties},
		{"world", 2, geo.World},
	}
	type scored struct {
		candidate Candidate
		class     int
		prio      int
	}
	var matches []scored
	for _, table := range tables {
		for _, record := range table.records {
			class, ok := candidateMatchClass(query, base, record)
			if !ok {
				continue
			}
			matches = append(matches, scored{
				candidate: Candidate{
					Name:      record.Name,
					Region:    record.Region,
					City:      record.City,
					Country:   record.Region,
					Level:     table.level,
					Longitude: record.Lng,
					Latitude:  record.Lat,
					Timezone:  record.TZ,
					BBox:      record.BBox,
				},
				class: class,
				prio:  table.prio,
			})
		}
	}
	sort.SliceStable(matches, func(i, j int) bool {
		a, b := matches[i], matches[j]
		if a.class != b.class {
			return a.class < b.class
		}
		if a.prio != b.prio {
			return a.prio < b.prio
		}
		if a.candidate.Name != b.candidate.Name {
			return a.candidate.Name < b.candidate.Name
		}
		if a.candidate.Region != b.candidate.Region {
			return a.candidate.Region < b.candidate.Region
		}
		return a.candidate.City < b.candidate.City
	})
	seen := make(map[string]bool, len(matches))
	out := make([]Candidate, 0, candidateLimit)
	for _, match := range matches {
		key := match.candidate.Name + "|" + match.candidate.Region + "|" +
			match.candidate.City + "|" + strconv.FormatFloat(match.candidate.Longitude, 'f', 6, 64)
		if seen[key] {
			continue
		}
		seen[key] = true
		out = append(out, match.candidate)
		if len(out) >= candidateLimit {
			break
		}
	}
	return out, nil
}

// candidateMatchClass ranks record against query; lower is a better match.
func candidateMatchClass(query, base string, record geoRecord) (int, bool) {
	if class, ok := nameMatchClass(query, base, record.Name, record); ok {
		return class, true
	}
	for _, alias := range record.Aliases {
		if class, ok := nameMatchClass(query, base, alias, record); ok {
			return class, true
		}
	}
	return 0, false
}

func nameMatchClass(query, base, name string, record geoRecord) (int, bool) {
	if name == query {
		return 0, true
	}
	if compoundMatches(query, record) {
		return 1, true
	}
	if base != "" && stripRegionSuffix(name) == base {
		return 2, true
	}
	if strings.HasPrefix(name, query) {
		return 3, true
	}
	if base != "" && strings.Contains(name, base) {
		return 4, true
	}
	return 0, false
}

// compoundMatches reports whether query is a region/city prefix plus the
// place name, e.g. 北京朝阳区, 湖南长沙, 杭州西湖区.
func compoundMatches(query string, record geoRecord) bool {
	runes := []rune(query)
	for i := 1; i < len(runes); i++ {
		prefix := string(runes[:i])
		rest := string(runes[i:])
		if record.Name != rest && stripRegionSuffix(record.Name) != stripRegionSuffix(rest) {
			continue
		}
		if strings.HasPrefix(record.Region, prefix) ||
			(record.City != "" && strings.HasPrefix(record.City, prefix)) {
			return true
		}
	}
	return false
}

// resolveRecords 唯一记录直接返回；多条同名 fail-closed 并给出省/市限定示例。
func resolveRecords(query string, records []geoRecord) (Coords, bool, error) {
	if len(records) == 1 {
		return resultFrom(records[0]), true, nil
	}
	labels := make([]string, 0, len(records))
	for _, record := range records {
		labels = append(labels, recordLabel(record))
	}
	sort.Strings(labels)
	return Coords{}, false, fmt.Errorf(
		"城市名 '%s' 存在歧义：%s；请使用“省/市+名称”（如 '%s'）或直接提供经纬度",
		query, strings.Join(labels, "、"), compoundExample(records),
	)
}

// recordLabel 输出“省(+市)+区县名”，供歧义候选展示。
func recordLabel(record geoRecord) string {
	if record.City != "" {
		return record.Region + record.City + record.Name
	}
	return record.Region + record.Name
}

// compoundExample 从候选推导消歧示例：优先市名（“南京鼓楼区”），否则省名简称。
func compoundExample(records []geoRecord) string {
	record := records[0]
	if record.City != "" {
		return strings.TrimSuffix(record.City, "市") + record.Name
	}
	return shortRegion(record.Region) + record.Name
}

// shortRegion 去掉省级后缀得到可查询前缀（北京市→北京、内蒙古自治区→内蒙古）。
func shortRegion(region string) string {
	for _, suffix := range []string{"维吾尔自治区", "壮族自治区", "回族自治区", "自治区", "特别行政区", "省", "市"} {
		if trimmed := strings.TrimSuffix(region, suffix); trimmed != region {
			return trimmed
		}
	}
	return region
}

// searchCompound 解析“省/市 + 名称”复合查询（如“北京朝阳区”“杭州西湖区”）：
// 取可识别的名称后缀，用前缀匹配候选的市/省限定；无法唯一确定时继续回退。
func searchCompound(query string) (Coords, bool, error) {
	runes := []rune(query)
	for i := 1; i < len(runes); i++ {
		key := string(runes[i:])
		records, ok := geoIndex[key]
		if !ok {
			continue
		}
		filtered := filterByPrefix(records, string(runes[:i]))
		switch len(filtered) {
		case 1:
			return resultFrom(filtered[0]), true, nil
		case 0:
			continue
		default:
			return resolveRecords(query, filtered)
		}
	}
	return Coords{}, false, nil
}

// filterByPrefix 按省/市前缀过滤候选：市名前缀（南京→南京市）优先，其次省名前缀。
func filterByPrefix(records []geoRecord, prefix string) []geoRecord {
	hint := ""
	if recs := geoIndex[prefix]; len(recs) > 0 {
		hint = recs[0].Region
	} else if recs := geoIndex[prefix+"市"]; len(recs) > 0 {
		hint = recs[0].Region
	}
	var out []geoRecord
	for _, record := range records {
		if record.City != "" && strings.HasPrefix(record.City, prefix) {
			out = append(out, record)
			continue
		}
		if strings.HasPrefix(record.Region, prefix) || strings.HasPrefix(prefix, record.Region) {
			out = append(out, record)
			continue
		}
		if hint != "" && record.Region == hint && record.City == "" {
			out = append(out, record)
		}
	}
	return out
}

// suffixCandidates 收集后缀回退候选（name+后缀 与 去后缀同名），按记录去重。
func suffixCandidates(name string) []geoRecord {
	seen := make(map[string]bool)
	var out []geoRecord
	add := func(records []geoRecord) {
		for _, record := range records {
			label := recordLabel(record)
			if seen[label] {
				continue
			}
			seen[label] = true
			out = append(out, record)
		}
	}
	for _, suffix := range regionSuffixes {
		add(geoIndex[name+suffix])
	}
	if base := stripRegionSuffix(name); base != "" {
		for key, records := range geoIndex {
			if stripRegionSuffix(key) == base {
				add(records)
			}
		}
	}
	return out
}

// stripRegionSuffix removes trailing administrative-division suffixes.
func stripRegionSuffix(s string) string {
	for _, suf := range regionSuffixes {
		s = strings.TrimSuffix(s, suf)
	}
	return s
}

func resultFrom(record geoRecord) Coords {
	return searchResult{
		Name: record.Name, Longitude: record.Lng, Latitude: record.Lat,
		Country: record.Region, Timezone: record.TZ, BBox: record.BBox,
	}
}

func searchNominatim(ctx context.Context, query string) (Coords, error) {
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
		return Coords{}, fmt.Errorf("search: new request: %w", err)
	}
	req.Header.Set("User-Agent", "Liki-Engine (https://liki.hk; contact: api@liki.hk)")

	resp, err := httpClient.Do(req)
	if err != nil {
		return Coords{}, fmt.Errorf("search: get: %w", err)
	}
	defer func() { _ = resp.Body.Close() }()

	if resp.StatusCode != http.StatusOK {
		return Coords{}, fmt.Errorf("search: status %d", resp.StatusCode)
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
		return Coords{}, fmt.Errorf("search: decode: %w", err)
	}
	if len(results) == 0 {
		return Coords{}, fmt.Errorf("search: no results for %s", query)
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
		return Coords{}, fmt.Errorf("同名行政区跨省歧义，请提供省/市限定（如 '辽宁朝阳'）：%s", query)
	}
	if len(admins) == 0 {
		admins = []int{0}
	}
	r := results[admins[0]]
	lon, err := parseFloat(r.Lon)
	if err != nil {
		return Coords{}, fmt.Errorf("search: parse lon: %w", err)
	}
	lat, err := parseFloat(r.Lat)
	if err != nil {
		return Coords{}, fmt.Errorf("search: parse lat: %w", err)
	}
	return Coords{
		Name:      r.Name,
		Longitude: lon,
		Latitude:  lat,
		Country:   countryToZh(r.Address.Country),
	}, nil
}

func parseFloat(s string) (float64, error) {
	f, err := strconv.ParseFloat(strings.TrimSpace(s), 64)
	if err != nil {
		return 0, fmt.Errorf("parseFloat: %q: %w", s, err)
	}
	return f, nil
}
