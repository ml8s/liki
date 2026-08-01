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
	}
}
