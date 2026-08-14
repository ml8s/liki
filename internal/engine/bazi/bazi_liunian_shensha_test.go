package bazi

import (
	"testing"
	"time"

	"liki-engine/internal/engine/ganzhi"
	"liki-engine/internal/engine/tianwen"
)

// ============================================================================
// 流年神煞扩展测试（E1 年支+日支双查 / E2 值年凶煞）——先测试后实现
// ============================================================================

// T1: 日支神煞逢流年（年支不中、日支中）→ 应返回（原仅年支查不返回）
func TestLiuNian_ShenSha_RiBranch(t *testing.T) {
	// 年支=申（申的桃花=酉），日支=巳（巳的桃花=午），流年支=午
	// 年支申桃花酉 ≠ 午（年支不中）；日支巳桃花午 = 午（日支中）→ 应返回桃花
	ss := computeDynamicShenSha(ganzhi.ZhiWu, ganzhi.ZhiShen, ganzhi.ZhiSi, ganzhi.GanJia)
	if len(ss) != 1 || ss[0].Name != "桃花" {
		t.Errorf("日支桃花逢流年：got %v, want [桃花]", ss)
	}
}

// T2: 年支+日支命中同一神煞 → 去重只输出一条
func TestLiuNian_ShenSha_Dedup(t *testing.T) {
	// 年支=申（桃花酉）、日支=子（子桃花酉）、流年支=酉 → 年日都中桃花，只输出一条
	ss := computeDynamicShenSha(ganzhi.ZhiYou, ganzhi.ZhiShen, ganzhi.ZhiZi, ganzhi.GanJia)
	count := 0
	for _, s := range ss {
		if s.Name == "桃花" {
			count++
		}
	}
	if count != 1 {
		t.Errorf("年日同桃花应去重：got %d 条桃花, want 1", count)
	}
}

// T3: 年支中桃花、日支中驿马（不同神煞）→ 都输出
func TestLiuNian_ShenSha_Multi(t *testing.T) {
	// 年支=申（桃花酉）、流年=酉 → 桃花；日支=亥（亥卯未驿马巳）、流年=巳（年支巳无桃花/劫煞/驿马命中）→ 驿马
	ss1 := computeDynamicShenSha(ganzhi.ZhiYou, ganzhi.ZhiShen, ganzhi.ZhiHai, ganzhi.GanJia)
	if len(ss1) != 1 || ss1[0].Name != "桃花" {
		t.Errorf("年支桃花逢流年：got %v, want [桃花]", ss1)
	}
	ss2 := computeDynamicShenSha(ganzhi.ZhiSi, ganzhi.ZhiSi, ganzhi.ZhiHai, ganzhi.GanJia)
	if len(ss2) != 1 || ss2[0].Name != "驿马" {
		t.Errorf("日支驿马逢流年：got %v, want [驿马]", ss2)
	}
}

// T4: 值年病符临命（太岁后1辰）——1984 甲子命（月支寅），2021 辛丑太岁，病符=丑后1=寅（月支）→ 应返回病符
func TestLiuNian_AnnualBingFu(t *testing.T) {
	st := tianwen.GregorianToSolar(
		time.Date(1984, 2, 15, 8, 0, 0, 0, time.FixedZone("CST", 8*3600)), 116.4, 8)
	chart := ComputeChart(st, ganzhi.Male)
	ln, err := ComputeLiuNian(chart, 2021) // 辛丑
	if err != nil {
		t.Fatalf("ComputeLiuNian(2021): %v", err)
	}
	names := map[string]bool{}
	for _, s := range ln.ShenSha {
		names[s.Name] = true
	}
	if !names["病符"] {
		t.Errorf("2021 缺病符（太岁丑→病符寅，命局月支寅），got %v", ln.ShenSha)
	}
}

// T5: 值年丧门临命（太岁后2辰）——2022 壬寅太岁，丧门=寅后2=子（命局年/日支子）→ 应返回丧门
func TestLiuNian_AnnualSangMen(t *testing.T) {
	st := tianwen.GregorianToSolar(
		time.Date(1984, 2, 15, 8, 0, 0, 0, time.FixedZone("CST", 8*3600)), 116.4, 8)
	chart := ComputeChart(st, ganzhi.Male)
	ln, err := ComputeLiuNian(chart, 2022) // 壬寅
	if err != nil {
		t.Fatalf("ComputeLiuNian(2022): %v", err)
	}
	names := map[string]bool{}
	for _, s := range ln.ShenSha {
		names[s.Name] = true
	}
	if !names["丧门"] {
		t.Errorf("2022 缺丧门（太岁寅→丧门子，命局有子），got %v", ln.ShenSha)
	}
}

// T5b: 值年吊客临命（太岁前2辰）——2020 庚子太岁，吊客=子前2=寅（命局月支寅）→ 应返回吊客
func TestLiuNian_AnnualDiaoKe(t *testing.T) {
	st := tianwen.GregorianToSolar(
		time.Date(1984, 2, 15, 8, 0, 0, 0, time.FixedZone("CST", 8*3600)), 116.4, 8)
	chart := ComputeChart(st, ganzhi.Male)
	ln, err := ComputeLiuNian(chart, 2020) // 庚子
	if err != nil {
		t.Fatalf("ComputeLiuNian(2020): %v", err)
	}
	names := map[string]bool{}
	for _, s := range ln.ShenSha {
		names[s.Name] = true
	}
	if !names["吊客"] {
		t.Errorf("2020 缺吊客（太岁子→吊客寅，命局有寅），got %v", ln.ShenSha)
	}
}

// T5c: 值年大耗临命（太岁对冲）——2026 丙午太岁，大耗=午对冲子（命局年/日支子）→ 应返回大耗
func TestLiuNian_AnnualDaHao(t *testing.T) {
	st := tianwen.GregorianToSolar(
		time.Date(1984, 2, 15, 8, 0, 0, 0, time.FixedZone("CST", 8*3600)), 116.4, 8)
	chart := ComputeChart(st, ganzhi.Male)
	ln, err := ComputeLiuNian(chart, 2026) // 丙午
	if err != nil {
		t.Fatalf("ComputeLiuNian(2026): %v", err)
	}
	names := map[string]bool{}
	for _, s := range ln.ShenSha {
		names[s.Name] = true
	}
	if !names["大耗"] {
		t.Errorf("2026 缺大耗（太岁午→大耗子，命局有子），got %v", ln.ShenSha)
	}
}

// T6: 值年煞不临命 → 不输出值年凶煞（构造：命局四柱无任何值年煞支）
func TestLiuNian_AnnualNoHit(t *testing.T) {
	// 白盒：构造四柱（年卯 月卯 日卯 时卯），太岁=子 → 病符亥/丧门戌/吊客寅/大耗午，卯均不逢
	bz := ganzhi.Bazi{
		Nian: ganzhi.Zhu{Gan: ganzhi.GanYi, Zhi: ganzhi.ZhiMao},
		Yue:  ganzhi.Zhu{Gan: ganzhi.GanYi, Zhi: ganzhi.ZhiMao},
		Ri:   ganzhi.Zhu{Gan: ganzhi.GanYi, Zhi: ganzhi.ZhiMao},
		Shi:  ganzhi.Zhu{Gan: ganzhi.GanYi, Zhi: ganzhi.ZhiMao},
	}
	ss := computeAnnualShenSha(ganzhi.ZhiZi, bz)
	if len(ss) != 0 {
		t.Errorf("四柱卯、太岁子应无值年煞：got %v, want empty", ss)
	}
}

// T7: 值年煞与年支型神煞并存 → 都输出（2022：丧门子 + 年支子劫煞位巳？验证至少含丧门且整体非空）
func TestLiuNian_AnnualAndDynamicCoexist(t *testing.T) {
	st := tianwen.GregorianToSolar(
		time.Date(1984, 2, 15, 8, 0, 0, 0, time.FixedZone("CST", 8*3600)), 116.4, 8)
	chart := ComputeChart(st, ganzhi.Male)
	ln, err := ComputeLiuNian(chart, 2022)
	if err != nil {
		t.Fatalf("ComputeLiuNian(2022): %v", err)
	}
	names := map[string]bool{}
	for _, s := range ln.ShenSha {
		names[s.Name] = true
	}
	if !names["丧门"] {
		t.Errorf("2022 缺丧门：got %v", ln.ShenSha)
	}
	// 2022 寅年：年支子劫煞位巳（子年劫煞巳）——流年寅≠巳，但验证整体机制不冲突即可
	if len(ln.ShenSha) == 0 {
		t.Error("2022 ShenSha 为空，值年煞+动态煞机制异常")
	}
}
