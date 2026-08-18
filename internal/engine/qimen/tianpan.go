package qimen

import "liki-engine/internal/engine/ganzhi"

// starOrder8 is the clockwise order of 8 stars (starting from 天蓬),
// excluding 天禽 which 寄坤2 (resides with 天芮).
var starOrder8 = [8]StarIndex{
	StarTianPeng, StarTianRui, StarTianChong, StarTianFu,
	StarTianXin, StarTianZhu, StarTianRen, StarTianYing,
}

// placeTianPan arranges the heaven plate: 8 stars and their associated heaven stems.
// 值符星 fits to the gong where 时干 sits on the earth plate.
// Other stars follow clockwise, skipping 中5 (the void central palace).
// 天禽星寄坤2 — always in the same gong as 天芮.
// Heaven stem at each gong = the earth stem of the star's home gong.
// If the duty star is 天禽 (旬首 in 中5), it is treated as 天芮 (寄坤2), and 天禽
// then follows 天芮's position.
func placeTianPan(driveZhu ganzhi.Zhu, dutyStar StarIndex, dipan [9]ganzhi.Gan) ([9]StarIndex, [9]ganzhi.Gan) {
	var stars [9]StarIndex
	var stems [9]ganzhi.Gan

	// 甲遁于旬首 — when the driving stem is 甲, use the xunShou instead.
	searchGan := driveZhu.Gan
	if driveZhu.Gan == ganzhi.GanJia {
		searchGan = findXunShou(driveZhu)
	}

	// Find the gong where the driving stem sits on the earth plate.
	driveGanPalace := 0
	for i := 0; i < 9; i++ {
		if dipan[i] == searchGan {
			driveGanPalace = i
			break
		}
	}
	// 中5寄坤2：时干（或旬首）落中5时，值符星寄于坤2（与地盘寄宫一致）。
	if driveGanPalace == 4 {
		driveGanPalace = 1 // 中5 → 坤2
	}

	// 天禽寄坤2: treat duty 天禽 as 天芮 so it flies like the 坤2 star.
	duty8 := dutyStar
	if dutyStar == StarTianQin {
		duty8 = StarTianRui
	}

	// Find the index of duty8 in starOrder8.
	dutyIdx := 0
	for i, s := range starOrder8 {
		if s == duty8 {
			dutyIdx = i
			break
		}
	}

	// Place 8 stars clockwise from duty star over the 8 non-central palaces
	// (skipping 中5, the void central palace).
	step := 0
	for i := 0; i < 9; i++ {
		pos := (driveGanPalace + i) % 9
		if pos == 4 {
			continue // 中5 虚空
		}
		star := starOrder8[(dutyIdx+step)%8]
		step++
		stars[pos] = star

		// Heaven stem = earth stem from the star's home gong.
		homePalace := starHomePalace(star)
		stems[pos] = dipan[homePalace]
	}

	// 天禽寄坤2: 天禽不占独立星位，隐含寄于天芮所在宫（与天芮同宫）。
	// 排盘只列天芮，天禽随天芮（主流转盘法）。

	return stars, stems
}
