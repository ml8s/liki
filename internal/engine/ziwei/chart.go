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
	shiZhi := bz.Shi.Zhi
	nianGan := bz.Nian.Gan
	nianZhi := bz.Nian.Zhi

	mingZhi, shenZhi := computeMingShen(lunarMonth, shiZhi)
	palaceZhis := arrangePalaceZhis(mingZhi)
	shenGong := findShenGongIndex(palaceZhis, shenZhi)

	mingGan, palaceGans := arrangePalaceGans(nianGan, mingZhi, (lunarMonth-int(shiZhi)+12)%12)
	ju := determineJuShu(mingGan, mingZhi)
	ziweiAnXingIdx := findZiwei(ju, lunarDay)
	tianfuAnXingIdx := (12 - ziweiAnXingIdx) % 12
	ziweiPos := anXingIdxToPalace(ziweiAnXingIdx, mingZhi)
	mainByPalace := placeMainStars(ziweiAnXingIdx, tianfuAnXingIdx, mingZhi)
	minorByPalace := placeMinorStars(ganzhi.Zhu{Gan: nianGan, Zhi: nianZhi}, lunarMonth, shiZhi, mingZhi)

	var palaces [12]gong
	for i := 0; i < 12; i++ {
		starInfos := make([]starInfo, 0)
		// mainByPalace/minorByPalace 为 zhiIdx 坐标，经本宫支反查
		zm1 := zhiToZhiIdx(palaceZhis[i])
		for _, s := range mainByPalace[zm1] {
			starInfos = append(starInfos, starInfo{Star: s, Name: starName(s), IsMajor: true})
		}
		for _, s := range minorByPalace[zm1] {
			starInfos = append(starInfos, starInfo{Star: s, Name: starName(s), IsMajor: false})
		}
		palaces[i] = gong{
			Index:        gongIndex(i),
			Name:         gongLabels[i],
			Gan:          palaceGans[i],
			Zhi:          palaceZhis[i],
			IsBodyPalace: gongIndex(i) == shenGong,
			Stars:        starInfos,
		}
	}

	return Chart{
		GongWei:        palaces,
		MingGong:       0,
		ShenGong:       shenGong,
		JuShu:          ju,
		JuShuName:      juShuName(ju),
		ZiweiPos:       ziweiPos,
		NianGan:        nianGan,
		NianZhi:        nianZhi,
		ShiZhi:         shiZhi,
		LunarMonth:     lunarMonth,
		LunarDay:       lunarDay,
		BirthLunarMonth: lt.Month,
		BirthIsLeap:    lt.Leap,
	}
}

// buildChartDetail enriches a core chart with siHua, brightness, and patterns.
func buildChartDetail(chart Chart) Chart {
	siHua := computeSiHua(chart.NianGan)
	for i := range chart.GongWei {
		for j, s := range chart.GongWei[i].Stars {
			if s.IsMajor {
				chart.GongWei[i].Stars[j].Brightness = miaoWang(s.Star, chart.GongWei[i].Zhi).String()
				if h, ok := siHua[s.Star]; ok {
					chart.GongWei[i].Stars[j].SiHua = string(h)
				}
			}
		}
	}
	chart.SiHua = siHua
	chart.Patterns = findPatterns(chart.GongWei)
	// 来因宫
	ygIdx := yuanGongPalace(chart.GongWei, chart.NianGan)
	chart.GongWei[ygIdx].IsYuanGong = true
	return chart
}


