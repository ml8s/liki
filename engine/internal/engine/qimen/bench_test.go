package qimen

import (
	"testing"
	"time"

	"liki-engine/internal/engine/tianwen"
)

var cst = time.FixedZone("CST", 8*3600)

// BenchmarkComputeChart 基准 wire 路径的计算核心：
// ResolveMethod（默认解析）+ ComputeChartWithMethod —— 与 qimen_chart 的
// 生产计算管线一致（wire 侧仅多出入参用神校验，不属于计算成本）。
func BenchmarkComputeChart(b *testing.B) {
	stgt := time.Time(time.Date(2026, 6, 28, 12, 0, 0, 0, cst))
	st := tianwen.ApparentSolarFromCivil(
		time.Date(2026, 6, 28, 12, 0, 0, 0, cst),
		116.4, 8,
	)
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		ComputeChart(stgt, st)
	}
}

func ComputeChart(gt time.Time, st time.Time) Chart {
	method, _ := ResolveMethod("", "", "", "")
	chart, _ := ComputeChartWithMethod(gt, st, method)
	return chart
}
