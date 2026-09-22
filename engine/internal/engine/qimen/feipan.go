package qimen

import "liki-engine/internal/engine/ganzhi"

func flyTo(palace GongIndex, delta int) GongIndex {
	return GongIndex(floorMod(int(palace)-1+delta, 9) + 1)
}

func flyDelta(from, to GongIndex) int {
	return floorMod(int(to)-int(from), 9)
}

func placeFlyTianPan(leadZhu ganzhi.Zhu, source GongIndex, dipan [9]ganzhi.Gan) ([9][]TianPanSymbol, GongIndex) {
	searchGan := resolveJiaDunGan(leadZhu.Gan, leadZhu.Zhi)
	landing := findEarthGanPalace(searchGan, dipan)
	delta := flyDelta(source, landing)
	var result [9][]TianPanSymbol
	for home := 1; home <= 9; home++ {
		target := flyTo(GongIndex(home), delta)
		result[int(target)-1] = append(result[int(target)-1], TianPanSymbol{
			Gan: dipan[home-1], Star: palaceStar[home-1],
		})
	}
	return result, landing
}

func placeFlyRenPan(
	leadZhu ganzhi.Zhu, source GongIndex, yinDun bool,
) ([9]DoorIndex, GongIndex) {
	offset := ganzhi.SixtyCycleIndex(leadZhu.Gan, leadZhu.Zhi) % 10
	position := int(source) - 1
	if yinDun {
		position -= offset
	} else {
		position += offset
	}
	landing := GongIndex(floorMod(position, 9) + 1)
	sourceEffective := projectCenter(source)
	delta := flyDelta(sourceEffective, landing)
	var doors [9]DoorIndex
	for _, home := range outerRing {
		target := flyTo(home, delta)
		doors[int(target)-1] = palaceDoor[int(home)-1]
	}
	return doors, landing
}

func placeFlyHidden(dipan [9]ganzhi.Gan, source, landing GongIndex) [9]ganzhi.Gan {
	delta := flyDelta(projectCenter(source), landing)
	var result [9]ganzhi.Gan
	for home := 1; home <= 9; home++ {
		target := flyTo(GongIndex(home), delta)
		result[int(target)-1] = dipan[home-1]
	}
	return result
}

func placeFlyShenPan(yinDun bool, dutyStarPalace GongIndex) [9]SpiritIndex {
	var result [9]SpiritIndex
	if dutyStarPalace < 1 || dutyStarPalace > 9 {
		return result
	}
	start := int(dutyStarPalace) - 1
	for i, spirit := range flySpiritOrder {
		position := start + i
		if yinDun {
			position = start - i
		}
		result[floorMod(position, 9)] = spirit
	}
	return result
}

func projectCenter(palace GongIndex) GongIndex {
	if palace == GongZhong {
		return centerLodgingPalace
	}
	return palace
}

func findEarthGanPalace(gan ganzhi.Gan, dipan [9]ganzhi.Gan) GongIndex {
	for i, candidate := range dipan {
		if candidate == gan {
			return GongIndex(i + 1)
		}
	}
	return 0
}
