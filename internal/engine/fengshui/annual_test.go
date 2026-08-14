package fengshui

import "testing"

// 紫白流年飞星 — 命理锚点测试。
// 口诀：上元甲子(1864)一白入中，中元甲子(1924)四绿入中，下元甲子(1984)七赤入中，逐年逆行。

func TestComputeAnnualFlyingStars_RuZhong_KnownYears(t *testing.T) {
	tests := []struct {
		year     int
		wantName string
	}{
		{1864, "一白贪狼"}, // 上元甲子
		{1924, "四绿文曲"}, // 中元甲子
		{1984, "七赤破军"}, // 下元甲子
		{1985, "六白武曲"}, // 逐年逆行
		{1990, "一白贪狼"},
		{1993, "七赤破军"}, // 9 年周期回到七赤
		{2024, "三碧禄存"},
		{2025, "二黑巨门"}, // 2024 三碧 → 逆行 → 2025 二黑
	}
	for _, tt := range tests {
		board := ComputeAnnualFlyingStars(tt.year)
		if board.RuZhong != tt.wantName {
			t.Errorf("year %d: 入中星 = %s, want %s", tt.year, board.RuZhong, tt.wantName)
		}
		if board.Year != tt.year {
			t.Errorf("year %d: board.Year = %d", tt.year, board.Year)
		}
		if len(board.GongWei) != 9 {
			t.Errorf("year %d: gong_wei len = %d, want 9", tt.year, len(board.GongWei))
		}
	}
}

func TestComputeAnnualFlyingStars_Pre1864(t *testing.T) {
	// 1864 前按 60 年甲子周期回推：1804 = 1864-60（上元甲子前推一轮，七赤入中）
	board := ComputeAnnualFlyingStars(1804)
	if board.RuZhong != "七赤破军" {
		t.Errorf("year 1804: 入中星 = %s, want 七赤破军", board.RuZhong)
	}
	// 1844 = 1864 前 20 年，逆推 20 年：一白 → 九紫 → 八白 → 七赤 → 六白 → 五黄 → 四绿 → 三碧
	board = ComputeAnnualFlyingStars(1844)
	if board.RuZhong != "三碧禄存" {
		t.Errorf("year 1844: 入中星 = %s, want 三碧禄存", board.RuZhong)
	}
}

func TestComputeAnnualFlyingStars_Distribution_1984(t *testing.T) {
	// 1984 七赤入中，按洛书顺序飞布：
	// 中5=七赤、乾6=八白、兑7=九紫、艮8=一白、离9=二黑、坎1=三碧、坤2=四绿、震3=五黄、巽4=六白
	want := map[int]int{1: 3, 2: 4, 3: 5, 4: 6, 5: 7, 6: 8, 7: 9, 8: 1, 9: 2}
	board := ComputeAnnualFlyingStars(1984)
	if board.RuZhong != "七赤破军" {
		t.Fatalf("1984 入中星 = %s, want 七赤破军", board.RuZhong)
	}
	seen := map[int]bool{}
	for _, s := range board.GongWei {
		wantXing, ok := want[s.GongNum]
		if !ok {
			t.Errorf("unexpected gong_num %d", s.GongNum)
			continue
		}
		if s.Xing != wantXing {
			t.Errorf("gong %d: 飞星 = %d(%s), want %d", s.GongNum, s.Xing, s.XingName, wantXing)
		}
		if seen[s.GongNum] {
			t.Errorf("duplicate gong %d", s.GongNum)
		}
		seen[s.GongNum] = true
		// 中宫标记
		if s.GongNum == 5 && !s.RuZhong {
			t.Errorf("gong 5 should be 入中")
		}
		if s.GongNum != 5 && s.RuZhong {
			t.Errorf("gong %d should not be 入中", s.GongNum)
		}
	}
	for i := 1; i <= 9; i++ {
		if !seen[i] {
			t.Errorf("missing gong %d", i)
		}
	}
}

func TestComputeAnnualFlyingStars_StarRatings(t *testing.T) {
	// 星的吉凶是命理经典定论（符号固有属性），此处校验共享评级表与星表一致
	board := ComputeAnnualFlyingStars(2024) // 三碧入中
	for _, s := range board.GongWei {
		ref := StarByNumber(s.Xing)
		if s.XingName != ref.Name {
			t.Errorf("gong %d: xing_name = %s, want %s", s.GongNum, s.XingName, ref.Name)
		}
		if s.Wuxing != ref.Element.String() {
			t.Errorf("gong %d: wuxing = %s, want %s", s.GongNum, s.Wuxing, ref.Element.String())
		}
		if s.Rating != StarRatings[s.Xing] {
			t.Errorf("gong %d: rating = %s, want %s", s.GongNum, s.Rating, StarRatings[s.Xing])
		}
	}
}
