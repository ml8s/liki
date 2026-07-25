package xuankong

import "testing"

func TestComputeAnnual_RuZhong(t *testing.T) {
	tests := []struct {
		year     int
		wantRZ   int  // expected 入中星
	}{
		{year: 1984, wantRZ: 6}, // 1984: 84+84/4=84+21=105, 105%9=6? 
		// Actually let me check: 1984: 84/4=21. 84+21=105. 105%9=6. But 6%9=6, not 0. So 入中=6.
		// Alternative formula: 入中 = (year_suffix + year_suffix/4) % 9
		// 1984: 84 + 21 = 105. 105 % 9 = 6. 入中=6.
		{year: 2024, wantRZ: 3}, // 24+6=30, 30%9=3
		{year: 2025, wantRZ: 4}, // 25+6=31, 31%9=4
		{year: 2026, wantRZ: 5}, // 26+6=32, 32%9=5
		{year: 2000, wantRZ: 9}, // 0+0=0 → 9? Actually 0+0=0, 0%9=0 → wantRZ should be 9
	}
	for _, tt := range tests {
		t.Run("", func(t *testing.T) {
			got := computeNianRuZhong(tt.year)
			if got != tt.wantRZ {
				t.Errorf("computeNianRuZhong(%d) = %d, want %d", tt.year, got, tt.wantRZ)
			}
		})
	}
}

func TestComputeAnnual_Stars(t *testing.T) {
	board := ComputeAnnual(2024)
	if board.Year != 2024 {
		t.Errorf("year=%d, want 2024", board.Year)
	}
	if board.RuZhong != 3 {
		t.Errorf("ru_zhong=%d, want 3", board.RuZhong)
	}
	if len(board.Stars) != 9 {
		t.Errorf("stars=%d, want 9", len(board.Stars))
	}
	// 入中星=3: 中=3, 乾=4, 兑=5, 艮=6, 离=7, 坎=8, 坤=9, 震=1, 巽=2
	expected := []struct {
		pi   int
		star int
	}{
		{5, 3}, // 中宫5, 三碧
		{6, 4}, // 乾6, 四绿
		{7, 5}, // 兑7, 五黄
		{8, 6}, // 艮8, 六白
		{9, 7}, // 离9, 七赤
		{1, 8}, // 坎1, 八白
		{2, 9}, // 坤2, 九紫
		{3, 1}, // 震3, 一白
		{4, 2}, // 巽4, 二黑
	}
	for i, exp := range expected {
		s := board.Stars[i]
		if s.Palace != exp.pi {
			t.Errorf("star[%d].palace=%d, want %d", i, s.Palace, exp.pi)
		}
		if s.Star != exp.star {
			t.Errorf("star[%d].star=%d, want %d", i, s.Star, exp.star)
		}
		// 验证已知吉凶
		if s.Star == 5 && s.Rating != "大凶" {
			t.Errorf("star[%d]=五黄: rating=%q, want 大凶", i, s.Rating)
		}
		if s.Star == 8 && s.Rating != "大吉" {
			t.Errorf("star[%d]=八白: rating=%q, want 大吉", i, s.Rating)
		}
	}
}
