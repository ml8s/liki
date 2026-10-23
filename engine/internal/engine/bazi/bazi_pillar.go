package bazi

import (
	"liki-engine/internal/engine/ganzhi"
	"time"
)

func computeChartCore(gt time.Time, st time.Time, bz ganzhi.Bazi, gender ganzhi.Gender) Chart {
	ny := computeNaYin(bz)
	ps := bz.Slice()

	makePI := func(i int) zhuInfo {
		return zhuInfo{Zhu: ps[i], NaYin: ny[i]}
	}

	cr := Chart{
		Nian:   makePI(0),
		Yue:    makePI(1),
		Ri:     makePI(2),
		Shi:    makePI(3),
		Gender: gender,
	}
	cr.DaYun = computeDaYun(gt, bz, gender)
	cr.BirthYear = gt.Year()
	cr.ZiShiRule = "晚子时(23:00-24:00)日柱不变、时柱按次日日干起（lunar 约定）"
	return cr
}

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
	return PillarAnalysis{
		HiddenStems: HiddenStemFacts(bz),
		NaYin:       computeNaYin(bz),
		TenGods:     TenGodProjections(bz, HiddenStemFacts(bz)),
		SelfSitting: computeSelfSitting(bz),
		LifeStages:  computeDayMasterTrend(bz),
		ShenSha:     ShenShaProjections(bz, gender),
		VoidPalaces: VoidPalaces(bz),
	}
}

func computeFullFromCore(c Chart, bz ganzhi.Bazi) FullChart {
	analysis := AnalyzePillars(bz, c.Gender)
	return assembleFullChart(c, bz, analysis)
}

// assembleFullChart projects the pillar analysis into the published FullChart.
func assembleFullChart(c Chart, bz ganzhi.Bazi, analysis PillarAnalysis) FullChart {
	hs := analysis.HiddenStems
	ny := analysis.NaYin
	tgTable := analysis.TenGods
	selfSitting := analysis.SelfSitting
	dayMasterTrend := analysis.LifeStages
	shensha := analysis.ShenSha
	voidHits := analysis.VoidPalaces
	ps := bz.Slice()

	makeFull := func(i int) fullZhuInfo {
		isVoid := false
		for _, vh := range voidHits {
			if vh == i {
				isVoid = true
				break
			}
		}
		pi := fullZhuInfo{
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

	return FullChart{
		BirthYear:  c.BirthYear,
		Nian:       makeFull(0),
		Yue:        makeFull(1),
		Ri:         makeFull(2),
		Shi:        makeFull(3),
		DaYun:      c.DaYun,
		Gender:     c.Gender,
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

func computeElementCount(bz ganzhi.Bazi, hs [4]HiddenStemProjection) map[ganzhi.Wuxing]int {
	wc := make(map[ganzhi.Wuxing]int, 5)
	for _, z := range bz.Slice() {
		wc[ganzhi.GanWuxing(z.Gan)]++
	}
	for _, h := range hs {
		wc[ganzhi.GanWuxing(h.Main)]++
		if h.Mid != nil {
			wc[ganzhi.GanWuxing(*h.Mid)]++
		}
		if h.Minor != nil {
			wc[ganzhi.GanWuxing(*h.Minor)]++
		}
	}
	return wc
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
