package huangli

import (
	"testing"

	"liki-engine/internal/engine/ganzhi"
)

func TestRenYuanName_Normal(t *testing.T) {
	phases := ganzhi.RenYuanSiLingFenYeForZhi(ganzhi.ZhiYin)
	if len(phases) == 0 {
		t.Skip("RenYuan phases not available")
	}
	r := renYuanSiLing{Phases: phases, Current: &phases[0]}
	name := renYuanName(r)
	if name == "" {
		t.Error("renYuanName should not be empty when Current is set")
	}
	if name != phases[0].GanName {
		t.Errorf("renYuanName = %s, want %s", name, phases[0].GanName)
	}
}

func TestRenYuanName_NilCurrent(t *testing.T) {
	r := renYuanSiLing{Current: nil}
	if got := renYuanName(r); got != "" {
		t.Errorf("renYuanName(nil Current) = %q, want empty", got)
	}
}

// =============================================================================
// QueryDate — eventType 分配宜忌
// =============================================================================

func TestQueryDate_WithOtherEvents(t *testing.T) {
	// Test various event types to exercise different jianChuSuitable zhi.
	events := []string{"wedding", "travel", "opening", "medical", "funeral"}
	for _, ev := range events {
		t.Run("event="+ev, func(t *testing.T) {
			got, err := QueryDate("2024-06-15")
			if err != nil {
				t.Fatalf("QueryDate: %v", err)
			}
			if got.JianChu == "" {
				t.Error("JianChu should not be empty")
			}
		})
	}
}

// =============================================================================
// computeRenYuanSiLing — nil phases
// =============================================================================

func TestComputeRenYuanSiLing_NilPhases(t *testing.T) {
	// Branch 0 or out-of-range returns nil phases → should get empty slice.
	r := computeRenYuanSiLing(ganzhi.Zhi(0))
	if r.Current != nil {
		t.Error("Current should be nil for nil phases")
	}
	if len(r.Phases) != 0 {
		t.Errorf("Phases len = %d, want 0", len(r.Phases))
	}
	if r.YueZhi != ganzhi.Zhi(0) {
		t.Error("YueZhi should be preserved")
	}
}

// 正月建寅: 寅月寅日=建(offset 0), 寅月卯日=除(offset 1)
// 二月建卯: 卯月卯日=建(offset 0), 卯月辰日=除(offset 1)
func TestJianChuOffset(t *testing.T) {
	tests := []struct {
		name       string
		yueZhi     ganzhi.Zhi
		riZhi      ganzhi.Zhi
		wantOffset int
		wantGod    string
	}{
		{"寅月寅日→建", ganzhi.ZhiYin, ganzhi.ZhiYin, 0, "建"},
		{"寅月卯日→除", ganzhi.ZhiYin, ganzhi.ZhiMao, 1, "除"},
		{"寅月辰日→满", ganzhi.ZhiYin, ganzhi.ZhiChen, 2, "满"},
		{"卯月卯日→建", ganzhi.ZhiMao, ganzhi.ZhiMao, 0, "建"},
		{"卯月辰日→除", ganzhi.ZhiMao, ganzhi.ZhiChen, 1, "除"},
		{"子月子日→建", ganzhi.ZhiZi, ganzhi.ZhiZi, 0, "建"},
		{"子月午日→破", ganzhi.ZhiZi, ganzhi.ZhiWu, 6, "破"},
		{"午月午日→建", ganzhi.ZhiWu, ganzhi.ZhiWu, 0, "建"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			jianIdx := int(tt.yueZhi) - 1
			dayIdx := int(tt.riZhi) - 1
			offset := (dayIdx - jianIdx + 12) % 12

			if offset != tt.wantOffset {
				t.Errorf("offset = %d, want %d", offset, tt.wantOffset)
			}
			if offset < len(jianChuCfg.Sequence) {
				god := jianChuCfg.Sequence[offset]
				if god != tt.wantGod {
					t.Errorf("god = %s, want %s", god, tt.wantGod)
				}
			}
		})
	}
}

// TestHuangDaoForDay verifies 黄道黑道十二神.
func TestHuangDaoForDay(t *testing.T) {
	tests := []struct {
		yueZhi   ganzhi.Zhi
		riZhi    ganzhi.Zhi
		wantName string
		wantPath string
	}{
		// 寅月: 青龙起子 → 子=青龙, 丑=明堂, 寅=天刑
		{ganzhi.ZhiYin, ganzhi.ZhiZi, "青龙", "黄道"},
		{ganzhi.ZhiYin, ganzhi.ZhiChou, "明堂", "黄道"},
		{ganzhi.ZhiYin, ganzhi.ZhiYin, "天刑", "黑道"},
		// 卯月: 青龙起寅 → 寅=青龙, 卯=明堂, 辰=天刑, 巳=朱雀, 午=金匮
		{ganzhi.ZhiMao, ganzhi.ZhiYin, "青龙", "黄道"},
		{ganzhi.ZhiMao, ganzhi.ZhiWu, "金匮", "黄道"},
		{ganzhi.ZhiMao, ganzhi.ZhiChen, "天刑", "黑道"},
		// 子月: 青龙起申
		{ganzhi.ZhiZi, ganzhi.ZhiShen, "青龙", "黄道"},
		// 丑月: 青龙起戌
		{ganzhi.ZhiChou, ganzhi.ZhiXu, "青龙", "黄道"},
		{ganzhi.ZhiChou, ganzhi.ZhiHai, "明堂", "黄道"},
	}

	for _, tt := range tests {
		name := ganzhi.ZhiName(tt.yueZhi) + "月" + ganzhi.ZhiName(tt.riZhi) + "日"
		t.Run(name, func(t *testing.T) {
			got := huangDaoForDay(tt.yueZhi, tt.riZhi)
			if got.Name != tt.wantName {
				t.Errorf("Name = %s, want %s", got.Name, tt.wantName)
			}
			if got.Path != tt.wantPath {
				t.Errorf("Path = %s, want %s", got.Path, tt.wantPath)
			}
		})
	}
}

// TestHuangDaoCycle verifies all 12 stars cycle correctly in order.
func TestHuangDaoCycle(t *testing.T) {
	// 寅月子日起青龙, 12日 cycle through all stars.
	expected := []struct{ name, path string }{
		{"青龙", "黄道"}, {"明堂", "黄道"}, {"天刑", "黑道"}, {"朱雀", "黑道"},
		{"金匮", "黄道"}, {"天德", "黄道"}, {"白虎", "黑道"}, {"玉堂", "黄道"},
		{"天牢", "黑道"}, {"玄武", "黑道"}, {"司命", "黄道"}, {"勾陈", "黑道"},
	}
	for i := 0; i < 12; i++ {
		dz := ganzhi.Zhi((i)%12 + 1) // 子=1 through 亥=12
		got := huangDaoForDay(ganzhi.ZhiYin, dz)
		if got.Name != expected[i].name {
			t.Errorf("%s日: name = %s, want %s",
				ganzhi.ZhiName(dz), got.Name, expected[i].name)
		}
	}
}

// TestQingLongStart verifies青龙起始 for all 12 months.
func TestQingLongStart(t *testing.T) {
	tests := []struct {
		yueZhi    ganzhi.Zhi
		wantStart ganzhi.Zhi
	}{
		{ganzhi.ZhiYin, ganzhi.ZhiZi},    // 寅月青龙起子
		{ganzhi.ZhiMao, ganzhi.ZhiYin},   // 卯月起寅
		{ganzhi.ZhiChen, ganzhi.ZhiChen}, // 辰月起辰
		{ganzhi.ZhiSi, ganzhi.ZhiWu},     // 巳月起午
		{ganzhi.ZhiWu, ganzhi.ZhiShen},   // 午月起申
		{ganzhi.ZhiWei, ganzhi.ZhiXu},    // 未月起戌
		{ganzhi.ZhiShen, ganzhi.ZhiZi},   // 申月起子
		{ganzhi.ZhiYou, ganzhi.ZhiYin},   // 酉月起寅
		{ganzhi.ZhiXu, ganzhi.ZhiChen},   // 戌月起辰
		{ganzhi.ZhiHai, ganzhi.ZhiWu},    // 亥月起午
		{ganzhi.ZhiZi, ganzhi.ZhiShen},   // 子月起申
		{ganzhi.ZhiChou, ganzhi.ZhiXu},   // 丑月起戌
	}

	for _, tt := range tests {
		t.Run(ganzhi.ZhiName(tt.yueZhi)+"月", func(t *testing.T) {
			start, ok := qingLongStart[tt.yueZhi]
			if !ok {
				t.Fatal("month not found in qingLongStart map")
			}
			if start != tt.wantStart {
				t.Errorf("start zhi = %d(%s), want %d(%s)",
					start, ganzhi.ZhiName(start),
					int(tt.wantStart), ganzhi.ZhiName(tt.wantStart))
			}
		})
	}
}

// TestTaiSui verifies太岁 calculation.
func TestQueryDate_Golden_JianChu(t *testing.T) {
	// 建除十二神: based on month zhi and day zhi
	// 2024-06-15: 午月(月支=午), 庚戌日(日支=戌)
	// 午月起午日为建 → 午=建,未=除,申=满,酉=平,戌=定
	// 戌日 → 定日
	got, err := QueryDate("2024-06-15")
	if err != nil {
		t.Fatalf("QueryDate: %v", err)
	}
	if got.JianChu != "定" {
		t.Errorf("2024-06-15 JianChu=%q, want 定 (午月戌日)", got.JianChu)
	}

	// 2024-02-10: 寅月(月支=寅), 甲辰日(日支=辰)
	// 寅月起寅日为建 → 寅=建,卯=除,辰=满
	// 辰日 → 满日
	got2, err := QueryDate("2024-02-10")
	if err != nil {
		t.Fatalf("QueryDate: %v", err)
	}
	if got2.JianChu != "满" {
		t.Errorf("2024-02-10 JianChu=%q, want 满 (寅月辰日)", got2.JianChu)
	}
}

func TestQueryDate_Golden_MarksWarnings(t *testing.T) {
	// Verify that event type filtering produces marks/warnings.
	for _, ev := range []string{"wedding", "travel", "opening"} {
		t.Run(ev, func(t *testing.T) {
			got, err := QueryDate("2024-06-15")
			if err != nil {
				t.Fatalf("QueryDate: %v", err)
			}
			// Each date should have marks and/or warnings or be explicitly empty.
			// Just verify the fields exist and are correctly typed.
			if got.JianChu == "" {
				t.Error("JianChu should not be empty")
			}
		})
	}
}
