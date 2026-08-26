package qiming

import "testing"

// build 测试：消费 combos → given name，只管双名

func TestBuildDoubleName(t *testing.T) {
	// 双名：first × second 笛卡尔积
	combos := []Combo{
		{ID: 0, First: []string{"囊", "耀"}, Second: []string{"倔", "倜"}},
	}
	names := BuildNames(combos)
	if len(names) != 4 {
		t.Fatalf("双名应 2×2=4 个，实际 %d: %v", len(names), names)
	}
	// 验证组合正确
	expect := map[string]bool{"囊倔": true, "囊倜": true, "耀倔": true, "耀倜": true}
	for _, n := range names {
		if !expect[n] {
			t.Errorf("意外组合: %s", n)
		}
	}
}

func TestBuildMultipleCombos(t *testing.T) {
	combos := []Combo{
		{ID: 0, First: []string{"囊"}, Second: []string{"倔"}},
		{ID: 1, First: []string{"中", "丰"}, Second: []string{"乳"}},
	}
	names := BuildNames(combos)
	if len(names) != 3 { // 1×1 + 2×1
		t.Fatalf("应 3 个，实际 %d: %v", len(names), names)
	}
}

func TestBuildSkipSingleName(t *testing.T) {
	// 无 second 的 combo（单名）跳过
	combos := []Combo{
		{ID: 0, First: []string{"中", "丰"}},                   // 单名，跳过
		{ID: 1, First: []string{"囊"}, Second: []string{"倔"}}, // 双名，保留
	}
	names := BuildNames(combos)
	if len(names) != 1 {
		t.Fatalf("应只 1 个（单名跳过），实际 %d: %v", len(names), names)
	}
	if names[0] != "囊倔" {
		t.Errorf("应 囊倔，实际 %s", names[0])
	}
}

func TestBuildEmptySecond(t *testing.T) {
	// second 为空数组 → 跳过
	combos := []Combo{
		{ID: 0, First: []string{"囊"}, Second: []string{}},
	}
	names := BuildNames(combos)
	if len(names) != 0 {
		t.Fatalf("second 空应无结果，实际 %d", len(names))
	}
}
