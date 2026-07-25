package fengshui

import "testing"

func Test24MountainsYinYang(t *testing.T) {
	// Build index of天元龙 by trigram.
	tianYinYang := make(map[string]string)
	for _, m := range Mountains24Table {
		if m.YuanLong == "天元龙" {
			tianYinYang[m.Trigram] = m.YinYang
		}
	}

	// Verify天元龙 follows四正阴/四隅阳.
	for _, m := range Mountains24Table {
		if m.YuanLong != "天元龙" {
			continue
		}
		var expected string
		switch m.Trigram {
		case "坎", "震", "离", "兑":
			expected = "阴" // 四正卦天元=阴
		default:
			expected = "阳" // 四隅卦(乾坤艮巽)天元=阳
		}
		if m.YinYang != expected {
			t.Errorf("%s(%s 天元): yinyang=%s, want %s",
				m.Name, m.Trigram, m.YinYang, expected)
		}
	}

	// Verify人元龙 = 天元龙 (same yin-yang).
	for _, m := range Mountains24Table {
		if m.YuanLong != "人元龙" {
			continue
		}
		expected := tianYinYang[m.Trigram]
		if m.YinYang != expected {
			t.Errorf("%s(%s 人元): yinyang=%s, want %s (应同天元)",
				m.Name, m.Trigram, m.YinYang, expected)
		}
	}

	// Verify地元龙 = 天元龙 reversed.
	for _, m := range Mountains24Table {
		if m.YuanLong != "地元龙" {
			continue
		}
		tianYY := tianYinYang[m.Trigram]
		var expected string
		if tianYY == "阳" {
			expected = "阴"
		} else {
			expected = "阳"
		}
		if m.YinYang != expected {
			t.Errorf("%s(%s 地元): yinyang=%s, want %s (应反天元)",
				m.Name, m.Trigram, m.YinYang, expected)
		}
	}
}

// TestStarTable verifies紫白飞星 data integrity.
func TestStarTable(t *testing.T) {
	// Verify 1-9 all present and non-empty.
	for i := 1; i <= 9; i++ {
		s := StarTable[i]
		if s.Number != i {
			t.Errorf("star %d: number = %d", i, s.Number)
		}
		if s.Name == "" {
			t.Errorf("star %d: empty name", i)
		}
		if s.Color == "" {
			t.Errorf("star %d: empty color", i)
		}
	}

	// Verify auspicious stars: 1,4,6,8,9 (一白四绿六白八白九紫)
	auspicious := map[int]bool{1: true, 4: true, 6: true, 8: true, 9: true}
	for i := 1; i <= 9; i++ {
		if auspicious[i] != StarTable[i].Auspicious {
			t.Errorf("star %d (%s): auspicious=%v, want %v",
				i, StarTable[i].Name, StarTable[i].Auspicious, auspicious[i])
		}
	}
}

// TestPalaceTable verifies九宫 data integrity.
func TestPalaceTable(t *testing.T) {
	palaceNames := map[int]string{
		1: "坎", 2: "坤", 3: "震", 4: "巽", 5: "中",
		6: "乾", 7: "兑", 8: "艮", 9: "离",
	}
	for i := 1; i <= 9; i++ {
		p := PalaceTable[i]
		if p.Number != i {
			t.Errorf("palace %d: number = %d", i, p.Number)
		}
		if p.Name != palaceNames[i] {
			t.Errorf("palace %d: name = %s, want %s", i, p.Name, palaceNames[i])
		}
	}
}

// TestLuoshuFlyOrder verifies the飞星 order skips center.
func TestLuoshuFlyOrder(t *testing.T) {
	// The 8-cell fly order: 6,7,8,9,1,2,3,4
	expected := [8]int{6, 7, 8, 9, 1, 2, 3, 4}
	if LuoshuFlyOrder != expected {
		t.Errorf("LuoshuFlyOrder = %v, want %v", LuoshuFlyOrder, expected)
	}
}
