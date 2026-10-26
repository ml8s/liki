package ziwei

import (
	"testing"
	"time"

	"liki-engine/internal/engine/tianwen"
)

func benchLiuNianChart(b *testing.B) Chart {
	b.Helper()
	st := tianwen.ApparentSolarFromCivil(
		time.Date(1990, 6, 1, 12, 0, 0, 0, time.FixedZone("CST", 8*3600)), 116.4, 8)
	lt := lunarTimeFromSolar(st)
	c, err := ComputeChart(lt, Male)
	if err != nil {
		b.Fatal(err)
	}
	return c
}

func BenchmarkComputeLiuNian(b *testing.B) {
	chart := benchLiuNianChart(b)
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		ComputeLiuNian(chart, 2026)
	}
}
