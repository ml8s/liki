package ziwei

import "testing"

// ── 空宫测试 ──
// 命宫无主星是常见场景(约30%)，紫微斗数用对宫借星

func TestJudgment_EmptyMingGong_NoCrash(t *testing.T) {
	// 命理: 空宫(命宫无主星)不应崩溃
	var palaces [12]palace
	for i := range palaces {
		palaces[i] = palace{Index: palaceIndex(i), Zhi: Zhi(i%12 + 1)}
	}
	chart := Chart{Palaces: palaces, SiHua: siHuaResult{}}

	result := ComputeJudgment(chart)
	if result.Rating != "下" {
		t.Errorf("空宫无四化→ rating=%q rule=%d, 命理应为下(无格局无四化)", result.Rating, result.Rule)
	}
	if len(result.Patterns) != 0 {
		t.Errorf("空宫应有0格局, got %d", len(result.Patterns))
	}
}

func TestJudgment_EmptyMingGong_WithSiHua(t *testing.T) {
	// 命理: 空宫+紫微化禄 → 财荫夹印(score=1)→规则5 "中"
	var palaces [12]palace
	for i := range palaces {
		palaces[i] = palace{Index: palaceIndex(i), Zhi: Zhi(i%12 + 1)}
	}
	palaces[4] = palace{Index: 4, Zhi: 5, Stars: []starInfo{
		{Star: ZiWei, Name: "紫微", IsMajor: true, SiHua: "禄"}}}
	chart := Chart{Palaces: palaces, SiHua: siHuaResult{ZiWei: HuaLu}}

	result := ComputeJudgment(chart)
	// 找到"财荫夹印"格局
	found := false
	for _, p := range result.Patterns {
		if p.Name == "财荫夹印" { found = true; break }
	}
	if !found {
		t.Error("空宫+紫微化禄→应有财荫夹印格局")
	}
	if result.Rating != "中" {
		t.Errorf("空宫+紫微化禄→ rating=%q, 命理应为中(中等格局)", result.Rating)
	}
}
