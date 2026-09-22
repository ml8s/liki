package bazi

import (
	"encoding/json"
	"testing"

	_ "embed"
)

//go:embed data/tiaohou_reference.json
var tiaohouRefJSON []byte

//go:embed data/tiaohou.json
var tiaohouProdJSON []byte

// ── 穷通宝鉴120条原文对照测试 ──
// tiaohou_reference.json 按穷通宝鉴原文独立录入
// 与生产数据逐条对比, 差异即 error

func TestTiaoHou_All120_AgainstReference(t *testing.T) {
	type entry struct {
		RiYuan    string `json:"ri_yuan"`
		YueZhi    string `json:"month_branch"`
		Primary   string `json:"primary"`
		Secondary string `json:"secondary"`
	}

	var ref, prod []entry
	if err := json.Unmarshal(tiaohouRefJSON, &ref); err != nil {
		t.Fatal(err)
	}
	if err := json.Unmarshal(tiaohouProdJSON, &prod); err != nil {
		t.Fatal(err)
	}
	if len(ref) != 120 {
		t.Fatalf("reference entries = %d, want 120", len(ref))
	}
	if len(prod) != 120 {
		t.Fatalf("production entries = %d, want 120", len(prod))
	}

	refMap := make(map[string]entry, 120)
	for _, e := range ref {
		refMap[e.RiYuan+e.YueZhi] = e
	}
	prodMap := make(map[string]entry, 120)
	for _, e := range prod {
		prodMap[e.RiYuan+e.YueZhi] = e
	}

	mismatches := 0
	for _, e := range ref {
		key := e.RiYuan + e.YueZhi
		p, ok := prodMap[key]
		if !ok {
			t.Errorf("生产数据缺少 %s%s月", e.RiYuan, e.YueZhi)
			mismatches++
			continue
		}
		if p.Primary != e.Primary || p.Secondary != e.Secondary {
			t.Errorf("%s%s月: production=%s+%s, 穷通原文=%s+%s",
				e.RiYuan, e.YueZhi, p.Primary, p.Secondary, e.Primary, e.Secondary)
			mismatches++
		}
	}

	// 检查生产数据多出的条目
	for _, e := range prod {
		key := e.RiYuan + e.YueZhi
		if _, ok := refMap[key]; !ok {
			t.Errorf("生产数据多余 %s%s月", e.RiYuan, e.YueZhi)
			mismatches++
		}
	}

	if mismatches > 0 {
		if mismatches > 0 {
			t.Errorf("共 %d/120 条与穷通原文不一致", mismatches)
		}
	}
}
