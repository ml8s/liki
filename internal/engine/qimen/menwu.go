package qimen

import "liki-engine/internal/engine/ganzhi"

// doorEntry holds named door-palace data (door/palace set at runtime).
type doorEntry struct {
	DoorName    string
	PalaceName  string
	Name        string
	Meaning     string
}

// computeDoorInteractions returns door interactions for each palace.
func computeDoorInteractions(pan pan) [9]DoorInteraction {
	var result [9]DoorInteraction
	for i := 0; i < 9; i++ {
		p := pan.Palaces[i]
		if p.Door == 0 {
			continue
		}
		key := [2]int{int(p.Door), i}
		if entry, ok := doorPalaceTable[key]; ok {
			result[i] = DoorInteraction{
				Door:    p.Door,
				Palace:  PalaceIndex(i + 1),
				Name:    entry.Name,
				Meaning: entry.Meaning,
			}
		} else {
			// Generic: door name + palace name
			result[i] = DoorInteraction{
				Door:    p.Door,
				Palace:  PalaceIndex(i + 1),
				Name:    p.Door.String() + "加" + PalaceIndex(i+1).String(),
				Meaning: doorAuspicious(p.Door),
			}
		}
	}
	return result
}

// doorAuspicious returns a generic description based on whether the door is auspicious.
func doorAuspicious(d DoorIndex) string {
	switch d {
	case DoorXiu, DoorSheng, DoorKai:
		return "吉门得地，谋事可成"
	case DoorDu, DoorJing:
		return "中平之门，需择时而行"
	case DoorShang, DoorSi, DoorJingMen:
		return "凶门当位，行事多阻"
	}
	return ""
}

// menPo checks if a door is 门迫 (door overcomes palace) at the given palace.
func menPo(door DoorIndex, palace PalaceIndex) bool {
	de := doorWuxing(door)
	pe := palaceWuxing(palace)
	return de != 0 && pe != 0 && ganzhi.Ke(de, pe)
}

// menZhi checks if a door is 门制 (palace overcomes door) at the given palace.
func menZhi(door DoorIndex, palace PalaceIndex) bool {
	de := doorWuxing(door)
	pe := palaceWuxing(palace)
	return de != 0 && pe != 0 && ganzhi.Ke(pe, de)
}

// doorWuxing returns the element of a door.
func doorWuxing(d DoorIndex) ganzhi.Wuxing {
	switch d {
	case DoorXiu:
		return ganzhi.WxShui
	case DoorSheng, DoorSi:
		return ganzhi.WxTu
	case DoorShang, DoorDu:
		return ganzhi.WxMu
	case DoorJing:
		return ganzhi.WxHuo
	case DoorJingMen, DoorKai:
		return ganzhi.WxJin
	}
	return 0
}

// findMenPo returns palaces where the door is 门迫.
func findMenPo(pan pan) []PalaceIndex {
	var result []PalaceIndex
	for i, p := range pan.Palaces {
		if p.Door != 0 && menPo(p.Door, PalaceIndex(i+1)) {
			result = append(result, PalaceIndex(i+1))
		}
	}
	return result
}

// findMenZhi returns palaces where the door is 门制.
func findMenZhi(pan pan) []PalaceIndex {
	var result []PalaceIndex
	for i, p := range pan.Palaces {
		if p.Door != 0 && menZhi(p.Door, PalaceIndex(i+1)) {
			result = append(result, PalaceIndex(i+1))
		}
	}
	return result
}
