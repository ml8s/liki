package agent

import (
	"reflect"
	"testing"

	"liki-engine/internal/engine/liuyao"
)

// TestLiuyaoChartBenGuaSchemaMatchesTable guards the liuyao.chart `ben_gua`
// enum against the actual 64-hexagram table (zhouyi.json). The enum must be the
// exact 64 卦名 in table order: no duplicates, no omissions.
func TestLiuyaoChartBenGuaSchemaMatchesTable(t *testing.T) {
	reg := NewRPCRegistry()
	schemas := resultDataSchemas(t, reg)

	schema, ok := schemas["liuyao.chart"].(map[string]any)
	if !ok {
		t.Fatalf("liuyao.chart result data schema missing")
	}
	properties, _ := schema["properties"].(map[string]any)
	benGua, _ := properties["ben_gua"].(map[string]any)
	rawEnum, _ := benGua["enum"].([]any)
	if len(rawEnum) == 0 {
		t.Fatalf("liuyao.chart ben_gua enum is empty")
	}

	got := make([]string, 0, len(rawEnum))
	seen := make(map[string]bool, len(rawEnum))
	for _, raw := range rawEnum {
		name, ok := raw.(string)
		if !ok {
			t.Fatalf("ben_gua enum has non-string value %#v", raw)
		}
		if seen[name] {
			t.Fatalf("ben_gua enum has duplicate %q", name)
		}
		seen[name] = true
		got = append(got, name)
	}

	want := liuyao.HexagramNames()
	if len(want) != 64 {
		t.Fatalf("hexagram table has %d names, want 64", len(want))
	}
	if !reflect.DeepEqual(got, want) {
		t.Fatalf("ben_gua enum (%d) does not match hexagram table (%d)", len(got), len(want))
	}
}
