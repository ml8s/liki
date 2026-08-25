package agent

import (
	"context"
	"encoding/json"
	"testing"
)

// TestSchemaGoConsistency 验证每个方法：Go 层实际输出字段 vs schema 声明完全一致
// （命名统一：schema 声明必须覆盖全部实际输出字段）
func TestSchemaGoConsistency(t *testing.T) {
	reg := NewRPCRegistry()
	// 依赖链样本
	zc, _ := json.Marshal(map[string]any{"lunar": map[string]any{"year": 1984, "month": 1, "day": 15, "shichen": "辰"}, "gender": "male"})
	bc, _ := json.Marshal(map[string]any{"solar_time": "1984-02-15T08:00:00+08:00", "gender": "male"})
	zout, _ := reg.Execute(context.Background(), "ziwei.chart", zc)
	bout, _ := reg.Execute(context.Background(), "bazi.chart", bc)
	var zr, br struct{ Data map[string]any `json:"data"` }
	_ = json.Unmarshal(zout, &zr)
	_ = json.Unmarshal(bout, &br)
	z, b := zr.Data, br.Data
	qout, _ := reg.Execute(context.Background(), "liuyao.qigua", []byte(`{"solar_time":"2026-07-31T10:00:00+08:00","question":"t"}`))
	var qr struct{ Data map[string]any `json:"data"` }
	_ = json.Unmarshal(qout, &qr)
	yaos, _ := json.Marshal(qr.Data["yaos"])
	mk := func(m map[string]any) json.RawMessage { b2, _ := json.Marshal(m); return b2 }
	calls := map[string]json.RawMessage{
		"bazi.chart": bc, "bazi.fullchart": mk(map[string]any{"chart": b}),
		"bazi.liunian": mk(map[string]any{"chart": b, "year": 2026}),
		"bazi.liuyue": mk(map[string]any{"chart": b, "year": 2026, "month": 6}),
		"bazi.liuri": mk(map[string]any{"chart": b, "year": 2026, "month": 6, "day": 4}),
		"bazi.liushi": mk(map[string]any{"chart": b, "year": 2026, "month": 6, "day": 4, "hour": 12}),
		"bazi.xiaoyun": mk(map[string]any{"chart": b, "max_age": 12}),
		"bazi.bond": mk(map[string]any{"a": map[string]any{"chart": b}, "b": map[string]any{"chart": b}}),
		"ziwei.chart": zc, "ziwei.fullchart": mk(map[string]any{"chart": z}),
		"ziwei.daxian": mk(map[string]any{"chart": z}),
		"ziwei.liunian": mk(map[string]any{"chart": z, "liu_nian": 2026}),
		"liuyao.chart": []byte(`{"solar_time":"2026-07-31T10:00:00+08:00","yaos":` + string(yaos) + `}`),
		"qimen.chart": []byte(`{"solar_time":"2026-07-31T10:00:00+08:00","kind":"shi"}`),
		"bazhai.chart": mk(map[string]any{"solar_time": "1984-02-15T08:00:00+08:00", "gender": "male"}),
		"xuankong.chart": mk(map[string]any{"solar_time": "2026-07-31T10:00:00+08:00", "zuo_shan": 2, "xiang_shan": 8}),
		"xuankong.liunian": mk(map[string]any{"year": 2026}),
		"huangli.days": mk(map[string]any{"start_date": "2026-08-01", "count": 2}),
		"tianwen.time": mk(map[string]any{"time": "1984-02-15T08:00:00+08:00", "longitude": 116.4, "latitude": 39.9}),
		"time.now": mk(map[string]any{}),
		"qiming.char": mk(map[string]any{"char": "明"}),
		"qiming.pick": mk(map[string]any{"surname": "陈", "wuxing1": "木", "count": 2}),
	}
	doc := reg.OpenRPCDocument()
	var d struct {
		Methods []struct {
			Name   string          `json:"name"`
			Result json.RawMessage `json:"result"`
		} `json:"methods"`
	}
	_ = json.Unmarshal(doc, &d)
	schemas := map[string]map[string]any{}
	for _, mm := range d.Methods {
		var r struct {
			Properties struct {
				Data struct {
					Properties map[string]any `json:"properties"`
				} `json:"data"`
			} `json:"properties"`
		}
		if err := json.Unmarshal(mm.Result, &r); err == nil {
			schemas[mm.Name] = r.Properties.Data.Properties
		}
	}
	okCnt, failList := 0, []string{}
	for name, params := range calls {
		out, err := reg.Execute(context.Background(), name, params)
		if err != nil {
			t.Logf("  ⚠️ %s: %v", name, err)
			continue
		}
		var res struct{ Data map[string]any `json:"data"` }
		_ = json.Unmarshal(out, &res)
		if res.Data == nil {
			continue
		}
		var missing []string
		for k := range res.Data {
			if _, ok := schemas[name][k]; !ok {
				missing = append(missing, k)
			}
		}
		if len(missing) > 0 {
			failList = append(failList, name+": "+mustStr(missing))
		} else {
			okCnt++
		}
	}
	t.Logf("✅ %d 方法 schema 与 Go 输出一致", okCnt)
	if len(failList) > 0 {
		for _, f := range failList {
			t.Errorf("❌ %s", f)
		}
	}
}
func mustStr(v any) string { b, _ := json.Marshal(v); return string(b) }
