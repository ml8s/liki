package qimen

import "liki-engine/internal/engine/ganzhi"

// doorOrder is the clockwise order of 8 doors.
var doorOrder = [8]DoorIndex{
	DoorXiu, DoorSheng, DoorShang, DoorDu,
	DoorJing, DoorSi, DoorJingMen, DoorKai,
}

// placeRenPan arranges the human plate: 8 doors on the 9 palaces.
// 值使门 fits to the gong where 时支 (or drive zhi) sits on the earth plate.
// 阳遁顺排、阴遁逆排（与八神阳顺阴逆一致）。
func placeRenPan(driveZhi ganzhi.Zhi, dutyDoor DoorIndex, yinDun bool) [9]DoorIndex {
	var doors [9]DoorIndex

	driveZhiPalace := zhiPalace(driveZhi)

	// Find the index of dutyDoor in doorOrder.
	dutyIdx := 0
	for i, d := range doorOrder {
		if d == dutyDoor {
			dutyIdx = i
			break
		}
	}

	// Place doors from duty door starting at driveZhiPalace.
	// 阳遁顺时针、阴遁逆时针。
	doorIdx := dutyIdx
	startPos := int(driveZhiPalace) - 1 // GongIndex is 1-based, convert to 0-based
	for i := 0; i < 9; i++ {
		var pos int
		if yinDun {
			pos = (startPos - i + 9) % 9
		} else {
			pos = (startPos + i) % 9
		}
		if pos == 4 {
			continue // 中宫无门
		}
		doors[pos] = doorOrder[doorIdx%8]
		doorIdx++
	}

	return doors
}
