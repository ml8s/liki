package liuyao

import (
	"math/rand"
	"testing"
)

func FuzzShakeCoins(f *testing.F) {
	f.Add(int64(0))
	f.Add(int64(42))
	f.Add(int64(-1))
	f.Add(int64(1 << 62))

	f.Fuzz(func(t *testing.T, seed int64) {
		rng := rand.New(rand.NewSource(seed))
		yaos := shakeCoins(rng)

		for i, y := range yaos {
			if y < 6 || y > 9 {
				t.Errorf("shakeCoins(seed=%d)[%d] = %d, want 6-9", seed, i, y)
			}
		}

		// Reproducibility: same seed must produce same result
		rng2 := rand.New(rand.NewSource(seed))
		yaos2 := shakeCoins(rng2)
		if yaos != yaos2 {
			t.Errorf("shakeCoins not reproducible with seed=%d", seed)
		}
	})
}
