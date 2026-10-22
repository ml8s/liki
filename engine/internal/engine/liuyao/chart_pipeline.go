package liuyao

import (
	"liki-engine/internal/engine/ganzhi"
)

func validateYaoValues(yaos [6]int) bool {
	for _, value := range yaos {
		if value < 6 || value > 9 {
			return false
		}
	}
	return true
}

func newBaseChart(yaos [6]int, bz ganzhi.Bazi) Chart {
	chart := computeGuaPan(shakeCoinsFixed(yaos), bz.Ri)
	chart.TombSchool = TombSchool{EarthBranch: "辰", School: "engine_default_chen"}
	chart.YueZhi = bz.Yue.Zhi
	chart.YueGan = bz.Yue.Gan
	chart.XunKong = ganzhi.XunKong(bz.Ri.Gan, bz.Ri.Zhi)
	return chart
}

func applyYaoStrength(chart *Chart) {
	for i := 0; i < 6; i++ {
		chart.WangShuai[i] = ganzhi.WangShuaiOf(
			ganzhi.ZhiWuxing(chart.Lines[i].Zhi),
			chart.YueZhi,
		)
		chart.DayRelations[i] = dayInteraction(chart.Lines[i].Zhi, chart.RiZhi)
	}
	computeLineDerived(chart)
}

func resolveYongshen(chart *Chart, yongShen YongShen) {
	position := chart.findYongShen(yongShen)
	chart.YongShen = YongShenResult{Name: yongShen.String(), Position: position}

	if position == 0 {
		chart.YongShen.FuShen = chart.findFuShen(yongShen)
		chart.YongShen.FuShenCandidates = chart.findFuShenAll(yongShen)
		resolveHiddenYongshen(chart, position)
		return
	}

	line := chart.Lines[position-1]
	chart.YongShen.WangShuai = chart.WangShuai[position-1].String()
	chart.YongShen.YuePo = line.YuePo
	chart.YongShen.XunKong = line.XunKong
	chart.YongShen.MuKu = line.MuKu
	chart.YongShen.LiuShou = line.LiuShou
	chart.YongShen.ChangSheng = line.ChangShengYue
}

func resolveHiddenYongshen(chart *Chart, position int) {
	if position != 0 {
		return
	}
	fuShen := chart.YongShen.FuShen
	if fuShen == nil {
		return
	}

	fuZhi := fuShenZhi(fuShen)
	fuElement := ganzhi.ZhiWuxing(fuZhi)
	chart.YongShen.IsHidden = true
	chart.YongShen.WangShuai = ganzhi.WangShuaiOf(fuElement, chart.YueZhi).String()
	chart.YongShen.YuePo = ganzhi.IsLiuChong(fuZhi, chart.YueZhi)
	chart.YongShen.XunKong = fuZhi == chart.XunKong[0] || fuZhi == chart.XunKong[1]
	chart.YongShen.MuKu = len(fuTombSources(chart, fuElement)) > 0
	chart.YongShen.ChangSheng = lifeStageOf(fuElement, chart.YueZhi)
}

func deriveAdvancedFacts(chart *Chart, yongShen YongShen) {
	chart.DongYaoRelations = computeDongYaoRelations(chart, yongShen)
	chart.Patterns = ComputePatterns(chart, yongShen)
	chart.TimingCandidates = computeTimingCandidates(chart, yongShen)
	chart.DayClashFacts = computeDayClashFacts(chart)
	chart.MovingTransformations = computeMovingTransformations(chart)
	chart.SanHeCandidates = computeSanHeCandidates(chart)
	chart.ForceChain = computeForceChain(chart, yongShen)
	chart.YongShenCandidates = computeYongShenCandidates(chart, yongShen)
	chart.HiddenLines = computeHiddenLines(chart)
	chart.BranchRelationFacts = computeBranchRelationFacts(chart)
	chart.Conflicts = computeConflicts(chart)
}

func attachGuaCi(chart *Chart) {
	if guaCi, err := GetGuaCi(int(chart.BenGua)); err == nil {
		chart.GuaCi = guaCi
	}
}
