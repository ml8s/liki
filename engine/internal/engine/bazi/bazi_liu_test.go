package bazi

import (
	"testing"
	"time"

	"liki-engine/internal/engine/ganzhi"
	"liki-engine/internal/engine/tianwen"
)

// helper: make a Chart from a known BaZi for testing liu* functions.
func makeTestChart(nian, yue, ri, shi ganzhi.Zhu, daYun *DaYun) Chart {
	return Chart{
		Nian:  zhuInfo{Zhu: nian, NaYin: ganzhi.NayinLabel(nian.Gan, nian.Zhi)},
		Yue:   zhuInfo{Zhu: yue, NaYin: ganzhi.NayinLabel(yue.Gan, yue.Zhi)},
		Ri:    zhuInfo{Zhu: ri, NaYin: ganzhi.NayinLabel(ri.Gan, ri.Zhi)},
		Shi:   zhuInfo{Zhu: shi, NaYin: ganzhi.NayinLabel(shi.Gan, shi.Zhi)},
		DaYun: daYun,
	}
}

// TestComputeLiuYue_WuHuDun verifies the month pillar follows 五虎遁 rule:
// 甲年 → 正月丙寅 → 午月=庚午.
func TestComputeLiuYue_WuHuDun(t *testing.T) {
	// 甲子年 丙寅月 戊辰日 庚申时
	cb := makeTestChart(
		ganzhi.Zhu{Gan: ganzhi.GanJia, Zhi: ganzhi.ZhiZi},
		ganzhi.Zhu{Gan: ganzhi.GanBing, Zhi: ganzhi.ZhiYin},
		ganzhi.Zhu{Gan: ganzhi.GanWu, Zhi: ganzhi.ZhiChen},
		ganzhi.Zhu{Gan: ganzhi.GanGeng, Zhi: ganzhi.ZhiShen},
		nil,
	)

	// 2024 = 甲辰年, June = 午月.
	// 甲年正月丙寅 → 午月 (正月+4) → 丙+4=庚 → 庚午月.
	ly, err := ComputeLiuYue(cb, 2024, 6)
	if err != nil {
		t.Fatal(err)
	}

	if ly.YueGan != ganzhi.GanGeng || ly.YueZhi != ganzhi.ZhiWu {
		t.Errorf("Month pillar = %s%s, want 庚午",
			ganzhi.GanName(ly.YueGan), ganzhi.ZhiName(ly.YueZhi))
	}
	if ly.ShiShen == "" {
		t.Error("ShiShen is empty")
	}
	if ly.Element == "" {
		t.Error("Element is empty")
	}
}

// TestComputeLiuYue_EthYear verifies 乙年 (2025=乙巳) month pillar.
func TestComputeLiuYue_EthYear(t *testing.T) {
	cb := makeTestChart(
		ganzhi.Zhu{Gan: ganzhi.GanYi, Zhi: ganzhi.ZhiChou},
		ganzhi.Zhu{Gan: ganzhi.GanWu, Zhi: ganzhi.ZhiYin},
		ganzhi.Zhu{Gan: ganzhi.GanJi, Zhi: ganzhi.ZhiSi},
		ganzhi.Zhu{Gan: ganzhi.GanXin, Zhi: ganzhi.ZhiWei},
		nil,
	)

	// 乙年正月戊寅 → 午月 (正月+4) → 戊+4=壬 → 壬午月.
	ly, err := ComputeLiuYue(cb, 2025, 6)
	if err != nil {
		t.Fatal(err)
	}
	if ly.YueGan != ganzhi.GanRen || ly.YueZhi != ganzhi.ZhiWu {
		t.Errorf("Month pillar = %s%s, want 壬午",
			ganzhi.GanName(ly.YueGan), ganzhi.ZhiName(ly.YueZhi))
	}
}

// TestComputeLiuYue_InvalidMonth verifies error for out-of-range month.
func TestComputeLiuYue_InvalidMonth(t *testing.T) {
	cb := makeTestChart(
		ganzhi.Zhu{Gan: ganzhi.GanJia, Zhi: ganzhi.ZhiZi},
		ganzhi.Zhu{Gan: ganzhi.GanBing, Zhi: ganzhi.ZhiYin},
		ganzhi.Zhu{Gan: ganzhi.GanWu, Zhi: ganzhi.ZhiChen},
		ganzhi.Zhu{Gan: ganzhi.GanGeng, Zhi: ganzhi.ZhiShen},
		nil,
	)
	_, err := ComputeLiuYue(cb, 2024, 0)
	if err == nil {
		t.Error("expected error for month=0")
	}
	_, err = ComputeLiuYue(cb, 2024, 13)
	if err == nil {
		t.Error("expected error for month=13")
	}
}

// TestComputeLiuRi_DayPillar verifies consecutive day pillars are sequential.
func TestComputeLiuRi_DayPillar(t *testing.T) {
	cb := makeTestChart(
		ganzhi.Zhu{Gan: ganzhi.GanJia, Zhi: ganzhi.ZhiZi},
		ganzhi.Zhu{Gan: ganzhi.GanBing, Zhi: ganzhi.ZhiYin},
		ganzhi.Zhu{Gan: ganzhi.GanWu, Zhi: ganzhi.ZhiChen},
		ganzhi.Zhu{Gan: ganzhi.GanGeng, Zhi: ganzhi.ZhiShen},
		nil,
	)

	base := time.Date(2024, 6, 10, 0, 0, 0, 0, time.UTC)
	prevIdx := -1
	for i := 0; i < 5; i++ {
		day := base.AddDate(0, 0, i)
		lr, err := ComputeLiuRi(cb, day.Year(), int(day.Month()), day.Day())
		if err != nil {
			t.Fatalf("day %s: %v", day.Format("2006-01-02"), err)
		}
		idx := ganzhi.SixtyCycleIndex(lr.RiGan, lr.RiZhi)
		if idx < 0 || idx >= 60 {
			t.Errorf("day %s: index = %d, want [0,59]", day.Format("2006-01-02"), idx)
		}
		if prevIdx >= 0 && idx != (prevIdx+1)%60 {
			t.Errorf("day %s: index %d, prev %d, want %d",
				day.Format("2006-01-02"), idx, prevIdx, (prevIdx+1)%60)
		}
		prevIdx = idx
	}
}

// TestComputeLiuRi_HasRequiredFields verifies all LiuRi fields are populated.
func TestComputeLiuRi_HasRequiredFields(t *testing.T) {
	st := tianwen.GregorianToSolar(
		time.Date(1982, 10, 13, 6, 45, 0, 0, time.FixedZone("", 8*3600)), 114.134, 8)
	chart := ComputeChart(st, ganzhi.Male)

	lr, err := ComputeLiuRi(chart, 2024, 6, 15)
	if err != nil {
		t.Fatal(err)
	}
	if lr.RiGan < 1 || lr.RiGan > 10 {
		t.Errorf("RiGan = %d", lr.RiGan)
	}
	if lr.RiZhi < 1 || lr.RiZhi > 12 {
		t.Errorf("RiZhi = %d", lr.RiZhi)
	}
	if lr.DayName == "" {
		t.Error("DayName is empty")
	}
	if lr.ShiShen == "" {
		t.Error("ShiShen is empty")
	}
	if lr.DayNaYin == "" {
		t.Error("DayNaYin is empty")
	}
	if len(lr.GanRels) == 0 {
		t.Error("GanRels is empty: day gan should interact with bazi")
	}
	if len(lr.ZhiRels) == 0 {
		t.Error("ZhiRels is empty: day zhi should interact with bazi")
	}
}

// TestComputeLiuShi_WuShuDun verifies the hour pillar follows 五鼠遁 rule.
// computeLiuShi computes RiZhu from the given date, not from the chart.
// 2024-06-15 = 乙日. 乙日: 子时→丙子, 丑时→丁丑, 寅时→戊寅, ...
func TestComputeLiuShi_WuShuDun(t *testing.T) {
	cb := makeTestChart(
		ganzhi.Zhu{Gan: ganzhi.GanJia, Zhi: ganzhi.ZhiZi},
		ganzhi.Zhu{Gan: ganzhi.GanBing, Zhi: ganzhi.ZhiYin},
		ganzhi.Zhu{Gan: ganzhi.GanWu, Zhi: ganzhi.ZhiChen},
		ganzhi.Zhu{Gan: ganzhi.GanGeng, Zhi: ganzhi.ZhiShen},
		nil,
	)

	// 2024-06-15 = 乙日 (day gan 2).
	// 乙日五鼠遁: 子→丙, 丑→丁, 寅→戊, 午→壬, 亥→丁.
	tests := []struct {
		hour    int
		wantGan ganzhi.Gan
		wantZhi ganzhi.Zhi
	}{
		{0, ganzhi.GanBing, ganzhi.ZhiZi},   // 乙日 子时: (2*2+1-2)%10=3 → 丙子
		{2, ganzhi.GanDing, ganzhi.ZhiChou}, // 乙日 丑时: (2*2+2-2)%10=4 → 丁丑
		{4, ganzhi.GanWu, ganzhi.ZhiYin},    // 乙日 寅时: (2*2+3-2)%10=5 → 戊寅
		{12, ganzhi.GanRen, ganzhi.ZhiWu},   // 乙日 午时: (2*2+7-2)%10=9 → 壬午
		{22, ganzhi.GanDing, ganzhi.ZhiHai}, // 乙日 亥时: (2*2+12-2)%10=14%10=4 → 丁亥
	}

	for _, tt := range tests {
		ls, err := ComputeLiuShi(cb, 2024, 6, 15, tt.hour)
		if err != nil {
			t.Fatalf("hour %d: %v", tt.hour, err)
		}
		if ls.ShiGan != tt.wantGan || ls.ShiZhi != tt.wantZhi {
			t.Errorf("hour %d: pillar = %s%s, want %s%s",
				tt.hour,
				ganzhi.GanName(ls.ShiGan), ganzhi.ZhiName(ls.ShiZhi),
				ganzhi.GanName(tt.wantGan), ganzhi.ZhiName(tt.wantZhi))
		}
	}
}

// TestComputeLiuShi_HasRequiredFields verifies all LiuShi fields are populated.
func TestComputeLiuShi_HasRequiredFields(t *testing.T) {
	cb := makeTestChart(
		ganzhi.Zhu{Gan: ganzhi.GanJia, Zhi: ganzhi.ZhiZi},
		ganzhi.Zhu{Gan: ganzhi.GanBing, Zhi: ganzhi.ZhiYin},
		ganzhi.Zhu{Gan: ganzhi.GanWu, Zhi: ganzhi.ZhiChen},
		ganzhi.Zhu{Gan: ganzhi.GanGeng, Zhi: ganzhi.ZhiShen},
		nil,
	)

	ls, err := ComputeLiuShi(cb, 2024, 6, 15, 12)
	if err != nil {
		t.Fatal(err)
	}
	if ls.ShiGan < 1 || ls.ShiGan > 10 {
		t.Errorf("ShiGan = %d", ls.ShiGan)
	}
	if ls.ShiZhi < 1 || ls.ShiZhi > 12 {
		t.Errorf("ShiZhi = %d", ls.ShiZhi)
	}
	if ls.HourName == "" {
		t.Error("HourName is empty")
	}
	if ls.ShiShen == "" {
		t.Error("ShiShen is empty")
	}
	if len(ls.GanRels) == 0 {
		t.Error("GanRels is empty: hour gan should interact with bazi")
	}
	if len(ls.ZhiRels) == 0 {
		t.Error("ZhiRels is empty: hour zhi should interact with bazi")
	}
}

// 流月神煞（E1 年支+日支双查——流月也走 computeDynamicShenSha）
func TestComputeLiuYue_ShenSha(t *testing.T) {
	// beijing-1984（甲子 丙寅 己卯 戊辰，男）2026 年 6 月（丙午月）：月支午
	// 动态：年支子灾煞午=午✓、日干己羊刃午=午✓（2026 golden 验证过流年午）——流月午同样命中
	cb := makeTestChart(
		ganzhi.Zhu{Gan: ganzhi.GanJia, Zhi: ganzhi.ZhiZi},
		ganzhi.Zhu{Gan: ganzhi.GanBing, Zhi: ganzhi.ZhiYin},
		ganzhi.Zhu{Gan: ganzhi.GanJi, Zhi: ganzhi.ZhiMao},
		ganzhi.Zhu{Gan: ganzhi.GanWu, Zhi: ganzhi.ZhiChen},
		nil,
	)
	ly, err := ComputeLiuYue(cb, 2026, 6)
	if err != nil {
		t.Fatalf("ComputeLiuYue: %v", err)
	}
	names := map[string]bool{}
	for _, s := range ly.ShenSha {
		names[s.Name] = true
	}
	if !names["灾煞"] {
		t.Errorf("流月午缺灾煞（年支子灾煞午）：got %v", ly.ShenSha)
	}
	// 己羊刃=巳非午（戊羊刃才在午）——流月午不应命中羊刃
	if names["羊刃"] {
		t.Errorf("流月午不应命中羊刃（己羊刃=巳）：got %v", ly.ShenSha)
	}
	// 流月不应含值年煞（病符/丧门/吊客/大耗——流年概念）
	for _, annual := range []string{"病符", "丧门", "吊客", "大耗"} {
		if names[annual] {
			t.Errorf("流月不应含值年煞 %s：got %v", annual, ly.ShenSha)
		}
	}
}

// 流日神煞（E1 年支+日支双查）
func TestComputeLiuRi_ShenSha(t *testing.T) {
	// beijing-1984（甲子 丙寅 己卯 戊辰，男）2026 年 6 月 4 日：流日支需命中某神煞
	cb := makeTestChart(
		ganzhi.Zhu{Gan: ganzhi.GanJia, Zhi: ganzhi.ZhiZi},
		ganzhi.Zhu{Gan: ganzhi.GanBing, Zhi: ganzhi.ZhiYin},
		ganzhi.Zhu{Gan: ganzhi.GanJi, Zhi: ganzhi.ZhiMao},
		ganzhi.Zhu{Gan: ganzhi.GanWu, Zhi: ganzhi.ZhiChen},
		nil,
	)
	lr, err := ComputeLiuRi(cb, 2026, 6, 4)
	if err != nil {
		t.Fatalf("ComputeLiuRi: %v", err)
	}
	// 流日支=子（2026-06-04）：年支子桃花酉≠子、日支卯桃花子=子✓（日支双查生效）
	names := map[string]bool{}
	for _, s := range lr.ShenSha {
		names[s.Name] = true
	}
	if !names["桃花"] {
		t.Errorf("流日子缺桃花（日支卯桃花子）：got %v", lr.ShenSha)
	}
	// 流日不应含值年煞
	for _, annual := range []string{"病符", "丧门", "吊客", "大耗"} {
		if names[annual] {
			t.Errorf("流日不应含值年煞 %s：got %v", annual, lr.ShenSha)
		}
	}
}
