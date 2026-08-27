package huangli

import (
	"testing"

	"liki-engine/internal/engine/ganzhi"
)

// =============================================================================
// mansionForDay — 二十八宿
// =============================================================================

func TestMansionForDay_ReferencePoints(t *testing.T) {
	// The standard reference: 甲子日 → 虚宿 (mansion index 10).
	// Formula: (sbIdx + 10) % 28, where sbIdx = SixtyCycleIndex (0-59).
	tests := []struct {
		gan, zhi  int
		sbIdx     int
		wantName  string
		wantGroup string
	}{
		{1, 1, 0, "虚日鼠", "北方玄武"},    // 甲子
		{2, 2, 1, "危月燕", "北方玄武"},    // 乙丑
		{3, 3, 2, "室火猪", "北方玄武"},    // 丙寅
		{4, 4, 3, "壁水貐", "北方玄武"},    // 丁卯
		{5, 5, 4, "奎木狼", "西方白虎"},    // 戊辰
		{7, 7, 6, "胃土雉", "西方白虎"},    // 庚午
		{1, 9, 20, "氐土貉", "东方青龙"},   // 甲申
		{3, 7, 42, "星日马", "南方朱雀"},   // 丙午
		{9, 1, 48, "氐土貉", "东方青龙"},   // 壬子
		{10, 12, 59, "壁水貐", "北方玄武"}, // 癸亥
	}

	for _, tt := range tests {
		g := ganzhi.Gan(tt.gan)
		z := ganzhi.Zhi(tt.zhi)
		name := ganzhi.GanName(g) + ganzhi.ZhiName(z)
		t.Run(name, func(t *testing.T) {
			sbIdx := ganzhi.SixtyCycleIndex(g, z)
			if sbIdx != tt.sbIdx {
				t.Fatalf("SixtyCycleIndex = %d, want %d", sbIdx, tt.sbIdx)
			}
			got := mansionForDay(ganzhi.Zhu{Gan: g, Zhi: z})
			if got.Name != tt.wantName {
				t.Errorf("Name = %q, want %q", got.Name, tt.wantName)
			}
			if got.Group != tt.wantGroup {
				t.Errorf("Group = %q, want %q", got.Group, tt.wantGroup)
			}
		})
	}
}

func TestMansionForDay_Cycle(t *testing.T) {
	// Verify sequential六十甲子 days advance mansions one step at a time.
	for sbIdx := 0; sbIdx < 59; sbIdx++ {
		g1 := ganzhi.Gan((sbIdx % 10) + 1)
		z1 := ganzhi.Zhi((sbIdx % 12) + 1)
		g2 := ganzhi.Gan(((sbIdx + 1) % 10) + 1)
		z2 := ganzhi.Zhi(((sbIdx + 1) % 12) + 1)

		m1 := mansionForDay(ganzhi.Zhu{Gan: g1, Zhi: z1})
		m2 := mansionForDay(ganzhi.Zhu{Gan: g2, Zhi: z2})

		expectedNext := (m1.Index + 1) % 28
		if m2.Index != expectedNext {
			t.Errorf("after %s%s(idx=%d): want mansion index %d, got %s(idx=%d)",
				ganzhi.GanName(g1), ganzhi.ZhiName(z1), sbIdx,
				expectedNext, m2.Name, m2.Index)
		}
	}
}

// =============================================================================
// QueryDate — 黄历单日查询
// =============================================================================

func TestQueryDate_KnownDates(t *testing.T) {
	tests := []struct {
		dateStr     string
		wantDayGan  string
		wantDayZhi  string
		wantMansion string
	}{
		{"2024-02-10", "甲", "辰", "鬼金羊"}, // 春节 2024
		{"2024-01-01", "甲", "子", "虚日鼠"}, // sbIdx=0
		{"2025-01-01", "庚", "午", "胃土雉"}, // sbIdx=6
	}

	for _, tt := range tests {
		t.Run(tt.dateStr, func(t *testing.T) {
			got, err := QueryDate(tt.dateStr)
			if err != nil {
				t.Fatalf("QueryDate: %v", err)
			}
			if ganzhi.GanName(got.RiZhu.Gan) != tt.wantDayGan {
				t.Errorf("RiGan = %s, want %s",
					ganzhi.GanName(got.RiZhu.Gan), tt.wantDayGan)
			}
			if ganzhi.ZhiName(got.RiZhu.Zhi) != tt.wantDayZhi {
				t.Errorf("RiZhi = %s, want %s",
					ganzhi.ZhiName(got.RiZhu.Zhi), tt.wantDayZhi)
			}
			if got.Mansion.Name != tt.wantMansion {
				t.Errorf("Mansion = %q, want %q", got.Mansion.Name, tt.wantMansion)
			}
			if got.Date != tt.dateStr {
				t.Errorf("Date = %q, want %q", got.Date, tt.dateStr)
			}
		})
	}
}

func TestQueryDate_WithEvent(t *testing.T) {
	got, err := QueryDate("2024-02-10")
	if err != nil {
		t.Fatalf("QueryDate: %v", err)
	}
	if got.JianChu == "" {
		t.Error("JianChu should not be empty")
	}
}

func TestQueryDate_NotEmpty(t *testing.T) {
	// Verify key fields are always populated.
	got, err := QueryDate("2024-06-15")
	if err != nil {
		t.Fatalf("QueryDate: %v", err)
	}
	if got.Wuxing == "" {
		t.Error("Wuxing should not be empty")
	}
	if got.NaYin == "" {
		t.Error("NaYin should not be empty")
	}
	if got.JianChu == "" {
		t.Error("JianChu should not be empty")
	}
	if got.XiShen == "" {
		t.Error("XiShen should not be empty")
	}
	if got.CaiShen == "" {
		t.Error("CaiShen should not be empty")
	}
	if got.Mansion.Name == "" {
		t.Error("Mansion should not be empty")
	}
}

func TestQueryDate_InvalidDate(t *testing.T) {
	_, err := QueryDate("not-a-date")
	if err == nil {
		t.Fatal("expected error for invalid date")
	}
}

// =============================================================================
// computeJieQiDepth — 节气深度
// =============================================================================

func TestComputeJieQiDepth_Normal(t *testing.T) {
	d := computeJieQiDepth(2024, 6, 15)
	if d.TermName == "" {
		t.Error("TermName should not be empty")
	}
	if d.NextTermName == "" {
		t.Error("NextTermName should not be empty")
	}
	if d.DaysIn < 0 {
		t.Errorf("DaysIn should be non-negative, got %d", d.DaysIn)
	}
	if d.DaysToNext < 0 {
		t.Errorf("DaysToNext should be non-negative, got %d", d.DaysToNext)
	}
}

func TestComputeJieQiDepth_LiChunBoundary(t *testing.T) {
	tests := []struct {
		y, m, d  int
		wantTerm string
	}{
		{2024, 2, 5, "立春"},  // after立春(~Feb 4)
		{2024, 1, 25, "大寒"}, // between大寒(~Jan 21) and立春
		{2024, 1, 10, "小寒"}, // between小寒(~Jan 6) and大寒
	}

	for _, tt := range tests {
		label := jieQiNames[0] // placeholder for test name
		t.Run(label, func(t *testing.T) {
			d := computeJieQiDepth(tt.y, tt.m, tt.d)

			if d.TermName != tt.wantTerm {
				t.Errorf("%04d-%02d-%02d: TermName = %q, want %q (daysIn=%d, next=%s)",
					tt.y, tt.m, tt.d, d.TermName, tt.wantTerm, d.DaysIn, d.NextTermName)
			}
			// DaysIn must be < 20 (terms are ~15 days apart).
			if d.DaysIn > 20 {
				t.Errorf("%04d-%02d-%02d: DaysIn = %d > 20 — likely year boundary bug",
					tt.y, tt.m, tt.d, d.DaysIn)
			}
			if d.DaysIn < 0 {
				t.Errorf("DaysIn should be non-negative, got %d", d.DaysIn)
			}
		})
	}
}

func TestComputeJieQiDepth_Consistency(t *testing.T) {
	// For each month's 10th, verify next term follows current in cycle.
	for m := 1; m <= 12; m++ {
		d := computeJieQiDepth(2024, m, 10)
		for i := 0; i < 24; i++ {
			if jieQiNames[i] == d.TermName {
				expectedNext := jieQiNames[(i+1)%24]
				if d.NextTermName != expectedNext {
					t.Errorf("2024-%02d-10: %s → next should be %s, got %s",
						m, d.TermName, expectedNext, d.NextTermName)
				}
				break
			}
		}
		// Sum should be reasonable (~15 days between terms).
		if d.DaysIn+d.DaysToNext < 10 || d.DaysIn+d.DaysToNext > 20 {
			t.Logf("2024-%02d-10: %s(%d) → %s(%d), sum=%d",
				m, d.TermName, d.DaysIn, d.NextTermName, d.DaysToNext,
				d.DaysIn+d.DaysToNext)
		}
	}
}

// =============================================================================
// renYuanSiLing — 人元司令分野
// =============================================================================

func TestComputeRenYuanSiLingForDate_Phases(t *testing.T) {
	phases := ganzhi.RenYuanSiLingFenYeForZhi(ganzhi.ZhiYin)
	if len(phases) == 0 {
		t.Skip("RenYuan phases not available for 寅")
	}

	// Day 3 → first phase
	r := computeRenYuanSiLingForDate(ganzhi.ZhiYin, 3)
	if r.Current == nil {
		t.Fatal("Current phase should not be nil")
	}
	if r.Current.GanName != phases[0].GanName {
		t.Errorf("Day 3: want %s, got %s", phases[0].GanName, r.Current.GanName)
	}

	// Day 25 → last phase
	r2 := computeRenYuanSiLingForDate(ganzhi.ZhiYin, 25)
	if r2.Current == nil {
		t.Fatal("Current phase should not be nil")
	}
	lastPhase := phases[len(phases)-1]
	if r2.Current.GanName != lastPhase.GanName {
		t.Errorf("Day 25: want %s, got %s", lastPhase.GanName, r2.Current.GanName)
	}
}

func TestFindCurrentRenYuanSiLingFenYe_Empty(t *testing.T) {
	got := findCurrentRenYuanSiLingFenYe(nil, 5)
	if got != nil {
		t.Error("expected nil for nil phases")
	}
}

// =============================================================================
// 择日 — 喜神/财神/福神/彭祖百忌
// =============================================================================

func TestXiShenDirection(t *testing.T) {
	tests := []struct {
		gan  ganzhi.Gan
		want string
	}{
		{ganzhi.GanJia, "东北"}, {ganzhi.GanJi, "东北"},
		{ganzhi.GanYi, "西北"}, {ganzhi.GanGeng, "西北"},
		{ganzhi.GanBing, "西南"}, {ganzhi.GanXin, "西南"},
		{ganzhi.GanDing, "正南"}, {ganzhi.GanRen, "正南"},
		{ganzhi.GanWu, "东南"}, {ganzhi.GanGui, "东南"},
	}
	for _, tt := range tests {
		t.Run(ganzhi.GanName(tt.gan), func(t *testing.T) {
			if got := xiShenDirection(tt.gan); got != tt.want {
				t.Errorf("xiShen(%s) = %q, want %q",
					ganzhi.GanName(tt.gan), got, tt.want)
			}
		})
	}
}

func TestCaiShenDirection(t *testing.T) {
	tests := []struct {
		gan  ganzhi.Gan
		want string
	}{
		{ganzhi.GanJia, "东北"}, {ganzhi.GanYi, "东北"},
		{ganzhi.GanBing, "正西"}, {ganzhi.GanDing, "正西"},
		{ganzhi.GanWu, "正北"}, {ganzhi.GanJi, "正北"},
		{ganzhi.GanGeng, "正东"}, {ganzhi.GanXin, "正东"},
		{ganzhi.GanRen, "正南"}, {ganzhi.GanGui, "正南"},
	}
	for _, tt := range tests {
		t.Run(ganzhi.GanName(tt.gan), func(t *testing.T) {
			if got := caiShenDirection(tt.gan); got != tt.want {
				t.Errorf("caiShen(%s) = %q, want %q",
					ganzhi.GanName(tt.gan), got, tt.want)
			}
		})
	}
}

func TestFuShenDirection(t *testing.T) {
	tests := []struct {
		gan  ganzhi.Gan
		want string
	}{
		{ganzhi.GanJia, "东南"}, {ganzhi.GanYi, "东南"},
		{ganzhi.GanBing, "西北"}, {ganzhi.GanDing, "正东"},
		{ganzhi.GanWu, "正南"}, {ganzhi.GanJi, "正南"},
		{ganzhi.GanGeng, "西南"}, {ganzhi.GanXin, "西南"},
		{ganzhi.GanRen, "西北"}, {ganzhi.GanGui, "正西"},
	}
	for _, tt := range tests {
		t.Run(ganzhi.GanName(tt.gan), func(t *testing.T) {
			if got := fuShenDirection(tt.gan); got != tt.want {
				t.Errorf("fuShen(%s) = %q, want %q",
					ganzhi.GanName(tt.gan), got, tt.want)
			}
		})
	}
}

func TestPengZuTaboos(t *testing.T) {
	if got := pengZuGanTaboo(ganzhi.GanJia); got != "甲不开仓财物耗散" {
		t.Errorf("甲 taboo = %q", got)
	}
	if got := pengZuZhiTaboo(ganzhi.ZhiZi); got != "子不问卜自惹祸殃" {
		t.Errorf("子 taboo = %q", got)
	}
	// Spot check: 午不苫盖屋主更张
	if got := pengZuZhiTaboo(ganzhi.ZhiWu); got != "午不苫盖屋主更张" {
		t.Errorf("午 taboo = %q", got)
	}
}

func TestZeRiFunctions_InvalidInput(t *testing.T) {
	if got := xiShenDirection(ganzhi.Gan(0)); got != "" {
		t.Errorf("xiShen(0) = %q, want empty", got)
	}
	if got := caiShenDirection(ganzhi.Gan(0)); got != "" {
		t.Errorf("caiShen(0) = %q, want empty", got)
	}
	if got := fuShenDirection(ganzhi.Gan(0)); got != "" {
		t.Errorf("fuShen(0) = %q, want empty", got)
	}
	if got := pengZuGanTaboo(ganzhi.Gan(0)); got != "" {
		t.Errorf("pengZuGanTaboo(0) = %q, want empty", got)
	}
	if got := pengZuZhiTaboo(ganzhi.Zhi(0)); got != "" {
		t.Errorf("pengZuZhiTaboo(0) = %q, want empty", got)
	}
}

// =============================================================================
// evaluateZhi — 地支关系评估
// =============================================================================
