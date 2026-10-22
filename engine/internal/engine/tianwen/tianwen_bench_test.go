package tianwen

import (
	"testing"
	"time"
)

func BenchmarkLunarDateAt(b *testing.B) {
	gt := time.Time(time.Date(2026, 6, 28, 12, 0, 0, 0, time.UTC))
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		LunarDateAt(gt)
	}
}

func BenchmarkApparentSolarFromCivil(b *testing.B) {
	t := time.Date(2026, 6, 28, 12, 0, 0, 0, time.FixedZone("CST", 8*3600))
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		ApparentSolarFromCivil(t, 116.4, 8)
	}
}
