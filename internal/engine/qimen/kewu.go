package qimen

import "liki-engine/internal/engine/ganzhi"

// stemEntry holds the named interaction data (without stem fields set at runtime).
type stemEntry struct {
	Name        string
	PatternName string
	Meaning     string
	Auspicious  bool
}

// computeGanInteractions returns the 十干克应 for each gong.
func computeGanInteractions(pan pan) [9]GanInteraction {
	var result [9]GanInteraction
	for i := 0; i < 9; i++ {
		p := pan.GongWei[i]
		key := [2]ganzhi.Gan{p.EarthStem, p.HeavenStem}
		if entry, ok := ganInteractionTable[key]; ok {
			result[i] = GanInteraction{
				EarthStem:  p.EarthStem,
				HeavenStem: p.HeavenStem,
				Name:       entry.Name,
				Meaning:    entry.PatternName + "：" + entry.Meaning,
				Auspicious: entry.Auspicious,
			}
		} else {
			result[i] = genericGanInteraction(p.EarthStem, p.HeavenStem)
		}
	}
	return result
}

// genericGanInteraction generates a five-element-based description for unnamed combinations.
func genericGanInteraction(earth, heaven ganzhi.Gan) GanInteraction {
	eWuxing := ganzhi.GanWuxing(earth)
	hWuxing := ganzhi.GanWuxing(heaven)
	name := ganzhi.GanName(earth) + "+" + ganzhi.GanName(heaven)

	var meaning string
	var auspicious bool
	if eWuxing == hWuxing {
		meaning = "比和，静守为宜"
		auspicious = false
	} else if ganzhi.Sheng(hWuxing, eWuxing) { // heaven generates earth → 上生下
		meaning = "上生下，谋事可成"
		auspicious = true
	} else if ganzhi.Sheng(eWuxing, hWuxing) { // earth generates heaven → 下生上
		meaning = "下生上，耗损有忧"
		auspicious = false
	} else if ganzhi.Ke(hWuxing, eWuxing) { // heaven overcomes earth → 上克下
		meaning = "上克下，主胜于客"
		auspicious = false
	} else { // earth overcomes heaven → 下克上
		meaning = "下克上，客胜于主"
		auspicious = true
	}

	return GanInteraction{
		EarthStem:  earth,
		HeavenStem: heaven,
		Name:       name,
		Meaning:    meaning,
		Auspicious: auspicious,
	}
}
