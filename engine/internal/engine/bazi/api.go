// Package bazi provides 八字 computation.
//
// Types
//
//	Chart, NatalCore,
//	DaYunStep, DaYun,
//	Bond,
//	LiuNian, LiuYue, LiuRi, LiuShi,
//	XiaoYunZhu,
//	FuYiResult, TiaoHouResult, GeJuResult
package bazi

import (
	"fmt"
	"time"

	"liki-engine/internal/engine/ganzhi"
	"liki-engine/internal/engine/tianwen"
)

// ComputeChart produces the complete domain chart from civil birth time
// (年/月柱) and true-solar time (日/时柱).
func ComputeChart(gt time.Time, st time.Time, gender ganzhi.Gender) Chart {
	return ComputeChartFromCore(computeNatalCore(gt, st, tianwen.ComputeBazi(gt, st), gender))
}

// ComputeChartAt is the standard-time entry for new callers.
func ComputeChartAt(
	standard time.Time,
	longitude float64,
	gender ganzhi.Gender,
) Chart {
	solar := tianwen.ApparentSolarFromStandard(standard, longitude)
	return ComputeChart(standard, solar, gender)
}

// ComputeNatalCoreAt is the standard-time entry for derived-method callers.
func ComputeNatalCoreAt(
	standard time.Time,
	longitude float64,
	gender ganzhi.Gender,
) NatalCore {
	solar := tianwen.ApparentSolarFromStandard(standard, longitude)
	return ComputeNatalCore(standard, solar, gender)
}

// ComputeNatalCore produces the narrow facts consumed by derived methods.
func ComputeNatalCore(gt time.Time, st time.Time, gender ganzhi.Gender) NatalCore {
	return computeNatalCore(gt, st, tianwen.ComputeBazi(gt, st), gender)
}

// ComputeChartFromCore fills the complete chart from a narrow natal core.
func ComputeChartFromCore(core NatalCore) Chart {
	bz := core.ToBazi()
	analysis := AnalyzePillars(bz, core.Gender)
	chart := assembleChart(core, bz, analysis)
	fillDaYunRoots(&chart, bz)
	applyChartExtra(&chart, core)
	applyInteractions(&chart, bz)
	applyYongshenEvidence(&chart, core, analysis)
	applyDerivedFacts(&chart)
	chart.ZiShiRule = "晚子时(23:00-24:00)日柱不变、时柱按次日日干起（lunar 约定）"
	return chart
}

// applyChartExtra derives the supplementary 三元 / 胎元 / 命宫 / 纳音 projections.
func applyChartExtra(chart *Chart, core NatalCore) {
	applyChartExtraFields(chart, core)
	chart.Zodiac = ganzhi.ZhiAnimal(core.ToBazi().Nian.Zhi)
}

// applyInteractions derives the complete stem/branch interaction groups.
func applyInteractions(chart *Chart, bz ganzhi.Bazi) {
	hehui := ComputeHeHui(bz)
	chart.GanHe = hehui.GanHe
	chart.GanChong = hehui.GanChong
	chart.ZhiLiuHe = hehui.ZhiLiuHe
	chart.SanHe = hehui.SanHe
	chart.SanHePartial = hehui.SanHePartial
	chart.SanHui = hehui.SanHui
	chart.LiuChong = hehui.LiuChong
	chart.LiuHai = hehui.LiuHai
	chart.LiuXing = hehui.LiuXing
	chart.LiuPo = hehui.LiuPo
	chart.AnHe = hehui.AnHe
}

// applyYongshenEvidence derives FuYi/TiaoHou candidates and GeJu structure.
func applyYongshenEvidence(chart *Chart, core NatalCore, analysis PillarAnalysis) {
	chart.FuYi, chart.TiaoHou, chart.GeJu = computeYongShenEvidenceFromAnalysis(core, analysis)
	chart.ShenShaSchool = ShenShaSchool{
		DualReference: []string{"天乙贵人", "桃花", "驿马", "华盖", "将星", "劫煞", "灾煞"},
		Policy:        "union_of_year_and_day_references",
	}
}

// applyDerivedFacts derives the read-only facts consumed by counsel factors.
func applyDerivedFacts(chart *Chart) {
	var elementStates map[string]ElementState
	chart.TenGodStates, elementStates = computeTenGodStates(*chart)
	chart.ElementStates = elementStateList(elementStates)
	chart.LuRoots = computeLuRoots(*chart)
	chart.GuanXi = computeGuanXi(*chart)
	facts := computeChartConclusions(*chart)
	chart.GuanSha = GuanSha{QuQing: facts.OfficerKillingCleaned}
	chart.CaiKu = CaiKu{Xian: facts.WealthTombPresent, XingRuKu: facts.WealthStarInTomb}
	chart.GeJu.GeShenTou = facts.PatternGodTransparent
}

func fillDaYunRoots(chart *Chart, bz ganzhi.Bazi) {
	if chart.DaYun == nil {
		return
	}
	for index := range chart.DaYun.Steps {
		step := &chart.DaYun.Steps[index]
		step.RootRefs = daYunRootRefs(ganzhi.Zhu{Gan: step.Gan, Zhi: step.Zhi}, bz)
		rooted := len(step.RootRefs) > 0
		step.Rooted = &rooted
	}
}

// ComputeLiuNianAtTime computes the year pillar from a resolved 立春 year.
func ComputeLiuNianAtTime(core NatalCore, t LiuNianTime) (*LiuNian, error) {
	cd := dayunStepForYear(core.DaYun, core.BirthYear, t.LichunYear)
	return computeLiuNian(core.ToBazi(), t, cd)
}

// ComputeLiuYueAtTime computes the month pillar from a resolved solar-term month.
func ComputeLiuYueAtTime(core NatalCore, t LiuYueTime) (*LiuYue, error) {
	return computeLiuYue(core.ToBazi(), t)
}

// ComputeLiuRi computes the day pillar and its interactions with the bazi chart.
func ComputeLiuRi(core NatalCore, t LiuRiTime) (*LiuRi, error) {
	// 流日的流年/大运按该日所属的立春年取，而不是公历年。
	nianTime, err := ResolveLiuNianTime(t.Date)
	if err != nil {
		return nil, fmt.Errorf("computeLiuRi: liunian time: %w", err)
	}
	ds := dayunStepForYear(core.DaYun, core.BirthYear, nianTime.LichunYear)
	var dzZhu *ganzhi.Zhu
	if ds != nil {
		dzZhu = &ganzhi.Zhu{Gan: ds.Gan, Zhi: ds.Zhi}
	}
	ln, err := ComputeLiuNianAtTime(core, nianTime)
	if err != nil {
		return nil, fmt.Errorf("computeLiuRi: liunian: %w", err)
	}
	var lnZhu *ganzhi.Zhu
	if ln != nil {
		lnZhu = &ganzhi.Zhu{Gan: ln.NianGan, Zhi: ln.NianZhi}
	}
	return computeLiuRi(core.ToBazi(), t, dzZhu, lnZhu)
}

// ComputeLiuShi computes the hour pillar and its interactions with the bazi chart.
func ComputeLiuShi(core NatalCore, t LiuShiTime) (*LiuShi, error) {
	return computeLiuShi(core.ToBazi(), t)
}

// dayunStepForYear returns the DaYun step governing the given year, based on
// 虚岁 = year - birthYear + 1 matching the step's qi_sui/zhi_sui (虚岁) interval.
// Returns nil when: dy is nil / empty, birthYear unknown (0), year earlier than
// birth year, not yet in 大运, or past
// all steps.
func dayunStepForYear(dy *DaYun, birthYear, year int) *DaYunStep {
	if dy == nil || len(dy.Steps) == 0 || birthYear <= 0 || year < birthYear {
		return nil
	}
	for i := range dy.Steps {
		if year >= dy.Steps[i].StartYear && year <= dy.Steps[i].EndYear {
			return &dy.Steps[i]
		}
	}
	return nil
}

// ComputeXiaoYun computes the minor fortune (小运) by multiple schools.
func ComputeXiaoYun(core NatalCore, maxAge int) []XiaoYunSet {
	return computeXiaoYun(core.ToBazi(), core.Gender, maxAge)
}
