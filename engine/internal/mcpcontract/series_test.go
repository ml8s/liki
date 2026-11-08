package mcpcontract

import (
	"context"
	"encoding/json"
	"testing"

	"liki-engine/internal/engine/ganzhi"
	"liki-engine/internal/engine/tianwen"
)

func civilDateForLunar(t *testing.T, year, month, day int, leap bool) string {
	t.Helper()
	date, err := tianwen.ShanghaiLunarReference().CivilDateFromLunar(tianwen.LunarDate{
		Year: year, Month: month, Day: day, Leap: leap,
	})
	if err != nil {
		t.Fatalf("lunar %d-%d-%d leap=%v has no civil date: %v", year, month, day, leap, err)
	}
	return date.String()
}

func seriesChart(t *testing.T) map[string]any {
	t.Helper()
	return testBaziChart(t, "1984-02-15T00:00:00Z", "Asia/Shanghai", 116.4, "male")
}

var namedSeriesRoots = []string{"liu_nian", "liu_yue", "liu_ri", "liu_shi", "xiao_yun"}

func decodeSeriesItems(t *testing.T, raw json.RawMessage) []map[string]any {
	t.Helper()
	var document map[string]any
	decode(t, raw, &document)
	for _, root := range namedSeriesRoots {
		if items, ok := document[root].([]any); ok {
			out := make([]map[string]any, 0, len(items))
			for _, item := range items {
				object, ok := item.(map[string]any)
				if !ok {
					t.Fatalf("%s series item is not an object: %#v", root, item)
				}
				out = append(out, object)
			}
			return out
		}
	}
	t.Fatalf("series document has no named domain root: %#v", document)
	return nil
}

func firstSeriesItem(t *testing.T, document map[string]any) map[string]any {
	t.Helper()
	for _, root := range namedSeriesRoots {
		if items, ok := document[root].([]any); ok && len(items) > 0 {
			item, ok := items[0].(map[string]any)
			if !ok {
				t.Fatalf("%s series item is not an object: %#v", root, items[0])
			}
			return item
		}
	}
	t.Fatalf("series document has no named domain root: %#v", document)
	return nil
}

func callSeries(t *testing.T, reg *Registry, tool string, args map[string]any) json.RawMessage {
	t.Helper()
	raw, err := reg.Call(context.Background(), tool, mustJSON(t, args))
	if err != nil {
		t.Fatalf("%s: %v", tool, err)
	}
	if err := reg.ValidateOutput(tool, raw); err != nil {
		t.Fatalf("%s output contract: %v", tool, err)
	}
	return raw
}

func TestSeriesSingleIsDefault(t *testing.T) {
	reg := NewEngine()
	raw := callSeries(t, reg, "bazi_liupan", map[string]any{
		"subject": seriesChart(t), "scope": "liu_nian",
		"standard_time": "2026-06-15T12:00:00+08:00",
	})
	items := decodeSeriesItems(t, raw)
	if len(items) != 1 {
		t.Fatalf("got %d items, want 1", len(items))
	}
	if _, ok := items[0]["year"]; !ok {
		t.Fatal("missing year field")
	}
}

func TestSeriesBatchN(t *testing.T) {
	reg := NewEngine()
	raw := callSeries(t, reg, "bazi_liupan", map[string]any{
		"subject": seriesChart(t), "scope": "liu_nian",
		"standard_time": "2026-06-15T12:00:00+08:00", "count": 3,
	})
	items := decodeSeriesItems(t, raw)
	if len(items) != 3 {
		t.Fatalf("want 3 items, got %d", len(items))
	}
	for i, item := range items {
		y, _ := item["year"].(float64)
		if int(y) != 2026+i {
			t.Errorf("items[%d].year=%v, want %d", i, y, 2026+i)
		}
	}
}

func TestSeriesBatchFirstEqualsSingle(t *testing.T) {
	reg := NewEngine()
	args := func(count int) map[string]any {
		return map[string]any{
			"subject": seriesChart(t), "scope": "liu_nian",
			"standard_time": "2026-06-15T12:00:00+08:00", "count": count,
		}
	}
	var s BaziLiuNianSeries
	decode(t, callSeries(t, reg, "bazi_liupan", args(1)), &s)
	var b BaziLiuNianSeries
	decode(t, callSeries(t, reg, "bazi_liupan", args(3)), &b)
	if len(s.LiuNian) == 0 || len(b.LiuNian) == 0 {
		t.Fatal("empty batch")
	}
	if b.LiuNian[0].Year != s.LiuNian[0].Year || b.LiuNian[0].ShiShen != s.LiuNian[0].ShiShen {
		t.Errorf("batch[0] != single: year %d/%d shishen %s/%s",
			b.LiuNian[0].Year, s.LiuNian[0].Year, b.LiuNian[0].ShiShen, s.LiuNian[0].ShiShen)
	}
}

func TestSeriesCountBounds(t *testing.T) {
	reg := NewEngine()
	for _, count := range []int{-1, MaxSeriesCount + 1} {
		if _, err := reg.Call(context.Background(), "bazi_liupan", mustJSON(t, map[string]any{
			"subject": seriesChart(t), "scope": "liu_nian",
			"standard_time": "2026-06-15T12:00:00+08:00", "count": count,
		})); err == nil {
			t.Errorf("count=%d should error", count)
		}
	}
}

func TestSeriesCountSchemaMatchesRuntimeBound(t *testing.T) {
	registry := NewEngine()
	tools := []string{"bazi_liupan", "ziwei_liupan", "xuankong_liunian"}
	for _, tool := range tools {
		t.Run(tool, func(t *testing.T) {
			definition, err := registry.Definition(tool)
			if err != nil {
				t.Fatal(err)
			}
			var schema struct {
				Properties map[string]struct {
					Minimum *float64 `json:"minimum"`
					Maximum *float64 `json:"maximum"`
				} `json:"properties"`
				OneOf []struct {
					Properties map[string]struct {
						Minimum *float64 `json:"minimum"`
						Maximum *float64 `json:"maximum"`
					} `json:"properties"`
				} `json:"oneOf"`
			}
			decode(t, definition.InputSchema, &schema)
			var minimum, maximum *float64
			check := func(value struct {
				Minimum *float64 `json:"minimum"`
				Maximum *float64 `json:"maximum"`
			}) {
				if value.Minimum != nil {
					minimum, maximum = value.Minimum, value.Maximum
				}
			}
			check(schema.Properties["count"])
			if minimum == nil {
				check(schema.Properties["count"])
			}
			if minimum == nil {
				for _, branch := range schema.OneOf {
					check(branch.Properties["count"])
					check(branch.Properties["count"])
					if minimum != nil {
						break
					}
				}
			}
			if minimum == nil || *minimum != 1 || maximum == nil || *maximum != MaxSeriesCount {
				t.Fatalf("count bounds = %#v/%#v, want [1,%d]", minimum, maximum, MaxSeriesCount)
			}
		})
	}
}

func TestSeriesRollovers(t *testing.T) {
	reg := NewEngine()
	baziChart := seriesChart(t)

	months := decodeSeriesItems(t, callSeries(t, reg, "bazi_liupan", map[string]any{
		"subject": baziChart, "scope": "liu_yue",
		"standard_time": "2026-11-15T12:00:00+08:00", "count": 3,
	}))
	prevEnd := ""
	for i, item := range months {
		resolved, ok := item["jie_qi"].(map[string]any)
		if !ok {
			t.Fatalf("months[%d]: missing resolved", i)
		}
		start, _ := resolved["jieqi_start"].(string)
		end, _ := resolved["jieqi_end"].(string)
		if start == "" || end == "" {
			t.Fatalf("months[%d]: missing jieqi boundary", i)
		}
		if i > 0 && start != prevEnd {
			t.Errorf("months[%d].jieqi_start=%s, want previous end %s", i, start, prevEnd)
		}
		prevEnd = end
	}

	hours := decodeSeriesItems(t, callSeries(t, reg, "bazi_liupan", map[string]any{
		"subject": baziChart, "scope": "liu_shi",
		"standard_time": "2026-01-01T23:00:00+08:00", "longitude": 116.4, "count": 2,
	}))
	if len(hours) != 2 {
		t.Fatalf("want 2 hours, got %d", len(hours))
	}
	if hourRange, ok := hours[0]["hour_range"].(string); !ok || hourRange == "" {
		t.Fatalf("bazi liu shi hour_range = %#v, want a non-empty time range", hours[0]["hour_range"])
	}
	first, _ := hours[0]["shi_chen"].(map[string]any)["shichen"].(string)
	second, _ := hours[1]["shi_chen"].(map[string]any)["shichen"].(string)
	firstZhi, err := ganzhi.ParseZhi(first)
	if err != nil {
		t.Fatal(err)
	}
	secondZhi, err := ganzhi.ParseZhi(second)
	if err != nil {
		t.Fatal(err)
	}
	if secondZhi != firstZhi%12+1 {
		t.Errorf("时辰 step = %s -> %s, want the next shichen", first, second)
	}
}

func TestSeriesApparentSolarCellsDoNotRepeat(t *testing.T) {
	reg := NewEngine()
	baziChart := seriesChart(t)

	days := decodeSeriesItems(t, callSeries(t, reg, "bazi_liupan", map[string]any{
		"subject": baziChart, "scope": "liu_ri",
		"standard_time": "2024-01-01T22:52:00Z", "longitude": 18, "count": 2,
	}))
	if len(days) != 2 {
		t.Fatalf("bazi liu_ri returned %d items, want 2", len(days))
	}
	firstDay := days[0]["ri_jie"].(map[string]any)["solar_date"].(string)
	secondDay := days[1]["ri_jie"].(map[string]any)["solar_date"].(string)
	if firstDay == secondDay {
		t.Fatalf("bazi liu_ri repeated LAST date %s", firstDay)
	}

	if got := days[0]["ri_jie"].(map[string]any)["standard_time"].(string); got != "2024-01-01T22:52:00Z" {
		t.Fatalf("bazi liu_ri first selector = %s, want the input selector", got)
	}

	hours := decodeSeriesItems(t, callSeries(t, reg, "bazi_liupan", map[string]any{
		"subject": baziChart, "scope": "liu_shi",
		"standard_time": "2024-01-02T20:28:00Z", "longitude": -171, "count": 2,
	}))
	if len(hours) != 2 {
		t.Fatalf("bazi liu_shi returned %d items, want 2", len(hours))
	}
	if got := hours[0]["shi_chen"].(map[string]any)["standard_time"].(string); got != "2024-01-02T20:28:00Z" {
		t.Fatalf("bazi liu_shi first selector = %s, want the input selector", got)
	}
	if first := hours[0]["shi_chen"].(map[string]any)["shichen"].(string); first != "巳" {
		t.Fatalf("first shichen = %s, want 巳", first)
	}
	if second := hours[1]["shi_chen"].(map[string]any)["shichen"].(string); second != "午" {
		t.Fatalf("second shichen = %s, want 午", second)
	}

	ziweiChart := map[string]any{
		"standard_time": "2024-01-02T20:28:00Z", "longitude": -171, "gender": "male",
	}
	ziweiHours := decodeSeriesItems(t, callSeries(t, reg, "ziwei_liupan", map[string]any{
		"subject": ziweiChart, "scope": "liu_shi",
		"standard_time": "2024-01-02T20:28:00Z", "longitude": -171, "count": 2,
	}))
	if len(ziweiHours) != 2 {
		t.Fatalf("ziwei liu_shi returned %d items, want 2", len(ziweiHours))
	}
	if first := ziweiHours[0]["shi_chen"].(map[string]any)["shichen"].(string); first != "巳" {
		t.Fatalf("first ziwei shichen = %s, want 巳", first)
	}
	if second := ziweiHours[1]["shi_chen"].(map[string]any)["shichen"].(string); second != "午" {
		t.Fatalf("second ziwei shichen = %s, want 午", second)
	}
}

func TestSeriesXiaoYunDefaultOne(t *testing.T) {
	reg := NewEngine()
	raw := callSeries(t, reg, "bazi_liupan", map[string]any{
		"subject": seriesChart(t), "scope": "xiao_yun",
	})
	var out BaziXiaoYunOutput
	decode(t, raw, &out)
	for _, set := range out.XiaoYun {
		if len(set.Zhus) != 1 {
			t.Errorf("xiaoyun default count=1: got %d zhus", len(set.Zhus))
		}
	}
}

func TestSeriesBatchOutputContract(t *testing.T) {
	reg := NewEngine()
	baziChart := seriesChart(t)
	ziweiChart := map[string]any{
		"standard_time": "1984-02-16T12:00:00+08:00", "longitude": 116.4, "gender": "male",
	}
	ziweiDate := civilDateForLunar(t, 2026, 5, 10, false)
	ziweiCivilTime := ziweiDate + "T11:00:00+08:00"

	cases := []struct {
		tool string
		args map[string]any
	}{
		{"bazi_liupan", map[string]any{"subject": baziChart, "scope": "liu_nian", "standard_time": "2026-06-15T12:00:00+08:00", "count": 2}},
		{"bazi_liupan", map[string]any{"subject": baziChart, "scope": "liu_yue", "standard_time": "2026-11-15T12:00:00+08:00", "count": 2}},
		{"bazi_liupan", map[string]any{"subject": baziChart, "scope": "liu_ri", "standard_time": "2026-01-01T12:00:00+08:00", "longitude": 116.4, "count": 2}},
		{"bazi_liupan", map[string]any{"subject": baziChart, "scope": "liu_shi", "standard_time": "2026-01-01T23:00:00+08:00", "longitude": 116.4, "count": 2}},
		{"ziwei_liupan", map[string]any{"subject": ziweiChart, "scope": "liu_nian", "standard_time": "2026-06-15T12:00:00+08:00", "count": 2}},
		{"ziwei_liupan", map[string]any{"subject": ziweiChart, "scope": "liu_yue", "standard_time": ziweiDate + "T12:00:00+08:00", "count": 2}},
		{"ziwei_liupan", map[string]any{"subject": ziweiChart, "scope": "liu_ri", "standard_time": ziweiDate + "T12:00:00+08:00", "count": 2}},
		{"ziwei_liupan", map[string]any{"subject": ziweiChart, "scope": "liu_shi", "standard_time": ziweiCivilTime, "longitude": 116.4, "count": 2}},
		{"xuankong_liunian", map[string]any{"standard_time": "2026-06-15T12:00:00+08:00", "zuo_shan": 2, "xiang_shan": 14, "count": 2}},
	}

	for _, tc := range cases {
		t.Run(tc.tool+"/batch", func(t *testing.T) {
			raw := callSeries(t, reg, tc.tool, tc.args)
			items := decodeSeriesItems(t, raw)
			if len(items) < 2 {
				t.Fatalf("%s: batch output has %d items, want at least 2", tc.tool, len(items))
			}
		})
	}
}
