package qimen

import (
	"liki-engine/internal/engine/ganzhi"
	"liki-engine/internal/engine/tianwen"
)

// determineJuShu computes the bureau number and yin/yang dun for a given date.
func determineJuShu(year, month, day int, dayGan ganzhi.Gan, dayZhi ganzhi.Zhi) juShu {
	idx := tianwen.SolarTermIndex(year, month, day)
	entry := solarTermBureau[idx]

	yuan := determineYuan(ganzhi.Zhu{Gan: dayGan, Zhi: dayZhi})

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
func determineYuan(dayZhu ganzhi.Zhu) int {
	dayIdx := ganzhi.SixtyCycleIndex(dayZhu.Gan, dayZhu.Zhi)

	for _, start := range []int{0, 15, 30, 45} {
		if inCycleRange(dayIdx, start, 5) {
			return 0
		}
	}
	for _, start := range []int{10, 25, 40, 55} {
		if inCycleRange(dayIdx, start, 5) {
			return 1
		}
	}
	return 2
}

func inCycleRange(idx, start, length int) bool {
	for i := 0; i < length; i++ {
		if (idx+60)%60 == (start+i)%60 {
			return true
		}
	}
	return false
}
