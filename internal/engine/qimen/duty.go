package qimen

import "liki-engine/internal/engine/ganzhi"

// 六甲旬首 → 六仪
var liuJiaLiuYi = [6]ganzhi.Gan{ganzhi.GanWu, ganzhi.GanJi, ganzhi.GanGeng, ganzhi.GanXin, ganzhi.GanRen, ganzhi.GanGui}

// palaceStar maps gong index (0-based) to its resident star.
var palaceStar = [9]StarIndex{
	StarTianPeng,  // 坎1
	StarTianRui,   // 坤2
	StarTianChong, // 震3
	StarTianFu,    // 巽4
	StarTianQin,   // 中5
	StarTianXin,   // 乾6
	StarTianZhu,   // 兑7
	StarTianRen,   // 艮8
	StarTianYing,  // 离9
}

// palaceDoor maps gong index (0-based) to its resident door.
var palaceDoor = [9]DoorIndex{
	DoorXiu,     // 坎1=休
	DoorSi,      // 坤2=死
	DoorShang,   // 震3=伤
	DoorDu,      // 巽4=杜
	DoorSi,      // 中5=寄坤2=死
	DoorKai,     // 乾6=开
	DoorJingMen, // 兑7=惊
	DoorSheng,   // 艮8=生
	DoorJing,    // 离9=景
}

// findDuty determines 值符星 and 值使门 from the driving pillar and earth plate.
func findDuty(driveZhu ganzhi.Zhu, dipan [9]ganzhi.Gan) duty {
	xunShou := findXunShou(driveZhu)

	var targetPalace int
	for i := 0; i < 9; i++ {
		if dipan[i] == xunShou {
			targetPalace = i
			break
		}
	}

	return duty{
		Star: palaceStar[targetPalace],
		Door: palaceDoor[targetPalace],
	}
}

// findXunShou returns the 六仪 that corresponds to the 六甲旬 of the given pillar.
func findXunShou(zhu ganzhi.Zhu) ganzhi.Gan {
	idx := ganzhi.SixtyCycleIndex(zhu.Gan, zhu.Zhi) // 0-59
	return liuJiaLiuYi[idx/10]
}
