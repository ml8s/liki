package qiming

import (
	"testing"
)

func BenchmarkComposeNames(b *testing.B) {
	yongChars, err := GetChars("金")
	if err != nil || len(yongChars) == 0 {
		b.Skip("chars returned empty")
	}
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		ComposeNames("王", yongChars, yongChars, nil)
	}
}
