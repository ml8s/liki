package agent

import (
	"encoding/json"
	"testing"
)

// TestDiscoverMethods_Filter 验证 rpc.discover 按域过滤：
// 1) 过滤结果只含指定域方法，schema（params/description）完整不丢失
// 2) 具体方法过滤只返回自身（rpc.discover 是元方法，不混入）
// 3) 空 pattern 返回全量
func TestDiscoverMethods_Filter(t *testing.T) {
	reg := NewRPCRegistry()

	doc := reg.DiscoverMethods([]string{"bazi"})
	var out openRPCDoc
	_ = json.Unmarshal(doc, &out)
	names := map[string]bool{}
	for _, m := range out.Methods {
		names[m.Name] = true
		if len(m.Name) < 6 || m.Name[:5] != "bazi." {
			t.Errorf("bazi 过滤混入: %s", m.Name)
		}
		if len(m.Params) == 0 {
			t.Errorf("%s missing params", m.Name)
		}
		if m.Description == "" {
			t.Errorf("%s missing description", m.Name)
		}
	}
	if !names["bazi.chart"] {
		t.Errorf("bazi 域缺方法: %v", names)
	}
	if names["ziwei.chart"] || names["qimen.chart"] {
		t.Errorf("bazi 过滤混入其他域: %v", names)
	}

	doc2 := reg.DiscoverMethods([]string{"xuankong.chart"})
	_ = json.Unmarshal(doc2, &out)
	if len(out.Methods) != 1 || out.Methods[0].Name != "xuankong.chart" {
		t.Fatalf("xuankong.chart 过滤: got %v", out.Methods)
	}
	if len(out.Methods[0].Params) == 0 {
		t.Errorf("xuankong.chart missing params")
	}

	doc3 := reg.DiscoverMethods(nil)
	_ = json.Unmarshal(doc3, &out)
	if len(out.Methods) <= 2 {
		t.Fatalf("空过滤: got %d", len(out.Methods))
	}
}
