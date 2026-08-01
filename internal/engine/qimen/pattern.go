package qimen

import "liki-engine/internal/engine/ganzhi"

// findPatterns detects pan-level奇门格局.
// Per-gong stem interaction patterns are handled by computeGanInteractions.
func findPatterns(pan pan) []Pattern {
	var patterns []Pattern

	// 三奇得使: 乙/丙/丁 at duty door gong.
	if dutyPal := dutyDoorPalace(pan); dutyPal >= 1 {
		p := pan.GongWei[dutyPal-1]
		if p.HeavenStem == ganzhi.GanYi || p.HeavenStem == ganzhi.GanBing || p.HeavenStem == ganzhi.GanDing {
			patterns = append(patterns, Pattern{
				Name: "三奇得使", Description: "吉门得奇，百事可成",
				Auspicious: true, GongWei: []GongIndex{dutyPal},
			})
		}
		// 玉女守门: 值使门宫有丁.
		if hasStemAtPalace(pan, ganzhi.GanDing, dutyPal) {
			patterns = append(patterns, Pattern{
				Name: "玉女守门", Description: "值使门宫有丁，百事大吉",
				Auspicious: true, GongWei: []GongIndex{dutyPal},
			})
		}
	}

	// 天遁: 丙+生门+丁 in the pan.
	if hasStem(pan, ganzhi.GanBing) && hasDoor(pan, DoorSheng) && hasStem(pan, ganzhi.GanDing) {
		patterns = append(patterns, Pattern{
			Name: "天遁", Description: "丙+生门+丁，远行出兵大吉",
			Auspicious: true,
		})
	}
	// 地遁: 乙+开门+己.
	if hasStem(pan, ganzhi.GanYi) && hasDoor(pan, DoorKai) && hasStem(pan, ganzhi.GanJi) {
		patterns = append(patterns, Pattern{
			Name: "地遁", Description: "乙+开门+己，安营立寨大吉",
			Auspicious: true,
		})
	}
	// 人遁: 丁+休门+太阴.
	if hasStem(pan, ganzhi.GanDing) && hasDoor(pan, DoorXiu) && hasSpirit(pan, SpiritTaiYin) {
		patterns = append(patterns, Pattern{
			Name: "人遁", Description: "丁+休门+太阴，和谈联姻得吉",
			Auspicious: true,
		})
	}

	// 伏吟: duty star in its home gong.
	dutyHome := starHomePalace(pan.DutyStar)
	if pal := pan.GongWei[dutyHome]; pal.Star == pan.DutyStar {
		patterns = append(patterns, Pattern{
			Name: "伏吟", Description: "值符归位，凡事闭塞，静守为吉",
			Auspicious: false,
		})
	}

	// 反吟: duty star in opposite gong.
	var dutyPos int
	for i, p := range pan.GongWei {
		if p.Star == pan.DutyStar {
			dutyPos = i
			break
		}
	}
	opposite := 8 - dutyHome
	if dutyPos == opposite {
		patterns = append(patterns, Pattern{
			Name: "反吟", Description: "值符反位，凡事反复，动则有成",
			Auspicious: false,
		})
	}

	return patterns
}

// dutyDoorPalace returns the 1-based gong where the duty door sits.
func dutyDoorPalace(pan pan) GongIndex {
	for i, p := range pan.GongWei {
		if p.Door == pan.DutyDoor {
			return GongIndex(i + 1)
		}
	}
	return 0
}

func hasStem(pan pan, g ganzhi.Gan) bool {
	for _, p := range pan.GongWei {
		if p.EarthStem == g || p.HeavenStem == g {
			return true
		}
	}
	return false
}

func hasStemAtPalace(pan pan, g ganzhi.Gan, gong GongIndex) bool {
	if gong < 1 || gong > 9 {
		return false
	}
	p := pan.GongWei[gong-1]
	return p.EarthStem == g || p.HeavenStem == g
}

func hasDoor(pan pan, d DoorIndex) bool {
	for _, p := range pan.GongWei {
		if p.Door == d {
			return true
		}
	}
	return false
}

func hasSpirit(pan pan, s SpiritIndex) bool {
	for _, p := range pan.GongWei {
		if p.Spirit == s {
			return true
		}
	}
	return false
}
