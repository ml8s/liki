package bazi

import (
	"testing"
	"time"

	"liki-engine/internal/engine/ganzhi"
	"liki-engine/internal/engine/tianwen"
)

var cst = time.FixedZone("CST", 8*3600)

func BenchmarkComputeChart(b *testing.B) {
	stgt := time.Time(time.Date(1984, 2, 15, 8, 0, 0, 0, cst))
	st := tianwen.ApparentSolarFromCivil(
		time.Date(1984, 2, 15, 8, 0, 0, 0, cst),
		116.4, 8,
	)
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		ComputeChart(stgt, st, ganzhi.Male)
	}
}

func BenchmarkComputeBond(b *testing.B) {
	stAgt := time.Time(time.Date(1984, 2, 15, 8, 0, 0, 0, cst))
	stA := tianwen.ApparentSolarFromCivil(
		time.Date(1984, 2, 15, 8, 0, 0, 0, cst),
		116.4, 8,
	)
	stBgt := time.Time(time.Date(1986, 8, 20, 12, 0, 0, 0, cst))
	stB := tianwen.ApparentSolarFromCivil(
		time.Date(1986, 8, 20, 12, 0, 0, 0, cst),
		121.5, 8,
	)
	chartA := ComputeChart(stAgt, stA, ganzhi.Male)
	chartB := ComputeChart(stBgt, stB, ganzhi.Female)
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		ComputeBond(chartA, chartB)
	}
}

func BenchmarkComputeLiuNian(b *testing.B) {
	stgt := time.Time(time.Date(1984, 2, 15, 8, 0, 0, 0, cst))
	st := tianwen.ApparentSolarFromCivil(
		time.Date(1984, 2, 15, 8, 0, 0, 0, cst),
		116.4, 8,
	)
	chart := ComputeChart(stgt, st, ganzhi.Male)
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		if _, err := ComputeLiuNian(chart, 2026); err != nil {
			b.Fatal(err)
		}
	}
}

func BenchmarkComputeXiaoYun(b *testing.B) {
	stgt := time.Time(time.Date(1984, 2, 15, 8, 0, 0, 0, cst))
	st := tianwen.ApparentSolarFromCivil(
		time.Date(1984, 2, 15, 8, 0, 0, 0, cst),
		116.4, 8,
	)
	chart := ComputeChart(stgt, st, ganzhi.Male)
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		ComputeXiaoYun(chart, 120)
	}
}
