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

func BenchmarkComputeJudgment(b *testing.B) {
	st := tianwen.GregorianToSolar(
		time.Date(2026, 6, 28, 12, 0, 0, 0, cst),
		116.4, 8,
	)
	chart := ComputeChart(st, ShiQiMen)
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		ComputeJudgment(chart, EventGeneral)
	}
}

func BenchmarkComputeTimeSelection(b *testing.B) {
	start := time.Date(2026, 7, 20, 0, 0, 0, 0, cst)
	end := time.Date(2026, 7, 21, 0, 0, 0, 0, cst)
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		ComputeTimeSelection(start, end, EventTravel, 3, 116.4, 8)
	}
}
