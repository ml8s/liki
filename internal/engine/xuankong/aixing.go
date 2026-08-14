package xuankong

import (
	"liki-engine/internal/engine/fengshui"
)

// xuanKongStar holds the three stars (运星, 山星, 向星) for one palace.
type xuanKongStar struct {
	PalaceNum    int                  `json:"gong_num"`
	PeriodStar   fengshui.FlyingStar  `json:"yun_xing"`
	MountainStar fengshui.FlyingStar  `json:"shan_xing"`
	FacingStar   fengshui.FlyingStar  `json:"xiang_xing"`
}

// Chart is the complete 玄空飞星排盘 for a given坐向 and 运.
type Chart struct {
	Yun           SanYuanYun     `json:"yun"`
	SitMountain   int            `json:"zuo_shan"`  // 0-23,坐山 index
	FaceMountain  int            `json:"xiang_shan"` // 0-23,朝向 index
	Palaces       [9]xuanKongStar `json:"gong_wei"`
	WangShan      bool           `json:"wang_shan"`      // 旺山：坐宫山星=当令
	WangXiang     bool           `json:"wang_xiang"`     // 旺向：向宫向星=当令
	ShanXing      bool           `json:"shan_xing"`      // 双星会坐：坐宫山向星皆=当令
	XiangXing     bool           `json:"xiang_xing"`     // 双星会向：向宫山向星皆=当令
	XiaShui       bool           `json:"xia_shui"`       // 上山下水：向宫山星=当令 且 坐宫向星=当令
	FanYin        bool           `json:"fan_yin"`        // 运盘反吟（恒 false，运盘恒顺飞）
	FuYin         bool           `json:"fu_yin"`         // 运盘伏吟（五运顺飞全盘重合）
	XingJiaHui    [9]xingJiaHui  `json:"xing_jia_hui"`
	ShouShanChuSha shouShanChuSha `json:"shou_shan_chu_sha"`
}

func computeChart(sitMountain, faceMountain int, year int) Chart {
	if sitMountain < 0 || sitMountain > 23 || faceMountain < 0 || faceMountain > 23 {
		return Chart{}
	}

	yun := ComputeSanYuanYun(year)
	yunNum := yun.YunNumber

	// 1. Period star distribution.
	periodStars := flyStars(yunNum, true)

	// 2. Mountain star.
	sitPalace := mountainPalace(sitMountain)
	sitPeriodStar := periodStars[sitPalace-1]
	shanNum := sitPeriodStar.Number
	// 下卦（正向）不用替星：API 输入为山向 index（无兼向度数），按正向排盘。
	shanForward := shanXiangForward(shanNum, sitMountain)
	mountainStars := flyStars(shanNum, shanForward)

	// 3. Facing star.
	facePalace := mountainPalace(faceMountain)
	facePeriodStar := periodStars[facePalace-1]
	xiangNum := facePeriodStar.Number
	// 同上：正向不用替星。
	xiangForward := shanXiangForward(xiangNum, faceMountain)
	facingStars := flyStars(xiangNum, xiangForward)

	// 4. Assemble the pan.
	pan := Chart{
		Yun:          yun,
		SitMountain:  sitMountain,
		FaceMountain: faceMountain,
	}

	for i := 0; i < 9; i++ {
		pan.Palaces[i] = xuanKongStar{
			PalaceNum:    i + 1,
			PeriodStar:   periodStars[i],
			MountainStar: mountainStars[i],
			FacingStar:   facingStars[i],
		}
	}

	// 5. Evaluate 旺山旺向/上山下水/反吟伏吟.
	pan.evaluate()

	// 6. 双星加会 + 收山出煞.
	pan.XingJiaHui = pan.computeXingJiaHui()
	pan.ShouShanChuSha = pan.computeShouShanChuSha()

	return pan
}

// shanXiangForward determines whether a mountain/facing star board flies forward.
//
// 规则（《沈氏玄空学》《易学经世真诠》）：入中星 n 对应元旦盘（洛书）某宫，
// 在该宫三山中取与坐山/向首同元龙（天/地/人）的一山，其阴阳定顺逆——阳顺阴逆。
// 例：七运子山午向，山星 3 入中（3=震宫 甲卯乙，子=天元 → 卯=阴 → 逆飞）；
// 向星 2 入中（2=坤宫 未坤申，午=天元 → 坤=阳 → 顺飞）→ 双星会坐。
// 入中星为 5（五黄无卦）时，按坐山/向首自身三元龙阴阳（5 落其宫）。
func shanXiangForward(centerNum int, mountainIdx int) bool {
	if centerNum == 5 {
		return fengshui.Mountains24Table[mountainIdx].YinYang == "阳"
	}
	trigram := luoshuPalaceName[centerNum]
	target := fengshui.Mountains24Table[mountainIdx]
	for i := 0; i < 24; i++ {
		m := fengshui.Mountains24Table[i]
		if m.Trigram == trigram && m.YuanLong == target.YuanLong {
			return m.YinYang == "阳"
		}
	}
	return false
}

// luoshuPalaceName maps a flying-star number to its 洛书（元旦盘）palace trigram.
var luoshuPalaceName = [10]string{"", "坎", "坤", "震", "巽", "中", "乾", "兑", "艮", "离"}

// flyStars distributes num stars following luoshu fly order.
func flyStars(centerNum int, forward bool) [9]fengshui.FlyingStar {
	var stars [9]fengshui.FlyingStar
	stars[4] = fengshui.StarByNumber(centerNum)

	for i, pn := range fengshui.LuoshuFlyOrder {
		var starNum int
		if forward {
			starNum = (centerNum + i + 1) % 9
		} else {
			starNum = (centerNum - i - 1 + 9) % 9
		}
		if starNum == 0 {
			starNum = 9
		}
		stars[pn-1] = fengshui.StarByNumber(starNum)
	}
	return stars
}

// mountainPalace returns which palace (1-9) a given 24-mountain index belongs to.
func mountainPalace(idx int) int {
	idx = idx % 24
	switch {
	case idx <= 1 || idx == 23:
		return 1 // 坎(子癸壬)
	case idx >= 2 && idx <= 4:
		return 8 // 艮(丑艮寅)
	case idx >= 5 && idx <= 7:
		return 3 // 震(甲卯乙)
	case idx >= 8 && idx <= 10:
		return 4 // 巽(辰巽巳)
	case idx >= 11 && idx <= 13:
		return 9 // 离(丙午丁)
	case idx >= 14 && idx <= 16:
		return 2 // 坤(未坤申)
	case idx >= 17 && idx <= 19:
		return 7 // 兑(庚酉辛)
	default: // 20-22
		return 6 // 乾(戌乾亥)
	}
}

// evaluate computes the four 格局 (四大局) using the standard definitions
// (《沈氏玄空学》, 当令=运星数):
//
//	旺山旺向   : 坐宫山星=当令 且 向宫向星=当令
//	双星会坐   : 坐宫山星=当令 且 坐宫向星=当令（财星上山）
//	双星会向   : 向宫山星=当令 且 向宫向星=当令（丁星下水）
//	上山下水   : 向宫山星=当令（山星下水）且 坐宫向星=当令（向星上山）
//
// 伏吟（运盘）：五运运盘顺飞与地盘全盘重合（入中=宫序恒等）；运盘恒顺飞，
// 故运盘无反吟（fan_yin 恒 false，反吟须看山向星盘与运盘对冲，此处不判定）。
func (p *Chart) evaluate() {
	sitPalace := mountainPalace(p.SitMountain)
	facePalace := mountainPalace(p.FaceMountain)
	yunNum := p.Yun.YunNumber

	sitMStar := p.Palaces[sitPalace-1].MountainStar.Number
	sitFStar := p.Palaces[sitPalace-1].FacingStar.Number
	faceMStar := p.Palaces[facePalace-1].MountainStar.Number
	faceFStar := p.Palaces[facePalace-1].FacingStar.Number

	p.WangShan = sitMStar == yunNum
	p.WangXiang = faceFStar == yunNum
	p.ShanXing = sitMStar == yunNum && sitFStar == yunNum  // 双星会坐
	p.XiangXing = faceMStar == yunNum && faceFStar == yunNum // 双星会向
	p.XiaShui = faceMStar == yunNum && sitFStar == yunNum   // 上山下水

	// 运盘伏吟：运盘与地盘全盘重合（仅五运顺飞成立）。
	p.FuYin = p.Yun.YunNumber == 5 && p.Palaces[4].PeriodStar.Number == 5
	p.FanYin = false // 运盘恒顺飞，无反吟
}

func tiXingShanStar(sitIdx int) int {
	return needTiXing[sitIdx%24]
}

// tiXingXiangStar returns the 替星 for a facing mountain (兼向替卦用，规则同上).
func tiXingXiangStar(faceIdx int) int {
	return needTiXing[faceIdx%24]
}

// needTiXing maps mountain index → 替星数（替卦十三山）；非十三山无替（0）。
var needTiXing = [24]int{
	0, // 子(0)
	0, // 癸(1)
	7, // 丑(2) → 破军7
	7, // 艮(3) → 破军7（天元龙亦替）
	9, // 寅(4) → 右弼9
	1, // 甲(5) → 贪狼1
	2, // 卯(6) → 巨门2
	2, // 乙(7) → 巨门2
	6, // 辰(8) → 武曲6
	6, // 巽(9) → 武曲6（天元龙亦替）
	6, // 巳(10) → 武曲6
	7, // 丙(11) → 破军7
	0, // 午(12)
	0, // 丁(13)
	0, // 未(14)
	0, // 坤(15)
	1, // 申(16) → 贪狼1
	9, // 庚(17) → 右弼9
	0, // 酉(18)
	0, // 辛(19)
	0, // 戌(20)
	0, // 乾(21)
	0, // 亥(22)
	2, // 壬(23) → 巨门2
}

// -- 双星加会 (Double Star Combination) --------------------------------

type xingJiaHui struct {
	ShanNum    int    `json:"shan_num"`
	XiangNum   int    `json:"xiang_num"`
	Name       string `json:"name"`
	Meaning    string `json:"meaning"`
	Auspicious bool   `json:"auspicious"`
}

func (p *Chart) computeXingJiaHui() [9]xingJiaHui {
	var result [9]xingJiaHui
	for i, pal := range p.Palaces {
		key := [2]int{pal.MountainStar.Number, pal.FacingStar.Number}
		if entry, ok := xingJiaHuiTable[key]; ok {
			result[i] = entry
		} else {
			result[i] = xingJiaHui{
				ShanNum:    pal.MountainStar.Number,
				XiangNum:   pal.FacingStar.Number,
				Name:       "双星到向",
				Meaning:    "山向配合，需参合判断",
				Auspicious: pal.MountainStar.Auspicious && pal.FacingStar.Auspicious,
			}
		}
	}
	return result
}

// -- 收山出煞 (Mountain Containment & Sha Removal) --------------------

type shouShanChuSha struct {
	ZhengShen  int    `json:"zheng_shen"`
	LingShen   int    `json:"ling_shen"`
	ShouShanOK bool   `json:"shou_shan"`
	ChuShaOK   bool   `json:"chu_sha"`
	Assessment string `json:"assessment"`
}

func (p *Chart) computeShouShanChuSha() shouShanChuSha {
	yunNum := p.Yun.YunNumber
	zhengShen := yunNum
	lingShen := 10 - yunNum

	sitPalace := mountainPalace(p.SitMountain)
	facePalace := mountainPalace(p.FaceMountain)

	sitMStar := p.Palaces[sitPalace-1].MountainStar.Number
	faceFStar := p.Palaces[facePalace-1].FacingStar.Number

	shouShanOK := sitMStar == zhengShen
	chuShaOK := faceFStar == lingShen

	var assessment string
	if shouShanOK && chuShaOK {
		assessment = "收山出煞俱得，丁财两旺"
	} else if shouShanOK {
		assessment = "收山得宜，旺丁；出煞未得，财弱"
	} else if chuShaOK {
		assessment = "出煞得宜，旺财；收山未得，丁弱"
	} else {
		assessment = "收山出煞俱失，宜择时改向"
	}

	return shouShanChuSha{
		ZhengShen:  zhengShen,
		LingShen:   lingShen,
		ShouShanOK: shouShanOK,
		ChuShaOK:   chuShaOK,
		Assessment: assessment,
	}
}
