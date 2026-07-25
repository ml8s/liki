package qimen

import "liki-engine/internal/engine/ganzhi"

// doorOrder is the clockwise order of 8 doors.
var doorOrder = [8]DoorIndex{
	DoorXiu, DoorSheng, DoorShang, DoorDu,
	DoorJing, DoorSi, DoorJingMen, DoorKai,
}

// placeRenPan arranges the human plate: 8 doors on the 9 palaces.
// 值使门 fits to the palace where 时支 (or drive branch) sits on the earth plate.
func placeRenPan(driveZhi ganzhi.Zhi, dutyDoor DoorIndex) [9]DoorIndex {
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

	// Place doors clockwise from duty door starting at driveZhiPalace.
	doorIdx := dutyIdx
	startPos := int(driveZhiPalace) - 1 // PalaceIndex is 1-based, convert to 0-based
	for i := 0; i < 9; i++ {
		pos := (startPos + i) % 9
		if pos == 4 {
			continue // 中宫无门
		}
		doors[pos] = doorOrder[doorIdx%8]
		doorIdx++
	}

	return doors
}
