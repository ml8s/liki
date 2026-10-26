package bazi

import "testing"

// TestBaziGolden3_AllLayers verifies 第三轮 golden（bazi_golden3.json，32 例）：
// 补全更多海外经度、南半球立春反向、极端年份（1900–2050）、闰月、晚子时。
func TestBaziGolden3_AllLayers(t *testing.T) {
	for _, gc := range loadBaziGoldenFile(t, "testdata/bazi_golden3.json") {
		t.Run(gc.Name, func(t *testing.T) {
			verifyBaziGoldenCase(t, gc, goldenChecks{MingGong: true})
		})
	}
}
