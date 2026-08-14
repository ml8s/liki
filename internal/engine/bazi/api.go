// Package bazi provides 八字 computation.
//
// Types
//   Chart, Chart,
//   DaYunStep, DaYun,
//   Bond, XunGong,
//   LiuNian, LiuYue, LiuRi, LiuShi,
//   XiaoYunZhu, XiaoXian,
//   YongShenResult, FuYiResult, TiaoHouResult, GeJuResult
package bazi

import (
	"fmt"

	"liki-engine/internal/engine/ganzhi"
	"liki-engine/internal/engine/tianwen"
)

// ComputeChart produces a lean Chart (四柱+纳音+大运+性别) from solar birth time and gender.
// Use ComputeFullChart to expand to full chart with 十神/藏干/神煞/长生/空亡.
func ComputeChart(st tianwen.SolarTime, gender ganzhi.Gender) Chart {
	return computeChartCore(tianwen.ComputeBazi(st), st, gender)
}

// ComputeFullChart expands a lean Chart to FullChart with all fields (十神/藏干/神煞/长生/空亡/自合/魁罡).
func ComputeFullChart(c Chart) FullChart {
	bz := c.ToBazi()
	fc := computeFullFromCore(c, bz)
	extra := ComputeChartExtra(c)
	fc.SanYuan = extra.SanYuan
	fc.GongJia = extra.GongJia
	fc.NayinRel = extra.NayinRel
	fc.ChangSheng = extra.ChangSheng
	fc.SanQiName = extra.SanQiName
	hehui := ComputeHeHui(c)
	fc.GanHe = hehui.GanHe
	fc.ZhiLiuHe = hehui.ZhiLiuHe
	fc.SanHe = hehui.SanHe
	fc.SanHui = hehui.SanHui
	fc.LiuChong = hehui.LiuChong
	fc.LiuHai = hehui.LiuHai
	fc.LiuXing = hehui.LiuXing
	fc.YongShen = ComputeYongShen(c) // 用神三派归完整命盘（chart 纯排盘不含）
	return fc
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
// Returns nil when: dy is nil / empty, birthYear unknown (0 — legacy chart
// without birth_year), year earlier than birth year, not yet in 大运, or past
// all steps.
//
// Note: differs from ComputeCurrentStepIndex (实岁, used to fill
// current_step_index); liunian/liuri 需要的是"查询年份行何运"，故按虚岁区间匹配。
func dayunStepForYear(dy *DaYun, birthYear, year int) *DaYunStep {
	if dy == nil || len(dy.Steps) == 0 || birthYear <= 0 || year < birthYear {
		return nil
	}
	age := year - birthYear + 1 // 虚岁
	for i := range dy.Steps {
		if age >= dy.Steps[i].AgeStart && age <= dy.Steps[i].AgeEnd {
			return &dy.Steps[i]
		}
	}
	return nil
}

// ComputeCurrentStepIndex determines the current DaYun step index based on age.
// Returns -1 if not yet in DaYun or past all DaYun steps.
func ComputeCurrentStepIndex(dy *DaYun, birthYear, currentYear, currentYearDay, birthYearDay int) int {
	if dy == nil || len(dy.Steps) == 0 {
		return -1
	}
	age := currentYear - birthYear
	if currentYearDay < birthYearDay {
		age--
	}
	for i, step := range dy.Steps {
		if age >= step.AgeStart && age <= step.AgeEnd {
			return i
		}
	}
	return -1
}

// ComputeXiaoYun computes the minor fortune (小运) pillars.
func ComputeXiaoYun(c Chart, maxAge int) []XiaoYunZhu {
	return computeXiaoYun(c.ToBazi(), c.Gender, maxAge)
}
