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
	return fc
}

// ComputeLiuNian computes the year pillar and its interactions with the bazi chart.
func ComputeLiuNian(chart Chart, year int) (*LiuNian, error) {
	cd := currentDaYunStep(chart.DaYun)
	return computeLiuNian(chart.ToBazi(), year, cd)
}

// ComputeLiuYue computes the month pillar and its interactions with the bazi chart.
func ComputeLiuYue(c Chart, year, month int) (*LiuYue, error) {
	return computeLiuYue(c.ToBazi(), year, month)
}

// ComputeLiuRi computes the day pillar and its interactions with the bazi chart.
func ComputeLiuRi(c Chart, year, month, day int) (*LiuRi, error) {
	ds := currentDaYunStep(c.DaYun)
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
		lnZhu = &ganzhi.Zhu{Gan: ln.YearGan, Zhi: ln.YearZhi}
	}
	return computeLiuRi(c.ToBazi(), year, month, day, dzZhu, lnZhu)
}

// ComputeLiuShi computes the hour pillar and its interactions with the bazi chart.
func ComputeLiuShi(c Chart, year, month, day, hour int) (*LiuShi, error) {
	return computeLiuShi(c.ToBazi(), year, month, day, hour)
}

// currentDaYunStep returns the current DaYun step matching current_step_index, or nil if not applicable.
func currentDaYunStep(dy *DaYun) *DaYunStep {
	if dy == nil || dy.CurrentStepIndex < 0 || dy.CurrentStepIndex >= len(dy.Steps) {
		return nil
	}
	return &dy.Steps[dy.CurrentStepIndex]
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
