package bazhai

import (
	"liki-engine/internal/engine/ganzhi"
)

// gua is a bagua trigram.
type gua struct {
	Index   int    `json:"index"`
	Name    string `json:"name"`
	Wuxing  string `json:"wuxing"`
	YinYang string `json:"yin_yang"`
}

// guaTable maps 洛书数 (1-9) to trigrams.
// 洛书: 1=坎 2=坤 3=震 4=巽 5=中宫(男寄坤女寄艮) 6=乾 7=兑 8=艮 9=离
var guaTable = [10]gua{
	{},
	{1, "坎", "水", "阳"}, // 洛书 1 = 坎
	{2, "坤", "土", "阴"}, // 洛书 2 = 坤
	{3, "震", "木", "阳"}, // 洛书 3 = 震
	{4, "巽", "木", "阴"}, // 洛书 4 = 巽
	{5, "坤", "土", "阴"}, // 洛书 5 = 中宫(男寄坤)
	{6, "乾", "金", "阳"}, // 洛书 6 = 乾
	{7, "兑", "金", "阴"}, // 洛书 7 = 兑
	{8, "艮", "土", "阳"}, // 洛书 8 = 艮
	{9, "离", "火", "阴"}, // 洛书 9 = 离
}

func ganNaJia(stem ganzhi.Gan) gua {
	switch stem {
	case 1, 9: return guaTable[6]  // 甲壬→乾
	case 2, 10: return guaTable[2] // 乙癸→坤
	case 3: return guaTable[8]     // 丙→艮
	case 4: return guaTable[7]     // 丁→兑
	case 5: return guaTable[1]     // 戊→坎
	case 6: return guaTable[9]     // 己→离
	case 7: return guaTable[3]     // 庚→震
	case 8: return guaTable[4]     // 辛→巽
	default: return guaTable[0]
	}
}

func zhuNaJia(p ganzhi.Zhu) gua { return ganNaJia(p.Gan) }

// MingGua is the 命卦 result.
type MingGua struct {
	Gua       gua    `json:"gua"`
	Group     string `json:"group"`
}

// ComputeMingGua computes the 命卦 from gender and birth year.
func ComputeMingGua(gender ganzhi.Gender, birthYear int) MingGua {
	n := (birthYear%100 - 4) % 9
	if n <= 0 { n += 9 }
	if gender == ganzhi.Female {
		n = 11 - n
		if n > 9 { n = 1 + (n-1)%9 }
	}
	if n == 5 {
		if gender == ganzhi.Male { n = 2 } else { n = 8 }
	}
	g := guaTable[n]
	group := "东四命"
	if westGroup[n] {
		group = "西四命"
	}
	return MingGua{Gua: g, Group: group}
}

// Chart is the complete八宅合参 result.
type Chart struct {
	MingGua     MingGua          `json:"ming_gua"`
	BaZhaiDirs  baZhaiDirections `json:"ba_zhai_dirs"`
	YearStars   yearStarResult   `json:"year_stars"`
	ZhuBagua [4]gua           `json:"pillar_bagua"`
}

// computeChart computes a complete八宅合参 from bazi, gender and birth year.
func computeChart(bz ganzhi.Bazi, gender ganzhi.Gender, year int) Chart {
	mg := ComputeMingGua(gender, year)
	return Chart{
		MingGua:    mg,
		BaZhaiDirs: baZhaiDirectionsForGua(mg.Gua.Index),
		YearStars:  computeYearStars(year),
		ZhuBagua: [4]gua{
			zhuNaJia(bz.Nian),
			zhuNaJia(bz.Yue),
			zhuNaJia(bz.Ri),
			zhuNaJia(bz.Shi),
		},
	}
}
