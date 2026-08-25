package tianwen

import (
	"testing"
	"time"
)

func BenchmarkSolarToLunar(b *testing.B) {
	gt := GregorianTime(time.Date(2026, 6, 28, 12, 0, 0, 0, time.UTC))
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		SolarToLunar(gt)
	}
}

func BenchmarkGregorianToSolar(b *testing.B) {
	t := time.Date(2026, 6, 28, 12, 0, 0, 0, time.FixedZone("CST", 8*3600))
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		GregorianToSolar(t, 116.4, 8)
	}
}
