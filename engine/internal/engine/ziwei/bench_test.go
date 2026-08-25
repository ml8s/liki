package ziwei

import (
	"testing"
	"time"

	"liki-engine/internal/engine/ganzhi"
	"liki-engine/internal/engine/tianwen"
)

func BenchmarkComputeChart(b *testing.B) {
	st := tianwen.GregorianToSolar(
		time.Date(1990, 6, 1, 12, 0, 0, 0, time.FixedZone("CST", 8*3600)), 116.4, 8)
	lt := tianwen.SolarToLunar(tianwen.GregorianTime(st.Time()))
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		ComputeChart(lt, ganzhi.Male)
	}
}

