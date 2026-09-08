package qimen

import "liki-engine/internal/engine/ganzhi"

// ganEntry holds the named interaction data (without gan fields set at runtime).
type ganEntry struct {
	PatternName string
	Meaning     string
	Auspicious  bool
}

// computeGanInteractions returns the 十干克应 for each gong.
func computeGanInteractions(pan pan) []GanInteraction {
	result := []GanInteraction{}
	for i := 0; i < 9; i++ {
		p := pan.GongWei[i]
		for _, heavenGan := range heavenGans(p) {
			key := [2]ganzhi.Gan{p.DiPanGan, heavenGan}
			if entry, ok := ganInteractionTable[key]; ok {
				result = append(result, GanInteraction{
					Gong:       GongIndex(i + 1),
					DiPanGan:   p.DiPanGan,
					TianPanGan: heavenGan,
					Name:       heavenGan.String() + "+" + p.DiPanGan.String(),
					Meaning:    entry.PatternName + "：" + entry.Meaning,
					Auspicious: entry.Auspicious,
				})
			}
		}
	}
	return result
}
