package city

import (
	"context"
	"fmt"
	"strconv"
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

func TestBuiltin_CountyExact(t *testing.T) {
	// 抚远市（佳木斯下辖，中国最东端）应精确命中内置表
	candidates, err := ResolveCandidates(context.Background(), "抚远市")
	if err != nil {
		t.Fatalf("ResolveCandidates: %v", err)
	}
	if len(candidates) == 0 {
		t.Fatal("no candidates")
	}
	r := candidates[0]
	if got, want := r.Longitude, 134.294501; got != want {
		t.Errorf("longitude = %v, want %v", got, want)
	}
	if r.Country == "" {
		t.Error("country should be populated from embedded region")
	}
}

func TestBuiltin_CountySuffixMatch(t *testing.T) {
	// 不带"市"后缀应回退命中县级条目
	candidates, err := ResolveCandidates(context.Background(), "抚远")
	if err != nil {
		t.Fatalf("ResolveCandidates: %v", err)
	}
	if len(candidates) == 0 {
		t.Fatal("no candidates")
	}
	if got, want := candidates[0].Longitude, 134.294501; got != want {
		t.Errorf("longitude = %v, want %v", got, want)
	}
}

func TestBuiltin_PrefectureFallback(t *testing.T) {
	// 只给地级市"佳木斯" → 命中 cities 代表坐标（向阳区）
	candidates, err := ResolveCandidates(context.Background(), "佳木斯")
	if err != nil {
		t.Fatalf("ResolveCandidates: %v", err)
	}
	if len(candidates) == 0 {
		t.Fatal("no candidates")
	}
	if got, want := candidates[0].Longitude, 130.361634; got != want {
		t.Errorf("longitude = %v, want %v", got, want)
	}
}

func TestBuiltin_WorldCity(t *testing.T) {
	candidates, err := ResolveCandidates(context.Background(), "东京")
	if err != nil {
		t.Fatalf("ResolveCandidates: %v", err)
	}
	if len(candidates) == 0 {
		t.Fatal("no candidates")
	}
	if got, want := candidates[0].Longitude, 139.692; got != want {
		t.Errorf("longitude = %v, want %v", got, want)
	}
	if got, want := candidates[0].Country, "日本"; got != want {
		t.Errorf("country = %q, want %q", got, want)
	}
}

func TestBuiltin_ConflictingSuffix(t *testing.T) {
	// 输入带错误后缀（表里是"抚远市"，输入"抚远区"）→ 去后缀回退应命中
	candidates, err := ResolveCandidates(context.Background(), "抚远区")
	if err != nil {
		t.Fatalf("ResolveCandidates: %v", err)
	}
	if len(candidates) == 0 {
		t.Fatal("no candidates")
	}
	if got, want := candidates[0].Longitude, 134.294501; got != want {
		t.Errorf("longitude = %v, want %v", got, want)
	}
}

func TestBuiltin_Municipality(t *testing.T) {
	// 直辖市："北京" → 应命中 cities["北京市"]
	candidates, err := ResolveCandidates(context.Background(), "北京")
	if err != nil {
		t.Fatalf("ResolveCandidates: %v", err)
	}
	if len(candidates) == 0 {
		t.Fatal("no candidates")
	}
	if got, want := candidates[0].Longitude, 116.4074; got < want-0.5 || got > want+0.5 {
		t.Errorf("longitude = %v, want near %v (北京)", got, want)
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

func parseFloat(s string) (float64, error) {
	f, err := strconv.ParseFloat(strings.TrimSpace(s), 64)
	if err != nil {
		return 0, fmt.Errorf("parseFloat: %q: %w", s, err)
	}
	return f, nil
}
