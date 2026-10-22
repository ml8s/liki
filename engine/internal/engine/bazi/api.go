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

// ComputeChartAt is the explicit-time entry for new callers. at must already
// be normalized to UTC; loc is the explicit IANA location.
func ComputeChartAt(
	at time.Time,
	loc *time.Location,
	longitude float64,
	gender ganzhi.Gender,
) Chart {
	civil := at.In(loc)
	solar := tianwen.ApparentSolarCivil(at, loc, longitude)
	return ComputeChart(
		time.Time(civil),
		time.Time(solar),
		gender,
	)
}

// ComputeFullChart expands a lean Chart to FullChart with all fields (十神/藏干/神煞/长生/空亡/自合/魁罡).
func ComputeFullChart(c Chart) FullChart {
	bz := c.ToBazi()
	fc := computeFullFromCore(c, bz)
	enrichDaYunRoots(&fc, bz)
	applyChartExtra(&fc, c)
	applyInteractions(&fc, c)
	applyYongshenSchools(&fc, c)
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
func applyYongshenSchools(fc *FullChart, c Chart) {
	fc.FuYi, fc.TiaoHou, fc.GeJu = ComputeYongShenSchools(c)
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
func ComputeLiuNian(chart Chart, year int) (*LiuNian, error) {
	cd := dayunStepForYear(chart.DaYun, chart.BirthYear, year)
	return computeLiuNian(chart.ToBazi(), year, cd)
}

// ComputeLiuYue computes the month pillar and its interactions with the bazi chart.
func ComputeLiuYue(c Chart, year, month int) (*LiuYue, error) {
	return computeLiuYue(c.ToBazi(), year, month)
}

// ComputeLiuRi computes the day pillar and its interactions with the bazi chart.
func ComputeLiuRi(c Chart, year, month, day int) (*LiuRi, error) {
	ds := dayunStepForYear(c.DaYun, c.BirthYear, year)
	var dzZhu *ganzhi.Zhu
	if ds != nil {
		dzZhu = &ganzhi.Zhu{Gan: ds.Gan, Zhi: ds.Zhi}
	}
	ln, err := ComputeLiuNian(c, year)
	if err != nil {
		return nil, fmt.Errorf("computeLiuRi: liunian: %w", err)
	}
	var lnZhu *ganzhi.Zhu
	if ln != nil {
		lnZhu = &ganzhi.Zhu{Gan: ln.NianGan, Zhi: ln.NianZhi}
	}
	return computeLiuRi(c.ToBazi(), year, month, day, dzZhu, lnZhu)
}

// ComputeLiuShi computes the hour pillar and its interactions with the bazi chart.
func ComputeLiuShi(c Chart, year, month, day, hour int) (*LiuShi, error) {
	return computeLiuShi(c.ToBazi(), year, month, day, hour)
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
func ComputeDaYunAt(
	at time.Time,
	loc *time.Location,
	longitude float64,
	gender ganzhi.Gender,
) (*DaYun, error) {
	chart := ComputeChartAt(at, loc, longitude, gender)
	if chart.DaYun == nil {
		return nil, fmt.Errorf("dayun is unavailable")
	}
	return chart.DaYun, nil
}

// ComputeLiuNianAt is the explicit-time entry for a year's flow analysis.
func ComputeLiuNianAt(
	at time.Time,
	loc *time.Location,
	longitude float64,
	year int,
) (*LiuNian, error) {
	chart := ComputeChartAt(at, loc, longitude, ganzhi.Gender(""))
	return ComputeLiuNian(chart, year)
}

// ComputeLiuYueAt is the explicit-time entry for a month's flow analysis.
func ComputeLiuYueAt(
	at time.Time,
	loc *time.Location,
	longitude float64,
	year, month int,
) (*LiuYue, error) {
	chart := ComputeChartAt(at, loc, longitude, ganzhi.Gender(""))
	return ComputeLiuYue(chart, year, month)
}

// ComputeLiuRiAt is the explicit-time entry for a day's flow analysis.
func ComputeLiuRiAt(
	at time.Time,
	loc *time.Location,
	longitude float64,
	year, month, day int,
) (*LiuRi, error) {
	chart := ComputeChartAt(at, loc, longitude, ganzhi.Gender(""))
	return ComputeLiuRi(chart, year, month, day)
}

// ComputeLiuShiAt is the explicit-time entry for an hour's flow analysis.
func ComputeLiuShiAt(
	at time.Time,
	loc *time.Location,
	longitude float64,
	year, month, day, hour int,
) (*LiuShi, error) {
	chart := ComputeChartAt(at, loc, longitude, ganzhi.Gender(""))
	return ComputeLiuShi(chart, year, month, day, hour)
}

// ComputeXiaoYunAt is the explicit-time entry for minor-fortune analysis.
func ComputeXiaoYunAt(
	at time.Time,
	loc *time.Location,
	longitude float64,
	maxAge int,
) []XiaoYunSet {
	chart := ComputeChartAt(at, loc, longitude, ganzhi.Male)
	return ComputeXiaoYun(chart, maxAge)
}
