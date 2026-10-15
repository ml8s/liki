package agent

import (
	"context"
	"encoding/json"
	"testing"
)

func TestSchemaShenShaEnum(t *testing.T) {
	reg := NewRPCRegistry()
	bc := mustJSON(t, map[string]any{
		"solar_time": "1984-02-15T08:00:00+08:00",
		"gender":     "male",
	})
	b := executeAndDecode(t, reg, "bazi.chart", bc)["data"].(map[string]any)

	enums := map[string][]string{}
	for _, name := range reg.Names() {
		method, ok := reg.Method(name)
		if !ok {
			continue
		}
		var result struct {
			Properties struct {
				Data struct {
					Properties map[string]any `json:"properties"`
				} `json:"data"`
			} `json:"properties"`
		}
		decodeJSON(t, method.Result, &result)
		shenSha, ok := result.Properties.Data.Properties["shensha"].(map[string]any)
		if !ok {
			continue
		}
		items, _ := shenSha["items"].(map[string]any)
		properties, _ := items["properties"].(map[string]any)
		nameSchema, _ := properties["name"].(map[string]any)
		values, _ := nameSchema["enum"].([]any)
		for _, value := range values {
			enums[name] = append(enums[name], value.(string))
		}
	}

	calls := map[string]json.RawMessage{
		"bazi.liunian": mustJSON(t, map[string]any{"chart": b, "year": 2026}),
		"bazi.liuyue":  mustJSON(t, map[string]any{"chart": b, "year": 2026, "month": 6}),
		"bazi.liuri":   mustJSON(t, map[string]any{"chart": b, "year": 2026, "month": 6, "day": 4}),
	}
	for name, params := range calls {
		enum, ok := enums[name]
		if !ok || len(enum) == 0 {
			t.Errorf("%s schema has no shensha.name enum", name)
			continue
		}
		output, err := reg.Execute(context.Background(), name, params)
		if err != nil {
			t.Errorf("%s: %v", name, err)
			continue
		}
		var result struct {
			Data struct {
				ShenSha []struct {
					Name string `json:"name"`
				} `json:"shensha"`
			} `json:"data"`
		}
		decodeJSON(t, output, &result)
		allowed := make(map[string]bool, len(enum))
		for _, value := range enum {
			allowed[value] = true
		}
		for _, shenSha := range result.Data.ShenSha {
			if !allowed[shenSha.Name] {
				t.Errorf("%s returned shensha %q outside schema enum", name, shenSha.Name)
			}
		}
	}
}
