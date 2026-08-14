package xuankong

import (
	"testing"
	"time"

	"liki-engine/internal/engine/tianwen"
)

func BenchmarkComputeChart(b *testing.B) {
	st := tianwen.GregorianToSolar(
		time.Date(2020, 6, 1, 12, 0, 0, 0, time.FixedZone("CST", 8*3600)),
		116.4, 8)
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		ComputeChart(st, 0, 12)
	}
}

func BenchmarkComputeLiuNian(b *testing.B) {
	st := tianwen.GregorianToSolar(
		time.Date(2020, 6, 1, 12, 0, 0, 0, time.FixedZone("CST", 8*3600)),
		116.4, 8)
	chart := ComputeChart(st, 0, 12)
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		ComputeLiuNian(2027, &chart)
	}
}
