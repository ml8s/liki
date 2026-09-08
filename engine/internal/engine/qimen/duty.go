package qimen

import "liki-engine/internal/engine/ganzhi"

// findDuty determines 值符星 and 值使门 from the lead pillar and earth plate.
func findDuty(leadZhu ganzhi.Zhu, dipan [9]ganzhi.Gan) duty {
	xunShou := findXunShou(leadZhu)

	var targetPalace int
	for i := 0; i < 9; i++ {
		if dipan[i] == xunShou {
			targetPalace = i
			break
		}
	}

	palace := GongIndex(targetPalace + 1)
	doorPalace := palace
	if doorPalace == GongZhong {
		doorPalace = centerLodgingPalace
	}
	return duty{
		Star:   palaceStar[targetPalace],
		Door:   palaceDoor[int(doorPalace)-1],
		Palace: palace,
	}
}

// findXunShou returns the 六仪 that corresponds to the 六甲旬 of the given pillar.
func findXunShou(zhu ganzhi.Zhu) ganzhi.Gan {
	idx := ganzhi.SixtyCycleIndex(zhu.Gan, zhu.Zhi) // 0-59
	return liuJiaLiuYi[idx/10]
}
