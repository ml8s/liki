package bazhai

import "liki-engine/internal/engine/fengshui"

// -- 紫白年飞星 (Annual Purple-White Flying Stars) -------------------------------

// yearStarResult is the 流年紫白飞星 inside bazhai.chart.
// schema 与玄空共用（fengshui.ComputeAnnualFlyingStars），字段与 xuankong 一致。
type yearStarResult struct {
	Year    int                          `json:"year"`
	RuZhong string                       `json:"ru_zhong"`
	Palaces []fengshui.AnnualFlyingStar `json:"gong_wei"`
}

// computeYearStars computes the annual purple-white flying star distribution
// via the shared fengshui implementation (口诀：上元甲子一白/中元四绿/下元七赤，逐年逆行).
func computeYearStars(year int) yearStarResult {
	board := fengshui.ComputeAnnualFlyingStars(year)
	return yearStarResult{Year: board.Year, RuZhong: board.RuZhong, Palaces: board.GongWei}
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
