package bazi

import "testing"

// TestBaziGolden4_AllLayers verifies 第四轮 golden（bazi_golden4.json，24 例）：
// 补全 12 时辰覆盖（丑/寅/卯/酉/戌）、时辰切换边界（亥→子/子→丑/午→未/未→申）、
// 农历正月初一/除夕、闰月位置（闰二月/闰四月）、立春当天起运边界、大运全步骤。
func TestBaziGolden4_AllLayers(t *testing.T) {
	for _, gc := range loadBaziGoldenFile(t, "testdata/bazi_golden4.json") {
		t.Run(gc.Name, func(t *testing.T) {
			verifyBaziGoldenCase(t, gc, goldenChecks{})
		})
	}
}
