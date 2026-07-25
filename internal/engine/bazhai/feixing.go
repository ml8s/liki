package bazhai

import "liki-engine/internal/engine/fengshui"

// -- 紫白年飞星 (Annual Purple-White Flying Stars) -------------------------------

type yearStarResult struct {
	Year       int                `json:"year"`
	CenterStar fengshui.FlyingStar `json:"center_star"`
	Palaces    [9]palaceStar      `json:"palaces"`
}

type palaceStar struct {
	PalaceNum int                `json:"palace_num"`
	Star      fengshui.FlyingStar `json:"star"`
}

// computeYearStars computes the annual purple-white flying star distribution.
//
// Formula: 下元甲子(1984)七赤入中, each year the center star decreases by 1.
func computeYearStars(year int) yearStarResult {
	var jiaZiYear int
	var jiaZiStar int
	switch {
	case year >= 1984:
		jiaZiYear, jiaZiStar = 1984, 7
	case year >= 1924:
		jiaZiYear, jiaZiStar = 1924, 4
	case year >= 1864:
		jiaZiYear, jiaZiStar = 1864, 1
	default:
		cycle := (1864 - year) / 60
		if (1864-year)%60 != 0 {
			cycle++
		}
		jiaZiYear = 1864 - cycle*60
		nBack := (1864 - jiaZiYear) / 60
		phase := (3 - nBack%3) % 3
		jiaZiStar = []int{1, 4, 7}[phase]
	}

	diff := year - jiaZiYear
	centerNum := (jiaZiStar - diff%9 + 9) % 9
	if centerNum == 0 {
		centerNum = 9
	}

	centerStar := fengshui.StarByNumber(centerNum)

	var palaces [9]palaceStar
	palaces[4] = palaceStar{PalaceNum: 5, Star: centerStar}

	for i, pn := range fengshui.LuoshuFlyOrder {
		starNum := (centerNum + i + 1) % 9
		if starNum == 0 {
			starNum = 9
		}
		palaces[pn-1] = palaceStar{PalaceNum: pn, Star: fengshui.StarByNumber(starNum)}
	}

	return yearStarResult{Year: year, CenterStar: centerStar, Palaces: palaces}
}

// -- 八宅 四吉四凶 ------------------------------------------

type dirPattern struct {
	shengQi, tianYi, yanNian, fuWei int
	huoHai, wuGui, liuSha, jueMing  int
}

// Standard 八宅大游年 patterns per gua number.
// Four auspicious: 生气, 天医, 延年, 伏位
// Four inauspicious: 祸害, 五鬼, 六煞, 绝命
func eightMansionDirs(guaNum int) (auspicious [4]int, inauspicious [4]int) {
	p, ok := eightMansionPatterns[guaNum]
	if !ok {
		return
	}
	auspicious = [4]int{p.shengQi, p.tianYi, p.yanNian, p.fuWei}
	inauspicious = [4]int{p.huoHai, p.wuGui, p.liuSha, p.jueMing}
	return
}

// baZhaiDirections holds the八宅 four-auspicious-four-inauspicious directions by name.
type baZhaiDirections struct {
	ShengQi []string `json:"sheng_qi"`
	TianYi  []string `json:"tian_yi"`
	YanNian []string `json:"yan_nian"`
	FuWei   []string `json:"fu_wei"`
	HuoHai  []string `json:"huo_hai"`
	WuGui   []string `json:"wu_gui"`
	LiuSha  []string `json:"liu_sha"`
	JueMing []string `json:"jue_ming"`
}

func baZhaiDirectionsForGua(guaNum int) baZhaiDirections {
	aus, inaus := eightMansionDirs(guaNum)
	dirs := palaceDirs
	return baZhaiDirections{
		ShengQi: []string{dirs[aus[0]]},
		TianYi:  []string{dirs[aus[1]]},
		YanNian: []string{dirs[aus[2]]},
		FuWei:   []string{dirs[aus[3]]},
		HuoHai:  []string{dirs[inaus[0]]},
		WuGui:   []string{dirs[inaus[1]]},
		LiuSha:  []string{dirs[inaus[2]]},
		JueMing: []string{dirs[inaus[3]]},
	}
}
