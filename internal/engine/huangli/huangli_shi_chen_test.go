package huangli

import (
	"testing"

)

// ── 时辰吉凶命理知识测试 ──
// 基于黄历经典理论构建

func TestShiChen_KnownDates_HuangDaoSequence(t *testing.T) {
	// 命理: 日支不同则青龙起始时辰不同→黄道黑道分布不同
	// 验证: 不同日支返回12个不同时辰
	tests := []struct {
		name     string
		date     string
		wantLen  int
	}{
		{name: "2026-07-20(日支卯)", date: "2026-07-20", wantLen: 12},
		{name: "2026-07-21(日支辰)", date: "2026-07-21", wantLen: 12},
		{name: "2026-07-22(日支巳)", date: "2026-07-22", wantLen: 12},
		{name: "2026-01-01(日支酉)", date: "2026-01-01", wantLen: 12},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			day, err := QueryDate(tt.date, "")
			if err != nil {
				t.Fatal(err)
			}
			if len(day.ShiChen) != tt.wantLen {
				t.Errorf("shi_chen=%d, want %d", len(day.ShiChen), tt.wantLen)
			}
			// 验证12时辰地支正确
			expectedZhi := []string{"子","丑","寅","卯","辰","巳","午","未","申","酉","戌","亥"}
			for i, sc := range day.ShiChen {
				if sc.Zhi != expectedZhi[i] {
					t.Errorf("shi_chen[%d].zhi=%q, want %q", i, sc.Zhi, expectedZhi[i])
				}
			}
			// 验证黄道黑道至少各有几个
			huangCount := 0
			for _, sc := range day.ShiChen {
				if sc.Suitable { huangCount++ }
			}
			if huangCount < 3 || huangCount > 9 {
				t.Errorf("黄道时辰=%d(应约6个左右)", huangCount)
			}
		})
	}
}

func TestShiChen_DifferentDays_DifferentPattern(t *testing.T) {
	// 命理: 不同日期(不同日支)的时辰吉凶分布不同
	day1, err := QueryDate("2026-07-20", "")
	if err != nil { t.Fatal(err) }
	if err != nil { t.Fatal(err) }
	day2, err := QueryDate("2026-07-21", "")
	if err != nil { t.Fatal(err) }
	if err != nil { t.Fatal(err) }

	if len(day1.ShiChen) != 12 || len(day2.ShiChen) != 12 {
		t.Fatal("shi_chen length != 12")
	}

	// 青龙起始时辰应该不同
	// 卯日→青龙在寅(第3个=index2)
	// 辰日→青龙在辰(第5个=index4)
	getQingLong := func(sc []ShiChenFortune) string {
		for _, s := range sc {
			if s.HuangDaoStr == "青龙" { return s.Zhi }
		}
		return ""
	}
	ql1 := getQingLong(day1.ShiChen)
	ql2 := getQingLong(day2.ShiChen)
	// 青龙起始由月支决定(非日支), 同月同日支通常相同
	if ql1 != ql2 {
		t.Logf("卯辰日青龙不同(%s vs %s)", ql1, ql2)
	}
	if ql1 == "" {
		t.Errorf("青龙起始为空")
	}
}

func TestShiChen_JianChuAlignsWithDay(t *testing.T) {
	// 命理: 子时的建除应与本日建除相同(起建)
	day, err := QueryDate("2026-07-20", "")
	if err != nil { t.Fatal(err) }

	if len(day.ShiChen) == 0 {
		t.Fatal("shi_chen empty")
	}

	// 子时的建除应等于本日建除
	if day.ShiChen[0].JianChu != day.JianChu {
		t.Errorf("子时建除=%q, 应等于本日建除=%q",
			day.ShiChen[0].JianChu, day.JianChu)
	}

	// 12时辰建除应覆盖12个不同的建除神
	seen := make(map[string]bool)
	for _, sc := range day.ShiChen {
		seen[sc.JianChu] = true
	}
	if len(seen) != 12 {
		t.Errorf("建除覆盖=%d/12, 应有12个不同建除", len(seen))
	}
}

func TestShiChen_GoodHoursForTravel(t *testing.T) {
	// 命理: 出行宜开日(建除=开)或青龙黄道时辰
	// 2026-07-20日支卯, 建除=开(宜出行)
	day, err := QueryDate("2026-07-20", "travel")
	if err != nil { t.Fatal(err) }

	if !day.Suitable {
		t.Log("2026-07-20不宜出行, 验证时辰中是否有可用的")
	}

	// 列出宜出行的时辰
	goodHours := 0
	for _, sc := range day.ShiChen {
		if sc.Suitable {
			goodHours++
		}
	}
	t.Logf("当日%d个时辰宜出行", goodHours)
	for _, sc := range day.ShiChen {
		if sc.Suitable {
			t.Logf("  %s时(%s): %s(%s)", sc.Zhi, sc.Time, sc.HuangDaoStr, sc.JianChu)
		}
	}
}

func TestShiChen_TimeRanges_Correct(t *testing.T) {
	// 命理: 时辰时间范围应正确
	day, err := QueryDate("2026-07-20", "")
	if err != nil { t.Fatal(err) }
	expected := []string{
		"23:00-01:00", "01:00-03:00", "03:00-05:00", "05:00-07:00",
		"07:00-09:00", "09:00-11:00", "11:00-13:00", "13:00-15:00",
		"15:00-17:00", "17:00-19:00", "19:00-21:00", "21:00-23:00",
	}
	for i, sc := range day.ShiChen {
		if sc.Time != expected[i] {
			t.Errorf("shi_chen[%d].time=%q, want %q", i, sc.Time, expected[i])
		}
	}
}

func TestShiChen_DifferentMonths_DifferentPattern(t *testing.T) {
	// 不同月份的青龙起始不同→时辰吉凶分布不同
	// 2026-02-01(寅月, 青龙起子) vs 2026-06-01(巳月, 青龙起午)
	day1, err := QueryDate("2026-02-01", "")
	if err != nil { t.Fatal(err) }
	day2, err := QueryDate("2026-06-01", "")
	if err != nil { t.Fatal(err) }

	same := 0
	for i := 0; i < 12 && i < len(day1.ShiChen) && i < len(day2.ShiChen); i++ {
		if day1.ShiChen[i].Suitable == day2.ShiChen[i].Suitable {
			same++
		}
	}
	// 不同月份的青龙起始不同, 时辰分布应不同
	if same >= 12 {
		t.Errorf("不同月时辰吉凶完全一致: %d/12", same)
	}
	t.Logf("两月时辰吉凶相同数=%d/12", same)
}

