package agent

import (
	"context"
	"encoding/json"
	"testing"
)

// TestSchema_ShenShaEnum 保障：bazi.liunian/liuyue/liuri 实际返回的神煞 name
// 必须都在 schema 声明的 enum 内（防返回未知神煞、防 schema 与实现脱节）
func TestSchema_ShenShaEnum(t *testing.T) {
	reg := NewRPCRegistry()

	// 构造 beijing-1984 chart
	bc, _ := json.Marshal(map[string]any{"solar_time": "1984-02-15T08:00:00+08:00", "gender": "male"})
	bout, err := reg.Execute(context.Background(), "bazi.chart", bc)
	if err != nil {
		t.Fatalf("bazi.chart: %v", err)
	}
	var br struct{ Data map[string]any `json:"data"` }
	_ = json.Unmarshal(bout, &br)
	b := br.Data

	// OpenRPCDocument 拿 schema enum
	doc := reg.OpenRPCDocument()
	var dd struct {
		Methods []struct {
			Name   string          `json:"name"`
			Result json.RawMessage `json:"result"`
		} `json:"methods"`
	}
	_ = json.Unmarshal(doc, &dd)

	enums := map[string][]string{}
	for _, m := range dd.Methods {
		var r struct {
			Properties struct {
				Data struct {
					Properties map[string]any `json:"properties"`
				} `json:"data"`
			} `json:"properties"`
		}
		if err := json.Unmarshal(m.Result, &r); err == nil {
			ss, _ := r.Properties.Data.Properties["shensha"].(map[string]any)
			items, _ := ss["items"].(map[string]any)
			props, _ := items["properties"].(map[string]any)
			nameSchema, _ := props["name"].(map[string]any)
			enum, _ := nameSchema["enum"].([]any)
			for _, e := range enum {
				enums[m.Name] = append(enums[m.Name], e.(string))
			}
		}
	}

	mk := func(m map[string]any) json.RawMessage { b2, _ := json.Marshal(m); return b2 }
	calls := map[string]json.RawMessage{
		"bazi.liunian": mk(map[string]any{"chart": b, "year": 2026}),
		"bazi.liuyue":  mk(map[string]any{"chart": b, "year": 2026, "month": 6}),
		"bazi.liuri":   mk(map[string]any{"chart": b, "year": 2026, "month": 6, "day": 4}),
	}
	for name, params := range calls {
		enum, ok := enums[name]
		if !ok || len(enum) == 0 {
			t.Errorf("%s schema 无 shensha.name enum", name)
			continue
		}
		out, err := reg.Execute(context.Background(), name, params)
		if err != nil {
			t.Errorf("%s 执行失败: %v", name, err)
			continue
		}
		var res struct{ Data struct {
			ShenSha []struct{ Name string `json:"name"` } `json:"shensha"`
		} `json:"data"` }
		_ = json.Unmarshal(out, &res)
		enumSet := map[string]bool{}
		for _, e := range enum {
			enumSet[e] = true
		}
		for _, s := range res.Data.ShenSha {
			if !enumSet[s.Name] {
				t.Errorf("%s 返回神煞 %q 不在 schema enum 内（%v）", name, s.Name, enum)
			}
		}
	}
}
