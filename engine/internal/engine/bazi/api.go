// Package bazi provides 八字 computation.
//
// Types
//
//	Chart,
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

// ComputeChart produces a lean Chart (四柱+纳音+大运+性别) from civil birth time
// (年/月柱) and true-solar time (日/时柱).
// Use ComputeFullChart to expand to full chart with 十神/藏干/神煞/长生/空亡.
func ComputeChart(gt time.Time, st time.Time, gender ganzhi.Gender) Chart {
	return computeChartCore(gt, st, tianwen.ComputeBazi(gt, st), gender)
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

// NewChartFromPillars assembles the published lean chart from already decoded
// wire values. It is the typed inverse of the public chart projection and keeps
// callers from constructing bazi.Chart through a JSON round trip.
func NewChartFromPillars(
	pillars [4]ganzhi.Zhu,
	naYin [4]string,
	daYun *DaYun,
	gender ganzhi.Gender,
	birthYear int,
	ziShiRule string,
) Chart {
	return Chart{
		Nian:      baziZhu(pillars[0], naYin[0]),
		Yue:       baziZhu(pillars[1], naYin[1]),
		Ri:        baziZhu(pillars[2], naYin[2]),
		Shi:       baziZhu(pillars[3], naYin[3]),
		DaYun:     daYun,
		Gender:    gender,
		BirthYear: birthYear,
		ZiShiRule: ziShiRule,
	}
}

func baziZhu(pillar ganzhi.Zhu, naYin string) zhuInfo {
	return zhuInfo{Zhu: pillar, NaYin: naYin}
}

// ComputeFullChart expands a lean Chart to FullChart with all fields (十神/藏干/神煞/长生/空亡/自合/魁罡).
func ComputeFullChart(c Chart) FullChart {
	bz := c.ToBazi()
	analysis := AnalyzePillars(bz, c.Gender)
	fc := assembleFullChart(c, bz, analysis)
	enrichDaYunRoots(&fc, bz)
	applyChartExtra(&fc, c)
	applyInteractions(&fc, c)
	applyYongshenSchools(&fc, c, analysis)
	applyDerivedFacts(&fc)
	return fc
}

// applyChartExtra derives the supplementary 三元 / 胎元 / 命宫 / 纳音 projections.
func applyChartExtra(fc *FullChart, c Chart) ChartExtra {
	extra := ComputeChartExtra(c)
	fc.SanYuan = extra.SanYuan
	fc.TaiXi = extra.TaiXi
	fc.GongJia = extra.GongJia
	fc.NayinRel = extra.NayinRel
	fc.ChangSheng = extra.ChangSheng
	fc.SanQiName = extra.SanQiName
	fc.Zodiac = ganzhi.ZhiAnimal(c.ToBazi().Nian.Zhi)
	return extra
}

// applyInteractions derives the complete stem/branch interaction groups.
func applyInteractions(fc *FullChart, c Chart) {
	hehui := ComputeHeHui(c)
	fc.GanHe = hehui.GanHe
	fc.GanChong = hehui.GanChong
	fc.ZhiLiuHe = hehui.ZhiLiuHe
	fc.SanHe = hehui.SanHe
	fc.SanHePartial = hehui.SanHePartial
	fc.SanHui = hehui.SanHui
	fc.LiuChong = hehui.LiuChong
	fc.LiuHai = hehui.LiuHai
	fc.LiuXing = hehui.LiuXing
	fc.LiuPo = hehui.LiuPo
	fc.AnHe = hehui.AnHe
}

// applyYongshenSchools derives the three independent yongshen evidence sources.
func applyYongshenSchools(fc *FullChart, c Chart, analysis PillarAnalysis) {
	fc.FuYi, fc.TiaoHou, fc.GeJu = computeYongShenSchoolsFromAnalysis(c, analysis)
	fc.ShenShaSchool = ShenShaSchool{
		DualReference: []string{"天乙贵人", "桃花", "驿马", "华盖", "将星", "劫煞", "灾煞"},
		Policy:        "union_of_year_and_day_references",
	}
}

// applyDerivedFacts derives the read-only facts consumed by counsel factors.
func applyDerivedFacts(fc *FullChart) {
	var elementStates map[string]ElementState
	fc.TenGodStates, elementStates = computeTenGodStates(*fc)
	fc.ElementStates = elementStateList(elementStates)
	fc.LuRoots = computeLuRoots(*fc)
	fc.RelationGroups = computeRelationGroups(*fc)
	fc.AtomicFacts = computeAtomicFacts(*fc)
}

func enrichDaYunRoots(fc *FullChart, bz ganzhi.Bazi) {
	if fc.DaYun == nil {
		return
	}
	for index := range fc.DaYun.Steps {
		step := &fc.DaYun.Steps[index]
		step.RootRefs = daYunRootRefs(ganzhi.Zhu{Gan: step.Gan, Zhi: step.Zhi}, bz)
		rooted := len(step.RootRefs) > 0
		step.Rooted = &rooted
	}
}

// ComputeLiuNian computes the year pillar and its interactions with the bazi chart.
// The year selector is transitional while the wire still carries a Gregorian
// year; it resolves at mid-year to preserve the existing contract.
func ComputeLiuNian(chart Chart, year int) (*LiuNian, error) {
	t, err := ResolveLiuNianTime(tianwen.NewCivilDate(year, 6, 15))
	if err != nil {
		return nil, err
	}
	return ComputeLiuNianAtTime(chart, t)
}

// ComputeLiuNianAtTime computes the year pillar from a resolved 立春 year.
func ComputeLiuNianAtTime(chart Chart, t LiuNianTime) (*LiuNian, error) {
	cd := dayunStepForYear(chart.DaYun, chart.BirthYear, t.LichunYear)
	return computeLiuNian(chart.ToBazi(), t, cd)
}

// ComputeLiuYue computes the month pillar and its interactions with the bazi chart.
// The year/month selector is transitional while the wire still carries a
// Gregorian month; it resolves at mid-month to preserve the existing contract.
func ComputeLiuYue(c Chart, year, month int) (*LiuYue, error) {
	t, err := ResolveLiuYueTime(tianwen.NewCivilDate(year, month, 15))
	if err != nil {
		return nil, err
	}
	return ComputeLiuYueAtTime(c, t)
}

// ComputeLiuYueAtTime computes the month pillar from a resolved solar-term month.
func ComputeLiuYueAtTime(c Chart, t LiuYueTime) (*LiuYue, error) {
	return computeLiuYue(c.ToBazi(), t)
}

// ComputeLiuRi computes the day pillar and its interactions with the bazi chart.
func ComputeLiuRi(c Chart, t LiuRiTime) (*LiuRi, error) {
	// 流日的流年/大运按该日所属的立春年取，而不是公历年。
	nianTime, err := ResolveLiuNianTime(t.Date)
	if err != nil {
		return nil, fmt.Errorf("computeLiuRi: liunian time: %w", err)
	}
	ds := dayunStepForYear(c.DaYun, c.BirthYear, nianTime.LichunYear)
	var dzZhu *ganzhi.Zhu
	if ds != nil {
		dzZhu = &ganzhi.Zhu{Gan: ds.Gan, Zhi: ds.Zhi}
	}
	ln, err := ComputeLiuNianAtTime(c, nianTime)
	if err != nil {
		return nil, fmt.Errorf("computeLiuRi: liunian: %w", err)
	}
	var lnZhu *ganzhi.Zhu
	if ln != nil {
		lnZhu = &ganzhi.Zhu{Gan: ln.NianGan, Zhi: ln.NianZhi}
	}
	return computeLiuRi(c.ToBazi(), t, dzZhu, lnZhu)
}

// ComputeLiuShi computes the hour pillar and its interactions with the bazi chart.
func ComputeLiuShi(c Chart, t LiuShiTime) (*LiuShi, error) {
	return computeLiuShi(c.ToBazi(), t)
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
func ComputeXiaoYun(c Chart, maxAge int) []XiaoYunSet {
	return computeXiaoYun(c.ToBazi(), c.Gender, maxAge)
}

// ComputeDaYunAt is the explicit-time entry for the big-fortune cycle.
// ComputeLiuNianAt is the explicit-time entry for a year's flow analysis.
// ComputeLiuYueAt is the explicit-time entry for a month's flow analysis.
// ComputeLiuRiAt is the explicit-time entry for a day's flow analysis.
// ComputeLiuShiAt is the explicit-time entry for an hour's flow analysis.
// ComputeXiaoYunAt is the explicit-time entry for minor-fortune analysis.
