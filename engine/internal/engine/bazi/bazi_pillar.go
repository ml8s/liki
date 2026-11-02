package bazi

import (
	"liki-engine/internal/engine/ganzhi"
)

// PillarAnalysis holds the structural projections derived directly from the
// four pillars. It is a concrete Bazi-domain value, not a generic fact bag.
type PillarAnalysis struct {
	HiddenStems [4]HiddenStemProjection
	TenGods     [4][]TenGodProjection
	LifeStages  [4]string
	SelfSitting [4]string
	NaYin       [4]string
	ShenSha     [4][]ShenShaProjection
	VoidPalaces []int
}

// AnalyzePillars derives the structural facts shared by the full-chart view
// and downstream rule evaluators.
func AnalyzePillars(bz ganzhi.Bazi, gender ganzhi.Gender) PillarAnalysis {
	hiddenStems := HiddenStemFacts(bz)
	return PillarAnalysis{
		HiddenStems: hiddenStems,
		NaYin:       computeNaYin(bz),
		TenGods:     TenGodProjections(bz, hiddenStems),
		SelfSitting: computeSelfSitting(bz),
		LifeStages:  computeDayMasterTrend(bz),
		ShenSha:     ShenShaProjections(bz, gender),
		VoidPalaces: VoidPalaces(bz),
	}
}

// assembleChart projects pillar analysis into the published Chart pillars.
func assembleChart(core NatalCore, bz ganzhi.Bazi, analysis PillarAnalysis) Chart {
	hs := analysis.HiddenStems
	ny := analysis.NaYin
	tgTable := analysis.TenGods
	selfSitting := analysis.SelfSitting
	dayMasterTrend := analysis.LifeStages
	shensha := analysis.ShenSha
	voidHits := analysis.VoidPalaces
	ps := bz.Slice()

	makeFull := func(i int) Pillar {
		isVoid := false
		for _, vh := range voidHits {
			if vh == i {
				isVoid = true
				break
			}
		}
		pi := Pillar{
			Zhu: ps[i], DayMasterTrend: dayMasterTrend[i],
			Xun: ganzhi.XunName(ps[i].Gan, ps[i].Zhi), XunKong: xunKongPairLabel(ps[i].Gan, ps[i].Zhi),
			NaYin: ny[i], CangGan: hs[i], ShiShens: tgTable[i], SelfSitting: selfSitting[i],
			ShenSha: shensha[i], IsVoid: isVoid,
		}
		pi.IsSelfHe = isSelfHe(ps[i])
		pi.SelfHeName = selfHeName(ps[i])
		pi.IsKuiGang = isKuiGang(ps[i])
		return pi
	}

	return Chart{
		BirthYear:  core.BirthYear,
		Nian:       makeFull(0),
		Yue:        makeFull(1),
		Ri:         makeFull(2),
		Shi:        makeFull(3),
		DaYun:      core.DaYun,
		Gender:     core.Gender,
		DayXun:     ganzhi.XunName(bz.Ri.Gan, bz.Ri.Zhi),
		DayXunKong: xunKongPairLabel(bz.Ri.Gan, bz.Ri.Zhi),
	}
}

// xunKongPairLabel 按干支所居之旬输出旬空两字，如甲申旬 → "午未"。
func xunKongPairLabel(gan ganzhi.Gan, zhi ganzhi.Zhi) string {
	v := ganzhi.XunKong(gan, zhi)
	return ganzhi.ZhiName(v[0]) + ganzhi.ZhiName(v[1])
}

// computeDayMasterTrend returns the day master's twelve-life stage on each pillar branch.
func computeDayMasterTrend(bz ganzhi.Bazi) [4]string {
	row := ganzhi.ChangShengTable[bz.Ri.Gan]
	var out [4]string
	for i, pillar := range bz.Slice() {
		for stage, branch := range row {
			if branch == pillar.Zhi {
				out[i] = ganzhi.StageNamesZH[stage]
				break
			}
		}
	}
	return out
}

func HiddenStemFacts(bz ganzhi.Bazi) [4]HiddenStemProjection {
	var hs [4]HiddenStemProjection
	for i, z := range bz.Slice() {
		if z.Zhi == 0 {
			continue
		}
		qi := ganzhi.CangGanForZhi(z.Zhi)
		hs[i] = HiddenStemProjection{Main: *qi.Main}
		if qi.Mid != nil {
			mid := *qi.Mid
			hs[i].Mid = &mid
		}
		if qi.Minor != nil {
			minor := *qi.Minor
			hs[i].Minor = &minor
		}
	}
	return hs
}

func computeNaYin(bz ganzhi.Bazi) [4]string {
	var ny [4]string
	for i, z := range bz.Slice() {
		ny[i] = ganzhi.NayinLabel(z.Gan, z.Zhi)
	}
	return ny
}

func TenGodProjections(bz ganzhi.Bazi, hs [4]HiddenStemProjection) [4][]TenGodProjection {
	dm := bz.Ri.Gan
	var table [4][]TenGodProjection
	ps := bz.Slice()
	for i := range ps {
		var entries []TenGodProjection
		entries = append(entries, TenGodProjection{
			ShiShen: ganzhi.ShiShenFromGan(dm, ps[i].Gan),
			Name:    ganzhi.GanName(ps[i].Gan),
			Source:  sourceGan,
			Gan:     ps[i].Gan,
		})
		entries = append(entries, TenGodProjection{
			ShiShen: ganzhi.ShiShenFromGan(dm, hs[i].Main),
			Name:    ganzhi.GanName(hs[i].Main),
			Source:  sourceMainQi,
			Gan:     hs[i].Main,
		})
		if hs[i].Mid != nil {
			entries = append(entries, TenGodProjection{
				ShiShen: ganzhi.ShiShenFromGan(dm, *hs[i].Mid),
				Name:    ganzhi.GanName(*hs[i].Mid),
				Source:  sourceMidQi,
				Gan:     *hs[i].Mid,
			})
		}
		if hs[i].Minor != nil {
			entries = append(entries, TenGodProjection{
				ShiShen: ganzhi.ShiShenFromGan(dm, *hs[i].Minor),
				Name:    ganzhi.GanName(*hs[i].Minor),
				Source:  sourceMinQi,
				Gan:     *hs[i].Minor,
			})
		}
		table[i] = entries
	}
	return table
}

func computeSelfSitting(bz ganzhi.Bazi) [4]string {
	var table [4]string
	for i, z := range bz.Slice() {
		stages, ok := ganzhi.ChangShengTable[z.Gan]
		if !ok {
			continue
		}
		for stageIdx, b := range stages {
			if b == z.Zhi {
				table[i] = ganzhi.StageNamesZH[stageIdx]
				break
			}
		}
	}
	return table
}
