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
	solar := tianwen.CivilTimeFromLunarDate(tianwen.LunarDate{
		Year: year, Month: month, Day: day, Leap: leap,
	})
	if solar.IsZero() {
		t.Fatalf("lunar %d-%d-%d leap=%v has no civil date", year, month, day, leap)
	}
	return tianwen.CivilDateOf(solar).String()
}

func seriesChart(t *testing.T, reg *Registry) map[string]any {
	t.Helper()
	return testBaziChart(t, reg, "1984-02-15T00:00:00Z", "Asia/Shanghai", 116.4, "male")
}

func decodeSeriesItems(t *testing.T, raw json.RawMessage) []map[string]any {
	t.Helper()
	var out struct {
		Items []map[string]any `json:"items"`
	}
	decode(t, raw, &out)
	return out.Items
}

// firstSeriesItem unwraps the object-rooted series document's first item.
func firstSeriesItem(t *testing.T, document map[string]any) map[string]any {
	t.Helper()
	items, ok := document["items"].([]any)
	if !ok || len(items) == 0 {
		t.Fatalf("series document has no items: %#v", document)
	}
	item, ok := items[0].(map[string]any)
	if !ok {
		t.Fatalf("series item is not an object: %#v", items[0])
	}
	return item
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
	chart := seriesChart(t, reg)
	for _, count := range []int{0, 1} {
		raw := callSeries(t, reg, "bazi_liunian", map[string]any{
			"standard_time": "2026-06-15T12:00:00+08:00", "count": count, "chart": chart,
		})
		items := decodeSeriesItems(t, raw)
		if len(items) != 1 {
			t.Errorf("count=%d: got %d items, want 1", count, len(items))
		}
		if _, ok := items[0]["year"]; !ok {
			t.Errorf("count=%d: missing year field", count)
		}
	}
}

func TestSeriesBatchN(t *testing.T) {
	reg := NewEngine()
	chart := seriesChart(t, reg)
	raw := callSeries(t, reg, "bazi_liunian", map[string]any{
		"standard_time": "2026-06-15T12:00:00+08:00", "count": 3, "chart": chart,
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
	chart := seriesChart(t, reg)
	single := callSeries(t, reg, "bazi_liunian", map[string]any{
		"standard_time": "2026-06-15T12:00:00+08:00", "count": 1, "chart": chart,
	})
	batch := callSeries(t, reg, "bazi_liunian", map[string]any{
		"standard_time": "2026-06-15T12:00:00+08:00", "count": 3, "chart": chart,
	})
	var s struct {
		Items []BaziLiuNianOutput `json:"items"`
	}
	decode(t, single, &s)
	var b struct {
		Items []BaziLiuNianOutput `json:"items"`
	}
	decode(t, batch, &b)
	if len(s.Items) == 0 || len(b.Items) == 0 {
		t.Fatal("empty batch")
	}
	if b.Items[0].Year != s.Items[0].Year || b.Items[0].ShiShen != s.Items[0].ShiShen {
		t.Errorf("batch[0] != single: year %d/%d shishen %s/%s",
			b.Items[0].Year, s.Items[0].Year, b.Items[0].ShiShen, s.Items[0].ShiShen)
	}
}

func TestSeriesCountBounds(t *testing.T) {
	reg := NewEngine()
	chart := seriesChart(t, reg)
	for _, count := range []int{-1, MaxSeriesCount + 1} {
		if _, err := reg.Call(context.Background(), "bazi_liunian", mustJSON(t, map[string]any{
			"standard_time": "2026-06-15T12:00:00+08:00", "count": count, "chart": chart,
		})); err == nil {
			t.Errorf("count=%d should error", count)
		}
	}
}

func TestSeriesCountSchemaMatchesRuntimeBound(t *testing.T) {
	registry := NewEngine()
	tools := []string{
		"bazi_liunian", "bazi_liuyue", "bazi_liuri", "bazi_liushi", "bazi_xiaoyun",
		"ziwei_liunian", "ziwei_liuyue", "ziwei_liuri", "ziwei_liushi",
		"xuankong_liunian",
	}
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
			}
			decode(t, definition.InputSchema, &schema)
			count, ok := schema.Properties["count"]
			if !ok {
				t.Fatal("count property is missing")
			}
			if count.Minimum == nil || *count.Minimum != 0 ||
				count.Maximum == nil || *count.Maximum != MaxSeriesCount {
				t.Fatalf("count bounds = %#v/%#v, want [0,%d]", count.Minimum, count.Maximum, MaxSeriesCount)
			}
		})
	}
}

func TestSeriesMonthRollover(t *testing.T) {
	reg := NewEngine()
	chart := seriesChart(t, reg)
	raw := callSeries(t, reg, "bazi_liuyue", map[string]any{
		"standard_time": "2026-11-15T12:00:00+08:00", "count": 3, "chart": chart,
	})
	items := decodeSeriesItems(t, raw)
	prevEnd := ""
	for i, item := range items {
		resolved, ok := item["resolved"].(map[string]any)
		if !ok {
			t.Fatalf("items[%d]: missing resolved", i)
		}
		start, _ := resolved["jieqi_start"].(string)
		end, _ := resolved["jieqi_end"].(string)
		if start == "" || end == "" {
			t.Fatalf("items[%d]: missing jieqi boundary", i)
		}
		if i > 0 && start != prevEnd {
			t.Errorf("items[%d].jieqi_start=%s, want previous end %s", i, start, prevEnd)
		}
		prevEnd = end
	}
}

func TestSeriesHourRollover(t *testing.T) {
	reg := NewEngine()
	chart := seriesChart(t, reg)
	raw := callSeries(t, reg, "bazi_liushi", map[string]any{
		"standard_time": "2026-01-01T23:00:00+08:00",
		"longitude":     116.4, "count": 2, "chart": chart,
	})
	items := decodeSeriesItems(t, raw)
	if len(items) != 2 {
		t.Fatalf("want 2, got %d", len(items))
	}
	first, _ := items[0]["resolved"].(map[string]any)["shichen"].(string)
	second, _ := items[1]["resolved"].(map[string]any)["shichen"].(string)
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
	baziChart := seriesChart(t, reg)

	days := decodeSeriesItems(t, callSeries(t, reg, "bazi_liuri", map[string]any{
		"standard_time": "2024-01-01T22:52:00Z",
		"longitude":     18,
		"count":         2,
		"chart":         baziChart,
	}))
	if len(days) != 2 {
		t.Fatalf("bazi_liuri returned %d items, want 2", len(days))
	}
	if got := days[0]["resolved"].(map[string]any)["standard_time"].(string); got != "2024-01-01T22:52:00Z" {
		t.Fatalf("bazi_liuri first selector = %s, want the input selector", got)
	}
	firstDay := days[0]["resolved"].(map[string]any)["last_date"].(string)
	secondDay := days[1]["resolved"].(map[string]any)["last_date"].(string)
	if firstDay == secondDay {
		t.Fatalf("bazi_liuri repeated LAST date %s", firstDay)
	}

	hours := decodeSeriesItems(t, callSeries(t, reg, "bazi_liushi", map[string]any{
		"standard_time": "2024-01-02T20:28:00Z",
		"longitude":     -171,
		"count":         2,
		"chart":         baziChart,
	}))
	if len(hours) != 2 {
		t.Fatalf("bazi_liushi returned %d items, want 2", len(hours))
	}
	if got := hours[0]["resolved"].(map[string]any)["standard_time"].(string); got != "2024-01-02T20:28:00Z" {
		t.Fatalf("bazi_liushi first selector = %s, want the input selector", got)
	}
	firstShichen := hours[0]["resolved"].(map[string]any)["shichen"].(string)
	secondShichen := hours[1]["resolved"].(map[string]any)["shichen"].(string)
	if firstShichen != "巳" || secondShichen != "午" {
		t.Fatalf("bazi_liushi shichen = %s -> %s, want 巳 -> 午", firstShichen, secondShichen)
	}

	ziweiChart := callSeries(t, reg, "ziwei_chart", map[string]any{
		"standard_time": "2024-01-02T20:28:00Z",
		"longitude":     -171,
		"gender":        "male",
	})
	ziweiHours := decodeSeriesItems(t, callSeries(t, reg, "ziwei_liushi", map[string]any{
		"standard_time": "2024-01-02T20:28:00Z",
		"longitude":     -171,
		"count":         2,
		"chart":         ziweiChart,
	}))
	if len(ziweiHours) != 2 {
		t.Fatalf("ziwei_liushi returned %d items, want 2", len(ziweiHours))
	}
	firstZiweiShichen := ziweiHours[0]["resolved"].(map[string]any)["shichen"].(string)
	secondZiweiShichen := ziweiHours[1]["resolved"].(map[string]any)["shichen"].(string)
	if firstZiweiShichen != "巳" || secondZiweiShichen != "午" {
		t.Fatalf("ziwei_liushi shichen = %s -> %s, want 巳 -> 午", firstZiweiShichen, secondZiweiShichen)
	}
}

func TestSeriesXiaoYunDefaultOne(t *testing.T) {
	reg := NewEngine()
	chart := seriesChart(t, reg)
	raw := callSeries(t, reg, "bazi_xiaoyun", map[string]any{
		"chart": chart,
	})
	var out BaziXiaoYunOutput
	decode(t, raw, &out)
	for _, set := range out.Items {
		if len(set.Zhus) != 1 {
			t.Errorf("xiaoyun default count=1: got %d zhus", len(set.Zhus))
		}
	}
}

func TestSeriesBatchOutputContractForEveryCountTool(t *testing.T) {
	reg := NewEngine()
	baziChart := seriesChart(t, reg)
	ziweiChart := callSeries(t, reg, "ziwei_chart", map[string]any{
		"standard_time": "1984-02-16T12:00:00+08:00",
		"longitude":     116.4, "gender": "male",
	})
	xuankongChart := callSeries(t, reg, "xuankong_chart", map[string]any{
		"standard_time": "2026-07-31T12:00:00+08:00",
		"zuo_shan":      2, "xiang_shan": 14,
	})
	ziweiDate := civilDateForLunar(t, 2026, 5, 10, false)
	ziweiCivilTime := ziweiDate + "T11:00:00+08:00"

	cases := []struct {
		tool string
		args map[string]any
	}{
		{"bazi_liunian", map[string]any{"chart": baziChart, "standard_time": "2026-06-15T12:00:00+08:00", "count": 2}},
		{"bazi_liuyue", map[string]any{"chart": baziChart, "standard_time": "2026-11-15T12:00:00+08:00", "count": 2}},
		{"bazi_liuri", map[string]any{"chart": baziChart, "standard_time": "2026-01-01T12:00:00+08:00", "longitude": 116.4, "count": 2}},
		{"bazi_liushi", map[string]any{"chart": baziChart, "standard_time": "2026-01-01T23:00:00+08:00", "longitude": 116.4, "count": 2}},
		{"ziwei_liunian", map[string]any{"chart": ziweiChart, "standard_time": "2026-06-15T12:00:00+08:00", "count": 2}},
		{"ziwei_liuyue", map[string]any{"chart": ziweiChart, "standard_time": civilDateForLunar(t, 2026, 5, 10, false) + "T12:00:00+08:00", "count": 2}},
		{"ziwei_liuri", map[string]any{"chart": ziweiChart, "standard_time": civilDateForLunar(t, 2026, 5, 10, false) + "T12:00:00+08:00", "count": 2}},
		{"ziwei_liushi", map[string]any{"chart": ziweiChart, "standard_time": ziweiCivilTime, "longitude": 116.4, "count": 2}},
		{"xuankong_liunian", map[string]any{"chart": xuankongChart, "standard_time": "2026-06-15T12:00:00+08:00", "count": 2}},
	}

	for _, tc := range cases {
		t.Run(tc.tool, func(t *testing.T) {
			raw := callSeries(t, reg, tc.tool, tc.args)
			items := decodeSeriesItems(t, raw)
			if len(items) < 2 {
				t.Fatalf("%s: batch output has %d items, want at least 2", tc.tool, len(items))
			}
		})
	}
}
