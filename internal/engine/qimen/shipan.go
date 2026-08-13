package qimen

import (
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
		RiGanPalace:      findGanPalaceIdx(p, bz.Ri.Gan),
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
