package qimen

import "liki-engine/internal/engine/ganzhi"

// doorEntry holds named door-gong data (door/gong set at runtime).
type doorEntry struct {
	DoorName    string
	GongName  string
	Name        string
	Meaning     string
}

// computeMenInteractions returns door interactions for each gong.
func computeMenInteractions(pan pan) [9]MenInteraction {
	var result [9]MenInteraction
	for i := 0; i < 9; i++ {
		p := pan.GongWei[i]
		if p.Door == 0 {
			continue
		}
		key := [2]int{int(p.Door), i}
		if entry, ok := menGongTable[key]; ok {
			result[i] = MenInteraction{
				Door:    p.Door,
				Gong:  GongIndex(i + 1),
				Name:    entry.Name,
				Meaning: entry.Meaning,
			}
		} else {
			// Generic: door name + gong name
			result[i] = MenInteraction{
				Door:    p.Door,
				Gong:  GongIndex(i + 1),
				Name:    p.Door.String() + "加" + GongIndex(i+1).String(),
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

// menPo checks if a door is 门迫 (door overcomes gong) at the given gong.
func menPo(door DoorIndex, gong GongIndex) bool {
	de := doorWuxing(door)
	pe := palaceWuxing(gong)
	return de != 0 && pe != 0 && ganzhi.Ke(de, pe)
}

// menZhi checks if a door is 门制 (gong overcomes door) at the given gong.
func menZhi(door DoorIndex, gong GongIndex) bool {
	de := doorWuxing(door)
	pe := palaceWuxing(gong)
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
func findMenPo(pan pan) []GongIndex {
	var result []GongIndex
	for i, p := range pan.GongWei {
		if p.Door != 0 && menPo(p.Door, GongIndex(i+1)) {
			result = append(result, GongIndex(i+1))
		}
	}
	return result
}

// findMenZhi returns palaces where the door is 门制.
func findMenZhi(pan pan) []GongIndex {
	var result []GongIndex
	for i, p := range pan.GongWei {
		if p.Door != 0 && menZhi(p.Door, GongIndex(i+1)) {
			result = append(result, GongIndex(i+1))
		}
	}
	return result
}
