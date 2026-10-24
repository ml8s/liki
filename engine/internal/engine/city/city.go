package city

import (
	"context"
	_ "embed"
	"encoding/json"
	"fmt"
	"sort"
	"strconv"
	"strings"
	"sync"
)

//go:embed cities_data.json
var citiesData []byte

// httpDoer is the interface for HTTP clients, allowing test injection.

// SetHTTPClient replaces the HTTP client used for Nominatim queries. Call from
// tests to inject a mock transport.
// HttpClient returns the current HTTP client. Useful for save/restore in tests.
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

// Candidate is one possible place match. Candidate mode never guesses: callers
// receive the ranked list and pick the intended place themselves.
type Candidate struct {
	Name        string            `json:"name"`
	Names       map[string]string `json:"names,omitempty"`
	Region      string            `json:"region"`
	RegionNames map[string]string `json:"region_names,omitempty"`
	City        string            `json:"city,omitempty"`
	Country     string            `json:"country"`
	Level       string            `json:"level"`
	Longitude   float64           `json:"longitude"`
	Latitude    float64           `json:"latitude"`
	Timezone    string            `json:"timezone,omitempty"`
	BBox        *BBox             `json:"bbox,omitempty"`
}

// candidateLimit caps how many candidates a single query may return.
const candidateLimit = 20

// geoRecord 是内嵌地理表的一条记录。县级表存在跨省同名（如朝阳区：北京/长春），
// 因此名称不是唯一键：同名记录靠 Region（省）与 City（地级市，可空）消歧。
type geoRecord struct {
	Name        string            `json:"name"`
	Names       map[string]string `json:"names,omitempty"`
	Region      string            `json:"region"`
	RegionNames map[string]string `json:"region_names,omitempty"`
	City        string            `json:"city,omitempty"`
	Lng         float64           `json:"lng"`
	Lat         float64           `json:"lat"`
	TZ          string            `json:"tz,omitempty"`
	BBox        *BBox             `json:"bbox,omitempty"`
	Aliases     []string          `json:"aliases,omitempty"`
}

type geoDB struct {
	Counties []geoRecord `json:"counties"`
	Cities   []geoRecord `json:"cities"`
	World    []geoRecord `json:"world"`
}

var (
	geoOnce  sync.Once
	geo      geoDB
	geoIndex map[string][]geoRecord
)

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

// SearchCoords resolves a city name to coordinates. The embedded table
// (counties/cities/world, WGS-84) is checked first; Nominatim (OSM) is used
// as a fallback for long-tail places.
// ResolveCoords resolves a city name to typed coordinates. The embedded table
// (counties/cities/world, WGS-84) is checked first; Nominatim (OSM) is used
// as a fallback for long-tail places.
// searchBuiltin resolves a place name against the embedded table.
//
// 唯一名直接命中；同名多条（跨省同名区县，如朝阳区：北京/长春）必须显式消歧：
// 先尝试“省/市 + 名称”复合形式（如“北京朝阳区”“南京鼓楼区”），无法消歧时
// fail-closed 列出候选，绝不静默返回某一个同名地点的坐标。
// 后缀回退（name+后缀、去后缀同名）同样按记录消歧。
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
					Name:        record.Name,
					Names:       record.Names,
					Region:      record.Region,
					RegionNames: record.RegionNames,
					City:        record.City,
					Country:     record.Region,
					Level:       table.level,
					Longitude:   record.Lng,
					Latitude:    record.Lat,
					Timezone:    record.TZ,
					BBox:        record.BBox,
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
// recordLabel 输出“省(+市)+区县名”，供歧义候选展示。
// compoundExample 从候选推导消歧示例：优先市名（“南京鼓楼区”），否则省名简称。
// shortRegion 去掉省级后缀得到可查询前缀（北京市→北京、内蒙古自治区→内蒙古）。
// searchCompound 解析“省/市 + 名称”复合查询（如“北京朝阳区”“杭州西湖区”）：
// 取可识别的名称后缀，用前缀匹配候选的市/省限定；无法唯一确定时继续回退。
// filterByPrefix 按省/市前缀过滤候选：市名前缀（南京→南京市）优先，其次省名前缀。
// suffixCandidates 收集后缀回退候选（name+后缀 与 去后缀同名），按记录去重。
// stripRegionSuffix removes trailing administrative-division suffixes.
func stripRegionSuffix(s string) string {
	for _, suf := range regionSuffixes {
		s = strings.TrimSuffix(s, suf)
	}
	return s
}
