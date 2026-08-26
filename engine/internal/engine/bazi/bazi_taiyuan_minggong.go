package bazi

import (
	"liki-engine/internal/engine/ganzhi"
)

// SanYuan holds the 胎元, 命宫 and 身宫 pillars for a chart.
type SanYuan struct {
	TaiYuan  ganzhi.Zhu `json:"tai_yuan"`
	MingGong ganzhi.Zhu `json:"ming_gong"`
	ShenGong ganzhi.Zhu `json:"shen_gong"`
}

// computeSanYuan computes the 三垣 (three palaces): 胎元, 命宫, 身宫.
// 命宫/身宫采用 lunar-typescript (6tail) 标准算法：
//   - MONTH_ZHI 寅起索引（寅=1 卯=2 ... 丑=12）
//   - ZHI 子起索引（子=1 丑=2 ... 亥=12）
//   - 命宫 offset = 14 - (月支寅起 + 时支寅起)，≥14 时用 26 - 和
//   - 身宫 offset = (月支寅起 + 时支子起) mod 12
//   - 天干 = 年干*2 + offset（mod 10，0→10）
func computeSanYuan(monthZhu ganzhi.Zhu, yearStem ganzhi.Gan, timeZhi ganzhi.Zhi) SanYuan {
	// 胎元: month stem+1 (mod 10), month branch+3 (mod 12)
	tyStem := int(monthZhu.Gan) + 1
	if tyStem > 10 {
		tyStem -= 10
	}
	tyBranch := int(monthZhu.Zhi) + 3
	if tyBranch > 12 {
		tyBranch -= 12
	}
	taiYuan := ganzhi.Zhu{Gan: ganzhi.Gan(tyStem), Zhi: ganzhi.Zhi(tyBranch)}

	// 月支 → 寅起索引（寅=1 卯=2 ... 丑=12）
	monthZhiIndex := (int(monthZhu.Zhi)+9)%12 + 1
	// 时支 → 寅起索引
	timeZhiIndexM := (int(timeZhi)+9)%12 + 1
	// 时支 → 子起索引（子=1 ... 亥=12，即 ganzhi.Zhi 原值）
	timeZhiIndexZ := int(timeZhi)

	// 命宫
	mgOffset := monthZhiIndex + timeZhiIndexM
	if mgOffset >= 14 {
		mgOffset = 26 - mgOffset
	} else {
		mgOffset = 14 - mgOffset
	}
	mgZhi := (mgOffset+1)%12 + 1
	mgStem := (int(yearStem)*2 + mgOffset) % 10
	if mgStem == 0 {
		mgStem = 10
	}
	mingGong := ganzhi.Zhu{Gan: ganzhi.Gan(mgStem), Zhi: ganzhi.Zhi(mgZhi)}

	// 身宫
	sgOffset := monthZhiIndex + timeZhiIndexZ
	if sgOffset > 12 {
		sgOffset -= 12
	}
	sgZhi := (sgOffset+1)%12 + 1
	sgStem := (int(yearStem)*2 + sgOffset) % 10
	if sgStem == 0 {
		sgStem = 10
	}
	shenGong := ganzhi.Zhu{Gan: ganzhi.Gan(sgStem), Zhi: ganzhi.Zhi(sgZhi)}

	return SanYuan{TaiYuan: taiYuan, MingGong: mingGong, ShenGong: shenGong}
}
