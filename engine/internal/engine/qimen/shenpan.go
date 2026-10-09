package qimen

// placeShenPan arranges the spirit plate around the outer ring.
// The leading spirit follows the duty star; Yang Dun proceeds clockwise and
// Yin Dun counter-clockwise.
func placeShenPan(yinDun bool, tianStarPos GongIndex) [9]SpiritIndex {
	var spirits [9]SpiritIndex
	start := outerRingIndex(tianStarPos)
	if start < 0 {
		return spirits
	}
	for step, si := 0, 0; si < 8; step++ {
		ringIdx := (start + step) % 8
		if yinDun {
			ringIdx = (start - step + 8) % 8
		}
		spirits[int(outerRing[ringIdx])-1] = spiritOrder[si]
		si++
	}
	return spirits
}

func spiritDisplayName(spirit SpiritIndex, yinDun bool, school School) string {
	if school == SchoolMingFaFeiPan && int(spirit) < len(mingfaSpiritNames) &&
		mingfaSpiritNames[int(spirit)] != "" {
		return mingfaSpiritNames[int(spirit)]
	}
	if school == SchoolLuoShuFeiPan && int(spirit) < len(flySpiritNames) && flySpiritNames[spirit] != "" {
		return flySpiritNames[spirit]
	}
	if yinDun {
		return spirit.YinName()
	}
	return spirit.YangName()
}
