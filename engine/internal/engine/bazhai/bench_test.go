package bazhai

import (
	"testing"
	"time"

	"liki-engine/internal/engine/ganzhi"
	"liki-engine/internal/engine/tianwen"
)

func BenchmarkComputeChart(b *testing.B) {
	st := tianwen.GregorianToSolar(
		time.Date(1984, 2, 15, 8, 0, 0, 0, time.FixedZone("CST", 8*3600)),
		116.4, 8)
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		ComputeChart(st, ganzhi.Male)
	}
}

func BenchmarkComputeLayout(b *testing.B) {
	st := tianwen.GregorianToSolar(
		time.Date(1984, 2, 15, 8, 0, 0, 0, time.FixedZone("CST", 8*3600)),
		116.4, 8)
	chart := ComputeChart(st, ganzhi.Male)
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		ComputeLayout(chart, "乾", "坤", "离")
	}
}
