package agent

import (
	"context"
	"encoding/json"
	"sort"
	"testing"
)

func TestAllMethodsSchema(t *testing.T) {
	reg := NewRPCRegistry()
	zc := mustJSON(t, map[string]any{
		"lunar":  map[string]any{"year": 1984, "month": 1, "day": 15, "shichen": "辰"},
		"gender": "male",
	})
	bc := mustJSON(t, map[string]any{
		"solar_time": "1984-02-15T08:00:00+08:00",
		"gender":     "male",
	})
	zr := executeAndDecode(t, reg, "ziwei.chart", zc)
	br := executeAndDecode(t, reg, "bazi.chart", bc)
	z, b := zr["data"].(map[string]any), br["data"].(map[string]any)
	qr := executeAndDecode(t, reg, "liuyao.qigua", []byte(`{"seed":12345}`))
	yaos := mustJSON(t, qr["data"].(map[string]any)["yaos"])
	xr := executeAndDecode(t, reg, "xuankong.chart", []byte(
		`{"solar_time":"2026-07-31T10:00:00+08:00","zuo_shan":2,"xiang_shan":8}`,
	))
	x := xr["data"].(map[string]any)

	calls := map[string]json.RawMessage{
		"bazi.chart":      bc,
		"bazi.fullchart":  mustJSON(t, map[string]any{"chart": b}),
		"bazi.liunian":    mustJSON(t, map[string]any{"chart": b, "year": 2026}),
		"bazi.liuyue":     mustJSON(t, map[string]any{"chart": b, "year": 2026, "month": 6}),
		"bazi.liuri":      mustJSON(t, map[string]any{"chart": b, "year": 2026, "month": 6, "day": 4}),
		"bazi.liushi":     mustJSON(t, map[string]any{"chart": b, "year": 2026, "month": 6, "day": 4, "hour": 12}),
		"bazi.xiaoyun":    mustJSON(t, map[string]any{"chart": b, "max_age": 12}),
		"bazi.bond":       mustJSON(t, map[string]any{"a": map[string]any{"chart": b}, "b": map[string]any{"chart": b}}),
		"ziwei.chart":     zc,
		"ziwei.fullchart": mustJSON(t, map[string]any{"chart": z}),
		"ziwei.daxian":    mustJSON(t, map[string]any{"chart": z}),
		"ziwei.liunian":   mustJSON(t, map[string]any{"chart": z, "lunar_year": 2026}),
		"ziwei.liuyue":    mustJSON(t, map[string]any{"chart": z, "lunar_year": 2026, "lunar_month": 6}),
		"ziwei.liuri":     mustJSON(t, map[string]any{"chart": z, "lunar_year": 2026, "lunar_month": 6, "lunar_day": 4}),
		"ziwei.liushi":    mustJSON(t, map[string]any{"chart": z, "lunar_year": 2026, "lunar_month": 6, "lunar_day": 4, "shi_zhi": "午"}),
		"ziwei.bond":      mustJSON(t, map[string]any{"a": z, "b": z}),
		"liuyao.qigua":    mustJSON(t, map[string]any{"seed": 12345}),
		"liuyao.chart":    []byte(`{"solar_time":"2026-07-31T10:00:00+08:00","yaos":` + string(yaos) + `}`),
		"qimen.chart":     []byte(`{"solar_time":"2026-07-31T10:00:00+08:00","kind":"shi"}`),
		"bazhai.chart":    mustJSON(t, map[string]any{"solar_time": "1984-02-15T08:00:00+08:00", "gender": "male"}),
		"bazhai.layout": mustJSON(t, map[string]any{
			"chart":      map[string]any{"solar_time": "1984-02-15T08:00:00+08:00", "gender": "male"},
			"door_gua":   "乾",
			"master_gua": "乾",
			"stove_gua":  "乾",
		}),
		"xuankong.chart":   mustJSON(t, map[string]any{"solar_time": "2026-07-31T10:00:00+08:00", "zuo_shan": 2, "xiang_shan": 8}),
		"xuankong.liunian": mustJSON(t, map[string]any{"chart": x, "year": 2026}),
		"huangli.days":     mustJSON(t, map[string]any{"start_date": "2026-08-01", "count": 2}),
		"tianwen.time":     mustJSON(t, map[string]any{"time": "1984-02-15T08:00:00+08:00", "longitude": 116.4, "latitude": 39.9}),
		"time.now":         mustJSON(t, map[string]any{}),
		"qiming.char":      mustJSON(t, map[string]any{"char": "明"}),
		"qiming.pick":      mustJSON(t, map[string]any{"wuxing1": "木", "wuxing2": "火", "count": 2}),
		"qiming.compose":   mustJSON(t, map[string]any{"first": []string{"德"}, "second": []string{"明"}}),
		"qiming.check":     mustJSON(t, map[string]any{"given_names": []string{"德明"}}),
	}

	schemas := resultDataSchemas(t, reg)
	names := make([]string, 0, len(calls))
	for name := range calls {
		names = append(names, name)
	}
	sort.Strings(names)
	for _, name := range names {
		output := executeAndDecode(t, reg, name, calls[name])
		assertSchemaAllows(t, name, output["data"], schemas[name], "data")
	}
}

func resultDataSchemas(t *testing.T, reg *RPCRegistry) map[string]any {
	t.Helper()
	var document struct {
		Methods []struct {
			Name   string `json:"name"`
			Result struct {
				Properties struct {
					Data json.RawMessage `json:"data"`
				} `json:"properties"`
			} `json:"result"`
		} `json:"methods"`
	}
	decodeJSON(t, reg.OpenRPCDocument(), &document)
	schemas := make(map[string]any, len(document.Methods))
	for _, method := range document.Methods {
		if len(method.Result.Properties.Data) == 0 {
			continue
		}
		var schema any
		decodeJSON(t, method.Result.Properties.Data, &schema)
		schemas[method.Name] = schema
	}
	return schemas
}

func assertSchemaAllows(t *testing.T, method string, value any, schema any, path string) {
	t.Helper()
	switch actual := value.(type) {
	case map[string]any:
		objectSchema, ok := schema.(map[string]any)
		if !ok {
			t.Errorf("%s %s: object output has non-object schema", method, path)
			return
		}
		properties, _ := objectSchema["properties"].(map[string]any)
		keys := make([]string, 0, len(actual))
		for key := range actual {
			keys = append(keys, key)
		}
		sort.Strings(keys)
		for _, key := range keys {
			child, declared := properties[key]
			if !declared {
				additional, exists := objectSchema["additionalProperties"]
				switch schema := additional.(type) {
				case map[string]any:
					child = schema
				case bool:
					if !schema {
						t.Errorf("%s %s.%s is not declared in schema", method, path, key)
						continue
					}
				default:
					if !exists {
						continue
					}
				}
			}
			assertSchemaAllows(t, method, actual[key], child, path+"."+key)
		}
	case []any:
		arraySchema, ok := schema.(map[string]any)
		if !ok {
			t.Errorf("%s %s: array output has non-array schema", method, path)
			return
		}
		items := arraySchema["items"]
		if items == nil || items == true {
			return
		}
		for _, item := range actual {
			assertSchemaAllows(t, method, item, items, path+"[]")
		}
	}
}

func executeAndDecode(t *testing.T, reg *RPCRegistry, method string, params json.RawMessage) map[string]any {
	t.Helper()
	output, err := reg.Execute(context.Background(), method, params)
	if err != nil {
		t.Fatalf("%s: %v", method, err)
	}
	var value map[string]any
	decodeJSON(t, output, &value)
	return value
}

func decodeJSON(t *testing.T, data []byte, value any) {
	t.Helper()
	if err := json.Unmarshal(data, value); err != nil {
		t.Fatalf("decode JSON: %v", err)
	}
}

func mustJSON(t *testing.T, value any) json.RawMessage {
	t.Helper()
	encoded, err := json.Marshal(value)
	if err != nil {
		t.Fatalf("encode JSON: %v", err)
	}
	return encoded
}
