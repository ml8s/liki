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
func computeSanYuan(monthZhu ganzhi.Zhu, yearStem ganzhi.Gan, birthMonth, birthHour int) SanYuan {
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

	// 命宫: 以子为正月逆数至生月
	monthOnZi := (1 - (birthMonth - 1) + 12) % 12
	if monthOnZi == 0 {
		monthOnZi = 12
	}

	hourBranch := ganzhi.Zhi((birthHour+1)/2%12 + 1)
	mgBranch := (monthOnZi + int(hourBranch) - 1) % 12
	if mgBranch == 0 {
		mgBranch = 12
	}

	mgMonthIdx := ((mgBranch - 3 + 12) % 12) + 1
	mgStem := (int(yearStem)*2 + mgMonthIdx) % 10
	if mgStem == 0 {
		mgStem = 10
	}

	mingGong := ganzhi.Zhu{Gan: ganzhi.Gan(mgStem), Zhi: ganzhi.Zhi(mgBranch)}

	// 身宫: 以子为正月顺数至生月
	shenStart := birthMonth

	sgBranch := (shenStart - int(hourBranch) + 1 + 12) % 12
	if sgBranch == 0 {
		sgBranch = 12
	}

	sgMonthIdx := ((sgBranch - 3 + 12) % 12) + 1
	sgStem := (int(yearStem)*2 + sgMonthIdx) % 10
	if sgStem == 0 {
		sgStem = 10
	}

	shenGong := ganzhi.Zhu{Gan: ganzhi.Gan(sgStem), Zhi: ganzhi.Zhi(sgBranch)}

	return SanYuan{TaiYuan: taiYuan, MingGong: mingGong, ShenGong: shenGong}
}
