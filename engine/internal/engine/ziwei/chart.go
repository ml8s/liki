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
		starInfos := make([]starInfo, 0)
		for _, star := range mainByPalace[zhiIdx] {
			starInfos = append(starInfos, starInfo{Star: star, Name: starName(star), IsMajor: true})
		}
		for _, star := range minorByPalace[zhiIdx] {
			starInfos = append(starInfos, starInfo{Star: star, Name: starName(star), IsMajor: false})
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

// SanFangInfo describes one palace in the 三方四正 (命宫/财帛/官禄/迁移 三合).
type SanFangInfo struct {
	Name    string   `json:"name"`
	ZhuXing []string `json:"zhu_xing"`
	FuXing  []string `json:"fu_xing"`
	SiHua   []string `json:"si_hua"`
}

// buildSanFangInfo collects the major/minor stars and 四化 of the 三方四正 palaces.
func buildSanFangInfo(c Chart, sfPalaces [4]gongIndex) []SanFangInfo {
	var result []SanFangInfo
	for _, pi := range sfPalaces {
		p := c.GongWei[pi]
		info := SanFangInfo{
			Name: gongLabels[pi], ZhuXing: make([]string, 0),
			FuXing: make([]string, 0), SiHua: make([]string, 0),
		}
		for _, s := range p.Stars {
			if s.SiHua != "" {
				info.SiHua = append(info.SiHua, s.SiHua)
			}
			if s.IsMajor {
				info.ZhuXing = append(info.ZhuXing, s.Name)
			} else {
				info.FuXing = append(info.FuXing, s.Name)
			}
		}
		result = append(result, info)
	}
	return result
}

// buildChartDetail enriches a core chart with siHua, brightness, and patterns.
func buildChartDetail(chart Chart) Chart {
	chart = applyBrightnessAndSiHua(chart)
	chart.SiHua = computeSiHua(chart.NianGan)
	chart.KongGong = computeKongGong(chart.GongWei)
	chart.Patterns = findPatterns(chart.GongWei)
	chart.SanFang = buildSanFangInfo(chart, sanFang(0))
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
		info := kongGongInfo{
			GongName: gongLabels[gongIndex(i)],
			DuiGong:  gongLabels[gongIndex(dui)],
		}
		for _, s := range gongWei[dui].Stars {
			if s.IsMajor {
				info.JieXing = append(info.JieXing, s.Name)
			}
		}
		out = append(out, info)
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
