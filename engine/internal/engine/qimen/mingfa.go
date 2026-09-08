package qimen

import "liki-engine/internal/engine/ganzhi"

func findMingFaDuty(leadZhu ganzhi.Zhu, dipan [9]ganzhi.Gan) duty {
	result := findDuty(leadZhu, dipan)
	if result.Palace == GongZhong {
		result.Door = DoorZhong
	}
	return result
}

func placeMingFaTianPan(
	leadZhu ganzhi.Zhu, source GongIndex, dipan [9]ganzhi.Gan,
) ([9][]TianPanSymbol, GongIndex) {
	return placeFlyTianPan(leadZhu, source, dipan)
}

func placeMingFaRenPan(
	leadZhu ganzhi.Zhu, dutyDoor DoorIndex, source GongIndex, yinDun bool,
) ([9]DoorIndex, GongIndex) {
	offset := ganzhi.SixtyCycleIndex(leadZhu.Gan, leadZhu.Zhi) % 10
	position := int(source) - 1
	if yinDun {
		position -= offset
	} else {
		position += offset
	}
	landing := GongIndex(floorMod(position, 9) + 1)
	start := -1
	for i, door := range mingfaDoorOrder {
		if door == dutyDoor {
			start = i
			break
		}
	}
	if start < 0 {
		return [9]DoorIndex{}, 0
	}
	return placeMingFaSequence(start, int(landing)-1), landing
}

func placeMingFaSequence(start, landing int) [9]DoorIndex {
	var doors [9]DoorIndex
	for i := 0; i < len(mingfaDoorOrder); i++ {
		index := (start + i) % len(mingfaDoorOrder)
		palace := floorMod(landing+i, len(mingfaDoorOrder))
		doors[palace] = mingfaDoorOrder[index]
	}
	return doors
}

func placeMingFaShenPan(yinDun bool, dutyStarPalace GongIndex) [9]SpiritIndex {
	var spirits [9]SpiritIndex
	if dutyStarPalace < 1 || dutyStarPalace > 9 {
		return spirits
	}
	start := int(dutyStarPalace) - 1
	for i, spirit := range mingfaSpiritOrder {
		position := start + i
		if yinDun {
			position = start - i
		}
		spirits[floorMod(position, len(mingfaSpiritOrder))] = spirit
	}
	return spirits
}

func placeMingFaHidden(
	leadZhu ganzhi.Zhu, landing GongIndex, yinDun bool,
) ([9]string, [9]ganzhi.Gan) {
	var pillars [9]string
	var stems [9]ganzhi.Gan
	index := ganzhi.SixtyCycleIndex(leadZhu.Gan, leadZhu.Zhi)
	xunStart := index / 10 * 10
	start := index - xunStart
	palace := int(landing) - 1
	placed := 0
	for step := 0; placed < len(pillars); step++ {
		if step >= 12 {
			return pillars, stems
		}
		candidate := ganzhi.SixtyToZhu(xunStart + (start+step)%10)
		if step > 0 && candidate.Gan == ganzhi.GanJia {
			continue
		}
		name := ganzhi.GanName(candidate.Gan) + ganzhi.ZhiName(candidate.Zhi)
		pillars[palace] = name
		stems[palace] = candidate.Gan
		placed++
		if yinDun {
			palace = floorMod(palace-1, len(pillars))
		} else {
			palace = floorMod(palace+1, len(pillars))
		}
	}
	return pillars, stems
}
