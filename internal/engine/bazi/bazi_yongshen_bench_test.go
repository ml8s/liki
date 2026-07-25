package bazi

import (
	"testing"
	"time"

	"liki-engine/internal/engine/ganzhi"
	"liki-engine/internal/engine/tianwen"
)

func BenchmarkComputeYongShen(b *testing.B) {
	st := tianwen.GregorianToSolar(
		time.Date(1984, 2, 15, 8, 0, 0, 0, cst),
		116.4, 8,
	)
	chart := ComputeChart(st, ganzhi.Male)
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		ComputeYongShen(chart)
	}
}
