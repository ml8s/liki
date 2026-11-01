package ziwei

import (
	"liki-engine/internal/engine/ganzhi"
	"liki-engine/internal/engine/tianwen"
)

// computeChart builds the core ziwei chart (palaces + stars, no brightness/patterns).
func computeChart(bz ganzhi.Bazi, lt tianwen.LunarTime) Chart {
	lunarMonth := publishedLunarMonth(lt)
	lunarDay := lt.Day
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
	palaces := assemblePalaces(
		palaceZhis,
		palaceGans,
		shenGong,
		mainByPalace,
		minorByPalace,
	)
	return assembleCoreChart(
		palaces,
		coreCalendarFacts{
			ShenGong:        shenGong,
			JuShu:           ju,
			JuShuName:       juShuName(ju),
			ZiweiPos:        ziweiPos,
			NianGan:         nianGan,
			NianZhi:         nianZhi,
			ShiZhi:          shiZhi,
			LunarMonth:      lunarMonth,
			LunarDay:        lunarDay,
			BirthLunarMonth: lt.Month,
			BirthIsLeap:     lt.Leap,
		},
	)
}

// leap month belongs to the marked month; the second half belongs to the next.
func publishedLunarMonth(lt tianwen.LunarTime) int {
	if lt.Leap && lt.Day > 15 {
		return lt.Month + 1
	}
	return lt.Month
}

type coreCalendarFacts struct {
	JuShu           juShu
	JuShuName       string
	ZiweiPos        gongIndex
	ShenGong        gongIndex
	NianGan         Gan
	NianZhi         Zhi
	ShiZhi          Zhi
	LunarMonth      int
	LunarDay        int
	BirthLunarMonth int
	BirthIsLeap     bool
}

func assemblePalaces(
	palaceZhis [12]Zhi,
	palaceGans [12]Gan,
	shenGong gongIndex,
	mainByPalace map[int][]starIndex,
	minorByPalace map[int][]starIndex,
) [12]gong {
	var palaces [12]gong
	for i := 0; i < 12; i++ {
		zhiIdx := zhiToZhiIdx(palaceZhis[i])
		major := mainByPalace[zhiIdx]
		minor := minorByPalace[zhiIdx]
		starInfos := make([]starInfo, 0, len(major)+len(minor))
		for _, star := range major {
			starInfos = append(starInfos, newStarInfo(star, true))
		}
		for _, star := range minor {
			starInfos = append(starInfos, newStarInfo(star, false))
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
	return palaces
}

func assembleCoreChart(palaces [12]gong, facts coreCalendarFacts) Chart {
	return Chart{
		GongWei:         palaces,
		MingGong:        0,
		ShenGong:        facts.ShenGong,
		JuShu:           facts.JuShu,
		JuShuName:       facts.JuShuName,
		ZiweiPos:        facts.ZiweiPos,
		NianGan:         facts.NianGan,
		NianZhi:         facts.NianZhi,
		ShiZhi:          facts.ShiZhi,
		LunarMonth:      facts.LunarMonth,
		LunarDay:        facts.LunarDay,
		BirthLunarMonth: facts.BirthLunarMonth,
		BirthIsLeap:     facts.BirthIsLeap,
	}
}

// sanFangSiZheng returns the palace relation only. Star facts remain owned by
// each palace in ShiErGong.
func sanFangSiZheng(palaces [4]gongIndex) SanFangSiZheng {
	return SanFangSiZheng{MingGong: palaces[:]}
}

// buildChartDetail enriches a core chart with siHua, brightness, and patterns.
func buildChartDetail(chart Chart) Chart {
	chart = applyBrightnessAndSiHua(chart)
	chart.SiHua = computeSiHua(chart.NianGan)
	chart.KongGong = computeKongGong(chart.GongWei)
	chart.Patterns = findPatterns(chart.GongWei)
	chart.SanFangSiZheng = sanFangSiZheng(sanFang(0))
	return markYuanGong(chart)
}

func applyBrightnessAndSiHua(chart Chart) Chart {
	siHua := computeSiHua(chart.NianGan)
	for i := range chart.GongWei {
		for j, star := range chart.GongWei[i].Stars {
			if star.IsMajor || star.Star == WenChang || star.Star == WenQu {
				chart.GongWei[i].Stars[j].Brightness = miaoWang(star.Star, chart.GongWei[i].Zhi).String()
			}
			if transformation, ok := siHua[star.Star]; ok {
				chart.GongWei[i].Stars[j].SiHua = string(transformation)
			}
		}
	}
	return chart
}

func markYuanGong(chart Chart) Chart {
	yuanGongIdx := yuanGongPalace(chart.GongWei, chart.NianGan)
	chart.GongWei[yuanGongIdx].IsYuanGong = true
	return chart
}

// computeKongGong 找出所有无主星宫位，借对宫主星（紫微空宫借星——确定性派生）。
func computeKongGong(gongWei [12]gong) []kongGongInfo {
	var out []kongGongInfo
	for i := range gongWei {
		if hasMajorStar(gongWei[i]) {
			continue
		}
		dui := (i + 6) % 12
		out = append(out, kongGongInfo{KongGong: gongIndex(i), JieGong: gongIndex(dui)})
	}
	return out
}

func hasMajorStar(g gong) bool {
	for _, s := range g.Stars {
		if s.IsMajor {
			return true
		}
	}
	return false
}
