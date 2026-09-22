package qimen

import "liki-engine/internal/engine/ganzhi"

// findPatterns evaluates the table-driven pattern rules. All symbol conditions
// are evaluated within one palace; pan-wide scattered symbols never combine.
func findPatterns(chart pan) []Pattern {
	patterns := []Pattern{}
	for _, rule := range patternRules {
		switch rule.DutyStarPosition {
		case "home", "opposite":
			if palace := dutyStarSpecialPosition(chart, rule.DutyStarPosition); palace != 0 {
				patterns = append(patterns, newPattern(rule, []GongIndex{palace}))
			}
		default:
			var palaces []GongIndex
			for i, palace := range chart.GongWei {
				if GongIndex(i+1) == GongZhong {
					continue
				}
				if rule.RequiresDutyDoor && dutyDoorPalace(chart) != GongIndex(i+1) {
					continue
				}
				if matchesPatternCondition(rule.Conditions, palace, GongIndex(i+1)) {
					palaces = append(palaces, GongIndex(i+1))
				}
			}
			if len(palaces) > 0 {
				patterns = append(patterns, newPattern(rule, palaces))
			}
		}
	}
	return patterns
}

func newPattern(rule patternRule, palaces []GongIndex) Pattern {
	return Pattern{
		Name: rule.Name, Description: rule.Description, Basis: rule.Basis,
		Auspicious: rule.Auspicious, GongWei: palaces,
	}
}

func matchesPatternCondition(conditions []patternCondition, palace Gong, position GongIndex) bool {
	for _, condition := range conditions {
		if !containsGan(condition.HeavenGan, heavenGans(palace)) ||
			!containsGan(condition.EarthGan, []ganzhi.Gan{palace.DiPanGan}) ||
			!containsDoor(condition.Doors, palace.Door) ||
			!containsSpirit(condition.Spirits, palace.Spirit) ||
			!containsPalace(condition.Palaces, position) {
			continue
		}
		return true
	}
	return false
}

func heavenGans(palace Gong) []ganzhi.Gan {
	result := make([]ganzhi.Gan, 0, len(palace.TianPan)+1)
	for _, item := range palace.TianPan {
		result = append(result, item.Gan)
	}
	if palace.TianPanGan != nil {
		result = append(result, *palace.TianPanGan)
	}
	return result
}

func containsGan(expected []ganzhi.Gan, actual []ganzhi.Gan) bool {
	if len(expected) == 0 {
		return true
	}
	for _, want := range expected {
		for _, got := range actual {
			if want == got {
				return true
			}
		}
	}
	return false
}

func containsDoor(expected []DoorIndex, actual DoorIndex) bool {
	if len(expected) == 0 {
		return true
	}
	for _, want := range expected {
		if want == actual {
			return true
		}
	}
	return false
}

func containsSpirit(expected []SpiritIndex, actual SpiritIndex) bool {
	if len(expected) == 0 {
		return true
	}
	for _, want := range expected {
		if want == actual {
			return true
		}
	}
	return false
}

func containsPalace(expected []GongIndex, actual GongIndex) bool {
	if len(expected) == 0 {
		return true
	}
	for _, want := range expected {
		if want == actual {
			return true
		}
	}
	return false
}

func dutyDoorPalace(chart pan) GongIndex {
	for i, palace := range chart.GongWei {
		if palace.Door == chart.DutyDoor {
			return GongIndex(i + 1)
		}
	}
	return 0
}

func dutyStarSpecialPosition(chart pan, position string) GongIndex {
	duty := chart.DutyStar
	if chart.School == SchoolZhuanPan && duty == tianQinStar {
		duty = tianQinFollows
	}
	current := findStarPalace(chart, duty)
	home := GongIndex(starHomePalace(duty) + 1)
	if position == "home" {
		if current == home {
			return current
		}
		return 0
	}
	opposite := palaceOpposite(home)
	if current == opposite {
		return current
	}
	return 0
}

func palaceOpposite(p GongIndex) GongIndex {
	index := outerRingIndex(p)
	if index < 0 {
		return 0
	}
	return outerRing[(index+4)%8]
}
