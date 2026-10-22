package bazi

import (
	"testing"
	"time"

	"liki-engine/internal/engine/ganzhi"
	"liki-engine/internal/engine/tianwen"
)

func BenchmarkComputeYongShenSchools(b *testing.B) {
	stgt := time.Time(time.Date(1984, 2, 15, 8, 0, 0, 0, cst))
	st := tianwen.ApparentSolarFromCivil(
		time.Date(1984, 2, 15, 8, 0, 0, 0, cst),
		116.4, 8,
	)
	chart := ComputeChart(stgt, st, ganzhi.Male)
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_, _, _ = ComputeYongShenSchools(chart)
	}
}
