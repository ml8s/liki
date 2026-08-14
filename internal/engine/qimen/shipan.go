package qimen

import (
	"fmt"

	"liki-engine/internal/engine/ganzhi"
)

// Chart bundles a complete奇门盘 with all analysis layers.
type Chart struct {
	Pan              pan               `json:"pan"`
	GanInteractions [9]GanInteraction `json:"gan_interaction"`
	MenInteractions [9]MenInteraction `json:"men_interaction"`
	XingInteractions [9]XingInteraction `json:"xing_interaction"`
	WangShuai        [9]WangShuai       `json:"wang_shuai"`
	MenPo            []GongIndex      `json:"men_po"`
	MenZhi           []GongIndex      `json:"men_zhi"`
	Patterns         []Pattern          `json:"patterns"`
	YingQi           YingQi             `json:"ying_qi"`
	RiGanPalace      GongIndex          `json:"ri_gan_gong"`       // 日干落宫（排盘固有）
	ShiGanPalace     GongIndex          `json:"shi_gan_gong"`      // 时干落宫（排盘固有）
	RiShiShengKe     string             `json:"ri_shi_sheng_ke"`   // 日干宫-时干宫五行生克（确定性派生）
	KongWangAffected bool               `json:"kong_wang_affected"` // 日干宫或时干宫是否空亡（确定性派生）
	MaXingAffected   bool               `json:"ma_xing_affected"`   // 日干宫或时干宫是否马星（确定性派生）
	DutyStarPalace   GongIndex          `json:"zhi_fu_xing_gong"`  // 值符星落宫（排盘固有）
	DutyDoorPalace   GongIndex          `json:"zhi_shi_men_gong"`  // 值使门落宫（排盘固有）
}

// computeChart computes a complete奇门盘 with all analyses.
func computeChart(bz ganzhi.Bazi, kind ChartKind, y, m, d int) Chart {
	ju := determineJuShu(y, m, d, bz.Ri.Gan, bz.Ri.Zhi)

	var driveZhu ganzhi.Zhu
	switch kind {
	case RiQiMen:
		driveZhu = bz.Ri
	case YueQiMen:
		driveZhu = bz.Yue
	case NianQiMen:
		driveZhu = bz.Nian
	default: // ShiQiMen
		driveZhu = bz.Shi
	}

	p := computePan(ju, driveZhu, bz.Ri.Gan)
	p.RiGan = bz.Ri.Gan
	p.RiZhi = bz.Ri.Zhi

	riGanP := findGanPalaceIdx(p, bz.Ri.Gan)
	shiGanP := findGanPalaceIdx(p, p.DriveGan)

	kongWangAffected := false
	for _, k := range p.KongWang {
		if k == riGanP || k == shiGanP {
			kongWangAffected = true
			break
		}
	}

	return Chart{
		Pan:              p,
		GanInteractions: computeGanInteractions(p),
		MenInteractions: computeMenInteractions(p),
		XingInteractions: computeXingInteractions(p),
		WangShuai:        computeWangShuai(p),
		MenPo:            findMenPo(p),
		MenZhi:           findMenZhi(p),
		Patterns:         findPatterns(p),
		YingQi:           computeYingQi(p),
		RiGanPalace:      riGanP,
		ShiGanPalace:     shiGanP,
		RiShiShengKe:     analyzeShengKe(riGanP, shiGanP),
		KongWangAffected: kongWangAffected,
		MaXingAffected:   p.MaXing == riGanP || p.MaXing == shiGanP,
		DutyStarPalace:   findStarPalaceIdx(p, p.DutyStar),
		DutyDoorPalace:   findDoorPalaceIdx(p, p.DutyDoor),
	}
}


// findGanPalaceIdx finds which gong a heavenly stem resides in (earth plate).
func findGanPalaceIdx(p pan, g ganzhi.Gan) GongIndex {
	for i, pg := range p.GongWei {
		if pg.EarthStem == g || pg.HeavenStem == g {
			return GongIndex(i + 1)
		}
	}
	return 0
}

// analyzeShengKe analyzes the 五行生克 between 日干宫 and 时干宫（确定性派生）。
func analyzeShengKe(subjectP, eventP GongIndex) string {
	if subjectP > 0 && eventP > 0 {
		sp := palaceWuxing(subjectP)
		ep := palaceWuxing(eventP)
		if sp == ep {
			return fmt.Sprintf("日干(%d宫)与时干(%d宫)比和", subjectP, eventP)
		}
		if ganzhi.Sheng(sp, ep) {
			return fmt.Sprintf("日干(%d宫)生时干(%d宫)", subjectP, eventP)
		}
		if ganzhi.Sheng(ep, sp) {
			return fmt.Sprintf("时干(%d宫)生日干(%d宫)", eventP, subjectP)
		}
		if ganzhi.Ke(sp, ep) {
			return fmt.Sprintf("日干(%d宫)克时干(%d宫)", subjectP, eventP)
		}
		return fmt.Sprintf("时干(%d宫)克日干(%d宫)", eventP, subjectP)
	}
	return "无显著生克关系"
}

// findStarPalaceIdx finds which gong a star resides in.
func findStarPalaceIdx(p pan, s StarIndex) GongIndex {
	for i, pg := range p.GongWei {
		if pg.Star == s {
			return GongIndex(i + 1)
		}
	}
	return 0
}

// findDoorPalaceIdx finds which gong a door resides in.
func findDoorPalaceIdx(p pan, d DoorIndex) GongIndex {
	for i, pg := range p.GongWei {
		if pg.Door == d {
			return GongIndex(i + 1)
		}
	}
	return 0
}
