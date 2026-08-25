package ziwei

import (
	"testing"

	"liki-engine/internal/engine/tianwen"
)

// TestLeapMonth 验证闰月排盘与 iztro fixLeap 一致：
// 闰月前半月（日≤15）按本月算，后半月（日>15）按下月算。
func TestLeapMonth(t *testing.T) {
	tests := []struct {
		name      string
		day       int
		leap      bool
		wantGong  string // 命宫地支
		wantJu    string
	}{
		{"闰六月十四(前半月按六月)", 14, true, "丑", "火六局"},
		{"闰六月十六(后半月按下月)", 16, true, "寅", "土五局"},
		{"六月十六(非闰对照)", 16, false, "丑", "火六局"},
	}
	for _, tt := range tests {
		lt := tianwen.LunarTime{Year: 2025, Month: 6, Day: tt.day, Leap: tt.leap, Shichen: 7}
		c := ComputeChart(lt, "male")
		mg := c.GongWei[c.MingGong]
		if mg.Zhi.String() != tt.wantGong || c.JuShuName != tt.wantJu {
			t.Errorf("%s: 命宫=%s 局=%s, want 命宫=%s 局=%s",
				tt.name, mg.Zhi.String(), c.JuShuName, tt.wantGong, tt.wantJu)
		}
	}
}
