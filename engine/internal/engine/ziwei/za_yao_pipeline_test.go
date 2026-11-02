package ziwei

import (
	"encoding/json"
	"testing"
)

// TestChartPipelineDeterministic guards against map-iteration ordering bugs:
// za_yao (and the resulting chart digest) must be identical across runs.
func TestChartPipelineDeterministic(t *testing.T) {
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
	first := dump(fillChartAreas(base))
	second := dump(fillChartAreas(base))
	if first != second {
		t.Fatalf("chart pipeline not deterministic: za_yao order unstable")
	}
}
