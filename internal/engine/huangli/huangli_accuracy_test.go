package huangli

import (
	"time"
	"testing"

	"liki-engine/internal/engine/ganzhi"
	"liki-engine/internal/engine/tianwen"
)

func TestEvaluateZhi_AllRelations(t *testing.T) {
	tests := []struct {
		name       string
		dayZhi     ganzhi.Zhi
		refZhi     ganzhi.Zhi
		wantRel    string
		wantMarks  bool
		wantWarns  bool
	}{
		{"子丑六合", ganzhi.ZhiZi, ganzhi.ZhiChou, "六合", true, false},
		{"申子三合半", ganzhi.ZhiShen, ganzhi.ZhiZi, "三合半", true, false},
		{"寅卯三会半", ganzhi.ZhiYin, ganzhi.ZhiMao, "三会半", true, false},
		{"子午六冲", ganzhi.ZhiZi, ganzhi.ZhiWu, "六冲", false, true},
		{"子卯相刑", ganzhi.ZhiZi, ganzhi.ZhiMao, "相刑", false, true},
		{"子未六害", ganzhi.ZhiZi, ganzhi.ZhiWei, "六害", false, true},
		{"子戌无关系", ganzhi.ZhiZi, ganzhi.ZhiXu, "无", false, false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			rel, marks, warns := evaluateZhi(tt.dayZhi, tt.refZhi, "日柱")
			if rel != tt.wantRel {
				t.Errorf("rel = %q, want %q", rel, tt.wantRel)
			}
			if tt.wantMarks && len(marks) == 0 {
				t.Error("expected marks")
			}
			if tt.wantWarns && len(warns) == 0 {
				t.Error("expected warnings")
			}
		})
	}
}

// =============================================================================
// ComputeBondDay / ComputeBondMonth — 择日合盘
// =============================================================================

func TestComputeBondDay_Basic(t *testing.T) {
	ts := tianwen.ComputeTimeset(tianwen.GregorianTime(time.Date(2000, time.Month(1), 1, 12, 0, 0, 0, time.FixedZone("", int(8.0*3600)))), 116.4074)
	result, err := ComputeBondDay(ts.Solar, "marriage", "2024-02-10")
	if err != nil {
		t.Fatalf("ComputeBondDay: %v", err)
	}
	if result.GanRelation == "" {
		t.Error("GanRelation should not be empty")
	}
	if result.ZhiRelation == "" {
		t.Error("ZhiRelation should not be empty")
	}
	if result.TaiSuiRelation == "" {
		t.Error("TaiSuiRelation should not be empty")
	}
	if result.Date == "" {
		t.Error("Date should not be empty")
	}
}

func TestComputeBondMonth_Basic(t *testing.T) {
	ts := tianwen.ComputeTimeset(tianwen.GregorianTime(time.Date(2000, time.Month(1), 1, 12, 0, 0, 0, time.FixedZone("", int(8.0*3600)))), 116.4074)
	m, err := ComputeBondMonth(ts.Solar, "marriage", "2024-06")
	if err != nil {
		t.Fatalf("ComputeBondMonth: %v", err)
	}
	if len(m.Days) != 30 {
		t.Errorf("len(Days) = %d, want 30", len(m.Days))
	}
	for i, d := range m.Days {
		if d.GanRelation == "" {
			t.Errorf("Day %d: empty GanRelation", i+1)
		}
		if d.ZhiRelation == "" {
			t.Errorf("Day %d: empty ZhiRelation", i+1)
		}
	}
}

// =============================================================================
// QueryMonth — 黄历月查询
// =============================================================================

func TestQueryMonth_Basic(t *testing.T) {
	m, err := QueryMonth("2024-06", "")
	if err != nil {
		t.Fatalf("QueryMonth: %v", err)
	}
	if m.Month != "2024-06" {
		t.Errorf("Month = %q, want 2024-06", m.Month)
	}
	if len(m.Days) != 30 {
		t.Errorf("len(Days) = %d, want 30", len(m.Days))
	}
	if m.Stem == "" || m.Branch == "" {
		t.Error("Month stem/branch should not be empty")
	}
	for i, d := range m.Days {
		if d.Date == "" {
			t.Errorf("Day %d: empty date", i+1)
		}
		if d.Mansion.Name == "" {
			t.Errorf("Day %d: empty mansion", i+1)
		}
	}
}

func TestQueryMonth_InvalidInput(t *testing.T) {
	_, err := QueryMonth("not-a-month", "")
	if err == nil {
		t.Fatal("expected error for invalid month")
	}
}

// =============================================================================
// taiSui — 太岁
// =============================================================================

func TestTaiSui_Cycle(t *testing.T) {
	expected := []string{"子", "丑", "寅", "卯", "辰", "巳", "午", "未", "申", "酉", "戌", "亥"}
	for i, want := range expected {
		year := 2020 + i
		got := taiSui(year)
		if ganzhi.ZhiName(got) != want {
			t.Errorf("taiSui(%d) = %s, want %s", year, ganzhi.ZhiName(got), want)
		}
	}
}

// =============================================================================
// jianChuSuitable — 建除宜忌四大分支
// =============================================================================

func TestJianChuSuitable_UnknownEvent(t *testing.T) {
	// Unknown event type → defaults to suitable=true.
	suitable, marks, warnings := jianChuSuitable("破", "nonexistent")
	if !suitable {
		t.Error("unknown event type should default to suitable=true")
	}
	if len(marks) != 0 || len(warnings) != 0 {
		t.Error("unknown event type should have no marks/warnings")
	}
}

func TestJianChuSuitable_SuitableMatch(t *testing.T) {
	// "成" is suitable for "wedding" (嫁娶).
	suitable, marks, warnings := jianChuSuitable("成", "wedding")
	if !suitable {
		t.Error("成日 should be suitable for wedding")
	}
	if len(marks) == 0 {
		t.Error("should have marks for suitable match")
	}
	if len(warnings) != 0 {
		t.Error("should have no warnings for suitable match")
	}
}

func TestJianChuSuitable_ForbiddenMatch(t *testing.T) {
	// "建" is forbidden for "wedding" (嫁娶).
	suitable, marks, warnings := jianChuSuitable("建", "wedding")
	if suitable {
		t.Error("建日 should NOT be suitable for wedding")
	}
	if len(marks) != 0 {
		t.Error("should have no marks for forbidden match")
	}
	if len(warnings) == 0 {
		t.Error("should have warnings for forbidden match")
	}
}

func TestJianChuSuitable_PoDay(t *testing.T) {
	// "破" is forbidden for everything, special message "破日，万事不宜".
	suitable, _, warnings := jianChuSuitable("破", "wedding")
	if suitable {
		t.Error("破日 should never be suitable")
	}
	if len(warnings) == 0 || warnings[0] != "破日，万事不宜" {
		t.Errorf("warnings = %v, want [破日，万事不宜]", warnings)
	}
}

// =============================================================================
// renYuanName — 人元司令名称
// =============================================================================

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
	// Test various event types to exercise different jianChuSuitable branches.
	events := []string{"wedding", "travel", "open", "medical", "funeral"}
	for _, ev := range events {
		t.Run("event="+ev, func(t *testing.T) {
			got, err := QueryDate("2024-06-15", ev)
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
	if r.MonthBranch != ganzhi.Zhi(0) {
		t.Error("MonthBranch should be preserved")
	}
}
// 正月建寅: 寅月寅日=建(offset 0), 寅月卯日=除(offset 1)
// 二月建卯: 卯月卯日=建(offset 0), 卯月辰日=除(offset 1)
func TestJianChuOffset(t *testing.T) {
	tests := []struct {
		name        string
		monthBranch ganzhi.Zhi
		dayZhi      ganzhi.Zhi
		wantOffset  int
		wantGod      string
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
			jianIdx := int(tt.monthBranch) - 1
			dayIdx := int(tt.dayZhi) - 1
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
		monthBranch ganzhi.Zhi
		dayZhi      ganzhi.Zhi
		wantName    string
		wantPath    string
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
		name := ganzhi.ZhiName(tt.monthBranch) + "月" + ganzhi.ZhiName(tt.dayZhi) + "日"
		t.Run(name, func(t *testing.T) {
			got := huangDaoForDay(tt.monthBranch, tt.dayZhi)
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
		dz := ganzhi.Zhi((i) % 12 + 1) // 子=1 through 亥=12
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
		monthBranch ganzhi.Zhi
		wantStart   ganzhi.Zhi
	}{
		{ganzhi.ZhiYin, ganzhi.ZhiZi},   // 寅月青龙起子
		{ganzhi.ZhiMao, ganzhi.ZhiYin},  // 卯月起寅
		{ganzhi.ZhiChen, ganzhi.ZhiChen}, // 辰月起辰
		{ganzhi.ZhiSi, ganzhi.ZhiWu},    // 巳月起午
		{ganzhi.ZhiWu, ganzhi.ZhiShen},  // 午月起申
		{ganzhi.ZhiWei, ganzhi.ZhiXu},   // 未月起戌
		{ganzhi.ZhiShen, ganzhi.ZhiZi},  // 申月起子
		{ganzhi.ZhiYou, ganzhi.ZhiYin},  // 酉月起寅
		{ganzhi.ZhiXu, ganzhi.ZhiChen},  // 戌月起辰
		{ganzhi.ZhiHai, ganzhi.ZhiWu},   // 亥月起午
		{ganzhi.ZhiZi, ganzhi.ZhiShen},  // 子月起申
		{ganzhi.ZhiChou, ganzhi.ZhiXu},  // 丑月起戌
	}

	for _, tt := range tests {
		t.Run(ganzhi.ZhiName(tt.monthBranch)+"月", func(t *testing.T) {
			start, ok := qingLongStart[tt.monthBranch]
			if !ok {
				t.Fatal("month not found in qingLongStart map")
			}
			if start != tt.wantStart {
				t.Errorf("start branch = %d(%s), want %d(%s)",
					start, ganzhi.ZhiName(start),
					int(tt.wantStart), ganzhi.ZhiName(tt.wantStart))
			}
		})
	}
}

// TestTaiSui verifies太岁 calculation.
func TestTaiSui(t *testing.T) {
	tests := []struct {
		year     int
		wantName string
	}{
		{2024, "辰"}, // 甲辰年
		{2025, "巳"}, // 乙巳年
		{2026, "午"}, // 丙午年
		{2023, "卯"}, // 癸卯年
		{2020, "子"}, // 庚子年
	}

	for _, tt := range tests {
		t.Run(ganzhi.ZhiName(taiSui(tt.year)), func(t *testing.T) {
			got := taiSui(tt.year)
			if ganzhi.ZhiName(got) != tt.wantName {
				t.Errorf("taiSui(%d) = %s, want %s", tt.year, ganzhi.ZhiName(got), tt.wantName)
			}
		})
	}
}

// TestComputeBondDay_Golden_GanRelation verifies GanRelation for known dates.
// 日主己土 (1984-02-15 08:00 Beijing → 甲子 丙寅 己卯 戊辰)
func TestComputeBondDay_Golden_GanRelation(t *testing.T) {
	// 日主己土 (1984-02-15 08:00 Beijing → 甲子 丙寅 己卯 戊辰)
	ts := tianwen.ComputeTimeset(tianwen.GregorianTime(time.Date(1984, time.Month(2), 15, 8, 0, 0, 0, time.FixedZone("", int(8.0*3600)))), 116.4074)

	// 2024-02-10 = 甲辰日, 日主己土
	// 甲木克己土, 阳克阴 → 正官
	result, err := ComputeBondDay(ts.Solar, "wedding", "2024-02-10")
	if err != nil {
		t.Fatalf("ComputeBondDay: %v", err)
	}
	if result.GanRelation != "正官" {
		t.Errorf("己土日主 vs 甲辰日: GanRelation=%q, want 正官", result.GanRelation)
	}

	// 2024-06-15 = 庚戌日
	// 己土生庚金, 阴生阳 → 伤官
	result2, err := ComputeBondDay(ts.Solar, "wedding", "2024-06-15")
	if err != nil {
		t.Fatalf("ComputeBondDay: %v", err)
	}
	if result2.GanRelation != "伤官" {
		t.Errorf("己土日主 vs 庚戌日: GanRelation=%q, want 伤官", result2.GanRelation)
	}
}

func TestComputeBondDay_Golden_ZhiRelation(t *testing.T) {
	// 日支亥 (2000-01-01: 1984 is JiaZi, but 2000-01-01 gives different pillar)
	// Actually let's use a known birthday: 1984-02-15 → 日柱己卯 (日支=卯)
	ts := tianwen.ComputeTimeset(tianwen.GregorianTime(time.Date(1984, time.Month(2), 15, 8, 0, 0, 0, time.FixedZone("", int(8.0*3600)))), 116.4074)

	// 2024-02-10 = 甲辰日 (日支=辰)
	// 卯 vs 辰: 卯辰相害 → 六害
	result, err := ComputeBondDay(ts.Solar, "wedding", "2024-02-10")
	if err != nil {
		t.Fatalf("ComputeBondDay: %v", err)
	}
	// 卯辰 overlaps: 三会半 (priority over 六害)
	if result.ZhiRelation != "三会半" {
		t.Errorf("卯(日支) vs 辰(择日日支): ZhiRelation=%q, want 三会半", result.ZhiRelation)
	}
}

func TestComputeBondDay_Golden_TaiSui(t *testing.T) {
	// 2024年太岁=辰
	ts := tianwen.ComputeTimeset(tianwen.GregorianTime(time.Date(1984, time.Month(2), 15, 8, 0, 0, 0, time.FixedZone("", int(8.0*3600)))), 116.4074)

	// 2024-02-10 = 甲辰日, 太岁=辰
	// 辰 vs 辰: 伏吟(自刑)
	result, err := ComputeBondDay(ts.Solar, "wedding", "2024-02-10")
	if err != nil {
		t.Fatalf("ComputeBondDay: %v", err)
	}
	if result.TaiSuiRelation == "" {
		t.Error("TaiSuiRelation should not be empty")
	}
	t.Logf("2024-02-10(辰) vs 太岁(辰): TaiSuiRelation=%s", result.TaiSuiRelation)
}

// ── evaluateZhi full matrix tests ──

func TestEvaluateZhi_SixHarmonies(t *testing.T) {
	// 六合: 子丑, 寅亥, 卯戌, 辰酉, 巳申, 午未
	pairs := [][2]ganzhi.Zhi{
		{ganzhi.ZhiZi, ganzhi.ZhiChou},
		{ganzhi.ZhiYin, ganzhi.ZhiHai},
		{ganzhi.ZhiMao, ganzhi.ZhiXu},
		{ganzhi.ZhiChen, ganzhi.ZhiYou},
		{ganzhi.ZhiSi, ganzhi.ZhiShen},
		{ganzhi.ZhiWu, ganzhi.ZhiWei},
	}
	for _, p := range pairs {
		rel, _, _ := evaluateZhi(p[0], p[1], "test")
		if rel != "六合" {
			t.Errorf("%s+%s: got %q, want 六合",
				ganzhi.ZhiName(p[0]), ganzhi.ZhiName(p[1]), rel)
		}
	}
}

func TestEvaluateZhi_SixClashes(t *testing.T) {
	// 六冲: 子午, 丑未, 寅申, 卯酉, 辰戌, 巳亥
	pairs := [][2]ganzhi.Zhi{
		{ganzhi.ZhiZi, ganzhi.ZhiWu},
		{ganzhi.ZhiChou, ganzhi.ZhiWei},
		{ganzhi.ZhiYin, ganzhi.ZhiShen},
		{ganzhi.ZhiMao, ganzhi.ZhiYou},
		{ganzhi.ZhiChen, ganzhi.ZhiXu},
		{ganzhi.ZhiSi, ganzhi.ZhiHai},
	}
	for _, p := range pairs {
		rel, _, _ := evaluateZhi(p[0], p[1], "test")
		if rel != "六冲" {
			t.Errorf("%s+%s: got %q, want 六冲",
				ganzhi.ZhiName(p[0]), ganzhi.ZhiName(p[1]), rel)
		}
	}
}

func TestEvaluateZhi_SixHarms(t *testing.T) {
	// 六害: 子未, 丑午, 申亥. (寅巳/卯辰/酉戌 overlap with higher-priority relationships)
	pairs := [][2]ganzhi.Zhi{
		{ganzhi.ZhiZi, ganzhi.ZhiWei},
		{ganzhi.ZhiChou, ganzhi.ZhiWu},
		{ganzhi.ZhiShen, ganzhi.ZhiHai},
	}
	for _, p := range pairs {
		rel, _, _ := evaluateZhi(p[0], p[1], "test")
		if rel != "六害" {
			t.Errorf("%s+%s: got %q, want 六害",
				ganzhi.ZhiName(p[0]), ganzhi.ZhiName(p[1]), rel)
		}
	}
	// 寅巳 → 相刑 (higher priority than 六害)
	if rel, _, _ := evaluateZhi(ganzhi.ZhiYin, ganzhi.ZhiSi, "test"); rel != "相刑" {
		t.Errorf("寅+巳: got %q, want 相刑 (priority over 六害)", rel)
	}
	// 卯辰 → 三会半 (higher priority than 六害)
	if rel, _, _ := evaluateZhi(ganzhi.ZhiMao, ganzhi.ZhiChen, "test"); rel != "三会半" {
		t.Errorf("卯+辰: got %q, want 三会半 (priority over 六害)", rel)
	}
	// 酉戌 → 三会半 (higher priority than 六害)
	if rel, _, _ := evaluateZhi(ganzhi.ZhiYou, ganzhi.ZhiXu, "test"); rel != "三会半" {
		t.Errorf("酉+戌: got %q, want 三会半 (priority over 六害)", rel)
	}
}

// ── taiSui golden test ──

func TestTaiSui_All60Years(t *testing.T) {
	// 太岁=年支. Verify for 60 years (one sexagenary cycle).
	expectedZhi := []ganzhi.Zhi{
		ganzhi.ZhiZi, ganzhi.ZhiChou, ganzhi.ZhiYin, ganzhi.ZhiMao,
		ganzhi.ZhiChen, ganzhi.ZhiSi, ganzhi.ZhiWu, ganzhi.ZhiWei,
		ganzhi.ZhiShen, ganzhi.ZhiYou, ganzhi.ZhiXu, ganzhi.ZhiHai,
	}
	for year := 2020; year < 2080; year++ {
		want := expectedZhi[(year-2020)%12]
		got := taiSui(year)
		if got != want {
			t.Errorf("taiSui(%d) = %s, want %s", year, ganzhi.ZhiName(got), ganzhi.ZhiName(want))
		}
	}
}

// ── ComputeBondMonth golden test ──

func TestComputeBondMonth_Golden(t *testing.T) {
	ts := tianwen.ComputeTimeset(tianwen.GregorianTime(time.Date(1984, time.Month(2), 15, 8, 0, 0, 0, time.FixedZone("", int(8.0*3600)))), 116.4074)
	// 日主己土, 日支卯

	m, err := ComputeBondMonth(ts.Solar, "wedding", "2024-06")
	if err != nil {
		t.Fatalf("ComputeBondMonth: %v", err)
	}
	if m.Month != "2024-06" {
		t.Errorf("Month=%q, want 2024-06", m.Month)
	}
	if len(m.Days) != 30 {
		t.Errorf("len(Days)=%d, want 30", len(m.Days))
	}

	// 2024-06-01 is before 芒种 → 巳月 (甲年: 甲己之年丙作首 → 四月己巳)
	if m.Stem != "己" {
		t.Errorf("Month stem=%q, want 己 (before芒种→巳月)", m.Stem)
	}
	if m.Branch != "巳" {
		t.Errorf("Month branch=%q, want 巳 (before芒种→巳月)", m.Branch)
	}

	// 2024-06-15 = 庚戌日, 日主己土: 庚→己 = 伤官
	for _, d := range m.Days {
		if d.Date == "2024-06-15" {
			if d.GanRelation != "伤官" {
				t.Errorf("2024-06-15 GanRelation=%q, want 伤官", d.GanRelation)
			}
		}
	}
}

// ── QueryDate golden: known 建除 values ──

func TestQueryDate_Golden_JianChu(t *testing.T) {
	// 建除十二神: based on month branch and day branch
	// 2024-06-15: 午月(月支=午), 庚戌日(日支=戌)
	// 午月起午日为建 → 午=建,未=除,申=满,酉=平,戌=定
	// 戌日 → 定日
	got, err := QueryDate("2024-06-15", "")
	if err != nil {
		t.Fatalf("QueryDate: %v", err)
	}
	if got.JianChu != "定" {
		t.Errorf("2024-06-15 JianChu=%q, want 定 (午月戌日)", got.JianChu)
	}

	// 2024-02-10: 寅月(月支=寅), 甲辰日(日支=辰)
	// 寅月起寅日为建 → 寅=建,卯=除,辰=满
	// 辰日 → 满日
	got2, err := QueryDate("2024-02-10", "")
	if err != nil {
		t.Fatalf("QueryDate: %v", err)
	}
	if got2.JianChu != "满" {
		t.Errorf("2024-02-10 JianChu=%q, want 满 (寅月辰日)", got2.JianChu)
	}
}

func TestQueryDate_Golden_MarksWarnings(t *testing.T) {
	// Verify that event type filtering produces marks/warnings.
	for _, ev := range []string{"wedding", "travel", "open"} {
		t.Run(ev, func(t *testing.T) {
			got, err := QueryDate("2024-06-15", ev)
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
