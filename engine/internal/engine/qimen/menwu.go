package qimen

import "liki-engine/internal/engine/ganzhi"

// doorEntry holds named door-gong data (door/gong set at runtime).
type doorEntry struct {
	Name    string
	Meaning string
}

// computeMenInteractions returns door interactions for each gong.
func computeMenInteractions(pan pan) []MenInteraction {
	result := []MenInteraction{}
	for i := 0; i < 9; i++ {
		p := pan.GongWei[i]
		if p.Door == 0 {
			continue
		}
		key := [2]int{int(p.Door), i}
		if entry, ok := menGongTable[key]; ok {
			result = append(result, MenInteraction{
				Door:    p.Door,
				Gong:    GongIndex(i + 1),
				Name:    entry.Name,
				Meaning: entry.Meaning,
			})
		}
	}
	return result
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
	if d >= 1 && int(d) <= len(doorWuxingTable) {
		return doorWuxingTable[int(d)-1]
	}
	return 0
}

// findMenPo returns palaces where the door is 门迫.
func findMenPo(pan pan) []GongIndex {
	result := []GongIndex{}
	for i, p := range pan.GongWei {
		if p.Door != 0 && menPo(p.Door, GongIndex(i+1)) {
			result = append(result, GongIndex(i+1))
		}
	}
	return result
}

// findMenZhi returns palaces where the door is 门制.
func findMenZhi(pan pan) []GongIndex {
	result := []GongIndex{}
	for i, p := range pan.GongWei {
		if p.Door != 0 && menZhi(p.Door, GongIndex(i+1)) {
			result = append(result, GongIndex(i+1))
		}
	}
	return result
}
