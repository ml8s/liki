package qimen

import "liki-engine/internal/engine/ganzhi"

// StarInteraction holds star-palace interaction data.
type StarInteraction struct {
	Star       string `json:"xing"`
	Palace     string `json:"gong"`
	Name       string `json:"name"`
	Meaning    string `json:"meaning"`
	Auspicious bool   `json:"auspicious"`
}

// computeStarInteractions returns star-palace 克应 for each palace.
func computeStarInteractions(pan pan) [9]StarInteraction {
	var result [9]StarInteraction
	for i := 0; i < 9; i++ {
		p := pan.Palaces[i]
		if p.Star == 0 {
			continue
		}
		key := [2]int{int(p.Star), i}
		if entry, ok := starPalaceTable[key]; ok {
			result[i] = entry
		} else {
			// Generic five-element-based description.
			result[i] = genericStarInteraction(p.Star, PalaceIndex(i+1))
		}
	}
	return result
}

func genericStarInteraction(star StarIndex, pal PalaceIndex) StarInteraction {
	return StarInteraction{
		Star:       star.String(),
		Palace:     pal.String(),
		Name:       star.String() + "加" + pal.String(),
		Meaning:    starNature(star) + "临" + pal.String() + "宫",
		Auspicious: isAuspiciousStar(star),
	}
}

func starNature(s StarIndex) string {
	switch s {
	case StarTianPeng:
		return "水性之精"
	case StarTianRui:
		return "土性之精"
	case StarTianChong:
		return "木性之精"
	case StarTianFu:
		return "木性文明"
	case StarTianQin:
		return "土性中和"
	case StarTianXin:
		return "金性肃杀"
	case StarTianZhu:
		return "金性锐利"
	case StarTianRen:
		return "土性厚重"
	case StarTianYing:
		return "火性光明"
	}
	return ""
}

func isAuspiciousStar(s StarIndex) bool {
	switch s {
	case StarTianFu, StarTianQin, StarTianXin, StarTianRen:
		return true
	default:
		return false
	}
}

// WangShuai represents 旺衰 state of a star in a palace.
type WangShuai struct {
	Star   StarIndex   `json:"xing"`
	Palace PalaceIndex `json:"gong"`
	State  string      `json:"state"` // 旺/相/休/囚/废
}

// computeWangShuai computes the 旺衰 state for each star in the pan.
func computeWangShuai(pan pan) [9]WangShuai {
	var result [9]WangShuai
	for i, p := range pan.Palaces {
		if p.Star == 0 {
			continue
		}
		sw := starWuxing(p.Star)
		pw := palaceWuxing(PalaceIndex(i + 1))
		result[i] = WangShuai{
			Star:   p.Star,
			Palace: PalaceIndex(i + 1),
			State:  wuxingState(sw, pw),
		}
	}
	return result
}

// starWuxing returns the element of a star.
func starWuxing(s StarIndex) ganzhi.Wuxing {
	switch s {
	case StarTianPeng:
		return ganzhi.WxShui
	case StarTianRui, StarTianQin, StarTianRen:
		return ganzhi.WxTu
	case StarTianChong, StarTianFu:
		return ganzhi.WxMu
	case StarTianXin, StarTianZhu:
		return ganzhi.WxJin
	case StarTianYing:
		return ganzhi.WxHuo
	}
	return 0
}

// wuxingState returns 旺/相/休/囚/废 for star(at starElem) in palace(at palElem).
func wuxingState(starElem, palElem ganzhi.Wuxing) string {
	if starElem == palElem {
		return "旺"
	}
	if ganzhi.Sheng(palElem, starElem) { // palace generates star → 相
		return "相"
	}
	if ganzhi.Sheng(starElem, palElem) { // star generates palace → 休
		return "休"
	}
	if ganzhi.Ke(starElem, palElem) { // star overcomes palace → 囚
		return "囚"
	}
	return "废"
}
