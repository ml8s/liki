package qiming

import "testing"

// pick 测试：combos 结构（first/second 纯 char，按笔画拆 id）

func TestPickSingleWuge(t *testing.T) {
	// 单名 + 五格：19 个吉 s1，每 combo 仅 first
	combos, err := PickChars("陈", "木", "", 1, true)
	if err != nil {
		t.Fatalf("PickChars: %v", err)
	}
	if len(combos) == 0 {
		t.Fatal("单名五格应返回 19 个 combo（吉 s1），实际 0")
	}
	for _, c := range combos {
		if len(c.First) == 0 {
			t.Errorf("combo %d first 为空", c.ID)
		}
		if len(c.Second) != 0 {
			t.Errorf("combo %d 单名不应有 second", c.ID)
		}
		// 纯 char 验证
		for _, ch := range c.First {
			if len([]rune(ch)) != 1 {
				t.Errorf("first 含非单字: %q", ch)
			}
		}
	}
}

func TestPickDoubleWuge(t *testing.T) {
	// 双名 + 五格：3 个 combo，每 combo first+second
	combos, err := PickChars("陈", "木", "火", 2, true)
	if err != nil {
		t.Fatalf("PickChars: %v", err)
	}
	if len(combos) == 0 {
		t.Fatal("双名五格应返回 3 个 combo（三才后），实际 0")
	}
	for _, c := range combos {
		if len(c.First) == 0 {
			t.Errorf("combo %d first 为空", c.ID)
		}
		if len(c.Second) == 0 {
			t.Errorf("combo %d second 为空（双名必须有）", c.ID)
		}
	}
}

func TestPickSingleNoWuge(t *testing.T) {
	// 单名 + 无五格：按笔画拆 id，每 combo 仅 first
	combos, err := PickChars("", "木", "", 1, false)
	if err != nil {
		t.Fatalf("PickChars: %v", err)
	}
	if len(combos) == 0 {
		t.Fatal("单名无五格应返回按笔画拆的 combo，实际 0")
	}
	for _, c := range combos {
		if len(c.First) == 0 {
			t.Errorf("combo %d first 为空", c.ID)
		}
		if len(c.Second) != 0 {
			t.Errorf("combo %d 单名不应有 second", c.ID)
		}
	}
}

func TestPickDoubleNoWuge(t *testing.T) {
	// 双名 + 无五格：两池（木+火），按笔画拆 id
	combos, err := PickChars("", "木", "火", 2, false)
	if err != nil {
		t.Fatalf("PickChars: %v", err)
	}
	if len(combos) == 0 {
		t.Fatal("双名无五格应返回 combo，实际 0")
	}
	for _, c := range combos {
		if len(c.First) == 0 {
			t.Errorf("combo %d first 为空", c.ID)
		}
		if len(c.Second) == 0 {
			t.Errorf("combo %d second 为空（双名必须有）", c.ID)
		}
	}
}
