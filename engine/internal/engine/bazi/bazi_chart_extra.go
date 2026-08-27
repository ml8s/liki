package bazi

import (
	"liki-engine/internal/engine/ganzhi"
)

// ChartExtra holds supplementary chart data not needed for core analysis.
type ChartExtra struct {
	SanYuan    SanYuan             `json:"san_yuan"`
	GongJia    []GongJia           `json:"gong_jia,omitempty"`
	NayinRel   []NayinRelEntry     `json:"nayin_rel"`
	ChangSheng [12]ChangShengStage `json:"chang_sheng"`
	SanQiName  string              `json:"san_qi_name,omitempty"`
}

// NayinRelEntry describes a nayin relation between two pillars.
type NayinRelEntry struct {
	A        string `json:"a"`
	B        string `json:"b"`
	Relation string `json:"relation"`
}

// ChangShengStage describes one of the 12 life stages.
type ChangShengStage struct {
	Name  string     `json:"name"`
	Index ganzhi.Zhi `json:"index"`
}

func ComputeChartExtra(c Chart) ChartExtra {
	bz := c.ToBazi()

	stages := [12]ChangShengStage{}
	zhi := ganzhi.ChangShengTable[c.Ri.Gan]
	for i := 0; i < 12; i++ {
		stages[i] = ChangShengStage{
			Name:  ganzhi.StageNamesZH[i],
			Index: zhi[i],
		}
	}

	nayins := c.NaYinArray()
	var nayinRels []NayinRelEntry
	for i := 0; i < 4; i++ {
		for j := i + 1; j < 4; j++ {
			ae := ganzhi.NayinWuxing(nayins[i])
			be := ganzhi.NayinWuxing(nayins[j])
			rel := "相同"
			if ae != 0 && be != 0 && ae != be {
				if ganzhi.Sheng(ae, be) || ganzhi.Sheng(be, ae) {
					rel = "相生"
				} else {
					rel = "相克"
				}
			}
			nayinRels = append(nayinRels, NayinRelEntry{
				A: zhuLabels[i], B: zhuLabels[j], Relation: rel,
			})
		}
	}

	return ChartExtra{
		SanYuan:    computeSanYuan(bz.Yue, bz.Nian.Gan, bz.Shi.Zhi),
		GongJia:    computeGongJia(bz),
		NayinRel:   nayinRels,
		ChangSheng: stages,
		SanQiName:  sanQiName(sanQiType(bz)),
	}
}
