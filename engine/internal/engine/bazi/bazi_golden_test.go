package bazi

import "testing"

// TestBaziGolden_AllLayers verifies 四柱/大运/三元（命宫/身宫/胎元）against
// lunar-typescript golden. 覆盖：常规/立春边界/晚子时/早子时/闰月/海外。
func TestBaziGolden_AllLayers(t *testing.T) {
	for _, gc := range loadBaziGoldenFile(t, "testdata/bazi_golden.json") {
		t.Run(gc.Name, func(t *testing.T) {
			verifyBaziGoldenCase(t, gc, goldenChecks{MingGong: true, ShenGong: true, TaiYuan: true})
		})
	}
}
