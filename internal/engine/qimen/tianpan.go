package qimen

import "liki-engine/internal/engine/ganzhi"

// starOrder is the clockwise order of 9 stars (starting from 天蓬).
var starOrder = [9]StarIndex{
	StarTianPeng, StarTianRui, StarTianChong,
	StarTianFu, StarTianQin, StarTianXin,
	StarTianZhu, StarTianRen, StarTianYing,
}

// placeTianPan arranges the heaven plate: 9 stars and their associated heaven stems.
// 值符星 fits to the palace where 时干 sits on the earth plate.
// Other stars follow clockwise.
// Heaven stem at each palace = the earth stem of the star's original palace.
func placeTianPan(driveZhu ganzhi.Zhu, dutyStar StarIndex, dipan [9]ganzhi.Gan) ([9]StarIndex, [9]ganzhi.Gan) {
	var stars [9]StarIndex
	var stems [9]ganzhi.Gan

	// 甲遁于旬首 — when the driving stem is 甲, use the xunShou instead.
	searchGan := driveZhu.Gan
	if driveZhu.Gan == ganzhi.GanJia {
		searchGan = findXunShou(driveZhu)
	}

	// Find the palace where the driving stem sits on the earth plate.
	driveGanPalace := 0
	for i := 0; i < 9; i++ {
		if dipan[i] == searchGan {
			driveGanPalace = i
			break
		}
	}

	// Find the index of dutyStar in starOrder.
	dutyIdx := 0
	for i, s := range starOrder {
		if s == dutyStar {
			dutyIdx = i
			break
		}
	}

	// Place stars clockwise from duty star starting at driveGanPalace.
	for i := 0; i < 9; i++ {
		pos := (driveGanPalace + i) % 9
		star := starOrder[(dutyIdx+i)%9]
		stars[pos] = star

		// Heaven stem = earth stem from the star's home palace.
		homePalace := starHomePalace(star)
		stems[pos] = dipan[homePalace]
	}

	return stars, stems
}
