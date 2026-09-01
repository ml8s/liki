package qiming

import (
	"strings"
	"testing"
)

// singleStrokes 构造单姓（Total==Last，Compound=false）便于测试复用。
func singleStrokes(n int) SurnameStrokes {
	return SurnameStrokes{Total: n, Last: n}
}

// =============================================================================
// SurnameStrokesOf — 复姓/单姓笔画信息
// =============================================================================

func TestSurnameStrokesOf_Compound(t *testing.T) {
	// 欧=15, 阳=17 → Total=32, Last=17（最后一字），复姓
	ss, err := SurnameStrokesOf("欧阳")
	if err != nil {
		t.Fatalf("SurnameStrokesOf(欧阳): %v", err)
	}
	if ss.Total != 32 {
		t.Errorf("Total = %d, want 32", ss.Total)
	}
	if ss.Last != 17 {
		t.Errorf("Last = %d, want 17", ss.Last)
	}
	if !ss.Compound {
		t.Error("Compound = false, want true")
	}
}

func TestSurnameStrokesOf_Single(t *testing.T) {
	ss, err := SurnameStrokesOf("王")
	if err != nil {
		t.Fatalf("SurnameStrokesOf(王): %v", err)
	}
	if ss.Total != 4 || ss.Last != 4 {
		t.Errorf("single surname: Total=%d Last=%d, want 4/4", ss.Total, ss.Last)
	}
	if ss.Compound {
		t.Error("Compound = true, want false for single surname")
	}
}

func TestSurnameStrokesOf_NotFound(t *testing.T) {
	if _, err := SurnameStrokesOf("𠀀"); err == nil {
		t.Error("expected error for unknown surname char")
	}
	if _, err := SurnameStrokesOf(""); err == nil {
		t.Error("expected error for empty surname")
	}
}

// =============================================================================
// computeWuGeFromStrokes — 复姓五格
// =============================================================================

func TestComputeWuGeFromStrokes_Compound(t *testing.T) {
	// 欧阳(Total 14, Last 6) + 佳(8) + 桐(10)
	ss := SurnameStrokes{Total: 14, Last: 6, Compound: true}
	wg := computeWuGeFromStrokes(ss, 8, 10)

	// 复姓：天格 = 姓氏笔画之和（不加 1）= 14
	if wg.TianGe.Stroke != 14 {
		t.Errorf("天格 = %d, want 14", wg.TianGe.Stroke)
	}
	// 人格 = 姓氏最后一字(6) + 名第一字(8) = 14
	if wg.RenGe.Stroke != 14 {
		t.Errorf("人格 = %d, want 14", wg.RenGe.Stroke)
	}
	// 地格 = 名笔画之和 = 18
	if wg.DiGe.Stroke != 18 {
		t.Errorf("地格 = %d, want 18", wg.DiGe.Stroke)
	}
	// 总格 = 姓全部(14) + 名全部(18) = 32
	if wg.ZongGe.Stroke != 32 {
		t.Errorf("总格 = %d, want 32", wg.ZongGe.Stroke)
	}
	// 外格 = 总格 - 人格 + 1 = 32 - 14 + 1 = 19
	if wg.WaiGe.Stroke != 19 {
		t.Errorf("外格 = %d, want 19", wg.WaiGe.Stroke)
	}
}

func TestComputeWuGeFromStrokes_CompoundVsSingle(t *testing.T) {
	// 同样笔画的名（佳8 桐10），复姓与单姓五格必须不同
	compound := SurnameStrokes{Total: 14, Last: 6, Compound: true}
	single := singleStrokes(14) // 单姓 14 笔画

	wgc := computeWuGeFromStrokes(compound, 8, 10)
	wgs := computeWuGeFromStrokes(single, 8, 10)

	// 复姓天格=14（不加1），单姓天格=15
	if wgc.TianGe.Stroke == wgs.TianGe.Stroke {
		t.Errorf("复姓/单姓天格不应相同: both %d", wgc.TianGe.Stroke)
	}
	if wgc.TianGe.Stroke != 14 || wgs.TianGe.Stroke != 15 {
		t.Errorf("天格 复姓=%d want 14, 单姓=%d want 15", wgc.TianGe.Stroke, wgs.TianGe.Stroke)
	}
	// 复姓人格=Last(6)+8=14，单姓人格=14+8=22
	if wgc.RenGe.Stroke != 14 || wgs.RenGe.Stroke != 22 {
		t.Errorf("人格 复姓=%d want 14, 单姓=%d want 22", wgc.RenGe.Stroke, wgs.RenGe.Stroke)
	}
}

// =============================================================================
// SancaiHarmonious — 复姓三才
// =============================================================================

func TestSancaiHarmonious_Compound(t *testing.T) {
	// 复姓：天格=Total=14→? 人格=Last+s1, 地格=s1+s2
	// 只验证公式路径不 panic 且与单姓走同一套三才表
	ss := SurnameStrokes{Total: 14, Last: 6, Compound: true}
	// 欧阳+佳(8)+桐(10): 天格14, 人格14, 地格18
	_ = SancaiHarmonious(ss, 8, 10)

	// 与 computeWuGeFromStrokes 的三才组合应一致
	wg := computeWuGeFromStrokes(ss, 8, 10)
	key := wg.TianGe.Element + wg.RenGe.Element + wg.DiGe.Element
	got := SancaiHarmonious(ss, 8, 10)
	want := false
	if v, ok := sanCaiCfg[key]; ok {
		want = v.Fortune == fortuneAuspicious || v.Fortune == fortuneGood
	}
	if got != want {
		t.Errorf("SancaiHarmonious = %v, want %v (key=%s)", got, want, key)
	}
}

// =============================================================================
// EvaluateNames — 复姓全链路
// =============================================================================

func TestEvaluateNames_CompoundSurname(t *testing.T) {
	results, err := EvaluateNames("欧阳", []string{"佳桐"}, "", nil, nil, true)
	if err != nil {
		if strings.Contains(err.Error(), "not found") {
			t.Skip("test chars not in DB: " + err.Error())
		}
		t.Fatalf("EvaluateNames: %v", err)
	}
	if len(results) != 1 {
		t.Fatalf("expected 1 result, got %d", len(results))
	}
	r := results[0]
	if r.Name != "欧阳佳桐" {
		t.Errorf("Name = %q, want 欧阳佳桐", r.Name)
	}
	// 复姓天格 = 32（欧15+阳17，不加1）
	if r.WuGe.TianGe.Stroke != 32 {
		t.Errorf("复姓天格 = %d, want 32", r.WuGe.TianGe.Stroke)
	}
	// 人格 = 阳(17) + 佳(8) = 25
	if r.WuGe.RenGe.Stroke != 25 {
		t.Errorf("复姓人格 = %d, want 25", r.WuGe.RenGe.Stroke)
	}
}

func TestPickChars_CompoundSurname(t *testing.T) {
	// 复姓 + 五格：必须能正常取字（内部走 ListViableStrokes/FilterSancai）
	combos, err := PickChars("欧阳", "木", "", 2, true)
	if err != nil {
		t.Fatalf("PickChars(欧阳, wuge): %v", err)
	}
	if len(combos) == 0 {
		t.Fatal("复姓五格应返回至少一个 combo")
	}
	for _, c := range combos {
		if len(c.First) == 0 || len(c.Second) == 0 {
			t.Errorf("combo %d 复姓双名必须 first+second 非空", c.ID)
		}
	}
}

func TestEvaluateNames_NoWuge(t *testing.T) {
	// wuge=false：不输出五格三才（WuGe/SanCai 为 nil）
	results, err := EvaluateNames("王", []string{"明辉"}, "", nil, nil, false)
	if err != nil {
		t.Fatalf("EvaluateNames: %v", err)
	}
	if len(results) != 1 {
		t.Fatalf("expected 1 result, got %d", len(results))
	}
	r := results[0]
	if r.WuGe != nil {
		t.Error("wuge=false 时 WuGe 应为 nil")
	}
	if r.SanCai != nil {
		t.Error("wuge=false 时 SanCai 应为 nil")
	}
	// 名字/姓氏/五行匹配仍应保留
	if r.Name != "王明辉" || r.GivenName != "明辉" {
		t.Errorf("name/given_name 应保留: %q/%q", r.Name, r.GivenName)
	}
}

func TestEvaluateNames_WithWuge(t *testing.T) {
	// wuge=true：五格三才照常输出
	results, err := EvaluateNames("王", []string{"明辉"}, "", nil, nil, true)
	if err != nil {
		t.Fatalf("EvaluateNames: %v", err)
	}
	if len(results) != 1 {
		t.Fatalf("expected 1 result, got %d", len(results))
	}
	r := results[0]
	if r.WuGe == nil || r.SanCai == nil {
		t.Fatal("wuge=true 时 WuGe/SanCai 不应为 nil")
	}
	if r.WuGe.TianGe.Stroke == 0 {
		t.Error("天格应可计算")
	}
}

// =============================================================================
// SurnameStroke — 复姓总笔画
// =============================================================================

func TestSurnameStroke_Compound(t *testing.T) {
	got, err := SurnameStroke("欧阳")
	if err != nil {
		t.Fatalf("SurnameStroke(欧阳): %v", err)
	}
	if got != 32 {
		t.Errorf("SurnameStroke(欧阳) = %d, want 32", got)
	}
}
