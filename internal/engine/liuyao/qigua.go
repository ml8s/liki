package liuyao

import (
	"math/rand"

	"liki-engine/internal/engine/ganzhi"
)

func shakeCoins(rng *rand.Rand) [6]YaoType {
	var yaos [6]YaoType
	for i := 0; i < 6; i++ {
		sum := 0
		for j := 0; j < 3; j++ {
			if rng.Intn(2) == 0 { sum += 2 } else { sum += 3 }
		}
		yaos[i] = YaoType(sum)
	}
	return yaos
}

func shakeCoinsFixed(results [6]int) [6]YaoType {
	var y [6]YaoType
	for i := 0; i < 6; i++ { y[i] = YaoType(results[i]) }
	return y
}

func yaoTypeToYang(y YaoType) int {
	if y.IsYang() { return 1 }
	return 0
}

// yaosToBin encodes 6 yaos as a 0-63 binary index.
// Upper trigram = lines 4-6, lower = lines 1-3; yang=1, yin=0.
// Trigram: 乾=7, 兑=6, 离=5, 震=4, 巽=3, 坎=2, 艮=1, 坤=0.
func yaosToBin(yaos [6]YaoType) int {
	upper, lower := 0, 0
	for i := 0; i < 3; i++ { upper = upper<<1 | yaoTypeToYang(yaos[5-i]) }
	for i := 0; i < 3; i++ { lower = lower<<1 | yaoTypeToYang(yaos[2-i]) }
	return upper*8 + lower
}

func yaosToGua(yaos [6]YaoType) guaIndex {
	return binaryToGuaTable[yaosToBin(yaos)]
}

// binaryToGuaTable maps binary encoding (upper*8+lower) to guaTable position (palace order).
// Trigram encoding from yaosToBin (top line at MSB):
//
//	乾=7(111) 兑=3(011) 离=5(101) 震=1(001)
//	巽=6(110) 坎=2(010) 艮=4(100) 坤=0(000)
var binaryToGuaTable = [64]guaIndex{
	56, 57, 47, 58, 13, 46, 28, 59, //  0-7:   坤地 地雷复 地水师 地泽临 地山谦 地火明夷 地风升 地天泰
	25, 24, 26, 15, 14, 45, 27, 60, //  8-15:  雷地豫 震为雷 雷水解 雷泽归妹 雷山小过 雷火丰 雷风恒 雷天大壮
	63, 42, 40, 41, 12, 43, 29, 62, // 16-23: 水地比 水雷屯 坎为水 水泽节 水山蹇 水火既济 水风井 水天需
	10, 31, 9, 8, 11, 44, 30, 61, // 24-31: 泽地萃 泽雷随 泽水困 兑为泽 泽山咸 泽火革 泽风大过 泽天夬
	5, 38, 20, 51, 48, 49, 39, 50, // 32-39: 山地剥 山雷颐 山水蒙 山泽损 艮为山 山火贲 山风蛊 山天大畜
	6, 37, 19, 52, 17, 16, 18, 7, // 40-47: 火地晋 火雷噬嗑 火水未济 火泽睽 火山旅 离为火 火风鼎 火天大有
	4, 35, 21, 54, 55, 34, 32, 33, // 48-55: 风地观 风雷益 风水涣 风泽中孚 风山渐 风火家人 巽为风 风天小畜
	3, 36, 22, 53, 2, 23, 1, 0, // 56-63: 天地否 天雷无妄 天水讼 天泽履 天山遁 天火同人 天风姤 乾为天
}

func dongYao(yaos [6]YaoType) []int {
	var dy []int
	for i := 0; i < 6; i++ {
		if yaos[i].IsChanging() { dy = append(dy, i+1) }
	}
	return dy
}

func invertDongYao(benGua guaIndex, dy []int) (guaIndex, bool) {
	if len(dy) == 0 { return 0, false }
	// Find binary encoding for current guaTable index.
	var benBin int
	for bin, gIdx := range binaryToGuaTable {
		if gIdx == benGua {
			benBin = bin
			break
		}
	}
	val := benBin
	for _, pos := range dy { val ^= 1 << (pos - 1) }
	return binaryToGuaTable[val], true
}

func computeGuaPan(yaos [6]YaoType, riZhu ganzhi.Zhu) Chart {
	benGua := yaosToGua(yaos)
	meta := guaTable[benGua]
	dy := dongYao(yaos)
	bianGua, hasBian := invertDongYao(benGua, dy)

	benElem := palaceWuxing[meta.PalaceIdx]
	lines := zhuangGua(benGua, riZhu.Gan, false, benElem)
	for i := 0; i < 6; i++ { lines[i].Position, lines[i].Type = i+1, yaos[i] }

	var bianLines [6]Line
	if hasBian {
		bianLines = zhuangGua(bianGua, riZhu.Gan, true, benElem)
		for i := 0; i < 6; i++ {
			bianLines[i].Position = i + 1
			if yaos[i].IsChanging() {
				if yaos[i] == LaoYang {
					bianLines[i].Type = ShaoYin
				} else {
					bianLines[i].Type = ShaoYang
				}
			} else {
				bianLines[i].Type = yaos[i]
			}
		}
	}

	return Chart{
		Name:         meta.Name,
		BenGua:       benGua,
		BianGua:      bianGua,
		Palace:       palaceNames[meta.PalaceIdx],
		PalaceWuxing: palaceWuxing[meta.PalaceIdx],
		Lines:        lines,
		BianLines:    bianLines,
		DayGan:       riZhu.Gan,
		DayZhi:       riZhu.Zhi,
		DongYao:      dy,
	}
}

func zhuangGua(gua guaIndex, dayGan ganzhi.Gan, isBian bool, palaceElem ganzhi.Wuxing) [6]Line {
	meta := guaTable[gua]
	naZhi := naZhiTable[meta.PalaceIdx]
	naGan := naGanTable[meta.PalaceIdx]
	shouOrder := dayGanShouOrder(dayGan)
	var lines [6]Line
	for i := 0; i < 6; i++ {
		z := naZhi[i]
		zwx := ganzhi.ZhiWuxing(z)
		qin := computeLiuQin(zwx, palaceElem)
		shiYing := ""
		if !isBian {
			shi := meta.ShiPos - 1
			if i == shi { shiYing = "世" } else if i == (shi+3)%6 { shiYing = "应" }
		}
		lines[i] = Line{Gan: naGan, Zhi: z, Wuxing: zwx, LiuQin: qin, ShiYing: shiYing, LiuShou: shouOrder[i]}
	}
	return lines
}

func computeLiuQin(lineElem, palaceElem ganzhi.Wuxing) LiuQin {
	if lineElem == palaceElem {
		return QinXiongDi
	}
	if ganzhi.Sheng(palaceElem, lineElem) {
		return QinZiSun
	}
	if ganzhi.Sheng(lineElem, palaceElem) {
		return QinFumu
	}
	if ganzhi.Ke(palaceElem, lineElem) {
		return QinQiCai
	}
	return QinGuanGui
}

// YongShenResult holds the 用神 analysis result.
type YongShenResult struct {
	Name     string   `json:"name"` // 用神六亲名
	Position int      `json:"position"` // line position 1-6, 0 if not found
	FuShen   *FuShen  `json:"fu_shen,omitempty"`
}

// computeChart computes a complete 六爻 chart from bazi, question type, and yaos (required).
func computeChart(bz ganzhi.Bazi, yongShen YongShen, yaos [6]int) Chart {
	yts := shakeCoinsFixed(yaos)
	chart := computeGuaPan(yts, bz.Ri)

	// Month building from bazi.
	chart.MonthZhi = bz.Yue.Zhi
	chart.MonthGan = bz.Yue.Gan

	// 用神.
	pos, _ := chart.findYongShen(yongShen)
	chart.YongShen = YongShenResult{Name: yongShen.String(), Position: pos}
	if pos == 0 {
		chart.YongShen.FuShen = chart.findFuShen(yongShen)
	}

	// 旺衰 + 日建关系.
	for i := 0; i < 6; i++ {
		chart.WangShuai[i] = ganzhi.WangShuaiOf(ganzhi.ZhiWuxing(chart.Lines[i].Zhi), chart.MonthZhi)
		chart.DayRelations[i] = dayInteraction(chart.Lines[i].Zhi, chart.DayZhi)
	}

	// 应期.
	chart.YingQi = computeYingQi(&chart, yongShen)

	gc, err := GetGuaCi(int(chart.BenGua))
	if err == nil {
		chart.GuaCi = gc
	}
	return chart
}
