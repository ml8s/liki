package ziwei


import (
	"liki-engine/internal/engine/ganzhi"
	"liki-engine/internal/engine/tianwen"
)
// computeChart builds the core ziwei chart (palaces + stars, no brightness/patterns).
func computeChart(bz ganzhi.Bazi, lt tianwen.LunarTime) Chart {
	lunarMonth, lunarDay := lt.Month, lt.Day
	// 闰月处理（与 iztro fixLeap 一致）：闰月后半月（日>15）按下月算，前半月按本月算
	if lt.Leap && lunarDay > 15 {
		lunarMonth++
	}
	hourZhi := bz.Shi.Zhi
	yearGan := bz.Nian.Gan
	yearZhi := bz.Nian.Zhi

	mingZhi, shenZhi := computeMingShen(lunarMonth, hourZhi)
	palaceZhis := arrangePalaceZhis(mingZhi)
	shenGong := findShenGongIndex(palaceZhis, shenZhi)

	mingGan, palaceGans := arrangePalaceGans(yearGan, mingZhi, (lunarMonth-int(hourZhi)+12)%12)
	ju := determineJuShu(mingGan, mingZhi)
	iztroZW := findZiwei(ju, lunarDay)
	iztroTF := (12 - iztroZW) % 12
	ziweiPos := iztroIdxToPalace(iztroZW, mingZhi)
	mainByPalace := placeMainStars(iztroZW, iztroTF, mingZhi)
	minorByPalace := placeMinorStars(ganzhi.Zhu{Gan: yearGan, Zhi: yearZhi}, lunarMonth, hourZhi, mingZhi)

	var palaces [12]palace
	for i := 0; i < 12; i++ {
		var starInfos []starInfo
		// mainByPalace/minorByPalace 为 zhiMinus1 坐标，经本宫支反查
		zm1 := zhiToZhiMinus1(palaceZhis[i])
		for _, s := range mainByPalace[zm1] {
			starInfos = append(starInfos, starInfo{Star: s, Name: starName(s), IsMajor: true})
		}
		for _, s := range minorByPalace[zm1] {
			starInfos = append(starInfos, starInfo{Star: s, Name: starName(s), IsMajor: false})
		}
		palaces[i] = palace{
			Index:        palaceIndex(i),
			Name:         PalaceNames[i],
			Gan:          palaceGans[i],
			Zhi:          palaceZhis[i],
			IsBodyPalace: palaceIndex(i) == shenGong,
			Stars:        starInfos,
		}
	}

	return Chart{
		Palaces:        palaces,
		MingGong:       0,
		ShenGong:       shenGong,
		JuShu:          ju,
		JuShuName:      juShuName(ju),
		ZiweiPos:       ziweiPos,
		NianGan:        yearGan,
		NianZhi:        yearZhi,
		ShiZhi:         hourZhi,
		LunarMonth:     lunarMonth,
		LunarDay:       lunarDay,
		BirthLunarMonth: lt.Month,
		BirthIsLeap:    lt.Leap,
	}
}

// buildChartDetail enriches a core chart with siHua, brightness, and patterns.
func buildChartDetail(chart Chart) Chart {
	siHua := computeSiHua(chart.NianGan)
	for i := range chart.Palaces {
		for j, s := range chart.Palaces[i].Stars {
			if s.IsMajor {
				chart.Palaces[i].Stars[j].Brightness = miaoWang(s.Star, chart.Palaces[i].Zhi).String()
				if h, ok := siHua[s.Star]; ok {
					chart.Palaces[i].Stars[j].SiHua = string(h)
				}
			}
		}
	}
	chart.SiHua = siHua
	chart.Patterns = findPatterns(chart.Palaces)
	// 来因宫
	ygIdx := yuanGongPalace(chart.Palaces, chart.NianGan)
	chart.Palaces[ygIdx].IsYuanGong = true
	return chart
}


