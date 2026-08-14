package qimen

import (
	"liki-engine/internal/engine/ganzhi"
	"liki-engine/internal/engine/tianwen"
)

// determineJuShu computes the bureau number and yin/yang dun for a given date.
func determineJuShu(year, month, day int, riGan ganzhi.Gan, riZhi ganzhi.Zhi) juShu {
	idx := tianwen.SolarTermIndex(year, month, day)
	entry := solarTermBureau[idx]

	yuan := determineYuan(ganzhi.Zhu{Gan: riGan, Zhi: riZhi})

	var ju int
	var yuanName string
	switch yuan {
	case 0:
		ju, yuanName = entry[0], "上元"
	case 1:
		ju, yuanName = entry[1], "中元"
	default:
		ju, yuanName = entry[2], "下元"
	}

	return juShu{
		Number: ju,
		YinDun: entry[3] == 0,
		Yuan:   yuanName,
	}
}

// determineYuan returns 0=上元, 1=中元, 2=下元 based on the day pillar's position in the 60-cycle.
// determineYuan returns 0=上元, 1=中元, 2=下元 based on the day pillar's position in the 60-cycle.
//
// 三元符头规则（《奇门遁甲》拆补法）：日干支序数 mod 15，
// 0-4（甲子/己卯/甲午/己酉 符头段）→ 上元，5-9（己巳/甲申/己亥/甲寅）→ 中元，
// 10-14（甲戌/己丑/甲辰/己未）→ 下元。即 (idx%15)/5：0上 1中 2下。
func determineYuan(dayZhu ganzhi.Zhu) int {
	dayIdx := ganzhi.SixtyCycleIndex(dayZhu.Gan, dayZhu.Zhi)
	return (dayIdx % 15) / 5
}
