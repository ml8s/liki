package bazhai

import (
	"testing"

	"liki-engine/internal/engine/ganzhi"
)

func BenchmarkComputeChart(b *testing.B) {
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		ComputeChart(ganzhi.Male, 1984)
	}
}

func BenchmarkComputeLayout(b *testing.B) {
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		ComputeLayout("兑", "乾", "坤", "艮")
	}
}
