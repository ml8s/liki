package bazhai

import (
	"liki-engine/internal/engine/ganzhi"
	"liki-engine/internal/engine/luoshu"
)

// gua is a bagua trigram (bazhai wire contract: index/name/wuxing/yin_yang).
type gua struct {
	Index   int    `json:"index"`
	Name    string `json:"name"`
	Wuxing  string `json:"wuxing"`
	YinYang string `json:"yin_yang"`
}

// guaTable maps 洛书数 (1-9) to trigrams, sourced from the shared luoshu table
// (names/wuxing/yin_yang) so there is a single 洛书九宫 truth. 中宫 (5) 的寄宫
// 由 ComputeMingGua 显式重映射（男寄坤/女寄艮），此处不再重复。
var guaTable [10]gua

func init() {
	for n := 1; n <= 9; n++ {
		p := luoshu.PalaceTable[n]
		guaTable[n] = gua{Index: n, Name: p.Name, Wuxing: p.Element.String(), YinYang: p.YinYang}
	}
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
