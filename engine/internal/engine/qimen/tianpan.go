package qimen

import "liki-engine/internal/engine/ganzhi"

// placeTianPan rotates the eight-star heaven plate.
// The duty star moves to the lead gan's earth-palace; if that target is the
// central palace, it is lodged in Kun2. TianQin follows TianRui and carries the
// central palace's earth gan.
func placeTianPan(
	leadZhu ganzhi.Zhu, dutyStar StarIndex, dipan [9]ganzhi.Gan,
) ([9][]TianPanSymbol, GongIndex) {
	var result [9][]TianPanSymbol
	searchGan := resolveJiaDunGan(leadZhu.Gan, leadZhu.Zhi)

	target := 0
	for i, g := range dipan {
		if g == searchGan {
			target = i
			break
		}
	}
	if GongIndex(target+1) == GongZhong {
		target = int(centerLodgingPalace) - 1
	}
	landing := findEarthGanPalace(searchGan, dipan)

	ringDutyStar := dutyStar
	if ringDutyStar == tianQinStar {
		ringDutyStar = tianQinFollows
	}
	dutyIdx := ringIndexForStar(ringDutyStar)
	targetIdx := outerRingIndex(GongIndex(target + 1))

	for offset := 0; offset < 8; offset++ {
		palace := outerRing[(targetIdx+offset)%8]
		star := starOrder8[(dutyIdx+offset)%8]
		idx := int(palace) - 1
		result[idx] = append(result[idx], TianPanSymbol{
			Gan:  dipan[starHomePalace(star)],
			Star: star,
		})
		if tianQinCarriesGan && star == tianQinFollows {
			result[idx] = append(result[idx], TianPanSymbol{
				Gan:  dipan[starHomePalace(tianQinStar)],
				Star: tianQinStar,
			})
		}
	}
	return result, landing
}

func outerRingIndex(p GongIndex) int {
	for i, candidate := range outerRing {
		if candidate == p {
			return i
		}
	}
	return -1
}

func ringIndexForStar(s StarIndex) int {
	for i, candidate := range starOrder8 {
		if candidate == s {
			return i
		}
	}
	return -1
}

func ringIndexForDoor(d DoorIndex) int {
	for i, candidate := range doorOrder {
		if candidate == d {
			return i
		}
	}
	return -1
}
