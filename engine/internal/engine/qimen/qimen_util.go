package qimen

import "liki-engine/internal/engine/ganzhi"

func zhiPalace(z ganzhi.Zhi) GongIndex {
	if z >= 1 && int(z) <= len(zhiPalaceTable) {
		return zhiPalaceTable[int(z)-1]
	}
	return 0
}

func palaceWuxing(p GongIndex) ganzhi.Wuxing {
	if p >= 1 && int(p) <= len(gongWuxingTable) {
		return gongWuxingTable[int(p)-1]
	}
	return 0
}

func palaceIdentity(p GongIndex) PalaceIdentity {
	if p < 1 || p > 9 {
		return PalaceIdentity{Name: "?", Luoshu: int(p)}
	}
	identity := PalaceIdentity{Name: p.String(), Luoshu: int(p)}
	if ringIndex := outerRingIndex(p); ringIndex >= 0 {
		value := ringIndex
		identity.RingIndex = &value
	}
	return identity
}

func starHomePalace(s StarIndex) int {
	for i, star := range palaceStar {
		if star == s {
			return i
		}
	}
	return int(GongZhong) - 1
}
