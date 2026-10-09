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

// MingGua is the 命卦 result.
type MingGua struct {
	Gua          gua    `json:"gua"`
	Group        string `json:"group"`
	YearBoundary string `json:"year_boundary"`
}

// ComputeMingGua computes the 命卦 from gender and birth year.
//
// 公式（《八宅明镜》通行算法，2000 年为分界）：
//
//	2000 年前：男 (100-年后两位)%9；女 (年后两位-4)%9
//	2000 年后：男 (99-年后两位)%9； 女 (年后两位+6)%9
//	余数 0 → 9（离）；余数 5 → 男寄坤、女寄艮
func ComputeMingGua(gender ganzhi.Gender, birthYear int) MingGua {
	y := birthYear % 100
	var n int
	if birthYear < 2000 {
		if gender == ganzhi.Male {
			n = (100 - y) % 9
		} else {
			n = ((y-4)%9 + 9) % 9
		}
	} else {
		if gender == ganzhi.Male {
			n = (99 - y) % 9
		} else {
			n = (y + 6) % 9
		}
	}
	if n == 0 {
		n = 9
	}
	if n == 5 {
		if gender == ganzhi.Male {
			n = 2
		} else {
			n = 8
		}
	}
	g := guaTable[n]
	group := "东四命"
	if westGroup[n] {
		group = "西四命"
	}
	return MingGua{Gua: g, Group: group, YearBoundary: "gregorian_calendar_year"}
}

// Chart is the complete八宅合参 result.
type Chart struct {
	MingGua    MingGua          `json:"ming_gua"`
	BaZhaiDirs baZhaiDirections `json:"ba_zhai_dirs"`
	YearStars  yearStarResult   `json:"liu_nian_xing"`
}
