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
	SitMountain   int            `json:"sit_mountain"`  // 0-23,坐山 index
	FaceMountain  int            `json:"face_mountain"` // 0-23,朝向 index
	Palaces       [9]xuanKongStar `json:"gong_wei"`
	WangShan      bool           `json:"wang_shan"`
	WangXiang     bool           `json:"wang_xiang"`
	ShanXing      bool           `json:"shan_xing"`
	XiaShui       bool           `json:"xia_shui"`
	FanYin        bool           `json:"fan_yin"`
	FuYin         bool           `json:"fu_yin"`
	XingJiaHui    [9]xingJiaHui  `json:"xing_jia_hui"`
	ShouShanChuSha shouShanChuSha `json:"shou_shan_chu_sha"`
}

func computeChart(sitMountain, faceMountain int, year int) Chart {
	if sitMountain < 0 || sitMountain > 23 || faceMountain < 0 || faceMountain > 23 {
		return Chart{}
	}

	sit := fengshui.Mountains24Table[sitMountain]
	face := fengshui.Mountains24Table[faceMountain]

	yun := ComputeSanYuanYun(year)
	yunNum := yun.YunNumber

	// 1. Period star distribution.
	periodStars := flyStars(yunNum, true)

	// 2. Mountain star.
	sitPalace := mountainPalace(sitMountain)
	sitPeriodStar := periodStars[sitPalace-1]
	shanNum := sitPeriodStar.Number

	if ti := tiXingShanStar(sitMountain); ti > 0 {
		shanNum = ti
	}
	shanForward := sit.YinYang == "阳"
	mountainStars := flyStars(shanNum, shanForward)

	// 3. Facing star.
	facePalace := mountainPalace(faceMountain)
	facePeriodStar := periodStars[facePalace-1]
	xiangNum := facePeriodStar.Number

	if ti := tiXingXiangStar(faceMountain); ti > 0 {
		xiangNum = ti
	}
	xiangForward := face.YinYang == "阳"
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

func (p *Chart) evaluate() {
	sitPalace := mountainPalace(p.SitMountain)
	facePalace := mountainPalace(p.FaceMountain)
	yunNum := p.Yun.YunNumber

	sitMStar := p.Palaces[sitPalace-1].MountainStar.Number
	faceFStar := p.Palaces[facePalace-1].FacingStar.Number

	p.WangShan = sitMStar == yunNum
	p.WangXiang = faceFStar == yunNum

	sitPStar := p.Palaces[sitPalace-1].PeriodStar.Number
	facePStar := p.Palaces[facePalace-1].PeriodStar.Number
	p.ShanXing = sitPStar != yunNum && sitMStar != yunNum
	p.XiaShui = facePStar != yunNum && faceFStar != yunNum

	opposite := 10 - sitPalace
	p.FanYin = p.Palaces[opposite-1].PeriodStar.Number == yunNum

	for i := 0; i < 9; i++ {
		if p.Palaces[i].PeriodStar.Number == i+1 {
			p.FuYin = true
			break
		}
	}
}

// tiXingTable maps 24 mountain index → substitute star number (1-9).
// Based on玄空替星口诀: 子癸甲申→1(贪狼), 壬卯乙未坤→2(巨门),
// 乾亥辰巽巳戌→6(武曲), 酉辛丑艮丙→7(破军), 寅午庚丁→9(右弼).
var tiXingTable = [24]int{
	1, // 子(0) → 1 贪狼
	1, // 癸(1) → 1 贪狼
	7, // 丑(2) → 7 破军
	7, // 艮(3) → 7 破军
	9, // 寅(4) → 9 右弼
	1, // 甲(5) → 1 贪狼
	2, // 卯(6) → 2 巨门
	2, // 乙(7) → 2 巨门
	6, // 辰(8) → 6 武曲
	6, // 巽(9) → 6 武曲
	6, // 巳(10) → 6 武曲
	7, // 丙(11) → 7 破军
	9, // 午(12) → 9 右弼
	9, // 丁(13) → 9 右弼
	2, // 未(14) → 2 巨门
	2, // 坤(15) → 2 巨门
	1, // 申(16) → 1 贪狼
	9, // 庚(17) → 9 右弼
	7, // 酉(18) → 7 破军
	7, // 辛(19) → 7 破军
	6, // 戌(20) → 6 武曲
	6, // 乾(21) → 6 武曲
	6, // 亥(22) → 6 武曲
	2, // 壬(23) → 2 巨门
}

func tiXingShanStar(sitIdx int) int {
	num := tiXingTable[sitIdx%24]
	sit := fengshui.Mountains24Table[sitIdx%24]
	if sit.YuanLong != "天元龙" {
		return num
	}
	return 0
}

func tiXingXiangStar(faceIdx int) int {
	num := tiXingTable[faceIdx%24]
	face := fengshui.Mountains24Table[faceIdx%24]
	if face.YuanLong != "天元龙" {
		return num
	}
	return 0
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
