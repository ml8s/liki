package qimen

import "liki-engine/internal/engine/ganzhi"

// placeRenPan first flies the duty door from the ten-day-cycle head palace by
// the lead pillar's offset, then aligns the fixed door ring at that palace. The
// flight starts from the real source palace (including Zhong5); only the
// resulting ring position is projected to Kun2.
func placeRenPan(
	leadZhu ganzhi.Zhu, dutyDoor DoorIndex, yinDun bool, dipan [9]ganzhi.Gan,
) ([9]DoorIndex, GongIndex) {
	var doors [9]DoorIndex
	xunShou := findXunShou(leadZhu)
	source := 0
	for i, g := range dipan {
		if g == xunShou {
			source = i
			break
		}
	}
	offset := ganzhi.SixtyCycleIndex(leadZhu.Gan, leadZhu.Zhi) % 10
	targetPos := source + offset
	if yinDun {
		targetPos = source - offset
	}
	targetPos = ((targetPos % 9) + 9) % 9
	realTarget := GongIndex(targetPos + 1)
	if targetPos == int(GongZhong)-1 {
		targetPos = int(centerLodgingPalace) - 1
	}

	target := GongIndex(targetPos + 1)
	targetRing := outerRingIndex(target)
	dutyRing := ringIndexForDoor(dutyDoor)
	for step := 0; step < 8; step++ {
		palace := outerRing[(targetRing+step)%8]
		doors[int(palace)-1] = doorOrder[(dutyRing+step)%8]
	}
	return doors, realTarget
}

func placeHiddenPan(dipan [9]ganzhi.Gan, source, landing GongIndex) [9]ganzhi.Gan {
	var result [9]ganzhi.Gan
	sourceRing := outerRingIndex(projectCenter(source))
	landingRing := outerRingIndex(projectCenter(landing))
	if sourceRing < 0 || landingRing < 0 {
		return result
	}
	shift := floorMod(landingRing-sourceRing, 8)
	for i, palace := range outerRing {
		target := outerRing[(i+shift)%8]
		result[int(target)-1] = dipan[int(palace)-1]
	}
	result[int(GongZhong)-1] = dipan[int(GongZhong)-1]
	return result
}
