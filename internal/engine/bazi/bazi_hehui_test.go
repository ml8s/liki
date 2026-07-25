package bazi

import (
	"testing"

	"liki-engine/internal/engine/ganzhi"
)

func TestComputeFullTripleHeHui_SanHeWater(t *testing.T) {
	// 申子辰 → 三合水局
	bz := ganzhi.Bazi{
		Nian: ganzhi.Zhu{Gan: ganzhi.GanJia, Zhi: ganzhi.ZhiShen}, // 甲申
		Yue:  ganzhi.Zhu{Gan: ganzhi.GanBing, Zhi: ganzhi.ZhiZi},  // 丙子
		Ri:   ganzhi.Zhu{Gan: ganzhi.GanWu, Zhi: ganzhi.ZhiChen},  // 戊辰
		Shi:  ganzhi.Zhu{Gan: ganzhi.GanGeng, Zhi: ganzhi.ZhiWu},  // 庚午
	}
	got := combinedTriple(bz)
	if len(got) != 1 {
		t.Fatalf("len = %d, want 1", len(got))
	}
	if got[0].Type != relSanHe {
		t.Errorf("Type = %q, want %q", got[0].Type, relSanHe)
	}
	if got[0].Element != "水" {
		t.Errorf("Element = %q, want 水", got[0].Element)
	}
}

func TestComputeFullTripleHeHui_SanHeFire(t *testing.T) {
	// 寅午戌 → 三合火局
	bz := ganzhi.Bazi{
		Nian: ganzhi.Zhu{Gan: ganzhi.GanJia, Zhi: ganzhi.ZhiYin},  // 甲寅
		Yue:  ganzhi.Zhu{Gan: ganzhi.GanBing, Zhi: ganzhi.ZhiWu},  // 丙午
		Ri:   ganzhi.Zhu{Gan: ganzhi.GanWu, Zhi: ganzhi.ZhiXu},    // 戊戌
		Shi:  ganzhi.Zhu{Gan: ganzhi.GanGeng, Zhi: ganzhi.ZhiShen}, // 庚申
	}
	got := combinedTriple(bz)
	if len(got) != 1 {
		t.Fatalf("len = %d, want 1", len(got))
	}
	if got[0].Type != relSanHe {
		t.Errorf("Type = %q, want %q", got[0].Type, relSanHe)
	}
	if got[0].Element != "火" {
		t.Errorf("Element = %q, want 火", got[0].Element)
	}
}

func TestComputeFullTripleHeHui_SanHeWood(t *testing.T) {
	// 亥卯未 → 三合木局
	bz := ganzhi.Bazi{
		Nian: ganzhi.Zhu{Gan: ganzhi.GanYi, Zhi: ganzhi.ZhiHai}, // 乙亥
		Yue:  ganzhi.Zhu{Gan: ganzhi.GanDing, Zhi: ganzhi.ZhiMao}, // 丁卯
		Ri:   ganzhi.Zhu{Gan: ganzhi.GanJi, Zhi: ganzhi.ZhiWei},  // 己未
		Shi:  ganzhi.Zhu{Gan: ganzhi.GanXin, Zhi: ganzhi.ZhiYou}, // 辛酉
	}
	got := combinedTriple(bz)
	if len(got) != 1 {
		t.Fatalf("len = %d, want 1", len(got))
	}
	if got[0].Type != relSanHe {
		t.Errorf("Type = %q, want %q", got[0].Type, relSanHe)
	}
	if got[0].Element != "木" {
		t.Errorf("Element = %q, want 木", got[0].Element)
	}
}

func TestComputeFullTripleHeHui_SanHeMetal(t *testing.T) {
	// 巳酉丑 → 三合金局
	bz := ganzhi.Bazi{
		Nian: ganzhi.Zhu{Gan: ganzhi.GanBing, Zhi: ganzhi.ZhiSi},  // 丙巳
		Yue:  ganzhi.Zhu{Gan: ganzhi.GanJi, Zhi: ganzhi.ZhiYou},   // 己酉
		Ri:   ganzhi.Zhu{Gan: ganzhi.GanWu, Zhi: ganzhi.ZhiChou},  // 戊丑
		Shi:  ganzhi.Zhu{Gan: ganzhi.GanGeng, Zhi: ganzhi.ZhiShen}, // 庚申
	}
	got := combinedTriple(bz)
	if len(got) != 1 {
		t.Fatalf("len = %d, want 1", len(got))
	}
	if got[0].Type != relSanHe {
		t.Errorf("Type = %q, want %q", got[0].Type, relSanHe)
	}
	if got[0].Element != "金" {
		t.Errorf("Element = %q, want 金", got[0].Element)
	}
}

func TestComputeFullTripleHeHui_SanHuiWood(t *testing.T) {
	// 寅卯辰 → 三会木方 (东方)
	bz := ganzhi.Bazi{
		Nian: ganzhi.Zhu{Gan: ganzhi.GanJia, Zhi: ganzhi.ZhiYin},  // 甲寅
		Yue:  ganzhi.Zhu{Gan: ganzhi.GanBing, Zhi: ganzhi.ZhiMao},  // 丙卯
		Ri:   ganzhi.Zhu{Gan: ganzhi.GanWu, Zhi: ganzhi.ZhiChen},   // 戊辰
		Shi:  ganzhi.Zhu{Gan: ganzhi.GanGeng, Zhi: ganzhi.ZhiWu},   // 庚午
	}
	got := combinedTriple(bz)
	if len(got) != 1 {
		t.Fatalf("len = %d, want 1", len(got))
	}
	if got[0].Type != relSanHui {
		t.Errorf("Type = %q, want %q", got[0].Type, relSanHui)
	}
	if got[0].Element != "木" {
		t.Errorf("Element = %q, want 木", got[0].Element)
	}
}

func TestComputeFullTripleHeHui_SanHuiWater(t *testing.T) {
	// 亥子丑 → 三会水方 (北方)
	bz := ganzhi.Bazi{
		Nian: ganzhi.Zhu{Gan: ganzhi.GanYi, Zhi: ganzhi.ZhiHai},  // 乙亥
		Yue:  ganzhi.Zhu{Gan: ganzhi.GanDing, Zhi: ganzhi.ZhiZi},  // 丁子
		Ri:   ganzhi.Zhu{Gan: ganzhi.GanJi, Zhi: ganzhi.ZhiChou},  // 己丑
		Shi:  ganzhi.Zhu{Gan: ganzhi.GanXin, Zhi: ganzhi.ZhiWei},  // 辛未
	}
	got := combinedTriple(bz)
	if len(got) != 1 {
		t.Fatalf("len = %d, want 1", len(got))
	}
	if got[0].Type != relSanHui {
		t.Errorf("Type = %q, want %q", got[0].Type, relSanHui)
	}
	if got[0].Element != "水" {
		t.Errorf("Element = %q, want 水", got[0].Element)
	}
}

func TestComputeFullTripleHeHui_NoMatch(t *testing.T) {
	// No He or Hui pattern present.
	bz := ganzhi.Bazi{
		Nian: ganzhi.Zhu{Gan: ganzhi.GanJia, Zhi: ganzhi.ZhiZi},   // 甲子
		Yue:  ganzhi.Zhu{Gan: ganzhi.GanBing, Zhi: ganzhi.ZhiYin},  // 丙寅
		Ri:   ganzhi.Zhu{Gan: ganzhi.GanWu, Zhi: ganzhi.ZhiWu},     // 戊午
		Shi:  ganzhi.Zhu{Gan: ganzhi.GanGeng, Zhi: ganzhi.ZhiShen}, // 庚申
	}
	got := combinedTriple(bz)
	if len(got) != 0 {
		t.Errorf("len = %d, want 0", len(got))
	}
}

// --- ComputeHeHui tests ---


func chartFrom(ng, nz, yg, yz, rg, rz, sg, sz string) Chart {
	return Chart{
		Nian:   zhuInfo{Zhu: ganzhi.Zhu{Gan: mustParseGan(ng), Zhi: mustParseZhi(nz)}},
		Yue:    zhuInfo{Zhu: ganzhi.Zhu{Gan: mustParseGan(yg), Zhi: mustParseZhi(yz)}},
		Ri:     zhuInfo{Zhu: ganzhi.Zhu{Gan: mustParseGan(rg), Zhi: mustParseZhi(rz)}},
		Shi:    zhuInfo{Zhu: ganzhi.Zhu{Gan: mustParseGan(sg), Zhi: mustParseZhi(sz)}},
	}
}

// chartFromBazi converts a ganzhi.Bazi to a minimal Chart for hehui analysis.
func chartFromBazi(bz ganzhi.Bazi) Chart {
	return Chart{
		Nian: zhuInfo{Zhu: bz.Nian},
		Yue:  zhuInfo{Zhu: bz.Yue},
		Ri:   zhuInfo{Zhu: bz.Ri},
		Shi:  zhuInfo{Zhu: bz.Shi},
	}
}

// combinedTriple returns combined SanHe + SanHui from ComputeHeHui (replaces legacy computeFullTripleHeHui).
func combinedTriple(bz ganzhi.Bazi) []TripleGroup {
	r := ComputeHeHui(chartFromBazi(bz))
	return append(append([]TripleGroup{}, r.SanHe...), r.SanHui...)
}

//nolint:errcheck
func mustParseGan(s string) ganzhi.Gan { g, _ := ganzhi.ParseGan(s); return g }
//nolint:errcheck
func mustParseZhi(s string) ganzhi.Zhi { z, _ := ganzhi.ParseZhi(s); return z }

func TestComputeHeHui_GanHe(t *testing.T) {
	// 甲子 己巳 甲子 己巳 → all 3 adjacent pairs are 甲己合
	ch := chartFrom("甲", "子", "己", "巳", "甲", "子", "己", "巳")
	r := ComputeHeHui(ch)

	if len(r.GanHe) != 3 {
		t.Fatalf("got %d gan he pairs, want 3 (all adjacent)", len(r.GanHe))
	}
	if r.GanHe[0].PillarA != 0 || r.GanHe[0].PillarB != 1 {
		t.Errorf("pair 0: want [0,1], got [%d,%d]", r.GanHe[0].PillarA, r.GanHe[0].PillarB)
	}
	if r.GanHe[1].PillarA != 1 || r.GanHe[1].PillarB != 2 {
		t.Errorf("pair 1: want [1,2], got [%d,%d]", r.GanHe[1].PillarA, r.GanHe[1].PillarB)
	}
	if r.GanHe[2].PillarA != 2 || r.GanHe[2].PillarB != 3 {
		t.Errorf("pair 2: want [2,3], got [%d,%d]", r.GanHe[2].PillarA, r.GanHe[2].PillarB)
	}
}

func TestComputeHeHui_ZhiLiuHe(t *testing.T) {
	ch := chartFrom("甲", "子", "乙", "丑", "丙", "寅", "丁", "卯")
	r := ComputeHeHui(ch)
	found := false
	for _, zh := range r.ZhiLiuHe {
		if (zh.PillarA == 0 && zh.PillarB == 1) || (zh.PillarA == 1 && zh.PillarB == 0) {
			found = true
		}
	}
	if !found {
		t.Error("子丑六合 not found")
	}
}

func TestComputeHeHui_LiuChong(t *testing.T) {
	ch := chartFrom("甲", "子", "乙", "丑", "丙", "午", "丁", "卯")
	r := ComputeHeHui(ch)
	found := false
	for _, lc := range r.LiuChong {
		if (lc.PillarA == 0 && lc.PillarB == 2) || (lc.PillarA == 2 && lc.PillarB == 0) {
			found = true
		}
	}
	if !found {
		t.Error("子午冲 not found")
	}
}

func TestComputeHeHui_LiuHai(t *testing.T) {
	ch := chartFrom("甲", "子", "乙", "未", "丙", "寅", "丁", "卯")
	r := ComputeHeHui(ch)
	found := false
	for _, lh := range r.LiuHai {
		if (lh.PillarA == 0 && lh.PillarB == 1) || (lh.PillarA == 1 && lh.PillarB == 0) {
			found = true
		}
	}
	if !found {
		t.Error("子未害 not found")
	}
}

func TestComputeHeHui_Empty(t *testing.T) {
	// 甲丑 丙辰 戊寅 庚戌 — no gan he, no complete san he/hui
	ch := chartFrom("甲", "丑", "丙", "辰", "戊", "寅", "庚", "戌")
	r := ComputeHeHui(ch)
	if len(r.GanHe) != 0 {
		t.Errorf("expected no gan he, got %d", len(r.GanHe))
	}
	if len(r.SanHe) != 0 {
		t.Errorf("expected no san he, got %d", len(r.SanHe))
	}
	if len(r.SanHui) != 0 {
		t.Errorf("expected no san hui, got %d", len(r.SanHui))
	}
}

func TestComputeHeHui_LiuXing(t *testing.T) {
	// 甲子 乙卯 丙子 丁卯 — 子卯相刑 (无礼之刑) at two pillar pairs
	ch := chartFrom("甲", "子", "乙", "卯", "丙", "子", "丁", "卯")
	r := ComputeHeHui(ch)
	if len(r.LiuXing) == 0 {
		t.Fatal("no 相刑 found, want 子卯相刑")
	}
	foundNianYue := false
	for _, lx := range r.LiuXing {
		if (lx.PillarA == 0 && lx.PillarB == 1) || (lx.PillarA == 1 && lx.PillarB == 0) {
			foundNianYue = true
		}
	}
	if !foundNianYue {
		t.Error("Nian-Yue 子卯相刑 not found")
	}
}

func TestComputeFullTripleHeHui_DualPattern(t *testing.T) {
	// 寅午戌 (三合火) + 巳午未 (三会火方) — 午 repeats in both.
	bz := ganzhi.Bazi{
		Nian: ganzhi.Zhu{Gan: ganzhi.GanJia, Zhi: ganzhi.ZhiYin}, // 甲寅
		Yue:  ganzhi.Zhu{Gan: ganzhi.GanBing, Zhi: ganzhi.ZhiWu}, // 丙午
		Ri:   ganzhi.Zhu{Gan: ganzhi.GanWu, Zhi: ganzhi.ZhiXu},   // 戊戌
		Shi:  ganzhi.Zhu{Gan: ganzhi.GanDing, Zhi: ganzhi.ZhiSi}, // 丁巳
	}
	got := combinedTriple(bz)
	if len(got) != 1 {
		t.Fatalf("len = %d, want 1 (fire sanhe only, 巳午未 needs 未)", len(got))
	}
}
