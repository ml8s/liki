package ziwei

import (
	"testing"
	"time"

	"liki-engine/internal/engine/ganzhi"
	"liki-engine/internal/engine/tianwen"
)

func BenchmarkComputeChart(b *testing.B) {
	st := tianwen.ApparentSolarFromCivil(
		time.Date(1990, 6, 1, 12, 0, 0, 0, time.FixedZone("CST", 8*3600)), 116.4, 8)
	lt := lunarTimeFromSolar(st)
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_, _ = ComputeChart(lt, ganzhi.Male)
	}
}
