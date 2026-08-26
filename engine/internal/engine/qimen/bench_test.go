package qimen

import (
	"testing"
	"time"

	"liki-engine/internal/engine/tianwen"
)

var cst = time.FixedZone("CST", 8*3600)

func BenchmarkComputeChart(b *testing.B) {
	st := tianwen.GregorianToSolar(
		time.Date(2026, 6, 28, 12, 0, 0, 0, cst),
		116.4, 8,
	)
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		ComputeChart(st, ShiQiMen)
	}
}
