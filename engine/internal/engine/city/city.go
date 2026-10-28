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

// Coords is the resolved place result. It is the typed domain boundary; MCP
// applies its own public DTO on top of it.
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
	records  []indexedRecord
	index    map[string]map[int]struct{}
)

type indexedRecord struct {
	tableIndex int
	record     geoRecord
	identity   string
}

func loadGeo() {
	if err := json.Unmarshal(citiesData, &geo); err != nil {
		panic("city: load embedded data: " + err.Error())
	}
	geoIndex = make(map[string][]geoRecord, len(geo.Counties)+len(geo.Cities)+len(geo.World))
	// cities 与 counties 在少数地级市上存在完全相同的记录（同名同省同坐标），
	// 合并索引时按记录身份去重，避免把同一地点报成歧义。
	seen := make(map[string]struct{}, len(geo.Counties)+len(geo.Cities)+len(geo.World))
	for tableIndex, table := range [][]geoRecord{geo.Cities, geo.Counties, geo.World} {
		for _, record := range table {
			key := fmt.Sprintf("%s|%s|%s|%.6f|%.6f", record.Name, record.Region, record.City, record.Lng, record.Lat)
			if _, ok := seen[key]; ok {
				continue
			}
			seen[key] = struct{}{}
			identity := strconv.Itoa(tableIndex) + "|" + key
			records = append(records, indexedRecord{
				tableIndex: tableIndex,
				record:     record,
				identity:   identity,
			})
			geoIndex[record.Name] = append(geoIndex[record.Name], record)
			for _, alias := range record.Aliases {
				if alias == "" || alias == record.Name {
					continue
				}
				geoIndex[alias] = append(geoIndex[alias], record)
			}
		}
	}
	index = make(map[string]map[int]struct{})
	for id, item := range records {
		for text := range searchTextKeys(item.record) {
			candidates, ok := index[text]
			if !ok {
				candidates = make(map[int]struct{})
				index[text] = candidates
			}
			candidates[id] = struct{}{}
		}
	}
}

// regionSuffixes are administrative-division suffixes stripped for fuzzy match.
var regionSuffixes = []string{"市", "区", "县", "旗", "盟", "州"}

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
	query = strings.ToLower(query)
	base := stripRegionSuffix(query)
	type scored struct {
		candidate Candidate
		class     int
		prio      int
	}
	var matches []scored
	for id := range candidateIDs(query, base) {
		item := records[id]
		record := item.record
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
				Level:       tableNames[item.tableIndex],
				Longitude:   record.Lng,
				Latitude:    record.Lat,
				Timezone:    record.TZ,
				BBox:        record.BBox,
			},
			class: class,
			prio:  tablePrios[item.tableIndex],
		})
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

var tableNames = []string{"city", "county", "world"}
var tablePrios = []int{0, 1, 2}

func candidateIDs(query, base string) map[int]struct{} {
	ids := make(map[int]struct{})
	add := func(key string) {
		for id := range index[key] {
			ids[id] = struct{}{}
		}
	}
	add(query)
	if base != "" {
		add(base)
	}
	runes := []rune(query)
	for length := 1; length < len(runes); length++ {
		add(string(runes[:length]))
		add(string(runes[length:]))
	}
	baseRunes := []rune(base)
	for length := 1; length < len(baseRunes); length++ {
		for start := 0; start+length <= len(baseRunes); start++ {
			add(string(baseRunes[start : start+length]))
		}
	}
	return ids
}

func searchTextKeys(record geoRecord) map[string]struct{} {
	keys := make(map[string]struct{})
	addText := func(text string) {
		text = strings.ToLower(strings.TrimSpace(text))
		if text == "" {
			return
		}
		runes := []rune(text)
		for start := 0; start < len(runes); start++ {
			for length := 1; length <= len(runes)-start; length++ {
				keys[string(runes[start:start+length])] = struct{}{}
			}
		}
	}
	addText(record.Name)
	addText(stripRegionSuffix(record.Name))
	for _, alias := range record.Aliases {
		addText(alias)
	}
	for locale := range record.Names {
		addText(record.Names[locale])
	}
	return keys
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
	name = strings.ToLower(name)
	if name == query {
		return 0, true
	}
	if compoundMatches(query, record) {
		return 1, true
	}
	if base != "" && stripRegionSuffix(name) == base {
		return 2, true
	}
	if isLatinPlace(query) {
		if latinWordMatch(name, query) {
			return 3, true
		}
		if base != "" && latinWordMatch(name, base) {
			return 4, true
		}
		return 0, false
	}
	if strings.HasPrefix(name, query) {
		return 3, true
	}
	if base != "" && strings.Contains(name, base) {
		return 4, true
	}
	return 0, false
}

func isLatinPlace(value string) bool {
	if value == "" {
		return false
	}
	for _, r := range value {
		if r > 127 {
			return false
		}
	}
	return true
}

func isASCIILetter(value byte) bool {
	return value >= 'a' && value <= 'z' || value >= 'A' && value <= 'Z'
}

func latinWordMatch(name, text string) bool {
	if text == "" {
		return false
	}
	for start := 0; start < len(name); {
		index := strings.Index(name[start:], text)
		if index < 0 {
			return false
		}
		index += start
		end := index + len(text)
		if (index == 0 || !isASCIILetter(name[index-1])) &&
			(end == len(name) || !isASCIILetter(name[end])) {
			return true
		}
		start = index + 1
	}
	return false
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
