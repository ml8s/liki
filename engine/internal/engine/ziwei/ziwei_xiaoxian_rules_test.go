package ziwei

import (
	"testing"

	"liki-engine/internal/engine/ganzhi"
)

// 小限起宫（anXingXiaoXian 的三合局起宫）固化测试——入表前锁定。
// 《紫微斗数》：小限以生年支三合局定起宫（寅午戌→辰、申子辰→戌、
// 巳酉丑→未、亥卯未→丑），男顺女逆，虚岁起。

func TestXiaoXianStartPalaceByYearBranch(t *testing.T) {
	cases := []struct {
		nianZhi ganzhi.Zhi
		want    int
	}{
		{ganzhi.ZhiYin, 2},  // 寅（午戌）→ 辰=2
		{ganzhi.ZhiWu, 2},   // 午（寅戌）→ 辰=2
		{ganzhi.ZhiXu, 2},   // 戌（寅午）→ 辰=2
		{ganzhi.ZhiShen, 8}, // 申（子辰）→ 戌=8
		{ganzhi.ZhiZi, 8},   // 子（申辰）→ 戌=8
		{ganzhi.ZhiChen, 8}, // 辰（申子）→ 戌=8
		{ganzhi.ZhiSi, 5},   // 巳（酉丑）→ 未=5
		{ganzhi.ZhiYou, 5},  // 酉（巳丑）→ 未=5
		{ganzhi.ZhiChou, 5}, // 丑（巳酉）→ 未=5
		{ganzhi.ZhiHai, 11}, // 亥（卯未）→ 丑=11
		{ganzhi.ZhiMao, 11}, // 卯（亥未）→ 丑=11
		{ganzhi.ZhiWei, 11}, // 未（亥卯）→ 丑=11
	}
	for _, tc := range cases {
		got := xiaoXianStartIndex(tc.nianZhi)
		if got != tc.want {
			t.Errorf("年支 %d 小限起宫 = %d, want %d", tc.nianZhi, got, tc.want)
		}
	}
}