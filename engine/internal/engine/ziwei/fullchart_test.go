package ziwei

import (
	"encoding/json"
	"testing"
)

// TestComputeFullChartDeterministic guards against map-iteration ordering bugs:
// za_yao (and the resulting chart digest) must be identical across runs.
func TestComputeFullChartDeterministic(t *testing.T) {
	base := Chart{
		MingGong:   0,
		NianZhi:    9, // 酉
		ShiZhi:     1, // 丑
		LunarMonth: 4,
		LunarDay:   26,
		NianGan:    4,
		Gender:     "male",
	}
	// 12 宫地支各不相同，确保多宫杂曜参与
	for i := 0; i < 12; i++ {
		base.GongWei[i].Zhi = Zhi(i + 1)
	}
	dump := func(c Chart) string {
		b, _ := json.Marshal(c)
		return string(b)
	}
	first := dump(ComputeFullChart(base, 0, 0))
	second := dump(ComputeFullChart(base, 0, 0))
	if first != second {
		t.Fatalf("ComputeFullChart not deterministic: za_yao order unstable")
	}
}
