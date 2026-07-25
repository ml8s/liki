package qimen

import "liki-engine/internal/engine/ganzhi"

// stemEntry holds the named interaction data (without stem fields set at runtime).
type stemEntry struct {
	Name        string
	PatternName string
	Meaning     string
	Auspicious  bool
}

// computeStemInteractions returns the 十干克应 for each palace.
func computeStemInteractions(pan pan) [9]StemInteraction {
	var result [9]StemInteraction
	for i := 0; i < 9; i++ {
		p := pan.Palaces[i]
		key := [2]ganzhi.Gan{p.EarthStem, p.HeavenStem}
		if entry, ok := stemInteractionTable[key]; ok {
			result[i] = StemInteraction{
				EarthStem:  p.EarthStem,
				HeavenStem: p.HeavenStem,
				Name:       entry.Name,
				Meaning:    entry.PatternName + "：" + entry.Meaning,
				Auspicious: entry.Auspicious,
			}
		} else {
			result[i] = genericStemInteraction(p.EarthStem, p.HeavenStem)
		}
	}
	return result
}

// genericStemInteraction generates a five-element-based description for unnamed combinations.
func genericStemInteraction(earth, heaven ganzhi.Gan) StemInteraction {
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

	return StemInteraction{
		EarthStem:  earth,
		HeavenStem: heaven,
		Name:       name,
		Meaning:    meaning,
		Auspicious: auspicious,
	}
}
