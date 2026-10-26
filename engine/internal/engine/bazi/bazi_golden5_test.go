package bazi

import "testing"

// TestBaziGolden5_AllLayers verifies 第五轮 golden（bazi_golden5.json）：
// 三元（命宫/身宫/胎元）+ 起运月/日起运精度（月/日级）。
func TestBaziGolden5_AllLayers(t *testing.T) {
	for _, gc := range loadBaziGoldenFile(t, "testdata/bazi_golden5.json") {
		t.Run(gc.Name, func(t *testing.T) {
			verifyBaziGoldenCase(t, gc, goldenChecks{
				MingGong: true, ShenGong: true, TaiYuan: true, QiYunMonthDay: true,
			})
		})
	}
}
